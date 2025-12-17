package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("启动AI答题服务...")

	// 获取配置文件路径
	configPath := GetConfigPath()
	log.Printf("使用配置文件: %s", configPath)

	// 确保配置文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("创建配置目录失败: %v", err)
		}

		// 写入默认配置
		defaultConfig := `# ocs-ai配置

# 本地服务端配置
server:
  port: 8080  # 本地服务端口
  host: "127.0.0.1"  # 本地服务主机

# AI回答配置
ai:
  base_url: "https://api.openai.com/v1/chat/completions"  # AI接口地址
  api_key: ""  # AI接口密钥
  model: "gpt-4o-mini"  # 使用的模型名称
  system_prompt: "你是一名专业的答题助手，请根据题目和选项给出最可能的正确答案，并简要解释理由。"
  prompt_template: |
    题目：{{title}}
    选项：{{options}}
    类型：{{type}}

    请根据题目信息给出最可能的正确答案，并在必要时提供简要推理。
  temperature: 0.2  # 随机性
  top_p: 1.0  # nucleus sampling
  max_tokens: 512  # 最大输出token
  timeout: 30  # 请求超时时间（秒）
`
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			log.Fatalf("创建默认配置文件失败: %v", err)
		}
		log.Printf("已创建默认配置文件: %s", configPath)
	}

	// 加载配置
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建服务器
	server := NewServer(config)

	// 启动服务器
	if err := server.Start(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	// 打印使用说明
	printUsage(config)

	// 等待关闭信号
	server.WaitForShutdown()
}

// printUsage 打印使用说明
func printUsage(config *Config) {
	addr := fmt.Sprintf("http://%s:%d", config.Server.Host, config.Server.Port)
	fmt.Println("\n=========== OCS-AI ===========")
	fmt.Println("服务已成功启动！")
	fmt.Printf("API地址: %s\n", addr)
	fmt.Println("\n使用说明:")
	fmt.Println("1. 替换脚本中的API地址:")
	fmt.Printf("   将题库地址修改为 '%s/query'\n", addr)
	fmt.Printf("2. 健康检查: %s/health\n", addr)
	fmt.Println("3. 按Ctrl+C停止服务")
	fmt.Println("=====================================")
}
