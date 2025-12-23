// Package services 提供业务逻辑层的服务实现
// ai_service.go 实现了AI相关的业务逻辑，包括图片分析和自然语言查询转换
// 使用智谱AI GLM-4 Vision模型，API格式兼容OpenAI，支持国内直接访问
package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"image-manager/internal/config"
)

// AIService AI服务结构体
// 提供AI相关的功能，包括图片分析和自然语言查询转换
type AIService struct {
	cfg config.Config  // 应用配置信息，包含AI API密钥、URL等
}

// NewAIService 创建AI服务实例
// 参数:
//   - cfg: 应用配置
// 返回: AIService指针
func NewAIService(cfg config.Config) *AIService {
	return &AIService{
		cfg: cfg,
	}
}

// AnalyzeImageRequest OpenAI API请求结构
type AnalyzeImageRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

// Message API消息结构
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// TextContent 文本内容
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ImageURL 图片URL结构
type ImageURL struct {
	URL string `json:"url"`
}

// ImageContent 图片内容
type ImageContent struct {
	Type     string   `json:"type"`
	ImageURL ImageURL `json:"image_url"`
}

// AnalyzeImageResponse AI API响应结构（兼容OpenAI格式，支持智谱AI）
type AnalyzeImageResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`  // AI返回的内容
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`  // 错误消息
	} `json:"error,omitempty"`
}

