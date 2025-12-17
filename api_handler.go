package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultAISystemPrompt   = "你是一名专业的答题助手，请根据题目内容和选项快速给出最可能的正确答案，并提供必要的推理。"
	defaultAIPromptTemplate = "题目：{{title}}\n选项：{{options}}\n类型：{{type}}\n\n请给出最可能的正确答案，并在必要时给出简要解释。"
)

// APIHandler 处理API请求
type APIHandler struct {
	config   *Config
	aiClient *http.Client
}

// NewAPIHandler 创建一个新的APIHandler
func NewAPIHandler(config *Config) *APIHandler {
	timeout := time.Duration(config.AI.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &APIHandler{
		config: config,
		aiClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// HandleQuery 处理查询请求
func (h *APIHandler) HandleQuery(c *gin.Context) {
	var req QueryRequest

	// 绑定请求参数，同时支持GET和POST
	if c.Request.Method == "GET" {
		req.Token = c.Query("token")
		req.Title = c.Query("title")
		req.Options = c.Query("options")
		req.Type = c.Query("type")
		if moreStr := c.Query("more"); moreStr != "" {
			if parsed, err := strconv.ParseBool(moreStr); err == nil {
				req.More = &parsed
			}
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("解析请求失败: %v", err)
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    0,
				Message: fmt.Sprintf("解析请求失败: %v", err),
			})
			return
		}
	}

	// 验证必要参数
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    0,
			Message: "缺少必要参数: title",
		})
		return
	}

	var moreFlag bool
	if req.More != nil {
		moreFlag = *req.More
	}
	log.Printf("收到查询请求: 题目=%s, 选项=%s, 类型=%s, More=%t", req.Title, req.Options, req.Type, moreFlag)

	h.handleAIQuery(c, &req)
}


func (h *APIHandler) handleAIQuery(c *gin.Context, req *QueryRequest) {
	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    0,
			Message: "缺少必要参数: title",
		})
		return
	}

	baseURL := strings.TrimSpace(h.config.AI.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/chat/completions"
	}

	model := strings.TrimSpace(h.config.AI.Model)
	if model == "" {
		model = "gpt-4o-mini"
	}

	systemPrompt := strings.TrimSpace(h.config.AI.SystemPrompt)
	if systemPrompt == "" {
		systemPrompt = defaultAISystemPrompt
	}

	promptTemplate := strings.TrimSpace(h.config.AI.PromptTemplate)
	if promptTemplate == "" {
		promptTemplate = defaultAIPromptTemplate
	}

	prompt := h.renderPrompt(promptTemplate, req)

	messages := make([]map[string]string, 0, 2)
	if systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})

	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}

	if h.config.AI.Temperature > 0 {
		payload["temperature"] = h.config.AI.Temperature
	}
	if h.config.AI.TopP > 0 {
		payload["top_p"] = h.config.AI.TopP
	}
	if h.config.AI.MaxTokens > 0 {
		payload["max_tokens"] = h.config.AI.MaxTokens
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("序列化AI请求失败: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    0,
			Message: fmt.Sprintf("构建AI请求失败: %v", err),
		})
		return
	}

	reqBody := bytes.NewBuffer(body)
	httpReq, err := http.NewRequest("POST", baseURL, reqBody)
	if err != nil {
		log.Printf("创建AI请求失败: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    0,
			Message: fmt.Sprintf("创建AI请求失败: %v", err),
		})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if strings.TrimSpace(h.config.AI.APIKey) != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.TrimSpace(h.config.AI.APIKey)))
	}

	log.Printf("转发请求到AI模型: %s, 模型=%s", baseURL, model)

	resp, err := h.aiClient.Do(httpReq)
	if err != nil {
		log.Printf("请求AI接口失败: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    0,
			Message: fmt.Sprintf("请求AI接口失败: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取AI响应失败: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    0,
			Message: fmt.Sprintf("读取AI响应失败: %v", err),
		})
		return
	}

	if resp.StatusCode >= 400 {
		log.Printf("AI接口返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		c.JSON(resp.StatusCode, ErrorResponse{
			Code:    0,
			Message: fmt.Sprintf("AI接口返回错误: %s", strings.TrimSpace(string(respBody))),
		})
		return
	}

	var aiResp OpenAIResponse
	if err := json.Unmarshal(respBody, &aiResp); err != nil {
		log.Printf("解析AI响应失败: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    0,
			Message: fmt.Sprintf("解析AI响应失败: %v", err),
		})
		return
	}

	if aiResp.Error != nil {
		log.Printf("AI接口返回错误: %s", aiResp.Error.Message)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    0,
			Message: fmt.Sprintf("AI接口错误: %s", aiResp.Error.Message),
		})
		return
	}

	if len(aiResp.Choices) == 0 || strings.TrimSpace(aiResp.Choices[0].Message.Content) == "" {
		log.Println("AI接口未返回有效答案")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    0,
			Message: "AI接口未返回有效答案",
		})
		return
	}

	rawAnswer := strings.TrimSpace(aiResp.Choices[0].Message.Content)
	formattedAnswer := formatAIAnswer(rawAnswer)
	log.Printf("AI回答成功: 题目=%s, 答案=%s", req.Title, formattedAnswer)

	var queryResp QueryResponse
	queryResp.Code = 1
	queryResp.Message = "AI回答成功"
	queryResp.Times = -1

	if req.More != nil && *req.More {
		queryResp.Data.Results = []QueryResult{{
			Question: req.Title,
			Answer:   formattedAnswer,
		}}
	} else {
		queryResp.Data.Question = req.Title
		queryResp.Data.Answer = formattedAnswer
	}

	c.JSON(http.StatusOK, queryResp)
}

