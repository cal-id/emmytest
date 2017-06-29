# Benchmarking emmy on a Raspberry PI

## Introduction

This directory contains the resources to benchmark protocols of the `emmy` library on a Raspberry PI. The current protocols are `ZKP_schnorr`, `dlog_equality` and `dlog_equality_blinded_transcript`.


### Layout of this directory
For each protocol, there is a container directory eg `ZKP_schnorr/`.

There is also an additional directory (`utils`) which contains a go library to generate the dlog groups at different key sizes.

Within each container directory:
- `benchmarking.go` is a go script that is compiled into an executable program that will run the given protocol.
    - It input parameters including key size.
    - Each execution will return a line of 'csv' output with the input parameters and the time it took to run.
    - These are compiled and added to $PATH in the 'Installation' section below.
- `basic_test.sh` is a bash script that iterates over all key sizes, executing the above protocol and combining the results into 'csv' in stdout.
- `outputs_and_analysis` is a directory containing:
    - `basic_outputX.csv` is the output from running `./basic_test.sh`. There are three because three repeats were done for each test.
    - `analysis.X` is a mathematica notebook plotting a graph for each repeat of the results.

## Specific Setup

### Emmy status

These tests were run on emmy at commit `dedcd8c`.


### Kit

The PI used is a Raspberry PI Model 3 B V1.2. This is (the one)[https://www.element14.com/community/community/raspberry-pi/blog/2016/11/21/how-to-identify-which-model-of-the-raspberry-pi-you-have] with WiFi and Bluetooth.

## Installation

These are the commands to:
- Install go
- Install emmy and revert to commit `dedcd8c`
- Install the benchmarks

```bash
cd ~
wget https://storage.googleapis.com/golang/go1.8.3.linux-armv6l.tar.gz  # download go for arm
tar -C /usr/local -xzf go1.8.3.linux-armv6l.tar.gz                      # extract to /usr/local

# Setup go path
export GOPATH="$HOME/go"
export PATH="$PATH:/usr/local/go/bin:$GOPATH/bin"

# Get + Install emmy at the correct commit
go get github.com/xlab-si/emmy
cd $GOPATH/src/github.com/xlab-si/emmy
git checkout dedcd8c
go get .

# Download and install this repo
mkdir $GOPATH/src/github.com/cal-id
cd $GOPATH/src/github.com/cal-id
git clone https://github.com/cal-id/emmytest
cd $GOPATH/src/github.com/cal-id/emmytest/benchmarking
for f in dlog_equality/ dlog_equality_blinded_transcript/ ZKP_schnorr/; do 
    cd $f
    go install .
    cd ..
done
```

To verify they installed, each of these commands should return a line of numbers ($GOPATH/bin is in $PATH).

``` bash
dlog_equality
dlog_equality_blinded_transcript/
ZKP_schnorr/
```

## Creating the output files
