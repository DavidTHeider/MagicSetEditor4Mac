// **Eine Zahl, die keine ist, bleibt am Feld stehen** — F03
// (#2245),
// am nativen Fenster und mit echten Tasten:
//
//   1. *Sammlung*: ein Exemplar bearbeiten, in die Menge „2.5" tippen — der
//      Satz steht am Feld, das Feld ist als ungültig ausgezeichnet, und
//      *Sichern* ruht; eine ganze Zahl gibt es wieder frei,
//   2. *Decks › Aus einem Set*: „2.5" als Menge je Karte — *Aus dem Set
//      anlegen* zeigt den Satz am Feld und legt kein Deck an.
//
// **Was der Lauf vorher und nachher belegt**: vorher (bis F03) ging „2.5"
// als Kommazahl an ein `int` in Go, und die Absage kam als Satz über JSON
// statt am Feld — das steht in den Tests; hier steht der Weg danach.
const satz = 'Eine ganze Zahl ab 1.'

const mengenfeld = "[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && e.labels?.[0]?.innerText.trim() === 'Menge')"
const deckmengenfeld = "[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && e.labels?.[0]?.innerText.trim() === 'Menge je Karte')"

// amFeld liest Auszeichnung und Beschreibung eines Feldes.
const amFeld = (feld) => `const f = ${feld}; if (!f) return null; const ids = (f.getAttribute('aria-describedby') ?? '').split(' ').filter(Boolean); return { ungueltig: f.getAttribute('aria-invalid'), beschreibung: ids.map((id) => document.getElementById(id)?.innerText.trim() ?? '').join(' ') }`

// knopf findet einen Knopf an seinem Wort — **ohne das Zeichen davor**:
// „✎ Bearbeiten" heißt am Fenster so, und ein Vergleich mit dem ganzen Text
// fand ihn im dritten nativen Lauf (2026-10-10) nicht.
// leere räumt ein Feld und gibt ihm den Fokus — **statt ⌘A**: die erste
// Taste danach ging am Fenster verloren, und aus „2" wurde nichts (vierter
// nativer Lauf, 2026-10-10). Derselbe Griff wie in `d02.mjs`.
const leere = (feld) => `const f = ${feld}; if (!f) return false; f.focus(); Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(f, ''); f.dispatchEvent(new Event('input', { bubbles: true })); return document.activeElement === f && f.value === ''`

const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.replace(/^[^\\p{L}\\p{N}]+/u, '').trim() === ${JSON.stringify(text)})`

export default async function f03(p) {
  const ort = p.ort({ zuhause: true })
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'exemplar', 'bolt-m10', '3')
  // **Ein Set mit einer Karte** — aus einem leeren entstünde kein Deck, und
  // *Aus einem Set* bietet es gar nicht an (`Deckausset`).
  const setID = JSON.parse(p.mse(ort, 'set-neu', 'Mengenprobe', 'probe', 'probe-schlicht', '--json')).id
  p.mse(ort, 'karte', String(setID), 'anlegen')
  await p.starte({ ort })

  p.notiz('1 — Sammlung: „2.5" als Menge eines Exemplars')
  await p.taste('3', 'befehl')
  await p.warte('das Exemplar steht da', () => p.seite("return document.body.innerText.includes('bolt-m10')"), 20000)
  // **Die Zeile über ihre Bezeichnung** — ein Raster aus `gridcell`, ohne
  // Knopf und ohne `td`; das Wahlkästchen in der ersten Zelle wählt nichts
  // für den Inspector (erste native Läufe, 2026-10-10).
  await p.klickeAuf("[...document.querySelectorAll('[role=row]')].find((e) => e.checkVisibility() && e.innerText.includes('bolt-m10'))?.querySelectorAll('[role=gridcell]')[1]", 'das Exemplar bolt-m10')
  await p.klickeAuf(knopf('Bearbeiten'), 'Bearbeiten')
  await p.warte('die Menge steht da', () => p.seite(`return !!(${mengenfeld})`), 10000)
  await p.klickeAuf(mengenfeld, 'das Feld »Menge«')
  await p.warte('die Menge ist leer', () => p.seite(leere(mengenfeld)), 10000)
  await p.tippeInsFeld('2.5')
  const nachEingabe = await p.warte('der Satz steht am Feld', async () => {
    const lage = await p.seite(amFeld(mengenfeld))
    return lage?.ungueltig === 'true' ? lage : null
  }, 10000)
  p.zeige(JSON.stringify(nachEingabe))
  p.pruefe('der Satz am Feld sagt, was gilt', nachEingabe.beschreibung.includes(satz), nachEingabe.beschreibung)
  p.pruefe('*Sichern* ruht', await p.seite(`return ${knopf('Sichern')}?.disabled === true`))
  await p.bild('f03-inspector-2-5')

  await p.warte('die Menge ist wieder leer', () => p.seite(leere(mengenfeld)), 10000)
  await p.tippeInsFeld('2')
  const danach = await p.warte('das Feld ist wieder gültig', async () => {
    const lage = await p.seite(amFeld(mengenfeld))
    return lage?.ungueltig === 'false' ? lage : null
  }, 10000)
  p.pruefe('eine ganze Zahl gibt *Sichern* frei', danach.ungueltig === 'false' && await p.seite(`return ${knopf('Sichern')}?.disabled === false`))

  p.notiz('2 — Decks: „2.5" als Menge je Karte')
  await p.taste('Escape')
  await p.taste('4', 'befehl')
  // **Erst den Kasten aufklappen** — *Aus einem Set* steht unter *Deck
  // anlegen* und wird erst beim Aufklappen gezeichnet (`Deckanlagekasten`;
  // fünfter und sechster nativer Lauf, 2026-10-10).
  await p.warte('»Deck anlegen« steht da', () => p.seite(`return !!(${knopf('Deck anlegen')})`), 20000)
  await p.klickeAuf(knopf('Deck anlegen'), 'Deck anlegen')
  await p.warte('»Aus einem Set« steht da', () => p.seite(`return !!(${deckmengenfeld})`), 20000)
  const deckeVorher = p.mse(ort, 'decks').trim()
  await p.klickeAuf(deckmengenfeld, 'das Feld »Menge je Karte«')
  await p.warte('die Menge je Karte ist leer', () => p.seite(leere(deckmengenfeld)), 10000)
  await p.tippeInsFeld('2.5')
  await p.klickeAuf(knopf('Aus dem Set anlegen'), 'Aus dem Set anlegen')
  const amDeck = await p.warte('der Satz steht am Feld', async () => {
    const lage = await p.seite(amFeld(deckmengenfeld))
    return lage?.ungueltig === 'true' ? lage : null
  }, 10000)
  p.pruefe('der Satz am Feld sagt, was gilt', amDeck.beschreibung.includes(satz), amDeck.beschreibung)
  p.pruefe('kein Deck ist entstanden', p.mse(ort, 'decks').trim() === deckeVorher)
  await p.bild('f03-deck-aus-set-2-5')

  p.pruefe('⌘Q beendet', await p.beende())
}
