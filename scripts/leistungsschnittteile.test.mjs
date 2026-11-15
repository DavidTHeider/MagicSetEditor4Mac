// Prüfungen der Teile des Messvertrags — PERF-01
// (#2168).

import { test } from 'node:test'
import assert from 'node:assert/strict'

import { auswaehlen, beobachte, benchLesen, fremdLast, gestoert, warteAufRuhe, zuViel } from './leistungsschnittteile.mjs'

test('die Ausgabe von go test -bench wird je Name gelesen, mit Speicher', () => {
  const gelesen = benchLesen([
    'goos: darwin',
    'BenchmarkZerteileKaltUndWarm/kalt-10   \t   12345\t     98765 ns/op\t   4096 B/op\t      42 allocs/op',
    'BenchmarkZerteileKaltUndWarm/kalt-10   \t   12000\t    100000 ns/op\t   4100 B/op\t      42 allocs/op',
    'BenchmarkWegKopflos/Karte_je_Ruf-10    \t       3\t 733296181 ns/op',
    '--- BENCH: BenchmarkWegKopflos',
    'PASS',
  ].join('\n'))

  assert.deepEqual([...gelesen.keys()], ['BenchmarkZerteileKaltUndWarm/kalt', 'BenchmarkWegKopflos/Karte_je_Ruf'])
  assert.deepEqual(gelesen.get('BenchmarkZerteileKaltUndWarm/kalt'), {
    ns: [98765, 100000], bytes: [4096, 4100], zuteilungen: [42, 42], durchlaeufe: [12345, 12000],
  })
  assert.deepEqual(gelesen.get('BenchmarkWegKopflos/Karte_je_Ruf').bytes, [])
})

test('eigene Kennzahlen zwischen ns/op und B/op stören das Lesen nicht', () => {
  const gelesen = benchLesen('BenchmarkWegKopflos/Skript_je_Ruf-10   \t       5\t  13878700 ns/op\t        12.00 cpu-ms/op\t 5242880 B/op\t   51234 allocs/op')

  assert.deepEqual(gelesen.get('BenchmarkWegKopflos/Skript_je_Ruf'), {
    ns: [13878700], bytes: [5242880], zuteilungen: [51234], durchlaeufe: [5],
  })
})

test('die fremde Last zählt nur Prozesse außerhalb des eigenen Baums', () => {
  const ps = [
    '  100     1  91.1 /System/Library/StorageManagementService',
    '  200     1  20.0 /usr/bin/fremd',
    '  300   250  50.0 /tmp/app.test',
    '  250     1   1.0 node',
  ].join('\n')
  const last = fremdLast(ps, 250)

  assert.equal(last.summe, 111.1)
  assert.deepEqual(last.groesste.map((p) => p.pid), [100, 200])
})

const ruhig = { last: [1.2, 1.5, 1.6], fremd: { summe: 120, groesste: [] } }
const laut = { last: [9.5, 8, 7], fremd: { summe: 420, groesste: [] } }
const grenzen = { last: 5, fremd: 250 }

test('eine Umgebung unter beiden Grenzen ist ruhig', () => {
  assert.equal(zuViel(ruhig, grenzen), null)
  assert.match(zuViel(laut, grenzen), /Last 9\.5 über 5/)
  assert.match(zuViel({ last: [1, 1, 1], fremd: { summe: 300, groesste: [] } }, grenzen), /fremde Prozesse 300 % über 250 %/)
})

test('der Normalfall misst sofort, ohne zu warten', async () => {
  let geschlafen = 0
  const ergebnis = await warteAufRuhe({
    messen: () => ruhig, schlafe: async (s) => { geschlafen += s }, grenzen, frist: 60, takt: 10,
  })

  assert.equal(ergebnis.ruhig, true)
  assert.equal(geschlafen, 0)
})

test('hohe fremde Last pausiert, bis sie fällt', async () => {
  const folge = [laut, laut, ruhig]
  let geschlafen = 0
  const ergebnis = await warteAufRuhe({
    messen: () => folge.shift(), schlafe: async (s) => { geschlafen += s }, grenzen, frist: 60, takt: 10,
  })

  assert.equal(ergebnis.ruhig, true)
  assert.equal(ergebnis.gewartetS, 20)
  assert.equal(geschlafen, 20)
})

