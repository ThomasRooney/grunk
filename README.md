grunk
=====

Grunk is a command based media player to play plug.dj streams, with as little overhead as can be reasonably achieved.

The motivation for producing this is that running a web browser to access music streaming services can often have too much overhead, such as higher battery life and cpu usage, alongside the requirement for having a modern web-browser installed.

Grunk essentially manually accesses the undocumented plug.dj API backend, using an authentication cookie that can either be automatically stolen from google-chrome or provided as a 'cookie.dat' file in the same directory. It then uses the information provided to work out the download link for the original source music file, and then stream it to a ffplay instance as it downloads.

Grunk should work on both OSX and Linux based systems, assuming the dependencies are provided.

dependencies
====

	ffplay
	go v1.2+ to compile
	google-chrome (optional) to grab the cookie automatically

'make' should compile ./grunk, which can then be run to start the stream. The default room is 'tastycat', but this is configurable

credits
====
I stole the youtube handling code from here: https://github.com/lepidosteus/youtube-dl
