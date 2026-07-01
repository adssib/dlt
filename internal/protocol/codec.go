package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

// Conn is a newline-delimited JSON message transport over a net.Conn. Each
// message is one Envelope, JSON-encoded, terminated by '\n' (see ADR-0004).
//
// Conn is intended for a single reader goroutine and a single writer goroutine.
type Conn struct {
	raw net.Conn
	r   *bufio.Reader
	w   *bufio.Writer
}

// NewConn wraps a net.Conn with buffered reading/writing.
func NewConn(c net.Conn) *Conn {
	return &Conn{raw: c, r: bufio.NewReader(c), w: bufio.NewWriter(c)}
}

// Close closes the underlying connection.
func (c *Conn) Close() error { return c.raw.Close() }

// WriteMsg wraps a typed message (Register, Progress, StartTest, ...) in an
// Envelope and writes it as one newline-terminated JSON line.
func (c *Conn) WriteMsg(v any) error {
	t, err := msgTypeOf(v)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	line, err := json.Marshal(Envelope{Type: t, Payload: payload})
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	if _, err := c.w.Write(line); err != nil {
		return err
	}
	if err := c.w.WriteByte('\n'); err != nil {
		return err
	}
	return c.w.Flush()
}

// ReadMsg reads exactly one newline-delimited Envelope. The caller inspects
// .Type and calls Envelope.Decode to unmarshal .Payload into the matching struct.
func (c *Conn) ReadMsg() (Envelope, error) {
	line, err := c.r.ReadBytes('\n')
	if err != nil {
		return Envelope{}, err
	}
	var env Envelope
	if err := json.Unmarshal(line, &env); err != nil {
		return Envelope{}, fmt.Errorf("unmarshal envelope: %w", err)
	}
	return env, nil
}

// msgTypeOf maps a concrete message struct to its wire tag.
func msgTypeOf(v any) (MsgType, error) {
	switch v.(type) {
	case Register, *Register:
		return MsgRegister, nil
	case Progress, *Progress:
		return MsgProgress, nil
	case Results, *Results:
		return MsgResults, nil
	case StartTest, *StartTest:
		return MsgStartTest, nil
	case StopTest, *StopTest:
		return MsgStopTest, nil
	default:
		return "", fmt.Errorf("protocol: unknown message type %T", v)
	}
}
