// **⌘Z am Tisch** — P8
// (#2376),
// Befund C-07, Vermerk 0197, Artboard `Tisch09`, am Bündel aus
// `scripts/bauen.sh abnahme`.
//
// # Was es belegt
//
// **S11 mit der Tastatur und mit der Maus**: ein Klick tappt ein Land, ⌘Z
// (über das Menü „Widerrufen“ des Fensters) nimmt es zurück, der Knopf
// „Letzten Zug zurücknehmen“ ebenso. **Abgewiesen mit Grund und ohne
// Wirkung**: eine zweite Rücknahme, eine nach einem fremden Folgezug — die
// Fassung bleibt. **⌘Z trifft den Tisch nicht** von einer anderen Seite aus
// und nicht hinter dem Vorhang.
//
// # Was es nicht belegt
//
// **VoiceOver** hört ein Mensch. **Ein Gast**, der zurücknimmt, geht über
// den Wirt wie jeder Befehl — der Lauf über zwei Anwendungen gehört zu P9.
import { writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { begonnen, bisZug, heisst, kartenknopf, knopf, stand, zumTisch } from '../huellenspieltisch.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const karte = { waldboden: '1', bergpfad: '2', moosbaer: '3', glutlaeufer: '4', funkenregen: '5', rindenwall: '6' }

const absagen = "return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter((t) => t !== '')"
const vorhang = "[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && (document.getElementById(s.getAttribute('aria-labelledby') ?? '')?.textContent ?? '').trim().startsWith('Gerät an '))"

function bestand(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  p.mse(ort, 'mtgjson-lesen', join(wurzel, 'testdaten/mtgjson/tisch/TSP.json'), 'magic')
  p.mse(ort, 'deck', 'neu', 'Waldschritt', 'magic')
  for (const [k, n] of [[karte.waldboden, 2], [karte.moosbaer, 1], [karte.glutlaeufer, 1], [karte.funkenregen, 1], [karte.rindenwall, 2]]) p.mse(ort, 'deck', 'legen', '1', `--karte=${k}`, String(n))
  p.mse(ort, 'deck', 'neu', 'Glutpfad', 'magic')
  for (const [k, n] of [[karte.bergpfad, 16], [karte.glutlaeufer, 4], [karte.rindenwall, 4]]) p.mse(ort, 'deck', 'legen', '2', `--karte=${k}`, String(n))

  return ort
}

export default async function rueckgaengig09(p) {
  const ort = bestand(p)
  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const z = wegzaehler(p, datensatz)

  await p.starte({ ort })
  await p.groesse(1440, 900)

  const partie = begonnen(p, ort, '1', '2', false, ['Ida', 'Jo'])
  p.pruefe('die Partie steht in Idas erster Hauptphase', bisZug(p, ort, partie, (zug) => zug.phase === 'hauptVorKampf'))
  await zumTisch(p, 'Ida · Jo', 'Ida')
  const fassung = () => Number(stand(p, ort, partie).fassung)
  const getappt = () => Boolean(stand(p, ort, partie).objekte.find((o) => heisst(o, 'Waldboden') && o.zone === 'tisch')?.getappt)
  const tappen = async () => {
    await p.klickeAuf(kartenknopf('Waldboden', 'Schlachtfeld'), 'Waldboden auf dem Spielfeld')
    await p.warte('getappt', getappt, 10000)
  }

  p.notiz('Vorbereitung: ein Land liegt')
  await p.klickeAuf(kartenknopf('Waldboden', 'Hand'), 'Waldboden auf der Hand')
  await p.warte('das Land liegt', () => p.seite(`return !!${kartenknopf('Waldboden', 'Schlachtfeld')}`), 10000)

  p.notiz('S11 — ⌘Z nimmt das Tappen zurück')
  await tappen()
  let vor = fassung()
  await z.beginne('S11', 'tisch', 'tastatur')
  await z.taste('z', 'befehl')
  await p.warte('enttappt', () => !getappt(), 10000).catch(() => false)
  await z.ende({ wirkung: !getappt() && fassung() === vor + 1 })
  await p.bild('nach-befehl-z')

  p.notiz('Eine zweite Rücknahme wird mit Grund abgewiesen und ändert nichts')
  vor = fassung()
  await p.taste('z', 'befehl')
  await p.warte('die Absage steht', async () => (await p.seite(absagen)).some((m) => m.includes('Rücknahme nimmt man nicht zurück')), 10000).catch(() => false)
  p.pruefe('abgewiesen, mit Grund, ohne Wirkung', (await p.seite(absagen)).some((m) => m.includes('Rücknahme nimmt man nicht zurück')) && fassung() === vor, `${(await p.seite(absagen)).join(' | ')} · Fassung ${String(vor)} → ${String(fassung())}`)

  p.notiz('Nach einem fremden Folgezug wird abgewiesen')
  await tappen()
  p.mse(ort, 'partie-befehl', partie, 'bo', `fremd-${String(Date.now())}`, 'spieler.leben', '{"um":-1}')
  const nachFremd = fassung()
  await p.warte('das Fenster kennt den fremden Zug', () => p.seite(`return /Fassung ${String(nachFremd)}\\b/.test(document.body.innerText)`), 15000)
  await p.taste('z', 'befehl')
  await p.warte('die Absage steht', async () => (await p.seite(absagen)).some((m) => m.includes('anderer Sitz gehandelt')), 10000).catch(() => false)
  p.pruefe('nach Jos Zug abgewiesen, das Land bleibt getappt', (await p.seite(absagen)).some((m) => m.includes('anderer Sitz gehandelt')) && getappt() && fassung() === nachFremd, (await p.seite(absagen)).join(' | '))
  await p.bild('fremder-folgezug')

  p.notiz('S11 — mit der Maus über den Knopf neben dem Tisch')
  await p.klickeAuf(kartenknopf('Waldboden', 'Schlachtfeld'), 'Waldboden, getappt')
  await p.warte('enttappt', () => !getappt(), 10000)
  vor = fassung()
  await z.beginne('S11', 'tisch', 'maus')
  await z.klickAuf(knopf('Letzten Zug zurücknehmen'), '»Letzten Zug zurücknehmen«')
  await p.warte('wieder getappt', getappt, 10000).catch(() => false)
  await z.ende({ wirkung: getappt() && fassung() === vor + 1 })

  p.notiz('⌘Z auf einer anderen Seite trifft den Tisch nicht')
  vor = fassung()
  await p.taste('1', 'befehl')
  await p.ruhe()
  await p.taste('z', 'befehl')
  await p.ruhe()
  p.pruefe('vom Start aus bleibt der Tisch, wie er war', fassung() === vor && getappt(), `Fassung ${String(vor)} → ${String(fassung())}`)
  await p.bild('start-nach-befehl-z')
  // **Zurück an den offenen Tisch** — ⌘4 zeigt die Decks des Bereichs,
  // „Partie anfangen“ die Spielseite mit dem Tisch (Weg, kein Messwert).
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  await p.warte('der Tisch steht wieder', () => p.seite("return !!document.querySelector('canvas')?.checkVisibility()"), 15000)

  p.notiz('Hinter dem Vorhang tut ⌘Z nichts')
  await p.warte('»Passen« steht bereit', () => p.seite(`const k = ${knopf('Passen')}; return !!k && !k.disabled`), 15000)
  await p.klickeAuf(knopf('Passen'), '»Passen«')
  await p.warte('der Vorhang ist zu', () => p.seite(`return !!${vorhang}`), 15000)
  vor = fassung()
  await p.taste('z', 'befehl')
  await p.ruhe()
  p.pruefe('hinter dem Vorhang bleibt der Stand', fassung() === vor && (await p.seite(`return !!${vorhang}`)), `Fassung ${String(vor)} → ${String(fassung())}`)

  writeFileSync(join(wurzel, 'docs/belege/rueckgaengig09-wege.json'), `${JSON.stringify(datensatz, null, 2)}\n`)
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
