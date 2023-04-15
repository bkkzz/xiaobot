package xiaobot

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/longbai/edgetts"
)

func (mt *MiBot) startEdgeServer() {
	mt.tempDir, _ = os.MkdirTemp("", "xiaogpt-tts-")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(mt.tempDir, filepath.Base(r.URL.Path)))
	})

	mt.hostname = getHostname()
	rd := rand.Uint32() % 50
	mt.port = int(8000 + rd)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", mt.hostname, mt.port), nil)
	if err != nil {
		fmt.Printf("Error starting HTTP server: %v\n", err)
	}

	fmt.Printf("Serving on %s:%d\n", mt.hostname, mt.port)
}

const voiceFormat = "audio-24khz-48kbitrate-mono-mp3"

func (mt *MiBot) textToMp3(text string, ttsLang string) (string, error) {
	ttsEdge := edgetts.EdgeTTS{}
	ssml := edgetts.CreateSSML(text, ttsLang)
	log.Println(ssml)
	b, err := ttsEdge.GetAudio(ssml, voiceFormat)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return "", err
	}

	filename := fmt.Sprintf("%s.mp3", time.Now().Format("20060102150405"))
	fp := filepath.Join(mt.tempDir, filename)
	err = os.WriteFile(fp, b, 0644)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return "", err
	}
	return filename, nil
}

func (mt *MiBot) edgeTTS(text string, ttsLang string) error {
	filename, err := mt.textToMp3(text, ttsLang)
	if err != nil {
		return err
	}
	u := fmt.Sprintf("http://%s:%d/%s", mt.hostname, mt.port, filename)

	fmt.Printf("play: %s\n", u)
	v, err := mt.minaService.PlayByUrl(mt.deviceID, u)
	log.Printf("play url %v, Error: %v\n", v, err)
	if err != nil {
		return err
	}
	time.Sleep(calculateTtsElapse(text))
	mt.waitForTTSFinish()
	return nil
}
