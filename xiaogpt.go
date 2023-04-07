package xiaogpt

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "math/rand"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/longbai/miservice"
)

type MiGPT struct {
    config      Config
    miTokenHome string

    deviceID string
    parentID interface{} // Adjust the type accordingly

    account     *miservice.Account
    minaService *miservice.AIService
    miioService *miservice.IOService

    pollingEvent   *sync.Cond
    newRecordEvent *sync.Cond
    tempDir        string

    Chatbot        Bot
    LastTimestamp  int64
    LastRecord     map[string]interface{}
    InConversation bool
}

func NewMiGPT(config Config) *MiGPT {
    miTokenHome := filepath.Join(os.Getenv("HOME"), ".mi.token")
    lastTimestamp := time.Now().Unix() * 1000

    return &MiGPT{
        config:         config,
        Chatbot:        &GhatGptBot{},
        miTokenHome:    miTokenHome,
        LastTimestamp:  lastTimestamp,
        pollingEvent:   sync.NewCond(&sync.Mutex{}),
        newRecordEvent: sync.NewCond(&sync.Mutex{}),
    }
}

func (mgpt *MiGPT) PollLatestAsk() {
    go func() {
        for {
            log.Printf("Now listening xiaoai new message timestamp: %d", mgpt.LastTimestamp)
            _, err := mgpt.GetLatestAskFromXiaoai()
            if err != nil {
                log.Printf("Error getting latest ask: %v", err)
            }
            start := time.Now()
            mgpt.pollingEvent.L.Lock()
            mgpt.pollingEvent.Wait()
            mgpt.pollingEvent.L.Unlock()
            elapsed := time.Since(start)
            if elapsed < time.Second {
                time.Sleep(time.Second - elapsed)
            }
        }
    }()
}

func (mgpt *MiGPT) InitAllData() error {
    err := mgpt.loginMiboy()
    if err != nil {
        return err
    }
    err = mgpt.initDataHardware()
    if err != nil {
        return err
    }
    // Update the cookie jar here
    // mgpt.cookieJar = ...

    if mgpt.config.EnableEdgeTTS {
        go mgpt.StartHTTPServer()
    }
    return nil
}

func (mgpt *MiGPT) loginMiboy() error {
    t := miservice.NewTokenStore(mgpt.miTokenHome)
    account := miservice.NewAccount(
        mgpt.config.Account,
        mgpt.config.Password,
        t,
    )
    mgpt.account = account
    err := account.Login("micoapi")
    if err != nil {
        return err
    }
    mgpt.minaService = miservice.NewAIService(account)
    mgpt.miioService = miservice.NewIOService(account, nil)
    return nil
}

func (mgpt *MiGPT) initDataHardware() error {
    if mgpt.config.Cookie != "" {
        // if you use cookie do not need init
        return nil
    }
    hardwareData, err := mgpt.minaService.DeviceList(0)
    if err != nil {
        return err
    }
    for _, h := range hardwareData {
        if did := mgpt.config.MiDID; did != "" {
            if h.MiotDID == did {
                mgpt.deviceID = h.DeviceID
                break
            }
        }
        if h.Hardware == mgpt.config.Hardware {
            mgpt.deviceID = h.DeviceID
            break
        }
    }
    if mgpt.deviceID == "" {
        return errors.New("we have no hardware: " + mgpt.config.Hardware + " please use micli mina to check")
    }
    if mgpt.config.MiDID == "" {
        devices, _, err := mgpt.miioService.DeviceList(false, 0)
        if err != nil {
            return err
        }
        found := false
        for _, d := range devices {
            if strings.HasSuffix(d.Model, strings.ToLower(mgpt.config.Hardware)) {
                mgpt.config.MiDID = d.Did
                found = true
                break
            }
        }
        if !found {
            return errors.New("cannot find did for hardware: " + mgpt.config.Hardware + " please set it via MI_DID env")
        }
    }
    return nil
}

func (mgpt *MiGPT) getCookie(u string) (http.CookieJar, error) {
    if mgpt.config.Cookie != "" {
        cookieJar, err := parseCookieString(mgpt.config.Cookie)
        if err != nil {
            return nil, err
        }
        // set deviceID from cookie
        cookieMap := make(map[string]string)
        ul, err := url.Parse(u)
        if err != nil {
            return nil, err
        }
        for _, cookie := range cookieJar.Cookies(ul) {
            cookieMap[cookie.Name] = cookie.Value
        }
        mgpt.deviceID = cookieMap["deviceId"]
        return cookieJar, nil
    } else {
        file, err := os.Open(mgpt.miTokenHome)
        if err != nil {
            return nil, err
        }
        defer file.Close()
        var userData map[string]interface{}
        err = json.NewDecoder(file).Decode(&userData)
        if err != nil {
            return nil, err
        }

        userId := userData["userId"].(string)
        serviceToken := userData["micoapi"].(string)
        cookieString := fmt.Sprintf("deviceId=%s;serviceToken=%s;userId=%s",
            mgpt.deviceID, serviceToken, userId)
        return parseCookieString(cookieString)
    }
}

