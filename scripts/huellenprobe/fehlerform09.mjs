// **Eine Bindungspanik im Fenster: ein Satz mit Kennung, kein Paketpfad** —
// F03 (#2342),
// Befunde A-03 und A-24.
//
// Die Panik kommt über die Störung des Rufbuchs (`stoere`, `panik`): die
// Hülle beantwortet `Sets` **so, wie Wails eine Bindungspanik schickt** —
// 422, `kind: RuntimeError`, `<Paketpfad>.Bindungen.Sets: panic: Probepanik`
// (`internal/fenster/rufstoerung.go`). Gemessen wird, was das Fenster daraus
// macht.
//
// # Was es prüft
//
// Im Fenster steht der Satz für unerwartete Fehler mit einer Kennung, **kein
// Paketpfad und kein `panic`**; im Protokoll steht unter derselben Kennung
// der Rohtext (`art=ruf`, `bereich=sets`). Danach, ohne Störung, zeichnet
// die Seite wieder.
//
// # Was es nicht belegt
//
// Eine echte Panik in Go — dort fängt Wails, und `panikMelden` schreibt sie
// (`internal/fenster/riegel_test.go`); und einen `TypeError` aus dem eigenen
// Code (`frontend/src/unerwartet.test.ts`). Kommandozeile und MCP belegen die
// Läufe in `internal/flaechen` gegen die gebauten Programme.
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'

import { stoere } from '../huellenrufbuch.mjs'

const alarm = "return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).join(' / ')"

export default async function fehlerform09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  await p.starte({ ort })
  await p.groesse(1440, 900)

  p.notiz('Sets wirft eine Panik, und die Seite wird aufgeschlagen')
  await stoere(p, [{ bindung: 'Sets', panik: true }])
  await p.taste('2', 'befehl')
  const meldung = await p.warte('die Seite meldet den Fehler', async () => (await p.seite(alarm)) || false, 15000)
  const kennung = /([0-9a-f]{4}-[0-9a-f]{4})/.exec(meldung)?.[1] ?? ''
  p.zeige(meldung.slice(0, 300))
  p.pruefe('im Fenster steht ein Satz mit Kennung', kennung !== '')
  const fenstertext = await p.seite('return document.body.innerText')
  p.pruefe('kein Paketpfad und kein »panic« im Fenster', !/github\.com|panic|internal\/app/.test(fenstertext))
  await p.bild('f03-panik-sets')

  const protokoll = readFileSync(join(dirname(ort.bestand), 'mse.log'), 'utf8')
  const zeile = protokoll.split('\n').find((z) => z.includes(`fehlerkennung=${kennung}`)) ?? ''
  p.zeige(zeile.slice(0, 300))
  p.pruefe('im Protokoll steht unter der Kennung der Rohtext', kennung !== '' && zeile.includes('art=ruf') && zeile.includes('bereich=sets') && zeile.includes('panic: Probepanik'))

  p.notiz('Ohne Störung zeichnet die Seite wieder')
  await stoere(p, [])
  await p.taste('1', 'befehl')
  await p.taste('2', 'befehl')
  await p.warte('die Sets stehen', async () => !(await p.seite(alarm)), 15000)
  p.pruefe('⌘Q beendet', await p.beende())
}
