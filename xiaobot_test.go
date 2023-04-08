package xiaobot

import (
    "log"
    "os"
    "testing"
)

func TestMiBot(t *testing.T) {
    user := os.Getenv("MI_USER")
    if user == "" {
        t.Skip("MI_USER not set")
    }
    p := "tmp.token"
    opt := map[string]interface{}{
        "hardware":   "LX01",
        "token_path": p,
        "openai_key": "test",
    }
    cfg, err := NewConfigFromOptions(opt)
    if err != nil {
        t.Error(err)
    }
    bt := NewMiBot(cfg)
    err = bt.InitAllData()
    if err != nil {
        t.Error(err)
    }
    r, err := bt.GetLatestAskFromXiaoAi()
    if err != nil {
        t.Error(err)
    }
    log.Println(r)

    b, err := bt.GetIfXiaoAiIsPlaying()
    if err != nil {
        t.Error(err)
    }
    log.Println(b)
    os.Remove(p)
}
