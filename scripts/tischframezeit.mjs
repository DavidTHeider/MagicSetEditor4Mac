// **Was die Grafikkarte kostet** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Wozu es dieses Skript gibt
//
// **`tischmass.mjs` misst die Rechnung, dieses hier die Rasterung.** #1460
// verlangt für die Leistungsszene „60 FPS als Ziel mit benanntem Mac/macOS,
// Auflösung, Dauer, Warm-/Kaltcache und Framezeitverteilung"; die Rechnung
// allein beantwortet davon nichts.
//
// # Warum ein eigener Vite und keine Seite im Baum
//
// **Weil eine Messseite in `frontend/src` Produktivcode wäre** — gemessen,
// gelintet, und ohne Aufrufer in der Anwendung (§2). Sie wird deshalb nach
// `$TMPDIR` geschrieben; ein Alias zeigt auf `frontend/src`, damit gemessen
// wird, was die Anwendung wirklich tut, und nicht eine Kopie davon.
//
// # Und warum Chrome und nicht die Browservorschau
//
// **Weil die den WebGL-Kontext verliert** — gemessen am 2026-09-16, PixiJS
// meldete „Could not retrieve shader source". Was dort herauskäme, wäre eine
// Zahl über den Ausfall.
//
// **Der Renderer wird ausgelesen und genannt.** Headless Chrome nimmt je nach
// Fassung SwiftShader (Software) oder ANGLE über Metal; eine Framezeit ohne
// den Namen des Renderers ist keine Aussage über diese Maschine (§10).
//
// Aufruf:
//
//	node scripts/tischframezeit.mjs
import { execFileSync, spawn } from 'node:child_process'
import { mkdtemp, readFile, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { verbinden, warte, werte } from './abzuegecdp.mjs'

const wurzel = new URL('..', import.meta.url).pathname
const bilder = Number(process.env.TISCHBILDER ?? 240)

// **Was während der Messung geschieht** — PLAY-06.AK-02
// (#1714):
// „mit aktivem Drag, Animation, Warm-/Kaltcache".
//
//   - `ruhe`: der Tisch steht, jedes Bild zeichnet denselben Plan.
//   - `zug`: eine Karte folgt einem Zeiger — jedes Bild bekommt einen neuen
//     Plan aus `mitZug`, **derselbe Weg wie beim Ziehen am Tisch**
//     (`Tischflaeche.tsx`). Nachgestellt ist nur der Zeiger; was gezeichnet
//     wird, ist das, was die Anwendung zeichnet.
//   - `animation`: eine Karte springt alle zwanzig Bilder zwischen zwei
//     Orten, und der Maler lässt sie fliegen (`bewegt: true`, 220 ms).
//     **Gezeichnet wird nur beim Sprung**; dazwischen treibt der Ticker den
//     Flug allein, bis er landet — so wie am Tisch, wo ein neuer Plan einen
//     Flug auslöst und nicht jedes Bild. Bis zum 2026-09-25 rief diese Seite
//     `zeichne` in **jedem** Bild, und jeder Ruf setzte den laufenden Flug
//     neu an; gemessen war eine Zeichenlast, keine vollständige Bewegung
//     (Befund 4 des Zweitblicks über Phase 6.9).
//
// **Vier Schalter für die Gegenproben** (`docs/TISCHLEISTUNG.md`):
// `TISCHSPRUNG=n` springt alle n Bilder, `TISCHGLEICH=1` zeichnet dabei
// denselben Plan, `TISCHBEWEGT=0` ohne Flug, `TISCHALT=1` in jedem Bild.
const art = process.env.TISCHART ?? 'ruhe'
const sprung = Number(process.env.TISCHSPRUNG ?? 20)
const gleich = process.env.TISCHGLEICH === '1'
const ohneFlug = process.env.TISCHBEWEGT === '0'
const zeichenlast = process.env.TISCHALT === '1'
if (!['ruhe', 'zug', 'animation'].includes(art)) {
  console.error(`TISCHART kennt ruhe, zug und animation — nicht ${art}`)
  process.exit(2)
}
const breite = Number(process.env.TISCHBREITE ?? 1440)
const hoehe = Number(process.env.TISCHHOEHE ?? 900)

const ort = await mkdtemp(join(tmpdir(), 'tischframezeit-'))

// **Die Seite misst selbst.** Sie kann es besser als CDP: `requestAnimationFrame`
// gibt ihr die Zeit zwischen zwei Bildern, und eine Messung von außen läge um
// die Runde über den Draht daneben.
await writeFile(
  join(ort, 'index.html'),
  `<!doctype html><meta charset="utf-8">
<style>html,body{margin:0;height:100%;background:#d4bb94}canvas{display:block}</style>
<div id="tisch" style="width:100%;height:100%"></div>
<script type="module" src="/messung.ts"></script>`,
)

await writeFile(
  join(ort, 'messung.ts'),
  `import { szene } from '@mse/test/tischszenen'
import { mitZug, tischplan } from '@mse/tisch/tischplan'
import { eingepasst } from '@mse/tisch/kamera'
import { tischmaler } from '@mse/tisch/pixinaht'

const leinwand = document.createElement('canvas')
document.getElementById('tisch')!.append(leinwand)

const plan = tischplan(szene('extreme-board-state'))
const malerei = await tischmaler(leinwand)

if (malerei.art === 'ohne') {
  ;(window as never as Record<string, unknown>).messung = { fehler: malerei.grund }
} else {
  const maler = malerei.maler
  const kamera = eingepasst(plan, maler.masse(), 16)
  const zeiten: number[] = []
  const arbeit: number[] = []
  let vorher = performance.now()

  const bild = () => {
    const anfang = performance.now()
    const nummer = zeiten.length
    const gezogen = plan.lagen[Math.floor(plan.lagen.length / 2)]
    const diesmal =
      ${JSON.stringify(art)} === 'zug' && gezogen
        ? mitZug(plan, { kennung: gezogen.kennung, dx: 180 * Math.sin(nummer / 15), dy: 90 * Math.cos(nummer / 15) })
        : ${JSON.stringify(art)} === 'animation' && ${JSON.stringify(!gleich)} && gezogen && Math.floor(nummer / ${String(sprung)}) % 2 === 1
          ? mitZug(plan, { kennung: gezogen.kennung, dx: 240, dy: -120 })
          : plan
    const animation = ${JSON.stringify(art)} === 'animation'
    const gezeichnet = !animation || nummer % ${String(sprung)} === 0 || ${JSON.stringify(zeichenlast)}
    if (gezeichnet) maler.zeichne({ plan: diesmal, kamera, bild: () => undefined, bewegt: animation && ${JSON.stringify(!ohneFlug)} })
    const jetzt = performance.now()
    if (gezeichnet) arbeit.push(jetzt - anfang)
    zeiten.push(jetzt - vorher)
    vorher = jetzt

    if (zeiten.length < ${String(bilder)}) requestAnimationFrame(bild)
    else {
      const gl = leinwand.getContext('webgl2') ?? leinwand.getContext('webgl')
      const info = gl?.getExtension('WEBGL_debug_renderer_info')
      ;(window as never as Record<string, unknown>).messung = {
        zeiten,
        arbeit,
        renderer: info && gl ? String(gl.getParameter(info.UNMASKED_RENDERER_WEBGL)) : 'unbekannt',
        masse: maler.masse(),
      }
    }
  }

  requestAnimationFrame(bild)
}`,
)

await writeFile(
  join(ort, 'vite.config.mjs'),
  // **Ohne `defineConfig`.** Die Datei liegt außerhalb von `node_modules`
  // und kann `vite` nicht auflösen; ein nacktes Objekt braucht es nicht.
  `export default {
  root: ${JSON.stringify(ort)},
  resolve: { alias: { '@mse': ${JSON.stringify(join(wurzel, 'frontend/src'))} } },
  server: { port: 0, strictPort: false },
  cacheDir: ${JSON.stringify(join(ort, '.vite'))},
}`,
)

const vite = spawn(
  join(wurzel, 'frontend/node_modules/.bin/vite'),
  ['--config', join(ort, 'vite.config.mjs')],
  { cwd: join(wurzel, 'frontend'), stdio: ['ignore', 'pipe', 'pipe'] },
)

const adresse = await new Promise((fertig, scheitern) => {
  const frist = setTimeout(() => {
    scheitern(new Error('Vite kam nicht hoch'))
  }, 30000)

  vite.stdout.on('data', (teil) => {
    const treffer = /http:\/\/localhost:(\d+)/.exec(String(teil))
    if (treffer) {
      clearTimeout(frist)
      fertig(`http://localhost:${treffer[1]}/`)
    }
  })

  // **Was Vite klagt, kommt hierher.** Ohne diese Zeile endete ein
  // gescheiterter Start in „kam nicht hoch" und sonst nichts — genau die
  // Gestalt, die §4 für jede Stufe verbietet.
  vite.stderr.on('data', (teil) => {
    process.stderr.write(String(teil))
  })
})

// **Die Messseite ruft den Maler mit einem Objekt** — `Zeichenwunsch` aus
// `tisch/maler.ts`. Dieses Skript läuft nicht in der Pipeline; ein veralteter
// Aufruf fällt hier niemandem auf.

// **Und in WebKit, wenn man es verlangt** (`TISCHBROWSER=webkit`, #1828):
// dieselbe Messseite in der WebKit-Probe — ein WKWebView, eingerichtet wie in
// Wails. PLAY-06.AK-01 verlangt die native WebView **getrennt** von der
// Browservorschau; Chrome misst Blink und ANGLE, die Anwendung läuft auf
// WebKit. **`TISCHSICHTBAR=1`** legt das Fenster auf den Bildschirm; ohne
// liegt es daneben, und WebKit zeichnet trotzdem (Verdeckung abgeschaltet).
const webkit = process.env.TISCHBROWSER === 'webkit'

let ergebnis
if (webkit) {
  const drehbuch = join(ort, 'framezeit.probe')
  const ausgabe = join(ort, 'webkit')
  await writeFile(
    drehbuch,
    // **Die Frist wächst mit der Zahl der Bilder** — eine lange Sitzung von
    // drei Minuten liefe sonst an zwei Minuten Frist ab.
    `motor wk2\nlade ${adresse}\nfrist ${String(Math.max(120000, Math.ceil((bilder / 50) * 1000)))}\nbis window.messung\nsichere messung window.messung\n`,
  )

  // **Der Speicher wird von außen gelesen** — PLAY-06.AK-03. WebKit gibt der
  // Seite kein `performance.memory`; was ein Tisch belegt, steht im
  // Webprozess und im GPU-Prozess von WebKit, und beide zählt `ps`. **Jede
  // Sekunde**, damit eine lange Sitzung zeigt, ob der Speicher wächst.
  //
  // **Nur die Prozesse dieser Messung.** `ps` zeigt alle WebKit-Prozesse des
  // Rechners — auch die fremder Programme, die WebKit nutzen. Gezählt wird
  // deshalb nur, wer **nach** dem Start dieses Skripts entstand.
  const prozesse = () =>
    execFileSync('ps', ['-Ao', 'pid=,rss=,comm='])
      .toString()
      .split('\n')
      .map((z) => z.trim().split(/\s+/))
      .filter((teile) => teile.length >= 3)
      .map(([pid, rss, ...name]) => ({ pid, rss: Number(rss), name: name.join(' ') }))
  const vorher = new Set(prozesse().map((p) => p.pid))
  const speicher = []
  const lesen = () => {
    try {
      const neue = prozesse().filter((p) => !vorher.has(p.pid))
      const summe = (name) =>
        neue.filter((p) => p.name.includes(name)).reduce((alle, p) => alle + p.rss, 0)
      speicher.push({ web: summe('WebKit.WebContent'), gpu: summe('WebKit.GPU') })
    } catch {
      // Ein verpasster Blick ist keine Messung — die nächste Sekunde zählt.
    }
  }
  const takt = setInterval(lesen, 1000)

  try {
    await new Promise((fertig, scheitern) => {
      const probe = spawn(
        join(wurzel, 'scripts/webkitprobe.sh'),
        [drehbuch, '--ausgabe', ausgabe, '--breite', String(breite), '--hoehe', String(hoehe)].concat(
          process.env.TISCHSICHTBAR ? ['--sichtbar'] : [],
        ),
        { stdio: ['ignore', 'ignore', 'inherit'] },
      )
      probe.on('exit', (code) => {
        if (code === 0) fertig(undefined)
        else scheitern(new Error(`webkitprobe endete mit ${String(code)}`))
      })
    })
    ergebnis = JSON.parse(await readFile(join(ausgabe, 'messung.json'), 'utf8'))
    ergebnis.speicher = speicher
  } finally {
    clearInterval(takt)
    vite.kill()
  }
} else {
  const profil = await mkdtemp(join(tmpdir(), 'tischframezeit-chrome-'))
  const chrome = spawn(
    '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    [
      '--headless=new',
      '--remote-debugging-port=0',
      `--user-data-dir=${profil}`,
      `--window-size=${String(breite)},${String(hoehe)}`,
      '--hide-scrollbars',

      // **Die Grafikkarte wird ausdrücklich zugelassen.** Ohne diese Schalter
      // nimmt headless Chrome SwiftShader — und misst dann die CPU noch einmal.
      '--enable-gpu',
      '--use-angle=metal',
      '--ignore-gpu-blocklist',
      'about:blank',
    ],
    { stdio: ['ignore', 'ignore', 'pipe'] },
  )

  try {
    const ws = await verbinden(profil)
    await werte(ws, `location.href = ${JSON.stringify(adresse)}`)

    for (let versuch = 0; versuch < 60; versuch += 1) {
      await warte(1000)
      const roh = await werte(ws, 'JSON.stringify(window.messung ?? null)')
      if (roh && roh !== 'null') {
        ergebnis = JSON.parse(roh)
        break
      }
    }
  } finally {
    chrome.kill()
    vite.kill()
  }
}

