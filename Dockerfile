FROM golang:1.4-wheezy

RUN apt-get update -y && apt-get install -y subversion git

WORKDIR /commitbot
ENV GOPATH /commitbot/lib

RUN go get github.com/thoj/go-ircevent

ADD commitbot.go commitbot.go

RUN go build commitbot.go
RUN cp commitbot /usr/bin/commitbot

ENTRYPOINT ["commitbot"]
CMD ["irc:6667", "commitbot", "rbruns", "#commits", "svn://nebula"]
