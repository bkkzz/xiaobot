package jarvis

type NewBing struct {
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
