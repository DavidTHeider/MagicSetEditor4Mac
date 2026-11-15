// **Der Messvertrag am nativen Fenster: je Kernweg Zeit und Speicher** —
// PERF01 (#2351),
// Befunde A-13, A-14 und A-25 des Audits 0.9.0.
//
// # Der Vertrag, vor dem Lauf festgelegt
//
// Die Reihen und ihre Grenzen stehen in `scripts/messvertrag.mjs`; dieses
// Drehbuch misst sie und schreibt je Reihe eine Zeile mit Urteil. **Eine
// gerissene Grenze macht den Lauf rot** — nicht nur die Tafel.
//
// - **Start kalt und warm**: je zwanzig Starts über LaunchServices bis die
//   Leiste antwortet, wie in `budgets` (B02) — kalt je ein frisches Zuhause,
//   warm dasselbe nach ⌘Q. Der Dateicache des Systems bleibt warm.
// - **Set öffnen**: zwanzigmal die Kachel von »Probesatz 0.9.0« (3.788
//   Karten), in der Seite gemessen vom `mousedown` bis die erste Karte in der
//   Liste steht und zwei Bilder gezeichnet sind; dazwischen ⌘W.
// - **Feld → Vorschaubild**: 24 Tasten im Namensfeld einer Karte dieses Sets,
//   gemessen wie in `budgets`.
// - **Sammlung durchsuchen**: 24 Suchen über 25.000 Zeilen — abwechselnd
//   vier Ziffern getippt und das Feld geleert —, gemessen vom **letzten**
//   `input` bis die Zählzeile »… von …« wechselt und zwei Bilder stehen.
// - **Ausgeben, Druckbogen, Zurückspielen**: je dreimal, gemessen vom Klick
//   bis die Seite das Ende meldet — die 60 Karten des Drucksets als PNG und
//   als Bogen in A4, die Sicherung des ganzen Bestands über Einstellungen ›
//   Daten.
// - **Fußabdruck nach jedem Weg**: Anwendung und ihre WebKit-Prozesse
//   (`footprint`), ohne fremde WebKit-Prozesse des Systems.
// - **Bündel, logisch**: die Größe aller Dateien des gebauten Bündels
//   (`du -A`), nicht der belegte Platz.
//
// # Was es nicht belegt
//
// Einen anderen Mac — **diese Maschine, dieser Lauf**; Last und Ausstattung
// stehen im Beleg. Eine Mindestausstattung belegt erst ein Gast mit 8 GB.
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync, rmSync } from 'node:fs'
import { cpus, loadavg, totalmem } from 'node:os'
import { join, resolve } from 'node:path'

import { MESSVERTRAG, fussabdrucksumme, tafel, urteil } from '../messvertrag.mjs'
import { namensfeld, vorschauzeit } from './budgets.mjs'
import { DRUCK, GROSS, ZEILEN, befuellt, pack } from './budgetsorte09.mjs'
import { fussabdruck } from './dauerbetrieb.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))
const ZUGE = 20

export const prozesse = () => execFileSync('ps', ['-Ao', 'pid=,comm=']).toString().split('\n')
  .map((z) => z.trim().split(/\s+/)).filter((t) => t.length >= 2).map(([pid, ...name]) => ({ pid, name: name.join(' ') }))

// speicher gibt den Fußabdruck der Anwendung und **ihrer** WebKit-Prozesse —
// die nach `vorher` entstanden; `EnhancedSecurity` startet das System selbst.
// **Fehlt ein Prozess, ist die Messung gescheitert** (−1, Zweitblick über
// Phase 3, Befund 12) und keine, die mit null für ihn zählt.
export function speicher(vorher, anwendung) {
  const web = prozesse().filter((x) => !vorher.has(x.pid) && /WebKit\.(WebContent|GPU|Networking)$/.test(x.name)).map((x) => x.pid)
  const werte = fussabdruck([anwendung, ...web])

  return { anwendung: fussabdrucksumme(werte, [anwendung]), web: fussabdrucksumme(werte, web), zusammen: fussabdrucksumme(werte, [anwendung, ...web]) }
}

