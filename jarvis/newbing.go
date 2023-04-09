package jarvis

type NewBing struct {
    Cookie string
    Proxy  string
}

func (g *NewBing) Ask(msg string) (string, error) {
    return "", nil
}

func (g *NewBing) AskStream() {

}

func (g *NewBing) GetHistory() []string {
    return []string{}
}

func (g *NewBing) SetHistoryPrompt(newPrompt string) {

}
