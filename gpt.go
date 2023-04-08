package xiaobot

type GhatGptBot struct {
}

func (g *GhatGptBot) Ask(msg string) (string, error) {
    return "", nil
}

func (g *GhatGptBot) AskStream() {

}

func (g *GhatGptBot) GetHistory() []string {
    return []string{}
}

func (g *GhatGptBot) SetHistoryPrompt(newPrompt string) {

}
