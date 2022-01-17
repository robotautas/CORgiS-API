package main

import (
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

// CONSIDER REFACTORING USING PACKAGE LEVEL VARIABLES WITH MUTEXES INSTEAD

var pool = newPool()

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

func addToKillList(id int) {
	client := pool.Get()
	defer client.Close()
	_, err := client.Do("LPUSH", "killList", id)
	if err != nil {
		panic(err)
	}
}

func removeFromKillList(id int) {
	client := pool.Get()
	defer client.Close()
	_, err := client.Do("LREM", "killList", 1, id)
	if err != nil {
		panic(err)
	}
}

// checks if int value present in redis array. Name should be either 'killList' or "activeTaskIds"
func idInRedisArray(name string, id int) bool {
	client := pool.Get()
	defer client.Close()
	res, err := client.Do("LRANGE", name, 0, -1)

	if err != nil {
		panic(err)
	}

	list := res.([]interface{})
	for _, i := range list {
		s := string(i.([]byte))
		num, err := strconv.Atoi(s)

		if err != nil {
			panic(err)
		}

		if num == id {
			return true
		}
	}
	return false
}

func storeActiveTask(id int, task string) {
	client := pool.Get()
	defer client.Close()
	_, err := client.Do("LPUSH", "activeTaskIds", id)
	if err != nil {
		panic(err)
	}
	_, err = client.Do("JSON.SET", id, ".", task)
	if err != nil {
		panic(err)
	}
}

func removeActiveTask(id int) {
	client := pool.Get()
	defer client.Close()
	_, err := client.Do("DEL", id)
	if err != nil {
		panic(err)
	}
}

func readActiveTask(id int) string {
	client := pool.Get()
	defer client.Close()
	task, err := client.Do("JSON.GET", id)
	if err != nil {
		panic(err)
	}
	return string(task.([]uint8))
}

func getActiveTaskIds() []int {
	client := pool.Get()
	defer client.Close()
	res, err := client.Do("LRANGE", "activeTaskIds", 0, -1)
	if err != nil {
		panic(err)
	}
	interfaceList := res.([]interface{})
	var list []int
	for _, i := range interfaceList {
		s := string(i.([]byte))
		num, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
		list = append(list, num)
	}
	return list
}

func storeInstruction(id int) {
	client := pool.Get()
	defer client.Close()
	_, err := client.Do("LPUSH", "instructions", id)
	if err != nil {
		panic(err)
	}
}

func getTasksTimeInterval(id int) (time.Time, time.Time) {
	client := pool.Get()
	defer client.Close()
	s, err := client.Do("JSON.GET", id, ".start")
	if err != nil {
		panic(err)
	}
	f, err := client.Do("JSON.GET", id, ".finish")
	if err != nil {
		panic(err)
	}
	// ties šita vieta buvo problema, kuomet skundėsi kad interfeisas nil vietoje []uint8
	// deja atkartoti situacijos nepavyko.
	startString := string(s.([]uint8))[1 : len(string(s.([]uint8)))-1]
	finishString := string(f.([]uint8))[1 : len(string(f.([]uint8)))-1]
	startTime, err := time.Parse(time.RFC3339, startString)
	if err != nil {
		panic(err)
	}
	finishTime, err := time.Parse(time.RFC3339, finishString)
	if err != nil {
		panic(err)
	}
	return startTime, finishTime

}

func removeIdFromRedisArr(redisArr string, id int) {
	client := pool.Get()
	defer client.Close()
	_, err := client.Do("LREM", redisArr, 1, id)
	if err != nil {
		panic(err)
	}
}

//collect all active tasks, exclude requirements that are default, merge into one task
func getAllRunningTasksNonDefaultRequirements() {
	var activeTasks []Task
	for _, id := range getActiveTaskIds() {
		JSONById := readActiveTask(id)
		task := JSONToTask(JSONById)
		activeTasks = append(activeTasks, task)
	}
	for _, t := range activeTasks {
		printInfo("ATATATATA %v", t)
	}
}

func flushRedis() {
	client := pool.Get()
	defer client.Close()
	_, err := client.Do("FLUSHALL")
	if err != nil {
		panic(err)
	}
}
