package protocol

import (
	"net"
	"testing"
)

// TestConn_RoundTrip sends two messages back-to-back over an in-memory pipe and
// checks both decode correctly — proving the newline framing separates them and
// the type tags survive the round trip. (Codec is plumbing, so this is GREEN.)
func TestConn_RoundTrip(t *testing.T) {
	p1, p2 := net.Pipe()
	defer p1.Close()
	defer p2.Close()
	sender, receiver := NewConn(p1), NewConn(p2)

	reg := Register{WorkerID: "w1", MaxConcurrency: 42}
	start := StartTest{TestID: "t1", Config: TestConfig{
		TargetURL:            "http://target:8080/",
		RequestsPerWorker:    100,
		ConcurrencyPerWorker: 10,
	}}

	// net.Pipe is synchronous, so write from a goroutine while we read.
	go func() {
		_ = sender.WriteMsg(reg)
		_ = sender.WriteMsg(start)
	}()

	// message 1
	env, err := receiver.ReadMsg()
	if err != nil {
		t.Fatalf("ReadMsg 1: %v", err)
	}
	if env.Type != MsgRegister {
		t.Fatalf("msg1 type = %q, want %q", env.Type, MsgRegister)
	}
	var gotReg Register
	if err := env.Decode(&gotReg); err != nil {
		t.Fatalf("decode register: %v", err)
	}
	if gotReg != reg {
		t.Errorf("register round-trip = %+v, want %+v", gotReg, reg)
	}

	// message 2 (proves framing split the two lines)
	env, err = receiver.ReadMsg()
	if err != nil {
		t.Fatalf("ReadMsg 2: %v", err)
	}
	if env.Type != MsgStartTest {
		t.Fatalf("msg2 type = %q, want %q", env.Type, MsgStartTest)
	}
	var gotStart StartTest
	if err := env.Decode(&gotStart); err != nil {
		t.Fatalf("decode start_test: %v", err)
	}
	if gotStart.Config.RequestsPerWorker != 100 || gotStart.TestID != "t1" {
		t.Errorf("start_test round-trip lost data: %+v", gotStart)
	}
}
