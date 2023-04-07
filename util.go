package xiaogpt

import (
    "errors"
    "net"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "os"
    "regexp"
    "strings"
)

// parseCookieString returns an http.CookieJar from a string containing cookies
func parseCookieString(cookieString string) (http.CookieJar, error) {
    cookieMap := make(map[string]string)
    headers := http.Header{}
    headers.Add("Cookie", cookieString)

    cookies := headers["Cookie"]
    for _, cookie := range cookies {
        parts := strings.Split(cookie, ";")
        for _, part := range parts {
            nameValue := strings.SplitN(strings.TrimSpace(part), "=", 2)
            if len(nameValue) == 2 {
                cookieMap[nameValue[0]] = nameValue[1]
            }
        }
    }

    cookieJar, err := cookiejar.New(nil)
    if err != nil {
        return nil, err
    }

    u, _ := url.Parse("http://dummy.url")
    for key, value := range cookieMap {
        cookieJar.SetCookies(u, []*http.Cookie{
            {
                Name:  key,
                Value: value,
            },
        })
    }

    return cookieJar, nil
}

var noElapseChars = regexp.MustCompile(`([「」『』《》“”'\"()（）]|(?<!-)-(?!-))`)

// calculateTtsElapse returns the elapsed time for TTS
func calculateTtsElapse(text string) float64 {
    speed := 4.5
    cleanText := noElapseChars.ReplaceAllString(text, "")
    return float64(len(cleanText)) / speed
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

func InArray(needle string, haystack []string) bool {
    for _, item := range haystack {
        if item == needle {
            return true
        }
    }
    return false
}
