package receiver

import (
	"context"
	"errors"
	"net"
	"reflect"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

const (
	// CLIENT is an end user.
	CLIENT = 0
)

const (
	Connected    = 1
	Disconnected = 2
)

type client struct {
	typ        int
	mu         sync.Mutex
	receiver   *Receiver
	conn       net.Conn
	info       info
	status     int
	ctx        context.Context
	cancelFunc context.CancelFunc
	//session    *sessions.Session

	mutexHandler sync.Mutex
}

type info struct {
	clientID  string // 客户端的ID
	username  string // 验证的用户名
	password  []byte // 验证的密码
	keepalive uint16 // 是否保持连接
	localIP   string // 本地IP
	remoteIP  string // 远程连接IP
}

var (
	DisconnectedPacket = packets.NewControlPacket(packets.Disconnect).(*packets.DisconnectPacket)
	//r                  = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func (c *client) init() {
	c.status = Connected
	c.info.localIP, _, _ = net.SplitHostPort(c.conn.LocalAddr().String())
	remoteAddr := c.conn.RemoteAddr()
	//remoteNetwork := remoteAddr.Network()
	c.info.remoteIP = ""
	c.info.remoteIP, _, _ = net.SplitHostPort(remoteAddr.String())
	c.ctx, c.cancelFunc = context.WithCancel(context.Background())
}

func (c *client) Close() {
	if c.status == Disconnected {
		return
	}

	c.cancelFunc()

	c.status = Disconnected
	//wait for message complete
	time.Sleep(1 * time.Second)
	c.status = Disconnected

	b := c.receiver

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	if b != nil {
		b.removeClient(c)
	}
}

func (c *client) readLoop() {
	nc := c.conn
	b := c.receiver
	if nc == nil || b == nil {
		return
	}

	keepAlive := time.Second * time.Duration(c.info.keepalive)
	timeOut := keepAlive + (keepAlive / 2)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			//add read timeout
			if keepAlive > 0 {
				if err := nc.SetReadDeadline(time.Now().Add(timeOut)); err != nil {
					log.Error("set read timeout error", zap.Error(err), zap.String("ClientID", c.info.clientID))
					msg := &Message{
						client: c,
						packet: DisconnectedPacket,
					}
					b.SubmitWork(c.info.clientID, msg)
					return
				}
			}

			packet, err := packets.ReadPacket(nc)
			if err != nil {
				//log.Error("read packet error", zap.Error(err), zap.String("ClientID", c.info.clientID))
				msg := &Message{
					client: c,
					packet: DisconnectedPacket,
				}
				b.SubmitWork(c.info.clientID, msg)
				return
			}

			msg := &Message{
				client: c,
				packet: packet,
			}
			b.SubmitWork(c.info.clientID, msg)
		}
	}

}

