FROM ubuntu

COPY --from=golang:1.18-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

RUN apt-get -y update \
  && apt-get -y install nginx

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -tags netgo -ldflags '-s -w' -o edge-cli

EXPOSE 1313

CMD [ "/edge-cli daemon --repo=.whypfs" ]