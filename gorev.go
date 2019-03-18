package gorev

import (
	"errors"
	"fmt"
)

type Params map[string]string

type Work func(p Params) error

type Task struct {
	Name       string
	NextTask   *Task
	PrevTask   *Task
	Forward    Work
	Backward   Work
}

func NewTask(name string, forward, backward Work) *Task {
	t := Task{}
	t.Name = name
	t.Forward = forward
	t.Backward = backward
	return &t
}

func (t *Task) Then(t2 *Task) *Task {
	first := t
	for t.NextTask != nil {
		t = t.NextTask
	}
	t.NextTask = t2
	t2.PrevTask = t
	return first
}

func (t *Task) handle(p Params) (err error) {
	rollback := p["error"] != ""
	if !rollback {
		err = t.Forward(p)
		if err != nil {
			rollback = true
			p["error"] = err.Error()
		}
		t.PrintStatus(false, err)
	}
	if rollback {
		rerr := t.Backward(p)
		t.PrintStatus(true, rerr)
	}
	return
}

func (t *Task) Exec(p Params) (err error) {
	err = t.handle(p)
	rollback := p["error"] != ""
	if rollback {
		if t.PrevTask != nil {
			return t.PrevTask.Exec(p)
		} else {
			return errors.New(p["error"])
		}
	} else if t.NextTask != nil {
		return t.NextTask.Exec(p)
	} 
	return
}

func (t *Task) PrintStatus(rollback bool, err error) {
	var msg string
	if rollback {
		msg = "ROLLBACK - "
	}
	if err == nil {
		msg += "DONE"
	} else {
		msg += "FAIL"
	}
	msg += " task '" + t.Name + "'."
	if err != nil {
		if !rollback {
			msg += "\nBEGINNING ROLLBACK"
		}
		msg += "\n\terror: " + err.Error()
	}
	fmt.Println(msg)
}

