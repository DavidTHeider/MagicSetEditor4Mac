#!/usr/bin/env node
// **Die Fuzzkampagne** — QA-06.AK-03 und AK-04
// (#1737).
//
//     make fuzzkampagne                                  # alle Ziele aus RELEASE.md
//     node scripts/fuzzkampagne.mjs FuzzRahmen FuzzChat  # nur diese
//     ZEIT=2m PARALLEL=4 SPEICHER=2GiB node scripts/fuzzkampagne.mjs
//     ROHDATEN=a.json BERICHT=… node scripts/fuzzkampagne.mjs FuzzA …  # eine Hälfte
//     node scripts/fuzzkampagne.mjs --zusammen a.json b.json         # der Bericht
//
// # Warum getrennt von der Pipeline
//
// **`make ci` fährt jedes Ziel nur mit seinem Saatkorpus** — das ist die
// Regressionsprobe, und jeder frühere Fund unter `testdata/fuzz/` läuft
// darin mit. Das Fuzzen selbst braucht Zeit, und eine Pipeline, die es
// enthielte, wäre entweder langsam oder fuzzte nicht ernsthaft.
//
// # Das Budget
//
// Je Ziel **eine Zeit** (`ZEIT`, Vorgabe 1m), **eine Zahl Arbeiter**
// (`PARALLEL`, Vorgabe die Hälfte der Kerne — der Rechner bleibt bedienbar)
// und **eine Speichergrenze** (`SPEICHER`, Vorgabe 2GiB, als `GOMEMLIMIT`).
// Minimiert wird höchstens 5 s je Eingabe: mit der Vorgabe von 60 s fanden
// die langsamen Ziele in einer Minute nichts (HAERTUNG.md).
//
// # Der Bericht
//
// `docs/belege/fuzzkampagne.md`: je Ziel Laufzeit, Ausführungen, **neue
// Eingaben** — das ist Gos Maß für erreichte neue Kanten — und, wenn es einen
// gibt, den Fund mit der Datei, die ihn wiederholt.
//
// **Jeder Teil ist erforderlich** (Z5, `scripts/fuzzbericht.mjs`): der
// Fuzzlauf, die Übernahme des erarbeiteten Korpus und der Abdeckungslauf. Ein
// Ziel, bei dem einer gescheitert ist, steht als **unvollständig** da, mit dem
// Grund, und der Anteil ist dann ein Strich und keine Messung. Der Exitcode ist
// 0, **1 bei einem Fund** und **3, wenn ein Teil gescheitert ist** (der Fund
// wiegt schwerer).
//
// # Der Anteil (QA-06.AK-04, QA-06c #1976)
//
// **Je Ziel der Anteil der Anweisungen seines Pakets, den Saat und
// erarbeitetes Korpus betreten.** Gos Fuzzer führt Kantenzähler, gibt sie aber
// nicht heraus — `new interesting` ist eine Zahl ohne Nenner. Gemessen wird
// deshalb nach dem Fuzzen ein Lauf über das ganze Korpus mit
// `-coverprofile -coverpkg=<paket>`: das erarbeitete Korpus liegt in
// `$(go env GOCACHE)/fuzz/<importpfad>/<Ziel>` und wird dafür **auf Zeit** nach
// `testdata/fuzz/<Ziel>` kopiert und danach wieder entfernt. **Was das nicht
// ist**: Kantenabdeckung — eine betretene Anweisung sagt nicht, über welche
// Kante sie erreicht wurde. Und der Nenner ist das ganze Paket, auch was das
// Ziel gar nicht erreichen soll.

