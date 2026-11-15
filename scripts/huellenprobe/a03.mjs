// **Kontraste in den Überlagerungen, hell und dunkel** — A03
// (#2217),
// Befund F-03 des Audits zu 0.8.0.
//
// # Was der Audit sah
//
// Im Gast unter macOS 15.2 stand „Erste Schritte" als dunkle Fläche mit
// **schwarzer** Überschrift da. Vermutet war ein älteres WebKit.
//
// # Was hier gemessen wird
//
// **Jeder sichtbare Text** in den Überlagerungen, die Mantine an `body`
// hängt — alle drei Arten, die die Anwendung benutzt: ein Dialog (die ersten
// Schritte, die Befehlspalette), eine Schublade (Paket beziehen) und ein
// Popover (ein Deck umbenennen) —, je
// Erscheinungsbild gegen seinen gerenderten Grund (`huellenkontrast.mjs`):
// 4,5 : 1, große Schrift 3 : 1. Nicht eine Liste ausgewählter Elemente,
// denn genau die Überschrift ohne eigene Farbklasse stand auf keiner.
import { rechnung, texteMessen } from '../huellenkontrast.mjs'

// kantenMessen misst jede sichtbare Kante unter `wurzel`, die eine Breite hat
// — den Rahmen der Überlagerung, Eingabefelder, Knöpfe, Trennlinien — gegen
// den Grund daneben: 3 : 1 (WCAG 1.4.11). Befund 7 des Zweitblicks über Phase A:
// bis dahin maß das Drehbuch nur Text; **und alle vier Seiten**, denn eine
// Trennlinie steht oft unten (Nachlauf über Phase A).
const kantenMessen = (p, wurzel) => p.seite(`${rechnung}
  const wurzeln = [...document.querySelectorAll(${JSON.stringify(wurzel)})].filter((w) => w.checkVisibility())
  // **Jede sichtbare Kante mit Farbe** — Rahmen, Felder, Knöpfe und
  // Trennlinien; die Leitlinie verlangt 3 : 1 für alle (§3, Sichtprüfung).
  const seiten = ['Top', 'Right', 'Bottom', 'Left']
  const mitKante = (e, seite) => {
    const stil = getComputedStyle(e)
    return parseFloat(stil['border' + seite + 'Width']) > 0 && alsRGBA(stil['border' + seite + 'Color'])[3] > 0
  }
  const name = (e) => 'Kante ' + e.tagName.toLowerCase() + (e.getAttribute('aria-label') ? ' »' + e.getAttribute('aria-label').slice(0, 24) + '«' : '')
  return wurzeln.flatMap((w) => [w, ...w.querySelectorAll('*')]).filter((e) => e.checkVisibility())
    .flatMap((e) => seiten.filter((seite) => mitKante(e, seite)).map((seite) => messe(name(e) + ' ' + seite.toLowerCase(), e, 'rand', 3, seite)))
    .filter((m) => !m.fehlt)`)

// eingabenMessen misst, was in den Feldern steht — den Wert, und wo keiner
// steht, den Platzhalter — gegen den Grund des Felds: 4,5 : 1. Nachlauf über
// Phase A: ein Feld hat keinen Textknoten, und `texteMessen` übersah es.
const eingabenMessen = (p, wurzel) => p.seite(`${rechnung}
  const wurzeln = [...document.querySelectorAll(${JSON.stringify(wurzel)})].filter((w) => w.checkVisibility())
  const felder = wurzeln.flatMap((w) => [...w.querySelectorAll('input:not([type=hidden]):not([type=checkbox]):not([type=radio]), textarea')]).filter((e) => e.checkVisibility())
  return felder.map((e) => {
    const name = 'Feld »' + (e.getAttribute('aria-label') ?? e.placeholder ?? e.name ?? '').slice(0, 24) + '«'
    if (e.value !== '') return messe(name + ' (Wert)', e, 'text', 4.5)
    return e.placeholder ? messe(name + ' (Platzhalter)', e, 'platzhalter', 4.5) : null
  }).filter((m) => m !== null)`)

