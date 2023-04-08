package xiaobot

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "net/url"
    "regexp"
    "strconv"
    "strings"
    "time"

    "github.com/longbai/miservice"
)

type MiBot struct {
    config *Config
    token  miservice.TokenStore

    deviceID string
    parentID interface{} // Adjust the type accordingly

    account     *miservice.Account
    minaService *miservice.AIService
    miioService *miservice.IOService

    ChatBot        Bot
    LastTimestamp  int64
    InConversation bool

    tempDir  string
    port     int
    hostname string

    records chan *Records
}

func NewMiBot(config *Config) *MiBot {
    lastTimestamp := time.Now().Unix() * 1000
    tokens := miservice.NewTokenStore(config.TokenPath)
    return &MiBot{
        config:        config,
        ChatBot:       &GhatGptBot{},
        token:         tokens,
        LastTimestamp: lastTimestamp,
        records:       make(chan *Records, 100),
    }
}

func (mt *MiBot) PollLatestAsk() {
    for {
        log.Printf("Now listening xiaoai new message timestamp: %d", mt.LastTimestamp)
        start := time.Now()
        _, err := mt.GetLatestAskFromXiaoAi()
        if err != nil {
            log.Printf("Error getting latest ask: %v", err)
        }
        elapsed := time.Since(start)
        if elapsed < time.Second {
            time.Sleep(time.Second - elapsed)
        }
    }
}

func (mt *MiBot) InitAllData() error {
    err := mt.loginMiBot()
    if err != nil {
        return err
    }
    err = mt.initDataHardware()
    if err != nil {
        return err
    }

    if mt.config.EnableEdgeTTS {
        go mt.StartHTTPServer()
    }
    return nil
}

func (mt *MiBot) loginMiBot() error {
    account := miservice.NewAccount(
        mt.config.Account,
        mt.config.Password,
        mt.token,
    )
    mt.account = account
    err := account.Login(MicoApi)
    if err != nil {
        return err
    }
    mt.minaService = miservice.NewAIService(account)
    mt.miioService = miservice.NewIOService(account, nil)
    return nil
}

func (mt *MiBot) initDataHardware() error {
    if mt.config.MiDID == "" {
        devices, err := mt.miioService.DeviceList(false, 0)
        if err != nil {
            return err
        }
        found := false
        for _, d := range devices {
            if strings.HasSuffix(d.Model, strings.ToLower(mt.config.Hardware)) {
                mt.config.MiDID = d.Did
                found = true
                break
            }
        }
        if !found {
            return errors.New("cannot find did for hardware: " + mt.config.Hardware + " please set it via MI_DID env")
        }
    }

    hardwareData, err := mt.minaService.DeviceList(0)
    if err != nil {
        return err
    }
    for _, h := range hardwareData {
        if h.MiotDID == mt.config.MiDID {
            mt.deviceID = h.DeviceID
            break
        }
    }
    if mt.deviceID == "" {
        for _, h := range hardwareData {
            if h.Hardware == mt.config.Hardware {
                mt.deviceID = h.DeviceID
                break
            }
        }
    }
    fmt.Println("deivd", mt.deviceID)
    if mt.deviceID == "" {
        return errors.New("we have no hardware: " + mt.config.Hardware + " please use micli mina to check")
    }

    return nil
}

func queryIn(q string, keywords []string) bool {
    for _, k := range keywords {
        if strings.HasPrefix(q, k) {
            return true
        }
    }
    return false
}

func (mt *MiBot) NeedAskGPT(r *Records) bool {
    //record map[string]interface{}
    //query := record["query"].(string)
    query := ""
    return mt.InConversation && !strings.HasPrefix(query, WakeupKeyword) || queryIn(query, mt.config.Keywords)
}

func (mt *MiBot) NeedChangePrompt(r *Records) bool {
    //record map[string]interface{}
    //query := record["query"].(string)
    query := ""
    return queryIn(query, mt.config.ChangePromptKeywords)
}

func (mt *MiBot) ChangePrompt(newPrompt string) {
    newPrompt = strings.TrimPrefix(newPrompt, mt.config.ChangePromptKeywords[0])
    newPrompt = "以下都" + newPrompt
    log.Printf("Prompt from %s change to %s\n", mt.config.Prompt, newPrompt)
    mt.config.Prompt = newPrompt
    if len(mt.ChatBot.GetHistory()) > 0 {
        mt.ChatBot.SetHistoryPrompt(newPrompt)
    }
}

type ret struct {
    Code    int    `json:"code,omitempty"`
    Message string `json:"message,omitempty"`
    Data    string `json:"data,omitempty"`
}

func (mt *MiBot) GetLatestAskFromXiaoAi() (*Records, error) {
    retries := 2
    for i := 0; i < retries; i++ {
        u := strings.Replace(LatestAskApi, "{hardware}", mt.config.Hardware, -1)
        u = strings.Replace(u, "{timestamp}", strconv.FormatInt(time.Now().Unix()*1000, 10), -1)
        var result ret
        err := mt.account.Request(MicoApi, u, nil, func(tokens *miservice.Tokens, cookie map[string]string) url.Values {
            cookie["deviceId"] = mt.deviceID
            return nil
        }, nil, false, &result)
        if err != nil {
            log.Println("get latest ask from xiaoai error, retry", err)
            continue
        }

        lastQuery, err := mt.GetLastQuery(&result)
        if err == nil {
            return lastQuery, nil
        }
    }
    return nil, fmt.Errorf("failed to get latest ask from xiaoai after %d retries", retries)
}

