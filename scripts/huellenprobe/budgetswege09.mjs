// **Markttisch, Handelszug und Kennzahlleisten unter dem Messvertrag** —
// #2532,
// Gesamtreview über 0.9.0, Befund 6.
//
// # Der Vertrag, vor dem Lauf festgelegt
//
// Die Reihen und Grenzen stehen in `scripts/messvertrag.mjs` (Beleg
// `huellenprobe/budgetswege09.md`); **eine gerissene Grenze macht den Lauf
// rot** und ist ein Befund mit eigenem Issue, keine angepasste Grenze.
//
// - **Der Bestand** ist der aus `budgetsorte09` — 3.788 Karten im großen
//   Set, **25.000 Exemplare** in der Sammlung —, dazu ein Deck mit einem
//   Druck, der fehlt. Gegenüber ist **Bert** als eigener Prozess `mse
//   horchen`, mit zwei Sturmrufern und seiner Sammlung zum Handeln frei.
// - **Markttisch aufschlagen**: zwanzigmal »Neuer Tausch mit Bert«, gemessen
//   in der Seite vom `mousedown` bis **ein neues** eigenes Fach mit einer
//   Karte und Berts Fach mit dem Sturmrufer stehen und zwei Bilder gezeichnet
//   sind. Das eigene Fach lädt die ganze Sammlung (`useMarktfaecher`).
// - **Anbieten bis zum neuen Stand**: zwanzigmal je ein frischer Tisch, eine
//   eigene Karte nach „Von dir“ gezogen, dann vom `mousedown` auf »Anbieten
//   und zustimmen« bis der Knopf fort ist und „Angeboten“ dasteht; danach
//   hält `mse handel liste`, dass der Vorgang angeboten ist.
// - **Kennzahlleisten**: Set beim Aufschlagen von »Probesatz 0.9.0« (vom
//   `mousedown` auf die Kachel), Sammlung und Deck im Wechsel über ⌘3 und ⌘4
//   (**von vor der Taste** — die Messung beginnt, bevor sie gedrückt ist),
//   Pakete über die Palette (vom `mousedown` auf die Zeile) — je bis die
//   Leiste ihre Zahlen trägt.
// - **Fußabdruck** von Anwendung und ihren WebKit-Prozessen nach jedem Weg.
//
// # Was es nicht belegt
//
// **Diese Maschine, dieser Lauf**; Last und Ausstattung stehen im Beleg.
// **Besuchte Seiten bleiben im Baum** (`Activity`): ab dem zweiten Besuch
// misst eine Leiste ihr Erscheinen, nicht ihr erstes Rechnen — das erste
// steckt im ersten Wert. **Zwei Rechner** (#1998) und Preise aus dem Netz.
import { existsSync } from 'node:fs'
import { cpus, loadavg, totalmem } from 'node:os'
import { join } from 'node:path'

import { gekoppelt, handel, knopf } from '../huellenhandel.mjs'
import { ziehe } from '../huellenrufbuch.mjs'
import { mitteVon } from '../huellenspieltisch.mjs'
import { MESSVERTRAG, tafel, urteil } from '../messvertrag.mjs'
import { prozesse, speicher } from './budgets09.mjs'
import { GROSS, ZEILEN, befuellt, pack } from './budgetsorte09.mjs'
import { horche } from './p12markttisch09.mjs'

const BELEG = 'huellenprobe/budgetswege09.md'
const ZUGE = 20
const sturm = '00000000-0000-4000-8000-0000000b1002'

const reihe = (anfang) => MESSVERTRAG.find((e) => e.beleg === BELEG && e.reihe.startsWith(anfang)).reihe

