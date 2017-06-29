# Benchmarking emmy on a Raspberry PI

## Introduction

This directory contains the resources to benchmark protocols of the `emmy` library on a Raspberry PI. The current protocols are `ZKP_schnorr`, `dlog_equality` and `dlog_equality_blinded_transcript`.


### Layout
For each protocol, there is a container directory eg `ZKP_schnorr/`.

There is also an additional directory (`utils/`) which contains a go library to generate the dlog groups at different key sizes.

Within each container directory:
- `benchmarking.go` is a go script that is compiled into an executable program that will run the given protocol.
    - e.g. `ZKP_schnorr -N 16 -L 128` records the time for a single run of the ZKP_schnorr protocol with key size (N=16, L=128)
    - It takes cli input parameters including key size.
    - Each execution will return a line of 'csv' output with the input parameters and the time it took to run.
    - These are compiled and added to $PATH in the 'Installation' section below.
- `basic_test.sh` is a bash script that iterates over all key sizes, executing the above protocol and combining the results into 'csv' in stdout.
- `outputs_and_analysis/basic_output[1,2,3].csv` is the output from running `./basic_test.sh`. There are three because three repeats were done for each test.
- `outputs_and_analysis/analysis.[nb,cdf,pdf]` is a mathematica notebook plotting a graph for each repeat of the results.

## Specific Setup

### Emmy status

These tests were run on emmy at commit `dedcd8c`.


### Kit

The PI used is a Raspberry PI Model 3 B V1.2. This is [the one](https://www.element14.com/community/community/raspberry-pi/blog/2016/11/21/how-to-identify-which-model-of-the-raspberry-pi-you-have) with WiFi and Bluetooth.

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

To verify they installed correctly, each of these commands should return a line of numbers (provided $GOPATH/bin is in $PATH).

``` bash
$ dlog_equality
8, 16, 216039, 181, 49957, 23458
$ dlog_equality_blinded_transcript
8, 16, 415412, 678951, 181, 38011, 35532
$ ZKP_schnorr
3, 8, 16, 9106837, 229, 60457, 45206
```

## Method Explaination

### Key Sizes
Each protocol takes a struct (`github.com/xlab-si/emmy/dlog.ZpDLog`) which contains the group parameters (`p`, `q`, `g`). See [this wikipedia page](https://en.wikipedia.org/wiki/Digital_Signature_Algorithm) for an explaination. `g` is just a generator of the group so it is of the same order of magnitude as `p`.

This means there are two 'key size' parameters:
- N = number of bits in `q`
- L = number of bits in `p`

As `p` must be a multiple of `q`, only `L > N` is tested.

### Generating the groups
The emmy library provides a single dlog group to test with. New groups must be generated to test different key sizes.

There is a go crypto library that can be used to generate the DSA groups `crypto/dsa`. The `utils/generate_dlog.go` library provided here adapts this to remove the check that (L,N) must be some specific values to allow testing a much wider range of key sizes.

### Writing to the csv files

This is an example of how the dlog_equality results were created. The other results were created in the same way.

```bash
dlog_equality/basic_tests.sh > dlog_equality/outputs_and_analysis/basic_output1.csv
dlog_equality/basic_tests.sh > dlog_equality/outputs_and_analysis/basic_output2.csv
dlog_equality/basic_tests.sh > dlog_equality/outputs_and_analysis/basic_output3.csv
```

### Different schnorr protocols
Emmy offers the option to run the schnorr protocol with these options:
- "Sigma": common.Sigma
- "ZKP":   common.ZKP
- "ZKPOK": common.ZKPOK

Each of the option adds an additional step to the previous. These results just record the timings for the default (and most complex) ZKPOK. To obtain results for the other options, specify them with the `-prot` flag eg: `ZKP_schnorr -prot sigma -N 256 -L 1024`.
