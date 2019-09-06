package bencoding

import (
	"errors"
)

const (
	endToken             = 'e'
	intToken             = 'i'
	listToken            = 'l'
	dictToken            = 'd'
	stringSeparatorToken = ':'
	structTagKey         = "bencode"
	optionOmitEmpty      = "omitempty"
)

var (
	ErrNonStringKey   = errors.New("map key is not of type string")
	ErrNonPointer     = errors.New("value is not a pointer")
	ErrInvalidType    = errors.New("type is not valid")
	ErrInvalidValue   = errors.New("value is not valid")
	ErrEmptyData      = errors.New("data is empty")
	ErrCanNotSet      = errors.New("value can not be set")
	ErrLeadingZero    = errors.New("integer must not have leading zeros")
	ErrLengthTooBig   = errors.New("string length would exceed length of data")
	ErrRemainingData  = errors.New("data was not fully consumed")
	ErrInvalidBool    = errors.New("unexpected value when decoding bool")
	ErrInvalidInteger = errors.New("unexpected value when decoding integer")
	ErrInvalidString  = errors.New("unexpected value when decoding string")
	ErrInvalidList    = errors.New("unexpected value when decoding list")
	ErrInvalidDict    = errors.New("unexpected value when decoding dict")
	ErrInvalidStruct  = errors.New("unexpected value when decoding struct")
	ErrInvalidToken   = errors.New("unexpected value when decoding")
)
