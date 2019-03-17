package main

import (
	"./settings"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var TALKATIVE = false

type Task struct {
	arg1     int
	arg2     int
	operator string
	result   int
}

func toString(task *Task) {
	fmt.Println(task.arg1, task.operator, task.arg2, "=", task.result)
}

func toStringWithoutResult(task *Task) {
	fmt.Println(task.arg1, task.operator, task.arg2)
}

func userInterface(pullTableCmd chan []Task, pushTableCmd chan int, pullWareCmd chan []Task, pushWareCmd chan int) {
	var version int
	for {
		fmt.Println("1-Show TaskTable 2-Show Warehouse")
		fmt.Print("Option: ")
		fmt.Scanf("%d\n", &version)
		switch version {
		case 1:
			pushTableCmd <- 1
			tab := <-pullTableCmd
			for _, v := range tab {
				toStringWithoutResult(&v)
			}
		case 2:
			pushWareCmd <- 1
			tab := <-pullWareCmd
			for _, v := range tab {
				toString(&v)
			}
		}
	}
}

func taskTable(pullTasks chan Task, popTask chan chan Task, pullTableCmd chan []Task, pushTableCmd chan int) {
	var tTable = make([]Task, 0)
	for {
		select {
		case cmd := <-pushTableCmd:
			switch cmd {
			case 1:
				pullTableCmd <- tTable
			default:
			}
		default:
		}
		length := len(tTable)
		if length > 0 {
			select {
			case read := <-popTask:
				read <- tTable[0]
				tTable = tTable[1:]
			default:
			}
		}
		if length < settings.MaxTasks {
			select {
			case write := <-pullTasks:
				tTable = append(tTable, write)
			default:
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func warehouse(pullResult chan Task, popResult chan chan Task, pullWareCmd chan []Task, pushWareCmd chan int) {
	var wh = make([]Task, 0)
	for {
		select {
		case cmd := <-pushWareCmd:
			switch cmd {
			case 1:
				pullWareCmd <- wh
			default:
			}
		default:
		}
		length := len(wh)
		if length > 0 {
			select {
			case read := <-popResult:
				read <- wh[0]
				wh = wh[1:]
			default:
			}
		}
		if length < settings.MaxResults {
			select {
			case write := <-pullResult:
				wh = append(wh, write)
			default:
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func worker(id int, popTask chan chan Task, pullResult chan Task) {
	for {
		taskC := make(chan Task, 1)
		popTask <- taskC
		task := <-taskC
		if TALKATIVE {
			fmt.Println("Worker		", id, " do 				", task.arg1, task.operator, task.arg2)
		}
		switch task.operator {
		case "+":
			r := task.arg1 + task.arg2
			task.result = r
			pullResult <- task
		case "-":
			r := task.arg1 - task.arg2
			task.result = r
			pullResult <- task
		case "*":
			r := task.arg2 * task.arg1
			task.result = r
			pullResult <- task
		}
		time.Sleep(settings.WorkerSpeed * time.Millisecond)
	}
}

func boss(id int, pullTask chan Task) {
	tab := []string{"+", "-", "*"}
	for {
		task := Task{rand.Intn(1000), rand.Intn(1000), tab[rand.Intn(len(tab))], 0}
		if TALKATIVE {
			fmt.Println("Boss		", id, " pull new Task 	", task.arg1, task.operator, task.arg2)
		}
		pullTask <- task
		time.Sleep(settings.BossSpeed * time.Millisecond)
	}
}

func client(id int, popResult chan chan Task) {
	for {
		result := make(chan Task)
		popResult <- result
		task := <-result
		if TALKATIVE {
			fmt.Println("Client		", id, " take Result 	", task.arg1, task.operator, task.arg2, "=", task.result)
		}
		time.Sleep(settings.ClientSpeed * time.Millisecond)
	}
}

func main() {
	var version int
	fmt.Println("1 - TALKATIVE   2 - CALM")
	fmt.Print("Enter version: ")
	fmt.Scanf("%d", &version)

	if version == 1 {
		TALKATIVE = true
	}
	popTask := make(chan chan Task, 2)
	pullTask := make(chan Task, 2)
	popResult := make(chan chan Task, 2)
	pullResult := make(chan Task, 2)
	pullTableCmd := make(chan []Task, 2)
	pushTableCmd := make(chan int, 2)
	pullWareCmd := make(chan []Task, 2)
	pushWareCmd := make(chan int, 2)

	if !TALKATIVE {
		go userInterface(pullTableCmd, pushTableCmd, pullWareCmd, pushWareCmd)
	}
	go taskTable(pullTask, popTask, pullTableCmd, pushTableCmd)
	go warehouse(pullResult, popResult, pullWareCmd, pushWareCmd)

	for i := 0; i < settings.Boss; i++ {
		go boss(i, pullTask)
	}
	for i := 0; i < settings.Workers; i++ {
		go worker(i, popTask, pullResult)
	}
	for i := 0; i < settings.Clients; i++ {
		go client(i, popResult)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
