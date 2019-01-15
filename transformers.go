package rewledis

import (
	"errors"

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
)

var (
	noneTransformerInstance = NoneTransformer()
)

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

	argInfo := parseArg(args[0])
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
		argInfo := parseArg(args[i])

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
			valueInfo := parseArg(args[i])
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
			valueInfo := parseArg(args[i])
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
		argInfo := parseArg(args[i])

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
		argInfo := parseArg(args[1])
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
		argInfo := parseArg(args[i])

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
			valueInfo := parseArg(args[i])
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
			valueInfo := parseArg(args[i])
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

// UnsafeCommandTransformer performs transformations for the UNSAFE Redis
// command provided by rewledis.
func UnsafeCommandTransformer(rewriter *Rewriter, command *RedisCommand, args []interface{}) (SendLedisFunc, error) {
	if len(args) < 1 {
		return nil, ErrInvalidSyntax
	}
	argInfo := parseArg(args[0])
	if !argInfo.IsStringLike() {
		return nil, ErrInvalidArgumentType
	}

	if argInfo.EqualFoldEither(stringLEDIS, bytesLEDIS) {
		if len(args) < 2 {
			return nil, ErrInvalidSyntax
		}
		commandArgInfo := parseArg(args[1])
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
	} else {
		return nil, ErrSubCommandUnknown
	}
}
