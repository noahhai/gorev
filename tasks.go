package gorev

import (
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
	subTasks []*Task
	resp      string
}

func Group(tasks ...*Task) *Task {
	t := Task{
		subTasks: tasks,
	}
	for _, st := range tasks {
		t.Name = t.Name + " " + st.Name
	}
	return &t
}

func NewTask(name string, forward, backward Work) *Task {
	t := Task{}
	t.Name = name
	t.Forward = forward
	t.Backward = backward
	return &t
}

func (t *Task) SetAutoResponse(resp string) *Task {
	t.resp = resp
	next := t.NextTask
	for next != nil {
		next.resp = resp
		next = next.NextTask
	}
	prev := t.PrevTask
	for prev != nil {
		prev.resp = resp
		prev = prev.NextTask
	}
	return t
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
	if t2.resp == "" && t.resp != "" {
		t2.resp = t.resp
	}
	return first
}

func (t *Task) handle(p Params) (err error) {
	isErr := p[flag_error] != nil
	rollback := p[flag_status] == flag_status_rollback
	exit := p[flag_status] == flag_status_exit
	if exit {
		if p[flag_error] != nil {
			return p[flag_error].(error)
		}
		return nil
	}
	if rollback {
		fmt.Printf("\nROLLBACK - START task '%s'\n", t.Name)
		var rerr error
		if len(t.subTasks) > 0 {
			for _, sub := range t.subTasks {
				if e := sub.Backward(p); e != nil {
					rerr = e
					break
				}
			}
		} else {
			rerr = t.Backward(p)
		}
		t.PrintStatus(true, rerr)
		return rerr
	}
	if isErr {
		if t.check("Would you like to rollback") {
			p[flag_status] = flag_status_rollback
		} else {
			p[flag_status] = flag_status_exit
		}
	}
	if !isErr {
		fmt.Printf("\nSTART task '%s'\n", t.Name)

		err = ValidateParamConditions(p, t.Condition)
		if err != nil {
			fmt.Printf("VALIDATION FAIL:\n\tERROR: %v\n", err)
			p[flag_error] = err
			return
		}
		if len(t.subTasks) > 0 {
			for _, sub := range t.subTasks {
				if e := sub.Forward(p); e != nil {
					err = e
					break
				}
			}
		} else {
			err = t.Forward(p)
		}
		if err != nil {
			p[flag_error] = err
		}
		t.PrintStatus(false, err)
	}
	if p[flag_error] != nil {
		return t.handle(p)
	}
	return
}

func (t *Task) check(msg string) bool {
	fmt.Printf("%s (y/n)? ", msg)
	var response string
	if t.resp != "" {
		fmt.Printf("auto-response set -> '%s'\n", t.resp)
		response = t.resp
	} else {
		_, err := fmt.Scanln(&response)
		t.SetAutoResponse(response)
		if err != nil || response == "" || len(response) < 1 {
			return false
		}
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
				return p[flag_error].(error)
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
