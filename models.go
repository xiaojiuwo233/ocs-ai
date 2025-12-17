package main

// QueryRequest 查询请求
type QueryRequest struct {
	Token   string `json:"token" form:"token"`
	Title   string `json:"title" form:"title"`
	Options string `json:"options" form:"options"`
	Type    string `json:"type" form:"type"`
	More    *bool  `json:"more,omitempty" form:"more"`
}

// QueryResponse 查询响应
type QueryResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Times   int               `json:"times,omitempty"`
	Data    QueryResponseData `json:"data"`
}

// QueryResponseData 查询结果详情
type QueryResponseData struct {
	Question string        `json:"question,omitempty"`
	Answer   string        `json:"answer,omitempty"`
	Results  []QueryResult `json:"results,omitempty"`
}

// QueryResult 查询题目和答案
type QueryResult struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// OpenAIResponse AI接口响应
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type,omitempty"`
		Code    any    `json:"code,omitempty"`
	} `json:"error,omitempty"`
}
