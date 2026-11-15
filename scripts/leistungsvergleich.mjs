#!/usr/bin/env node
// **Der Leistungsvergleich mit MSE2** — QA-04
// (#1735).
//
//     node scripts/leistungsvergleich.mjs                       # alle Teile
//     node scripts/leistungsvergleich.mjs start navigation      # nur diese
//     WIEDERHOLUNGEN=6 GROSS=3 LANGLAUF=120 LAST=10 node scripts/leistungsvergleich.mjs
//
// Was gepaart wird, unter welchen Bedingungen und mit welchem Vorbehalt, steht
// in `docs/LEISTUNGSVERGLEICH.md`; der Beleg landet in
// `docs/belege/leistungsvergleich.json`, der Bericht daneben als `.md`.
//
// # Was er braucht
//
// Ein gebautes MSE2 (`MSE2_BINAER`) mit dem Verweis `data` auf die Pakete
// daneben — `scripts/mse2-referenz.sh` legt ihn —, die Pakete
// (`MSE_FREMDPAKETE`), die kanonischen Sets (`MSE_KANON`) und die Schriften
// für MSE2 (`MSE2_SCHRIFTEN`). **Fehlt eines, misst er nicht** und sagt, was
// fehlt. MSE4Mac baut er aus diesem Stand, 0.5.0 aus dem Tag `v0.5.0` in einem
// eigenen Arbeitsbaum.
//
// # Was er nicht anfasst
//
// **Nichts unter `~/Library`** außer zu lesen: jedes Programm bekommt ein
// eigenes Zuhause und einen eigenen Bestand unter `ARBEIT` (Vorgabe
// `$TMPDIR/leistungsvergleich`). Die Pakete liegen dort als **APFS-Klon**
// (`cp -c`): ein Verweis genügt nicht, denn `filepath.WalkDir` folgt keinem
// (`internal/msepaket/paketwurzeln.go`), und eine Kopie belegte ein Gigabyte.

import { execFileSync, spawn } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, readdirSync, rmSync, statSync, writeFileSync } from 'node:fs'
import { cpus, loadavg, tmpdir, totalmem } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { performance } from 'node:perf_hooks'
import { fileURLToPath } from 'node:url'

import { kartenvorlagen, waehleKarten } from './leistungsauswahl.mjs'
import { fahreTeile, schalterFuer } from './leistungskontingent.mjs'
import { belegJson, berichtMarkdown, bilanz, bilanzZeile, schritteZusammenfassen, zusammenfassen } from './leistungsbericht.mjs'
import { bilderImVerzeichnis, kartenImSet, quelltextwert, sitzungsschritte, wiederhole } from './leistungsphasen.mjs'
import { prozesse, residentByte } from './leistungsprozess.mjs'
import { mse2, mse4mac } from './leistungsteilnehmer.mjs'

const wurzel = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const u = process.env
const k = {
  mse2: u.MSE2_BINAER ?? join(u.HOME, 'mse2-quelle/build/magicseteditor'),
  pakete: u.MSE_FREMDPAKETE ?? '/Users/david/full-magic-pack/data',
  kanon: u.MSE_KANON ?? join(u.HOME, 'Library/Application Support/MagicSetEditor4Mac/kanonische-sets'),
  schriften: u.MSE2_SCHRIFTEN ?? join(u.HOME, 'Library/Fonts/mse2-referenz'),
  arbeit: u.ARBEIT ?? join(tmpdir(), 'leistungsvergleich'),
  wiederholungen: Number(u.WIEDERHOLUNGEN ?? 6),
  navigation: Number(u.NAVIGATION ?? 2),
  bilder: Number(u.BILDER ?? 1),
  gross: Number(u.GROSS ?? 3),
  langlauf: Number(u.LANGLAUF ?? 120),
  langlaufTeilnehmer: (u.LANGLAUF_TEILNEHMER ?? 'mse2,heute').split(','),
  last: Number(u.LAST ?? 10),
  ruheHoechstens: Number(u.RUHE_HOECHSTENS ?? 900),
  lastHoechstens: Number(u.LAST_HOECHSTENS ?? 60),
  lastBudget: Number(u.LAST_BUDGET ?? 600),
  beleg: u.BELEG ?? join(wurzel, 'docs/belege/leistungsvergleich.json'),
}
const alleTeile = ['start', 'oeffnen', 'sichern', 'ausfuhr', 'skript', 'navigation', 'bilder', 'gross', 'langlauf', 'last']
const teile = process.argv.slice(2).length > 0 ? process.argv.slice(2) : alleTeile
const { messe, starte } = prozesse(spawn)
const sh = (befehl, argumente, optionen = {}) => execFileSync(befehl, argumente, { encoding: 'utf8', maxBuffer: 1 << 28, ...optionen }).trim()
const melde = (f) => console.log(`  ${f.aufgabe} ${f.satz} ${f.programm} #${f.wiederholung}: ` +
  `${Math.round(f.wandMs)} ms${f.fehler ? ` — ${f.fehler}` : ''}`)

