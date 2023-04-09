package xiaobot

import (
    "github.com/BurntSushi/toml"
    "log"
    "os"
    "testing"
    "time"
)

func TestMiBot(t *testing.T) {
    user := os.Getenv("MI_USER")
    if user == "" {
        t.Skip("MI_USER not set")
    }
    p := "tmp.token"
    opt := map[string]interface{}{
        "hardware":        "LX01",
        "token_path":      p,
        "openai_key":      "test",
        "enable_edge_tts": true,
    }
    cfg, err := NewConfigFromOptions(opt)
    if err != nil {
        t.Error(err)
    }
    bt := NewMiBot(cfg)
    err = bt.initAllData()
    if err != nil {
        t.Error(err)
    }

    r, err := bt.getLatestAsk()
    if err != nil {
        t.Error(err)
    }
    log.Println(r)

    b, err := bt.speakerIsPlaying()
    if err != nil {
        t.Error(err)
    }
    log.Println(b)

    err = bt.stopSpeaker()
    if err != nil {
        t.Error(err)
    }

    err = bt.miTTS("小爱同学吃了一条鱼", false)
    if err != nil {
        t.Error(err)
    }
    time.Sleep(1 * time.Second)
    err = bt.edgeTTS("小爱同学吃了两条鱼", "zh-CN-XiaoxiaoNeural")
    if err != nil {
        t.Error(err)
    }
    os.Remove(p)
}

func TestMi0(t *testing.T) {
    b := Config{}
    toml.NewEncoder(os.Stdout).Encode(b)
}
