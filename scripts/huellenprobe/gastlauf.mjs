// **Was jeder Gastlauf braucht** — aus `untergrenzegast.mjs` herausgelöst, als
// der Aufstieg (I03, #2261)
// dieselben Griffe brauchte: das Abnahmebündel dieses Stands, ein geteiltes
// Verzeichnis, die VM ohne Anzeige, `lume ssh`, und am Ende `TERM` (`lume
// stop` hing). Ohne VM fährt derselbe Treiber am Wirt — die Positivkontrolle.
import { execFileSync, spawn, spawnSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { basename, join, resolve } from 'node:path'

import { bauabdruck, dateiabdruck } from '../baustand.mjs'

export const wurzel = resolve(import.meta.dirname, '../..')
export const buendel = join(wurzel, 'build/bin/MagicSetEditor4Mac.app')
const lumeUmgebung = { ...process.env, LUME_TELEMETRY_ENABLED: 'false', LUME_UPDATE_CHECK: 'false' }
const lume = (...argumente) => execFileSync('lume', argumente, { env: lumeUmgebung, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] })

// gastziel liest aus dem Aufruf, wo gefahren wird — **ausdrücklich**: ein
// Gast mit einem Namen, der einer ist, oder `--wirt`, wo der Läufer es
// erlaubt. Befund 1 des Zweitblicks über Phase I: ein leerer Name überstand
// `??`, und `fahre()` las »keine VM« als »am Wirt«.
export function gastziel(eingabe, { wirtErlaubt }) {
  if (eingabe === undefined) return { vm: 'mse080-152', amWirt: false }
  if (eingabe === '--wirt') {
    if (!wirtErlaubt) throw new Error('Dieser Lauf läuft nie am Wirt — er richtet ein, was nur in einen Wegwerf-Gast gehört.')

    return { vm: null, amWirt: true }
  }
  if (!/^[A-Za-z0-9][A-Za-z0-9._-]*$/.test(eingabe)) throw new Error(`»${eingabe}« ist kein Gast`)

  return { vm: eingabe, amWirt: false }
}

// abnahmestand gibt den Baustand — oder endet, wenn das Bündel nicht der
// Abnahmebau dieses Stands ist oder die VM schon läuft.
export function abnahmestand(vm) {
  const stand = JSON.parse(readFileSync(join(buendel, 'Contents/Resources/baustand.json'), 'utf8'))
  if (stand.art !== 'abnahme' || stand.abdruck !== bauabdruck(wurzel)) {
    console.error('Das Bündel ist nicht der Abnahmebau dieses Stands — ./scripts/bauen.sh abnahme')
    process.exit(2)
  }
  if (vm && /running/.test(lume('ls').split('\n').find((z) => z.startsWith(`${vm} `)) ?? '')) {
    console.error(`Die VM ${vm} läuft schon — sie startet hier mit dem geteilten Verzeichnis dieses Laufs`)
    process.exit(2)
  }

  return stand
}

// imGeteilten legt ein geteiltes Verzeichnis an, gibt es `tun` und nimmt es
// **in jedem Fall** wieder weg — nach Erfolg, nach einer Ausnahme, auch in
// der Vorbereitung, und nach `SIGINT` oder `SIGTERM`. Befund 1 des Nachlaufs
// über Phase I: die Läufer räumten erst ab dem Gastlauf auf, ein
// gescheiterter Compileraufruf davor ließ das Verzeichnis samt Bündel liegen,
// und ein Signal übersprang jedes `finally` — ebenso ein `process.exit` im
// Läufer, deshalb auch beim `exit`.
export async function imGeteilten(name, tun) {
  const geteilt = mkdtempSync(join(tmpdir(), `${name}-gast-`))
  const weg = () => rmSync(geteilt, { recursive: true, force: true })
  const signal = (art) => {
    weg()
    process.exit(art === 'SIGINT' ? 130 : 143)
  }
  process.once('SIGINT', signal)
  process.once('SIGTERM', signal)
  process.once('exit', weg)
  try {
    return await tun(geteilt)
  } finally {
    process.off('SIGINT', signal)
    process.off('SIGTERM', signal)
    process.off('exit', weg)
    weg()
  }
}

// fuelleGeteilt legt in ein geteiltes Verzeichnis, was der Treiber braucht:
// das Bündel, die Saat, den Treiber und `Gast.pm`.
export function fuelleGeteilt(geteilt, name, treiber) {
  const ordner = join(geteilt, name)
  mkdirSync(ordner, { recursive: true })
  execFileSync('cp', ['-cR', buendel, join(ordner, 'Abnahme.app')])
  execFileSync('cp', ['-cR', join(wurzel, 'saat/probe-werkstatt.mse-set'), join(ordner, 'probe-werkstatt.mse-set')])
  for (const datei of [treiber, 'Gast.pm']) execFileSync('cp', ['-c', join(wurzel, 'scripts/huellenprobe/gast', datei), join(ordner, datei)])

  return ordner
}

