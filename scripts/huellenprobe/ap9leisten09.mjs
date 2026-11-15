// **Die Kennzahlleisten über Sammlung und Deck, am nativen Fenster** — AP-9
// (#2366),
// Befund B-15, Entwurf `Bereiche09`.
//
// # Was es prüft
//
// **Sammlung**, **Deck**, **Pakete** und **Set**. Sammlung: die Leiste nennt die Exemplare und die Einträge, den Wert
// („kein Preis“ — ohne Netz gibt es keinen) und die Exemplare ohne Preis
// **wie `mse portfolio`**. „Exemplare“ nimmt eine Suche zurück, und die Liste
// zeigt wieder alle Einträge; „Wert“ klappt einen eingeklappten Navigator
// auf und führt den Fokus zur Übersicht. **Deck**: Hauptdeck mit Sideboard,
// Prüfung („ohne Format“ — mit Format fragte die Prüfung Scryfall), und was
// in der Sammlung fehlt **wie `mse deck bestand`**. „Fehlt in der Sammlung“
// schlägt die Lasche „Prüfung“ auf, wo dieselben Karten schon stehen, auch
// bei eingeklapptem Inspector; „Hauptdeck“ führt den Fokus zu den Plätzen.
// **Pakete**: installiert **wie `mse pakete`**, „Mit Befund“ filtert die
// Liste auf so viele, wie die Leiste nennt, „Installiert“ nimmt eine Suche
// zurück. **Set**: nach „Jetzt prüfen“ die Karten mit Befund **wie `mse
// probleme`**; der Klick filtert die Kartenliste und schlägt eine Karte mit
// Befund auf. Hell und dunkel.
//
// # Was es nicht belegt
//
// **Wert und Preisalter mit echten Preisen** — sie kämen aus dem Netz; die
// Einheitstests prüfen die Zeile (`sammlungskennzahlen.test.tsx`). **Ein
// Urteil gegen ein Format** — aus demselben Grund. **VoiceOver** hört ein
// Mensch.
import { execFileSync } from 'node:child_process'
import { randomUUID } from 'node:crypto'
import { join } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'

// gebuendeltesMse ruft das `mse` aus dem Bündel — `.go/bin/mse` kennt die
// mitgelieferte Saat nicht (`ap8kacheln09`, erster Lauf).
const gebuendeltesMse = (ort, ...argumente) =>
  execFileSync(join(process.cwd(), 'build/bin/MagicSetEditor4Mac.app/Contents/MacOS/mse'), argumente, { env: kindumgebung(ort) }).toString()

// deckbestand fragt `mse deck bestand` — **Ausgang 1 heißt „es fehlt etwas“**
// (`mse hilfe`), kein Fehler; `p.mse` bräche dort ab.
function deckbestand(ort, deckID) {
  try {
    return execFileSync(join(process.cwd(), '.go/bin/mse'), ['deck', 'bestand', deckID], { env: kindumgebung(ort) }).toString()
  } catch (fehler) {
    if (fehler.status === 1) return fehler.stdout.toString()
    throw fehler
  }
}

const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim() === ${JSON.stringify(name)})`
const leiste = (name) => `[...document.querySelectorAll('[role=toolbar][aria-label]')].find((u) => u.checkVisibility() && u.getAttribute('aria-label') === ${JSON.stringify(name)})`
const zahl = (name, was) => `[...(${leiste(name)})?.querySelectorAll(':scope > div') ?? []].find((l) => l.innerText.trim().startsWith(${JSON.stringify(was)}))`
const zahltext = (name, was) => `return ${zahl(name, was)}?.innerText.replace(/\\s+/g, ' ').trim() ?? null`
const ziel = (name, was) => `${zahl(name, was)}?.querySelector('button')`
const fokus = "return document.activeElement?.innerText?.trim() ?? null"
const textDa = (muster) => `return [...document.querySelectorAll('main p, main [role=status]')].some((e) => e.checkVisibility() && ${muster}.test(e.innerText.trim()))`
const lascheGewaehlt = (name) => `return [...document.querySelectorAll('[role=tab][aria-selected=true]')].some((t) => t.checkVisibility() && t.innerText.trim() === ${JSON.stringify(name)})`

async function zurSeite(p, seite) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld(seite)
  await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(seite)})`, `»${seite}« in der Palette`)
}

