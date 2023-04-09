package xiaobot

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/longbai/xiaobot/jarvis"
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
    records     chan Record

    assistant      jarvis.Jarvis
    LastTimestamp  int64
    InConversation bool

    tempDir  string
    port     int
    hostname string
}

func NewMiBot(config *Config) *MiBot {
    lastTimestamp := time.Now().Unix() * 1000
    tokens := miservice.NewTokenStore(config.TokenPath)
    return &MiBot{
        config:        config,
        token:         tokens,
        LastTimestamp: lastTimestamp,
        records:       make(chan Record, 100),
    }
}

func (mt *MiBot) pollLatestAsk() {
    for {
        log.Printf("Now listening xiaoai new message timestamp: %d", mt.LastTimestamp)
        start := time.Now()
        records, err := mt.getLatestAsk()
        if err != nil {
            log.Printf("Error getting latest ask: %v", err)
        } else {
            if len(records.Records) > 0 {
                r := records.Records[0]
                if r.Time*1000 > mt.LastTimestamp {
                    mt.LastTimestamp = r.Time * 1000
                    for _, r := range records.Records {
                        mt.records <- r
                    }
                }
            }
        }
        elapsed := time.Since(start)
        if elapsed < time.Second {
            time.Sleep(time.Second - elapsed)
        }
    }
}

