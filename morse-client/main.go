package main

import (
    "log"
    "os"
)

func main() {
    if len(os.Args[1:]) != 2 {
        log.Println("usage: morse-client username url:port") 
        return
    }
    a := initConnection(os.Args[1], os.Args[2])
    a.ListenToServer()
}
