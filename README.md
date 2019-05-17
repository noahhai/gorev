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

## Conditions
Parameter validation conditions may be specified on a task using AND, OR and XOR relationships. This may be useful for example when using this library in a CLI or anywhere there are different options for necessary/sufficient parameters. 

For example, in a deployment CLI, --cert-path would only be a necessary parameter if --ssl is specified. In this case we can require --cert-path XOR --ssl=false.

With more declarative validation, the code footprint can be reduced.

```go
c := Condition{
    Xor: []Condition{
        {
            Key: "UseLocalDatabase",
            Value:true,
        },
        {
            And: []Condition{
                {
                    Key:"ExternalDatabaseHost",
                },
                {
                    Key:"ExternalDatabasePort",
                },
                {
                    Key:"ExternalDatabaseUser",
                },
                {
                    Key:"ExternalDatabasePass",
                },
            },
         },
    },
 },

task := Task{
	Condition: c,
	...
}
```