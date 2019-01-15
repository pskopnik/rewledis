// Package args provides utilities for working with redigo command arguments.
package args

import (
	"bytes"
	"errors"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

// AsSimpleStrings converts the elements of args to their string form.
// See AsSimpleString for information how conversion is performed.
func AsSimpleStrings(args []interface{}) []string {
	return AppendAsSimpleStrings(nil, args)
}

// AppendAsSimpleStrings appends the string form of args to the stringArgs
// slice.
// See AsSimpleString for information how conversion is performed.
func AppendAsSimpleStrings(stringArgs []string, args []interface{}) []string {
	baseIndex := len(stringArgs)
	stringArgs = append(stringArgs, make([]string, len(args))...)

	for i, arg := range args {
		stringArgs[baseIndex+i] = AsSimpleString(arg)
	}

	return stringArgs
}

// AsSimpleString converts an argument passed to redigo's Conn.Send()
// method to a string. Only arguments which are sent as redis strings are
// converted. For all other argument types, this function returns the empty
// string ("").
//
// AsSimpleString does not perform recursion and only supports one level of
// the redis.Argument interface. This allows inlining.
func AsSimpleString(arg interface{}) string {
	switch typedArg := arg.(type) {
	case string:
		return typedArg
	case []byte:
		return string(typedArg)
	case redis.Argument:
		nestedArg := typedArg.RedisArg()
		switch typedArg := nestedArg.(type) {
		case string:
			return typedArg
		case []byte:
			return string(typedArg)
		default:
			return ""
		}
	default:
		return ""
	}
}

// Error variables related to Info and associated functions.
var (
	ErrInvalidTypeForOperation = errors.New("invalid type for the operation")
)

// Parse extracts information about arg and returns an Info structure
// containing this information.
func Parse(arg interface{}) Info {
	var info Info

	info.Arg = arg
	info.UnwrappedArg = arg

	parseRecursive(&info)

	return info
}

func parseRecursive(info *Info) {
	switch arg := info.UnwrappedArg.(type) {
	case string:
		info.Type = TypeString
	case []byte:
		info.Type = TypeBytes
	case int64:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case int32:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case int16:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case int8:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case int:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case uint64:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case uint32:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case uint16:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case uint8:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case uint:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case uintptr:
		info.Type = TypeInt
		info.intValue = int64(arg)
	case float64:
		info.Type = TypeFloat
		info.floatValue = float64(arg)
	case float32:
		info.Type = TypeFloat
		info.floatValue = float64(arg)
	case complex64:
		info.Type = TypeComplex
	case complex128:
		info.Type = TypeComplex
	case bool:
		info.Type = TypeBool
		if arg {
			info.intValue = 1
		} else {
			info.intValue = 0
		}
	case nil:
		info.Type = TypeNil
	case redis.Argument:
		info.WrappingLevel++
		info.UnwrappedArg = arg

		parseRecursive(info)
	}

	// The default case implemented by redigo is omitted.
	// All types which are meant to be handled by the default case (builtin
	// numeric types) **should** have been handled by one of the cases above.
}

type Type int8

const (
	TypeUnset Type = iota
	TypeString
	TypeBytes
	TypeInt
	TypeUint
	TypeFloat
	TypeBool
	TypeComplex
	TypeNil
)

type Info struct {
	Arg          interface{}
	UnwrappedArg interface{}
	// intValue stores the int, uint or bool value of the argument described.
	intValue int64
	// float64Value stores the float value of the argument described.
	floatValue    float64
	WrappingLevel int
	Type          Type
}

func (i *Info) Complex128Value() complex128 {
	if val, ok := i.UnwrappedArg.(complex128); ok {
		return val
	}
	return complex128(i.UnwrappedArg.(complex64))
}

func (i *Info) BytesValue() []byte {
	return i.UnwrappedArg.([]byte)
}

func (i *Info) StringValue() string {
	return i.UnwrappedArg.(string)
}

func (i *Info) Int64Value() int64 {
	return i.intValue
}

func (i *Info) Uint64Value() uint64 {
	return uint64(i.intValue)
}

func (i *Info) Float64Value() float64 {
	return i.floatValue
}

func (i *Info) BoolValue() bool {
	return i.intValue != 0
}

// IsWrapped returns true iff i describes an argument wrapped in
// redis.Argument at least once.
func (i *Info) IsWrapped() bool {
	return i.WrappingLevel > 0
}

// IsStringLike returns true iff i describes either a String or Bytes argument.
func (i *Info) IsStringLike() bool {
	return i.Type == TypeString || i.Type == TypeBytes
}

// EqualEither checks string equality of i to either tString OR tBytes using
// whichever matches the type of i.
func (i *Info) EqualEither(tString string, tBytes []byte) bool {
	switch i.Type {
	case TypeString:
		return i.StringValue() == tString
	case TypeBytes:
		return bytes.Equal(i.BytesValue(), tBytes)
	default:
		return false
	}
}

// EqualEither checks string equality under unicode case folding of i to
// either tString OR tBytes using whichever matches the type of i.
func (i *Info) EqualFoldEither(tString string, tBytes []byte) bool {
	switch i.Type {
	case TypeString:
		return strings.EqualFold(i.StringValue(), tString)
	case TypeBytes:
		return bytes.EqualFold(i.BytesValue(), tBytes)
	default:
		return false
	}
}

// ConvertToRedisString returns the argument described by i in the form of a
// string. Depending on the type of i, a conversion might be performed. All
// conversions match the conversions performed by redigo when writing
// arguments to the connection.
func (i *Info) ConvertToRedisString() (string, error) {
	switch i.Type {
	case TypeString:
		return i.StringValue(), nil
	case TypeBytes:
		return string(i.BytesValue()), nil
	case TypeInt:
		return strconv.FormatInt(i.Int64Value(), 10), nil
	case TypeUint:
		return strconv.FormatUint(i.Uint64Value(), 10), nil
	case TypeFloat:
		return strconv.FormatFloat(i.Float64Value(), 'f', -1, 64), nil
	case TypeBool:
		if i.BoolValue() {
			return "1", nil
		} else {
			return "0", nil
		}
	case TypeNil:
		return "", nil
	default:
		return "", ErrInvalidTypeForOperation
	}
}

// ConvertToRedisBytesString returns the argument described by i in the form
// of a byte slice. If i is a byte slice, a reference is returned instead of a
// copy. The returned slice must not be modified to ensure the consistency of
// i. Depending on the type of i, a conversion might be performed. All
// conversions match the conversions performed by redigo when writing
// arguments to the connection.
func (i *Info) ConvertToRedisBytesString() ([]byte, error) {
	switch i.Type {
	case TypeString:
		return []byte(i.StringValue()), nil
	case TypeBytes:
		return i.BytesValue(), nil
	case TypeInt:
		return strconv.AppendInt(nil, i.Int64Value(), 10), nil
	case TypeUint:
		return strconv.AppendUint(nil, i.Uint64Value(), 10), nil
	case TypeFloat:
		return strconv.AppendFloat(nil, i.Float64Value(), 'f', -1, 64), nil
	case TypeBool:
		if i.BoolValue() {
			return []byte("1"), nil
		} else {
			return []byte("0"), nil
		}
	case TypeNil:
		return []byte(""), nil
	default:
		return nil, ErrInvalidTypeForOperation
	}
}

// AppendRedisBytesString appends the argument described by i in the form of a
// byte slice. Depending on the type of i, a conversion might be performed.
// All conversions match the conversions performed by redigo when writing
// arguments to the connection.
func (i *Info) AppendRedisBytesString(buf []byte) ([]byte, error) {
	switch i.Type {
	case TypeString:
		return append(buf, i.StringValue()...), nil
	case TypeBytes:
		return append(buf, i.BytesValue()...), nil
	case TypeInt:
		return strconv.AppendInt(buf, i.Int64Value(), 10), nil
	case TypeUint:
		return strconv.AppendUint(buf, i.Uint64Value(), 10), nil
	case TypeFloat:
		return strconv.AppendFloat(buf, i.Float64Value(), 'f', -1, 64), nil
	case TypeBool:
		if i.BoolValue() {
			return append(buf, '1'), nil
		} else {
			return append(buf, '0'), nil
		}
	case TypeNil:
		return buf, nil
	default:
		return buf, ErrInvalidTypeForOperation
	}
}

// ConvertToInt returns the argument described by i in the form of an int64.
// Depending on the type of i, a conversion might be performed. Any error
// occuring during conversion is returned.
//
// TODO: Can Nil be converted to an int?
func (i *Info) ConvertToInt() (int64, error) {
	switch i.Type {
	case TypeString:
		return strconv.ParseInt(i.StringValue(), 10, 64)
	case TypeBytes:
		return strconv.ParseInt(string(i.BytesValue()), 10, 64)
	case TypeInt:
		return i.Int64Value(), nil
	case TypeUint:
		return int64(i.Uint64Value()), nil
	case TypeFloat:
		return int64(i.Float64Value()), nil
	case TypeBool:
		if i.BoolValue() {
			return 1, nil
		} else {
			return 0, nil
		}
	case TypeNil:
		return 0, ErrInvalidTypeForOperation
	default:
		return 0, ErrInvalidTypeForOperation
	}
}

// ConvertToUint returns the argument described by i in the form of an uint64.
// Depending on the type of i, a conversion might be performed. Any error
// occuring during conversion is returned.
//
// TODO: Can Nil be converted to a uint?
func (i *Info) ConvertToUint() (uint64, error) {
	switch i.Type {
	case TypeString:
		return strconv.ParseUint(i.StringValue(), 10, 64)
	case TypeBytes:
		return strconv.ParseUint(string(i.BytesValue()), 10, 64)
	case TypeInt:
		return uint64(i.Int64Value()), nil
	case TypeUint:
		return i.Uint64Value(), nil
	case TypeFloat:
		return uint64(i.Float64Value()), nil
	case TypeBool:
		if i.BoolValue() {
			return 1, nil
		} else {
			return 0, nil
		}
	case TypeNil:
		return 0, ErrInvalidTypeForOperation
	default:
		return 0, ErrInvalidTypeForOperation
	}
}
