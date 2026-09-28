// 使用 Edge/Chrome DevTools Protocol 驱动无头浏览器：注入登录态并逐页截图
import { spawn } from 'child_process'
import { writeFileSync, mkdirSync } from 'fs'

const EDGE = 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'
const OUT = 'C:/Users/gongz/AppData/Local/Temp/mbti-shots'
const BASE = 'http://127.0.0.1:5310'
mkdirSync(OUT, { recursive: true })

const PORT = 9231
const userDir = 'C:/Users/gongz/AppData/Local/Temp/edge-cdpA'
const edge = spawn(EDGE, [
  '--headless=new',
  `--remote-debugging-port=${PORT}`,
  `--user-data-dir=${userDir}`,
  '--disable-gpu',
  '--no-first-run',
  '--window-size=1280,2400',
  'about:blank',
])

const sleep = (ms) => new Promise(r => setTimeout(r, ms))

async function getWebSocket() {
  for (let i = 0; i < 30; i++) {
    try {
      await fetch(`http://127.0.0.1:${PORT}/json/version`)
      const targets = await fetch(`http://127.0.0.1:${PORT}/json`).then(r => r.json())
      const page = targets.find(t => t.type === 'page')
      if (page) return page
    } catch {}
    await sleep(500)
  }
  throw new Error('CDP 未就绪')
}

// 极简 WebSocket JSON-RPC 客户端（Node 22 原生 WebSocket）
class CDP {
  constructor(wsUrl) { this.ws = new WebSocket(wsUrl); this.id = 0; this.pending = new Map()
    this.ws.onmessage = (e) => {
      const msg = JSON.parse(e.data)
      if (msg.id && this.pending.has(msg.id)) {
        const { resolve, reject } = this.pending.get(msg.id)
        this.pending.delete(msg.id)
        msg.error ? reject(new Error(JSON.stringify(msg.error))) : resolve(msg.result)
      }
    }
  }
  ready() { return new Promise((res, rej) => { this.ws.onopen = res; this.ws.onerror = rej }) }
  send(method, params = {}) {
    const id = ++this.id
    return new Promise((resolve, reject) => {
      this.pending.set(id, { resolve, reject })
      this.ws.send(JSON.stringify({ id, method, params }))
    })
  }
  close() { this.ws.close() }
}

async function shoot(cdp, name, { width = 1280, height = 1700, wait = 1800 } = {}) {
  await cdp.send('Emulation.setDeviceMetricsOverride', { width, height, deviceScaleFactor: 1, mobile: false })
  await sleep(wait)
  const { data } = await cdp.send('Page.captureScreenshot', { format: 'png' })
  writeFileSync(`${OUT}/${name}.png`, Buffer.from(data, 'base64'))
  console.log('📸', name)
}

;(async () => {
  const target = await getWebSocket()
  const cdp = new CDP(target.webSocketDebuggerUrl)
  await cdp.ready()
  await cdp.send('Page.enable')
  await cdp.send('Runtime.enable')

  // 1) 首页（未登录）
  await cdp.send('Page.navigate', { url: `${BASE}/` })
  await shoot(cdp, '01-home', { height: 1500, wait: 2200 })

  // 2) 注入登录态（alice 的 token）
  const login = await fetch('http://127.0.0.1:8088/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'alice', password: 'alice123' }),
  }).then(r => r.json())

  await cdp.send('Runtime.evaluate', {
    expression: `localStorage.setItem('mbti_token', ${JSON.stringify(login.token)})`,
  })

  // 3) 首页（登录后，能看到已有的帖与余额）
  await cdp.send('Page.navigate', { url: `${BASE}/` })
  await shoot(cdp, '02-home-loggedin', { height: 2200, wait: 2200 })

  // 4) 展开第一张帖的双面卡片：点“展开”
  await cdp.send('Runtime.evaluate', { expression: `
    [...document.querySelectorAll('.caret')][0]?.click();
  `})
  await sleep(1200)
  await shoot(cdp, '03-card-front', { height: 2400 })
  // 翻到背面
  await cdp.send('Runtime.evaluate', { expression: `
    [...document.querySelectorAll('.flip-tabs button')][1]?.click();
  `})
  await sleep(1000)
  await shoot(cdp, '04-card-back', { height: 1900 })

  // 5) 测评答题页
  await cdp.send('Page.navigate', { url: `${BASE}/quiz` })
  await shoot(cdp, '05-quiz', { height: 1100, wait: 2000 })

  // 6) 个人中心
  await cdp.send('Page.navigate', { url: `${BASE}/me` })
  await shoot(cdp, '06-profile', { height: 2100, wait: 2200 })
  // 切到流水 tab
  await cdp.send('Runtime.evaluate', { expression: `
    [...document.querySelectorAll('.rec-tabs button')][2]?.click();
  `})
  await sleep(900)
  await shoot(cdp, '07-profile-ledger', { height: 1800, wait: 600 })

  console.log('完成，输出目录:', OUT)
  cdp.close()
  edge.kill()
  process.exit(0)
})().catch(e => { console.error(e); edge.kill(); process.exit(1) })
