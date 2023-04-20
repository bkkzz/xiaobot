package main

import (
	"encoding/json"
	"github.com/longbai/xiaobot"
	"log"
	"net/http"
)

var submit = false

func config(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost || request.Body == nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	if submit {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	var config xiaobot.Config
	err := json.NewDecoder(request.Body).Decode(&config)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	// 简单的验证
	if config.Account == "" {
		http.Error(writer, "missing xiaomi account field", http.StatusBadRequest)
		return
	}
	log.Printf("Received config: %+v\n", config)
	bot := xiaobot.NewMiBot(&config)
	go bot.Run()
	submit = true
	writer.WriteHeader(http.StatusOK)
}

func Router() *http.ServeMux {
	h := http.NewServeMux()
	h.HandleFunc("/submit-config", config)

	h.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(index))
	})
	return h
}

const index = `<!DOCTYPE html>
<html>
<head>
  <title>配置页</title>
<style>
    body {
      font-family: Arial, Helvetica, sans-serif;
    }

    form {
      display: flex;
      flex-direction: column;
      max-width: 400px;
    }

    label {
      margin-bottom: 10px;
    }

    input[type="text"],
    input[type="password"],
    textarea,
    input[type="checkbox"] {
      margin-bottom: 20px;
      padding: 10px;
      border-radius: 5px;
      border: 1px solid #ccc;
    }

    label.checkbox {
      display: inline-block;
      margin-right: 20px;
    }

    input[type="submit"] {
      padding: 10px 20px;
      background-color: #007bff;
      color: #fff;
      border: none;
      border-radius: 5px;
      cursor: pointer;
    }

    input[type="submit"]:hover {
      background-color: #0069d9;
    }

    #error-message {
      color: red;
      margin-top: 10px;
    }

    .row {
      display: flex;
      flex-wrap: wrap;
      margin-bottom: 20px;
    }

    .row label {
      flex-basis: 30%;
      margin-right: 10px;
    }

    .row input[type="text"],
    .row textarea,
    .row select {
      flex-basis: 70%;
    }
  </style>
</head>
<body>
  <form id="config-form">
    <div class="row">
      <label for="hardware">硬件类型:</label>
      <input type="text" id="hardware" name="hardware">
    </div>

    <div class="row">
      <label for="account">账号:</label>
      <input type="text" id="account" name="account" required>
    </div>

    <div class="row">
      <label for="password">密码:</label>
      <input type="password" id="password" name="password" required>
    </div>

    <div class="row">
      <label for="openai-key">OpenAI Key:</label>
      <input type="text" id="openai-key" name="openai_key">
    </div>

    <div class="row">
      <label for="openai-backend">OpenAI Backend:</label>
      <input type="text" id="openai-backend" name="openai_backend">
    </div>

    <div class="row">
      <label for="proxy">代理:</label>
      <input type="text" id="proxy" name="proxy">
    </div>

    <div class="row">
      <label for="mi-did">MiDID:</label>
      <input type="text" id="mi-did" name="mi_did">
</div>

<div class="row">
  <label for="keyword">关键词:</label>
  <input type="text" id="keyword" name="keyword">
</div>

<div class="row">
  <label for="change-prompt-keyword">修改提示关键词:</label>
  <input type="text" id="change-prompt-keyword" name="change_prompt_keyword">
</div>

<div class="row">
  <label for="prompt">提示:</label>
  <input type="text" id="prompt" name="prompt">
</div>

<div class="row">
  <label for="mute-xiaoai" class="checkbox">静音小爱:</label>
  <input type="checkbox" id="mute-xiaoai" name="mute_xiaoai">
</div>

<div class="row">
  <label for="bot">机器人:</label>
  <input type="text" id="bot" name="bot">
</div>

<div class="row">
  <label for="api-base">API基础路径:</label>
  <input type="text" id="api-base" name="api_base">
</div>

<div class="row">
  <label for="use-command" class="checkbox">使用命令:</label>
  <input type="checkbox" id="use-command" name="use_command">
</div>

<div class="row">
  <label for="verbose" class="checkbox">详细:</label>
  <input type="checkbox" id="verbose" name="verbose">
</div>

<div class="row">
  <label for="start-conversation">开始对话:</label>
  <input type="text" id="start-conversation" name="start_conversation">
</div>

<div class="row">
  <label for="end-conversation">结束对话:</label>
  <input type="text" id="end-conversation" name="end_conversation">
</div>

<div class="row">
  <label for="stream" class="checkbox">流:</label>
  <input type="checkbox" id="stream" name="stream">
</div>

<div class="row">
  <label for="enable-edge-tts" class="checkbox">启用边缘TTS:</label>
  <input type="checkbox" id="enable-edge-tts" name="enable_edge_tts">
</div>

<div class="row">
  <label for="edge-tts-voice">边缘TTS语音:</label>
  <input type="text" id="edge-tts-voice" name="edge_tts_voice">
</div>

<div class="row">
  <label for="gpt-options">GPT选项:</label>
  <textarea id="gpt-options" name="gpt_options"></textarea>
</div>

<div class="row">
  <label for="token-path">令牌路径:</label>
  <input type="text" id="token-path" name="token_path">
</div>

<input type="submit" value="提交">

</form>
<div id="error-message"></div>
  <script src="https://cdn.jsdelivr.net/npm/umbrellajs@0.8.3/umbrella.min.js"></script>
  <script>
    const form = u('#config-form');
	const errorMessage = u('#error-message');
    form.on('submit', async (event) => {
      event.preventDefault();
// 验证账号和密码
      const account = form.find('#account').val();
      const password = form.find('#password').val();
      if (account.trim() === '' || password.trim() === '') {
        errorMessage.html('账号和密码不能为空');
errorMessage.show();
        return;
      }
      const config = {
        hardware: form.find('#hardware').val(),
        account: form.find('#account').val(),
        password: form.find('#password').val(),
        openai_key: form.find('#openai-key').val(),
        openai_backend: form.find('#openai-backend').val(),
        proxy: form.find('#proxy').val(),
        mi_did: form.find('#mi-did').val(),
        keyword: form.find('#keyword').val(),
        change_prompt_keyword: form.find('#change-prompt-keyword').val(),
        prompt: form.find('#prompt').val(),
        mute_xiaoai: form.find('#mute-xiaoai').is(':checked'),
        bot: form.find('#bot').val(),
        api_base: form.find('#api-base').val(),
        use_command: form.find('#use-command').is(':checked'),
        verbose: form.find('#verbose').is(':checked'),
        start_conversation: form.find('#start-conversation').val(),
        end_conversation: form.find('#end-conversation').val(),
        stream: form.find('#stream').is(':checked'),
        enable_edge_tts: form.find('#enable-edge-tts').is(':checked'),
        edge_tts_voice: form.find('#edge-tts-voice').val(),
        gpt_options: JSON.parse(form.find('#gpt-options').val()),
        token_path: form.find('#token-path').val(),
      };
      const response = await u.post('/submit-config', {json: config});
      if (response.ok) {
        console.log(response.json());
      } else {
        const error = await response.json();
        errorMessage.html(error.message);
errorMessage.show();
      }
    });
  </script>
</body>
</html>
`
