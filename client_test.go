package managesieve

import (
	"bufio"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/test/bufconn"
)

func newTestClient(conn net.Conn) *Client {
	return &Client{conn: conn, scanner: bufio.NewScanner(conn)}
}

func TestLogin(t *testing.T) {
	l := bufconn.Listen(10)
	done := make(chan struct{})
	go func() {
		s, err := l.Accept()
		require.NoError(t, err)
		close(done)
		scanner := bufio.NewScanner(s)
		scanner.Scan()
		_, _ = s.Write([]byte(`OK "Logged in."`))
		_, _ = s.Write([]byte("\r\n"))
	}()
	conn, err := l.Dial()
	require.NoError(t, err)

	<-done
	c := newTestClient(conn)
	assert.NoError(t, c.Login("username", "password"))
}

func TestGetScript(t *testing.T) {
	l := bufconn.Listen(10)
	done := make(chan struct{})
	go func() {
		s, err := l.Accept()
		require.NoError(t, err)
		close(done)
		scanner := bufio.NewScanner(s)
		scanner.Scan()
		res := `
{54}
#this is my wonderful script
reject "I reject all";

OK`
		_, _ = s.Write([]byte(res))
		_, _ = s.Write([]byte("\r\n"))
	}()
	conn, err := l.Dial()
	require.NoError(t, err)

	<-done
	c := newTestClient(conn)
	assert.NoError(t, c.GetScript("myscript"))
}

func TestPutScript(t *testing.T) {
	l := bufconn.Listen(10)
	done := make(chan struct{})
	go func() {
		s, err := l.Accept()
		require.NoError(t, err)
		close(done)
		scanner := bufio.NewScanner(s)
		scanner.Scan()
		_, _ = s.Write([]byte("OK\r\n"))
	}()
	conn, err := l.Dial()
	require.NoError(t, err)

	<-done
	c := newTestClient(conn)
	content := `require ["fileinto"]; if envelope :contains "to" "tmartin+sent" { fileinto "INBOX.sent";}`
	assert.NoError(t, c.PutScript("foo", content))
}

func TestSetActive(t *testing.T) {
	l := bufconn.Listen(10)
	done := make(chan struct{})
	go func() {
		s, err := l.Accept()
		require.NoError(t, err)
		close(done)
		scanner := bufio.NewScanner(s)
		scanner.Scan()
		_, _ = s.Write([]byte("OK\r\n"))
	}()
	conn, err := l.Dial()
	require.NoError(t, err)

	<-done
	c := newTestClient(conn)
	assert.NoError(t, c.SetActive(""))
}

func TestDeleteScript(t *testing.T) {
	l := bufconn.Listen(10)
	done := make(chan struct{})
	go func() {
		s, err := l.Accept()
		require.NoError(t, err)
		close(done)
		scanner := bufio.NewScanner(s)
		scanner.Scan()
		_, _ = s.Write([]byte(`No (ACTIVE) "You may not delete an active script"`))
		_, _ = s.Write([]byte("\r\n"))
	}()
	conn, err := l.Dial()
	require.NoError(t, err)

	<-done
	c := newTestClient(conn)
	assert.Error(t, c.DeleteScript("myscript"))
}
