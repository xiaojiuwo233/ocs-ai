package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 应用配置结构体
type Config struct {
	Server struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	} `mapstructure:"server"`

	AI struct {
		BaseURL        string  `mapstructure:"base_url"`
		APIKey         string  `mapstructure:"api_key"`
		Model          string  `mapstructure:"model"`
		SystemPrompt   string  `mapstructure:"system_prompt"`
		PromptTemplate string  `mapstructure:"prompt_template"`
		Temperature    float64 `mapstructure:"temperature"`
		TopP           float64 `mapstructure:"top_p"`
		MaxTokens      int     `mapstructure:"max_tokens"`
		Timeout        int     `mapstructure:"timeout"`
	} `mapstructure:"ai"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	var config Config

	// 确保配置文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 设置viper
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 映射配置
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置默认值
	if config.AI.Timeout <= 0 {
		config.AI.Timeout = 30
	}
	if config.AI.BaseURL == "" {
		config.AI.BaseURL = "https://api.openai.com/v1/chat/completions"
	}

	return &config, nil
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	// 获取当前执行目录
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(dir, "config.yaml")
}
