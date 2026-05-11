package network

import (
	"io"
	"log"
	"net"

	"github.com/sverrehu/spacegame/internal/utils"
)

type Transceiver interface {
	Send(msg Message) error
	Receive() (Message, error)
	Close()
}

type TcpTransceiver struct {
	conn             net.Conn
	incoming         *utils.BlockingQueue
	outgoing         *utils.BlockingQueue
	incomingError    error
	outgoingError    error
	IncomingCallback func(msg Message) error
}

func NewTransceiver(conn net.Conn, incomingCallback func(msg Message) error) *TcpTransceiver {
	t := TcpTransceiver{conn: conn, incoming: utils.NewBlockingQueue(), outgoing: utils.NewBlockingQueue(), incomingError: nil, outgoingError: nil, IncomingCallback: incomingCallback}
	go t.sender()
	go t.receiver()
	return &t
}

func (t *TcpTransceiver) Send(msg Message) error {
	if t.outgoingError != nil {
		return t.outgoingError
	}
	t.outgoing.Enqueue(msg)
	return nil
}

func (t *TcpTransceiver) Receive() (Message, error) {
	if t.incomingError != nil {
		return nil, t.incomingError
	}
	return t.incoming.Dequeue().(Message), nil
}

func (t *TcpTransceiver) Close() {
	_ = t.conn.Close()
}

func (t *TcpTransceiver) sender() {
	for {
		msg := t.outgoing.Dequeue().(Message)
		if msg.GetType() == MessageAbort {
			log.Println("Got abort; terminating sender.")
			break
		}
		bytes := EncodeMessage(msg)
		_, err := t.conn.Write(encInt32(int32(len(bytes))))
		if err != nil {
			t.outgoingError = err
			break
		}
		_, err = t.conn.Write(bytes)
		if err != nil {
			t.outgoingError = err
			break
		}
	}
	log.Println("Transceiver sender terminated")
}

func (t *TcpTransceiver) receiver() {
	for {
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(t.conn, lenBuf)
		if err != nil {
			t.incomingError = err
			break
		}
		l := decInt32(lenBuf)
		bytes := make([]byte, l)
		_, err = io.ReadFull(t.conn, bytes)
		if err != nil {
			t.incomingError = err
			break
		}
		msg := DecodeMessage(bytes)
		if msg.GetType() == MessageAbort {
			break
		}
		err = t.IncomingCallback(msg)
		if err != nil {
			t.incomingError = err
			break
		}
	}
	log.Println("Transceiver receiver terminated")
}

func encInt32(v int32) []byte {
	return []byte{
		byte((v >> 24) & 0xff),
		byte((v >> 16) & 0xff),
		byte((v >> 8) & 0xff),
		byte(v & 0xff),
	}
}

func decInt32(b []byte) int32 {
	return int32(b[0])<<24 + int32(b[1])<<16 + int32(b[2])<<8 + int32(b[3])
}
