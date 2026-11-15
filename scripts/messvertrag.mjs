// **Der Messvertrag für 0.9.0: je Kernweg Zeit und Speicher, mit Grenze** —
// PERF01 (#2351),
// Befunde A-13, A-14, A-25 und A-26 des Audits 0.9.0.
//
// # Warum es das gibt
//
// **Bis PERF01 hatten nur Start, Vorschau und Rollen ein Budget**, und für
// Speicher stand nur ein Wachstumsziel da. Ein großes Set öffnen, eine
// Sammlung mit 25.000 Zeilen durchsuchen, Ausgeben, Drucken, Zurückspielen,
// der Fußabdruck nach zwei Stunden, ein Dienst nach einem Schub Rasterungen,
// das Bündel und der Leerlauf: ohne Grenze, und damit bemerkte niemand eine
// Verschlechterung, bevor Nutzende sie spüren.
//
// # Wie
//
// **Eine Tafel, aus der drei Dinge folgen**: die Zeile, die ein Drehbuch in
// seinen Beleg schreibt ([tafel]), das Urteil, mit dem es rot wird
// ([urteil]), und die Einträge, mit denen der Belegplan die Grenze hält
// ([planeintraege], `docs/belege/BELEGPLAN.json`). Ein Test hält Plan und
// Tafel gleich (`messvertrag.test.mjs`) — **eine Grenze steht nur einmal**.
//
// **Die Grenzen sind die vorläufigen aus dem Issue**, festgelegt vor der
// ersten Messung; wo eine reißt, wird das entschieden und nicht die Grenze
// passend gemacht.
//
// # Was er nicht hält
//
// **Ob ein Drehbuch misst, was seine Zeile sagt** — das sagt sein Kopf. Und
// **diese Maschine**: jeder Beleg nennt sie, und eine Grenze, die hier hält,
// ist keine Zusage für einen Mac mit 8 GB.

// MESSVERTRAG ist die Tafel: je Reihe das Maß, die Grenze und der Beleg, in
// dem sie steht. `P95` verlangt mindestens zwanzig Werte; `max` hält jeden.
export const MESSVERTRAG = [
  ...[
    ['Start kalt (ms)', 'P95', 2000],
    ['Start warm (ms)', 'P95', 2000],
    ['Set öffnen, 3.788 Karten (ms)', 'P95', 1000],
    ['Feld → Vorschaubild (ms)', 'P95', 300],
    ['Sammlung durchsuchen, 25.000 Zeilen (ms)', 'P95', 500],
    ['Set ausgeben als Bilder, 60 Karten (ms)', 'max', 60000],
    ['Druckbogen, 60 Karten (ms)', 'max', 60000],
    ['Zurückspielen, 25.000 Zeilen (ms)', 'max', 60000],
    ['Fußabdruck nach jedem Weg (MiB)', 'max', 1536],
    ['Bündel, logisch (MiB)', 'max', 500],
  ].map(([reihe, mass, grenze]) => ({ reihe, mass, grenze, beleg: 'huellenprobe/budgets09.md' })),
  ...['sichtbar', 'ausgeblendet', 'minimiert'].flatMap((zustand) => ['Anwendung', 'Webinhalt', 'GPU', 'Netz'].map((prozess) =>
    ({ reihe: `Leerlauf ${zustand}, ${prozess} (% CPU)`, mass: 'max', grenze: 1, stellen: 2, mindestSekunden: 600, beleg: 'huellenprobe/leerlauf09.md' }))),
  { reihe: 'Fußabdruck nach 120 Minuten (MiB)', mass: 'max', grenze: 1536, beleg: 'huellenprobe/dauerbetrieb.md' },
  ...['mse-rest', 'mse-mcp'].flatMap((dienst) => ['', ', mit GOMEMLIMIT'].map((grenze) =>
    ({ reihe: `Dienst nach Rendern und Ruhe, ${dienst}${grenze} (MiB)`, mass: 'max', grenze: 512, beleg: 'dienstspeicher.md' }))),
  // **PERF02** (#2352): das Grundgerüst der Kennzahlen in 500 ms, die Prüfung
  // eines Sets mit 3.788 Karten in 60 s — beides die Grenzen aus dem Issue.
  ...['5 Karten', '3.788 Karten', '25.000 Karten in 50 Sets'].flatMap((bestand) =>
    ['Kennzahl nach dem Start, kalt', 'Kennzahl nach dem Start, warm', 'Kennzahl beim Aufruf der Startseite'].map((weg) =>
      ({ reihe: `${weg}, ${bestand} (ms)`, mass: 'P95', grenze: 500, beleg: 'huellenprobe/kennzahlen09.md', paket: 'PERF02' }))),
  { reihe: 'Hintergrundprüfung, 3.788 Karten (ms)', mass: 'max', grenze: 60000, beleg: 'huellenprobe/kennzahlen09.md', paket: 'PERF02' },
  // **Die Kernwege nach Phase 3** (#2532, Gesamtreview über 0.9.0, Befund 6):
  // der Markttisch lädt fürs eigene Fach die ganze Sammlung, deshalb mit den
  // 25.000 Zeilen aus `budgetsorte09`. Die Grenzen stehen vor der Messung:
  // aufschlagen wie ein Set öffnen (1 s), anbieten doppelt so lang, weil die
  // Leitung dazwischen liegt, eine Leiste wie die Kennzahl der Startseite.
  ...[
    ['Markttisch aufschlagen, 25.000 eigene Exemplare (ms)', 'P95', 1000],
    ['Anbieten bis zum neuen Stand (ms)', 'P95', 2000],
    ...['Set, 3.788 Karten', 'Sammlung, 25.000 Zeilen', 'Deck', 'Pakete'].map((leiste) => [`Kennzahlleiste ${leiste} (ms)`, 'P95', 500]),
    ['Fußabdruck nach Markttisch, Handel und Leisten (MiB)', 'max', 1536],
  ].map(([reihe, mass, grenze]) => ({ reihe, mass, grenze, beleg: 'huellenprobe/budgetswege09.md', paket: '#2532' })),
]

