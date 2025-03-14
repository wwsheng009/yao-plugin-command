package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/yaoapp/kun/grpc"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// Charset 定义字符集类型
type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

func ConvertByte2String(byte []byte, charset Charset) string {
	var str string
	switch charset {
	case GB18030:
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}
	return str
}

// CommandExecutor 命令执行器结构体
type CommandExecutor struct {
	Logger hclog.Logger
}

// NewCommandExecutor 创建新的命令执行器
func NewCommandExecutor(logger hclog.Logger) *CommandExecutor {
	return &CommandExecutor{
		Logger: logger,
	}
}

// CommandArgs 命令参数结构体
type CommandArgs struct {
	cmdArgs     []string
	isRemote    bool
	isOk        bool
	errStr      string
	outputStr   string
	statusCode  int
	statusText  string
}

// parseArgs 解析命令参数
func (e *CommandExecutor) parseArgs(args ...interface{}) *CommandArgs {
	cmdArgs := make([]string, 0)
	
	for _, val := range args {
		switch data := val.(type) {
		case string:
			cmdArgs = append(cmdArgs, data)
		case float32, float64:
			cmdArgs = append(cmdArgs, fmt.Sprintf("%f", data))
			e.Logger.Log(hclog.Trace, "paramter float", val, fmt.Sprintf("%f", data))
		case int, int16, int32, int64:
			cmdArgs = append(cmdArgs, fmt.Sprintf("%d", data))
		default:
			cmdArgs = append(cmdArgs, fmt.Sprintf("%v", data))
			typeName := reflect.TypeOf(val).Name()
			e.Logger.Log(hclog.Trace, "paramter type name", val, typeName)
		}
	}

	return &CommandArgs{
		cmdArgs:     cmdArgs,
		isRemote:    false,
		isOk:        true,
		errStr:      "",
		outputStr:   "",
		statusCode:  0,
		statusText:  "",
	}
}

