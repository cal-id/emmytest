#! /bin/bash

# Runs 64 benchmarks at different key sizes
# and prints csv to standardout

# Echo the headers for the csv, these are
# determined by the order the variables
# are printed in the go code.
echo "N, L, Time Proof (ns), Time Verify (ns), Q, P, G"

for L in $(seq 128 128 1024); do
    for N in $(seq 16 16 256); do
        if [ $L -gt $N ]; then
            dlog_equality_blinded_transcript -N $N -L $L
        fi
    done
done
