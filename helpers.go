package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// formats multi command string from provided map
func formatMultiCommand(m map[string]interface{}) string {
	commandString := "<"
	for k, v := range m {
		if strings.HasPrefix(k, "V") ||
			strings.HasPrefix(k, "T") {
			value := strconv.Itoa(v.(int))
			command := "SET_" + k + "=" + value + ";"
			commandString += command
		} else if strings.HasPrefix(k, "PUMP") {
			value := v.(string)
			command := k + "_" + value + ";"
			commandString += command
		}
	}
	return commandString + ">"
}

// Validates raw arduino output against regex pattern and few other conditions.
func outputIsValid(s string, re *regexp.Regexp) bool {
	if len(s) > 168 && len(s) < 215 {
		if s[:4] == "V00=" &&
			strings.HasSuffix(s, ";") &&
			len(re.FindAll([]byte(s), -1)) >= 28 {
			// log.Output(1, s)
			return true
		}
	}
	return false
}

// Transforms output like "V00=0;V01=0;V02=0; ... S01=00;PUMP=0;" to map, for convenient writing to influxdb.
func outputToMap(s string) map[string]interface{} {
	// println("STRINGAS: ", s)
	res := make(map[string]interface{})
	splitted_s := strings.Split(s, ";")
	for _, i := range splitted_s[:len(splitted_s)-1] {
		splitted_i := strings.Split(i, "=")
		// hex -> dec
		if strings.HasPrefix(i, "V") {
			num, err := strconv.ParseInt(splitted_i[1], 16, 64)
			check(err)
			res[splitted_i[0]] = num
		} else {
			number, err := strconv.ParseInt(splitted_i[1], 10, 64)
			check(err)
			res[splitted_i[0]] = number
		}
	}
	return res
}

// checks if string is in slice
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// validates provided url value
func URLValueValid(p string, v string) bool {
	if len(v) > 0 {
		valueToInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return false
		}
		if stringInSlice(p, VxxParams) && valueToInt >= 0 && valueToInt < 256 {
			return true
		} else if stringInSlice(p, TxxParams) && valueToInt >= 0 && valueToInt < 1000 {
			return true
		}
	} else if stringInSlice(p, pumpParams) {
		return true
	}
	return false
}

// validates provided URL param
func URLParamValid(s string) bool {
	if stringInSlice(s, VxxParams) ||
		stringInSlice(s, TxxParams) ||
		stringInSlice(s, pumpParams) ||
		s == "Sleep" {
		return true
	}
	return false
}

// validates map made from POST request JSON STRING (bussiness logic)
func validateJSONTasks(t []Task) bool {
	for _, task := range t {
		for p, v := range task.Vxx {
			if !stringInSlice(p, VxxParams) {
				log.Output(1, fmt.Sprintf("JSON validation failed, parameter %v is not valid", p))
				return false
			}
			for _, r := range v {
				if r[0] < 0 || r[0] > 7 {
					err := fmt.Sprintf("JSON validation failed. Combination %v is invalid, first value must be in range 0-7.", r)
					log.Output(1, err)
					return false
				} else if r[1] < 0 || r[1] > 1 {
					err := fmt.Sprintf("JSON validation failed. Combination %v is invalid, second value must be 0 or 1.", r)
					log.Output(1, err)
					return false
				}
			}
		}
		for p, v := range task.Txx {
			if !stringInSlice(p, TxxParams) {
				log.Output(1, fmt.Sprintf("JSON validation failed, parameter %v is not valid", p))
				return false
			} else if v <= 0 || v >= 1000 {
				err := fmt.Sprintf("JSON validation failed. Temperature value %v is out of range (0-999)", v)
				log.Output(1, err)
				return false
			}
		}
	}
	return true
}

// validates if key-value pair suitable for arduino command (helper for validateJSONMap())
func JSONValueValid(p string, v int) bool {
	if stringInSlice(p, VxxParams) && v >= 0 && v < 256 {
		return true
	} else if stringInSlice(p, TxxParams) && v >= 0 && v < 1000 {
		return true
	} else if stringInSlice(p, pumpParams) {
		return true
	}
	return false
}

