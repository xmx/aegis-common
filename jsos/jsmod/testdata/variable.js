import time from 'time'
import url from 'net/url'
import http from 'net/http'
import httputil from 'net/http/httputil'

const target = url.parse('https://mirrors.zju.edu.cn/')
const proxy = httputil.newSingleHostReverseProxy(target)

const mux = http.newServeMux()
mux.handleFunc('/', (w, r) => {
    w.header().set('Content-Type', 'text/html')
    w.write('<h1>HELLO</h1>')
})

const opt = {
    addr: '0.0.0.0:8888',
    // handler: proxy,
    handler: mux,
    readTimeout: time.minute,
    readHeaderTimeout: 5 * time.second,
}

const handle = http.serve(opt)
handle.wait()
