package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"

	"github.com/hopstee/sshvault/internal/storage"
	"github.com/hopstee/sshvault/internal/utils"
	"github.com/spf13/cobra"
)

const SSHKeyTTLInAgent = "8h"

var (
	ErrHostNotTrusted = errors.New("host not trusted")
)

func (c *Command) connectCmd() {
	var remember bool
	cmd := &cobra.Command{
		Use:                   "conn [name]",
		Short:                 "Connect to an SSH connection",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			conn, err := c.storage.Find(name)
			if err != nil {
				slog.Error("failed to find connection", "err", err)
				return
			}

			sshArgs := []string{"-p", strconv.Itoa(conn.Port)}

			c.processAuth(conn, sshArgs, remember)
		},
	}

	cmd.Flags().BoolVar(
		&remember,
		"remember",
		true,
		fmt.Sprintf("Add SSH key to ssh-agent for %s", SSHKeyTTLInAgent),
	)
	c.Cmd.AddCommand(cmd)
}

func (c *Command) processAuth(conn storage.Record, sshArgs []string, remember bool) {
	var sshCmd *exec.Cmd
	var err error

	if conn.AuthType == "" {
		if err := c.backfillAuth(&conn); err != nil {
			return
		}
	}

	switch conn.AuthType {
	case storage.PasswordAuth:
		sshCmd, sshArgs, err = c.processPasswordAuth(conn, sshArgs)
		if err != nil {
			return
		}
	case storage.KeyAuth:
		sshArgs = c.processKeyAuth(conn, sshArgs, remember)
	case storage.AgentAuth:
		sshArgs = c.processAgentAuth(conn, sshArgs)
	default:
		slog.Error("unknown auth type")
		return
	}

	if sshCmd == nil {
		sshCmd = exec.Command("ssh", sshArgs...)
	}

	c.provideStdInOutErrToSSHCmd(sshCmd)

	if err := sshCmd.Run(); err != nil {
		slog.Error("ssh failed", slog.Any("error", err))
	}
}

func (c *Command) processPasswordAuth(conn storage.Record, sshArgs []string) (*exec.Cmd, []string, error) {
	password, err := c.keyring.Get(conn.PasswordKey)
	if err != nil {
		slog.Error("failed to get password from keyring", slog.Any("error", err))
		return nil, sshArgs, err
	}

	sshArgs = append(
		sshArgs,
		fmt.Sprintf("%s@%s", conn.User, conn.Address),
	)

	if err := utils.CheckBinary("sshpass"); err != nil {
		slog.Error("failed check binary", slog.Any("error", err))
		return nil, sshArgs, err
	}

	if !utils.HostKnown(conn.Address) {
		if err := c.verifyHost(conn); err != nil {
			slog.Error("failed to verify host")
			return nil, sshArgs, err
		}
	}

	sshCmd := exec.Command(
		"sshpass",
		append(
			[]string{"-e", "ssh"},
			sshArgs...,
		)...,
	)
	sshCmd.Env = append(os.Environ(), "SSHPASS="+password)

	return sshCmd, sshArgs, nil
}

func (c *Command) processKeyAuth(conn storage.Record, sshArgs []string, remember bool) []string {
	if remember {
		var addArgs []string

		if remember {
			addArgs = append(addArgs, "-t", "8h")
		}

		addArgs = append(addArgs, conn.PathToKey)

		addCmd := exec.Command("ssh-add", addArgs...)

		c.provideStdInOutErrToSSHCmd(addCmd)

		_ = addCmd.Run()
	}

	return append(
		sshArgs,
		"-i", conn.PathToKey,
		fmt.Sprintf("%s@%s", conn.User, conn.Address),
	)
}

func (c *Command) processAgentAuth(conn storage.Record, sshArgs []string) []string {
	return append(
		sshArgs,
		fmt.Sprintf("%s@%s", conn.User, conn.Address),
	)
}

func (c *Command) verifyHost(conn storage.Record) error {
	key, err := utils.GetHostKey(conn.Port, conn.Address)
	if err != nil {
		slog.Error("failed to get host key")
		return err
	}

	fp, err := utils.GetFingerprint(conn.Port, conn.Address)
	if err != nil {
		slog.Error("failed to get fingerprint")
		return err
	}

	fmt.Println("New host detected:")
	fmt.Println(fp)

	fmt.Print("Trust this host? (yes/no): ")
	var answer string
	fmt.Scanln(&answer)

	if answer != "yes" {
		return ErrHostNotTrusted
	}

	if err := utils.AddToKnownHosts(key); err != nil {
		slog.Error("failed add to known hosts")
		return err
	}

	slog.Info("Host added")
	return nil
}

func (c *Command) provideStdInOutErrToSSHCmd(sshCmd *exec.Cmd) {
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
}

func (c *Command) backfillAuth(conn *storage.Record) error {
	var passwordKey, pathToKey string
	var authType storage.AuthType

	p := &CreateParams{
		Name: conn.Name,
	}

	c.selectAuthType(p.Name, &passwordKey, &pathToKey, &authType)

	if err := c.storage.Update(conn.Name, storage.UpsertDto{
		Name:        conn.Name,
		Address:     conn.Address,
		User:        conn.User,
		PathToKey:   pathToKey,
		PasswordKey: passwordKey,
		Port:        conn.Port,
		AuthType:    authType,
	}); err != nil {
		slog.Error("failed to backfill connection auth data", slog.Any("error", err))
		return err
	}

	conn.PasswordKey = passwordKey
	conn.PathToKey = pathToKey
	conn.AuthType = authType

	return nil
}
