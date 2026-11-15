// **Die Probleme auf den Kacheln, am nativen Fenster** — AP-8, Teil 2a
// (#2365),
// Befund B-13, Entwurf `Start09`, mit dem Vertrag aus PERF02 (#2352).
//
// # Was es prüft
//
// Vor der Prüfung sagen die Kacheln, wie viel geprüft ist, und **nie 0**;
// darunter steht der Anstoß. Nach „Jetzt prüfen“ tragen sie die Zahlen — und
// **dieselben wie die Kommandozeile** (`mse probleme`) für dasselbe Set. Die
// Zeile nennt das Alter und die Karten. Die Abschnitte darunter bleiben, wo
// sie waren (AP-4). Mit Problemen führt die Kachel zur gefilterten
// Setliste, sonst zur ganzen. Hell und dunkel.
//
// # Was es nicht belegt
//
// **Den Abbruch über die Hintergrundaufgaben** — die Einheitstests prüfen
// ihn in Go. **VoiceOver** hört ein Mensch.
import { execFileSync } from 'node:child_process'
import { join } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'

// gebuendeltesMse ruft das `mse` **aus dem Bündel** — das, das Anwendende
// bekommen. `p.mse` nimmt `.go/bin/mse`, und das kennt die mitgelieferte Saat
// nicht: es meldete für jede Karte „liegt in keinem Paketverzeichnis“, wo das
// Fenster und das gebündelte `mse` einen Befund finden (erster Lauf).
const gebuendeltesMse = (ort, ...argumente) =>
  execFileSync(join(process.cwd(), 'build/bin/MagicSetEditor4Mac.app/Contents/MacOS/mse'), argumente, { env: kindumgebung(ort) }).toString()

const liste = "[...document.querySelectorAll('ul[aria-label]')].find((u) => u.checkVisibility() && u.getAttribute('aria-label') === 'Dein Überblick')"
const kachel = (wort) => `[...(${liste})?.querySelectorAll('button') ?? []].find((b) => b.innerText.replace(/\\s+/g, ' ').includes(${JSON.stringify(wort)}))`
const kacheltext = (wort) => `return ${kachel(wort)}?.innerText.replace(/\\s+/g, ' ').trim() ?? null`
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim() === ${JSON.stringify(name)})`
const textDa = (muster) => `return [...document.querySelectorAll('main p')].some((e) => e.checkVisibility() && ${muster}.test(e.innerText.trim()))`
const oben = "return Math.round([...document.querySelectorAll('main h2')].find((h) => h.checkVisibility() && h.innerText.trim() === 'Weiterarbeiten')?.getBoundingClientRect().top ?? -1)"

async function lauf(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const [set] = JSON.parse(p.mse(ort, 'sets', '--json')).sets
  // **Dieselbe Prüfung über die Kommandozeile** — ihre Zahl ist der Maßstab
  // für die des Fensters.
  const erwartet = JSON.parse(gebuendeltesMse(ort, 'probleme', String(set.id), '--json'))

  await p.starte({ ort })
  await p.groesse(1440, 838)
  await p.taste('1', 'befehl')

  p.notiz(`${bild}: vor der Prüfung — nie 0`)
  await p.warte('die Kachel der Sets steht', () => p.seite(`return !!${kachel('Set')}`), 20000)
  const setsVorher = await p.seite(kacheltext(' Set'))
  const kartenVorher = await p.seite(kacheltext(' Karte'))
  p.pruefe('die Kachel der Sets sagt, wie viel geprüft ist', /0 von 1 geprüft/.test(setsVorher ?? ''), setsVorher)
  p.pruefe('die Kachel der Karten sagt „nicht vollständig geprüft“', /nicht vollständig geprüft/.test(kartenVorher ?? '') && !/mit Befund/.test(kartenVorher ?? ''), kartenVorher)
  p.pruefe('darunter steht der Anstoß', await p.seite(textDa('/^Probleme: 1 von 1 Sets nicht geprüft$/')))
  const vorher = await p.seite(oben)
  await p.bild(`vorher-${bild}`)

  p.notiz(`${bild}: nach „Jetzt prüfen“ — dieselben Zahlen wie die Kommandozeile`)
  await p.klickeAuf(knopf('Jetzt prüfen'), '»Jetzt prüfen«')
  await p.warte('die Prüfung ist fertig', () => p.seite(textDa('/^Probleme geprüft /')), 120000)
  const zeile = await p.seite("return [...document.querySelectorAll('main p')].map((e) => e.innerText.trim()).find((t) => t.startsWith('Probleme geprüft')) ?? ''")
  p.pruefe('die Zeile nennt Alter und Karten', zeile.endsWith(`· ${erwartet.geprueft} von ${erwartet.gesamt} Karten`), `${zeile} — erwartet ${erwartet.geprueft} von ${erwartet.gesamt}`)
  const setsNachher = await p.seite(kacheltext(' Set'))
  const kartenNachher = await p.seite(kacheltext(' Karte'))
  p.pruefe('die Kachel der Sets zählt wie die Kommandozeile', (setsNachher ?? '').endsWith(`${erwartet.zahl > 0 ? 1 : 0} mit Problemen`), `${setsNachher} — Kommandozeile: ${erwartet.zahl}`)
  p.pruefe('die Kachel der Karten zählt wie die Kommandozeile', (kartenNachher ?? '').endsWith(`${erwartet.zahl} mit Befund`), `${kartenNachher} — Kommandozeile: ${erwartet.zahl}`)
  const nachher = await p.seite(oben)
  p.pruefe('die Abschnitte darunter bleiben, wo sie waren', vorher >= 0 && Math.abs(vorher - nachher) <= 1, `${vorher} → ${nachher}`)
  await p.bild(`nachher-${bild}`)

  p.notiz(`${bild}: die Kachel führt zur Setliste`)
  await p.klickeAuf(kachel(' Set'), 'die Kachel der Sets')
  await p.warte('die Setliste steht', () => p.seite("return [...document.querySelectorAll('main h1')].some((h) => h.checkVisibility() && h.innerText.trim() === 'Sets')"), 15000)
  const gefiltert = await p.seite(textDa('/^1 Set mit Problemen$/'))
  p.pruefe(erwartet.zahl > 0 ? 'die Setliste ist auf das Set mit Problemen gefiltert' : 'ohne Probleme steht die ganze Setliste da', gefiltert === erwartet.zahl > 0, String(gefiltert))
  await p.bild(`setliste-${bild}`)
  p.pruefe('⌘Q beendet', await p.beende())
}

export default async function ap8kacheln09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
