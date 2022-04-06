package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	avg            int
	warningMessage string
)

func doHTTP(url string) int {
	client := &http.Client{
		Timeout: time.Millisecond * 500,
	}
	resp, err := client.Get(url)
	if os.IsTimeout(err) {
		return -1
	} else {
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		num, _ := strconv.Atoi(string(data))
		return num
	}
}

func getAvg(sl []int) (int, string) {
	warning := ""
	sum := 0
	length := len(sl)
	for i, v := range sl {
		if v == -1 {
			length--
			warning += fmt.Sprintf("%d ", i+1)
			continue
		}
		sum += v
	}
	if length == 0 {
		return 0, "All sensors unreleasable"
	}
	if warning != "" {
		warning = "Unreachable sensors: " + warning
	} else {
		warning = "All data received without issues"
	}
	return sum / length, warning
}

func main() {
	urls := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
		"http://localhost:8084",
	}
	signals := make([]int, len(urls))

	go func() {
		for {
			for i, url := range urls {
				signals[i] = doHTTP(url)
			}
			avg, warningMessage = getAvg(signals)
			time.Sleep(time.Second * 40)
		}
	}()

	err := http.ListenAndServe(":3000", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := fmt.Fprintf(writer, "Avg: %d\nSensors: %d\n%s", avg, signals, warningMessage)
		if err != nil {
			fmt.Println(err)
		}
	}))

	if err != nil {
		fmt.Println(err)
	}
}
