// **Bis zur sichtbaren Kennzahl, und die Prüfung eines ganzen Sets** — PERF02
// (#2352).
//
// # Was gemessen wird
//
// - **Die Kennzahlen der Startseite** in drei künstlichen Beständen — ein Set
//   mit 5 Karten, eines mit 3.788 (`GROSS` aus `budgetsorte09`), und 50 Sets
//   zu je 500 Karten, zusammen 25.000. Je Bestand **30 Starts kalt** (ein
//   frisch geklontes Zuhause je Start), **30 warm** (dasselbe Zuhause nach
//   ⌘Q) und **30 Aufrufe der Startseite** (⌘2, dann ⌘1).
// - **Gemessen an der Marke** `mse:kennzahlen`, die der Überblick bei seinem
//   Erscheinen setzt — und er erscheint erst mit seinen Zahlen. Beim Start
//   ist das die Zeit seit dem Laden der Seite, beim Aufruf die seit dem
//   Augenblick vor ⌘1. **Eine besuchte Seite bleibt verborgen stehen**
//   (`<Activity>`, UX-03.AK-02): ⌘2 verbirgt den Überblick, statt ihn
//   abzubauen, und ⌘1 zeigt ihn mit seinen Zahlen und setzt die Marke neu,
//   weil die Effekte wieder laufen. Gezählt werden deshalb nur **sichtbare**
//   Zahlen — der erste Lauf wartete zehn Sekunden auf einen Überblick, der nie
//   verschwand. Daneben, ohne Grenze: die Zeit seit dem Öffnen über
//   LaunchServices, der Fußabdruck von Anwendung und WebKit an der Kennzahl
//   und die Bindungsrufe bis dahin.
// - **Kein Rasterlauf je Startseitenaufruf**: unter den Bindungsrufen eines
//   Aufrufs steht keiner, der zeichnet, eine Vorschau holt oder prüft.
// - **Die Prüfung eines ganzen Sets**: `mse probleme` über die 3.788 Karten,
//   mit dem Stil M15 und den Schriften aus dem Pack — derselbe Weg wie an MCP
//   und Dienst.
//
// # Was es nicht misst
//
// **Bilder und Problemfälle ändern keine Kennzahl** — sie sind Zählungen über
// SQL. In die Prüfung gehen sie ein; dort misst `BenchmarkWegProblempruefung`
// sie am kanonischen `Black`. **Und das Fenster zeigt noch keine
// Problemzahl**: Fortschritt und Alter am Fenster kommen mit AP-8
// (#2365).
import { execFileSync } from 'node:child_process'
import { existsSync, writeFileSync } from 'node:fs'
import { cpus, loadavg, totalmem } from 'node:os'
import { join, resolve } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'
import { mischen } from '../huellenrufbuch.mjs'
import { MESSVERTRAG, fussabdrucksumme, kennzahlen, tafel, urteil } from '../messvertrag.mjs'
import { GROSS, mitPack, mtgjson, pack } from './budgetsorte09.mjs'
import { fussabdruck } from './dauerbetrieb.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ZUEGE = Number(process.env.MSE_KENNZAHLEN_ZUEGE ?? 30)
const MARKE = 'mse:kennzahlen'
const BELEG = 'huellenprobe/kennzahlen09.md'
// UEBERBLICK sind die Zahlen des Überblicks — Sets, Karten, Exemplare, Decks.
const UEBERBLICK = 'ul[aria-label="Dein Überblick"] .tabular-nums'

// BESTAENDE sind die drei Größen des Issues — je Bestand die Sets, die er
// bekommt, und die Kartenzahl, die der Überblick nennen muss.
export const BESTAENDE = [
  { name: '5 Karten', karten: 5, sets: [{ code: 'K05', name: 'Kennzahlsatz klein', karten: 5, vorsatz: 'Kleinkarte' }] },
  { name: '3.788 Karten', karten: GROSS.karten, sets: [GROSS] },
  {
    name: '25.000 Karten in 50 Sets', karten: 25000,
    sets: Array.from({ length: 50 }, (_, i) => ({ code: `K${String(i + 10)}`, name: `Kennzahlsatz ${String(i + 1).padStart(2, '0')}`, karten: 500, vorsatz: `Kennkarte ${String(i + 1).padStart(2, '0')}` })),
  },
]

// eigenesSet ist `mtgjson` mit eigenen Kennungen je Set — die Vorlage gibt
// allen Sets außer `GROSS` dieselben, und 50 Sets mit gleichen Kennungen
// prüften das Einfahren und nicht die Zählung.
function eigenesSet(set, stelle) {
  const daten = mtgjson(set)
  daten.data.cards.forEach((karte, i) => { karte.uuid = `00000000-0000-4000-b${String(stelle).padStart(3, '0')}-${(i + 1).toString(16).padStart(12, '0')}` })

  return daten
}