// kennzahlen fasst eine Reihe zusammen — P95 erst ab zwanzig Werten, wie in
// `huellenprobe/budgets.mjs`.
export function kennzahlen(werte) {
  const s = [...werte].sort((a, b) => a - b)
  const q = (anteil) => s[Math.min(s.length - 1, Math.floor(anteil * s.length))]

  return { n: s.length, min: s[0], median: q(0.5), p95: s.length >= 20 ? q(0.95) : undefined, max: s.at(-1) }
}

// gueltig sind die Werte einer Reihe, die eine Messung sind — ein Drehbuch
// trägt −1 ein, wenn eine scheiterte.
const gueltig = (werte) => werte.filter((w) => w >= 0)

// urteil sagt, ob eine Reihe ihre Grenze hält. **Eine gescheiterte Messung
// ist kein Wert**: im ersten Lauf von `budgets09` standen 23 Mal −1 in der
// Reihe der Vorschau, und ihr P95 von −1 „hielt".
//
// **Und eine zu kurze Messung auch nicht** (Zweitblick über Phase 3, Befund
// 13): wer `mindestSekunden` trägt, verlangt, dass so lange gemessen wurde —
// `lauf.sekunden` sagt, wie lange es war.
export function urteil(eintrag, werte, lauf = {}) {
  if (werte.length === 0) return 'nicht gemessen'
  if (gueltig(werte).length < werte.length) return 'gescheitert'
  if (eintrag.mindestSekunden !== undefined && !((lauf.sekunden ?? 0) >= eintrag.mindestSekunden)) return 'zu kurz'
  const k = kennzahlen(werte)
  const gemessen = eintrag.mass === 'P95' ? k.p95 : k.max
  if (gemessen === undefined) return 'zu wenige Werte'

  return gemessen <= eintrag.grenze ? 'gehalten' : 'gerissen'
}

const zahl = (wert, stellen) => (wert === undefined ? '—' : String(Math.round(wert * 10 ** stellen) / 10 ** stellen))
const grenztext = (eintrag) => `${eintrag.mass} ≤ ${String(eintrag.grenze)}`

// tafel gibt die Zeile einer Reihe im Beleg:
// `| Reihe | n | min | Median | P95 | max | Grenze | Urteil |`.
export function tafel(eintrag, werte, lauf = {}) {
  const stellen = eintrag.stellen ?? 1
  const k = gueltig(werte).length === 0 ? {} : kennzahlen(gueltig(werte))
  const spalten = [k.min, k.median, k.p95, k.max].map((w) => zahl(w, stellen))

  return `| ${eintrag.reihe} | ${String(werte.length)} | ${spalten.join(' | ')} | ${grenztext(eintrag)} | ${urteil(eintrag, werte, lauf)} |`
}

const roh = (text) => text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

// muster ist der Ausdruck, mit dem der Belegplan den gemessenen Wert liest —
// die Spalte P95 oder max. Ein Strich statt einer Zahl trifft nicht, und der
// Plan wird rot: ohne zwanzig Werte kein P95.
//
// **Und das Urteil der Zeile muss eines über eine Zahl sein** — gehalten oder
// gerissen (Zweitblick über Phase 3, Befund 12): eine Reihe mit einer
// gescheiterten Messung zeigt das Höchstmaß der übrigen, und ohne das Urteil
// las der Plan es als gehalten. Gerissen trifft, damit der Plan sagt, um wie
// viel; gescheitert, zu kurz und zu wenige Werte treffen nicht.
export function muster(eintrag) {
  const zahlen = eintrag.mass === 'P95'
    ? '\\| \\d+ \\| [\\d.]+ \\| [\\d.]+ \\| ([\\d.]+) \\| [\\d.]+ \\| [^|]+ \\| (?:gehalten|gerissen) \\|'
    : '\\| \\d+ \\| [\\d.]+ \\| [\\d.]+ \\| (?:[\\d.]+|—) \\| ([\\d.]+) \\| [^|]+ \\| (?:gehalten|gerissen) \\|'

  return `\\| ${roh(eintrag.reihe)} ${zahlen}`
}

// planeintraege gibt je Beleg die Grenzen für den Belegplan.
export function planeintraege() {
  const je = new Map()
  for (const e of MESSVERTRAG) {
    if (!je.has(e.beleg)) je.set(e.beleg, { beleg: e.beleg, werte: [] })
    je.get(e.beleg).werte.push({ name: e.reihe, muster: muster(e), hoechstens: e.grenze })
  }

  return [...je.values()].map((e) => ({ schritt: `${MESSVERTRAG.find((x) => x.beleg === e.beleg).paket ?? 'PERF01'}: ${e.beleg}`, fristTage: 30, ...e }))
}

// fussabdrucksumme zählt den Fußabdruck der genannten Prozesse zusammen —
// **−1, wenn einer fehlt** (Zweitblick über Phase 3, Befund 12): `footprint`
// nennt einen Prozess, den es nicht messen konnte, gar nicht, und eine Null
// an seiner Stelle hielt jede Grenze.
export function fussabdrucksumme(werte, pids) {
  if (pids.some((pid) => !werte.has(String(pid)))) return -1

  return pids.reduce((summe, pid) => summe + werte.get(String(pid)), 0)
}
