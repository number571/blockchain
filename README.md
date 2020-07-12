# Blockchain
> Blockchain example. 

### Compile:
```
$ make
```

### Run nodes and client:
```
$ ./node -serve::8080 -newuser:node1.key -newchain:chain1.db -loadaddr:addr.json
$ ./node -serve::9090 -newuser:node2.key -newchain:chain2.db -loadaddr:addr.json
$ ./client -loaduser:node1.key -loadaddr:addr.json
```