func (m *MiGPT) SimulateXiaoaiQuestion() (map[string]string, error) {
    data := map[string]string{}
    dataDict := make(map[string]interface{})
    err := json.Unmarshal([]byte(data["data"]), &dataDict)
    if err != nil {
        return nil, err
    }
    records := dataDict["records"].([]interface{})
    record := records[0].(map[string]interface{})
    fmt.Print("Enter the new query: ")
    var query string
    _, _ = fmt.Scanln(&query)
    record["query"] = query
    record["time"] = time.Now().Unix() * 1000
    data["data"], _ = json.Marshal(dataDict)
    time.Sleep(1 * time.Second)
    return data, nil
}

func queryIn(q string, keywords []string) bool {
    for _, k := range keywords {
        if strings.HasPrefix(q, k) {
            return true
        }
    }
    return false
}

func (m *MiGPT) NeedAskGPT(record map[string]interface{}) bool {
    query := record["query"].(string)
    return m.InConversation && !strings.HasPrefix(query, WAKEUP_KEYWORD) || queryIn(query, m.config.Keywords)
}

func (m *MiGPT) NeedChangePrompt(record map[string]interface{}) bool {
    query := record["query"].(string)
    return queryIn(query, m.config.ChangePromptKeywords)
}

func (m *MiGPT) ChangePrompt(newPrompt string) {
    newPrompt = strings.TrimPrefix(newPrompt, m.config.ChangePromptKeywords[0])
    newPrompt = "以下都" + newPrompt
    fmt.Printf("Prompt from %s change to %s\n", m.config.Prompt, newPrompt)
    m.config.Prompt = newPrompt
    if len(m.Chatbot.GetHistory()) > 0 {
        m.Chatbot.SetHistoryPrompt(newPrompt)
    }
}

func (m *MiGPT) GetLatestAskFromXiaoai() (map[string]interface{}, error) {
    retries := 2
    for i := 0; i < retries; i++ {
        r, err := GetLatestAskAPI(m.config.Hardware, strconv.FormatInt(time.Now().Unix()*1000, 10))
        if err != nil {
            fmt.Println("get latest ask from xiaoai error, retry")
            continue
        }
        data := make(map[string]interface{})
        err = json.Unmarshal(r, &data)
        if err != nil {
            fmt.Println("get latest ask from xiaoai error, retry")
            continue
        }
        lastQuery, err := m.GetLastQuery(data)
        if err == nil {
            return lastQuery, nil
        }
    }
    return nil, fmt.Errorf("failed to get latest ask from xiaoai after %d retries", retries)
}

func (m *MiGPT) GetLastQuery(data map[string]interface{}) (map[string]interface{}, error) {
    if d, ok := data["data"].(string); ok {
        recordsData := make(map[string]interface{})
        err := json.Unmarshal([]byte(d), &recordsData)
        if err != nil {
            return nil, err
        }
        records := recordsData["records"].([]interface{})
        if len(records) == 0 {
            return nil, fmt.Errorf("no records found")
        }
        lastRecord := records[0].(map[string]interface{})
        timestamp := int64(lastRecord["time"].(float64))
        if timestamp > m.LastTimestamp {
            m.LastTimestamp = timestamp
            m.LastRecord = lastRecord
            m.newRecordEvent.Signal()
            return lastRecord, nil
        }
    }
    return nil, fmt.Errorf("failed to get last query")
}

func (m *MiGPT) DoTTS(value string, waitForFinish bool) {
    if !m.Config.UseCommand {
        if err := m.MinaService.TextToSpeech(m.DeviceID, value); err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    } else {
        // Implement miio_command here
    }
    if waitForFinish {
        elapse := CalculateTTSElapse(value)
        time.Sleep(elapse)
        m.WaitForTTSFinish()
    }
}

func (m *MiGPT) WaitForTTSFinish() {
    for {
        // Implement the GetIfXiaoaiIsPlaying function
        isPlaying, _ := GetIfXiaoaiIsPlaying()
        if !isPlaying {
            break
        }
        time.Sleep(1 * time.Second)
    }
}

func (m *MiGPT) StartHTTPServer() {
    // Set the port range
    portRange := make([]int, 40)
    for i := 8050; i < 8090; i++ {
        portRange[i-8050] = i
    }

    // Get a random port from the range
    rand.Seed(time.Now().UnixNano())
    m.Port = rand.Intn(len(portRange))

    m.TempDir, _ = ioutil.TempDir("", "xiaogpt-tts-")

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, filepath.Join(m.TempDir, filepath.Base(r.URL.Path)))
    })

    go func() {
        err := http.ListenAndServe(fmt.Sprintf(":%d", m.Port), nil)
        if err != nil {
            fmt.Printf("Error starting HTTP server: %v\n", err)
        }
    }()

    // Implement the GetHostname function
    m.Hostname = getHostname()

    fmt.Printf("Serving on %s:%d\n", m.Hostname, m.Port)
}

func (m *MiGPT) Text2MP3(text string, ttsLang string) (string, float64, error) {
    // Implement the EdgeTTS Communicate function and Stream method
    return "", 0, fmt.Errorf("not implemented")
}

