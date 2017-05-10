package main

import (
    "log"
    "github.com/xlab-si/emmy/common"
    "github.com/xlab-si/emmy/dlogproofs"
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


    // Create the verifier directly (dlogproofs.NewSchnorrProtocolServer is
    // just a struct containing this verifier and the protocolType)
    verifier := dlogproofs.NewSchnorrVerifier(protocolType)

    // Create the prover directly (dlogproofs.NewSchnorrProtocolClient is just
    // a struct that includes the prover and the GRPC connection information
    // alongside the protocolType)
    prover, err := dlogproofs.NewSchnorrProver(protocolType)
    // Prover has: DLog
    if err != nil {
        log.Fatalf("error when creating Schnorr prover: %v", err)
    }
    log.Println("Prover.DLog: ", *prover.DLog)

    // IN MY NOTES, an P before a variable means that that variable is used
    // exclusively with the pedersen committer / receiver.
    // Variables g, p, q are shared between the pedersen and the sigma

    // --------------------- Vars and their meaning ---------------------------

    // SHARED
    // p                                                                      ?
    // q                                              the order of the subgroup
    // g                                                          the generator

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
    // t = g ^ s                                          considered public key


    // -------------------- Equivalent of OpeningMsg() ------------------------
    // Prover should first set up the group

    // CLIENT
    // Ph = g^Pa where Pa is a trapdoor (the commitment)
    // p
    // g = generator for the group
    // q is the order of the subgroup
    Ph, p, q, g := prover.GetOpeningMsg()
    // Normally {H:Ph, P:p, OrderOfSubgroup:q, G:g} -> server

    // SERVER
    // store 'received' values in both the pedersen committer and the verifier
    verifier.SetCommitmentGroup(p, q, g)
    verifier.SetGroup(p, q, g)
    // Store Ph in the verifer,
    // get the challenge (== committed value)
    // and return the commitment, Pc = g^Px * Ph^Pr
    // Pr is a random
    // Px is the commited value
    Pc := verifier.GetOpeningMsgReply(Ph)
    // Normally {X1: commitment} -> client

    // CLIENT
    // Store the commitment to check later
    prover.PedersenReceiver.SetCommitment(Pc)

    // ----------------- Equivalent of ProofRandomData() ----------------------
    // CLIENT
    // Create a secret
    s := big.NewInt(345345345334)
    // Store the secret and store a random number r
    // Return:
    // x = g ^ r
    // t = g ^ s
    x, t := prover.GetProofRandomData(s)
    // Normally {X:x, P:p, OrderOfSubgroup:q, G:g, T:t} -> server

    // SERVER
    verifier.SetProofRandomData(x, t)  // store in verifier
    // This line reveals the committed value, Px from the server to the client
    // It reveals the random value to prove that it committed to this value
    // earlier
    Px, Pr := verifier.GetChallenge()
    // Normally {X: Px, R:Pr} -> client

    // CLIENT
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
    // CLIENT
    // z = r + Px * s
    // Pa = trapdoor, the original Pa from OpeningMsg()
    z, Pa := prover.GetProofData(Px)
    // Normally {Z:z, Trapdoor:Pa} -> server

    // SERVER
    // Check that the pedersen committer verifies the trapdoor against h
    // Also check that g^z == t^Px * x
    // The RHS of that equation is: (g^s)^Pa * g^r
    valid := verifier.Verify(z, Pa)

    if ! valid{
        log.Fatalf("Trapdoor or z was not verified")
    } else {
        log.Println("Worked!")
    }

}
