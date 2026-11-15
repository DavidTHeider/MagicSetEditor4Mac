// **Claude Desktop einrichten und widerrufen, am nativen Fenster** — OPS-09b
// (#1961):
//
// 1. vorher: *Einstellungen › Verbindungen* (⌘,) führt fünf Ziele, je Art mit
//    ihrer Vorschau — ohne offenes Set, denn die Dateien gehören dem
//    Menschen und keinem Set. Claude Desktop ist offen, seine Datei hat einen
//    fremden Eintrag und eine eigene Einrückung,
// 2. *Einrichten* — der eigene Eintrag steht in der Datei, der fremde auch,
//    und die Kommandozeile sagt denselben Stand — **über den Link in
//    `~/.local/bin`**, wie OPS-03 ihn legt, auf ein `mse` im Bündel neben
//    `mse-mcp`, wie `scripts/werkzeuge.sh` es für die Freigabe tut. Ohne
//    Link sähe dieser Schritt nicht, was er beim ersten Lauf fand: dass
//    `mse` hinter dem Link einen anderen Ort für `mse-mcp` annahm,
// 3. *Widerrufen* — die Datei ist Byte für Byte die von vorher.
//
// **Nicht gedrückt** werden Claude Code und der Dienst: das eine riefe das
// echte `claude`, das andere meldete einen Dienst bei launchd an — beides
// braucht die Zustimmung des Menschen davor (Vermerk 0159, BLOCKED).
import { execFileSync } from 'node:child_process'
import { chmodSync, copyFileSync, mkdirSync, readFileSync, symlinkSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { kindumgebung } from '../huellenorte.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const buendel = join(wurzel, 'build/bin/MagicSetEditor4Mac.app/Contents/MacOS')

// mseUeberDenLink legt `mse` ins Bündel und einen Link darauf ins Zuhause —
// und fragt über den Link.
function mseUeberDenLink(ort, ...argumente) {
  const imBuendel = join(buendel, 'mse')
  copyFileSync(join(wurzel, '.go/bin/mse'), imBuendel)
  chmodSync(imBuendel, 0o755)
  const lokal = join(ort.heim, '.local', 'bin')
  mkdirSync(lokal, { recursive: true })
  symlinkSync(imBuendel, join(lokal, 'mse'))

  return execFileSync(join(lokal, 'mse'), argumente, {
    encoding: 'utf8',
    env: kindumgebung(ort),
  })
}

const desktop = 'Der Werkzeugserver für Claude Desktop'

// zeile gibt den Ausdruck für die Zeile eines Ziels — über seinen Namen.
const zeile = (name) => `[...document.querySelectorAll('li[aria-labelledby]')].find((l) => l.checkVisibility() && document.getElementById(l.getAttribute('aria-labelledby'))?.innerText.trim() === ${JSON.stringify(name)})`
const knopfIn = (name, text) => `[...(${zeile(name)})?.querySelectorAll('button') ?? []].find((b) => b.innerText.trim() === ${JSON.stringify(text)})`
// **Die Marke trägt `data-stand`** seit UXR-19 (#1927) — vorher war es der
// zweite `span` der Zeile, und der ist seitdem der Name.
const standVon = (p, name) => p.seite(`return (${zeile(name)})?.querySelector('[data-stand]')?.innerText ?? ''`)

export default async function ops09b(p) {
  const ort = p.ort({ zuhause: true })
  const ordner = join(ort.heim, 'Library', 'Application Support', 'Claude')
  const datei = join(ordner, 'claude_desktop_config.json')
  mkdirSync(ordner, { recursive: true })
  const vorher = '{\n    "mcpServers": {\n        "anderer": {"command": "/bin/echo"}\n    },\n    "globalShortcut": ""\n}\n'
  writeFileSync(datei, vorher)

  await p.starte({ ort })

  p.notiz('1 — vorher: Einstellungen › Verbindungen, fünf Ziele, je Art ihre Vorschau')
  await p.taste(',', 'befehl')
  await p.wartetext('Netz, Gegenüber, CLI, MCP, REST', 20000)
  // **Seit AP-10 (#2367) unter Verbindungen** und nicht mehr unter Diagnose.
  await p.klickeText('Verbindungen')
  // **Seit UXR-19 (#1927) in einem Abschnitt** — erst aufklappen.
  await p.klickeText('Werkzeuge ohne Fenster')
  await p.wartetext('Kommandozeile und Werkzeugserver einrichten', 20000)
  // **Auf die Einträge warten und nicht auf die Überschrift** — die Liste
  // kommt aus einem eigenen Ruf und steht einen Augenblick später da; im
  // Gesamtlauf vom 2026-10-06 las das Drehbuch sie, bevor sie stand.
  await p.warte('die fünf Ziele stehen da', () => p.seite("return [...document.querySelectorAll('li[aria-labelledby]')].filter((l) => l.checkVisibility()).length >= 5"), 20000)
  await p.seite(`(${zeile(desktop)})?.scrollIntoView({ block: 'center' }); return true`)
  const namen = await p.seite("return [...document.querySelectorAll('li[aria-labelledby]')].filter((l) => l.checkVisibility()).map((l) => document.getElementById(l.getAttribute('aria-labelledby'))?.innerText.trim() ?? '')")
  for (const name of ['Der Werkzeugserver für Codex', desktop, 'Der Werkzeugserver für Claude Code', 'Der Dienst für Skripte als LaunchAgent']) {
    p.pruefe(`»${name}« steht da`, namen.includes(name), namen.join(' | '))
  }
  const vorschauen = await p.seite('return [...document.querySelectorAll(\'li[aria-labelledby] summary\')].filter((s) => s.checkVisibility()).map((s) => s.innerText.trim())')
  for (const satz of ['Dieser Eintrag kommt in die Datei', 'Dieser Befehl läuft', 'Diese Datei entsteht, und dieser Befehl lädt sie']) {
    p.pruefe(`die Vorschau »${satz}« steht da`, vorschauen.includes(satz), vorschauen.join(' | '))
  }
  p.pruefe('Claude Desktop ist offen', (await standVon(p, desktop)) === 'nicht eingerichtet', await standVon(p, desktop))
  await p.bild('vorher')

  p.notiz('2 — Einrichten: der eigene Eintrag kommt, der fremde bleibt')
  await p.klickeAuf(knopfIn(desktop, 'Einrichten'), '»Einrichten« an Claude Desktop')
  await p.warte('Claude Desktop ist eingerichtet', async () => (await standVon(p, desktop)) === 'eingerichtet', 15000)
  const eingerichtet = readFileSync(datei, 'utf8')
  p.pruefe('die Datei nennt mse-mcp unter magicseteditor4mac', eingerichtet.includes('"magicseteditor4mac"') && eingerichtet.includes('mse-mcp'), eingerichtet)
  p.pruefe('der fremde Eintrag steht weiter da', eingerichtet.includes('"anderer"'))
  const zeilen = mseUeberDenLink(ort, 'schnittstellen')
  p.pruefe('mse schnittstellen über den Link sagt denselben Stand', /claudedesktop\s+eingerichtet/.test(zeilen), zeilen)
  await p.bild('eingerichtet')

  p.notiz('3 — Widerrufen: Byte für Byte wie vorher')
  await p.klickeAuf(knopfIn(desktop, 'Widerrufen'), '»Widerrufen« an Claude Desktop')
  await p.warte('Claude Desktop ist wieder offen', async () => (await standVon(p, desktop)) === 'nicht eingerichtet', 15000)
  p.pruefe('die Datei ist Byte für Byte die von vorher', readFileSync(datei, 'utf8') === vorher)
  await p.bild('widerrufen')

  p.pruefe('⌘Q beendet', await p.beende())
}
