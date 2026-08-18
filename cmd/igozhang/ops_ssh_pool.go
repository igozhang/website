package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const opsSSHIdleTimeout = 5 * time.Minute

type pooledSSH struct {
	client   *ssh.Client
	lastUsed time.Time
}

type sshConnPool struct {
	mu    sync.Mutex
	conns map[string]*pooledSSH
}

func newSSHConnPool() *sshConnPool {
	return &sshConnPool{conns: make(map[string]*pooledSSH)}
}

var opsSSHPool = newSSHConnPool()

func (p *sshConnPool) runJanitor() {
	tick := time.NewTicker(60 * time.Second)
	defer tick.Stop()
	for range tick.C {
		now := time.Now()
		p.mu.Lock()
		for k, ent := range p.conns {
			if ent.client == nil || now.Sub(ent.lastUsed) > opsSSHIdleTimeout {
				if ent.client != nil {
					_ = ent.client.Close()
				}
				delete(p.conns, k)
			}
		}
		p.mu.Unlock()
	}
}

func (p *sshConnPool) remove(key string) {
	p.mu.Lock()
	if ent, ok := p.conns[key]; ok {
		if ent.client != nil {
			_ = ent.client.Close()
		}
		delete(p.conns, key)
	}
	p.mu.Unlock()
}

func dialSSH(tgt *sshTarget) (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:%d", tgt.Host, tgt.Port)
	config := &ssh.ClientConfig{
		User:            tgt.User,
		Auth:            []ssh.AuthMethod{ssh.Password(tgt.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}
	return ssh.Dial("tcp", addr, config)
}

func (p *sshConnPool) get(tgt *sshTarget) (*ssh.Client, error) {
	key := tgt.poolKey()
	now := time.Now()

	p.mu.Lock()
	if ent, ok := p.conns[key]; ok && ent.client != nil {
		if now.Sub(ent.lastUsed) <= opsSSHIdleTimeout {
			ent.lastUsed = now
			c := ent.client
			p.mu.Unlock()
			return c, nil
		}
		_ = ent.client.Close()
		delete(p.conns, key)
	}
	p.mu.Unlock()

	client, err := dialSSH(tgt)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	if old, ok := p.conns[key]; ok && old.client != nil {
		_ = client.Close()
		old.lastUsed = now
		p.mu.Unlock()
		return old.client, nil
	}
	p.conns[key] = &pooledSSH{client: client, lastUsed: now}
	p.mu.Unlock()
	return client, nil
}

func (p *sshConnPool) touch(tgt *sshTarget) {
	key := tgt.poolKey()
	p.mu.Lock()
	if ent, ok := p.conns[key]; ok {
		ent.lastUsed = time.Now()
	}
	p.mu.Unlock()
}
