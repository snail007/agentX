package gitx

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"

	"fmt"
	"path/filepath"

	"strings"

	"os"
	"os/exec"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type String string

type Gitx struct{}

type URL struct {
	URL        string `json:"url"`
	SSHKEY     string `json:"sshkey"`
	SSHKEYSALT string `json:"sshkeysalt"`
	PATH       string `json:"path"`
	BRANCH     string `json:"branch"`
	USER       string `json:"user"`
	PASSWORD   string `json:"password"`
}

func checkout(name string, url URL) (output string, err error) {
	cmd := exec.Command("git", "checkout", "-f", name)
	cmd.Dir = url.PATH
	var outputB []byte
	outputB, err = cmd.CombinedOutput()
	if err == nil {
		output = string(outputB)
	}
	return
}
func cleanBranch(url URL) (err error) {
	//var hash string
	//var hashObj plumbing.Hash
	var r *git.Repository
	//var ref *plumbing.Reference
	if r, err = git.PlainOpen(url.PATH); err != nil {
		return
	}
	refs, err := r.References()
	if err != nil {
		return
	}
	h, err := r.Head()
	if err != nil {
		return
	}
	headBranchName := h.Name().String()
	refs.ForEach(func(ref0 *plumbing.Reference) (e error) {
		//fmt.Println(ref0.Name().String(), ref0.Hash().String())
		if headBranchName != ref0.Name().String() && strings.HasPrefix(ref0.Name().String(), "refs/heads/") {
			err = r.Storer.RemoveReference(ref0.Name())
			if err != nil {
				e = err
				return
			}
		}
		return nil
	})
	return
}

var branchNamePrefix = "refs/heads/auto-"

func createBranchName(url URL) (name string, err error) {
	// var hash string
	// var hashObj plumbing.Hash
	var ref *plumbing.Reference
	_, _, _, ref, err = getHash(url)
	if err != nil {
		return
	}
	f := filepath.Base(ref.Name().Short())
	name = branchNamePrefix + f + "-" + newLenChars(8, stdChars)
	return
}

var stdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func newLenChars(length int, chars []byte) string {
	if length == 0 {
		return ""
	}
	clen := len(chars)
	if clen < 2 || clen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}
	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			panic("Error reading random bytes: " + err.Error())
		}
		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				continue // Skip this number to avoid modulo bias.
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}
func (x *Gitx) Publish(in *URL, out *string) (err error) {
	url := *in
	if isEmpty(url.PATH) {
		err = nil
		//clone
		_, err = clone(url)
	} else {
		p := filepath.Join(url.PATH, ".git")
		if _, err := os.Stat(p); err != nil || isEmpty(p) {
			return errors.New("target path is not a git repository")
		}
		//if target is not a repository,git.PlainOpen
		//will create two folders (objects,refs) in target path
		objectsBeforeIsExists := false
		refsBeforeIsExists := false
		p1 := filepath.Join(p, "objects")
		p2 := filepath.Join(p, "refs")
		if _, err := os.Stat(p1); err == nil {
			objectsBeforeIsExists = true
		}
		if _, err := os.Stat(p2); err == nil {
			refsBeforeIsExists = true
		}
		_, err = git.PlainOpen(url.PATH)
		if _, err := os.Stat(p1); err == nil && !objectsBeforeIsExists {
			os.RemoveAll(p1)
		}
		if _, err := os.Stat(p2); err == nil && !refsBeforeIsExists {
			os.RemoveAll(p2)
		}
		//clean objects,refs folder if necessary
		if err == nil {
			//fetch
			_, err = fetch(url)
		}
	}
	if err != nil {
		return
	}
	branchShortName := ""
	branchShortName, _, err = createBranch(url)
	if err != nil {
		return
	}
	//output := ""
	_, err = checkout(branchShortName, url)
	if err != nil {
		return
	}
	err = cleanBranch(url)
	return
}
func createBranch(url URL) (branchShortName, branchName string, err error) {
	_, hashObj, r, _, err := getHash(url)
	if err != nil {
		return
	}
	branchName, err = createBranchName(url)
	if err != nil {
		return
	}
	branchShortName = filepath.Base(branchName)
	err = r.Storer.SetReference(plumbing.NewHashReference(plumbing.ReferenceName(branchName), hashObj))
	return
}
func headHash(url URL) (hash, name string, hashObj plumbing.Hash, err error) {
	r, err := git.PlainOpen(url.PATH)
	if err != nil {
		return
	}
	h, err := r.Head()
	if err != nil {
		return
	}
	hash = h.Hash().String()
	name = h.Name().Short()
	hashObj = h.Hash()
	return
}
func getHash(url URL) (hash string, hashObj plumbing.Hash, r *git.Repository, ref *plumbing.Reference, err error) {
	if r, err = git.PlainOpen(url.PATH); err != nil {
		return
	}
	refs, err := r.References()
	if err != nil {
		return
	}
	refs.ForEach(func(ref0 *plumbing.Reference) (e error) {
		//fmt.Println(ref0.Name().String(), ref0.Hash().String())
		if !strings.HasPrefix(ref0.Name().String(), "refs/heads/") {
			if url.BRANCH == filepath.Base(ref0.Name().String()) {
				hash = ref0.Hash().String()
				ref = ref0
				hashObj = ref0.Hash()
				return fmt.Errorf("")
			}
		}
		return nil
	})
	if hash == "" {
		err = fmt.Errorf("reference not found")
	}
	return
}
func fetch(url URL) (r *git.Repository, err error) {
	if err = validate(url); err != nil {
		err = fmt.Errorf("config error : %s", err)
		return
	}
	if r, err = git.PlainOpen(url.PATH); err != nil {
		return
	}
	var opt git.FetchOptions
	if opt, err = fetchOptions(url); err != nil {
		return
	}
	if err = r.Fetch(&opt); err == git.NoErrAlreadyUpToDate {
		err = nil
	}
	return
}
func clone(url URL) (r *git.Repository, err error) {
	if err = validate(url); err != nil {
		err = fmt.Errorf("config error : %s", err)
		return
	}
	var opt git.CloneOptions
	opt, err = cloneOptions(url)
	if err != nil {
		return
	}
	r, err = git.PlainClone(url.PATH, false, &opt)
	return
}
func cloneOptions(url URL) (opt git.CloneOptions, err error) {
	opt.URL = url.URL
	if isNeedAuth(url) {
		opt.Auth, err = getAuth(url)
		if err != nil {
			return
		}
	}
	return
}
func fetchOptions(url URL) (opt git.FetchOptions, err error) {
	opt.RemoteName = git.DefaultRemoteName
	opt.RefSpecs = []config.RefSpec{"+refs/heads/*:refs/remotes/" + git.DefaultRemoteName + "/*"}
	if isNeedAuth(url) {
		opt.Auth, err = getAuth(url)
		if err != nil {
			return
		}
	}
	return
}
func getAuth(url URL) (auth transport.AuthMethod, err error) {
	var signer ssh.Signer
	if isHTTP(url) {
		customClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 60 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		client.InstallProtocol("https", githttp.NewClient(customClient))
		client.InstallProtocol("http", githttp.NewClient(customClient))
		auth = githttp.NewBasicAuth(url.USER, url.PASSWORD)
	} else {
		if url.SSHKEYSALT != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(url.SSHKEY), []byte(url.SSHKEYSALT))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(url.SSHKEY))
		}
		if err != nil {
			return
		}
		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}
	return
}

func validate(url URL) (err error) {
	if url.PATH == "" {
		err = fmt.Errorf("path requied")
		return
	}
	if url.URL == "" {
		err = fmt.Errorf("url requied")
		return
	}
	if isHTTP(url) {
		if (url.USER != "" && url.PASSWORD == "") ||
			(url.USER == "" && url.PASSWORD != "") {
			err = fmt.Errorf("user and password requied")
			return
		}
	} else if url.SSHKEY == "" {
		err = fmt.Errorf("SSHKEY requied")
		return
	}
	return
}
func isHTTP(url URL) bool {
	return strings.HasPrefix(url.URL, "http://") || strings.HasPrefix(url.URL, "https://")
}
func isNeedAuth(url URL) bool {
	if isHTTP(url) {
		return url.USER != "" && url.PASSWORD != ""
	}
	return url.SSHKEY != ""
}
func isEmpty(path string) bool {
	fs, e := filepath.Glob(filepath.Join(path, "*"))
	if e != nil {
		return false
	}
	if len(fs) > 0 {
		return false
	}
	return true
}
