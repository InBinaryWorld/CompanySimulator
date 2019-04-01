package main
//Szafraniak Krzysztof
// 244932

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
//Guard function
func maybe(guard bool, c chan Task) chan Task {
	if !guard {
		return nil
	}
	return c
}

//Guard function
func maybeChan(guard bool, c chan chan Task) chan chan Task {
	if !guard {
		return nil
	}
	return c
}

//Print task
func toString(task *Task) {
	fmt.Println(task.arg1, task.operator, task.arg2, "=", task.result)
}

//Print task without result
func toStringWithoutResult(task *Task) {
	fmt.Println(task.arg1, task.operator, task.arg2)
}

//Function allow to communicate with user
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

//Responsible for storage task
func taskTable(pullTasks chan Task, pushTask chan chan Task, pullTableCmd chan []Task, pushTableCmd chan int) {
	var tTable = make([]Task, 0)
	var length = 0
	for {
		select {
		case cmd := <-pushTableCmd:
			switch cmd {
			case 1:
				pullTableCmd <- tTable
			default:
			}
		case read := <-maybeChan(length > 0, pushTask):
			read <- tTable[0]
			tTable = tTable[1:]
			length--
		case write := <-maybe(length < settings.MaxTasks, pullTasks):
			tTable = append(tTable, write)
			length++
		default:
		}
		time.Sleep(time.Millisecond)
	}
}

//Storage solved tasks
func warehouse(pullResult chan Task, pushResult chan chan Task, pullWareCmd chan []Task, pushWareCmd chan int) {
	var wh = make([]Task, 0)
	var length = 0
	for {
		select {
		case cmd := <-pushWareCmd:
			switch cmd {
			case 1:
				pullWareCmd <- wh
			default:
			}
		case read := <-maybeChan(length > 0, pushResult):
			read <- wh[0]
			wh = wh[1:]
			length--
		case write := <-maybe(length < settings.MaxResults, pullResult):
			wh = append(wh, write)
			length++
		default:
		}
		time.Sleep(time.Millisecond)
	}
}

//pop task, solve it and push it to warehouse
func worker(id int, popTask chan chan Task, pullResult chan Task) {
	for {
		taskC := make(chan Task, 1)
		popTask <- taskC
		task := <-taskC
		if TALKATIVE {
			fmt.Println("Worker		", id, " do 			", task.arg1, task.operator, task.arg2)
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
		time.Sleep(time.Duration(rand.Intn(40)+80) * time.Millisecond * settings.WorkerSpeed)
	}
}

//Create new tasks
func boss(id int, pullTask chan Task) {
	tab := []string{"+", "-", "*"}
	for {
		task := Task{rand.Intn(1000), rand.Intn(1000), tab[rand.Intn(len(tab))], 0}
		if TALKATIVE {
			fmt.Println("Boss		", id, " pull new Task 	", task.arg1, task.operator, task.arg2)
		}
		pullTask <- task
		time.Sleep(time.Duration(rand.Intn(40)+80) * time.Millisecond * settings.BossSpeed)
	}
}

//pop tasks from warehouse
func client(id int, popResult chan chan Task) {
	for {
		result := make(chan Task)
		popResult <- result
		task := <-result
		if TALKATIVE {
			fmt.Println("Client		", id, " take Result 	", task.arg1, task.operator, task.arg2, "=", task.result)
		}
		time.Sleep(time.Duration(rand.Intn(40)+80) * time.Millisecond * settings.ClientSpeed)
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
	pushTask := make(chan chan Task, 2)
	pullTask := make(chan Task, 2)
	pushResult := make(chan chan Task, 2)
	pullResult := make(chan Task, 2)
	pullTableCmd := make(chan []Task, 2)
	pushTableCmd := make(chan int, 2)
	pullWareCmd := make(chan []Task, 2)
	pushWareCmd := make(chan int, 2)

	if !TALKATIVE {
		go userInterface(pullTableCmd, pushTableCmd, pullWareCmd, pushWareCmd)
	}
	//Start taskTable and warehouse
	go taskTable(pullTask, pushTask, pullTableCmd, pushTableCmd)
	go warehouse(pullResult, pushResult, pullWareCmd, pushWareCmd)

	//Start Bosses
	for i := 0; i < settings.Boss; i++ {
		go boss(i, pullTask)
	}

	//Start Workers
	for i := 0; i < settings.Workers; i++ {
		go worker(i, pushTask, pullResult)
	}

	//Start Clients
	for i := 0; i < settings.Clients; i++ {
		go client(i, pushResult)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
