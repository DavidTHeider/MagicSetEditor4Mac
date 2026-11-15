// **Start bis zum ersten bedienbaren Kernweg** — E05
// (#2242).
//
// B02 maß den Start bis „bedienbar" (die Leiste antwortet) mit einem Bestand,
// den `mse` vorher angelegt hatte. E05 verlangt drei Dinge dazu, und dieses
// Drehbuch fährt sie, je Reihe fünf Starts:
//
//   - **Saat**: ein völlig frisches Zuhause — die Anwendung legt Bestand und
//     Pakete selbst an,
//   - **warm**: dasselbe Zuhause, das Programm zuvor mit ⌘Q beendet,
//   - **älter**: ein Bestand, den 0.7.0 geschrieben hat
//     (`MSE_E05_ALTBESTAND`, mit dem `mse` aus dem signierten Abbild angelegt).
//
// Je Start: **bedienbar** (wie `budgets.mjs`, alle 20 ms gefragt), dann die
// **erste Eingabe** — ⌘K, bis die Palette steht — und der **erste
// Kernweg** — ⌘2, bis die Setliste steht. Die erste Eingabe zeigt, ob ein
// Vorwärmen sie blockiert.
//
// **Was er nicht belegt**: einen kalten Dateicache des Systems (`purge`
// braucht ein Kennwort) und den ersten Start nach einer Installation, den
// Gatekeeper prüft — den zeigt der Gastlauf (`budgets-gast.md`).
import { copyFileSync, existsSync } from 'node:fs'

import { kennzahlen, paletteOffen } from './budgets.mjs'

const setliste = "return document.querySelector('main')?.dataset.seite === 'sets' && document.querySelectorAll('main button').length > 0"

// kernweg misst nach dem Start die erste Eingabe und den ersten Weg.
async function kernweg(p) {
  const a = Date.now()
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite(paletteOffen), 15000, 10)
  const eingabe = Date.now() - a
  await p.taste('Escape')
  await p.warte('die Palette ist zu', async () => !(await p.seite(paletteOffen)), 10000)
  const b = Date.now()
  await p.taste('2', 'befehl')
  await p.warte('die Setliste steht', () => p.seite(setliste), 15000, 10)
  return { eingabe, weg: Date.now() - b }
}

export default async function e05start(p) {
  const alt = process.env.MSE_E05_ALTBESTAND
  const reihen = { saat: [], warm: [], älter: [] }
  const zeile = async (reihe, ort) => {
    const beginn = Date.now()
    await p.starte({ ort, takt: 20 })
    const bereit = p.bereitUm - beginn
    const { eingabe, weg } = await kernweg(p)
    reihen[reihe].push({ bereit, eingabe, weg })
    p.pruefe(`${reihe} ${String(reihen[reihe].length)}: ⌘Q beendet`, await p.beende())
  }

  p.notiz('Saat — jedes Mal ein frisches Zuhause')
  for (let i = 0; i < 5; i++) await zeile('saat', p.ort())

  p.notiz('warm — dasselbe Zuhause')
  const warm = p.ort()
  await p.starte({ ort: warm, takt: 20 })
  p.pruefe('erster Start des warmen Zuhauses: ⌘Q beendet', await p.beende())
  for (let i = 0; i < 5; i++) await zeile('warm', warm)

  p.notiz('älter — ein Bestand von 0.7.0')
  if (alt && existsSync(alt)) {
    for (let i = 0; i < 5; i++) {
      const ort = p.ort()
      copyFileSync(alt, ort.bestand)
      await zeile('älter', ort)
    }
  } else {
    p.zeige('nicht gemessen: MSE_E05_ALTBESTAND nennt keinen Bestand von 0.7.0')
  }

  p.notiz('Die Reihen')
  p.zeilen.push('| Reihe | Starts | bedienbar: Median · max | erste Eingabe (⌘K): Median · max | Kernweg (⌘2): Median · max |', '| --- | ---: | ---: | ---: | ---: |')
  for (const [reihe, werte] of Object.entries(reihen)) {
    if (werte.length === 0) continue
    const k = (feld) => kennzahlen(werte.map((w) => w[feld]))
    p.zeilen.push(`| ${reihe} | ${String(werte.length)} | ${String(k('bereit').median)} · ${String(k('bereit').max)} ms | ${String(k('eingabe').median)} · ${String(k('eingabe').max)} ms | ${String(k('weg').median)} · ${String(k('weg').max)} ms |`)
  }
  p.zeilen.push('', `Rohwerte — ${JSON.stringify(reihen)}`)
  const gemessen = Object.values(reihen).filter((w) => w.length > 0)
  p.pruefe('bedienbar im Median ≤ 2 s, in jeder Reihe', gemessen.every((w) => kennzahlen(w.map((x) => x.bereit)).median <= 2000))
  p.pruefe('die erste Eingabe wartet auf kein Vorwärmen (≤ 500 ms, jeder Start)', gemessen.every((w) => w.every((x) => x.eingabe <= 500)), JSON.stringify(gemessen.flat().map((x) => x.eingabe)))
}
