package main

// The backbone of the program. The sound-writing loop is constantly running
// in C. 

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -L/usr/lib -lao -lcurses
#include <curses.h>
#include "audio-output.h"
Out O;
*/
import "C"

import (
    "encoding/gob"
    "io"
    "log"
    "net"
)

// The User type contains all of a client's relevant audio playback info.
// There are MAX_USERS User structs that are initialized once. They are neither
// created nor destroyed; they are simply marked active or inactive when a
// client connects or disconnects from the chat. Every User contains a direct
// pointer to a C AudioInstance struct, where the on/off value and pitch of the
// User's sound are modified directly.

type User struct {
    On uint8
    Key uint8
    Hz float64
    Name string
    Instance *C.AudioInstance
}

// The Audio struct contains all User info and connections to the server. It
// also contains a C Out struct, which is the master of all AudioInstance and
// libao structs.

type Audio struct {
    Server net.Conn
    Reader *gob.Decoder
    Writer *gob.Encoder
    UserKey uint8
    Users []User
    Out *C.Out
    ToUI chan Msg
    FromUI chan Msg
    FromServer chan Msg
}

// The main loop that initializes sound playback, then the user interface, then
// listens for messages from the server, which it parses and adjusts user
// information based upon.

func (a *Audio) ListenToServer() {
    var m Msg
    defer a.Server.Close()
    log.Println("Initializing audio ...")
    if int(a.UserKey) >= USERS_MAX {
        log.Fatal("Invalid user key.")
    }
    a.Users = make([]User, USERS_MAX)
    a.Out = &C.O
    if err := C.initOut(a.Out, C.uint(USERS_MAX)); err < 0 {
        log.Fatal("Error initializing C-side audio output.")
    }
    go C.playback(a.Out)
    defer C.destroyOut(a.Out)
    log.Println("Audio running .")
    log.Println("Getting audio instance pointers ...")
    for i, _ := range a.Users {
        a.Users[i].Instance = C.getInstance(a.Out, C.uint(i))
    }
    log.Println("Got them.")
    log.Println("Launching user interface ...")
    a.ToUI = make(chan Msg)
    a.FromUI = make(chan Msg)
    a.FromServer = make(chan Msg)
    a.Users[a.UserKey].Instance.newPitch = 440.0
    ui := UI{FromAudio: a.ToUI, ToAudio: a.FromUI}
    go ui.ListenToAudio(&a.Users[a.UserKey].Instance.on)
    go a.ListenToAllMsgs()
    for {
        if err := a.Reader.Decode(&m); err != nil {
            if err == io.EOF {
                C.endwin()
                log.Fatal("Server closed.")
            } else {
                log.Println(err)
            }
        }
        m.On -= 1 // Decoding back to potential zeros
        m.Key -= 1
        a.FromServer <- m
    }
}

// A loop that listens to Msgs from the server and the UI alike, routing them
// appropriately.

func (a *Audio) ListenToAllMsgs() {
    var m Msg
    for {
        select {
        case m = <- a.FromServer:
            a.HandleMsg(&m)
        case m = <- a.FromUI:
            switch {
            case m.Type > MSG_INTERNAL && m.Type < MSG_ERROR_OK:
                // Internal Msgs are routed back into Audio
                a.HandleMsg(&m)
            case m.Type >= MSG_ERROR_OK:
                // Errors are ignored for now
            default:
                m.On += 1 // Encoding away potential zero values
                m.Key = a.UserKey + 1
                if err := a.Writer.Encode(m); err != nil {
                    log.Println(err)
                }
            }
        }
    }
}

// Actions taken in response to Msgs from server and UI alike.

func (a *Audio) HandleMsg(m *Msg) {
    switch m.Type {
    // On/off events for the local User are engaged ASAP on the C level, but
    // there is no harm in receiving redundant Msgs from the server as well.
    case MSG_ON:
        a.Users[m.Key].Instance.on = 1
        a.Users[m.Key].On = 1 
    case MSG_OFF:
        a.Users[m.Key].Instance.on = 0
        a.Users[m.Key].On = 0
    case MSG_HZ:
        m.Name = a.Users[m.Key].Name
        a.Users[m.Key].Instance.newPitch = C.double(m.Hz)
        a.Users[m.Key].Hz = m.Hz
        a.ToUI <- *m
    case MSG_ENTER:
        a.Users[m.Key].Instance.on = C.uint(m.On)
        a.Users[m.Key].Instance.newPitch = C.double(m.Hz)
        a.Users[m.Key].On = m.On
        a.Users[m.Key].Key = m.Key
        a.Users[m.Key].Hz = m.Hz
        a.Users[m.Key].Name = m.Name
        a.ToUI <- *m
    case MSG_LEAVE:
        m.Name = a.Users[m.Key].Name
        a.Users[m.Key].Instance.on = 0
        a.Users[m.Key].Instance.newPitch = 0.0
        a.Users[m.Key].On = 0
        a.Users[m.Key].Hz = 0.0
        a.Users[m.Key].Name = ""
        a.ToUI <- *m
    case MSG_INTERNAL_VOLUME:
        a.Out.masterAmplitude = C.double(m.Hz)
    case MSG_INTERNAL_NAMES:
        m.Type = MSG_HZ
        for _, u := range a.Users {
            if u.Name != "" {
                m.Name = u.Name
                m.Hz = u.Hz
                a.ToUI <- *m
            }
        }
    }
}
