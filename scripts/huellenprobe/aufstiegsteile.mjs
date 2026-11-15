// **Die Teile der Aufstiegsproben** — M01 und M02
// (#2345,
// #2346):
// 0.8.0 aus dem Tag, ein künstlicher Bestand und die Vergleiche. Kein
// Drehbuch: `aufstieg09` und `wanderung09` nehmen sie.
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { existsSync, mkdirSync, readdirSync, readFileSync, realpathSync, symlinkSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { deflateSync } from 'node:zlib'

export const wurzel = resolve(import.meta.dirname, '..', '..')

// crc32 für die Teile eines PNG.
const tafel = Array.from({ length: 256 }, (_, n) => {
  let c = n
  for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1
  return c >>> 0
})
const crc32 = (puffer) => {
  let c = 0xffffffff
  for (const b of puffer) c = tafel[(c ^ b) & 0xff] ^ (c >>> 8)
  return (c ^ 0xffffffff) >>> 0
}
const teil = (art, daten) => {
  const laenge = Buffer.alloc(4)
  laenge.writeUInt32BE(daten.length)
  const kopf = Buffer.concat([Buffer.from(art, 'ascii'), daten])
  const pruef = Buffer.alloc(4)
  pruef.writeUInt32BE(crc32(kopf))
  return Buffer.concat([laenge, kopf, pruef])
}

// bild gibt ein PNG von 8 × 8 Punkten in einer Farbe, die aus `n` folgt.
function bild(n) {
  const kopf = Buffer.alloc(13)
  kopf.writeUInt32BE(8, 0)
  kopf.writeUInt32BE(8, 4)
  kopf[8] = 8
  kopf[9] = 2
  const zeilen = []
  for (let y = 0; y < 8; y++) {
    zeilen.push(0)
    for (let x = 0; x < 8; x++) zeilen.push((n * 37) & 255, (n * 91 + x) & 255, (n * 13 + y) & 255)
  }
  return Buffer.concat([Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]), teil('IHDR', kopf),
    teil('IDAT', deflateSync(Buffer.from(zeilen))), teil('IEND', Buffer.alloc(0))])
}

// setErzeugen legt ein Set als Ordner an — dieselbe Gestalt wie die Saat.
export function setErzeugen(ordner, nummer, { karten = 500, bilder = 20 } = {}) {
  mkdirSync(ordner, { recursive: true })
  const zeilen = ['mse version: 2.0.0', 'game: probe', 'stylesheet: probe-kunst', 'set info:', `\ttitle: Aufstiegsset ${nummer}`, '\tcopyright: Probe']
  for (let k = 0; k < karten; k++) {
    zeilen.push('card:', `\tname: Karte ${nummer}-${k}`, `\tfarbe: ${k % 2 ? 'rot' : 'blau'}`, `\ttext: Probetext ${nummer}-${k} — ein Satz, der im Bestand wiederkommen muss.`)
    if (k < bilder) {
      writeFileSync(join(ordner, `bild${k}.png`), bild(nummer * 1000 + k))
      zeilen.push(`\tkunst: bild${k}.png`)
    }
  }
  writeFileSync(join(ordner, 'set'), `${zeilen.join('\n')}\n`)
}

// fassung080 baut Kommandozeile und Abnahmebündel aus dem Tag v0.8.0.
export function fassung080(arbeit, { buendel: mitBuendel = true } = {}) {
  const baum = join(arbeit, 'v080')
  execFileSync('git', ['worktree', 'add', '--detach', baum, 'v0.8.0'], { cwd: wurzel, stdio: 'ignore' })
  const entfernen = () => execFileSync('git', ['worktree', 'remove', '--force', baum], { cwd: wurzel, stdio: 'ignore' })
  try {
    mkdirSync(join(baum, '.go'))
    for (const teilort of ['mod', 'build', 'bin']) symlinkSync(realpathSync(join(wurzel, '.go', teilort)), join(baum, '.go', teilort))
    symlinkSync(realpathSync(join(wurzel, 'frontend/node_modules')), join(baum, 'frontend/node_modules'))
    const umgebung = { ...process.env, GOPROXY: 'off' }
    const mse = join(arbeit, 'mse080')
    execFileSync('go', ['build', '-o', mse, './cmd/mse'], { cwd: baum, env: umgebung, stdio: 'ignore' })
    if (mitBuendel) execFileSync('./scripts/bauen.sh', ['abnahme'], { cwd: baum, env: umgebung, stdio: 'ignore' })
    const buendel = join(baum, 'build/bin/MagicSetEditor4Mac.app')
    const abdruck = mitBuendel ? createHash('sha256').update(readFileSync(join(buendel, 'Contents/MacOS/MagicSetEditor4Mac'))).digest('hex').slice(0, 12) : '—'
    const stand = execFileSync('git', ['rev-parse', '--short', 'HEAD'], { cwd: baum }).toString().trim()
    return { mse, buendel, abdruck, stand, entfernen }
  } catch (fehler) {
    entfernen()
    throw fehler
  }
}

