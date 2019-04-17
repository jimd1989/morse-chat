[↑](https://dalrym.pl/projects/index.html) &nbsp; &nbsp; [⌂](https://dalrym.pl/index.html)

This page documents two separate but complementary pieces of software: morse-server and morse-client. The two of them allow for users to chat with real time audio [morse code](https://en.wikipedia.org/wiki/Morse_code).

## Requirements and installation

Both pieces of software require that [Go](https://golang.org) and [curses](http://http://invisible-island.net/ncurses/man/) be installed. morse-client needs [libao](https://xiph.org/ao/) as well. It's likely that all three are available through your package manager of choice.

Both programs are installed with:

+ ``make``
+ ``make install`` (may have to be root)
+ ``make uninstall`` (to remove)

## morse-server

The morse-server accepts TCP connections from morse-client sessions and routes messages between them. Its invocation is simple:

    morse-server url:port max-connections
 
Where ``url:port`` is unsurprisingly where the program listens for connections, and ``max-connections`` is the number of concurrent sessions to allow. After setting up, it will spit messages about sessions out to stderr.

## morse-client

The morse-client is where an individual user does his or her chatting. It is invoked with:

    morse-client username url:port

Rules about username length and maximum connections are determined serverside. If the client parameters are acceptable, the user will be thrown into a simple curses window after connecting. Here one can click and hold the mouse to make noise. It will be audible to all connected clients. Ideally users will communicate in morse, but there's nothing stopping you from doing whatever you want with your sound.

## Screenshot

[![two clients chatting](https://dalrym.pl/media/img/morse.gif)](https://dalrym.pl/media/img/morse.gif)

## Issues

+ Scrolling back through chat history in the curses window is not supported. All relevant info can be viewed at any time with the 'n' and 'h' keys.
+ Unicode names are not supported by the default curses library. Linking alternative versions should be simple enough, but I have avoided doing so in the interests of portability.
+ Click resolution is tight but less than ideal. This is most likely due to the way sound buffers are written.
+ libao is licensed under the GPL, which unfortunately makes this project GPL as well.
