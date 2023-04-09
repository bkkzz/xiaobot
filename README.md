# xiaobot

Play ChatGPT with Xiaomi AI Speaker

fork from [xiaogpt](https://github.com/yihong0618/xiaogpt) and convert to Go

## 支持的 AI 类型

- ChatGPT API
- NewBing WIP
## 准备

1. ChatGPT Key
2. 小爱音响
3. 能正常联网的环境或 proxy

## 使用

- 跑起来之后就可以问小爱同学问题了，“帮我"开头的问题，会发送一份给 ChatGPT 然后小爱同学用 tts 回答
- 默认用目前 ubus, 如果你的设备不支持 ubus 可以使用配置 `use_command` 来使用 command 来 tts
- 配置 `mute_xiaoai` 选项，可以快速停掉小爱的回答
- 
- 如果有能力可以自行替换唤醒词，也可以去掉唤醒词
- 使用 gpt-3 的 api 那样可以更流畅的对话，速度快, 请 google 如何用 [openai api](https://platform.openai.com/account/api-keys) 命令 --use_gpt3

## config.toml
如果想通过单一配置文件启动也是可以的, 可以通过 `-c` 参数指定配置文件, config 文件必须是合法的 JSON 或者toml 格式
参数优先级
- 环境变量 > config

```shell
xiaobot -c config.toml
```
## 配置项说明

| 参数                  | 说明                                     | 默认值                              |
| --------------------- |----------------------------------------| ----------------------------------- |
| hardware              | 设备型号                                   |                                     |
| account               | 小爱账户                                   |                                     |
| password              | 小爱账户密码                                 |                                     |
| openai_key            | openai的apikey                          |                                     |
| cookie                | 小爱账户cookie （如果用上面密码登录可以不填）             |                                     |
| mi_did                | 设备did                                  |                                     |
| use_command           | 使用 MI command 与小爱交互                    | `false`                             |
| mute_xiaoai           | 快速停掉小爱自己的回答                            | `true`                              |
| verbose               | 是否打印详细日志                               | `false`                             |
| bot                   | 使用的 bot 类型，目前支持gpt3,chatgptapi和newbing | `chatgptapi`                        |
| enable_edge_tts       | 使用Edge TTS        WIP                  | `false`                             |
| edge_tts_voice        | Edge TTS 的嗓音        WIP                | `zh-CN-XiaoxiaoNeural`              |
| prompt                | 自定义prompt                              | `请用100字以内回答`                 |
| keyword               | 自定义请求词列表                               | `["请问"]`                          |
| change_prompt_keyword | 更改提示词触发列表                              | `["更改提示词"]`                    |
| start_conversation    | 开始持续对话关键词                              | `开始持续对话`                      |
| end_conversation      | 结束持续对话关键词                              | `结束持续对话`                      |
| stream                | 使用流式响应，获得更快的响应 WIP                     | `false`                             |
| proxy                 | 支持 HTTP 代理，传入 http proxy URL           | ""                                  |
| gpt_options           | OpenAI API 的参数字典                       | `{}`                                |
| bing_cookie_path      | NewBing使用的cookie路径，参考[这里]获取            | 也可通过环境变量 `COOKIE_FILE` 设置 |
| bing_cookies          | NewBing使用的cookie字典，参考[这里]获取            |                                     |


## 注意

1. 请开启小爱同学的蓝牙
2. 如果要更改提示词和 PROMPT 在代码最上面自行更改
3. 目前已知 LX04 和 L05B L05C 可能需要使用 `use_command`

## TODO
1. 支持流式响应
2. 支持Newbing
3. UI 优化

## 感谢

- [xiaomi](https://www.mi.com/)
- [openai](https://openai.com/)
- [yihong0618](https://github.com/yihong0618)

