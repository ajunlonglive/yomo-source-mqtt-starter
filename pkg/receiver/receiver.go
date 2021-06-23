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
	handler      func(topic string, payload []byte, writer ISourceWriter) error
	sourceClient ISourceClient

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

	return b, nil
}

func (b *Receiver) Start(handler func(topic string, payload []byte, writer ISourceWriter) error, sourceStream ISourceClient) {

	if b == nil {
		panic("receiver is null")
	}

	if handler == nil {
		panic("handler is null")
	}
	// 注册处理器
	b.handler = handler

	if sourceStream == nil {
		panic("please register ISourceClient")
	}
	// 注册ISourceStream
	b.sourceClient = sourceStream

	if len(b.config.ServerAddr) == 0 {
		panic("must set ServerAddr")
	}

	go b.StartClientListening(false)

	//log.Debug("start listening",
	//	zap.String("ServerAddr", b.config.ServerAddr),
	//	zap.Int("Worker", b.config.Worker),
	//	zap.Bool("Debug", b.config.Debug))
}

func (b *Receiver) StartClientListening(Tls bool) {
	var (
		err      error
		listener net.Listener
		addr     = b.config.ServerAddr
	)

	for {
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			log.Error("Listen error", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}
		log.Info("start listening client", zap.String("addr", addr), zap.String("id", b.id))
		break // successfully listening
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Accept error", zap.Error(err), zap.String("addr", addr), zap.String("id", b.id))

			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				// 如果网络抖动则延时后再次Accept
				delay := time.Duration(1000) * time.Millisecond
				log.Info("Temporary Client Accept Error(%v), sleeping %vms", zap.Error(ne), zap.Duration("sleeping", delay))
				time.Sleep(delay)
			}
			continue
		}

		log.Info("Accept client", zap.Any("RemoteAddr", conn.RemoteAddr()), zap.String("id", b.id))
		go b.handleConnection(CLIENT, conn)
	}
}

func (b *Receiver) handleConnection(typ int, conn net.Conn) {
	//process connect packet
	packet, err := packets.ReadPacket(conn)
	if err != nil {
		log.Error("read connect packet error", zap.Error(err))
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

	// 响应CONNACK连接报文确认
	connack := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
	connack.SessionPresent = msg.CleanSession
	connack.ReturnCode = msg.Validate()

	if connack.ReturnCode != packets.Accepted {
		err = connack.Write(conn)
		if err != nil {
			log.Error("send connack error", zap.Error(err), zap.String("clientID", msg.ClientIdentifier))
			return
		}
		return
	}

	if typ == CLIENT && !b.CheckConnectAuth(string(msg.ClientIdentifier), string(msg.Username), string(msg.Password)) {
		connack.ReturnCode = packets.ErrRefusedNotAuthorised
		err = connack.Write(conn)
		if err != nil {
			log.Error("send connack error", zap.Error(err), zap.String("clientID", msg.ClientIdentifier))
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
	if b.wpool == nil {
		b.wpool = pool.New(b.config.Worker)
	}

	b.wpool.Submit(clientId, func() {
		ProcessMessage(msg)
	})

	//go ProcessMessage(msg)
}
