// Prüfungen des Belegplans — B06
// (#2225).
//
// **Die Positivkontrollen stehen hier**: ein überfälliger Lauf und ein
// absichtlich verschlechterter Messwert müssen rot werden, sonst sagt ein
// grüner Plan nichts.
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { angabenAus, pruefe } from './belegplan.mjs'

const heute = new Date('2026-10-09T12:00:00Z')

const huelle = (datum, wert = '120') => `# Hüllenprobe: budgets

**bestanden** — Commit \`ffadb0e5e\`,
${datum} 07:00 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus \`scripts/bauen.sh abnahme\`, gestartet über LaunchServices.
Lauf \`2026-10-09T07-00-20Z-44654\`, Bauabdruck \`17c8bcbdc895\`.

| Feld → Vorschaubild (ms) | 24 | 80 | ${wert} | 140 | 200 | Median ≤ 150 |
`

const eintrag = { schritt: 'Budgets', beleg: 'budgets.md', fristTage: 30, wert: { muster: 'Feld → Vorschaubild \\(ms\\) \\| \\d+ \\| [\\d.]+ \\| ([\\d.]+)', hoechstens: 150 } }

test('ein Hüllenbeleg nennt Commit, Datum, Maschine und Befehl', () => {
  const angaben = angabenAus(huelle('2026-10-09'), 'md')
  assert.equal(angaben.commit, 'ffadb0e5e')
  assert.equal(angaben.datum.toISOString().slice(0, 10), '2026-10-09')
  assert.match(angaben.maschine, /Mac17,3/)
  assert.match(angaben.befehl, /bauen\.sh abnahme/)
})

test('ein frischer Beleg im Rahmen hält', () => {
  assert.deepEqual(pruefe(eintrag, huelle('2026-10-09'), heute), [])
})

test('ein überfälliger Lauf wird rot — die Positivkontrolle der Frist', () => {
  assert.match(pruefe(eintrag, huelle('2026-08-01'), heute).join(' '), /überfällig/)
})

test('ein absichtlich verschlechterter Messwert wird rot — die Positivkontrolle der Schwelle', () => {
  assert.match(pruefe(eintrag, huelle('2026-10-09', '180'), heute).join(' '), /180.*150/)
})

test('ein Messwert, den das Muster nicht findet, ist rot und nicht still', () => {
  assert.match(pruefe(eintrag, '# leer\n\nCommit `abc`', heute).join(' '), /nicht gefunden/)
})

test('ein Beleg ohne Commit ist rot', () => {
  assert.match(pruefe({ ...eintrag, wert: undefined }, '# ohne\n\n2026-10-09 07:00 UTC, Mac17,3', heute).join(' '), /Commit/)
})

test('ein fehlender Beleg ist rot', () => {
  assert.match(pruefe(eintrag, undefined, heute).join(' '), /fehlt/)
})

test('ein Laufbeleg (JSON) nennt Commit, Ende und Befehl', () => {
  const roh = JSON.stringify({ commit: 'abc123', ende: '2026-10-08T10:00:00Z', befehl: ['go', 'test'], rueckgabe: 0, werkzeuge: { go: '1.27.1' }, gemessen: { quote: 0.82 } })
  const angaben = angabenAus(roh, 'json')
  assert.equal(angaben.commit, 'abc123')
  assert.equal(angaben.rueckgabe, 0)
  const mitSchwelle = { schritt: 'Quote', beleg: 'x.json', fristTage: 30, wert: { pfad: 'gemessen.quote', mindestens: 0.8 } }
  assert.deepEqual(pruefe(mitSchwelle, roh, heute), [])
  assert.match(pruefe({ ...mitSchwelle, wert: { pfad: 'gemessen.quote', mindestens: 0.9 } }, roh, heute).join(' '), /0\.82.*0\.9/)
})

test('ein Laufbeleg mit roter Rückgabe ist rot', () => {
  const roh = JSON.stringify({ commit: 'abc', ende: '2026-10-08T10:00:00Z', befehl: ['x'], rueckgabe: 1 })
  assert.match(pruefe({ schritt: 'x', beleg: 'x.json', fristTage: 30 }, roh, heute).join(' '), /Rückgabe 1/)
})

