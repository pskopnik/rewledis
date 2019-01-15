package rewledis

import (
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Error variables related to the LedisConn type.
var (
	ErrTimeoutNotSupported = errors.New("rewledis: connection does not support ConnWithTimeout")
	ErrConnClosed          = errors.New("rewledis: connection closed")
)

var _ redis.Conn = &LedisConn{}

// LedisConn is a rewriting wrapper around a connection to a LedisDB server.
// Many Redis commands can be issued on this connection. The commands are
// dynamically rewritten to their LedisDB equivalent. See the repository's
// README for more information on which commands are supported.
type LedisConn struct {
	slots    SlotDeque
	rewriter *Rewriter
	// err stores any error which occurred during processing in Send(),
	// Receive(), Flush() and Do(). err may store errors occuring during
	// rewriting or errors returned by the methods of the underlying connection.
	err  error
	conn redis.Conn
}

// RawConn returns the underlying connection to the LedisDB server.
//
// This method must be used with care, as the stack of internally stored reply
// procedures (Slots) can no longer be maintained. RawConn should only be used
// on connections without any pending replies.
func (l *LedisConn) RawConn() redis.Conn {
	return l.conn
}

func (l *LedisConn) fatal(err error) error {
	if l.err == nil {
		l.err = err
	}
	if l.conn != nil {
		_ = l.Close()
		l.conn = nil
	}

	return err
}

// Close closes the connection.
func (l *LedisConn) Close() error {
	if l.conn == nil {
		return ErrConnClosed
	}

	err := l.conn.Close()
	l.conn = nil
	l.slots.Clear()
	return err
}

// Err returns a non-nil value when the connection is not usable.
func (l *LedisConn) Err() error {
	if l.err != nil {
		return l.err
	}
	if l.conn == nil {
		return ErrConnClosed
	}
	// All errors should be returned by now (l.err captures all errors). To be
	// safe, also check Err() of the underlying connection.
	return l.conn.Err()
}

// Send writes the command to the client's output buffer.
func (l *LedisConn) Send(commandName string, args ...interface{}) error {
	if l.conn == nil {
		return ErrConnClosed
	}

	slot, err := l.rewriteAndSend(commandName, args...)
	if err != nil {
		return l.fatal(err)
	}

	l.slots.PushBack(slot)

	return nil
}

// Flush flushes the output buffer to the Redis server.
func (l *LedisConn) Flush() error {
	if l.conn == nil {
		return ErrConnClosed
	}

	err := l.conn.Flush()
	if err != nil {
		return l.fatal(err)
	}

	return nil
}

// Receive receives a single reply from the Redis server.
func (l *LedisConn) Receive() (interface{}, error) {
	if l.conn == nil {
		return nil, ErrConnClosed
	}

	slot := l.slots.PopFront()

	var repliesArray [4]interface{}
	replies, err := l.receiveRepliesAppend(slot.RepliesCount, repliesArray[:0])
	if err != nil {
		return nil, l.fatal(err)
	}

	reply, err := slot.ProcessFunc(replies)
	if err != nil {
		return nil, l.fatal(err)
	}

	return reply, nil
}

// Receive receives a single reply from the Redis server. The timeout
// overrides the read timeout set when dialing the connection.
func (l *LedisConn) ReceiveWithTimeout(timeout time.Duration) (interface{}, error) {
	if l.conn == nil {
		return nil, ErrConnClosed
	}

	connWithTimeout, ok := l.conn.(redis.ConnWithTimeout)
	if !ok {
		return nil, ErrTimeoutNotSupported
	}

	slot := l.slots.PopFront()

	var repliesArray [4]interface{}
	replies, err := l.receiveRepliesWithTimeoutAppend(connWithTimeout, slot.RepliesCount, timeout, repliesArray[:0])
	if err != nil {
		return nil, l.fatal(err)
	}

	reply, err := slot.ProcessFunc(replies)
	if err != nil {
		return nil, l.fatal(err)
	}

	return reply, nil
}

// Do sends a command to the server and returns the received reply.
func (l *LedisConn) Do(commandName string, args ...interface{}) (interface{}, error) {
	if l.conn == nil {
		return nil, ErrConnClosed
	}

	var err error
	var slot Slot

	if len(commandName) > 0 {
		slot, err = l.rewriteAndSend(commandName, args...)
		if err != nil {
			return nil, l.fatal(err)
		}
	}

	err = l.conn.Flush()
	if err != nil {
		return nil, l.fatal(err)
	}

	err = l.consumeSlots()
	if err != nil {
		return nil, l.fatal(err)
	}

	if len(commandName) > 0 {
		var repliesArray [4]interface{}
		replies, err := l.receiveRepliesAppend(slot.RepliesCount, repliesArray[:0])
		if err != nil {
			return nil, l.fatal(err)
		}

		reply, err := slot.ProcessFunc(replies)
		if err != nil {
			return nil, l.fatal(err)
		}

		return reply, nil
	}

	return nil, nil
}

// Do sends a command to the server and returns the received reply. The
// timeout overrides the read timeout set when dialing the connection.
func (l *LedisConn) DoWithTimeout(timeout time.Duration, commandName string, args ...interface{}) (interface{}, error) {
	if l.conn == nil {
		return nil, ErrConnClosed
	}

	connWithTimeout, ok := l.conn.(redis.ConnWithTimeout)
	if !ok {
		return nil, ErrTimeoutNotSupported
	}

	var err error
	var slot Slot

	if len(commandName) > 0 {
		slot, err = l.rewriteAndSend(commandName, args...)
		if err != nil {
			return nil, l.fatal(err)
		}
	}

	err = l.conn.Flush()
	if err != nil {
		return nil, l.fatal(err)
	}

	err = l.consumeSlots()
	if err != nil {
		return nil, l.fatal(err)
	}

	if len(commandName) > 0 {
		var repliesArray [4]interface{}
		replies, err := l.receiveRepliesWithTimeoutAppend(connWithTimeout, slot.RepliesCount, timeout, repliesArray[:0])
		if err != nil {
			return nil, l.fatal(err)
		}

		reply, err := slot.ProcessFunc(replies)
		if err != nil {
			return nil, l.fatal(err)
		}

		return reply, nil
	}

	return nil, nil
}

func (l *LedisConn) consumeSlots() error {
	for i := 0; i < l.slots.Len(); i++ {
		for j := 0; j < l.slots.At(i).RepliesCount; j++ {
			_, err := l.conn.Receive()
			if err != nil {
				return err
			}
		}
	}

	l.slots.Clear()

	return nil
}

func (l *LedisConn) rewriteAndSend(commandName string, args ...interface{}) (Slot, error) {
	sendLedisFunc, err := l.rewriter.Rewrite(commandName, args...)
	if err != nil {
		return Slot{}, err
	}

	slot, err := sendLedisFunc(l.conn)
	if err != nil {
		return Slot{}, err
	}

	return slot, nil
}

func (l *LedisConn) receiveRepliesAppend(count int, replies []interface{}) ([]interface{}, error) {
	baseInd := len(replies)
	replies = append(replies, make([]interface{}, count)...)

	for i := 0; i < count; i++ {
		reply, err := l.conn.Receive()
		if err != nil {
			return nil, err
		}
		replies[baseInd+i] = reply
	}

	return replies, nil
}

func (l *LedisConn) receiveRepliesWithTimeoutAppend(
	connWithTimeout redis.ConnWithTimeout,
	count int,
	timeout time.Duration,
	replies []interface{},
) ([]interface{}, error) {
	baseInd := len(replies)
	replies = append(replies, make([]interface{}, count)...)

	deadline := time.Now().Add(timeout)

	for i := 0; i < count; i++ {
		timeout := time.Now().Sub(deadline)
		reply, err := connWithTimeout.ReceiveWithTimeout(timeout)
		if err != nil {
			return nil, err
		}
		replies[baseInd+i] = reply
	}

	return replies, nil
}
