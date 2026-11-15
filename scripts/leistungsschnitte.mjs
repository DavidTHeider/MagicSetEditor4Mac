// **Der Messvertrag der Leistung** — PERF-01
// (#2168).
//
//     node scripts/leistungsschnitte.mjs                    # alle Schnitte
//     node scripts/leistungsschnitte.mjs ast-frisch mcp     # nur diese
//     WIEDERHOLUNGEN=5 DAUER=2s LASTGRENZE=5 FREMDGRENZE=250 PAUSE=120 node …
//     ERZWINGE_PAUSE=1 PAUSE=10 BELEG=/tmp/x.json node …    # der Gegenfall
//
// Für die Schnitte über dem Pack: `MSE_FREMDPAKETE` (das `data` des Packs)
// und `MSE_KANON` (die kanonischen Sets). Fehlt eins, gelten sie als nicht
// gemessen, und der Lauf als unvollständig.
//
// # Was er zusagt — und woran es hängt
//
// - **Benannte Schnitte** (AK-01): jeder ist eine Messung, die es im
//   Repositorium schon gibt, und fährt als eigener Prozess aus einem einmal
//   gebauten Testprogramm — dessen SHA-256 steht im Beleg, ebenso der der
//   Eingänge. **AST frisch und AST aus dem Speicher sind getrennt**, und der
//   zweite heißt, was er ist: nur Treffer (AK-02).
// - **Die Umgebung vor und nach jedem Schnitt** (AK-03): Last und die fünf
//   größten fremden Prozesse. Liegt sie über der vorher festgelegten Grenze,
//   wartet der Schnitt bis `PAUSE` Sekunden; danach gilt er als nicht
//   gemessen und der Lauf als **unvollständig**. Fremde Prozesse hält der
//   Lauf nicht an. `ERZWINGE_PAUSE=1` setzt die Grenze unter jede Last — der
//   Gegenfall, dass die Pause greift. **Während des Schnitts** gilt die
//   Grenze für fremde Prozesse weiter, mit einer Probe alle fünf Sekunden;
//   darüber heißt er „gestört, nicht gemessen" (Befund 4 des Zweitblicks
//   über Phase I).
// - **Speicher, getrennt benannt** (AK-04): `B/op` und `allocs/op` sind die
//   Go-Zuteilung je Durchlauf; „RSS" und „Spitze" sind der höchste residente
//   Stand und der Spitzenverbrauch des **ganzen** Testprozesses nach
//   `/usr/bin/time -l`, Aufbau eingeschlossen. **Nicht erhoben**: der lebende
//   Heap, der Adressraum von SQLite als WebAssembly.
// - **Echter Transport** (AK-05): `mse-mcp` über stdio, dieselben drei
//   Ausdrücke wie der Leistungsvergleich (`leistungsteile.mjs`, `skript`),
//   Start und Aufwärmen ausgeschlossen, gemessen von der geschriebenen Zeile
//   bis zur gelesenen Antwort (`McpSitzung.werkzeug`).
//
// **Was er nicht ist**: ein Vergleich zweier Stände. Für den gilt die Regel
// aus `docs/LEISTUNGSSCHNITTE.md` — A/B/A unter derselben Grenze.