if (!ergebnis) {
  console.error('Keine Messung — die Seite hat nichts abgelegt.')
  process.exit(1)
}

if (ergebnis.fehler) {
  console.error(`Kein Renderer: ${String(ergebnis.fehler)}`)
  process.exit(1)
}

// **Das erste Bild zählt gesondert.** Es bezahlt den Aufbau der Szene — jedes
// Sprite, jeden Shader — und ist die Antwort auf „Kaltcache", nicht ein
// Ausreißer, den man wegwerfen darf.
const [kalt, ...warm] = ergebnis.zeiten
const [kaltarbeit, ...warmarbeit] = ergebnis.arbeit

const verteilung = (werte) => {
  const sortiert = [...werte].sort((a, b) => a - b)

  return (anteil) =>
    sortiert[Math.min(sortiert.length - 1, Math.floor(sortiert.length * anteil))]
}

const perzentil = verteilung(warm)
const arbeitP = verteilung(warmarbeit)

const maschine = () => {
  const modell = execFileSync('sysctl', ['-n', 'hw.model']).toString().trim()
  const os = execFileSync('sw_vers', ['-productVersion']).toString().trim()

  return `${modell}, macOS ${os}`
}

// **Die Zahl, nicht nur der Anteil**: ein Bild von 36 ms unter 10.799 stand als
// „0.0 %" da und wurde als „kein Bild" gelesen (Nachlauf über Phase 6.9).
const ueber = (grenze, n = warm.filter((eine) => eine > grenze).length) =>
  `${String(n)} Bilder, ${((n / warm.length) * 100).toFixed(2)} %`

