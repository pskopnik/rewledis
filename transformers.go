package rewledis

import (
	"context"
	"errors"
	"fmt"
	"time"

	rewledisArgs "github.com/pskopnik/rewledis/args"

	"github.com/gomodule/redigo/redis"
)

// Error variables related to transformers.
var (
	ErrSubCommandNotImplemented   = errors.New("sub command not implemented")
	ErrSubCommandUnknown          = errors.New("sub command unknown")
	ErrInvalidArgumentCombination = errors.New("invalid argument combination")
	ErrInvalidSyntax              = errors.New("invalid syntax")
	ErrInvalidArgumentType        = errors.New("invalid argument type")
	ErrNoEmulationPossible        = errors.New("no emulation possible for the issued command")
)

const (
	stringEX = "EX"
	stringPX = "PX"
	stringXX = "XX"
	stringNX = "NX"

	stringINCR = "INCR"
	stringCH   = "CH"

	stringEXISTS = "EXISTS"
	stringFLUSH  = "FLUSH"
	stringLOAD   = "LOAD"

	stringREPLACE  = "REPLACE"
	stringABSTTL   = "ABSTTL"
	stringIDLETIME = "IDLETIME"
	stringFREQ     = "FREQ"

	stringLEDIS = "LEDIS"
	stringSELF  = "SELF"
)

var (
	bytesEX = []byte("EX")
	bytesPX = []byte("PX")
	bytesXX = []byte("XX")
	bytesNX = []byte("NX")

	bytesINCR = []byte("INCR")
	bytesCH   = []byte("CH")

	bytesEXISTS = []byte("EXISTS")
	bytesFLUSH  = []byte("FLUSH")
	bytesLOAD   = []byte("LOAD")

	bytesREPLACE  = []byte("REPLACE")
	bytesABSTTL   = []byte("ABSTTL")
	bytesIDLETIME = []byte("IDLETIME")
	bytesFREQ     = []byte("FREQ")

	bytesLEDIS = []byte("LEDIS")
	bytesSELF  = []byte("SELF")
)

var (
	noneTransformerInstance = NoneTransformer()
)

