

I was tired of running curl with different parameters to get the default health and other OpenSearch statuses.

I looked at opensearch-client .. but it didn't support the basics.. it just allowed me to run the curl commands again.

So I asked claude to generate me this little tool that reads a config file and makes life easier for me. 



```
OpenSearch CLI Tool

Usage: opensearch-cli [command] [options]

Commands:
  health              Show cluster health
  status              Show cluster status/stats
  indices             List all indices
  index <name>        Show health for specific index
  nodes               Show node information
  shards              Show shard allocation

Options:
  -c, --config <path> Config file path (default: ~/.osc-config.yaml)

Example config.yaml:
  endpoint: https://localhost:9200
  username: admin
  password: admin
  insecure: true
```