// symboleMessen misst jedes sichtbare Symbol eines Bedienelements — das ×
// einer Schublade, ein Pfeil — an seiner ersten gezeichneten Form: 3 : 1
// (WCAG 1.4.11). **Am 2026-10-09 trug keine der vier Überlagerungen eines** —
// das × der Schublade ist ein eigener Knopf mit Zeichen (`Schublade.tsx`) —;
// die Prüfung steht für das erste, das dazukommt, und meldet sich nur dann.
const symboleMessen = (p, wurzel) => p.seite(`${rechnung}
  const wurzeln = [...document.querySelectorAll(${JSON.stringify(wurzel)})].filter((w) => w.checkVisibility())
  const symbole = wurzeln.flatMap((w) => [...w.querySelectorAll('button svg, a svg, [role=img]')]).filter((e) => e.checkVisibility())
  const name = (e) => 'Symbol ' + (e.closest('button, a')?.getAttribute('aria-label') ?? e.getAttribute('aria-label') ?? '').slice(0, 24)
  return symbole.map((e) => messe(name(e), e.querySelector('path, circle, rect, line, polyline, polygon') ?? e, 'symbol', 3))`)

async function pruefeUeberlagerung(p, wahl, was, wurzel) {
  const messungen = await texteMessen(p, wurzel)
  const zuSchwach = messungen.filter((m) => m.fehlt || m.verhaeltnis < m.grenze)
  const kleinstes = Math.min(...messungen.map((m) => m.verhaeltnis ?? 0))
  p.pruefe(`${wahl}: ${was} — ${String(messungen.length)} Texte gemessen`, messungen.length > 0)
  p.pruefe(`${wahl}: ${was} — jeder Text hält seine Grenze (kleinstes ${String(kleinstes).replace('.', ',')} : 1)`,
    zuSchwach.length === 0, zuSchwach.map((m) => `${m.name} ${String(m.verhaeltnis).replace('.', ',')}`).join('; '))
  const felder = await eingabenMessen(p, wurzel)
  const schwacheFelder = felder.filter((m) => m.fehlt || m.verhaeltnis < m.grenze)
  if (felder.length > 0) {
    p.pruefe(`${wahl}: ${was} — ${String(felder.length)} Felder halten 4,5 : 1 in Wert und Platzhalter`,
      schwacheFelder.length === 0, schwacheFelder.map((m) => `${m.name} ${m.fehlt ? 'nicht messbar' : String(m.verhaeltnis).replace('.', ',')}`).join('; '))
  }
  const symbole = await symboleMessen(p, wurzel)
  const schwacheSymbole = symbole.filter((m) => m.fehlt || m.verhaeltnis < m.grenze)
  if (symbole.length > 0) {
    p.pruefe(`${wahl}: ${was} — ${String(symbole.length)} Symbole halten 3 : 1`,
      schwacheSymbole.length === 0, schwacheSymbole.map((m) => `${m.name} ${m.fehlt ? 'nicht messbar' : String(m.verhaeltnis).replace('.', ',')}`).join('; '))
  }
  // **Und Mantine steht, wie das Fenster steht** — dieselbe Gegenprobe.
  p.pruefe(`${wahl}: Mantine folgt dem Erscheinungsbild`,
    (await p.seite('return document.documentElement.dataset.mantineColorScheme')) === (wahl === 'dunkel' ? 'dark' : 'light'))
  const kanten = await kantenMessen(p, wurzel)
  const schwacheKanten = kanten.filter((m) => m.verhaeltnis < m.grenze)
  p.pruefe(`${wahl}: ${was} — ${String(kanten.length)} Kanten halten 3 : 1 (kleinste ${String(Math.min(...kanten.map((m) => m.verhaeltnis))).replace('.', ',')} : 1)`,
    kanten.length > 0 && schwacheKanten.length === 0, schwacheKanten.map((m) => `${m.name} ${String(m.verhaeltnis).replace('.', ',')}`).join('; '))
}

// oeffneSeite geht über die Befehlspalette zu einer Seite.
async function oeffneSeite(p, name) {
  await p.taste('k', 'befehl')
  await p.warte('die Befehlspalette steht da', () => p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())"), 20000)
  const zeile = `[...document.querySelectorAll('[data-palettenzeile]')].find((z) => z.innerText.trim().split('\\n')[0] === ${JSON.stringify(name)})`
  await p.seite(`(${zeile}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(zeile, `»${name}« in der Palette`)
  await p.warte(`die Seite »${name}« steht da`, () => p.seite(`return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() === ${JSON.stringify(name)}`), 20000)
}

// knopfDruecken drückt den sichtbaren Knopf mit diesem Text.
async function knopfDruecken(p, text) {
  const knopf = `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(text)})`
  await p.warte(`»${text}« steht da`, () => p.seite(`return !!(${knopf})`), 20000)
  await p.seite(`(${knopf}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(knopf, `»${text}«`)
}

