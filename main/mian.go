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

type MachineManager struct {
	push chan Task
	pull chan Task
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
func userInterface(pullTableCmd chan []Task, pushTableCmd chan int, pullWareCmd chan []Task,
	pushWareCmd chan int, pushWorker [settings.Workers]chan int, pullWorker chan string) {
	var version int
	for {
		fmt.Println("1-Show TaskTable 2-Show Warehouse 3-Worker Info")
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
		case 3:
			for _, k := range pushWorker {
				k <- 1
			}
			for i := 0; i < settings.Workers; i++ {
				fmt.Println(<-pullWorker)
			}
		}
		time.Sleep(time.Millisecond)
	}
}

//Responsible for storage task
func taskTable(pushTasks chan Task, pullTask chan chan Task, pullTableCmd chan []Task, pushTableCmd chan int) {
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
		case read := <-maybeChan(length > 0, pullTask):
			read <- tTable[0]
			tTable = tTable[1:]
			length--
		case write := <-maybe(length < settings.MaxTasks, pushTasks):
			tTable = append(tTable, write)
			length++
		default:
		}
		time.Sleep(time.Millisecond)
	}
}

//Storage solved tasks
func warehouse(pushResult chan Task, pullResult chan chan Task, pullWareCmd chan []Task, pushWareCmd chan int) {
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
		case read := <-maybeChan(length > 0, pullResult):
			read <- wh[0]
			wh = wh[1:]
			length--
		case write := <-maybe(length < settings.MaxResults, pushResult):
			wh = append(wh, write)
			length++
		default:
		}
		time.Sleep(time.Millisecond)
	}
}

func addingMachine(id int, manager MachineManager) {
	for {
		task := <-manager.push
		time.Sleep(time.Duration(time.Millisecond * settings.AddingMachinesSpeed))
		if TALKATIVE {
			fmt.Println("Adding		", id, " do 			", task.arg1, task.operator, task.arg2)
		}
		switch task.operator {
		case "+":
			r := task.arg1 + task.arg2
			task.result = r
			manager.pull <- task
		case "-":
			r := task.arg1 - task.arg2
			task.result = r
			manager.pull <- task

		}
	}
}

func multiplyingMachine(id int, manager MachineManager) {
	for {

		task := <-manager.push
		time.Sleep(time.Duration(time.Millisecond * settings.MultiplyingMachinesSpeed))
		if TALKATIVE {
			fmt.Println("Multi		", id, " do 			", task.arg1, task.operator, task.arg2)
		}
		r := task.arg2 * task.arg1
		task.result = r
		manager.pull <- task
	}
}

//pop task, solve it and push it to warehouse
func worker(id int, pullTask chan chan Task, pushResult chan Task, addManagers [settings.AddMachines]MachineManager,
	multiManagers [settings.MultiMachines]MachineManager, pushCmd chan int, pullCmd chan string, r *rand.Rand) {
	counter := 0
	isPatient := r.Intn(2) == 0

	for {
		select {
		case cmd := <-pushCmd:
			if cmd == 1 {
				str := fmt.Sprintf("Worker %d Patient : %t Done: %d", id, isPatient, counter)
				pullCmd <- str
			}
		default:
		}

		taskC := make(chan Task, 1)
		pullTask <- taskC
		task := <-taskC
		if TALKATIVE {
			fmt.Println("Worker		", id, " do 			", task.arg1, task.operator, task.arg2)
		}

		var manager MachineManager
		var maxIdx int
		var idx int

		if task.operator == "*" {
			maxIdx = settings.MultiMachines
			idx = r.Intn(maxIdx)
			manager = multiManagers[idx]
		} else {
			maxIdx = settings.AddMachines
			idx = r.Intn(maxIdx)
			manager = addManagers[idx]
		}

		if isPatient {
			manager.push <- task
			pushResult <- <-manager.pull
		} else {
			done := false
			for done == false {
				select {
				case manager.push <- task:
					pushResult <- <-manager.pull
					done = true
				case <-time.After(time.Millisecond * settings.UnPatientTime):
					idx++
					if idx == maxIdx {
						idx = 0
					}
					if task.operator == "*" {
						manager = multiManagers[idx]
					} else {
						manager = addManagers[idx]
					}
				}
			}
		}
		counter++
		time.Sleep(time.Duration(time.Millisecond * settings.WorkerSpeed))
	}
}

//Create new tasks
func boss(id int, pullTask chan Task) {
	tab := []string{"+", "-","*"}
	for {
		task := Task{rand.Intn(1000), rand.Intn(1000), tab[rand.Intn(len(tab))], 0}
		if TALKATIVE {
			fmt.Println("Boss		", id, " pull new Task 	", task.arg1, task.operator, task.arg2)
		}
		pullTask <- task
		time.Sleep(time.Duration(time.Millisecond * settings.BossSpeed))
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
		time.Sleep(time.Duration(time.Millisecond * settings.ClientSpeed))
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
	pullTask := make(chan chan Task, 2)
	pushTask := make(chan Task, 2)
	pullResult := make(chan chan Task, 2)
	pushResult := make(chan Task, 2)
	pullTableCmd := make(chan []Task, 2)
	pushTableCmd := make(chan int, 2)
	pullWareCmd := make(chan []Task, 2)
	pushWareCmd := make(chan int, 2)

	var addingManager [settings.AddMachines]MachineManager
	var multiplyingManager [settings.MultiMachines]MachineManager
	var pushWorkerCmd [settings.Workers]chan int
	pullWorkerCmd := make(chan string, settings.Workers)

	for i := 0; i < settings.AddMachines; i++ {
		addingManager[i] = MachineManager{make(chan Task, 1), make(chan Task, 1)}
	}

	for i := 0; i < settings.MultiMachines; i++ {
		multiplyingManager[i] = MachineManager{make(chan Task, 1), make(chan Task, 1)}
	}

	for i := 0; i < settings.Workers; i++ {
		pushWorkerCmd[i] = make(chan int, 2)
	}

	if !TALKATIVE {
		go userInterface(pullTableCmd, pushTableCmd, pullWareCmd, pushWareCmd, pushWorkerCmd, pullWorkerCmd)
	}
	//Start taskTable and warehouse
	go taskTable(pushTask, pullTask, pullTableCmd, pushTableCmd)
	go warehouse(pushResult, pullResult, pullWareCmd, pushWareCmd)

	for i := 0; i < settings.AddMachines; i++ {
		go addingMachine(i, addingManager[i])
	}

	for i := 0; i < settings.MultiMachines; i++ {
		go multiplyingMachine(i, multiplyingManager[i])
	}

	//Start Bosses
	for i := 0; i < settings.Boss; i++ {
		go boss(i, pushTask)
	}

	//Start Workers

	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	for i := 0; i < settings.Workers; i++ {
		go worker(i, pullTask, pushResult, addingManager, multiplyingManager, pushWorkerCmd[i], pullWorkerCmd, r)
	}

	//Start Clients
	for i := 0; i < settings.Clients; i++ {
		go client(i, pullResult)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