// befuellen legt einen Bestand an und gibt die Kennung seines ersten Sets.
function befuellen(p, ort, bestand) {
  mitPack(p, ort)
  bestand.sets.forEach((set, i) => {
    const datei = join(ort.ort, `${set.code}.json`)
    writeFileSync(datei, JSON.stringify(set === GROSS ? mtgjson(set) : eigenesSet(set, i)))
    p.mse(ort, 'mtgjson-lesen', datei, 'magic')
  })
  const zeilen = p.mse(ort, 'sets').trim().split('\n').map((z) => z.split('\t'))
  const erstes = zeilen.find((t) => t[1] === bestand.sets[0].name)?.[0]
  if (erstes !== undefined) p.mse(ort, 'stile', erstes, 'wechseln', 'magic-m15')
  p.zeige(`${bestand.name}: ${String(zeilen.length)} Sets, erstes Set ${String(erstes)}`)

  return erstes
}

// klon gibt ein frisches Zuhause mit dem Bestand von `quelle` — ein Klon auf
// APFS, kein Kopieren.
function klon(p, quelle) {
  const ziel = p.ort()
  execFileSync('cp', ['-cR', `${quelle.heim}/.`, ziel.heim])
  for (const endung of ['', '-wal', '-shm']) {
    if (existsSync(quelle.bestand + endung)) execFileSync('cp', ['-c', quelle.bestand + endung, ziel.bestand + endung])
  }

  return ziel
}

const prozesse = () => execFileSync('ps', ['-Ao', 'pid=,comm=']).toString().split('\n')
  .map((z) => z.trim().split(/\s+/)).filter((t) => t.length >= 2).map(([pid, ...name]) => ({ pid, name: name.join(' ') }))

// speicher ist der Fußabdruck der Anwendung und der WebKit-Prozesse, die
// seit `vorher` dazukamen, in MiB — **−1, wenn einer fehlt** (Zweitblick über
// Phase 3, Befund 12).
function speicher(vorher, anwendung) {
  const web = prozesse().filter((x) => !vorher.has(x.pid) && /WebKit\.(WebContent|GPU|Networking)$/.test(x.name)).map((x) => x.pid)

  return fussabdrucksumme(fussabdruck([anwendung, ...web]), [anwendung, ...web])
}

// rufe holt das Rufbuch seit dem letzten Holen — Zahl und Namen.
async function rufe(p) {
  const antwort = await p.leitung.frage({ was: 'rufbuch', leeren: true })
  if (!antwort.ok) return { anzahl: -1, namen: [] }
  mischen(p.rufbuch ??= {}, antwort.daten)

  return { anzahl: Object.values(antwort.daten).reduce((s, z) => s + z.anzahl, 0), namen: Object.keys(antwort.daten) }
}

// marke wartet, bis der Überblick mit der erwarteten Kartenzahl steht, und
// gibt die Zeit seiner Marke.
async function marke(p, karten, seit = 0) {
  const steht = `const m = performance.getEntriesByName('${MARKE}', 'mark').at(-1); const z = [...document.querySelectorAll('${UEBERBLICK}')].filter((e) => e.checkVisibility()).map((e) => Number(e.textContent)); return m && m.startTime > ${String(seit)} && z[1] === ${String(karten)} ? m.startTime : null`
  await p.warte(`der Überblick nennt ${String(karten)} Karten`, async () => (await p.seite(steht)) !== null, 30000, 20)

  return { zeit: await p.seite(steht), ursprung: await p.seite('return performance.timeOrigin') }
}

// **Lesen ist erlaubt, Prüfen nicht** — seit AP-8 (#2365) liest die
// Startseite die Prüfstände (`Setprobleme`, ohne zu zeichnen); verboten bleibt,
// was prüft (`SetproblemePruefen`). Hier stand `problem`, und das traf auch das
// Lesen.
const ZEICHNEND = /vorschau|zeichne|raster|problemepruefen/i

async function startreihe(p, quelle, bestand, werte, art) {
  let ort = quelle
  // **Warm einmal mehr**: der erste Start in einem Zuhause ist kalt und zählt nicht.
  for (let i = 0; i < ZUEGE + (art === 'warm' ? 1 : 0); i++) {
    if (art === 'kalt') ort = klon(p, quelle)
    const vorher = new Set(prozesse().map((x) => x.pid))
    const beginn = Date.now()
    await p.starte({ ort, takt: 20 })
    const { zeit, ursprung } = await marke(p, bestand.karten)
    const gerufen = await rufe(p)
    if (art === 'kalt' || i > 0) {
      werte[`Kennzahl nach dem Start, ${art}, ${bestand.name} (ms)`].push(zeit)
      werte.neben[`Öffnen bis Kennzahl, ${art}, ${bestand.name} (ms)`].push(ursprung + zeit - beginn)
      werte.neben[`Fußabdruck an der Kennzahl, ${art}, ${bestand.name} (MiB)`].push(speicher(vorher, p.pids.at(-1)))
      werte.neben[`Bindungsrufe bis zur Kennzahl, ${art}, ${bestand.name}`].push(gerufen.anzahl)
    }
    p.pruefe(`${bestand.name}, ${art} ${String(i)}: ⌘Q beendet`, await p.beende())
  }
}

