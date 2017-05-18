package main


/*
 * Go code that runs the emmy server and client for a
 * zero knowledge proof using goroutines rather than
 * separate cli interfaces. Eventually, there will be
 * timings produced for each key size.
 */

import (
    "github.com/xlab-si/emmy/common"
    "log"
    "github.com/xlab-si/emmy/dlogproofs"
    "math/big"
    "errors"
    "flag"
    emmyDlog "github.com/xlab-si/emmy/dlog"
)


/*
 * Sets the protocol type and then starts everything going
 */
func main(){
    cliProtocolTypePtr := flag.String("prot", "ZKPOK", "Protocol type: Sigma, ZKP, ZKPOK")
    Nptr := flag.Int("N", 8, "N = bit length of q, must be divisible by 8")
    Lptr := flag.Int("L", 16, "L = bit length of p, must be divisible by 8")

    flag.Parse()

    protocolTypeFlagMap := map[string]common.ProtocolType{
        "Sigma": common.Sigma,
        "ZKP": common.ZKP,
        "ZKPOK": common.ZKPOK,
    }

    protocolType := protocolTypeFlagMap[*cliProtocolTypePtr]
    runWithProtocolType(protocolType, *Nptr, *Lptr)
}

/*
 * Runs the server and then the client with a given key size (N, L)
 * TODO: Time the client
 */
func runWithProtocolType(protocolType common.ProtocolType, N int, L int) {
    log.Println("Starting up, using this protocol type: ", protocolType)

    // Create a channel to be published on after the server is running
    publishWhenServerRunning := make(chan bool)

    // Instead of loading the standard dlog from the config file using:
    // dlog := config.LoadPseudonymsysDLog()
    // Generate one of a specific length
    dlog, err := generate_dlog(N, L)
    if err != nil{
        log.Fatalf("There was an error: ", err)
    }
    log.Println("Q:", (*dlog).OrderOfSubgroup,
                "P:", (*dlog).P,
                "G:", (*dlog).G)

    // Start the server running in the background
    go runServer(protocolType, dlog, publishWhenServerRunning)
    <-publishWhenServerRunning  // wait for the server to start

    // Run the client which runs the proof
    err = runClient(protocolType, dlog)
    if err != nil{
        log.Fatalf("There was an error: ", err)
    }
}

/*
 * Runs the client to prove knowledge against the server which should already
 * be running. It takes a protocolType and a dlog.
 */
func runClient(protocolType common.ProtocolType, dlog *emmyDlog.ZpDLog) error {
    schnorrProtocolClient, err := dlogproofs.NewSchnorrProtocolClient(dlog, protocolType)
    if err != nil {
        return err
    }
    // Choose a secret which is less than the lowest q (8 bits = 1 byte)
    secret := big.NewInt(200)
    // Run the proof
    isProved, err := schnorrProtocolClient.Run(dlog.G, secret)
    // Check for errors and raise if necessary
    if err != nil {
        return err
    }
    if isProved != true {
        return errors.New("knowledge NOT proved")
    }
    log.Println("knowledge proved")
    return nil
}


/*
 * Runs the server to prove knowledge against. This should be run as a
 * goroutine so that it can run in the background.
 * It takes:
 *   - protocolType to use for the proof
 *   - dlog         to use for the proof
 *   - channel      to report when the server is running and the client
 *                  can be run.
 */
func runServer(protocolType common.ProtocolType, dlog *emmyDlog.ZpDLog, channel chan bool){
    schnorrServer := dlogproofs.NewSchnorrProtocolServer(dlog, protocolType)
    go func() {
        channel <- true
    } ()
    schnorrServer.Listen()
}