const fehlt = [
  [k.mse2, 'MSE2 (MSE2_BINAER)'], [join(dirname(k.mse2), 'data'), 'der Verweis data neben MSE2 (scripts/mse2-referenz.sh)'],
  [k.pakete, 'die Pakete (MSE_FREMDPAKETE)'], [k.kanon, 'die kanonischen Sets (MSE_KANON)'],
  [k.schriften, 'die Schriften für MSE2 (MSE2_SCHRIFTEN)'],
].filter(([pfad]) => !existsSync(pfad))
const unbekannt = teile.filter((t) => !alleTeile.includes(t))
if (fehlt.length > 0 || unbekannt.length > 0) {
  for (const [pfad, was] of fehlt) console.error(`fehlt: ${was} — ${pfad}`)
  for (const t of unbekannt) console.error(`unbekannter Teil: ${t} — es gibt ${alleTeile.join(', ')}`)
  console.error('Es wird nicht gemessen.')
  process.exit(1)
}
mkdirSync(k.arbeit, { recursive: true })

// ------------------------------------------------------------ Bauen

/** bauen übersetzt mse, mse-mcp und mse-rest aus `quelle` nach `ziel` — einmal je Stand. */
function bauen(quelle, ziel) {
  if (existsSync(join(ziel, 'mse-rest'))) return ziel
  mkdirSync(ziel, { recursive: true })
  sh('bash', ['-c', `source "${wurzel}/scripts/activate.sh" >/dev/null && cd "${quelle}" && ` +
    ['mse', 'mse-mcp', 'mse-rest'].map((p) => `go build -o "${ziel}/${p}" ./cmd/${p}`).join(' && ')])
  return ziel
}

const commit = sh('git', ['-C', wurzel, 'rev-parse', '--short', 'HEAD'])
const baumSauber = sh('git', ['-C', wurzel, 'status', '--porcelain']) === ''
console.log(`== bauen: MSE4Mac ${commit} und 0.5.0`)
// **Ältere Bauten dieses Stands weg** — je Commit ein Satz von drei
// Programmen zu je 80 MB, und ein Läufer, der nach jedem Commit neu baut,
// hinterließe sie sonst alle.
for (const alt of readdirSync(k.arbeit).filter((n) => n.startsWith('heute-') && n !== `heute-${commit}` && n !== 'heute-zuhause')) {
  rmSync(join(k.arbeit, alt), { recursive: true, force: true })
}
const heuteBin = bauen(wurzel, join(k.arbeit, `heute-${commit}`, 'bin'))
const altQuelle = join(k.arbeit, 'v050-quelle')
if (!existsSync(altQuelle)) sh('git', ['-C', wurzel, 'worktree', 'add', '--detach', altQuelle, 'v0.5.0'])
const altBin = bauen(altQuelle, join(k.arbeit, 'v050', 'bin'))

/** zuhauseMitPaketen legt ein Zuhause an, dessen Paketverzeichnis das Pack als Klon trägt. */
function zuhauseMitPaketen(ort) {
  const pack = join(ort, 'Library/Application Support/MagicSetEditor4Mac/pakete/full-magic-pack')
  if (!existsSync(join(pack, 'data'))) {
    mkdirSync(pack, { recursive: true })
    sh('cp', ['-cR', k.pakete, join(pack, 'data')])
    sh('cp', ['-cR', join(dirname(k.pakete), 'Magic - Fonts'), join(pack, 'Magic - Fonts')])
  }
  return ort
}

