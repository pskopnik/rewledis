package rewledis

import (
	"context"
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

var (
	ErrUnknownLedisTypeString   = errors.New("input string does not represent a known LedisType value")
	ErrInvalidLedisType         = errors.New("input LedisType value is unknown or otherwise invalid for this operation")
	ErrNoCorrespondingLedisType = errors.New("there is no LedisType value corresponding to the input RedisType value")
)

type LedisType int8

const (
	LedisTypeNone LedisType = iota
	LedisTypeKV
	LedisTypeList
	LedisTypeHash
	LedisTypeSet
	LedisTypeZSet
)

func (l LedisType) String() string {
	switch l {
	case LedisTypeNone:
		return "None"
	case LedisTypeKV:
		return "KV"
	case LedisTypeList:
		return "List"
	case LedisTypeHash:
		return "Hash"
	case LedisTypeSet:
		return "Set"
	case LedisTypeZSet:
		return "ZSet"
	default:
		return fmt.Sprintf("LedisType(%d)", l)
	}
}

func ParseLedisType(str string) (LedisType, error) {
	switch str {
	case "None":
		return LedisTypeNone, nil
	case "KV":
		return LedisTypeKV, nil
	case "List":
		return LedisTypeList, nil
	case "Hash":
		return LedisTypeHash, nil
	case "Set":
		return LedisTypeSet, nil
	case "ZSet":
		return LedisTypeZSet, nil
	default:
		return LedisTypeNone, ErrUnknownLedisTypeString
	}

}
func ParseLedisTypeFromLedis(str string) (LedisType, error) {
	switch str {
	case "KV":
		return LedisTypeKV, nil
	case "LIST":
		return LedisTypeList, nil
	case "HASH":
		return LedisTypeHash, nil
	case "SET":
		return LedisTypeSet, nil
	case "ZSET":
		return LedisTypeZSet, nil
	default:
		return LedisTypeNone, ErrUnknownLedisTypeString
	}
}

func LedisTypeFromRedisType(redisType RedisType) (LedisType, error) {
	switch redisType {
	case RedisTypeNone:
		return LedisTypeNone, nil
	case RedisTypeString:
		return LedisTypeKV, nil
	case RedisTypeList:
		return LedisTypeList, nil
	case RedisTypeHash:
		return LedisTypeHash, nil
	case RedisTypeSet:
		return LedisTypeSet, nil
	case RedisTypeZSet:
		return LedisTypeZSet, nil
	case RedisTypeGeneric:
		return LedisTypeNone, ErrNoCorrespondingLedisType
	default:
		return LedisTypeNone, ErrInvalidRedisType
	}
}

var (
	ErrUnknownRedisTypeString = errors.New("input string does not represent a known RedisType value")
	ErrInvalidRedisType       = errors.New("input RedisType value is unknown or otherwise invalid for this operation")
)

type RedisType int8

const (
	RedisTypeNone RedisType = iota
	RedisTypeString
	RedisTypeList
	RedisTypeHash
	RedisTypeSet
	RedisTypeZSet

	RedisTypeGeneric = -2
)

func (r RedisType) String() string {
	switch r {
	case RedisTypeNone:
		return "None"
	case RedisTypeString:
		return "String"
	case RedisTypeList:
		return "List"
	case RedisTypeHash:
		return "Hash"
	case RedisTypeSet:
		return "Set"
	case RedisTypeZSet:
		return "ZSet"
	case RedisTypeGeneric:
		return "Generic"
	default:
		return fmt.Sprintf("RedisType(%d)", r)
	}
}

func ParseRedisType(str string) (RedisType, error) {
	switch str {
	case "None":
		return RedisTypeNone, nil
	case "String":
		return RedisTypeString, nil
	case "List":
		return RedisTypeList, nil
	case "Hash":
		return RedisTypeHash, nil
	case "Set":
		return RedisTypeSet, nil
	case "ZSet":
		return RedisTypeZSet, nil
	case "Generic":
		return RedisTypeGeneric, nil
	default:
		return 0, ErrUnknownRedisTypeString
	}

}
func ParseRedisTypeFromRedis(str string) (RedisType, error) {
	switch str {
	case "string":
		return RedisTypeString, nil
	case "list":
		return RedisTypeList, nil
	case "hash":
		return RedisTypeHash, nil
	case "set":
		return RedisTypeSet, nil
	case "zset":
		return RedisTypeZSet, nil
	default:
		return 0, ErrUnknownRedisTypeString
	}
}

type KeyExtractor interface {
	Keys(args []string) []string
}

type KeyExtractorFunc func(args []string) []string

func (k KeyExtractorFunc) Keys(args []string) []string {
	return k(args)
}

// KeysAtIndices returns a KeyExtractor.
// The KeyExtractor returns the arguments with the indices passed to
// KeysAtIndices.
func KeysAtIndices(indices ...int) KeyExtractor {
	return KeyExtractorFunc(func(args []string) []string {
		keys := make([]string, len(indices))

		for ind, argIndex := range indices {
			keys[ind] = args[argIndex]
		}

		return keys
	})
}

// KeysFromIndex returns a KeyExtractor.
// The KeyExtractor returns all arguments starting at the index passed as a
// parameter. If the optional parameter skip is passed, skip arguments are
// skipped after each key argument.
func KeysFromIndex(index int, skip ...int) KeyExtractor {
	return KeysFromUntilIndex(index, 0, skip...)
}

// KeysFromIndexUntil returns a KeyExtractor.
// The KeyExtractor returns all arguments between the two indices passed as
// parameters, including the argument at from and excluding the argument at
// until. It works in the same way as slice ranges: args[from:until].
//
// If until is 0 the maximum possible value is presumed. If until is less than
// 0 it is subtracted from the length of the arguments.
//
// If the optional parameter skip is passed, skip arguments are skipped after
// each key argument.
func KeysFromUntilIndex(from, until int, skip ...int) KeyExtractor {
	if len(skip) > 1 {
		panic("optional skip parameter must contain at most one value")
	}

	var skipVal int
	if len(skip) == 1 {
		skipVal = skip[0]
	}

	return KeyExtractorFunc(func(args []string) []string {
		untilIndex := until
		if untilIndex <= 0 {
			untilIndex = len(args) + untilIndex
		} else if untilIndex >= len(args) {
			untilIndex = len(args) - 1
		}

		keys := make([]string, 0, (untilIndex-from+skipVal)/(1+skipVal))

		for i := from; i < untilIndex; i += 1 + skipVal {
			keys = append(keys, args[i])
		}

		return keys
	})
}

type Slot struct {
	RepliesCount int
	ProcessFunc  func([]interface{}) (interface{}, error)
}

type SendLedisFunc func(ledisConn redis.Conn) (Slot, error)

// type Performable []SerializedCommand

// type SerializedCommand struct {
// 	command string
// 	args []interface{}
// }

// type TransformFunc func(command *RedisCommand, rewriter *Rewriter, args []interface{}) (Performable, error)

// type TransformFunc func(command *RedisCommand, rewriter *Rewriter, args []interface{}) (Slot, error)

type TransformFunc func(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error)

func NoneTransformer() TransformFunc {
	return TransformFunc(func(_ *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
		return SendLedisFunc(func(ledisConn redis.Conn) (Slot, error) {
			err := ledisConn.Send(command.Name, args...)
			if err != nil {
				return Slot{}, err
			}

			return Slot{
				RepliesCount: 1,
				ProcessFunc: func(replies []interface{}) (interface{}, error) {
					return replies[0], nil
				},
			}, nil
		}), nil
	})
}

type KeyTypeAggregation struct {
	None []string
	KV   []string
	List []string
	Hash []string
	Set  []string
	ZSet []string
}

func (k *KeyTypeAggregation) AppendKeys(typesInfo []TypeInfo) {
	for i := range typesInfo {
		switch typesInfo[i].Type {
		case LedisTypeNone:
			k.None = append(k.None, typesInfo[i].Key)
		case LedisTypeKV:
			k.KV = append(k.KV, typesInfo[i].Key)
		case LedisTypeList:
			k.List = append(k.List, typesInfo[i].Key)
		case LedisTypeHash:
			k.Hash = append(k.Hash, typesInfo[i].Key)
		case LedisTypeSet:
			k.Set = append(k.Set, typesInfo[i].Key)
		case LedisTypeZSet:
			k.ZSet = append(k.ZSet, typesInfo[i].Key)
		}
	}
}

type TypeSpecificCommands struct {
	None string
	KV   string
	List string
	Hash string
	Set  string
	ZSet string
}

var (
	ErrInvalidAggregatorValue = errors.New("invalid Aggregator value")
)

type Aggregator int8

const (
	AggregatorSum Aggregator = iota
	AggregatorCountOne
)

type TypeSpecificBulkTransformerConfig struct {
	Commands   TypeSpecificCommands
	Debulk     bool
	Aggregator Aggregator
}

func TypeSpecificBulkTransformer(config *TypeSpecificBulkTransformerConfig) TransformFunc {
	return TransformFunc(func(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
		var typeInfoArray [12]TypeInfo
		var stringArgsArray [12]string

		stringArgs := keyArgumentsToStringsAppend(args, stringArgsArray[:0])
		keys := command.Keys(stringArgs)

		resolver := rewriter.Resolver()
		ctx := context.Background()
		typesInfo, err := resolver.ResolveAppend(ctx, keys, typeInfoArray[:0])
		if err != nil {
			return nil, err
		}

		keyTypeAggregation := KeyTypeAggregation{}
		keyTypeAggregation.AppendKeys(typesInfo)

		return SendLedisFunc(func(ledisConn redis.Conn) (Slot, error) {
			var sentCount, repliesCount int
			var err error

			sentCount, err = somethingSomething(config.Commands.None, keyTypeAggregation.None, ledisConn, config.Debulk)
			if err != nil {
				return Slot{}, err
			}
			repliesCount += sentCount
			sentCount, err = somethingSomething(config.Commands.KV, keyTypeAggregation.KV, ledisConn, config.Debulk)
			if err != nil {
				return Slot{}, err
			}
			repliesCount += sentCount
			sentCount, err = somethingSomething(config.Commands.List, keyTypeAggregation.List, ledisConn, config.Debulk)
			if err != nil {
				return Slot{}, err
			}
			repliesCount += sentCount
			sentCount, err = somethingSomething(config.Commands.Hash, keyTypeAggregation.Hash, ledisConn, config.Debulk)
			if err != nil {
				return Slot{}, err
			}
			repliesCount += sentCount
			sentCount, err = somethingSomething(config.Commands.Set, keyTypeAggregation.Set, ledisConn, config.Debulk)
			if err != nil {
				return Slot{}, err
			}
			repliesCount += sentCount
			sentCount, err = somethingSomething(config.Commands.ZSet, keyTypeAggregation.ZSet, ledisConn, config.Debulk)
			if err != nil {
				return Slot{}, err
			}
			repliesCount += sentCount

			return Slot{
				RepliesCount: repliesCount,
				ProcessFunc: func(replies []interface{}) (interface{}, error) {
					switch config.Aggregator {
					case AggregatorSum:
						var sum int
						for _, reply := range replies {
							value, err := redis.Int(reply, nil)
							if err != nil {
								return nil, err
							}
							sum += value
						}
						return sum, nil
					case AggregatorCountOne:
						return len(replies), nil
					default:
						return nil, ErrInvalidAggregatorValue
					}
				},
			}, nil
		}), nil
	})
}

func somethingSomething(command string, keys []string, conn redis.Conn, debulk bool) (int, error) {
	var sentCount int

	if len(command) > 0 {
		if debulk {
			for _, key := range keys {
				sentCount++
				err := conn.Send(command, key)
				if err != nil {
					return sentCount, err
				}
			}
		} else {
			sentCount++

			var argsArray [12]interface{}
			args := append(argsArray[:0], make([]interface{}, len(keys))...)
			for i := range keys {
				args[i] = keys[i]
			}

			err := conn.Send(command, args...)
			if err != nil {
				return sentCount, err
			}
		}
	}

	return sentCount, nil
}

var (
	ErrUnknownRedisCommandName = errors.New("input string does not represent a known RedisCommand name")
)

var _ KeyExtractor = &RedisCommand{}

type RedisCommand struct {
	Name          string
	KeyType       RedisType
	KeyExtractor  KeyExtractor
	TransformFunc TransformFunc
	Syntax        string
}

func (r RedisCommand) Keys(args []string) []string {
	return r.KeyExtractor.Keys(args)
}

func keyArgumentsToStrings(args []interface{}) []string {
	return keyArgumentsToStringsAppend(args, nil)
}

func keyArgumentsToStringsAppend(args []interface{}, stringArgs []string) []string {
	baseIndex := len(stringArgs)
	stringArgs = append(stringArgs, make([]string, len(args))...)

	for i, arg := range args {
		switch typedArg := arg.(type) {
		case string:
			stringArgs[baseIndex+i] = typedArg
		case []byte:
			stringArgs[baseIndex+i] = string(typedArg)
		case redis.Argument:
			nestedArg := typedArg.RedisArg()
			switch typedArg := nestedArg.(type) {
			case string:
				stringArgs[baseIndex+i] = typedArg
			case []byte:
				stringArgs[baseIndex+i] = string(typedArg)
			}
		}
	}

	return stringArgs
}
