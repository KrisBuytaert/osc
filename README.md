
I was tired of running curl with different parameters to get the default health and other OpenSearch statuses.

I looked at opensearch-client, but it didn't support the basics — it just allowed me to run the curl commands again.

So I asked Claude to generate this little tool that reads a config file and makes life easier for me.

## Install

Download the latest binary for your platform from the [releases page](../../releases), or build from source:

```bash
make build   # linux/amd64, outputs ./osc
make install # copies to /usr/local/bin/osc
```

## Config

Default config path: `~/.osc-config.yaml`

```yaml
endpoint: https://localhost:9200
username: admin
password: admin
insecure: true                              # skip TLS verification
cert: /etc/opensearch/tls/admin-cert.pem   # optional client cert
key:  /etc/opensearch/tls/admin-key.pem    # optional client key
```

Generate a starter config:

```bash
osc generate-config
```

Override the config path or endpoint at runtime:

```bash
osc -c /path/to/config.yaml health
osc -e https://my-cluster:9200 health
```

## Commands

### Cluster

| Command | Description |
|---|---|
| `health` | Cluster health (status, shard counts, pending tasks) |
| `status` | Full cluster stats |
| `settings` | Cluster settings (persistent + transient) |
| `nodes` | Node list with roles, load, heap |
| `versions` | OpenSearch and JVM version per node |

### Indices

| Command | Description |
|---|---|
| `indices` | List all indices with health, doc count, store size |
| `index <name>` | Health for a specific index (wildcards work) |

### Shards

| Command | Description |
|---|---|
| `shards` | Full shard allocation table |
| `unassigned` | Only UNASSIGNED shards with their reason |
| `disk` | Disk used / available / percent per node |
| `recovery` | Active (non-done) shard recoveries |

### Allocation & maintenance

| Command | Description |
|---|---|
| `allocation-explain [body]` | Why a shard can't be allocated; pass a JSON body to target a specific shard |
| `reroute <body>` | POST to `_cluster/reroute`; body from argument or stdin |
| `rebalance` | Re-enable allocation (`enable: all`) and trigger `retry_failed` reroute |
| `drain <node>` | Exclude a node from shard allocation to drain it for maintenance |
| `undrain` | Clear all node allocation exclusions |

### ISM

| Command | Description |
|---|---|
| `ism` | List all ISM (Index State Management) policies |

### Raw requests

| Command | Description |
|---|---|
| `get <path>` | Raw GET — e.g. `get '/_cat/indices?v'` |
| `post <path> [body]` | Raw POST — body from argument or stdin |
| `put <path> [body]` | Raw PUT — body from argument or stdin |

### Other

| Command | Description |
|---|---|
| `generate-config` | Write a config file with current defaults |

## Options

```
-c, --config <path>   Config file path       (default: ~/.osc-config.yaml)
-e, --endpoint <url>  OpenSearch endpoint    (default: https://localhost:9200)
--cert <path>         Client certificate PEM (default: /etc/opensearch/tls/admin-cert.pem)
--key  <path>         Client key PEM         (default: /etc/opensearch/tls/admin-key.pem)
--curl                Print equivalent curl command to stderr before each request
```

## Examples

```bash
# Quick cluster health check
osc health

# Find out why the cluster is yellow
osc unassigned
osc allocation-explain

# Target a specific shard
osc allocation-explain '{"index":"my-index","shard":0,"primary":false}'

# Drain a node for maintenance, then restore
osc drain my-node-01
osc undrain

# Re-enable allocation after a maintenance window
osc rebalance

# Force-assign a stale primary (accept data loss)
osc reroute '{"commands":[{"allocate_stale_primary":{"index":"my-index","shard":0,"node":"my-node","accept_data_loss":true}}]}'

# Pipe a reroute body from a file
osc reroute < reroute.json

# Change a cluster setting
osc put /_cluster/settings '{"transient":{"cluster.routing.allocation.enable":"primaries"}}'

# Raw access to any API endpoint
osc get '/_cat/indices?v&h=index,pri,rep,docs.count,store.size&s=store.size:desc'
```

## Building from source

```bash
make build      # linux/amd64 static binary → ./osc
make build-all  # all platforms → dist/
make test       # unit tests
make lint       # gofmt + go vet
```
