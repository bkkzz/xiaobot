package xiaobot

import (
    "errors"
    "net"
    "net/url"
    "os"
    "regexp"
    "strings"
    "time"
)

//var noElapseChars = regexp.MustCompile(`([「」『』《》“”'\"()（）]|(?<!-)-(?!-))`)
var regex1 = regexp.MustCompile(`[「」『』《》“”'\"()（）]`)
var regex2 = regexp.MustCompile(`(^|[^-])-($|[^-])`)

// calculateTtsElapse returns the elapsed time for TTS
func calculateTtsElapse(text string) time.Duration {
    speed := 4.5

    // Replace the first part of the regex
    result := regex1.ReplaceAllString(text, "")

    // Replace the second part of the regex
    result = regex2.ReplaceAllString(result, "")

    v := float64(len(result)) / speed
    return time.Duration(v+1) * time.Second
}

var endingPunctuations = []string{"。", "？", "！", "；", ".", "?", "!", ";"}

func splitSentences(textStream <-chan string) <-chan string {
    result := make(chan string)
    go func() {
        cur := ""
        for text := range textStream {
            cur += text
            for _, punc := range endingPunctuations {
                if strings.HasSuffix(cur, punc) {
                    result <- cur
                    cur = ""
                    break
                }
            }
        }
        if cur != "" {
            result <- cur
        }
        close(result)
    }()
    return result
}

func findKeyByPartialString(dictionary map[string]string, partialKey string) string {
    for key, value := range dictionary {
        if strings.Contains(partialKey, key) {
            return value
        }
    }
    return ""
}

func validateProxy(proxyStr string) error {
    parsed, err := url.Parse(proxyStr)
    if err != nil {
        return err
    }
    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        return errors.New("Proxy scheme must be http or https")
    }
    if parsed.Hostname() == "" || parsed.Port() == "" {
        return errors.New("Proxy hostname and port must be set")
    }
    return nil
}

func getHostname() string {
    if hostname, exists := os.LookupEnv("XIAOGPT_HOSTNAME"); exists {
        return hostname
    }

    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return ""
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String()
}

func Normalize(message string) string {
    message = strings.TrimSpace(message)
    message = strings.ReplaceAll(message, " ", "--")
    message = strings.ReplaceAll(message, "\n", "，")
    message = strings.ReplaceAll(message, "\"", "，")
    return message
}
