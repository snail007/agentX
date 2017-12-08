package systemx

import (
	"agentX/utils"
	"fmt"
	"time"
)

type SystemX struct{}
type Command struct {
	Cmd     string `json:"cmd"`
	Async   bool   `json:"async"`
	Timeout int    `json:"timeout"`
	User    string `json:"user"`
}

func (x *SystemX) Passwd(out *string) (err error) {
	*out, _ = utils.FileGetContents("/etc/passwd")
	return nil
}
func (x *SystemX) Time(out *string) (err error) {
	*out = fmt.Sprintf("%d", time.Now().Unix())
	return nil
}
func (x *SystemX) Exec(command *Command, out *string) (err error) {
	*out = fmt.Sprintf("%d", time.Now().Unix())
	return nil
}
