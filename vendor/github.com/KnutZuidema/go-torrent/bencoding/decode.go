package bencoding

import (
	"bytes"
	"reflect"
	"strconv"
)

// Unmarshaler if a type implements this interface the UnmarshalBEncode method will be called when unmarshaling it into
// a bencode representation.
type Unmarshaler interface {
	UnmarshalBEncode(data []byte) error
}

// Unmarshal decodes data in the bencodeing format into a given object.
//
// returns ErrEmptyData if the data is empty
//
// returns an error if the data is not a valid bencoded object
//
// returns ErrRemainingData if the data is not fully consumed after decoding is complete
//
// returns ErrNonPointer if the target is not a pointer
//
// returns ErrInvalidType if the target type is not valid for bencoding, namely float64, float32, complex128, complex64
//
// returns an error if the data does not represent an object of the target type (structs count as a single type)
//
// returns ErrNonStringKey if the key type of a map is not string
//
// to add custom decoding types may implement the Unmarshaler interface
//
// []byte is decoded like a string
//
// bool is decoded from the integer values 1 and 0
func Unmarshal(data []byte, target interface{}) error {
	n, err := decode(data, target)
	if err != nil {
		return err
	}
	if n != len(data) {
		return ErrRemainingData
	}
	return nil
}

func decode(data []byte, target interface{}) (int, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}
	switch v := target.(type) {
	case Unmarshaler:
		n, d, err := readNextValue(data)
		if err != nil {
			return 0, err
		}
		err = v.UnmarshalBEncode(d)
		if err != nil {
			return 0, err
		}
		return n, nil
	}
	val, err := valueOfTarget(target, reflect.Invalid)
	if err != nil {
		return 0, err
	}
	switch val.Kind() {
	case reflect.Bool:
		var tmp int64
		n, err := decodeInt(data, &tmp)
		if err != nil {
			return 0, err
		}
		if tmp > 1 {
			return 0, ErrInvalidBool
		}
		val.SetBool(tmp == 1)
		return n, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var tmp int64
		n, err := decodeInt(data, &tmp)
		if err != nil {
			return 0, err
		}
		val.SetInt(tmp)
		return n, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var tmp uint64
		n, err := decodeUint(data, &tmp)
		if err != nil {
			return 0, err
		}
		val.SetUint(tmp)
		return n, nil
	case reflect.String:
		return decodeString(data, target.(*string))
	case reflect.Slice:
		// []byte is decoded like a string
		if val.Type().Elem().Kind() == reflect.Uint8 {
			var tmp string
			n, err := decodeString(data, &tmp)
			if err != nil {
				return 0, err
			}
			val.SetBytes([]byte(tmp))
			return n, nil
		}
		return decodeList(data, target)
	case reflect.Map:
		return decodeDict(data, target)
	case reflect.Struct:
		return decodeStruct(data, target)
	}
	return 0, ErrInvalidType
}

func decodeString(data []byte, target *string) (int, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}
	var read int
	var length []byte
	for data[read] != stringSeparatorToken {
		if read >= len(data) {
			return 0, ErrEmptyData
		}
		switch data[read] {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			length = append(length, data[read])
		default:
			return 0, ErrInvalidString
		}
		read++
	}
	read++
	if len(length) == 0 {
		return 0, ErrEmptyData
	}
	if len(length) > 1 && length[0] == '0' {
		return 0, ErrLeadingZero
	}
	l, err := strconv.Atoi(string(length))
	if err != nil {
		return 0, err
	}
	if read+l > len(data) {
		return 0, ErrLengthTooBig
	}
	*target = string(data[read : read+l])
	return read + l, nil
}

