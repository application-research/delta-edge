FROM golang:1.18-alpine
RUN apk add build-base
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -tags netgo -ldflags '-s -w' -o edge

EXPOSE 1313

CMD [ "./edge daemon --repo=.whypfs" ]