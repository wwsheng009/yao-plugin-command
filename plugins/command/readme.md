# command runner

excute the system script from yao plugin

## build

```sh
GOARCH=amd64 GOOS=linux go build -o ../cmdt.so .

go build -o ../cmdt.so .
# windows
go build -o ../cmdt.dll .
```
## usage

call sh command with paramters
```
plugins.cmdt.sh para1 para2 para3 ...
```

call bash command with paramters
```
plugins.cmdt.bash para1 para2 para3 ...
```

call csh command with paramters
```
plugins.cmdt.csh para1 para2 para3 ...
```

call zsh command with paramters
```
plugins.cmdt.zsh para1 para2 para3 ...
```

call fish command with paramters
```
plugins.cmdt.fish para1 para2 para3 ...
```

run the script or other command
```
plugins.cmdt.run command para1 para2 para3 ...
```

call the command directly

Attention: the command name will convert to lowercase auto.
```
plugins.cmdt.<command> para1 para2 para3 ...
```

## remote

```
yao run plugins.cmdt.remote 172.18.3.234 22 root wwsheng850524 ls 

```

## test

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


