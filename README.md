# Blockchain
Blockchain example. Template version.

### Compile:
```
$ make
```

### Run node and client:
```
$ ./node -serve::8080 -newuser:node.key -newchain:chain.db -loadaddr:addr.json
$ ./client -loaduser:node.key -loadaddr:addr.json
```
