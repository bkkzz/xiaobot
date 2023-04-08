package xiaobot

type Bot interface {
    Ask(msg string) (string, error)
    AskStream()

    GetHistory() []string
    SetHistoryPrompt(newPrompt string)
}
