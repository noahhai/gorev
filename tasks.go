package gorev

import (
	"errors"
	"fmt"
	"strings"
)

const (
	flag_status          = "internal_status"
	flag_error           = "internal_error"
	flag_status_rollback = "rollback"
	flag_status_exit     = "exit"
)

type Params map[string]interface{}

func (p Params) StringOrEmpty(key string) string {
	if v, ok := p[key]; ok && v != nil {
		return v.(string)
	}
	return ""
}

func (p Params) BoolOrDefault(key string, defaultVal bool) bool {
	if v, ok := p[key]; ok && v != nil {
		if vBool, okBool := v.(bool); okBool {
			return vBool
		}
	}
	return defaultVal
}

type Work func(p Params) error

var WorkPassthrough Work = func(p Params) error {
	return nil
}

type Task struct {
	Name      string
	Condition Condition
	NextTask  *Task
	PrevTask  *Task
	Forward   Work
	Backward  Work
}

func NewTask(name string, forward, backward Work) *Task {
	t := Task{}
	t.Name = name
	t.Forward = forward
	t.Backward = backward
	return &t
}

func (t *Task) WithCondition(c Condition) *Task {
	t.Condition = c
	return t
}

func (t *Task) last() *Task {
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
	exit := p[flag_status] == flag_status_exit
	if exit {
		return nil
	}
	if !rollback {
		fmt.Printf("START task '%s'\n", t.Name)

		err = ValidateParamConditions(p, t.Condition)
		if err != nil {
			fmt.Printf("VALIDATION FAIL:\n\tERROR: %v\n", err)
			p[flag_status] = flag_status_rollback
			p[flag_error] = err.Error()
			return
		}

		err = t.Forward(p)
		if err != nil {
			rollback = true
			p[flag_status] = flag_status_rollback
			p[flag_error] = err.Error()
		}
		t.PrintStatus(false, err)
	}
	if rollback {
		if !t.check("would you like to rollback") {
			p[flag_status] = flag_status_exit
			return
		}
		fmt.Printf("ROLLBACK - START task '%s'\n", t.Name)

		rerr := t.Backward(p)
		t.PrintStatus(true, rerr)
	}
	return
}

func (t *Task) check(msg string) bool {
	fmt.Printf("%s (y/n)? ", msg)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil || response == "" || len(response) < 1 {
		return false
	}
	return strings.ToLower(response)[0] == 'y'
}

func (t *Task) Rollback(p Params) (err error) {
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
		msg += "\n\tERROR: " + err.Error()
	}
	fmt.Println(msg)
}
