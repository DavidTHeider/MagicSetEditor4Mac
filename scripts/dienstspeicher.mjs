#!/usr/bin/env node
// **Was ein Dienst nach einem Schub Rasterungen behält** — PERF01
// (#2351),
// Befund A-13 des Audits 0.9.0.
//
//     source scripts/activate.sh
//     node scripts/dienstspeicher.mjs            # Beleg in docs/belege/dienstspeicher.md
//     DAUER=10 RUHE=30 node scripts/dienstspeicher.mjs
//
// # Der Vertrag, vor dem Lauf festgelegt
//
// - **Gegen frisch gebaute Programme und einen eigenen Bestand** — wie der
//   Lastlauf (`scripts/lastlauf.mjs`): eigenes Zuhause, eigener Bestand, ein
//   Schlüsselbund, den es nicht gibt.
// - **Ohne Kontingente** (`grenzen.kontingente aus`) — was die Maschine
//   schafft, nicht was der Schutz daraus macht: der ungünstigste Fall.
// - **`mse-rest`**: `DAUER` Sekunden (Vorgabe 30) mit acht Klienten das
//   Profil `rest-rendern` — je Ruf ein Kartenbild in eine Datei.
// - **`mse-mcp` über HTTP**: dieselbe Zeit mit vier Klienten
//   `schreiben.set.bogen` — je Ruf ein Druckbogen des Sets.
// - **Danach `RUHE` Sekunden (Vorgabe 120) ohne Ruf**, und dann der
//   **Fußabdruck** des Prozesses (`footprint`): das, wonach macOS den
//   Speicherdruck misst, und nicht RSS, das zurückgegebene Seiten mitzählt
//   (LEISTUNGSSCHNITTE §9.1). Während der Last alle zwei Sekunden der
//   Höchstwert.
// - **Jeder Dienst zweimal**: ohne und **mit `GOMEMLIMIT`** (Vorgabe
//   `384MiB`, `GRENZE`) — drei Viertel der Grenze aus dem Messvertrag.
// - **Bewertet**: höchstens 512 MiB nach der Ruhe (`scripts/messvertrag.mjs`).
//
// # Was er nicht misst
//
// Eine zweite Maschine; Klient und Dienst teilen sich diesen Rechner. Und
// einen Dienst, der Tage läuft — die Ruhe ist zwei Minuten.
import { execFileSync, spawn } from 'node:child_process'
import { cpSync, mkdirSync, mkdtempSync, rmSync, symlinkSync, writeFileSync } from 'node:fs'
import { cpus, loadavg, tmpdir, totalmem } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { performance } from 'node:perf_hooks'
import { fileURLToPath } from 'node:url'

import { fussabdruck } from './huellenprobe/dauerbetrieb.mjs'
import { Kinderregister, beende, starteNacheinander } from './lastlaufdienste.mjs'
import { mcpUeberHTTP } from './lastlaufnetz.mjs'
import { profile } from './lastlaufprofile.mjs'
import { MESSVERTRAG, tafel, urteil } from './messvertrag.mjs'

const wurzel = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const dauer = Number(process.env.DAUER ?? 30)
const ruhe = Number(process.env.RUHE ?? 120)
const grenze = process.env.GRENZE ?? '384MiB'
const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))

const arbeit = mkdtempSync(join(tmpdir(), 'dienstspeicher-'))
const kinder = new Kinderregister(spawn)
process.on('exit', () => { kinder.beendeSofort(); rmSync(arbeit, { recursive: true, force: true }) })
for (const signal of ['SIGINT', 'SIGTERM', 'SIGHUP']) process.once(signal, () => { kinder.beendeAlle().finally(() => process.exit(130)) })

const zuhause = join(arbeit, 'zuhause')
const anwendungsort = join(zuhause, 'Library', 'Application Support', 'MagicSetEditor4Mac')
const programme = { mse: join(arbeit, 'bin', 'mse'), rest: join(arbeit, 'bin', 'mse-rest'), mcp: join(arbeit, 'bin', 'mse-mcp') }
for (const [name, ziel] of Object.entries({ mse: programme.mse, 'mse-rest': programme.rest, 'mse-mcp': programme.mcp })) {
  execFileSync('go', ['build', '-o', ziel, `./cmd/${name}`], { cwd: wurzel, stdio: 'inherit' })
}
mkdirSync(join(arbeit, 'bilder'), { recursive: true })
mkdirSync(join(arbeit, 'boegen'), { recursive: true })
cpSync(join(wurzel, 'saat'), join(anwendungsort, 'pakete'), { recursive: true })
symlinkSync(join(anwendungsort, 'pakete', 'probe-schlicht.mse-style'), join(anwendungsort, 'pakete', 'probe-stil.mse-style'))

