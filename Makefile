
build:
	CGO_ENABLED=0 go build -o yaoapp/plugins/command.so

windows:
	GOOS=windows CGO_ENABLED=0 GOARCH=amd64 go build -o yaoapp/plugins/command.dll

.PHONY:	clean
clean:
	rm -f yaoapp/plugins/command.*
