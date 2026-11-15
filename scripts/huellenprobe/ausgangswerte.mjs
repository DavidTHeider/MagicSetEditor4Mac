// **Ausgangswerte der Kernwege** — UX-01.AK-03
// (#1673):
// je Weg die **Schritte**, die **Fehlversuche** und die **beobachtete
// Dauer**, am nativen Fenster.
//
// # Was gezählt wird
//
// **Eine Handlung ist, was ein Mensch einmal tut**: ein Klick, ein Kürzel,
// ein getippter Wert (samt ⇥, das ihn schreibt), eine Wahl in einer Liste,
// eine Wahl im Dateidialog. **Rollen zählt nicht** — die Probe rollt ein
// Ziel in den Blick, bevor sie klickt. Was vorher angelegt wird, damit es den Weg
// überhaupt gibt — ein Set, ein Deck, eine Sicherung —, entsteht über `mse`
// und zählt nicht.
//
// **Die Dauer ist Maschinenzeit** und wird so genannt: vom ersten Handgriff
// bis die Wirkung im Bestand steht, gegangen von der Probe. **Sie ist keine
// Bearbeitungszeit eines Menschen** — ein Mensch liest, sucht und zögert,
// und das misst sie nicht. **Fehlversuche** hat die Probe keine, weil sie
// den Weg kennt; die Zahl steht trotzdem da, damit die Tafel dieselben
// Spalten hat wie die eines Menschen.
//
// # Was hier nicht steht
//
// **Pairing und Handel** — sie brauchen ein Gegenüber mit eigenem Bestand,
// eigener Identität und eigenem Horchposten (QA-10). Ihre Ausgangswerte
// stehen in `docs/UX-ARCHITEKTUR.md` mit ihrer Herkunft.
import { existsSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { zaehler } from '../huellenzaehler.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')

// setzeWahl setzt ein `<select>` unter einer Beschriftung — wie eine Wahl
// aus der Liste.
//
// **`innerhalb` grenzt die Suche ein**: seit FMT-04 (2026-09-30) steht im Navigator der
// Decks »Oder aus einem Set« mit einer eigenen Auswahl *Set*, und bis zum
// 2026-10-05 setzte dieses Drehbuch sie statt der in »Karten suchen«.
const setzeWahl = (beschriftung, wert, innerhalb = 'document') => `const s = [...((${innerhalb})?.querySelectorAll('select') ?? [])].find((w) => w.checkVisibility() && (w.labels?.[0]?.textContent ?? '').includes(${JSON.stringify(beschriftung)})); if (!s) return false; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, ${JSON.stringify(wert)}); s.dispatchEvent(new Event('change', { bubbles: true })); return true`

// deckNamens liest das Deck dieses Namens aus dem Bestand — oder `{}`.
const deckNamens = (p, ort, name) => (JSON.parse(p.mse(ort, 'decks', '--json')).decks ?? []).find((d) => d.name === name) ?? {}

// mengeImDeck ist die Menge der eigenen Karte `karteID` im Deck, 0 ohne Platz.
const mengeImDeck = (p, ort, deckID, karteID) => (JSON.parse(p.mse(ort, 'deck', deckID, '--json')).deck?.plaetze ?? [])
  .filter((platz) => platz.karteID === karteID)
  .reduce((summe, platz) => summe + platz.menge, 0)

// kartensuche ist der Kasten »Karten suchen« am offenen Deck.
const kartensuche = "[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText.trim() === 'Karten suchen')"

