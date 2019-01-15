package rewledis

import (
	"context"
	"errors"

	"github.com/gomodule/redigo/redis"
)

type ArgsExtractor interface {
	AppendArgs(args []interface{}, extracted []interface{}) []interface{}
	Args(args []interface{}) []interface{}
}

var _ ArgsExtractor = ArgsExtractorFunc(nil)

type ArgsExtractorFunc func(args []interface{}, extracted []interface{}) []interface{}

func (a ArgsExtractorFunc) AppendArgs(args []interface{}, extracted []interface{}) []interface{} {
	return a(args, extracted)
}

func (a ArgsExtractorFunc) Args(args []interface{}) []interface{} {
	return a(args, nil)
}

// ArgsAtIndices returns an ArgsExtractor.
// The ArgsExtractor returns the arguments with the indices passed to
// ArgsAtIndices.
func ArgsAtIndices(indices ...int) ArgsExtractor {
	return ArgsExtractorFunc(func(args []interface{}, extracted []interface{}) []interface{} {
		baseInd := len(extracted)
		extracted = append(extracted, make([]interface{}, len(indices))...)

		for ind, argIndex := range indices {
			extracted[baseInd+ind] = args[argIndex]
		}

		return extracted
	})
}

// ArgsFromIndex returns an ArgsExtractor.
// The ArgsExtractor returns all arguments starting at the index passed as a
// parameter. If the optional parameter skip is passed, skip arguments are
// skipped after each key argument.
func ArgsFromIndex(index int, skip ...int) ArgsExtractor {
	return ArgsFromUntilIndex(index, 0, skip...)
}

