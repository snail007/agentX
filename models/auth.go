package models

import "net/http"

type String string

type Auth struct{}

func (a *Auth) Login(r *http.Request, req *[]string, reply *[]string) error {
	*reply = *req
	return nil
}
