# Blockchain

> Cryptocurrency from scratch. Supplemented by training manuals

> More information about blockchain in the [youtube.com/watch?v=mp3I1HtEKfU](https://www.youtube.com/watch?v=mp3I1HtEKfU "Blockchain")


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
