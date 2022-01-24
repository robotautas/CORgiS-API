package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial.v1/enumerator"
)

// preventing race conditions, as board likes to take some miliseconds to proceed commands
var mutexArduino sync.Mutex

// TODO: refactor every single command sent to arduino to use this function
//(and see what happens :)
func sendCommand(command string) {
	mutexArduino.Lock()
	defer mutexArduino.Unlock()
	_, err := arduino.Write([]byte(command))
	check(err)
	if command != "<GET_ALL;>" {
		printInfo("%v command sent!", command)
	}
	// take a nap
	time.Sleep(time.Millisecond * 20)
}

// Scan ports for arduino, return first port whose serial number meets one of the S/Ns in serial_numbers.txt file.
// If arduino not found, it makes a program wait for device.

func findArduinoPort() string {
	if OS != "windows" {
		for {
			ports, err := enumerator.GetDetailedPortsList()
			check(err)
			for _, port := range ports {
				if port.IsUSB {
					for _, sn := range getSerialNumbers("serial_numbers.txt") {
						if sn == port.SerialNumber {
							return port.Name
						}
					}
				}
			}
			printError("Arduino device not found. Check if connected!")
			time.Sleep(time.Second * 1)
		}
	}
	return "COM10"
}

//Reads serial numbers from file, removes whitespace and returns array
func getSerialNumbers(path string) []string {
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sn := scanner.Text()
		// TODO use strings.TrimSpace!
		if strings.Contains(sn, " ") {
			sn = strings.ReplaceAll(sn, " ", "")
		}
		lines = append(lines, sn)
	}
	printInfo("Reading serial numbers: %v", lines)
	return lines
}

// excecutes instruction represented as []Task
func excecuteInstruction(c chan int, tasks []Task) {
	// unique instruction id created and stored to package level list
	var instructionId int

	activeInstructionsLimit := 1000
	// activeTasksLimit := 3000

	if len(instructionIds) > activeInstructionsLimit {
		printWarning("%d active instructions limit exceeded, abandoning instruction.", activeInstructionsLimit)
		return
	}

	mutex.Lock()
	for {
		random := randInt(1000, 9999)
		// printDebug("INSTRUCTION ID DEBUG %v", random)
		if !intInSlice(instructionIds, random) {
			instructionId = random
			instructionIds = append(instructionIds, instructionId)
			break
		}
	}
	mutex.Unlock()
	c <- instructionId

	// register and excecute tasks one by one
	for _, task := range tasks {
		// var taskId int

		// if len(getActiveTaskIds()) > activeTasksLimit {
		// 	printWarning("%d active tasks limit exceeded, abandoning instruction.", activeInstructionsLimit)
		// 	return
		// }

		// var taskId int

		// for {
		// 	random := randInt(1000, 9999)
		// 	if !idInRedisArray("activeTaskIds", random) {
		// 		taskJSON := taskToJSON(task)
		// 		taskId = random
		// 		storeActiveTask(random, taskJSON)
		// 		break
		// 	}
		// }

		currentSettings := outputToMap(singleOutputRead())
		printDebug("Starting task %v.", task.Id)
		for k, v := range task.Vxx {
			currentSetting := int(currentSettings[k].(int64))
			changedSetting := vxxRequirementsToDec(currentSetting, v)
			command := fmt.Sprintf("<SET_%v=%v;>", k, changedSetting)
			// arduino.ResetInputBuffer()
			// _, err := arduino.Write([]byte(command))
			// check(err)
			// time.Sleep(time.Millisecond * 20)
			sendCommand(command)
		}

		for k, v := range task.Txx {
			command := fmt.Sprintf("<SET_%v=%v;>", k, v)
			// printDebug("TEMP COMMAND: %s", command)
			// _, err := arduino.Write([]byte(command))
			// check(err)
			// time.Sleep(time.Millisecond * 20)
			sendCommand(command)
		}

		if task.Pump == "ON" {
			command := "<PUMP_ON;>"
			sendCommand(command)
		} else if task.Pump == "OFF" {
			command := "<PUMP_OFF;>"
			sendCommand(command)
		}
		printDebug("Task %v started!", task.Id)
		for i := 0; i < task.Sleep; i++ {
			time.Sleep(1 * time.Second)
		}
		task.Stop(task.Id)
	}
	printInfo("Instruction %v done!", instructionId)
}

