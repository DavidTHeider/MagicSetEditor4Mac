// **Wie lange die eigene Vorschau beim Blättern braucht** — #2309, Folge aus
// dem Nachlauf über J05 (#2269), Befund 7.
//
// Das Formular ordnet seine Felder nach den Flächen der Vorschau. Beim
// Blättern übernahm es bis hierher die der vorigen Karte, solange beide
// denselben Stil tragen — und der Kanon zeigt, dass sich die Leseordnung
// dabei in 601 von 1.033 Schritten ändert
// (`internal/app/feldordnungkanon_intern_test.go`). Wartet das Formular
// stattdessen auf die eigene Vorschau, steht es so lange unsichtbar da. Wie
// lange das ist, misst dieses Drehbuch: »Black« aus dem Kanon, 30 Schritte
// mit ↓ in der Kartenliste, je Schritt die Zeit vom Tastendruck bis zum
// geladenen Vorschaubild der neuen Karte.
//
// **Was es nicht misst**: wie es aussieht. Ob ein kurzes Unsichtbarsein
// ruhiger wirkt als ein Umsortieren, sieht nur ein Mensch.
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync } from 'node:fs'
import { homedir } from 'node:os'
import { join } from 'node:path'

export const nurEinzeln = true

const pack = process.env.MSE_PROBE_PACK ?? join(homedir(), 'full-magic-pack')
const kanon = process.env.MSE_PROBE_KANON ?? join(homedir(), 'Library/Application Support/MagicSetEditor4Mac/kanonische-sets')
const schritte = 30

const vorschaubild = "document.querySelector('img[alt=\"Vorschau der Karte\"]')"

// messen setzt die Uhr und wartet im Takt der Seite auf ein neues, fertig
// geladenes Vorschaubild. **Dabei schaut es aufs Formular** (#2309): sobald
// das der neuen Karte steht (der Name ist ein anderer), merkt es sich seine
// Ordnung und ob es unsichtbar stand; 100 ms nach dem Bild noch einmal.
const messen = `const namensfeld = () => [...document.querySelectorAll('input')].find((e) => e.getAttribute('aria-label') === 'name')
const ordnung = () => { const f = namensfeld()?.closest('form'); return f ? [...f.querySelectorAll('input, textarea, select, [contenteditable=true]')].map((e) => e.getAttribute('aria-label') ?? '').join('|') : '' }
const vorher = (${vorschaubild})?.src
const vorherName = namensfeld()?.value
window.__blatt = { beginn: performance.now(), dauer: null, zuerst: null, zuletzt: null, unsichtbar: false }
const pruefe = () => {
  const b = window.__blatt
  const feld = namensfeld()
  if (feld && feld.value !== vorherName) {
    if (b.zuerst === null) b.zuerst = ordnung()
    if (feld.closest('[aria-busy=true]')) b.unsichtbar = true
  }
  const i = (${vorschaubild})
  if (i && i.src !== vorher && i.complete && i.naturalWidth > 0) { b.dauer = performance.now() - b.beginn; setTimeout(() => { b.zuletzt = ordnung() }, 100) }
  else if (performance.now() - b.beginn < 10000) requestAnimationFrame(pruefe)
}
requestAnimationFrame(pruefe)
return true`

function kennzahlen(werte) {
  const s = [...werte].sort((a, b) => a - b)
  const q = (anteil) => s[Math.min(s.length - 1, Math.floor(anteil * s.length))]
  const rund = (z) => Math.round(z * 10) / 10

  return { n: s.length, min: rund(s[0]), median: rund(q(0.5)), p95: rund(q(0.95)), max: rund(s.at(-1)) }
}

export default async function blaettern(p) {
  if (!existsSync(join(pack, 'data')) || !existsSync(join(kanon, 'Black.mse-set'))) {
    p.pruefe(`Pack (${pack}) und Black.mse-set (${kanon}) liegen da`, false)

    return
  }

  const ort = p.ort()
  const pakete = join(ort.heim, 'Library/Application Support/MagicSetEditor4Mac/pakete')
  mkdirSync(pakete, { recursive: true })
  execFileSync('cp', ['-cR', `${join(pack, 'data')}/.`, pakete])
  execFileSync('cp', ['-c', join(kanon, 'Black.mse-set'), join(ort.ort, 'Black.mse-set')])
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', join(ort.ort, 'Black.mse-set'))
  const setname = p.mse(ort, 'sets').trim().split('\n')[0].split('\t')[1]

  await p.starte({ ort })
  await p.taste('2', 'befehl')
  await p.wartetext(setname, 30000)
  await p.klickeText(setname)
  await p.warte('die erste Vorschau steht', () => p.seite(`const i = (${vorschaubild}); return !!i && i.complete && i.naturalWidth > 0`), 60000)
  await p.klickeAuf("document.querySelector('[role=grid] [aria-selected=true] [role=gridcell]')", 'die gewählte Zeile der Kartenliste')

  const zeiten = []
  let umsortiert = 0
  let unsichtbar = 0
  for (let schritt = 0; schritt < schritte; schritt += 1) {
    await p.seite(messen)
    await p.taste('ArrowDown')
    const blatt = await p.warte(`Schritt ${String(schritt + 1)}: das neue Vorschaubild steht`, async () => {
      const b = await p.seite('return window.__blatt')

      return b.zuletzt === null ? false : b
    }, 15000)
    zeiten.push(blatt.dauer)
    if (blatt.zuerst !== null && blatt.zuerst !== blatt.zuletzt) umsortiert += 1
    if (blatt.unsichtbar) unsichtbar += 1
  }

  const k = kennzahlen(zeiten)
  p.zeilen.push('', '| Reihe | n | min | Median | P95 | max |', '| --- | --- | --- | --- | --- | --- |',
    `| ↓ → neues Vorschaubild (ms) | ${String(k.n)} | ${String(k.min)} | ${String(k.median)} | ${String(k.p95)} | ${String(k.max)} |`,
    '', `Rohwerte: ${JSON.stringify(zeiten.map(Math.round))}`)
  p.pruefe(`${String(schritte)} Schritte, jeder mit eigenem Vorschaubild`, zeiten.length === schritte)
  p.pruefe('das Formular sortiert sich nach der Vorschau nicht um', umsortiert === 0, `${String(umsortiert)} von ${String(schritte)}`)
  p.pruefe('das Formular steht beim Blättern nicht unsichtbar da', unsichtbar === 0, `${String(unsichtbar)} von ${String(schritte)}`)
  p.pruefe('⌘Q beendet', await p.beende())
}
