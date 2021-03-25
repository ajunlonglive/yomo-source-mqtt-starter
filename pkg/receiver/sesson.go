package receiver

import (
	"github.com/eclipse/paho.mqtt.golang/packets"
	"go.uber.org/zap"
)

func (b *Receiver) getSession(cli *client, req *packets.ConnectPacket, resp *packets.ConnackPacket) error {
	var err error

	if len(req.ClientIdentifier) == 0 {
		req.CleanSession = true
	}

	cid := req.ClientIdentifier

	if !req.CleanSession {
		if cli.session, err = b.sessionMgr.Get(cid); err == nil {
			resp.SessionPresent = true

			if err := cli.session.Update(req); err != nil {
				return err
			}
		}
	}

	if cli.session == nil {
		log.Debug("#10 [getSession] ", zap.String("cid", cid))
		if cli.session, err = b.sessionMgr.New(cid); err != nil {
			return err
		}

		resp.SessionPresent = false

		if err := cli.session.Init(req); err != nil {
			return err
		}
	}

	return nil
}
