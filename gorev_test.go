package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestFlow(t *testing.T) {
	task1 := NewTask("task1", Work(func(p Params)error{fmt.Println("doing work from task 1"); return nil}),Work(func(p Params)error{fmt.Println("undoing work from task 1"); return nil}))
	task2 := NewTask("task2", Work(func(p Params)error{fmt.Println("doing work from task 2"); return nil}),Work(func(p Params)error{fmt.Println("undoing work from task 2"); return nil}))
	task3 := NewTask("task3", Work(func(p Params)error{fmt.Println("doing work from task 3"); return nil;}),Work(func(p Params)error{fmt.Println("undoing work from task 3"); return nil}))
	task4 := NewTask("task4", Work(func(p Params)error{fmt.Println("doing work from task 4"); return nil;}),Work(func(p Params)error{fmt.Println("undoing work from task 4"); return nil}))
	task5 := NewTask("task5", Work(func(p Params)error{fmt.Println("doing work from task 5"); return errors.New("some error!")}),Work(func(p Params)error{fmt.Println("undoing work from task 5"); return nil}))

	p := Params{}
	err := task1.Then(task2).Then(task3).Then(task4).Then(task5).Exec(p)
	if err == nil {
		t.Error("error was not null")
	}

}