// AnalyzeImage 分析图片并生成标签
// 调用智谱AI GLM-4 Vision模型分析图片内容，返回标签列表（如风景、人物、动物等）
// 参数:
//   - imageData: 图片的二进制数据
//   - mimeType: 图片的MIME类型（如image/jpeg）
//   - existingTags: 标签库中已有的标签列表，AI会优先从中选择
// 返回: 标签名称列表和错误信息
func (s *AIService) AnalyzeImage(imageData []byte, mimeType string, existingTags []string) ([]string, error) {
	// 如果AI功能未启用或API密钥为空，返回空标签列表
	if !s.cfg.AIEnabled {
		log.Printf("AI功能未启用，跳过图片分析")
		return []string{}, nil
	}
	if s.cfg.AIApiKey == "" {
		log.Printf("AI API密钥为空，跳过图片分析")
		return []string{}, nil
	}

	// 将图片数据编码为base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	
	// 构建图片数据URL（data URI格式）
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	// 构建提示词，要求AI分析图片并返回简短的关键字标签
	// GLM-4v可能会返回<|observation|>标记，我们需要更直接明确的指令
	prompt := `请分析这张图片，直接输出5-15个简短的关键字标签，用中文逗号分隔。

输出要求：
1. 只输出标签，不要任何标记、前缀、后缀或说明
2. 标签格式：标签1,标签2,标签3（用中文逗号分隔）
3. 每个标签1-4个字，简短的关键字
4. 优先从已有标签库中选择，如果没有合适的可以生成新标签

已有标签库：`
	
	// 如果有已有标签，添加到提示词中
	if len(existingTags) > 0 {
		prompt += "\n" + strings.Join(existingTags, "、")
		prompt += "\n\n请优先从上述标签库中选择，如果图片内容匹配不上，再生成新的简短关键字标签。"
	} else {
		prompt += "\n（暂无已有标签，请生成新的简短关键字标签）"
	}
	
	prompt += `

**重要**：
1. 必须使用逗号分隔格式（格式：标签1,标签2,标签3）
2. 不要使用"标签1: xxx"这种格式
3. 不要直接复制示例，要根据实际图片内容生成标签
4. 只输出标签，不要有任何其他文字

请根据图片内容，直接用逗号分隔输出标签：`

	// 构建请求内容，包含文本和图片
	// 使用map结构以确保JSON序列化正确
	content := []interface{}{
		map[string]interface{}{
			"type": "text",
			"text": prompt,
		},
		map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]interface{}{
				"url": dataURL,
			},
		},
	}

	// 构建API请求
	// 注意：智谱AI GLM-4v在某些情况下可能会忽略system message，所以我们把要求都放在user message中
	// 如果仍然返回<|observation|>，我们会在后续清理中移除它
	reqBody := AnalyzeImageRequest{
		Model: s.cfg.AIModel,
		Messages: []Message{
			{
				Role:    "user",
				Content: content,
			},
		},
		MaxTokens: 500,  // 增加token限制，确保模型有足够空间输出完整标签
	}

	// 将请求体序列化为JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 打印请求信息（用于调试，但隐藏图片数据以减少日志大小）
	log.Printf("AI API请求信息 - 模型: %s, URL: %s, Content数组长度: %d, MaxTokens: %d", 
		s.cfg.AIModel, s.cfg.AIApiURL, len(content), reqBody.MaxTokens)
	// 打印请求体（但不包含完整的base64图片数据，只显示结构）
	reqPreview := fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":[...%d个元素...]}],"max_tokens":%d}`, 
		s.cfg.AIModel, len(content), reqBody.MaxTokens)
	log.Printf("AI API请求体结构预览: %s", reqPreview)

	// 创建HTTP请求
	req, err := http.NewRequest("POST", s.cfg.AIApiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.cfg.AIApiKey))
	log.Printf("AI API请求头已设置，Content-Type: application/json")

	// 创建HTTP客户端并发送请求（设置30秒超时）
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求AI API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 打印完整的原始响应内容（包括所有字段）
	log.Printf("AI API完整响应内容（状态码: %d）: %s", resp.StatusCode, string(respBody))

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		log.Printf("AI API返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
		return []string{}, nil  // 如果API调用失败，返回空标签列表，不影响上传流程
	}
	
	log.Printf("AI API调用成功，开始解析响应")

	// 解析响应
	var aiResp AnalyzeImageResponse
	if err := json.Unmarshal(respBody, &aiResp); err != nil {
		log.Printf("解析AI API响应失败: %v, 原始响应: %s", err, string(respBody))
		return []string{}, nil
	}

	// 打印解析后的响应结构（包括error字段等）
	if aiResp.Error != nil {
		log.Printf("AI API返回错误字段: %+v, 完整响应: %+v", aiResp.Error, aiResp)
		return []string{}, nil
	}

	// 打印解析后的完整响应结构（用于调试）
	log.Printf("AI API解析后的完整响应结构: %+v", aiResp)

	// 提取响应内容
	if len(aiResp.Choices) == 0 {
		log.Printf("AI API响应中没有Choices字段，返回空标签列表，完整响应: %+v", aiResp)
		return []string{}, nil
	}

	contentStr := aiResp.Choices[0].Message.Content
	log.Printf("AI返回的原始内容: %s", contentStr)
	
	// 如果AI只返回了<|observation|>标记，说明AI没有理解要求
	// 智谱AI GLM-4v在处理图片时可能会自动添加<|observation|>标记
	// 我们需要彻底移除这些标记，只保留标签内容
	
	// 先移除已知的特殊标记（包括可能的大小写变体）
	specialMarkers := []string{
		"<|observation|>", "<|think|>", "<|system|>", "<|user|>", "<|assistant|>",
		"<|endoftext|>", "<|end_of_text|>", "<|startoftext|>", "<|start_of_text|>",
		"<|OBSERVATION|>", "<|THINK|>", "<|SYSTEM|>", "<|USER|>", "<|ASSISTANT|>",
	}
	for _, marker := range specialMarkers {
		contentStr = strings.ReplaceAll(contentStr, marker, "")
	}
	// 使用正则表达式移除所有<|...|>格式的标记（作为备用方案，处理未知的特殊标记）
	// 这个正则会匹配所有<|...|>格式的内容，包括可能的空白字符
	markerRegex := regexp.MustCompile(`<\|[^|]*\|>`)
	contentStr = markerRegex.ReplaceAllString(contentStr, "")
	
	// 如果清理后内容为空或只有空白，说明AI只返回了标记，记录警告
	trimmed := strings.TrimSpace(contentStr)
	if trimmed == "" || trimmed == "<|observation|>" {
		log.Printf("警告：AI只返回了<|observation|>标记，没有实际标签内容。这可能是GLM-4v的默认行为，模型认为观察已完成。原始内容: %s", aiResp.Choices[0].Message.Content)
		// GLM-4v在处理图片时可能会先输出<|observation|>然后停止
		// 这种情况下返回空标签列表，避免错误处理
		return []string{}, nil
	}
	
	log.Printf("清理特殊标记后的内容: %s", contentStr)
	
	// 解析标签：支持多种格式
	// 1. 逗号分隔格式：标签1,标签2,标签3
	// 2. "标签X: xxx"格式：标签1: xxx 标签2: xxx
	// 3. 其他分隔符（分号、换行等）
	
	tags := []string{}
	
	// 首先尝试解析"标签X: xxx"格式（如：标签1: 异虫 标签2: 星际争霸）
	// 使用正则表达式匹配"标签数字: 内容"的模式
	// Go的regexp不支持前瞻断言，所以使用FindAllStringSubmatch来匹配所有"标签数字: 内容"的模式
	tagPatternRegex := regexp.MustCompile(`标签\d+\s*[:：]\s*([^标签]+)`)
	matches := tagPatternRegex.FindAllStringSubmatch(contentStr, -1)
	if len(matches) > 0 {
		log.Printf("检测到'标签X: xxx'格式，开始解析，匹配到%d个标签", len(matches))
		for _, match := range matches {
			if len(match) > 1 {
				tag := strings.TrimSpace(match[1])
				// 清理标签内容，移除多余的空白字符和标点
				// 由于匹配可能包含下一个"标签"之前的所有内容，需要进一步清理
				tag = strings.Trim(tag, "，,。.！!？?；;：: \n\r\t")
				// 如果标签内容中包含"标签"字样（可能是下一个标签的前缀），移除它
				if idx := strings.Index(tag, "标签"); idx > 0 {
					tag = tag[:idx]
					tag = strings.TrimSpace(tag)
				}
				if tag != "" {
					tags = append(tags, tag)
				}
			}
		}
	}
	
	// 如果没有匹配到"标签X: xxx"格式，则按多种分隔符解析
	if len(tags) == 0 {
		// 支持多种分隔符：中文逗号、英文逗号、顿号、分号（中英文）
		// 先统一替换为英文逗号，便于统一处理
		normalized := contentStr
		normalized = strings.ReplaceAll(normalized, "，", ",")  // 中文逗号
		normalized = strings.ReplaceAll(normalized, "、", ",")  // 顿号
		normalized = strings.ReplaceAll(normalized, "；", ",")  // 中文分号
		normalized = strings.ReplaceAll(normalized, ";", ",")   // 英文分号
		log.Printf("标准化后的内容: %s", normalized)
		
		// 按逗号分割
		for _, part := range strings.Split(normalized, ",") {
			tag := strings.TrimSpace(part)
			// 移除可能的标点符号和多余字符（包括"标签X:"前缀）
			tag = strings.Trim(tag, "，,。.！!？?；;：: \n\r\t")
			// 移除"标签X:"前缀（如果还有残留）
			tagPrefixRegex := regexp.MustCompile(`^标签\d+\s*[:：]\s*`)
			tag = tagPrefixRegex.ReplaceAllString(tag, "")
			tag = strings.TrimSpace(tag)
			// 再次清理可能残留的特殊标记
			for _, marker := range specialMarkers {
				tag = strings.ReplaceAll(tag, marker, "")
			}
			tag = markerRegex.ReplaceAllString(tag, "")
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	log.Printf("解析后的标签列表: %v (共%d个标签)", tags, len(tags))
	return tags, nil
}

// ConvertQueryToFilters 将自然语言查询转换为图片搜索过滤器
// 使用智谱AI GLM-4模型将用户的自然语言描述转换为结构化的搜索条件
// 参数:
//   - query: 自然语言查询（如"找一些风景照片"、"显示上个月拍的猫的照片"）
//   - existingTags: 标签库中已有的标签列表，AI会优先从中选择标签
// 返回: 过滤器映射（包含keyword、tags、start_date等）和错误信息
func (s *AIService) ConvertQueryToFilters(query string, existingTags []string) (map[string]string, error) {
	// 如果AI功能未启用或API密钥为空，返回空过滤器
	if !s.cfg.AIEnabled || s.cfg.AIApiKey == "" {
		return map[string]string{"keyword": query}, nil  // 降级为关键词搜索
	}

	// 构建提示词，要求AI将自然语言转换为JSON格式的过滤器
	// 注意：tags字段中的多个标签之间是OR关系（满足其中一个即可），以放宽检索条件
	// **重要**：每次调用都是完全独立的查询，不要参考任何历史对话或之前的查询
	prompt := fmt.Sprintf(`**你必须只返回一个有效的JSON对象，不要有任何说明文字、解释或示例。直接输出JSON，不要任何其他内容。**

用户查询：%s

转换规则：
1. **优先生成标签**：尽量从用户查询中提取标签用于检索，除非用户明确说"只搜索文件名"、"只用文件名搜索"等表示只检索文件名的意图，否则应该生成tags字段。即使查询中包含文件名关键词，也应该同时生成相关标签。
2. tags字段中的多个标签之间是"或(OR)"关系，即图片只要有其中任意一个标签即可匹配（放宽检索条件）
3. 标签必须从已有的标签库中选择，如果查询中的标签不在标签库中，请忽略或使用相近的标签。如果标签库中有多个相关标签，可以生成多个标签（用逗号分隔）。
4. **严格要求**：只生成用户明确提到的条件！如果用户没有明确提到日期、时间、文件大小、分辨率等条件，请不要生成这些字段。不要根据查询内容自行推断或添加额外的筛选条件。
5. **独立查询**：不要从历史对话中推断任何信息，只基于当前查询内容进行转换。

已有标签库：`, query)
	
	// 如果有已有标签，添加到提示词中
	if len(existingTags) > 0 {
		prompt += "\n" + strings.Join(existingTags, "、")
		prompt += "\n\n请优先从上述标签库中选择标签。"
	} else {
		prompt += "\n（暂无已有标签）"
	}
	
	prompt += `

请返回一个JSON对象，**只能包含以下字段**（只包含用户明确提到的条件，不要添加任何其他字段如background、feature等）：
- keyword: 关键词（字符串，用于搜索文件名。只有用户明确提到文件名、文件关键词时才生成。注意：即使生成了keyword，也应该尽量同时生成tags）
- tags: 标签（字符串，多个标签用逗号分隔，如"风景,山"。这些标签会被用于OR查询，且必须是标签库中存在的标签。**优先生成标签**：除非用户明确说"只搜索文件名"，否则应该尽量从查询中提取标签。可以从查询的主题、内容、类型等方面提取相关标签，如果标签库中有多个相关标签可以都生成）
- start_date: 开始日期（字符串，格式：YYYY-MM-DD，例如"2024-06-15"。只有用户明确提到创建时间、上传时间范围时才生成，必须根据用户查询中的实际日期生成，不要使用固定的默认日期）
- end_date: 结束日期（字符串，格式：YYYY-MM-DD，例如"2024-12-31"。只有用户明确提到创建时间、上传时间范围时才生成，必须根据用户查询中的实际日期生成，不要使用固定的默认日期）
- taken_start: 拍摄开始时间（字符串，格式：YYYY-MM-DD HH:MM，例如"2024-06-15 08:00"。只有用户明确提到拍摄时间、拍照时间范围时才生成，必须根据用户查询中的实际时间生成，不要使用固定的默认时间）
- taken_end: 拍摄结束时间（字符串，格式：YYYY-MM-DD HH:MM，例如"2024-12-31 23:59"。只有用户明确提到拍摄时间、拍照时间范围时才生成，必须根据用户查询中的实际时间生成，不要使用固定的默认时间）
- width_min: 最小宽度（整数，像素。只有用户明确提到宽度、分辨率、尺寸时才生成）
- width_max: 最大宽度（整数，像素。只有用户明确提到宽度、分辨率、尺寸时才生成）
- height_min: 最小高度（整数，像素。只有用户明确提到高度、分辨率、尺寸时才生成）
- height_max: 最大高度（整数，像素。只有用户明确提到高度、分辨率、尺寸时才生成）
- size_min: 最小文件大小（数字，单位：MB，可以是小数，如1.5表示1.5MB。只有用户明确提到文件大小、文件体积时才生成）
- size_max: 最大文件大小（数字，单位：MB，可以是小数，如2.5表示2.5MB。只有用户明确提到文件大小、文件体积时才生成）

**输出格式要求（必须严格遵守）**：
1. **只输出JSON对象，不要有任何其他文字**（不要说明、不要解释、不要示例）
2. 只能返回上述字段，绝对不要添加任何其他字段（如background、feature、description等）
3. **关键**：对于文件大小，日期，宽度，高度等字段，除非用户明确提到这个字段相关的词，否则不要自行推断或添加条件！如果用户没有提到日期时间，绝对不要生成start_date、end_date、taken_start、taken_end字段！
4. **重要**：日期字段中的示例（如"2024-06-15"）只是格式说明，不要使用这些示例值。必须根据用户查询中提到的实际日期来生成，如果用户没有提到日期，就不要生成这些字段！
5. 直接输出JSON，格式如：{"tags": "风景"} 或 {"keyword": "test", "tags": "test"}

**重要**：你的响应必须是一个有效的JSON对象，从第一个{开始，到最后一个}结束，中间不要有任何其他文字。`

	// 构建API请求
	// 使用system message明确要求只返回JSON
	systemPrompt := "你是一个JSON转换工具。你只能返回有效的JSON对象，不要有任何说明文字、解释或示例。直接输出JSON，从{开始，到}结束。"
	reqBody := AnalyzeImageRequest{
		Model: s.cfg.AIModel,
		Messages: []Message{
			{
				Role: "system",
				Content: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": systemPrompt,
					},
				},
			},
			{
				Role: "user",
				Content: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
		MaxTokens: 500,
	}

	// 序列化请求
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", s.cfg.AIApiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.cfg.AIApiKey))

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求AI API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 打印完整的原始响应内容（包括所有字段）
	log.Printf("AI API完整响应内容（状态码: %d）: %s", resp.StatusCode, string(respBody))

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		log.Printf("AI API返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
		return map[string]string{"keyword": query}, nil  // 降级为关键词搜索
	}

	// 解析响应
	var aiResp AnalyzeImageResponse
	if err := json.Unmarshal(respBody, &aiResp); err != nil {
		log.Printf("解析AI API响应失败: %v, 原始响应: %s", err, string(respBody))
		return map[string]string{"keyword": query}, nil
	}

	// 打印解析后的完整响应结构（用于调试）
	log.Printf("AI API解析后的完整响应结构: %+v", aiResp)

	// 检查错误
	if aiResp.Error != nil {
		log.Printf("AI API返回错误字段: %+v, 完整响应: %+v", aiResp.Error, aiResp)
		return map[string]string{"keyword": query}, nil
	}

	// 提取响应内容
	if len(aiResp.Choices) == 0 {
		log.Printf("AI API响应中没有Choices字段，完整响应: %+v", aiResp)
		return map[string]string{"keyword": query}, nil
	}

	contentStr := aiResp.Choices[0].Message.Content
	
	// 尝试从响应中提取JSON（可能包含markdown代码块或说明文字）
	jsonStr := ""
	
	// 方法1：尝试提取markdown代码块中的JSON
	if strings.Contains(contentStr, "```json") {
		start := strings.Index(contentStr, "```json") + 7
		end := strings.Index(contentStr[start:], "```")
		if end > 0 {
			jsonStr = strings.TrimSpace(contentStr[start : start+end])
		}
	} else if strings.Contains(contentStr, "```") {
		start := strings.Index(contentStr, "```") + 3
		end := strings.Index(contentStr[start:], "```")
		if end > 0 {
			jsonStr = strings.TrimSpace(contentStr[start : start+end])
		}
	}
	
	// 方法2：如果没有找到代码块，尝试直接查找第一个JSON对象（从{开始到}结束）
	if jsonStr == "" {
		startIdx := strings.Index(contentStr, "{")
		if startIdx >= 0 {
			// 从第一个{开始，查找匹配的}
			depth := 0
			endIdx := startIdx
			for i := startIdx; i < len(contentStr); i++ {
				if contentStr[i] == '{' {
					depth++
				} else if contentStr[i] == '}' {
					depth--
					if depth == 0 {
						endIdx = i + 1
						break
					}
				}
			}
			if endIdx > startIdx {
				jsonStr = strings.TrimSpace(contentStr[startIdx:endIdx])
			}
		}
	}
	
	// 方法3：使用正则表达式匹配JSON对象（作为最后的手段）
	if jsonStr == "" {
		// 匹配 {...} 格式的JSON对象
		jsonRegex := regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)
		matches := jsonRegex.FindString(contentStr)
		if matches != "" {
			jsonStr = strings.TrimSpace(matches)
		}
	}
	
	// 如果还是没找到JSON，使用整个内容（可能是纯JSON）
	if jsonStr == "" {
		jsonStr = strings.TrimSpace(contentStr)
	}

	// 解析JSON为过滤器映射
	// 先使用interface{}类型解析，因为AI可能返回数字类型的值（如width_min、height_min等）
	var rawFilters map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawFilters); err != nil {
		// 检查是否完全没有JSON（AI可能返回了纯说明文字）
		if !strings.Contains(contentStr, "{") && !strings.Contains(contentStr, "}") {
			log.Printf("AI返回了纯文字内容，未包含JSON。原始内容: %s", contentStr)
		} else {
			log.Printf("解析AI返回的JSON失败: %v, 提取的JSON字符串: %s, 原始内容: %s", err, jsonStr, contentStr)
		}
		// 如果JSON解析失败，降级为关键词搜索
		return map[string]string{"keyword": query}, nil
	}

	// 定义允许的过滤器字段列表（只允许这些字段）
	allowedFields := map[string]bool{
		"keyword":     true, // 关键词
		"tags":        true, // 标签
		"start_date":  true, // 创建开始日期
		"end_date":    true, // 创建结束日期
		"taken_start": true, // 拍摄开始时间
		"taken_end":   true, // 拍摄结束时间
		"width_min":   true, // 最小宽度
		"width_max":   true, // 最大宽度
		"height_min":  true, // 最小高度
		"height_max":  true, // 最大高度
		"size_min":    true, // 最小文件大小
		"size_max":    true, // 最大文件大小
	}

	// 将interface{}类型的值转换为string类型，并过滤掉不在允许列表中的字段
	filters := make(map[string]string)
	for k, v := range rawFilters {
		// 跳过不在允许列表中的字段（如background、feature等）
		if !allowedFields[k] {
			continue
		}
		
		// 跳过空值
		if v == nil {
			continue
		}
		
		var strValue string
		switch val := v.(type) {
		case string:
			strValue = val
		case float64:
			// JSON数字会被解析为float64，转换为整数字符串
			strValue = fmt.Sprintf("%.0f", val)
		case int:
			strValue = fmt.Sprintf("%d", val)
		case int64:
			strValue = fmt.Sprintf("%d", val)
		default:
			// 其他类型也尝试转换为字符串
			strValue = fmt.Sprintf("%v", val)
		}
		
		// 只有非空字符串才添加到filters中
		if strValue != "" {
			filters[k] = strValue
		}
	}

	return filters, nil
}

