package vterm

import (
	"io"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/ssh"
)

type sshTTY struct {
	cli    *ssh.Client
	sess   *ssh.Session
	stdin  io.WriteCloser
	stdout io.Reader
	mtx    sync.RWMutex
	siz    atomic.Pointer[Winsize]
}

// OpenSSH 打开一个 SSH 虚拟终端，此 OpenSSH 不是开源那个 OpenSSH，此处只是代表动作。
func OpenSSH() {

}