const umgebung = { ...process.env, HOME: zuhause, MSE_BESTAND: join(anwendungsort, 'bestand.db'), MSE_SCHLUESSELBUND: join(arbeit, 'keiner.keychain'), MSE_DIENST_GESPERRT: 'an' }
delete umgebung.GOMEMLIMIT
const mse = (argumente) => execFileSync(programme.mse, argumente, { env: umgebung, encoding: 'utf8' })
const setID = String(JSON.parse(mse(['oeffnen', join(wurzel, 'testdaten', 'probe.mse-set'), '--json'])).setID)
const karten = JSON.parse(mse(['karten', setID, '--json'])).karten.map((karte) => karte.id)

// unterLast fährt `last` und nimmt alle zwei Sekunden den Fußabdruck von `pid`.
async function unterLast(pid, last) {
  let hoechst = 0
  let laeuft = true
  const proben = (async () => {
    while (laeuft) {
      hoechst = Math.max(hoechst, fussabdruck([pid]).get(String(pid)) ?? 0)
      await schlafe(2000)
    }
  })()
  const ergebnis = await last()
  laeuft = false
  await proben

  return { hoechst, ergebnis }
}

// bogenlast ruft mit `klienten` Klienten `schreiben.set.bogen`, bis die Zeit um ist.
async function bogenlast(mcp, klienten) {
  const ende = performance.now() + dauer * 1000
  let gelungen = 0
  let abgewiesen = 0
  let grund = ''
  await Promise.all(Array.from({ length: klienten }, async (_, klient) => {
    for (let lauf = 0; performance.now() < ende; lauf += 1) {
      const antwort = await mcpUeberHTTP(mcp, 'schreiben.set.bogen', { setID: Number(setID), ziel: join(arbeit, 'boegen', `${String(klient)}-${String(lauf)}-${String(Date.now())}.pdf`) })
      if (antwort.status === 200 && !antwort.isError) gelungen += 1
      else { abgewiesen += 1; grund ||= `${String(antwort.status)} ${(antwort.inhalt ?? antwort.text ?? '').slice(0, 160)}` }
    }
  }))

  return { gelungen, abgewiesen, grund }
}

// fall startet einen Dienst, fährt die Last, wartet die Ruhe ab und misst.
async function fall(dienst, mitGrenze) {
  const env = mitGrenze ? { ...umgebung, GOMEMLIMIT: grenze } : umgebung
  const [programm, argumente, verbindungen] = dienst === 'mse-rest' ? [programme.rest, ['-port', '0'], 16] : [programme.mcp, ['-adresse', '127.0.0.1', '-port', '0'], 64]
  mse(['einstellungen', 'setzen', 'grenzen.kontingente', 'aus'])
  const [laufend] = await starteNacheinander(kinder, [[programm, argumente, env, verbindungen]])
  try {
    const pid = laufend.kind.pid
    const anfang = fussabdruck([pid]).get(String(pid))
    const k = { rest: laufend, setID, karten, arbeit, dauer, klienten: 8, fall: `${dienst}-${mitGrenze ? 'grenze' : 'frei'}` }
    const { hoechst, ergebnis } = await unterLast(pid, () => (dienst === 'mse-rest' ? profile['rest-rendern'](k).then((e) => {
      const proben = e.gruppen.REST
      // **Jedes 2xx ist gelungen**: ein Bild antwortet mit 201. Der erste
      // Lauf zählte nur 200 und nannte 11.361 Bilder „abgewiesen".
      const abgewiesen = proben.filter((x) => x.status < 200 || x.status > 299)
      return { gelungen: proben.length - abgewiesen.length, abgewiesen: abgewiesen.length, grund: abgewiesen[0] ? `${String(abgewiesen[0].status)} ${(abgewiesen[0].text ?? abgewiesen[0].fehler ?? '').slice(0, 160)}` : '' }
    }) : bogenlast(laufend, 4)))
    await schlafe(ruhe * 1000)
    const danach = fussabdruck([pid]).get(String(pid))

    return { dienst, mitGrenze, anfang, hoechst, danach, ...ergebnis }
  } finally {
    await beende(laufend)
  }
}