// **Ein Zuhause je MSE2-Prozess**, gesammelt an einem Ort, der zu Beginn
// jedes Laufs geleert wird.
const mse2Zuhause = join(k.arbeit, 'mse2-zuhause')
rmSync(mse2Zuhause, { recursive: true, force: true })
mkdirSync(mse2Zuhause, { recursive: true })
const teilnehmer = [
  mse2({ binaer: k.mse2, arbeit: mse2Zuhause, messe, starte }),
  mse4mac({ name: 'heute', verzeichnis: heuteBin, zuhause: zuhauseMitPaketen(join(k.arbeit, 'heute-zuhause')), messe, starte }),
  mse4mac({ name: 'v050', verzeichnis: altBin, zuhause: zuhauseMitPaketen(join(k.arbeit, 'v050-zuhause')), messe, starte }),
]
const mse4macs = teilnehmer.filter((t) => t.name !== 'mse2')

/** bestandsort gibt einen frischen Ort für eine Datenbank — ein eigenes Verzeichnis, denn die Bilder folgen ihr. */
function bestandsort(t, name) {
  const ort = join(k.arbeit, 'bestaende', `${t.name}-${name}`)
  rmSync(ort, { recursive: true, force: true })
  mkdirSync(ort, { recursive: true })
  return join(ort, 'bestand.db')
}
const satz = (name) => ({ name, pfad: join(k.kanon, `${name}.mse-set`) })
const setLesen = (pfad) => {
  try {
    return statSync(pfad).isDirectory() ? readFileSync(join(pfad, 'set'), 'utf8') : sh('unzip', ['-p', pfad, 'set'])
  } catch {
    return null
  }
}

console.log('== Bestände vorbereiten')
const kleine = ['Archenemy', 'Planechase', 'Vanguard', 'Multicolor']
for (const t of mse4macs) {
  t.leer = bestandsort(t, 'leer')
  await t.start(t.leer)
  t.vorrat = bestandsort(t, 'vorrat')
  t.ids = {}
  for (const name of kleine) {
    const m = await t.oeffnen(satz(name), t.vorrat)
    if (m.fehler || !m.setID) throw new Error(`${t.name}: ${name} lässt sich nicht einfahren — ${m.fehler}\n${m.fehlerText}`)
    t.ids[name] = m.setID
  }
}

// ------------------------------------------------------------ Auswahl (AK-04)

const planPfad = join(wurzel, 'docs/belege/kanonstichprobe.json')
const vorlagen = Object.fromEntries([...kleine, 'Black'].map((name) => [name, kartenvorlagen(setLesen(satz(name).pfad))]))
const kartenzahl = Object.fromEntries(Object.entries(vorlagen).map(([name, karten]) => [name, karten.length]))
const auswahl = waehleKarten(JSON.parse(readFileSync(planPfad, 'utf8')), vorlagen, [
  { schicht: 'Normal', satz: 'Multicolor', vorlage: 'm15-altered', anzahl: 12 },
  ...['Archenemy', 'Planechase', 'Vanguard'].map((s) => ({ schicht: 'Sonderformat', satz: s, anzahl: 6 })),
  ...['m15-aftermath', 'm15-mainframe-planeswalker', 'm15-mainframe-dfc'].map((v) => ({ schicht: 'Problemfall', satz: 'Multicolor', vorlage: v, anzahl: 3 })),
])

// ------------------------------------------------------------ Ruhe

/** pipelineLaeuft sagt, ob auf diesem Rechner gerade eine Pipeline oder ein Riegel läuft — aus irgendeinem Baum. */
const pipelineLaeuft = () => /scripts\/ci\.sh|vorpruefung\.sh/.test(sh('ps', ['-axo', 'command']))

