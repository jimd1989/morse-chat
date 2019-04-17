package main

// The main routing hub for sending msgs between clients and the server.
// Contains information on all connected users.

import (
    "encoding/gob"
    "errors"
    "io"
    "log"
    "net"
)

// The Client type is a bridge between the server and the application running
// on the connected user's computer. It holds this user's state, which is
// broadcast to all other users by means of the Clients type. Likewise, it
// receives state changes from other users by means of Clients.

type Client struct {
    On uint8
    Key uint8
    Hz float64
    Name string
    Reader *gob.Decoder
    Writer *gob.Encoder
    FromServer chan OMsg
}

// Client.ListenToClient() initializes the connection, adds Client data to the
// master Clients array, and awaits Msgs from the user, which it passes along
// to Clients. After initialization, it spawns the Client.ListenToServer()
// process in a separate goroutine. 

func (cli *Client) ListenToClient(c net.Conn, cs *Clients) {
    var m Msg
    defer log.Println(c.RemoteAddr(), "disconnected")
    defer c.Close()
    log.Println(c.RemoteAddr(), "connected")
    cli.Reader = gob.NewDecoder(c)
    cli.Writer = gob.NewEncoder(c)
    cli.FromServer = make(chan OMsg)
    if err := cli.Reader.Decode(&cli.Name); err != nil {
        log.Println(c.RemoteAddr(), err)
        return
    }
    cli.Hz = 440.0
    m.Type = MSG_ENTER
    m.Hz = cli.Hz
    m.Name = cli.Name
    m.Client = cli
    cs.FromClient <- m
    om := <- cli.FromServer
    if om.Type > MSG_ERROR_OK {
        return
    }
    go cli.ListenToServer(c)
    m.Client = nil
    for {
        m.Key = 255
        m.Name = ""
        if err := cli.Reader.Decode(&m); err != nil {
            if err == io.EOF {
                cli.Kick(&m, cs)
                return
            }
            log.Println(c.RemoteAddr(), err)
        }
        m.On-- // Decoding values back to potential zero
        m.Key--
        if m.Key != cli.Key {
            log.Println(c.RemoteAddr(), "invalid key")
            cli.Kick(&m, cs)
            return
        }
        cs.FromClient <- m
    }
}

// Client.ListenToServer() is a simple loop that accepts OMsgs from the
// Clients struct and encodes them back to the user. It runs in its own
// goroutine (as opposed to being in the Clients' main thread) so that the
// encoding process can take place in parallel if possible.

func (cli *Client) ListenToServer(c net.Conn) {
    for {
        om := <- cli.FromServer
        if err := cli.Writer.Encode(om); err != nil {
            log.Println(c.RemoteAddr(), err)
        }
    }
}

func (cli *Client) Kick(m *Msg, cs *Clients) {
    m.Type = MSG_LEAVE
    m.Key = cli.Key
    cs.FromClient <- *m
}

// The Clients type accepts Msgs from every Client type, which it uses to
// update user states, then dispatches the changes back to each individual
// Client through another OMsg. Client info is stored in an array, which the
// Client.Key() field addresses. This means that adding a new Client is an
// O(n) operation that must check every array index for duplicate names, but 
// all subsequent operations are able to address the index directly, without 
// need for hashing.

type Clients struct {
    FromClient chan Msg
    Available []uint8
    All []*Client
}

// Gob will not transmit zero-valued variables. Clients.NewOMsg() removes
// irrelevant fields before transmitting a given Msg type. It also ensures that
// relevant zero values, such as Msg.Key and Msg.On are transmitted by adding
// 1 to them ahead of time.

func (cs *Clients) NewOMsg(m *Msg) OMsg {
    om := OMsg{m.Type, m.On + 1, m.Key + 1, m.Hz, m.Name}
    switch {
    case m.Type == MSG_ON || m.Type == MSG_OFF:
        om.Hz = 0.0
        om.Name = ""
    case m.Type == MSG_HZ:
        om.On = 0
        om.Name = ""
    case m.Type == MSG_ENTER:
        // Keep everything
    default:
        om.On = 0
        om.Hz = 0.0
        om.Name = ""
    }
    return om
}

