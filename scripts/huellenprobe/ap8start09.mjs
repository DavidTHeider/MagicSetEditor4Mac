// **Die Aufmerksamkeit nach Art, am nativen Fenster** — AP-8, Teil 1
// (#2365),
// Befund B-14, Entwurf `Start09`.
//
// # Was es prüft
//
// Drei Pakete mit einem Befund im Skript stehen als **eine** Zeile da, die
// Gründe auf Wunsch darunter, in der Sprache des Fensters. Die Zeile öffnet
// die Paketliste mit **genau** diesen drei; der Filter nennt sich und lässt
// sich lösen. Ein Deck, dem Karten in der Sammlung fehlen, geht mit seiner
// Lasche „Prüfung“ auf. Hell und dunkel.
//
// # Was es nicht belegt
//
// **Die Kennzahlen der Kacheln** (B-13) — das ist Teil 2. **VoiceOver** hört
// ein Mensch.
import { mkdirSync, writeFileSync } from 'node:fs'
import { randomUUID } from 'node:crypto'
import { join } from 'node:path'

const region = "[...document.querySelectorAll('section[aria-label]')].find((r) => r.checkVisibility() && r.getAttribute('aria-label') === 'Was Aufmerksamkeit braucht')"
const knopfIn = (wo, muster) => `[...(${wo})?.querySelectorAll('button') ?? []].find((b) => b.checkVisibility() && ${muster}.test(b.innerText.replace(/\\s+/g, ' ').trim()))`
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim() === ${JSON.stringify(name)})`
const sichtbarerText = (text) => `[...document.querySelectorAll('main *')].some((e) => e.checkVisibility() && e.children.length === 0 && e.textContent.trim() === ${JSON.stringify(text)})`

// befundpaket legt ein Spielpaket mit einer Feldart an, die MSE2 nicht kennt.
function befundpaket(p, ort, name) {
  const ordner = join(ort.ort, 'quellen', `${name}.mse-game`)
  mkdirSync(ordner, { recursive: true })
  writeFileSync(join(ordner, 'game'), `mse version: 2.0.0\nshort name: ${name}\ncard field:\n\ttype: bild\n\tname: bild\n`)
  p.mse(ort, 'paket-installieren', ordner)
}

async function lauf(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  for (const name of ['probe-befund-eins', 'probe-befund-zwei', 'probe-befund-drei']) befundpaket(p, ort, name)
  const deckID = p.mse(ort, 'deck', 'neu', 'Lückenhaft', 'magic').trim()
  p.mse(ort, 'deck', 'legen', deckID, randomUUID(), '2')

  await p.starte({ ort })
  await p.groesse(1440, 838)
  await p.taste('1', 'befehl')

  p.notiz(`${bild}: drei Pakete in einer Zeile`)
  const sammel = knopfIn(region, '/^3 eigene Pakete mit Befund im Skript$/')
  await p.warte('die Sammelzeile steht', () => p.seite(`return !!${sammel}`), 30000)
  for (const name of ['probe-befund-eins', 'probe-befund-zwei', 'probe-befund-drei']) {
    p.pruefe(`»${name}« steht nicht als eigene Zeile da`, !(await p.seite(`return !!${knopfIn(region, `/^${name}/`)}`)))
  }
  await p.klickeAuf(knopfIn(region, '/^Gründe zeigen$/'), '»Gründe zeigen«')
  const gruende = await p.seite(`return [...(${region}).querySelectorAll('ul ul li')].filter((e) => e.checkVisibility()).map((e) => e.innerText.replace(/\\s+/g, ' ').trim())`)
  p.pruefe('darunter stehen die drei mit ihrem Grund', gruende.length === 3 && gruende.every((g) => /bild/.test(g)), gruende.join(' · '))
  await p.bild(`aufmerksamkeit-${bild}`)

  p.notiz(`${bild}: die Zeile öffnet die Paketliste gefiltert`)
  await p.klickeAuf(sammel, 'die Sammelzeile')
  await p.warte('der Filter nennt sich', () => p.seite(`return ${sichtbarerText('3 eigene Pakete mit Befund im Skript')}`), 15000)
  // **Die Zählung der Seite** — „3 von 11 Paketen“; erst wenn sie steht, ist
  // die Liste geladen (der Satz des Filters steht schon vorher da).
  const zaehlung = () => p.seite("const z = [...document.querySelectorAll('[role=status]')].map((e) => e.innerText.trim()).find((t) => /^\\d+ von \\d+ Paketen?$/.test(t)); return z ? z.match(/\\d+/g).map(Number) : null")
  const geladen = await p.warte('die Liste ist geladen', async () => { const z = await zaehlung(); return z !== null && z[1] > 0 ? z : null }, 20000)
  const namen = await p.seite("return document.querySelector('main').innerText")
  const saat = ['probe-schlicht', 'probe-kunst', 'probe-bild', 'probe-einzeiler', 'probe-spielskript'].filter((name) => namen.includes(name))
  p.pruefe('die Liste zeigt genau die drei', geladen[0] === 3 && ['probe-befund-eins', 'probe-befund-zwei', 'probe-befund-drei'].every((name) => namen.includes(name)) && saat.length === 0, `${geladen.join(' von ')}; Saat daneben: ${saat.join(', ') || 'keine'}`)
  await p.bild(`paketliste-gefiltert-${bild}`)
  await p.klickeAuf(knopf('Alle Pakete zeigen'), '»Alle Pakete zeigen«')
  await p.warte('der Filter ist gelöst', async () => !(await p.seite(`return ${sichtbarerText('3 eigene Pakete mit Befund im Skript')}`)), 10000)
  const alle = await p.warte('alle stehen da', async () => { const z = await zaehlung(); return z !== null && z[0] === z[1] ? z : null }, 10000)
  p.pruefe('danach stehen alle Pakete da', alle[0] > 3, alle.join(' von '))

  p.notiz(`${bild}: das Deck geht mit seiner Prüfung auf`)
  await p.taste('1', 'befehl')
  const deckzeile = knopfIn(region, '/^Lückenhaft/')
  await p.warte('die Deckzeile steht', () => p.seite(`return !!${deckzeile}`), 15000)
  await p.klickeAuf(deckzeile, '»Lückenhaft«')
  await p.warte('die Lasche »Prüfung« ist offen', () => p.seite("return [...document.querySelectorAll('[role=tab]')].some((t) => t.checkVisibility() && t.innerText.trim() === 'Prüfung' && t.getAttribute('aria-selected') === 'true')"), 20000)
  p.pruefe('die Prüfung ist aufgeschlagen', true)
  await p.bild(`deck-pruefung-${bild}`)
  p.pruefe('⌘Q beendet', await p.beende())
}

export default async function ap8start09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
