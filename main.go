// main.go
package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Endpoint string `yaml:"endpoint"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Insecure bool   `yaml:"insecure"`
	CertFile string `yaml:"cert"`
	KeyFile  string `yaml:"key"`
}

type ClusterHealth struct {
	ClusterName         string  `json:"cluster_name"`
	Status              string  `json:"status"`
	NumberOfNodes       int     `json:"number_of_nodes"`
	NumberOfDataNodes   int     `json:"number_of_data_nodes"`
	ActivePrimaryShards int     `json:"active_primary_shards"`
	ActiveShards        int     `json:"active_shards"`
	RelocatingShards    int     `json:"relocating_shards"`
	InitializingShards  int     `json:"initializing_shards"`
	UnassignedShards    int     `json:"unassigned_shards"`
	DelayedUnassigned   int     `json:"delayed_unassigned_shards"`
	PendingTasks        int     `json:"number_of_pending_tasks"`
	InFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitTime     int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercent float64 `json:"active_shards_percent_as_number"`
}

type IndexHealth struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	UUID         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
}

type OSClient struct {
	config     Config
	httpClient *http.Client
	showCurl   bool
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func newClient(config Config, showCurl bool) (*OSClient, error) {
	tlsCfg := &tls.Config{InsecureSkipVerify: config.Insecure}

	if config.CertFile != "" && config.KeyFile != "" {
		_, certErr := os.Stat(config.CertFile)
		_, keyErr := os.Stat(config.KeyFile)
		if certErr == nil && keyErr == nil {
			cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client cert/key: %w", err)
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
		}
	}

	return &OSClient{
		config:   config,
		showCurl: showCurl,
		httpClient: &http.Client{
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
			Timeout:   120 * time.Second,
		},
	}, nil
}

func (c *OSClient) buildCurlCommand(method, path string, body []byte) string {
	url := strings.TrimRight(c.config.Endpoint, "/") + path

	var curlCmd strings.Builder
	curlCmd.WriteString("curl -X ")
	curlCmd.WriteString(method)

	if c.config.Username != "" && c.config.Password != "" {
		curlCmd.WriteString(" -u '")
		curlCmd.WriteString(c.config.Username)
		curlCmd.WriteString(":")
		curlCmd.WriteString(c.config.Password)
		curlCmd.WriteString("'")
	}

	if c.config.Insecure {
		curlCmd.WriteString(" -k")
	}

	if c.config.CertFile != "" && c.config.KeyFile != "" {
		curlCmd.WriteString(" --cert '")
		curlCmd.WriteString(c.config.CertFile)
		curlCmd.WriteString("' --key '")
		curlCmd.WriteString(c.config.KeyFile)
		curlCmd.WriteString("'")
	}

	if len(body) > 0 {
		curlCmd.WriteString(" -H 'Content-Type: application/json' -d '")
		curlCmd.WriteString(string(body))
		curlCmd.WriteString("'")
	}

	curlCmd.WriteString(" '")
	curlCmd.WriteString(url)
	curlCmd.WriteString("'")

	return curlCmd.String()
}

func (c *OSClient) request(method, path string, body []byte) ([]byte, error) {
	if c.showCurl {
		fmt.Fprintf(os.Stderr, "\n# Equivalent curl command:\n%s\n\n", c.buildCurlCommand(method, path, body))
	}

	url := strings.TrimRight(c.config.Endpoint, "/") + path
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = strings.NewReader(string(body))
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func stdinIsTTY() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func prettifyJSON(data []byte) (string, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}

func (c *OSClient) GetClusterHealth() (*ClusterHealth, error) {
	data, err := c.request("GET", "/_cluster/health", nil)
	if err != nil {
		return nil, err
	}

	var health ClusterHealth
	if err := json.Unmarshal(data, &health); err != nil {
		return nil, err
	}

	return &health, nil
}

func (c *OSClient) GetClusterStatus() (map[string]interface{}, error) {
	data, err := c.request("GET", "/_cluster/stats", nil)
	if err != nil {
		return nil, err
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

func (c *OSClient) ListIndices() ([]IndexHealth, error) {
	data, err := c.request("GET", "/_cat/indices?format=json&h=health,status,index,uuid,pri,rep,docs.count,docs.deleted,store.size,pri.store.size", nil)
	if err != nil {
		return nil, err
	}

	var indices []IndexHealth
	if err := json.Unmarshal(data, &indices); err != nil {
		return nil, err
	}

	return indices, nil
}

func (c *OSClient) GetIndexHealth(index string) ([]IndexHealth, error) {
	path := fmt.Sprintf("/_cat/indices/%s?format=json&h=health,status,index,uuid,pri,rep,docs.count,docs.deleted,store.size,pri.store.size", index)
	data, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var indices []IndexHealth
	if err := json.Unmarshal(data, &indices); err != nil {
		return nil, err
	}

	return indices, nil
}

func (c *OSClient) GetNodes() (string, error) {
	data, err := c.request("GET", "/_cat/nodes?v", nil)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *OSClient) GetShards() (string, error) {
	data, err := c.request("GET", "/_cat/shards?v", nil)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *OSClient) GetUnassignedShards() (string, error) {
	data, err := c.request("GET", "/_cat/shards?v&h=index,shard,prirep,state,unassigned.reason,node", nil)
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	var result []string
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[0] == "index" || fields[3] == "UNASSIGNED" {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n"), nil
}

func (c *OSClient) GetVersions() (string, error) {
	data, err := c.request("GET", "/_cat/nodes?v&h=name,version,build.type,jvm.version", nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *OSClient) GetDiskUsage() (string, error) {
	data, err := c.request("GET", "/_cat/nodes?v&h=name,ip,diskUsed,diskAvail,diskUsedPercent", nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *OSClient) GetRecovery() (string, error) {
	data, err := c.request("GET", "/_cat/recovery?v&h=index,shard,type,stage,source_node,target_node,bytes_percent,files_percent", nil)
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	var result []string
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[3] != "done" {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n"), nil
}

func printClusterHealth(health *ClusterHealth) {
	fmt.Println("\n=== Cluster Health ===")
	data, _ := json.MarshalIndent(health, "", "  ")
	fmt.Println(string(data))
}

func printIndices(indices []IndexHealth) {
	fmt.Println("\n=== Indices ===")
	data, _ := json.MarshalIndent(indices, "", "  ")
	fmt.Println(string(data))
}

func printUsage() {
	fmt.Println("OpenSearch CLI Tool")
	fmt.Println("\nUsage: opensearch-cli [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  health                     Show cluster health")
	fmt.Println("  status                     Show cluster status/stats")
	fmt.Println("  settings                   Show cluster settings")
	fmt.Println("  indices                    List all indices")
	fmt.Println("  index <name>               Show health for specific index")
	fmt.Println("  nodes                      Show node information")
	fmt.Println("  versions                   Show OpenSearch and JVM version per node")
	fmt.Println("  shards                     Show shard allocation")
	fmt.Println("  unassigned                 Show only unassigned shards with reasons")
	fmt.Println("  disk                       Show disk usage per node")
	fmt.Println("  recovery                   Show active (non-done) shard recoveries")
	fmt.Println("  allocation-explain [body]  Explain shard allocation; optional JSON body targets a shard")
	fmt.Println("  reroute <body>             POST _cluster/reroute; body from arg or stdin")
	fmt.Println("  ism                        List ISM policies")
	fmt.Println("  rebalance                  Re-enable allocation and retry failed shard assignments")
	fmt.Println("  drain <node>               Exclude a node from receiving shards (triggers drain)")
	fmt.Println("  undrain                    Clear node allocation exclusions")
	fmt.Println("  generate-config            Write a config file with current defaults")
	fmt.Println("  get <path>                 Raw GET request (e.g. get /_cat/indices?v)")
	fmt.Println("  post <path> [body]         Raw POST request; body from arg or stdin")
	fmt.Println("  put  <path> [body]         Raw PUT request; body from arg or stdin")
	fmt.Println("\nOptions:")
	fmt.Println("  -c, --config <path>    Config file path (default: ~/.osc-config.yaml)")
	fmt.Println("  -e, --endpoint <url>   OpenSearch endpoint (default: https://localhost:9200)")
	fmt.Println("  --cert <path>          Client certificate PEM (default: /etc/opensearch/tls/admin-cert.pem)")
	fmt.Println("  --key  <path>          Client key PEM       (default: /etc/opensearch/tls/admin-key.pem)")
	fmt.Println("  --curl                 Print equivalent curl command to stderr")
	fmt.Println("\nExample config.yaml:")
	fmt.Println("  endpoint: https://localhost:9200")
	fmt.Println("  username: admin")
	fmt.Println("  password: admin")
	fmt.Println("  insecure: true")
	fmt.Println("  cert: /etc/opensearch/tls/admin-cert.pem")
	fmt.Println("  key:  /etc/opensearch/tls/admin-key.pem")
}

func getDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return homeDir + "/.osc-config.yaml"
}

func main() {
	args := os.Args[1:]

	configPath := getDefaultConfigPath()
	explicitConfig := false
	endpoint := "https://localhost:9200"
	certFile := "/etc/opensearch/tls/admin-cert.pem"
	keyFile := "/etc/opensearch/tls/admin-key.pem"
	var command string
	var indexName string
	var extraArg string
	var showCurl bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c", "--config":
			if i+1 < len(args) {
				configPath = args[i+1]
				explicitConfig = true
				i++
			}
		case "-e", "--endpoint":
			if i+1 < len(args) {
				endpoint = args[i+1]
				i++
			}
		case "--cert":
			if i+1 < len(args) {
				certFile = args[i+1]
				i++
			}
		case "--key":
			if i+1 < len(args) {
				keyFile = args[i+1]
				i++
			}
		case "--curl":
			showCurl = true
		default:
			if command == "" {
				command = args[i]
			} else if indexName == "" {
				indexName = args[i]
			} else if extraArg == "" {
				extraArg = args[i]
			}
		}
	}

	command = strings.ToLower(command)

	if command == "" {
		printUsage()
		os.Exit(1)
	}

	config, err := loadConfig(configPath)
	if err != nil {
		if explicitConfig || !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		config = &Config{}
	}

	// CLI flags fill in defaults not set in config file
	if config.Endpoint == "" {
		config.Endpoint = endpoint
	}
	if config.CertFile == "" {
		config.CertFile = certFile
	}
	if config.KeyFile == "" {
		config.KeyFile = keyFile
	}

	if command == "generate-config" {
		if _, err := os.Stat(configPath); err == nil {
			fmt.Fprintf(os.Stderr, "Config file already exists: %s\n", configPath)
			os.Exit(1)
		}
		data, err := yaml.Marshal(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating config: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config written to %s\n", configPath)
		os.Exit(0)
	}

	client, err := newClient(*config, showCurl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "health":
		health, err := client.GetClusterHealth()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printClusterHealth(health)

	case "status":
		stats, err := client.GetClusterStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println("\n=== Cluster Status ===")
		fmt.Println(string(data))

	case "indices":
		indices, err := client.ListIndices()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printIndices(indices)

	case "index":
		if indexName == "" {
			fmt.Fprintf(os.Stderr, "Error: index name required\n")
			os.Exit(1)
		}
		indices, err := client.GetIndexHealth(indexName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printIndices(indices)

	case "nodes":
		nodes, err := client.GetNodes()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n=== Nodes ===")
		fmt.Println(nodes)

	case "shards":
		shards, err := client.GetShards()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n=== Shards ===")
		fmt.Println(shards)

	case "versions":
		data, err := client.GetVersions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n=== Node Versions ===")
		fmt.Println(data)

	case "unassigned":
		data, err := client.GetUnassignedShards()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n=== Unassigned Shards ===")
		fmt.Println(data)

	case "disk":
		data, err := client.GetDiskUsage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n=== Disk Usage per Node ===")
		fmt.Println(data)

	case "recovery":
		data, err := client.GetRecovery()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n=== Active Recoveries ===")
		fmt.Println(data)

	case "settings":
		data, err := client.request("GET", "/_cluster/settings?pretty", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n=== Cluster Settings ===")
		if pretty, err := prettifyJSON(data); err == nil {
			fmt.Println(pretty)
		} else {
			fmt.Println(string(data))
		}

	case "allocation-explain":
		method := "GET"
		var body []byte
		if indexName != "" {
			method = "POST"
			body = []byte(indexName)
		}
		data, err := client.request(method, "/_cluster/allocation/explain?pretty", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n=== Allocation Explain ===")
		if pretty, err := prettifyJSON(data); err == nil {
			fmt.Println(pretty)
		} else {
			fmt.Println(string(data))
		}

	case "reroute":
		var body []byte
		if indexName != "" {
			body = []byte(indexName)
		} else {
			if stdinIsTTY() {
				fmt.Fprintf(os.Stderr, "Error: reroute body required (pass as argument or pipe via stdin)\n")
				os.Exit(1)
			}
			body, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
		}
		if len(body) == 0 {
			fmt.Fprintf(os.Stderr, "Error: reroute body required\n")
			os.Exit(1)
		}
		data, err := client.request("POST", "/_cluster/reroute?pretty", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if pretty, err := prettifyJSON(data); err == nil {
			fmt.Println(pretty)
		} else {
			fmt.Println(string(data))
		}

	case "put":
		if indexName == "" {
			fmt.Fprintf(os.Stderr, "Error: path required\n")
			os.Exit(1)
		}
		var body []byte
		if extraArg != "" {
			body = []byte(extraArg)
		} else {
			if stdinIsTTY() {
				fmt.Fprintf(os.Stderr, "Error: body required (pass as argument or pipe via stdin)\n")
				os.Exit(1)
			}
			body, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
		}
		data, err := client.request("PUT", indexName, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if pretty, err := prettifyJSON(data); err == nil {
			fmt.Println(pretty)
		} else {
			fmt.Println(string(data))
		}

	case "ism":
		data, err := client.request("GET", "/_plugins/_ism/policies", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n=== ISM Policies ===")
		if pretty, err := prettifyJSON(data); err == nil {
			fmt.Println(pretty)
		} else {
			fmt.Println(string(data))
		}

	case "rebalance":
		_, err = client.request("PUT", "/_cluster/settings", []byte(`{"transient":{"cluster.routing.allocation.enable":"all"}}`))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error enabling allocation: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("cluster.routing.allocation.enable set to all")
		data, err := client.request("POST", "/_cluster/reroute?retry_failed=true", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error triggering reroute: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Reroute triggered:")
		if pretty, err := prettifyJSON(data); err == nil {
			fmt.Println(pretty)
		} else {
			fmt.Println(string(data))
		}

	case "drain":
		if indexName == "" {
			fmt.Fprintf(os.Stderr, "Error: node name required\n")
			os.Exit(1)
		}
		body := fmt.Sprintf(`{"transient":{"cluster.routing.allocation.exclude._name":%q}}`, indexName)
		data, err := client.request("PUT", "/_cluster/settings", []byte(body))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Node %q excluded from shard allocation\n", indexName)
		if pretty, err := prettifyJSON(data); err == nil {
			fmt.Println(pretty)
		} else {
			fmt.Println(string(data))
		}

	case "undrain":
		data, err := client.request("PUT", "/_cluster/settings", []byte(`{"transient":{"cluster.routing.allocation.exclude._name":null,"cluster.routing.allocation.exclude._ip":null}}`))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Node exclusions cleared")
		if pretty, err := prettifyJSON(data); err == nil {
			fmt.Println(pretty)
		} else {
			fmt.Println(string(data))
		}

	case "get":
		if indexName == "" {
			fmt.Fprintf(os.Stderr, "Error: path required\n")
			os.Exit(1)
		}
		data, err := client.request("GET", indexName, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if pretty, err := prettifyJSON(data); err == nil {
			fmt.Println(pretty)
		} else {
			fmt.Println(string(data))
		}

	case "post":
		if indexName == "" {
			fmt.Fprintf(os.Stderr, "Error: path required\n")
			os.Exit(1)
		}
		var body []byte
		if extraArg != "" {
			body = []byte(extraArg)
		} else {
			if stdinIsTTY() {
				fmt.Fprintf(os.Stderr, "Error: body required (pass as argument or pipe via stdin)\n")
				os.Exit(1)
			}
			body, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
		}
		data, err := client.request("POST", indexName, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if pretty, err := prettifyJSON(data); err == nil {
			fmt.Println(pretty)
		} else {
			fmt.Println(string(data))
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
