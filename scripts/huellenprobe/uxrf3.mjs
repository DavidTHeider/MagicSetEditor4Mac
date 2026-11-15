// **Die eigenen Schlüsselwörter als Liste mit einem Editor, am nativen
// Fenster** — UXR-F3
// (#2009):
//
// - mit **fünf** eigenen Wörtern zeichnet die Schublade höchstens zwölf
//   Felder (R3) — gemessen mit derselben Messung wie `uirevue`, ohne Wahl und
//   mit einem gewählten Wort,
// - beim Öffnen ist nichts gewählt und die Seite sagt, was zu tun ist (MSE:
//   `ItemList::onChangeSet` wählt nichts),
// - ↓ und ↑ in der Liste wählen und nehmen den Fokus mit — **echte Tasten**
//   über die Ereignisschlange, wie in der Kartenliste,
// - eine **offene Eingabe** wird geschrieben, wenn ein Klick auf ein anderes
//   Wort den Editor ablöst (WebKit gibt dem Knopf keinen Fokus; die
//   Kommandozeile liest den Wert zurück),
// - *Schlüsselwort anlegen* wählt das neue Wort und setzt den Fokus in seinen
//   Namen, und *Löschen* wählt das folgende.
//
// Die fünf Wörter legt die Kommandozeile an, bevor das Fenster aufgeht —
// dieselbe Naht wie jede andere Fläche.
//
// **Was es nicht tut**: urteilen, ob die Anordnung dem Artboard gleicht und
// ob die Liste neben dem Editor (ab 48 rem Platz) und darüber (schmal) gut
// aussieht. Das sieht ein Mensch an den Bildern. Und Pos1 und Ende fehlen:
// die Leitung kennt diese beiden Tasten nicht (`tastencodes`,
// `internal/fenster/abnahme.go`).
import { messung, oeffnen, zurSeite } from './uirevue.mjs'

const groessen = [
  { name: 'normal', breite: 1280, hoehe: 800 },
  { name: 'eng', breite: 1000, hoehe: 700 },
]

const woerter = ['Beben', 'Kicker', 'Rausch', 'Sog', 'Wacht']

const gitter = `document.querySelector('[role=dialog] [role=grid][aria-label="Eigene Schlüsselwörter"]')`
const alleZeilen = `[...(${gitter})?.querySelectorAll('[role=row][data-wort]') ?? []]`
const zeile = (name) => `${alleZeilen}.find((z) => z.getAttribute('aria-label') === ${JSON.stringify(name)})`
const zeilenknopf = (name) => `(${zeile(name)})?.querySelector('button')`
const editor = `document.querySelector('[role=dialog] [role=group][aria-label^="Bearbeiten — "]')`
const feld = (aufschrift) => `[...(${editor})?.querySelectorAll('input') ?? []].find((e) => e.labels?.[0]?.innerText.trim() === ${JSON.stringify(aufschrift)})`
const knopf = (name) => `[...document.querySelectorAll('[role=dialog] button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(name)} || (!b.hasAttribute('aria-label') && b.innerText.trim().endsWith(${JSON.stringify(name)}))))`

const gewaehlt = `return (${alleZeilen}.find((z) => z.getAttribute('aria-selected') === 'true'))?.getAttribute('aria-label') ?? null`
const imEditor = `return (${editor})?.getAttribute('aria-label') ?? null`
const fokusZeile = "return document.activeElement?.closest('[role=row][data-wort]')?.getAttribute('aria-label') ?? null"

async function wartePause() {
  await new Promise((los) => setTimeout(los, 600))
}

