#!/usr/bin/env node
// **Der Lastlauf über REST und MCP** — QA-08.AK-01 und AK-04
// (#1739).
//
//     source scripts/activate.sh
//     node scripts/lastlauf.mjs                        # alle Profile, mit und ohne Kontingente
//     node scripts/lastlauf.mjs rest-lesen fairness    # nur diese
//     DAUER=5 KLIENTEN=4 VARIANTEN=ohne node scripts/lastlauf.mjs
//     LASTLAUF_PRUEFEN=1 …                            # zählt je Gruppe die Status
//
// # Gegen wen
//
// **Gegen frisch gebaute Programme und einen eigenen Bestand** — nie gegen
// den echten. Der Lauf baut `mse`, `mse-rest` und `mse-mcp` in einen Ordner
// unter `TMPDIR`, legt dort ein **eigenes Zuhause** an (`HOME`), darin den
// Bestand (`MSE_BESTAND`) und die Saat als Pakete, und gibt jedem Prozess
// einen **Schlüsselbund, den es nicht gibt** (`MSE_SCHLUESSELBUND`): ohne ihn
// griffen die Programme auf den Anmeldebund dieses Rechners zu
// (`docs/TESTGRENZEN.md`). Gebaut wird mit der aktivierten Umgebung, gefahren
// mit der verlegten — `go run` unter einem fremden `HOME` legte dort eine
// Werkzeugkette an.
//
// # Was ein Fall ist
//
// **Ein Profil gegen einen frisch gestarteten Dienst** — `mse-rest` und
// `mse-mcp` über HTTP, beide auf `127.0.0.1`. Frisch, weil Guthaben und
// Verteilung im Arbeitsspeicher eines Prozesses leben: ein zweites Profil
// gegen denselben Prozess erbte das leere Guthaben des ersten. Jedes Profil
// läuft zweimal: **mit** Kontingenten (die Vorgabe) und **ohne**
// (`grenzen.kontingente aus`, AK-03). Ohne zeigt, was die Maschine kann; mit
// zeigt, was der Schutz daraus macht.
//
// # Was gemessen wird
//
// Je Gruppe Rufe, Durchsatz, gelungene Rufe je Sekunde, Anteil der
// Drosselung, Fehlerrate, p50/p95/p99 beim Klienten, `Retry-After`, die
// Verteilung aus `/metriken` des Dienstes und der höchste belegte
// Arbeitsspeicher (RSS) der beiden Dienste. Das Profil `ueberlast` misst
// dazu die Erholung, `rest-auftraege` den Abbruch.
//
// # Was er nicht misst
//
// - **Eine zweite Maschine.** Klient und Dienst teilen sich diesen Rechner;
//   ein schneller Klient nimmt dem Dienst Kerne weg (0007).
// - **Das Fenster.** Es hat kein Guthaben und keine Uhr am Eingang
//   (`internal/service/laufzeitwerk.go`).
// - **TLS.** Alles läuft auf der Schleife, wo TLS nicht verlangt ist; jenseits
//   davon käme der Handschlag dazu.
// - **Große Bestände.** Die Saat ist das Probeset aus `testdaten/`; ein Set
//   mit tausend Karten liest und schreibt anders.