// get slice of active tasks ids, whose time interval overlaps with the given task
func (t *Task) overlappingTasks() []int {
	tStart := t.StartTime
	tFinish := t.FinishTime
	var ids []int
	for _, id := range getActiveTaskIds() {
		idStart, idFinish := getTasksTimeInterval(id)
		if timeIntervalsOverlap(tStart, tFinish, idStart, idFinish) {
			ids = append(ids, id)
		}
	}
	return ids
}

// checks if time interval a1 -> a2 overlaps with b1 -> b2
func timeIntervalsOverlap(a1, a2, b1, b2 time.Time) bool {
	if a1.Before(b1) && a2.Before(b1) {
		return false
	} else if a1.After(b2) && a2.After(b2) {
		return false
	}
	return true
}

// determines if task's excecution won't affect other active tasks
// ids are the ids of active tasks, whose time intervals overlap with t
func (t *Task) conflictsWith(ids []int) bool {
	for _, id := range ids {
		JSONById := readActiveTask(id)
		comparedTask := JSONToTask(JSONById)
		for kT, vT := range t.Vxx {
			for kC, vC := range comparedTask.Vxx {
				if kT == kC {
					for _, reqT := range vT {
						for _, reqC := range vC {
							if reqT[0] == reqC[0] {
								if reqT[1] != reqC[1] {
									printWarning("Conflict detected: task %d uses %s: %v, which conflicts with requested %s: %v", id, kC, reqC, kT, reqT)
									return true
								}
							}
						}
					}
				}
			}
		}
	}
	return false
}

// check if requirement is used in active and running tasks, doesn't matter conflicting or not
// e.g. see if V00 value {{1, 1}, {2, 1}, {3, 1}} has requirement {7, 1} in it
// similar to conflictsWith function, except it searches for other usages of boad channel,
// no matter if conflicting. Used in task stopping.
func requirementUsedElsewhere(param string, req [2]int, instrID int) bool {
	for _, id := range getActiveTaskIds() {
		JSONById := readActiveTask(id)
		comparedTask := JSONToTask(JSONById)
		// make sure we're not checking future tasks
		// don't checck tasks from the same instruction, cause time.now might be milliseconds ahead.
		if time.Now().Local().After(comparedTask.StartTime) &&
			comparedTask.InstructionId != instrID {
			for k, v := range comparedTask.Vxx {
				if k == param {
					for _, i := range v {
						if i[0] == req[0] {
							printError("req %v is used in task %v, where %v=%v", req, comparedTask.Id, k, v)
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func check(err error) {
	if err != nil {
		color.Set(color.FgRed)
		panic(err.Error())
		color.Unset()
	}
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func intInSlice(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// remove requirement like {3, 0} from requirements array, like {{1, 1}, {3, 0}}
func removeRequirement(reqs [][2]int, req [2]int) [][2]int {
	for idx, item := range reqs {
		if item == req {
			return append(reqs[:idx], reqs[idx+1:]...)
		}
	}
	return nil
}

func APIRules() string {
	text := fmt.Sprintf(`This is an API part of middleware between graphitizer microcontroller and user interface.
API sends commands to microcontroller through HTTP GET or POST requests.

SET request examples:
http://127.0.0.1:9999/set?param=V00&value=255
http://127.0.0.1:9999/set?param=T01&value=80
http://127.0.0.1:9999/set?param=PUMP_OFF

GET_ALL:
http://127.0.0.1:9999/getall

START:
Expects a set of instructions in JSON from a post request to /start endpoint like

%v

Vxx, Txx and PUMP parameters are .
responds with unique id of a process, for pausing/stopping it later.

STOP:
stops a process with unique id, provided in URL
http://127.0.0.1:9999/set?id=2345


`, exampleJSON)
	return text
}