// ab misst in der Seite bis eine Bedingung gilt und zwei Bilder gezeichnet
// sind — vom `ereignis` an oder, ohne es, **von jetzt** — und legt das
// Ergebnis in `window.__abWert`.
//
// **Gelesen wird es, sobald die Seite wieder antwortet** ([gemessen]): im
// ersten Lauf blockierte das Aufschlagen des Markttischs die Seite länger als
// die 30 s, die ein Ruf ins Fenster wartet, und das Drehbuch brach ab, statt
// die Zahl zu nennen. Die Zeit misst die Seite selbst; nach zwei Minuten gilt
// die Messung als gescheitert (−1).
const ab = (bedingung, ereignis) => `window.__abWert = undefined
let beginn = ${ereignis ? '0' : 'performance.now()'}
${ereignis ? `document.addEventListener(${JSON.stringify(ereignis)}, () => { beginn = performance.now() }, { capture: true, once: true })` : ''}
const frist = setTimeout(() => { window.__abWert = -1 }, 120000)
const schau = () => {
  if (beginn && (${bedingung})) requestAnimationFrame(() => requestAnimationFrame(() => { clearTimeout(frist); window.__abWert = performance.now() - beginn }))
  else requestAnimationFrame(schau)
}
requestAnimationFrame(schau)
return true`

// gemessen wartet auf den Wert aus [ab] — **ein Ruf, der in der blockierten
// Seite verfällt, ist kein Ergebnis**, sondern ein Grund, weiter zu warten.
const gemessen = (p) => p.warte('die Messung ist fertig',
  () => p.seite('return window.__abWert ?? null').catch(() => null), 150000, 250)

const region = (name) => `[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && (s.getAttribute('aria-label') === ${JSON.stringify(name)} || s.querySelector('p')?.innerText.trim() === ${JSON.stringify(name)}))`
const ersteEigene = `[...((${region('Mein Fach')})?.querySelectorAll('button[aria-pressed]') ?? [])].find((b) => b.checkVisibility())`
const schnellwahl = knopf('Neuer Tausch mit Bert')
const anbieten = knopf('Anbieten und zustimmen')

// **Ein neues Fach** — das alte trägt eine Marke; „Neuer Tausch“ baut den
// Tisch neu auf (`neu` zählt in `Handeln.tsx`), und ein stehengebliebenes
// erfüllte die Bedingung schon vor dem Klick.
const neuerTisch = `(() => { const mein = ${region('Mein Fach')}; const bert = ${region('Fach von Bert')}; return !!mein && !mein.dataset.gemessen && !!(${ersteEigene}) && (bert?.innerText ?? '').includes('Sturmrufer') })()`
const markiereTisch = `const m = ${region('Mein Fach')}; if (m) m.dataset.gemessen = '1'; return true`
// **„Stand: Angeboten“ steht in einem Absatz** (`Handelsstandanzeige.tsx`) —
// im ersten Lauf suchte die Bedingung ein Element mit „Angeboten“ allein und
// wartete je Runde bis zur Frist.
const angeboten = `!(${anbieten}) && [...document.querySelectorAll('main p')].some((e) => e.checkVisibility() && e.innerText.trim() === 'Stand: Angeboten')`

const leiste = (name, muster) => `(() => { const u = [...document.querySelectorAll('[role=toolbar][aria-label]')].find((x) => x.checkVisibility() && x.getAttribute('aria-label') === ${JSON.stringify(name)}); return !!u && ${muster}.test(u.innerText) })()`
const kachel = (name) => `[...document.querySelectorAll('main a, main button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(name)})`
const palettenzeile = (seite) => `[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(seite)})`

async function aufschlagen(p, werte) {
  await p.taste('5', 'befehl')
  await p.warte('»Neuer Tausch mit Bert« steht', () => p.seite(`return !!${schnellwahl}`), 30000)
  for (let i = 0; i < ZUGE; i++) {
    await p.seite(markiereTisch)
    await p.seite(ab(neuerTisch, 'mousedown'))
    await p.klickeAuf(schnellwahl, '»Neuer Tausch mit Bert«')
    werte[reihe('Markttisch aufschlagen')].push(await gemessen(p))
    // **Der Tisch ist die Vorgabe**; stünde die Liste, maß die Reihe nichts.
    if (i === 0) {
      const tisch = await p.seite(`return ${knopf('Als Tisch')}?.getAttribute('aria-pressed') ?? null`)
      p.pruefe('der Markttisch steht, nicht die Liste', tisch === 'true', String(tisch))
    }
  }
}

