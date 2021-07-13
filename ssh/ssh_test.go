package ssh_test

import (
	"net/url"
	"path/filepath"
	"testing"

	"github.com/git-lfs/git-lfs/lfshttp"
	"github.com/git-lfs/git-lfs/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHGetLFSExeAndArgs(t *testing.T) {
	cli, err := lfshttp.NewClient(nil)
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Path = "user/repo"

	exe, args := ssh.GetLFSExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, "download", "GET")
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{
		"--",
		"user@foo.com",
		"git-lfs-authenticate user/repo download",
	}, args)

	exe, args = ssh.GetLFSExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, "upload", "GET")
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{
		"--",
		"user@foo.com",
		"git-lfs-authenticate user/repo upload",
	}, args)
}

func TestSSHGetExeAndArgsSsh(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         "",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"--", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCustomPort(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         "",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"-p", "8888", "--", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlink(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlink(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandPrecedence(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "sshcmd",
		"GIT_SSH":         "bad",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandArgs(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "sshcmd --args 1",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd --args 1 user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandArgsWithMixedQuotes(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "sshcmd foo 'bar \"baz\"'",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd foo 'bar \"baz\"' user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandCustomPort(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "sshcmd",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd -p 8888 user@foo.com"}, args)
}

func TestSSHGetExeAndArgsCoreSshCommand(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "sshcmd --args 2",
	}, map[string]string{
		"core.sshcommand": "sshcmd --args 1",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd --args 2 user@foo.com"}, args)
}

func TestSSHGetExeAndArgsCoreSshCommandArgsWithMixedQuotes(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"core.sshcommand": "sshcmd foo 'bar \"baz\"'",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd foo 'bar \"baz\"' user@foo.com"}, args)
}

func TestSSHGetExeAndArgsConfigVersusEnv(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"core.sshcommand": "sshcmd --args 1",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd --args 1 user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCommand(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", plink + " user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCommandCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", plink + " -P 8888 user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCommand(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", plink + " -batch user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCommandCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", plink + " -batch -P 8888 user@foo.com"}, args)
}

func TestSSHGetLFSExeAndArgsWithCustomSSH(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH": "not-ssh",
	}, nil))
	require.Nil(t, err)

	u, err := url.Parse("ssh://git@host.com:12345/repo")
	require.Nil(t, err)

	e := lfshttp.EndpointFromSshUrl(u)
	t.Logf("ENDPOINT: %+v", e)
	assert.Equal(t, "12345", e.SSHMetadata.Port)
	assert.Equal(t, "git@host.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "repo", e.SSHMetadata.Path)

	exe, args := ssh.GetLFSExeAndArgs(cli.OSEnv(), cli.GitEnv(), &e.SSHMetadata, "download", "GET")
	assert.Equal(t, "not-ssh", exe)
	assert.Equal(t, []string{"-p", "12345", "git@host.com", "git-lfs-authenticate repo download"}, args)
}

func TestSSHGetLFSExeAndArgsInvalidOptionsAsHost(t *testing.T) {
	cli, err := lfshttp.NewClient(nil)
	require.Nil(t, err)

	u, err := url.Parse("ssh://-oProxyCommand=gnome-calculator/repo")
	require.Nil(t, err)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", u.Host)

	e := lfshttp.EndpointFromSshUrl(u)
	t.Logf("ENDPOINT: %+v", e)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "repo", e.SSHMetadata.Path)

	exe, args := ssh.GetLFSExeAndArgs(cli.OSEnv(), cli.GitEnv(), &e.SSHMetadata, "download", "GET")
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"--", "-oProxyCommand=gnome-calculator", "git-lfs-authenticate repo download"}, args)
}

func TestSSHGetLFSExeAndArgsInvalidOptionsAsHostWithCustomSSH(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH": "not-ssh",
	}, nil))
	require.Nil(t, err)

	u, err := url.Parse("ssh://--oProxyCommand=gnome-calculator/repo")
	require.Nil(t, err)
	assert.Equal(t, "--oProxyCommand=gnome-calculator", u.Host)

	e := lfshttp.EndpointFromSshUrl(u)
	t.Logf("ENDPOINT: %+v", e)
	assert.Equal(t, "--oProxyCommand=gnome-calculator", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "repo", e.SSHMetadata.Path)

	exe, args := ssh.GetLFSExeAndArgs(cli.OSEnv(), cli.GitEnv(), &e.SSHMetadata, "download", "GET")
	assert.Equal(t, "not-ssh", exe)
	assert.Equal(t, []string{"oProxyCommand=gnome-calculator", "git-lfs-authenticate repo download"}, args)
}

func TestSSHGetExeAndArgsInvalidOptionsAsHost(t *testing.T) {
	cli, err := lfshttp.NewClient(nil)
	require.Nil(t, err)

	u, err := url.Parse("ssh://-oProxyCommand=gnome-calculator")
	require.Nil(t, err)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", u.Host)

	e := lfshttp.EndpointFromSshUrl(u)
	t.Logf("ENDPOINT: %+v", e)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)

	exe, args, needShell := ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &e.SSHMetadata)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"--", "-oProxyCommand=gnome-calculator"}, args)
	assert.Equal(t, false, needShell)
}

func TestSSHGetExeAndArgsInvalidOptionsAsPath(t *testing.T) {
	cli, err := lfshttp.NewClient(nil)
	require.Nil(t, err)

	u, err := url.Parse("ssh://git@git-host.com/-oProxyCommand=gnome-calculator")
	require.Nil(t, err)
	assert.Equal(t, "git-host.com", u.Host)

	e := lfshttp.EndpointFromSshUrl(u)
	t.Logf("ENDPOINT: %+v", e)
	assert.Equal(t, "git@git-host.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", e.SSHMetadata.Path)

	exe, args, needShell := ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &e.SSHMetadata)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"--", "git@git-host.com"}, args)
	assert.Equal(t, false, needShell)
}

func TestParseBareSSHUrl(t *testing.T) {
	e := lfshttp.EndpointFromBareSshUrl("git@git-host.com:repo.git")
	t.Logf("endpoint: %+v", e)
	assert.Equal(t, "git@git-host.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "repo.git", e.SSHMetadata.Path)

	e = lfshttp.EndpointFromBareSshUrl("git@git-host.com/should-be-a-colon.git")
	t.Logf("endpoint: %+v", e)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)

	e = lfshttp.EndpointFromBareSshUrl("-oProxyCommand=gnome-calculator")
	t.Logf("endpoint: %+v", e)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)

	e = lfshttp.EndpointFromBareSshUrl("git@git-host.com:-oProxyCommand=gnome-calculator")
	t.Logf("endpoint: %+v", e)
	assert.Equal(t, "git@git-host.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", e.SSHMetadata.Path)
}
