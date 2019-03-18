# gorev
Simple Library for Building Task Flows and Handling Rollback

## Example (see tests)
```golang
task1 := gorev.NewTask("task1",   
 gorev.Work(func(p gorev.Params) error { fmt.Println("doing work from task 1"); return nil }),   
 gorev.Work(func(p gorev.Params) error { fmt.Println("undoing work from task 1"); return nil }),  
)  
task2 := gorev.NewTask("task2",   
 gorev.Work(func(p gorev.Params) error { fmt.Println("doing work from task 2"); return nil }),   
 gorev.Work(func(p gorev.Params) error { fmt.Println("undoing work from task 2"); return nil }),  
)  
task3 := gorev.NewTask("task3",   
 gorev.Work(func(p gorev.Params) error { fmt.Println("doing work from task 3"); return errors.New("some error!") }),   
 gorev.Work(func(p gorev.Params) error { fmt.Println("undoing work from task 3"); return nil }),  
)  
  
p := gorev.Params{}  
tasks := task1.Then(task2).Then(task3)  
err := tasks.Exec(p)  
// or, rollback  
err = tasks.Rollback(p)
```
