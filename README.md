# cluster
c-lightning node manager in a terminal

# running

This program expects your c-lightning `lightning-rpc` socket to be in your
current directory. If it's somewhere else you can pass the location using:

    go run . --rpc=/path/to/lightning-rpc

# running on localhost with a remote c-lightning node

You can use `socat` to teleport the remote socket to localhost.
To do this (assuming your c-lightning node is listening on 192.168.1.10)

### on the remote:
`socat TCP-LISTEN:3333,reuseaddr,fork UNIX-CONNECT:lightning-rpc`

### on localhost:
`socat UNIX-LISTEN:lightning-rpc,reuseaddr,fork TCP4:192.168.1.10:3333`

Bear in mind the above socat commands will expose your c-lightning RPC via port 3333 to anyone on your network and _access all your c-lightning funds_ so make sure your network is protected from unauthorized access.
