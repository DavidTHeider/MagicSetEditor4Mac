// **Ein Fehler beim Zeichnen, am nativen Fenster** — F01
// (#2340),
// Befund A-01.
//
// Der Weg aus dem Issue: **Menü → Fehler → Diagnosepaket → Wiederaufnahme
// mit unverändertem Bestand**. Der Fehler kommt über die Naht der
// Hüllenprobe (`Fehlergrenze.tsx`, `__mseProbewurf` und `mse.probewurf`):
// ein Bereich wirft beim nächsten Zeichnen, ohne dass ein echter Fehler im
// Baum liegt.
//
// # Was es prüft
//
// - **Bereich**: Satz, Kennung, Fokus auf dem Rückweg, die Seitenleiste
//   bleibt bedienbar; hell und dunkel — **und in jedem Arbeitsbereich**,
//   über die Palette, die Liste aus `navigation.ts`.
// - **Diagnose**: an Ort und Stelle; das gesicherte Paket nennt die Kennung
//   in `protokoll.txt`.
// - **Wiederaufnahme** mit der Tastatur (↩ auf dem Rückweg); der Bestand ist
//   danach unverändert.
// - **Wurzel**: ein Wurf beim Laden zeigt einen Satz statt eines leeren
//   Fensters, und »Fenster neu laden« bringt die Anwendung zurück.
//
// # Was es nicht belegt
//
// VoiceOver — ob der Satz angesagt wird, hört ein Mensch. Und einen echten
// Fehler: die Naht wirft auf Bestellung.
import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { paletteOffen } from './budgets.mjs'
import { seitenliste } from './zustaende.mjs'

