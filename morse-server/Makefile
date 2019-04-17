.POSIX:
.SUFFIXES:
all:
	go build -o "morse-server"
install:
	cp morse-server /usr/local/bin
	cp morse-server.1 /usr/local/share/man/man1
uninstall:
	rm /usr/local/bin/morse-server
	rm /usr/local/share/man/man1/morse-server.1
