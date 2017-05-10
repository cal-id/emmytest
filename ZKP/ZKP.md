# Notes on ZKP

To make pedersenReceiver public use this command:
```bash
grep -r -l "pedersenReceiver" . | xargs sed -i "s/pedersenReceiver/PedersenReceiver/"
```
## In detail how ZKPOK works

gRPC sets up a client-server relationship where the client starts every message and the server returns a response

1. "Not client" creates `schnorrServer := dlogproofs.NewSchnorrProtocolServer(protocolType)`
    * Creates `verifier := dlogproofs.NewSchnorrVerifier(protocolType)`
        * Creates `pedersenCommitter = commitments.NewPedersenCommitter()`
            * Returns `PedersenCommitter` struct
        * Returns `SchnorrVerifier` struct with `pedersenCommitter` initilised
    * Returns `verifier` and `protocolType` in a struct as `SchnorrProtocolServer`
2. "Not client" calls `schnorrServer.Listen()`
    * Registers `SchnorrProtocolServer` to the GPRC with `pb.RegisterSchnorrProtocolServer(s, server)`
        * Registers `"OpeningMsg"`
            * Takes `PedersenFirst`: {H, P, OrderOfSubgroup, G}
            * Returns `BigInt`: {X1}
        * Registers `"ProofRandomData"` handled by `_SchnorrProtocol_ProofRandomData_Handler`
            * Takes `SchnorrProofRandomData`: {X, P, OrderOfSubgroup, G, T}
            * Returns `PedersenDecommitment`: {X, R}
        * Registers `"ProofData"` handled by `_SchnorrProtocol_ProofData_Handler`
            * Takes `SchnorrProofData`: {Z, Trapdoor}
            * Returns `Status`: {Success}
3. "Client" creates `schnorrProtocolClient, err := dlogproofs.NewSchnorrProtocolClient(protocolType)`
    * Sets up the grpc connection
    * Creates `client := pb.NewSchnorrProtocolClient(conn)`
        * Returns a struct with client connection
    * Creates `prover, err := NewSchnorrProver(protocolType)`
    * Returns `protocolClient` struct with {client, conn, prover, protocolType}
4. "Client" creates `secret := big.NewInt(345345345334)`
5. "Client" calls run `isProved, err := schnorrProtocolClient.Run(secret)`
    * Runs OpeningMsg `commitment, _ := client.OpeningMsg()`
        * gets h, p, q, g using `client.prover.GetOpeningMsg()` (stores as `msg`)
        * sends it to the server using `(*client.client).OpeningMsg(context.Background(), msg)`
            * server sets group `s.verifier.SetCommitmentGroup(p, q, g)`
            * server sets group again `s.verifier.SetGroup(p, q, g)`
            * Returns the commitment `commitment := s.verifier.GetOpeningMsgReply(h)`
        * Returns the commitment
    * Stores the commitment `client.prover.pedersenReceiver.SetCommitment(commitment)`
        * Literally provides access to private variable
    * Sends first message of sigma and stores challenge `challenge, r, err := client.ProofRandomData(secret)`
        * Stores x and t `x, t := client.prover.GetProofRandomData(secret)`
        * Sends to the server (with previous p, q, G)`reply, err := (*client.client).ProofRandomData(context.Background(), msg)`
            * Sever stores: `s.verifier.SetProofRandomData(x, t)`
            * Returns challenge `challenge, r2 := s.verifier.GetChallenge()`
        * Returns challenge = `reply.X`, r2 = `reply.R`
    * Checks decommitment was valid `success = client.prover.pedersenReceiver.CheckDecommitment(r, challenge)`
        * Literally
    * Sends trapdoor and receives result `proved, err := client.ProofData(challenge)`
        * Generate trapdoor `z, trapdoor := client.prover.GetProofData(challenge)`
        * Send to server `(*client.client).ProofData(context.Background(), msg))`
            * Check if valid `valid := s.verifier.Verify(z, trapdoor)`
            * Returns valid
        * Returns valid
