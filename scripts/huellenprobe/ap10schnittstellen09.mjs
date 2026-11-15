// **Die Schnittstellen über Palette und Einstellungen finden, am nativen
// Fenster** — AP-10
// (#2367),
// Befund B-19, Abnahmekriterium 3: „Schnittstellen über Palette und
// Einstellungen auffindbar.“
//
// # Was es prüft
//
// 1. **Palette**: ⌘K, „MCP“ getippt — „Schnittstellen einrichten
//    (Kommandozeile, MCP, REST)“ steht unter *Tun*; ↩ darauf schlägt
//    *Einstellungen › Verbindungen* auf, der Abschnitt *Werkzeuge ohne
//    Fenster* steht offen, sein Schalter hat den Fokus, und die Einrichtung
//    der Kommandozeile ist zu sehen.
// 2. **Einstellungen**: *Verbindungen* nennt im Untertitel „CLI, MCP, REST“
//    und führt den Abschnitt als ersten; *Diagnose* führt ihn nicht mehr.
//
// Hell und dunkel. **Eingerichtet wird nichts** — die Probe drückt keinen
// Knopf der Einrichtung; was die Knöpfe tun, belegen `g08`, `ops03` und
// `ops09b`.
//
// # Was es nicht belegt
//
// **VoiceOver** hört ein Mensch.
const zeile = (anfang) => `[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim().startsWith(${JSON.stringify(anfang)}))`
const kastenknopf = "[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Werkzeuge ohne Fenster')"
const ueberschrift = (name) => `[...document.querySelectorAll('h2')].find((h) => h.checkVisibility() && h.innerText.trim().startsWith(${JSON.stringify(name)}))`
const gruppeneintrag = (name) => `[...document.querySelectorAll('button, a')].find((b) => b.checkVisibility() && b.innerText.trim().startsWith(${JSON.stringify(name)}))`

async function lauf(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  await p.starte({ ort })
  await p.groesse(1440, 838)

  p.notiz(`${bild}: ⌘K, „MCP“ — die Schnittstellen stehen unter Tun`)
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"), 10000)
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld('MCP')
  await p.warte('»Schnittstellen einrichten« steht in der Palette', () => p.seite(`return !!${zeile('Schnittstellen einrichten')}`), 10000)
  await p.bild(`palette-${bild}`)

  p.notiz(`${bild}: die Zeile schlägt den Abschnitt unter Verbindungen auf`)
  await p.klickeAuf(zeile('Schnittstellen einrichten'), '»Schnittstellen einrichten«')
  await p.warte('Verbindungen steht offen', () => p.seite(`return !!${ueberschrift('Verbindungen')}`), 15000)
  const kasten = await p.warte('der Abschnitt steht offen und hat den Fokus', () => p.seite(`const k = ${kastenknopf}; return k && k.getAttribute('aria-expanded') === 'true' && document.activeElement === k ? 'offen, mit Fokus' : null`), 15000)
  p.pruefe(`${bild}: »Werkzeuge ohne Fenster« ist aufgeschlagen, der Schalter hat den Fokus`, kasten === 'offen, mit Fokus', kasten)
  await p.wartetext('Die Kommandozeile einrichten', 20000)
  await p.bild(`werkzeuge-${bild}`)

  p.notiz(`${bild}: Untertitel und Ort in den Einstellungen`)
  const untertitel = await p.seite(`return ${gruppeneintrag('Verbindungen')}?.innerText.replaceAll('\\n', ' · ') ?? ''`)
  p.pruefe(`${bild}: der Eintrag „Verbindungen“ nennt CLI, MCP, REST`, untertitel.includes('CLI, MCP, REST'), untertitel)
  const erster = await p.seite("return [...document.querySelectorAll('section > button[aria-expanded]')].filter((b) => b.checkVisibility()).map((b) => b.innerText.trim())[0] ?? ''")
  p.pruefe(`${bild}: der erste Abschnitt unter Verbindungen ist »Werkzeuge ohne Fenster«`, erster === 'Werkzeuge ohne Fenster', erster)
  await p.klickeAuf(gruppeneintrag('Diagnose'), 'die Gruppe »Diagnose«')
  await p.warte('Diagnose steht offen', () => p.seite(`return !!${ueberschrift('Diagnose')}`), 15000)
  await p.wartetext('Protokoll', 15000)
  p.pruefe(`${bild}: unter Diagnose steht der Abschnitt nicht mehr`, !(await p.seite(`return !!${kastenknopf}`)))
  p.pruefe(`${bild}: ⌘Q beendet`, await p.beende())
}

export default async function ap10schnittstellen09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
