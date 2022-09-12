FROM golang:1.18-alpine

LABEL "rabitechs"="rabitechs"

WORKDIR /app

COPY . .

RUN go mod download

RUN apk update && apk add curl \
    git \
    protobuf \
    bash \
    make \
    openssh-client && \
    rm -rf /var/cache/apk/*


RUN go get github.com/Masterminds/glide

RUN go mod vendor

RUN go build -o=./auth /app/cmd/api

EXPOSE 4002 4002

CMD [ "./auth" ]