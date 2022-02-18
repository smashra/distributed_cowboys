package main

import (
	"log"
  "os"
  "fmt"
	"github.com/nats-io/nats.go"
)


type MType int32

const (
  SHOT MType = iota
  START
  HEALTH
)

type Msg struct {
  MsgType MType
  Cowboy
}

type Cowboy struct {
  Name string
  Health int32
}


type Shootout struct {
  Cowboy
  Damage int32
}


var TOPIC = os.Getenv("TOPIC") 
var NATSURL string 
var NATSEP =  os.Getenv("NATSEP")
var CLIENTPORT =  os.Getenv("CLIENTPORT")


func init() {

  if TOPIC == "" {
    TOPIC = "target"
  }
  
  if NATSEP != "" && CLIENTPORT != "" {
    NATSURL = "nats://" + NATSEP + ":" + CLIENTPORT
  } else {
    NATSURL = "demo.nats.io"
  }
}

func main() {
  
  fmt.Println("Commence shooting .....")

	nc, err := nats.Connect(NATSURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}
	defer ec.Close()

  fmt.Println(NATSURL)

	// Publish the message
	if err := ec.Publish(TOPIC, &Msg{MsgType: START,}); err != nil {
		log.Fatal(err)
	}
}




