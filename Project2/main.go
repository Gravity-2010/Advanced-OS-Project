package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"primes_grpc/pb"

	"google.golang.org/grpc"
)

// DispatcherServer
type DispatcherServer struct {
	pb.UnimplementedDispatcherServer
	jobsCh chan *pb.JobResponse
}

func (s *DispatcherServer) PullJob(ctx context.Context, req *pb.PullRequest) (*pb.JobResponse, error) {
	job, ok := <-s.jobsCh
	if !ok {
		return &pb.JobResponse{Done: true}, nil
	}
	return job, nil
}

func (s *DispatcherServer) start(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Dispatcher listening failed: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDispatcherServer(grpcServer, s)

	log.Printf("Dispatcher listening on %s", addr)
	grpcServer.Serve(lis)
}

// ConsolidatorServer
type ConsolidatorServer struct {
	pb.UnimplementedConsolidatorServer
	mu          sync.Mutex
	totalPrimes int64
	received    int
	expected    int
	jobQueue    chan *pb.JobResponse
	done        chan struct{}
	jobCounts   []int32
}

func (s *ConsolidatorServer) PushResult(ctx context.Context, req *pb.ResultRequest) (*pb.ResultAck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalPrimes += req.PrimeCount
	s.received++
	s.jobCounts = append(s.jobCounts, req.JobsDone)

	if s.received == s.expected {
		close(s.jobQueue)
		close(s.done)
	}
	return &pb.ResultAck{}, nil
}

func (s *ConsolidatorServer) start(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Consolidator listening failed: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterConsolidatorServer(grpcServer, s)

	log.Printf("Consolidator listening on %s", addr)
	grpcServer.Serve(lis)
}

func main() {
	N, _ := strconv.ParseInt(os.Args[1], 10, 64)
	C, _ := strconv.ParseInt(os.Args[2], 10, 64)
	dataPath := os.Args[3]
	cfgPath := os.Args[4]
	cfg := parseConfig(cfgPath)

	fmt.Printf("N: %d, C: %d, datafile: %s\n", N, C, dataPath)

	fileInfo, err := os.Stat(dataPath)
	if err != nil {
		log.Fatalf("Failed to stat data file: %v", err)
	}
	fileSize := fileInfo.Size()
	numSegments := int((fileSize + N - 1) / N)

	jobQueue := make(chan *pb.JobResponse, numSegments)
	done := make(chan struct{})

	for i := 0; i < numSegments; i++ {
		offset := int64(i) * N
		size := N
		if offset+size > fileSize {
			size = fileSize - offset
		}
		jobQueue <- &pb.JobResponse{
			Done:        false,
			SegmentId:   int64(i),
			ByteOffset:  offset,
			SegmentSize: size,
		}
	}

	dispServer := &DispatcherServer{jobsCh: jobQueue}
	consServer := &ConsolidatorServer{
		expected: numSegments,
		jobQueue: jobQueue,
		done:     done,
	}

	go dispServer.start(cfg.DispatcherAddr)
	go consServer.start(cfg.ConsolidatorAddr)

	start := time.Now()

	<-done

	elapsed := time.Since(start)
	fmt.Printf("Total primes: %d\n", consServer.totalPrimes)
	fmt.Printf("Elapsed time: %d ms\n", elapsed.Milliseconds())

	counts := consServer.jobCounts
	sort.Slice(counts, func(i, j int) bool { return counts[i] < counts[j] })

	min := counts[0]
	max := counts[len(counts)-1]

	sum := int32(0)
	for _, c := range counts {
		sum += c
	}
	avg := float64(sum) / float64(len(counts))

	n := len(counts)
	var median float64
	if n%2 == 0 {
		median = float64(counts[n/2-1]+counts[n/2]) / 2.0
	} else {
		median = float64(counts[n/2])
	}

	fmt.Printf("Min jobs: %d\n", min)
	fmt.Printf("Max jobs: %d\n", max)
	fmt.Printf("Avg jobs: %.1f\n", avg)
	fmt.Printf("Median jobs: %.1f\n", median)
}

type Config struct {
	DispatcherAddr   string
	ConsolidatorAddr string
	FileServerAddr   string
}

func parseConfig(path string) Config {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		switch parts[0] {
		case "dispatcher":
			cfg.DispatcherAddr = parts[1] + ":" + parts[2]
		case "consolidator":
			cfg.ConsolidatorAddr = parts[1] + ":" + parts[2]
		case "fileserver":
			cfg.FileServerAddr = parts[1] + ":" + parts[2]
		}
	}
	return cfg
}
