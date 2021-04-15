package receiver

import (
	"context"
	"io"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/yomorun/yomo/pkg/quic"
)

type ISourceWriter interface {
	Write(b []byte) (int, error)
}

type sourceWriter struct {
	w io.Writer
}

func (s *sourceWriter) Write(b []byte) (int, error) {
	n, e := s.w.Write(b)
	if e != nil {
		// 通过代理方式，隐性注入自定义错误类型，用于错误重试的逻辑
		return n, NewSourceError(e.Error())
	}
	return n, e
}

func NewSourceWriter(w io.Writer) ISourceWriter {
	return &sourceWriter{w: w}
}

type ISourceStream interface {
	GetWriter() ISourceWriter
	Create() quic.Stream
	Close() error
	Init()
}

type sourceStreamImpl struct {
	zipperAddr string

	client      quic.Client
	stream      quic.Stream
	mutexStream sync.Mutex

	streamErrorMax      int
	streamErrorInterval int
}

func NewSourceStream(zipperAddr string) ISourceStream {
	if len(zipperAddr) == 0 {
		panic("please provider addr of zipper")
	}

	return &sourceStreamImpl{
		zipperAddr:          zipperAddr,
		streamErrorMax:      10,
		streamErrorInterval: 1000,
	}
}

func (s *sourceStreamImpl) Init() {
	if s.stream == nil {
		s.stream = s.Create()
	}
}

func (s *sourceStreamImpl) Close() error {
	return s.stream.Close()
}

func (s *sourceStreamImpl) GetWriter() ISourceWriter {
	if s.stream == nil {
		s.stream = s.Create()
		return NewSourceWriter(s.stream)
	}

	return NewSourceWriter(s.stream)
}

func (s *sourceStreamImpl) Create() quic.Stream {
	var (
		err  error
		errs = 0
	)

	s.mutexStream.Lock()
	defer s.mutexStream.Unlock()

CLIENT:
	if s.client == nil {
		for {
			s.client, err = quic.NewClient(s.zipperAddr)
			if err != nil {
				log.Error("NewClient error", zap.String("zipperAddr", s.zipperAddr), zap.Error(err))
				time.Sleep(time.Duration(s.streamErrorInterval) * time.Millisecond)
				continue
			}
			log.Info("connect to zipper", zap.String("zipperAddr", s.zipperAddr))
			break
		}
	}

	for {
		s.stream, err = s.client.CreateStream(context.Background())
		if err != nil {
			log.Error("CreateStream error", zap.Error(err))
			errs++
			if errs > s.streamErrorMax {
				// if greater than the number of errors, a new connection is established
				s.client = nil
				goto CLIENT
			}
			time.Sleep(time.Duration(s.streamErrorInterval) * time.Millisecond)
			continue
		}
		log.Info("create stream", zap.String("zipperAddr", s.zipperAddr), zap.Any("StreamID", s.stream.StreamID()))
		break
	}

	return s.stream
}
