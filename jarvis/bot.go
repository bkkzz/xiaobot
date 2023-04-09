package jarvis

type Jarvis interface {
    Ask(msg string) (string, error)
    AskStream()

    GetHistory() []string
    SetHistoryPrompt(newPrompt string)
}
