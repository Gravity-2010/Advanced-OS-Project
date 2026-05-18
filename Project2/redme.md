Prime Counter - Distributed OS Project 2

INSTRUCTIONS
To build:
    go build -o main_bin .
    go build -o fileserver_bin ./fileserver/
    go build -o worker_bin ./worker/

    Or use the run script directly with go run.

To run everything:
    ./run.sh <N> <C> <M> <datafile> <configfile>

Example:
    ./run.sh 65536 4096 16 data1gb.dat primes_config.txt

To generate a sample datafile:
    head -c 1M /dev/urandom > test.dat
    head -c 1G /dev/urandom > data1gb.dat

CONFIG FILE (primes_config.txt)
dispatcher   localhost 5001
consolidator localhost 5002
fileserver   localhost 5003

Note: N is segment size in bytes, C is chunk size in bytes, M is number of workers.
      Start order: fileserver first, then main, then workers.
      Fileserver stops automatically when run.sh completes.