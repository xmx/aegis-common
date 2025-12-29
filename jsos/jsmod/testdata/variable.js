import console from 'console'
import http from 'net/http'

const mux = http.newServeMux()
mux.handleFunc('/ping', (w, r) => {
    console.log(`访问了`)
    w.write(`PONG`)
})

http.listenAndServe(':8800', mux)