async function aufrufe(p, ort, bestand, werte) {
  await p.starte({ ort })
  await marke(p, bestand.karten)
  await rufe(p)
  const gerufen = new Set()
  for (let i = 0; i < ZUEGE; i++) {
    await p.taste('2', 'befehl')
    await p.warte('die Startseite ist verborgen', () => p.seite(`return [...document.querySelectorAll('${UEBERBLICK}')].every((e) => !e.checkVisibility())`), 10000, 20)
    await p.ruhe()
    const vor = await p.seite('return performance.now()')
    await p.taste('1', 'befehl')
    const { zeit } = await marke(p, bestand.karten, vor)
    werte[`Kennzahl beim Aufruf der Startseite, ${bestand.name} (ms)`].push(zeit - vor)
    for (const name of (await rufe(p)).namen) gerufen.add(name)
  }
  const zeichnend = [...gerufen].filter((n) => ZEICHNEND.test(n))
  p.pruefe(`${bestand.name}: kein Aufruf der Startseite zeichnet oder prüft`, zeichnend.length === 0, zeichnend.join(', ') || [...gerufen].sort().join(', '))
  p.pruefe(`${bestand.name}: ⌘Q beendet`, await p.beende())
}

// pruefung fährt `mse probleme` über das große Set und misst die Wanduhr.
function pruefung(p, ort, setID, werte) {
  const beginn = Date.now()
  const roh = execFileSync(join(wurzel, '.go/bin/mse'), ['probleme', setID, '--json'], { cwd: wurzel, env: kindumgebung(ort), maxBuffer: 256 << 20 }).toString()
  werte['Hintergrundprüfung, 3.788 Karten (ms)'].push(Date.now() - beginn)
  const stand = JSON.parse(roh)
  p.pruefe('die Prüfung ist vollständig und aktuell', stand.urteil === 'aktuell' && stand.sicher && stand.geprueft === GROSS.karten, `${stand.urteil}, ${String(stand.geprueft)} von ${String(stand.gesamt)}, ${String(stand.zahl)} mit Problemen`)
}

export default async function kennzahlen09(p) {
  p.zeige(`Last vor dem Lauf ${loadavg().map((z) => z.toFixed(2)).join(' / ')} auf ${String(cpus().length)} Kernen, ${String(Math.round(totalmem() / 2 ** 30))} GiB; ${String(ZUEGE)} Züge je Reihe`)
  const eigene = MESSVERTRAG.filter((e) => e.beleg === BELEG)
  const werte = Object.fromEntries(eigene.map((e) => [e.reihe, []]))
  werte.neben = {}
  for (const b of BESTAENDE) {
    for (const art of ['kalt', 'warm']) {
      for (const reihe of [`Öffnen bis Kennzahl, ${art}, ${b.name} (ms)`, `Fußabdruck an der Kennzahl, ${art}, ${b.name} (MiB)`, `Bindungsrufe bis zur Kennzahl, ${art}, ${b.name}`]) werte.neben[reihe] = []
    }
  }

  if (!existsSync(join(pack, 'data'))) {
    p.pruefe(`der Pack liegt unter ${pack}`, false)
  } else {
    for (const bestand of BESTAENDE) {
      p.notiz(`Bestand: ${bestand.name}`)
      const ort = p.ort()
      const erstes = befuellen(p, ort, bestand)
      await startreihe(p, ort, bestand, werte, 'kalt')
      await startreihe(p, ort, bestand, werte, 'warm')
      await aufrufe(p, ort, bestand, werte)
      if (bestand.sets[0] === GROSS && erstes !== undefined) pruefung(p, ort, erstes, werte)
    }
  }

  p.notiz('Ergebnis')
  p.zeilen.push('', '| Reihe | n | min | Median | P95 | max | Grenze | Urteil |', '| --- | --- | --- | --- | --- | --- | --- | --- |')
  for (const e of eigene) {
    p.zeilen.push(tafel(e, werte[e.reihe]))
    p.pruefe(`${e.reihe}: ${e.mass} ≤ ${String(e.grenze)}`, urteil(e, werte[e.reihe]) === 'gehalten', urteil(e, werte[e.reihe]))
  }
  p.zeilen.push('', '**Daneben, ohne Grenze**', '', '| Reihe | n | min | Median | P95 | max |', '| --- | --- | --- | --- | --- | --- |')
  for (const [reihe, liste] of Object.entries(werte.neben)) {
    // **Eine gescheiterte Messung (−1) ist kein Wert** — gezählt wird sie
    // daneben, nicht mitgerechnet (Zweitblick über Phase 3, Befund 12).
    const gueltige = liste.filter((w) => w >= 0)
    const k = gueltige.length === 0 ? {} : kennzahlen(gueltige)
    const n = `${String(gueltige.length)}${gueltige.length < liste.length ? ` (${String(liste.length - gueltige.length)} gescheitert)` : ''}`
    p.zeilen.push(`| ${reihe} | ${n} | ${[k.min, k.median, k.p95, k.max].map((w) => (w === undefined ? '—' : String(Math.round(w)))).join(' | ')} |`)
  }
  p.zeilen.push('', `Rohwerte — ${eigene.map((e) => `${e.reihe}: ${JSON.stringify(werte[e.reihe].map(Math.round))}`).join('; ')}.`)
}
