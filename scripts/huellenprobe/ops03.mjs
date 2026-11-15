// **Alles zurücknehmen, was die Einrichtung anlegte, am nativen Fenster** —
// OPS03 (#2349),
// Befund A-10 des Audits 0.9.0. Zielartboard `Betrieb09`.
//
// # Was hier geschieht — in einem Wegwerfzuhause
//
// 1. **Vorher**: *Einstellungen › Verbindungen › Werkzeuge ohne Fenster* nennt
//    sechs Ziele, darunter die Verweise in `~/.local/bin` — offen. In jeder
//    Datei, in die gleich etwas kommt, steht schon ein fremder Eintrag.
// 2. **Einzeln eingerichtet** — die Verweise **mit der Tastatur**, Shellprofil,
//    Codex und Claude Desktop mit der Maus.
// 3. **Alles zurücknehmen**: je Ziel „zurückgenommen“, der Block der
//    Kommandozeile darüber liest nach, und jede Nachbardatei ist bytegleich
//    wie vor dem Einrichten.
// 4. **Ein zweites Mal, an der Kommandozeile** — `mse schnittstellen
//    widerrufen`: jedes Ziel unverändert, kein Fehler.
// 5. **Dunkel**: der Abschnitt im dunklen Bild.
//
// # Was es nicht belegt
//
// **Claude Code und den Dienst.** Das eine riefe das echte `claude`, das
// andere meldete einen Dienst beim launchd dieses Rechners an (Vermerk 0159)
// — die Abnahmeleitung sperrt `launchctl` ohnehin (`MSE_DIENST_GESPERRT`).
// Beide belegen die Läufe in `internal/app` mit eingesetztem Befehlsweg.
// **Ein Programm, das schon fehlt**, auch: ein laufendes Bündel lässt sich hier
// nicht wegnehmen (`TestAllesWiderrufenNimmtNurEigenesAuchOhneProgramm`).
import { chmodSync, copyFileSync, existsSync, lstatSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

const wurzel = resolve(import.meta.dirname, '..', '..')
const buendel = join(wurzel, 'build/bin/MagicSetEditor4Mac.app/Contents/MacOS')

const namen = {
  befehle: 'Die Verweise auf mse und mse-mcp in ~/.local/bin',
  befehlszeile: 'Die Kommandozeile im PATH der Anmeldeshell',
  codex: 'Der Werkzeugserver für Codex',
  claudedesktop: 'Der Werkzeugserver für Claude Desktop',
}

const zeile = (name) => `[...document.querySelectorAll('li[aria-labelledby]')].find((l) => l.checkVisibility() && document.getElementById(l.getAttribute('aria-labelledby'))?.innerText.trim() === ${JSON.stringify(name)})`
const knopfIn = (name, text) => `[...(${zeile(name)})?.querySelectorAll('button') ?? []].find((b) => b.innerText.trim() === ${JSON.stringify(text)})`
// **Der Schlüssel des Stands und nicht sein Text** — „nicht eingerichtet“
// enthält „eingerichtet“.
const standVon = (p, name) => p.seite(`return (${zeile(name)})?.querySelector('[data-stand]')?.getAttribute('data-stand') ?? ''`)
const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(text)})`

// nachbarn legt neben jedes Ziel einen fremden Eintrag und gibt Pfad und
// Inhalt.
function nachbarn(heim) {
  const dateien = {
    [join(heim, '.zprofile')]: 'export EDITOR=vi\n',
    [join(heim, '.codex', 'config.toml')]: '[mcp_servers.anderer]\ncommand = "/opt/anderer"\n',
    [join(heim, 'Library', 'Application Support', 'Claude', 'claude_desktop_config.json')]: '{\n  "mcpServers": {\n    "anderer": {"command": "/opt/anderer"}\n  }\n}\n',
    [join(heim, '.local', 'bin', 'mein-werkzeug')]: '#!/bin/sh\n',
  }
  for (const [pfad, inhalt] of Object.entries(dateien)) {
    mkdirSync(join(pfad, '..'), { recursive: true })
    writeFileSync(pfad, inhalt)
  }

  return dateien
}

async function zuDenWerkzeugen(p) {
  await p.taste(',', 'befehl')
  await p.wartetext('Netz, Gegenüber, CLI, MCP, REST', 20000)
  // **Seit AP-10 (#2367) unter Verbindungen** und nicht mehr unter Diagnose.
  await p.klickeText('Verbindungen')
  await p.klickeText('Werkzeuge ohne Fenster')
  await p.warte('sechs Ziele stehen da', () => p.seite("return [...document.querySelectorAll('li[aria-labelledby]')].filter((l) => l.checkVisibility()).length >= 6"), 20000)
}

export default async function ops03(p) {
  const ort = p.ort({ zuhause: true })
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')

  // **`mse` im Bündel**, wie `scripts/werkzeuge.sh` es für die Freigabe legt —
  // sonst zeigten die Verweise ins Leere und hießen gleich »veraltet«.
  copyFileSync(join(wurzel, '.go/bin/mse'), join(buendel, 'mse'))
  chmodSync(join(buendel, 'mse'), 0o755)

  const vorher = nachbarn(ort.heim)

  await p.starte({ ort })

  p.notiz('1 — vorher: sechs Ziele, die Verweise offen, fremde Einträge daneben')
  await zuDenWerkzeugen(p)
  p.pruefe('die Verweise sind ein Ziel und offen', (await standVon(p, namen.befehle)) === 'offen', await standVon(p, namen.befehle))
  await p.seite(`(${zeile(namen.befehle)}).scrollIntoView({ block: 'center' }); return true`)
  await p.bild('vorher')

  p.notiz('2 — einzeln einrichten: die Verweise mit der Tastatur, drei mit der Maus')
  await p.seite(`(${knopfIn(namen.befehle, 'Einrichten')}).focus(); return true`)
  await p.taste('Enter')
  await p.warte('die Verweise sind eingerichtet', async () => (await standVon(p, namen.befehle)) === 'eingerichtet', 20000)
  for (const name of ['befehlszeile', 'codex', 'claudedesktop']) {
    await p.seite(`(${zeile(namen[name])}).scrollIntoView({ block: 'center' }); return true`)
    await p.klickeAuf(knopfIn(namen[name], 'Einrichten'), `»Einrichten« an ${name}`)
    await p.warte(`${name} ist eingerichtet`, async () => (await standVon(p, namen[name])) === 'eingerichtet', 20000)
  }
  const verweis = join(ort.heim, '.local', 'bin', 'mse')
  p.pruefe('der Verweis auf mse liegt da', existsSync(verweis) && lstatSync(verweis).isSymbolicLink())
  p.pruefe('der eigene Codex-Eintrag steht neben dem fremden', readFileSync(join(ort.heim, '.codex', 'config.toml'), 'utf8').includes('magicseteditor4mac'))
  await p.bild('eingerichtet')

  p.notiz('3 — Alles zurücknehmen')
  await p.seite(`(${knopf('Alles zurücknehmen')}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(knopf('Alles zurücknehmen'), '»Alles zurücknehmen«')
  await p.warte('jedes eingerichtete Ziel ist zurückgenommen', async () => {
    for (const name of Object.values(namen)) {
      if ((await standVon(p, name)) !== 'offen') return false
    }

    return true
  }, 30000)
  for (const [pfad, inhalt] of Object.entries(vorher)) {
    p.pruefe(`${pfad.replace(ort.heim, '~')} ist bytegleich wie vorher`, readFileSync(pfad, 'utf8') === inhalt, readFileSync(pfad, 'utf8'))
  }
  p.pruefe('die Verweise sind fort', !existsSync(verweis) && !existsSync(join(ort.heim, '.local', 'bin', 'mse-mcp')))
  const ausgaenge = await p.seite("return [...document.querySelectorAll('li, p')].filter((e) => e.checkVisibility() && /zurückgenommen/.test(e.innerText)).length")
  p.pruefe('die Ausgänge nennen das Zurücknehmen', ausgaenge > 0, String(ausgaenge))
  const kommandozeile = await p.seite("return [...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText.trim() === 'Die Kommandozeile einrichten')?.innerText ?? ''")
  p.pruefe('der Block der Kommandozeile hat nachgelesen — beide Verweise »noch nicht eingerichtet«', /mse noch nicht eingerichtet/.test(kommandozeile) && /mse-mcp noch nicht eingerichtet/.test(kommandozeile), kommandozeile.replace(/\n/g, ' · '))
  await p.bild('zurueckgenommen')

  p.notiz('4 — ein zweites Mal, an der Kommandozeile')
  const zweites = p.mse(ort, 'schnittstellen', 'widerrufen')
  p.zeige(zweites)
  p.pruefe('nichts mehr zu tun, und kein Fehler', zweites.split('\n').filter(Boolean).every((z) => z.endsWith('\tunveraendert')), zweites)
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('5 — dunkel')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'dunkel')
  await p.starte({ ort })
  await zuDenWerkzeugen(p)
  await p.seite(`(${zeile(namen.befehle)}).scrollIntoView({ block: 'center' }); return true`)
  p.pruefe('Mantine steht dunkel', (await p.seite('return document.documentElement.dataset.mantineColorScheme')) === 'dark')
  await p.bild('dunkel')
  p.pruefe('⌘Q beendet', await p.beende())
}
