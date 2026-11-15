// **Rücknahme, Zurücksetzen und „−" am nativen Fenster** — DOM-03.AK-03
// (#1698):
//
// - zwei Exemplare in „Kiste 4", beide gewählt, Lagerort „Ordner Rot" —
//   danach *Zurücknehmen*, und die Kommandozeile liest wieder „Kiste 4",
// - eine Suche, die eines ausblendet, bietet *Suche und Filter zurücksetzen*
//   neben „1 von 2" an, und der Knopf holt beide zurück,
// - an einer Deckzeile mit einer Karte ruht „−", und *Herausnehmen* steht
//   daneben.
//
// Exemplare, Lagerort und Deck legt die Kommandozeile an, bevor das Fenster
// aufgeht — dieselbe Naht wie jede andere Fläche.
import { randomUUID } from 'node:crypto'

// **Das Zeichen des Katalogs steht vor dem Text** — gesucht wird deshalb nach
// dem Ende, wo es keine Bezeichnung gibt.
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(name)} || (!b.hasAttribute('aria-label') && b.innerText.trim().endsWith(${JSON.stringify(name)}))))`

const haken = (name) => `[...document.querySelectorAll('input[type=checkbox]')].find((e) => e.checkVisibility() && e.getAttribute('aria-label') === ${JSON.stringify(name)})`

// feldNach findet ein Eingabefeld über den Text seiner Beschriftung.
const feldNach = (name) => `[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && document.getElementById(e.getAttribute('aria-labelledby') ?? '')?.innerText.trim() === ${JSON.stringify(name)})`

// **Die Zahl der Sicht und nicht irgendeine** — „2 von 2 gewählt" steht
// daneben und passte auf dieselbe Suche.
const sichtzahl = "return [...document.querySelectorAll('[role=status]')].map((s) => s.innerText.trim()).find((t) => /^\\d+ von \\d+$/.test(t)) ?? ''"

const lagerorte = (p, ort) =>
  JSON.parse(p.mse(ort, 'sammlung', '--json')).exemplare.map((e) => e.lagerort).sort().join(', ')

async function zurSeite(p, name) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld(name)
  await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(name)})`, `»${name}« in der Palette`)
}

export default async function dom03ak03(p) {
  const ort = p.ort()
  const erster = randomUUID()
  const zweiter = randomUUID()
  p.mse(ort, 'exemplar', erster, '1')
  p.mse(ort, 'exemplar', zweiter, '1')
  const zeilen = JSON.parse(p.mse(ort, 'sammlung', '--json')).exemplare.map((e) => `${e.id}:${e.fassung}`)
  p.mse(ort, 'masse-aendern', 'lagerort=Kiste 4', ...zeilen)
  const deckID = p.mse(ort, 'deck', 'neu', 'Einzelstück', 'magic').trim()
  p.mse(ort, 'deck', 'legen', deckID, erster, '1')

  await p.starte({ ort })
  await zurSeite(p, 'Sammlung')
  await p.warte('die Liste steht', () => p.seite(`return !!${haken(erster)}`), 20000)

  p.notiz('AK-03 — eine Lagerortänderung für zwei Zeilen, und zurück')
  for (const druck of [erster, zweiter]) {
    await p.seite(`${haken(druck)}.scrollIntoView({ block: 'center' }); return true`)
    await p.klickeAuf(haken(druck), `der Haken an ${druck.slice(0, 8)}`)
    await p.warte('der Haken sitzt', () => p.seite(`return ${haken(druck)}.checked`), 10000)
  }
  const feld = feldNach('Lagerort setzen auf')
  await p.seite(`${feld}.scrollIntoView({ block: 'center' }); ${feld}.focus(); return true`)
  await p.tippeInsFeld('Ordner Rot')
  await p.klickeAuf(knopf('Ausgewählte ändern'), 'Ausgewählte ändern')
  await p.wartetext('2 geändert.', 15000)
  p.pruefe('die Kommandozeile liest »Ordner Rot«', lagerorte(p, ort) === 'Ordner Rot, Ordner Rot', lagerorte(p, ort))
  await p.bild('geaendert')

  const zurueck = knopf('Den Lagerort von 2 Zeilen zurücknehmen')
  await p.seite(`${zurueck}.scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(zurueck, 'Zurücknehmen')
  await p.wartetext('2 zurückgenommen.', 15000)
  p.pruefe('die Kommandozeile liest wieder »Kiste 4«', lagerorte(p, ort) === 'Kiste 4, Kiste 4', lagerorte(p, ort))
  p.pruefe('die Rücknahme steht nicht mehr da', !(await p.seite(`return !!${zurueck}`)))
  await p.bild('zurueckgenommen')

  p.notiz('AK-03 — eine Suche, die etwas ausblendet, nimmt ein Knopf zurück')
  const suche = feldNach('Suchen')
  await p.seite(`${suche}.scrollIntoView({ block: 'center' }); ${suche}.focus(); return true`)
  await p.tippeInsFeld(erster.slice(0, 8))
  await p.warte('die Sicht zeigt 1 von 2', async () => (await p.seite(sichtzahl)) === '1 von 2', 15000)
  await p.bild('eingeschraenkt')
  await p.klickeAuf(knopf('Suche und Filter zurücksetzen'), 'Suche und Filter zurücksetzen')
  await p.warte('die Sicht zeigt 2 von 2', async () => (await p.seite(sichtzahl)) === '2 von 2', 15000)
  p.pruefe('die Suche ist leer', await p.seite(`return ${suche}.value === ''`))
  p.pruefe('der Knopf geht mit der Einschränkung', !(await p.seite(`return !!${knopf('Suche und Filter zurücksetzen')}`)))
  await p.bild('zurueckgesetzt')

  p.notiz('AK-03 — „−" ruht bei eins')
  await zurSeite(p, 'Decks')
  await p.klickeAuf(knopf('Öffnen: Einzelstück'), 'Öffnen: Einzelstück')
  const weniger = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '').startsWith('Eine weniger: '))"
  await p.warte('die Zeile steht', () => p.seite(`return !!${weniger}`), 15000)
  p.pruefe('»−« ruht bei einer Karte', await p.seite(`return ${weniger}.disabled || ${weniger}.getAttribute('aria-disabled') === 'true'`))
  p.pruefe('»Herausnehmen« steht daneben', await p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '').startsWith('Herausnehmen: '))"))
  const gelesen = JSON.parse(p.mse(ort, 'deck', deckID, '--json')).deck
  p.pruefe('das Deck führt die Karte weiter', gelesen.plaetze.length === 1 && gelesen.plaetze[0].menge === 1, JSON.stringify(gelesen.plaetze))
  await p.bild('minus-ruht')

  p.pruefe('⌘Q beendet', await p.beende())
}
