// **Phase G am nativen Fenster: U09, U14, U15** —
// #2157,
// #2159,
// #2160.
//
// Eine Sammlung von 260 Exemplaren mit Mengen 1 bis 3 — über dem ersten
// Fenster von 200 Zeilen. Gemessen wird:
//
// - **U15**: der Überblick der Startseite nennt die Summe der Mengen.
// - **U09**: wer in der Sammlung ans Ende springt, sieht dort echte Zeilen —
//   das nachgeholte Fenster kommt an.
// - **U14**: die Paketseite zeigt den Abschnitt „Abhängigkeiten" aus
//   derselben Antwort, die die Zeilen zählt.
//
// **Was sie nicht belegt**: die Zahl der Rufe — die zählen die Fenstertests
// (`sammlungsunterwegs.test.tsx`, `abhaengigkeitsquelle.test.tsx`); und nur
// 1280×800, hell, Deutsch.
import { randomUUID } from 'node:crypto'

import { oeffnen, zurSeite } from './uirevue.mjs'

const ueberblick = "[...document.querySelectorAll('main ul button')].map((b) => b.innerText.replace(/\\s+/g, ' ').trim())"

export default async function phaseg2(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  let summe = 0
  for (let stelle = 0; stelle < 260; stelle++) {
    const menge = (stelle % 3) + 1
    summe += menge
    p.mse(ort, 'exemplar', randomUUID(), String(menge))
  }
  p.zeige(`260 Exemplare, ${String(summe)} Stück`)

  await oeffnen(p, ort, { name: 'normal', breite: 1280, hoehe: 800 })

  p.notiz('U15 — der Überblick zählt die Stücke')
  await p.taste('1', 'befehl')
  await p.warte('der Überblick nennt die Summe', async () => {
    const zahlen = await p.seite(`return ${ueberblick}`)
    return zahlen.some((z) => z.startsWith(`${String(summe)} `))
  }, 20000)
  p.pruefe('der Überblick nennt die Summe der Mengen', true, String(summe))
  await p.bild('g2-u15-start')

  p.notiz('U09 — ans Ende der Sammlung springen')
  await p.taste('3', 'befehl')
  await p.warte('das Gitter der Sammlung steht', () => p.seite("return !!document.querySelector('[role=grid][aria-rowcount]')"), 20000)
  const zeilen = await p.seite("return document.querySelector('[role=grid][aria-rowcount]').getAttribute('aria-rowcount')")
  p.zeige(`aria-rowcount ${zeilen}`)
  await p.seite("const g = document.querySelector('[role=grid][aria-rowcount]'); let e = g; while (e && e.scrollHeight <= e.clientHeight) e = e.parentElement; (e ?? g).scrollTop = 1e9; return true")
  await p.warte('am Ende stehen echte Zeilen', () => p.seite(`const z = [...document.querySelectorAll('[role=row][aria-rowindex]')].filter((r) => Number(r.getAttribute('aria-rowindex')) > 230); return z.length > 0 && z.every((r) => r.innerText.trim().length > 0)`), 20000)
  p.pruefe('das nachgeholte Fenster ist angekommen', true)
  await p.bild('g2-u09-ende')

  p.notiz('U14 — die Paketseite und ihr Abschnitt')
  await zurSeite(p, 'Pakete')
  await p.klickeText('Abhängigkeiten')
  await p.warte('der Abschnitt antwortet', () => p.seite("return [...document.querySelectorAll('section')].some((s) => s.innerText.includes('Jede Abhängigkeit ist erfüllt.') || s.querySelector('li'))"), 20000)
  p.pruefe('der Abschnitt steht mit seiner Antwort', true)
  await p.bild('g2-u14-pakete')

  p.pruefe('⌘Q beendet', await p.beende())
}
