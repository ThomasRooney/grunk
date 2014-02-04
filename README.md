grunk
=====

Grunk is a gui-less media player to play plug.dj streams, run from the command line

Grunk essentially manually accesses the undocumented plug.dj API backend, using an authentication cookie that can either be automatically stolen from google-chrome or provided as a 'cookie.dat' file in the same directory. It then uses the information provided to work out the download link for the original source music file, and then stream it to a ffplay instance as it downloads.

Grunk should work on both OSX and Linux based systems, assuming the dependencies are provided.

install
=====
	go get github.com/ThomasRooney/grunk

Following the install process (if $GOPATH is setup correctly), `$GOPATH/bin` holds the grunk executable. I recommend adding `$GOPATH\bin` to `$PATH` such that musical enjoyment can begin with the command `grunk`.

dependencies
====

	ffplay
	go v1.2+ to compile
	google-chrome (optional) to grab the cookie automatically


credits
====
	github.com/lepidosteus/youtube-dl
	github.com/mattn/go-sqlite3
