# OCS-AI答题

使用自己的AI API代替题库搜索，进行答题，支持 OpenAI 兼容的 API 接口。

[OCS官网](https://docs.ocsjs.com/)

## 注意

你需要首先获取AI的API才能使用

- [DeepSeek](https://platform.deepseek.com/) 
- [智谱 新人免费2000万Token](https://www.bigmodel.cn/invite?icode=XMLGgIZ1306F2SuhHqZPSn3uFJ1nZ0jLLgipQkYjpcA%3D)
- [硅基流动](https://cloud.siliconflow.cn/i/TNRegNF7)

项目处于初期阶段，可能会出现匹配失败的情况 如遇到此问题，请前往 issue 提交日志

## 快速开始

1. 复制配置文件：
   ```bash
   cp config.example.yaml config.yaml
   ```

2. 编辑 `config.yaml`，填写 AI 接口配置：
   - `ai.base_url`: AI 接口地址
   - `ai.api_key`: API 密钥
   - `ai.model`: 模型名称

3. 运行服务：
   ```bash
   go run .
   ```
   或通过 Release 下载编译好的exe文件并启动

## 对接 OCS

导入以下题库配置：

```
[
    {
        "name": "ocs-AI",
        "homepage": "http://127.0.0.1:8080/health",
        "url": "http://127.0.0.1:8080/query",
        "method": "get",
        "type": "GM_xmlhttpRequest",
        "contentType": "json",
        "data": {
            "token": "123456",
            "title": "${title}",
            "options": "${options}",
            "type": "${type}"
        },
        "handler": "return (res)=>res.code === 0 ? [res.data.answer, undefined] : [res.data.question,res.data.answer]"
    }
]
```

## 编译

```bash
go build -o ocs-ai.exe .
```

## 开源协议

MIT