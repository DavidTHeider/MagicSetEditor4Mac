// **Apples Gerätemodell am nativen Fenster** — AI-05.AK-09
// (#2042):
//
// - am Kartenformular steht an **Name** ein „Vorschlagen", und ein Druck holt
//   einen Vorschlag vom **echten** Gerätemodell dieses Macs — mit Vorrang,
//   ohne dass jemand etwas einstellt (AK-03),
// - „Übernehmen" schreibt ihn ins Feld (AK-06),
// - unter *Einstellungen › KI* steht der Sitzungsschalter; aus beendet das
//   laufende `fm serve` **sofort** (AK-04) — gemessen am Prozess, nicht am
//   Quelltext.
//
// **Braucht macOS 27 mit aktiviertem Modell.** Fehlt es, endet der Lauf rot
// mit BLOCKED und nicht grün.
import { execFileSync } from 'node:child_process'

import { oeffnen, zurSeite } from './uirevue.mjs'

// fmLaeuft nennt die Kennungen der laufenden `fm serve` dieses Rechners.
function fmLaeuft() {
  try {
    return execFileSync('/usr/bin/pgrep', ['-f', '/usr/bin/fm serve --socket'], { encoding: 'utf8' }).trim().split('\n').filter(Boolean)
  } catch {
    return []
  }
}

const namensfeld = "document.querySelector('input[aria-label=\"name\"], textarea[aria-label=\"name\"]')"
const vorschlagsknopf = `[...${namensfeld}.closest('label').querySelectorAll('button')].find((b) => b.innerText.trim() === 'Vorschlagen')`

