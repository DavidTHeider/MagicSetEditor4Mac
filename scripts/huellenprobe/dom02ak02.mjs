// **Ein Symbol der Bibliothek bearbeiten — am nativen Fenster** — DOM-02.AK-02
// und AK-03 (#1697):
//
// - *Zeichnung bearbeiten* öffnet das Symbol im Editor, eine Leiste nennt es,
// - *In der Bibliothek sichern* nennt die Sets mit der vorigen Fassung,
// - *In diese Sets übernehmen* schreibt ihre Kopie neu, und der Editor des
//   Setsymbols zeigt die neue Zeichnung,
// - *Vorige Fassung zurückholen* holt sie zurück — die Kommandozeile sieht
//   dieselbe Zeichnung,
// - *In den Fassungen dieses Sets* zeichnet das Symbol in jeder Fassung des
//   Stils in drei Größen.
//
// **Ein eigenes Spiel mit Fassungen**: die Saat hat keines, und ohne sie
// zeigte die Probe nur den leeren Zustand. Es liegt im eigenen Zuhause und
// wird über die Kommandozeile eingebunden.
import { mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { eintrag, knopf, knopfIn, zumSymbolfach } from './dom02.mjs'

const spiel = `mse version: 2.0.0
short name: Fass
set field:
\ttype: symbol
\tname: symbol
card field:
\ttype: text
\tname: name
default set style:
\tsymbol:
\t\tvariation:
\t\t\tname: common
\t\t\tborder radius: 0.10
\t\t\tfill type: solid
\t\t\tfill color: rgb(0,0,0)
\t\t\tborder color: rgb(255,255,255)
\t\tvariation:
\t\t\tname: rare
\t\t\tfill type: linear gradient
\t\t\tfill color 1: rgb(255,200,0)
\t\t\tfill color 2: rgb(120,80,0)
`

const stil = `mse version: 2.0.0
game: fass
short name: Schlicht
card width: 375
card height: 523
card dpi: 150
card style:
\tname:
\t\tleft: 24
\t\ttop: 24
\t\twidth: 327
\t\theight: 28
set info style:
\tsymbol:
\t\tvariation:
\t\t\tname: mythic
\t\t\tfill color: rgb(200,60,0)
`

// eingebunden legt ein Paket in den Ort und bindet es ein.
function eingebunden(p, ort, name, datei, inhalt) {
  const ordner = join(ort.ort, 'quellen', name)
  mkdirSync(ordner, { recursive: true })
  writeFileSync(join(ordner, datei), inhalt)
  p.mse(ort, 'paket-installieren', ordner)
}

// teile zählt die Teile der Zeichnung, wie die Kommandozeile sie gibt.
const teile = (p, ort, id) => JSON.parse(p.mse(ort, 'symbolbibliothek', 'teile', id, '--json')).teile.length
const bibliothek = (p, ort) => JSON.parse(p.mse(ort, 'symbolbibliothek', '--json'))
const leiste = (name) => `${name} aus der Bibliothek`

export default async function dom02ak02(p) {
  const ort = p.ort({ zuhause: true })
  eingebunden(p, ort, 'fass.mse-game', 'game', spiel)
  eingebunden(p, ort, 'fass-schlicht.mse-style', 'style', stil)
  p.mse(ort, 'set-neu', 'Kronenset', 'fass', 'fass-schlicht')

  await p.starte({ ort })

  p.notiz('Vorbereitung — das Setsymbol zeichnen und als »Krone« aufnehmen')
  await zumSymbolfach(p, 'Kronenset')
  await p.klickeText('Kreis')
  await p.wartetext('Kreis — 4 Punkte')
  await p.klickeText('⤓ Ins Set legen')
  await p.wartetext('abgelegt. Beim Sichern des Sets', 10000)
  await p.klickeAuf("[...document.querySelectorAll('label')].find((l) => l.innerText.startsWith('Name für das Setsymbol'))?.querySelector('input')", 'das Namensfeld')
  await p.tippeInsFeld('Krone')
  await p.klickeAuf(knopf('Setsymbol aufnehmen'), 'Setsymbol aufnehmen')
  await p.wartetext('Krone steht jetzt in der Bibliothek.', 10000)
  const id = bibliothek(p, ort).find((s) => s.name === 'Krone')?.id
  p.pruefe('die Kommandozeile kennt Krone', id !== undefined, bibliothek(p, ort).map((s) => `${s.id}:${s.name}:${s.verwendet.join('+')}`).join(', '))
  await p.bild('vorher')

  p.notiz('AK-02 — im Editor bearbeiten und in der Bibliothek sichern')
  await p.klickeAuf(knopfIn(eintrag('Krone'), 'Zeichnung bearbeiten'), 'Zeichnung bearbeiten an Krone')
  await p.wartetext(leiste('Krone'), 10000)
  await p.wartetext('Kreis — 4 Punkte', 10000)
  await p.bild('bearbeiten')
  await p.klickeText('Rechteck')
  await p.wartetext('Rechteck — 4 Punkte')
  await p.klickeAuf(knopf('In der Bibliothek sichern'), 'In der Bibliothek sichern')
  await p.wartetext('Diese Sets tragen noch die vorige Fassung: Kronenset.', 10000)
  p.pruefe('die Bibliothek hat die neue Zeichnung', teile(p, ort, id) === 2, teile(p, ort, id))
  await p.bild('gesichert')

  p.notiz('AK-03 — in das Set übernehmen, und das Setsymbol zeigt die neue')
  await p.klickeAuf(knopf('In diese Sets übernehmen'), 'In diese Sets übernehmen')
  await p.wartetext('Übernommen in Kronenset. Ihre Karten werden neu gezeichnet.', 10000)
  await p.warte('Krone wird wieder von Kronenset getragen', () => p.seite(`return (${eintrag('Krone')})?.innerText.includes('Getragen von Kronenset') ?? false`), 10000)
  await p.bild('uebernommen')
  await p.klickeAuf(knopf('Zurück zum Setsymbol'), 'Zurück zum Setsymbol')
  await p.warte('die Leiste ist fort', () => p.seite(`return !document.body.innerText.includes(${JSON.stringify(leiste('Krone'))})`), 10000)
  await p.wartetext('Rechteck — 4 Punkte', 10000)
  p.pruefe('das Setsymbol hat beide Teile', await p.seite("return document.body.innerText.includes('Kreis — 4 Punkte')"))

  p.notiz('AK-02 — die vorige Fassung zurückholen')
  await p.klickeAuf(knopfIn(eintrag('Krone'), 'Vorige Fassung zurückholen'), 'Vorige Fassung zurückholen an Krone')
  await p.wartetext('Die vorige Fassung von Krone ist zurück.', 10000)
  p.pruefe('die Kommandozeile sieht die vorige Zeichnung', teile(p, ort, id) === 1, teile(p, ort, id))
  await p.bild('zurueckgeholt')

  p.notiz('AK-02 — das Symbol in den Fassungen des Stils')
  await p.klickeAuf(knopfIn(eintrag('Krone'), 'In den Fassungen dieses Sets'), 'Fassungen an Krone')
  await p.warte('die Fassungen sind gezeichnet', () => p.seite(`return (${eintrag('Krone')})?.querySelectorAll('img[alt^="Krone in "][src^="data:image/png"]').length ?? 0`).then((n) => n >= 9), 20000)
  const namen = await p.seite(`return [...(${eintrag('Krone')}).querySelectorAll('img[alt^="Krone in "]')].map((b) => b.alt).join(' | ')`)
  p.pruefe('drei Fassungen in drei Größen', ['common', 'rare', 'mythic'].every((f) => [16, 32, 64].every((g) => namen.includes(`Krone in ${f}, ${g} Punkte`))), namen)
  await p.bild('fassungen')

  await p.beende()
}
