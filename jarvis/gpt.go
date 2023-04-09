package jarvis

type GhatGpt struct {
}

func (g *GhatGpt) Ask(msg string) (string, error) {
    return "", nil
}

func (g *GhatGpt) AskStream() {

}

func (g *GhatGpt) GetHistory() []string {
    return []string{}
}

func (g *GhatGpt) SetHistoryPrompt(newPrompt string) {

}
