package main

/*
 * Go code that runs the zero knowledge proof of equality of
 * discrete logarithms. This will print the results
 * in the form "N, L, Time, Q, P, G" to standard out.
 * anything "log.Print..." goes to standard error.
 *
 * Installation:
 *      from this directory run:
 *           go install
 *
 * Usage:
 *      assuming that $GOPATH/bin is in $PATH
 *           dlog_equality [-N n] [-L l] [-prot p] >> csvOutputForTests.csv
 *
 */

import (
	"flag"
	"fmt"
	"github.com/cal-id/emmytest/benchmarking/utils"
	"github.com/xlab-si/emmy/common"
	emmyDlog "github.com/xlab-si/emmy/dlog"
	"github.com/xlab-si/emmy/dlogproofs"
	"log"
	"math/big"
	"time"
)

/*
 * Reads the args and set everything going
 */
func main() {
	var N, L int
	flag.IntVar(&N, "N", 8, "N = bit length of q, must be divisible by 8")
	flag.IntVar(&L, "L", 16, "L = bit length of p, must be divisible by 8")

	flag.Parse()

	if N%8 != 0 || L%8 != 0 {
		log.Fatal("N and L must be multiples of 8.")
	}

	if N >= L {
		log.Fatal("L must be greater than N.")
	}
	run(N, L)
}

/*
 * Record the time to proof equality of two discrete
 * logs with key size (N, L)
 */
func run(N int, L int) {
	// Instead of loading the standard dlog from the config file using:
	// dlog := config.LoadPseudonymsysDLog()
	// Generate one of a specific length
	dlog, err := utils.GenerateDlog(N, L)
	if err != nil {
		log.Fatal("There was an error: ", err)
	}

	start := time.Now()          // Start the benchmark timer
	runProof(dlog)               // Run the proof
	elapsed := time.Since(start) // Record the benchmark time

	fmt.Printf("%v, %v, %v, %v, %v, %v\n", N, L,
		elapsed.Nanoseconds(),
		(*dlog).OrderOfSubgroup, (*dlog).P, (*dlog).G)

}

/*
 * For a given dlog, this proves knowledge of
 * the secret as the discrete logarithms:
 * log_{g1}{t1} = secret = log_{g2}{t2}
 *
 * This is adapted from emmy cli.go
 */
func runProof(dlog *emmyDlog.ZpDLog) {

	// Choose a secret which is less than the lowest q (8 bits = 1 byte)
	secret := big.NewInt(200)

	groupOrder := new(big.Int).Sub(dlog.P, big.NewInt(1))
	g1, _ := common.GetGeneratorOfZnSubgroup(dlog.P, groupOrder, dlog.OrderOfSubgroup)
	g2, _ := common.GetGeneratorOfZnSubgroup(dlog.P, groupOrder, dlog.OrderOfSubgroup)

	t1, _ := dlog.Exponentiate(g1, secret)
	t2, _ := dlog.Exponentiate(g2, secret)
	proved := dlogproofs.RunDLogEquality(secret, g1, g2, t1, t2, dlog)

	// Check for errors and raise if necessary
	if proved != true {
		log.Fatalf("knowledge NOT proved")
	}
}
