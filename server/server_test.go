package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const testPort = 1024

func TestTftpServer_Listen_GetNonExistentFile(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	sut := New(testPort, 10*time.Millisecond)
	go sut.Listen()
	defer sut.Stop()

	client := newTestClient(testPort, t)
	client.Start("binary")
	client.Get("test.txt")
	client.Quit()
	output := client.Output()

	if !strings.Contains(output, "Error code 0:") {
		t.Error("File returned when not expected:", output)
	}
}

type testClient struct {
	t   *testing.T
	cmd *exec.Cmd
	in  io.WriteCloser
	out io.ReadCloser
}

func newTestClient(port int, t *testing.T) testClient {
	tftpCmd := exec.Command("tftp", "localhost", fmt.Sprintf("%d", port))

	tftpIn, err := tftpCmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	tftpOut, err := tftpCmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	return testClient{t: t, cmd: tftpCmd, in: tftpIn, out: tftpOut}
}

func (client *testClient) Start(mode string) {
	if err := client.cmd.Start(); err != nil {
		client.t.Fatal(err)
	}
	client.writeln(mode)
}

func (client *testClient) Get(file string) {
	client.writeln(fmt.Sprintf("get %s\n", file))
}

func (client *testClient) Quit() {
	client.writeln("quit")
	client.in.Close()
}

func (client *testClient) Output() string {
	output, err := ioutil.ReadAll(client.out)
	if err != nil {
		client.t.Error(err)
		return ""
	} else {
		return string(output)
	}
}

func (client *testClient) writeln(s string) {
	if _, err := client.in.Write([]byte(s)); err != nil {
		client.t.Error(err)
	}
	if _, err := client.in.Write([]byte("\n")); err != nil {
		client.t.Error(err)
	}
}
