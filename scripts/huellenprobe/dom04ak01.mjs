// **Der Turnierassistent am nativen Fenster** — DOM-04.AK-01
// (#1699):
//
// - *Turnier*: Name, Spiel und ein Format dieses Spiels, *Weiter*,
// - *Regeln*: ein Match aus einem Spiel, *Anlegen*,
// - *Teilnehmende*: *Weiter* ruht mit einer Person und geht mit zweien,
// - *Erste Runde*: der Vorschlag steht da, *Runde 1 auslosen* legt sie an —
//   und die Kommandozeile liest das Turnier mit Format und einer Runde,
// - eine **Vorlage**: die Kopie trägt die Namen und keine Runde.
//
// Das Format legt die Kommandozeile an, bevor das Fenster aufgeht — dieselbe
// Naht wie jede andere Fläche.
//
// **Seit UXR-12 (#1920) steht der Assistent in einer Schublade** hinter
// *Turnier anlegen …* im Kopf der Seite; nach der ersten Runde und nach
// *Später weitermachen* ist sie wieder zu.

// **Das Zeichen des Katalogs steht vor dem Text** — gesucht wird deshalb nach
// dem Ende, wo es keine Bezeichnung gibt.
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(name)} || (!b.hasAttribute('aria-label') && b.innerText.trim().endsWith(${JSON.stringify(name)}))))`

// feldNach findet ein Feld über den Text seiner Beschriftung — in der
// Schublade *Turnier anlegen* (ein `section` mit diesem Namen).
const feldNach = (name) => `[...document.querySelector('section[aria-label="Turnier anlegen"]').querySelectorAll('input, select')].find((e) => e.checkVisibility() && ((e.labels?.[0]?.innerText.trim().startsWith(${JSON.stringify(name)})) || document.getElementById(e.getAttribute('aria-labelledby') ?? '')?.innerText.trim() === ${JSON.stringify(name)}))`

const schritt = "return document.querySelector('[aria-label=\"Ein neues Turnier in vier Schritten\"] [aria-current=step]')?.innerText.trim() ?? ''"

const schubladeOffen = "return !!document.querySelector('section[role=dialog][aria-label=\"Turnier anlegen\"]')?.checkVisibility()"

async function assistentOeffnen(p) {
  await klicke(p, 'Turnier anlegen …')
  await p.warte('die Schublade steht', () => p.seite(schubladeOffen), 15000)
  await p.warte('der Assistent steht beim ersten Schritt', async () => (await p.seite(schritt)).endsWith('Turnier'), 15000)
}

async function klicke(p, name) {
  await p.seite(`${knopf(name)}.scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(knopf(name), name)
}

async function tippe(p, feld, text) {
  await p.seite(`${feldNach(feld)}.scrollIntoView({ block: 'center' }); ${feldNach(feld)}.focus(); return true`)
  await p.tippeInsFeld(text)
}

async function anmelden(p, name) {
  await tippe(p, 'Name', name)
  await klicke(p, 'Anmelden')
  await p.warte(`${name} steht in der Liste`, () => p.seite(`return [...document.querySelectorAll('li')].some((l) => l.checkVisibility() && l.innerText.includes(${JSON.stringify(name)}))`), 15000)
}