func ProcessMessage(msg *Message) {
	client := msg.client
	ca := msg.packet
	if ca == nil {
		return
	}

	if client.typ == CLIENT {
		log.Debug("Debug Recv message", zap.String("message type", reflect.TypeOf(msg.packet).String()[9:]), zap.String("clientID", client.info.clientID))
	}

	switch ca.(type) {
	//case *packets.ConnackPacket:
	//case *packets.ConnectPacket:
	case *packets.PublishPacket:
		packet := ca.(*packets.PublishPacket)
		client.ProcessPublish(packet)

		// fix: https://github.com/yomorun/yomo-source-mqtt-starter/issues/1
		switch packet.Qos {
		case 1:
			puback := packets.NewControlPacket(packets.Puback).(*packets.PubackPacket)
			puback.MessageID = msg.packet.Details().MessageID
			if err := client.WriterPacket(puback); err != nil {
				log.Error("send response error", zap.Error(err), zap.String("ClientID", client.info.clientID))
				return
			}
		case 2:
			pubcomp := packets.NewControlPacket(packets.Pubcomp).(*packets.PubcompPacket)
			pubcomp.MessageID = msg.packet.Details().MessageID
			if err := client.WriterPacket(pubcomp); err != nil {
				log.Error("send pubcomp error, ", zap.Error(err), zap.String("ClientID", client.info.clientID))
				return
			}
		}

	case *packets.PubackPacket:
		puback := packets.NewControlPacket(packets.Puback).(*packets.PubackPacket)
		puback.MessageID = msg.packet.Details().MessageID
		if err := client.WriterPacket(puback); err != nil {
			log.Error("send puback error, ", zap.Error(err), zap.String("ClientID", client.info.clientID))
			return
		}
	case *packets.PubrecPacket:
		pubrec := packets.NewControlPacket(packets.Pubrec).(*packets.PubrecPacket)
		pubrec.MessageID = msg.packet.Details().MessageID
		if err := client.WriterPacket(pubrec); err != nil {
			log.Error("send pubrec error, ", zap.Error(err), zap.String("ClientID", client.info.clientID))
			return
		}
	case *packets.PubrelPacket:
		pubrel := packets.NewControlPacket(packets.Pubrel).(*packets.PubrelPacket)
		pubrel.MessageID = msg.packet.Details().MessageID
		if err := client.WriterPacket(pubrel); err != nil {
			log.Error("send pubrel error, ", zap.Error(err), zap.String("ClientID", client.info.clientID))
			return
		}
	case *packets.PubcompPacket:
		pubcomp := packets.NewControlPacket(packets.Pubcomp).(*packets.PubcompPacket)
		pubcomp.MessageID = msg.packet.Details().MessageID
		if err := client.WriterPacket(pubcomp); err != nil {
			log.Error("send pubcomp error, ", zap.Error(err), zap.String("ClientID", client.info.clientID))
			return
		}
	case *packets.PingreqPacket:
		client.ProcessPing()
	case *packets.DisconnectPacket:
		client.Close()
	default:
		log.Warn("Recv unknown message.......",
			zap.String("ClientID", client.info.clientID),
			zap.Any("TypeOfPacket", reflect.TypeOf(ca).Kind()))
	}

}

func (c *client) ProcessPublish(packet *packets.PublishPacket) {
	switch c.typ {
	case CLIENT:
		c.processClientPublish(packet)
	default:
		log.Warn("unknown client typ", zap.String("ClientID", c.info.clientID), zap.Int("typ", c.typ))
	}
}

func (c *client) processClientPublish(packet *packets.PublishPacket) {
	topic := packet.TopicName

	// handler message by user
	log.Debug("receive message from",
		zap.Uint16("MessageID", packet.MessageID),
		zap.String("clientID", c.info.clientID),
		zap.String("remoteIP", c.info.remoteIP),
		zap.String("username", c.info.username),
		zap.Any("topic", topic))

	// 错误重连
	c.mutexHandler.Lock()
	writer := c.receiver.sourceClient.GetWriter()
	err := c.receiver.handler(topic, packet.Payload, writer)
	if err != nil {
		if se, ok := err.(*SourceError); ok {
			log.Error("stream write error", zap.Error(se), zap.String("topic", topic))
			// 先关闭原来的流
			e := c.receiver.sourceClient.Close()
			if e != nil {
				log.Error("sourceClient Close", zap.Error(e), zap.String("topic", topic))
			}
			// 重新连接
			c.receiver.sourceClient.Retry()
		} else {
			log.Error("handler error", zap.Error(err), zap.String("topic", topic))
		}
	}
	c.mutexHandler.Unlock()
}

func (c *client) ProcessPing() {
	if c.status == Disconnected {
		return
	}
	resp := packets.NewControlPacket(packets.Pingresp).(*packets.PingrespPacket)
	err := c.WriterPacket(resp)
	if err != nil {
		log.Error("send PingResponse error, ", zap.Error(err), zap.String("ClientID", c.info.clientID))
		return
	}
}

func (c *client) WriterPacket(packet packets.ControlPacket) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error("recover error, ", zap.Any("recover", err))
		}
	}()
	if c.status == Disconnected {
		return nil
	}

	if packet == nil {
		return nil
	}
	if c.conn == nil {
		c.Close()
		return errors.New("connect lost")
	}

	c.mu.Lock()
	err := packet.Write(c.conn)
	c.mu.Unlock()
	return err
}
