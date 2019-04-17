package main

// All user input comes through a curses session, written in C. 

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -L/usr/lib -lcurses
#include "curses-ui.h"
Screen S;
*/
import "C"

import (
    "os"
    "strconv"
)

// The UI contains a pointer to the C Screen struct, which captures key and 
// mouse events. These events are communicated to the Audio struct and server
// by Msgs.

type UI struct {
    FromAudio chan Msg
    ToAudio chan Msg
    Screen *C.Screen
}

// The display loop. Updates to Audio are signaled through Msgs, and the curses
// session is updated accordingly. Scrolling through Msg history is not
// implemented for the time being, since all relevant information can be
// obtained through pressing the 'n' or 'h' keys whenever.

func (ui *UI) ListenToAudio(userAudioOn *C.uint) {
    ui.Screen = &C.S
    C.initScreen(ui.Screen, userAudioOn)
    helpMessage()
    go ui.ListenToInput()
    for {
        m := <- ui.FromAudio
        ui.HandleAudioMsg(&m)
    }
}

func (ui *UI) HandleAudioMsg(m *Msg) {
    switch m.Type {
    case MSG_HZ:
        s := C.CString(m.Name + " = " + 
        strconv.FormatFloat(m.Hz, 'f', 3, 64) + "Hz.")
        C.cursesPrintln(s)
    case MSG_ENTER:
        s := C.CString(m.Name + " has joined at " +
        strconv.FormatFloat(m.Hz, 'f', 3, 64) + "Hz.")
        C.cursesPrintln(s)
    case MSG_LEAVE:
        s := C.CString(m.Name + " has left.")
        C.cursesPrintln(s)
    }
}

// The input loop. Key and mouse events are sent back to Audio to update state. 

func (ui *UI) ListenToInput() {
    for {
        C.getInput(ui.Screen)
        ui.HandleInput(ui.Screen.ch)
    }
}

func (ui *UI) HandleInput(ch C.int) {
    var m Msg
    switch ch {
        // The C code returns mouse on/off events as keys 'o' and 'p'.
    case KEY_O:
        m.Type = MSG_ON
        ui.ToAudio <- m
    case KEY_P:
        m.Type = MSG_OFF
        ui.ToAudio <- m
    case KEY_V:
        s := C.CString("Enter new volume: (0.0 to 1.0)")
        C.cursesPrintln(s)
        d := C.getText()
        if d > VOLUME_MAX {
            d = VOLUME_MAX
        } else if d < VOLUME_MIN {
            d = VOLUME_MIN
        }
        m.Type = MSG_INTERNAL_VOLUME
        m.Hz = float64(d)
        s = C.CString("Volume = " + strconv.FormatFloat(m.Hz, 'f', 3, 64))
        C.cursesPrintln(s)
        ui.ToAudio <- m
    case KEY_E:
        s := C.CString("Enter new Hz value:")
        C.cursesPrintln(s)
        d := C.getText()
        m.Type = MSG_HZ
        if d > FREQ_MAX {
            d = FREQ_MAX
        } else if d < FREQ_MIN {
            d = FREQ_MIN
        }
        m.Hz = float64(d)
        ui.ToAudio <- m
    case KEY_N:
        m.Type = MSG_INTERNAL_NAMES
        ui.ToAudio <- m
    case KEY_Q:
        C.endwin()
        os.Exit(1)
    case KEY_H:
        helpMessage()
    case KEY_ENTER:
        s := C.CString(" ")
        C.cursesPrintln(s)
    }
}

func helpMessage() {
    s := C.CString("click - sound")
    C.cursesPrintln(s)
    s = C.CString("o - lock sound on")
    C.cursesPrintln(s)
    s = C.CString("p - lock sound off")
    C.cursesPrintln(s)
    s = C.CString("e - pitch")
    C.cursesPrintln(s)
    s = C.CString("v - volume")
    C.cursesPrintln(s)
    s = C.CString("n - list names")
    C.cursesPrintln(s)
    s = C.CString("q - quit")
    C.cursesPrintln(s)
    s = C.CString("h - help")
    C.cursesPrintln(s)
}