const flaeche = "[...document.querySelectorAll('[role=alert]')].find((e) => e.checkVisibility() && e.querySelector('h1, h2'))"
const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(text)})`
const bindung = (p, ausdruck) => p.seite(`const b = window.go.app.Bindungen; return await ${ausdruck}`)

export default async function fehlergrenze09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const bestand = () => p.mse(ort, 'sets') + p.mse(ort, 'karten', '1')
  const vorher = bestand()

  await p.starte({ ort })
  await p.groesse(1440, 900)

  p.notiz('Menü → Fehler: die Einstellungen über die Menüleiste, und ihr Bereich wirft')
  await p.seite("globalThis.__mseProbewurf = 'einstellungen'; return true")
  await p.klick('MagicSetEditor4Mac', 'Einstellungen …')
  const satz = await p.warte('die Fehlerfläche steht', () => p.seite(`return (${flaeche})?.innerText ?? false`), 15000)
  const kennung = /([0-9a-f]{4}-[0-9a-f]{4})/.exec(satz)?.[1] ?? ''
  p.pruefe('ein Satz mit Kennung steht da', kennung !== '', satz.replace(/\s+/g, ' '))
  p.pruefe('der Fokus liegt auf »Bereich neu laden«', await p.seite(`return document.activeElement === ${knopf('Bereich neu laden')}`))
  p.pruefe('die Seitenleiste bleibt bedienbar', await p.seite("return !!document.querySelector('nav[aria-label=\"Hauptnavigation\"]')?.checkVisibility()"))
  await p.bild('f01-bereich-hell')

  p.notiz('→ Diagnosepaket, an Ort und Stelle')
  await p.klickeAuf(knopf('Diagnose öffnen'), '»Diagnose öffnen«')
  await p.wartetext('Zusammenstellen', 15000)
  await p.klickeText('Zusammenstellen')
  const sichern = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith('Paket sichern'))"
  await p.warte('die Teile stehen', () => p.seite(`return !!${sichern}`), 20000)
  const paket = join(ort.ort, 'f01.msediagnose')
  await p.dialogantwort(paket)
  await p.klickeAuf(sichern, '»Paket sichern«')
  await p.warte('das Paket liegt da', () => Promise.resolve(existsSync(paket)), 20000)
  const protokoll = execFileSync('unzip', ['-p', paket, 'protokoll.txt']).toString()
  p.pruefe('das Protokoll im Paket nennt die Kennung', kennung !== '' && protokoll.includes(kennung), protokoll.split('\n').filter((z) => z.includes(kennung)).join(' ').slice(0, 300))
  await p.bild('f01-diagnose')

  p.notiz('→ Wiederaufnahme mit der Tastatur')
  await p.seite(`(${knopf('Bereich neu laden')}).focus(); return true`)
  await p.taste('Enter')
  await p.warte('die Einstellungen zeichnen', () => p.seite(`return document.querySelector('main')?.dataset.seite === 'einstellungen' && !(${flaeche})`), 15000)
  p.pruefe('der Bestand ist unverändert', bestand() === vorher)
  await p.bild('f01-wieder')

  p.notiz('In jedem Arbeitsbereich: über die Palette hin, Wurf, Rückweg')
  const { anwendung, set } = seitenliste(readFileSync(resolve('frontend/src/navigation.ts'), 'utf8'))
  const ohne = []
  for (const seite of [...anwendung, ...set]) {
    const text = await wirfIn(p, seite)
    if (text === null) {
      ohne.push(seite)
      continue
    }
    const fokus = await p.seite(`return document.activeElement === ${knopf('Bereich neu laden')}`)
    const diagnose = await p.seite(`return !!${knopf('Diagnose öffnen')}`)
    p.pruefe(`${seite}: Satz mit Kennung, Fokus auf dem Rückweg, Diagnose erreichbar`, /[0-9a-f]{4}-[0-9a-f]{4}/.test(text) && fokus && diagnose)
    await p.klickeAuf(knopf('Bereich neu laden'), '»Bereich neu laden«')
    await p.warte(`${seite}: der Bereich zeichnet wieder`, () => p.seite(`return !(${flaeche})`), 15000)
  }
  p.zeige(`ohne Zeile in der Palette (nicht erreicht): ${ohne.join(', ') || '—'}`)
  p.pruefe('der Bestand ist nach dem Durchgang unverändert', bestand() === vorher)

  p.notiz('Die Wurzel: ein Wurf beim Laden')
  // **Neu geladen wird nach der Antwort** — ein `location.reload()` im
  // selben Ruf nimmt dessen Antwort mit, und die Probe wartete 30 s.
  await p.seite("sessionStorage.setItem('mse.probewurf', 'fenster'); setTimeout(() => location.reload(), 30); return true")
  const wurzel = await p.warte('die Fläche der Wurzel steht', () => p.seite(`return (${flaeche})?.innerText ?? false`), 20000)
  p.pruefe('statt eines leeren Fensters ein Satz mit Kennung', /[0-9a-f]{4}-[0-9a-f]{4}/.test(wurzel), wurzel.replace(/\s+/g, ' '))
  p.pruefe('der Fokus liegt auf »Fenster neu laden«', await p.seite(`return document.activeElement === ${knopf('Fenster neu laden')}`))
  await p.bild('f01-wurzel')
  await p.taste('Enter')
  await zurueck(p)
  p.pruefe('der Bestand ist auch danach unverändert', bestand() === vorher)

  // **Dunkel über die Einstellung, nicht über das Attribut**: das Fenster
  // übernimmt das Erscheinungsbild aus seinem Einstellungsstand, also wird
  // gesetzt, neu geladen und der Fehler noch einmal ausgelöst.
  p.notiz('Dunkel: dieselbe Fläche')
  await bindung(p, "b.EinstellungSetzen('erscheinungsbild', 'dunkel')")
  await p.seite('setTimeout(() => location.reload(), 30); return true')
  await zurueck(p)
  p.pruefe('das Fenster ist dunkel', (await p.seite('return document.documentElement.dataset.erscheinungsbild')) === 'dunkel')
  p.pruefe('dunkel: Satz mit Kennung', /[0-9a-f]{4}-[0-9a-f]{4}/.test((await wirfIn(p, 'einstellungen')) ?? ''))
  await p.bild('f01-bereich-dunkel')
  // **↩ statt Klick**: der Fokus steht auf dem Rückweg, und der native Klick
  // traf nach dem Neuladen am 2026-10-14 ein `DIV` über dem Knopf.
  p.pruefe('dunkel: der Fokus liegt auf »Bereich neu laden«', await p.seite(`return document.activeElement === ${knopf('Bereich neu laden')}`))
  await p.taste('Enter')
  await p.warte('dunkel: der Bereich zeichnet wieder', () => p.seite(`return !(${flaeche})`), 15000)
  await bindung(p, "b.EinstellungSetzen('erscheinungsbild', 'hell')")
  p.pruefe('⌘Q beendet', await p.beende())
}

// zurueck wartet, bis die Anwendung nach einem Neuladen wieder steht, und
// schließt »Erste Schritte«, das ein frischer Start zeigt.
//
// **Erst warten, dann fragen**: ein Ruf mitten ins Neuladen bekommt keine
// Antwort und hält die Probe bis zu seiner Frist auf — so am 2026-10-14,
// dieselbe Pause wie in `g07palette.mjs`.
async function zurueck(p) {
  await new Promise((weiter) => { setTimeout(weiter, 1500) })
  await p.warte('die Anwendung ist zurück', () => p.seite("return document.readyState === 'complete' && document.querySelectorAll('nav button').length > 3"), 30000)
  if (await p.seite(`return !!${knopf("Los geht's")}`)) {
    await p.klickeAuf(knopf("Los geht's"), '»Los geht\'s«')
  }
}

// wirfIn lässt den Bereich `seite` beim nächsten Zeichnen werfen, geht über
// die Palette hin und gibt den Text der Fehlerfläche — `null`, wenn die
// Palette die Seite nicht anbietet.
async function wirfIn(p, seite) {
  await p.seite(`globalThis.__mseProbewurf = ${JSON.stringify(seite)}; return true`)
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite(paletteOffen), 15000)
  const zeile = `document.querySelector('[data-palettenzeile][data-seite=${JSON.stringify(seite)}]')`
  if (!(await p.seite(`return !!(${zeile})`))) {
    await p.taste('Escape')
    await p.seite('delete globalThis.__mseProbewurf; return true')

    return null
  }
  await p.klickeAuf(zeile, `»${seite}« in der Palette`)

  return p.warte(`${seite}: die Fehlerfläche steht`, () => p.seite(`return document.querySelector('main')?.dataset.seite === ${JSON.stringify(seite)} && ((${flaeche})?.innerText ?? false)`), 15000)
}
