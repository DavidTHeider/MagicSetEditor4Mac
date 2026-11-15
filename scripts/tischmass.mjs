// **Was der Tischplan kostet** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Was hier gemessen wird
//
// **Die Rechnung, nicht die Rasterung.** #1460 verlangt für die
// Leistungsszene „60 FPS als Ziel mit benanntem Mac/macOS, Auflösung, Dauer,
// Warm-/Kaltcache und Framezeitverteilung". Davon liegt **eine Hälfte** in
// diesem Prozess: der Weg von der Partie zum Plan und weiter zu den Sprites,
// den PixiJS bekommt. Die andere Hälfte — was die Grafikkarte daraus macht —
// braucht ein echtes Fenster und steht in `docs/TISCHLEISTUNG.md` als
// **offen** und nicht als gemessen.
//
// **Ein Bild hat 16,67 ms.** Wenn schon die Rechnung sie füllt, ist die Frage
// nach der Grafikkarte gegenstandslos; deshalb wird sie zuerst gestellt.
//
// # Wie gemessen wird
//
// **Verteilung statt Mittelwert.** Das Issue sagt es ausdrücklich: „keine
// reine FPS-Mittelwertbehauptung". Ausgegeben werden Median, 95. und 99.
// Perzentil und das Maximum — ein Mittelwert verbirgt genau den Ausreißer,
// den ein Auge als Ruckeln sieht.
//
// Aufruf:
//
//	node scripts/tischmass.mjs
import { performance } from 'node:perf_hooks'
import { execFileSync } from 'node:child_process'
import { createJiti } from '../frontend/node_modules/jiti/lib/jiti.mjs'

// **`jiti` und keine neue Abhängigkeit.** Es liegt ohnehin im Baum (ESLint
// bringt es mit) und lädt TypeScript ohne Übersetzungsschritt. Ein eigener
// Läufer dafür wäre ein Werkzeug mehr in der Tabelle aus §1 — für ein Skript,
// das von Hand gefahren wird.
const jiti = createJiti(import.meta.url)

const laeufe = Number(process.env.TISCHMASS_LAEUFE ?? 200)

const perzentil = (werte, anteil) => {
  const sortiert = [...werte].sort((a, b) => a - b)
  return sortiert[Math.min(sortiert.length - 1, Math.floor(sortiert.length * anteil))]
}

const messen = (was, tun) => {
  // **Erst warmlaufen.** Der erste Lauf bezahlt die Übersetzung des JIT; ihn
  // mitzuzählen hieße, den Kaltstart als Framezeit auszugeben.
  for (let i = 0; i < 20; i += 1) tun()

  const zeiten = []
  for (let i = 0; i < laeufe; i += 1) {
    const anfang = performance.now()
    tun()
    zeiten.push(performance.now() - anfang)
  }

  return {
    was,
    median: perzentil(zeiten, 0.5),
    p95: perzentil(zeiten, 0.95),
    p99: perzentil(zeiten, 0.99),
    max: Math.max(...zeiten),
  }
}

const zeile = (mass) =>
  `| ${mass.was.padEnd(24)} | ${mass.median.toFixed(3).padStart(8)} | ` +
  `${mass.p95.toFixed(3).padStart(8)} | ${mass.p99.toFixed(3).padStart(8)} | ` +
  `${mass.max.toFixed(3).padStart(8)} |`

const maschine = () => {
  const modell = execFileSync('sysctl', ['-n', 'hw.model']).toString().trim()
  const kerne = execFileSync('sysctl', ['-n', 'hw.ncpu']).toString().trim()
  const os = execFileSync('sw_vers', ['-productVersion']).toString().trim()

  return `${modell}, ${kerne} Kerne, macOS ${os}, Node ${process.version}`
}

const { szene } = await jiti.import('../frontend/src/test/tischszenen.ts')
const { tischplan } = await jiti.import('../frontend/src/tisch/tischplan.ts')
const { kartenbilder } = await jiti.import('../frontend/src/tisch/weltbild.ts')
const { abgeglichen } = await jiti.import(
  '../frontend/src/tisch/szenenabgleich.ts',
)

const partie = szene('extreme-board-state')
const plan = tischplan(partie)
const bild = () => undefined
const bilder = kartenbilder(plan, bild)
const vorher = new Set(bilder.map((eines) => eines.kennung))

const masse = [
  messen('tischplan()', () => tischplan(partie)),
  messen('kartenbilder()', () => kartenbilder(plan, bild)),
  messen('abgeglichen()', () => abgeglichen(vorher, bilder)),
  messen('alles zusammen', () => {
    const einer = tischplan(partie)
    const seine = kartenbilder(einer, bild)
    abgeglichen(vorher, seine)
  }),
]

console.log(`Maschine: ${maschine()}`)
console.log(`Szene:    extreme-board-state — ${String(partie.sitze.length)} Sitze, ` +
  `${String(partie.objekte.length)} Objekte, ${String(bilder.length)} Sprites`)
console.log(`Läufe:    ${String(laeufe)} je Messung, 20 vorab zum Warmlaufen\n`)
console.log('| Was                      |   Median |      p95 |      p99 |      max |')
console.log('| ------------------------ | -------- | -------- | -------- | -------- |')
for (const mass of masse) console.log(zeile(mass))
console.log('\nAlle Zeiten in Millisekunden. Ein Bild bei 60 Hz hat 16,667 ms.')
