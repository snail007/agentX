package gitx

import "net/http"

type String string

type Gitx struct{}

func (a *Gitx) Login(r *http.Request, req *[]string, reply *[]string) error {
	*reply = *req
	return nil
}