import { execFileSync, spawn } from 'node:child_process'
import { cpSync, mkdirSync, mkdtempSync, rmSync, symlinkSync, writeFileSync } from 'node:fs'
import { cpus, loadavg, totalmem, tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { performance } from 'node:perf_hooks'
import { fileURLToPath } from 'node:url'

import { auswerten, markdown, serverseitig } from './lastlaufbericht.mjs'
import { Kinderregister, beende, mcpUeberStdio, starteNacheinander } from './lastlaufdienste.mjs'
import { mcpUeberHTTP, ruf, speicher } from './lastlaufnetz.mjs'
import { profile, reihenfolge, rohpfadprobe } from './lastlaufprofile.mjs'
import { hoeflich, pruefeWirkungen } from './lastlaufwirkung.mjs'

const wurzel = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const dauer = Number(process.env.DAUER ?? 8)
const klienten = Number(process.env.KLIENTEN ?? 8)
const varianten = (process.env.VARIANTEN ?? 'mit,ohne').split(',')
const bericht = process.env.BERICHT ?? 'docs/belege/lastlauf.md'
const gewaehlt = process.argv.slice(2)
const faelle = gewaehlt.length === 0 ? reihenfolge : reihenfolge.filter((name) => gewaehlt.includes(name))

if (faelle.length === 0) {
  console.error(`kein Profil gefunden — es gibt ${reihenfolge.join(', ')}`)
  process.exit(2)
}

// **Der eigene Ort** — und alles, was die Programme darunter suchen.
const arbeit = mkdtempSync(join(tmpdir(), 'lastlauf-'))

// **Kein verborgener Winkel** — QA-LAST01 (#2092). Der Dienst weist jeden Pfad
// mit einem Glied ab, das mit einem Punkt beginnt (`internal/domain/dienstpfad.go`),
// und der Lauf für 0.7.0 lag unter `~/.codex/…`: jeder Ruf mit einem Pfad kam
// als 400 zurück, und der Lauf endete grün. Das sagt er jetzt vorher.
if (arbeit.split('/').some((glied) => glied.startsWith('.'))) {
  console.error(`Der Arbeitsort ${arbeit} liegt in einem verborgenen Ordner — der Dienst weist jeden Pfad dort ab. TMPDIR auf einen sichtbaren Ort setzen.`)
  rmSync(arbeit, { recursive: true, force: true })
  process.exit(2)
}
// **Jeder Prozess steht im Register** (`lastlaufdienste.mjs`) und wird beendet,
// wo der Lauf endet: in `finally`, bei einem Signal und beim Verlassen. Ein
// Bestand mit Hunderten eingelesener Sets bliebe sonst in `TMPDIR` liegen —
// und die Dienste liefen weiter.
const kinder = new Kinderregister(spawn)
process.on('exit', () => {
  kinder.beendeSofort()
  rmSync(arbeit, { recursive: true, force: true })
})
for (const signal of ['SIGINT', 'SIGTERM', 'SIGHUP']) {
  process.once(signal, () => { kinder.beendeAlle().finally(() => process.exit(130)) })
}
const zuhause = join(arbeit, 'zuhause')
const anwendungsort = join(zuhause, 'Library', 'Application Support', 'MagicSetEditor4Mac')
const pakete = join(anwendungsort, 'pakete')
const programme = { mse: join(arbeit, 'bin', 'mse'), rest: join(arbeit, 'bin', 'mse-rest'), mcp: join(arbeit, 'bin', 'mse-mcp') }

for (const [name, ziel] of Object.entries({ mse: programme.mse, 'mse-rest': programme.rest, 'mse-mcp': programme.mcp })) {
  execFileSync('go', ['build', '-o', ziel, `./cmd/${name}`], { cwd: wurzel, stdio: 'inherit' })
}

mkdirSync(join(arbeit, 'bilder'), { recursive: true })
mkdirSync(join(arbeit, 'boegen'), { recursive: true })
// **Die Eingabe der Einfuhr liegt im eigenen Ort** und nicht im Quellbaum —
// der kann selbst in einem verborgenen Ordner liegen.
const eingang = join(arbeit, 'eingang', 'probe.json')
mkdirSync(dirname(eingang), { recursive: true })
cpSync(join(wurzel, 'testdaten', 'mtgjson', 'probe.json'), eingang)
cpSync(join(wurzel, 'saat'), pakete, { recursive: true })
// **Der Stil des Probesets heißt anders als der der Saat** — derselbe
// Verweis wie in `internal/flaechen/wortlisten_test.go`.
symlinkSync(join(pakete, 'probe-schlicht.mse-style'), join(pakete, 'probe-stil.mse-style'))

const umgebung = {
  ...process.env,
  HOME: zuhause,
  MSE_BESTAND: join(anwendungsort, 'bestand.db'),
  MSE_SCHLUESSELBUND: join(arbeit, 'keiner.keychain'),
}

function mse(argumente, eingabe) {
  return execFileSync(programme.mse, argumente, { env: umgebung, encoding: 'utf8', input: eingabe })
}

// **Die Saat**: das Probeset, seine Karten und zwei Token unter zwei Namen —
// zwei Identitäten für die Frage nach der Fairness.
const setID = String(JSON.parse(mse(['oeffnen', join(wurzel, 'testdaten', 'probe.mse-set'), '--json'])).setID)
const karten = JSON.parse(mse(['karten', setID, '--json'])).karten.map((karte) => karte.id)
const profilID = JSON.parse(mse(['profil-neu', '1', 'Lastlauf', '--json'])).profil.id
const token = Object.fromEntries(['a', 'b'].map((name) => [name, mse(['token-anlegen', profilID],
  JSON.stringify({ principal: `Rechner ${name.toUpperCase()}`, bezeichnung: `Lastlauf ${name}`, scope: 'lesen' })).trim()]))

// metriken fragt die Verteilung ab — **höflich**: eine Drosselung wird
// abgewartet, denn die Auskunft selbst ist keine Last.
async function metriken(abfrage) {
  const antwort = await hoeflich(abfrage)

  return antwort.status === 200 && !antwort.gedrosselt ? antwort : { text: '' }
}

// fahre startet die Dienste, fährt ein Profil und gibt den Fall zurück.
//
// **Die Dienste werden in jedem Fall beendet** (`finally`) — auch wenn der
// zweite nicht anläuft, ein Profil abbricht oder ein Ruf über stdio nie
// beantwortet wird. Das Register beendet alles, was dieses Profil gestartet hat,
// die Prozesse über stdio eingeschlossen.
async function fahre(name, variante, zusatz = '', ablauf = profile[name]) {
  try {
    const [rest, mcp] = await starteNacheinander(kinder, [
      [programme.rest, ['-port', '0'], umgebung, 16],
      [programme.mcp, ['-adresse', '127.0.0.1', '-port', '0'], umgebung, 64],
    ])
    const messen = speicher([rest.kind.pid, mcp.kind.pid])
    const k = { rest, mcp, stdio: () => mcpUeberStdio(kinder, programme.mcp, [], umgebung), setID, karten, token, arbeit, dauer, klienten, mtgjson: eingang, fall: `${variante}-${name}${zusatz}` }

    const beginn = performance.now()
    const ergebnis = await ablauf(k)
    if (process.env.LASTLAUF_PRUEFEN) {
      for (const [g, p] of Object.entries(ergebnis.gruppen)) {
        const zahl = {}
        p.forEach((x) => { zahl[x.status] = (zahl[x.status] ?? 0) + 1 })
        console.error(g, JSON.stringify(zahl), p.slice(0, 2).map((x) => `${x.status} ${(x.text ?? '').slice(0, 160)}`))
      }
    }
    const sekunden = ergebnis.sekunden ?? (performance.now() - beginn) / 1000
    const rss = messen()

    const restStand = await metriken(() => ruf(rest, 'GET', '/metriken'))
    const mcpStand = await metriken(async () => {
      const antwort = await mcpUeberHTTP(mcp, 'lesen.metriken', {})
      return { ...antwort, text: antwort.inhalt }
    })

    // **Hat der Ruf getan, was er zusagte?** (`lastlaufwirkung.mjs`) Erst nach
    // den Zahlen des Dienstes: die Prüfung ist selbst eine Handvoll Rufe, und
    // sie soll keine Zahl verschieben.
    const wirkung = await pruefeWirkungen(ergebnis.gruppen, { abfrage: (weg) => ruf(rest, 'GET', weg) })

    await Promise.all([beende(rest), beende(mcp)])

    const gruppen = Object.fromEntries(Object.entries(ergebnis.gruppen).map(([gruppe, proben]) =>
      [gruppe, auswerten(proben, ergebnis.dauern?.[gruppe] ?? sekunden)]))
    const dienst = {}

    for (const gruppe of Object.keys(gruppen)) {
      if (gruppe.startsWith('MCP-stdio')) {
        dienst[gruppe] = serverseitig(ergebnis.stdioMetriken?.inhalt ?? '', 'mcp')
      } else {
        dienst[gruppe] = gruppe.startsWith('MCP') ? serverseitig(mcpStand.text, 'mcp') : serverseitig(restStand.text, 'dienst')
      }
    }

    // **Fehlende Wirkung steht im Befund des Falls**, mit Zahl und Grund.
    const luecken = [
      ...Object.entries(wirkung).filter(([, w]) => w.fehlend > 0).map(([gruppe, w]) =>
        `Wirkung fehlt (${gruppe}): ${w.fehlend} von ${w.geprueft} geprüften Rufen — ${w.beispiele.join('; ')}`),
      ...Object.entries(wirkung).filter(([, w]) => w.ungeprueft).map(([gruppe, w]) =>
        `Wirkung ungeprüft (${gruppe}): ${w.ungeprueft}`),
    ]

    const fall = {
      name: `${name}${zusatz}`, variante, gruppen, dienst, wirkung, befund: [ergebnis.befund, ...luecken].filter(Boolean).join(' · ') || undefined,
      speicherMiB: Math.round(Math.max(...Object.values(rss).map((s) => s.hoechst))),
      speicher: rss,
    }

    console.error(`${variante}\t${fall.name}\t${Object.entries(gruppen).map(([g, w]) =>
      `${g}: ${w.gelungenJeSekunde}/s, ${Math.round(w.anteilGedrosselt * 100)} % gedrosselt, p95 ${w.msP95} ms`).join(' · ')}`)

    return fall
  } finally {
    await kinder.beendeAlle()
  }
}

// **Die Last der Maschine vorher und nachher** — Klient und Dienst teilen
// sich die Kerne mit allem anderen, was gerade läuft (§4: eine Pipeline
// braucht die Maschine allein, ein Lastlauf auch).
const lastVorher = loadavg()
const ergebnisse = []

// **Erst die Rohpfade, dann die Last** — QA-LAST01 (#2092): jeder Weg, der
// eine Wirkung zusagt, einmal und ohne Last. Wirkt einer nicht, misst die Last
// nur, wie schnell der Dienst ablehnt; dann endet der Lauf hier, mit Namen.
const roh = await fahre('rohpfade', 'vorprobe', '', rohpfadprobe)
const unbelegt = Object.entries(roh.wirkung).filter(([, w]) => w.ungeprueft || w.fehlend > 0 || w.geprueft === 0)
if (unbelegt.length > 0) {
  for (const [gruppe, w] of unbelegt) console.error(`Rohpfad ${gruppe}: ${w.ungeprueft ?? (w.beispiele.join('; ') || 'nicht geprüft')}`)
  process.exit(1)
}
console.error(`Rohpfade: ${Object.keys(roh.wirkung).join(', ')} — jeder einmal mit Wirkung`)

for (const variante of varianten) {
  mse(['einstellungen', 'setzen', 'grenzen.kontingente', variante === 'mit' ? 'an' : 'aus'])

  for (const name of faelle) {
    ergebnisse.push(await fahre(name, variante))
  }

  // **Einmal mit ausführlichem Protokoll und Unified Log** — der Nachschärfung
  // zum Issue nach war dort die Last zu finden und nicht im Behandler.
  if (variante === 'ohne' && faelle.includes('rest-lesen')) {
    mse(['einstellungen', 'setzen', 'protokoll.stufe', 'ausfuehrlich'])
    mse(['einstellungen', 'setzen', 'protokoll.system', 'an'])
    ergebnisse.push(await fahre('rest-lesen', variante, ' (Protokoll ausführlich)'))
    mse(['einstellungen', 'setzen', 'protokoll.stufe', 'info'])
  }
}

const commit = execFileSync('git', ['rev-parse', '--short=8', 'HEAD'], { cwd: wurzel, encoding: 'utf8' }).trim()
const veraendert = execFileSync('git', ['status', '--porcelain'], { cwd: wurzel, encoding: 'utf8' }).trim() !== ''
const system = execFileSync('sw_vers', ['-productVersion'], { encoding: 'utf8' }).trim()
const beleg = {
  commit, veraendert, zeit: new Date().toISOString().replace(/\.\d+Z$/, ' UTC').replace('T', ' '),
  maschine: `${cpus()[0].model}, ${cpus().length} Kerne, ${Math.round(totalmem() / 2 ** 30)} GiB, macOS ${system}, ` +
    `${execFileSync('go', ['version'], { encoding: 'utf8' }).trim().replace(/^go version /, '')}, Node ${process.version}`,
  eingaben: { dauer, klienten, varianten, profile: faelle, karten: karten.length },
  last: { vorher: lastVorher.map((wert) => Math.round(wert * 10) / 10), nachher: loadavg().map((wert) => Math.round(wert * 10) / 10) },
  rohpfade: Object.keys(roh.wirkung),
  varianten, faelle: ergebnisse,
}

mkdirSync(dirname(resolve(wurzel, bericht)), { recursive: true })
writeFileSync(resolve(wurzel, bericht), markdown(beleg))
writeFileSync(resolve(wurzel, bericht.replace(/\.md$/, '.json')), `${JSON.stringify(beleg, null, 2)}\n`)
console.error(`Beleg: ${bericht}`)

// **Ein Lauf, in dem Rufe ohne Wirkung blieben, endet nicht grün** — der Beleg
// liegt trotzdem da, mit dem Befund beim Fall.
const ohneWirkung = ergebnisse.flatMap((f) => Object.values(f.wirkung ?? {})).reduce((summe, w) => summe + w.fehlend, 0)
if (ohneWirkung > 0) {
  console.error(`${ohneWirkung} geprüfte Rufe ohne Wirkung — siehe die Befunde im Beleg`)
  process.exitCode = 1
}

// **Und eine Gruppe, die nichts durchbrachte, auch nicht** — QA-LAST01 (#2092).
// **Mit Kontingenten ist die Drosselung die Zusage**: eine Gruppe, die dort
// ausschließlich gedrosselt wurde, steht ungeprüft im Beleg, macht den Lauf
// aber nicht rot. Ohne Kontingente, oder mit einem anderen Status, schon.
const ungeprueft = ergebnisse.flatMap((f) => Object.entries(f.wirkung ?? {})
  .filter(([, w]) => w.ungeprueft && !(f.variante === 'mit' && w.nurGedrosselt))
  .map(([g]) => `${f.variante}-${f.name}/${g}`))
if (ungeprueft.length > 0) {
  console.error(`ungeprüft: ${ungeprueft.join(', ')} — siehe die Befunde im Beleg`)
  process.exitCode = 1
}
