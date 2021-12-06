package receiver

import (
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/yomorun/yomo"
)

const DataTag uint8 = 0x33

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

type ISourceClient interface {
	GetWriter() ISourceWriter
	Create() yomo.Source
	Close() error
	Retry()
	Init()
}

type sourceClientImpl struct {
	zipperAddr string
	appName    string
	cli        yomo.Source

	mutexCli sync.Mutex

	clientErrorInterval int
}

func NewSourceStream(appName string, zipperAddr string) ISourceClient {
	if len(zipperAddr) == 0 {
		panic("please provider addr of zipper")
	}

	return &sourceClientImpl{
		zipperAddr:          zipperAddr,
		appName:             appName,
		clientErrorInterval: 1000,
	}
}

func (s *sourceClientImpl) Init() {
	if s.cli == nil {
		s.cli = s.Create()
	}
}

func (s *sourceClientImpl) Close() error {
	return s.cli.Close()
}

func (s *sourceClientImpl) Retry() {
	// s.cli.Retry()
}

func (s *sourceClientImpl) GetWriter() ISourceWriter {
	if s.cli == nil {
		s.cli = s.Create()
		return NewSourceWriter(s.cli)
	}

	return NewSourceWriter(s.cli)
}

func (s *sourceClientImpl) Create() yomo.Source {
	var err error

	s.mutexCli.Lock()
	defer s.mutexCli.Unlock()

	if s.cli == nil {
		for {
			source := yomo.NewSource(
				"yomo-source",
				yomo.WithZipperAddr(s.zipperAddr),
			)
			err = source.Connect()
			if err != nil {
				log.Error("[source] ❌ Emit the data to YoMo-Zipper failure with err: %v", zap.Error(err))
				time.Sleep(time.Duration(s.clientErrorInterval) * time.Millisecond)
				continue
			}
			source.SetDataTag(DataTag)
			s.cli = source
			log.Info("connect to zipper", zap.String("zipperAddr", s.zipperAddr))
			break
		}
	}

	return s.cli
}

func (s *sourceClientImpl) splitZipperAddr() (ip string, port int) {
	ss := strings.Split(s.zipperAddr, ":")
	ip = ss[0]
	port, _ = strconv.Atoi(ss[1])
	return
}
