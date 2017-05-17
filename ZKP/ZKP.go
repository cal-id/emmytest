package main

import (
    "log"
    "github.com/xlab-si/emmy/common"
    "github.com/xlab-si/emmy/dlogproofs"
    "github.com/xlab-si/emmy/config"
    "math/big"

)



/*
 * Attempt at implementing an ZKP without using GRPC
 */
func main(){
    var protocolType common.ProtocolType = common.ZKPOK
    // Other options are listed here for the future
    // protocolType = common.Sigma
    // protocolType = common.ZKP
    // protocolType = common.ZKPOK

    // Print the protocol type being used
    log.Println("Using this protocol type: ", protocolType)

    // Load dlog from config -> this is used by the verifier and prover
    dlog := config.LoadPseudonymsysDLog()

    // Create the verifier directly (dlogproofs.NewSchnorrProtocolServer is
    // just a struct containing this verifier and the protocolType)
    verifier := dlogproofs.NewSchnorrVerifier(dlog, protocolType)

    // Create the prover directly (dlogproofs.NewSchnorrProtocolClient is just
    // a struct that includes the prover and the GRPC connection information
    // alongside the protocolType)
    prover := dlogproofs.NewSchnorrProver(dlog, protocolType)
    // Prover has: DLog
    log.Println("Prover.DLog: ", *prover.DLog)

    // IN MY NOTES, an P before a variable means that that variable is used
    // exclusively with the pedersen committer / receiver.
    // Variables g, p, q are shared between the pedersen and the sigma

    // --------------------- Vars and their meaning ---------------------------

    // SHARED   these are stored under the dlog global variable
    // p                              a prime which is the modulus of the group
    // q                               another prime: the order of the subgroup
    // g                                                          the generator
    // p = q * r + 1            for some integer r
    // The group Zp* contains integers [1, p-1] under addition. This it NOT
    // a prime order group (order is p-1 because 0 is not included). However,
    // it does contain a q (prime) order subgroup produced by generator g which
    // is a member of both groups.

    // PEDERSEN COMMITTER / RECEIVER
    // Pa                                            trapdoor created by client
    // Ph = g^Pa        the shared value which forces the client to stick to Pa
    // Px                                    the commited value = the challenge
    // Pr                       a random value to be revealed when decommitting
    // Pc = g^Px * Ph^Pr               the commitment to Px before revealing Px


    // SIGMA ONLY
    // r                                             random number to send over
    // s                              secret = the key only known to the client
    // x = g ^ r                                         known as the challenge
    // b = g ^ s                                          considered public key
    // z = r + Px * s


    // -------------------- Equivalent of OpeningMsg() ------------------------
    // Prover should first set up the group

    // CLIENT / PROVER / PEDERSEN RECIEVER
    // Ph = g^Pa where Pa is a trapdoor (the commitment)
    Ph := prover.GetOpeningMsg()
    // Normally {H:Ph} -> server

    // SERVER / VERIFIER / PEDERSEN COMMITTER
    // Store Ph in the verifer,
    // get the challenge (== committed value)
    // and return the commitment, Pc = g^Px * Ph^Pr
    // Pr is a random
    // Px is the commited value
    Pc := verifier.GetOpeningMsgReply(Ph)
    // Normally {X1: commitment} -> client

    // CLIENT / PROVER / PEDERSEN RECIEVER
    // Store the commitment to check later
    prover.PedersenReceiver.SetCommitment(Pc)

    // ----------------- Equivalent of ProofRandomData() ----------------------
    // CLIENT / PROVER / PEDERSEN RECIEVER
    // Create a secret
    s := big.NewInt(345345345334)
    // Store the secret and store a random number r
    // Return:
    // x = g ^ r
    x := prover.GetProofRandomData(s, dlog.G)
    // b = g ^ s
    b, _ := prover.DLog.Exponentiate(dlog.G, s)
    // Normally {X:x, A:dlog.G, B:b} -> server

    // SERVER / VERIFIER / PEDERSEN COMMITTER
    verifier.SetProofRandomData(x, dlog.G, b)  // store in verifier
    // This line reveals the committed value, Px from the server to the client
    // It reveals the random value to prove that it committed to this value
    // earlier
    Px, Pr := verifier.GetChallenge()
    // Normally {X:Px, R:Pr} -> client

    // CLIENT / PROVER / PEDERSEN RECIEVER
    // Check that Px is the value that was committed to using Pr in the first
    // message response.
    // In other words, show that Px and Pr produce Pc (Pc was stored in
    // SetCommitment(commitment))
    success := prover.PedersenReceiver.CheckDecommitment(Pr, Px)
    if ! success{
        // If the decommitment failed then something went wrong.
        log.Fatalf("Commitment failed")
    }

    // ---------------------- Equivalent of ProofData() -----------------------
    // CLIENT / PROVER / PEDERSEN RECIEVER
    // z = r + Px * s
    // Pa = trapdoor, the original Pa from OpeningMsg()
    z, Pa := prover.GetProofData(Px)
    // Normally {Z:z, Trapdoor:Pa} -> server

    // SERVER / VERIFIER / PEDERSEN COMMITTER
    // Check that the pedersen committer verifies the trapdoor against Ph
    //     this is as simple as checking that g ^ Pa == Ph (Pa = trapdoor which
    //     was generated before OpeningMsg())
    // Also check that g^z == b ^ Px * x
    //     The RHS of that equation is: (g ^ s) ^ Px * g ^ r
    valid := verifier.Verify(z, Pa)

    if ! valid{
        log.Fatalf("Trapdoor or z was not verified")
    } else {
        log.Println("Worked!")
    }

}
