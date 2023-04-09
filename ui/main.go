package main

import (
    "fmt"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "github.com/longbai/xiaobot"
    "strings"
)

func main() {
    a := app.New()
    w := a.NewWindow("XiaoBot")

    config := xiaobot.Config{}

    hardwareEntry := widget.NewEntry()
    accountEntry := widget.NewEntry()
    passwordEntry := widget.NewPasswordEntry()
    openAIKeyEntry := widget.NewPasswordEntry()
    openAIBackendEntry := widget.NewEntry()
    proxyEntry := widget.NewEntry()
    miDIDEntry := widget.NewEntry()
    keywordsEntry := widget.NewEntry()
    changePromptKeywordsEntry := widget.NewEntry()
    promptEntry := widget.NewEntry()
    muteXiaoAICheck := widget.NewCheck("", nil)
    botEntry := widget.NewSelectEntry([]string{"gpt", "newbing"})
    apiBaseEntry := widget.NewEntry()
    useCommandCheck := widget.NewCheck("", nil)
    verboseCheck := widget.NewCheck("", nil)
    startConversationEntry := widget.NewEntry()
    endConversationEntry := widget.NewEntry()
    streamCheck := widget.NewCheck("", nil)
    enableEdgeTTSCheck := widget.NewCheck("", nil)
    edgeTTSVoiceEntry := widget.NewEntry()
    tokenPathEntry := widget.NewEntry()

    submit := false
    form := &widget.Form{
        Items: []*widget.FormItem{
            {Text: "Hardware:", Widget: hardwareEntry},
            {Text: "Account:", Widget: accountEntry},
            {Text: "Password:", Widget: passwordEntry},
            {Text: "OpenAI Key:", Widget: openAIKeyEntry},
            {Text: "OpenAI Backend:", Widget: openAIBackendEntry},
            {Text: "Proxy:", Widget: proxyEntry},
            {Text: "MiDID:", Widget: miDIDEntry},
            {Text: "Keywords:", Widget: keywordsEntry},
            {Text: "Change Prompt Keywords:", Widget: changePromptKeywordsEntry},
            {Text: "Prompt:", Widget: promptEntry},
            {Text: "Mute XiaoAI:", Widget: muteXiaoAICheck},
            {Text: "Bot:", Widget: botEntry},
            {Text: "API Base:", Widget: apiBaseEntry},
            {Text: "Use Command:", Widget: useCommandCheck},
            {Text: "Verbose:", Widget: verboseCheck},
            {Text: "Start Conversation:", Widget: startConversationEntry},
            {Text: "End Conversation:", Widget: endConversationEntry},
            {Text: "Stream:", Widget: streamCheck},
            {Text: "Enable Edge TTS:", Widget: enableEdgeTTSCheck},
            {Text: "Edge TTS Voice:", Widget: edgeTTSVoiceEntry},
            {Text: "Token Path:", Widget: tokenPathEntry},
        },
        OnSubmit: func() {
            if submit {
                return
            }
            config.Hardware = hardwareEntry.Text
            config.Account = accountEntry.Text
            config.Password = passwordEntry.Text
            config.OpenAIKey = openAIKeyEntry.Text
            config.OpenAIBackend = openAIBackendEntry.Text
            config.Proxy = proxyEntry.Text
            config.MiDID = miDIDEntry.Text
            config.Keywords = strings.Split(keywordsEntry.Text, ",")
            config.ChangePromptKeywords = strings.Split(changePromptKeywordsEntry.Text, ",")
            config.Prompt = promptEntry.Text
            config.MuteXiaoAI = muteXiaoAICheck.Checked
            config.Bot = botEntry.Text
            config.APIBase = apiBaseEntry.Text
            config.UseCommand = useCommandCheck.Checked
            config.Verbose = verboseCheck.Checked
            config.StartConversation = startConversationEntry.Text
            config.EndConversation = endConversationEntry.Text
            config.Stream = streamCheck.Checked
            config.EnableEdgeTTS = enableEdgeTTSCheck.Checked
            config.EdgeTTSVoice = edgeTTSVoiceEntry.Text
            config.TokenPath = tokenPathEntry.Text

            bot := xiaobot.NewMiBot(&config)
            go bot.Run()
            submit = true
        },
    }

    w.SetContent(container.NewVBox(
        form,
    ))

    w.SetOnClosed(func() {
        // Execute your desired function when the window is closed
        fmt.Println("Window closed, executing function...")
    })
    w.SetFullScreen(true)
    w.ShowAndRun()
}
