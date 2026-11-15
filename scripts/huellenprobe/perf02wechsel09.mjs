// **Ein Wechsel während der Prüfung gibt keine falsche Zahl** — PERF02
// (#2352),
// zweites Abnahmekriterium, am nativen Fenster.
//
// # Was es prüft
//
// Ein Set mit 3.788 erfundenen Karten und dem Stil M15 (`GROSS` aus
// `budgetsorte09`), dessen Prüfung am Fenster lange genug läuft. Während sie
// läuft, ändert **eine andere Fläche** — die Kommandozeile — eine Karte.
// Danach darf die Startseite **keine Zahl** für den alten Stand zeigen: die
// Kachel sagt „nicht vollständig geprüft“, die Zeile darunter fordert zur
// Prüfung auf. Eine zweite Prüfung bringt dieselben Zahlen wie das gebündelte
// `mse probleme` über denselben Stand.
//
// # Was es nicht belegt
//
// **Einen Profilwechsel während der Prüfung** — den prüfen die Einheitstests
// (`startprofilwechsel.test.tsx`), und die Kacheln hängen dort unter
// `key={profilstand}`. **Eine fehlende Schrift** ist ein Befund der Karte und
// keine Null (`PERF02-2352.md`).
import { execFileSync } from 'node:child_process'
import { writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'
import { GROSS, mitPack, mtgjson } from './budgetsorte09.mjs'

const gebuendeltesMse = (ort, ...argumente) =>
  execFileSync(join(process.cwd(), 'build/bin/MagicSetEditor4Mac.app/Contents/MacOS/mse'), argumente, { env: kindumgebung(ort), maxBuffer: 64 << 20 }).toString()

const liste = "[...document.querySelectorAll('ul[aria-label]')].find((u) => u.checkVisibility() && u.getAttribute('aria-label') === 'Dein Überblick')"
const kachel = (wort) => `[...(${liste})?.querySelectorAll('button') ?? []].find((b) => b.innerText.replace(/\\s+/g, ' ').includes(${JSON.stringify(wort)}))`
const kacheltext = (wort) => `return ${kachel(wort)}?.innerText.replace(/\\s+/g, ' ').trim() ?? null`
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim() === ${JSON.stringify(name)})`
const zeilen = "return [...document.querySelectorAll('main p')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim())"

export default async function perf02wechsel09(p) {
  const ort = p.ort()
  mitPack(p, ort)
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  const datei = join(ort.ort, `${GROSS.code}.json`)
  writeFileSync(datei, JSON.stringify(mtgjson(GROSS)))
  p.mse(ort, 'mtgjson-lesen', datei, 'magic')
  const setID = p.mse(ort, 'sets').trim().split('\n').map((z) => z.split('\t')).find((t) => t[1] === GROSS.name)?.[0]
  p.zeige(`${GROSS.name}: Set ${String(setID)}, ${p.mse(ort, 'stile', String(setID), 'wechseln', 'magic-m15').split('\n')[0]}`)

  await p.starte({ ort })
  await p.groesse(1440, 838)
  await p.taste('1', 'befehl')
  await p.warte('die Kachel der Sets steht', () => p.seite(`return !!${kachel('Set')}`), 30000)

  p.notiz('Während der Prüfung ändert die Kommandozeile eine Karte')
  // **Solange sie läuft, ruht der Knopf** (`angehalten`, §7 Regel 4) — das ist
  // das Zeichen, auf das hier gewartet wird.
  const ruht = `return ${knopf('Jetzt prüfen')}?.disabled === true`
  await p.klickeAuf(knopf('Jetzt prüfen'), '»Jetzt prüfen«')
  const beginn = Date.now()
  await p.warte('die Prüfung läuft', () => p.seite(ruht), 15000)
  p.mse(ort, 'feld', String(setID), '1', 'name', 'Mitten in der Prüfung geändert')
  const geaendert = Date.now() - beginn
  const nochDabei = await p.seite(ruht)
  await p.warte('die Prüfung ist zu Ende', async () => !(await p.seite(ruht)), 300000)
  const dauer = Date.now() - beginn
  p.zeige(`geändert nach ${String(geaendert)} ms, Prüfung zu Ende nach ${String(dauer)} ms`)
  p.pruefe('die Änderung kam, während die Prüfung lief', nochDabei === true && geaendert < dauer, `${String(geaendert)} < ${String(dauer)}`)

  await p.warte('die Startseite hat nachgelesen', async () => (await p.seite(zeilen)).some((t) => /^Probleme/.test(t)), 30000)
  const sets = await p.seite(kacheltext(' Set'))
  const karten = await p.seite(kacheltext(' Karte'))
  const zeile = (await p.seite(zeilen)).find((t) => /^Probleme/.test(t)) ?? ''
  p.pruefe('die Kachel der Karten nennt keine Zahl für den alten Stand', /nicht vollständig geprüft/.test(karten ?? '') && !/mit Befund/.test(karten ?? ''), karten)
  p.pruefe('die Kachel der Sets sagt, dass nicht alles geprüft ist', /0 von 1 geprüft/.test(sets ?? ''), sets)
  p.pruefe('die Zeile fordert wieder zur Prüfung auf', zeile === 'Probleme: 1 von 1 Sets nicht geprüft' && (await p.seite(`return !!${knopf('Jetzt prüfen')}`)), zeile)
  await p.bild('nach-dem-wechsel')

  p.notiz('Eine zweite Prüfung zählt wie die Kommandozeile')
  const erwartet = JSON.parse(gebuendeltesMse(ort, 'probleme', String(setID), '--json'))
  await p.klickeAuf(knopf('Jetzt prüfen'), '»Jetzt prüfen«')
  await p.warte('die Prüfung ist fertig', async () => (await p.seite(zeilen)).some((t) => /^Probleme geprüft /.test(t)), 240000)
  const nachher = await p.seite(kacheltext(' Karte'))
  p.pruefe('die Kachel der Karten zählt wie die Kommandozeile', (nachher ?? '').endsWith(`${String(erwartet.zahl)} mit Befund`), `${String(nachher)} — Kommandozeile: ${String(erwartet.zahl)}`)
  await p.bild('zweite-pruefung')
  p.pruefe('⌘Q beendet', await p.beende())
}
