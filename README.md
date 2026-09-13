# Kademlia DHT in Go

A Kademlia distributed hash table implementation written in Go. The
repository contains the routing-table and bucket implementation, RPC-based
node communication, iterative lookups, value storage, republishing, and local
simulation tooling.

## Features

- SHA-1 node and key identifiers
- K-bucket routing tables
- Ping, store, and find-value RPCs
- Iterative lookups with configurable parallelism (`a`)
- Persistent local storage backed by SQLite
- Key republishing and bucket refresh routines
- Unit tests for nodes, routing tables, priority lists, and XOR distance
- Local multi-node simulation scripts

## Requirements

- Go 1.24 or newer
- Python 3 for the simulation helpers
- `lsof` for the local process cleanup and node launcher scripts
- `screen` is useful for manually managing long-running node processes

Install Go dependencies from the repository root:

```bash
go mod download
```

## Test

Run all Go tests from the repository root:

```bash
go test ./...
```

To include coverage for the Kademlia package:

```bash
go test -cover ./kademlia
```

## Run Nodes

The simulation executable is the command-line entry point for both nodes and
clients. A bootstrap file contains one `IP:port` entry per line.

Build the executable expected by the helper scripts:

```bash
cd simulation
go build -o simulation .
```

Then start the local Kademlia network. The `a` parameter controls lookup
parallelism:

```bash
cd scripts/local
a=1
bash run_nodes.sh "$a"
```

The launcher reads node addresses from
`simulation/scripts/local/bootstraps/nodes_bootstraps/`, clears local node
databases, and writes node logs to `simulation/scripts/local/logs/`.

To start one node manually instead, run this from the `simulation` directory:

```bash
./simulation \
    -ip 127.0.0.1 \
    -port 40001 \
    -file ./scripts/local/bootstraps/nodes_bootstraps/bootstrap0.txt \
    -mode node \
    -a 1
```

Useful flags:

| Flag | Default | Description |
| --- | --- | --- |
| `-ip` | `127.0.0.1` | Address to bind |
| `-port` | `9000` | Listening port |
| `-file` | empty | Bootstrap file |
| `-mode` | `node` | `node` or `client` |
| `-a` | `1` | Lookup parallelism |
| `-delay` | `3` | Bootstrap delay in seconds |
| `-experiment` | `false` | Run a generated store/find experiment |
| `-keys` | `100` | Number of experiment keys |
| `-output` | `logs.txt` | Experiment output file |

## Run The Interactive Client

Start the nodes first, then open another terminal and run:

```bash
cd simulation
a=1
bash run_client.sh ./scripts/local/bootstraps/client_bootstraps/bootstrap0.txt "$a"
```

The client connects using the bootstrap file and writes stderr output to
`simulation/client.log`.

The client supports these commands:

```text
store <key> <value>
find <key>
quit
```

For example:

```text
store hello world
find hello
```

## Generate Bootstrap Files

Bootstrap files can be generated for a local network:

```bash
cd simulation/scripts/local

python3 bootstrap.py \
    --ip 127.0.0.1 \
    --start-port 40001 \
    --count 10 \
    --output-dir ./bootstraps/generated_nodes \
    --chunk-size 2 \
    --overlap 1
```

Each generated file contains the addresses that a node can use while joining
the network. Keep the generated files aligned with the ports assigned to the
nodes.

## Run A Local Simulation

The Python launcher can also be called directly from
`simulation/scripts/local/`:

```bash
python3 run_nodes.py \
    --exec ../../simulation \
    --ip 127.0.0.1 \
    --boot-dir ./bootstraps/nodes_bootstraps \
    --out-dir ./logs \
    --a 1
```

## Run Experiments

Run a generated store/find workload from the `simulation` directory. The
experiment creates random keys, stores values through the network, and then
looks them up:

```bash
cd simulation
a=1
./simulation \
    -file ./scripts/local/bootstraps/client_bootstraps/bootstrap0.txt \
    -mode client \
    -a "$a" \
    -output output.txt \
    -experiment \
    -keys 100 \
    -delay 0
```

The experiment output is written to `simulation/output.txt`.

## Project Layout

```text
kademlia/                         Core DHT implementation and tests
logging/                          Shared logging helpers
simulation/simulation.go          Node, client, and experiment entry point
simulation/client.go              Interactive client commands
simulation/scripts/local/         Local bootstrap and process helpers
simulation/scripts/remote/        Remote machine helpers
simulation/scripts/local/logs/    Generated simulation logs
dbs/                              Generated SQLite node databases
```

The `dbs/` and log directories are runtime output. Do not commit generated
databases, logs, machine lists, or IP-address reports.

## Remote Simulation Scripts

The scripts under `simulation/scripts/remote` are intended for a specific
university lab environment. They use the local `pass` password manager to read
the SSH password and expect machine names in `./output/machines.txt`.

Review and configure the username, machine list, remote paths, and SSH host-key
policy before using them. They are not required for local development.

## Collaborators

- [kmyrto](https://github.com/kmyrto)