export default async function a03(p) {
  for (const wahl of ['hell', 'dunkel']) {
    const ort = p.ort()
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', wahl)
    p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    p.mse(ort, 'deck', 'neu', 'Kontrastdeck', 'probe')
    await p.starte({ ort })
    p.notiz(`Erscheinungsbild ${wahl}`)
    p.pruefe(`${wahl}: das Erscheinungsbild ist gesetzt`, (await p.seite('return document.documentElement.dataset.erscheinungsbild ?? ""')) === wahl)

    // **Die ersten Schritte, aus der Hilfe zurückgeholt** — beim Start
    // schließt die Probe sie (`bereit`).
    await p.taste('?', 'befehl')
    await p.wartetext('Führungen', 20000)
    await p.klickeText('Führungen')
    await p.wartetext('Die ersten Schritte jetzt zeigen', 20000)
    await p.klickeText('Die ersten Schritte jetzt zeigen')
    await p.warte('die ersten Schritte stehen da', () => p.seite("return !!document.querySelector('[data-einfuehrung]')?.checkVisibility()"), 20000)
    await p.bild(`erste-schritte-${wahl}`)
    await pruefeUeberlagerung(p, wahl, 'die ersten Schritte', '[data-einfuehrung]')
    await p.taste('Escape')
    await p.warte('die ersten Schritte sind zu', async () => !(await p.seite("return !!document.querySelector('[data-einfuehrung]')?.checkVisibility()")))

    await p.taste('k', 'befehl')
    await p.warte('die Befehlspalette steht da', () => p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())"), 20000)
    await p.bild(`palette-${wahl}`)
    await pruefeUeberlagerung(p, wahl, 'die Befehlspalette', '[role=dialog]')
    await p.taste('Escape')
    await p.warte('die Befehlspalette ist zu', async () => !(await p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())")))

    // **Die Schublade** — über die Palette zu den Paketen.
    await oeffneSeite(p, 'Pakete')
    await knopfDruecken(p, 'Paket beziehen …')
    await p.warte('die Schublade steht da', () => p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())"), 20000)
    await p.bild(`schublade-${wahl}`)
    await pruefeUeberlagerung(p, wahl, 'die Schublade »Paket beziehen«', '[role=dialog]')
    await p.klickeAuf("document.querySelector('[role=dialog] button[aria-label^=Schließen]')", 'das × der Schublade')
    await p.warte('die Schublade ist zu', async () => !(await p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())")))

    // **Das Popover** — ein Deck umbenennen.
    await oeffneSeite(p, 'Decks')
    await p.warte('das Deck steht da', () => p.seite("return document.body.innerText.includes('Kontrastdeck')"), 20000)
    await knopfDruecken(p, 'Öffnen')
    // **Umbenennen steht seit AP-6 (#2363) im Menü des offenen Decks** —
    // der Knopf „Umbenennen: …“ daneben gibt es nicht mehr.
    const menue = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.getAttribute('aria-label') === 'Handlungen — Kontrastdeck')"
    await p.warte('das Menü des Decks steht da', () => p.seite(`return !!(${menue})`), 20000)
    await p.seite(`(${menue}).scrollIntoView({ block: 'center' }); return true`)
    await p.klickeAuf(menue, 'das Menü am Kontrastdeck')
    const umbenennen = "[...document.querySelectorAll('[role=menuitem]')].find((m) => m.checkVisibility() && m.innerText.includes('Umbenennen'))"
    await p.warte('»Umbenennen« steht im Menü', () => p.seite(`return !!(${umbenennen})`), 10000)
    await p.klickeAuf(umbenennen, '»Umbenennen« im Menü')
    // **Eine Gruppe und kein Dialog** (`Popover.tsx`) — erkannt an Mantines
    // Klasse, die auch `unstyled` trägt.
    await p.warte('das Popover steht da', () => p.seite("return [...document.querySelectorAll('.mantine-Popover-dropdown')].some((d) => d.checkVisibility())"), 20000)
    await p.bild(`popover-${wahl}`)
    await pruefeUeberlagerung(p, wahl, 'das Popover »Umbenennen«', '.mantine-Popover-dropdown')
    await p.taste('Escape')

    p.pruefe(`${wahl}: ⌘Q beendet`, await p.beende())
  }
}
