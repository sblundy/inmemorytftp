package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"
)

const testPort = 1024
const dummyFilename = "dummy.txt"

func TestTftpServer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	sut := New(testPort, 10*time.Millisecond)
	go sut.Listen()
	defer sut.Stop()

	//Tests
	getNonExistentFile(t)
	writeDummyFile(t)
	getDummyFile(t)
}

func getNonExistentFile(t *testing.T) {
	client := newTestClient(testPort, t)
	defer client.Close()
	client.Start("binary")
	client.Get("test.txt")
	client.Quit()
	output := client.Output()

	if !strings.Contains(output, "Error code 256: File not found") {
		t.Error("File returned when not expected:", output)
	}
}

func writeDummyFile(t *testing.T) {
	client := newTestClient(testPort, t)
	defer client.Close()
	client.Start("binary")
	client.Put(dummyFilename, dummyFileContents())
	client.Quit()
	output := client.Output()

	if !strings.Contains(output, "Sent 1000 bytes in") {
		t.Error("File not transmitted", output)
	}
}

func getDummyFile(t *testing.T) {
	client := newTestClient(testPort, t)
	defer client.Close()
	client.Start("binary")
	client.Get(dummyFilename)
	client.Quit()
	output := client.Output()
	client.cmd.Wait()
	contents, ok := client.OpenLocalFile(dummyFilename)

	if !strings.Contains(output, "Received 1000 bytes in") {
		t.Error("File not retreived", output)
	}
	if !ok {
		t.Error("Unable to open local file", output)
	} else if !bytes.Equal(contents, dummyFileContents()) {
		t.Error("File contents mangled", output)
	}
}
func dummyFileContents() []byte {
	return []byte(strings.Repeat("1234567890", 100))
}

type testClient struct {
	t   *testing.T
	cmd *exec.Cmd
	in  io.WriteCloser
	out io.ReadCloser
}

func newTestClient(port int, t *testing.T) testClient {
	tftpCmd := exec.Command("tftp", "localhost", fmt.Sprintf("%d", port))
	tmpDir, err := ioutil.TempDir("", "test-client")
	if err != nil {
		t.Fatal("Unable to create temp dir", err)
	}
	tftpCmd.Dir = tmpDir

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

func (client *testClient) Put(file string, contents []byte) {
	f, err := os.Create(path.Join(client.cmd.Dir, file))
	if err != nil {
		client.t.Fatal("Unable to create dummy file", err)
	}
	defer f.Close()
	f.Write(contents)
	client.writeln(fmt.Sprintf("put %s\n", file))
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
func (client testClient) OpenLocalFile(filename string) ([]byte, bool) {
	f, err := os.Open(path.Join(client.cmd.Dir, filename))
	if err != nil {
		fmt.Println("Unable to open file", err)
		return nil, false
	}
	defer f.Close()
	buff := make([]byte, 1000)
	n, err := f.Read(buff)
	if err != nil {
		fmt.Println("Unable to read file", err)
		return nil, false
	}
	return buff[:n], true
}

func (client testClient) Close() {
	if !client.t.Failed() {
		os.RemoveAll(client.cmd.Dir)
	}
}