func (mt *MiBot) initAllData() error {
    err := mt.loginMiBot()
    if err != nil {
        return err
    }
    err = mt.initDataHardware()
    if err != nil {
        return err
    }
    switch mt.config.Bot {
    case "gpt":
        mt.assistant = &jarvis.GhatGpt{Key: mt.config.OpenAIKey, Backend: mt.config.OpenAIBackend, Proxy: mt.config.Proxy}
    case "newbing":
        mt.assistant = &jarvis.NewBing{}
    default:
        return errors.New("unknown bot: " + mt.config.Bot)
    }

    if mt.config.EnableEdgeTTS {
        go mt.startEdgeServer()
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
    if mt.deviceID == "" {
        return errors.New("we have no hardware: " + mt.config.Hardware + " please use micli mina to check")
    }

    return nil
}

func queryIn(q string, keywords []string) bool {
    log.Println(q, keywords)
    for _, k := range keywords {
        if strings.HasPrefix(q, k) {
            return true
        }
    }
    return false
}

func (mt *MiBot) needAskJarvis(query string) bool {
    return (mt.InConversation && !strings.HasPrefix(query, WakeupKeyword)) || queryIn(query, mt.config.Keywords)
}

func (mt *MiBot) needChangePrompt(query string) bool {
    return queryIn(query, mt.config.ChangePromptKeywords)
}

func (mt *MiBot) changePrompt(newPrompt string) {
    newPrompt = strings.TrimPrefix(newPrompt, mt.config.ChangePromptKeywords[0])
    newPrompt = "以下都" + newPrompt
    log.Printf("Prompt from %s change to %s\n", mt.config.Prompt, newPrompt)
    mt.config.Prompt = newPrompt
    if len(mt.assistant.GetHistory()) > 0 {
        mt.assistant.SetHistoryPrompt(newPrompt)
    }
}

type ret struct {
    Code    int    `json:"code,omitempty"`
    Message string `json:"message,omitempty"`
    Data    string `json:"data,omitempty"`
}

func (mt *MiBot) getLatestAsk() (*Records, error) {
    retries := 2
    var err error
    for i := 0; i < retries; i++ {
        u := strings.Replace(LatestAskApi, "{hardware}", mt.config.Hardware, -1)
        u = strings.Replace(u, "{timestamp}", strconv.FormatInt(time.Now().Unix()*1000, 10), -1)
        var result ret
        err = mt.account.Request(MicoApi, u, nil, func(tokens *miservice.Tokens, cookie map[string]string) url.Values {
            cookie["deviceId"] = mt.deviceID
            return nil
        }, nil, false, &result)
        if err != nil {
            log.Println("get latest ask from xiaoai error, retry", err)
            continue
        }

        var records Records
        err = json.Unmarshal([]byte(result.Data), &records)
        if err != nil {
            log.Println("get latest ask from xiaoai error", err)
            continue
        }
        return &records, nil
    }

    return nil, err
}

type Record struct {
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
}

type Records struct {
    BitSet      []int    `json:"bitSet"`
    Records     []Record `json:"records"`
    NextEndTime int64    `json:"nextEndTime"`
}

func (mt *MiBot) askJarvis(query string) (string, error) {
    return mt.assistant.Ask(query)
}

type StatusInfo struct {
    Status   int `json:"status"`
    Volume   int `json:"volume"`
    LoopType int `json:"loop_type"`
}

func (mt *MiBot) speakerIsPlaying() (bool, error) {
    res, err := mt.minaService.PlayerGetStatus(mt.deviceID)
    if err != nil {
        return false, err
    }
    var info StatusInfo
    json.Unmarshal([]byte(res.Data.Info), &info)
    return info.Status == 1, nil
}

func (mt *MiBot) stopSpeaker() error {
    _, err := mt.minaService.PlayerPause(mt.deviceID)
    return err
}

const WakeupIndex = 1
const TTSIndex = 0

func (mt *MiBot) hardwareCommand(index int) string {
    v, ok := HardwareCommandDict[mt.config.Hardware]
    if !ok {
        v = DefaultCommand
    }
    return v[index]
}

func actionId(action string) []int {
    ids := strings.Split(action, "-")
    siid, _ := strconv.Atoi(ids[0])
    iid, _ := strconv.Atoi(ids[1])
    return []int{siid, iid}
}

func (mt *MiBot) wakeUp() {
    w := mt.hardwareCommand(WakeupIndex)
    mt.miioService.MiotAction(mt.deviceID, actionId(w), []interface{}{WakeupKeyword, 0})
}

func (mt *MiBot) miTTS(value string, waitForFinish bool) error {
    if mt.config.UseCommand {
        t := mt.hardwareCommand(TTSIndex)
        c, err := mt.miioService.MiotAction(mt.deviceID, actionId(t), []interface{}{value})
        log.Println("TTS command result", c, err)
        return err
    } else {
        if _, err := mt.minaService.TextToSpeech(mt.deviceID, value); err != nil {
            log.Printf("Error: %v\n", err)
            return err
        }
    }
    if waitForFinish {
        elapse := calculateTtsElapse(value)
        time.Sleep(elapse)
        mt.waitForTTSFinish()
    }
    return nil
}

func (mt *MiBot) waitForTTSFinish() {
    for {
        isPlaying, _ := mt.speakerIsPlaying()
        if !isPlaying {
            break
        }
        time.Sleep(1 * time.Second)
    }
}

func (mt *MiBot) Run() error {
    log.Printf("Running xiaogpt now, 用`%s`开头来提问\n", strings.Join(mt.config.Keywords, "/"))
    log.Printf("或用`%s`开始持续对话\n", mt.config.StartConversation)
    if err := mt.initAllData(); err != nil {
        return err
    }
    go mt.pollLatestAsk()
    for {
        record := <-mt.records
        query := record.Query

        if query == mt.config.StartConversation {
            if !mt.InConversation {
                log.Println("开始对话")
                mt.InConversation = true
                mt.wakeUp()
            }
            mt.stopSpeaker()
            continue
        } else if query == mt.config.EndConversation {
            if mt.InConversation {
                log.Println("结束对话")
                mt.InConversation = false
            }
            mt.stopSpeaker()
            continue
        }

        if mt.needChangePrompt(query) {
            log.Println("需要改变提示语", record)
            mt.changePrompt(query)
        }

        if !mt.needAskJarvis(query) {
            log.Println("不需要问GPT", query, mt.config.Keywords)
            continue
        }

        // Drop 帮我回答
        regexPattern := fmt.Sprintf("^%s", strings.Join(mt.config.Keywords, "|"))
        re := regexp.MustCompile(regexPattern)
        query = re.ReplaceAllString(query, "")

        log.Println(strings.Repeat("-", 20))
        log.Printf("问题：%s？\n", query)
        //if len(mt.assistant.GetHistory()) == 0 {
        //    query = fmt.Sprintf("%s，%s", query, mt.config.Prompt)
        //}
        if mt.config.MuteXiaoAI {
            mt.stopSpeaker()
        } else {
            time.Sleep(8 * time.Second)
        }
        mt.miTTS("正在问GPT请耐心等待", false)

        if len(record.Answers) != 0 {
            var str string
            for _, a := range record.Answers {
                str += a.Tts.Text
            }
            log.Printf("以下是小爱的回答: %s\n", str)
        } else {
            log.Println("小爱没回")
        }

        message, err := mt.askJarvis(query)
        if err != nil {
            log.Printf("AskGPT error: %v", err)
            continue
        }

        log.Print("以下是GPT的回答: ")
        if mt.config.EnableEdgeTTS {
            ttsLang := findKeyByPartialString(EdgeTtsDict, query)
            if ttsLang == "" {
                ttsLang = mt.config.EdgeTTSVoice
            }
            mt.edgeTTS(message, ttsLang)
        } else {
            mt.miTTS(message, true)
        }
        if mt.InConversation {
            log.Printf("继续对话, 或用`%s`结束对话\n", mt.config.EndConversation)
            mt.wakeUp()
        }
    }
    return nil
}