async function bieten(p, ort, werte) {
  const vorher = handel(p, ort).zeilen.filter((x) => x.vorgang.zustand === 'angeboten').length
  for (let i = 0; i < ZUGE; i++) {
    await p.seite(markiereTisch)
    await p.klickeAuf(schnellwahl, '»Neuer Tausch mit Bert«')
    await p.warte('ein neuer Tisch steht', () => p.seite(`return ${neuerTisch}`), 30000)
    const leer = await p.seite(`return (${region('Von dir')})?.innerText ?? ''`)
    const von = await mitteVon(p, ersteEigene, false)
    const nach = await mitteVon(p, region('Von dir'), false)
    if (von && nach) await ziehe(p, [von, { x: (von.x + nach.x) / 2, y: (von.y + nach.y) / 2 }, nach])
    const gelegt = await p.warte('eine Karte liegt bei „Von dir“', async () => (await p.seite(`return (${region('Von dir')})?.innerText ?? ''`)) !== leer, 10000).catch(() => false)
    p.pruefe(`Tisch ${String(i + 1)}: eine Karte liegt bei „Von dir“`, gelegt !== false)
    await p.seite(ab(angeboten, 'mousedown'))
    await p.klickeAuf(anbieten, '»Anbieten und zustimmen«')
    werte[reihe('Anbieten bis zum neuen Stand')].push(await gemessen(p))
  }
  const danach = handel(p, ort).zeilen.filter((x) => x.vorgang.zustand === 'angeboten').length
  p.pruefe(`${String(ZUGE)} Vorgänge sind angeboten`, danach - vorher === ZUGE, `${String(vorher)} → ${String(danach)}`)
}

async function leisten(p, werte) {
  for (let i = 0; i < ZUGE; i++) {
    // **⌘W erst, wenn ein Set offen ist** — davor schlösse es das Fenster.
    await p.taste(i === 0 ? '2' : 'w', 'befehl')
    await p.warte(`»${GROSS.name}« steht als Kachel`, () => p.seite(`return !!(${kachel(GROSS.name)})`), 30000)
    await p.seite(ab(leiste('Kennzahlen des Sets', /Karten\s*[\d.]+/), 'mousedown'))
    await p.klickeAuf(kachel(GROSS.name), `»${GROSS.name}«`)
    werte[reihe('Kennzahlleiste Set')].push(await gemessen(p))
  }
  await p.taste('4', 'befehl')
  await p.warte('die Deckliste steht', () => p.seite(`return !!${knopf('Öffnen: Bestandsdeck')}`), 20000)
  await p.klickeAuf(knopf('Öffnen: Bestandsdeck'), '»Öffnen: Bestandsdeck«')
  await p.taste('2', 'befehl')
  for (let i = 0; i < ZUGE; i++) {
    await p.seite(ab(leiste('Kennzahlen der Sammlung', /Exemplare\s*[\d.]+/)))
    await p.taste('3', 'befehl')
    werte[reihe('Kennzahlleiste Sammlung')].push(await gemessen(p))
    await p.seite(ab(leiste('Kennzahlen des Decks', /Fehlt in der Sammlung\s*\d/)))
    await p.taste('4', 'befehl')
    werte[reihe('Kennzahlleiste Deck')].push(await gemessen(p))
  }
  for (let i = 0; i < ZUGE; i++) {
    await p.taste('3', 'befehl')
    await p.taste('k', 'befehl')
    await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
    await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
    await p.tippeInsFeld('Pakete')
    await p.warte('»Pakete« steht in der Palette', () => p.seite(`return !!(${palettenzeile('Pakete')})`), 10000)
    await p.seite(ab(leiste('Kennzahlen der Pakete', /Installiert\s*\d[\s\S]*Mit Befund\s*\d/), 'mousedown'))
    await p.klickeAuf(palettenzeile('Pakete'), '»Pakete« in der Palette')
    werte[reihe('Kennzahlleiste Pakete')].push(await gemessen(p))
  }
}

