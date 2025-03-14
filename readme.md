# command runner

excute the system script from yao plugin

## build

```sh
GOARCH=amd64 GOOS=linux go build -o -o yaoapp/plugins/cmdt.so .

# linux
go build -o yaoapp/plugins/cmdt.so .

# windows
go build -o yaoapp/plugins/cmdt.dll .
```
## usage

call sh command with paramters
```sh
yao run plugins.cmdt.sh para1 para2 para3 ...
```

call bash command with paramters
```
yao run plugins.cmdt.bash para1 para2 para3 ...
```

call csh command with paramters
```
yao run plugins.cmdt.csh para1 para2 para3 ...
```

call zsh command with paramters
```
yao run plugins.cmdt.zsh para1 para2 para3 ...
```

call fish command with paramters
```
yao run plugins.cmdt.fish para1 para2 para3 ...
```

run the script or other command
```
yao run plugins.cmdt.run command para1 para2 para3 ...
```

call the command directly

Attention: the command name will convert to lowercase auto.
```
yao run plugins.cmdt.<command> para1 para2 para3 ...
```

## remote

```
yao run plugins.cmdt.remote 172.18.3.234 22 root wwsheng850524 ls 

```

## test

windows

```cmd
yao run plugins.cmdt.cmd "dir"

```

linux

```sh

yao run plugins.cmdt.sh "cd ~;ls -la"

# run date
yao run plugins.cmdt.run "date" 

# failed
yao run plugins.cmdt.sh "top" 

# run script file
yao run plugins.cmdt.sh "~/demo.sh" 

# execute sh command
yao run plugins.cmdt.run "sh" "/root/demo.sh" 

# command line with "-c" flag not supported 
yao run plugins.cmdt.run "sh" "-c" "/root/demo.sh" 

# run script file directly
yao run plugins.cmdt.run "/root/demo.sh" 

# run the date command directly
yao run plugins.cmdt.date
```


