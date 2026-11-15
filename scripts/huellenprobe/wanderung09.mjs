// **Abbruch mitten in der Wanderung und volle Platte** — M02
// (#2346),
// Befunde A-11 und A-12.
//
// Die Abnahme: `SIGKILL` mitten in der Wanderung und eine volle Platte lassen
// den Originalstand oder den vollständig neuen zurück, nie eine halbe
// Wanderung — Tafeln, Buchführung und Bildverweise verglichen. **Gleichzeitige
// Starter** belegt `aufstieg09`, den Dienst vor einem neueren Bestand
// `internal/flaechen` (`neuererbestand_test.go`).
//
// # Wie der Abbruch die Wanderung trifft
//
// Ein `mse` aus diesem Baum, gebaut **mit `-tags abnahme`**: dort hält
// `MSE_ABNAHME_WANDERUNGSHALT` die Wanderung **nach** ihren Anweisungen und
// **vor** dem Festschreiben an und legt eine Marke
// (`internal/storage/wanderungshalt_abnahme.go`). Das Drehbuch wartet auf
// die Marke und beendet den Prozess dann hart. Der Bestand kommt von 0.8.0
// aus dem Tag — 10 Sets mit je 500 Karten.
//
// # Wie die Platte voll wird
//
// Ein eigenes Datenträgerabbild (`hdiutil`, 64 MB, nur für diesen Lauf), darauf
// der Bestand von 0.8.0 und eine Füllung, die weniger Platz lässt, als die
// Sicherung vor der Wanderung braucht. Danach wird es wieder ausgehängt.
//
// # Was es nicht belegt
//
// Einen laufenden Dienst unter launchd — dafür braucht es einen Gast
// (offen in #2346), und den Aufstieg durch das Fenster.
import { execFileSync, spawn } from 'node:child_process'
import { copyFileSync, existsSync, mkdirSync, readdirSync, rmSync, statSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'
import { bilder, buchfuehrung, fassung080, inhalt, setErzeugen, sicherungen, sqlite, summe, unterschied, wanderungenNachDerGrundlinie, wurzel } from './aufstiegsteile.mjs'

const SETS = 10

export default async function wanderung09(p) {
  const ort = p.ort()
  const arbeit = join(ort.ort, 'arbeit')
  mkdirSync(arbeit)

  p.notiz('0.8.0 aus dem Tag legt den Bestand an')
  const alt = fassung080(arbeit, { buendel: false })
  try {
    const altRuf = (...a) => execFileSync(alt.mse, a, { cwd: wurzel, env: kindumgebung(ort), stdio: ['ignore', 'pipe', 'pipe'] }).toString()
    altRuf('paket-installieren', 'saat/probe.mse-game')
    altRuf('paket-installieren', 'saat/probe-kunst.mse-style')
    for (let s = 1; s <= SETS; s++) {
      const ordner = join(arbeit, `wanderung-${s}.mse-set`)
      setErzeugen(ordner, s)
      altRuf('oeffnen', ordner)
    }
  } finally {
    alt.entfernen()
  }
  execFileSync('sqlite3', [ort.bestand, 'PRAGMA wal_checkpoint(TRUNCATE);'], { stdio: 'ignore' })
  const stand080 = join(arbeit, 'stand-0.8.0.db')
  copyFileSync(ort.bestand, stand080)
  const vorher = inhalt(stand080)
  const bilderVorher = bilder(stand080)
  p.pruefe('0.8.0 führt 97 Wanderungen und 5.000 Karten',
    buchfuehrung(stand080).split('\n').length === 97 && sqlite(stand080, 'SELECT count(*) FROM cards') === '5000')

  p.notiz('mse mit dem Halt des Abnahmebaus')
  const abnahme = join(arbeit, 'mse-abnahme')
  execFileSync('go', ['build', '-tags', 'abnahme', '-o', abnahme, './cmd/mse'], { cwd: wurzel, env: { ...process.env, GOPROXY: 'off' }, stdio: 'ignore' })

  await abbruch(p, abnahme, stand080, vorher, bilderVorher)
  vollePlatte(p, stand080, vorher)
}

// abbruch beendet `mse` in der offenen Transaktion der Wanderung.
async function abbruch(p, abnahme, stand080, vorher, bilderVorher) {
  p.notiz('SIGKILL mitten in der Wanderung')
  const ort = p.ort()
  copyFileSync(stand080, ort.bestand)
  const marke = join(ort.ort, 'wanderungshalt')
  const kind = spawn(abnahme, ['sets'], {
    cwd: wurzel, stdio: 'ignore',
    env: { ...kindumgebung(ort), MSE_ABNAHME_WANDERUNGSHALT: '60s', MSE_ABNAHME_WANDERUNGSMARKE: marke },
  })
  const beendet = new Promise((los) => kind.once('exit', (code, signal) => los({ code, signal })))
  await p.warte('die Wanderung hält vor dem Festschreiben', async () => existsSync(marke), 30000)
  p.pruefe('vor dem Halt entstand die Sicherung', sicherungen(ort).length === 1, sicherungen(ort).join(', '))
  kind.kill('SIGKILL')
  const ende = await beendet
  p.pruefe('der Prozess endete am Signal', ende.signal === 'SIGKILL', JSON.stringify(ende))

  p.pruefe('der Bestand ist heil', sqlite(ort.bestand, 'PRAGMA quick_check') === 'ok')
  p.pruefe('er führt die 97 Wanderungen von vorher — keine halbe', buchfuehrung(ort.bestand).split('\n').length === 97)
  const anders = unterschied(vorher, inhalt(ort.bestand))
  p.pruefe('jede Tafel trägt den Inhalt von vorher', anders.length === 0, anders.join(', '))
  p.pruefe('die neue Tafel gibt es nicht', sqlite(ort.bestand, "SELECT count(*) FROM sqlite_master WHERE name = 'sicherungsvermerke'") === '0')

  p.notiz('Der nächste Start wandert zu Ende')
  p.mse(ort, 'sets')
  const nachher = buchfuehrung(ort.bestand).split('\n')
  const neue = wanderungenNachDerGrundlinie()
  p.pruefe(`jetzt ${String(97 + neue.length)} Wanderungen`, nachher.length === 97 + neue.length && nachher.at(-1).startsWith(neue.at(-1)), nachher.at(-1))
  p.pruefe('mit demselben Inhalt', unterschied(vorher, inhalt(ort.bestand)).length === 0)
  p.pruefe('und denselben Bildern', bilder(ort.bestand).abdruck === bilderVorher.abdruck)
  p.pruefe('die Sicherung von vorher wurde wieder genommen, keine zweite', sicherungen(ort).length === 1, sicherungen(ort).join(', '))
}

// vollePlatte lässt die Sicherung vor der Wanderung am Platz scheitern.
function vollePlatte(p, stand080, vorher) {
  p.notiz('Die Platte ist voll')
  const ort = p.ort()
  const abbild = join(ort.ort, 'voll.dmg')
  const halt = join(ort.ort, 'voll')
  mkdirSync(halt)
  execFileSync('hdiutil', ['create', '-size', '64m', '-fs', 'HFS+', '-volname', 'mse-voll', abbild], { stdio: 'ignore' })
  execFileSync('hdiutil', ['attach', '-nobrowse', '-mountpoint', halt, abbild], { stdio: 'ignore' })
  try {
    const voll = { ...ort, bestand: join(halt, 'bestand.db') }
    copyFileSync(stand080, voll.bestand)
    const groesse = statSync(voll.bestand).size
    const frei = () => Number(execFileSync('df', ['-k', halt]).toString().trim().split('\n').at(-1).split(/\s+/)[3]) * 1024
    // **Weniger Platz, als die Sicherung braucht** — sie ist so groß wie
    // der Bestand.
    const fuellung = join(halt, 'fuellung')
    writeFileSync(fuellung, Buffer.alloc(Math.max(frei() - Math.floor(groesse / 2), 0)))
    p.zeige(`Bestand ${Math.round(groesse / 1024)} KB, frei ${Math.round(frei() / 1024)} KB`)
    const summeVorher = summe(voll.bestand)

    let satz = ''
    try {
      p.mse(voll, 'sets')
    } catch (fehler) {
      satz = `${fehler.stderr}`.trim()
    }
    p.zeige(satz.split('\n')[0] ?? '')
    // **Beides, seit #2407**: SQLite meldet den vollen Datenträger hier als
    // Fehler des Datenträgers, ohne `ENOSPC` in der Kette.
    p.pruefe('der Start scheitert und sagt, dass kaum Platz frei ist und der Datenträger zu prüfen', /kaum Platz frei/.test(satz) && /schaffe Platz/.test(satz) && /prüfe den Datenträger/.test(satz), satz.split('\n')[0])
    p.pruefe('ohne den Text des Treibers', !/sqlite3:|disk I\/O/.test(satz))
    p.pruefe('der Bestand ist bitgleich', summe(voll.bestand) === summeVorher)
    // **Auch keine halbe**: sie entsteht unter einem verborgenen Namen.
    const reste = readdirSync(halt).filter((n) => n.includes('bestand-vor-'))
    p.pruefe('keine Sicherung bleibt liegen, auch keine halbe', reste.length === 0, reste.join(', '))

    p.notiz('Platz frei, und noch einmal')
    rmSync(fuellung)
    p.mse(voll, 'sets')
    const nachher = buchfuehrung(voll.bestand).split('\n')
    p.pruefe(`jetzt ${String(97 + wanderungenNachDerGrundlinie().length)} Wanderungen`, nachher.length === 97 + wanderungenNachDerGrundlinie().length, nachher.at(-1))
    p.pruefe('mit demselben Inhalt', unterschied(vorher, inhalt(voll.bestand)).length === 0)
    p.pruefe('und genau einer Sicherung davor', sicherungen({ ort: halt }).length === 1)
  } finally {
    for (let versuch = 0; versuch < 3; versuch++) {
      try {
        execFileSync('hdiutil', ['detach', halt], { stdio: 'ignore' })
        break
      } catch {
        execFileSync('sleep', ['1'])
      }
    }
  }
}
