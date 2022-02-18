FROM golang:latest

WORKDIR /app

COPY go.mod ./

COPY go.sum ./

RUN go mod download

COPY src/castai.go ./

RUN go build castai.go

ENTRYPOINT ["./castai"]

