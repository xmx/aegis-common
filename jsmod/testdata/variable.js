import console from 'console'
import config from 'aegis/config'

console.log("开始初始化配置")

const cfg = {
    addr: 'xxxxxxx',
}

config.set(cfg)

console.log("配置初始化完毕")