// wirkung misst in der Seite von einem Ereignis bis eine Bedingung gilt und
// zwei Bilder gezeichnet sind — **angelegt vor** dem Griff, wie in `budgets`.
// **Die Frist ist kürzer als die der Leitung** (30 s): sonst gewinnt die
// Leitung, und statt einer gescheiterten Messung bricht das Drehbuch ab.
const wirkung = (ereignis, bedingung) => `window.__wirkung = new Promise((fertig) => {
  let beginn = 0
  document.addEventListener(${JSON.stringify(ereignis)}, () => { beginn = performance.now() }, { capture: true, once: true })
  const frist = setTimeout(() => fertig(-1), 25000)
  const schau = () => {
    if (beginn && (${bedingung})) requestAnimationFrame(() => requestAnimationFrame(() => { clearTimeout(frist); fertig(performance.now() - beginn) }))
    else requestAnimationFrame(schau)
  }
  requestAnimationFrame(schau)
})
return true`

const kachel = (name) => `[...document.querySelectorAll('main a, main button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(name)})`
const ersteKarte = `!![...document.querySelectorAll('main [role=row], main [role=option], main li')].find((e) => e.checkVisibility() && (e.innerText ?? '').includes(${JSON.stringify(`${GROSS.vorsatz} 0001`)}))`
const suchfeld = "[...document.querySelectorAll('main input[type=search]')].find((e) => e.checkVisibility())"
const zaehlzeile = "([...document.querySelectorAll('main *')].filter((e) => e.children.length === 0 && e.checkVisibility()).map((e) => (e.innerText ?? '').trim()).find((t) => /^[\\d.\\s]+ von [\\d.\\s]+$/.test(t)) ?? '')"
const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.replace(/\\s+/g, ' ').trim().replace(/^[^\\p{L}\\p{N}]+/u, '') === ${JSON.stringify(text)})`
const steht = (text) => `return document.body.innerText.includes(${JSON.stringify(text)})`

// suchzeit misst vom letzten `input` des Suchfelds, bis die Zählzeile
// wechselt und das Feld den erwarteten Wert trägt.
const suchzeit = (erwartet) => `window.__suche = new Promise((fertig) => {
  const feld = ${suchfeld}
  const vorher = ${zaehlzeile}
  let letzte = 0
  const merke = () => { letzte = performance.now() }
  feld.addEventListener('input', merke, { capture: true })
  const ende = (ms) => { feld.removeEventListener('input', merke, { capture: true }); fertig({ ms, zeile: ${zaehlzeile} }) }
  const frist = setTimeout(() => ende(-1), 25000)
  const schau = () => {
    if (letzte && feld.value === ${JSON.stringify(erwartet)} && ${zaehlzeile} !== vorher) requestAnimationFrame(() => requestAnimationFrame(() => { clearTimeout(frist); ende(performance.now() - letzte) }))
    else requestAnimationFrame(schau)
  }
  requestAnimationFrame(schau)
})
return true`

async function klicke(p, text) {
  await p.warte(`»${text}« steht da`, () => p.seite(`return !!(${knopf(text)}) && !(${knopf(text)}).disabled`), 20000)
  await p.seite(`(${knopf(text)}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(knopf(text), `»${text}«`)
}

// dauer misst in Millisekunden vom Klick bis zur Wirkung, gefragt alle 50 ms.
async function dauer(p, text, wirkt, frist = 180000) {
  await p.warte(`»${text}« steht da`, () => p.seite(`return !!(${knopf(text)}) && !(${knopf(text)}).disabled`), 20000)
  // **Erst wenn die Seite ruht**: unter »Diese Sicherung zurückspielen?«
  // erscheint die Kostentafel später und schiebt den Knopf nach unten — der
  // Klick traf im zweiten Lauf ihre Überschrift.
  await p.ruhe()
  await p.seite(`(${knopf(text)}).scrollIntoView({ block: 'center' }); return true`)
  const beginn = Date.now()
  await p.klickeAuf(knopf(text), `»${text}«`)
  await p.warte(`nach »${text}«`, wirkt, frist, 50)

  return Date.now() - beginn
}

