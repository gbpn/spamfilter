package main

import (
	"container/ring"
	"fmt"
	"time"
)

//Log is an in memory ring buffer log
type Log struct {
	ring *ring.Ring
}

func (l *Log) init(size int) {
	l.ring = ring.New(size)

	for i := 0; i < size; i++ {
		l.ring.Value = ""
		l.ring = l.ring.Next()
	}
}

func (l *Log) record(severity string, msg string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	save := fmt.Sprintf("%s %s %s", now, severity, msg)
	l.ring = l.ring.Prev()
	l.ring.Value = save
}

func (l *Log) notice(msg string) {
	l.record("NOTICE", msg)
}

func (l *Log) warn(msg string) {
	l.record("WARN", msg)
}

func (l *Log) fatal(msg string) {
	l.record("FATAL", msg)
}


func (l *Log) tostdout() {
	l.ring.Do(func(p any) {
		if len(p.(string)) == 0 {
			return
		}

		fmt.Println(p)
	})
}

func (l *Log) toarray() ([]string) {
	var res []string
	l.ring.Do(func(p any) {
		res = append(res, p.(string))
	})
	return res
}

