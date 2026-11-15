// **„Wie es wird" vor dem Schreiben, am nativen Fenster** — UXR-F7
// (#2015):
//
// - der Druckbogen zeigt über seinem Knopf, wie er würde — „9 je Seite · 3
//   Seiten" und das Raster einer Seite —, **bevor** etwas geschrieben ist,
// - eine andere Wahl (Letter) ändert die Zahl, und der Bogen, den der Knopf
//   danach legt, hat genau die Seitenzahl, die der Kasten nannte,
// - die Exportvorlage nennt Dateiname und -größe; die Datei, die der Knopf
//   danach schreibt, ist genau so groß,
// - bis dahin liegt nichts auf der Platte.
//
// **Ein eigenes Zuhause**: die Saat (`probe-liste.mse-export-template`) liegt
// im verwalteten Paketverzeichnis, und das Set entsteht in der Probe.
//
// **Was sie nicht belegt**: ob jemand das Raster als „so sieht eine Seite aus"
// liest — das sieht nur ein Mensch (UXR-F7, Sichtprüfung).
import { existsSync, mkdirSync, mkdtempSync, readdirSync, statSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(name)})`
const kasten = "[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && document.getElementById(s.getAttribute('aria-labelledby') ?? '')?.innerText.trim() === 'Wie es wird')"
const status = (anfang) => `const e = [...document.querySelectorAll('[role=status]')].find((s) => s.checkVisibility() && s.innerText.startsWith(${JSON.stringify(anfang)})); return e ? e.innerText : false`

// zahlen liest, was im Kasten steht.
const zahlen = `const k = ${kasten}; return k ? k.innerText.replace(/\\s+/g, ' ').trim() : false`

function probeset(ordner, zahl) {
  const pfad = join(ordner, 'wieeswird.mse-set')
  mkdirSync(pfad)

  let text = 'mse version: 2.0.0\ngame: probe\nstylesheet: schlicht\nset info:\n\ttitle: Wie-es-wird-Set\n'
  for (let nummer = 1; nummer <= zahl; nummer++) text += `card:\n\tname: Karte ${String(nummer)}\n`
  writeFileSync(join(pfad, 'set'), text)

  return pfad
}

// setzeWahl stellt eine Auswahl ein — React hört auf `change`.
const setzeWahl = (beschriftung, wert) => `const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && (w.labels?.[0]?.textContent ?? '').includes(${JSON.stringify(beschriftung)})); if (!s) return false; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, ${JSON.stringify(wert)}); s.dispatchEvent(new Event('change', { bubbles: true })); return true`

// setzeText füllt ein Textfeld — React hört auf `input`.
const setzeText = (beschriftung, wert) => `const f = [...document.querySelectorAll('input[type=text], input:not([type])')].find((w) => w.checkVisibility() && (w.labels?.[0]?.textContent ?? '').includes(${JSON.stringify(beschriftung)})); if (!f) return false; Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(f, ${JSON.stringify(wert)}); f.dispatchEvent(new Event('input', { bubbles: true })); return true`

export default async function uxrf7(p) {
  const ort = p.ort({ zuhause: true })
  const ordner = mkdtempSync(join(tmpdir(), 'uxrf7-'))
  p.mse(ort, 'oeffnen', probeset(ordner, 20))
  const bogen = join(ordner, 'bogen.pdf')
  const liste = join(ordner, 'liste.txt')
  const vorher = readdirSync(ordner)

  await p.starte({ ort })
  await p.taste('2', 'befehl')
  await p.wartetext('Wie-es-wird-Set', 20000)
  await p.klickeText('Wie-es-wird-Set')
  await p.wartetext('Karte anlegen', 20000)

  p.notiz('Der Druckbogen — der Kasten steht vor dem Knopf, und nichts ist geschrieben')
  await p.klick('Ablage', 'Ausgeben', 'Druckbogen …')
  await p.warte('der Druckbogen steht', () => p.seite(`return !!${knopf('Druckbogen legen …')}`), 20000)
  const a4 = await p.warte('der Kasten nennt die Zahlen', () => p.seite(`const t = (() => { ${zahlen} })(); return t && t.includes('je Seite') ? t : false`), 20000)
  p.zeige(`Kasten: ${a4}`)
  p.pruefe('9 je Seite auf 3 Seiten', a4.includes('9 je Seite · 3 Seiten'), a4)
  p.pruefe('mit Karten, Papier und Auflösung', a4.includes('20 Karten · Papier 210 × 297 mm · 300 dpi'), a4)
  p.pruefe('und der halben letzten Seite', a4.includes('Auf der letzten Seite: 2 Karten.'), a4)
  p.pruefe('der Kasten steht vor dem Knopf', await p.seite(`return !!(${kasten}.compareDocumentPosition(${knopf('Druckbogen legen …')}) & Node.DOCUMENT_POSITION_FOLLOWING)`))
  p.pruefe('das Raster hat neun Plätze, belegt alle neun', await p.seite(`const r = (${kasten}).querySelector('[aria-hidden=true]'); return r && r.children.length === 9 && [...r.children].every((c) => /(^|\\s)bg-marke(\\s|$)/.test(c.className))`))
  p.pruefe('das Raster steht nicht im Zugänglichkeitsbaum', await p.seite(`return (${kasten}).querySelector('[aria-hidden=true]')?.getAttribute('aria-hidden') === 'true'`))
  p.pruefe('bis hierher liegt nichts auf der Platte', readdirSync(ordner).length === vorher.length && !existsSync(bogen), readdirSync(ordner).join(', '))
  await p.bild('bogen-wie-es-wird')

  p.notiz('Letter — die Zahl ändert sich, und der Bogen hat genau diese Seitenzahl')
  await p.seite(setzeWahl('Papier', 'letter'))
  const letter = await p.warte('der Kasten nennt Letter', () => p.seite(`const t = (() => { ${zahlen} })(); return t && /215[,.]9 × 279[,.]4/.test(t) ? t : false`), 20000)
  p.zeige(`Kasten: ${letter}`)
  const seiten = Number(/je Seite · (\d+) Seiten/.exec(letter)?.[1])
  p.pruefe('Letter nennt andere Zahlen und 215,9 × 279,4 mm (Komma oder Punkt nach der Darstellung)', !letter.includes('9 je Seite') && /215[,.]9 × 279[,.]4 mm/.test(letter), letter)
  await p.bild('bogen-letter')
  await p.dialogantwort(bogen)
  await p.klickeAuf(knopf('Druckbogen legen …'), 'Druckbogen legen …')
  const gelegt = await p.warte('der Bogen ist gelegt', () => p.seite(status('20 Karten')), 120000)
  p.pruefe('der Bogen liegt da', existsSync(bogen) && statSync(bogen).size > 0, gelegt.split('\n')[0])
  p.pruefe('mit genau der Seitenzahl, die der Kasten nannte', gelegt.startsWith(`20 Karten auf ${String(seiten)} Seiten`), `Kasten ${String(seiten)}, Satz: ${gelegt.split('\n')[0]}`)
  await p.bild('bogen-gelegt')

  p.notiz('Die Exportvorlage — Dateiname und -größe vor dem Schreiben')
  await p.klick('Ablage', 'Ausgeben', 'Über eine Exportvorlage …')
  await p.klickeAuf("document.querySelector('input[type=radio][value=vorlage]')", 'das Ziel Exportvorlage')
  const roh = await p.warte('der Kasten nennt die Datei', () => p.seite(`const t = (() => { ${zahlen} })(); return t && t.includes('Datei:') ? t : false`), 30000)
  p.zeige(`Kasten: ${roh}`)
  const vorGroesse = Number(/Datei: [^\s]+ \((\d+) B\)/.exec(roh)?.[1])
  p.pruefe('er nennt Name und Größe', /Datei: ausfuhr\.txt \(\d+ B\)/.test(roh), roh)
  p.pruefe('und die Karten, über die die Vorlage lief', roh.includes('Ausgewertet über 20 Karten.'), roh)
  p.pruefe('bis hierher liegt keine Liste da', !existsSync(liste))
  await p.bild('vorlage-wie-es-wird')

  p.notiz('Eine Option ändern — die Größe folgt, und die Datei danach ist genau so groß')
  await p.seite(setzeText('Was über der Liste steht', 'Meine Karten und noch ein langer Zusatz'))
  const nachher = await p.warte('die Größe ist gewachsen', () => p.seite(`const t = (() => { ${zahlen} })(); const m = t && /Datei: [^\\s]+ \\((\\d+) B\\)/.exec(t); return m && Number(m[1]) > ${String(vorGroesse)} ? t : false`), 30000)
  const groesse = Number(/Datei: [^\s]+ \((\d+) B\)/.exec(nachher)?.[1])
  p.pruefe('sie wuchs um die Länge des Zusatzes', groesse - vorGroesse === 'Meine Karten und noch ein langer Zusatz'.length - 'Meine Karten'.length, `${String(vorGroesse)} → ${String(groesse)}`)
  await p.dialogantwort(liste)
  await p.klickeAuf(knopf('Ausführen'), 'Ausführen')
  await p.warte('die Liste liegt da', () => Promise.resolve(existsSync(liste)), 60000)
  p.pruefe('die Datei ist genau so groß, wie der Kasten sagte', statSync(liste).size === groesse, `${String(statSync(liste).size)} Bytes, Kasten ${String(groesse)}`)
  await p.bild('vorlage-geschrieben')

  p.pruefe('⌘Q beendet', await p.beende())
}