func decodeInt(data []byte, target *int64) (int, error) {
	var tmp string
	n, err := decodeIntStr(data, &tmp)
	if err != nil {
		return 0, err
	}
	*target, err = strconv.ParseInt(tmp, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func decodeUint(data []byte, target *uint64) (int, error) {
	var tmp string
	n, err := decodeIntStr(data, &tmp)
	if err != nil {
		return 0, err
	}
	*target, err = strconv.ParseUint(tmp, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func decodeIntStr(data []byte, target *string) (int, error) {
	var read int
	if len(data) == 0 {
		return 0, ErrEmptyData
	}
	if data[read] != intToken {
		return 0, ErrInvalidInteger
	}
	read++
	var value []byte
	if data[read] == '-' {
		value = append(value, data[read])
		read++
	}
	for data[read] != endToken {
		switch data[read] {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			value = append(value, data[read])
		default:
			return 0, ErrInvalidInteger
		}
		read++
	}
	read++
	if len(value) == 0 {
		return 0, ErrEmptyData
	}
	if len(value) > 1 && value[0] == '0' {
		return 0, ErrLeadingZero
	}
	if bytes.Equal(value, []byte("-0")) {
		return 0, ErrInvalidInteger
	}
	*target = string(value)
	return read, nil
}

func decodeList(data []byte, target interface{}) (int, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}
	var read int
	if data[0] != listToken {
		return 0, ErrInvalidList
	}
	read++
	val, err := valueOfTarget(target, reflect.Slice)
	if err != nil {
		return 0, err
	}
	res := reflect.MakeSlice(reflect.SliceOf(val.Type().Elem()), 0, 0)
	for data[read] != endToken {
		v := reflect.New(val.Type().Elem())
		r, err := decode(data[read:], v.Interface())
		if err != nil {
			return 0, err
		}
		read += r
		res = reflect.Append(res, v.Elem())
	}
	read++
	val.Set(res)
	return read, nil
}

func decodeDict(data []byte, target interface{}) (int, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}
	var read int
	if data[read] != dictToken {
		return 0, ErrInvalidDict
	}
	read++
	val, err := valueOfTarget(target, reflect.Map)
	if err != nil {
		return 0, err
	}
	if val.Type().Key().Kind() != reflect.String {
		return 0, ErrNonStringKey
	}
	res := reflect.MakeMap(reflect.MapOf(reflect.TypeOf(""), val.Type().Elem()))
	for data[read] != endToken {
		var key string
		r, err := decodeString(data[read:], &key)
		if err != nil {
			return 0, err
		}
		read += r
		value := reflect.New(val.Type().Elem()).Interface()
		r, err = decode(data[read:], value)
		if err != nil {
			return 0, err
		}
		read += r
		res.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value).Elem())
	}
	read++
	val.Set(res)
	return read, nil
}

func decodeStruct(data []byte, target interface{}) (int, error) {
	val, err := valueOfTarget(target, reflect.Struct)
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, ErrEmptyData
	}
	if data[0] != dictToken {
		return 0, ErrInvalidStruct
	}
	read := 1
	nameMapping := map[string]int{}
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		tag := newTag(typ.Field(i))
		nameMapping[tag.Name] = i
	}
	for data[read] != endToken {
		var key string
		r, err := decodeString(data[read:], &key)
		if err != nil {
			return 0, err
		}
		read += r
		num, ok := nameMapping[key]
		if !ok {
			r, _, err := readNextValue(data[read:])
			if err != nil {
				return 0, err
			}
			read += r
			continue
		}
		v := reflect.New(val.Field(num).Type())
		r, err = decode(data[read:], v.Interface())
		if err != nil {
			return 0, err
		}
		read += r
		val.Field(num).Set(v.Elem())
	}
	read++
	return read, nil
}

func dereference(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	return val
}

func valueOfTarget(in interface{}, wantKind reflect.Kind) (reflect.Value, error) {
	rv := reflect.ValueOf(in)
	if !rv.IsValid() {
		return reflect.Value{}, ErrInvalidValue
	}
	if rv.Kind() != reflect.Ptr {
		return reflect.Value{}, ErrNonPointer
	}
	val := dereference(rv)
	if wantKind != reflect.Invalid && val.Kind() != wantKind {
		return reflect.Value{}, ErrInvalidType
	}
	if !val.CanSet() {
		return reflect.Value{}, ErrCanNotSet
	}
	return val, nil
}

func readNextValue(data []byte) (int, []byte, error) {
	if len(data) == 0 {
		return 0, nil, ErrEmptyData
	}
	var read int
	switch data[0] {
	case intToken:
		var i int64
		r, err := decodeInt(data, &i)
		if err != nil {
			return 0, nil, err
		}
		read += r
	case listToken, dictToken:
		read++ // read start token
		for data[read] != endToken {
			n, _, err := readNextValue(data[read:])
			if err != nil {
				return 0, nil, err
			}
			read += n
		}
		read++ // read end token
	default:
		var s string
		r, err := decodeString(data, &s)
		if err != nil {
			return 0, nil, err
		}
		read += r
	}
	return read, data[:read], nil
}
