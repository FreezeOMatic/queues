package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	const queueBufferLength = 10
	args := os.Args
	if len(args) != 2 {
		fmt.Println("please enter the port as argument")
		return
	}

	queues := make(map[string]chan string)

	port := ":" + args[1]
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:
			//localhost:8000/color?timeout=10
			var timeout int
			queueName := strings.TrimPrefix(r.URL.String(), "/")

			if r.URL.Query().Has("timeout") {
				tStr := r.URL.Query().Get("timeout")
				t, err := strconv.Atoi(tStr)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
				}
				timeout = t
				queueName = queueName[0:strings.Index(queueName, "?")]
			}

			queueChan := getQueueChan(queues, queueName, queueBufferLength)
			var val string
			if timeout != 0 {
				select {
				case val = <-queueChan:
					w.Write([]byte(val))
				case <-time.After(time.Duration(timeout) * time.Second):
					w.WriteHeader(http.StatusNotFound)
				}
			} else {
				if len(queueChan) == 0 {
					w.WriteHeader(http.StatusNotFound)
				} else {
					val = <-queueChan
					w.Write([]byte(val))
				}
			}

		case http.MethodPut:
			//localhost:8000/color?v=red&v=blue
			if !r.URL.Query().Has("v") {
				w.WriteHeader(http.StatusBadRequest)
			}

			queueName := strings.TrimPrefix(r.URL.String(), "/")
			queueName = queueName[0:strings.Index(queueName, "?")]

			queueChan := getQueueChan(queues, queueName, queueBufferLength)

			for k, v := range r.URL.Query() {
				if k == "v" {
					for _, value := range v {
						queueChan <- value
					}
				}
			}
			fmt.Println(len(queueChan))
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusBadRequest) // unsupported method
		}
	})
	err := http.ListenAndServe(port, nil)

	fmt.Println("Error while start listening port:", port, err)

}

func getQueueChan(queues map[string]chan string, queueName string, queueBufferLenght int) (ch chan string) {
	_, ok := queues[queueName]
	if ok {
		ch = queues[queueName]
		return
	} else {
		queues[queueName] = make(chan string, queueBufferLenght)
		ch = queues[queueName]
		return
	}
}
