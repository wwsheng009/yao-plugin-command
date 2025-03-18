package main

import (
	"fmt"
	"testing"
)

func TestTemplatePages(t *testing.T) {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	// grpc.Serve(plugin) 不要使用server
	res, err := plugin.Exec("run", "sh", "-c", "ls -la") //普通的go程序，用于开发调试
	if err != nil {
		fmt.Println(err.Error())
	} else {
		m := res.MustMap()
		fmt.Println("error:", m.Get("error"))
		fmt.Println("output:", m.Get("output"))
		fmt.Println("status:", m.Get("status"))
	}
}
func TestWindowsCmd(t *testing.T) {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	// grpc.Serve(plugin) 不要使用server
	res, err := plugin.Exec("run", "cmd", "/c", "dir") //普通的go程序，用于开发调试
	if err != nil {
		fmt.Println(err.Error())
	} else {
		m := res.MustMap()
		fmt.Println("error:", m.Get("error"))
		fmt.Println("output:", m.Get("output"))
		fmt.Println("status:", m.Get("status"))
	}
}
func TestRemoteCmd(t *testing.T) {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	// grpc.Serve(plugin) 不要使用server
	res, err := plugin.Exec("remote", "172.18.3.234", "22", "yao", "Abcd1234", "cat", "info.txt") //普通的go程序，用于开发调试
	if err != nil {
		fmt.Println(err.Error())
	} else {
		m := res.MustMap()
		fmt.Println("error:", m.Get("error"))
		fmt.Println("output:", m.Get("output"))
		fmt.Println("status:", m.Get("status"))
	}
}

func TestRemoteKey(t *testing.T) {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	ssh_key := `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACCGkIz2KfKzQg9mcH9b2krgwd7JDKkyYHwzv9Cr8rmVpAAAAJjENtlkxDbZ
ZAAAAAtzc2gtZWQyNTUxOQAAACCGkIz2KfKzQg9mcH9b2krgwd7JDKkyYHwzv9Cr8rmVpA
AAAEB+qi3nvgfkUNbI90Z65jhoxF+rhLUJmjOMqrpR7tObUIaQjPYp8rNCD2Zwf1vaSuDB
3skMqTJgfDO/0KvyuZWkAAAAEnJvb3RAd3dzaGVuZy10azE0cAECAw==
-----END OPENSSH PRIVATE KEY-----`
	// grpc.Serve(plugin) 不要使用server
	res, err := plugin.Exec("remote_key", "172.18.3.202", "22", "root", ssh_key, "ls -l") //普通的go程序，用于开发调试
	if err != nil {
		fmt.Println(err.Error())
	} else {
		m := res.MustMap()
		fmt.Println("error:", m.Get("error"))
		fmt.Println("output:", m.Get("output"))
		fmt.Println("status:", m.Get("status"))
	}
}
func TestCopyFile(t *testing.T) {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	// grpc.Serve(plugin) 不要使用server
	res, err := plugin.Exec("remote_copy_file", "172.18.3.234", "22", "yao", "Abcd1234", "readme.md", "readme.md") //普通的go程序，用于开发调试
	if err != nil {
		fmt.Println(err.Error())
	} else {
		m := res.MustMap()
		fmt.Println("error:", m.Get("error"))
		fmt.Println("output:", m.Get("output"))
		fmt.Println("status:", m.Get("status"))
	}
}
func TestCopyFolder(t *testing.T) {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	// grpc.Serve(plugin) 不要使用server
	res, err := plugin.Exec("remote_copy_folder", "172.18.3.234", "22", "yao", "Abcd1234", "./logs", "/home/yao/logs") //普通的go程序，用于开发调试
	if err != nil {
		fmt.Println(err.Error())
	} else {
		m := res.MustMap()
		fmt.Println("error:", m.Get("error"))
		fmt.Println("output:", m.Get("output"))
		fmt.Println("status:", m.Get("status"))
	}
}
func TestWriteFile(t *testing.T) {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	// grpc.Serve(plugin) 不要使用server
	res, err := plugin.Exec("remote_write_file", "172.18.3.234", "22", "yao", "Abcd1234", "Hello World", "readme.md") //普通的go程序，用于开发调试
	if err != nil {
		fmt.Println(err.Error())
	} else {
		m := res.MustMap()
		fmt.Println("error:", m.Get("error"))
		fmt.Println("output:", m.Get("output"))
		fmt.Println("status:", m.Get("status"))
	}
}

func TestScanNormal(t *testing.T) {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	res, err := plugin.Exec("scan", "172.18.3.220", "172.18.3.236", "22")
	if err != nil {
		t.Fatalf("扫描失败: %v", err)
	}
	m := res.MustMap()
	if m.Get("status") != 200 {
		t.Errorf("预期状态码200，实际得到%d", m.Get("status"))
	}
	results := m.Get("data").(map[string]interface{})["output"].(string)
	if results == "[]" {
		t.Error("未扫描到任何开放端口")
	}
}

func TestScanInvalidPorts(t *testing.T) {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	res, _ := plugin.Exec("scan", "127.0.0.1", "127.0.0.1", "abc,99999")
	m := res.MustMap()
	if m.Get("status") != 200 {
		t.Errorf("无效端口处理异常，状态码%d", m.Get("status"))
	}
	results := m.Get("data").(map[string]interface{})["output"].(string)
	if results == "[]" {
		t.Error("无效端口参数应返回空结果")
	}
}

func TestScanServiceDetection(t *testing.T) {
	// 需要模拟服务响应，此处测试基本识别逻辑
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	res, _ := plugin.Exec("scan", "127.0.0.1", "127.0.0.1", "22")
	m := res.MustMap()
	results := m.Get("data").(map[string]interface{})["output"].(string)
	if results == "[]" {
		t.Error("服务识别功能异常")
	}
}

func TestScanHostAlive(t *testing.T) {
	plugin := &CmdPlugin{}
	plugin.setLogFile()
	res, _ := plugin.Exec("scan", "127.0.0.1", "127.0.0.1")
	m := res.MustMap()
	results := m.Get("data").(map[string]interface{})["output"].(string)
	if results == "[]" {
		t.Log("注意：本机未开放测试端口，存活检测通过但无扫描结果")
	}
}
