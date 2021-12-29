# tcp_serve_mux
a serve mux used in tcp net.conn using go generics with go1.18beta1



```go
package mux // import "tcpServeMux"

func FrameDecodeFuncJson(conn net.Conn) ([]byte, error)
func MsgDecodeFuncJson[T any](bs []byte) (string, *T, error)
type FrameDecoder func(conn net.Conn) ([]byte, error)
type MsgDecoder[T any] func(bs []byte) (string, *T, error)
type MsgHandleFunc[T any, CTX any] func(ctx *CTX, msg *T)
    func HandleNullMsgWithCause[T any, CTX any](msgs ...string) MsgHandleFunc[T, CTX]
type MsgMux[T any, CTX any] struct{ ... }
    func NewMsgMux[T any, CTX any]() *MsgMux[T, CTX]
```


## usage example:
```go
func main() {
	mux := msgGW.NewMsgMux[model.Message, msgGW.ConnContext]()

	gw := msgGW.NewGateway()

	mux.RegisterMsgHandleFunc(model.MsgActionType.Heartbeat, gw.HandleHeartbeat)
	mux.RegisterInitHandleFunc(gw.HandleOnline)

	mux.RegisterFrameDecodeFunc(msgGW.FrameDecodeFuncJson)
	mux.RegisterMsgDecodeFunc(model.MsgDecodeFuncJson)

	addr := ":" + env.GetMyTcpPort()

	lis, err := net.Listen("tcp", addr)
	log.FatalIfErr(err)

	log.Logger.Info().Msg("msg_gateway serving on " + addr)

	log.FatalIfErr(mux.Serve(lis))
}

```