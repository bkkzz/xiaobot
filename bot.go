package xiaogpt

type Bot interface {
    Ask()
    AskStream()

    GetHistory() []string
    SetHistoryPrompt(newPrompt string)
}

type GhatGptBot struct {
}

func (g *GhatGptBot) Ask() {

}

func (g *GhatGptBot) AskStream() {

}

func (g *GhatGptBot) GetHistory() []string {
    return []string{}
}

func (g *GhatGptBot) SetHistoryPrompt(newPrompt string) {

}