func (cs *Clients) NameExists(name string) bool {
    for _, cli := range cs.All {
        if cli != nil && cli.Name == name {
            return true
        }
    }
    return false
}

func (cs *Clients) On(m *Msg) error {
    cs.All[m.Key].On = 1
    return nil
}

func (cs *Clients) Off(m *Msg) error {
    cs.All[m.Key].On = 0
    return nil
}

func (cs *Clients) Hz(m *Msg) error {
    cs.All[m.Key].Hz = m.Hz
    return nil
}

func (cs *Clients) Enter(m *Msg) error {
    // Setting up an individual user's session with the server must be handled
    // within the Clients' thread to avoid race conditions. There is a lot of
    // back and forth communication between the server and the client in this
    // single method, which is ugly but unavoidable.
    var err error
    var om OMsg
    if len(m.Name) <= 0 || len(m.Name) > NAME_MAX {
        m.Type = MSG_ERROR_NAME_LEN
    } else if exists := cs.NameExists(m.Name); exists {
        m.Type = MSG_ERROR_NAME_EXISTS
    } else if len(cs.Available) == 0 {
        m.Type = MSG_ERROR_USERS_MAX
    }
    m.Key = uint8(USERS_MAX)
    om = cs.NewOMsg(m)
    err = m.Client.Writer.Encode(om)
    if err != nil || m.Type != MSG_ENTER {
        if err != nil {
            log.Println(err)
        }
        m.Type = MSG_ERROR_INIT
        m.Client.FromServer <- om
        err = errors.New("Error initializing new user.")
        return err
    }
    m.Client.Key = cs.Available[len(cs.Available) - 1]
    cs.Available = cs.Available[:len(cs.Available) - 1]
    m.Key = m.Client.Key
    om = cs.NewOMsg(m)
    err = m.Client.Writer.Encode(om)
    if err != nil {
        log.Println(err)
        m.Type = MSG_ERROR_INIT
        m.Client.FromServer <- om
        err = errors.New("Error initializing new user.")
        return err
    }
    for _, cli := range cs.All {
        if cli != nil {
            clim := &Msg{MSG_ENTER, cli.On, cli.Key, cli.Hz, cli.Name, nil}
            om = cs.NewOMsg(clim)
            err = m.Client.Writer.Encode(om)
            if err != nil {
                log.Println(err)
            }
        }
    }
    om = cs.NewOMsg(m)
    cs.All[m.Client.Key] = m.Client
    m.Client.FromServer <- om
    return nil
}

func (cs *Clients) Leave(m *Msg) error {
    var err error
    if m.Key >= uint8(USERS_MAX) {
        err = errors.New("Key out of range.")
        log.Println(err)
        return err
    }
    if cs.All[m.Key] == nil {
        err = errors.New("Inactive index targeted by key.")
        log.Println(err)
        return err
    }
    cs.Available = append(cs.Available, cs.All[m.Key].Key)
    cs.All[m.Key] = nil
    return nil
}

// The main server loop. Accepts Msgs from connected clients, updates state
// based upon their contents, and sends updates back to clients as OMsgs.

func (cs *Clients) Listen() {
    var err error
    var m Msg
    cs.All = make([]*Client, USERS_MAX)
    cs.Available = make([]uint8, USERS_MAX)
    for i, _ := range cs.Available {
        cs.Available[i] = uint8(i)
    }
    for {
        m = <- cs.FromClient
        switch m.Type {
        case MSG_ON:
            err = cs.On(&m)
        case MSG_OFF:
            err = cs.Off(&m)
        case MSG_HZ:
            err = cs.Hz(&m)
        case MSG_ENTER:
            err = cs.Enter(&m)
        case MSG_LEAVE:
            err = cs.Leave(&m)
        }
        if err == nil {
            om := cs.NewOMsg(&m)
            for _, cli := range cs.All {
                if cli != nil {
                    cli.FromServer <- om
                }
            }
        }
    }
}
