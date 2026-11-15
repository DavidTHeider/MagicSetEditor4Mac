// **Die Befunde am Objekt, am nativen Fenster** — AP-8, Teil 2b
// (#2365),
// Befunde B-13 und B-14, Entwurf `Start09`.
//
// # Was es prüft
//
// **Die Kachel „Sammlung“** nennt die Exemplare ohne Preis — **dieselbe Zahl
// wie `mse portfolio`**; Preise holt die Probe nicht, denn sie kämen aus dem
// Netz. **Die Kachel „Decks“** nennt die Decks mit fehlenden Karten — zwei
// von drei, wie `mse deck bestand` sie sieht — und führt zur Deckliste mit
// genau ihnen; der Filter nennt sich und lässt sich lösen. **Nach „Jetzt
// prüfen“** steht das Set mit Befund in der Aufmerksamkeit, und **die Kachel
// „Karten“** führt in seine Kartenliste, gefiltert auf die Karten mit Befund
// — so viele, wie das gebündelte `mse probleme` zählt, und aufgeschlagen
// ist eine von ihnen; „Alle Karten zeigen“
// gibt alle zurück. **Die Zeile der Aufmerksamkeit** tut dasselbe über die
// Tastatur (Fokus und ↩). Nichts verschiebt sich, wenn die Unterzeilen
// kommen. Hell und dunkel.
//
// # Was es nicht belegt
//
// **Wert und Alter des ältesten Preises** — ohne Netz gibt es keinen Preis;
// die Einheitstests prüfen die Zeile (`sammlungskachel.test.ts`).
// **Den Tabweg** zur Zeile — der Fokus wird gesetzt, nicht erwandert.
// **VoiceOver** hört ein Mensch.
import { execFileSync } from 'node:child_process'
import { randomUUID } from 'node:crypto'
import { join } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'

// gebuendeltesMse ruft das `mse` **aus dem Bündel** — `.go/bin/mse` kennt
// die mitgelieferte Saat nicht (`ap8kacheln09`, erster Lauf).
const gebuendeltesMse = (ort, ...argumente) =>
  execFileSync(join(process.cwd(), 'build/bin/MagicSetEditor4Mac.app/Contents/MacOS/mse'), argumente, { env: kindumgebung(ort) }).toString()

