// **Was gerade läuft, am nativen Fenster** — TRADE-01
// (#1715):
//
// - AK-01/02: zwei Handel an der Kommandozeile — ein Entwurf und einer,
//   dem dieses Profil schon zugestimmt hat. Die Liste im Bereich *Handeln*
//   nennt beide, mit „Du bist dran: anbieten" und „Wartet auf …",
// - AK-02: ein Klick schlägt Gegenüber und Vorgang auf,
// - AK-03: der Entwurf lässt sich erst nach der Rückfrage löschen, und
//   danach steht nur noch der andere in der Liste — seit UXR-15 (#1923)
//   über das Menü „Mit diesem Handel".
//
// **Die Positionen kommen von der Standardeingabe** (`mse handel
// eroeffnen`), und `p.mse` hat keine — das Drehbuch ruft `mse` hier selbst,
// mit demselben Bestand und demselben Bund.
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { kindumgebung } from '../huellenorte.mjs'

const wurzel = fileURLToPath(new URL('../..', import.meta.url))

const positionen = JSON.stringify([
  { von: 'bgegenueber', printing: 'bolt-m10', menge: 1, guete: 'nearMint', sprache: 'de', finish: 'nonfoil' },
])

// eroeffne legt einen Handel an und gibt seine Kennung.
function eroeffne(ort, gegenueber) {
  const aus = execFileSync(`${wurzel}.go/bin/mse`, ['handel', 'eroeffnen', '', 'tausch', gegenueber], {
    cwd: wurzel,
    input: positionen.replaceAll('bgegenueber', gegenueber),
    env: kindumgebung(ort),
  }).toString()

  return aus.split('\t')[0]
}

const liste = "[...document.querySelectorAll('section')].find((s) => s.querySelector('h2')?.innerText.startsWith('Was gerade läuft'))"
const zeilen = `return [...(${liste})?.querySelectorAll('li button') ?? []].map((b) => b.innerText.replace(/\\s+/g, ' ').trim())`

export default async function trade01(p) {
  const ort = p.ort()

  p.notiz('1 — zwei Handel an der Kommandozeile')
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'peer-anlegen')
  const entwurf = eroeffne(ort, 'bgegenueber')
  const angeboten = eroeffne(ort, 'bzweiter')
  p.mse(ort, 'handel', 'anbieten', angeboten)
  p.mse(ort, 'handel', 'zustimmen', angeboten)
  p.zeige(`Entwurf ${entwurf}, angeboten ${angeboten}`)

  p.notiz('2 — die Liste im Bereich Handeln')
  await p.starte({ ort })
  await p.taste('5', 'befehl')
  await p.warte('die Liste steht', async () => ((await p.seite(zeilen)) ?? []).length === 2, 20000)
  const vorher = await p.seite(zeilen)
  p.zeige(vorher.join(' | '))
  p.pruefe('der Entwurf ist dran', vorher.some((z) => z.includes('Du bist dran: anbieten') && z.includes('Entwurf')), vorher.join(' | '))
  p.pruefe('der angebotene wartet', vorher.some((z) => z.includes('Wartet auf bzweiter')), vorher.join(' | '))
  await p.bild('liste')

  p.notiz('3 — ein Klick schlägt ihn auf')
  await p.klickeAuf(`[...(${liste}).querySelectorAll('li button')].find((b) => b.innerText.includes('anbieten'))`, 'die Zeile des Entwurfs')
  await p.wartetext('Mit diesem Handel', 20000)
  p.pruefe('die Zeile ist gewählt', await p.seite(`return (${liste}).querySelector('li button[aria-current=true]')?.innerText.includes('anbieten') ?? false`))
  p.pruefe('der Entwurf wird gespeichert, nicht vorgeschlagen', await p.seite("return [...document.querySelectorAll('button')].some((b) => b.innerText.includes('Entwurf speichern'))"))
  await p.bild('aufgeschlagen')

  p.notiz('4 — den Entwurf löschen, nach der Rückfrage')
  await p.klickeText('Mit diesem Handel')
  await p.klickeAuf("[...document.querySelectorAll('[role=menu] [role=menuitem]')].find((b) => b.checkVisibility() && b.innerText.includes('Entwurf löschen'))", '»Entwurf löschen« im Menü')
  await p.wartetext('Diesen Entwurf löschen?', 20000)
  await p.bild('rueckfrage')
  await p.klickeAuf("[...document.querySelectorAll('[role=group] button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith('Löschen'))", '»Löschen«')
  await p.warte('nur noch einer', async () => ((await p.seite(zeilen)) ?? []).length === 1, 20000)
  const nachher = await p.seite(zeilen)
  p.pruefe('übrig ist der angebotene', nachher.length === 1 && nachher[0].includes('bzweiter'), nachher.join(' | '))
  await p.bild('geloescht')

  p.pruefe('⌘Q beendet', await p.beende())
}
