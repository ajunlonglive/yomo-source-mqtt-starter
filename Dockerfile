FROM golang:buster

RUN apt-get update && \
    apt-get install nano iputils-ping telnet net-tools ifstat -y

RUN cp  /usr/share/zoneinfo/Asia/Shanghai /etc/localtime  && \
    echo 'Asia/Shanghai'  > /etc/timezone

WORKDIR $GOPATH/src/source
RUN go mod init source && go get github.com/yomorun/yomo-source-mqtt-starter
RUN go get -d -v ./...

EXPOSE 1883
