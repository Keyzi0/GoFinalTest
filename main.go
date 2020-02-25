package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

//webCounter counts the "Go" string in the body of Get response
func webCounter(client *http.Client, url string) int32 {
	cnt := 0
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s \n", err.Error())
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		cnt = strings.Count(string(bodyBytes), `Go`)
	}
	return int32(cnt)
}

func main() {
	// fmt.Println(`Service started...`)
	var finalSum int32
	maxWorkersCnt := 5
	workerCount := 0

	scanner := bufio.NewScanner(os.Stdin)
	wgWorkers := new(sync.WaitGroup)
	chTask := make(chan string)

	for scanner.Scan() {
		if workerCount < maxWorkersCnt {
			wgWorkers.Add(1)
			go func() {
				defer wgWorkers.Done()
				var client http.Client
				var cnt int32
				for {
					url, status := <-chTask
					if !status {
						return
					}
					cnt = webCounter(&client, url)
					fmt.Printf("Count for %s: %d\n", url, cnt)
					atomic.AddInt32(&finalSum, int32(cnt))
				}
			}()
			workerCount++
		}
		chTask <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(os.Stderr, "reading standard input:", err)
	}
	close(chTask)
	wgWorkers.Wait()

	fmt.Println("Total:", finalSum)
}