// ArgsFromIndexUntil returns an ArgsExtractor.
// The ArgsExtractor returns all arguments between the two indices passed as
// parameters, including the argument at from and excluding the argument at
// until. It works in the same way as slice ranges: args[from:until].
//
// If until is 0 the maximum possible value is presumed. If until is less than
// 0 it is subtracted from the length of the arguments.
//
// If the optional parameter skip is passed, skip arguments are skipped after
// each key argument.
func ArgsFromUntilIndex(from, until int, skip ...int) ArgsExtractor {
	if len(skip) > 1 {
		panic("optional skip parameter must contain at most one value")
	}

	var skipVal int
	if len(skip) == 1 {
		skipVal = skip[0]
	}

	return ArgsExtractorFunc(func(args []interface{}, extracted []interface{}) []interface{} {
		untilIndex := until
		if untilIndex <= 0 {
			untilIndex = len(args) + untilIndex
		} else if untilIndex >= len(args) {
			untilIndex = len(args) - 1
		}

		ind := len(extracted)
		extracted = append(extracted, make([]interface{}, (untilIndex-from+skipVal)/(1+skipVal))...)

		for i := from; i < untilIndex; i += 1 + skipVal {
			extracted[ind] = args[i]
			ind++
		}

		return extracted
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
	ErrInvalidAggregationValue = errors.New("invalid Aggregation value")
)

type Aggregation int8

const (
	AggregationSum Aggregation = iota
	AggregationCountOne
	AggregationFirst
)

type TypeSpecificBulkTransformerConfig struct {
	Commands            TypeSpecificCommands
	Debulk              bool
	Aggregation         Aggregation
	AppendArgsExtractor ArgsExtractor
}

func TypeSpecificBulkTransformer(config *TypeSpecificBulkTransformerConfig) TransformFunc {
	return TransformFunc(func(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
		var typeInfoArray [12]TypeInfo
		var extractedArgsArray [12]interface{}
		var keysArray [12]string

		keyArgs := command.KeyExtractor.AppendArgs(args, extractedArgsArray[:0])
		keys := appendArgsAsStrings(keyArgs, keysArray[:0])

		resolver := rewriter.Resolver()
		ctx, cancel := context.WithCancel(context.Background())
		typesInfo, err := resolver.ResolveAppend(ctx, keys, typeInfoArray[:0])
		cancel()
		if err != nil {
			return nil, err
		}

		var appendArgs []interface{}
		if config.AppendArgsExtractor != nil {
			appendArgs = config.AppendArgsExtractor.AppendArgs(args, extractedArgsArray[:0])
		}

		keyTypeAggregation := KeyTypeAggregation{}
		keyTypeAggregation.AppendKeys(typesInfo)

		return SendLedisFunc(func(ledisConn redis.Conn) (Slot, error) {
			repliesCount, err := sendBulkForAllTypes(config, keyTypeAggregation, ledisConn, appendArgs)
			if err != nil {
				return Slot{}, err
			}

			return Slot{
				RepliesCount: repliesCount,
				ProcessFunc:  Aggregator(config.Aggregation),
			}, nil
		}), nil
	})
}

func Aggregator(aggregation Aggregation) func([]interface{}) (interface{}, error) {
	return func(replies []interface{}) (interface{}, error) {
		switch aggregation {
		case AggregationSum:
			var sum int64
			for _, reply := range replies {
				value, err := redis.Int64(reply, nil)
				if err != nil {
					return nil, err
				}
				sum += value
			}
			return sum, nil
		case AggregationCountOne:
			return int64(len(replies)), nil
		case AggregationFirst:
			if len(replies) == 0 {
				return nil, nil
			}
			return replies[0], nil
		default:
			return nil, ErrInvalidAggregationValue
		}
	}
}

func sendBulkForAllTypes(
	config *TypeSpecificBulkTransformerConfig,
	keyTypeAggregation KeyTypeAggregation,
	ledisConn redis.Conn,
	appendArgs []interface{},
) (int, error) {
	var sentCount, repliesCount int
	var err error

	sentCount, err = sendBulk(config.Commands.None, keyTypeAggregation.None, ledisConn, config.Debulk, appendArgs)
	if err != nil {
		return repliesCount, err
	}
	repliesCount += sentCount
	sentCount, err = sendBulk(config.Commands.KV, keyTypeAggregation.KV, ledisConn, config.Debulk, appendArgs)
	if err != nil {
		return repliesCount, err
	}
	repliesCount += sentCount
	sentCount, err = sendBulk(config.Commands.List, keyTypeAggregation.List, ledisConn, config.Debulk, appendArgs)
	if err != nil {
		return repliesCount, err
	}
	repliesCount += sentCount
	sentCount, err = sendBulk(config.Commands.Hash, keyTypeAggregation.Hash, ledisConn, config.Debulk, appendArgs)
	if err != nil {
		return repliesCount, err
	}
	repliesCount += sentCount
	sentCount, err = sendBulk(config.Commands.Set, keyTypeAggregation.Set, ledisConn, config.Debulk, appendArgs)
	if err != nil {
		return repliesCount, err
	}
	repliesCount += sentCount
	sentCount, err = sendBulk(config.Commands.ZSet, keyTypeAggregation.ZSet, ledisConn, config.Debulk, appendArgs)
	if err != nil {
		return repliesCount, err
	}
	repliesCount += sentCount

	return repliesCount, nil
}

func sendBulk(command string, keys []string, conn redis.Conn, debulk bool, appendArgs []interface{}) (int, error) {
	var sentCount int
	var argsArray [12]interface{}

	if len(command) > 0 && len(keys) > 0 {
		if debulk {
			for _, key := range keys {
				sentCount++

				args := append(argsArray[:0], key)
				args = append(args, appendArgs...)

				err := conn.Send(command, args...)
				if err != nil {
					return sentCount, err
				}
			}
		} else {
			sentCount++

			args := append(argsArray[:0], make([]interface{}, len(keys))...)
			for i := range keys {
				args[i] = keys[i]
			}
			args = append(args, appendArgs...)

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

type RedisCommand struct {
	Name          string
	KeyType       RedisType
	KeyExtractor  ArgsExtractor
	TransformFunc TransformFunc
	Syntax        string
}

func (r RedisCommand) Keys(args []interface{}) []string {
	var keyArgsArray [12]interface{}
	keyArgs := r.KeyExtractor.AppendArgs(args, keyArgsArray[:0])
	return argsAsStrings(keyArgs)
}

func (r RedisCommand) AppendKeys(args []interface{}, keys []string) []string {
	var keyArgsArray [12]interface{}
	keyArgs := r.KeyExtractor.AppendArgs(args, keyArgsArray[:0])
	return appendArgsAsStrings(keyArgs, keys)
}
