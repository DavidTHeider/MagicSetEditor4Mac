// **Aus welchem Stand ein Bündel gebaut ist, und was ein Lauf belegt** — T06
// (#2099).
//
// # Der Bauabdruck
//
// Ein SHA-256 über den Inhalt der Baueingänge an HEAD, den Unterschied des
// Baums dazu und jede noch nicht versionierte Datei — **nur über die
// Baueingänge**. `bauen.sh abnahme` legt ihn ins Bündel (`Contents/Resources/baustand.json`), und die
// Hüllenprobe hält ihn gegen den Baum, bevor sie misst. Bis zum 2026-10-06
// nannte ein Beleg nur HEAD und eine Zahl offener Dateien; zwei Läufe am
// selben HEAD konnten verschiedene Bündel messen, und niemand sah es.
//
// # Was er nicht sieht
//
// **Was außerhalb der Baueingänge liegt** — Doku, Belege, die Drehbücher
// selbst. Eine Änderung an einem Drehbuch braucht keinen neuen Bau, und eine
// Zeile in `docs/` ändert kein Bündel. Und **was außerhalb des Baums liegt**:
// Werkzeugfassungen, `node_modules`, der Modulzwischenspeicher. Die stehen im
// Manifest des Laufs, nicht im Abdruck.
import { createHash } from 'node:crypto'
import { execFileSync } from 'node:child_process'
import { readdirSync, readFileSync, rmSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

// baueingaenge sind die Pfade, aus denen `scripts/bauen.sh` ein Bündel macht.
export const baueingaenge = [
  'main.go', 'go.mod', 'go.sum', 'internal', 'cmd',
  'frontend/src', 'frontend/public', 'frontend/index.html', 'frontend/package.json',
  'frontend/package-lock.json', 'frontend/vite.config.ts', 'frontend/tsconfig.json',
  'saat', 'build', 'wails.json', 'scripts/bauen.sh', 'scripts/saat.sh', 'scripts/macos.sh',
]

const git = (wurzel, ...argumente) => execFileSync('git', ['-C', wurzel, ...argumente], { maxBuffer: 256 * 1024 * 1024 })

// bauabdruck ist der Abdruck der Baueingänge im Baum unter `wurzel`.
export function bauabdruck(wurzel) {
  const summe = createHash('sha256')
  // **Der Inhalt der Eingänge an HEAD, nicht HEAD selbst** — sonst hieße ein
  // Bündel nach jedem Doku-Commit veraltet, obwohl aus demselben Stand gebaut.
  summe.update(git(wurzel, 'ls-tree', '-r', 'HEAD', '--', ...baueingaenge))
  summe.update(git(wurzel, 'diff', 'HEAD', '--binary', '--', ...baueingaenge))
  const neu = git(wurzel, 'ls-files', '--others', '--exclude-standard', '-z', '--', ...baueingaenge)
    .toString().split('\0').filter(Boolean).sort()
  for (const pfad of neu) {
    summe.update(`${pfad}\0`)
    summe.update(readFileSync(join(wurzel, pfad)))
  }

  return summe.digest('hex')
}

// dateiabdruck ist der SHA-256 einer Datei — oder `null`, wenn es sie nicht gibt.
export function dateiabdruck(pfad) {
  try {
    return createHash('sha256').update(readFileSync(pfad)).digest('hex')
  } catch {
    return null
  }
}

// verzeichnisabdruck ist der SHA-256 über Namen und Inhalt jeder Datei
// unter `ort` — die gebaute Oberfläche etwa. `null`, wenn es `ort` nicht gibt.
export function verzeichnisabdruck(ort) {
  let namen
  try {
    namen = readdirSync(ort, { recursive: true }).map(String).sort()
  } catch {
    return null
  }
  const summe = createHash('sha256')
  for (const name of namen) {
    const pfad = join(ort, name)
    if (!statSync(pfad).isFile()) continue
    summe.update(`${name}\0`)
    summe.update(readFileSync(pfad))
  }

  return summe.digest('hex')
}

// laufkennung ist der Name eines Laufs: Zeit in UTC und Prozess — sortiert
// sich wie die Zeit.
export const laufkennung = () => `${new Date().toISOString().slice(0, 19).replaceAll(':', '-')}Z-${String(process.pid)}`

// laeufeRaeumen lässt unter `ort` die jüngsten `behalten` Läufe stehen und
// gibt die Namen der geräumten. **Ein Lauf hinterlässt Bilder**, und sie
// sollen den Platz nicht stillschweigend füllen.
export function laeufeRaeumen(ort, behalten) {
  let namen
  try {
    namen = readdirSync(ort).filter((n) => statSync(join(ort, n)).isDirectory()).sort()
  } catch {
    return []
  }
  const weg = namen.slice(0, Math.max(0, namen.length - behalten))
  for (const name of weg) rmSync(join(ort, name), { recursive: true, force: true })

  return weg
}

// Als Programm für `scripts/bauen.sh`: `node scripts/baustand.mjs abdruck`.
if (process.argv[1] === fileURLToPath(import.meta.url)) {
  if (process.argv[2] !== 'abdruck') {
    console.error('Aufruf: node scripts/baustand.mjs abdruck')
    process.exit(2)
  }
  process.stdout.write(bauabdruck(fileURLToPath(new URL('..', import.meta.url))))
}
