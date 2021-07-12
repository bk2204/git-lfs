package ssh

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/subprocess"
)

type PktlineConnection struct {
	mu  sync.Mutex
	cmd *subprocess.Cmd
	pl  Pktline
}

func (conn *PktlineConnection) Lock() {
	conn.mu.Lock()
}

func (conn *PktlineConnection) Unlock() {
	conn.mu.Unlock()
}

func (conn *PktlineConnection) Start() error {
	conn.Lock()
	defer conn.Unlock()
	return conn.negotiateVersion()
}

func (conn *PktlineConnection) End() error {
	conn.Lock()
	defer conn.Unlock()
	err := conn.SendMessage("quit", nil, true)
	if err != nil {
		return err
	}
	_, _, err = conn.ReadStatus(true)
	conn.cmd.Wait()
	return err
}

func (conn *PktlineConnection) negotiateVersion() error {
	pkts, err := conn.pl.ReadPacketList()
	if err != nil {
		return errors.NewProtocolError("Unable to negotiate version with remote side (unable to read capabilities)", err)
	}
	ok := false
	for _, line := range pkts {
		if line == "version=1" {
			ok = true
		}
	}
	if !ok {
		return errors.NewProtocolError("Unable to negotiate version with remote side (missing version=1)", nil)
	}
	err = conn.SendMessage("version 1", nil, false)
	if err != nil {
		return errors.NewProtocolError("Unable to negotiate version with remote side (unable to send version)", err)
	}
	status, _, err := conn.ReadStatus(true)
	if err != nil {
		return errors.NewProtocolError("Unable to negotiate version with remote side (unable to read status)", err)
	}
	if status != 200 {
		return errors.NewProtocolError(fmt.Sprintf("Unable to negotiate version with remote side (unexpected status %d)", status), nil)
	}
	return nil
}

func (conn *PktlineConnection) SendMessage(command string, lines []string, delim bool) error {
	err := conn.pl.WritePacketText(command)
	if err != nil {
		return err
	}
	if delim {
		err = conn.pl.WriteDelim()
		if err != nil {
			return err
		}
	}
	for _, line := range lines {
		err = conn.pl.WritePacketText(line)
		if err != nil {
			return err
		}
	}
	return conn.pl.WriteFlush()
}

func (conn *PktlineConnection) SendMessageWithLines(command string, args []string, lines []string) error {
	err := conn.pl.WritePacketText(command)
	if err != nil {
		return err
	}
	for _, arg := range args {
		err = conn.pl.WritePacketText(arg)
		if err != nil {
			return err
		}
	}
	err = conn.pl.WriteDelim()
	if err != nil {
		return err
	}
	for _, line := range lines {
		err = conn.pl.WritePacketText(line)
		if err != nil {
			return err
		}
	}
	return conn.pl.WriteFlush()
}

func (conn *PktlineConnection) SendMessageWithData(command string, args []string, data io.Reader) error {
	err := conn.pl.WritePacketText(command)
	if err != nil {
		return err
	}
	for _, arg := range args {
		err = conn.pl.WritePacketText(arg)
		if err != nil {
			return err
		}
	}
	err = conn.pl.WriteDelim()
	if err != nil {
		return err
	}
	buf := make([]byte, 32768)
	for {
		n, err := data.Read(buf)
		if n > 0 {
			conn.pl.WritePacket(buf[0:n])
		}
		if err != nil {
			break
		}
	}
	return conn.pl.WriteFlush()
}

func (conn *PktlineConnection) ReadStatus(delim bool) (int, []string, error) {
	args := make([]string, 0, 100)
	status := 0
	seenDelim := false
	seenStatus := false
	for {
		s, pktLen, err := conn.pl.ReadPacketTextWithLength()
		if err != nil {
			return 0, nil, errors.NewProtocolError("error reading packet", err)
		}
		switch {
		case pktLen == 0:
			if status == 0 {
				return 0, nil, errors.NewProtocolError("no status seen", nil)
			} else if !seenDelim {
				return 0, nil, errors.NewProtocolError("no delimiter seen", nil)
			}
			return status, args, nil
		case pktLen == 1:
			seenDelim = true
		case seenDelim:
			args = append(args, s)
		case !seenStatus:
			ok := false
			if strings.HasPrefix(s, "status ") {
				status, err = strconv.Atoi(s[7:])
				ok = err == nil
			}
			if !ok {
				return 0, nil, errors.NewProtocolError(fmt.Sprintf("expected status line, got %q", s), err)
			}
			if !delim {
				seenDelim = true
			}
			seenStatus = true
		default:
			// Otherwise, this is an optional argument which we are
			// ignoring.
		}
	}
}

// ReadStatusWithData reads a status, arguments, and any binary data.  Note that
// the reader must be fully exhausted before invoking any other read methods.
func (conn *PktlineConnection) ReadStatusWithData() (int, []string, io.Reader, error) {
	args := make([]string, 0, 100)
	status := 0
	seenStatus := false
	for {
		s, pktLen, err := conn.pl.ReadPacketTextWithLength()
		if err != nil {
			return 0, nil, nil, errors.NewProtocolError("error reading packet", err)
		}
		if pktLen == 0 {
			if status == 0 {
				return 0, nil, nil, errors.NewProtocolError("no status seen", nil)
			}
			return status, args, nil, nil
		} else if pktLen == 1 {
			break
		} else if !seenStatus {
			ok := false
			if strings.HasPrefix(s, "status ") {
				status, err = strconv.Atoi(s[7:])
				ok = err == nil
			}
			if !ok {
				return 0, nil, nil, errors.NewProtocolError(fmt.Sprintf("expected status line, got %q", s), err)
			}
			seenStatus = true
		} else {
			args = append(args, s)
		}
	}

	return status, args, pktlineReader(conn.pl), nil
}

// ReadStatusWithArguments reads a status, arguments, and a set of text lines.
func (conn *PktlineConnection) ReadStatusWithArguments() (int, []string, []string, error) {
	args := make([]string, 0, 100)
	lines := make([]string, 0, 100)
	status := 0
	seenDelim := false
	seenStatus := false
	for {
		s, pktLen, err := conn.pl.ReadPacketTextWithLength()
		if err != nil {
			return 0, nil, nil, errors.NewProtocolError("error reading packet", err)
		}
		switch {
		case pktLen == 0:
			if status == 0 {
				return 0, nil, nil, errors.NewProtocolError("no status seen", nil)
			}
			return status, args, lines, nil
		case pktLen == 1:
			seenDelim = true
		case seenDelim:
			lines = append(lines, s)
		case !seenStatus:
			ok := false
			if strings.HasPrefix(s, "status ") {
				status, err = strconv.Atoi(s[7:])
				ok = err == nil
			}
			if !ok {
				return 0, nil, nil, errors.NewProtocolError(fmt.Sprintf("expected status line, got %q", s), err)
			}
			seenStatus = true
		default:
			args = append(args, s)
		}
	}
}
