// README screenshots and a browser smoke test, against a demo instance (DEMO_DATA=true seeds seven people and half a year of check-ins):
//
//   DATA_DIR=/tmp/horoscope-demo DEMO_DATA=true ADMIN_EMAIL=demo@example.com ADMIN_PASSWORD=demodemo1 ADMIN_NAME=Mara PORT=18306 ./pocket-horoscope
//   CDP=http://127.0.0.1:19271 BASE=http://127.0.0.1:18306 node scripts/screenshots.mjs   # any Chrome with remote debugging
//   SHOTS=0 … node scripts/screenshots.mjs                                                  # checks only
//   OUT=/tmp/shots … node scripts/screenshots.mjs                                           # somewhere else than docs/screenshots
// playwright-core comes from web/node_modules, or set PW_FROM to a package.json next to an install of it.
import { createRequire } from 'node:module'
import assert from 'node:assert/strict'

const require = createRequire(process.env.PW_FROM || new URL('../web/package.json', import.meta.url))
let pw
try { pw = require('playwright') } catch { pw = require('playwright-core') }
const { chromium } = pw
const base = process.env.BASE || 'http://127.0.0.1:18306'
const shots = process.env.SHOTS !== '0'
const dir = process.env.OUT ? process.env.OUT.replace(/\/?$/, '/') : new URL('../docs/screenshots/', import.meta.url).pathname
const browser = process.env.CDP ? await chromium.connectOverCDP(process.env.CDP) : await chromium.launch()
const errors = []

async function session({ phone = false, dark = false } = {}) {
  const ctx = await browser.newContext({
    viewport: phone ? { width: 390, height: 844 } : { width: 1440, height: 900 },
    deviceScaleFactor: phone ? 3 : 2, isMobile: phone, hasTouch: phone, colorScheme: dark ? 'dark' : 'light',
    userAgent: phone ? 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1' : undefined,
  })
  const page = await ctx.newPage()
  page.on('pageerror', e => errors.push(`pageerror: ${e}`))
  page.on('console', m => { if (m.type() === 'error' && !/Failed to load resource/.test(m.text())) errors.push(m.text()) })
  await page.goto(base + '/login')
  const status = await page.evaluate(() => fetch('/api/v1/auth/status').then(r => r.json()))
  if (status.auth_mode === 'local') {
    const res = await page.evaluate(() => fetch('/api/v1/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email: 'demo@example.com', password: 'demodemo1' }) }).then(r => r.status))
    assert.equal(res, 200, 'demo sign-in')
  }
  return { ctx, page }
}
async function go(page, path, wait = 1500, still = true) {
  await page.goto(base + path, { waitUntil: 'networkidle' })
  // headless Chrome freezes CSS fade-ins: show everything at once
  if (still) await page.addStyleTag({ content: '*, *::before, *::after { transition: none !important; animation: none !important; }' })
  await page.evaluate(() => document.fonts.ready)
  await page.waitForTimeout(wait)
}
async function shot(page, name) { if (shots) await page.screenshot({ path: dir + name + '.png' }) }
async function personId(page, rel) {
  const ps = await page.evaluate(() => fetch('/api/v1/people').then(r => r.json()))
  const p = ps.find(x => x.relation === rel)
  assert.ok(p, `demo person ${rel}`)
  return p.id
}

// ---- desktop ----
{
  const { ctx, page } = await session()
  await go(page, '/')
  assert.ok(await page.locator('h1, h2').first().isVisible(), 'today renders')
  await shot(page, 'desktop-today')
  const me = await personId(page, 'me')
  await go(page, `/people/${me}`, 2000)
  assert.ok(await page.locator('svg').first().isVisible(), 'chart wheel renders')
  await shot(page, 'desktop-chart')
  await go(page, '/people')
  await shot(page, 'desktop-people')
  const partner = await personId(page, 'partner')
  await go(page, `/match?b=${partner}`, 2000)
  await shot(page, 'desktop-match')
  await go(page, '/sky/moon', 2000)
  await shot(page, 'desktop-moon')
  await go(page, '/sky', 2000)
  await shot(page, 'desktop-sky')
  await go(page, '/transits', 2000)
  await shot(page, 'desktop-transits')
  await go(page, '/journal/insights', 1500)
  await shot(page, 'desktop-insights')
  await go(page, '/journal')
  await shot(page, 'desktop-journal')
  await go(page, '/sky/retrogrades', 2000)
  await shot(page, 'desktop-retrogrades')
  await go(page, '/settings/charts')
  await shot(page, 'desktop-settings')
  await ctx.close()
}
// ---- desktop, dark ----
{
  const { ctx, page } = await session({ dark: true })
  await go(page, '/')
  await shot(page, 'desktop-today-dark')
  const me = await personId(page, 'me')
  await go(page, `/people/${me}`, 2000)
  await shot(page, 'desktop-chart-dark')
  await ctx.close()
}
// ---- phone ----
{
  const { ctx, page } = await session({ phone: true })
  await go(page, '/')
  await shot(page, 'phone-today')
  const me = await personId(page, 'me')
  await go(page, `/people/${me}`, 2000)
  await shot(page, 'phone-chart')
  await go(page, '/sky/moon', 2000)
  await shot(page, 'phone-moon')
  await go(page, '/people')
  await shot(page, 'phone-people')
  await go(page, '/journal')
  await shot(page, 'phone-journal')
  await ctx.close()
}
{
  const { ctx, page } = await session({ phone: true, dark: true })
  await go(page, '/')
  await shot(page, 'phone-today-dark')
  await ctx.close()
}

await browser.close()
if (errors.length) { console.error(errors.join('\n')); process.exit(1) }
console.log('ok', shots ? `screenshots in ${dir}` : '(checks only)')