const liste = "[...document.querySelectorAll('ul[aria-label]')].find((u) => u.checkVisibility() && u.getAttribute('aria-label') === 'Dein Überblick')"
const kachel = (wort) => `[...(${liste})?.querySelectorAll('button') ?? []].find((b) => b.innerText.replace(/\\s+/g, ' ').includes(${JSON.stringify(wort)}))`
const kacheltext = (wort) => `return ${kachel(wort)}?.innerText.replace(/\\s+/g, ' ').trim() ?? null`
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim() === ${JSON.stringify(name)})`
const textDa = (muster) => `return [...document.querySelectorAll('main p')].some((e) => e.checkVisibility() && ${muster}.test(e.innerText.trim()))`
const zeile = (muster) => `[...document.querySelectorAll('section[aria-label="Was Aufmerksamkeit braucht"] button')].find((b) => b.checkVisibility() && ${muster}.test(b.innerText.replace(/\\s+/g, ' ').trim()))`
const zeilen = "return Number([...document.querySelectorAll('[role=grid]')].find((g) => g.checkVisibility() && g.getAttribute('aria-label') === 'Karten')?.getAttribute('aria-rowcount') ?? -1)"
const sichtbar = (name) => `return [...document.querySelectorAll('main *')].some((e) => e.children.length === 0 && e.checkVisibility() && e.innerText?.trim() === ${JSON.stringify(name)})`
const oben = "return Math.round([...document.querySelectorAll('main h2')].find((h) => h.checkVisibility() && h.innerText.trim() === 'Weiterarbeiten')?.getBoundingClientRect().top ?? -1)"
const ueberschrift = (name) => `return [...document.querySelectorAll('main h1')].some((h) => h.checkVisibility() && h.innerText.trim() === ${JSON.stringify(name)})`

const gewaehlteZeile = "return [...document.querySelectorAll('[role=grid] [role=row][aria-selected=true]')].some((z) => z.checkVisibility())"

const karten = (zahl) => `${zahl} ${zahl === 1 ? 'Karte' : 'Karten'} mit Befund`

async function lauf(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const [set] = JSON.parse(p.mse(ort, 'sets', '--json')).sets
  const erwartet = JSON.parse(gebuendeltesMse(ort, 'probleme', String(set.id), '--json'))
  p.pruefe('die Saat hat Karten mit Befund', erwartet.zahl > 0, `${erwartet.zahl} von ${erwartet.gesamt}`)

  // **Künstliche Exemplare ohne Preis** und drei Decks: zweien fehlt ein
  // Druck, der nicht in der Sammlung liegt, das dritte ist leer.
  p.mse(ort, 'exemplar', randomUUID(), '2')
  p.mse(ort, 'exemplar', randomUUID(), '3')
  for (const [nummer, name] of [['1', 'Waldschritt'], ['2', 'Moorwacht'], ['3', 'Glutpfad']]) {
    p.mse(ort, 'deck', 'neu', name, 'magic')
    if (name !== 'Glutpfad') p.mse(ort, 'deck', 'legen', nummer, randomUUID(), '4')
  }
  const portfolio = JSON.parse(gebuendeltesMse(ort, 'portfolio', '--json'))

  await p.starte({ ort })
  await p.groesse(1440, 838)
  await p.taste('1', 'befehl')

  p.notiz(`${bild}: Sammlung und Decks — ohne Netz, aus dem Bestand`)
  await p.warte('die Kachel der Sets steht', () => p.seite(`return !!${kachel(' Set')}`), 20000)
  const vorher = await p.seite(oben)
  await p.warte('die Kachel der Decks trägt ihre Unterzeile', () => p.seite(`return /mit fehlenden Karten/.test(${kachel(' Decks')}?.innerText ?? '')`), 20000)
  const sammlung = await p.seite(kacheltext(' Exemplar'))
  p.pruefe('die Kachel der Sammlung nennt die Exemplare ohne Preis wie `mse portfolio`', (sammlung ?? '').endsWith(`${portfolio.ohnePreis} ohne Preis`) && portfolio.zeilen.length === 0, `${sammlung} — Kommandozeile: ${portfolio.ohnePreis} ohne Preis, ${portfolio.zeilen.length} Währungen`)
  const decks = await p.seite(kacheltext(' Decks'))
  p.pruefe('die Kachel der Decks nennt zwei mit fehlenden Karten', (decks ?? '').endsWith('2 mit fehlenden Karten'), decks)
  const danach = await p.seite(oben)
  p.pruefe('die Unterzeilen verschieben nichts', vorher >= 0 && Math.abs(vorher - danach) <= 1, `${vorher} → ${danach}`)
  await p.bild(`kacheln-${bild}`)

  p.notiz(`${bild}: die Kachel der Decks führt zu genau ihnen`)
  await p.klickeAuf(kachel(' Decks'), 'die Kachel der Decks')
  await p.warte('die Deckliste ist gefiltert', () => p.seite(textDa('/^2 Decks mit fehlenden Karten$/')), 15000)
  p.pruefe('die zwei Decks stehen da, das dritte nicht', (await p.seite(sichtbar('Waldschritt'))) && (await p.seite(sichtbar('Moorwacht'))) && !(await p.seite(sichtbar('Glutpfad'))))
  await p.bild(`decks-${bild}`)
  await p.klickeAuf(knopf('Alle Decks zeigen'), '»Alle Decks zeigen«')
  await p.warte('alle Decks stehen wieder da', () => p.seite(sichtbar('Glutpfad')), 10000)
  p.pruefe('der Filter ist gelöst', !(await p.seite(textDa('/^2 Decks mit fehlenden Karten$/'))))

  p.notiz(`${bild}: nach „Jetzt prüfen“ — die Karten mit Befund am Objekt`)
  await p.taste('1', 'befehl')
  await p.klickeAuf(knopf('Jetzt prüfen'), '»Jetzt prüfen«')
  await p.warte('die Prüfung ist fertig', () => p.seite(textDa('/^Probleme geprüft /')), 120000)
  await p.warte('die Aufmerksamkeit nennt das Set', () => p.seite(`return !!${zeile(new RegExp(`${karten(erwartet.zahl)}$`).toString())}`), 15000)
  await p.bild(`aufmerksamkeit-${bild}`)

  await p.klickeAuf(kachel(' Karte'), 'die Kachel der Karten')
  await p.warte('die Kartenliste ist gefiltert', () => p.seite(textDa(`/^${karten(erwartet.zahl)}$/`)), 15000)
  const gefiltert = await p.seite(zeilen)
  p.pruefe('die Liste zeigt so viele Karten, wie `mse probleme` zählt', gefiltert === erwartet.zahl, `${gefiltert} — Kommandozeile: ${erwartet.zahl}`)
  // **Aufgeschlagen ist eine Karte mit Befund** — die gewählte Zeile steht
  // in der gefilterten Liste. Der erste Lauf fand hier die vorige Karte mit
  // „Keine Probleme an dieser Karte“.
  await p.warte('die gewählte Zeile steht in der gefilterten Liste', () => p.seite(gewaehlteZeile), 10000)
  await p.bild(`karten-${bild}`)
  await p.klickeAuf(knopf('Alle Karten zeigen'), '»Alle Karten zeigen«')
  await p.warte('alle Karten stehen wieder da', async () => (await p.seite(zeilen)) === erwartet.gesamt, 10000)
  p.pruefe('der Filter ist gelöst', !(await p.seite(textDa(`/^${karten(erwartet.zahl)}$/`))))

  p.notiz(`${bild}: die Zeile der Aufmerksamkeit, mit der Tastatur`)
  await p.taste('1', 'befehl')
  await p.warte('die Startseite steht', () => p.seite(ueberschrift('MagicSetEditor4Mac')), 10000)
  await p.seite(`${zeile(new RegExp(`${karten(erwartet.zahl)}$`).toString())}?.focus(); return true`)
  await p.taste('Enter')
  await p.warte('die Kartenliste ist wieder gefiltert', () => p.seite(textDa(`/^${karten(erwartet.zahl)}$/`)), 15000)
  p.pruefe('und zeigt dieselben Karten', (await p.seite(zeilen)) === erwartet.zahl)
  p.pruefe('⌘Q beendet', await p.beende())
}

export default async function ap8befunde09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
