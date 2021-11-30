package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	client "github.com/influxdata/influxdb1-client"
)

// Returns a database connection
func getDBConnection() *client.Client {
	host, err := url.Parse(fmt.Sprintf("http://%s:%d", "localhost", 8086))
	check(err)
	conf := client.Config{
		URL: *host,
	}
	con, err := client.NewClient(conf)
	check(err)
	return con
}

// Checks if database 'data' is present
func databaseDataExists(con *client.Client) bool {
	q := client.Query{
		Command: "show databases",
	}
	response, err := con.Query(q)
	check(err)
	for _, v := range response.Results[0].Series[0].Values {
		if v[0] == "data" {
			return true
		}
	}
	return false
}

// create database data with retention policy 1h
func createDatabaseData1h(con *client.Client) {
	q := client.Query{
		Command: "CREATE DATABASE \"data\" WITH DURATION 1h REPLICATION 1",
	}
	_, err := con.Query(q)
	check(err)
}

// write transformed outputs from arduino to database
func writeLineToDatabase(con *client.Client, output map[string]interface{}) {
	pt := client.Point{
		Measurement: "outputs",
		Fields:      output,
		Time:        time.Now()}
	pts := []client.Point{pt}
	bp := client.BatchPoints{
		Points:          pts,
		Database:        "data",
		RetentionPolicy: "autogen", // pabandyti koreguoti.
	}
	_, err := con.Write(bp)
	if err != nil {
		log.Fatal(err)
	}
}