// Transforms the running task to neutral and excecute
// but only the settings that are not shared
func (t *Task) Stop(taskId int) {
	printInfo("Stopping task %v", taskId)
	removeActiveTask(taskId)
	removeIdFromRedisArr("activeTaskIds", taskId)
	removeFromKillList(taskId)

	// neutralise task - set all task requirements to defaults
	t.resetToDefaults(&defaults)

	// get all current settings in decimal values, for changing later
	currentSettings := outputToMap(singleOutputRead())
	activeIds := t.overlappingTasks()

	// make copy of t for aggregation
	stoppingTask := *t
	printWarning("STOPPING TASK BEFORE REMOVED REQUIREMENT %v", stoppingTask)

	for k, v := range t.Vxx {
		updatedReqs := stoppingTask.Vxx[k]
		for _, req := range v {
			if requirementUsedElsewhere(k, req, t.InstructionId) {
				updatedReqs = removeRequirement(updatedReqs, req)
				printWarning("REMOVED REQ %v", req)
			}
			stoppingTask.Vxx[k] = updatedReqs
		}
	}

	printWarning("STOPPING TASK AFTER REMOVED REQUIREMENT %v", stoppingTask)

	// if the stopping task is the only one running at the time, just excecute neutralised task
	if len(activeIds) < 1 {
		printDebug("excecuting neutralised task - excecuting neutralised task")
		for k, v := range t.Vxx {
			currentSetting := int(currentSettings[k].(int64))
			changedSetting := vxxRequirementsToDec(currentSetting, v)
			command := fmt.Sprintf("<SET_%v=%v;>", k, changedSetting)
			sendCommand(command)
		}
		//else iterate over active tasks to identify if stopping task will not interfer
		//and excecute only the parts which will not
	} else {
		for k, v := range stoppingTask.Vxx {
			currentSetting := int(currentSettings[k].(int64))
			changedSetting := vxxRequirementsToDec(currentSetting, v)
			command := fmt.Sprintf("<SET_%v=%v;>", k, changedSetting)
			sendCommand(command)
		}
		printInfo("Task %v stopped.", taskId)
	}
}

// changes task's values to system defaults
func (t *Task) resetToDefaults(defaults *Task) {
	// k like "V00", v like {{2, 0}, {3, 1}}
	for k, v := range t.Vxx {
		// requirement like {3, 1}
		var newValue [][2]int
		for _, requirement := range v {
			// if requirement differs from default requirement, change it to default
			if requirement != defaults.Vxx[k][requirement[0]] {
				newValue = append(newValue, defaults.Vxx[k][requirement[0]])
			}
			fmt.Printf("New value: %v\n", newValue)
		}
		t.Vxx[k] = newValue
	}
	// TODO: temperatures??? need business logic.
	if t.Pump != defaults.Pump {
		t.Pump = defaults.Pump
	}
}

// reads serial output untill it matches validation check
func singleOutputRead() string {
	for {
		// arduino.ResetInputBuffer()
		arduino.ResetOutputBuffer()
		// _, err := arduino.Write([]byte("<GET_ALL;>"))
		// check(err)
		sendCommand("<GET_ALL;>")
		// time.Sleep(30 * time.Millisecond)
		scanner := bufio.NewScanner(arduino)
		scanner.Scan()
		output := scanner.Text()
		if outputIsValid(output, re) {
			return output
		}
		// time.Sleep(20 * time.Millisecond)
	}
}
