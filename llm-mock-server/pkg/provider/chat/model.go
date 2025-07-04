package chat

const (
	completionMockId = "chatcmpl-llm-mock"

	objectChatCompletion      = "chat.completion"
	objectChatCompletionChunk = "chat.completion.chunk"

	roleAssistant = "assistant"

	stopReason = "stop"

	contentTypeText     = "text"
	contentTypeImageUrl = "image_url"
)

var (
	completionMockCreated int64 = 10
	completionMockUsage         = usage{
		PromptTokens:     9,
		CompletionTokens: 1,
		TotalTokens:      10,
	}
)

type chatCompletionRequest struct {
	Model            string                 `json:"model" validate:"required"`
	Messages         []chatMessage          `json:"messages" validate:"required,min=1"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	FrequencyPenalty float64                `json:"frequency_penalty,omitempty"`
	N                int                    `json:"n,omitempty"`
	PresencePenalty  float64                `json:"presence_penalty,omitempty"`
	Seed             int                    `json:"seed,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	StreamOptions    *streamOptions         `json:"stream_options,omitempty"`
	Temperature      float64                `json:"temperature,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	Tools            []tool                 `json:"tools,omitempty"`
	ToolChoice       *toolChoice            `json:"tool_choice,omitempty"`
	User             string                 `json:"user,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
	ResponseFormat   map[string]interface{} `json:"response_format,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type tool struct {
	Type     string   `json:"type"`
	Function function `json:"function"`
}

type function struct {
	Description string                 `json:"description,omitempty"`
	Name        string                 `json:"name"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type toolChoice struct {
	Type     string   `json:"type"`
	Function function `json:"function"`
}

type chatCompletionResponse struct {
	Id                string                 `json:"id,omitempty"`
	Choices           []chatCompletionChoice `json:"choices"`
	Created           int64                  `json:"created,omitempty"`
	Model             string                 `json:"model,omitempty"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
	Object            string                 `json:"object,omitempty"`
	Usage             *usage                 `json:"usage"`
}

type chatCompletionChoice struct {
	Index        int                    `json:"index"`
	Message      *chatMessage           `json:"message,omitempty"`
	Delta        *chatMessage           `json:"delta,omitempty"`
	FinishReason *string                `json:"finish_reason"`
	Logprobs     map[string]interface{} `json:"logprobs"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type chatMessage struct {
	Name      string     `json:"name,omitempty"`
	Role      string     `json:"role,omitempty"`
	Content   any        `json:"content,omitempty"`
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
}

type messageContent struct {
	Type     string    `json:"type,omitempty"`
	Text     string    `json:"text"`
	ImageUrl *imageUrl `json:"image_url,omitempty"`
}

type imageUrl struct {
	Url    string `json:"url,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func (m *chatMessage) IsEmpty() bool {
	if m.IsStringContent() && m.Content != "" {
		return false
	}
	anyList, ok := m.Content.([]any)
	if ok && len(anyList) > 0 {
		return false
	}
	if len(m.ToolCalls) != 0 {
		nonEmpty := false
		for _, toolCall := range m.ToolCalls {
			if !toolCall.Function.IsEmpty() {
				nonEmpty = true
				break
			}
		}
		if nonEmpty {
			return false
		}
	}
	return true
}

func (m *chatMessage) IsStringContent() bool {
	_, ok := m.Content.(string)
	return ok
}

func (m *chatMessage) StringContent() string {
	content, ok := m.Content.(string)
	if ok {
		return content
	}
	contentList, ok := m.Content.([]any)
	if ok {
		var contentStr string
		for _, contentItem := range contentList {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}
			if contentMap["type"] == contentTypeText {
				if subStr, ok := contentMap[contentTypeText].(string); ok {
					contentStr += subStr + "\n"
				}
			}
		}
		return contentStr
	}
	return ""
}

func (m *chatMessage) ParseContent() []messageContent {
	var contentList []messageContent
	content, ok := m.Content.(string)
	if ok {
		contentList = append(contentList, messageContent{
			Type: contentTypeText,
			Text: content,
		})
		return contentList
	}
	anyList, ok := m.Content.([]any)
	if ok {
		for _, contentItem := range anyList {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}
			switch contentMap["type"] {
			case contentTypeText:
				if subStr, ok := contentMap[contentTypeText].(string); ok {
					contentList = append(contentList, messageContent{
						Type: contentTypeText,
						Text: subStr,
					})
				}
			case contentTypeImageUrl:
				if subObj, ok := contentMap[contentTypeImageUrl].(map[string]any); ok {
					contentList = append(contentList, messageContent{
						Type: contentTypeImageUrl,
						ImageUrl: &imageUrl{
							Url: subObj["url"].(string),
						},
					})
				}
			}
		}
		return contentList
	}
	return nil
}

type toolCall struct {
	Index    int          `json:"index"`
	Id       string       `json:"id"`
	Type     string       `json:"type"`
	Function functionCall `json:"function"`
}

type functionCall struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func (m *functionCall) IsEmpty() bool {
	return m.Name == "" && m.Arguments == ""
}
