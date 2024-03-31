# Blockchain

> Cryptocurrency from scratch. Supplemented by training manuals
> 
> More information about blockchain in the [youtube.com/watch?v=mp3I1HtEKfU](https://www.youtube.com/watch?v=mp3I1HtEKfU "Blockchain")

> [!IMPORTANT]
> The current implementation in the master branch is irrelevant for the modern version of the Go language. To compile programs successfully, use the [goup](https://github.com/number571/blockchain/tree/goup) branch.

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
