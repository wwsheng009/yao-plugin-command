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
