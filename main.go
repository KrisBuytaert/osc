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
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func newClient(config Config) *OSClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.Insecure},
	}

	return &OSClient{
		config: config,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

func (c *OSClient) request(method, path string) ([]byte, error) {
	url := strings.TrimRight(c.config.Endpoint, "/") + path
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
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
	data, err := c.request("GET", "/_cluster/health")
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
	data, err := c.request("GET", "/_cluster/stats")
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
	data, err := c.request("GET", "/_cat/indices?format=json&h=health,status,index,uuid,pri,rep,docs.count,docs.deleted,store.size,pri.store.size")
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
	data, err := c.request("GET", path)
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
	data, err := c.request("GET", "/_cat/nodes?v")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *OSClient) GetShards() (string, error) {
	data, err := c.request("GET", "/_cat/shards?v")
	if err != nil {
		return "", err
	}

	return string(data), nil
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
	fmt.Println("  health              Show cluster health")
	fmt.Println("  status              Show cluster status/stats")
	fmt.Println("  indices             List all indices")
	fmt.Println("  index <name>        Show health for specific index")
	fmt.Println("  nodes               Show node information")
	fmt.Println("  shards              Show shard allocation")
	fmt.Println("\nOptions:")
	fmt.Println("  -c, --config <path> Config file path (default: ~/.osc-config.yaml)")
	fmt.Println("\nExample config.yaml:")
	fmt.Println("  endpoint: https://localhost:9200")
	fmt.Println("  username: admin")
	fmt.Println("  password: admin")
	fmt.Println("  insecure: true")
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
	var command string
	var indexName string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		if args[i] == "-c" || args[i] == "--config" {
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		} else if command == "" {
			command = args[i]
		} else if indexName == "" {
			indexName = args[i]
		}
	}

	if command == "" {
		printUsage()
		os.Exit(1)
	}

	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	client := newClient(*config)

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

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
