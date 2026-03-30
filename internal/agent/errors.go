package agent

import "net/http"

// AgentErrorClassifier классифицирует ошибки HTTP-запросов агента для определения необходимости ретрая.
type AgentErrorClassifier struct{}

// NewAgentErrorClassifier создаёт новый AgentErrorClassifier.
func NewAgentErrorClassifier() *AgentErrorClassifier {
	return &AgentErrorClassifier{}
}

func (c *AgentErrorClassifier) isRetriable(err error, resp *http.Response) bool {
	if err != nil {
		return true
	}
	if resp != nil && resp.StatusCode >= 500 {
		return true
	}

	return false
}
