// **Deck kopieren und Fehlkarten auf die Wunschliste, am nativen Fenster** —
// DOM-03.AK-02 (#1698):
//
// - *Duplizieren* an der Zeile schlägt „… (Kopie)" vor, *Kopie anlegen* legt sie
//   an und schlägt sie auf — mit denselben vier Karten,
// - *Was fehlt, auf die Wunschliste* setzt die drei fehlenden darauf, und ein
//   zweiter Klick legt nichts doppelt an,
// - die Kommandozeile liest danach über denselben Bestand denselben Stand.
//
// Deck, Platz und ein eigenes Exemplar legt die Kommandozeile an, bevor das
// Fenster aufgeht — dieselbe Naht wie jede andere Fläche.
import { randomUUID } from 'node:crypto'

// **Das Zeichen des Katalogs steht vor dem Text** („+ Kopie anlegen") — gesucht
// wird deshalb nach dem Ende, wo es keine Bezeichnung gibt.
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(name)} || (!b.hasAttribute('aria-label') && b.innerText.trim().endsWith(${JSON.stringify(name)}))))`

const text = (p) => p.seite('return document.body.innerText')

const status = "return [...document.querySelectorAll('[role=status]')].find((s) => s.checkVisibility() && s.innerText.includes('Wunschliste'))?.innerText ?? ''"

export default async function dom03ak02(p) {
  const ort = p.ort()
  const druck = randomUUID()
  const deckID = p.mse(ort, 'deck', 'neu', 'Aschenrat', 'magic').trim()
  p.mse(ort, 'deck', 'legen', deckID, druck, '4')
  p.mse(ort, 'exemplar', druck, '1')

  await p.starte({ ort })

  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld('Decks')
  await p.klickeAuf("[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Decks')", '»Decks« in der Palette')
  await p.wartetext('Aschenrat', 20000)

  p.notiz('AK-02 — ein Deck kopieren')
  // **Duplizieren steht am offenen Deck** (UXR-11, #1919) — erst aufschlagen.
  await p.klickeAuf(knopf('Öffnen: Aschenrat'), 'Öffnen: Aschenrat')
  await p.warte('Duplizieren steht da', () => p.seite(`return !!${knopf('Duplizieren: Aschenrat')}`), 15000)
  await p.klickeAuf(knopf('Duplizieren: Aschenrat'), 'Duplizieren: Aschenrat')
  await p.warte('das Namensfeld trägt den Vorschlag', () => p.seite("return [...document.querySelectorAll('input')].some((e) => e.checkVisibility() && e.value === 'Aschenrat (Kopie)')"), 10000)
  p.pruefe('der Vorschlag ist »Aschenrat (Kopie)«', true)
  await p.bild('kopie-vorschlag')

  await p.klickeAuf(knopf('Kopie anlegen'), 'Kopie anlegen')
  await p.warte('die Kopie steht in der Liste', async () => (await text(p)).includes('Aschenrat (Kopie)'), 15000)

  const liste = JSON.parse(p.mse(ort, 'decks', '--json'))
  const kopie = (liste.decks ?? []).find((d) => d.name === 'Aschenrat (Kopie)')
  p.pruefe('der Bestand führt die Kopie mit eigener Kennung', !!kopie && kopie.id !== deckID, JSON.stringify(kopie ?? {}))
  const gelesen = JSON.parse(p.mse(ort, 'deck', kopie.id, '--json')).deck
  p.pruefe('die Kopie trägt dieselben vier Karten', gelesen.hauptdeck === 4 && gelesen.plaetze.length === 1 && gelesen.plaetze[0].printing === druck)

  // **Die Kopie geht auf** — der Inspector zeigt ihre Laschen.
  await p.warte('die Kopie ist aufgeschlagen', () => p.seite("return [...document.querySelectorAll('[role=tab]')].some((t) => t.checkVisibility() && t.innerText.trim() === 'Zuordnung')"), 15000)
  await p.bild('kopie-offen')

  p.notiz('AK-02 — was fehlt, auf die Wunschliste')
  await p.klickeAuf("[...document.querySelectorAll('[role=tab]')].find((t) => t.checkVisibility() && t.innerText.trim() === 'Zuordnung')", 'die Lasche »Zuordnung«')
  await p.warte('der Knopf steht', () => p.seite(`return !!${knopf('Was fehlt, auf die Wunschliste')}`), 15000)
  await p.klickeAuf(knopf('Was fehlt, auf die Wunschliste'), 'Was fehlt, auf die Wunschliste')
  const erstes = await p.warte('die Antwort steht', async () => (await p.seite(status)) || false, 15000)
  p.pruefe('drei fehlende sind neu darauf', erstes.includes('Neu auf der Wunschliste: 1') && erstes.includes('schon darauf: 0'), erstes)
  await p.bild('wunschliste-neu')

  await p.klickeAuf(knopf('Was fehlt, auf die Wunschliste'), 'Was fehlt, auf die Wunschliste — noch einmal')
  await p.warte('die zweite Antwort steht', async () => (await p.seite(status)).includes('schon darauf: 1'), 15000)
  p.pruefe('ein zweiter Klick legt nichts doppelt an', (await p.seite(status)).includes('Neu auf der Wunschliste: 0'))
  await p.bild('wunschliste-schon')

  p.notiz('Derselbe Stand an der Kommandozeile')
  const ausCLI = p.mse(ort, 'deck', 'wuenschen', kopie.id)
  p.pruefe('die Kommandozeile sieht den Wunsch über drei', ausCLI.startsWith('schon gewünscht\t3 von 3'), ausCLI.trim())

  p.pruefe('⌘Q beendet', await p.beende())
}
