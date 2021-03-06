package clients

import (
	"fmt"
	"github.com/defgenx/funicular/internal/utils"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func NewSSHConfig(user string, password string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         2 * time.Second,
	}
}

func NewSSHClient(host string, port uint32, sshConfig *ssh.ClientConfig) (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := ssh.Dial("tcp", addr, sshConfig)

	if err != nil {
		err = utils.ErrorPrintf("unable to connect to [%s]: %v", addr, err)
		return nil, err
	}
	return conn, err
}

func NewSFTPClient(sshClient *ssh.Client) (*sftp.Client, error) {
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		err = utils.ErrorPrintf("unable to start sftp subsytem: %v", err)
		return nil, err
	}
	return client, err
}

//------------------------------------------------------------------------------

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

// SFTP Manager Construct
func NewSFTPManager(host string, port uint32, sshConfig *ssh.ClientConfig) *SFTPManager {
	return &SFTPManager{
		host:      host,
		port:      port,
		Conns:     make([]*SFTPWrapper, 0),
		sshConfig: sshConfig,
		log:       log.New(os.Stdout, "SFTPManager", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (sm *SFTPManager) AddClient() (*SFTPWrapper, error) {
	sshConn, sftpConn, err := sm.newConnections()
	if err != nil {
		return nil, err
	}
	sftpStrut := NewSFTPWrapper(sshConn, sftpConn)
	go sm.reconnect(sftpStrut)
	sm.Conns = append(sm.Conns, sftpStrut)
	return sftpStrut, err
}

// Private method to create ssh and sftp clients
func (sm *SFTPManager) newConnections() (*ssh.Client, *sftp.Client, error) {
	sshConn, err := NewSSHClient(sm.host, sm.port, sm.sshConfig)
	if err != nil {
		return sshConn, nil, err
	}
	sftpConn, err := NewSFTPClient(sshConn)
	if err != nil {
		return nil, sftpConn, err
	}

	return sshConn, sftpConn, err
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
		sshConn, sftpConn, err := sm.newConnections()
		if err != nil {
			log.Fatal(err)
		}
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

//------------------------------------------------------------------------------

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
		return utils.ErrorPrint("Connection was already closed")
	}
	var err = s.Client.Close()
	if err != nil {
		return utils.ErrorPrintf("unable to close ftp connection: %v", err)
	} else {
		s.shutdown <- true
		s.closed = true
	}
	return s.Client.Wait()
}
