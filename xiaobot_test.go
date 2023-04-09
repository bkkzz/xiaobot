package xiaobot

import (
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

// simulate_xiaoai_question
var miAskSimulateData = map[string]interface{}{
    "code":    0,
    "message": "Success",
    "data":    `{"bitSet":[0,1,1],"records":[{"bitSet":[0,1,1,1,1],"answers":[{"bitSet":[0,1,1,1],"type":"TTS","tts":{"bitSet":[0,1],"text":"Fake Answer"}}],"time":1677851434593,"query":"Fake Question","requestId":"fada34f8fa0c3f408ee6761ec7391d85"}],"nextEndTime":1677849207387}`,
}
