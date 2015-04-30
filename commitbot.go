package main

import (
	"encoding/xml"
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// annoying xml parsing nonsense

type Commit struct {
	Revision string `xml:"revision,attr"`
	Author   string `xml:"author"`
	Date     string `xml:"date"`
	Msg      string `xml:"msg"`
}

type Log struct {
	Commits []Commit `xml:"logentry"`
}

type Entry struct {
	Revision string `xml:"revision,attr"`
}

type Info struct {
	Entry Entry `xml:"entry"`
}

type Path struct {
	Props string `xml:"props,attr"`
	Kind  string `xml:"kind,attr"`
	Item  string `xml:"item,attr"`
	Path  string `xml:",chardata"`
}

type Diff struct {
	Paths []Path `xml:"paths>path"`
}

func failiferr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func getPathChanges(svnroot string, commit Commit) (Diff, error) {
	buff, err := exec.Command("svn", "diff", "--summarize", "--xml", "-c", commit.Revision, svnroot).Output()
	if err != nil {
		return Diff{}, err
	}

	var v Diff
	err = xml.Unmarshal(buff, &v)
	return v, err
}

func getHead(svnroot string) int {
	buff, err := exec.Command("svn", "info", "--xml", svnroot).Output()
	failiferr(err)

	var v Info
	err = xml.Unmarshal(buff, &v)
	failiferr(err)

	r, e := strconv.Atoi(v.Entry.Revision)
	failiferr(e)

	return r
}

func getLogFromHead(svnroot string, head int) (Log, error) {
	buff, err := exec.Command("svn", "log", "-r", strconv.Itoa(head)+":HEAD", "--xml", svnroot).Output()
	if err != nil {
		return Log{}, err
	}

	var v Log
	err = xml.Unmarshal(buff, &v)
	return v, err
}

func parseHead(c Commit) (Commit, int) {
	h, e := strconv.Atoi(c.Revision)
	failiferr(e)

	return c, h
}

func recentCommits(root string, head int) Log {
	log, err := getLogFromHead(root, head+1)
	if err != nil {
		return Log{}
	}
	return log
}

// taken from rosetta code
func commonPrefix(diff Diff) string {
	sep := byte(os.PathSeparator)

	switch len(diff.Paths) {
	case 0:
		return ""
	case 1:
		return diff.Paths[0].Path
	}

	c := []byte(diff.Paths[0].Path)
	c = append(c, sep)

	for _, path := range diff.Paths[1:] {
		v := path.Path + string(sep)

		if len(v) < len(c) {
			c = c[:len(v)]
		}
		for i := 0; i < len(c); i++ {
			if v[i] != c[i] {
				c = c[:i]
				break
			}
		}
	}

	for i := len(c) - 1; i >= 0; i-- {
		if c[i] == sep {
			c = c[:i]
			break
		}
	}

	return string(c)
}

func run_irc(server, nick, owner string, channels []string, input chan string) {
	io := irc.IRC(nick, owner)
	io.VerboseCallbackHandler = true

	io.Connect(server)

	io.AddCallback("001", func(e *irc.Event) {
		for _, c := range channels {
			io.Join(c)
		}
	})

	go func() {
		for s := range input {
			for _, c := range channels {
				io.Privmsgf(c, "%s\n", s)
			}
		}
	}()

	io.Loop()
}

func formatCommit(c Commit) string {
	return fmt.Sprintf("r%s: <%s> %s", c.Revision, c.Author, c.Msg)
}

func withChangeRoot(original, svnroot string, c Commit) string {
	diff, err := getPathChanges(svnroot, c)
	if err == nil {
		root := commonPrefix(diff)
		return fmt.Sprintf("%s (%s)", original, root)
	}
	return original
}

func main() {
	if len(os.Args) != 6 {
		failiferr(fmt.Errorf("usage: %s <server:port> <nick> <username> <#channel1,#channel2> <svnroot>", os.Args[0]))
	}

	server := os.Args[1]
	nick := os.Args[2]
	owner := os.Args[3]
	channels := strings.Split(os.Args[4], ",")
	sr := os.Args[5]

	head := getHead(sr) - 1
	svnchan := make(chan string)

	go run_irc(server, nick, owner, channels, svnchan)

	for {
		log := recentCommits(sr, head)
		for _, c := range log.Commits {
			_, head = parseHead(c)
			svnchan <- withChangeRoot(formatCommit(c), sr, c)
		}
		time.Sleep(10 * time.Second)
	}
}
