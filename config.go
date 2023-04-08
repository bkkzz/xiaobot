package xiaobot

import (
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
    "strings"

    "github.com/BurntSushi/toml"
)

type Config struct {
    Hardware             string                 `json:"hardware" toml:"hardware"`
    Account              string                 `json:"account" toml:"account"`
    Password             string                 `json:"password" toml:"password"`
    OpenAIKey            string                 `json:"openai_key" toml:"openai_key"`
    Proxy                string                 `json:"proxy,omitempty" toml:"proxy,omitempty"`
    MiDID                string                 `json:"mi_did" toml:"mi_did"`
    Keywords             []string               `json:"keyword" toml:"keyword"`
    ChangePromptKeywords []string               `json:"change_prompt_keyword" toml:"change_prompt_keyword"`
    Prompt               string                 `json:"prompt" toml:"prompt"`
    MuteXiaoAI           bool                   `json:"mute_xiaoai" toml:"mute_xiaoai"`
    Bot                  string                 `json:"bot" toml:"bot"`
    APIBase              string                 `json:"api_base,omitempty" toml:"api_base,omitempty"`
    UseCommand           bool                   `json:"use_command" toml:"use_command"`
    Verbose              bool                   `json:"verbose" toml:"verbose"`
    StartConversation    string                 `json:"start_conversation" toml:"start_conversation"`
    EndConversation      string                 `json:"end_conversation" toml:"end_conversation"`
    Stream               bool                   `json:"stream" toml:"stream"`
    EnableEdgeTTS        bool                   `json:"enable_edge_tts" toml:"enable_edge_tts"`
    EdgeTTSVoice         string                 `json:"edge_tts_voice" toml:"edge_tts_voice"`
    GPTOptions           map[string]interface{} `json:"gpt_options" toml:"gpt_options"`
    //BingCookiePath       string                 `json:"bing_cookie_path" toml:"bing_cookie_path"`
    //BingCookies          map[string]interface{} `json:"bing_cookies,omitempty" toml:"bing_cookies,omitempty"`
    TokenPath string `json:"token_path" toml:"token_path"`
}

func (c *Config) PostInit() error {
    if c.OpenAIKey == "" {
        return errors.New("no OpenAI key provided")
    }

    if v := os.Getenv("MI_USER"); v != "" {
        c.Account = v
    }
    if v := os.Getenv("MI_PASS"); v != "" {
        c.Password = v
    }
    if v := os.Getenv("OPENAI_API_KEY"); v != "" {
        c.OpenAIKey = v
    }
    if v := os.Getenv("MI_DID"); v != "" {
        c.MiDID = v
    }
    if c.EdgeTTSVoice == "" {
        c.EdgeTTSVoice = "zh-CN-XiaoxiaoNeural"
    }
    if c.TokenPath == "" {
        c.TokenPath = filepath.Join(os.Getenv("HOME"), ".mi.token")
    }
    return nil
}

func NewConfigFromFile(path string) (*Config, error) {
    config := &Config{}
    if err := config.ReadFromFile(path); err != nil {
        return nil, err
    }
    return config, nil
}

func NewConfigFromOptions(options map[string]interface{}) (*Config, error) {
    config := &Config{}

    if options["config"] != nil {
        err := config.ReadFromFile(options["config"].(string))
        if err != nil {
            return nil, err
        }
    }

    for key, value := range options {
        switch key {
        case "hardware":
            config.Hardware = value.(string)
        case "account":
            config.Account = value.(string)
        case "password":
            config.Password = value.(string)
        case "openai_key":
            config.OpenAIKey = value.(string)
        case "proxy":
            config.Proxy = value.(string)
        case "mi_did":
            config.MiDID = value.(string)
        case "keyword":
            config.Keywords = strings.Split(value.(string), ",")
        case "change_prompt_keyword":
            config.ChangePromptKeywords = strings.Split(value.(string), ",")
        case "prompt":
            config.Prompt = value.(string)
        case "mute_xiaoai":
            config.MuteXiaoAI = value.(bool)
        case "bot":
            config.Bot = value.(string)
        case "api_base":
            config.APIBase = value.(string)
        case "use_command":
            config.UseCommand = value.(bool)
        case "verbose":
            config.Verbose = value.(bool)
        case "start_conversation":
            config.StartConversation = value.(string)
        case "end_conversation":
            config.EndConversation = value.(string)
        case "stream":
            config.Stream = value.(bool)
        case "enable_edge_tts":
            config.EnableEdgeTTS = value.(bool)
        case "edge_tts_voice":
            config.EdgeTTSVoice = value.(string)
        case "gpt_options":
            config.GPTOptions = value.(map[string]interface{})
        case "token_path":
            config.TokenPath = value.(string)
        }
    }

    err := config.PostInit()
    if err != nil {
        return nil, err
    }
    return config, nil
}

