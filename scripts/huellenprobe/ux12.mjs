// **Marken an jedem Fachobjekt, am nativen Fenster** — UX-12
// (#1684):
//
// - AK-01: ein Deck bekommt über „+ Marke" an seiner Zeile eine Marke, und
//   sie steht danach in Normalform an der Zeile,
// - AK-03: „Nach Marken" filtert die Liste über Go — das andere Deck geht,
//   „Alles zeigen" bringt es zurück,
// - dieselbe Marke liest danach die Kommandozeile über denselben Bestand.
//
// Die zwei Decks legt die Kommandozeile an, bevor das Fenster aufgeht —
// dieselbe Naht wie jede andere Fläche.

const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText.trim()) === ${JSON.stringify(name)})`

const text = (p) => p.seite('return document.body.innerText')

export default async function ux12(p) {
  const ort = p.ort()
  p.mse(ort, 'deck', 'neu', 'Rotes Deck', 'magic')
  p.mse(ort, 'deck', 'neu', 'Blaues Deck', 'magic')

  await p.starte({ ort })

  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld('Decks')
  await p.klickeAuf("[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Decks')", '»Decks« in der Palette')
  await p.wartetext('Blaues Deck', 20000)

  p.notiz('AK-01 — eine Marke an einem Deck')
  await p.klickeAuf(knopf('Marken von Rotes Deck'), 'Marken von Rotes Deck')
  await p.warte('das Markenfeld steht', () => p.seite("return !!document.querySelector('input[placeholder=\"etwa zum Tausch\"]')"), 10000)
  await p.klickeAuf("document.querySelector('input[placeholder=\"etwa zum Tausch\"]')", 'das Markenfeld')
  await p.tippeInsFeld('Für Freitag')
  await p.klickeAuf(knopf('+ Marke'), '+ Marke')
  await p.warte('die Marke steht in Normalform an der Zeile', async () => (await text(p)).includes('für freitag'), 10000)
  p.pruefe('die Marke steht in Normalform an der Zeile', (await text(p)).includes('für freitag'))
  await p.bild('marke-gesetzt')

  p.notiz('AK-03 — nach der Marke filtern')
  await p.klickeAuf(knopf('Nach Marken'), 'Nach Marken')
  await p.klickeAuf(knopf('Nach »für freitag« filtern'), 'Nach »für freitag« filtern')
  await p.warte('das blaue Deck geht', async () => !(await text(p)).includes('Blaues Deck'), 10000)
  p.pruefe('nur das rote Deck bleibt', (await text(p)).includes('Rotes Deck') && !(await text(p)).includes('Blaues Deck'))
  await p.bild('gefiltert')

  await p.klickeAuf(knopf('Alles zeigen'), 'Alles zeigen')
  await p.warte('das blaue Deck kommt zurück', async () => (await text(p)).includes('Blaues Deck'), 10000)

  p.notiz('Dieselbe Marke an der Kommandozeile')
  const ausCLI = p.mse(ort, '--json', 'markenkatalog')
  p.pruefe('die Kommandozeile liest dieselbe Marke', ausCLI.includes('"für freitag"'), ausCLI.replace(/\s+/g, ' ').slice(0, 120))

  p.pruefe('⌘Q beendet', await p.beende())
}
