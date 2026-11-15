// **Der erste Zufallspack mit einem großen Set, am nativen Fenster** —
// #2466.
//
// *Blue* (3.693 Karten, kanonisches Set) mit dem vollen Pack: gemessen wird
// vom Wechsel auf „Zufallspack“ bis die Packarten zur Wahl stehen. Am
// 2026-10-17 waren es mit `readmebilder` 98 bis 104 s; an der Kommandozeile
// am 2026-10-19 80 s vor und 15,5 s nach der Änderung (`mse packarten`).
//
// **Was sie nicht belegt**: einen Mac mit weniger Kernen — die Zahl der
// Arbeiter folgt den Kernen, wie bei der Problemprüfung.
import { homedir } from 'node:os'
import { join } from 'node:path'

import { mitPack } from './budgetsorte09.mjs'
import { oeffnen, zurSeite } from './uirevue.mjs'

const kanon = join(homedir(), 'Library/Application Support/MagicSetEditor4Mac/kanonische-sets')
const laedt = "return [...document.querySelectorAll('[role=dialog] *, main *')].some((e) => e.checkVisibility() && e.childElementCount === 0 && e.textContent.trim().startsWith('Wird geladen'))"
const packarten = "return [...document.querySelectorAll('select option')].filter((o) => o.closest('select')?.checkVisibility()).length"

export default async function zufallspack09(p) {
  const ort = p.ort()
  mitPack(p, ort)
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  p.zeige(p.mse(ort, 'oeffnen', join(kanon, 'Blue.mse-set')).split('\n')[0])

  await oeffnen(p, ort, { name: '1440 × 900', breite: 1440, hoehe: 900 })

  p.notiz('Der erste Aufruf von »Zufallspack« mit Blue')
  const beginn = Date.now()
  await zurSeite(p, 'Zufallspack')
  await p.warte('die Packarten stehen zur Wahl', async () => !(await p.seite(laedt)) && (await p.seite(packarten)) > 0, 180000)
  const dauer = Date.now() - beginn
  p.zeige(`erster Aufruf bis zu den Packarten: ${String(dauer)} ms, ${String(await p.seite(packarten))} Packarten zur Wahl`)
  p.pruefe('der erste Aufruf braucht weniger als 30 s (vorher rund 100 s)', dauer < 30000, `${String(dauer)} ms`)
  await p.bild('zufallspack-blue')
  p.pruefe('⌘Q beendet', await p.beende())
}
