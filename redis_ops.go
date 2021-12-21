package main

import "github.com/gomodule/redigo/redis"

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

func idInKillList(id int) bool {
	client := pool.Get()
	defer client.Close()
	res, err := client.Do("LRANGE", "killList", 0, -1)
	if err != nil {
		panic(err)
	}
	list := res.([]int)
	for _, i := range list {
		println(i)
	}
	return true
}

func storeActiveTask(id int, task string) {
	client := pool.Get()
	defer client.Close()
	_, err := client.Do("SET", id, task)
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
	task, err := client.Do("GET", id)
	if err != nil {
		panic(err)
	}
	return string(task.([]uint8))
}
