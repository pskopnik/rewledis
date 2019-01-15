package rewledis

import (
	"errors"

	rewledisArgs "github.com/pskopnik/rewledis/args"

	"github.com/gomodule/redigo/redis"
)

type ArgsExtractor interface {
	AppendArgs(extracted []interface{}, args []interface{}) []interface{}
	Args(args []interface{}) []interface{}
}

var _ ArgsExtractor = ArgsExtractorFunc(nil)

type ArgsExtractorFunc func(extracted []interface{}, args []interface{}) []interface{}

func (a ArgsExtractorFunc) AppendArgs(extracted []interface{}, args []interface{}) []interface{} {
	return a(extracted, args)
}

func (a ArgsExtractorFunc) Args(args []interface{}) []interface{} {
	return a(nil, args)
}

// ArgsAtIndices returns an ArgsExtractor.
// The ArgsExtractor returns the arguments with the indices passed to
// ArgsAtIndices.
func ArgsAtIndices(indices ...int) ArgsExtractor {
	return ArgsExtractorFunc(func(extracted []interface{}, args []interface{}) []interface{} {
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

	return ArgsExtractorFunc(func(extracted []interface{}, args []interface{}) []interface{} {
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
	return r.AppendKeys(nil, args)
}

func (r RedisCommand) AppendKeys(keys []string, args []interface{}) []string {
	var keyArgsArray [12]interface{}
	keyArgs := r.KeyExtractor.AppendArgs(keyArgsArray[:0], args)
	return rewledisArgs.AppendAsSimpleStrings(keys, keyArgs)
}
