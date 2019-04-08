FROM golang:alpine
WORKDIR /app
RUN apk add git
COPY . /app
RUN go build -o /mcapi
ENTRYPOINT [ "/mcapi", "-config", "/config.json" ]
