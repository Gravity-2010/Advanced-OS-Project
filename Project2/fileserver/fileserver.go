package main

import (
	"io"
	"log"
	"net"
	"os"
	"strings"

	"primes_grpc/pb"

	"google.golang.org/grpc"
)

type FileServerImpl struct {
	pb.UnimplementedFileServerServer
	file *os.File
}

func (s *FileServerImpl) FetchSegment(req *pb.FetchRequest, stream pb.FileServer_FetchSegmentServer) error {

	remaining := req.SegmentSize
	buf := make([]byte, req.ChunkSize)
	offset := req.ByteOffset

	for remaining > 0 {
		toRead := min(remaining, int64(req.ChunkSize))
		n, err := s.file.ReadAt(buf[:toRead], offset)

		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		stream.Send(&pb.Chunk{Data: buf[:n]})
		offset += int64(n)
		remaining -= int64(n)
	}

	return nil
}

func main() {
	dataPath := os.Args[1]
	cfgPath := os.Args[2]

	cfg := parseConfig(cfgPath)

	f, err := os.Open(dataPath)
	if err != nil {
		log.Fatalf("cannot open data file: %v", err)
	}
	defer f.Close()

	lis, err := net.Listen("tcp", cfg.FileServerAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFileServerServer(grpcServer, &FileServerImpl{file: f})

	log.Printf("Fileserver listening on %s", cfg.FileServerAddr)
	grpcServer.Serve(lis)
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
