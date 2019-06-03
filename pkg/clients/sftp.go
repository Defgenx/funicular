package clients

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Manager SFTP connections structure
type SFTPManager struct {
	host      string
	port      uint32
	user      string
	password  string
	Conns     []*SFTPWrapper
	log       *log.Logger
	sshConfig *ssh.ClientConfig
}

// SFTP connection (with clients)
type SFTPWrapper struct {
	sync.Mutex
	connection *ssh.Client
	Client     *sftp.Client
	shutdown   chan bool
	closed     bool
	reconnects uint64
}

// SFTP Wrapper Construct
func NewSFTPWrapper(sshClient *ssh.Client, sftpClient *sftp.Client) *SFTPWrapper {
	return &SFTPWrapper{
		connection: sshClient,
		Client:     sftpClient,
		shutdown:   make(chan bool, 0),
		closed:     false,
		reconnects: 0,
	}
}

// SFTP Wrapper Close connection => chan notify ssh connection to close
func (s *SFTPWrapper) Close() error {
	s.Lock()
	defer s.Unlock()
	if s.closed == true {
		return fmt.Errorf("Connection was already closed")
	}
	var err = s.Client.Close()
	if err != nil {
		log.Fatalf("unable to close ftp connection: %v", err)
	} else {
		s.shutdown <- true
		s.closed = true
	}
	return s.Client.Wait()
}

// SFTP Manager Construct
func NewSFTPManager(host string, port uint32, user string, password string) *SFTPManager {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         2 * time.Second,
	}
	return &SFTPManager{
		host:      host,
		port:      port,
		user:      user,
		password:  password,
		Conns:     make([]*SFTPWrapper, 0),
		sshConfig: config,
		log:       log.New(os.Stdout, "SFTPManager", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (sm *SFTPManager) AddClient() (*SFTPWrapper, error) {
	sshConn, sftpConn := sm.newConnections()
	sftpStrut := NewSFTPWrapper(sshConn, sftpConn)
	go sm.reconnect(sftpStrut)
	sm.Conns = append(sm.Conns, sftpStrut)
	return sftpStrut, nil
}

// Private method to create ssh and sftp clients
func (sm *SFTPManager) newConnections() (*ssh.Client, *sftp.Client) {
	addr := fmt.Sprintf("%s:%d", sm.host, sm.port)
	conn, err := ssh.Dial("tcp", addr, sm.sshConfig)
	if err != nil {
		log.Fatalf("unable to connect to [%s]: %v", addr, err)
	}
	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatalf("unable to start sftp subsytem: %v", err)
	}

	return conn, client
}

// Private method to handle reconnect on error / close / timeout
func (sm *SFTPManager) reconnect(c *SFTPWrapper) {
	closed := make(chan error, 1)
	go func() {
		closed <- c.connection.Wait()
	}()

	select {
	case <-c.shutdown:
		_ = c.connection.Close()
		break
	case res := <-closed:
		sm.log.Printf("Connection closed, reconnecting: %s", res)
		sshConn, sftpConn := sm.newConnections()

		atomic.AddUint64(&c.reconnects, 1)
		c.Lock()
		c.connection = sshConn
		c.Client = sftpConn
		c.closed = false
		c.Unlock()
		// New connections set, rerun async reconnect
		sm.reconnect(c)
	}
}
