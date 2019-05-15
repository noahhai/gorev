package gorev

import (
	"errors"
	"fmt"
)

const (
	flag_status = "internal_status"
	flag_error = "internal_error"
	flag_status_rollback = "rollback"
)

type Params map[string]interface{}

type Work func(p Params) error

var WorkPassThrough Work = func(p Params)error{
	return nil
}

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

func (t *Task) last () *Task {
	for t.NextTask != nil {
		t = t.NextTask
	}
	return t
}

func (t *Task) Then(t2 *Task) *Task {
	first := t
	t = t.last()
	t.NextTask = t2
	t2.PrevTask = t
	return first
}

func (t *Task) handle(p Params) (err error) {
	rollback := p[flag_status] == flag_status_rollback
	if !rollback {
		fmt.Printf("START task '%s'\n", t.Name)
		err = t.Forward(p)
		if err != nil {
			rollback = true
			p[flag_status] = flag_status_rollback
			p[flag_error] = err.Error()
		}
		t.PrintStatus(false, err)
	}
	if rollback {
		fmt.Printf("ROLLBACK - START '%s'\n", t.Name)
		rerr := t.Backward(p)
		t.PrintStatus(true, rerr)
	}
	return
}

func(t *Task) Rollback(p Params)(err error) {
	p[flag_status] = flag_status_rollback
	return t.last().Exec(p)
}

func (t *Task) Exec(p Params) (err error) {
	err = t.handle(p)
	rollback := p[flag_status] == flag_status_rollback
	if rollback {
		if t.PrevTask != nil {
			return t.PrevTask.Exec(p)
		} else {
			if p[flag_error] == nil {
				return nil
			} else {
				return errors.New(p[flag_error].(string))
			}
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
	msg += " task '" + t.Name + "'"
	if err != nil {
		if !rollback {
			msg += "\nBEGINNING ROLLBACK"
		}
		msg += "\n\terror: " + err.Error()
	}
	fmt.Println(msg)
}