const zeilen = []
const gemessen = {}
const roh = []
for (const dienst of ['mse-rest', 'mse-mcp']) {
  for (const mitGrenze of [false, true]) {
    const f = await fall(dienst, mitGrenze)
    roh.push(f)
    // **Ohne gelungenen Ruf kein Wert**: gemessen wäre ein Dienst, der nur
    // abweist, und nicht einer unter Rasterlast — und der Beleg hieße
    // trotzdem „bestanden".
    gemessen[`Dienst nach Rendern und Ruhe, ${dienst}${mitGrenze ? ', mit GOMEMLIMIT' : ''} (MiB)`] = f.danach === undefined || f.gelungen === 0 ? [] : [f.danach]
    console.log(`${dienst}${mitGrenze ? ` mit GOMEMLIMIT=${grenze}` : ''}: ${String(f.gelungen)} gelungen, ${String(f.abgewiesen)} abgewiesen${f.grund ? ` (zuerst ${f.grund})` : ''}; Fußabdruck ${String(f.anfang)} → höchstens ${String(f.hoechst)} → nach ${String(ruhe)} s Ruhe ${String(f.danach)} MiB`)
  }
}

let rot = 0
for (const e of MESSVERTRAG.filter((x) => x.beleg === 'dienstspeicher.md')) {
  zeilen.push(tafel(e, gemessen[e.reihe] ?? []))
  if (urteil(e, gemessen[e.reihe] ?? []) !== 'gehalten') rot += 1
}
const commit = execFileSync('git', ['rev-parse', '--short=9', 'HEAD'], { cwd: wurzel, encoding: 'utf8' }).trim()
const ungesichert = execFileSync('git', ['status', '--porcelain'], { cwd: wurzel, encoding: 'utf8' }).split('\n').filter(Boolean).length
const maschine = `${execFileSync('sysctl', ['-n', 'hw.model'], { encoding: 'utf8' }).trim()}, ${cpus()[0].model}, ${String(cpus().length)} Kerne, ${String(Math.round(totalmem() / 2 ** 30))} GiB, macOS ${execFileSync('sw_vers', ['-productVersion'], { encoding: 'utf8' }).trim()}`
const beleg = `# Dienstspeicher nach Rendern und Ruhe

**${rot === 0 ? 'bestanden' : `${String(rot)} Befund(e)`}** — Commit \`${commit}\`${ungesichert ? ` **mit ${String(ungesichert)} ungesicherten Änderung(en)**` : ''},
${new Date().toISOString().slice(0, 16).replace('T', ' ')} UTC, ${maschine},
\`node scripts/dienstspeicher.mjs\` — Last ${String(dauer)} s, Ruhe ${String(ruhe)} s, \`GOMEMLIMIT=${grenze}\` im zweiten Fall; Last vor dem Lauf ${loadavg().map((z) => z.toFixed(2)).join(' / ')}.

| Reihe | n | min | Median | P95 | max | Grenze | Urteil |
| --- | --- | --- | --- | --- | --- | --- | --- |
${zeilen.join('\n')}

| Dienst | GOMEMLIMIT | gelungen | abgewiesen | Fußabdruck am Anfang | höchstens unter Last | nach der Ruhe |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
${roh.map((f) => `| ${f.dienst} | ${f.mitGrenze ? grenze : '—'} | ${String(f.gelungen)} | ${String(f.abgewiesen)} | ${String(f.anfang)} MiB | ${String(f.hoechst)} MiB | ${String(f.danach)} MiB |`).join('\n')}

**Was es nicht belegt**: eine zweite Maschine und einen Dienst, der Tage läuft. Gemessen ist der Fußabdruck nach \`footprint\`, nicht RSS.
`
writeFileSync(join(wurzel, process.env.BERICHT ?? 'docs/belege/dienstspeicher.md'), beleg)
console.log(rot === 0 ? '✓ Dienstspeicher: jede Grenze gehalten' : `✗ Dienstspeicher: ${String(rot)} Grenze(n) gerissen`)
process.exit(rot === 0 ? 0 : 1)
