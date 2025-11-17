import console from 'console'
import time from 'time'
import url from 'net/url'
import http from 'net/http'
import httputil from 'net/http/httputil'

const target = url.parse('https://mirrors.zju.edu.cn/')
const proxy = httputil.newSingleHostReverseProxy(target)

let cnt = 0
const mux = http.newServeMux()
mux.handleFunc('/', (w, r) => {
    cnt++
    w.header().set('Content-Type', 'text/html; charset=utf8')
    const content = `你访问了<strong>${r.url.path}</strong> ，网站总访问量 ${cnt}`
    w.write(content)

    const log = `[${new Date().toJSON()}] ${r.remoteAddr} 第 ${cnt} 次访问：${r.url.path} ${r.url.rawQuery}`
    console.log(log)
})

const addr = '0.0.0.0:8888'
const opt = {
    addr: addr,
    handler: mux,
    readTimeout: time.minute,
    readHeaderTimeout: 5 * time.second
}
const srv = http.createServer(opt)
const h = srv.listen(() => console.log(`HTTP 服务监听在 ${addr}`))
h.wait()
