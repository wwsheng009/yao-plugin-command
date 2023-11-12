package main

//插件模板
import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/yaoapp/kun/grpc"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// func init() {

// 	fmt.Println("init plugin")
// }

// 定义插件类型，包含grpc.Plugin
type CmdPlugin struct{ grpc.Plugin }

// 设置插件日志到单独的文件
func (plugin *CmdPlugin) setLogFile() {
	var output io.Writer = os.Stdout
	//开启日志
	logroot := os.Getenv("GOU_TEST_PLG_LOG")
	if logroot == "" {
		logroot = "./logs"
	}
	if logroot != "" {
		logfile, err := os.OpenFile(path.Join(logroot, "cmdt.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err == nil {
			output = logfile
		}
	}
	plugin.Plugin.SetLogger(output, grpc.Trace)
}

// 插件执行需要实现的方法
// 参数name是在调用插件时的方法名，比如调用插件demo的Hello方法是的规则是plugins.demo.Hello时。
//
// 注意：name会自动的变成小写
//
// args参数是一个数组，需要在插件中自行解析。判断它的长度与类型，再转入具体的go类型。
//
// Exec 插件入口函数
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

func (plugin *CmdPlugin) Exec(name string, args ...interface{}) (*grpc.Response, error) {
	// plugin.Logger.Log(hclog.Trace, "plugin method called", name)
	// plugin.Logger.Log(hclog.Trace, "args", args)
	var v = make(map[string]interface{})

	isOk := true
	cmdArgs := make([]string, 0)
	// if len(args) < 1 {
	// 	v = map[string]interface{}{"code": 400, "message": "参数不足，需要一个参数"}
	// 	isOk = false
	// }
	for i, val := range args {
		_, ok := val.(string)
		if !ok {
			v = map[string]interface{}{"code": 400, "message": "参数的类型需要是字符串", "args": args[i]}
			isOk = false
			break
		}
	}

	isRemote := false

	outputStr := ""
	errStr := ""
	statusText := ""
	for _, val := range args {
		param, ok := val.(string)
		if ok {
			cmdArgs = append(cmdArgs, param)
		}
	}
	switch name {
	case "cmd":
		cmdArgs = append([]string{name}, cmdArgs...)
		cmdArgs = append([]string{"/c"}, cmdArgs...)
	case "bash":
		cmdArgs = append([]string{name}, cmdArgs...)
		cmdArgs = append([]string{"-c"}, cmdArgs...)
	case "sh":
		cmdArgs = append([]string{name}, cmdArgs...)
		cmdArgs = append([]string{"-c"}, cmdArgs...)
	case "csh":
		cmdArgs = append([]string{name}, cmdArgs...)
		cmdArgs = append([]string{"-c"}, cmdArgs...)
	case "ksh":
		cmdArgs = append([]string{name}, cmdArgs...)
		cmdArgs = append([]string{"-c"}, cmdArgs...)
	case "zsh":
		cmdArgs = append([]string{name}, cmdArgs...)
		cmdArgs = append([]string{"-c"}, cmdArgs...)
	case "fish":
		cmdArgs = append([]string{name}, cmdArgs...)
		cmdArgs = append([]string{"-c"}, cmdArgs...)
	case "scp":
		if len(cmdArgs) < 2 {
			v = map[string]interface{}{"code": 400, "message": "参数不足，需要2个参数"}
			isOk = false
		}
		cmdArgs = append([]string{name}, cmdArgs...)
		cmdArgs = append([]string{"-r"}, cmdArgs...)
	case "run":

	case "remote":
		isRemote = true

		// host,port,user,password,command
		if len(cmdArgs) < 5 {
			v = map[string]interface{}{"code": 400, "message": "参数不足，需要5个参数"}
			isOk = false
		}

		commane_line := strings.Join(cmdArgs[4:], " ")
		plugin.Logger.Log(hclog.Trace, "excute remote command:"+commane_line)

		result, eStr, err := SSHRun(cmdArgs[0], cmdArgs[1], "", cmdArgs[2], cmdArgs[3], commane_line)
		if err != nil {
			errStr = err.Error()
		} else {
			errStr = eStr
		}
		outputStr = result
	case "remote_key":
		isRemote = true
		// host,port,key,command
		if len(args) < 4 {
			v = map[string]interface{}{"code": 400, "message": "参数不足，需要4个参数"}
			isOk = false
		}
		commane_line := strings.Join(cmdArgs[3:], " ")
		plugin.Logger.Log(hclog.Trace, "excute remote command:"+commane_line)

		result, eStr, err := SSHRun(cmdArgs[0], cmdArgs[1], cmdArgs[2], "", "", commane_line)
		if err != nil {
			errStr = err.Error()
		} else {
			errStr = eStr
		}
		outputStr = result
	case "remote_copy_file":
		isRemote = true
		// host,port,user,password,srcPath,targetPath
		if len(args) < 7 {
			v = map[string]interface{}{"code": 400, "message": "参数不足，需要6个参数"}
			isOk = false
		}
		err := SSHCopyFile(cmdArgs[0], cmdArgs[1], "", cmdArgs[2], cmdArgs[3], cmdArgs[4], cmdArgs[5])
		if err != nil {
			errStr = err.Error()
		} else {
			statusText = "ok"
		}
	case "remote_copy_file_key":
		isRemote = true
		// host,port,key,srcPath,targetPath
		if len(args) < 6 {
			v = map[string]interface{}{"code": 400, "message": "参数不足，需要5个参数"}
			isOk = false
		}
		err := SSHCopyFile(cmdArgs[0], cmdArgs[1], cmdArgs[2], "", "", cmdArgs[3], cmdArgs[4])
		if err != nil {
			errStr = err.Error()
		} else {
			statusText = "ok"
		}
	case "remote_copy_folder":
		isRemote = true
		// host,port,user,password,srcPath,targetPath
		if len(args) < 7 {
			v = map[string]interface{}{"code": 400, "message": "参数不足，需要6个参数"}
			isOk = false
		}
		err := SSHCopyFolder(cmdArgs[0], cmdArgs[1], "", cmdArgs[2], cmdArgs[3], cmdArgs[4], cmdArgs[5])
		if err != nil {
			errStr = err.Error()
		} else {
			statusText = "ok"
		}
	case "remote_copy_folder_key":
		isRemote = true
		// host,port,key,srcPath,targetPath
		if len(args) < 6 {
			v = map[string]interface{}{"code": 400, "message": "参数不足，需要5个参数"}
			isOk = false
		}
		err := SSHCopyFolder(cmdArgs[0], cmdArgs[1], cmdArgs[2], "", "", cmdArgs[3], cmdArgs[4])
		if err != nil {
			errStr = err.Error()
		} else {
			statusText = "ok"
		}
	case "remote_write_file":
		isRemote = true
		// host,port,user,password,srcPath,targetPath
		if len(args) < 7 {
			v = map[string]interface{}{"code": 400, "message": "参数不足，需要6个参数"}
			isOk = false
		}
		err := SSHWriteFile(cmdArgs[0], cmdArgs[1], "", cmdArgs[2], cmdArgs[3], cmdArgs[4], cmdArgs[5])
		if err != nil {
			errStr = err.Error()
		} else {
			statusText = "ok"
		}
	case "remote_write_file_key":
		isRemote = true
		// host,port,key,srcPath,targetPath
		if len(args) < 6 {
			v = map[string]interface{}{"code": 400, "message": "参数不足，需要5个参数"}
			isOk = false
		}
		err := SSHWriteFile(cmdArgs[0], cmdArgs[1], cmdArgs[2], "", "", cmdArgs[3], cmdArgs[4])
		if err != nil {
			errStr = err.Error()
		} else {
			statusText = "ok"
		}
	default:
		cmdArgs = append(cmdArgs, name)
	}
	// if command == "" {
	// 	v = map[string]interface{}{"output": "", "error": "Operation not supported:" + name, "status": "error"}
	// }

	if isOk && !isRemote {
		for _, val := range args {
			param, ok := val.(string)
			if ok {
				cmdArgs = append(cmdArgs, param)
			}
		}

		commane_line := strings.Join(cmdArgs, " ")
		plugin.Logger.Log(hclog.Trace, "excute command:"+commane_line)

		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

		// out, err := cmd.CombinedOutput()
		// if err != nil {
		// 	// log.Fatalf("cmd.Run() failed with %s\n", err)
		// 	errStr = err.Error()
		// }
		// outputStr = string(out)
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err := cmd.Run()

		errStr = errb.String()
		outputStr = outb.String()
		if err != nil {
			statusText = err.Error()
			plugin.Logger.Log(hclog.Trace, "error", err)
			if errStr == "" {
				errStr = statusText
				statusText = "error"
			}
		} else {
			statusText = "ok"
		}

	}
	if errStr != "" {
		statusText = "error"
		outputStr = ""
	}
	if runtime.GOOS == "windows" {
		outputStr = ConvertByte2String([]byte(outputStr), "GB18030")
		errStr = ConvertByte2String([]byte(errStr), "GB18030")
	}
	v = map[string]interface{}{"output": outputStr, "error": errStr, "status": statusText}
	//输出前需要转换成字节
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	//设置输出数据的类型
	//支持的类型：map/interface/string/integer,int/float,double/array,slice
	return &grpc.Response{Bytes: bytes, Type: "map"}, nil
}

// 生成插件时函数名修改成main
func main() {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	grpc.Serve(plugin)
}

// 调试时开启，需要直接调试时修改成main
// func main() {

// 	plugin := &DemoPlugin{}
// 	plugin.setLogFile()
// 	// grpc.Serve(plugin) 不要使用server
// 	plugin.Exec("run", "ls -a") //普通的go程序，用于开发调试
// }