// processCommandType 处理不同类型的命令
func (e *CommandExecutor) processCommandType(name string, args *CommandArgs) {
	switch name {
	case "cmd":
		args.cmdArgs = append([]string{name, "/c"}, args.cmdArgs...)
	case "bash", "sh", "csh", "ksh", "zsh", "fish":
		args.cmdArgs = append([]string{name, "-c"}, args.cmdArgs...)
	case "scp":
		if len(args.cmdArgs) < 2 {
			args.isOk = false
			args.errStr = "参数不足，需要2个参数"
		} else {
			args.cmdArgs = append([]string{name, "-r"}, args.cmdArgs...)
		}
	case "remote":
		args.isRemote = true
		if len(args.cmdArgs) < 5 {
			args.isOk = false
			args.errStr = "参数不足，需要5个参数"
		} else {
			commane_line := strings.Join(args.cmdArgs[4:], " ")
			e.Logger.Log(hclog.Trace, "excute remote command:"+commane_line)
			// args.cmdArgs[0]: 主机地址
			// args.cmdArgs[1]: 端口号
			// args.cmdArgs[2]: 用户名
			// args.cmdArgs[3]: 密码
			// args.cmdArgs[4:]: 命令行参数
			result, eStr, err := SSHRun(args.cmdArgs[0], args.cmdArgs[1], args.cmdArgs[2], args.cmdArgs[3], "", commane_line)
			if err != nil {
				args.errStr = err.Error()
			} else {
				args.errStr = eStr
			}
			args.outputStr = result
		}
	case "remote_key":
		args.isRemote = true
		if len(args.cmdArgs) < 5 {
			args.isOk = false
			args.errStr = "参数不足，需要5个参数"
		} else {
			commane_line := strings.Join(args.cmdArgs[4:], " ")
			e.Logger.Log(hclog.Trace, "excute remote command:"+commane_line)
			// args.cmdArgs[0]: 主机地址
			// args.cmdArgs[1]: 端口号
			// args.cmdArgs[2]: 用户名
			// args.cmdArgs[3]: 密钥文件路径
			// args.cmdArgs[4:]: 命令行参数
			result, eStr, err := SSHRun(args.cmdArgs[0], args.cmdArgs[1], args.cmdArgs[2], "", args.cmdArgs[3], commane_line)
			if err != nil {
				args.errStr = err.Error()
			} else {
				args.errStr = eStr
			}
			args.outputStr = result
		}
	case "remote_copy_file":
		args.isRemote = true
		if len(args.cmdArgs) < 6 {
			args.isOk = false
			args.errStr = "参数不足，需要6个参数"
		} else {
			// args.cmdArgs[0]: 主机地址
			// args.cmdArgs[1]: 端口号
			// args.cmdArgs[2]: 用户名
			// args.cmdArgs[3]: 密码
			// args.cmdArgs[4]: 本地文件路径
			// args.cmdArgs[5]: 远程文件路径
			err := SSHCopyFile(args.cmdArgs[0], args.cmdArgs[1], args.cmdArgs[2], args.cmdArgs[3], "", args.cmdArgs[4], args.cmdArgs[5])
			if err != nil {
				args.errStr = err.Error()
			} else {
				args.statusCode = 0
			}
		}
	case "remote_copy_file_key":
		args.isRemote = true
		if len(args.cmdArgs) < 6 {
			args.isOk = false
			args.errStr = "参数不足，需要6个参数"
		} else {
			// args.cmdArgs[0]: 主机地址
			// args.cmdArgs[1]: 端口号
			// args.cmdArgs[2]: 用户名
			// args.cmdArgs[3]: 密钥文件路径
			// args.cmdArgs[4]: 本地文件路径
			// args.cmdArgs[5]: 远程文件路径
			err := SSHCopyFile(args.cmdArgs[0], args.cmdArgs[1], args.cmdArgs[2], "", args.cmdArgs[3], args.cmdArgs[4], args.cmdArgs[5])
			if err != nil {
				args.errStr = err.Error()
			} else {
				args.statusCode = 0
			}
		}
	case "remote_copy_folder":
		args.isRemote = true
		if len(args.cmdArgs) < 6 {
			args.isOk = false
			args.errStr = "参数不足，需要6个参数"
		} else {
			// args.cmdArgs[0]: 主机地址
			// args.cmdArgs[1]: 端口号
			// args.cmdArgs[2]: 用户名
			// args.cmdArgs[3]: 密码
			// args.cmdArgs[4]: 本地文件夹路径
			// args.cmdArgs[5]: 远程文件夹路径
			err := SSHCopyFolder(args.cmdArgs[0], args.cmdArgs[1], args.cmdArgs[2], args.cmdArgs[3], "", args.cmdArgs[4], args.cmdArgs[5])
			if err != nil {
				args.errStr = err.Error()
			} else {
				args.statusCode = 0
			}
		}
	case "remote_copy_folder_key":
		args.isRemote = true
		if len(args.cmdArgs) < 6 {
			args.isOk = false
			args.errStr = "参数不足，需要6个参数"
		} else {
			// args.cmdArgs[0]: 主机地址
			// args.cmdArgs[1]: 端口号
			// args.cmdArgs[2]: 用户名
			// args.cmdArgs[3]: 密钥文件路径
			// args.cmdArgs[4]: 本地文件夹路径
			// args.cmdArgs[5]: 远程文件夹路径
			err := SSHCopyFolder(args.cmdArgs[0], args.cmdArgs[1], args.cmdArgs[2], "", args.cmdArgs[3], args.cmdArgs[4], args.cmdArgs[5])
			if err != nil {
				args.errStr = err.Error()
			} else {
				args.statusCode = 0
			}
		}
	case "remote_write_file":
		args.isRemote = true
		if len(args.cmdArgs) < 6 {
			args.isOk = false
			args.errStr = "参数不足，需要6个参数"
		} else {
			// 检查参数是否包含文件内容
			// args.cmdArgs[0]: 主机地址
			// args.cmdArgs[1]: 端口号
			// args.cmdArgs[2]: 用户名
			// args.cmdArgs[3]: 密码
			// args.cmdArgs[4]: 远程文件路径
			// args.cmdArgs[5]: 文件内容
			err := SSHWriteFile(args.cmdArgs[0], args.cmdArgs[1], args.cmdArgs[2], args.cmdArgs[3], "", args.cmdArgs[4], args.cmdArgs[5])
			if err != nil {
				args.errStr = err.Error()
			} else {
				args.statusCode = 0
			}
		}
	case "remote_write_file_key":
		args.isRemote = true
		if len(args.cmdArgs) < 6 {
			args.isOk = false
			args.errStr = "参数不足，需要6个参数"
		} else {
			// 检查参数是否包含文件内容
			// args.cmdArgs[0]: 主机地址
			// args.cmdArgs[1]: 端口号
			// args.cmdArgs[2]: 用户名
			// args.cmdArgs[3]: 密钥文件路径
			// args.cmdArgs[4]: 远程文件路径
			// args.cmdArgs[5]: 文件内容
			err := SSHWriteFile(args.cmdArgs[0], args.cmdArgs[1], args.cmdArgs[2], "", args.cmdArgs[3], args.cmdArgs[4], args.cmdArgs[5])
			if err != nil {
				args.errStr = err.Error()
			} else {
				args.statusCode = 0
			}
		}
	default:
		args.cmdArgs = append(args.cmdArgs, name)
	}
}

