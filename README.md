# Advanced Operating Systems Project

**High-Performance Prime Number Counter using Concurrency and Distributed Systems**

This repository contains two implementations developed for the **Advanced Operating Systems** course. Both projects efficiently count prime numbers (stored as uint64) in large binary files, focusing on performance, scalability, and systems programming concepts in Go.

## 📋 Table of Contents
- [Project 1: Concurrent Prime Counter](#project-1-concurrent-prime-counter)
- [Project 2: Distributed Prime Counter](#project-2-distributed-prime-counter)
- [Technologies](#technologies)
- [Key Learnings](#key-learnings)
- [Repository Structure](#repository-structure)

## Project 1: Concurrent Prime Counter

Multi-threaded implementation using the Producer-Consumer pattern with Goroutines.

### Features
- Dispatcher divides the input binary file into equal segments
- Multiple worker goroutines process segments concurrently
- Optimized chunk-based file reading
- Consolidator aggregates results with detailed statistics
- Performance metrics (execution time, min/max/avg/median jobs per worker)

### Usage

```bash
# Basic usage (default parameters)
go run Project1/main.go <datafile.dat>

# With custom parameters: M (workers), N (segment size), C (chunk size)
go run Project1/main.go <datafile.dat> <M> <N> <C>

# Example
go run Project1/main.go data1gb.dat 8 65536 4096
```

**Parameters:**
- `M` → Number of workers (default: 1)
- `N` → Segment size in bytes (default: 65536 → 64KB)
- `C` → Chunk size for reading (default: 1024 → 1KB)

## Project 2: Distributed Prime Counter

Fully distributed system using **gRPC** for communication between independent services.

### Architecture
- **File Server** – Serves binary data chunks on request
- **Dispatcher** – Divides work and distributes segments via gRPC
- **Workers** – Independent processes that pull jobs, process data, and return results
- **Consolidator** – Collects results and generates final statistics

### Usage

```bash
cd Project2

# Build all binaries
go build -o main_bin .
go build -o fileserver_bin ./fileserver
go build -o worker_bin ./worker

# Run the complete system (recommended)
./run.sh <N> <C> <datafile.dat> primes_config.txt
```

**Example:**
```bash
./run.sh 65536 4096 data1gb.dat primes_config.txt
```

**Configuration File** (`primes_config.txt`):
```txt
dispatcher   localhost 5001
consolidator localhost 5002
fileserver   localhost 5003
```

## Technologies

- **Language**: Go (Golang)
- **Concurrency**: Goroutines, Channels, Mutexes, WaitGroups
- **Distributed Systems**: gRPC, Protocol Buffers
- **I/O Optimization**: Chunked binary reading
- **Build & Orchestration**: Shell scripting

## Key Learnings & Highlights

- Practical implementation of Producer-Consumer and Worker Pool patterns
- Deep understanding of synchronization and load balancing
- Design of scalable distributed systems using microservices architecture
- Performance profiling and optimization of file I/O in concurrent & distributed environments
- Real-world application of gRPC for inter-process communication

## Repository Structure

```
Advanced-OS-Project/
├── Project1/
│   ├── main.go
│   ├── Project_Report.pdf
│   └── sample_output.txt
├── Project2/
│   ├── main.go
│   ├── fileserver/
│   ├── worker/
│   ├── pb/                  # Generated protobuf files
│   ├── primes.proto
│   ├── run.sh
│   ├── Project2_Report.pdf
│   └── sample_output.txt
├── Project1_os.zip
├── Project2.zip
└── README.md
```

---

**Built for Advanced Operating Systems Course**

Feel free to explore both implementations and compare their performance!