// fahre fährt den Treiber im Gast oder — nur auf ausdrückliches `amWirt` —
// am Wirt, mit `wirtUmgebung` für den Treiber.
export async function fahre({ vm, geteilt, name, treiber, amWirt, wirtUmgebung }) {
  if (amWirt === true) {
    const lauf = spawnSync('perl', [join(geteilt, name, treiber), geteilt], { encoding: 'utf8', env: { ...process.env, ...wirtUmgebung } })

    return { rueckgabe: lauf.status ?? 1, ausgabe: `${lauf.stdout}${lauf.stderr}` }
  }
  if (!vm) throw new Error('weder Gast noch Wirt — fahre() braucht eine VM oder ausdrücklich amWirt')
  const starte = () => spawn('lume', ['run', vm, '--display', process.env.MSE_GAST_ANZEIGE ?? 'none', '--shared-dir', geteilt], { env: lumeUmgebung, stdio: 'ignore', detached: true })
  let lauf = starte()
  let versuche = 1
  try {
    const bis = Date.now() + 300_000
    for (;;) {
      try {
        lume('ssh', vm, '-t', '30', 'true')
        break
      } catch (grund) {
        if (Date.now() > bis) throw new Error(`die VM antwortet nicht: ${String(grund.message).slice(0, 200)}`)
        // **Endete `lume run` vor der ersten Antwort, startet er neu** —
        // höchstens dreimal. Kurz nach einem Stopp endete er still, und
        // `fahre` wartete fünf Minuten auf eine VM, die nie lief (R02, am
        // 2026-10-15 drei von sieben Starts; kein `lume run` lief da mehr).
        if ((lauf.exitCode !== null || lauf.signalCode !== null) && versuche < 3) {
          await new Promise((weiter) => setTimeout(weiter, 10_000))
          lauf = starte()
          versuche++
        }
        await new Promise((weiter) => setTimeout(weiter, 5000))
      }
    }
    const imGast = `/Volumes/My Shared Files/${basename(geteilt)}`
    try {
      return { rueckgabe: 0, ausgabe: lume('ssh', vm, '-t', '3600', `perl '${imGast}/${name}/${treiber}' '${imGast}'`) }
    } catch (grund) {
      return { rueckgabe: grund.status ?? 1, ausgabe: `${grund.stdout ?? ''}${grund.stderr ?? ''}` }
    }
  } finally {
    // **Erst festschreiben, dann stoppen.** SIGTERM an `lume run` beendet die
    // VM hart, und was der Treiber zuletzt schrieb — sein Aufräumen — war nach
    // dem nächsten Start weg (gemessen am 2026-10-11: eine Datei ohne `sync`
    // verloren, mit `sync` erhalten). Die Pause gibt `cfprefsd` Zeit, die
    // Suchliste der Schlüsselbünde zu schreiben. Antwortet die VM nicht mehr,
    // gibt es nichts festzuschreiben.
    try {
      lume('ssh', vm, '-t', '30', 'sync; sleep 1; sync')
    } catch {
      // die VM antwortet nicht
    }
    // **Nur auf ein Ende warten, das noch kommt** — ist `lume run` schon
    // beendet, fiele das `exit` nie mehr, und es bliebe die volle Frist.
    // **Und die Frist geht mit dem Ende weg** — ein stehender Zeitgeber hielt
    // den Läufer nach jedem Gastlauf 30 s länger am Leben.
    if (lauf.exitCode === null && lauf.signalCode === null) {
      await new Promise((weiter) => {
        const frist = setTimeout(weiter, 30_000)
        lauf.once('exit', () => {
          clearTimeout(frist)
          weiter()
        })
        lauf.kill('SIGTERM')
      })
    }
  }
}

// kopfzeilen sind Commit, Maschine und Bündel eines Belegs.
export function kopfzeilen({ beginn, befehl, rueckgabe, wo, maschine, stand }) {
  const kopf = execFileSync('git', ['-C', wurzel, 'rev-parse', '--short', 'HEAD']).toString().trim()
  const programm = join(buendel, 'Contents/MacOS/MagicSetEditor4Mac')
  const minos = execFileSync('vtool', ['-show-build', programm]).toString().match(/minos (\S+)/)?.[1] ?? '?'

  return [
    `**Stand:** Commit \`${kopf}\` · **Datum:** ${beginn.toISOString().slice(0, 16).replace('T', ' ')} UTC · **Befehl:** \`${befehl}\` · **Rückgabe:** ${String(rueckgabe)}`,
    '',
    `**Maschine:** ${wo} — macOS ${maschine.system} (${maschine.bau}), ${maschine.modell}, ${maschine.kerne} Kerne${maschine.speicher ? `, ${String(maschine.speicher)} MB` : ''}${maschine.last ? `; Last ${maschine.last}, beim Messen ${maschine.lastBeimMessen ?? '?'}` : ''}.`,
    '',
    `**Bündel:** Abnahmebau, Baustand \`${stand.abdruck.slice(0, 12)}\`, Programm \`${dateiabdruck(programm).slice(0, 16)}…\`, gebaut mit \`minos ${minos}\`.`,
  ]
}

// schrittzeilen sind die Tafel der Schritte.
export function schrittzeilen(schritte, spalte) {
  return [
    `| ${spalte} | Schritt | Ergebnis | ms | Wert oder Fehler |`,
    '| --- | --- | --- | ---: | --- |',
    ...schritte.map((s) => `| ${s.flaeche ?? s.teil} | ${s.was} | ${s.ok ? 'gelungen' : '**gescheitert**'} | ${String(s.ms)} | ${(s.fehler ?? s.wert ?? '').replaceAll('|', '\\|').replaceAll('\n', ' ')} |`),
  ]
}
