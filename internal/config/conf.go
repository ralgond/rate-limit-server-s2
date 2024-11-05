package config

import (
	"encoding/xml"
	"fmt"
	"github.com/libertylocked/urlpattern"
	"os"
	"strings"
)

type LogConfig struct {
	Path  string `xml:"path,attr"`
	Level string `xml:"level,attr"`
}

type RateLimitURLConfig struct {
	Value            string `xml:"value,attr"`
	LimitParamString string `xml:"limit_param,attr"`
	Method           string `xml:"method,attr"`
	Pattern          *urlpattern.Pattern
	LimitParam       []string
}

type RateLimitWithIPConfig struct {
	RateLimitURL []*RateLimitURLConfig `xml:"url"`
}

type RateLimitWithTokenConfig struct {
	RateLimitURL []*RateLimitURLConfig `xml:"url"`
}

type TimeoutItemConfig struct {
	Value int `xml:"value,attr"`
}

type TimeoutConfig struct {
	FrontendReadHeaderTimeoutMS TimeoutItemConfig `xml:"frontend_read_header"`
	FrontendReadTimeoutMS       TimeoutItemConfig `xml:"frontend_read"`
	FrontendWriteTimeoutMS      TimeoutItemConfig `xml:"frontend_write"`
	FrontendIdleTimeoutMS       TimeoutItemConfig `xml:"frontend_idle"`
}

type RateLimitRedisConfig struct {
	Address  string `xml:"addr,attr"`
	PoolSize int    `xml:"pool_size,attr"`
	Cluster  bool   `xml:"cluster,attr"`
}

type UserSessionRedisConfig struct {
	Address  string `xml:"addr,attr"`
	PoolSize int    `xml:"pool_size,attr"`
	Cluster  bool   `xml:"cluster,attr"`
}

type ServerConfig struct {
	Address string `xml:"addr,attr"`
}

type LoadBalanceModeConfig struct {
	Value string `xml:"value,attr"`
}

type MaxConnectionsPerHostConfig struct {
	Value int `xml:"value,attr"`
}

type BackendConfig struct {
	MaxConnectionsPerHost MaxConnectionsPerHostConfig `xml:"max_connections_per_host"`
	LoadBalanceMode       LoadBalanceModeConfig       `xml:"load_balance_mode"`
	Servers               []ServerConfig              `xml:"server"`
}

type FrontendConfig struct {
	Address        string `xml:"bind,attr"`
	MaxConcurrency int    `xml:"maxcon,attr"`
}

type Config struct {
	Frontend           FrontendConfig           `xml:"frontend"`
	Backend            BackendConfig            `xml:"backend"`
	UserSessionRedis   UserSessionRedisConfig   `xml:"user_session_redis"`
	RateLimitRedis     RateLimitRedisConfig     `xml:"rate_limit_redis"`
	RateLimitWithIP    RateLimitWithIPConfig    `xml:"rate_limit_with_ip"`
	RateLimitWithToken RateLimitWithTokenConfig `xml:"rate_limit_with_token"`
	Timeout            TimeoutConfig            `xml:"timeout"`
	Log                LogConfig                `xml:"log"`
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		fmt.Println("Error decoding XML:", err)
		return nil, err
	}

	fmt.Printf("frontend bind: %s\n", config.Frontend.Address)
	fmt.Printf("frontend maxcon: %d\n", config.Frontend.MaxConcurrency)
	fmt.Printf("backend.max_connections_per_host: %d\n", config.Backend.MaxConnectionsPerHost.Value)
	fmt.Printf("backend.load_balance_mode: %s\n", config.Backend.LoadBalanceMode.Value)
	for i, server := range config.Backend.Servers {
		fmt.Printf("backend.server: [%d], %s\n", i, server.Address)
	}
	fmt.Printf("user_session_redis: addr=%s, pool_size=%d, cluster=%v\n",
		config.UserSessionRedis.Address, config.UserSessionRedis.PoolSize, config.UserSessionRedis.Cluster)
	fmt.Printf("rate_limit_redis: addr=%s, pool_size=%d, cluster=%v\n",
		config.RateLimitRedis.Address, config.RateLimitRedis.PoolSize, config.RateLimitRedis.Cluster)

	for i, url := range config.RateLimitWithIP.RateLimitURL {
		url.LimitParam = strings.Split(url.LimitParamString, " ")
		if len(url.LimitParam) < 3 {
			fmt.Println("length of url.LimitParam should ge 3")
			os.Exit(-1)
		}
		url.Pattern = urlpattern.NewPattern().Path(url.Value)
		fmt.Printf("rate_limit_with_ip: [%d], %s, %s \"%s\" %p\n",
			i, url.Value, url.Method, url.LimitParamString, url.Pattern)
	}

	for i, url := range config.RateLimitWithToken.RateLimitURL {
		url.LimitParam = strings.Split(url.LimitParamString, " ")
		if len(url.LimitParam) < 3 {
			fmt.Println("length of url.LimitParam should ge 3")
			os.Exit(-1)
		}
		url.Pattern = urlpattern.NewPattern().Path(url.Value)
		fmt.Printf("rate_limit_with_token: [%d], %s, %s \"%s\" %p\n",
			i, url.Value, url.Method, url.LimitParamString, url.Pattern)
	}

	fmt.Printf("read header timeout: %d\n", config.Timeout.FrontendReadHeaderTimeoutMS.Value)
	fmt.Printf("read timeout: %d\n", config.Timeout.FrontendReadTimeoutMS.Value)
	fmt.Printf("write timeout: %d\n", config.Timeout.FrontendWriteTimeoutMS.Value)
	fmt.Printf("idle timeout: %d\n", config.Timeout.FrontendIdleTimeoutMS.Value)
	return &config, nil
}