test('bleibt sie hoch, ist der Schnitt nicht gemessen — und der Grund steht da', async () => {
  const ergebnis = await warteAufRuhe({
    messen: () => laut, schlafe: async () => {}, grenzen, frist: 30, takt: 10,
  })

  assert.equal(ergebnis.ruhig, false)
  assert.equal(ergebnis.gewartetS, 30)
  assert.match(ergebnis.grund, /Last/)
})

// **Ein unbekannter Name misst nichts — und darf nicht „vollständig" heißen.**
// Gefunden am 2026-10-07: zsh teilt `$TEIL` nicht, der Läufer bekam einen
// einzigen Namen aus vier Wörtern, maß null Schnitte und meldete Erfolg.
test('die Auswahl weist unbekannte Namen ab und nimmt ohne Namen alle', () => {
  const schnitte = [{ name: 'a' }, { name: 'b' }]

  assert.deepEqual(auswaehlen(schnitte, []).map((s) => s.name), ['a', 'b'])
  assert.deepEqual(auswaehlen(schnitte, ['b']).map((s) => s.name), ['b'])
  assert.throws(() => auswaehlen(schnitte, ['a b']), /unbekannter Schnitt »a b«/)
})

// **Die Grenze gilt über den ganzen Schnitt, nicht nur bis zu seinem
// Anfang** — Befund 4 des Zweitblicks über Phase 0.7.0-I. Gezählt wird dabei
// nur die **fremde** Last: `loadavg` enthält die eigene Arbeit, und ein
// Schnitt, der zehn Kerne belegt, hebt sie selbst über die Grenze.
test('fremde Last während des Schnitts macht ihn ungültig, die eigene nicht', () => {
  const grenzen = { last: 5, fremd: 250 }
  const probe = (last, summe) => ({ last: [last, 0, 0], fremd: { summe, groesste: [{ name: 'fremd', cpu: summe }] } })

  assert.equal(gestoert([probe(9.5, 120), probe(8.1, 240)], grenzen), null)
  assert.equal(gestoert([probe(3, 120), probe(3, 310), probe(3, 90)], grenzen), 'fremde Prozesse während des Schnitts bis 310 % über 250 % (fremd 310)')
  assert.equal(gestoert([], grenzen), 'keine Probe während des Schnitts')
})

test('die Beobachtung nimmt Proben im Takt und eine letzte am Ende', async () => {
  let blick = 0
  const beobachtung = beobachte(() => {
    blick += 1

    return blick
  }, 5)
  await new Promise((los) => { setTimeout(los, 40) })
  const { proben, fehlgeschlagen } = beobachtung.ende()

  assert.ok(proben.length >= 3, `${String(proben.length)} Proben`)
  assert.equal(proben.at(-1), blick)
  assert.equal(fehlgeschlagen, 0)
})

// **Ein verpasster Blick ist kein Beleg für Ruhe** — Befund 3 des Nachlaufs
// über Phase I: fiel die Messung nach einer guten Probe aus, galt der Schnitt
// als gemessen, und die letzte gute stand als „nachher" im Beleg.
test('verpasste Proben werden gezählt, und dann ist die Grenze nicht belegt', async () => {
  let blick = 0
  const beobachtung = beobachte(() => {
    blick += 1
    if (blick > 1) throw new Error('ps antwortet nicht')

    return { last: [1, 0, 0], fremd: { summe: 100, groesste: [] } }
  }, 5)
  await new Promise((los) => { setTimeout(los, 25) })
  const { proben, fehlgeschlagen } = beobachtung.ende()

  assert.equal(proben.length, 1)
  assert.ok(fehlgeschlagen >= 2, `${String(fehlgeschlagen)} verpasst`)
  assert.equal(gestoert(proben, { last: 5, fremd: 250 }, fehlgeschlagen), `${String(fehlgeschlagen)} Proben ohne Antwort — dass die Grenze während des Schnitts hielt, ist nicht belegt`)
})