// **Ohne `-readonly`**: der Bestand läuft im WAL-Modus, und nach einem
// sauberen Schließen fehlt `-shm` — lesend darf SQLite es nicht anlegen und
// öffnet gar nicht (gemessen im ersten Lauf). Gefragt wird nur gelesen.
export const sqlite = (datei, sql) => execFileSync('sqlite3', [datei, sql], { maxBuffer: 1 << 30 }).toString().trim()

// inhalt gibt je Tafel einen Abdruck ihrer Zeilen — **ohne die Buchführung**: `migrationen`, `sicherungsvermerke` (0098) und
// `set_aenderungen` (0099, PERF02) zählen, was geschah, und nicht, was im
// Bestand steht — ein Zähler, der beim Aufstieg bei null beginnt, ist kein
// verlorener Inhalt. Ob jedes Set seinen Zähler hat, prüft [zaehlerJeSet].
//
// **Nur Tafeln mit Zeilen** (seit 0101, #2500): eine Tafel, die eine
// Wanderung neu und leer anlegt, ist kein Unterschied — und eine, die beim
// Aufstieg leer würde, fehlt danach im Abdruck und ist damit einer.
export function inhalt(datei) {
  const tafeln = sqlite(datei, "SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT IN ('migrationen', 'sicherungsvermerke', 'set_aenderungen', 'sqlite_sequence') ORDER BY name").split('\n')
  // **Als JSON und selbst sortiert**: manche Tafeln führen keine `rowid`.
  const zeilen = (name) => JSON.parse(execFileSync('sqlite3', ['-json', datei, `SELECT * FROM "${name}"`], { maxBuffer: 1 << 30 }).toString() || '[]').map((z) => JSON.stringify(z)).sort()
  return Object.fromEntries(tafeln.map((name) => [name, zeilen(name)]).filter(([, z]) => z.length > 0)
    .map(([name, z]) => [name, createHash('sha256').update(z.join('\n')).digest('hex').slice(0, 16)]))
}

// zaehlerJeSet sagt, ob jedes Set genau einen Änderungszähler hat (0099).
export const zaehlerJeSet = (datei) => sqlite(datei, 'SELECT (SELECT COUNT(*) FROM sets) = (SELECT COUNT(*) FROM set_aenderungen JOIN sets ON sets.id = set_aenderungen.set_id)') === '1'

export const unterschied = (a, b) => Object.keys({ ...a, ...b }).filter((k) => a[k] !== b[k])
export const summe = (datei) => createHash('sha256').update(readFileSync(datei)).digest('hex')

// bilder folgt jedem Bildverweis des Bestands zu seiner Datei — **ein
// eingelesenes Set liest seine Bilder aus seiner Quelle** (`sets.herkunft`,
// `internal/app/dienstablage.go`), es kopiert sie nicht. Erhalten heißt: jeder
// Verweis zeigt auf dieselbe vorhandene Datei.
export function bilder(datei) {
  const verweise = sqlite(datei, "SELECT s.herkunft || '/' || v.wert FROM card_values v JOIN cards c ON c.id = v.card_id JOIN sets s ON s.id = c.set_id WHERE v.feld = 'kunst' ORDER BY 1").split('\n').filter(Boolean)
  const abdruck = createHash('sha256')
  let zahl = 0
  for (const pfad of verweise) {
    abdruck.update(pfad)
    if (existsSync(pfad)) {
      abdruck.update(readFileSync(pfad))
      zahl++
    }
  }
  return { verweise: verweise.length, zahl, abdruck: abdruck.digest('hex').slice(0, 16) }
}

export const buchfuehrung = (datei) => sqlite(datei, "SELECT name || ' ' || pruefsumme FROM migrationen ORDER BY name")
// wanderungenNachDerGrundlinie nennt die Wanderungen, die dieser Stand über
// die 97 der Grundlinie hinaus mitbringt — **aus dem Ordner gelesen** und
// nicht gezählt: mit `0099_setstand.sql` (PERF02) hätte eine feste Zahl jedes
// Drehbuch rot gemacht, das sie nennt.
export const wanderungenNachDerGrundlinie = () => readdirSync(resolve(import.meta.dirname, '../../internal/storage/migrationen'))
  .filter((n) => /^\d{4}_.+\.sql$/.test(n)).sort()

export const sicherungen = (ort) => readdirSync(ort.ort).filter((n) => /^bestand-vor-.*\.db$/.test(n)).sort()
