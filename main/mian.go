package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Task struct {
	arg1     int
	arg2     int
	operator string
	result   int
}

func taskTable(pullTasks chan Task, popTask chan chan Task) {
	var tTable = make([]Task, 0)
	for {
		leng := len(tTable)
		if leng > 0 {
			select {
			case read := <-popTask:
				read <- tTable[0]
				tTable = tTable[1:]
			default:
			}
		}
		if leng < MaxTasks {
			select {
			case write := <-pullTasks:
				tTable = append(tTable, write)
			default:
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func warehouse(pullResult chan Task, popResult chan chan Task) {
	var wh = make([]Task, 0)
	for {
		leng := len(wh)
		if leng > 0 {
			select {
			case read := <-popResult:
				read <- wh[0]
				wh = wh[1:]
			default:
			}
		}
		if leng < MaxResults {
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

		fmt.Println("Worker		", id, " do 				", task.arg1, task.operator, task.arg2)
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
		time.Sleep(time.Duration(rand.Intn(40)+80) * WorkerSpeed * time.Microsecond * 10)
	}
}

func boss(id int, pullTask chan Task) {
	tab := []string{"+", "-", "*"}
	for {
		task := Task{rand.Intn(1000), rand.Intn(1000), tab[rand.Intn(len(tab))], 0}
		fmt.Println("Boss		", id, " pull new Task 	", task.arg1, task.operator, task.arg2)
		pullTask <- task
		time.Sleep(time.Duration(rand.Intn(40)+80) * BossSpeed * time.Microsecond * 10)
	}
}

func client(id int, popResult chan chan Task) {
	for {
		result := make(chan Task)
		popResult <- result
		task := <-result
		fmt.Println("Client		", id, " take Result 	", task.arg1, task.operator, task.arg2, "=", task.result)
		time.Sleep(time.Duration(rand.Intn(40)+80) * ClientSpeed * time.Microsecond * 10)
	}
}

func main() {
	popTask := make(chan chan Task, 1)
	pullTask := make(chan Task, 1)
	popResult := make(chan chan Task, 1)
	pullResult := make(chan Task, 1)

	go taskTable(pullTask, popTask)
	go warehouse(pullResult, popResult)

	for i := 0; i < Boss ; i++ {
		go boss(i,pullTask)
	}
	for i := 0; i < Workers ; i++ {
		go worker(i, popTask, pullResult)
	}
	for i := 0; i < Clients ; i++ {
		go client(i,popResult)
	}
	
	time.Sleep(200 * time.Second)
}
