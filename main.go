package main

import (
	"agentX/models"
	_ "agentX/routers"
	"net/http"

	"fmt"

	"os"

	"github.com/astaxie/beego"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
)

type Hook struct {
	f http.Handler
}

func (h *Hook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hook called")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
	if r.Method == "OPTIONS" {
		return
	}
	h.f.ServeHTTP(w, r)
}

func main() {
	fmt.Println(poster())
	err := initConfig()
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(0)
	}
	initLog()
	log.Info("agentX service stared")
	s := rpc.NewServer()
	s.RegisterCodec(json2.NewCodec(), "application/json")
	s.RegisterService(new(models.Auth), "")
	h := new(Hook)
	h.f = s
	beego.Handler("/", h)
	beego.Run()
	select {}
}
