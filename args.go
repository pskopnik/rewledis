package rewledis

import (
	"bytes"
	"errors"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

// argsAsSimleStrings converts the args to their string form.
// See argAsSimpleString for information how conversion is performed.
func argsAsSimpleStrings(args []interface{}) []string {
	return appendArgsAsSimpleStrings(nil, args)
}

// appendArgsAsSimpleStrings appends the string form of args to the stringArgs
// slice.
// See argAsSimpleString for information how conversion is performed.
func appendArgsAsSimpleStrings(stringArgs []string, args []interface{}) []string {
	baseIndex := len(stringArgs)
	stringArgs = append(stringArgs, make([]string, len(args))...)

	for i, arg := range args {
		stringArgs[baseIndex+i] = argAsSimpleString(arg)
	}

	return stringArgs
}

// argAsSimpleString converts an argument passed to redigo's Conn.Send()
// method to a string. Only arguments which are sent as redis strings are
// converted. For all other argument types, this function returns the empty
// string ("").
//
// argAsSimpleString does not perform recursion and only supports one level of
// the redis.Argument interface. This allows inlining.
func argAsSimpleString(arg interface{}) string {
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

// Error variables related to argInfo and associated functions.
var (
	ErrInvalidTypeForOperation = errors.New("invalid type for the operation")
)

func parseArg(arg interface{}) argInfo {
	var info argInfo

	info.Arg = arg
	info.UnwrappedArg = arg

	parseArgRecursive(&info)

	return info
}

func parseArgRecursive(info *argInfo) {
	switch arg := info.UnwrappedArg.(type) {
	case string:
		info.Type = argTypeString
	case []byte:
		info.Type = argTypeBytes
	case int64:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case int32:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case int16:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case int8:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case int:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case uint64:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case uint32:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case uint16:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case uint8:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case uint:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case uintptr:
		info.Type = argTypeInt
		info.intValue = int64(arg)
	case float64:
		info.Type = argTypeFloat
		info.floatValue = float64(arg)
	case float32:
		info.Type = argTypeFloat
		info.floatValue = float64(arg)
	case complex64:
		info.Type = argTypeComplex
	case complex128:
		info.Type = argTypeComplex
	case bool:
		info.Type = argTypeBool
		if arg {
			info.intValue = 1
		} else {
			info.intValue = 0
		}
	case nil:
		info.Type = argTypeNil
	case redis.Argument:
		info.WrappingLevel++
		info.UnwrappedArg = arg

		parseArgRecursive(info)
	}

	// The default case implemented by redigo is omitted.
	// All types which are meant to be handled by the default case (builtin
	// numeric types) **should** have been handled by one of the cases above.
}

type argType int8

const (
	argTypeUnset argType = iota
	argTypeString
	argTypeBytes
	argTypeInt
	argTypeUint
	argTypeFloat
	argTypeBool
	argTypeComplex
	argTypeNil
)

type argInfo struct {
	Arg          interface{}
	UnwrappedArg interface{}
	// intValue stores the int, uint or bool value of the argument described.
	intValue int64
	// float64Value stores the float value of the argument described.
	floatValue    float64
	WrappingLevel int
	Type          argType
}

func (a *argInfo) Complex128Value() complex128 {
	if val, ok := a.UnwrappedArg.(complex128); ok {
		return val
	}
	return complex128(a.UnwrappedArg.(complex64))
}

func (a *argInfo) BytesValue() []byte {
	return a.UnwrappedArg.([]byte)
}

func (a *argInfo) StringValue() string {
	return a.UnwrappedArg.(string)
}

func (a *argInfo) Int64Value() int64 {
	return a.intValue
}

func (a *argInfo) Uint64Value() uint64 {
	return uint64(a.intValue)
}

func (a *argInfo) Float64Value() float64 {
	return a.floatValue
}

func (a *argInfo) BoolValue() bool {
	return a.intValue != 0
}

// IsWrapped returns true iff a describes an argument wrapped in
// redis.Argument at least once.
func (a *argInfo) IsWrapped() bool {
	return a.WrappingLevel > 0
}

// IsStringLike returns true iff a describes either a String or Bytes argument.
func (a *argInfo) IsStringLike() bool {
	return a.Type == argTypeString || a.Type == argTypeBytes
}

// EqualEither checks string equality of a to either tString OR tBytes using
// whichever matches the type of a.
func (a *argInfo) EqualEither(tString string, tBytes []byte) bool {
	switch a.Type {
	case argTypeString:
		return a.StringValue() == tString
	case argTypeBytes:
		return bytes.Equal(a.BytesValue(), tBytes)
	default:
		return false
	}
}

// EqualEither checks string equality under unicode case folding of a to
// either tString OR tBytes using whichever matches the type of a.
func (a *argInfo) EqualFoldEither(tString string, tBytes []byte) bool {
	switch a.Type {
	case argTypeString:
		return strings.EqualFold(a.StringValue(), tString)
	case argTypeBytes:
		return bytes.EqualFold(a.BytesValue(), tBytes)
	default:
		return false
	}
}

// ConvertToRedisString returns the argument described by a expressed as a
// string. Depending on the type of a, a conversion might be performed. All
// conversions match the conversions performed by redigo when writing
// arguments to the connection.
func (a *argInfo) ConvertToRedisString() (string, error) {
	switch a.Type {
	case argTypeString:
		return a.StringValue(), nil
	case argTypeBytes:
		return string(a.BytesValue()), nil
	case argTypeInt:
		return strconv.FormatInt(a.Int64Value(), 10), nil
	case argTypeUint:
		return strconv.FormatUint(a.Uint64Value(), 10), nil
	case argTypeFloat:
		return strconv.FormatFloat(a.Float64Value(), 'f', -1, 64), nil
	case argTypeBool:
		if a.BoolValue() {
			return "1", nil
		} else {
			return "0", nil
		}
	case argTypeNil:
		return "", nil
	default:
		return "", ErrInvalidTypeForOperation
	}
}

// ConvertToRedisBytesString returns the argument described by a expressed as
// a byte slice. If a is a byte slice, a reference is returned instead of a
// copy. The returned slice must not be modified to ensure the consistency of
// a. Depending on the type of a, a conversion might be performed. All
// conversions match the conversions performed by redigo when writing
// arguments to the connection.
func (a *argInfo) ConvertToRedisBytesString() ([]byte, error) {
	switch a.Type {
	case argTypeString:
		return []byte(a.StringValue()), nil
	case argTypeBytes:
		return a.BytesValue(), nil
	case argTypeInt:
		return strconv.AppendInt(nil, a.Int64Value(), 10), nil
	case argTypeUint:
		return strconv.AppendUint(nil, a.Uint64Value(), 10), nil
	case argTypeFloat:
		return strconv.AppendFloat(nil, a.Float64Value(), 'f', -1, 64), nil
	case argTypeBool:
		if a.BoolValue() {
			return []byte("1"), nil
		} else {
			return []byte("0"), nil
		}
	case argTypeNil:
		return []byte(""), nil
	default:
		return nil, ErrInvalidTypeForOperation
	}
}

// AppendRedisBytesString appends the argument described by a expressed to a
// byte slice. Depending on the type of a, a conversion might be performed.
// All conversions match the conversions performed by redigo when writing
// arguments to the connection.
func (a *argInfo) AppendRedisBytesString(buf []byte) ([]byte, error) {
	switch a.Type {
	case argTypeString:
		return append(buf, a.StringValue()...), nil
	case argTypeBytes:
		return append(buf, a.BytesValue()...), nil
	case argTypeInt:
		return strconv.AppendInt(buf, a.Int64Value(), 10), nil
	case argTypeUint:
		return strconv.AppendUint(buf, a.Uint64Value(), 10), nil
	case argTypeFloat:
		return strconv.AppendFloat(buf, a.Float64Value(), 'f', -1, 64), nil
	case argTypeBool:
		if a.BoolValue() {
			return append(buf, '1'), nil
		} else {
			return append(buf, '0'), nil
		}
	case argTypeNil:
		return buf, nil
	default:
		return buf, ErrInvalidTypeForOperation
	}
}

// ConvertToInt returns the argument described by a expressed as an int64.
// Depending on the type of a, a conversion might be performed. Any error
// occuring during conversion is returned.
//
// TODO: Can Nil be converted to an int?
func (a *argInfo) ConvertToInt() (int64, error) {
	switch a.Type {
	case argTypeString:
		return strconv.ParseInt(a.StringValue(), 10, 64)
	case argTypeBytes:
		return strconv.ParseInt(string(a.BytesValue()), 10, 64)
	case argTypeInt:
		return a.Int64Value(), nil
	case argTypeUint:
		return int64(a.Uint64Value()), nil
	case argTypeFloat:
		return int64(a.Float64Value()), nil
	case argTypeBool:
		if a.BoolValue() {
			return 1, nil
		} else {
			return 0, nil
		}
	case argTypeNil:
		return 0, ErrInvalidTypeForOperation
	default:
		return 0, ErrInvalidTypeForOperation
	}
}

// ConvertToUint returns the argument described by a expressed as an uint64.
// Depending on the type of a, a conversion might be performed. Any error
// occuring during conversion is returned.
//
// TODO: Can Nil be converted to a uint?
func (a *argInfo) ConvertToUint() (uint64, error) {
	switch a.Type {
	case argTypeString:
		return strconv.ParseUint(a.StringValue(), 10, 64)
	case argTypeBytes:
		return strconv.ParseUint(string(a.BytesValue()), 10, 64)
	case argTypeInt:
		return uint64(a.Int64Value()), nil
	case argTypeUint:
		return a.Uint64Value(), nil
	case argTypeFloat:
		return uint64(a.Float64Value()), nil
	case argTypeBool:
		if a.BoolValue() {
			return 1, nil
		} else {
			return 0, nil
		}
	case argTypeNil:
		return 0, ErrInvalidTypeForOperation
	default:
		return 0, ErrInvalidTypeForOperation
	}
}
