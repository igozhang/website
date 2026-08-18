package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	sshPassFileName   = "ssh.pass"
	opsSSHTimeout     = 60 * time.Second
	opsSSHMaxOutput   = 256 * 1024
	opsSSHMaxCmdBytes = 4000
)

func sshPassPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), sshPassFileName)
}

// opsSSHShellCommand wraps user input in bash -lc so pipes, env, and $PATH behave like an interactive shell.
func opsSSHShellCommand(command string) string {
	escaped := strings.ReplaceAll(command, "'", "'\\''")
	return "bash -lc '" + escaped + "'"
}

func runSSHCommand(dataRoot string, targetID int, command string) (stdout, stderr string, exitCode int, err error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", "", -1, fmt.Errorf("empty command")
	}
	if len(command) > opsSSHMaxCmdBytes {
		return "", "", -1, fmt.Errorf("command too long (max %d bytes)", opsSSHMaxCmdBytes)
	}
	if strings.Contains(command, "\x00") {
		return "", "", -1, fmt.Errorf("invalid command")
	}

	tgt, err := sshTargetByID(dataRoot, targetID)
	if err != nil {
		return "", "", -1, err
	}

	client, err := opsSSHPool.get(tgt)
	if err != nil {
		return "", "", -1, fmt.Errorf("ssh dial %s: %w", tgt.label(), err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), opsSSHTimeout)
	defer cancel()

	type result struct {
		stdout []byte
		stderr []byte
		err    error
	}
	ch := make(chan result, 1)
	go func() {
		sess, sErr := client.NewSession()
		if sErr != nil {
			ch <- result{err: sErr}
			return
		}
		defer sess.Close()
		var outBuf, errBuf bytes.Buffer
		sess.Stdout = &outBuf
		sess.Stderr = &errBuf
		runErr := sess.Run(opsSSHShellCommand(command))
		ch <- result{stdout: outBuf.Bytes(), stderr: errBuf.Bytes(), err: runErr}
	}()

	var res result
	select {
	case <-ctx.Done():
		opsSSHPool.remove(tgt.poolKey())
		return "", "", -1, fmt.Errorf("command timeout (%s)", opsSSHTimeout)
	case res = <-ch:
	}

	opsSSHPool.touch(tgt)

	stdout = truncateSSHOut(string(res.stdout))
	stderr = truncateSSHOut(string(res.stderr))
	exitCode = 0
	if res.err != nil {
		if ee, ok := res.err.(*ssh.ExitError); ok {
			exitCode = ee.ExitStatus()
			if stderr == "" {
				if m := ee.Msg(); m != "" {
					stderr = m
				}
			}
		} else {
			opsSSHPool.remove(tgt.poolKey())
			if combined := strings.TrimSpace(stdout + stderr); combined != "" {
				return stdout, stderr, -1, fmt.Errorf("%w (output: %s)", res.err, combined)
			}
			return stdout, stderr, -1, res.err
		}
	}
	log.Printf("ops-ssh: %s user=%s exit=%d out=%d err=%d cmd_len=%d",
		tgt.label(), tgt.User, exitCode, len(stdout), len(stderr), len(command))
	return stdout, stderr, exitCode, nil
}

func truncateSSHOut(s string) string {
	if len(s) > opsSSHMaxOutput {
		return s[:opsSSHMaxOutput] + "\n…(输出已截断)"
	}
	return s
}

func opsSSHReachable(dataRoot string) {
	all, err := listSSHTargets(dataRoot)
	if err != nil {
		log.Printf("ops-ssh: %v (remote shell disabled until %s is configured)", err, sshPassFileName)
		return
	}
	log.Printf("ops-ssh: %d target(s) loaded (idle disconnect %s)", len(all), opsSSHIdleTimeout)
	for _, t := range all {
		log.Printf("ops-ssh:   [%d] %s", t.ID, t.label())
	}
}