export default async function ausgangswerte(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  const sicherung = join(ort.ort, 'ausgang.msesicherung')
  p.mse(ort, 'bestand-sichern', sicherung)
  await p.starte({ ort })
  const z = zaehler(p)

  p.notiz('Kartenanlage — in einem offenen Set eine Karte anlegen und benennen')
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.wartetext('Karte anlegen', 20000)
  const kartenVorher = p.mse(ort, 'karten', '1').split('\n').filter((z2) => /^\d+\t/.test(z2)).length
  z.beginne()
  await z.klick('Karte anlegen')
  const namensfeld = "[...document.querySelectorAll('input')].find((x) => x.checkVisibility() && x.getAttribute('aria-label') === 'name')"
  // **Das Feld der neuen Karte** — die vorige steht nach dem Klick noch rund
  // 100 ms da (gemessen am 2026-10-05), und bis dahin las die Probe den Namen
  // »Langtext« als Ausgangswert. Die neue trägt keinen Namen.
  await p.warte('das Namensfeld der neuen Karte steht', () => p.seite(`return (${namensfeld})?.value === ''`))
  await z.klickAuf(namensfeld, 'das Namensfeld')
  await p.seite(`(${namensfeld}).focus(); return true`)
  await z.tippe('Ausgangskarte')
  await z.taste('Tab')
  await p.warte('die Karte steht im Bestand', () => p.mse(ort, 'karten', '1').includes('Ausgangskarte'), 20000)
  z.ende('Kartenanlage')
  p.zeige(`Karten vorher ${String(kartenVorher)}`)

  p.notiz('Import — ein Verzeichnis MTG JSON einlesen')
  const setsVorher = p.mse(ort, 'sets').trim().split('\n').length
  z.beginne()
  await z.taste('2', 'befehl')
  await z.klick('Ein Set von woanders einlesen')
  await p.wartetext('Verzeichnis öffnen')
  await z.dialog(join(wurzel, 'testdaten/mtgjson/einzeln'))
  await z.klick('Verzeichnis öffnen …')
  await p.warte('die Sets aus dem Verzeichnis sind da', () => p.mse(ort, 'sets').trim().split('\n').length > setsVorher, 30000)
  z.ende('Import')

  p.notiz('Deckbau — ein Deck anlegen und eine Karte hineinlegen')
  z.beginne()
  // **⌘4 führt auf die Decks** — einen Klick auf „Decks“ danach gibt es
  // seit AP-6 (#2363) nicht mehr: die Laschen „Decks“ und „Formate“ sind
  // weg, und ⌘4 stand schon vorher auf dieser Seite.
  await z.taste('4', 'befehl')
  await p.wartetext('Deck anlegen', 20000)
  await z.klick('Deck anlegen')
  const deckname = "[...document.querySelectorAll('input')].find((x) => x.checkVisibility() && (x.labels?.[0]?.textContent ?? '').startsWith('Name'))"
  await p.warte('das Namensfeld des Decks', () => p.seite(`return !!(${deckname})`))
  // **Der Fokus steht im Namen** (AP-6) — geprüft; fehlt er, klickt der
  // Lauf hinein und zählt die Handlung, statt sie zu verschweigen.
  const namensfokus = await p.seite(`return document.activeElement === (${deckname})`)
  p.pruefe('nach »Deck anlegen« steht der Fokus im Namen', namensfokus)
  if (!namensfokus) await z.klickAuf(deckname, 'das Namensfeld des Decks')
  await z.tippe('Ausgangsdeck')
  await p.warte('das Spiel »probe« steht zur Wahl', () => p.seite("return [...document.querySelectorAll('select option')].some((o) => o.value === 'probe')"))
  // **Das Spiel des offenen Sets ist vorbelegt** (AP-6) — dieselbe Regel.
  const spielwahl = "[...document.querySelectorAll('select')].find((w) => w.checkVisibility() && (w.labels?.[0]?.textContent ?? '').startsWith('Spiel'))"
  const spielVorbelegt = (await p.seite(`return (${spielwahl})?.value`)) === 'probe'
  p.pruefe('das Spiel des offenen Sets ist vorbelegt', spielVorbelegt)
  if (!spielVorbelegt) await z.wahl(setzeWahl('Spiel', 'probe'))
  await z.klick('Anlegen')
  await p.warte('das Deck ist offen', () => p.seite("return [...document.querySelectorAll('select')].some((w) => (w.labels?.[0]?.textContent ?? '').startsWith('Set'))"), 20000)
  // **Im Kasten gewartet** — die Auswahl im Navigator kennt das Set früher,
  // und ein Wert, den die Auswahl noch nicht führt, bleibt leer.
  await p.warte('das Set steht zur Wahl', () => p.seite(`return [...((${kartensuche})?.querySelectorAll('select option') ?? [])].some((o) => o.value === '1')`))
  // **Und das Set ist vorbelegt** (AP-6) — das offene, wenn es zum Spiel
  // des Decks gehört.
  const setwahl = `[...((${kartensuche})?.querySelectorAll('select') ?? [])].find((w) => (w.labels?.[0]?.textContent ?? '').startsWith('Set'))`
  await p.warte('das Set ist vorbelegt', () => p.seite(`return (${setwahl})?.value === '1'`), 5000).catch(() => undefined)
  const setVorbelegt = (await p.seite(`return (${setwahl})?.value`)) === '1'
  p.pruefe('das offene Set ist in der Kartenwahl vorbelegt', setVorbelegt)
  if (!setVorbelegt) await z.wahl(setzeWahl('Set', '1', kartensuche))
  await p.warte('die Karte steht zur Wahl', () => p.seite(`return [...((${kartensuche})?.querySelectorAll('select') ?? [])].some((w) => (w.labels?.[0]?.textContent ?? '').startsWith('Karte') && w.options.length > 1)`))
  await z.wahl(setzeWahl('Karte', '1', kartensuche))
  // **Gewartet wird auf die Menge der gewählten Karte im Bestand** — T03
  // (#2097). Bis zum 2026-10-06 genügte der Deckname, und den gab es schon
  // seit »Anlegen«: ein Klick, der nichts legte, war grün.
  const ausgangsdeck = deckNamens(p, ort, 'Ausgangsdeck')
  p.pruefe('das Deck »Ausgangsdeck« steht im Bestand', !!ausgangsdeck.id)
  const vorher = mengeImDeck(p, ort, ausgangsdeck.id, '1')
  await z.klick('Ins Deck legen')
  await p.warte('die Karte 1 liegt im Deck', () => mengeImDeck(p, ort, ausgangsdeck.id, '1') > vorher, 20000)
  z.ende('Deckbau')

  p.notiz('Spielstart — eine Partie mit dem Deck anfangen')
  z.beginne()
  await z.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await z.klick('Partie anfangen')
  await p.wartetext('Wer sitzt am Tisch', 20000)
  // **Jeder Sitz braucht einen Namen** — sonst bleibt der Knopf gesperrt.
  for (const [nummer, name] of [[1, 'Ada'], [2, 'Bo']]) {
    const feld = `[...document.querySelectorAll('input')].find((x) => x.checkVisibility() && (x.labels?.[0]?.textContent ?? '').trim() === 'Name am Tisch ${String(nummer)}')`
    await z.klickAuf(feld, `das Feld »Name am Tisch ${String(nummer)}«`)
    await p.seite(`(${feld}).focus(); return true`)
    await z.tippe(name)
  }
  // **Beide Sitze wählen das Deck** — je eine Wahl aus der Liste.
  for (const sitz of [0, 1]) {
    await z.wahl(`const e = [...document.querySelectorAll('section[aria-label="Wer sitzt am Tisch"] select')][${String(sitz)}]; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(e, '1'); e.dispatchEvent(new Event('change', { bubbles: true })); return true`)
  }
  await z.klickAuf("[...document.querySelectorAll('main button')].filter((b) => b.checkVisibility() && b.innerText.trim() === 'Partie anfangen').at(-1)", '»Partie anfangen« unter der Anlage')
  await p.warte('die Leinwand zeichnet', () => p.seite("return !!document.querySelector('canvas') && document.querySelector('canvas').checkVisibility()"), 30000)
  z.ende('Spielstart')

  p.notiz('Wiederherstellung — eine Sicherung zurückspielen')
  z.beginne()
  await z.taste(',', 'befehl')
  await p.warte('die Gruppe »Daten« steht zur Wahl', () => p.seite("return [...document.querySelectorAll('button, a, [role=tab]')].some((e) => e.checkVisibility() && e.innerText.replace(/\\s+/g, ' ').trim().startsWith('Daten'))"), 20000)
  await z.klickAuf("[...document.querySelectorAll('button, a, [role=tab]')].find((e) => e.checkVisibility() && e.innerText.replace(/\\s+/g, ' ').trim().startsWith('Daten'))", 'die Gruppe »Daten«')
  await p.wartetext('Sicherung zurückspielen …', 20000)
  await z.dialog(sicherung)
  await z.klick('Sicherung zurückspielen …')
  await p.wartetext('Diese Sicherung zurückspielen?', 20000)
  await z.klick('Zurückspielen')
  await p.warte('der Bestand ist ersetzt', () => !p.mse(ort, 'decks').includes('Ausgangsdeck'), 30000)
  z.ende('Wiederherstellung')

  p.pruefe('die Sicherung lag da', existsSync(sicherung))
  p.pruefe('⌘Q beendet', await p.beende())
}
