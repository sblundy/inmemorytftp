# In-Memory TFTP
Simple In-Memory TFTP Server

Build
---
Prerequisites:
* GoLang 1.10
* The project is in the directory `$GOPATH/github.com/sblundy/inmemorytftp/`
* Integration tests depend on OS X `tftp` client
* `tests/stress_tests.py` require Python 2.7 and the [TFTPy](http://tftpy.sourceforge.net/) package

To build the project, you only need to execute `go build`

Execution
---
The executable `inmemorytftp` can be run with minimal setup. Note: by default it binds to port 69. This will probably 
require running as superuser or configuring a user with access to port 69. Alternatively, it can bind to another port (see below)

The executable takes to options
* `-port` to specify an alternative port to bind to
* `-h` to show the usage message

Testing
---
 The GoLang unit tests include a few integration tests that are run by default. Also provided is the `stress_tests.py` if
 you want to hammer the server a bit.