func (h *APIHandler) renderPrompt(template string, req *QueryRequest) string {
	prompt := template

	replacements := map[string]string{
		"{{title}}":   strings.TrimSpace(req.Title),
		"{{options}}": strings.TrimSpace(req.Options),
		"{{type}}":    strings.TrimSpace(req.Type),
	}

	for placeholder, value := range replacements {
		if value == "" {
			value = "无"
		}
		prompt = strings.ReplaceAll(prompt, placeholder, value)
	}

	return prompt
}

func formatAIAnswer(raw string) string {
	normalized := strings.TrimSpace(strings.ReplaceAll(raw, "\r\n", "\n"))
	if normalized == "" {
		return ""
	}

	var analysis string
	keywords := []string{"解析：", "解析:", "理由：", "理由:", "原因：", "原因:", "说明：", "说明:", "推理：", "推理:"}
	for _, kw := range keywords {
		if idx := strings.Index(normalized, kw); idx != -1 {
			analysis = strings.TrimSpace(normalized[idx+len(kw):])
			normalized = strings.TrimSpace(normalized[:idx])
			break
		}
	}

	answerSection := normalized
	for _, prefix := range []string{"正确答案：", "正确答案:", "答案：", "答案:"} {
		if strings.HasPrefix(answerSection, prefix) {
			answerSection = strings.TrimSpace(strings.TrimPrefix(answerSection, prefix))
			break
		}
	}

	if answerSection == "" {
		answerSection = normalized
	}

	// 忽略解析内容变量，确保编译通过
	_ = analysis
	return answerSection
}

// HandleInfo 处理信息请求
func (h *APIHandler) HandleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "AI模式",
		"data": gin.H{
			"model":    strings.TrimSpace(h.config.AI.Model),
			"endpoint": strings.TrimSpace(h.config.AI.BaseURL),
		},
	})
}

// HandleHealth 处理健康检查请求
func (h *APIHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"mode":   "ai",
		"ai": gin.H{
			"model":    strings.TrimSpace(h.config.AI.Model),
			"endpoint": strings.TrimSpace(h.config.AI.BaseURL),
		},
	})
}
