package mux

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"strings"
	"sync"
)

type MsgHandleFunc[T any, CTX any] func(ctx *CTX, msg *T)

func HandleNullMsgWithCause[T any, CTX any](msgs ...string) MsgHandleFunc[T, CTX] {
	msg := strings.Join(msgs, " ")
	return func(_ *CTX, _ *T) {
		log.Println("HandleNullMsg func called, because of " + msg)
	}
}

type (
	FrameDecoder      func(conn net.Conn) ([]byte, error)
	MsgDecoder[T any] func(bs []byte) (string, *T, error)
)

type MsgMux[T any, CTX any] struct {
	m               sync.Mutex
	FrameDecodeFunc FrameDecoder
	MsgDecodeFunc   MsgDecoder[T]
	route           map[string]MsgHandleFunc[T, CTX]
	onInit          func(conn net.Conn, ctx *CTX, msg *T)
}

func NewMsgMux[T any, CTX any]() *MsgMux[T, CTX] {
	return &MsgMux[T, CTX]{
		m:     sync.Mutex{},
		route: map[string]MsgHandleFunc[T, CTX]{},
	}
}

func (s *MsgMux[T, CTX]) RegisterFrameDecodeFunc(f FrameDecoder) {
	s.FrameDecodeFunc = f
}

func (s *MsgMux[T, CTX]) RegisterMsgDecodeFunc(f MsgDecoder[T]) {
	s.MsgDecodeFunc = f
}

func (s *MsgMux[T, CTX]) Serve(lis net.Listener) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalln("failed to accept conn from listener, error:" + err.Error())
			return err
		}

		go func() {
			defer conn.Close()
			ctx := new(CTX)

			if s.onInit != nil {
				_, msg, err := s.decodeOnce(ctx, conn)
				if err != nil {
					log.Fatalln("decode error:" + err.Error())
					return
				}
				s.onInit(conn, ctx, msg)
			}

			for {
				actionType, msg, err := s.decodeOnce(ctx, conn)
				if err != nil {
					log.Fatalln("decode error:" + err.Error())
					return
				}

				s.Handle(actionType, ctx, msg)
			}
		}()
	}
}

func (s *MsgMux[T, CTX]) decodeOnce(ctx *CTX, conn net.Conn) (_ string, t *T, err error) {
	frame, err := s.FrameDecodeFunc(conn)
	if err != nil {
		log.Fatalln("failed to decode frame from net.conn, error:" + err.Error())
		return
	}

	actionType, msg, err := s.MsgDecodeFunc(frame)
	if err != nil {
		log.Fatalln("failed to decode messsage from frame bytes, error:" + err.Error())
		return
	}

	return actionType, msg, nil
}

func (m *MsgMux[T, CTX]) Handle(actionType string, ctx *CTX, msg *T) {
	f := m.GetMsgHandleFunc(actionType)
	f(ctx, msg)
}

func (m *MsgMux[T, CTX]) RegisterMsgHandleFunc(path string, handleFunc MsgHandleFunc[T, CTX]) {
	m.m.Lock()
	defer m.m.Unlock()

	m.route[path] = handleFunc
}

func (m *MsgMux[T, CTX]) RegisterInitHandleFunc(handleFunc func(conn net.Conn, ctx *CTX, msg *T)) {
	m.onInit = handleFunc
}

func (m *MsgMux[T, CTX]) GetMsgHandleFunc(actionType string) MsgHandleFunc[T, CTX] {
	handleFunc, ok := m.route[actionType]
	if !ok {
		return HandleNullMsgWithCause[T, CTX]("unregistered action type:", actionType)
	}

	return handleFunc
}

func FrameDecodeFuncJson(conn net.Conn) ([]byte, error) {
	reader := bufio.NewReader(conn)

	frame, err := reader.ReadBytes('}')
	if err != nil {
		return nil, err
	}

	return frame, nil
}

func MsgDecodeFuncJson[T any](bs []byte) (string, *T, error) {
	var msg T
	err := json.Unmarshal(bs, &msg)
	if err != nil {
		return "", nil, err
	}

	return "", &msg, nil
}
