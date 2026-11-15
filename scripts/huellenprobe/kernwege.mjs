// **Die übrigen Kernwege aus ENG-02.AK-04** — `docs/WAILS3.md`, „Die
// Kernwege am Fenster"
// (#1727).
//
// Die Wege 1 (Start, ⌘Q, „wie zuletzt"), 2 (die Leiste) und 3 (⌘O, Sichern
// unter) stehen in `start`, `leiste` und `fix-f02`, Weg 7 (der Tisch) in
// `fix-f04`. Hier:
//
// - Weg 4: ein `.mse-set` aus dem Finder — bei laufender und bei
//   geschlossener Anwendung, über LaunchServices wie ein Doppelklick,
// - Weg 5: ein Bild auf ein Bildfeld ziehen, ein Set ins Fenster ziehen,
// - Weg 8: Startfehler und Wiederherstellung mit einem unbrauchbaren
//   Bestand.
//
// **Was hier nicht steht**: Weg 6 (Karten kopieren und in ein anderes Set
// einsetzen) fährt `kopieren` — über die Ersatzablage des Abnahmebaus, denn
// die Zwischenablage dieses Rechners fasst die Probe nicht an —, Weg 9
// (Profilwechsel mit zweitem Faktor) fährt `profilwechsel`.
import { execFileSync } from 'node:child_process'
import { mkdirSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

const wurzel = resolve(import.meta.dirname, '..', '..')
const buendel = join(wurzel, 'build/bin/MagicSetEditor4Mac.app')
const setImFinder = join(wurzel, 'saat/probe-werkstatt.mse-set')

const setsImBestand = (p, ort) => p.mse(ort, 'sets')
const zahlDerSets = (p, ort) => setsImBestand(p, ort).trim().split('\n').filter(Boolean).length

export default async function kernwege(p) {
  p.notiz('Weg 4 — ein Set aus dem Finder, bei geschlossener Anwendung')
  const ort = p.ort()
  await p.starte({ ort, datei: setImFinder })
  await p.warte('das Set ist offen', () => setsImBestand(p, ort).includes('Werkstatt'), 20000)
  p.pruefe('der Doppelklick bei geschlossener Anwendung öffnet das Set', true, setsImBestand(p, ort).trim().split('\n').slice(0, 2).join(' | '))

  p.notiz('Weg 5 — ein Bild auf ein Bildfeld ziehen')
  const ablage = join(ort.ort, 'ablage')
  mkdirSync(ablage)
  const bild = join(ablage, 'gezogen.png')
  execFileSync('cp', [join(wurzel, 'build/appicon.png'), bild])
  await p.klickeText('Werkstatt-Testkarten').catch(() => undefined)
  const feld = await p.warte('ein Bildfeld steht da', () => p.seite("const f = [...document.querySelectorAll('[data-bildfeld]')].find((e) => e.checkVisibility()); if (!f) return false; const r = f.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2, feld: f.dataset.bildfeld }"), 20000)
  await p.frage({ was: 'ablegen', pfad: [bild], x: feld.x, y: feld.y })
  await p.wartetext('Einsetzen', 15000)
  p.pruefe(`ein Bild auf dem Feld »${feld.feld}« führt zum Zuschnitt`, true)
  await p.klickeText('Abbrechen')

  p.notiz('Weg 4 — dasselbe Set noch einmal aus dem Finder, bei laufender Anwendung')
  const zweites = join(ablage, 'zweites.mse-set')
  execFileSync('cp', ['-R', setImFinder, zweites])
  execFileSync('open', ['-a', buendel, zweites])
  await p.warte('das zweite Set kommt an', () => zahlDerSets(p, ort) >= 2, 20000)
  p.pruefe('der Doppelklick bei laufender Anwendung öffnet es im selben Fenster', true, `${String(zahlDerSets(p, ort))} Sets`)

  p.notiz('Weg 5 — ein Set ins Fenster ziehen')
  const drittes = join(ablage, 'drittes.mse-set')
  execFileSync('cp', ['-R', setImFinder, drittes])
  const mitte = await p.seite('return { x: innerWidth / 2, y: innerHeight / 2 }')
  await p.frage({ was: 'ablegen', pfad: [drittes], x: mitte.x, y: mitte.y })
  await p.warte('das gezogene Set kommt an', () => zahlDerSets(p, ort) >= 3, 20000)
  p.pruefe('ein ins Fenster gezogenes Set wird geöffnet', true)
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('Weg 8 — ein unbrauchbarer Bestand')
  const kaputt = p.ort()
  writeFileSync(kaputt.bestand, 'kein SQLite, sondern Text')
  await p.starte({ ort: kaputt })
  const lage = await p.warte('das Fenster öffnet zur Wiederherstellung', () => p.seite("return /Sicherung|wiederherstell/i.test(document.body.innerText) ? document.querySelector('h1,h2')?.innerText ?? 'ja' : false"), 20000)
  p.pruefe('der Start scheitert nicht still, sondern zeigt den Weg zur Sicherung', true, lage)
  await p.bild('kernwege-wiederherstellung')
  p.pruefe('⌘Q beendet', await p.beende())
}