console.log(`Maschine:  ${maschine()}`)
console.log(`Browser:   ${webkit ? 'WebKit-Probe — WKWebView wie in Wails' : 'Chrome headless, ANGLE über Metal'}`)
console.log(`Renderer:  ${String(ergebnis.renderer)}`)
console.log(`Fläche:    ${String(ergebnis.masse.breite)} × ${String(ergebnis.masse.hoehe)} Punkte`)
console.log(`Szene:     extreme-board-state, ${art === 'zug' ? 'mit aktivem Ziehen' : art === 'animation' ? 'mit Animation' : 'ruhend'}`)
console.log(`Bilder:    ${String(warm.length)} nach dem ersten\n`)
// **Zwei Zahlen und nicht eine.** Der *Bildabstand* sagt, ob 60 Hz gehalten
// werden — bei VSync liegt er immer nahe 16,67 ms, gleich wie wenig Arbeit
// anfällt. Die *Zeichenzeit* sagt, wie viel Luft bleibt; ohne sie wüsste man
// nicht, ob der Tisch mit Mühe oder mit Leichtigkeit durchkommt.
console.log('Bildabstand (hält es 60 Hz?)')
console.log(`  erstes Bild, kalt:     ${kalt.toFixed(2)} ms`)
console.log(`  Median:                ${perzentil(0.5).toFixed(2)} ms`)
console.log(`  p95:                   ${perzentil(0.95).toFixed(2)} ms`)
console.log(`  p99:                   ${perzentil(0.99).toFixed(2)} ms`)
console.log(`  Maximum:               ${Math.max(...warm).toFixed(2)} ms`)
console.log(`  über 33,33 ms (30 Hz): ${ueber(33.33)}`)

