// **Den Neuanfang am Fenster freigeben und ausführen** — G12
// (#2282),
// am nativen Fenster:
//
//   1. *Einstellungen › Daten und Sicherung*, Abschnitt *Zurücksetzen*: „Die
//      ganze Anwendung" wählen und nachsehen — **kein Feld**, in das jemand
//      eine Kennung tippen müsste, und *Zurücksetzen* ruht,
//   2. *Freigabe erteilen*: „Freigegeben bis …" und darunter die Kennung für
//      die Kommandozeile; *Zurücksetzen* ist frei,
//   3. *Zurücksetzen*: die Bilanz nennt, wo der bisherige Bestand liegt — und
//      dort liegt er.
//
// **Der Weg vorher** steht im Befund des Issues: die Seite fragte nach einer
// Kennung, die keine Fläche ausstellte, und der Neuanfang ging nirgends.
// **Gefahren in einem eigenen Zuhause** — beiseitegelegt wird der Bestand der
// Probe und kein anderer.
import { existsSync } from 'node:fs'

const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(text)})`

// waehleAnwendung stellt die Art über das native Auswahlfeld, dessen Wert
// React liest — derselbe Griff wie in `f01.mjs`.
const waehleAnwendung = "const s = [...document.querySelectorAll('select')].find((e) => e.checkVisibility() && [...e.options].some((o) => o.value === 'anwendung')); if (!s) return false; s.scrollIntoView({ block: 'center' }); Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, 'anwendung'); s.dispatchEvent(new Event('change', { bubbles: true })); return s.value === 'anwendung'"

export default async function g12(p) {
  const ort = p.ort({ zuhause: true })
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  await p.starte({ ort })

  p.notiz('1 — Die ganze Anwendung, nachgesehen: kein Tippfeld, der Knopf ruht')
  await p.taste(',', 'befehl')
  await p.wartetext('Bestand, Pfade, Sicherung', 20000)
  await p.klickeText('Daten und Sicherung')
  // **Seit #2465 in einem Abschnitt** — erst aufklappen.
  await p.klickeText('Zurückgehen, aufräumen, zurücksetzen')
  await p.warte('das Auswahlfeld steht', () => p.seite(waehleAnwendung), 20000)
  await p.klickeAuf(knopf('Nachsehen, was betroffen wäre'), 'Nachsehen')
  await p.warte('die Freigabe wird angeboten', () => p.seite(`return !!(${knopf('Freigabe erteilen')})`), 20000)
  p.pruefe('kein Feld für eine Kennung', await p.seite("return ![...document.querySelectorAll('input')].some((e) => e.checkVisibility() && e.labels?.[0]?.innerText.includes('Freigabe'))"))
  p.pruefe('*Zurücksetzen* ruht', await p.seite(`return ${knopf('Zurücksetzen')}?.disabled === true`))
  await p.bild('g12-vor-der-freigabe')

  p.notiz('2 — Freigabe erteilen: bis wann, und die Kennung')
  await p.klickeAuf(knopf('Freigabe erteilen'), 'Freigabe erteilen')
  const status = await p.warte('die Freigabe steht da', () => p.seite("return [...document.querySelectorAll('[role=status]')].find((e) => e.checkVisibility() && e.innerText.includes('Freigegeben bis'))?.innerText ?? null"), 20000)
  p.zeige(status)
  const kennung = await p.seite("return [...document.querySelectorAll('[role=status] code')].find((e) => e.checkVisibility())?.innerText.trim() ?? ''")
  p.pruefe('die Kennung steht für die Kommandozeile da', /^[0-9a-f]{16,}$/.test(kennung), kennung)
  p.pruefe('*Zurücksetzen* ist frei', await p.seite(`return ${knopf('Zurücksetzen')}?.disabled === false`))
  await p.bild('g12-freigegeben')

  p.notiz('3 — Zurücksetzen: der bisherige Bestand liegt beiseite')
  await p.klickeAuf(knopf('Zurücksetzen'), 'Zurücksetzen')
  const bilanz = await p.warte('die Bilanz nennt den Ort', () => p.seite("return [...document.querySelectorAll('p')].find((e) => e.checkVisibility() && e.innerText.startsWith('Der bisherige Bestand liegt bei'))?.innerText ?? null"), 30000)
  p.zeige(bilanz)
  const beiseite = bilanz.replace(/^Der bisherige Bestand liegt bei /, '').replace(/\.$/, '')
  p.pruefe('dort liegt er', existsSync(beiseite), beiseite)
  await p.bild('g12-bilanz')

  p.pruefe('⌘Q beendet', await p.beende())
}
