package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	pathname := os.Args[1]

	M := 1
	N := 65536
	C := 1024

	var err error
	if len(os.Args) > 2 {
		M, err = strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Error converting M: %v\n", err)
			return
		}
	}
	if len(os.Args) > 3 {
		N, err = strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("Error converting N: %v\n", err)
			return
		}
	}
	if len(os.Args) > 4 {
		C, err = strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Printf("Error converting C: %v\n", err)
			return
		}
	}
	fmt.Printf("pathname: %s, M: %d, N: %d, C: %d\n", pathname, M, N, C)

	// fmt.Println(isPrime(7))
	// fmt.Println(isPrime(10))
	// fmt.Println(isPrime(2))
	// fmt.Println(isPrime(1))

	jobsCh := make(chan Job, M)
	resultsCh := make(chan Result, M)
	totalCh := make(chan int)

	var wg sync.WaitGroup
	wg.Add(M)

	go dispatcher(pathname, N, jobsCh)

	for i := 0; i < M; i++ {
		go worker(i, C, jobsCh, resultsCh, &wg)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	go consolidator(resultsCh, totalCh)

	totalPrimes := <-totalCh
	fmt.Printf("Total prime numbers found: %d\n", totalPrimes)
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
		file, err := os.Open(job.Pathname)
		if err != nil {
			fmt.Printf("Worker %d: Error opening file: %v\n", id, err)
			continue
		}

		_, err = file.Seek(job.Start, 0)
		if err != nil {
			fmt.Printf("Worker %d: Error seeking file: %v\n", id, err)
			file.Close()
			continue
		}

		// Creating a buffer of size C and reading from the file
		buffer := make([]byte, C)
		primeCount := 0
		bytesProcessed := int64(0)
		for {
			if bytesProcessed >= job.Length {
				break
			}
			length := int64(C)
			if bytesProcessed+length > job.Length {
				length = job.Length - bytesProcessed
			}
			bytesRead, err := file.Read(buffer[:length])
			if err != nil {
				fmt.Printf("Worker %d: Error reading file: %v\n", id, err)
				break
			}
			if bytesRead == 0 {
				break
			}

			reader := bytes.NewReader(buffer[:bytesRead])

			var num uint64
			for {
				err = binary.Read(reader, binary.LittleEndian, &num)
				if err == io.EOF {
					break
				}
				if err != nil {
					fmt.Printf("Worker %d: Error reading number from buffer: %v\n", id, err)
					break
				}
				if isPrime(num) {
					primeCount++
				}
			}

			bytesProcessed += int64(bytesRead)
		}

		file.Close()

		// Send the result back to the results channel
		resultsCh <- Result{
			Job:        job,
			Primecount: primeCount,
		}

		slog.Info("Job completed",
			"worker_id", id,
			"pathname", job.Pathname,
			"start", job.Start,
			"length", job.Length,
			"prime_count", primeCount,
		)
	}
}

// COnsolidating the results
func consolidator(resultCh <-chan Result, totalCh chan<- int) {
	totalPrimes := 0
	for result := range resultCh {
		totalPrimes += result.Primecount
	}
	totalCh <- totalPrimes
}
