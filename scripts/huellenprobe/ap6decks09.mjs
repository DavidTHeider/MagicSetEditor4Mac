// **Das Deck und seine Plätze hinter »…« und im Rechtsklick — am nativen
// Fenster** — AP-6, Teil 1
// (#2363),
// Entwurf `Bereiche09`.
//
// # Was es prüft
//
// Am offenen Deck steht „…“ statt Umbenennen, Duplizieren, Teilen und
// Löschen, an der aktiven Platzzeile „…“ statt −, +, Bereich und
// Herausnehmen; beide Menüs nennen die Handlungen und stehen ganz im Bild.
// Ein Rechtsklick auf ein **anderes** Deck öffnet es und zeigt **sein**
// Menü; „Eine mehr“ aus dem Menü des Platzes zählt die Menge im Bestand
// herauf (`mse deck`). Hell und dunkel.
//
// # Was es nicht belegt
//
// **Höchstens zwanzig Bedienelemente im Bild** — das zählt
// `uirevue09zustaende`, und es ist mit diesem Teil noch nicht erreicht.
// **VoiceOver** hört ein Mensch.
import { rechtsklick } from '../huellenrufbuch.mjs'

const menueOffen = "return [...document.querySelectorAll('[role=menu]')].some((m) => m.checkVisibility())"
const sichtbareEintraege = "return [...document.querySelectorAll('[role=menu] [role=menuitem]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.replace(/\\s+/g, ' ').trim())"
const ganzZuSehen = "const m = [...document.querySelectorAll('[role=menu]')].find((m) => m.checkVisibility()); if (!m) return null; const r = m.getBoundingClientRect(); return [[r.left + 3, r.top + 3], [r.right - 3, r.top + 3], [r.left + 3, r.bottom - 3], [r.right - 3, r.bottom - 3]].every(([x, y]) => m.contains(document.elementFromPoint(x, y)))"
const mitteVon = (ausdruck) => `const e = ${ausdruck}; if (!e) return null; e.scrollIntoView({ block: 'center' }); const r = e.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }`
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim() === ${JSON.stringify(name)})`
const textvon = (text) => `[...document.querySelectorAll('span')].find((e) => e.checkVisibility() && e.innerText.trim() === ${JSON.stringify(text)})`

async function menuepruefung(p, wo, bild, erwartet) {
  await p.warte(`das Menü an ${wo} ist offen`, () => p.seite(menueOffen), 10000)
  const eintraege = await p.seite(sichtbareEintraege)
  p.pruefe(`das Menü an ${wo} nennt ${erwartet.join(', ')}`, erwartet.every((wort) => eintraege.some((e) => e.toLowerCase().includes(wort))), eintraege.join(' · '))
  p.pruefe(`das Menü an ${wo} ist ganz zu sehen`, (await p.seite(ganzZuSehen)) === true)
  await p.bild(`menue-${wo.replace(/[^\p{L}\p{N}]+/gu, '-')}-${bild}`)

  return eintraege
}

async function lauf(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Waldschritt', 'probe')
  p.mse(ort, 'deck', 'neu', 'Glutpfad', 'probe')
  for (const karte of ['1', '2']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '2')
  const menge = () => JSON.parse(p.mse(ort, 'deck', '1', '--json')).deck.plaetze.reduce((summe, platz) => summe + platz.menge, 0)

  await p.starte({ ort })
  await p.groesse(1440, 838)
  await p.taste('4', 'befehl')
  await p.warte('die Deckliste steht', () => p.seite(`return !!${knopf('Öffnen: Waldschritt')}`), 20000)

  p.notiz(`${bild}: das offene Deck trägt »…« statt vier Knöpfen`)
  await p.klickeAuf(knopf('Öffnen: Waldschritt'), '»Öffnen: Waldschritt«')
  await p.warte('»…« am offenen Deck', () => p.seite(`return !!${knopf('Handlungen — Waldschritt')}`), 20000)
  for (const alt of ['Umbenennen: Waldschritt', 'Duplizieren: Waldschritt', 'Deck löschen: Waldschritt']) {
    p.pruefe(`»${alt}« steht nicht mehr als Knopf da`, !(await p.seite(`return !!${knopf(alt)}`)))
  }
  await p.klickeAuf(knopf('Handlungen — Waldschritt'), '»…« am Deck')
  await menuepruefung(p, 'dem offenen Deck', bild, ['umbenennen', 'duplizieren', 'teilen', 'löschen'])
  await p.taste('Escape')

  p.notiz(`${bild}: der Platz trägt »…« statt −, +, Bereich und Herausnehmen`)
  const platzmenue = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '').startsWith('Plätze: '))"
  await p.warte('»…« am aktiven Platz', () => p.seite(`return !!${platzmenue}`), 20000)
  p.pruefe('kein »Eine mehr« als Knopf in der Zeile', !(await p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '').startsWith('Eine mehr'))")))
  const vorher = menge()
  await p.klickeAuf(platzmenue, '»…« am Platz')
  await menuepruefung(p, 'dem Platz', bild, ['eine mehr', 'eine weniger', 'sideboard', 'herausnehmen'])
  await p.klickeAuf("[...document.querySelectorAll('[role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.trim().startsWith('Eine mehr'))", '»Eine mehr«')
  await p.warte('die Menge steigt im Bestand', () => menge() === vorher + 1, 15000).catch(() => false)
  p.pruefe('»Eine mehr« aus dem Menü zählt im Bestand herauf', menge() === vorher + 1, `${String(vorher)} → ${String(menge())}`)

  p.notiz(`${bild}: Rechtsklick auf ein anderes Deck`)
  const stelle = await p.seite(mitteVon(textvon('Glutpfad')))
  p.pruefe('das andere Deck steht da', !!stelle, JSON.stringify(stelle))
  await rechtsklick(p, stelle.x, stelle.y)
  await p.warte('das andere Deck geht auf', () => p.seite(`return !!${knopf('Handlungen — Glutpfad')}`), 20000)
  const objekt = await p.warte('sein Menü geht auf', () => p.seite("return [...document.querySelectorAll('[role=menu]')].find((m) => m.checkVisibility())?.closest('[data-objekt]')?.getAttribute('data-objekt') ?? null"), 10000).catch(() => null)
  p.pruefe('das Menü gehört dem anderen Deck', objekt === 'deck/2', String(objekt))
  await menuepruefung(p, 'dem anderen Deck', bild, ['umbenennen', 'löschen'])
  await p.taste('Escape')

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}

export default async function ap6decks09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