async function starts(p, werte) {
  for (let i = 0; i < ZUGE; i++) {
    const frisch = p.ort()
    p.mse(frisch, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
    p.mse(frisch, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    const beginn = Date.now()
    await p.starte({ ort: frisch, takt: 20 })
    werte['Start kalt (ms)'].push(p.bereitUm - beginn)
    p.pruefe(`Kaltstart ${String(i + 1)}: ⌘Q beendet`, await p.beende())
  }
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  for (let i = 0; i <= ZUGE; i++) {
    const beginn = Date.now()
    await p.starte({ ort, takt: 20 })
    // **Der erste Start in diesem Zuhause ist kalt** und zählt nicht.
    if (i > 0) werte['Start warm (ms)'].push(p.bereitUm - beginn)
    p.pruefe(`Warmstart ${String(i)}: ⌘Q beendet`, await p.beende())
  }
}

async function setOeffnen(p, werte) {
  for (let i = 0; i < ZUGE; i++) {
    await p.taste('w', 'befehl')
    await p.warte(`»${GROSS.name}« steht als Kachel`, () => p.seite(`return !!(${kachel(GROSS.name)})`), 30000)
    await p.seite(wirkung('mousedown', ersteKarte))
    await p.klickeAuf(kachel(GROSS.name), `»${GROSS.name}«`)
    werte['Set öffnen, 3.788 Karten (ms)'].push(await p.seite('return window.__wirkung'))
  }
}

async function vorschauen(p, werte) {
  const zeile = `[...document.querySelectorAll('main [role=row], main [role=option], main li')].find((e) => e.checkVisibility() && (e.innerText ?? '').includes(${JSON.stringify(`${GROSS.vorsatz} 0001`)}))`
  if (!(await p.seite(`return !!(${namensfeld})`))) await p.klickeAuf(zeile, `»${GROSS.vorsatz} 0001«`)
  await p.warte('das Namensfeld steht da', () => p.seite(`return !!(${namensfeld})`), 20000)
  await p.klickeAuf(namensfeld, 'das Namensfeld')
  let daneben = 0
  for (let i = 0; werte['Feld → Vorschaubild (ms)'].length < 24 && i < 40; i++) {
    if (!(await p.seite(`return document.activeElement === (${namensfeld})`))) await p.klickeAuf(namensfeld, 'das Namensfeld')
    const vorher = await p.seite(`return (${namensfeld}).value`)
    await p.seite(vorschauzeit)
    await p.taste(i % 2 === 0 ? 'x' : 'Backspace')
    const { ms, grund } = await p.seite('return window.__vorschaumessung')
    if ((await p.seite(`return (${namensfeld}).value`)) === vorher) { daneben++; continue }
    if (ms < 0 && !werte.__vorschaubild) {
      werte.__vorschaubild = true
      p.zeige(`die Vorschau: ${await p.seite("return [...document.querySelectorAll('img')].filter((b) => b.alt.includes('Vorschau')).map((b) => `${b.alt}: ${b.checkVisibility()} ${b.src.slice(0, 60)}`).join(' | ') || 'kein Bild mit »Vorschau«'")}`)
      await p.bild('vorschau-blieb')
    }
    p.pruefe(`Vorschau ${String(i + 1)} kam an`, ms >= 0, grund ?? '')
    werte['Feld → Vorschaubild (ms)'].push(ms)
  }
  p.zeige(`${String(daneben)} Tasten kamen nicht im Feld an und zählen nicht`)
}

async function suchen(p, werte) {
  await p.taste('3', 'befehl')
  await p.warte(`die Sammlung zählt ${String(ZEILEN)}`, () => p.seite(`return ${zaehlzeile}.replace(/[.\\s]/g, '').endsWith('von${String(ZEILEN)}')`), 60000)
  await p.klickeAuf(suchfeld, 'das Suchfeld der Sammlung')
  // **Geleert wird Zeichen für Zeichen**, wie getippt: ⌘A und ⌫ kamen im
  // ersten vollen Lauf vertauscht an, und das Feld behielt drei Ziffern.
  let vorher = ''
  for (let i = 0; i < 24; i++) {
    const begriff = i % 2 === 0 ? String(1000 + i * 397).padStart(4, '0') : ''
    await p.seite(suchzeit(begriff))
    if (begriff) await p.tippe(begriff)
    else for (const _ of vorher) await p.taste('Backspace')
    vorher = begriff
    const { ms, zeile } = await p.seite('return window.__suche')
    p.pruefe(`Suche ${String(i + 1)} »${begriff}« kam an`, ms >= 0, zeile)
    werte['Sammlung durchsuchen, 25.000 Zeilen (ms)'].push(ms)
  }
}

async function ausgeben(p, ort, werte) {
  await p.taste('w', 'befehl')
  await p.warte(`»${DRUCK.name}« steht als Kachel`, () => p.seite(`return !!(${kachel(DRUCK.name)})`), 30000)
  await p.klickeAuf(kachel(DRUCK.name), `»${DRUCK.name}«`)
  for (let i = 0; i < 3; i++) {
    const ordner = join(ort.ort, `bilder-${String(i)}`)
    mkdirSync(ordner, { recursive: true })
    await p.klick('Ablage', 'Ausgeben', 'Als Bilder …')
    await p.dialogantwort(ordner)
    await klicke(p, 'Ordner wählen …')
    await p.warte('der Ordner ist gewählt', async () => !(await p.seite(steht('Noch kein Ordner gewählt.'))), 20000)
    werte['Set ausgeben als Bilder, 60 Karten (ms)'].push(await dauer(p, 'Bilder ausgeben', () => p.seite(steht('Bilder geschrieben nach'))))
    werte['Fußabdruck nach jedem Weg (MiB)'].push(speicher(p.vorher, p.pids.at(-1)).zusammen)
  }
  for (let i = 0; i < 3; i++) {
    const bogen = join(ort.ort, `bogen-${String(i)}.pdf`)
    await p.klick('Ablage', 'Ausgeben', 'Druckbogen …')
    await p.dialogantwort(bogen)
    werte['Druckbogen, 60 Karten (ms)'].push(await dauer(p, 'Druckbogen legen …', async () => existsSync(bogen) && p.seite(steht(`${String(DRUCK.karten)} Karten auf`))))
    werte['Fußabdruck nach jedem Weg (MiB)'].push(speicher(p.vorher, p.pids.at(-1)).zusammen)
    rmSync(bogen, { force: true })
  }
}

async function zurueckspielen(p, sicherung, werte) {
  for (let i = 0; i < 3; i++) {
    await p.taste(',', 'befehl')
    const gruppe = "[...document.querySelectorAll('button, a, [role=tab]')].find((e) => e.checkVisibility() && e.innerText.replace(/\\s+/g, ' ').trim().startsWith('Daten'))"
    await p.warte('die Gruppe »Daten« steht zur Wahl', () => p.seite(`return !!(${gruppe})`), 20000)
    await p.klickeAuf(gruppe, 'die Gruppe »Daten«')
    await p.dialogantwort(sicherung)
    await klicke(p, 'Sicherung zurückspielen …')
    await p.wartetext('Diese Sicherung zurückspielen?', 20000)
    await p.wartetext('Was ein Zurückspielen kostet', 20000)
    werte['Zurückspielen, 25.000 Zeilen (ms)'].push(await dauer(p, 'Zurückspielen', () => p.seite(steht('Zurückgespielt.'))))
    werte['Fußabdruck nach jedem Weg (MiB)'].push(speicher(p.vorher, p.pids.at(-1)).zusammen)
  }
}

export default async function budgets09(p) {
  p.zeige(`Last vor dem Lauf ${loadavg().map((z) => z.toFixed(2)).join(' / ')} auf ${String(cpus().length)} Kernen, ${String(Math.round(totalmem() / 2 ** 30))} GiB`)
  const werte = Object.fromEntries(MESSVERTRAG.filter((e) => e.beleg === 'huellenprobe/budgets09.md').map((e) => [e.reihe, []]))
  const buendel = join(wurzel, 'build/bin/MagicSetEditor4Mac.app')
  werte['Bündel, logisch (MiB)'].push(Number(execFileSync('du', ['-A', '-k', '-s', buendel]).toString().split(/\s+/)[0]) / 1024)

  // **`MSE_BUDGETS_OHNE_START=1` lässt die Starts aus** — zur Fehlersuche an
  // den übrigen Wegen; dann fehlen zwei Reihen, und der Lauf ist rot.
  if (process.env.MSE_BUDGETS_OHNE_START !== '1') {
    p.notiz('Start bis bedienbar, kalt und warm')
    await starts(p, werte)
  }

  if (!existsSync(join(pack, 'data'))) {
    p.pruefe(`der Pack liegt unter ${pack}`, false)
  } else {
    p.notiz(`Der Bestand: ${String(GROSS.karten)} und ${String(DRUCK.karten)} Karten, ${String(ZEILEN)} Zeilen in der Sammlung`)
    const ort = p.ort()
    const bestand = befuellt(p, ort)
    p.pruefe(`die Sammlung hat ${String(ZEILEN)} Zeilen`, bestand.zeilen === ZEILEN, String(bestand.zeilen))
    p.vorher = new Set(prozesse().map((x) => x.pid))
    await p.starte({ ort })
    await p.taste('2', 'befehl')
    await p.warte(`»${GROSS.name}« steht als Kachel`, () => p.seite(`return !!(${kachel(GROSS.name)})`), 30000)
    await p.klickeAuf(kachel(GROSS.name), `»${GROSS.name}«`)
    await p.warte('die erste Karte steht', () => p.seite(`return ${ersteKarte}`), 30000)
    const merke = () => werte['Fußabdruck nach jedem Weg (MiB)'].push(speicher(p.vorher, p.pids.at(-1)).zusammen)

    // **`MSE_BUDGETS_WEGE` wählt Wege** zur Fehlersuche — ohne die übrigen
    // ist der Lauf rot.
    const wege = (process.env.MSE_BUDGETS_WEGE ?? 'oeffnen,vorschau,suche,ausgabe,rueckspiel').split(',')
    if (wege.includes('oeffnen')) { p.notiz('Set öffnen'); await setOeffnen(p, werte); merke() }
    if (wege.includes('vorschau')) { p.notiz('Feld → dekodiertes Vorschaubild'); await vorschauen(p, werte); merke() }
    if (wege.includes('suche')) { p.notiz('Sammlung durchsuchen'); await suchen(p, werte); merke() }
    if (wege.includes('ausgabe')) { p.notiz('Ausgeben und Druckbogen'); await ausgeben(p, ort, werte) }
    if (wege.includes('rueckspiel')) { p.notiz('Zurückspielen'); await zurueckspielen(p, bestand.sicherung, werte) }
    await schlafe(500)
    p.pruefe('⌘Q beendet', await p.beende())
  }

  p.notiz('Ergebnis')
  p.zeilen.push('', '| Reihe | n | min | Median | P95 | max | Grenze | Urteil |', '| --- | --- | --- | --- | --- | --- | --- | --- |')
  for (const e of MESSVERTRAG.filter((x) => x.beleg === 'huellenprobe/budgets09.md')) {
    p.zeilen.push(tafel(e, werte[e.reihe]))
    p.pruefe(`${e.reihe}: ${e.mass} ≤ ${String(e.grenze)}`, urteil(e, werte[e.reihe]) === 'gehalten', urteil(e, werte[e.reihe]))
  }
  delete werte.__vorschaubild
  p.zeilen.push('', `Rohwerte — ${Object.entries(werte).map(([reihe, liste]) => `${reihe}: ${JSON.stringify(liste.map(Math.round))}`).join('; ')}.`)
}
