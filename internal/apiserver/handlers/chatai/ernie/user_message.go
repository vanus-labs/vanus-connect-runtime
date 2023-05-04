package ernie

type userMessage struct {
	messages   []ChatCompletionMessage
	tokens     []int
	totalToken int
}

func (m *userMessage) cal(newToken, maxTokens int) {
	currToken := m.totalToken + newToken
	if currToken < maxTokens {
		return
	}
	var index, token int
	for index < len(m.tokens) {
		// question token
		token += m.tokens[index]
		// answer token
		token += m.tokens[index+1]
		index += 2
		if currToken-token < maxTokens {
			break
		}
	}
	m.totalToken -= token
	m.messages = m.messages[index:]
	m.tokens = m.tokens[index:]
}
