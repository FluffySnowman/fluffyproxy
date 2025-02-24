# FluffyProxy

<!--toc:start-->
- [FluffyProxy](#fluffyproxy)
- [Installation](#installation)
  - [Using curl or wget](#using-curl-or-wget)
      - [Curl](#curl)
      - [Wget](#wget)
  - [From source](#from-source)
- [Usage](#usage)
  - [How does this work ?](#how-does-this-work)
  - [Example with the configuration language](#example-with-the-configuration-language)
  - [Example with cli arguments](#example-with-cli-arguments)
- [List of all available keys for the coniguration language](#list-of-all-available-keys-for-the-coniguration-language)
<!--toc:end-->

A proxy built in golang that allows exposing a local service to the
web behind NAT/firewalls etc. similar to cloudflare tunnels or FRP.

Great for minecraft servers/game servers, web servers and any other
tcp service that you want to expose to the web from your local network
without port forwarding or changing any firewall/nat configuration.

<hr />

<!-- > [!NOTE] -->
<!-- > Documentation is not complete yet. This project is still in -->
<!-- > development- do not expect things to work properly since this -->
<!-- > project is not meant to be collaborated with cos this is just a tool -->
<!-- > I've mady for myself and the code happens to be public. -->

# Installation

## Using curl or wget

The commands below download the latest release from the releases page
into /usr/local/bin/fp and make it executable.

#### Curl

```sh
sudo curl -L -o /usr/local/bin/fp https://github.com/FluffySnowman/fluffyproxy/releases/download/v1.0.0/fp_linux_amd64 && sudo chmod +x /usr/local/bin/fp
```

#### Wget

```sh
sudo wget -O /usr/local/bin/fp https://github.com/FluffySnowman/fluffyproxy/releases/download/v1.0.0/fp_linux_amd64 && sudo chmod +x /usr/local/bin/fp
```

## From source

```sh
git clone https://github.com/FluffySnowman/fluffyproxy
cd fluffyproxy
make go/release
```

The executable will be at `./release/fp`

# Usage

## How does this work ?

When a user connects to the server, the server forwards the connection
to the client. The client then forwards the connection to the local
service. The client and server communicate using a control connection.

This essentially allows exposeing a local service to the web behing a
firewall or NAT- similar to how cloudflare tunnels or FRP work.

Here's the network architecture of our example:

Running the server on system1 and the client on system2.

Service (website,any tcp service) is on `10.69.42.16:8000` on system2
and `0.0.0.0:42069` is the address of the control server on system1.
When anything connects to `192.168.1.96:8989` on system1, the
connection is then forwarded/proxied all the way to `10.69.42.16:8000`
on system2 and vice versa.

## Example with the configuration language

Create a file `fp_server` and `fp_client` (or any other name you can
remember) in the current directory on each respective machine.

```sh
touch fp_server
touch fp_client
```

Add the following to the fp_server file:

```javascript
# server listens for external connections on the below addy
listen 192.168.1.96:8989

# the addy of the control connection that the client connects to
control 0.0.0.0:42069

# allowing all ips (for clients) to connect
client_whitelist *

# allowing all external ips to connect/access the service
external_whitelist *
```

Its recommended to use the `client_whitelist` since anyone with the
IP:PORT of the control server can connect to the server and control
it. The `external_whitelist` is also recommended to prevent anyone
unauthorized from accessing the service.

IPs for the client and external whitelist should be comma seperated
such as `client_whitelist 127.0.0.1,192.168.1.10,10.69.42.2`

Add the following to the fp_client file:

```javascript
# addy of the internel service to expose to the web
local 10.69.42.16:8000

# address of the server control connection
server 0.0.0.0:42069
```

Start the server:

```sh
fp -server -f fp_server
```

Start the client:

```sh
fp -client -f fp_client
```

Now the service on `10.69.42.16:8000` should be accessible from `192.168.1.96:42000`

> [!NOTE]
> The default address for the SERVER control `0.0.0.0:42069` with the
> listen address being `0.0.0.0:7000`.

> [!NOTE]
> the default control address of the SERVER (FROM THE CLIENT) is
> `0.0.0.0:42069` with the default local service address (on the
> client) is `0.0.0.0:8080`.

## Example with cli arguments

The cli arguments do the exact same thing as the configuration
language and are just a faster way of getting things done.

Server:

(allowing all ips with `*`)

```sh
fp -server -listen '192.168.1.96:8989' -control '0.0.0.0:42069' -client-whitelist "*" -external-whitelist "*"
```

Just like the config file, the `-client_whitelist` and
`-external-whitelist` should have comma seperated IPs.

Client:

```sh
fp -client -server-control-addr '0.0.0.0:42069' -local '10.69.42.16:8000'
```

# List of all available keys for the coniguration language

Server

```javascript
listen <address>    # Example: 0.0.0.0:80  - listen for external connections.
control <address>   # Example: 10.0.1.7000 - connected to by client
client_whitelist    # Example: 127.0.0.1,192.168.1.10,10.69.42.2
external_whitelist  # Example: 1.1.1.1,8.8.8.8,192.168.1.69
```

Client

```javascript
local <address>     # Example: 0.0.0.0:3000 - address of the local service
server <address>    # Example: 10.0.1.7000  - address of the server control
```

<hr />

<!-- ### Install script -->
<!-- This script will download the latest release from the releases page -->
<!-- into your current directory and then move it to `/usr/local/bin/fp`. -->

<!-- ### From the release page -->

<!-- Download the latest release from the [releases -->
<!-- page](github.com/FluffySnowman/fluffyproxy/releases/) and place it -->
<!-- anywhere in your `$PATH` or in a place you'll remember. -->



