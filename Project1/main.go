package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	pathname := os.Args[1]

	var err error
	M, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Error converting M to integer: %v\n", err)
		return
	}

	N, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Printf("Error converting N to integer: %v\n", err)
		return
	}

	C, err := strconv.Atoi(os.Args[4])
	if err != nil {
		fmt.Printf("Error converting C to integer: %v\n", err)
		return
	}

	fmt.Printf("pathname: %s, M: %s, N: %s, C: %s\n", pathname, M, N, C)

	fmt.Println(isPrime(7))
	fmt.Println(isPrime(10))
	fmt.Println(isPrime(2))
	fmt.Println(isPrime(1))
}

func isPrime(n uint64) bool {
	if n < 2 {
		return false
	}
	for i := uint64(2); i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

type Job struct {
	Pathname string
	Start    int64
	Length   int64
}

type Result struct {
	Job        Job
	Primecount int
}

func dispatcher(pathname string, L int, jobsCh chan<- Job) {
	fileInfo, err := os.Stat(pathname)
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return
	}

	fileSize := fileInfo.Size()

	for start := int64(0); start < fileSize; start += int64(L) {
		length := int64(L)
		if start+length > fileSize {
			length = fileSize - start
		}
		job := Job{
			Pathname: pathname,
			Start:    start,
			Length:   length,
		}
		jobsCh <- job
	}
	close(jobsCh)
}

func worker(id int, C int, jobsCh <-chan Job, resultsCh chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Duration(400+rand.Intn(200)) * time.Millisecond)
	for job := range jobsCh {
		// Simulate processing the job
	}
}
