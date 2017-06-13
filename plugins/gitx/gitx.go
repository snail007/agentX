package gitx

type String string

type Gitx struct{}

func (a *Gitx) Login(req *[]string, reply *[]string) error {
	*reply = *req
	return nil
}
