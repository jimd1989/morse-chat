package main

// The Msg type is the primary means of communication between client and server
// as well as between internal channels in each progam. A Msg contains a number
// of different fields for a variety of use cases, but it's rare that any one
// Msg of a given type transmits all of these at once.

import (
    "log"
)

// Zeros seem to be handled strangely by the gob protocol sometimes, which is
// why all bytes passed in Msgs must be no less than 1. The reader may see
// + 1 / -1 throughout this code because of that.

const (
    MSG_ON uint8 = iota + 1
    MSG_OFF
    MSG_HZ
    MSG_ENTER
    MSG_LEAVE
    MSG_INTERNAL
    MSG_INTERNAL_VOLUME
    MSG_INTERNAL_NAMES
    MSG_ERROR_OK
    MSG_ERROR_INIT
    MSG_ERROR_NAME_LEN
    MSG_ERROR_NAME_EXISTS
    MSG_ERROR_USERS_MAX
)

type Msg struct {
    Type uint8
    On uint8
    Key uint8
    Hz float64
    Name string
}

func errMsgDisplay(err byte) {
    switch err {
    case MSG_ERROR_INIT:
        log.Println("Error initializing server connection.")
    case MSG_ERROR_NAME_LEN:
        log.Println("User name is too long or two short.")
    case MSG_ERROR_NAME_EXISTS:
        log.Println("User name already taken.")
    case MSG_ERROR_USERS_MAX:
        log.Println("Room is full.")
    }
}
