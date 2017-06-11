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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
	if r.Method == "OPTIONS" {
		return
	}
	user, pass, ok := r.BasicAuth()
	if ok {
		log.Infof("user:%s,pass:%s", user, pass)
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
	go beego.Run()
	go func() {
		http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./clients/js"))))
		http.ListenAndServe(":25900", nil)
	}()
	select {}
}
