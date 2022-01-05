package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"go.bug.st/serial.v1/enumerator"
)

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
	printInfo("Reading serial numbers: %v\n", lines)
	return lines
}

// excecutes instruction represented as []Task
func excecuteInstruction(c chan int, tasks []Task) {
	// unique instruction id created and stored to package level list
	var instructionId int

	activeInstructionsLimit := 1000
	activeTasksLimit := 3000

	if len(instructionIds) > activeInstructionsLimit {
		printWarning("%d active instructions limit exceeded, abandoning instruction.", activeInstructionsLimit)
		return
	}

	mutex.Lock()
	for {
		random := randInt(1000, 9999)
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

		if len(getActiveTaskIds()) > activeTasksLimit {
			printWarning("%d active tasks limit exceeded, abandoning instruction.", activeInstructionsLimit)
			return
		}

		for {
			random := randInt(1000, 9999)
			if !idInRedisArray("activeTaskIds", random) {
				taskJSON := taskToJSON(task)
				storeActiveTask(random, taskJSON)
				break
			}
		}

		currentSettings := outputToMap(singleOutputRead())

		for k, v := range task.Vxx {
			currentSetting := int(currentSettings[k].(int64))
			changedSetting := vxxRequirementsToDec(currentSetting, v)
			command := fmt.Sprintf("<SET_%v=%v;>", k, changedSetting)
			// arduino.ResetInputBuffer()
			_, err := arduino.Write([]byte(command))
			check(err)
		}

		for k, v := range task.Txx {
			command := fmt.Sprintf("<SET_%v=%v;>", k, v)
			printDebug("TEMP COMMAND: %s\n", command)
			_, err := arduino.Write([]byte(command))
			check(err)
		}

		if task.Pump == "ON" {
			command := "<PUMP_ON;>"
			_, err := arduino.Write([]byte(command))
			check(err)
		} else if task.Pump == "OFF" {
			command := "<PUMP_OFF;>"
			_, err := arduino.Write([]byte(command))
			check(err)
		}

		for i := 0; i < task.Sleep; i++ {
			time.Sleep(1 * time.Second)
		}
	}
	printInfo("Instruction %d done!\n", c)
}

// reads serial output untill it matches validation check
func singleOutputRead() string {
	for {
		arduino.ResetInputBuffer()
		_, err := arduino.Write([]byte("<GET_ALL;>"))
		check(err)
		// time.Sleep(30 * time.Millisecond)
		scanner := bufio.NewScanner(arduino)
		scanner.Scan()
		output := scanner.Text()
		if outputIsValid(output, re) {
			return output
		}
		time.Sleep(20 * time.Millisecond)
	}
}
