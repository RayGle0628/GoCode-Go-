package packet

import (
	"chapter13/chatbycellnet/cellnet"
	"chapter13/chatbycellnet/cellnet/socket"
)

type MsgEvent struct {
	Ses cellnet.Session
	Msg interface{}
}

func invokeMsgFunc(ses cellnet.Session, f SessionMessageFunc, msg interface{}) {
	q := ses.Peer().Queue()

	// Peer有队列时，在队列线程调用用户处理函数
	if q != nil {
		q.Post(func() {

			f(ses, msg)
		})

	} else {

		// 在I/O线程调用用户处理函数
		f(ses, msg)
	}
}

type SessionMessageFunc func(ses cellnet.Session, raw interface{})

func NewMessageCallback(f SessionMessageFunc) cellnet.EventFunc {

	return func(raw interface{}) interface{} {

		switch ev := raw.(type) {
		case socket.RecvEvent: // 接收数据事件
			return onRecvLTVPacket(ev.Ses, f)
		case socket.SendEvent: // 发送数据事件
			return onSendLTVPacket(ev.Ses, ev.Msg)
		case socket.ConnectErrorEvent: // 连接错误事件
			invokeMsgFunc(ev.Ses, f, raw)
		case socket.SessionStartEvent: // 会话开始事件（连接上/接受连接）
			invokeMsgFunc(ev.Ses, f, raw)
		case socket.SessionClosedEvent: // 会话关闭事件
			invokeMsgFunc(ev.Ses, f, raw)
		case socket.SessionExitEvent: // 会话退出事件
			invokeMsgFunc(ev.Ses, f, raw)
		case socket.RecvErrorEvent: // 接收错误事件
			log.Errorf("<%s> socket.RecvErrorEvent: %s\n", ev.Ses.Peer().Name(), ev.Error)
			invokeMsgFunc(ev.Ses, f, raw)
		case socket.SendErrorEvent: // 发送错误事件
			log.Errorf("<%s> socket.SendErrorEvent: %s, msg: %#v\n", ev.Ses.Peer().Name(), ev.Error, ev.Msg)
			invokeMsgFunc(ev.Ses, f, raw)
		}

		return nil
	}
}