export default async function dom04ak01(p) {
  const ort = p.ort()
  const formatID = p.mse(ort, 'format', 'neu', 'konstruiert', 'Pauper', 'magic').trim()

  await p.starte({ ort })
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld('Turniere')
  await p.klickeAuf("[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Turniere')", '»Turniere« in der Palette')
  await p.warte('die Seite steht', () => p.seite(`return !!${knopf('Turnier anlegen …')}`), 20000)
  p.pruefe('ohne Turnier kein Feld auf der Seite', await p.seite("return [...document.querySelector('main').querySelectorAll('input, select, textarea')].filter((e) => e.checkVisibility()).length === 0"))
  await assistentOeffnen(p)

  p.notiz('AK-01 — Schritt 1: Turnier, mit Format')
  await tippe(p, 'Name', 'Freitag')
  await p.warte('Pauper steht zur Wahl', () => p.seite(`return [...${feldNach('Format')}.options].some((o) => o.value === ${JSON.stringify(formatID)})`), 15000)
  await p.seite(`const feld = ${feldNach('Format')}; feld.value = ${JSON.stringify(formatID)}; feld.dispatchEvent(new Event('change', { bubbles: true })); return true`)
  await p.bild('schritt-turnier')
  await klicke(p, 'Weiter')

  p.notiz('AK-01 — Schritt 2: Regeln')
  await p.warte('die Regeln stehen', async () => (await p.seite(schritt)).endsWith('Regeln'), 10000)
  await p.klickeAuf("[...document.querySelectorAll('label')].find((l) => l.checkVisibility() && l.innerText.includes('Ein Match aus einem Spiel'))", 'Ein Match aus einem Spiel')
  await p.bild('schritt-regeln')
  await klicke(p, 'Anlegen')

  p.notiz('AK-01 — Schritt 3: Teilnehmende')
  await p.warte('die Teilnehmenden stehen', async () => (await p.seite(schritt)).endsWith('Teilnehmende'), 15000)
  await anmelden(p, 'Anna')
  p.pruefe('»Weiter« ruht mit einer Person', await p.seite(`return ${knopf('Weiter')}.disabled`))
  await anmelden(p, 'Bert')
  await p.warte('»Weiter« geht mit zweien', () => p.seite(`return !${knopf('Weiter')}.disabled`), 10000)
  await p.bild('schritt-teilnehmende')
  await klicke(p, 'Weiter')

  p.notiz('AK-01 — Schritt 4: erste Runde')
  await p.warte('der Vorschlag steht', () => p.seite("return [...document.querySelectorAll('li')].some((l) => l.checkVisibility() && l.innerText.includes('Anna') && l.innerText.includes('Bert'))"), 15000)
  await p.bild('schritt-erste-runde')
  await klicke(p, 'Runde 1 auslosen')
  await p.warte('die Schublade ist zu', async () => !(await p.seite(schubladeOffen)), 15000)
  await p.warte('die Werkbank zeigt das Turnier', () => p.seite("return [...document.querySelectorAll('section[aria-label]')].some((e) => e.checkVisibility() && e.getAttribute('aria-label') === 'Freitag')"), 15000)

  const liste = JSON.parse(p.mse(ort, 'turniere', '--json')).turniere
  const freitag = liste.find((t) => t.name === 'Freitag')
  const gelesen = JSON.parse(p.mse(ort, 'turnier', freitag.id, '--json')).turnier
  p.pruefe('die Kommandozeile liest Format, Einzelspiel, zwei Namen und eine Runde',
    gelesen.formatID === formatID && gelesen.einzelspiel && gelesen.teilnehmer.length === 2 && gelesen.runden.length === 1,
    JSON.stringify({ formatID: gelesen.formatID, einzelspiel: gelesen.einzelspiel, teilnehmer: gelesen.teilnehmer.length, runden: gelesen.runden.length }))
  await p.bild('danach-werkbank')

  p.notiz('AK-01 und AK-03 — eine Vorlage')
  await assistentOeffnen(p)
  await p.seite(`const feld = ${feldNach('Vorlage')}; feld.value = ${JSON.stringify(freitag.id)}; feld.dispatchEvent(new Event('change', { bubbles: true })); return true`)
  await p.warte('das Spiel ruht', () => p.seite(`return ${feldNach('Spiel')}.disabled`), 10000)
  await tippe(p, 'Name', 'Freitag II')
  await klicke(p, 'Weiter')
  await p.warte('die Regeln aus der Vorlage', () => p.seite("return document.body.innerText.includes('Die Regeln kommen aus der Vorlage.')"), 10000)
  await p.bild('vorlage-regeln')
  await klicke(p, 'Anlegen')
  await p.warte('die Namen sind da', async () => (await p.seite(schritt)).endsWith('Teilnehmende'), 15000)
  await klicke(p, 'Später weitermachen')
  await p.warte('die Schublade ist zu', async () => !(await p.seite(schubladeOffen)), 15000)

  const zweites = JSON.parse(p.mse(ort, 'turniere', '--json')).turniere.find((t) => t.name === 'Freitag II')
  const kopie = JSON.parse(p.mse(ort, 'turnier', zweites.id, '--json')).turnier
  p.pruefe('die Kopie trägt die Namen, das Format und keine Runde',
    kopie.teilnehmer.length === 2 && kopie.formatID === formatID && kopie.runden.length === 0 && kopie.teilnehmer.every((t) => t.deckID === ''),
    JSON.stringify({ teilnehmer: kopie.teilnehmer.map((t) => t.name), runden: kopie.runden.length }))

  p.pruefe('⌘Q beendet', await p.beende())
}
