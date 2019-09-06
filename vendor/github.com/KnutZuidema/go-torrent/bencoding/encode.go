package bencoding

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Marshaler if a type implements this interface its MarshalBEncode method will be used when marshaling the object into
// a bencode representation
type Marshaler interface {
	MarshalBEncode() ([]byte, error)
}

// Marshal encodes the given value into a bencode representation
//
// returns ErrInvalidType if the value type is not valid for bencoding, namely float64, float32, complex128, complex64
//
// returns ErrNonStringKey if the key type of a map is not string
//
// []byte values are encoded like strings
//
// bool values are encoded as integers 1 if true and 0 if false
func Marshal(value interface{}) ([]byte, error) {
	value = normalize(value)
	switch v := value.(type) {
	case Marshaler:
		return v.MarshalBEncode()
	case string:
		return encodeString(v)
	case int64:
		return encodeInt(v)
	case uint64:
		return encodeUint(v)
	case bool:
		if v {
			return encodeInt(1)
		}
		return encodeInt(0)
	default:
		rv := dereference(reflect.ValueOf(v))
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			var list []interface{}
			for i := 0; i < rv.Len(); i++ {
				list = append(list, rv.Index(i).Interface())
			}
			return encodeList(list)
		case reflect.Map:
			if rv.Type().Key().Kind() != reflect.String {
				return nil, ErrNonStringKey
			}
			dict := make(map[string]interface{})
			iter := rv.MapRange()
			for iter.Next() {
				dict[iter.Key().Interface().(string)] = iter.Value().Interface()
			}
			return encodeDict(dict)
		case reflect.Struct:
			return encodeDict(structToMap(rv))
		default:
			return nil, ErrInvalidType
		}
	}
}

func encodeString(value string) ([]byte, error) {
	return []byte(strconv.Itoa(len(value)) + string(stringSeparatorToken) + value), nil
}

func encodeInt(value int64) ([]byte, error) {
	return []byte(string(intToken) + strconv.FormatInt(value, 10) + string(endToken)), nil
}

func encodeUint(value uint64) ([]byte, error) {
	return []byte(string(intToken) + strconv.FormatUint(value, 10) + string(endToken)), nil
}

func encodeList(value []interface{}) ([]byte, error) {
	res := []byte(string(listToken))
	for _, val := range value {
		d, err := Marshal(val)
		if err != nil {
			return nil, err
		}
		res = append(res, d...)
	}
	res = append(res, endToken)
	return res, nil
}

func encodeDict(value map[string]interface{}) ([]byte, error) {
	res := []byte(string(dictToken))
	type keyValuePair struct {
		key   string
		value interface{}
	}
	var sorted []keyValuePair
	for key, val := range value {
		sorted = append(sorted, keyValuePair{
			key:   key,
			value: val,
		})
	}
	sort.Slice(sorted, func(i, j int) bool {
		if len(sorted[i].key) == len(sorted[j].key) {
			for k := 0; k < len(sorted[i].key); k++ {
				if sorted[i].key[k] != sorted[j].key[k] {
					return sorted[i].key[k] < sorted[j].key[k]
				}
			}
			return true
		}
		return len(sorted[i].key) < len(sorted[j].key)
	})
	for _, pair := range sorted {
		key, _ := encodeString(pair.key)
		val, err := Marshal(pair.value)
		if err != nil {
			return nil, err
		}
		res = append(res, append(key, val...)...)
	}
	res = append(res, endToken)
	return res, nil
}

type tag struct {
	Name      string
	OmitEmpty bool
	Skip      bool
}

func newTag(f reflect.StructField) *tag {
	if f.PkgPath != "" {
		return &tag{
			Skip: true,
		}
	}
	v := f.Tag.Get(structTagKey)
	if v == "" {
		if f.Anonymous {
			return &tag{}
		}
		return &tag{
			Name: f.Name,
		}
	}
	split := strings.Split(v, ",")
	if split[0] == "-" {
		return &tag{
			Skip: true,
		}
	}
	t := &tag{Name: split[0]}
	if split[0] == "" && !f.Anonymous {
		t.Name = f.Name
	}
	for _, option := range split[1:] {
		if option == optionOmitEmpty {
			t.OmitEmpty = true
		}
	}
	return t
}

func structToMap(rv reflect.Value) map[string]interface{} {
	dict := make(map[string]interface{})
	typ := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		t := newTag(typ.Field(i))
		if t.Skip {
			continue
		}
		if t.OmitEmpty && isEmpty(rv) {
			continue
		}
		if typ.Field(i).Anonymous && t.Name == "" {
			for k, v := range structToMap(rv.Field(i)) {
				dict[k] = v
			}
		} else {
			dict[t.Name] = rv.Field(i).Interface()
		}
	}
	return dict
}

func isEmpty(rv reflect.Value) bool {
	return reflect.DeepEqual(rv.Interface(), reflect.Zero(rv.Type()).Interface())
}

func normalize(i interface{}) interface{} {
	switch v := i.(type) {
	case []byte:
		return string(v)
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	}
	return i
}
