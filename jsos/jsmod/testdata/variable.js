import console from 'console'
import crontab from 'crontab'
import runtime from 'runtime'
import os from 'os'

const handle = crontab.addJob('*/3 * * * * *', function () {
    console.log(`${os.getpid()} - ${runtime.numGoroutine()}`)
})

console.log(handle.id())
handle.wait()