// **Die Köpfe der vorhandenen Belege, wie sie dastehen** — der Plan prüft sie,
// ohne dass jeder Läufer dieselbe Zeile schreiben muss.
test('die Köpfe vorhandener Belege werden gelesen', () => {
  const mutationen = '# Mutationslauf\n\n**12 von 12 gefangen** — Commit `2f82e6c3e` **mit 4 ungesicherten Änderung(en)**,\n2026-10-06 16:52 UTC, go version go1.27.1 darwin/arm64.\nGefahren mit `node scripts/mutationslauf.mjs`; gefangen ist'
  const last = '# Lastlauf\n\n**25 Fälle** — Commit `da5e77cf`, 2026-10-06 17:05:29 UTC, Apple M5, 10 Kerne, 32 GiB, macOS 27.0.1, go1.27.1 darwin/arm64.\nGefahren mit `scripts/lastlauf.mjs`: je Fall'
  const vergleich = '# Leistungsvergleich — Bericht\n\n**Stand `c42b1b4ef`**, Lauf ab 2026-10-05T02:09:56.832Z. Erzeugt von\n`scripts/leistungsvergleich.mjs`; die Zahlen'
  for (const [kopf, commit, tagesdatum] of [[mutationen, '2f82e6c3e', '2026-10-06'], [last, 'da5e77cf', '2026-10-06'], [vergleich, 'c42b1b4ef', '2026-10-05']]) {
    const angaben = angabenAus(kopf, 'md')
    assert.equal(angaben.commit, commit)
    assert.equal(angaben.datum.toISOString().slice(0, 10), tagesdatum)
    assert.match(angaben.befehl, /scripts\//)
  }
  assert.notEqual(angabenAus(mutationen, 'md').maschine, '')
  assert.notEqual(angabenAus(last, 'md').maschine, '')
  // **Der Leistungsvergleich nennt keine Maschine im Kopf** — und genau das
  // soll der Plan sagen.
  assert.match(pruefe({ schritt: 'Leistungsvergleich', beleg: 'leistungsvergleich.md', fristTage: 30 }, vergleich, heute).join(' '), /keine Maschine/)
})

test('ein Messwert mit Dezimalkomma wird gelesen', () => {
  const roh = huelle('2026-10-09').replace('| 120 |', '| 97,5 |')
  const eintrag2 = { ...eintrag, wert: { muster: 'Feld → Vorschaubild \\(ms\\) \\| \\d+ \\| [\\d.]+ \\| ([\\d,]+)', hoechstens: 150 } }
  assert.deepEqual(pruefe(eintrag2, roh, heute), [])
})

// **Ein Beleg, der selbst einen Befund meldet, trägt keine Zusage** — der
// erste Plan nahm einen dreiminütigen Probelauf des Dauerbetriebs mit einem
// roten Speicherurteil als gültigen Beleg für zwei Stunden.
test('ein Hüllenbeleg mit Befunden ist rot', () => {
  const rot = huelle('2026-10-09').replace('**bestanden**', '**2 Befund(e)**')
  assert.match(pruefe(eintrag, rot, heute).join(' '), /meldet 2 Befund/)
})

test('ein Hüllenbeleg aus einem Stand mit ungesicherten Änderungen ist rot', () => {
  const offen = huelle('2026-10-09').replace('Commit `ffadb0e5e`,', 'Commit `ffadb0e5e` **mit 3 ungesicherten Änderung(en)**,')
  assert.match(pruefe(eintrag, offen, heute).join(' '), /ungesichert/)
})

// **Ein Laufbeleg urteilt mit `status`, nicht mit der Rückgabe** — ein
// Trendfall wie `TestDerAbstandZuMSE2` scheitert und lässt `go test` mit 1
// enden, und der Beleg sagt trotzdem PASS (`internal/laufbeleg`). Der erste
// Plan las die Rückgabe und meldete den Fremdpaketlauf als rot.
const laufbeleg = (status, rueckgabe) => JSON.stringify({ commit: 'abc1234', ende: '2026-10-08T02:46:14+02:00', werkzeuge: { go: 'go1.27.2' }, befehl: ['go', 'test'], rueckgabe, status })
const laufeintrag = { schritt: 'Fremdpakete', beleg: 'fremdpakete.json', fristTage: 30 }

test('ein Laufbeleg mit PASS hält, auch wenn ein Trendfall die Rückgabe 1 gab', () => {
  assert.deepEqual(pruefe(laufeintrag, laufbeleg('PASS', 1), heute), [])
})

test('ein Laufbeleg mit FAIL ist rot, auch bei Rückgabe 0', () => {
  assert.match(pruefe(laufeintrag, laufbeleg('FAIL', 0), heute).join(' '), /FAIL/)
})

// **Was der Zweitblick über Phase B fand** (Befunde 8 und 9): eine fehlende
// Maschine im JSON, ein Hüllenbeleg ohne Urteil, ein Bündel aus einem anderen
// Stand und ein Eintrag mit mehreren Grenzen.
test('ein Laufbeleg ohne Maschine ist rot und nicht ein leerer Text', () => {
  const ohne = JSON.stringify({ commit: 'abc1234', ende: '2026-10-08T10:00:00Z', befehl: ['go', 'test'], rueckgabe: 0, status: 'PASS' })
  assert.match(pruefe(laufeintrag, ohne, heute).join(' '), /keine Maschine/)
})

test('ein Hüllenbeleg ohne Urteil ist rot und nicht befundfrei', () => {
  const ohneUrteil = huelle('2026-10-09').replace('**bestanden** — ', '')
  assert.match(pruefe(eintrag, ohneUrteil, heute).join(' '), /kein Urteil/)
})

test('ein Hüllenbeleg, dessen Bündel aus einem anderen Stand stammt, ist rot', () => {
  const veraltet = huelle('2026-10-09').replace('Bauabdruck `17c8bcbdc895`.', 'Bauabdruck `17c8bcbdc895` — **das Bündel stammt aus einem anderen Stand**.')
  assert.match(pruefe(eintrag, veraltet, heute).join(' '), /anderen Stand/)
})

test('ein Eintrag prüft jede seiner Grenzen', () => {
  const zwei = {
    ...eintrag,
    wert: undefined,
    werte: [eintrag.wert, { muster: 'Feld → Vorschaubild \\(ms\\) \\| \\d+ \\| ([\\d.]+)', hoechstens: 50 }],
  }
  assert.match(pruefe(zwei, huelle('2026-10-09'), heute).join(' '), /80 über der Schwelle 50/)
  assert.deepEqual(pruefe({ ...zwei, werte: [eintrag.wert] }, huelle('2026-10-09'), heute), [])
})

// **Und eine leere Struktur ist keine Maschine** — Nachlauf des Zweitblicks
// über Phase B: `{}` und `[]` wurden zu `"{}"` und `"[]"` und bestanden.
test('ein Laufbeleg mit leerer Maschinenangabe ist rot', () => {
  for (const leer of [{}, [], '', '   ']) {
    const roh = JSON.stringify({ commit: 'abc1234', ende: '2026-10-08T10:00:00Z', befehl: ['go', 'test'], rueckgabe: 0, status: 'PASS', werkzeuge: leer })
    assert.match(pruefe(laufeintrag, roh, heute).join(' '), /keine Maschine/, JSON.stringify(leer))
  }
})
