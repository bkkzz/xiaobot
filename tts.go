package xiaobot

import (
    "fmt"
    "math/rand"
    "net/http"
    "os"
    "path/filepath"
    "sync"
    "time"
)

func (mt *MiBot) DoTTS(value string, waitForFinish bool) {
    if !mt.config.UseCommand {
        if _, err := mt.minaService.TextToSpeech(mt.deviceID, value); err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    } else {
        t := mt.ttsCommand()
        mt.miioService.MiotAction(mt.deviceID, actionId(t), []interface{}{value})
    }
    if waitForFinish {
        elapse := calculateTtsElapse(value)
        time.Sleep(time.Duration(elapse)*time.Second + 1*time.Second)
        mt.WaitForTTSFinish()
    }
}

func (mt *MiBot) WaitForTTSFinish() {
    for {
        isPlaying, _ := mt.getIfXiaoAiIsPlaying()
        if !isPlaying {
            break
        }
        time.Sleep(1 * time.Second)
    }
}

func (mt *MiBot) StartHTTPServer() {
    // Set the port range
    portRange := make([]int, 40)
    for i := 8050; i < 8090; i++ {
        portRange[i-8050] = i
    }

    // Get a random port from the range
    mt.port = rand.Intn(len(portRange))

    mt.tempDir, _ = os.MkdirTemp("", "xiaogpt-tts-")

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, filepath.Join(mt.tempDir, filepath.Base(r.URL.Path)))
    })

    go func() {
        err := http.ListenAndServe(fmt.Sprintf(":%d", mt.port), nil)
        if err != nil {
            fmt.Printf("Error starting HTTP server: %v\n", err)
        }
    }()

    // Implement the GetHostname function
    mt.hostname = getHostname()

    fmt.Printf("Serving on %s:%d\n", mt.hostname, mt.port)
}

// TODO implement
func (mt *MiBot) Text2MP3(text string, ttsLang string) (string, float64, error) {
    // Implement the EdgeTTS Communicate function and Stream method
    return "", 0, fmt.Errorf("not implemented")
}

func (mt *MiBot) EdgeTTS(textStream string, ttsLang string) {
    runTTS := func(textStream string, ttsLang string, queue chan<- interface{}) {

        url, duration, err := mt.Text2MP3(textStream, ttsLang)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
        queue <- struct {
            URL      string
            Duration float64
        }{URL: url, Duration: duration}

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
            if _, err := mt.minaService.PlayByUrl(mt.deviceID, data.URL); err != nil {
                fmt.Printf("Error: %v\n", err)
            }
            time.Sleep(time.Duration(data.Duration) * time.Second)
            mt.WaitForTTSFinish()
        }
    }()

    wg.Wait()
}
