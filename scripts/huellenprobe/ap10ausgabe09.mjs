// **Ausgeben: Fußleiste, gemerktes Ziel je Set und gemerkter Bildordner, am
// nativen Fenster** — AP-10
// (#2367),
// Befund B-20, Abnahmekriterium 1: „Ausgabe-Fußleiste mit Hauptknopf bei
// 1.440 × 838 und Mindesthöhe sichtbar; gemerktes Ziel über Neustart je Set
// korrekt.“
//
// # Was es prüft
//
// Zwei künstliche Sets, *Nordprobe* mit zwölf Karten und *Südprobe* mit drei.
//
// 1. **Die Fußleiste**: bei 1.440 × 838 und bei 1.024 × 700 — dem Mindestmaß
//    aus `internal/fenster/bildschirm.go` — steht der Hauptknopf von
//    Druckbogen, Bildern, Exportvorlage und MTG JSON im Fenster, ohne dass
//    jemand rollt; **und er bleibt stehen, wenn die Fläche bis unten gerollt
//    ist**. Neben dem Druckbogen steht „9 je Seite · 2 Seiten“. Mit ⇥ ist der
//    Knopf vom Ziel aus zu erreichen.
// 2. **Der Bildordner**: „Ordner wählen …“, ein Ordner, *Bilder ausgeben* —
//    zwölf Bilder liegen dort, und `mse einstellungen` kennt den Ordner.
// 3. **Das Ziel je Set über einen Neustart**: Nord zuletzt Druckbogen, Süd
//    MTG JSON; nach ⌘Q und neuem Start ist bei Nord der Druckbogen, bei Süd
//    MTG JSON vorgewählt. Bei Nord schlagen die Bilder den gemerkten Ordner
//    vor, und *Bilder ausgeben* schreibt dorthin, **ohne Dialog**.
//
// Hell und dunkel, je mit eigenem Bestand.
//
// # Was es nicht belegt
//
// **VoiceOver** hört ein Mensch. **Ob der Dialog im gemerkten Ordner
// aufgeht**, sieht die Probe nicht — sie beantwortet ihn, bevor er sich
// zeigt; das hält `internal/app/bildordnerablage_test.go` an der Übergabe.
import { mkdirSync, mkdtempSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

function probeset(ordner, name, zahl) {
  const pfad = join(ordner, `${name}.mse-set`)
  mkdirSync(pfad)

  let text = `mse version: 2.0.0\ngame: probe\nstylesheet: schlicht\nset info:\n\ttitle: ${name}\n`
  for (let nummer = 1; nummer <= zahl; nummer++) text += `card:\n\tname: ${name} ${String(nummer)}\n`
  writeFileSync(join(pfad, 'set'), text)

  return pfad
}

const ziel = (wert) => `document.querySelector('input[type=radio][value=${wert}]')`
const fussknopf = "[...document.querySelectorAll('[data-ausgabefuss=fest] button')].find((b) => b.checkVisibility())"
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && !b.closest('[inert]') && b.innerText.trim() === ${JSON.stringify(name)})`

// lage sagt, wo der Hauptknopf steht — vor und nach dem Rollen des
// Behälters, der die Fläche trägt.
const lage = `
  const k = ${fussknopf}
  if (!k) return null
  let behaelter = k.parentElement
  while (behaelter && !(behaelter.scrollHeight > behaelter.clientHeight && /(auto|scroll)/.test(getComputedStyle(behaelter).overflowY))) behaelter = behaelter.parentElement
  // **Ohne eigenen Behälter rollt das Dokument** — \`main\` hat kein overflow.
  const wurzel = document.scrollingElement
  if (!behaelter && wurzel && wurzel.scrollHeight > wurzel.clientHeight) behaelter = wurzel
  const vorher = k.getBoundingClientRect()
  const rollt = !!behaelter
  if (behaelter) behaelter.scrollTop = behaelter.scrollHeight
  await new Promise((fertig) => requestAnimationFrame(() => requestAnimationFrame(fertig)))
  const nachher = k.getBoundingClientRect()
  if (behaelter) behaelter.scrollTop = 0
  await new Promise((fertig) => requestAnimationFrame(() => requestAnimationFrame(fertig)))
  const im = (r) => r.top >= 0 && r.bottom <= innerHeight && r.height > 0
  return { text: k.innerText.trim(), vorher: im(vorher), nachher: im(nachher), rollt, unten: Math.round(vorher.bottom), hoehe: innerHeight,
    zeile: k.closest('[data-ausgabefuss]').innerText.trim().replaceAll('\\n', ' · ') }`

async function oeffneAusgeben(p, set) {
  await p.taste('2', 'befehl')
  await p.wartetext(set, 20000)
  await p.klickeText(set)
  await p.wartetext('Karte anlegen', 20000)
  await p.taste('9', 'wahl', 'befehl')
  await p.warte('der Assistent steht', () => p.seite(`return !!${ziel('druckbogen')}`), 20000)
}

async function waehle(p, wert) {
  await p.klickeAuf(ziel(wert), `das Ziel »${wert}«`)
  await p.warte(`»${wert}« ist gewählt`, () => p.seite(`return ${ziel(wert)}.checked`), 5000)
}

async function fussleiste(p, wert, bild) {
  await waehle(p, wert)
  for (const [breite, hoehe] of [[1440, 838], [1024, 700]]) {
    await p.groesse(breite, hoehe)
    const stand = await p.warte(`der Knopf von »${wert}« steht in der Fußleiste`, () => p.seite(lage), 15000)
    p.pruefe(`${bild}, ${String(breite)} × ${String(hoehe)}, ${wert}: der Hauptknopf steht im Fenster, ohne Rollen`, stand.vorher, JSON.stringify(stand))
    p.pruefe(`${bild}, ${String(breite)} × ${String(hoehe)}, ${wert}: und bleibt stehen, wenn bis unten gerollt ist`, stand.nachher, JSON.stringify(stand))
    if (wert === 'druckbogen') {
      p.pruefe(`${bild}, ${String(breite)} × ${String(hoehe)}: neben dem Druckbogen steht, was er legt`,
        stand.zeile.includes('9 je Seite · 2 Seiten'), stand.zeile)
      // **Sonst prüfte „bleibt stehen" nichts** — die Fläche muss rollen.
      p.pruefe(`${bild}, ${String(breite)} × ${String(hoehe)}: die Fläche des Druckbogens reicht über das Fenster`, stand.rollt, JSON.stringify(stand))
    }
    if (hoehe === 700) await verdeckt(p, bild, wert)
    await p.bild(`${wert}-${String(breite)}-${bild}`)
  }
  await p.groesse(1440, 838)

  // **Mit ⇥ vom Ziel aus erreichbar** — gezählt, nicht nur erreicht.
  await fokussiereZiel(p, wert)
  const weg = []
  while (weg.length < 60 && !(await p.seite(`return document.activeElement === ${fussknopf}`))) {
    weg.push(await tab(p))
  }
  p.pruefe(`${bild}, ${wert}: der Hauptknopf ist mit ⇥ erreichbar, ohne die Seite zu verlassen`,
    weg.length < 60 && !weg.some((halt) => halt.startsWith('body')), `${String(weg.length)} Schritte: ${weg.join(' → ')}`)
}

// tab drückt ⇥ und **wartet, bis der Fokus gewandert ist** — sonst läse der
// Beleg den Halt davor (gesehen am 2026-10-17: Namen doppelt und versetzt).
async function tab(p) {
  await p.seite('window.__probefokus = document.activeElement; return true')
  await p.taste('Tab')
  // **Eine eigene Schleife und nicht `p.warte`**: ein Fokus, der bleibt, ist
  // ein Halt im Weg und kein Fehler des Laufs — er steht als „(blieb)“ dort.
  const ende = Date.now() + 3000
  while (Date.now() < ende) {
    if (await p.seite('return document.activeElement !== window.__probefokus')) return p.seite(fokusname)
    await new Promise((weiter) => { setTimeout(weiter, 50) })
  }
  return `(blieb) ${await p.seite(fokusname)}`
}

// fokusname beschreibt, was den Fokus hat — für den Weg im Beleg.
const fokusname = "const e = document.activeElement; return e ? `${e.tagName.toLowerCase()}${e.type ? '[' + e.type + ']' : ''} »${(e.getAttribute('aria-label') ?? e.labels?.[0]?.innerText ?? e.innerText ?? '').trim().split('\\n')[0].slice(0, 40)}«` : 'nichts'"

// fokussiereZiel setzt den Fokus auf das Ziel und **wartet, bis er dort ist**
// — ein Fokus, der nicht ankam, ließe ⇥ vom Anfang des Dokuments laufen.
async function fokussiereZiel(p, wert) {
  await p.warte(`der Fokus steht auf »${wert}«`, () => p.seite(`${ziel(wert)}.focus(); return document.activeElement === ${ziel(wert)}`), 5000)
}

// verdeckt geht mit ⇥ vom Ziel bis zum Hauptknopf und misst an jedem Halt,
// ob die Fußleiste ihn verdeckt — WCAG 2.2, 2.4.11 *Focus Not Obscured
// (Minimum)*: kein fokussiertes Element ganz hinter ihr.
async function verdeckt(p, bild, wert) {
  await fokussiereZiel(p, wert)
  const halte = []
  for (let schritt = 0; schritt < 40; schritt++) {
    const name = await tab(p)
    const halt = await p.seite(`
      await new Promise((fertig) => requestAnimationFrame(() => requestAnimationFrame(fertig)))
      const e = document.activeElement
      const imFuss = !!e?.closest('[data-ausgabefuss]')
      // **Die sichtbare Leiste** — verborgene Ziele behalten ihre im Baum.
      const leiste = [...document.querySelectorAll('[data-ausgabefuss=fest]')].find((l) => l.checkVisibility())
      if (!e || !leiste) return { ganz: false, teils: false, imFuss, ohneLeiste: !leiste }
      const fuss = leiste.getBoundingClientRect()
      const r = e.getBoundingClientRect()
      const ganz = !imFuss && r.top >= fuss.top
      const teils = !imFuss && !ganz && r.bottom > fuss.top
      return { ganz, teils, imFuss, ohneLeiste: false }`)
    halt.name = name
    halte.push(halt)
    if (halt.imFuss) break
  }
  const ganz = halte.filter((h) => h.ganz).map((h) => h.name)
  const teils = halte.filter((h) => h.teils).map((h) => h.name)
  p.pruefe(`${bild}, 1024 × 700, ${wert}: kein Halt bis zum Hauptknopf liegt ganz hinter der Fußleiste (WCAG 2.4.11)`,
    ganz.length === 0 && halte.at(-1)?.imFuss === true && !halte.some((h) => h.ohneLeiste), `${String(halte.length)} Halte; ganz verdeckt: ${ganz.join(', ') || 'keiner'}; teils: ${teils.join(', ') || 'keiner'}`)
}

async function lauf(p, bild) {
  const ort = p.ort({ zuhause: true })
  const ordner = mkdtempSync(join(tmpdir(), 'ap10ausgabe-'))
  const bilder = join(ordner, 'bilder')
  mkdirSync(bilder)
  p.mse(ort, 'oeffnen', probeset(ordner, 'Nordprobe', 12))
  p.mse(ort, 'oeffnen', probeset(ordner, 'Südprobe', 3))
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')

  p.notiz(`${bild}: die Fußleiste der vier Ziele, bei 1.440 × 838 und 1.024 × 700`)
  await p.starte({ ort })
  await p.groesse(1440, 838)
  await oeffneAusgeben(p, 'Nordprobe')
  p.pruefe(`${bild}: ein frisches Set wählt nichts vor`, await p.seite("return !['setdatei', 'bilder', 'druckbogen', 'vorlage', 'mtgjson'].some((w) => document.querySelector(`input[type=radio][value=${w}]`)?.checked)"))
  for (const wert of ['druckbogen', 'vorlage', 'mtgjson']) await fussleiste(p, wert, bild)

  // **Die Bilder erst mit Ordner** — ohne ihn ist ihr Knopf angehalten und
  // mit ⇥ nicht zu erreichen, mit Absicht.
  p.notiz(`${bild}: der Bildordner wird gewählt, und die Bilder liegen dort`)
  await waehle(p, 'bilder')
  await p.dialogantwort(bilder)
  await p.klickeAuf(knopf('Ordner wählen …'), '»Ordner wählen …«')
  await p.warte('der Ordner steht da', () => p.seite(`return document.body.innerText.includes(${JSON.stringify(bilder)})`), 10000)
  await p.klickeAuf(fussknopf, '»Bilder ausgeben«')
  await p.warte('zwölf Bilder liegen im Ordner', () => readdirSync(bilder).filter((datei) => datei.endsWith('.png')).length === 12, 60000)
  const gemerkt = await p.warte('der Ordner ist gemerkt', () => {
    const eintrag = JSON.parse(p.mse(ort, 'einstellungen', '--json')).einstellungen.find((e) => e.schluessel === 'bildausfuhr.ordner')
    return eintrag?.wert === bilder ? eintrag.wert : null
  }, 10000)
  p.pruefe(`${bild}: mse einstellungen kennt den Bildordner`, gemerkt === bilder, gemerkt)
  await fussleiste(p, 'bilder', bild)

  p.notiz(`${bild}: Nord zuletzt Druckbogen, Süd MTG JSON`)
  await waehle(p, 'druckbogen')
  await oeffneAusgeben(p, 'Südprobe')
  await waehle(p, 'mtgjson')
  await p.warte('beide Ziele sind gemerkt', () => {
    const eintrag = JSON.parse(p.mse(ort, 'einstellungen', '--json')).einstellungen.find((e) => e.schluessel === 'ausgabe.ziele')
    const ziele = JSON.parse(eintrag?.wert || '[]').map((z) => z.ziel).join(' ')
    return ziele.endsWith('druckbogen mtgjson') ? ziele : null
  }, 10000)
  p.pruefe(`${bild}: ⌘Q beendet`, await p.beende())

  p.notiz(`${bild}: nach dem Neustart stehen beide Ziele und der Ordner wieder da`)
  rmSync(bilder, { recursive: true })
  mkdirSync(bilder)
  await p.starte({ ort })
  await p.groesse(1440, 838)
  await oeffneAusgeben(p, 'Nordprobe')
  await p.warte('bei Nord ist der Druckbogen vorgewählt', () => p.seite(`return ${ziel('druckbogen')}.checked`), 10000)
  p.pruefe(`${bild}: bei Nord steht die Fläche des Druckbogens`, await p.seite("return !!document.querySelector('[data-ausgabefuss=fest]') && document.body.innerText.includes('Druckbogen legen')"))
  await p.bild(`nord-nach-neustart-${bild}`)
  await oeffneAusgeben(p, 'Südprobe')
  await p.warte('bei Süd ist MTG JSON vorgewählt', () => p.seite(`return ${ziel('mtgjson')}.checked`), 10000)
  await p.bild(`sued-nach-neustart-${bild}`)

  await oeffneAusgeben(p, 'Nordprobe')
  await waehle(p, 'bilder')
  await p.warte('der gemerkte Ordner steht vorbelegt da', () => p.seite(`return document.body.innerText.includes(${JSON.stringify(bilder)}) && document.body.innerText.includes('Zuletzt gewählt — derselbe Ordner für alle Sets.')`), 10000)
  await p.bild(`bildordner-gemerkt-${bild}`)
  await p.klickeAuf(fussknopf, '»Bilder ausgeben«')
  await p.warte('die Bilder liegen wieder im gemerkten Ordner, ohne Dialog', () => readdirSync(bilder).filter((datei) => datei.endsWith('.png')).length === 12, 60000)
  p.pruefe(`${bild}: ⌘Q beendet`, await p.beende())
}

export default async function ap10ausgabe09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