func (m *MiGPT) EdgeTTS(textStream chan string, ttsLang string) {
    runTTS := func(textStream chan string, ttsLang string, queue chan<- interface{}) {
        for text := range textStream {
            url, duration, err := m.Text2MP3(text, ttsLang)
            if err != nil {
                fmt.Printf("Error: %v\n", err)
                continue
            }
            queue <- struct {
                URL      string
                Duration float64
            }{URL: url, Duration: duration}
        }
        queue <- "EOF"
    }

    queue := make(chan interface{})
    fmt.Printf("Edge TTS with voice=%s\n", ttsLang)

    go runTTS(textStream, ttsLang, queue)

    var wg sync.WaitGroup
    wg.Add(1)

    go func() {
        defer wg.Done()

        for {
            item := <-queue
            if item == "EOF" {
                break
            }
            data := item.(struct {
                URL      string
                Duration float64
            })
            fmt.Printf("play: %s\n", data.URL)
            if err := m.minaService.PlayByURL(m.DeviceID, data.URL); err != nil {
                fmt.Printf("Error: %v\n", err)
            }
            time.Sleep(time.Duration(data.Duration) * time.Second)
            m.WaitForTTSFinish()
        }
    }()

    wg.Wait()
}

func (m *MiGPT) Normalize(message string) string {
    message = strings.TrimSpace(message)
    message = strings.ReplaceAll(message, " ", "--")
    message = strings.ReplaceAll(message, "\n", "，")
    message = strings.ReplaceAll(message, "\"", "，")
    return message
}

func (m *MiGPT) AskGPT(query string) (chan string, error) {
    if !m.Config.Stream {
        // TODO: Implement synchronous ask_gpt logic
        return nil, fmt.Errorf("synchronous AskGPT not implemented")
    }

    queue := make(chan string)
    collectStream := func(queue chan string, wg *sync.WaitGroup) {
        defer wg.Done()

        // TODO: Replace the following loop with an actual implementation
        // of the chatbot.ask_stream method.
        for i := 0; i < 5; i++ {
            message := fmt.Sprintf("Sample message %d", i)
            queue <- m.Normalize(message)
        }
        queue <- "EOF"
    }

    var wg sync.WaitGroup
    wg.Add(1)

    go collectStream(queue, &wg)

    return queue, nil
}

func (m *MiGPT) StopIfXiaoaiIsPlaying() error {
    isPlaying, err := m.GetIfXiaoaiIsPlaying()
    if err != nil {
        return err
    }

    if isPlaying {
        // TODO: Implement MinaService.PlayerPause() method
        if err := m.MinaService.PlayerPause(m.DeviceID); err != nil {
            return err
        }
    }

    return nil
}

func (m *MiGPT) RunForever() {
    // TODO: Implement init_all_data() method
    // TODO: Implement poll_latest_ask() method

    fmt.Printf("Running xiaogpt now, 用`%s`开头来提问\n", "/".join(m.Config.Keyword))
    fmt.Printf("或用`%s`开始持续对话\n", m.Config.StartConversation)

    for {
        m.pollingEvent.Broadcast()
        m.newRecordEvent.L.Lock()
        m.newRecordEvent.Wait()
        m.newRecordEvent.L.Unlock()

        newRecord := m.LastRecord
        m.pollingEvent.Broadcast()

        query := strings.TrimSpace(newRecord["query"].(string))

        if query == m.config.StartConversation {
            if !m.InConversation {
                fmt.Println("开始对话")
                m.InConversation = true
                m.WakeupXiaoai()
            }
            m.StopIfXiaoaiIsPlaying()
            continue
        } else if query == m.config.EndConversation {
            if m.InConversation {
                fmt.Println("结束对话")
                m.InConversation = false
            }
            m.StopIfXiaoaiIsPlaying()
            continue
        }

        // TODO: Implement need_change_prompt() and _change_prompt() methods
        // TODO: Implement need_ask_gpt() method

        // Drop 帮我回答
        regexPattern := fmt.Sprintf("^%s", strings.Join(m.Config.Keyword, "|"))
        re := regexp.MustCompile(regexPattern)
        query = re.ReplaceAllString(query, "")

        fmt.Println(strings.Repeat("-", 20))
        fmt.Printf("问题：%s？\n", query)
        if len(m.Chatbot.History) == 0 {
            query = fmt.Sprintf("%s，%s", query, m.Config.Prompt)
        }
        if m.config.MuteXiaoai {
            m.StopIfXiaoaiIsPlaying()
        } else {
            time.Sleep(8 * time.Second)
        }
        m.DoTTS("正在问GPT请耐心等待", false)

        // TODO: Implement logic to handle Xiaoai's answer

        fmt.Print("以下是GPT的回答: ")
        // TODO: Implement logic to handle GPT's answer

        if m.InConversation {
            fmt.Printf("继续对话, 或用`%s`结束对话\n", m.Config.EndConversation)
            m.WakeupXiaoai()
        }
    }
}
