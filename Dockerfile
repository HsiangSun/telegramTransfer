FROM golang:alpine

RUN apk add  sqlite-libs sqlite-dev
RUN apk add  build-base

WORKDIR /app
ADD ./ /app

RUN go build -o /app/transfer main.go

#ADD /app/cacert.pem /etc/ssl/certs/

ADD ./config.toml /app

#CMD ["ping www.google.com"]
ENTRYPOINT ["/app/transfer"]
