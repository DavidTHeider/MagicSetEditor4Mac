// **Prozesse messen** — die zwei Griffe, die der Läufer den Teilnehmern gibt
// (QA-04, #1735).
//
// `messe` fährt einen ganzen Prozess, `starte` hält einen für eine Sitzung
// offen. Beide laufen unter `/usr/bin/time -l` — das ist die einzige Quelle
// für CPU-Zeit und Speicherspitze eines Kindprozesses, die macOS ohne
// Verwalterrechte gibt. Die Wandzeit misst der Läufer selbst.
//
// **Den Prozess erzeugt, wer `prozesse` ruft** — der Läufer reicht
// `child_process.spawn` herein. Ein Helfer, der selbst Prozesse startet, sähe
// für `internal/werkzeuge/nahtlauf_test.go` wie ein Lauf aus, den
// `docs/RELEASE.md` nennen müsste.

import { execFileSync } from 'node:child_process'
import { loadavg } from 'node:os'
import { performance } from 'node:perf_hooks'

import { zeitLesen } from './leistungsstatistik.mjs'

/** auswerten fasst Ende, Zeitangabe und Ausgaben eines Prozesses zusammen. */
function auswerten({ anfang, rueckgabe, signal, aus, fehler, lastVorher }) {
  const zeit = zeitLesen(fehler)
  const absturz = signal !== null || zeit === null || zeit.abgebrochen
  return {
    wandMs: performance.now() - anfang,
    cpuS: zeit ? Math.round((zeit.nutzer + zeit.system) * 100) / 100 : null,
    rssByte: zeit?.rss ?? null,
    spitzeByte: zeit?.spitze ?? null,
    rueckgabe, signal, absturz,
    fehler: absturz ? `abgebrochen (${signal ?? 'ohne Zeitangabe'})` : rueckgabe !== 0 ? `Rückgabe ${rueckgabe}` : null,
    last: lastVorher,
    aus,
    // **Nur das Ende der Fehlerausgabe** — davor steht Protokoll, und ein
    // Beleg mit Megabyte an Protokoll liest niemand.
    fehlerText: fehler.slice(-1500),
  }
}

/** prozesse gibt `messe` und `starte` über der Erzeugung `erzeuge` (wie `spawn`). */
export function prozesse(erzeuge) {
  return { messe: (...a) => messe(erzeuge, ...a), starte: (...a) => starte(erzeuge, ...a) }
}

/** messe fährt einen Befehl bis zum Ende — Eingabe auf stdin, Ausgaben gesammelt. */
function messe(erzeuge, befehl, argumente, { umgebung, ort, eingabe } = {}) {
  return new Promise((fertig) => {
    const lastVorher = loadavg()[0]
    const anfang = performance.now()
    const kind = erzeuge('/usr/bin/time', ['-l', befehl, ...argumente], { env: umgebung, cwd: ort })
    let aus = ''
    let fehler = ''
    kind.stdout.on('data', (d) => (aus += d))
    kind.stderr.on('data', (d) => (fehler += d))
    kind.on('close', (rueckgabe, signal) => fertig(auswerten({ anfang, rueckgabe, signal, aus, fehler, lastVorher })))
    kind.stdin.on('error', () => {})
    kind.stdin.end(eingabe ?? '')
  })
}

/**
 * starte hält einen Prozess für eine Sitzung offen. `schliessen` beendet die
 * Eingabe und wartet auf das Ende; `pid` gibt die Kennung des **Programms** —
 * nicht die von `time` —, damit der Speicher im Dauerbetrieb am richtigen
 * Prozess gelesen wird.
 */
function starte(erzeuge, befehl, argumente, { umgebung, ort } = {}) {
  const lastVorher = loadavg()[0]
  const anfang = performance.now()
  const kind = erzeuge('/usr/bin/time', ['-l', befehl, ...argumente], { env: umgebung, cwd: ort })
  let fehler = ''
  kind.stderr.on('data', (d) => (fehler += d))
  kind.stdin.on('error', () => {})
  const ende = new Promise((fertig) => kind.on('close', (rueckgabe, signal) => fertig({ rueckgabe, signal })))
  return {
    kind,
    pid() {
      try {
        return Number(execFileSync('pgrep', ['-P', String(kind.pid)], { encoding: 'utf8' }).trim().split('\n')[0])
      } catch {
        return null
      }
    },
    async schliessen() {
      kind.stdin.end()
      const { rueckgabe, signal } = await ende
      return auswerten({ anfang, rueckgabe, signal, aus: '', fehler, lastVorher })
    },
  }
}

/** residentByte liest den residenten Stand eines Prozesses — `null`, wenn er weg ist. */
export function residentByte(pid) {
  try {
    return Number(execFileSync('ps', ['-o', 'rss=', '-p', String(pid)], { encoding: 'utf8' }).trim()) * 1024
  } catch {
    return null
  }
}
