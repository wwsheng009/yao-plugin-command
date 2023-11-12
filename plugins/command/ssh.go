package main

import (
	"bytes"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func SSHCopyFolder(addr string, port string, privateKey string, user string, password string, localFolder, remoteFolder string) error {

	lPort := port
	if lPort == "" {
		lPort = "22"
	}
	var authMethod ssh.AuthMethod
	if privateKey != "" {
		key, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return err
		}
		authMethod = ssh.PublicKeys(key)
	} else if password != "" {
		authMethod = ssh.Password(password)
	}

	// Authentication
	config := &ssh.ClientConfig{
		User: user,
		// https://github.com/golang/go/issues/19767
		// as clientConfig is non-permissive by default
		// you can set ssh.InsercureIgnoreHostKey to allow any host
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(addr, lPort), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	// open an SFTP session over an existing ssh connection.
	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()
	// Create remote folder if it does not exist
	_, err = client.Stat(remoteFolder)
	if err != nil {
		if os.IsNotExist(err) {
			err = client.MkdirAll(remoteFolder)
			if err != nil {
				return errors.New("Failed to create remote folder: " + err.Error())
			}
		} else {
			return errors.New("Failed to stat remote folder: " + err.Error())
		}
	}

	// Copy local folder to remote folder
	err = filepath.Walk(localFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.New("Error Occurs: " + path + " " + err.Error())
		}

		relPath, err := filepath.Rel(localFolder, path)
		if err != nil {
			return errors.New("Failed to Get File Relation Path: " + err.Error())
		}

		remotePath := filepath.Join(remoteFolder, relPath)

		if info.IsDir() {
			_, err = client.Stat(remotePath)
			if err != nil {
				if os.IsNotExist(err) {
					err = client.MkdirAll(remotePath)
					if err != nil {
						return errors.New("Failed to create remote folder: " + remotePath + " " + err.Error())
					}
				} else {
					return errors.New("Failed to stat remote folder: " + remotePath + " " + err.Error())
				}
			}
			return nil
		}

		file, err := os.ReadFile(path)
		if err != nil {
			return errors.New("Failed to Read File: " + path + " " + err.Error())
		}

		dst, err := client.Create(remotePath)
		if err != nil {
			return errors.New("Failed to Create Remote File: " + remotePath + " " + err.Error())
		}

		_, err = dst.Write(file)
		if err != nil {
			return errors.New("Failed to Write Remote File: " + err.Error())

		}

		return nil
	})
	if err != nil {
		return errors.New("Failed to copy local folder to remote folder: " + err.Error())

	}

	// Create remote folder if it does not exist

	return nil
}

func SSHCopyFile(addr string, port string, privateKey string, user string, password string, srcPath, dstPath string) error {

	lPort := port
	if lPort == "" {
		lPort = "22"
	}
	var authMethod ssh.AuthMethod
	if privateKey != "" {
		key, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return err
		}
		authMethod = ssh.PublicKeys(key)
	} else if password != "" {
		authMethod = ssh.Password(password)
	}

	// Authentication
	config := &ssh.ClientConfig{
		User: user,
		// https://github.com/golang/go/issues/19767
		// as clientConfig is non-permissive by default
		// you can set ssh.InsercureIgnoreHostKey to allow any host
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(addr, lPort), config)
	if err != nil {
		return err
	}
	defer client.Close()

	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := sftp.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// write to file
	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}
	return nil
}

func SSHWriteFile(addr string, port string, privateKey string, user string, password string, data, dstPath string) error {

	lPort := port
	if lPort == "" {
		lPort = "22"
	}
	var authMethod ssh.AuthMethod
	if privateKey != "" {
		key, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return err
		}
		authMethod = ssh.PublicKeys(key)
	} else if password != "" {
		authMethod = ssh.Password(password)
	}

	// Authentication
	config := &ssh.ClientConfig{
		User: user,
		// https://github.com/golang/go/issues/19767
		// as clientConfig is non-permissive by default
		// you can set ssh.InsercureIgnoreHostKey to allow any host
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(addr, lPort), config)
	if err != nil {
		return err
	}
	defer client.Close()

	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftp.Close()

	// Convert String data to io.Reader
	srcFile := strings.NewReader(data)
	if err != nil {
		return err
	}
	// Create the destination file
	dstFile, err := sftp.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// write to file
	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}
	return nil
}

// e.g. output, err := SSHRun("root", "MY_IP", "PRIVATE_KEY", "ls")
func SSHRun(addr string, port string, privateKey string, user string, password string, cmd string) (string, string, error) {
	// privateKey could be read from a file, or retrieved from another storage
	// source, such as the Secret Service / GNOME Keyring

	lPort := port
	if lPort == "" {
		lPort = "22"
	}
	var authMethod ssh.AuthMethod
	if privateKey != "" {
		key, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return "", "", err
		}
		authMethod = ssh.PublicKeys(key)
	} else if password != "" {
		authMethod = ssh.Password(password)
	}

	// Authentication
	config := &ssh.ClientConfig{
		User: user,
		// https://github.com/golang/go/issues/19767
		// as clientConfig is non-permissive by default
		// you can set ssh.InsercureIgnoreHostKey to allow any host
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}
	// Connect
	client, err := ssh.Dial("tcp", net.JoinHostPort(addr, lPort), config)
	if err != nil {
		return "", "", err
	}
	defer client.Close()
	// Create a session. It is one session per command.
	session, err := client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()
	var b bytes.Buffer  // import "bytes"
	var er bytes.Buffer // import "bytes"

	session.Stdout = &b  // get output
	session.Stderr = &er // get output
	// you can also pass what gets input to the stdin, allowing you to pipe
	// content from client to server
	//      session.Stdin = bytes.NewBufferString("My input")

	// Finally, run the command
	err = session.Run(cmd)
	return b.String(), er.String(), err
}
