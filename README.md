# FluffyProxy

A proxy built in golang that allows exposing a local service to the
web behind NAT/firewalls etc. similar to cloudflare tunnels or FRP.

<hr />

> [!NOTE]
> Documentation is not complete yet. This project is still in
> development- do not expect things to work properly since this
> project is not meant to be collaborated with cos this is just a tool
> I've mady for myself and the code happens to be public.

## Installation

### Using curl or wget

The commands below download the latest release from the releases page
into /usr/local/bin/fp and make it executable.

##### Curl

```sh
sudo curl -L -o /usr/local/bin/fp https://github.com/FluffySnowman/fluffyproxy/releases/download/v0.1.0/fp_linux_amd64 && sudo chmod +x /usr/local/bin/fp
```

##### Wget

```sh
sudo wget -O /usr/local/bin/fp https://github.com/FluffySnowman/fluffyproxy/releases/download/v0.1.0/fp_linux_amd64 && sudo chmod +x /usr/local/bin/fp
```

### From source

```sh
git clone https://github.com/FluffySnowman/fluffyproxy
cd fluffyproxy
make go/release
```

The executable will be at `./release/fp`

## Usage

### How does this work ?

When a user connects to the server, the server forwards the connection
to the client. The client then forwards the connection to the local
service. The client and server communicate using a control connection.

This essentially allows exposeing a local service to the web behing a
firewall or NAT- similar to how cloudflare tunnels or FRP work.

Here's the network architecture of our example-

Running the server on system1 and the client on system2.

Service (website,any tcp service) is on `10.69.42.16:8000` on system2
and `0.0.0.0:42069` is the address of the control server on system1.
When anything connects to `192.168.1.96:8989` on system1, the
connection is then forwarded/proxied all the way to `10.69.42.16:8000`
on system2 and vice versa.

### Example with the configuration language

Create a file `server` and `client` (or any other name you can
remember) in the current directory on each respective machine.

```sh
touch server
touch client
```

Add the following to the server file:

```conf
# server listens for external connections on the below addy
listen 192.168.1.96:7001

# the addy of the control connection that the client connects to
control 192.168.1.96:7070
```

Add the following to the client file:

```conf
# addy of the internel service to expose to the web
local 10.69.42.16:8000

# address of the server control connection
server 192.168.1.96:7070
```

To start the server and client-

server:

```sh
```

```sh
```

### Example with cli arguments

The cli arguments do the exact same thing as the configuration
language and are just a faster way of getting things done.

Server:

```sh
fp -server -listen '192.168.1.96:8989' -control '0.0.0.0:42069'
```

Client:

```sh
fp -client -server-control-addr '0.0.0.0:42069' -local '10.69.42.16:8000'
```

<!-- ### Install script -->
<!-- This script will download the latest release from the releases page -->
<!-- into your current directory and then move it to `/usr/local/bin/fp`. -->

<!-- ### From the release page -->

<!-- Download the latest release from the [releases -->
<!-- page](github.com/FluffySnowman/fluffyproxy/releases/) and place it -->
<!-- anywhere in your `$PATH` or in a place you'll remember. -->



