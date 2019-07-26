package gorev

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExec(t *testing.T) {
	task1 := NewTask("task1", Work(func(p Params) error { fmt.Println("doing work from task 1"); return nil }), Work(func(p Params) error { fmt.Println("undoing work from task 1"); return nil }))
	task2 := NewTask("task2", Work(func(p Params) error { fmt.Println("doing work from task 2"); return nil }), Work(func(p Params) error { fmt.Println("undoing work from task 2"); return nil }))
	task3 := NewTask("task3", Work(func(p Params) error { fmt.Println("doing work from task 3"); return nil; }), Work(func(p Params) error { fmt.Println("undoing work from task 3"); return nil }))
	task4 := NewTask("task4", Work(func(p Params) error { fmt.Println("doing work from task 4"); return nil; }), Work(func(p Params) error { fmt.Println("undoing work from task 4"); return nil }))
	task5 := NewTask("task5", Work(func(p Params) error { fmt.Println("doing work from task 5"); return errors.New("some error!") }), Work(func(p Params) error { fmt.Println("undoing work from task 5"); return nil }))

	p := Params{}
	err := task1.SetAutoResponse("y").Then(task2).Then(task3).Then(task4).Then(task5).Exec(p)
	if err == nil {
		t.Error("unexpected: error was nil")
	}
}

func TestExecSub(t *testing.T) {
	task1 := NewTask("task1", Work(func(p Params) error { fmt.Println("doing work from task 1"); return nil }), Work(func(p Params) error { fmt.Println("undoing work from task 1"); return nil }))
	task2 := NewTask("task2", Work(func(p Params) error { fmt.Println("doing work from task 2"); return nil }), Work(func(p Params) error { fmt.Println("undoing work from task 2"); return nil }))
	task3 := NewTask("task3", Work(func(p Params) error { fmt.Println("doing work from task 3"); return nil; }), Work(func(p Params) error { fmt.Println("undoing work from task 3"); return nil }))
	task4 := NewTask("task4", Work(func(p Params) error { fmt.Println("doing work from task 4"); return nil; }), Work(func(p Params) error { fmt.Println("undoing work from task 4"); return nil }))
	task5 := NewTask("task5", Work(func(p Params) error { fmt.Println("doing work from task 5"); return errors.New("some error!") }), Work(func(p Params) error { fmt.Println("undoing work from task 5"); return nil }))

	grouped := Group(task3, task4, task5)

	p := Params{}
	err := task1.SetAutoResponse("y").Then(task2).Then(grouped).Exec(p)
	if err == nil {
		t.Error("unexpected: error was nil")
	}
}

func TestRollback(t *testing.T) {
	task1 := NewTask("task1", Work(func(p Params) error { fmt.Println("doing work from task 1"); return nil }), Work(func(p Params) error { fmt.Println("undoing work from task 1"); return nil }))
	task2 := NewTask("task2", Work(func(p Params) error { fmt.Println("doing work from task 2"); return nil }), Work(func(p Params) error { fmt.Println("undoing work from task 2"); return nil }))
	task3 := NewTask("task3", Work(func(p Params) error { fmt.Println("doing work from task 3"); return nil; }), Work(func(p Params) error { fmt.Println("undoing work from task 3"); return nil }))
	task4 := NewTask("task4", Work(func(p Params) error { fmt.Println("doing work from task 4"); return nil; }), Work(func(p Params) error { fmt.Println("undoing work from task 4"); return nil }))
	task5 := NewTask("task5", Work(func(p Params) error { fmt.Println("doing work from task 5"); return errors.New("some error!") }), Work(func(p Params) error { fmt.Println("undoing work from task 5"); return nil }))

	p := Params{}
	tasks := task1.Then(task2).Then(task3).Then(task4).Then(task5).SetAutoResponse("y")
	err := tasks.Rollback(p)
	if err != nil {
		t.Error("unexpected: rollback error was not not nil", err)
	}
}

var validationCases = []struct {
	Name    string
	Success bool
	Condition
	Params
}{
	{
		Name:    "Simple",
		Success: true,
		Params: Params{
			"A": "B",
		},
		Condition: Condition{
			Key: "A",
		},
	},
	{
		Name:    "SimpleFail",
		Success: false,
		Params: Params{
			"A": "B",
		},
		Condition: Condition{
			Key: "C",
		},
	},
	{
		Name:    "And",
		Success: true,
		Params: Params{
			"A": "B",
			"C": "D",
		},
		Condition: Condition{
			And: []Condition{
				{
					Key:   "A",
					Value: "B",
				},
				{
					Key: "C",
				},
			},
		},
	},
	{
		Name:    "AndFail",
		Success: false,
		Params: Params{
			"A": "B",
			"C": "D",
		},
		Condition: Condition{
			And: []Condition{
				{
					Key:   "A",
					Value: "B",
				},
				{
					Key:   "C",
					Value: "E",
				},
			},
		},
	},
	{
		Name:    "Or",
		Success: true,
		Params: Params{
			"A": "B",
			"C": "D",
		},
		Condition: Condition{
			Or: []Condition{
				{
					Key:   "A",
					Value: "B",
				},
				{
					Key:   "F",
					Value: "G",
				},
			},
		},
	},
	{
		Name:    "OrFail",
		Success: false,
		Params: Params{
			"A": "B",
			"C": "D",
		},
		Condition: Condition{
			Or: []Condition{
				{
					Key:   "A",
					Value: "N",
				},
				{
					Key:   "F",
					Value: "G",
				},
			},
		},
	},
	{
		Name:    "Xor",
		Success: true,
		Params: Params{
			"A": "B",
			"C": "D",
		},
		Condition: Condition{
			Xor: []Condition{
				{
					Key: "A",
				},
				{
					Key:   "F",
					Value: "G",
				},
			},
		},
	},
	{
		Name:    "XorFail",
		Success: false,
		Params: Params{
			"A": "B",
			"C": "D",
		},
		Condition: Condition{
			Xor: []Condition{
				{
					Key:   "A",
					Value: "B",
				},
				{
					Key:   "C",
					Value: "D",
				},
			},
		},
	},
	{
		Name:    "EmptyString",
		Success: false,
		Params: Params{
			"A": "",
		},
		Condition: Condition{
			Key:   "A",
		},
	},
	{
		Name:    "MatchString",
		Success: true,
		Params: Params{
			"A": "ABA",
		},
		Condition: Condition{
			Key:   "A",
			Value: "ABA|CBC",
			Comparison: Match,
		},
	},
	{
		Name:    "MatchStringFail",
		Success: false,
		Params: Params{
			"A": "ABA",
		},
		Condition: Condition{
			Key:   "A",
			Value: "ABAC|CBC",
			Comparison: Match,
		},
	},
	{
		Name:    "MatchStringBadRegex",
		Success: false,
		Params: Params{
			"A": "ABA",
		},
		Condition: Condition{
			Key:   "A",
			Value: "ABA[",
			Comparison: Match,
		},
	},
}

func TestParamValidation(t *testing.T) {
	for _, c := range validationCases {
		t.Run(c.Name, func(t *testing.T) {
			task := Task{
				Name:      c.Name,
				Forward:   WorkPassthrough,
				Backward:  WorkPassthrough,
				Condition: c.Condition,
			}
			err := task.Exec(c.Params)
			assert.Equal(t, c.Success, err == nil, err)
		})
	}
}
