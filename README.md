# nocloud-net-server

nocloud-net-server is a HTTP server for [NoCloud datasource of cloud-init](https://cloudinit.readthedocs.io/en/latest/topics/datasources/nocloud.html).
It can serve different instance data based on client's IP address.

## Installation

### Binary releases (linux/amd64 only for now)

See the [releases page](https://github.com/ryot4/nocloud-net-server/releases).
Download and extract the archive, and then put the binary into the directory listed in `$PATH`.

### Build from source

Make sure [Go distribution is installed](https://golang.org/doc/install), and then run `go install`.

    $ go install github.com/ryot4/nocloud-net-server@latest

## Usage

At first, prepare the datasource directory and instance data to serve.
You can serve different data for different clients by creating directories named after their IP addresses as in the following example.

    datasource/
    ├── 192.168.1.23/  # Data for 192.168.1.23 (no vendor-data)
    │   ├── meta-data
    │   └── user-data
    ├── meta-data      # Data for other clients
    ├── user-data
    └── vendor-data

Then, run `nocloud-net-server`. You can specify the listen address and the path to the datasource directory.

    $ nocloud-net-server -l address:port -d /path/to/datasource

You can also pass parameters with environment variables.

    $ NOCLOUD_NET_LISTEN_ADDRESS=address:port NOCLOUD_NET_DATASOURCE_DIR=/path/to/datasource nocloud-net-server

## Running behind a reverse proxy

If `X-Forwarded-For` header is set in the request, the server uses the client address in the header to find instance data.
