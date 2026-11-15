// **»Angreifer erklären« bietet nur Kreaturen an, am nativen Fenster** —
// #2390.
//
// Gefunden bei der Ausgangsmessung P0 (#2336): das Formular nannte jede
// eigene Bleibende, auch die Länder, und ließ den Kern abweisen. Angreifen
// können nur Kreaturen (CR 508.1a, Fassung 2026-09-25). **Gemessen wird,
// welche Zeilen das Formular trägt** — ein Land und zwei Kreaturen liegen
// auf dem Spielfeld, gelegt über `mse`.
//
// **Die Gegenprobe ist dasselbe Drehbuch an einem Bündel ohne die
// Korrektur** in `Kampffeld.tsx`; dort stand der Waldboden in der Liste.
//
// Der Bestand ist der aus `ausgangswerte09spielen`: selbst geschrieben und
// künstlich, `testdaten/mtgjson/tisch/TSP.json`.
import { join, resolve } from 'node:path'

import { begonnen, bisZug, heisst, knopf, schliessen, stand, zumTisch } from '../huellenspieltisch.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')

// Die Karten des Sets in der Reihenfolge, in der `mse karten 1` sie nennt.
const karte = { waldboden: '1', bergpfad: '2', moosbaer: '3', glutlaeufer: '4', funkenregen: '5', rindenwall: '6' }

export default async function angreifer09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'mtgjson-lesen', join(wurzel, 'testdaten/mtgjson/tisch/TSP.json'), 'magic')
  p.mse(ort, 'deck', 'neu', 'Waldschritt', 'magic')
  for (const [k, n] of [[karte.waldboden, 2], [karte.moosbaer, 1], [karte.glutlaeufer, 1], [karte.funkenregen, 1], [karte.rindenwall, 2]]) p.mse(ort, 'deck', 'legen', '1', `--karte=${k}`, String(n))
  p.mse(ort, 'deck', 'neu', 'Glutpfad', 'magic')
  for (const [k, n] of [[karte.bergpfad, 16], [karte.glutlaeufer, 4], [karte.rindenwall, 4]]) p.mse(ort, 'deck', 'legen', '2', `--karte=${k}`, String(n))

  const partie = begonnen(p, ort, '1', '2', false, ['Ida', 'Jo'])
  p.pruefe('die Partie steht in Idas erster Hauptphase', bisZug(p, ort, partie, (zug) => zug.aktiv === 'ada' && zug.phase === 'hauptVorKampf'))
  for (const name of ['Waldboden', 'Moosbär', 'Glutläufer']) {
    const ding = stand(p, ort, partie).objekte.find((o) => heisst(o, name) && o.zone === 'hand')
    if (ding) p.mse(ort, 'partie-befehl', partie, 'ada', `tisch-${ding.kennung}`, 'objekt.schieben', JSON.stringify({ objekt: ding.kennung, zone: 'tisch' }))
  }
  p.pruefe('Idas Zug steht bei den Angreifern', bisZug(p, ort, partie, (zug) => zug.aktiv === 'ada' && zug.schritt === 'angreifer'))

  await p.starte({ ort })
  await schliessen(p)
  await zumTisch(p, 'Ida · Jo', 'Ida')
  await p.warte('»Angreifer erklären« steht', () => p.seite(`return !!${knopf('Angreifer erklären')}`), 20000)
  await p.klickeAuf(knopf('Angreifer erklären'), '»Angreifer erklären«')
  const zeilen = "[...document.querySelectorAll('[role=group][aria-label=\"Angreifer erklären\"] select')].map((s) => s.getAttribute('aria-label') ?? '')"
  await p.warte('das Formular steht', async () => ((await p.seite(`return ${zeilen}`)) ?? []).length > 0, 10000).catch(() => false)
  const namen = (await p.seite(`return ${zeilen}`)) ?? []
  p.zeige(`Zeilen: ${JSON.stringify(namen)}`)
  p.pruefe('beide Kreaturen stehen zur Wahl', ['Moosbär', 'Glutläufer'].every((name) => namen.some((z) => z.startsWith(name))))
  p.pruefe('das Land steht nicht zur Wahl', !namen.some((z) => z.startsWith('Waldboden')))
  await p.bild('angreifer')
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
