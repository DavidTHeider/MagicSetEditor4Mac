// **Eine Partie in drei Schritten — am nativen Fenster** — P6
// (#2374),
// Befunde C-04 und C-23, Artboard `PartieStart09`.
//
// # Was es prüft
//
// **S1 aus dem Bereich**: ⌘4, der Hauptknopf — die Sitze stehen vorbelegt
// da (Name aus dem Profil, eigene Decks, nie ein fremdes). Mit der Maus und
// mit der Tastatur (⇥ zum Hauptknopf, ↩). **S1 aus dem Deck**: ⌘4,
// Rechtsklick auf das Deck, „Schnell spielen“. **Über die Palette**: ⌘K,
// „Schnell spielen mit …“. **S2 bis Zug 1**: die Starthand liegt schon da;
// „Hand behalten“ am ersten Sitz, die Sicht folgt zum zweiten, „Hand
// behalten“ — zwei Handlungen. Und was daneben erreichbar bleibt: Sitz
// hinzufügen, die Regeln, eine Leihe.
//
// # Wie gezählt wird
//
// Wie in `ausgangswerte09spielen` (`huellenwege.mjs`); vorher steht in
// `docs/belege/ausgangswerte09-spielen.json`, nachher in
// `docs/belege/partiestart09-wege.json`.
//
// # Was es nicht belegt
//
// **VoiceOver** hört ein Mensch. **Den Vorhang** zwischen zwei Menschen am
// selben Gerät prüft `vorhang09` (P7); hier wird er nur durchschritten.
import { writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { rechtsklick } from '../huellenrufbuch.mjs'
import { knopf, mitteVon, stand, tabZu } from '../huellenspieltisch.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const karte = { waldboden: '1', bergpfad: '2', moosbaer: '3', glutlaeufer: '4', funkenregen: '5', rindenwall: '6' }

const partien = (p, ort) => JSON.parse(p.mse(ort, 'partien', '--json')).partien
const anfangen = "[...document.querySelectorAll('main button')].filter((b) => b.checkVisibility() && b.innerText.trim() === 'Partie anfangen').at(-1)"
// formular sagt, ob die Sitze dastehen — der Hauptknopf allein steht auch
// oben auf der Deckliste (und führt erst zum Formular).
const formular = "return [...document.querySelectorAll('input')].some((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').startsWith('Name am Tisch'))"
const amTisch = () => "return !!document.querySelector('canvas')?.checkVisibility()"
const menuepunkt = (name) => `[...document.querySelectorAll('[role=menu] [role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.trim().startsWith(${JSON.stringify(name)}))`
// deckzeile ist der Knopf „Öffnen“ an der Zeile eines Decks — der Rechtsklick
// darauf wählt das Deck und öffnet sein Menü (AP-5).
const deckzeile = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '') === ${JSON.stringify(`Öffnen: ${name}`)})`

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

// zurueck schließt den Tisch und geht zum Start — Weg, kein Messwert.
async function zurueck(p) {
  if (await p.seite(`return !!${knopf('Partie schließen')}`)) {
    await p.klickeAuf(knopf('Partie schließen'), '»Partie schließen«')
    // **Sichtbar und nicht vorhanden** — `<Activity>` hält besuchte Seiten
    // verborgen im Baum.
    await p.warte('der Tisch ist zu', () => p.seite("return ![...document.querySelectorAll('[role=application]')].some((e) => e.checkVisibility())"), 10000)
  }
  await p.taste('1', 'befehl')
  await p.ruhe()
}

export default async function partiestart09(p) {
  const ort = bestand(p)
  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const z = wegzaehler(p, datensatz)

  await p.starte({ ort })
  await p.groesse(1440, 900)

  p.notiz('S1 — aus dem Bereich, mit der Maus')
  let vorher = partien(p, ort).length
  await z.beginne('S1', 'bereich', 'maus')
  await z.taste('4', 'befehl')
  await p.warte('»Partie anfangen« steht', () => p.seite(`return !!(${anfangen})`), 20000)
  await z.klickAuf(anfangen, '»Partie anfangen« oben')
  await p.warte('das Formular steht vorbelegt', async () => (await p.seite(formular)) && p.seite(`return !(${anfangen}).disabled`), 20000)
  const sitze = await p.seite("return [...document.querySelectorAll('input')].filter((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').startsWith('Name am Tisch')).map((e) => e.value)")
  const decks = await p.seite("return [...document.querySelectorAll('section[aria-label=\"Wer sitzt am Tisch\"] select')].map((s) => s.selectedOptions[0]?.text ?? '')")
  p.pruefe('die Sitze stehen vorbelegt da — Namen und eigene Decks', sitze.every((n) => n !== '') && decks.length === 2 && decks.every((d) => d.startsWith('Waldschritt') || d.startsWith('Glutpfad')), `${sitze.join(', ')} · ${decks.join(', ')}`)
  await p.bild('anfang-vorbelegt')
  await z.klickAuf(anfangen, '»Partie anfangen«')
  await p.warte('der Tisch steht', () => p.seite(amTisch()), 30000)
  await z.ende({ wirkung: partien(p, ort).length === vorher + 1 })

  p.notiz('S2 — bis Zug 1: behalten, die Sicht folgt, behalten')
  const partie = partien(p, ort)[0].id
  const sitz = (welcher) => stand(p, ort, partie, welcher).sitze
  p.pruefe('die Starthand liegt beim Anfangen schon da', sitz('sitz1').every((s) => s.eroeffnung === 'offen'), sitz('sitz1').map((s) => `${s.kennung}: ${s.eroeffnung}`).join(', '))
  await z.beginne('S2', 'tisch', 'maus')
  for (const [name, kennung] of [['erster', 'sitz1'], ['zweiter', 'sitz2']]) {
    // **Seit P7 steht vor dem zweiten Sitz der Vorhang** (#2375): „Ich bin …“
    // gehört zum Weg, wie in `vorhang09` — S2 sind damit drei Handlungen.
    if (name === 'zweiter') {
      const ichBin = knopf(`Ich bin ${sitze[1]}`)
      await p.warte('der Vorhang steht', () => p.seite(`return !!${ichBin}`), 15000)
      await z.klickAuf(ichBin, `»Ich bin ${sitze[1]}«`)
    }
    await p.warte(`»Hand behalten« steht (${name} Sitz)`, () => p.seite(`return !!${knopf('Hand behalten')}`), 15000)
    await z.text('Hand behalten')
    await p.warte(`${name} Sitz hat behalten`, () => sitz('sitz1').find((s) => s.kennung === kennung)?.eroeffnung === 'behalten', 15000).catch(() => false)
  }
  await p.warte('Zug 1 läuft', () => stand(p, ort, partie, 'sitz1').zug.nummer === 1, 15000).catch(() => false)
  await z.ende({ wirkung: stand(p, ort, partie, 'sitz1').zug.nummer === 1 && sitz('sitz1').every((s) => s.eroeffnung === 'behalten') })
  await p.bild('zug-eins')

  p.notiz('S1 — aus dem Bereich, mit der Tastatur')
  await zurueck(p)
  vorher = partien(p, ort).length
  await z.beginne('S1', 'bereich', 'tastatur')
  await z.taste('4', 'befehl')
  await p.warte('»Partie anfangen« steht', () => p.seite(`return !!(${anfangen})`), 20000)
  if ((await tabZu(p, z, anfangen)) && (await z.taste('Enter'), await p.warte('das Formular steht', () => p.seite(formular), 20000).catch(() => false)) && (await tabZu(p, z, anfangen))) {
    await z.taste('Enter')
    await p.warte('der Tisch steht', () => p.seite(amTisch()), 30000).catch(() => false)
    await z.ende({ wirkung: partien(p, ort).length === vorher + 1 })
  } else z.ohneGang('S1', 'bereich', 'tastatur', 'FEHLT', 'der Hauptknopf ist in 200 ⇥ nicht zu erreichen')

  p.notiz('S1 — aus dem Deck: Rechtsklick, „Schnell spielen“')
  await zurueck(p)
  vorher = partien(p, ort).length
  await z.beginne('S1', 'deck', 'maus')
  await z.taste('4', 'befehl')
  await p.warte('das Deck steht', () => p.seite(`return !!(${deckzeile('Glutpfad')})`), 20000)
  const zeile = await mitteVon(p, deckzeile('Glutpfad'))
  await z.wahl('return true')
  await rechtsklick(p, zeile.x, zeile.y)
  await p.warte('das Menü des Decks ist offen', () => p.seite(`return !!${menuepunkt('Schnell spielen')}`), 10000)
  await p.bild('deck-schnell-spielen')
  await z.klickAuf(menuepunkt('Schnell spielen'), '»Schnell spielen«')
  await p.warte('der Tisch steht', () => p.seite(amTisch()), 30000).catch(() => false)
  const ausDeck = partien(p, ort)[0]
  await z.ende({ wirkung: partien(p, ort).length === vorher + 1 })
  p.zeige(`aus dem Deck: ${JSON.stringify(ausDeck.sitze)}`)

  p.notiz('Über die Palette: „Schnell spielen mit …“')
  await zurueck(p)
  vorher = partien(p, ort).length
  await z.beginne('S1', 'palette', 'maus')
  await z.taste('k', 'befehl')
  const eintrag = "[...document.querySelectorAll('button')].find((e) => e.checkVisibility() && e.innerText.includes('Schnell spielen mit Waldschritt'))"
  await p.warte('die Palette nennt »Schnell spielen mit Waldschritt«', () => p.seite(`return !!${eintrag}`), 10000).catch(() => false)
  if (await p.seite(`return !!${eintrag}`)) {
    await z.klickAuf(eintrag, '»Schnell spielen mit Waldschritt«')
    await p.warte('der Tisch steht', () => p.seite(amTisch()), 30000).catch(() => false)
    await z.ende({ wirkung: partien(p, ort).length === vorher + 1 })
  } else z.ohneGang('S1', 'palette', 'maus', 'FEHLT', 'die Palette nennt den Eintrag nicht ohne Eingabe')

  p.notiz('Was daneben erreichbar bleibt')
  await zurueck(p)
  await p.taste('4', 'befehl')
  await p.warte('»Partie anfangen« steht', () => p.seite(`return !!(${anfangen})`), 20000)
  await p.klickeAuf(anfangen, '»Partie anfangen« oben')
  await p.warte('das Formular steht', () => p.seite(formular), 20000)
  const daneben = await p.seite(`return { sitz: !!${knopf('Sitz hinzufügen')}, regeln: document.body.innerText.includes('Nach welchen Regeln'), leihe: !!${knopf('Verleihen oder leihen')} }`)
  p.pruefe('Sitz hinzufügen, die Regeln und die Leihe bleiben erreichbar', daneben.sitz && daneben.regeln && daneben.leihe, JSON.stringify(daneben))

  writeFileSync(join(wurzel, 'docs/belege/partiestart09-wege.json'), `${JSON.stringify(datensatz, null, 2)}\n`)
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
