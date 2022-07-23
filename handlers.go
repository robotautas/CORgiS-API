package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func SetMultiHandler(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query().Encode()
	params := strings.Split(urlParams, "&")
	commands := "<"
	for _, i := range params {
		if strings.HasPrefix(i, "V") ||
			strings.HasPrefix(i, "T") {
			param := strings.Split(i, "=")[0]
			value := strings.Split(i, "=")[1]
			if !URLParamValid(param) {
				w.Write([]byte(fmt.Sprintf("error: unknown param %s!", param)))
				return
			} else if !URLValueValid(param, value) {
				w.Write([]byte(fmt.Sprintf("error: invalid value %s for param %s!", value, param)))
				return
			} else {
				commands = commands + fmt.Sprintf("SET_%s=%s;", param, value)
			}
		} else if i == "PUMP_ON=" {
			commands = commands + "PUMP_ON;"
		} else if i == "PUMP_OFF=" {
			commands = commands + "PUMP_OFF;"
		} else {
			w.Write([]byte(fmt.Sprintf("error: unable to parse %s!", i)))
			return
		}
	}
	commands = commands + ">"
	sendCommand(commands)
	w.Write([]byte(fmt.Sprintf("command %s sent!", commands)))
}

func SetHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	param := r.URL.Query().Get("param")
	value := r.URL.Query().Get("value")

	// make sure, that param & value combination is valid
	if !URLParamValid(param) {
		w.Write([]byte("error: incorrect param!"))
		printWarning("Invalid request! Parameter %v incorrect!", param)
		return
	}
	if !URLValueValid(param, value) {
		w.Write([]byte("error: incorrect value!"))
		printWarning("Invalid request! Value %v incorrect!", value)
		return
	}

	// format and send a command to the device
	command := ""
	if strings.HasPrefix(param, "PUMP") {
		command = "<" + param + ";>"
		_, err := arduino.Write([]byte(command))
		check(err)
		printInfo("Command sent: %v", command)
	} else {
		command = "<SET_" + param + "=" + value + ";>"
		_, err := arduino.Write([]byte(command))
		if err != nil {
			w.Write([]byte("error: could not send a command to device, check if connected!"))
		}
		printInfo("Command sent: %v", command)
	}

	time.Sleep(30 * time.Millisecond)

	// format and send a response depending on parameter
	if stringInSlice(param, VxxParams) {
		valueToInt, err := strconv.ParseInt(value, 10, 64)
		check(err)
		for {
			answer := outputToMap(singleOutputRead())
			if answer[param] == valueToInt {
				jsonString, err := json.Marshal(answer)
				check(err)
				w.Write([]byte(jsonString))
				printInfo("Valid response received.")
				break
			} else {
				printError("Response FAILED, %v != %v! Reading again..", answer[param], value)
				time.Sleep(20 * time.Millisecond)
			}
		}
	} else if stringInSlice(param, pumpParams) {
		for {
			answer := outputToMap(singleOutputRead())
			if param == "PUMP_ON" && answer["PUMP"] == int64(1) {
				jsonString, err := json.Marshal(answer)
				check(err)
				w.Write([]byte(jsonString))
				printInfo("Valid response received.")
				break
			} else if param == "PUMP_OFF" && answer["PUMP"] == int64(0) {
				jsonString, err := json.Marshal(answer)
				check(err)
				w.Write([]byte(jsonString))
				printInfo("Valid response received.")
				break
			} else {
				printError("Response FAILED! Param = '%v', pump value = '%v'", param, answer["PUMP"])
				time.Sleep(80 * time.Millisecond)
			}
		}
		// temperature is inertic, so it doesn't really need imediate response
	} else if stringInSlice(param, TxxParams) {
		answer := outputToMap(singleOutputRead())
		jsonString, err := json.Marshal(answer)
		check(err)
		w.Write([]byte(jsonString))
		printInfo("Valid response received.")
	} else {
		w.Write([]byte("error: something unexpected happened"))
	}
	finish := time.Since(start)
	printInfo("Response took %v", finish)
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	answer := outputToMap(singleOutputRead())
	jsonString, err := json.Marshal(answer)
	check(err)
	w.Write([]byte(jsonString))
	printInfo("Valid response received.")
}

// Accepts JSON string from request, starts a process routine
func StartHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var body []Task
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}

	// validate business logic
	if validateJSONTasks(body) {
		//add timestamps
		tasks := addTimeIntervals(body)

		// check for conflicts
		for _, task := range tasks {
			overlappingTasks := task.overlappingTasks()
			if task.conflictsWith(overlappingTasks) {
				response := "Conflicting instruction!"
				w.Write([]byte(response))
				return
			}
		}

		// create instruction id
		instruction := make(Instruction)
		var instructionId int
		mutex.Lock()
		for {
			random := randInt(1000, 9999)

			if !intInSlice(instructionIds, random) {
				instructionId = random
				// this became redundant, refactor to redis only in future
				instructionIds = append(instructionIds, instructionId)
				break
			}
		}
		mutex.Unlock()

		// modify tasks so that they know their ids, write them to redis
		var taskIds []int
		var modifiedTasks []Task
		for _, task := range tasks {

			for {
				random := randInt(1000, 9999)
				if !idInRedisArray("activeTaskIds", random) {
					task.addIds(instructionId, random)
					taskJSON := taskToJSON(task)
					storeActiveTask(random, taskJSON)
					taskIds = append(taskIds, random)
					modifiedTasks = append(modifiedTasks, task)
					break
				}
			}
		}
		tasks = modifiedTasks

		// write instruction-task map to redis
		instruction[instructionId] = taskIds
		addInstructionToRedis(instruction)

		for _, task := range tasks {
			// printWarning("%v", task)
			//debug - atspausdina pakkankamai info, kad galima atsekti, ar nedaromos klaidos
			for k, v := range task.Vxx {
				printDebug("%v: %v", k, v)
			}
		}

		go excecuteInstruction(instructionId, tasks)
		idToString := strconv.Itoa(instructionId)
		w.Write([]byte(idToString))
	}
}

func StopHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		panic(err)
	}
	killInstructionIds = append(killInstructionIds, idInt)
	addToKillList(idInt)
	fmt.Printf("%v\n", killInstructionIds)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(APIRules()))
}
