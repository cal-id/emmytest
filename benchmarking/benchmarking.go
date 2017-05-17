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
    "github.com/xlab-si/emmy/config"
)


func main(){
    protocolType := common.Sigma
    // protocolType := common.ZKP
    // protocolType := common.ZKPOK
    runWithProtocolType(protocolType)
}

func runWithProtocolType(protocolType common.ProtocolType) {
    log.Println("Starting up, using this protocol type: ", protocolType)

    publishWhenServerRunning := make(chan bool)
    go runServer(protocolType, publishWhenServerRunning)
    <-publishWhenServerRunning  // wait for the server to start
    runClient(protocolType)

}


func runClient(protocolType common.ProtocolType){
    log.Println(protocolType)
    dlog := config.LoadPseudonymsysDLog()

    schnorrProtocolClient, err := dlogproofs.NewSchnorrProtocolClient(dlog, protocolType)
    if err != nil {
        log.Fatalf("error when creating Schnorr protocol client: %v", err)
    }
    secret := big.NewInt(345345345334)
    isProved, err := schnorrProtocolClient.Run(dlog.G, secret)
    if isProved == true {
        log.Println("knowledge proved")
    } else {
        log.Println("knowledge NOT proved")
    }
}

func runServer(protocolType common.ProtocolType, channel chan bool){
    dlog := config.LoadPseudonymsysDLog()
    schnorrServer := dlogproofs.NewSchnorrProtocolServer(dlog, protocolType)
    go func() {
        channel <- true
    } ()
    schnorrServer.Listen()
}
