
build:
	CGO_ENABLED=0 go build -o ../command.so

windows:
	GOOS=windows CGO_ENABLED=0 GOARCH=amd64 go build -o ../command.dll

.PHONY:	clean
clean:
	rm -f "../command.so"
	rm -f "../command.dll"