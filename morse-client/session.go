package main

// The functions in this file are tasked with establishing a connection to the
// server. Once they've done so, all subsequent communication is handled by
// an Audio struct.

import (
    "encoding/gob"
    "log"
    "net"
)

func initConnection(name string, url string) Audio {
    a := Audio{}
    m := Msg{}
    log.Println("Connecting to", url, "as", name, "...")
    c, err := net.Dial("tcp", url)
    if err != nil {
        log.Fatal(err)
    }
    r := gob.NewDecoder(c)
    w := gob.NewEncoder(c)
    if err := w.Encode(name); err != nil {
        c.Close()
        log.Fatal(err)
    }
    if err := r.Decode(&m); err != nil {
        c.Close()
        log.Fatal(err)
    }
    if m.Type != MSG_ENTER {
        errMsgDisplay(m.Type)
        c.Close()
        log.Fatal("Re-connect when these conditions change.")
    }
    log.Println(name,"is okay.")
    USERS_MAX = int(m.Key - 1)
    if USERS_MAX == 0 || USERS_MAX > 254 {
        c.Close()
        log.Fatal("Error retrieving max user count from the server.") 
    }
    log.Println("Retrieving user key ...")
    if err := r.Decode(&m); err != nil {
        c.Close()
        log.Fatal(err)
    }
    a.Server = c
    a.Reader = r
    a.Writer = w
    a.UserKey = m.Key - 1
    return a
}

