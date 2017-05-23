package main

/*
 * Go code that runs the zero knowledge proof of equality of
 * discrete logarithms with a blinded transcript. This will print the results
 * in the form "N, L, TimeProof, TimeVerify, Q, P, G" to standard out.
 * anything "log.Print..." goes to standard error.
 *
 * Installation:
 *      from this directory run:
 *           go install
 *
 * Usage:
 *      assuming that $GOPATH/bin is in $PATH
 *           dlog_equality_blinded_transcript [-N n] [-L l] [-prot p] >> csvOutputForTests.csv
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

	// Proof timer
	start := time.Now()                          // Start the benchmark timer
	transcript, g1, t1, G2, T2 := runProof(dlog) // Run the proof
	elapsedProof := time.Since(start)            // Record the benchmark time

	// Verify timer
	start = time.Now()
	valid := dlogproofs.VerifyBlindedTranscript(transcript, dlog, g1, t1, G2, T2)
	if !valid {
		log.Fatalf("This dlog verify was invalid")
	}
	elapsedVerify := time.Since(start)

	fmt.Printf("%v, %v, %v, %v, %v, %v, %v\n", N, L,
		elapsedProof.Nanoseconds(), elapsedVerify.Nanoseconds(),
		(*dlog).OrderOfSubgroup, (*dlog).P, (*dlog).G)

}

/*
 * For a given dlog, the proves knowledge of
 * equality of discrete logs. It also returns
 * the transcript.
 *
 * This is adapted from emmy cli.go
 */
func runProof(dlog *emmyDlog.ZpDLog) ([]*big.Int, *big.Int, *big.Int, *big.Int, *big.Int) {

	// Choose a secret which is less than the lowest q (8 bits = 1 byte)
	secret := big.NewInt(200)

	eProver := dlogproofs.NewDLogEqualityBTranscriptProver(dlog)
	eVerifier := dlogproofs.NewDLogEqualityBTranscriptVerifier(dlog, nil)

	groupOrder := new(big.Int).Sub(eProver.DLog.P, big.NewInt(1))
	g1, _ := common.GetGeneratorOfZnSubgroup(eProver.DLog.P, groupOrder, eProver.DLog.OrderOfSubgroup)
	g2, _ := common.GetGeneratorOfZnSubgroup(eProver.DLog.P, groupOrder, eProver.DLog.OrderOfSubgroup)

	t1, _ := eProver.DLog.Exponentiate(g1, secret)
	t2, _ := eProver.DLog.Exponentiate(g2, secret)

	x1, x2 := eProver.GetProofRandomData(secret, g1, g2)

	challenge := eVerifier.GetChallenge(g1, g2, t1, t2, x1, x2)
	z := eProver.GetProofData(challenge)
	verified, transcript, G2, T2 := eVerifier.Verify(z)

	// Check for errors and raise if necessary
	if verified != true {
		log.Fatalf("knowledge NOT proved")
	}

	return transcript, g1, t1, G2, T2
}

func verifyProof() {

}
