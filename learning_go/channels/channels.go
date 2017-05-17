package main

/*
 * A simple go package to test out using channels with go routines.
 *
 * A channel "messages" is created and passed to each go routine 
 * "msgs". At the end of each go routine, it publishes a string
 * to the channel.
 *
 * In the main() func, after the goroutines are started, a for
 * loop waits for all of them to send a message (ie waits for
 * all of them to finish). As each of the messages are 
 * recieved, the output is printed to the console.
 */


import (
    "fmt"
    "time"
)

func f(from string, repeats int, msgs chan string){
    for i := 0 ; i < repeats ; i ++ {
        fmt.Println(from, ":", i)
        time.Sleep(time.Second)
    }
    msgs <- "@ done " + from
}

func main() {

    messages := make(chan string)

    go f("goroutine1-2", 2, messages)
    go f("goroutine1-5", 5, messages)
    go f("goroutine1-1", 1, messages)
    go f("goroutine1-7", 7, messages)
    go f("goroutine1-10", 10, messages)

    fmt.Println("waiting for done from all")
    for i := 0 ; i < 5 ; i ++ {
        fmt.Println(<-messages)
    }
    fmt.Println("all should be done. Waiting for input before exit")

    var input string
    fmt.Scanln(&input)
    fmt.Println("done")
}
