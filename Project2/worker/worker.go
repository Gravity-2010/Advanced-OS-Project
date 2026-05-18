package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"primes_grpc/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func isPrime(n uint64) bool {
	if n < 2 {
		return false
	}
	bigN := new(big.Int).SetUint64(n)
	return bigN.ProbablyPrime(20)
}

func main() {
	C, _ := strconv.ParseInt(os.Args[1], 10, 64)
	cfgPath := os.Args[2]
	cfg := parseConfig(cfgPath)
	workerId := int32(os.Getpid())

	dispConn, _ := grpc.Dial(cfg.DispatcherAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	fileConn, _ := grpc.Dial(cfg.FileServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	consConn, _ := grpc.Dial(cfg.ConsolidatorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer dispConn.Close()
	defer fileConn.Close()
	defer consConn.Close()

	dispClient := pb.NewDispatcherClient(dispConn)
	fileClient := pb.NewFileServerClient(fileConn)
	consClient := pb.NewConsolidatorClient(consConn)

	jobsDone := int32(0)

	for {
		// ① PullJob with retry (max 10 attempts)
		var job *pb.JobResponse
		retries := 0
		for {
			var err error
			job, err = dispClient.PullJob(context.Background(), &pb.PullRequest{})
			if err == nil {
				break
			}
			retries++
			if retries > 10 {
				log.Printf("Dispatcher unavailable, exiting")
				return
			}
			log.Printf("PullJob retrying: %v", err)
			time.Sleep(500 * time.Millisecond)
		}

		// Check if no more jobs
		if job.Done {
			break
		}

		// ② FetchSegment with retry + count primes
		primeCount := int64(0)
		var stream pb.FileServer_FetchSegmentClient
		for {
			var err error
			stream, err = fileClient.FetchSegment(context.Background(), &pb.FetchRequest{
				ByteOffset:  job.ByteOffset,
				SegmentSize: job.SegmentSize,
				ChunkSize:   C,
			})
			if err == nil {
				break
			}
			log.Printf("FetchSegment retrying: %v", err)
			time.Sleep(500 * time.Millisecond)
		}

		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("stream.Recv error: %v", err)
			}
			reader := bytes.NewReader(chunk.Data)
			var num uint64
			for {
				err := binary.Read(reader, binary.LittleEndian, &num)
				if err != nil {
					break
				}
				if isPrime(num) {
					primeCount++
				}
			}
		}

		// ③ PushResult
		jobsDone++
		consClient.PushResult(context.Background(), &pb.ResultRequest{
			SegmentId:  job.SegmentId,
			PrimeCount: primeCount,
			JobsDone:   jobsDone,
			WorkerId:   workerId,
		})
	}
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
