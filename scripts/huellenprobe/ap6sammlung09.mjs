// **Die Zeile der Sammlung im Rechtsklick und das Menü „Exemplar“ — am
// nativen Fenster** — AP-6, Teil 2
// (#2363),
// Abnahme „Kontext an der Sammlungszeile mit Ins Deck, Zum Tausch, Menge,
// Löschen“.
//
// # Was es prüft
//
// Ein Rechtsklick auf eine **andere** Zeile schlägt sie im Inspector auf und
// zeigt ihr Menü — Bearbeiten (dort die Menge), Löschen, Zum Tausch, Ins
// Deck —, ganz im Bild. Löschen stellt die Rückfrage im Inspector und
// löscht nichts (`mse sammlung`). ⇧F10 öffnet es an der gehaltenen Zeile,
// Escape gibt ihr den Fokus zurück. Die Menüleiste führt dasselbe Menü
// „Exemplar“; „Ins Deck“ daraus führt zu den Decks, und die Leiste verliert
// das Menü mit der Seite. Hell und dunkel.
//
// # Was es nicht belegt
//
// **Die Leserrolle** prüfen die Einheitstests (`useExemplarhandlungen`,
// `schreibknoepfe`). **VoiceOver** hört ein Mensch.
import { randomUUID } from 'node:crypto'

import { rechtsklick } from '../huellenrufbuch.mjs'

const menueOffen = "return [...document.querySelectorAll('[role=menu]')].some((m) => m.checkVisibility())"
const sichtbareEintraege = "return [...document.querySelectorAll('[role=menu] [role=menuitem]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.replace(/\\s+/g, ' ').trim())"
const ganzZuSehen = "const m = [...document.querySelectorAll('[role=menu]')].find((m) => m.checkVisibility()); if (!m) return null; const r = m.getBoundingClientRect(); return [[r.left + 3, r.top + 3], [r.right - 3, r.top + 3], [r.left + 3, r.bottom - 3], [r.right - 3, r.bottom - 3]].every(([x, y]) => m.contains(document.elementFromPoint(x, y)))"
const zeile = (text) => `[...document.querySelectorAll('[role=row]')].find((z) => z.checkVisibility() && z.innerText.includes(${JSON.stringify(text)}))`
const mitteVon = (ausdruck) => `const e = ${ausdruck}; if (!e) return null; e.scrollIntoView({ block: 'center' }); const r = e.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }`
const inspectorZeigt = (text) => `return [...document.querySelectorAll('aside, [role=complementary]')].some((a) => a.checkVisibility() && a.innerText.includes(${JSON.stringify(text)}))`
const eintrag = (name) => `[...document.querySelectorAll('[role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.includes(${JSON.stringify(name)}))`
const leistenmenue = async (p, titel) => (await p.menue()).find((m) => m.titel === titel)
const leistennamen = (menue) => (menue?.unter ?? []).filter((e) => !e.trenner && !e.verborgen).map((e) => e.titel)

async function zurSammlung(p) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld('Sammlung')
  await p.klickeAuf("[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Sammlung')", '»Sammlung« in der Palette')
}

async function lauf(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  const erster = randomUUID()
  const zweiter = randomUUID()
  p.mse(ort, 'exemplar', erster, '2')
  p.mse(ort, 'exemplar', zweiter, '3')
  const anzahl = () => JSON.parse(p.mse(ort, 'sammlung', '--json')).exemplare.length

  await p.starte({ ort })
  await p.groesse(1440, 838)
  await zurSammlung(p)
  await p.warte('die Liste steht', () => p.seite(`return !!${zeile(zweiter.slice(0, 8))}`), 20000)

  p.notiz(`${bild}: Rechtsklick auf eine Zeile`)
  const stelle = await p.seite(mitteVon(zeile(zweiter.slice(0, 8))))
  await rechtsklick(p, stelle.x, stelle.y)
  await p.warte('das Menü ist offen', () => p.seite(menueOffen), 10000)
  const eintraege = await p.seite(sichtbareEintraege)
  p.pruefe('das Menü nennt Bearbeiten, Löschen, Zum Tausch und Ins Deck', ['Bearbeiten', 'Löschen', 'Zum Tausch', 'Ins Deck'].every((name) => eintraege.some((e) => e.includes(name))) && eintraege.length === 4, eintraege.join(' · '))
  p.pruefe('das Menü ist ganz zu sehen', (await p.seite(ganzZuSehen)) === true)
  p.pruefe('der Inspector zeigt die angeklickte Zeile', await p.seite(inspectorZeigt(zweiter)), zweiter)
  await p.bild(`menue-zeile-${bild}`)

  p.notiz(`${bild}: Löschen fragt im Inspector`)
  await p.klickeAuf(eintrag('Löschen'), '»Löschen« im Menü')
  await p.warte('die Rückfrage steht', () => p.seite(inspectorZeigt('Dieses Exemplar löschen?')), 10000)
  p.pruefe('Löschen hat nichts gelöscht', anzahl() === 2, String(anzahl()))
  await p.bild(`rueckfrage-${bild}`)
  await p.klickeAuf("[...document.querySelectorAll('aside button, [role=complementary] button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Abbrechen')", '»Abbrechen«')

  p.notiz(`${bild}: ⇧F10 an der gehaltenen Zeile, Escape gibt ihr den Fokus zurück`)
  await p.seite(`${zeile(zweiter.slice(0, 8))}.focus(); return true`)
  await p.taste('F10', 'umschalt')
  await p.warte('⇧F10 öffnet das Menü', () => p.seite(menueOffen), 10000)
  p.pruefe('⇧F10: der Fokus steht im Menü', await p.seite("return document.activeElement?.closest('[role=menu]') !== null"))
  await p.taste('Escape')
  await p.warte('das Menü ist zu', async () => !(await p.seite(menueOffen)), 10000)
  p.pruefe('Escape gibt der Zeile den Fokus zurück', await p.seite(`return document.activeElement === ${zeile(zweiter.slice(0, 8))}`))

  p.notiz(`${bild}: das Menü »Exemplar« der Leiste`)
  const leiste = await p.warte('die Leiste führt »Exemplar«', () => leistenmenue(p, 'Exemplar'), 10000)
  p.pruefe('die Leiste nennt dieselben vier', JSON.stringify(leistennamen(leiste)) === JSON.stringify(['Bearbeiten', 'Löschen', 'Zum Tausch', 'Ins Deck']), leistennamen(leiste).join(' · '))
  await p.klick('Exemplar', 'Ins Deck')
  await p.warte('die Decks stehen da', () => p.seite("return [...document.querySelectorAll('main h1, main h2')].some((h) => h.checkVisibility() && h.innerText.trim() === 'Decks')"), 15000)
  await p.warte('die Leiste verliert »Exemplar« mit der Seite', async () => !(await leistenmenue(p, 'Exemplar')), 10000)
  await p.bild(`ins-deck-${bild}`)
  p.pruefe('⌘Q beendet', await p.beende())
}

export default async function ap6sammlung09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