/**
 * warteAufRuhe wartet vor jedem Fall und jeder Sitzung, bis **keine Pipeline**
 * läuft — höchstens `RUHE_HOECHSTENS` Sekunden — und danach, bis die Last
 * unter der Kernzahl liegt: höchstens `LAST_HOECHSTENS` Sekunden je Fall und
 * `LAST_BUDGET` Sekunden im ganzen Lauf. **Die zweite Grenze hat ein Budget**,
 * weil auf diesem Rechner mehrere Arbeitsbäume zugleich prüfen und eine Last
 * über zehn eher die Regel ist als die Ausnahme — ohne Budget wartete ein Lauf
 * mit zweihundert Fällen Stunden. Der Fall trägt, wie lange gewartet wurde,
 * und `last`, wie hoch sie war.
 */
let lastGewartetS = 0
async function warteAufRuhe() {
  const anfang = performance.now()
  const vergangen = () => (performance.now() - anfang) / 1000
  const pause = () => new Promise((weiter) => setTimeout(weiter, 5000))
  while (pipelineLaeuft() && vergangen() < k.ruheHoechstens) await pause()
  const lastAnfang = vergangen()
  const erlaubt = Math.min(k.lastHoechstens, k.lastBudget - lastGewartetS)
  while (loadavg()[0] > cpus().length && vergangen() - lastAnfang < erlaubt) await pause()
  lastGewartetS += vergangen() - lastAnfang
  return Math.round(vergangen() * 10) / 10
}

/** mitRuhe misst einen Fall erst in Ruhe und sagt danach, ob inzwischen eine Pipeline lief. */
const mitRuhe = (teilnehmer, anzahl, aufgabe, satz, einmal, meldung) => wiederhole(teilnehmer, anzahl, aufgabe, satz,
  async (t, i) => ({ ...await einmal(t, i), pipelineDanach: pipelineLaeuft() }), meldung, warteAufRuhe)

// ------------------------------------------------------------ Die Teile

const anfang = new Date().toISOString()
const c = {
  k, teilnehmer, mse4macs, satz, setLesen, kartenzahl, vorlagen, auswahl, bestandsort, melde, messe, starte,
  wiederhole: mitRuhe, warte: warteAufRuhe, pipelineLaeuft, quelltextwert, sitzungsschritte, bilderImVerzeichnis, kartenImSet, residentByte,
  faelle: [], schritte: [], befunde: [], extra: {},
  ausgabe(name) {
    const ort = join(k.arbeit, 'aus', name)
    rmSync(ort, { recursive: true, force: true })
    mkdirSync(ort, { recursive: true })
    return ort
  },
}

// ------------------------------------------------------------ Beleg

const hash = (pfad) => sh('shasum', ['-a', '256', pfad]).split(' ')[0]
const ueberVerzeichnis = (ort) => sh('bash', ['-c', `cd "${ort}" && find . -type f ! -name '.*' -print0 | sort -z | xargs -0 shasum -a 256 | shasum -a 256`]).split(' ')[0]
const fest = {
  format: 1, lauf: 'leistungsvergleich', commit, baumSauber, anfang,
  befehl: `node scripts/leistungsvergleich.mjs ${teile.join(' ')}`,
  maschine: {
    chip: sh('sysctl', ['-n', 'machdep.cpu.brand_string']), kerne: cpus().length, speicherByte: totalmem(),
    system: `${sh('sw_vers', ['-productName'])} ${sh('sw_vers', ['-productVersion'])} (${sh('sw_vers', ['-buildVersion'])})`,
    node: process.version,
  },
  programme: {
    mse2: { binaer: k.mse2, sha256: hash(k.mse2), quelle: 'f9d9356 vom 2021-07-06', patch: 'docs/mse2-macos.patch', wx: sh('bash', ['-c', `otool -L "${k.mse2}" | grep -o 'libwx_baseu-[0-9.]*' | head -1`]) },
    heute: { meta: sh(join(heuteBin, 'mse'), ['meta']) },
    v050: { meta: sh(join(altBin, 'mse'), ['meta']), tag: 'v0.5.0', commit: sh('git', ['-C', altQuelle, 'rev-parse', '--short', 'HEAD']) },
  },
  eingaben: {
    sets: Object.fromEntries([...kleine, 'Black'].map((name) => [name, { karten: kartenzahl[name], sha256: hash(satz(name).pfad) }])),
    plan: { pfad: 'docs/belege/kanonstichprobe.json', sha256: hash(planPfad) },
    pakete: { pfad: k.pakete, anzahl: Number(sh('bash', ['-c', `ls "${k.pakete}" | wc -l`])) },
    schriften: { mse2: k.schriften, sha256: ueberVerzeichnis(k.schriften) },
  },
}