export default async function budgetswege09(p) {
  p.zeige(`Last vor dem Lauf ${loadavg().map((z) => z.toFixed(2)).join(' / ')} auf ${String(cpus().length)} Kernen, ${String(Math.round(totalmem() / 2 ** 30))} GiB`)
  const werte = Object.fromEntries(MESSVERTRAG.filter((e) => e.beleg === BELEG).map((e) => [e.reihe, []]))

  if (!existsSync(join(pack, 'data'))) {
    p.pruefe(`der Pack liegt unter ${pack}`, false)
  } else {
    p.notiz(`Der Bestand: ${String(GROSS.karten)} Karten, ${String(ZEILEN)} Exemplare, ein Deck; Bert als zweiter Prozess`)
    const annaOrt = p.ort()
    const bertOrt = p.ort()
    const bestand = befuellt(p, annaOrt)
    p.pruefe(`die Sammlung hat ${String(ZEILEN)} Zeilen`, bestand.zeilen === ZEILEN, String(bestand.zeilen))
    p.mse(annaOrt, 'deck', 'neu', 'Bestandsdeck', 'probe')
    p.mse(annaOrt, 'deck', 'legen', '1', '00000000-0000-4000-a000-000000000001', '4')
    p.mse(annaOrt, 'deck', 'legen', '1', '00000000-0000-4000-a000-0000000fffff', '2')
    const wer = gekoppelt(p, { Anna: annaOrt, Bert: bertOrt })
    p.mse(bertOrt, 'exemplar', sturm, '2', 'nearMint', 'de', 'nonfoil', '', '', '', '', 'Sturmrufer', 'p12', '3', '', 'besitz')
    p.mse(bertOrt, 'kontakt-freigeben', wer.Anna.kennung, 'handel', 'sammlung')
    p.mse(annaOrt, 'kontakt-adresse', wer.Bert.kennung, await horche(p, bertOrt))

    p.vorher = new Set(prozesse().map((x) => x.pid))
    await p.starte({ ort: annaOrt })
    await p.groesse(1440, 900)
    const merke = () => werte[reihe('Fußabdruck nach Markttisch')].push(speicher(p.vorher, p.pids.at(-1)).zusammen)

    // **`MSE_BUDGETSWEGE` wählt Wege** zur Fehlersuche — ohne die übrigen
    // ist der Lauf rot.
    const wege = (process.env.MSE_BUDGETSWEGE ?? 'markt,bieten,leisten').split(',')
    if (wege.includes('markt')) { p.notiz('Markttisch aufschlagen'); await aufschlagen(p, werte); merke() }
    if (wege.includes('bieten')) { p.notiz('Anbieten bis zum neuen Stand'); await bieten(p, annaOrt, werte); merke() }
    if (wege.includes('leisten')) { p.notiz('Kennzahlleisten'); await leisten(p, werte); merke() }
    p.pruefe('⌘Q beendet', await p.beende())
  }

  p.notiz('Ergebnis')
  p.zeilen.push('', '| Reihe | n | min | Median | P95 | max | Grenze | Urteil |', '| --- | --- | --- | --- | --- | --- | --- | --- |')
  for (const e of MESSVERTRAG.filter((x) => x.beleg === BELEG)) {
    p.zeilen.push(tafel(e, werte[e.reihe]))
    p.pruefe(`${e.reihe}: ${e.mass} ≤ ${String(e.grenze)}`, urteil(e, werte[e.reihe]) === 'gehalten', urteil(e, werte[e.reihe]))
  }
  p.zeilen.push('', `Rohwerte — ${Object.entries(werte).map(([name, liste]) => `${name}: ${JSON.stringify(liste.map(Math.round))}`).join('; ')}.`)
}
