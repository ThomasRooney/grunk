grunk
=====

Grunk is a command based media player to play plug.dj streams, with as little overhead as can be reasonably achieved.

The motivation for producing this is that running a web browser to access music streaming services can often be too much overhead, in both battery life/cpu and the very requirement for having a fully compliant web-browser installed.

Grunk essentially manually accesses the undocumented plug.dj API backend, alongside an authentication cookie that can either be stolen from google-chrome or provided as a cookie.dat file. It'll then use the information provided to work out the file link for the original source file, and then stream it to a ffplay instance as it downloads.

dependencies
====

	ffplay
	go v1.2+ to compile
	google-chrome (optional) to grab the cookie automatically

'make' should compile ./grunk, which can then be run to start the stream. The default room is 'tastycat', but this is configurable

