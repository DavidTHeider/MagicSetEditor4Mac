// **Laufzeiten, Verbrauch und ein Leistungsprofil, am nativen Fenster** —
// OPS-05 (#1722):
//
// - AK-01/04: unter *Einstellungen › Diagnose* stehen Arbeitsspeicher,
//   Verbindungen zum Bestand, Zwischenspeicher, Erhebungszeitraum und
//   Aufwand — und was nicht gemessen ist,
// - AK-03: ein Leistungsprofil über eine Sekunde liegt danach neben dem
//   Bestand, und von zweien bleibt eines, weil `messung.profile` 1 sagt,
// - eine Dauer außerhalb der Grenze hält den Knopf an.
//
// **Die Dauer kommt über das Feld selbst** — gesetzt wie eine Eingabe, damit
// React sie als Änderung sieht (dieselbe Gestalt wie in `ops02.mjs`).

import { existsSync, readdirSync } from 'node:fs'
import { join } from 'node:path'

const hier = (p) => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''")

const feld = "[...document.querySelectorAll('input[type=number]')].find((f) => f.closest('label')?.innerText.includes('Dauer in Sekunden'))"
const knopf = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Profil aufzeichnen')"

async function dauer(p, wert) {
  await p.seite(`${feld}.scrollIntoView({ block: 'center' })`)
  return p.seite(`const f = ${feld}; Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(f, ${JSON.stringify(wert)}); f.dispatchEvent(new Event('input', { bubbles: true })); return f.value`)
}

const profile = (ordner) => (existsSync(ordner) ? readdirSync(ordner).filter((n) => n.endsWith('.pprof')) : [])

export default async function ops05(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'messung.profile', '1')
  await p.starte({ ort })

  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', async () => (await hier(p)) !== '', 20000)
  await p.klickeText('Diagnose')
  // **Seit #2465 in einem Abschnitt** — erst aufklappen.
  await p.klickeText('Gewirkt seit dem Start')
  await p.wartetext('Laufzeiten und Verbrauch', 20000)

  p.notiz('AK-01/04 — Zahlen mit Einheit, Zeitraum, Aufwand und was fehlt')
  const text = await p.seite('return document.body.innerText')
  // **Seit OPS02 (#2348) der Fußabdruck vorn**, „belegt“ steht nicht mehr da.
  p.pruefe('der Arbeitsspeicher steht mit Einheit da', /Arbeitsspeicher\s+[\d,.]+ (KB|MB|GB) (Fußabdruck|Go-Heap)/.test(text), (text.match(/Arbeitsspeicher[^\n]*\n?[^\n]*/) ?? [''])[0])
  p.pruefe('die Verbindungen zum Bestand stehen da', /\d+ offen, \d+ in Benutzung/.test(text))
  p.pruefe('die Zwischenspeicher stehen da', /\d+ Treffer, \d+ geholt/.test(text))
  // **Seit AP-11 (#2368) zwei Sätze** — der Aufwand nicht mehr als Anhang.
  p.pruefe('Zeitraum und Aufwand stehen da', /Erhoben von .+\. Aufwand der Erhebung: [\d.,\u202f]+ µs/.test(text))
  p.pruefe('was nicht gemessen ist, steht da', text.includes('Nicht gemessen: die Bindungen des Fensters'))
  p.pruefe('die Spur ist aus und sagt, wo sie einzuschalten ist', text.includes('Die Ablaufspur ist aus.'))
  await p.seite("[...document.querySelectorAll('h3')].find((h) => h.innerText.includes('Laufzeiten und Verbrauch')).scrollIntoView({ block: 'start' })")
  await p.bild('betriebszahlen')

  p.notiz('Eine Dauer außerhalb der Grenze hält den Knopf an')
  // **Seit UXR-19 (#1927) in einem Abschnitt** — erst aufklappen.
  await p.klickeText('Messung')
  await p.warte('das Feld der Dauer steht da', () => p.seite(`return !!(${feld})`))
  await dauer(p, '0')
  await p.warte('das Feld ist ungültig', () => p.seite(`return ${feld}.getAttribute('aria-invalid') === 'true'`))
  p.pruefe('der Knopf hält an', await p.seite(`return ${knopf}.disabled`))

  p.notiz('AK-03 — zwei Profile über je eine Sekunde, eines bleibt')
  const ordner = join(ort.ort, 'Leistungsprofile')
  const gesehen = []
  for (const lauf of [1, 2]) {
    await dauer(p, '1')
    await p.warte('der Knopf ist frei', () => p.seite(`return !${knopf}.disabled`))
    await p.klickeAuf(knopf, '»Profil aufzeichnen«')
    // **Gewartet wird auf ein neues Profil**, nicht auf irgendeines — das
    // zweite fände sonst das erste schon vor.
    const neu = await p.warte(`Profil ${String(lauf)} liegt auf der Platte`, () => profile(ordner).find((n) => !gesehen.includes(n)) ?? false, 20000)
    gesehen.push(neu)
    await p.warte(`Profil ${String(lauf)} steht im Fenster`, () => p.seite(`return document.body.innerText.includes(${JSON.stringify(neu)})`), 10000)
  }

  const liegen = profile(ordner)
  p.pruefe('von zweien bleibt eines', liegen.length === 1, liegen.join(', '))
  await p.seite(`${knopf}.scrollIntoView({ block: 'center' })`)
  await p.bild('leistungsprofil')

  p.pruefe('⌘Q beendet', await p.beende())
}
