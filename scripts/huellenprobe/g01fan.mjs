// **Der erfahrene Fan: Import, Deck, Netz** — G01
// (#2248),
// am nativen Fenster und **in einem Zug**.
//
// Er bringt mit, was er schon hat: eine Sammlung von zwei Drucken (über
// `mse` angelegt — das ist nicht der Weg, sondern sein Ausgangsbestand) und
// einen Freund, der ihm einen Einladungscode schickt (eine zweite Identität,
// ebenfalls über `mse`). Dann geht er drei Wege und provoziert je den Fehler,
// der einem Menschen dort begegnet:
//
//   1. **Import** — ein Set aus MTG JSON; vorher eine Datei, die keines ist,
//   2. **Deck** — aus den Exemplaren der Sammlung, über das Auswahlband
//      (A07.2, das B03 als ungefahren führte),
//   3. **Netz** — den Code des Freundes prüfen und bestätigen; vorher ein
//      verstümmelter Code.
//
// Gezählt wird wie in `ausgangswerte` (`huellenzaehler.mjs`); ein
// provozierter Fehlgriff zählt mit, denn auch ihn tut ein Mensch.
//
// **Was er nicht belegt**: zwei Rechner — der Freund ist eine zweite
// Identität auf diesem Mac (§17, außerhalb: zwei Geräte); und keinen
// Menschen.
import { execFileSync } from 'node:child_process'
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'
import { zaehler } from '../huellenzaehler.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const seite = "document.querySelector('main')?.dataset.seite"
const zeilenwert = (ausgabe, name) => ausgabe.split('\n').find((z2) => z2.startsWith(`${name}\t`))?.split('\t')[1] ?? ''
const setzeWahl = (beschriftung, wert) => `const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && (w.labels?.[0]?.textContent ?? '').startsWith(${JSON.stringify(beschriftung)})); if (!s) return false; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, ${JSON.stringify(wert)}); s.dispatchEvent(new Event('change', { bubbles: true })); return true`
const einfuegen = (ausdruck, text) => `const f = ${ausdruck}; Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype, 'value').set.call(f, ${JSON.stringify(text)}); f.dispatchEvent(new Event('input', { bubbles: true })); return f.value.length`
const meldungen = "return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter(Boolean).join(' | ')"

// zeilen sind die Exemplarzeilen der Sammlung — die Häkchen ihrer Wahl.
const zeilen = "[...document.querySelectorAll('main input[type=checkbox]')].filter((e) => e.checkVisibility() && e.closest('tr, [role=row]'))"

// mseMit gibt `mse` eine Eingabe — die Liste, die er schon besitzt.
const mseMit = (ort, eingabe, ...argumente) => execFileSync(join(wurzel, '.go/bin/mse'), argumente, { cwd: wurzel, env: kindumgebung(ort), input: eingabe }).toString()

async function abschnitt(p, z, tafel, { name, beginn, entscheidungen, fehler, ziel }, weg) {
  p.notiz(`${name} — Beginn: ${beginn}`)
  z.beginne()
  const erreicht = await weg()
  const { handlungen, dauer } = z.ende(name)
  p.pruefe(`${name}: Ziel erreicht — ${ziel}`, erreicht === true, typeof erreicht === 'string' ? erreicht : '')
  tafel.push({ name, beginn, entscheidungen, fehler, ziel, handlungen, dauer })
}

