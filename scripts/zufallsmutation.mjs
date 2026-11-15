#!/usr/bin/env node
// **Zufällige Mutation der Kernregeln** — B05
// (#2224).
//
//     node scripts/zufallsmutation.mjs [--samen 2224] [--je 40] [--nur <stelle,...>]
//
// # Was er fragt
//
// **Wie viele zufällig gezogene Fehler fangen die Tests?** Anders als
// `mutationslauf.mjs`, das je eine **gewählte** Wache abschaltet, zieht er mit
// festem Samen aus **allen** Stellen von `domain`, `service` und `storage`
// (`scripts/mutationsorte.go`) — geschichtet, je Paket gleich viele — und
// fährt je Mutant die Tests **seines** Pakets.
//
// # Wie gezählt wird
//
// - **gefangen**: der Testlauf wird rot,
// - **überlebt**: er bleibt grün — dann ist der Mutant entweder
//   **äquivalent** (er ändert kein Verhalten) oder eine **Lücke**; das sagt
//   die Durchsicht in `docs/belege/zufallsmutation-durchsicht.json`, nicht
//   der Lauf,
// - **nicht übersetzt** und **Frist** stehen getrennt und zählen nicht.
//
// Die Quote ist gefangen ÷ (gefangen + Lücken), mit Wilson-Intervall (95 %).
// **Keine Hochrechnung aufs Gesamtprojekt**: sie gilt für die drei Pakete
// und die fünf Arten aus `mutationsorte.go`.
//
// # Was er nicht tut
//
// **Er fasst den Arbeitsbaum nicht an**: gefahren wird in einem eigenen
// `git worktree` am selben Commit, der danach weggeräumt wird.
import { execFileSync, spawnSync } from 'node:child_process'
import { mkdtempSync, readFileSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'

import { geschichtet, mitAktivierung, wegraeumen, wilson } from './zufallsmutationkern.mjs'

const wurzel = resolve(import.meta.dirname, '..')
const pakete = ['internal/domain', 'internal/service', 'internal/storage']

const argument = (name, vorgabe) => {
  const stelle = process.argv.indexOf(name)
  return stelle < 0 ? vorgabe : process.argv[stelle + 1]
}
const samen = Number(argument('--samen', '2224'))
const je = Number(argument('--je', '40'))

const commit = execFileSync('git', ['rev-parse', 'HEAD'], { cwd: wurzel, encoding: 'utf8' }).trim()
const baum = mkdtempSync(join(tmpdir(), 'zufallsmutation-'))
execFileSync('git', ['worktree', 'add', '--detach', baum, commit], { cwd: wurzel, stdio: 'ignore' })

const isoliert = { ...process.env, HOME: mkdtempSync(join(tmpdir(), 'zufallsmutation-heim-')), GOFLAGS: '-mod=readonly' }
let umgebung = isoliert
const go = (argumente, frist) => spawnSync('go', argumente, { cwd: baum, env: umgebung, encoding: 'utf8', timeout: frist, maxBuffer: 64 << 20 })

try {
  // **Die Aktivierung sieht schon das Wegwerf-Zuhause**, und ihre Ausgabe
  // überschreibt es nicht (`mitAktivierung`).
  umgebung = mitAktivierung(isoliert, execFileSync('bash', ['-c', 'source scripts/activate.sh >/dev/null && env'],
    { cwd: baum, env: isoliert, encoding: 'utf8' }))

  const alle = JSON.parse(go(['run', 'scripts/mutationsorte.go', ...pakete], 600000).stdout)
  const gezogen = geschichtet(alle, samen, je)

  // **Der Grundlauf zuerst**: ein rotes Paket ohne Mutant macht jeden Fang
  // wertlos, und seine Dauer setzt die Frist.
  const frist = {}
  for (const paket of pakete) {
    const beginn = Date.now()
    const lauf = go(['test', '-count=1', `./${paket}/`], 1800000)
    if (lauf.status !== 0) throw new Error(`der Grundlauf von ${paket} ist rot:\n${lauf.stdout.slice(-2000)}`)
    frist[paket] = Math.max(60000, 3 * (Date.now() - beginn))
  }

  const ergebnisse = []
  for (const [nummer, m] of gezogen.entries()) {
    const datei = join(baum, m.datei)
    const vorher = readFileSync(datei)
    const text = vorher.toString('utf8')
    const bytes = Buffer.from(text, 'utf8')
    const mutiert = Buffer.concat([bytes.subarray(0, m.versatz), Buffer.from(m.ersatz), bytes.subarray(m.versatz + m.laenge)])
    if (bytes.subarray(m.versatz, m.versatz + m.laenge).toString() !== m.urtext) throw new Error(`Stelle verschoben: ${m.datei}:${m.zeile}`)
    writeFileSync(datei, mutiert)
    const beginn = Date.now()
    try {
      const lauf = go(['test', '-count=1', '-failfast', `./${m.paket}/`], frist[m.paket])
      const ausgabe = `${lauf.stdout}${lauf.stderr}`
      const urteil = lauf.error?.code === 'ETIMEDOUT' || lauf.signal ? 'frist'
        : /\[build failed\]|\[setup failed\]|cannot use|mismatched types|invalid operation/.test(ausgabe) ? 'nicht übersetzt'
        : lauf.status === 0 ? 'überlebt' : 'gefangen'
      ergebnisse.push({ nummer, ...m, urteil, dauerMs: Date.now() - beginn, test: urteil === 'gefangen' ? (ausgabe.match(/--- FAIL: (\S+)/)?.[1] ?? '') : '' })
    } finally {
      writeFileSync(datei, vorher)
    }
    console.log(`${String(nummer + 1).padStart(3)}/${gezogen.length} ${ergebnisse.at(-1).urteil.padEnd(15)} ${m.datei}:${m.zeile} ${m.urtext} → ${m.ersatz}`.slice(0, 200))
  }

  const ziel = join(wurzel, 'docs/belege/zufallsmutation.json')
  const maschine = `${execFileSync('sysctl', ['-n', 'hw.model'], { encoding: 'utf8' }).trim()}, macOS ${execFileSync('sw_vers', ['-productVersion'], { encoding: 'utf8' }).trim()}`
  const befehl = `node scripts/zufallsmutation.mjs --samen ${String(samen)} --je ${String(je)}`
  writeFileSync(ziel, `${JSON.stringify({ commit, samen, je, stellen: alle.length, frist, ende: new Date().toISOString(), maschine, befehl, rueckgabe: 0, ergebnisse }, null, 1)}\n`)
  const zahl = (u) => ergebnisse.filter((e) => e.urteil === u).length
  const [unten, oben] = wilson(zahl('gefangen'), zahl('gefangen') + zahl('überlebt'))
  console.log(`gefangen ${zahl('gefangen')}, überlebt ${zahl('überlebt')}, nicht übersetzt ${zahl('nicht übersetzt')}, Frist ${zahl('frist')} — vor der Durchsicht ${(100 * unten).toFixed(1)} bis ${(100 * oben).toFixed(1)} %`)
} finally {
  spawnSync('git', ['worktree', 'remove', '--force', baum], { cwd: wurzel })
  wegraeumen(baum)
  wegraeumen(isoliert.HOME)
}
