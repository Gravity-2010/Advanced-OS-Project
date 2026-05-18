#!/bin/bash
N=$1
C=$2
M=$3
DATA=$4
CONFIG=$5

echo "Starting fileserver..."
./fileserver_bin $DATA $CONFIG &
FILESERVER_PID=$!
sleep 1

echo "Starting main (N=$N C=$C)..."
./main_bin $N $C $DATA $CONFIG &
sleep 1

echo "Starting $M workers (C=$C)..."
for i in $(seq 1 $M); do
    ./worker_bin $C $CONFIG &
done

wait
kill $FILESERVER_PID 2>/dev/null
echo "Done"