import { createHash } from 'node:crypto'
import { execFileSync, spawn } from 'node:child_process'
import { existsSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, statSync, writeFileSync } from 'node:fs'
import { cpus, loadavg, tmpdir, totalmem } from 'node:os'
import { dirname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

import { prozesse } from './leistungsprozess.mjs'
import { auswaehlen, beobachte, benchLesen, fremdLast, gestoert, warteAufRuhe } from './leistungsschnittteile.mjs'
import { verteilung } from './leistungsstatistik.mjs'
import { mse4mac } from './leistungsteilnehmer.mjs'

const wurzel = fileURLToPath(new URL('..', import.meta.url))
const zahl = (name, vorgabe) => Number(process.env[name] ?? vorgabe)
const k = {
  wiederholungen: zahl('WIEDERHOLUNGEN', 5),
  dauer: process.env.DAUER ?? '2s',
  pause: zahl('PAUSE', 120),
  grenzen: process.env.ERZWINGE_PAUSE === '1'
    ? { last: -1, fremd: -1 }
    : { last: zahl('LASTGRENZE', cpus().length / 2), fremd: zahl('FREMDGRENZE', 250) },
  rufe: zahl('RUFE', 100),
  pakete: process.env.MSE_FREMDPAKETE ?? '',
  kanon: process.env.MSE_KANON ?? '',
}

const schnitte = [
  { name: 'ast-frisch', was: 'Zerteilen, frischer Neuaufbau des Baums', paket: 'msescript', bench: '^BenchmarkZerteileKaltUndWarm$/^kalt$' },
  { name: 'ast-speicher', was: 'Zerteilen, nur Treffer im Baumspeicher', paket: 'msescript', bench: '^BenchmarkZerteileKaltUndWarm$/^warm$' },
  { name: 'ebene', was: 'Rasterung einer Ebene, kalt und warm', paket: 'render', bench: '^BenchmarkEbeneKaltUndWarm$' },
  { name: 'einfahren', was: 'Kartenlesen: ein Set einfahren', paket: 'app', pack: true, bench: '^BenchmarkWegEinfahren$' },
  { name: 'naht-skript', was: 'öffentliche Naht: ein Ausdruck je Ruf', paket: 'app', pack: true, bench: '^BenchmarkWegKopflos$/^Skript_je_Ruf$' },
  { name: 'naht-karte', was: 'öffentliche Naht: eine Karte je Ruf', paket: 'app', pack: true, bench: '^BenchmarkWegKopflos$/^Karte_je_Ruf$' },
  { name: 'werk-frisch', was: 'ganze Rechnung: frisches Werk, Umgebung samt Abbildung', paket: 'app', pack: true, bench: '^BenchmarkWegKopflos$/^Karte,_frisches_Werk$' },
  { name: 'werk-gehalten', was: 'ganze Rechnung: das Werk gehalten, der Weg des Fensters', paket: 'app', pack: true, bench: '^BenchmarkWegKopflos$/^Karte,_Werk_gehalten$' },
  { name: 'ausfuhr', was: 'Ausfuhr über eine Vorlage, 70 Karten', paket: 'app', pack: true, bench: '^BenchmarkWegAusfuhr$' },
  { name: 'rastern', was: 'Rasterung einer Karte im Fensterweg', paket: 'app', pack: true, bench: '^BenchmarkWegRastern$' },
  { name: 'paare', was: 'dieselbe Karte vor, sofort und nach fünf anderen (PERF-04)', paket: 'app', pack: true, einmal: true, bench: '^BenchmarkKartenpaare$' },
  { name: 'mcp', was: 'echter Transport: mse-mcp über stdio, drei Ausdrücke', pack: true, transport: true },
]

const sh = (befehl, argumente, optionen = {}) => execFileSync(befehl, argumente, { cwd: wurzel, encoding: 'utf8', maxBuffer: 64 << 20, ...optionen })
const sha = (pfad) => createHash('sha256').update(readFileSync(pfad)).digest('hex')
const schlafe = (s) => new Promise((los) => { setTimeout(los, s * 1000) })
const umgebungMessen = () => ({
  last: loadavg().map((w) => Math.round(w * 100) / 100),
  fremd: fremdLast(sh('ps', ['-A', '-o', 'pid=,ppid=,pcpu=,comm=']), process.pid),
})

// packabdruck ist der Abdruck des Packs über Namen und Größen — der Inhalt
// von 4.000 Dateien wäre eine eigene Messung.
function packabdruck(ordner) {
  const summe = createHash('sha256')
  const dateien = readdirSync(ordner, { recursive: true, withFileTypes: true })
    .filter((e) => e.isFile()).map((e) => join(e.parentPath, e.name)).sort()
  for (const datei of dateien) summe.update(`${relative(ordner, datei)}\0${String(statSync(datei).size)}\n`)

  return { dateien: dateien.length, abdruck: summe.digest('hex') }
}

function stand() {
  const commit = sh('git', ['rev-parse', 'HEAD']).trim()
  const sauber = sh('git', ['status', '--porcelain']).trim() === ''
  const go = sh('bash', ['-c', 'source scripts/activate.sh >/dev/null 2>&1 && go version']).trim()

  return {
    commit, sauber, go, zeit: new Date().toISOString(),
    rechner: sh('sysctl', ['-n', 'machdep.cpu.brand_string']).trim(), kerne: cpus().length,
    speicherGB: Math.round(totalmem() / 2 ** 30),
    system: `${sh('sw_vers', ['-productVersion']).trim()} (${sh('sw_vers', ['-buildVersion']).trim()})`,
    wiederholungen: k.wiederholungen, dauer: k.dauer, grenzen: k.grenzen, pauseS: k.pause, rufe: k.rufe,
  }
}

function bauen(ort) {
  const programme = {}
  const go = (argumente) => sh('bash', ['-c', `source scripts/activate.sh >/dev/null 2>&1 && go ${argumente}`])
  for (const paket of new Set(schnitte.filter((s) => s.paket).map((s) => s.paket))) {
    const ziel = join(ort, `${paket}.test`)
    go(`test -c -tags fremdpakete -o "${ziel}" ./internal/${paket}/`)
    programme[paket] = { pfad: ziel, sha256: sha(ziel) }
  }
  for (const p of ['mse', 'mse-mcp']) {
    go(`build -o "${join(ort, p)}" ./cmd/${p}`)
    programme[p] = { pfad: join(ort, p), sha256: sha(join(ort, p)) }
  }

  return programme
}

async function benchSchnitt(s, programme, messe, ort, packzuhause) {
  const zuhause = mkdtempSync(join(ort, 'zuhause-'))
  const paarbeleg = join(zuhause, 'paare.json')
  const lauf = await messe(programme[s.paket].pfad, [
    '-test.run', '^$', '-test.bench', s.bench, '-test.benchmem',
    '-test.count', s.einmal ? '1' : String(k.wiederholungen),
    '-test.benchtime', s.einmal ? '1x' : s.pack ? '5x' : k.dauer,
  ], {
    ort: join(wurzel, 'internal', s.paket),
    umgebung: {
      ...process.env, HOME: zuhause, MSE_FREMDPAKETE: k.pakete, MSE_KANON: k.kanon,
      MSE_PAKETZUHAUSE: packzuhause ?? '', MSE_KARTENPAARE_BELEG: s.einmal ? paarbeleg : '',
    },
  })
  const werte = [...benchLesen(lauf.aus)].map(([name, w]) => ({
    name, ns: verteilung(w.ns), bytes: verteilung(w.bytes), zuteilungen: verteilung(w.zuteilungen),
    durchlaeufe: w.durchlaeufe,
  }))

  const paare = s.einmal && existsSync(paarbeleg) ? JSON.parse(readFileSync(paarbeleg, 'utf8')) : undefined

  return {
    ok: lauf.fehler === null && werte.length > 0 && (!s.einmal || paare !== undefined), fehler: lauf.fehler, werte, paare,
    prozess: { wandMs: Math.round(lauf.wandMs), cpuS: lauf.cpuS, rssByte: lauf.rssByte, spitzeByte: lauf.spitzeByte },
    rest: lauf.fehler === null ? undefined : lauf.aus.slice(-1500) + lauf.fehlerText,
  }
}

// packzuhauseAnlegen klont das Pack **einmal** in ein Zuhause, das alle
// Schnitte teilen — ein Klon je Prozess weckte `fseventsd` während der
// gemessenen Schleife (gemessen: 250–280 % CPU). Die Pause vor dem ersten
// Schnitt wartet, bis es ruhig ist.
function packzuhauseAnlegen(ort) {
  const zuhause = join(ort, 'packzuhause')
  const pack = join(zuhause, 'Library', 'Application Support', 'MagicSetEditor4Mac', 'pakete', 'full-magic-pack')
  mkdirSync(pack, { recursive: true })
  sh('cp', ['-cR', k.pakete, join(pack, 'data')])
  sh('cp', ['-cR', join(dirname(k.pakete), 'Magic - Fonts'), join(pack, 'Magic - Fonts')])

  return zuhause
}

async function transportSchnitt(ort, zuhause, { messe, starte }) {
  const bestand = join(mkdtempSync(join(ort, 'mcp-')), 'bestand.db')
  const t = mse4mac({ name: 'heute', verzeichnis: ort, zuhause, messe, starte })
  const geoeffnet = await t.oeffnen({ pfad: join(k.kanon, 'Archenemy.mse-set') }, bestand)
  if (geoeffnet.setID === null) return { ok: false, fehler: `oeffnen: ${String(geoeffnet.fehler)}`, werte: [] }
  // **Gemessen wird die Maschine, nicht die Drossel**: das Guthaben von
  // `mse-mcp` ist nach rund hundert Rufen leer (`leistungskontingent.mjs`).
  const aus = await t.kontingente(bestand, 'aus')
  if (aus.fehler !== null) return { ok: false, fehler: `Guthaben: ${aus.fehler}`, werte: [] }
  const sitzung = await t.sitzung({ id: geoeffnet.setID }, bestand)
  const ausdruecke = [['Rechnung', '1 + 1', '2'], ['Zeichenkette', 'to_upper("abc")', '"ABC"'], ['Kartenfeld', 'card.name', null]]
  const stelle = 3
  for (let i = 0; i < 10; i++) for (const [, a] of ausdruecke) await sitzung.skript(a, stelle)
  const werte = []
  for (const [name, ausdruck, erwartet] of ausdruecke) {
    const dauern = []
    let falsch = 0
    let erster
    for (let i = 0; i < k.rufe; i++) {
      const r = await sitzung.skript(ausdruck, stelle)
      if (!r.ok || (erwartet !== null && r.text !== erwartet)) {
        falsch++
        erster ??= r.text
      }
      dauern.push(r.dauerMs)
    }
    werte.push({ name: `lesen.skript: ${name}`, ms: verteilung(dauern.filter((d) => d !== null)), falsch, erster })
  }
  const ende = await sitzung.lauf.schliessen()

  return {
    ok: werte.every((w) => w.falsch === 0), fehler: ende.fehler, werte, bereitMs: Math.round(sitzung.bereitMs),
    prozess: { rssByte: ende.rssByte, spitzeByte: ende.spitzeByte, cpuS: ende.cpuS },
  }
}

async function hauptlauf() {
  const gewuenscht = process.argv.slice(2)
  const auswahl = auswaehlen(schnitte, gewuenscht)
  const ort = mkdtempSync(join(tmpdir(), 'leistungsschnitte-'))
  const beleg = { stand: stand(), eingaenge: {}, programme: {}, schnitte: [], vollstaendig: true }
  if (k.pakete !== '' && k.kanon !== '') {
    beleg.eingaenge = { pack: packabdruck(k.pakete), archenemy: sha(join(k.kanon, 'Archenemy.mse-set')) }
  }
  console.log(`Bauen nach ${ort} …`)
  const programme = bauen(ort)
  for (const [name, p] of Object.entries(programme)) beleg.programme[name] = p.sha256
  const werkzeuge = prozesse(spawn)
  const mitPack = auswahl.some((s) => s.pack) && k.pakete !== '' && existsSync(k.pakete)
  const packzuhause = mitPack ? packzuhauseAnlegen(ort) : undefined

  for (const s of auswahl) {
    const eintrag = { name: s.name, was: s.was, bench: s.bench ?? null }
    beleg.schnitte.push(eintrag)
    if (s.pack && (k.pakete === '' || k.kanon === '' || !existsSync(k.pakete))) {
      Object.assign(eintrag, { status: 'nicht gemessen', grund: 'MSE_FREMDPAKETE oder MSE_KANON fehlt' })
      beleg.vollstaendig = false
      continue
    }
    const vorher = await warteAufRuhe({ messen: umgebungMessen, schlafe, grenzen: k.grenzen, frist: k.pause, takt: 10 })
    eintrag.vorher = { ...vorher.umgebung, gewartetS: vorher.gewartetS }
    if (!vorher.ruhig) {
      Object.assign(eintrag, { status: 'pausiert, nicht gemessen', grund: vorher.grund })
      beleg.vollstaendig = false
      console.log(`⏸ ${s.name}: ${vorher.grund}`)
      continue
    }
    console.log(`▶ ${s.name} (Last ${String(vorher.umgebung.last[0])}, fremd ${String(vorher.umgebung.fremd.summe)} %)`)
    const beobachtung = beobachte(umgebungMessen, 5000)
    const ergebnis = s.transport ? await transportSchnitt(ort, packzuhause, werkzeuge) : await benchSchnitt(s, programme, werkzeuge.messe, ort, packzuhause)
    const { proben, fehlgeschlagen } = beobachtung.ende()
    const stoerung = gestoert(proben, k.grenzen, fehlgeschlagen)
    Object.assign(eintrag, ergebnis, {
      status: !ergebnis.ok ? 'gescheitert' : stoerung === null ? 'gemessen' : 'gestört, nicht gemessen',
      nachher: fehlgeschlagen === 0 ? proben.at(-1) : null,
      waehrend: { proben: proben.length, fehlgeschlagen, fremdHoechstens: Math.max(0, ...proben.map((p) => p.fremd.summe)) },
    })
    if (stoerung !== null) {
      eintrag.grund = stoerung
      console.log(`⚠ ${s.name}: ${stoerung}`)
    }
    if (!ergebnis.ok || stoerung !== null) beleg.vollstaendig = false
  }

  const ziel = process.env.BELEG ?? join(wurzel, 'docs', 'belege', 'leistungsschnitte.json')
  writeFileSync(ziel, `${JSON.stringify(beleg, null, 2)}\n`)
  sh('rm', ['-rf', ort])
  console.log(`${beleg.vollstaendig ? '✓ vollständig' : '✗ unvollständig'} → ${ziel}`)
  process.exitCode = beleg.vollstaendig ? 0 : 1
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await hauptlauf()
