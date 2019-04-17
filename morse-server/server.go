package main

// The main server loop. Accepts connections from clients and attempts to make
// sessions out of them.

import (
    "log"
    "net"
    "os"
    "strconv"
)

func main() {
    if len(os.Args[1:]) != 2 {
        log.Println("usage: morse-server url:port max-connections")
        return
    }
    max, err := strconv.Atoi(os.Args[2])
    if err != nil {
        log.Fatal(err)
    }
    USERS_MAX = max
    if USERS_MAX == 0 || USERS_MAX > 254 {
        log.Fatal("Server must accept 1 to 254 users.")
    }
    l, err := net.Listen("tcp", os.Args[1])
    if err != nil {
        log.Fatal(err)
    }
    cs := Clients{FromClient: make(chan Msg)}
    go cs.Listen()
    log.Println("Up and listening for clients ...")
    for {
        c, err := l.Accept()
        if err != nil {
            log.Println(err)
        } else {
            cli := Client{}
            go cli.ListenToClient(c, &cs)
        }
    }
    log.Println("Shutting down...")
    l.Close()
}