export default async function uxrf3(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const setID = JSON.parse(p.mse(ort, 'sets', '--json')).sets[0].id
  for (const name of woerter) {
    const kennung = p.mse(ort, 'schluesselwort', setID, 'anlegen').trim()
    p.mse(ort, 'schluesselwort', setID, 'setzen', kennung, 'wort', name)
    p.mse(ort, 'schluesselwort', setID, 'setzen', kennung, 'muster', `${name} `)
  }
  const aufDerPlatte = () => JSON.parse(p.mse(ort, 'schluesselwoerter', setID, '--json')).woerter.filter((w) => !w.fest)

  const zeilenmasse = []
  for (const groesse of groessen) {
    await oeffnen(p, ort, groesse)
    await zurSeite(p, 'Schlüsselwörter')
    await p.warte('die fünf Wörter stehen in der Liste', () => p.seite(`return ${alleZeilen}.length === ${String(woerter.length)}`), 20000)
    await wartePause()

    p.notiz(`${groesse.name} — ohne Wahl`)
    p.pruefe(`nichts gewählt (${groesse.name})`, (await p.seite(gewaehlt)) === null)
    p.pruefe(`die Seite sagt, was zu tun ist (${groesse.name})`, await p.seite("return document.body.innerText.includes('Wähle ein Schlüsselwort in der Liste')"))
    const ohneWahl = await p.seite(messung)
    p.zeige(`Messung ohne Wahl: ${JSON.stringify(ohneWahl)}`)
    p.pruefe(`R3: höchstens 12 Felder mit fünf Wörtern, ohne Wahl (${groesse.name})`, ohneWahl.felder <= 12, `${String(ohneWahl.felder)} Felder`)
    await p.bild(`schluesselwoerter-ohne-wahl-${groesse.name}`)

    p.notiz(`${groesse.name} — ein Klick wählt`)
    await p.klickeAuf(zeilenknopf('Rausch'), 'Rausch in der Liste')
    await p.warte('Rausch ist gewählt', async () => (await p.seite(gewaehlt)) === 'Rausch')
    await p.warte('der Editor zeigt Rausch', async () => (await p.seite(imEditor)) === 'Bearbeiten — Rausch')
    p.pruefe(`das Namensfeld trägt Rausch (${groesse.name})`, (await p.seite(`return (${feld('Name')})?.value ?? null`)) === 'Rausch')
    await wartePause()
    const mitWahl = await p.seite(messung)
    p.zeige(`Messung mit Wahl: ${JSON.stringify(mitWahl)}`)
    p.pruefe(`R3: höchstens 12 Felder mit fünf Wörtern, ein Wort gewählt (${groesse.name})`, mitWahl.felder <= 12, `${String(mitWahl.felder)} Felder`)
    p.pruefe(`R5: nichts rollt waagerecht (${groesse.name})`, !mitWahl.waagerecht)
    zeilenmasse.push(`| ${groesse.name} | ohne Wahl | ${String(ohneWahl.felder)} (${String(ohneWahl.felderImBild)}) | ${String(ohneWahl.bedien)} (${String(ohneWahl.bedienImBild)}) | ${String(ohneWahl.rolltiefe)} |`)
    zeilenmasse.push(`| ${groesse.name} | Rausch gewählt | ${String(mitWahl.felder)} (${String(mitWahl.felderImBild)}) | ${String(mitWahl.bedien)} (${String(mitWahl.bedienImBild)}) | ${String(mitWahl.rolltiefe)} |`)
    await p.bild(`schluesselwoerter-gewaehlt-${groesse.name}`)

    p.notiz(`${groesse.name} — die Tasten der Kartenliste`)
    // **Der Fokus auf die gewählte Zeile gesetzt**, dann echte Pfeiltasten: so
    // sitzt man nach einem Tabulator in der Liste.
    await p.seite(`(${zeile('Rausch')}).focus(); return true`)
    await p.taste('ArrowDown')
    await p.warte('↓ wählt Sog', async () => (await p.seite(gewaehlt)) === 'Sog')
    p.pruefe(`↓ nimmt den Fokus mit (${groesse.name})`, (await p.seite(fokusZeile)) === 'Sog')
    p.pruefe(`der Editor folgt der Wahl (${groesse.name})`, (await p.seite(imEditor)) === 'Bearbeiten — Sog')
    await p.taste('ArrowDown')
    await p.warte('↓ wählt Wacht', async () => (await p.seite(gewaehlt)) === 'Wacht')
    await p.taste('ArrowDown')
    await wartePause()
    p.pruefe(`am Ende ist Schluss: kein Umlauf (${groesse.name})`, (await p.seite(gewaehlt)) === 'Wacht')
    await p.taste('ArrowUp')
    await p.taste('ArrowUp')
    await p.taste('ArrowUp')
    await p.warte('↑ wählt Kicker', async () => (await p.seite(gewaehlt)) === 'Kicker')
    p.pruefe(`↑ nimmt den Fokus mit (${groesse.name})`, (await p.seite(fokusZeile)) === 'Kicker')
    await p.taste('ArrowUp')
    await p.taste('ArrowUp')
    await wartePause()
    p.pruefe(`am Anfang ist Schluss: kein Umlauf (${groesse.name})`, (await p.seite(gewaehlt)) === 'Beben')

    p.notiz(`${groesse.name} — eine offene Eingabe überlebt den Wechsel`)
    await p.klickeAuf(zeilenknopf('Kicker'), 'Kicker in der Liste')
    await p.warte('Kicker ist gewählt', async () => (await p.seite(imEditor)) === 'Bearbeiten — Kicker')
    await p.klickeAuf(feld('Muster'), 'das Musterfeld')
    // **Auf die Wirkung des Klicks warten** (§3): gemessen am Fenster hatte
    // direkt danach noch der Knopf »Kicker« den Fokus — WebKit verarbeitet
    // den Klick nach der Antwort der Leitung, und getippt wurde ins Leere.
    await p.warte('das Musterfeld hat den Fokus', () => p.seite(`return document.activeElement === ${feld('Muster')}`))
    // **Die Textmarke ans Ende**: ein Klick in die Mitte eines breiten Feldes
    // setzt sie hinter den Text, aber nicht zwingend.
    await p.seite(`const f = ${feld('Muster')}; f.setSelectionRange(f.value.length, f.value.length); return true`)
    await p.tippeInsFeld('Z')
    // **Ohne das Feld zu verlassen**: ein Klick auf ein anderes Wort ist der
    // Weg, auf dem WebKit dem Feld keinen `blur` schickt.
    await p.klickeAuf(zeilenknopf('Sog'), 'Sog in der Liste')
    await p.warte('Sog ist gewählt', async () => (await p.seite(imEditor)) === 'Bearbeiten — Sog')
    await p.warte('die Kommandozeile liest den Wert zurück', async () => aufDerPlatte().some((w) => w.wort === 'Kicker' && w.muster === 'Kicker Z'), 20000)
    p.pruefe(`das Muster von Kicker steht auf der Platte (${groesse.name})`, aufDerPlatte().some((w) => w.wort === 'Kicker' && w.muster === 'Kicker Z'))
    // **Zurück auf den Ausgangswert**, damit die zweite Größe von vorn beginnt.
    p.mse(ort, 'schluesselwort', setID, 'setzen', aufDerPlatte().find((w) => w.wort === 'Kicker').id, 'muster', 'Kicker ')

    p.pruefe(`⌘Q beendet die Anwendung (${groesse.name})`, await p.beende())
  }

  // **Anlegen und Löschen mit einem frischen Fenster** — die Größe ist
  // gleich, und die Liste hat wieder ihre fünf.
  await oeffnen(p, ort, groessen[0])
  await zurSeite(p, 'Schlüsselwörter')
  await p.warte('die fünf Wörter stehen in der Liste', () => p.seite(`return ${alleZeilen}.length === ${String(woerter.length)}`), 20000)

  p.notiz('Anlegen — das neue Wort ist gewählt, der Fokus steht im Namen')
  await p.klickeAuf(knopf('Schlüsselwort anlegen'), 'Schlüsselwort anlegen')
  await p.warte('das sechste Wort steht in der Liste', () => p.seite(`return ${alleZeilen}.length === ${String(woerter.length + 1)}`), 20000)
  await p.warte('das neue Wort ist gewählt', async () => (await p.seite(gewaehlt)) === 'ohne Namen')
  await p.warte('der Fokus steht im Namensfeld', () => p.seite(`const f = ${feld('Name')}; return !!f && document.activeElement === f`))
  p.pruefe('das neue Wort ist gewählt, sein Namensfeld hat den Fokus', true)
  await p.tippeInsFeld('Wirbel')
  await p.taste('Tab')
  await p.warte('der Name steht in der Liste', () => p.seite(`return !!(${zeile('Wirbel')})`), 20000)
  p.pruefe('der Name steht auf der Platte', aufDerPlatte().some((w) => w.wort === 'Wirbel'))
  await p.bild('schluesselwoerter-angelegt')

  p.notiz('Löschen — gewählt wird das folgende Wort, der Fokus geht mit')
  await p.klickeAuf(zeilenknopf('Rausch'), 'Rausch in der Liste')
  await p.warte('Rausch ist offen', async () => (await p.seite(imEditor)) === 'Bearbeiten — Rausch')
  await p.klickeAuf(knopf('Löschen — Rausch'), 'Löschen — Rausch')
  await p.warte('Rausch ist weg', () => p.seite(`return !(${zeile('Rausch')})`), 20000)
  await p.warte('Sog ist gewählt', async () => (await p.seite(gewaehlt)) === 'Sog')
  p.pruefe('das folgende Wort ist gewählt', true)
  await p.warte('der Fokus steht auf der Zeile von Sog', async () => (await p.seite(fokusZeile)) === 'Sog')
  p.pruefe('der Fokus ging mit', true)

  p.notiz('Die Tabelle')
  p.zeige('| Größe | Zustand | Felder (im Bild) | Bedienelemente (im Bild) | Rolltiefe |')
  p.zeige('| --- | --- | ---: | ---: | ---: |')
  for (const z of zeilenmasse) p.zeige(z)

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
