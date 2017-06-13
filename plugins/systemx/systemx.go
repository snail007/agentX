package systemx

import (
	"agentX/utils"
	"fmt"
	"time"
)

type SystemX struct{}

func (x *SystemX) Passwd(out *string) (err error) {
	*out, _ = utils.FileGetContents("/etc/passwd")
	return nil
}
func (x *SystemX) Time(out *string) (err error) {
	*out = fmt.Sprintf("%d", time.Now().Unix())
	return nil
}
