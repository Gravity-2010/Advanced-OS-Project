package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"math/rand"
	"os"
	"sort"
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

	start := time.Now()

	jobsCh := make(chan Job, M)
	resultsCh := make(chan Result, M)
	totalCh := make(chan int)

	jobCounts := make([]int, M)

	var wg sync.WaitGroup
	wg.Add(M)

	go dispatcher(pathname, N, jobsCh)

	for i := 0; i < M; i++ {
		go worker(i, C, jobsCh, resultsCh, &wg, jobCounts)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	go consolidator(resultsCh, totalCh)

	totalPrimes := <-totalCh
	elapsed := time.Since(start)

	// Statistics
	min, max := jobCounts[0], jobCounts[0]
	for _, count := range jobCounts {
		if count < min {
			min = count
		}
		if count > max {
			max = count
		}
	}

	total := 0
	for _, count := range jobCounts {
		total += count
	}
	average := float64(total) / float64(M)

	sorted := make([]int, M)
	copy(sorted, jobCounts)
	sort.Ints(sorted)
	var median float64
	if M%2 == 0 {
		median = float64(sorted[M/2-1]+sorted[M/2]) / 2.0
	} else {
		median = float64(sorted[M/2])
	}

	fmt.Printf("Total prime numbers found: %d\n", totalPrimes)
	fmt.Printf("Min jobs completed by a worker: %d\n", min)
	fmt.Printf("Max jobs completed by a worker: %d\n", max)
	fmt.Printf("Average jobs completed by a worker: %.1f\n", average)
	fmt.Printf("Median jobs completed by a worker: %.1f\n", median)
	fmt.Printf("Elapsed time (ms): %d ms\n", elapsed.Milliseconds())
}

func isPrime(n uint64) bool {
	if n < 2 {
		return false
	}
	bigN := new(big.Int).SetUint64(n)
	return bigN.ProbablyPrime(20)
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
	defer close(jobsCh)

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
}

func worker(id int, C int, jobsCh <-chan Job, resultsCh chan<- Result, wg *sync.WaitGroup, jobCounts []int) {
	defer wg.Done()
	time.Sleep(time.Duration(400+rand.Intn(200)) * time.Millisecond)
	// fmt.Printf("Worker %d started\n", id)
	for job := range jobsCh {
		file, err := os.Open(job.Pathname)
		// fmt.Printf("Worker %d opened file, processing job start=%d length=%d\n", id, job.Start, job.Length)
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
		jobCounts[id]++

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

// Consolidating the results
func consolidator(resultCh <-chan Result, totalCh chan<- int) {
	totalPrimes := 0
	for result := range resultCh {
		totalPrimes += result.Primecount
	}
	totalCh <- totalPrimes
}
