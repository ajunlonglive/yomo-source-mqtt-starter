package receiver

import (
	"net"
	"sync"
	"time"

	"github.com/yomorun/yomo-source-mqtt-starter/internal/logger"

	"github.com/yomorun/yomo-source-mqtt-starter/internal/comm"

	"github.com/yomorun/yomo-source-mqtt-starter/internal/pool"

	"go.uber.org/zap"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

var (
	log = logger.Prod().Named("receiver")
)

type Message struct {
	client *client
	packet packets.ControlPacket
}

type Receiver struct {
	handler func(topic string, payload []byte)

	id      string
	clients sync.Map
	config  *Config
	//sessionMgr *sessions.Manager
	wpool *pool.WorkerPool
}

func NewReceiver(config *Config) (*Receiver, error) {
	b := &Receiver{
		id:     comm.GenUniqueId(),
		config: config,
		wpool:  pool.New(config.Worker),
	}

	if config.Debug {
		log = logger.Debug().Named("receiver")
	}

	//var err error
	//b.sessionMgr, err = sessions.NewManager("mem")
	//if err != nil {
	//	log.Error("new session manager error ", zap.Error(err))
	//	return nil, err
	//}

	return b, nil

}

func (b *Receiver) Start(handler func(topic string, payload []byte)) {

	if b == nil {
		log.Error("receiver is null")
		return
	}

	if handler == nil {
		log.Error("handler is null")
		return
	}

	// 注册处理器
	b.handler = handler

	if len(b.config.ServerAddr) == 0 {
		panic("must set ServerAddr")
	}

	go b.StartClientListening(false)

	log.Debug("start listening",
		zap.String("ServerAddr", b.config.ServerAddr),
		zap.Int("Worker", b.config.Worker),
		zap.Bool("Debug", b.config.Debug))
}

func (b *Receiver) StartClientListening(Tls bool) {
	var err error
	var l net.Listener

	for {
		hp := b.config.ServerAddr
		l, err = net.Listen("tcp", hp)
		log.Info("start listening client", zap.String("host-port", hp), zap.String("id", b.id))

		if err != nil {
			log.Error("Error listening", zap.Error(err))
			time.Sleep(1 * time.Second)
		} else {
			break // successfully listening
		}
	}
	tmpDelay := 10 * comm.ACCEPT_MIN_SLEEP
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error("Accept Accept err.", zap.Error(err))
		}

		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Info("Temporary Client Accept Error(%v), sleeping %vms",
					zap.Error(ne), zap.Duration("sleeping", tmpDelay/time.Millisecond))
				time.Sleep(tmpDelay)
				tmpDelay *= 2
				if tmpDelay > comm.ACCEPT_MAX_SLEEP {
					tmpDelay = comm.ACCEPT_MAX_SLEEP
				}
			} else {
				log.Error("Accept error: %v", zap.Error(err))
			}
			continue
		}
		tmpDelay = comm.ACCEPT_MIN_SLEEP
		go b.handleConnection(CLIENT, conn)

	}
}

func (b *Receiver) handleConnection(typ int, conn net.Conn) {
	//process connect packet
	packet, err := packets.ReadPacket(conn)
	if err != nil {
		log.Error("read connect packet error: ", zap.Error(err))
		return
	}
	if packet == nil {
		log.Error("received nil packet")
		return
	}
	msg, ok := packet.(*packets.ConnectPacket)
	if !ok {
		log.Error("received msg that was not Connect")
		return
	}

	//log.Info("read connect from ", zap.String("clientID", msg.ClientIdentifier))

	connack := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
	connack.SessionPresent = msg.CleanSession
	connack.ReturnCode = msg.Validate()

	if connack.ReturnCode != packets.Accepted {
		err = connack.Write(conn)
		if err != nil {
			log.Error("send connack error, ", zap.Error(err), zap.String("clientID", msg.ClientIdentifier))
			return
		}
		return
	}

	if typ == CLIENT && !b.CheckConnectAuth(string(msg.ClientIdentifier), string(msg.Username), string(msg.Password)) {
		connack.ReturnCode = packets.ErrRefusedNotAuthorised
		err = connack.Write(conn)
		if err != nil {
			log.Error("send connack error, ", zap.Error(err), zap.String("clientID", msg.ClientIdentifier))
			return
		}
		return
	}

	err = connack.Write(conn)
	if err != nil {
		log.Error("send connack error, ", zap.Error(err), zap.String("clientID", msg.ClientIdentifier))
		return
	}

	info := info{
		clientID:  msg.ClientIdentifier,
		username:  msg.Username,
		password:  msg.Password,
		keepalive: msg.Keepalive,
	}

	c := &client{
		typ:      typ,
		receiver: b,
		conn:     conn,
		info:     info,
	}

	c.init()

	//err = b.getSession(c, msg, connack)
	//if err != nil {
	//	log.Error("get session error: ", zap.String("clientID", c.info.clientID))
	//	return
	//}

	//cid := c.info.clientID

	//var exist bool
	//var old interface{}
	//
	//switch typ {
	//case CLIENT:
	//	old, exist = b.clients.Load(cid)
	//	if exist {
	//		ol, ok := old.(*client)
	//		if ok {
	//			log.Warn("client exist, close old...",
	//				zap.Any("clientID", c.info.clientID),
	//				zap.Int("status,", ol.status),
	//				zap.Error(ol.ctx.Err()))
	//			ol.cancelFunc()
	//			ol.Close()
	//		}
	//	}
	//	b.clients.Store(cid, c)
	//}

	c.readLoop()
}

func (b *Receiver) removeClient(c *client) {
	clientId := string(c.info.clientID)
	typ := c.typ
	switch typ {
	case CLIENT:
		b.clients.Delete(clientId)
	}
	log.Info("delete client ", zap.String("clientId", clientId))
}

func (b *Receiver) SubmitWork(clientId string, msg *Message) {
	//if b.wpool == nil {
	//	b.wpool = pool.New(b.config.Worker)
	//}
	//
	//b.wpool.Submit(clientId, func() {
	//	ProcessMessage(msg)
	//})
	ProcessMessage(msg)
}
