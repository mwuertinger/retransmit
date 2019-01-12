# Retransmit
This tool allows to transmit data streams from UNIX shell pipelines over
unreliable connections. It works by trying to connect to the destination in an
infinite loop until all of the data has been transmitted. Should the connection
drop or hang the connection is re-established and the data is sent again.
Sequence numbers and checksums ensure that data is transferred reliably.

## Setup
You need to have [Go installed](https://golang.org/doc/install) in order to
build this application. Then you can simply run go get:
```
go get github.com/mwuertinger/retransmit
```

## Usage
The tool can operate in a send and a receive mode and has to run on both source
and destination hosts:
```
source$ some-program | retransmit send destination:1234 
```
```
destination$ retransmit recv 0.0.0.0:1234 | some-other-program
```

## Example
An example use case is to transfer a large BTRFS snapshot over an unreliable
connection. Normally one would just pipe the data strem through an SSH
connection:
```
btrfs send /path/to/snapshot | ssh destination btrfs receive /path/to/filesystem
```
However, this approach will not work if the filesystem is large and the
connection is unreliable. Since the data transfer cannot be resumed any problem
with the SSH connection will require a complete restart of the process.

This problem can solved by using `retransmit` on both sides:
```
source$ btrfs send /path/to/snapshot | retransmit send destination:10000
```
On the destination end start `retransmit` in receiving mode and pipe the data
into the target process:
```
destination$ retransmit recv 0.0.0.0:10000 | btrfs receive /path/to/filesystem
```

## Security
This tool follows the UNIX philosophy and concentrates on doing one job well.
Therefore it does not include encryption or authentication mechanisms. If you
are transmitting data over untrusted networks you should use a VPN or an SSH
tunnel.

## Performance
The algorithm for data transmission is very primitive. Data is packaged up into
frames which are sent sequentially over a TCP connection. Once a frame is
transmitted completely the sender waits for a confirmation frame from the
received. This behavior leads to sub-optimal bandwidth usage as there's a time
gap between the frames where no data is transmitted.