// executeLocalCommand 执行本地命令
func (e *CommandExecutor) executeLocalCommand(args *CommandArgs) {
	commane_line := strings.Join(args.cmdArgs, " ")
	e.Logger.Log(hclog.Trace, "excute command:"+commane_line)

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, args.cmdArgs[0], args.cmdArgs[1:]...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if err := cmd.Start(); err != nil {
		args.errStr = err.Error()
		args.isOk = false
	}

	if args.isOk {
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-ctx.Done():
			if err := cmd.Process.Kill(); err != nil {
				args.isOk = false
				args.errStr = err.Error()
			}
		case err := <-done:
			if err != nil {
				args.errStr = err.Error()
				args.isOk = false
			}
		}
	}

	if args.isOk {
		args.errStr = errb.String()
		args.outputStr = outb.String()
	}
}

// ExecuteCommand 执行命令
func (e *CommandExecutor) ExecuteCommand(name string, args ...interface{}) (*grpc.Response, error) {
	e.Logger.Log(hclog.Trace, "plugin method called", name)
	e.Logger.Log(hclog.Trace, "args", args)

	// 解析参数
	cmdArgs := e.parseArgs(args...)
	
	// 处理命令类型
	e.processCommandType(name, cmdArgs)

	// 执行本地命令
	if cmdArgs.isOk && !cmdArgs.isRemote {
		e.executeLocalCommand(cmdArgs)
	}

	if cmdArgs.errStr != "" {
		cmdArgs.statusCode = 503
		cmdArgs.outputStr = ""
	}

	if runtime.GOOS == "windows" {
		cmdArgs.outputStr = ConvertByte2String([]byte(cmdArgs.outputStr), "GB18030")
		cmdArgs.errStr = ConvertByte2String([]byte(cmdArgs.errStr), "GB18030")
	}

	if cmdArgs.statusCode == 0 {
		cmdArgs.errStr = "调用成功"
	}

	v := map[string]interface{}{"data": map[string]interface{}{"output": cmdArgs.outputStr}, "msg": cmdArgs.errStr, "status": cmdArgs.statusCode, "statusText": cmdArgs.statusText}

	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return &grpc.Response{Bytes: bytes, Type: "map"}, nil
} 