package locking

import (
	"fmt"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/ssh"
)

type sshLockClient struct {
	transfer *ssh.SSHTransfer
	*lfsapi.Client
}

func (c *sshLockClient) connection() *ssh.PktlineConnection {
	return c.transfer.Connection(0)
}

func (c *sshLockClient) parseLockResponse(status int, args []string, text []string) (*Lock, string, error) {
	var lock *Lock
	var message string
	var err error
	if status >= 200 && status <= 299 || status == 409 {
		lock = &Lock{}
		for _, entry := range args {
			if strings.HasPrefix(entry, "id=") {
				lock.Id = entry[3:]
			} else if strings.HasPrefix(entry, "path=") {
				lock.Path = entry[5:]
			} else if strings.HasPrefix(entry, "ownername=") {
				lock.Owner = &User{}
				lock.Owner.Name = entry[10:]
			} else if strings.HasPrefix(entry, "locked-at=") {
				lock.LockedAt, err = time.Parse(time.RFC3339, entry[10:])
				if err != nil {
					return lock, "", fmt.Errorf("lock response: invalid locked-at: %s", entry)
				}
			}
		}
	}
	if status > 299 && len(text) > 0 {
		message = text[0]
	}
	return lock, message, nil
}

type owner string

const (
	ownerOurs    = owner("ours")
	ownerTheirs  = owner("theirs")
	ownerUnknown = owner("")
)

func (c *sshLockClient) parseListLockResponse(status int, args []string, text []string) (all []Lock, ours []Lock, theirs []Lock, nextCursor string, message string, err error) {
	type lockData struct {
		lock Lock
		id   string
		who  owner
	}
	locks := make(map[string]*lockData)
	var last *lockData
	if status >= 200 && status <= 299 {
		for _, entry := range args {
			if strings.HasPrefix(entry, "next-cursor=") {
				nextCursor = entry[12:]
			}
		}
		for _, entry := range text {
			values := strings.SplitN(entry, " ", 3)
			if len(values) < 2 || (values[0] != "lock" && len(values) < 3) {
				return nil, nil, nil, "", "", fmt.Errorf("lock response: invalid response: %q", entry)
			}
			if values[0] != "lock" && (last == nil || last.id != values[1]) {
				return nil, nil, nil, "", "", fmt.Errorf("lock response: interspersed response: %q", entry)
			}
			if values[0] == "lock" {
				locks[values[1]] = &lockData{id: values[1]}
				last = locks[values[1]]
				last.lock.Id = values[1]
			} else if values[0] == "path" {
				last.lock.Path = values[2]
			} else if values[0] == "ownername" {
				last.lock.Owner = &User{}
				last.lock.Owner.Name = values[2]
			} else if values[0] == "owner" {
				last.who = owner(values[2])
			} else if values[0] == "locked-at" {
				last.lock.LockedAt, err = time.Parse(time.RFC3339, values[2])
				if err != nil {
					return nil, nil, nil, "", "", fmt.Errorf("lock response: invalid locked-at: %s", entry)
				}
			}
		}
		for _, lock := range locks {
			all = append(all, lock.lock)
			if lock.who == ownerOurs {
				ours = append(ours, lock.lock)
			} else if lock.who == ownerTheirs {
				theirs = append(theirs, lock.lock)
			}
		}
	} else if status > 299 && len(text) > 0 {
		message = text[0]
	}
	return all, ours, theirs, nextCursor, message, nil
}

func (c *sshLockClient) Lock(remote string, lockReq *lockRequest) (*lockResponse, int, error) {
	args := make([]string, 0, 3)
	args = append(args, fmt.Sprintf("path=%s", lockReq.Path))
	if lockReq.Ref != nil {
		args = append(args, fmt.Sprintf("refname=%s", lockReq.Ref.Name))
	}
	conn := c.connection()
	conn.Lock()
	defer conn.Unlock()
	err := conn.SendMessage("lock", args, false)
	if err != nil {
		return nil, 0, err
	}
	status, args, text, err := conn.ReadStatusWithArguments()
	if err != nil {
		return nil, status, err

	}
	var lock lockResponse
	lock.Lock, lock.Message, err = c.parseLockResponse(status, args, text)
	return &lock, status, err
}

func (c *sshLockClient) Unlock(ref *git.Ref, remote, id string, force bool) (*unlockResponse, int, error) {
	args := make([]string, 0, 3)
	if ref != nil {
		args = append(args, fmt.Sprintf("refname=%s", ref.Name))
	}
	conn := c.connection()
	conn.Lock()
	defer conn.Unlock()
	err := conn.SendMessage(fmt.Sprintf("unlock %s", id), args, false)
	if err != nil {
		return nil, 0, err
	}
	status, args, text, err := conn.ReadStatusWithArguments()
	if err != nil {
		return nil, status, err

	}
	var lock unlockResponse
	lock.Lock, lock.Message, err = c.parseLockResponse(status, args, text)
	return &lock, status, err
}

func (c *sshLockClient) Search(remote string, searchReq *lockSearchRequest) (*lockList, int, error) {
	values := searchReq.QueryValues()
	args := make([]string, 0, len(values))
	for key, value := range values {
		args = append(args, fmt.Sprintf("%s=%s", key, value))
	}
	conn := c.connection()
	conn.Lock()
	defer conn.Unlock()
	err := conn.SendMessage("list-lock", args, false)
	if err != nil {
		return nil, 0, err
	}
	status, args, text, err := conn.ReadStatusWithArguments()
	if err != nil {
		return nil, status, err
	}
	locks, _, _, nextCursor, message, err := c.parseListLockResponse(status, args, text)
	if err != nil {
		return nil, status, err
	}
	list := &lockList{
		Locks:      locks,
		NextCursor: nextCursor,
		Message:    message,
	}
	return list, status, nil
}

func (c *sshLockClient) SearchVerifiable(remote string, vreq *lockVerifiableRequest) (*lockVerifiableList, int, error) {
	args := make([]string, 0, 3)
	if vreq.Ref != nil {
		args = append(args, fmt.Sprintf("refname=%s", vreq.Ref.Name))
	}
	if len(vreq.Cursor) > 0 {
		args = append(args, fmt.Sprintf("cursor=%s", vreq.Cursor))
	}
	if vreq.Limit > 0 {
		args = append(args, fmt.Sprintf("limit=%d", vreq.Limit))
	}
	conn := c.connection()
	conn.Lock()
	defer conn.Unlock()
	err := conn.SendMessage("list-locks", args, false)
	if err != nil {
		return nil, 0, err
	}
	status, args, text, err := conn.ReadStatusWithArguments()
	if err != nil {
		return nil, status, err
	}
	_, ours, theirs, nextCursor, message, err := c.parseListLockResponse(status, args, text)
	if err != nil {
		return nil, status, err
	}
	list := &lockVerifiableList{
		Ours:       ours,
		Theirs:     theirs,
		NextCursor: nextCursor,
		Message:    message,
	}
	return list, status, nil
}
