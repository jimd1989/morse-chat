package main

// The Msg type is the primary means of communication between client and server
// as well as between internal channels in each progam. A Msg contains a number
// of different fields for a variety of use cases, but it's rare that any one
// Msg of a given type transmits all of these at once.

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

// Msg types are used server-side for internal communications.

type Msg struct {
    Type uint8
    On uint8
    Key uint8
    Hz float64
    Name string
    Client *Client
}

// The OMsg (Optimized Msg) is identical to a client-side Msg type. It omits
// the pointer to Client structs. Msgs are converted to OMsgs before being sent
// to clients.

type OMsg struct {
    Type uint8
    On uint8
    Key uint8
    Hz float64
    Name string
}