import { execFileSync, spawnSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { availableParallelism } from 'node:os'
import { readFileSync, writeFileSync } from 'node:fs'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { anteilMessen, berichtText, fuzzlaufAuswerten, kampagnenstatus, standPruefen, zieleAus, zielErgebnis, zieleWaehlen, zusammenfuehren } from './fuzzbericht.mjs'

const wurzel = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const zeit = process.env.ZEIT ?? '1m'
const parallel = process.env.PARALLEL ?? String(Math.max(1, Math.floor(availableParallelism() / 2)))
const speicher = process.env.SPEICHER ?? '2GiB'
// **Der Bericht an einem eigenen Ort**, etwa für den Nachtest nach einem Fund —
// der erste Bericht mit seinen Funden bleibt dann stehen.
const bericht = process.env.BERICHT ?? 'docs/belege/fuzzkampagne.md'

// **Die Ziele stehen in RELEASE.md** — gelesen von `zieleAus`.
const alle = zieleAus(readFileSync(join(wurzel, 'docs/RELEASE.md'), 'utf8'))

// **In Hälften** (C02, #2227): `ROHDATEN=pfad.json` legt neben dem Bericht die
// Ergebnisse ab, und `--zusammen a.json b.json` macht aus den Hälften den einen
// Bericht — vom selben Stand, mit demselben Budget, sonst bricht es ab.
if (process.argv[2] === '--zusammen') {
  const teile = process.argv.slice(3).map((pfad) => JSON.parse(readFileSync(pfad, 'utf8')))
  const ganz = zusammenfuehren(teile, alle)
  writeFileSync(join(wurzel, bericht), berichtText(ganz.ergebnisse, ganz.kopf))
  const status = kampagnenstatus(ganz.ergebnisse)
  console.log(`${String(ganz.ergebnisse.length)} Ziele, ${String(status.funde.length)} Funde, ${String(status.unvollstaendig.length)} unvollständig → ${bericht}`)
  process.exit(status.exitcode)
}

const { ziele, unbekannt } = zieleWaehlen(alle, process.argv.slice(2))

if (unbekannt.length > 0 || ziele.length === 0) {
  console.error(`${unbekannt.length > 0 ? `unbekannt: ${unbekannt.join(', ')}` : 'kein Ziel gefunden'} — die Namen stehen in docs/RELEASE.md`)
  process.exit(2)
}

// **Der Stand vor dem ersten Ziel und nach dem letzten** (Zweitblick über
// Phase C, Befund 4) — mit den ungesicherten Änderungen, aber ohne das
// erarbeitete Korpus, das der Abdeckungslauf auf Zeit in `testdata/fuzz`
// legt.
//
// **Und der Inhalt** (Nachlauf über Phase C): Commit und Zahl der Dateien
// sahen eine zweite Änderung an einer schon geänderten Datei nicht. `inhalt`
// ist ein Abdruck des Unterschieds zum Commit samt neuer Dateien, `saat` der
// des Korpus unter `testdata/fuzz` — er steht im Bericht und wird nicht
// verglichen, denn ein Fund legt seine Eingabe dort ab.
const git = (argumente) => execFileSync('git', argumente, { cwd: wurzel, encoding: 'utf8', maxBuffer: 256 << 20 })
const dateien = (liste) => liste.split('\0').filter(Boolean).map((pfad) => `${pfad}\0${readFileSync(join(wurzel, pfad), 'latin1')}`)
const abdruck = (teile) => createHash('sha256').update(teile.join('\0')).digest('hex').slice(0, 12)
const ohneKorpus = ['--', '.', ':!**/testdata/fuzz/**']
const standJetzt = () => ({
  commit: git(['rev-parse', '--short', 'HEAD']).trim(),
  offen: git(['status', '--porcelain', ...ohneKorpus]).split('\n').filter(Boolean).length,
  inhalt: abdruck([git(['diff', 'HEAD', '--binary', ...ohneKorpus]), ...dateien(git(['ls-files', '-z', '--others', '--exclude-standard', ...ohneKorpus]))]),
  saat: abdruck(dateien(git(['ls-files', '-z', '--cached', '--others', '--exclude-standard', '--', ':(glob)**/testdata/fuzz/**']))),
})
const anfang = standJetzt()

const ergebnisse = []
const modul = readFileSync(join(wurzel, 'go.mod'), 'utf8').match(/^module (\S+)/m)[1]
const zwischen = execFileSync('go', ['env', 'GOCACHE'], { cwd: wurzel, encoding: 'utf8' }).trim()
const goTest = (argumente) => spawnSync('go', argumente, { cwd: wurzel, encoding: 'utf8', maxBuffer: 64 << 20 })

for (const ziel of ziele) {
  const argumente = ['test', `./${ziel.paket}`, '-run', `^${ziel.name}$`, '-fuzz', `^${ziel.name}$`,
    '-fuzztime', zeit, '-fuzzminimizetime', '5s', '-parallel', parallel]
  const anfang = Date.now()
  const lauf = spawnSync('go', argumente, { cwd: wurzel, encoding: 'utf8', env: { ...process.env, GOMEMLIMIT: speicher }, maxBuffer: 64 << 20 })
  const sekunden = Math.round((Date.now() - anfang) / 1000)

  // **Jeder Teil wird gelesen** (`fuzzbericht.mjs`): der Fuzzlauf, die
  // Übernahme des erarbeiteten Korpus und der Abdeckungslauf.
  const fuzz = fuzzlaufAuswerten(lauf)
  const gemessen = anteilMessen(ziel, { wurzel, zwischen, modul, fund: fuzz.ausgang === 'Fund', goTest })
  const e = zielErgebnis(ziel, fuzz, gemessen, sekunden)

  ergebnisse.push(e)
  const stand = e.ausgang === 'Fund' ? `FUND ${e.fund}` : e.ausgang === 'unvollständig' ? `UNVOLLSTÄNDIG — ${e.fehler.join('; ')}` : 'ohne Fund'
  console.log(`${ziel.name} (${ziel.paket}): ${stand} — ${e.ausfuehrungen} Ausführungen, ${e.neu} neue`)
}

const pruefstand = standPruefen(anfang, standJetzt())
const goFassung = execFileSync('go', ['version'], { cwd: wurzel, encoding: 'utf8' }).trim()
const status = kampagnenstatus(ergebnisse)

const kopf = { ...pruefstand, zeit: new Date().toISOString().slice(0, 16).replace('T', ' '), go: goFassung, budget: zeit, parallel, speicher }
writeFileSync(join(wurzel, bericht), berichtText(ergebnisse, kopf))
if (process.env.ROHDATEN) writeFileSync(process.env.ROHDATEN, `${JSON.stringify({ ergebnisse, kopf }, null, 1)}\n`)
console.log(`${status.funde.length} Funde, ${status.unvollstaendig.length} unvollständig${pruefstand.geaendert ? ', der Stand änderte sich während des Laufs' : ''} → ${bericht}`)
// **0, 1 (Fund) oder 3 (ein Teil gescheitert)** — eine Kampagne, deren
// Abdeckungslauf nicht lief, endet nicht grün; **ebenso eine, deren Stand
// sich während des Laufs änderte**. Der Fund wiegt schwerer.
process.exit(status.exitcode === 0 && pruefstand.geaendert ? 3 : status.exitcode)