export default async function g01fan(p) {
  const ort = p.ort({ zuhause: true })
  const freund = p.ort({ zuhause: true })
  for (const o of [ort, freund]) p.mse(o, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  // **Seine Sammlung, mit Namen** — über die Liste und Scryfall, wie er sie
  // eingelesen hätte. Der erste Lauf legte zwei Exemplare über ihre Kennung
  // an, und die Tabelle zeigte dann die Kennung statt des Namens.
  p.zeige(mseMit(ort, '4 Lightning Bolt (M11) 149\n2 Counterspell (7ED) 67\n', 'liste-einlesen').trim())
  p.mse(freund, 'peer-anlegen')
  const code = zeilenwert(p.mse(freund, 'einladung-erstellen', 'Bert', 'handel'), 'code')
  const ordner = mkdtempSync(join(tmpdir(), 'g01fan-'))
  const keinSet = join(ordner, 'notizen.json')
  writeFileSync(keinSet, '{"das": "ist kein Set"}\n')
  const tafel = []

  await p.starte({ ort })
  const z = zaehler(p)

  await abschnitt(p, z, tafel, {
    name: 'Import',
    beginn: 'die Setseite (⌘2)',
    entscheidungen: 'Datei oder Verzeichnis; welches Spiel die Felder zuordnet',
    fehler: 'eine JSON-Datei, die kein Set ist',
    ziel: 'die Sets des Verzeichnisses im Bestand',
  }, async () => {
    const setsVorher = p.mse(ort, 'sets').trim().split('\n').filter(Boolean).length
    await z.taste('2', 'befehl')
    await z.klick('Ein Set von woanders einlesen')
    await p.wartetext('Verzeichnis öffnen', 20000)
    await z.dialog(keinSet)
    await z.klick('Datei öffnen …')
    const falsch = await p.warte('die Datei wird abgewiesen', async () => (await p.seite(meldungen)) || false, 30000)
    p.zeige(`abgewiesen mit: ${falsch}`)
    await p.bild('g01f-kein-set')
    await z.dialog(join(wurzel, 'testdaten/mtgjson/einzeln'))
    await z.klick('Verzeichnis öffnen …')
    await p.warte('die Sets aus dem Verzeichnis sind da', () => p.mse(ort, 'sets').trim().split('\n').filter(Boolean).length > setsVorher, 30000)
    await p.bild('g01f-eingelesen')

    return p.mse(ort, 'sets').trim().split('\n').filter(Boolean).length > setsVorher
  })

  await abschnitt(p, z, tafel, {
    name: 'Deck',
    beginn: 'die Sammlung mit zwei Drucken (⌘3)',
    entscheidungen: 'welche Exemplare; Name und Spiel des Decks',
    fehler: 'keiner provoziert — was das Spiel des Decks nicht kennt, sagt die Aufnahme',
    ziel: 'beide Drucke als Plätze im Deck',
  }, async () => {
    await z.taste('3', 'befehl')
    await p.warte('die Sammlung zeigt zwei Zeilen', () => p.seite(`return (${zeilen}).length >= 2`), 30000)
    const namen = await p.seite("return [...document.querySelectorAll('main tr, main [role=row]')].map((r) => r.innerText.replace(/\\s+/g, ' ').trim()).filter((t) => /Near Mint/.test(t))")
    p.zeige(`Zeilen: ${JSON.stringify(namen)}`)
    for (const stelle of [0, 1]) {
      await z.klickAuf(`(${zeilen})[${String(stelle)}]`, `die Wahl der Zeile ${String(stelle + 1)}`)
    }
    const band = await p.warte('das Auswahlband steht', () => p.seite("return [...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith('→ Deck'))?.innerText.trim() ?? false"), 10000)
    p.zeige(`Band: ${band}`)
    await z.klick(band)
    await p.warte('die Decks stehen', () => p.seite("return document.body.innerText.includes('Deck anlegen')"), 20000)
    await p.bild('g01f-stapel-wartet')
    await z.klick('Deck anlegen')
    const deckname = "[...document.querySelectorAll('input')].find((x) => x.checkVisibility() && (x.labels?.[0]?.textContent ?? '').startsWith('Name'))"
    await p.warte('das Namensfeld des Decks', () => p.seite(`return !!(${deckname})`))
    await z.klickAuf(deckname, 'das Namensfeld des Decks')
    await p.seite(`(${deckname}).focus(); return true`)
    await z.tippe('Blitz und Gegenzauber')
    const spiele = await p.seite("return [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && (w.labels?.[0]?.textContent ?? '').startsWith('Spiel'))?.options ? [...[...document.querySelectorAll('select')].find((w) => (w.labels?.[0]?.textContent ?? '').startsWith('Spiel')).options].map((o) => o.value) : []")
    p.zeige(`Spiele zur Wahl: ${JSON.stringify(spiele)}`)
    await z.wahl(setzeWahl('Spiel', spiele.find((s) => s !== '') ?? ''))
    await z.klick('Anlegen')
    await p.warte('»Alle ins Deck legen« steht', () => p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Alle ins Deck legen')"), 20000)
    await p.bild('g01f-deck-offen')
    await z.klick('Alle ins Deck legen')
    const deck = await p.warte('die Plätze stehen im Deck', () => {
      const decks = JSON.parse(p.mse(ort, 'decks', '--json')).decks ?? []
      const eines = decks.find((d) => d.name === 'Blitz und Gegenzauber')
      if (!eines) return false
      const plaetze = JSON.parse(p.mse(ort, 'deck', String(eines.id), '--json')).deck?.plaetze ?? []

      return plaetze.length >= 2 ? plaetze : false
    }, 20000).catch(async (grund) => { p.zeige(`Meldungen: ${await p.seite(meldungen)}`); throw grund })
    p.zeige(`Plätze: ${JSON.stringify(deck.map((d) => ({ menge: d.menge, printing: d.printing ?? d.druck ?? '' })))}`)
    await p.bild('g01f-deck-gelegt')

    return deck.length >= 2
  })

  await abschnitt(p, z, tafel, {
    name: 'Netz',
    beginn: 'Handeln (⌘5), mit dem Code des Freundes in der Ablage',
    entscheidungen: 'stimmen die dreißig Ziffern; darf er handeln',
    fehler: 'ein verstümmelter Code — und dem Profil fehlt noch die eigene Identität',
    ziel: 'der Freund bestätigt, mit Erlaubnis zu handeln',
  }, async () => {
    const eingabefeld = "document.querySelector('[data-einladungstausch] textarea:not([readonly])')"
    await z.taste('5', 'befehl')
    await z.klick('Neuer Tausch …')
    await z.klick('Gegenüber aufnehmen oder bestätigen')
    await p.wartetext('Code eingeben', 20000)
    await z.klick('Code eingeben')
    await p.warte('das Eingabefeld steht', () => p.seite(`return ${eingabefeld} !== null`))
    await z.wahl(einfuegen(eingabefeld, code.slice(0, -12)))
    await z.klick('Prüfen')
    const kaputt = await p.warte('der verstümmelte Code wird abgewiesen', async () => (await p.seite(meldungen)) || false, 20000)
    p.zeige(`abgewiesen mit: ${kaputt}`)
    await p.bild('g01f-code-kaputt')
    await z.wahl(einfuegen(eingabefeld, code))
    await z.klick('Prüfen')
    // **Ein Profil ohne Identität** bekommt ihr Anlegen am Fehler angeboten
    // (seit dem zweiten Lauf am 2026-10-10 — davor endete der Weg hier), und
    // derselbe Code wird danach noch einmal geprüft.
    const danach = await p.warte('der Vergleich oder das Anlegen der Identität steht', () => p.seite("const t = document.body.innerText; return t.includes('Bert bestätigen') ? 'vergleich' : ([...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim().endsWith('Identität anlegen') && b.closest('[data-einladungstausch]')) ? 'identitaet' : false)"), 20000)
    p.zeige(`nach dem Prüfen: ${danach}`)
    if (danach === 'identitaet') {
      await p.bild('g01f-identitaet-am-fehler')
      await z.klickAuf("[...document.querySelectorAll('[data-einladungstausch] button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith('Identität anlegen'))", '»Identität anlegen« am Fehler')
    }
    await p.wartetext('Bert bestätigen', 20000)
    await p.bild('g01f-vergleich')
    await z.text('Danach erlauben: Karten tauschen')
    await z.klick('Stimmt überein')
    await p.wartetext('Bert ist bestätigt.', 20000)
    const kontakte = p.mse(ort, 'kontakte')

    return /^bestaetigt\t/m.test(kontakte) && kontakte.includes('darf\thandel') ? true : kontakte
  })

  p.notiz('Die Tafel')
  for (const zeile of tafel) {
    p.zeige(`| ${zeile.name} | ${zeile.beginn} | ${zeile.entscheidungen} | ${zeile.fehler} | ${zeile.ziel} | ${String(zeile.handlungen)} | ${(zeile.dauer / 1000).toFixed(1).replace('.', ',')} s |`)
  }
  p.pruefe('⌘Q beendet', await p.beende())
}
