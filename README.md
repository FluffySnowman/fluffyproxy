# FluffyProxy

A proxy built in golang that allows exposing a local service to the
web behind NAT/firewalls etc. similar to cloudflare tunnels or FRP.

<hr />

> [!NOTE]
> Documentation is not complete yet. This project is still in
> development- do not expect things to work properly since this
> project is not meant to be collaborated with cos this is just a tool
> I've mady for myself and the code happens to be public.

## Usage

Running the server on system1 and the client on system2.

The server listens for control messages from the client and once
connected, forwards connections on the listen port to the client and
then to the local service which is available on the network of the
client.

##### Build

Note: cd into `./src` to build the project.

```sh
go build -ldflags "-s -w" -trimpath -o fp main.go
```

#### Server

```sh
./fp -server -listen '192.168.1.96:8989' -control '0.0.0.0:42069'
```

#### Client

```sh
./fp -client -server-control-addr '0.0.0.0:42069' -local '10.69.42.16:8000'
```

Service (website,any tcp service) is on `10.69.42.16:8000` on system2
and `0.0.0.0:42069` is the address of the control server on system1.
When anything connects to `192.168.1.96:8989` on system1, the
connection is then forwarded/proxied all the way to `10.69.42.16:8000`
on system2 and vice versa.