const paketzaehlung = "const z = [...document.querySelectorAll('[role=status]')].map((e) => e.innerText.trim()).find((t) => /^\\d+ von \\d+ Paketen?$/.test(t)); return z ? z.match(/\\d+/g).map(Number) : null"
const gewaehlteZeile = "return [...document.querySelectorAll('[role=grid] [role=row][aria-selected=true]')].some((z) => z.checkVisibility())"

async function lauf(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  // **Künstliche Exemplare ohne Preis** und ein Deck mit einem Druck, der
  // nicht in der Sammlung liegt.
  p.mse(ort, 'exemplar', randomUUID(), '2')
  p.mse(ort, 'exemplar', randomUUID(), '3')
  p.mse(ort, 'deck', 'neu', 'Waldschritt', 'probe')
  p.mse(ort, 'deck', 'legen', '1', randomUUID(), '4')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const [set] = JSON.parse(p.mse(ort, 'sets', '--json')).sets
  const probleme = JSON.parse(gebuendeltesMse(ort, 'probleme', String(set.id), '--json'))
  const installiert = JSON.parse(gebuendeltesMse(ort, 'pakete', '--json')).pakete.length
  const portfolio = JSON.parse(p.mse(ort, 'portfolio', '--json'))
  const fehlend = Number(/fehlend (\d+)/.exec(deckbestand(ort, '1'))?.[1] ?? -1)

  await p.starte({ ort })
  await p.groesse(1440, 838)

  p.notiz(`${bild}: die Leiste der Sammlung`)
  await zurSeite(p, 'Sammlung')
  const sammlung = 'Kennzahlen der Sammlung'
  await p.warte('die Leiste steht', () => p.seite(`return /ohne Preis\\s*\\d/i.test(${leiste(sammlung)}?.innerText ?? '')`), 20000)
  const exemplare = await p.seite(zahltext(sammlung, 'Exemplare'))
  p.pruefe('Exemplare und Einträge', exemplare === 'Exemplare 5 2 Einträge', exemplare)
  const wert = await p.seite(zahltext(sammlung, 'Wert'))
  p.pruefe('kein Wert ohne Preis, und kein Trend', wert === 'Wert kein Preis', wert)
  const ohne = await p.seite(zahltext(sammlung, 'Ohne Preis'))
  p.pruefe('ohne Preis wie `mse portfolio`', ohne === `Ohne Preis ${portfolio.ohnePreis}`, `${ohne} — Kommandozeile: ${portfolio.ohnePreis}`)
  await p.bild(`sammlung-${bild}`)

  // **Eine Tabstelle für die Leiste** — sonst verlöre der Kernweg
  // `sammlung` sein Ziel „≤ 13. Tabstelle“ für das Suchfeld.
  const tabstellen = await p.seite(`return [...(${leiste(sammlung)})?.querySelectorAll('button') ?? []].filter((b) => b.tabIndex === 0).length`)
  p.pruefe('die Leiste ist eine Tabstelle', tabstellen === 1, String(tabstellen))
  await p.seite(`${ziel(sammlung, 'Exemplare')}.focus(); return true`)
  await p.taste('ArrowRight')
  p.pruefe('→ geht in der Leiste weiter', /^Wert/.test((await p.seite(fokus)) ?? ''), await p.seite(fokus))

  await p.seite("const f = [...document.querySelectorAll('input[type=search]')].find((e) => e.checkVisibility()); f.focus(); return true")
  await p.tippeInsFeld('gibt-es-nicht')
  await p.warte('die Suche filtert', () => p.seite(textDa('/^0 von 2/')), 10000)
  await p.klickeAuf(ziel(sammlung, 'Exemplare'), '»Exemplare« in der Leiste')
  await p.warte('die Liste zeigt wieder alle Einträge', () => p.seite(textDa('/^2 von 2/')), 10000)
  p.pruefe('die Suche ist zurückgenommen', await p.seite("return [...document.querySelectorAll('input[type=search]')].find((e) => e.checkVisibility())?.value === ''"))

  await p.klickeAuf(knopf('Navigator ausblenden'), '»Navigator ausblenden«')
  await p.warte('der Navigator ist eingeklappt', () => p.seite(`return !!${knopf('Navigator einblenden')}`), 10000)
  await p.klickeAuf(ziel(sammlung, 'Wert'), '»Wert« in der Leiste')
  await p.warte('der Fokus steht an der Übersicht', async () => (await p.seite(fokus)) === 'Übersicht', 10000)
  p.pruefe('der Navigator ist wieder offen', await p.seite(`return !!${knopf('Navigator ausblenden')}`))

  p.notiz(`${bild}: die Leiste des Decks`)
  await p.taste('4', 'befehl')
  await p.warte('die Deckliste steht', () => p.seite(`return !!${knopf('Öffnen: Waldschritt')}`), 20000)
  await p.klickeAuf(knopf('Öffnen: Waldschritt'), '»Öffnen: Waldschritt«')
  const deck = 'Kennzahlen des Decks'
  await p.warte('die Leiste steht', () => p.seite(`return /Fehlt in der Sammlung\\s*\\d/.test(${leiste(deck)}?.innerText ?? '')`), 20000)
  const hauptdeck = await p.seite(zahltext(deck, 'Hauptdeck'))
  p.pruefe('Hauptdeck mit Sideboard', hauptdeck === 'Hauptdeck 4 Sideboard 0', hauptdeck)
  const pruefung = await p.seite(zahltext(deck, 'Prüfung'))
  p.pruefe('ohne Format kein Urteil', pruefung === 'Prüfung ohne Format', pruefung)
  const fehlt = await p.seite(zahltext(deck, 'Fehlt in der Sammlung'))
  p.pruefe('was fehlt, wie `mse deck bestand`', fehlt === `Fehlt in der Sammlung ${fehlend}`, `${fehlt} — Kommandozeile: ${fehlend}`)
  await p.bild(`deck-${bild}`)

  await p.klickeAuf(knopf('Inspector ausblenden'), '»Inspector ausblenden«')
  await p.warte('der Inspector ist eingeklappt', () => p.seite(`return !!${knopf('Inspector einblenden')}`), 10000)
  await p.klickeAuf(ziel(deck, 'Fehlt in der Sammlung'), '»Fehlt in der Sammlung« in der Leiste')
  await p.warte('die Lasche „Prüfung“ ist aufgeschlagen', () => p.seite(lascheGewaehlt('Prüfung')), 10000)
  await p.warte('dort stehen die fehlenden Karten', () => p.seite(`return [...document.querySelectorAll('p')].some((e) => e.checkVisibility() && e.innerText.trim() === '${fehlend} Karten fehlen.')`), 10000)
  p.pruefe('der Inspector ist wieder offen', await p.seite(`return !!${knopf('Inspector ausblenden')}`))
  await p.bild(`pruefung-${bild}`)

  await p.klickeAuf(ziel(deck, 'Hauptdeck'), '»Hauptdeck« in der Leiste')
  await p.warte('der Fokus steht an den Plätzen', async () => (await p.seite(fokus)) === 'Plätze', 10000)

  p.notiz(`${bild}: die Leiste der Pakete`)
  await zurSeite(p, 'Pakete')
  const pakete = 'Kennzahlen der Pakete'
  await p.warte('die Leiste steht', () => p.seite(`return /Installiert\\s*\\d/.test(${leiste(pakete)}?.innerText ?? '') && /Mit Befund\\s*\\d/.test(${leiste(pakete)}?.innerText ?? '')`), 20000)
  const anzahl = await p.seite(zahltext(pakete, 'Installiert'))
  p.pruefe('installiert wie `mse pakete`', anzahl === `Installiert ${installiert}`, `${anzahl} — Kommandozeile: ${installiert}`)
  const mitBefund = Number(/\d+/.exec((await p.seite(zahltext(pakete, 'Mit Befund'))) ?? '')?.[0] ?? -1)
  await p.bild(`pakete-${bild}`)
  if (mitBefund > 0) {
    await p.klickeAuf(ziel(pakete, 'Mit Befund'), '»Mit Befund« in der Leiste')
    const gefiltert = await p.warte('die Liste ist gefiltert', async () => { const z = await p.seite(paketzaehlung); return z !== null && z[0] === mitBefund ? z : null }, 15000)
    p.pruefe('die Liste zeigt so viele Pakete, wie die Leiste nennt', gefiltert[0] === mitBefund, gefiltert.join(' von '))
  } else {
    p.pruefe('null Befunde führen nirgends hin', !(await p.seite(`return !!${ziel(pakete, 'Mit Befund')}`)))
  }
  await p.seite("const f = [...document.querySelectorAll('input[type=search]')].find((e) => e.checkVisibility()); f.focus(); return true")
  await p.tippeInsFeld('gibt-es-nicht')
  await p.warte('die Suche filtert', async () => (await p.seite(paketzaehlung))?.[0] === 0, 10000)
  await p.klickeAuf(ziel(pakete, 'Installiert'), '»Installiert« in der Leiste')
  const alle = await p.warte('alle stehen wieder da', async () => { const z = await p.seite(paketzaehlung); return z !== null && z[0] === z[1] && z[0] > 0 ? z : null }, 10000)
  p.pruefe('„Installiert“ nimmt die Suche zurück', alle[0] === installiert, alle.join(' von '))

  p.notiz(`${bild}: die Leiste des Sets, nach „Jetzt prüfen“`)
  await p.taste('1', 'befehl')
  await p.klickeAuf(knopf('Jetzt prüfen'), '»Jetzt prüfen«')
  await p.warte('die Prüfung ist fertig', () => p.seite(textDa('/^Probleme geprüft /')), 120000)
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  const setleiste = 'Kennzahlen des Sets'
  await p.warte('die Leiste steht', () => p.seite(`return /Mit Befund\\s*\\d/.test(${leiste(setleiste)}?.innerText ?? '')`), 20000)
  const karten = await p.seite(zahltext(setleiste, 'Karten'))
  p.pruefe('Karten des Sets', karten === `Karten ${probleme.gesamt}`, karten)
  const befund = await p.seite(zahltext(setleiste, 'Mit Befund'))
  p.pruefe('mit Befund wie `mse probleme`', new RegExp(`^Mit Befund ${probleme.zahl} geprüft `).test(befund ?? ''), `${befund} — Kommandozeile: ${probleme.zahl}`)
  await p.bild(`set-${bild}`)
  await p.klickeAuf(ziel(setleiste, 'Mit Befund'), '»Mit Befund« in der Leiste')
  const kartentext = `${probleme.zahl} ${probleme.zahl === 1 ? 'Karte' : 'Karten'} mit Befund`
  await p.warte('die Kartenliste ist gefiltert', () => p.seite(textDa(`/^${kartentext}$/`)), 15000)
  await p.warte('aufgeschlagen ist eine Karte mit Befund', () => p.seite(gewaehlteZeile), 10000)
  p.pruefe('⌘Q beendet', await p.beende())
}

export default async function ap9leisten09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