// **Bei der Animation zählt die Zeichenzeit nur die Sprünge**; den Flug
// rendert der Ticker, und der steht im Bildabstand.
console.log(
  art === 'animation' && !zeichenlast
    ? `\nZeichenzeit je Sprung (${String(warmarbeit.length)} Sprünge nach dem ersten)`
    : '\nZeichenzeit (wie viel Luft bleibt?)',
)
console.log(`  erstes Bild, kalt:     ${kaltarbeit.toFixed(2)} ms`)
console.log(`  Median:                ${arbeitP(0.5).toFixed(2)} ms`)
console.log(`  p95:                   ${arbeitP(0.95).toFixed(2)} ms`)
console.log(`  p99:                   ${arbeitP(0.99).toFixed(2)} ms`)
console.log(`  Maximum:               ${Math.max(...warmarbeit).toFixed(2)} ms`)
console.log(
  `  Anteil eines Bildes:   ${((arbeitP(0.5) / 16.667) * 100).toFixed(1)} % im Median, ` +
    `${((Math.max(...warmarbeit) / 16.667) * 100).toFixed(1)} % im schlechtesten`,
)

// **Eine lange Sitzung zeigt sich in Dritteln** — PLAY-06.AK-03. Wird das
// letzte Drittel langsamer als das erste, wächst etwas mit der Zeit; eine
// Verteilung über alles verdeckte das.
if (warm.length >= 600) {
  const drittel = Math.floor(warm.length / 3)
  console.log('\nBildabstand je Drittel (wächst etwas mit der Zeit?)')
  for (const [name, teil] of [
    ['erstes', warm.slice(0, drittel)],
    ['zweites', warm.slice(drittel, 2 * drittel)],
    ['letztes', warm.slice(2 * drittel)],
  ]) {
    const p = verteilung(teil)
    console.log(`  ${name.padEnd(8)} Median ${p(0.5).toFixed(2)} ms, p99 ${p(0.99).toFixed(2)} ms`)
  }
}

if (ergebnis.speicher && ergebnis.speicher.length > 0) {
  const mb = (kb) => (kb / 1024).toFixed(0)
  const s = ergebnis.speicher
  const hoechst = (feld) => Math.max(...s.map((eine) => eine[feld]))
  console.log('\nSpeicher der WebKit-Prozesse (RSS, jede Sekunde gelesen)')
  console.log(`  Webprozess:  Anfang ${mb(s[0].web)} MB, Ende ${mb(s.at(-1).web)} MB, höchstens ${mb(hoechst('web'))} MB`)
  console.log(`  GPU-Prozess: Anfang ${mb(s[0].gpu)} MB, Ende ${mb(s.at(-1).gpu)} MB, höchstens ${mb(hoechst('gpu'))} MB`)
  console.log(`  ${String(s.length)} Blicke`)

  // **Der Verlauf**: ob der Speicher sich einpendelt, sagen die Enden nicht.
  if (s.length >= 60) {
    const schritt = Math.floor(s.length / 6)
    const punkte = s.filter((_, stelle) => stelle % schritt === 0 || stelle === s.length - 1)
    console.log(`  Webprozess im Verlauf: ${punkte.map((eine) => mb(eine.web)).join(' → ')} MB`)
  }
}