export default async function ai05(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  const vorher = fmLaeuft()
  p.zeige(`fm serve vor dem Lauf: ${vorher.length === 0 ? 'keiner' : vorher.join(', ')}`)

  await oeffnen(p, ort, { name: 'normal', breite: 1280, hoehe: 800 })
  await zurSeite(p, 'Karten')
  await p.warte('das Namensfeld steht da', () => p.seite(`return !!${namensfeld}`), 20000)

  p.notiz('AK-03 und AK-06 — ein Vorschlag am Feld Name, mit Vorrang')
  p.pruefe('der Knopf steht am Feld Name', await p.seite(`return !!${vorschlagsknopf}`))
  await p.klickeAuf(vorschlagsknopf, '»Vorschlagen« am Namen')
  // **Gewartet wird auf eine Wirkung — Vorschlag, Ablehnung oder Fehler** —
  // und nicht nur auf den Erfolg: sonst stünde nach 60 s „nicht da" statt
  // dessen, was das Fenster sagt.
  const ausgang = await p.warte('am Feld steht eine Antwort', () =>
    p.seite(`const feld = ${namensfeld}.closest('label'); const von = [...feld.querySelectorAll('span')].map((s) => s.innerText.trim()).find((t) => t.startsWith('von ')); if (von) return von; const meldung = feld.querySelector('[role=alert], [role=status]'); return meldung ? 'Meldung: ' + meldung.innerText.trim() : ''`), 60000)
  p.zeige(`am Feld: ${ausgang}`)
  const herkunft = ausgang
  p.pruefe('er kam vom Gerätemodell, mit Vorrang', herkunft === 'von apple/system · Vorrang', herkunft)
  const laufend = fmLaeuft()
  p.pruefe('ein fm serve läuft — bei Bedarf gestartet', laufend.length > vorher.length, laufend.join(', '))
  await p.bild('ai05-vorschlag')

  const zeile = "[...document.querySelectorAll('span')].find((s) => s.innerText.trim().startsWith('von apple/system')).parentElement"
  const wert = await p.seite(`return ${zeile}.querySelector('span')?.innerText.trim() ?? ''`)
  await p.klickeAuf(`[...${zeile}.querySelectorAll('button')].find((b) => b.innerText.trim() === 'Übernehmen')`, '»Übernehmen« am Vorschlag')
  await new Promise((los) => setTimeout(los, 3000))
  await p.bild('ai05-nach-uebernehmen')
  p.zeige(`im Bestand: ${JSON.parse(p.mse(ort, 'karten', '1', '--json')).karten.map((k) => k.name).join(', ')}`)
  p.zeige(`nach dem Druck: Feld »${await p.seite(`return ${namensfeld}.value`)}«, Meldungen: ${await p.seite("return [...document.querySelectorAll('[role=alert]')].map((m) => m.innerText.trim()).filter(Boolean).join(' | ')")}`)
  const nachher = await p.warte('am Feld steht der Name oder eine Meldung', () => p.seite(`const feld = ${namensfeld}; if (feld.value.includes(${JSON.stringify(wert)})) return 'Wert: ' + feld.value; const m = feld.closest('label').querySelector('[role=alert]'); return m ? 'Meldung: ' + m.innerText.trim() : ''`), 20000)
  p.pruefe('übernommen', nachher === `Wert: ${wert}`, nachher)

  p.notiz('AK-04 — für die Sitzung abschalten, am Prozess gemessen')
  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', () => p.seite("return !!document.querySelector('[role=tab]')"), 20000)
  await p.klickeText('KI')
  // **Seit #2465 stehen die Modelle in einem Abschnitt** — erst aufklappen.
  await p.klickeText('KI-Modelle')
  const schalter = "[...document.querySelectorAll('[role=switch]')].find((s) => s.closest('label')?.innerText.includes('In dieser Sitzung fragen'))"
  await p.warte('der Sitzungsschalter steht da', () => p.seite(`return !!${schalter}`), 30000)
  p.pruefe('er steht auf an', await p.seite(`return ${schalter}.checked`))
  // **Das Bild zeigt die Zeile des Geräts**, nicht den Kopf der Seite.
  await p.seite(`${schalter}.scrollIntoView({ block: 'center', behavior: 'instant' }); return true`)
  await p.bild('ai05-einstellungen-an')
  await p.klickeAuf(schalter, 'den Sitzungsschalter')
  await p.warte('er steht auf aus', () => p.seite(`return ${schalter}?.checked === false`), 20000)
  await p.warte('fm serve ist beendet', async () => fmLaeuft().length === vorher.length, 10000)
  p.pruefe('kein fm serve mehr', fmLaeuft().length === vorher.length, fmLaeuft().join(', '))
  await p.bild('ai05-einstellungen-aus')

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
  p.pruefe('und mit ihr jedes fm serve', fmLaeuft().length === vorher.length, fmLaeuft().join(', '))

  // **Und dasselbe im Dunkeln** — die Sichtprüfung verlangt beide
  // Erscheinungsbilder; der Sitzungsschalter ist nach dem Neustart wieder an.
  p.notiz('Sichtprüfung — dunkel, nach dem Neustart')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'dunkel')
  await oeffnen(p, ort, { name: 'dunkel', breite: 1280, hoehe: 800 })
  await zurSeite(p, 'Karten')
  await p.warte('das Namensfeld steht da', () => p.seite(`return !!${namensfeld}`), 20000)
  await p.klickeAuf(vorschlagsknopf, '»Vorschlagen« am Namen')
  const dunkel = await p.warte('am Feld steht eine Antwort', () =>
    p.seite(`return [...${namensfeld}.closest('label').querySelectorAll('span')].map((s) => s.innerText.trim()).find((t) => t.startsWith('von ')) ?? ''`), 60000)
  p.pruefe('nach dem Neustart fragt der Vorrang wieder das Gerät', dunkel === 'von apple/system · Vorrang', dunkel)
  await p.bild('ai05-vorschlag-dunkel')
  p.pruefe('⌘Q beendet die Anwendung (dunkel)', await p.beende())
  p.pruefe('und mit ihr jedes fm serve (dunkel)', fmLaeuft().length === vorher.length, fmLaeuft().join(', '))
}
