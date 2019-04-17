package main

// Global values, some of which are not technically constant, but which are
// referenced during runtime and changed infrequently.

const (
    
    // Curses keys

    KEY_ENTER = 10
    KEY_E = 101
    KEY_H = 104
    KEY_N = 110
    KEY_O = 111
    KEY_P = 112
    KEY_Q = 113
    KEY_V = 118

    // Min/max inputs

    FREQ_MIN = 20.0
    FREQ_MAX = 20000.0
    VOLUME_MIN = 0.0
    VOLUME_MAX = 1.0

    // Text buffer length (set HISTORY_LEN_MAX to 1 more than intended max)

    HISTORY_LEN_MAX = 31 
    HISTORY_MAX = HISTORY_LEN_MAX - 1
)



// The maximum number of clients in a chat room. Received from server at start
// up.

var USERS_MAX int
