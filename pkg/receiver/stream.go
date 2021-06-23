package receiver

import (
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	cl "github.com/yomorun/yomo/pkg/client"
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

type ISourceClient interface {
	GetWriter() ISourceWriter
	Create() cl.SourceClient
	Close() error
	Retry()
	Init()
}

type sourceClientImpl struct {
	zipperAddr string
	appName    string
	cli        cl.SourceClient

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
	s.cli.Retry()
}

func (s *sourceClientImpl) GetWriter() ISourceWriter {
	if s.cli == nil {
		s.cli = s.Create()
		return NewSourceWriter(s.cli)
	}

	return NewSourceWriter(s.cli)
}

func (s *sourceClientImpl) Create() cl.SourceClient {
	var err error

	s.mutexCli.Lock()
	defer s.mutexCli.Unlock()

	if s.cli == nil {
		for {
			ip, port := s.splitZipperAddr()
			s.cli, err = cl.NewSource(s.appName).Connect(ip, port)
			if err != nil {
				log.Error("NewClient error", zap.String("zipperAddr", s.zipperAddr), zap.Error(err))
				time.Sleep(time.Duration(s.clientErrorInterval) * time.Millisecond)
				continue
			}
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
