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

// Maybe in future add task id to task struct... :)
func (t *Task) Stop(taskId int) {
	printInfo("Stopping task %v", taskId)
	removeActiveTask(taskId)
	removeIdFromRedisArr("activeTaskIds", taskId)
	removeFromKillList(taskId)
	// getAllRunningTasksNonDefaultRequirements()
	t.resetToDefaults(&defaults)
	currentSettings := outputToMap(singleOutputRead())

	// return to defaults only values that are not in use
	// step 1
	// need to collect all requirements of all running
	// tasks into one Task
	// compare them to defaults
	// leave only requirements that are not default

	activeIds := getActiveTaskIds()

	if len(activeIds) < 1 {
		printDebug("no active instructions")
		printDebug("excecuting neutralised task")
		printError("%v", t)
		for k, v := range t.Vxx {
			currentSetting := int(currentSettings[k].(int64))
			changedSetting := vxxRequirementsToDec(currentSetting, v)
			command := fmt.Sprintf("<SET_%v=%v;>", k, changedSetting)
			sendCommand(command)
		}
	} else {
		for _, id := range getActiveTaskIds() {
			JSONById := readActiveTask(id)
			comparedTask := JSONToTask(JSONById)
			for kT, vT := range t.Vxx {
				// make a copy of current Vxx iteration
				newVT := make([][2]int, len(t.Vxx[kT]))
				copy(newVT, t.Vxx[kT])
				for kC, vC := range comparedTask.Vxx {
					// like if V00 = V00
					if kT == kC {
						// comparing two arrays like {{1, 1}, {5,0}}
						for _, reqT := range vT {
							for _, reqC := range vC {
								// if first num of [2]int slice matches
								// remove requirement from neutralised task
								// as it is used elsewhere and cant be excecuted
								if reqT[0] == reqC[0] {
									newVT = removeRequirement(newVT, reqT)
								}
							}
						}
					}
				}
				currentSetting := int(currentSettings[kT].(int64))
				changedSetting := vxxRequirementsToDec(currentSetting, newVT)
				command := fmt.Sprintf("<SET_%v=%v;>", kT, changedSetting)
				printError(command)
				sendCommand(command)
				printInfo("Task %v stopped.", taskId)
			}
		}
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
