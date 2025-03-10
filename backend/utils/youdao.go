package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Cibifang/online-reader/backend/config"
)

// YoudaoTranslateResponse represents the response from Youdao Translate API
type YoudaoTranslateResponse struct {
	ErrorCode   string   `json:"errorCode"`
	Query       string   `json:"query"`
	Translation []string `json:"translation"`
	Basic       struct {
		Explains []string `json:"explains"`
	} `json:"basic"`
	Web []struct {
		Key   string   `json:"key"`
		Value []string `json:"value"`
	} `json:"web"`
}

// TranslateWithYoudao translates a word using Youdao API
func TranslateWithYoudao(word string) (string, error) {
	if config.AppConfig.YoudaoAppKey == "your-app-key" || config.AppConfig.YoudaoAppSecret == "your-app-secret" {
		return "请先设置有道翻译API密钥", nil
	}

	salt := fmt.Sprintf("%d", rand.Intn(10000))
	curtime := fmt.Sprintf("%d", time.Now().Unix())

	signStr := config.AppConfig.YoudaoAppKey + truncate(word) + salt + curtime + config.AppConfig.YoudaoAppSecret
	sign := md5Sum(signStr)

	apiURL := "https://openapi.youdao.com/api"
	data := url.Values{}
	data.Set("q", word)
	data.Set("from", "en")
	data.Set("to", "zh-CHS")
	data.Set("appKey", config.AppConfig.YoudaoAppKey)
	data.Set("salt", salt)
	data.Set("sign", sign)
	data.Set("signType", "v3")
	data.Set("curtime", curtime)

	log.Printf("Youdao API Request: URL=%s, Word=%s, AppKey=%s, Salt=%s, CurTime=%s, Sign=%s",
		apiURL, word, config.AppConfig.YoudaoAppKey, salt, curtime, sign)

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return "", fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response body: %v", err)
	}

	log.Printf("Youdao API Raw Response: %s", string(body))

	var result YoudaoTranslateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("JSON parsing error: %v", err)
	}

	if result.ErrorCode != "0" {
		return "", fmt.Errorf("API error: %s", result.ErrorCode)
	}

	var translation strings.Builder

	if len(result.Translation) > 0 {
		translation.WriteString(strings.Join(result.Translation, ", "))
	}

	if result.Basic.Explains != nil && len(result.Basic.Explains) > 0 {
		translation.WriteString("\n解释: ")
		translation.WriteString(strings.Join(result.Basic.Explains, ", "))
	}

	if result.Web != nil && len(result.Web) > 0 {
		translation.WriteString("\n网络释义:\n")
		for _, item := range result.Web {
			translation.WriteString("- ")
			translation.WriteString(item.Key)
			translation.WriteString(": ")
			translation.WriteString(strings.Join(item.Value, ", "))
			translation.WriteString("\n")
		}
	}

	return translation.String(), nil
}

// truncate truncates a string for Youdao API
func truncate(q string) string {
	if len(q) <= 20 {
		return q
	}
	return q[:10] + fmt.Sprintf("%d", len(q)) + q[len(q)-10:]
}

// md5Sum calculates MD5 hash
func md5Sum(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}
