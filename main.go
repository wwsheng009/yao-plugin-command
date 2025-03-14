package main

//插件模板
import (
	"io"
	"os"
	"path"

	"github.com/yaoapp/kun/grpc"
)

// func init() {

// 	fmt.Println("init plugin")
// }

// 定义插件类型，包含grpc.Plugin
type CmdPlugin struct {
	grpc.Plugin
	executor *CommandExecutor
}

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
	plugin.executor = NewCommandExecutor(plugin.Logger)
}

// 插件执行需要实现的方法
// 参数name是在调用插件时的方法名，比如调用插件demo的Hello方法是的规则是plugins.demo.Hello时。
//
// 注意：name会自动的变成小写
//
// args参数是一个数组，需要在插件中自行解析。判断它的长度与类型，再转入具体的go类型。
//
// Exec 插件入口函数


func (plugin *CmdPlugin) Exec(name string, args ...interface{}) (*grpc.Response, error) {
	return plugin.executor.ExecuteCommand(name, args...)
}

// 生成插件时函数名修改成main
func main() {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	grpc.Serve(plugin)
}