// **Ein Exemplar über seinen Namen aufnehmen, am nativen Fenster** — AP-10
// (#2367),
// Befund B-20, Entwurf `Wege09`: „Nach Name, Set und Nummer — die
// Druckkennung steht nicht mehr im Hauptweg.“
//
// # Was es prüft
//
// In der Schublade *Ein Exemplar aufnehmen* führt das Feld „Karte“; die
// Druckkennung steht nicht im Bild. Ein Name, ↩, und die Drucke stehen zur
// Wahl — **von einer Gegenstelle auf 127.0.0.1**, die wie Scryfall antwortet
// (`MSE_SCRYFALL_ADRESSE`); ins Netz geht nichts. Gewählt wird der zweite
// Druck **mit der Tastatur**, aufgenommen mit *Aufnehmen*; danach trägt
// `mse sammlung --json` genau diesen Druck **mit Name, Set und Nummer**.
// Hell und dunkel.
//
// # Was es nicht belegt
//
// **Scryfall selbst** — die Gegenstelle antwortet, was die eingecheckte
// Vorrichtung `internal/scryfall/testdaten/suche.json` vormacht.
// **VoiceOver** hört ein Mensch.
import { createServer } from 'node:http'

const karte = (id, set, setname, nummer) => ({
  object: 'card', id, name: 'Lightning Bolt', lang: 'en', layout: 'normal', mana_cost: '{R}',
  type_line: 'Instant', rarity: 'common', set, set_name: setname, collector_number: nummer,
})
const drucke = [
  karte('00000000-0000-4000-8000-000000000149', 'm11', 'Magic 2011', '149'),
  karte('00000000-0000-4000-8000-000000000146', 'm10', 'Magic 2010', '146'),
]

const kopfknopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && !b.closest('[inert]') && b.innerText.trim().endsWith(${JSON.stringify(name)}))`
const schublade = "[...document.querySelectorAll('[role=dialog]')].find((d) => d.checkVisibility() && d.getAttribute('aria-label') === 'Ein Exemplar aufnehmen')"
const feld = (beschriftung) => `[...(${schublade})?.querySelectorAll('input') ?? []].find((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').trim().startsWith(${JSON.stringify(beschriftung)}))`
const knopfIn = (name) => `[...(${schublade})?.querySelectorAll('button') ?? []].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(name)})`
const drucke2 = `return [...(${schublade})?.querySelectorAll('input[type=radio]') ?? []].filter((r) => r.checkVisibility()).map((r) => r.closest('label').innerText.trim())`

async function lauf(p, bild, adresse, gefragt) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  await p.starte({ ort, umgebung: { MSE_SCRYFALL_ADRESSE: adresse } })
  await p.groesse(1440, 838)

  p.notiz(`${bild}: der Name führt, die Kennung steht nicht im Hauptweg`)
  await p.taste('3', 'befehl')
  await p.warte('der Kopf der Sammlung steht', () => p.seite(`return !!${kopfknopf('Aufnehmen …')}`), 20000)
  await p.klickeAuf(kopfknopf('Aufnehmen …'), '»Aufnehmen …«')
  await p.warte('die Schublade steht', () => p.seite(`return !!${schublade}`), 20000)
  p.pruefe('das Feld „Karte“ steht im Bild', await p.seite(`return !!${feld('Karte')}`))
  p.pruefe('die Druckkennung steht nicht im Hauptweg', !(await p.seite(`return !!${feld('Druck')}`)))
  // **Der Fokus steht schon im Feld** — Kernweg 6a, Budget 4 (AP-10).
  p.pruefe('beim Öffnen hat das Feld „Karte“ den Fokus', await p.warte('der Fokus steht im Feld „Karte“', () => p.seite(`return document.activeElement === ${feld('Karte')}`), 5000))

  p.notiz(`${bild}: ein Name, ↩, und die Drucke stehen zur Wahl`)
  await p.klickeAuf(feld('Karte'), 'das Feld »Karte«')
  await p.tippe('Lightning Bolt')
  await p.taste('Enter')
  const zurWahl = await p.warte('die Drucke stehen zur Wahl', async () => { const d = await p.seite(drucke2); return d.length === 2 ? d : null }, 15000)
  p.pruefe('beide Drucke mit Set und Nummer', zurWahl[0] === 'Lightning Bolt · Magic 2011 (M11) · Nr. 149' && zurWahl[1] === 'Lightning Bolt · Magic 2010 (M10) · Nr. 146', zurWahl.join(' | '))
  p.pruefe('gefragt wurde die Gegenstelle, nach Drucken', gefragt.some((url) => url.startsWith('/cards/search') && url.includes('unique=prints')), gefragt.join(', '))
  await p.bild(`drucke-${bild}`)

  p.notiz(`${bild}: der zweite Druck mit der Tastatur, dann Aufnehmen`)
  await p.seite(`[...(${schublade}).querySelectorAll('input[type=radio]')][0].focus(); return true`)
  await p.taste('ArrowDown')
  await p.warte('der zweite Druck ist gewählt', () => p.seite(`return [...(${schublade}).querySelectorAll('input[type=radio]')][1].checked`), 5000)
  await p.klickeAuf(knopfIn('Aufnehmen'), '»Aufnehmen«')
  const exemplar = await p.warte('das Exemplar liegt im Bestand', () => { const liste = JSON.parse(p.mse(ort, 'sammlung', '--json')).exemplare; return liste.length === 1 ? liste[0] : null }, 15000)
  p.pruefe('aufgenommen ist der gewählte Druck mit Name, Set und Nummer',
    exemplar.printing === drucke[1].id && exemplar.name === 'Lightning Bolt' && exemplar.setcode === 'm10' && exemplar.nummer === '146',
    `${exemplar.printing} · ${exemplar.name} · ${exemplar.setcode} · ${exemplar.nummer}`)
  p.pruefe('⌘Q beendet', await p.beende())
}

export default async function ap10exemplar09(p) {
  const gefragt = []
  const gegenstelle = createServer((anfrage, antwort) => {
    gefragt.push(anfrage.url)
    antwort.writeHead(200, { 'content-type': 'application/json' })
    antwort.end(JSON.stringify({ object: 'list', total_cards: drucke.length, has_more: false, data: drucke }))
  })
  await new Promise((fertig) => { gegenstelle.listen(0, '127.0.0.1', fertig) })
  const adresse = `http://127.0.0.1:${String(gegenstelle.address().port)}`
  try {
    for (const bild of ['hell', 'dunkel']) await lauf(p, bild, adresse, gefragt)
  } finally {
    gegenstelle.close()
  }
}
