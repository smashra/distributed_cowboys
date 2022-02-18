package main

import (
  "fmt"
  "time"
  "log"
  "os"
  "math/rand"
  "strconv"
  "sync/atomic"
  "encoding/json"
//  "io/ioutil"
  "github.com/nats-io/nats.go"
   "knative.dev/pkg/configmap"
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

func (s *Shootout) shoot() int32 {
    return atomic.AddInt32(&s.Health, -(s.Damage))
}

func (s *Shootout) health() int32 {
    return atomic.LoadInt32(&s.Health)
}

func (s *Shootout) updateHealth(health int32) {
    atomic.StoreInt32(&s.Health, health)
}

func shoot(start, stop chan bool, m []*Shootout) {
 
  <-start
  var others int32
  min, max := 1, len(m)
  
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

  for {
    time.Sleep(1 * time.Second)
    
    if m[0].health() <= 0 {
      fmt.Println(ID + " is dead.")
      break
    }

    for i := 1; i < len(m); i++ {
      if m[i].health() <= 0 {
        others++        
      }
    }  
    
    if others == int32(len(m) - 1) {
      fmt.Println(ID + " is the winner.")
      break 
    }

    others = 0

    index := rand.Intn(max - min) + min

    if m[index].health() > 0  {
      if err := ec.Publish(TOPIC,  &Msg{MsgType: SHOT, Cowboy: Cowboy{Name: m[index].Name,}}); err != nil {
		    log.Fatal(err)
	    }
      fmt.Println(ID + " shoots " + m[index].Name + " ...")
    }
  } 
  
  stop<- true
}

var ID = os.Getenv("ID") 
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

  rand.Seed(time.Now().UnixNano())
}

func main() {

  fmt.Println(ID + " is getting ready for the shootout " + NATSURL)
  start := make(chan bool) 
  stop := make(chan bool) 

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

  cowboys, shooters := loadConfig()

  go shoot(start, stop, shooters)
  
  //wait for start from mq; subscribe
  sub, err := ec.Subscribe(TOPIC, func(s *Msg) {
    switch s.MsgType {
      case START:
        start<-true
        fmt.Println(ID + " starts shooting ....")
        break
      case SHOT:
        if s.Name == ID {
          //fmt.Println(cowboys[ID])
          h := cowboys[ID].health()
          if h > 0 {
            cowboys[ID].shoot()
            h := cowboys[ID].health()
            //publish your stats
            if err := ec.Publish(TOPIC,  &Msg{MsgType: HEALTH, Cowboy: Cowboy{Name: ID, Health: h}}); err != nil {
              log.Fatal(err)
            }
            fmt.Println(ID + " got shot, health [" +  strconv.FormatInt(int64(h), 10) + "]" )
          }

        }
        break
      case HEALTH:
        if s.Name != ID {
          //fmt.Println(cowboys[s.Name])
          cowboys[s.Name].updateHealth(s.Health)
          //fmt.Println(cowboys[s.Name])
        }
    }
  })
  
  if err != nil {
    log.Fatal(err)
  } 

  <-stop
  if err := sub.Unsubscribe(); err != nil {
    log.Fatal(err)
  }
}

func loadConfig( ) (map[string]*Shootout, []*Shootout) {

  cowboys := make(map[string]*Shootout)

  type shooter struct {
    Name string
    Health int32
    Damage int32
  }

  var data []shooter
  //healthPlan, _ := ioutil.ReadFile("./shooters.json")
  cfg, err := configmap.Load("/config")
  if err != nil {
    log.Fatal("Unable to load config map!")
  }
  
  for _, v := range cfg {
    err := json.Unmarshal([]byte(v), &data)
    if err != nil {
        log.Fatal("Cannot unmarshal the json ", err)
    }
  }
  
//  err := json.Unmarshal(healthPlan, &data)
//  if err != nil {
//        log.Fatal("Cannot unmarshal the json ", err)
//  }
  
  for _, v := range data {
    cowboys[v.Name] =  &Shootout {
      Cowboy: Cowboy { Name: v.Name,
        Health: v.Health,
      },
      Damage: v.Damage,
    }
  }

  shooterSlice := make([]*Shootout, len(cowboys))
  shooterSlice[0] = cowboys[ID]
  k := 1 

  for key, val := range cowboys {
    if key != ID {
      shooterSlice[k] = val
      k++      
    }
  }  

  return cowboys, shooterSlice
}






