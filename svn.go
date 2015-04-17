package main

import (
	"encoding/xml"
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"os/exec"
	"strconv"
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

func failiferr(e error) {
	if e != nil {
		log.Fatal(e)
	}
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

// try to do something sensible now

func parseHead(c Commit) (Commit, int) {
	h, e := strconv.Atoi(c.Revision)
	failiferr(e)

	return c, h
}

func formatCommit(c Commit) string {
	return fmt.Sprintf("%s: <%s> %s", c.Revision, c.Author, c.Msg)
}

func recentCommits(root string, head int) Log {
	log, err := getLogFromHead(root, head+1)
	if err != nil {
		return Log{}
	}
	return log
}

func run_irc(server, nick string, channels []string, input chan string) {
	io := irc.IRC(nick, "rbruns")
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

func main() {
	sr := "svn://nebula/"
	head := getHead(sr) - 1
	svnchan := make(chan string)

	go run_irc("irc:6667", "commits", []string{"#commits"}, svnchan)

	for {
		log := recentCommits(sr, head)
		for _, c := range log.Commits {
			_, head = parseHead(c)
			svnchan <- formatCommit(c)
		}
		time.Sleep(10 * time.Second)
	}
}
