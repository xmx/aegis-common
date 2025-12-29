### Usage

```js
import cron from 'github.com/robfig/cron/v3'
import console from 'console'
import context from 'context'
import time from 'time'

const parser = cron.newParser(cron.minute | cron.hour | cron.dom | cron.month | cron.dow)
const ctab = cron.new(cron.withSeconds(), cron.withParser(parser))

let cnt = 0

ctab.addJob('* * * * *', () => {
  cnt++
  console.log(`ADD ${cnt}`)
})
ctab.schedule(cron.every(time.second), () => {
  cnt += 3
  console.log(`EVERY ${cnt}`)
})

ctab.run() // blocking
```

## const

- `second`
- `minute`
- `hour`
- `dom`
- `month`
- `dow`
- `descriptor`

## func

- `new`

- `descriptor`