func NoneTransformer() TransformFunc {
	return TransformFunc(
		func(_ *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
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
		},
	)
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
	return TransformFunc(
		func(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
			var typeInfoArray [12]TypeInfo
			var extractedArgsArray [12]interface{}
			var keysArray [12]string

			keyArgs := command.KeyExtractor.AppendArgs(extractedArgsArray[:0], args)
			keys := rewledisArgs.AppendAsSimpleStrings(keysArray[:0], keyArgs)

			resolver := rewriter.Resolver()
			ctx, cancel := context.WithCancel(context.Background())
			typesInfo, err := resolver.ResolveAppend(typeInfoArray[:0], ctx, keys)
			cancel()
			if err != nil {
				return nil, err
			}

			var appendArgs []interface{}
			if config.AppendArgsExtractor != nil {
				appendArgs = config.AppendArgsExtractor.AppendArgs(extractedArgsArray[:0], args)
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
		},
	)
}

// Aggregator creates an aggregator, i.e. a function which reduces any number
// of reply values to a single reply value.
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

// SetCommandTransformer performs transformations for the SET Redis
// command.
//
// This transformer mirrors Redis behaviour: When both XX and EX is supplied,
// nil is returned without applying any changes. When both EX and PX is
// supplied, the PX value takes precedence.
func SetCommandTransformer(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
	commandInfo, err := parseSetCommand(args)
	if err != nil {
		return nil, err
	}

	if commandInfo.XXSet && commandInfo.NXSet {
		return nil, ErrInvalidArgumentCombination
	}
	if commandInfo.EXSet && commandInfo.PXSet {
		return nil, ErrInvalidArgumentCombination
	}

	expSet := commandInfo.EXSet
	expDuration := commandInfo.EX
	if commandInfo.PXSet {
		expSet = true
		// Round up the expiration duration (ignoring overflow)
		expDuration = (commandInfo.PX + 999) / 1000
	}

	if commandInfo.XXSet {
		return nil, ErrNoEmulationPossible
	}
	// if expSet && commandInfo.NXSet {
	// 	// Emulation subject to race-conditions
	// }

	return SendLedisFunc(func(ledisConn redis.Conn) (Slot, error) {
		var err error
		var repliesCount int
		var transformNX bool

		if commandInfo.NXSet {
			transformNX = true
			repliesCount++
			err = ledisConn.Send("SETNX", args[0], args[1])
			if err != nil {
				return Slot{}, err
			}

			if expSet {
				repliesCount++
				err = ledisConn.Send("EXPIRE", args[0], expDuration)
				if err != nil {
					return Slot{}, err
				}
			}
		} else if expSet {
			repliesCount++
			err = ledisConn.Send("SETEX", args[0], expDuration, args[1])
			if err != nil {
				return Slot{}, err
			}
		} else {
			repliesCount++
			err = ledisConn.Send("SET", args[0], args[1])
			if err != nil {
				return Slot{}, err
			}
		}

		return Slot{
			RepliesCount: repliesCount,
			ProcessFunc: func(replies []interface{}) (interface{}, error) {
				if transformNX {
					wasSet, err := redis.Bool(replies[0], nil)
					if err != nil {
						return nil, err
					}
					if !wasSet {
						return nil, nil
					}
					return "OK", nil
				}

				return replies[0], nil
			},
		}, nil
	}), nil
}

type setCommandInfo struct {
	EXSet bool
	EX    int64
	PXSet bool
	PX    int64
	NXSet bool
	XXSet bool
}

func parseSetCommand(args []interface{}) (info setCommandInfo, err error) {
	if len(args) < 2 {
		err = ErrInvalidSyntax
		return
	}

	for i := 2; i < len(args); i++ {
		argInfo := rewledisArgs.Parse(args[i])

		if !argInfo.IsStringLike() {
			err = ErrInvalidArgumentType
			return
		}

		if argInfo.EqualFoldEither(stringEX, bytesEX) {
			if i+1 >= len(args) {
				err = ErrInvalidSyntax
				return
			}

			i++
			valueInfo := rewledisArgs.Parse(args[i])
			info.EX, err = valueInfo.ConvertToInt()
			if err != nil {
				return
			}
			info.EXSet = true
		} else if argInfo.EqualFoldEither(stringPX, bytesPX) {
			if i+1 >= len(args) {
				err = ErrInvalidSyntax
				return
			}

			i++
			valueInfo := rewledisArgs.Parse(args[i])
			info.PX, err = valueInfo.ConvertToInt()
			if err != nil {
				return
			}
			info.PXSet = true
		} else if argInfo.EqualFoldEither(stringNX, bytesNX) {
			info.NXSet = true
		} else if argInfo.EqualFoldEither(stringXX, bytesXX) {
			info.XXSet = true
		} else {
			err = ErrInvalidSyntax
			return
		}
	}

	return
}

var lremScript = redis.NewScript(2, `
local function reverse(arr)
	local i, j = 1, #arr

	while i < j do
		arr[i], arr[j] = arr[j], arr[i]

		i = i + 1
		j = j - 1
	end
end

local listKey = KEYS[1]
local tempListKey = KEYS[2]
local count = tonumber(ARGV[1])
local value = ARGV[2]

local removedCount = 0
local listLen = ledis.call('LLEN', listKey)

if count >= 0
then
	local processed = 0

	for i = 0, listLen - 1, 1 do
		element = ledis.call('LPOP', listKey)
		processed = processed + 1

		if element == value
		then
			removedCount = removedCount + 1
			if removedCount == count
			then
				break
			end
		else
			ledis.call('RPUSH', tempListKey, element)
		end
	end

	if processed < listLen
	then
		remainingElements = ledis.call('LRANGE', listKey, 0, -1)
		ledis.call('RPUSH', tempListKey, unpack(remainingElements))
	end
else
	local processed = 0

	for i = 0, listLen - 1, 1 do
		element = ledis.call('LINDEX', listKey, -1)
		processed = processed + 1

		if element == value
		then
			ledis.call('RPOP', listKey)
			removedCount = removedCount + 1
			if removedCount == -count
			then
				break
			end
		else
			ledis.call('RPOPLPUSH', listKey, tempListKey)
		end
	end

	if processed < listLen
	then
		remainingElements = ledis.call('LRANGE', listKey, 0, -1)
		reverse(remainingElements)
		ledis.call('LPUSH', tempListKey, unpack(remainingElements))
	end
end

-- move temporary list content to the original key

local tempListContent = ledis.call('LDUMP', tempListKey)
local listTTL = ledis.call('LTTL', listKey)

local restoreTTL = 0
if listTTL > -1
then
	restoreTTL = listTTL * 1000
end

ledis.call('RESTORE', listKey, restoreTTL, tempListContent)
ledis.call('LCLEAR', tempListKey)

return removedCount
`)

func LremCommandTransformer(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
	if len(args) < 3 {
		return nil, ErrInvalidSyntax
	}

	argInfo := rewledisArgs.Parse(args[0])
	listKey, err := argInfo.ConvertToRedisString()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	tempListKey := fmt.Sprintf("rewledis:temp:%d%d:%s", now.Unix(), now.Nanosecond(), listKey)

	ctx, cancel := context.WithCancel(context.Background())
	conn, err := rewriter.internalSubPool.getRaw(ctx)
	cancel()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	reply, err := redis.Values(conn.Do("SCRIPT", "EXISTS", lremScript.Hash()))
	if err != nil {
		return nil, err
	}

	var scriptExists int
	_, err = redis.Scan(reply, &scriptExists)
	if err != nil {
		return nil, err
	}

	if scriptExists == 0 {
		err = lremScript.Load(conn)
		if err != nil {
			return nil, err
		}
	}

	return SendLedisFunc(func(ledisConn redis.Conn) (Slot, error) {
		err := lremScript.SendHash(ledisConn, listKey, tempListKey, args[1], args[2])
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
}

func ZaddCommandTransformer(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
	commandInfo, err := parseZaddCommand(args)
	if err != nil {
		return nil, err
	}

	if commandInfo.XXSet && commandInfo.NXSet {
		return nil, ErrInvalidArgumentCombination
	}

	if commandInfo.XXSet || commandInfo.NXSet || commandInfo.CHSet {
		return nil, ErrNoEmulationPossible
	}

	if commandInfo.INCRSet {
		if len(args)-(commandInfo.NumFlags+1) != 2 {
			return nil, ErrInvalidSyntax
		}
	}

	return SendLedisFunc(func(ledisConn redis.Conn) (Slot, error) {
		var err error

		if commandInfo.INCRSet {
			err = ledisConn.Send(
				"ZINCRBY",
				args[0],
				args[commandInfo.NumFlags+1],
				args[commandInfo.NumFlags+2],
			)
			if err != nil {
				return Slot{}, err
			}
		} else {
			err = ledisConn.Send("ZADD", args[0], args[commandInfo.NumFlags+1:])
			if err != nil {
				return Slot{}, err
			}
		}

		return Slot{
			RepliesCount: 1,
			ProcessFunc: func(replies []interface{}) (interface{}, error) {
				return replies[0], nil
			},
		}, nil
	}), nil
}

type zaddCommandInfo struct {
	NumFlags int
	NXSet    bool
	XXSet    bool
	INCRSet  bool
	CHSet    bool
}

func parseZaddCommand(args []interface{}) (info zaddCommandInfo, err error) {
	if len(args) < 3 {
		err = ErrInvalidSyntax
		return
	}

	for i := 1; i < len(args); i++ {
		argInfo := rewledisArgs.Parse(args[i])

		if !argInfo.IsStringLike() {
			break
		}

		info.NumFlags++

		if argInfo.EqualFoldEither(stringNX, bytesNX) {
			info.NXSet = true
		} else if argInfo.EqualFoldEither(stringXX, bytesXX) {
			info.XXSet = true
		} else if argInfo.EqualFoldEither(stringINCR, bytesINCR) {
			info.INCRSet = true
		} else if argInfo.EqualFoldEither(stringCH, bytesCH) {
			info.CHSet = true
		} else {
			break
		}
	}

	if (len(args)-(info.NumFlags+1))%2 != 0 {
		err = ErrInvalidSyntax
		return
	}

	return
}

// RestoreCommandTransformer performs transformations for the RESTORE Redis
// command.
//
// The IDLETIME and FREQ modifiers are ignored and removed when passing on the
// command to LedisDB.
//
// TODO: Figure out how LedisDB performs RESTORE and incorporate into REPLACE
// handling.
func RestoreCommandTransformer(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
	commandInfo, err := parseRestoreCommand(args)
	if err != nil {
		return nil, err
	}

	// if commandInfo.REPLACESet {
	// 	// TODO
	// }

	var setExpireAt bool
	var expireAtTimestamp int64
	if commandInfo.ABSTTLSet {
		setExpireAt = true
		argInfo := rewledisArgs.Parse(args[1])
		expireAtTimestamp, err = argInfo.ConvertToInt()
		if err != nil {
			return nil, err
		}
	}

	return SendLedisFunc(func(ledisConn redis.Conn) (Slot, error) {
		var err error
		var repliesCount int

		if commandInfo.ABSTTLSet {
			repliesCount++
			err = ledisConn.Send("RESTORE", args[0], 0, args[2])
			if err != nil {
				return Slot{}, err
			}

			if setExpireAt {
				repliesCount++
				err = ledisConn.Send("EXPIREAT", args[0], expireAtTimestamp)
				if err != nil {
					return Slot{}, err
				}
			}
		} else {
			repliesCount++
			err = ledisConn.Send("RESTORE", args[0], args[1], args[2])
			if err != nil {
				return Slot{}, err
			}
		}

		return Slot{
			RepliesCount: repliesCount,
			ProcessFunc: func(replies []interface{}) (interface{}, error) {
				return replies[0], nil
			},
		}, nil
	}), nil
}

type restoreCommandInfo struct {
	REPLACESet  bool
	ABSTTLSet   bool
	IDLETIMESet bool
	IDLETIME    int64
	FREQSet     bool
	FREQ        int64
}

func parseRestoreCommand(args []interface{}) (info restoreCommandInfo, err error) {
	if len(args) < 3 {
		err = ErrInvalidSyntax
		return
	}

	for i := 1; i < len(args); i++ {
		argInfo := rewledisArgs.Parse(args[i])

		if !argInfo.IsStringLike() {
			err = ErrInvalidArgumentType
			return
		}

		if argInfo.EqualFoldEither(stringREPLACE, bytesREPLACE) {
			info.REPLACESet = true
		} else if argInfo.EqualFoldEither(stringABSTTL, bytesABSTTL) {
			info.ABSTTLSet = true
		} else if argInfo.EqualFoldEither(stringIDLETIME, bytesIDLETIME) {
			if i+1 >= len(args) {
				err = ErrInvalidSyntax
				return
			}

			i++
			valueInfo := rewledisArgs.Parse(args[i])
			info.IDLETIME, err = valueInfo.ConvertToInt()
			if err != nil {
				return
			}
			info.IDLETIMESet = true
		} else if argInfo.EqualFoldEither(stringFREQ, bytesFREQ) {
			if i+1 >= len(args) {
				err = ErrInvalidSyntax
				return
			}

			i++
			valueInfo := rewledisArgs.Parse(args[i])
			info.FREQ, err = valueInfo.ConvertToInt()
			if err != nil {
				return
			}
			info.IDLETIMESet = true
		} else {
			err = ErrInvalidSyntax
			return
		}
	}

	return
}

// PingCommandTransformer performs transformations for the PING Redis command.
func PingCommandTransformer(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
	if len(args) == 0 {
		return noneTransformerInstance(rewriter, command, args)
	} else if len(args) == 1 {
		argInfo := rewledisArgs.Parse(args[0])

		message, err := argInfo.ConvertToRedisString()
		if err != nil {
			return nil, err
		}

		return SendLedisFunc(func(ledisConn redis.Conn) (Slot, error) {
			err := ledisConn.Send(command.Name)
			if err != nil {
				return Slot{}, err
			}

			return Slot{
				RepliesCount: 1,
				ProcessFunc: func(_ []interface{}) (interface{}, error) {
					return message, nil
				},
			}, nil
		}), nil
	} else {
		return nil, ErrInvalidSyntax
	}
}

// TransactionTransformer performs transformations for transaction related
// Redis commands.
//
// Support for transaction commands has not yet been integrated into rewledis.
// The transformer drops all commands except DISCARD for which
// ErrNoEmulationPossible is returned. This "emulation" is subject to
// race-conditions.
func TransactionTransformer(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
	switch command.Name {
	case "DISCARD":
		return nil, ErrNoEmulationPossible
	case "EXEC":
		fallthrough
	case "MULTI":
		fallthrough
	case "UNWATCH":
		fallthrough
	case "WATCH":
		return SendLedisFunc(func(_ redis.Conn) (Slot, error) {
			return Slot{
				RepliesCount: 0,
				ProcessFunc: func(_ []interface{}) (interface{}, error) {
					return nil, nil
				},
			}, nil
		}), nil
	default:
		return nil, ErrSubCommandUnknown
	}
}

// ScriptCommandTransformer performs transformations for the SCRIPT Redis
// command.
//
// Only some sub-commands are supported. Issuing a not supported sub-command
// results in a ErrSubCommandNotImplemented error.
//
//     Implemented:
//       SCRIPT EXISTS sha1 [sha1 ...]
//       SCRIPT FLUSH
//       SCRIPT LOAD script
//     Not implemented:
//       SCRIPT DEBUG YES|SYNC|NO
//       SCRIPT KILL
func ScriptCommandTransformer(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
	if len(args) < 1 {
		return nil, ErrInvalidSyntax
	}

	argInfo := rewledisArgs.Parse(args[0])
	if !argInfo.IsStringLike() {
		return nil, ErrInvalidArgumentType
	}

	if argInfo.EqualFoldEither(stringEXISTS, bytesEXISTS) {
		return noneTransformerInstance(rewriter, command, args)
	} else if argInfo.EqualFoldEither(stringFLUSH, bytesFLUSH) {
		return noneTransformerInstance(rewriter, command, args)
	} else if argInfo.EqualFoldEither(stringLOAD, bytesLOAD) {
		return noneTransformerInstance(rewriter, command, args)
	} else {
		return nil, ErrSubCommandNotImplemented
	}
}

// UnsafeCommandTransformer performs transformations for the UNSAFE Redis
// command provided by rewledis.
func UnsafeCommandTransformer(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
	if len(args) < 1 {
		return nil, ErrInvalidSyntax
	}
	argInfo := rewledisArgs.Parse(args[0])
	if !argInfo.IsStringLike() {
		return nil, ErrInvalidArgumentType
	}

	if argInfo.EqualFoldEither(stringLEDIS, bytesLEDIS) {
		if len(args) < 2 {
			return nil, ErrInvalidSyntax
		}
		commandArgInfo := rewledisArgs.Parse(args[1])
		if !commandArgInfo.IsStringLike() {
			return nil, ErrInvalidArgumentType
		}
		commandName, err := commandArgInfo.ConvertToRedisString()
		if err != nil {
			return nil, err
		}

		return SendLedisFunc(func(ledisConn redis.Conn) (Slot, error) {
			err := ledisConn.Send(commandName, args[2:]...)
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
	} else if argInfo.EqualFoldEither(stringSELF, bytesSELF) {
		if len(args) != 1 {
			return nil, ErrInvalidSyntax
		}

		return SendLedisFunc(func(ledisConn redis.Conn) (Slot, error) {
			return Slot{
				RepliesCount: 0,
				ProcessFunc: func(_ []interface{}) (interface{}, error) {
					return ledisConn, nil
				},
			}, nil
		}), nil
	} else {
		return nil, ErrSubCommandUnknown
	}
}
