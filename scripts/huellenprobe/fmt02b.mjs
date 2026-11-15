// **Befund und Konflikt in der Werkstatt — am nativen Fenster** — FMT-02.AK-02,
// AK-03, S10.NAT-02 und die erste Hälfte von S10.NAT-03
// (#1691):
//
// - die Kopie einer mitgelieferten Vorlage hat **keinen** Befund — bis FMT-02
//   meldete die Prüfung dort einen Skriptfehler in `hell: hell.png`,
// - eine Feldart, die MSE2 nicht kennt, steht **unter dem Editor** mit Zeile
//   und Spalte, auch bei schmalem Fenster (S10.NAT-02),
// - die Korrektur nimmt den Befund weg (S10.NAT-03, erste Hälfte),
// - ändert jemand die Datei auf der Platte, meldet Sichern den Konflikt, und
//   „Meine behalten" sichert den eigenen Text gegen die fremde Fassung.
//
// **Ein eigenes Zuhause**: die Kopie landet im verwalteten Paketverzeichnis,
// und die fremde Änderung geht an diese Kopie.
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))`
const zweig = (name) => `[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))?.closest('li')`
const feld = "document.querySelector('textarea')"
const befund = "[...document.querySelectorAll('[role=status]')].find((e) => e.checkVisibility() && e.innerText.includes('Zeile '))"

// imBild sagt, ob ein Element ganz im sichtbaren Bereich steht — **gemessen am
// Fenster**, nicht am DOM: ein Knopf unter dem unteren Rand ist da und
// trotzdem nicht zu sehen.
const imBild = (ausdruck) => `const r = (${ausdruck}).getBoundingClientRect(); return r.width > 0 && r.top >= 0 && r.bottom <= window.innerHeight && r.right <= window.innerWidth`

// einfuegen setzt den ganzen Text, wie ein Einfügen es täte — Tabulatoren
// ließen sich über die Tastatur nicht tippen, sie wechselten den Fokus.
async function einfuegen(p, text) {
  await p.seite(`const t = ${feld}; const setze = Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype, 'value').set; setze.call(t, ${JSON.stringify(text)}); t.dispatchEvent(new Event('input', { bubbles: true })); return true`)
}

async function zurWerkstatt(p) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld('Vorlagen')
  await p.klickeAuf("[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Vorlagen')", '»Vorlagen« in der Palette')
  await p.wartetext('probe-schlicht.mse-style', 20000)
}

export default async function fmt02b(p) {
  const ort = p.ort({ zuhause: true })

  await p.starte({ ort })
  await p.groesse(1280, 820)

  p.notiz('Eine Kopie der mitgelieferten Vorlage — ohne Befund')
  await zurWerkstatt(p)
  await p.warte('der Zweig sagt »nur lesbar«', () => p.seite(`return (${zweig('probe-schlicht.mse-style')})?.innerText.includes('nur lesbar') ?? false`), 15000)
  await p.klickeAuf(`[...(${zweig('probe-schlicht.mse-style')})?.querySelectorAll('button') ?? []].find((b) => b.innerText.trim().endsWith('Eigene Kopie anlegen'))`, 'Eigene Kopie anlegen unter probe-schlicht')
  await p.wartetext('Das Original bleibt.', 15000)
  await p.warte('der Editor steht in der Kopie', () => p.seite(`return !!(${feld})`), 10000)
  await p.wartetext('Kein Befund in dieser Datei.', 15000)
  p.pruefe('die Kopie der Saat hat keinen Befund', !(await p.seite(`return !!(${befund})`)))
  const kopie = JSON.parse(p.mse(ort, 'pakete', '--json')).pakete.find((paket) => paket.name === 'probe-schlicht-kopie')
  const datei = join(kopie.pfad, 'style')
  p.pruefe('die Datei der Kopie liegt auf der Platte', existsSync(datei), datei)
  const gelesen = await p.seite(`return (${feld}).value`)
  await p.bild('ohne-befund')

  p.notiz('S10.NAT-02 — eine unbekannte Feldart, bei schmalem Fenster')
  await p.groesse(900, 820)
  await einfuegen(p, `${gelesen}\nstyling field:\n\ttype: bild\n\tname: probe\n`)
  await p.warte('der Befund steht unter dem Editor', () => p.seite(`return !!(${befund})`), 15000)
  const satz = await p.seite(`return (${befund}).innerText`)
  p.pruefe('er nennt die Art und die bekannten', satz.includes('»bild«') && satz.includes('package choice'), satz.slice(0, 160))
  p.pruefe('er steht ganz im sichtbaren Bereich', await p.seite(imBild(befund)))
  await p.bild('feldart-schmal')

  p.notiz('S10.NAT-03 — die Korrektur nimmt den Befund weg')
  await einfuegen(p, `${gelesen}\nstyling field:\n\ttype: image\n\tname: probe\n`)
  await p.wartetext('Kein Befund in dieser Datei.', 15000)
  p.pruefe('nach der Korrektur steht kein Befund mehr da', !(await p.seite(`return !!(${befund})`)))
  await p.groesse(1280, 820)

  p.notiz('AK-03 — jemand ändert die Datei, „Meine behalten" sichert den eigenen Text')
  await einfuegen(p, `${gelesen}# meins\n`)
  writeFileSync(datei, `${readFileSync(datei, 'utf8')}# fremd\n`)
  await p.klickeAuf(knopf('Sichern'), 'Sichern')
  await p.wartetext('Die Datei hat sich seit dem Aufschlagen geändert.', 15000)
  p.pruefe('der Konflikt bietet beide Wege', await p.seite(`return !!(${knopf('Ihre Fassung laden')}) && !!(${knopf('Meine behalten')})`))
  p.pruefe('beide stehen im sichtbaren Bereich', await p.seite(imBild(knopf('Ihre Fassung laden'))) && await p.seite(imBild(knopf('Meine behalten'))))
  await p.bild('konflikt')
  await p.klickeAuf(knopf('Meine behalten'), 'Meine behalten')
  await p.warte('der Konflikt ist fort', async () => !(await p.seite("return document.body.innerText.includes('Die Datei hat sich seit dem Aufschlagen geändert.')")), 10000)
  p.pruefe('der eigene Text steht noch da', await p.seite(`return (${feld}).value.endsWith('# meins\\n')`))
  await p.klickeAuf(knopf('Sichern'), 'Sichern')
  await p.wartetext('Gesichert.', 10000)
  const danach = readFileSync(datei, 'utf8')
  p.pruefe('auf der Platte steht der eigene Text', danach.endsWith('# meins\n') && !danach.includes('# fremd'), danach.slice(-40))
  await p.bild('behalten')

  await p.beende()
}