func (c *Config) ReadFromJson(configPath string) error {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return err
    }

    return json.Unmarshal(data, c)
}

func (c *Config) ReadFromTOML(configPath string) error {
    _, err := toml.DecodeFile(configPath, c)
    return err
}

func (c *Config) ReadFromFile(configPath string) error {
    if strings.HasSuffix(configPath, ".toml") {
        return c.ReadFromTOML(configPath)
    } else if strings.HasSuffix(configPath, ".json") {
        return c.ReadFromJson(configPath)
    } else {
        return errors.New("invalid config file type")
    }
}

const LatestAskApi = "https://userprofile.mina.mi.com/device_profile/v2/conversation?source=dialogu&hardware={hardware}&timestamp={timestamp}&limit=2"
const COOKIE_TEMPLATE = "deviceId={device_id}; serviceToken={service_token}; userId={user_id}"
const WakeupKeyword = "小爱同学"
const MicoApi = "micoapi"

var HARDWARE_COMMAND_DICT = map[string][2]string{
    //hardware: (tts_command, wakeup_command)
    "LX06":  {"5-1", "5-5"},
    "L05B":  {"5-3", "5-4"},
    "S12A":  {"5-1", "5-5"},
    "LX01":  {"5-1", "5-5"},
    "L06A":  {"5-1", "5-5"},
    "LX04":  {"5-1", "5-4"},
    "L05C":  {"5-3", "5-4"},
    "L17A":  {"7-3", "7-4"},
    "X08E":  {"7-3", "7-4"},
    "LX05A": {"5-1", "5-5"}, // 小爱红外版
    "LX5A":  {"5-1", "5-5"}, // 小爱红外版
    "L07A":  {"5-1", "5-5"}, // Redmi小爱音箱Play(l7a)
    "L15A":  {"7-3", "7-4"},
    "X6A":   {"7-3", "7-4"}, // 小米智能家庭屏6
    // add more here
}

var EDGE_TTS_DICT = map[string]string{
    "用英语": "en-US-AriaNeural",
    "用日语": "ja-JP-NanamiNeural",
    "用法语": "fr-BE-CharlineNeural",
    "用韩语": "ko-KR-SunHiNeural",
    "用德语": "de-AT-JonasNeural",
    //add more here
}

var DEFAULT_COMMAND = []string{"5-1", "5-5"}

var KEY_WORD = []string{"帮我", "请回答"}
var CHANGE_PROMPT_KEY_WORD = []string{"更改提示词"}
var PROMPT = "以下请用100字以内回答，请只回答文字不要带链接"

// simulate_xiaoai_question
var MI_ASK_SIMULATE_DATA = map[string]interface{}{
    "code":    0,
    "message": "Success",
    "data":    `{"bitSet":[0,1,1],"records":[{"bitSet":[0,1,1,1,1],"answers":[{"bitSet":[0,1,1,1],"type":"TTS","tts":{"bitSet":[0,1],"text":"Fake Answer"}}],"time":1677851434593,"query":"Fake Question","requestId":"fada34f8fa0c3f408ee6761ec7391d85"}],"nextEndTime":1677849207387}`,
}