/**
 * belegSchreiben legt den Beleg ab — **nach jedem Teil**, nicht erst am Ende:
 * ein Lauf, der nach zwanzig Minuten abbricht, hinterlässt sonst nichts.
 * `stand` sagt, ob er vollständig ist.
 */
function belegSchreiben(stand, lasten) {
  const beleg = {
    ...fest, stand, ende: new Date().toISOString(),
    bedingungen: {
      wiederholungen: k.wiederholungen, navigation: k.navigation, bilder: k.bilder, gross: k.gross,
      langlaufSekunden: k.langlauf, langlaufTeilnehmer: k.langlaufTeilnehmer, lastSekunden: k.last, teile, lasten,
      ruhe: { pipelineHoechstensS: k.ruheHoechstens, lastHoechstensS: k.lastHoechstens, lastBudgetS: k.lastBudget, aufLastGewartetS: Math.round(lastGewartetS) },
    },
    auswahl, befunde: c.befunde, ...c.extra,
    // **Was der Lauf gefunden hat, Fälle eingerechnet** — nicht nur, was er als
    // Ganzes angemerkt hat.
    bilanz: bilanz({ faelle: c.faelle, schritte: c.schritte, befunde: c.befunde }),
    zusammenfassung: zusammenfassen(c.faelle),
    schrittzusammenfassung: schritteZusammenfassen(c.schritte),
    // **Die Fehlerausgabe nur, wo ein Fehler war** — sonst stünde bei jedem
    // Lauf von MSE2 dasselbe Protokoll der Warnungen beim Laden im Beleg.
    faelle: c.faelle.map(({ fehlerText, ...f }) => (f.fehler ? { ...f, fehlerText } : f)),
    schritte: c.schritte,
  }
  writeFileSync(k.beleg, belegJson(beleg))
  writeFileSync(k.beleg.replace(/\.json$/, '.md'), berichtMarkdown({ ...beleg, schritte: beleg.schrittzusammenfassung }))
}

const funktionen = await import('./leistungsteile.mjs')
const lasten = {}

// **Das Guthaben von `heute` ist abgeschaltet, wo es nicht Gegenstand ist**
// (`leistungskontingent.mjs`): ein Vergleich misst nicht die Drosselung.
// Zurückgestellt wird vor der Lastgrenze und am Ende — und bei einem Signal.
const schalter = schalterFuer(mse4macs.find((t) => t.name === 'heute'))
for (const signal of ['SIGINT', 'SIGTERM', 'SIGHUP']) {
  process.once(signal, () => {
    schalter.zurueck().catch((fehler) => console.error(`Die Kontingente ließen sich nicht zurückstellen: ${fehler.message}`))
      .finally(() => process.exit(130))
  })
}
await fahreTeile(teile, schalter, async (teil) => {
  console.log(`== ${teil} (Last ${loadavg()[0].toFixed(1)})`)
  const vorher = loadavg()[0]
  await funktionen[teil](c)
  lasten[teil] = { vorher, nachher: loadavg()[0] }
  belegSchreiben(`Zwischenstand nach ${teil}`, lasten)
})
belegSchreiben('vollständig', lasten)
const gefunden = bilanz({ faelle: c.faelle, schritte: c.schritte, befunde: c.befunde })
console.log(`Beleg: ${k.beleg} — ${bilanzZeile(gefunden)}`)
for (const befund of gefunden.beispiele) console.log(`  Befund: ${befund}`)
// **Ein Lauf mit Befund endet nicht grün.** Der Beleg liegt trotzdem da: wer
// ihn liest, soll sehen, was nicht stimmt, statt eine Zahl zu zitieren.
process.exit(gefunden.befunde > 0 ? 1 : 0)
