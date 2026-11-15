// **Direkte Hand am Spieltisch — am nativen Fenster** — P4 und P5
// (#2372,
// #2373),
// Befunde C-02, C-03, C-08, C-09, C-20, C-21, C-24. Artboard `Tisch09`.
//
// # Was es prüft
//
// **Am gezeichneten Tisch, mit der Maus**: ein Klick spielt ein Land (S3),
// das zweite wird mit Grund abgewiesen; ein Klick tappt (S5), ein Doppelklick
// tappt **einmal** für Mana; ⌘-Klick wählt; der Rechtsklick legt eine Karte
// auf den Friedhof (S6); ⇧F10 öffnet dasselbe Menü; T an der Karte; Leben
// −1, ⇧ +5 und „Wert setzen“. **Drei schnelle Gesten** bei verzögerter
// Antwort: alle drei wirken, und das Warten steht da. **Der Bildabstand**,
// während eine Karte über den Tisch gezogen wird.
//
// **An der Liste, nur mit der Tastatur**, in einer zweiten Partie: Land
// spielen, tappen, Menü, Leben, Geste mit Leertaste, Pfeil und Eingabe,
// Escape, Aufgeben — und die beendete Partie zeigt ihr Ergebnis statt einer
// Aufforderung. Dazu axe in WebKit am laufenden Tisch.
//
// # Wie gezählt wird
//
// Wie in `ausgangswerte09spielen` (`huellenwege.mjs`); der Vorher-Stand steht
// in `docs/belege/ausgangswerte09-spielen.json`, der Nachher-Stand in
// `docs/belege/tischhand09-wege.json`. Ein nativer Rechts- oder Doppelklick
// zählt **eine** Handlung.
//
// # Was es nicht belegt
//
// **VoiceOver** hört ein Mensch, und **zwei Geräte** braucht es dafür nicht.
// Ob ⇧ an der Tastatur auch ⇧-Enter auf „+“ meint, misst es und behauptet es
// nicht. Die Bildzeit gilt für **diesen** Mac.
import { readFileSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { loslassen, rechtsklick, stoere, ziehe } from '../huellenrufbuch.mjs'
import { begonnen, bisZug, heisst, kartenknopf, knopf, meldungen, mitteVon, schliessen, stand, tabZu, zaehle, zonenfeld, zugpunkte, zumTisch } from '../huellenspieltisch.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))
const karte = { waldboden: '1', bergpfad: '2', moosbaer: '3', glutlaeufer: '4', funkenregen: '5', rindenwall: '6' }

const menueOffen = "return [...document.querySelectorAll('[role=menu]')].some((m) => m.checkVisibility())"
const menuepunkt = (name) => `[...document.querySelectorAll('[role=menu] [role=menuitem]')].find((e) => e.checkVisibility() && (e.getAttribute('aria-label') ?? e.innerText).replace(/\\s+/g, ' ').trim().replace(/^[^\\p{L}]+/u, '').startsWith(${JSON.stringify(name)}))`
const beschriftet = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '') === ${JSON.stringify(name)})`
const listenkarte = (name, zone) => `[...([...document.querySelectorAll('main li')].find((l) => l.checkVisibility() && l.firstElementChild?.innerText.trim() === ${JSON.stringify(zone)})?.querySelectorAll('button') ?? [])].find((b) => b.innerText.includes(${JSON.stringify(name)}))`
const lebensfeld = "[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && [...(e.labels ?? [])].some((l) => l.innerText.trim() === 'Neuer Stand'))"
const fokussiert = (ausdruck) => `const e = ${ausdruck}; return !!e && document.activeElement === e`
const ansage = "return document.querySelector('[data-gestenansage]')?.textContent ?? ''"

// klickMit klickt nativ, mit Klickzahl und Umschalter, und gibt zurück, was
// die Seite als `detail` sah — **der Beleg, dass der Doppelklick einer war**.
async function klickMit(p, ausdruck, beschreibung, art) {
  await p.seite("window.__details = []; if (!window.__horcher) { document.addEventListener('click', (e) => window.__details.push(e.detail), { capture: true }); window.__horcher = true } return true")
  await p.klickeAuf(ausdruck, beschreibung, art)
  await schlafe(400)

  return p.seite('return window.__details')
}

// mitschreiben hält jeden Spielbefehl fest, der die Seite verlässt — Art,
// Nutzlast, Zeit —, damit ein Befund sagt, **welcher** Befehl es war.
const mitschreiben = "if (!window.__befehle) { window.__befehle = []; const echt = window.fetch; window.fetch = (url, wahl) => { try { const ruf = JSON.parse(wahl?.body ?? '{}').args ?? {}; if (String(ruf.methodName ?? '').endsWith('PartieBefehl')) window.__befehle.push(`${(performance.now() / 1000).toFixed(2)} ${ruf.args?.[3]} ${JSON.stringify(ruf.args?.[4])} @${ruf.args?.[5]}`) } catch {} return echt(url, wahl) } } return true"
const befehle = (p) => p.seite('return window.__befehle ?? []')

// bildabstand zeichnet die Abstände der Bilder auf, solange `waehrend` läuft.
async function bildabstand(p, waehrend) {
  await p.seite("window.__bilder = []; window.__halt = false; let v = 0; const s = (t) => { if (v) window.__bilder.push(t - v); v = t; if (!window.__halt) requestAnimationFrame(s) }; requestAnimationFrame(s); return true")
  await waehrend()
  const zeiten = await p.seite('window.__halt = true; return window.__bilder')
  const sortiert = [...zeiten].sort((a, b) => a - b)
  const q = (anteil) => sortiert[Math.min(sortiert.length - 1, Math.floor(anteil * sortiert.length))]

  const lang = zeiten.map((z, i) => [i, Math.round(z)]).filter(([, z]) => z > 33)

  return { n: sortiert.length, median: q(0.5), p99: q(0.99), max: sortiert.at(-1), lang }
}

async function axeAmTisch(p) {
  const quelle = readFileSync(join(wurzel, 'frontend/node_modules/axe-core/axe.min.js'), 'utf8')
  await p.seite(`${quelle}; return true`)
  const befund = await p.seite("const r = await axe.run(document, { runOnly: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'] }); return r.violations.map((v) => `${v.id} (${v.nodes.length}): ${v.nodes.slice(0, 3).map((n) => { const e = document.querySelector(n.target.join(' ')); const k = e?.getBoundingClientRect(); return `${(e?.getAttribute('aria-label') ?? e?.innerText ?? '').replace(/\\s+/g, ' ').slice(0, 50)} ${k ? `${Math.round(k.width)}×${Math.round(k.height)}` : ''} — ${n.failureSummary?.split('\\n').slice(1, 2).join('')}` }).join(' | ')}`)")
  p.pruefe('axe in WebKit: keine Verstöße gegen WCAG 2.2 AA am Tisch', befund.length === 0, befund.join(' · '))
}

function bestand(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  p.mse(ort, 'mtgjson-lesen', join(wurzel, 'testdaten/mtgjson/tisch/TSP.json'), 'magic')
  p.mse(ort, 'deck', 'neu', 'Waldschritt', 'magic')
  for (const [k, n] of [[karte.waldboden, 2], [karte.moosbaer, 1], [karte.glutlaeufer, 1], [karte.funkenregen, 1], [karte.rindenwall, 2]]) p.mse(ort, 'deck', 'legen', '1', `--karte=${k}`, String(n))
  p.mse(ort, 'deck', 'neu', 'Glutpfad', 'magic')
  for (const [k, n] of [[karte.bergpfad, 16], [karte.glutlaeufer, 4], [karte.rindenwall, 4]]) p.mse(ort, 'deck', 'legen', '2', `--karte=${k}`, String(n))

  return ort
}

async function mitDerMaus(p, z, ort) {
  const partie = begonnen(p, ort, '1', '2', false, ['Ida', 'Jo'])
  p.pruefe('die Partie steht in Idas erster Hauptphase', bisZug(p, ort, partie, (zug) => zug.phase === 'hauptVorKampf'))
  await zumTisch(p, 'Ida · Jo', 'Ida')
  await p.warte('Idas Hand steht', () => p.seite(`return !!${kartenknopf('Waldboden', 'Hand')}`), 20000)
  await p.seite(mitschreiben)
  await p.bild('tisch-vorher')
  const objekt = (name, zone) => stand(p, ort, partie).objekte.find((o) => heisst(o, name) && o.zone === zone)
  const getappt = () => Boolean(objekt('Waldboden', 'tisch')?.getappt)
  const vorrat = () => stand(p, ort, partie).sitze.find((s) => s.kennung === 'ada').manavorrat
  const leben = () => stand(p, ort, partie).sitze.find((s) => s.kennung === 'ada').leben

  p.notiz('S3 — ein Klick spielt ein Land, das zweite wird mit Grund abgewiesen')
  await z.beginne('S3', 'tisch-klick', 'maus')
  await z.klickAuf(kartenknopf('Waldboden', 'Hand'), 'Waldboden auf der Hand')
  await p.warte('das Land liegt', () => zaehle(p, ort, partie, 'Waldboden', 'tisch') === 1, 10000).catch(() => false)
  await z.ende({ wirkung: zaehle(p, ort, partie, 'Waldboden', 'tisch') === 1 })
  await p.klickeAuf(kartenknopf('Waldboden', 'Hand'), 'den zweiten Waldboden')
  await p.warte('die Absage steht', async () => (await meldungen(p)).some((m) => m.includes('CR 305.2')), 10000).catch(() => false)
  const absage = await meldungen(p)
  p.pruefe('das zweite Land ist abgewiesen, mit Grund (CR 305.2)', absage.some((m) => m.includes('CR 305.2')) && zaehle(p, ort, partie, 'Waldboden', 'tisch') === 1, absage.join(' | '))
  p.zeige(`gestellt: ${(await befehle(p)).join(' · ')}`)
  await p.bild('zweites-land')

  p.notiz('S5 — ein Klick tappt, ein Doppelklick tappt einmal für Mana')
  await z.beginne('S5', 'tisch-klick', 'maus')
  await z.klickAuf(kartenknopf('Waldboden', 'Schlachtfeld'), 'Waldboden auf dem Spielfeld')
  await p.warte('getappt', getappt, 10000).catch(() => false)
  await z.ende({ wirkung: getappt() })
  await p.klickeAuf(kartenknopf('Waldboden', 'Schlachtfeld'), 'Waldboden, getappt')
  await p.warte('enttappt', () => !getappt(), 10000).catch(() => false)
  p.pruefe('ein zweiter Klick enttappt', !getappt())
  const details = await klickMit(p, kartenknopf('Waldboden', 'Schlachtfeld'), 'Waldboden, doppelt', { anzahl: 2 })
  await p.warte('für Mana getappt', () => getappt() && vorrat() !== '', 10000).catch(() => false)
  await schlafe(600)
  p.pruefe('der Doppelklick kam als einer an (detail 1, 2)', JSON.stringify(details) === '[1,2]', JSON.stringify(details))
  p.pruefe('der Doppelklick tappt einmal für Mana und nicht zweimal', getappt() && vorrat() === 'G', `getappt: ${String(getappt())}, Vorrat: »${vorrat()}«`)

  p.notiz('T an der Karte, ⌘-Klick wählt')
  const tasten = await p.seite(`return (${kartenknopf('Waldboden', 'Schlachtfeld')})?.getAttribute('aria-keyshortcuts') ?? ''`)
  p.pruefe('die Karte nennt ihre Einzeltasten', tasten.split(' ').includes('T'), tasten)
  await p.seite(`(${kartenknopf('Waldboden', 'Schlachtfeld')}).focus(); return true`)
  await p.taste('t')
  await p.warte('T enttappt', () => !getappt(), 10000).catch(() => false)
  p.pruefe('T stellt die Primärhandlung', !getappt())
  const fassung = stand(p, ort, partie).fassung
  await klickMit(p, kartenknopf('Moosbär', 'Hand'), 'Moosbär mit ⌘', { mit: ['befehl'] })
  p.pruefe('⌘-Klick wählt und stellt nichts', (await p.seite(`return (${kartenknopf('Moosbär', 'Hand')})?.getAttribute('aria-pressed')`)) === 'true' && stand(p, ort, partie).fassung === fassung)

  // **Die gewählte Karte steht in der Menüleiste** (#2482) — dieselben
  // Einträge wie ihr Rechtsklick.
  p.notiz('Die gewählte Karte in der Menüleiste')
  const leiste = await p.warte('die Leiste führt »Karte«', async () => (await p.menue()).find((m) => m.titel === 'Karte'), 10000).catch(() => undefined)
  const inDerLeiste = (leiste?.unter ?? []).filter((e) => !e.trenner && !e.verborgen).map((e) => e.titel)
  const amBaeren = await mitteVon(p, kartenknopf('Moosbär', 'Hand'))
  await rechtsklick(p, amBaeren.x, amBaeren.y)
  await p.warte('das Menü des Moosbären ist offen', () => p.seite(menueOffen), 10000)
  const imMenue = await p.seite("return [...document.querySelectorAll('[role=menu] [role=menuitem]')].filter((e) => e.checkVisibility()).map((e) => (e.getAttribute('aria-label') ?? e.innerText).replace(/\\s+/g, ' ').trim().replace(/^[^\\p{L}]+/u, ''))")
  await p.taste('Escape')
  p.pruefe('die Menüleiste führt dieselben Einträge wie der Rechtsklick', inDerLeiste.length > 0 && inDerLeiste.length === imMenue.length && inDerLeiste.every((name, i) => imMenue[i].startsWith(name)), `Leiste: ${inDerLeiste.join(' · ')} | Menü: ${imMenue.join(' · ')}`)

  p.notiz('S6 — Rechtsklick, „Auf den Friedhof“')
  const begraben = zaehle(p, ort, partie, 'Rindenwall', 'friedhof')
  await z.beginne('S6', 'tisch-menue', 'maus')
  const ziel = await mitteVon(p, kartenknopf('Rindenwall', 'Hand'))
  await z.wahl('return true')
  await rechtsklick(p, ziel.x, ziel.y)
  await p.warte('das Menü der Karte ist offen', () => p.seite(menueOffen), 10000)
  await p.bild('kartenmenue')
  await z.klickAuf(menuepunkt('Auf den Friedhof'), '»Auf den Friedhof«')
  await p.warte('im Friedhof', () => zaehle(p, ort, partie, 'Rindenwall', 'friedhof') > begraben, 10000).catch(() => false)
  await z.ende({ wirkung: zaehle(p, ort, partie, 'Rindenwall', 'friedhof') === begraben + 1 })
  await p.seite(`(${kartenknopf('Rindenwall', 'Hand')}).focus(); return true`)
  await p.taste('F10', 'umschalt')
  await p.warte('⇧F10 öffnet das Menü', () => p.seite(menueOffen), 10000).catch(() => false)
  p.pruefe('⇧F10: das Menü ist offen, der Fokus darin', (await p.seite("return document.activeElement?.getAttribute('role')")) === 'menuitem')

  await p.taste('Escape')

  p.notiz('S7 — Leben: dreimal −, dann ⇧ +, dann „Wert setzen“')
  const vorher = leben()
  await z.beginne('S7', 'tisch-regler', 'maus')
  for (let mal = 0; mal < 3; mal += 1) await z.klickAuf(beschriftet('Ein Leben weniger — mit ⇧ fünf'), '»−«')
  await p.warte('drei weniger', () => leben() === vorher - 3, 10000).catch(() => false)
  await z.ende({ wirkung: leben() === vorher - 3 })
  await klickMit(p, beschriftet('Ein Leben mehr — mit ⇧ fünf'), '»+« mit ⇧', { mit: ['umschalt'] })
  await p.warte('fünf mehr', () => leben() === vorher + 2, 10000).catch(() => false)
  p.pruefe('⇧-Klick auf + gibt fünf', leben() === vorher + 2, String(leben()))
  await p.klickeAuf(beschriftet(`Leben ${String(vorher + 2)} — Wert setzen`), '»Wert setzen«')
  await p.warte('das Feld steht', () => p.seite(`return !!${lebensfeld}`), 10000)
  await p.seite(`const e = ${lebensfeld}; e.focus(); Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(e, ''); e.dispatchEvent(new Event('input', { bubbles: true })); return true`)
  await p.tippeInsFeld('17')
  await p.klickeAuf(knopf('Setzen'), '»Setzen«')
  await p.warte('gesetzt', () => leben() === 17, 10000).catch(() => false)
  p.pruefe('„Wert setzen“ setzt 17', leben() === 17, String(leben()))
  await axeAmTisch(p)

  p.notiz('Drei schnelle Gesten bei verzögerter Antwort')
  await stoere(p, [{ bindung: 'PartieBefehl', verzoegerungMs: 900 }])
  let wartend = ''
  const namen = ['Glutläufer', 'Funkenregen', 'Rindenwall']
  const imFriedhof = Object.fromEntries(namen.map((name) => [name, zaehle(p, ort, partie, name, 'friedhof')]))
  for (const name of namen) {
    const [von, nach] = await zugpunkte(p, kartenknopf(name, 'Hand'), zonenfeld('friedhof', 'ada'))
    await ziehe(p, [von, { x: (von.x + nach.x) / 2, y: (von.y + nach.y) / 2 }, nach])
    wartend = (await p.seite("return [...document.querySelectorAll('[aria-live]')].map((e) => e.innerText.trim()).find((t) => t.startsWith('Wartet auf den Tisch')) ?? ''")) || wartend
  }
  await p.bild('drei-gesten-wartend')
  const alle = () => namen.every((name) => zaehle(p, ort, partie, name, 'friedhof') === imFriedhof[name] + 1)
  await p.warte('alle drei im Friedhof', alle, 15000).catch(() => false)
  await stoere(p, [])
  p.pruefe('das Warten steht am Tisch, höchstens drei', /: [123]$/.test(wartend), wartend)
  p.pruefe('drei Gesten, drei Wirkungen, keine still verloren', alle(), JSON.stringify(await meldungen(p)))
  p.zeige(`gestellt: ${(await befehle(p)).join(' · ')}`)

  p.notiz('Bildabstand in Ruhe und beim Ziehen')
  const ruhe = await bildabstand(p, () => schlafe(2000))
  p.zeige(`in Ruhe: ${JSON.stringify(ruhe)}`)
  const [von, nach] = await zugpunkte(p, kartenknopf('Moosbär', 'Hand'), zonenfeld('friedhof', 'ada'))
  const weg = Array.from({ length: 50 }, (_, i) => ({ x: von.x + ((nach.x - von.x) * i) / 49 + 40 * Math.sin(i / 4), y: von.y + ((nach.y - von.y) * i) / 49 }))
  const bilder = await bildabstand(p, () => ziehe(p, weg, true))
  await loslassen(p, nach.x, nach.y)
  p.pruefe(`Bildabstand beim Ziehen: Median ≤ 17 ms, p99 ≤ 20 ms (${String(bilder.n)} Bilder)`, bilder.median <= 17 && bilder.p99 <= 20, JSON.stringify(bilder))
  await schliessen(p)
}

async function mitDerTastatur(p, z, ort) {
  const partie = begonnen(p, ort, '1', '2', false, ['Ina', 'Jan'])
  bisZug(p, ort, partie, (zug) => zug.phase === 'hauptVorKampf')
  await zumTisch(p, 'Ina · Jan', 'Ina')
  await p.klickeAuf(knopf('Als Liste'), '»Als Liste«')
  const zahl = (name, zone) => zaehle(p, ort, partie, name, zone)
  const gang = async (weg, ausdruck, tasten, wirkung) => {
    await z.beginne(weg, 'liste', 'tastatur')
    const erreicht = await tabZu(p, z, ausdruck)
    if (!erreicht) return z.ohneGang(weg, 'liste', 'tastatur', 'FEHLT', 'in 200 ⇥ nicht zu erreichen')
    for (const taste of tasten) await z.taste(...taste)
    await p.warte(`${weg} wirkt`, wirkung, 10000).catch(() => false)

    return z.ende({ wirkung: await wirkung() })
  }

  p.notiz('Nur mit der Tastatur, an der Liste')
  await gang('S3', listenkarte('Waldboden', 'Hand'), [['Enter']], () => zahl('Waldboden', 'tisch') === 1)
  await gang('S5', listenkarte('Waldboden', 'Schlachtfeld'), [['t']], () => stand(p, ort, partie).objekte.some((o) => heisst(o, 'Waldboden') && o.zone === 'tisch' && o.getappt))
  await z.beginne('S6', 'liste-menue', 'tastatur')
  if (await tabZu(p, z, listenkarte('Rindenwall', 'Hand'))) {
    await z.taste('F10', 'umschalt')
    for (let schritt = 0; schritt < 8 && !(await p.seite(fokussiert(menuepunkt('Auf den Friedhof')))); schritt += 1) await z.taste('ArrowDown')
    await z.taste('Enter')
    await p.warte('im Friedhof', () => zahl('Rindenwall', 'friedhof') === 1, 10000).catch(() => false)
    await z.ende({ wirkung: zahl('Rindenwall', 'friedhof') === 1 })
  } else z.ohneGang('S6', 'liste-menue', 'tastatur', 'FEHLT', 'die Karte ist in 200 ⇥ nicht zu erreichen')

  const leben = () => stand(p, ort, partie).sitze.find((s) => s.kennung === 'ada').leben
  const vorher = leben()
  await gang('S7', beschriftet('Ein Leben weniger — mit ⇧ fünf'), [['Enter'], ['Enter'], ['Enter']], () => leben() === vorher - 3)

  p.notiz('Die Geste: Leertaste, Pfeil, Eingabe — und Escape')
  const glut = () => stand(p, ort, partie).objekte.filter((o) => heisst(o, 'Glutläufer')).map((o) => o.zone).join(', ')
  const erreicht = await tabZu(p, z, listenkarte('Glutläufer', 'Hand'))
  const vorEscape = glut()
  await p.taste(' ')
  await p.taste('Escape')
  // **Die Ansage kommt mit dem nächsten Zeichnen** — gewartet wird auf sie.
  await p.warte('die Ansage sagt den Abbruch', async () => (await p.seite(ansage)).startsWith('Abgebrochen'), 5000).catch(() => false)
  p.pruefe('Escape bricht ab, der Glutläufer bleibt, wo er war', (await p.seite(ansage)).startsWith('Abgebrochen') && glut() === vorEscape, `${await p.seite(ansage)} — erreicht: ${String(erreicht)}, vorher: ${vorEscape}, nachher: ${glut()}`)
  await z.beginne('S6', 'liste-geste', 'tastatur')
  await tabZu(p, z, listenkarte('Moosbär', 'Hand'))
  await z.taste(' ')
  for (let schritt = 0; schritt < 8 && !(await p.seite(ansage)).includes('Friedhof'); schritt += 1) await z.taste('ArrowRight')
  await z.taste('Enter')
  await p.warte('Moosbär im Friedhof', () => zahl('Moosbär', 'friedhof') === 1, 10000).catch(() => false)
  const imFriedhof = stand(p, ort, partie).objekte.find((o) => heisst(o, 'Moosbär') && o.zone === 'friedhof')
  await z.ende({ wirkung: imFriedhof?.eigner === 'ada' })

  p.notiz('Aufgeben mit der Tastatur — und das Ergebnis statt einer Aufforderung')
  await tabZu(p, z, knopf('Aufgeben'))
  await p.taste('Enter')
  await tabZu(p, z, knopf('Ja, aufgeben'))
  await p.taste('Enter')
  // **„Das Ergebnis“ steht in Großbuchstaben** (CSS), und `innerText` liest
  // sie so — gewartet wird auf den Satz darunter.
  await p.warte('das Ergebnis steht', () => p.seite("return document.body.innerText.includes('Gewonnen hat')"), 15000).catch(() => false)
  const offen = await p.seite(`return [${['Starthand ziehen', 'Behalten', 'Passen'].map((t) => `!!${knopf(t)}`).join(', ')}]`)
  p.pruefe('die beendete Partie zeigt ihr Ergebnis', await p.seite("return document.body.innerText.includes('Gewonnen hat')"))
  p.pruefe('und keine Starthand- oder Prioritätsaufforderung', offen.every((da) => !da), JSON.stringify(offen))
  await p.bild('beendet')
}

export default async function tischhand09(p) {
  const ort = bestand(p)
  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const z = wegzaehler(p, datensatz)

  await p.starte({ ort })
  await p.groesse(1440, 900)
  await mitDerMaus(p, z, ort)
  await mitDerTastatur(p, z, ort)
  writeFileSync(join(wurzel, 'docs/belege/tischhand09-wege.json'), `${JSON.stringify(datensatz, null, 2)}\n`)
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