type Records struct {
    BitSet  []int `json:"bitSet"`
    Records []struct {
        BitSet  []int `json:"bitSet"`
        Answers []struct {
            BitSet []int  `json:"bitSet"`
            Type   string `json:"type"`
            Tts    struct {
                BitSet []int  `json:"bitSet"`
                Text   string `json:"text"`
            } `json:"tts"`
        } `json:"answers"`
        Time      int64  `json:"time"`
        Query     string `json:"query"`
        RequestID string `json:"requestId"`
    } `json:"records"`
    NextEndTime int64 `json:"nextEndTime"`
}

func (mt *MiBot) GetLastQuery(data *ret) (*Records, error) {
    var records Records
    err := json.Unmarshal([]byte(data.Data), &records)
    if err != nil {
        return nil, err
    }
    return &records, nil
}

func (mt *MiBot) AskOther(query string) (string, error) {
    return mt.ChatBot.Ask(query)
}

type StatusInfo struct {
    Status   int `json:"status"`
    Volume   int `json:"volume"`
    LoopType int `json:"loop_type"`
}

func (mt *MiBot) GetIfXiaoAiIsPlaying() (bool, error) {
    res, err := mt.minaService.PlayerGetStatus(mt.deviceID)
    if err != nil {
        return false, err
    }
    var info StatusInfo
    json.Unmarshal([]byte(res.Data.Info), &info)
    return info.Status == 1, nil
}

func (mt *MiBot) StopIfXiaoAiIsPlaying() error {
    isPlaying, err := mt.GetIfXiaoAiIsPlaying()
    if err != nil {
        return err
    }

    if isPlaying {
        if _, err := mt.minaService.PlayerPause(mt.deviceID); err != nil {
            return err
        }
    }

    return nil
}

func (mt *MiBot) WakeupCommand() string {
    if r, ok := HARDWARE_COMMAND_DICT[mt.config.Hardware]; ok {
        return r[1]
    }
    return DEFAULT_COMMAND[1]
}

func (mt *MiBot) TtsCommand() string {
    if r, ok := HARDWARE_COMMAND_DICT[mt.config.Hardware]; ok {
        return r[0]
    }
    return DEFAULT_COMMAND[0]
}

func ActionId(action string) []int {
    ids := strings.Split(action, "-")
    siid, _ := strconv.Atoi(ids[0])
    iid, _ := strconv.Atoi(ids[1])
    return []int{siid, iid}
}

func (mt *MiBot) WakeUp() {
    w := mt.WakeupCommand()
    mt.miioService.MiotAction(mt.deviceID, ActionId(w), []interface{}{WakeupKeyword, 0})
}

func (r *Records) query() string {
    //    strings.TrimSpace(newRecord["query"].(string)
    return ""
}

func (r *Records) answer() string {
    return ""
}

func (mt *MiBot) Run() error {
    log.Printf("Running xiaogpt now, 用`%s`开头来提问\n", strings.Join(mt.config.Keywords, "/"))
    log.Printf("或用`%s`开始持续对话\n", mt.config.StartConversation)
    if err := mt.InitAllData(); err != nil {
        return err
    }
    go mt.PollLatestAsk()
    for {
        record := <-mt.records
        query := record.query()

        if query == mt.config.StartConversation {
            if !mt.InConversation {
                log.Println("开始对话")
                mt.InConversation = true
                mt.WakeUp()
            }
            mt.StopIfXiaoAiIsPlaying()
            continue
        } else if query == mt.config.EndConversation {
            if mt.InConversation {
                log.Println("结束对话")
                mt.InConversation = false
            }
            mt.StopIfXiaoAiIsPlaying()
            continue
        }

        if mt.NeedChangePrompt(record) {
            log.Println("需要改变提示语", record)
            mt.ChangePrompt(query)
        }

        if !mt.NeedAskGPT(record) {
            log.Println("不需要问GPT", record)
            continue
        }

        // Drop 帮我回答
        regexPattern := fmt.Sprintf("^%s", strings.Join(mt.config.Keywords, "|"))
        re := regexp.MustCompile(regexPattern)
        query = re.ReplaceAllString(query, "")

        log.Println(strings.Repeat("-", 20))
        log.Printf("问题：%s？\n", query)
        if len(mt.ChatBot.GetHistory()) == 0 {
            query = fmt.Sprintf("%s，%s", query, mt.config.Prompt)
        }
        if mt.config.MuteXiaoAI {
            mt.StopIfXiaoAiIsPlaying()
        } else {
            time.Sleep(8 * time.Second)
        }
        mt.DoTTS("正在问GPT请耐心等待", false)

        if record.answer() != "" {
            log.Printf("以下是小爱的回答: %s\n", record.answer())
        } else {
            log.Println("小爱没回")
        }

        message, err := mt.AskOther(query)
        if err != nil {
            log.Printf("AskGPT error: %v", err)
            continue
        }

        log.Print("以下是GPT的回答: ")
        if mt.config.EnableEdgeTTS {
            ttsLang := findKeyByPartialString(EDGE_TTS_DICT, query)
            if ttsLang == "" {
                ttsLang = mt.config.EdgeTTSVoice
            }
            mt.EdgeTTS(message, ttsLang)
        } else {
            mt.DoTTS(message, true)
        }
        if mt.InConversation {
            log.Printf("继续对话, 或用`%s`结束对话\n", mt.config.EndConversation)
            mt.WakeUp()
        }
    }
    return nil
}
