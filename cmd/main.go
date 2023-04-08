package main

import (
    "flag"
    "github.com/longbai/xiaobot"
)

func main() {
    cfgPath := flag.String("c", "config.toml", "config file path")
    flag.Parse()
    config, err := xiaobot.NewConfigFromFile(*cfgPath)
    if err != nil {
        panic(err)
    }
    bot := xiaobot.NewMiBot(config)
    err = bot.Run()
    if err != nil {
        panic(err)
    }
}
