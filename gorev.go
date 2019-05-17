package gorev

import (
	"errors"
	"fmt"
	"strings"
)

const (
	flag_status = "internal_status"
	flag_error = "internal_error"
	flag_status_rollback = "rollback"
)

type Params map[string]interface{}

type Work func(p Params) error

type Condition struct {
	And   []Condition
	Or    []Condition
	Xor   []Condition
	Key   string
	Value interface{}
}

var WorkPassthrough Work = func(p Params)error{
	return nil
}

type Task struct {
	Name       string
	Condition
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
		err = ValidateParamConditions(p, t.Condition)
		if err == nil {
			err = t.Forward(p)
		}
		if err != nil {
			rollback = true
			p[flag_status] = flag_status_rollback
			p[flag_error] = err.Error()
		}
		t.PrintStatus(false, err)
	}
	if rollback {
		fmt.Printf("ROLLBACK - START task '%s'\n", t.Name)
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
		msg += "\n\tERROR: " + err.Error()
	}
	fmt.Println(msg)
}


func ValidateParamConditions(params map[string]interface{}, condition Condition) error {
	if condition.Key != "" {
		if v, ok := params[condition.Key]; !ok || v == nil {
			return fmt.Errorf("missing param '%s'", condition.Key)
		} else if condition.Value != nil && condition.Value != v {
			return fmt.Errorf("param '%s' did not match expected '%v' (%v)", condition.Key, condition.Value, v)
		}
	}

	// And
	for _, c := range condition.And {
		if err := ValidateParamConditions(params, c); err != nil {
			return err
		}
	}

	// Xor
	var found []string
	var missing []string
	for _, c := range condition.Xor {
		if err := ValidateParamConditions(params, c); err != nil {
			missing = append(missing, c.Key)
		} else {
			found = append(found, c.Key)
		}
		if len(found) > 1 {
			return fmt.Errorf("2+ XOR param(s): %s", strings.Join(found, ", "))
		}
	}
	if len(condition.Xor) > 1 && len(found) < 1 {
		return fmt.Errorf("invalid/missing XOR param(s): %s", strings.Join(missing, ", "))
	}

	// Or
	missing = []string{}
	for _, c := range condition.Or {
		if err := ValidateParamConditions(params, c); err == nil {
			return nil
		}
		missing = append(missing, c.Key)
	}
	if len(condition.Or) > 1 {
		return fmt.Errorf("invalid/missing OR param(s): %s", strings.Join(missing, ", "))
	}
	return nil
}
