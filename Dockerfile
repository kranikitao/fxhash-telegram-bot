FROM golang:1.18.3 as builder
WORKDIR /go/src/github.com/kranikitao/fxhash-telegram-bot

COPY ./ /go/src/github.com/kranikitao/fxhash-telegram-bot

RUN make

FROM debian:11.3

RUN apt-get update && apt-get install -y wget
RUN mkdir -p /usr/local/share/ca-certificates/cacert.org
RUN wget -P /usr/local/share/ca-certificates/cacert.org http://www.cacert.org/certs/root.crt http://www.cacert.org/certs/class3.crt
RUN update-ca-certificates

RUN adduser --system --disabled-password --home /app appuser
USER appuser
WORKDIR /app/

COPY --from=builder /go/src/github.com/kranikitao/fxhash-telegram-bot/src/migrations /app/src/migrations
COPY --from=builder /go/src/github.com/kranikitao/fxhash-telegram-bot/botrunner /app/botrunner

CMD ["./botrunner"]
