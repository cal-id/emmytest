# Notes on ZKP

To make pedersenReceiver public use this command:
```bash
grep -r -l "pedersenReceiver" . | xargs sed -i "s/pedersenReceiver/PedersenReceiver/"
```

To make config file load use this command, viper doesn't load configuration with capital letters
```bash
sed -i "s#\$GOPATH#$GOPATH#" config/config.go
sed -i 's#"P"#"p"#;s#"Q"#"q"#;s#"G"#"g"#' config/config.go
```
## In detail how ZKPOK works

gRPC sets up a client-server relationship where the client starts every message and the server returns a response

1. "Not client" creates `schnorrServer := dlogproofs.NewSchnorrProtocolServer(dlog, protocolType)`
    * Creates `verifier := dlogproofs.NewSchnorrVerifier(dlog, protocolType)`
        * Creates `pedersenCommitter = commitments.NewPedersenCommitter(dlog)`
            * Returns `PedersenCommitter` struct
        * Returns `SchnorrVerifier` struct with `pedersenCommitter` initilised
    * Returns `verifier` and `protocolType` in a struct as `SchnorrProtocolServer`
2. "Not client" calls `schnorrServer.Listen()`
    * Registers `SchnorrProtocolServer` to the GPRC with `pb.RegisterSchnorrProtocolServer(s, server)`
        * Registers `"OpeningMsg"`
            * Takes `PedersenFirst`: {H}
            * Returns `BigInt`: {X1}
        * Registers `"ProofRandomData"` handled by `_SchnorrProtocol_ProofRandomData_Handler`
            * Takes `SchnorrProofRandomData`: {X, A, B}
            * Returns `PedersenDecommitment`: {X, R}
        * Registers `"ProofData"` handled by `_SchnorrProtocol_ProofData_Handler`
            * Takes `SchnorrProofData`: {Z, Trapdoor}
            * Returns `Status`: {Success}
3. "Client" creates `schnorrProtocolClient, err := dlogproofs.NewSchnorrProtocolClient(dlog, protocolType)`
    * Sets up the grpc connection
    * Creates `client := pb.NewSchnorrProtocolClient(conn)`
        * Returns a struct with client connection
    * Creates `prover, err := NewSchnorrProver(dlog, protocolType)`
    * Returns `protocolClient` struct with {client, conn, prover, protocolType}
4. "Client" creates `secret := big.NewInt(345345345334)`
5. "Client" calls run `isProved, err := schnorrProtocolClient.Run(dlog.G, secret)`
    * passes in dlog.G as 'a'
    * Runs OpeningMsg `commitment, _ := client.OpeningMsg()`
        * gets h using `client.prover.GetOpeningMsg()` (stores as `msg`)
        * sends it to the server using `(*client.client).OpeningMsg(context.Background(), msg)`
            * Returns the commitment `commitment := s.verifier.GetOpeningMsgReply(h)`
        * Returns the commitment
    * Stores the commitment `client.prover.pedersenReceiver.SetCommitment(commitment)`
        * Literally provides access to private variable
    * Sends first message of sigma and stores challenge `challenge, r, err := client.ProofRandomData(a, secret)`
        * Stores a, secret, r (random), x = a^r and returns x `x := client.prover.GetProofRandomData(secret, a)`
        * Stores b `b, _ := client.prover.DLog.Exponentiate(a, secret)`
        * Packaged message `msg = &pb.SchnorrProofRandomData{X: x.Bytes(), A: a.Bytes(), B: b.Bytes()}`
        * Sends to the server `reply, err := (*client.client).ProofRandomData(context.Background(), msg)`
            * Sever stores x, a, b: `s.verifier.SetProofRandomData(x, a, b)`
            * Returns challenge, r2 (pedersen decommitment) `challenge, r2 := s.verifier.GetChallenge()`
        * Returns challenge = `reply.X`, r2 = `reply.R`
    * Checks decommitment was valid `success = client.prover.pedersenReceiver.CheckDecommitment(r, challenge)`
        * Literally
    * Sends trapdoor and receives result `proved, err := client.ProofData(challenge)`
        * Generate trapdoor `z, trapdoor := client.prover.GetProofData(challenge)`
        * Send to server `(*client.client).ProofData(context.Background(), msg))`
            * Check if valid `valid := s.verifier.Verify(z, trapdoor)`
            * Returns valid
        * Returns valid

# Fixing break from 6e3f9b3 to 8cc75e9. This should work at bf5e70d

The code in emmy was changed (6e3f9b3 to 8cc75e9) so this no longer works. Here are notes on how to get it working for the latest commit at the time (bf5e70d)

## Key Changes

See [this github  comparison](https://github.com/xlab-si/emmy/compare/6e3f9b3815645f72806bc27b49fe4fd6eba0eded...b5fe70dfa272b9dc7f8d3af35a1b34cf75cf9a98) for the changes. I have edited the above notes to reflect the changes relative to the schnorr protocol.


1. The group `dlog` is loaded from a config file `dlog := config.LoadPseudonymsysDLog()` rather than being created by the prover and then transferred to the verifier using the `PedersenFirst` gRPC message. This means that `dlog` is passed into a couple of functions where it wasn't before. Also the `PedersenFirst` messge only transfers the `h` commitment rather than the group attributes as well. The `SchnorrProofRandomData` message no longer needs to transfer the group parameters but does transmit dlog.G for some reason...
2. The generator `dlog.G` for the group is now stored in the `SchnorrVerifier` and `SchnorrProver` types as a `.a` attribute. This is for convenience as it was already accessible under `.Dlog.G`. This means that the `dlog.G` (also referred to as `a`) is passed into a number of functions where it wasn't before (so that it can be set when initialising the struct)
