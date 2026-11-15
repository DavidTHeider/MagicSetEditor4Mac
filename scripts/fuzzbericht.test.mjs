// Prüfungen von scripts/fuzzbericht.mjs — Z5 des Zweitblicks über Phase 6.11.
//
// **Jeder Teil einer Kampagne ist erforderlich**: der Fuzzlauf, die Übernahme
// des erarbeiteten Korpus und der Abdeckungslauf. Bis dahin prüfte die
// Kampagne nur den ersten — scheiterte die Instrumentierung nach einem grünen
// Fuzzlauf, stand „0 mit Fund“ mit leerem Anteil und Exitcode 0 im Bericht.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { existsSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { abdeckungAuswerten, anteilMessen, berichtText, dauerSekunden, fuzzlaufAuswerten, kampagnenstatus, saettigung, standPruefen, zieleAus, zielErgebnis, zieleWaehlen, zusammenfuehren } from './fuzzbericht.mjs'

const gruen = {
  status: 0,
  stdout: 'fuzz: elapsed: 3s, execs: 1200 (400/sec), new interesting: 2 (total: 7)\nfuzz: elapsed: 6s, execs: 2400 (400/sec), new interesting: 3 (total: 8)\nPASS\n',
  stderr: '',
}
const fund = {
  status: 1,
  stdout: 'fuzz: elapsed: 1s, execs: 100 (100/sec), new interesting: 0 (total: 5)\n--- FAIL: FuzzRahmen (0.50s)\n    --- FAIL: FuzzRahmen (0.00s)\n        rahmen_test.go:12: index out of range\n\n    Failing input written to testdata/fuzz/FuzzRahmen/abc123\n',
  stderr: '',
}

test('ein grüner Fuzzlauf nennt Ausführungen, neue Eingaben und Korpus aus der letzten Fortschrittszeile', () => {
  const f = fuzzlaufAuswerten(gruen)
  assert.deepEqual([f.ausgang, f.ausfuehrungen, f.neu, f.korpus], ['ohne Fund', 2400, 3, 8])
  assert.equal(f.fehler, null)
})

test('ein Fuzzlauf mit gespeicherter Eingabe ist ein Fund — mit Grund und Datei', () => {
  const f = fuzzlaufAuswerten(fund)
  assert.equal(f.ausgang, 'Fund')
  assert.equal(f.fund, 'testdata/fuzz/FuzzRahmen/abc123')
  assert.match(f.grund, /index out of range/)
})

test('ein Fuzzlauf, der ohne gespeicherte Eingabe scheitert, ist weder grün noch ein Fund', () => {
  const f = fuzzlaufAuswerten({ status: 2, stdout: '', stderr: 'internal/rahmen/x.go:3:1: syntax error\nFAIL\tinternal/rahmen [build failed]\n' })
  assert.equal(f.ausgang, 'gescheitert')
  assert.match(f.fehler, /Status 2/)
  assert.match(f.fehler, /build failed/, 'das Ende der Ausgabe steht dabei')
})

test('ein Fuzzlauf, den ein Signal beendet oder der nie startete, ist gescheitert', () => {
  assert.equal(fuzzlaufAuswerten({ status: null, signal: 'SIGKILL', stdout: '', stderr: '' }).ausgang, 'gescheitert')
  const nie = fuzzlaufAuswerten({ status: null, error: new Error('spawn go ENOENT'), stdout: null, stderr: null })
  assert.equal(nie.ausgang, 'gescheitert')
  assert.match(nie.fehler, /ENOENT/)
})

test('ein „grüner“ Lauf ohne Fuzzziel hat nichts gefuzzt', () => {
  const f = fuzzlaufAuswerten({ status: 0, stdout: 'testing: warning: no fuzz tests to fuzz\nPASS\n', stderr: '' })
  assert.equal(f.ausgang, 'gescheitert')
  assert.match(f.fehler, /kein Fuzzziel/)
})

const profil = 'ok  \tmodul/internal/rahmen\t0.4s\tcoverage: 61.5% of statements in ./internal/rahmen\n'

test('der Abdeckungslauf nennt den Anteil', () => {
  assert.deepEqual(abdeckungAuswerten({ status: 0, stdout: profil, stderr: '' }, false), { prozent: '61.5', fehler: null })
})

test('ein Abdeckungslauf mit Status ungleich null ist ein Fehler, auch wenn eine Zahl dasteht', () => {
  const a = abdeckungAuswerten({ status: 1, stdout: profil, stderr: 'panic: Instrumentierung\n' }, false)
  assert.match(a.fehler, /Abdeckungslauf endete mit Status 1/)
  assert.match(a.fehler, /Instrumentierung/)
})

test('ein Abdeckungslauf ohne Zahl ist ein Fehler — nicht ein Strich', () => {
  const a = abdeckungAuswerten({ status: 0, stdout: 'ok\tmodul/x\t0.1s\n', stderr: '' }, false)
  assert.equal(a.prozent, null)
  assert.match(a.fehler, /keine Abdeckungszahl/)
})

test('bei einem Fund darf der Abdeckungslauf scheitern: die gespeicherte Eingabe lässt ihn rot enden', () => {
  const a = abdeckungAuswerten({ status: 1, stdout: profil, stderr: '' }, true)
  assert.equal(a.fehler, null)
  assert.equal(a.prozent, '61.5')
  assert.equal(abdeckungAuswerten({ status: 1, stdout: '', stderr: '' }, true).prozent, null)
})

/** werkstatt legt ein Arbeitsverzeichnis mit eingechecktem und erarbeitetem Korpus an. */
function werkstatt({ erarbeitet = ['a', 'b'] } = {}) {
  const wurzel = mkdtempSync(join(tmpdir(), 'fuzzbericht-'))
  const zwischen = mkdtempSync(join(tmpdir(), 'fuzzcache-'))
  const z = { paket: 'internal/rahmen', name: 'FuzzRahmen' }
  const quelle = join(zwischen, 'fuzz', 'modul', z.paket, z.name)
  if (erarbeitet) {
    mkdirSync(quelle, { recursive: true })
    for (const datei of erarbeitet) writeFileSync(join(quelle, datei), `eingabe ${datei}`)
  }
  mkdirSync(join(wurzel, z.paket, 'testdata', 'fuzz', z.name), { recursive: true })
  writeFileSync(join(wurzel, z.paket, 'testdata', 'fuzz', z.name, 'saat'), 'saat')
  return { wurzel, zwischen, ziel: z, eingecheckt: join(wurzel, z.paket, 'testdata', 'fuzz', z.name) }
}

test('der erarbeitete Korpus liegt während des Abdeckungslaufs im Baum — und danach nicht mehr', () => {
  const w = werkstatt()
  let waehrend
  const lauf = anteilMessen(w.ziel, {
    wurzel: w.wurzel, zwischen: w.zwischen, modul: 'modul', fund: false,
    goTest: () => { waehrend = readdirSync(w.eingecheckt).sort(); return { status: 0, stdout: profil, stderr: '' } },
  })
  assert.deepEqual(waehrend, ['kampagne-a', 'kampagne-b', 'saat'])
  assert.deepEqual(readdirSync(w.eingecheckt), ['saat'], 'aufgeräumt wird auch ohne Fehler')
  assert.deepEqual([lauf.prozent, lauf.korpusdateien, lauf.fehler], ['61.5', 2, []])
})

test('scheitert der Abdeckungslauf, steht das als Fehler da und der Anteil bleibt leer', () => {
  const w = werkstatt()
  const lauf = anteilMessen(w.ziel, {
    wurzel: w.wurzel, zwischen: w.zwischen, modul: 'modul', fund: false,
    goTest: () => ({ status: 1, stdout: '', stderr: 'build failed' }),
  })
  assert.equal(lauf.prozent, null)
  assert.equal(lauf.fehler.length, 1)
  assert.match(lauf.fehler[0], /^Abdeckungslauf/)
  assert.deepEqual(readdirSync(w.eingecheckt), ['saat'], 'aufgeräumt wird auch im Fehlerfall')
})

test('scheitert die Übernahme des Korpus, ist das ein eigener Fehler — und der Abdeckungslauf geht trotzdem, damit der Rest sichtbar bleibt', () => {
  const w = werkstatt()
  // Ein Verzeichnis im Cache kann `copyFileSync` nicht kopieren.
  mkdirSync(join(w.zwischen, 'fuzz', 'modul', w.ziel.paket, w.ziel.name, 'ordner'))
  let gerufen = 0
  const lauf = anteilMessen(w.ziel, {
    wurzel: w.wurzel, zwischen: w.zwischen, modul: 'modul', fund: false,
    goTest: () => { gerufen++; return { status: 0, stdout: profil, stderr: '' } },
  })
  assert.equal(gerufen, 1)
  assert.equal(lauf.fehler.length, 1)
  assert.match(lauf.fehler[0], /^Korpusübernahme/)
  assert.deepEqual(readdirSync(w.eingecheckt), ['saat'], 'was schon kopiert war, ist wieder weg')
})

test('ohne erarbeiteten Korpus gibt es nichts zu übernehmen — das ist kein Fehler', () => {
  const w = werkstatt({ erarbeitet: null })
  const lauf = anteilMessen(w.ziel, {
    wurzel: w.wurzel, zwischen: w.zwischen, modul: 'modul', fund: false,
    goTest: () => ({ status: 0, stdout: profil, stderr: '' }),
  })
  assert.deepEqual([lauf.korpusdateien, lauf.fehler], [0, []])
  assert.ok(existsSync(w.eingecheckt))
})

/** ergebnis baut den Eintrag eines Ziels aus Fuzzlauf und Anteil. */
const ergebnis = (name, f, anteil) => zielErgebnis({ paket: 'internal/x', name }, f, anteil, 60)

test('ein Ziel ist unvollständig, sobald ein Teil fehlt — auch wenn der Fuzzlauf grün war', () => {
  const e = ergebnis('FuzzA', fuzzlaufAuswerten(gruen), { prozent: null, korpusdateien: 0, fehler: ['Abdeckungslauf endete mit Status 1'] })
  assert.equal(e.ausgang, 'unvollständig')
  assert.deepEqual(e.fehler, ['Abdeckungslauf endete mit Status 1'])
  assert.equal(e.anteil, '–')
})

test('ein Fund bleibt ein Fund, auch wenn der Abdeckungslauf daneben scheiterte', () => {
  const e = ergebnis('FuzzA', fuzzlaufAuswerten(fund), { prozent: null, korpusdateien: 0, fehler: [] })
  assert.equal(e.ausgang, 'Fund')
})

test('ein gescheiterter Fuzzlauf macht das Ziel unvollständig, mit seinem Grund an erster Stelle', () => {
  const e = ergebnis('FuzzA', fuzzlaufAuswerten({ status: 2, stdout: '', stderr: 'build failed' }), { prozent: '40.0', korpusdateien: 1, fehler: [] })
  assert.equal(e.ausgang, 'unvollständig')
  assert.match(e.fehler[0], /Fuzzlauf endete mit Status 2/)
})

test('ein vollständiges Ziel ohne Fund ist ohne Fund', () => {
  const e = ergebnis('FuzzA', fuzzlaufAuswerten(gruen), { prozent: '61.5', korpusdateien: 2, fehler: [] })
  assert.equal(e.ausgang, 'ohne Fund')
  assert.equal(e.anteil, '61.5')
})

test('der Status der Kampagne: grün, mit Fund (1) oder unvollständig (3) — der Fund wiegt schwerer', () => {
  const ok = ergebnis('A', fuzzlaufAuswerten(gruen), { prozent: '1', korpusdateien: 0, fehler: [] })
  const mitFund = ergebnis('B', fuzzlaufAuswerten(fund), { prozent: '1', korpusdateien: 0, fehler: [] })
  const luecke = ergebnis('C', fuzzlaufAuswerten(gruen), { prozent: null, korpusdateien: 0, fehler: ['x'] })
  assert.deepEqual(kampagnenstatus([ok]), { funde: [], unvollstaendig: [], exitcode: 0 })
  assert.equal(kampagnenstatus([ok, luecke]).exitcode, 3)
  assert.equal(kampagnenstatus([ok, mitFund]).exitcode, 1)
  assert.equal(kampagnenstatus([luecke, mitFund]).exitcode, 1)
  assert.equal(kampagnenstatus([ok, luecke, mitFund]).unvollstaendig.length, 1)
})

test('der Bericht nennt die Zahl der Ziele mit Fund und die der unvollständigen — und beim Ziel, was fehlte', () => {
  const luecke = ergebnis('FuzzC', fuzzlaufAuswerten(gruen), { prozent: null, korpusdateien: 0, fehler: ['Abdeckungslauf endete mit Status 1 | roh'] })
  const ok = ergebnis('FuzzA', fuzzlaufAuswerten(gruen), { prozent: '61.5', korpusdateien: 2, fehler: [] })
  const text = berichtText([ok, luecke], { commit: 'abc', zeit: '2026-10-03 12:00', go: 'go1.27.1', budget: '1m', parallel: '4', speicher: '2GiB' })
  assert.match(text, /\*\*2 Ziele, 0 mit Fund, 1 unvollständig\*\*/)
  assert.match(text, /\| `FuzzA` .* \| 61\.5 % \| ohne Fund \|/)
  assert.match(text, /\| `FuzzC` .* \| – \| \*\*unvollständig\*\*: Abdeckungslauf endete mit Status 1 \\\| roh \|/)
  assert.doesNotMatch(text, /\| – % \|/, 'ein Strich trägt kein Prozentzeichen')
})

// **Ein angeforderter Name, den es nicht gibt, wird genannt** — N4 des
// Nachlaufs über die QA-Arbeit: `fuzzkampagne.mjs FuzzBekannt FuzzVertippt`
// fuhr das bekannte Ziel, überging das vertippte und endete mit 0.
test('die Auswahl der Ziele nennt jeden unbekannten Namen', () => {
  const alle = [{ paket: 'internal/a', name: 'FuzzA' }, { paket: 'internal/b', name: 'FuzzB' }]
  assert.deepEqual(zieleWaehlen(alle, []), { ziele: alle, unbekannt: [] })
  assert.deepEqual(zieleWaehlen(alle, ['FuzzB']), { ziele: [alle[1]], unbekannt: [] })
  assert.deepEqual(zieleWaehlen(alle, ['FuzzB', 'FuzzVertippt']), { ziele: [alle[1]], unbekannt: ['FuzzVertippt'] })
})


// **Zwei Hälften einer Kampagne werden ein Bericht** — C02 (#2227): die
// Kampagne fährt in zwei Prozessen über getrennte Ziele, damit 55 Ziele zu je
// zehn Minuten nicht neun Stunden brauchen. Zusammen geht nur, was vom selben
// Stand und mit demselben Budget kommt.
test('zwei Hälften vom selben Stand werden ein Bericht, in der Reihenfolge der Ziele', () => {
  const kopf = { commit: 'abc1234', zeit: '2026-10-10 06:00', go: 'go1.27.2', budget: '10m', parallel: '5', speicher: '2GiB' }
  const e = (name) => ({ name, paket: 'internal/x', sekunden: 600, ausfuehrungen: 1, neu: 0, korpus: 1, anteil: '1', ausgang: 'ohne Fund', fehler: [] })
  const ganz = zusammenfuehren(
    [{ ergebnisse: [e('FuzzB')], kopf }, { ergebnisse: [e('FuzzA'), e('FuzzC')], kopf: { ...kopf, zeit: '2026-10-10 07:00' } }],
    ['FuzzA', 'FuzzB', 'FuzzC'].map((name) => ({ paket: 'internal/x', name })),
  )
  assert.deepEqual(ganz.ergebnisse.map((x) => x.name), ['FuzzA', 'FuzzB', 'FuzzC'])
  assert.equal(ganz.kopf.zeit, '2026-10-10 07:00')
  assert.equal(ganz.kopf.commit, 'abc1234')
})

test('Hälften von verschiedenen Ständen, mit anderem Budget oder doppelten Zielen gehen nicht zusammen', () => {
  const kopf = { commit: 'abc1234', zeit: 'x', go: 'go1.27.2', budget: '10m', parallel: '5', speicher: '2GiB' }
  const e = (name) => ({ name, paket: 'internal/x', ausgang: 'ohne Fund', fehler: [] })
  const ab = [{ paket: 'internal/x', name: 'FuzzA' }, { paket: 'internal/x', name: 'FuzzB' }]
  assert.throws(() => zusammenfuehren([{ ergebnisse: [e('FuzzA')], kopf }, { ergebnisse: [e('FuzzB')], kopf: { ...kopf, commit: 'def5678' } }], ab), /Ständen/)
  assert.throws(() => zusammenfuehren([{ ergebnisse: [e('FuzzA')], kopf }, { ergebnisse: [e('FuzzB')], kopf: { ...kopf, budget: '1m' } }], ab), /Budget/)
  assert.throws(() => zusammenfuehren([{ ergebnisse: [e('FuzzA')], kopf }, { ergebnisse: [e('FuzzA')], kopf }], ab.slice(0, 1)), /doppelt/)
})

// **Ein Ziel ist Paket und Name** — Zweitblick über Phase C, Befund 1: die 60
// Ziele aus `RELEASE.md` haben nur 55 Namen (`FuzzModelle` viermal,
// `FuzzLies` und `FuzzLiesKarten` je zweimal), und die Zusammenführung wies
// zwei gültige Hälften als „doppelt" ab.
test('gleichnamige Ziele aus verschiedenen Paketen gehen zusammen, in der Reihenfolge der Ziele', () => {
  const kopf = { commit: 'abc1234', zeit: 'x', go: 'go1.27.2', budget: '10m', parallel: '5', speicher: '2GiB' }
  const e = (paket, name) => ({ name, paket, ausgang: 'ohne Fund', fehler: [] })
  const ziele = [{ paket: 'internal/ollama', name: 'FuzzModelle' }, { paket: 'internal/x', name: 'FuzzA' }, { paket: 'internal/openai', name: 'FuzzModelle' }]
  const ganz = zusammenfuehren([
    { ergebnisse: [e('internal/openai', 'FuzzModelle')], kopf },
    { ergebnisse: [e('internal/x', 'FuzzA'), e('internal/ollama', 'FuzzModelle')], kopf },
  ], ziele)
  assert.deepEqual(ganz.ergebnisse.map((x) => `${x.paket} ${x.name}`), ['internal/ollama FuzzModelle', 'internal/x FuzzA', 'internal/openai FuzzModelle'])
})

// **Eine fehlende Hälfte ist keine Kampagne** — Befund 2: ein einziges Ziel
// ergab einen zusammengeführten Bericht mit Exitcode 0.
test('ein fehlendes oder unbekanntes Ziel wird beim Zusammenführen genannt und abgewiesen', () => {
  const kopf = { commit: 'abc1234', zeit: 'x', go: 'go1.27.2', budget: '10m', parallel: '5', speicher: '2GiB' }
  const e = (name) => ({ name, paket: 'internal/x', ausgang: 'ohne Fund', fehler: [] })
  const ziele = ['FuzzA', 'FuzzB', 'FuzzC'].map((name) => ({ paket: 'internal/x', name }))
  assert.throws(() => zusammenfuehren([{ ergebnisse: [e('FuzzA')], kopf }, { ergebnisse: [e('FuzzC')], kopf }], ziele), /es fehlt: internal\/x FuzzB/)
  assert.throws(() => zusammenfuehren([{ ergebnisse: [e('FuzzA'), e('FuzzB')], kopf }, { ergebnisse: [e('FuzzC'), e('FuzzD')], kopf }], ziele), /unbekannt: internal\/x FuzzD/)
})

// **Sättigung braucht den Verlauf** — Befund 4: aus der letzten
// Fortschrittszeile lässt sich nicht ablesen, wann zuletzt eine neue Eingabe
// kam.
test('der Fuzzlauf behält seinen Verlauf und nennt, wann zuletzt eine neue Eingabe kam', () => {
  const zeilen = [
    'fuzz: elapsed: 0s, gathering baseline coverage: 0/7 completed',
    'fuzz: elapsed: 3s, execs: 1200 (400/sec), new interesting: 2 (total: 9)',
    'fuzz: elapsed: 1m0s, execs: 24000 (400/sec), new interesting: 5 (total: 12)',
    'fuzz: elapsed: 7m30s, execs: 180000 (400/sec), new interesting: 5 (total: 12)',
    'fuzz: elapsed: 10m0s, execs: 240000 (400/sec), new interesting: 5 (total: 12)',
  ]
  const f = fuzzlaufAuswerten({ status: 0, stdout: `${zeilen.join('\n')}\nPASS\n`, stderr: '' })
  assert.deepEqual(f.verlauf, [[3, 1200, 2], [60, 24000, 5], [450, 180000, 5], [600, 240000, 5]])
  assert.equal(f.zuletztNeu, 60)
  assert.equal(f.ausfuehrungen, 240000)
})

test('Go-Dauern werden zu Sekunden', () => {
  assert.equal(dauerSekunden('3s'), 3)
  assert.equal(dauerSekunden('1m0s'), 60)
  assert.equal(dauerSekunden('1h2m3s'), 3723)
  assert.equal(dauerSekunden('10m'), 600)
})

// **Gesättigt heißt hier: im letzten Viertel des Budgets kam nichts Neues** —
// ein Maß, das festgelegt ist, bevor die Zahlen da sind, und kein Beweis,
// dass nichts mehr zu finden wäre.
test('gesättigt ist ein Ziel, das im letzten Viertel seiner Zeit nichts Neues fand', () => {
  assert.equal(saettigung({ zuletztNeu: 60, verlauf: [[600, 1, 5]] }, '10m'), 'gesättigt')
  assert.equal(saettigung({ zuletztNeu: 450, verlauf: [[600, 1, 5]] }, '10m'), 'gesättigt')
  assert.equal(saettigung({ zuletztNeu: 451, verlauf: [[600, 1, 5]] }, '10m'), 'wächst noch')
  assert.equal(saettigung({ zuletztNeu: null, verlauf: [] }, '10m'), 'ohne Verlauf')
})

// **Sättigung nur bei voll beobachtetem Budget** — Nachlauf über Phase C:
// `FuzzLiesInhalt` endete nach 198 s mit einem Fund und galt als
// „gesättigt“. Ein Lauf, der vorher endet, sagt darüber nichts.
test('ein vorzeitig beendeter Lauf ist nicht bewertbar', () => {
  assert.equal(saettigung({ ausgang: 'Fund', zuletztNeu: 60, verlauf: [[198, 1, 5]] }, '10m'), 'nicht bewertbar')
  assert.equal(saettigung({ ausgang: 'ohne Fund', zuletztNeu: 60, verlauf: [[300, 1, 5]] }, '10m'), 'nicht bewertbar')
  // **Die letzte Meldung kommt alle drei Sekunden** — 597 s sind ein voller Lauf.
  assert.equal(saettigung({ ausgang: 'ohne Fund', zuletztNeu: 60, verlauf: [[597, 1, 5]] }, '10m'), 'gesättigt')
})

// **Der Stand gilt für den ganzen Lauf** — Befund 4: der Commit wurde erst am
// Ende gelesen, und ungesicherte Änderungen zählten gar nicht.
test('der Stand vor und nach dem Lauf wird verglichen, mit den ungesicherten Änderungen', () => {
  const stand = (commit, offen, inhalt = 'i1') => ({ commit, offen, inhalt, saat: 's1' })
  assert.deepEqual(standPruefen(stand('abc', 0), stand('abc', 0)), { commit: 'abc', geaendert: false, offen: 0, inhalt: 'i1', saat: 's1' })
  assert.deepEqual(standPruefen(stand('abc', 0), stand('def', 0)), { commit: 'abc', geaendert: true, offen: 0, inhalt: 'i1', saat: 's1' })
  assert.deepEqual(standPruefen(stand('abc', 2), stand('abc', 1)), { commit: 'abc', geaendert: false, offen: 2, inhalt: 'i1', saat: 's1' })
})

// **Der Inhalt zählt, nicht nur die Zahl der Dateien** — Nachlauf über
// Phase C: eine Änderung an einer schon geänderten Datei ließ Commit und
// Zahl stehen. Die Saat steht im Kopf; verglichen wird sie nicht, denn ein
// Fund legt seine Eingabe selbst dort ab.
test('ein geänderter Inhalt bei gleicher Zahl ist ein geänderter Stand', () => {
  const anfang = { commit: 'abc', offen: 1, inhalt: 'i1', saat: 's1' }
  assert.equal(standPruefen(anfang, { ...anfang, inhalt: 'i2' }).geaendert, true)
  assert.equal(standPruefen(anfang, { ...anfang, saat: 's2' }).geaendert, false)
})

test('Hälften, deren Stand sich während des Laufs änderte, gehen nicht zusammen', () => {
  const kopf = { commit: 'abc1234', zeit: 'x', go: 'go1.27.2', budget: '10m', parallel: '5', speicher: '2GiB', geaendert: false, offen: 0 }
  const e = (name) => ({ name, paket: 'internal/x', ausgang: 'ohne Fund', fehler: [] })
  const ziele = ['FuzzA', 'FuzzB'].map((name) => ({ paket: 'internal/x', name }))
  assert.throws(() => zusammenfuehren([{ ergebnisse: [e('FuzzA')], kopf }, { ergebnisse: [e('FuzzB')], kopf: { ...kopf, geaendert: true } }], ziele), /während des Laufs/)
})

// **Derselbe Commit heißt nicht derselbe Quelltext** — J05 (#2269), Befund C02:
// zwei Hälften, die je für sich stabil liefen, aber mit verschiedenen
// ungesicherten Änderungen am selben Commit, wurden zusammengeführt, und der
// Kopf trug den Abdruck nur der ersten. Der Bericht sähe aus wie einer über
// einen Stand und wäre keiner. `saat` bleibt ausgenommen: ein Fund legt seine
// Eingabe selbst dort ab, und die spätere Hälfte beginnt mit einem anderen
// Korpus.
test('Hälften vom selben Commit mit verschiedenem Inhalt gehen nicht zusammen', () => {
  const kopf = { commit: 'abc1234', zeit: 'x', go: 'go1.27.2', budget: '10m', parallel: '5', speicher: '2GiB', geaendert: false, offen: 1, inhalt: 'i1', saat: 's1' }
  const e = (name) => ({ name, paket: 'internal/x', ausgang: 'ohne Fund', fehler: [] })
  const ziele = ['FuzzA', 'FuzzB'].map((name) => ({ paket: 'internal/x', name }))
  const teile = (zweiter) => [{ ergebnisse: [e('FuzzA')], kopf }, { ergebnisse: [e('FuzzB')], kopf: zweiter }]

  assert.throws(() => zusammenfuehren(teile({ ...kopf, inhalt: 'i2' }), ziele), /verschiedenen Inhalt.*i1 und i2/)
  assert.throws(() => zusammenfuehren(teile({ ...kopf, inhalt: undefined }), ziele), /verschiedenen Inhalt/)

  // Die Gegenprobe zur Gegenprobe: gleicher Inhalt geht, und der Kopf nennt ihn.
  // Ein anderer Korpus allein ist kein Grund.
  assert.equal(zusammenfuehren(teile(kopf), ziele).kopf.inhalt, 'i1')
  assert.equal(zusammenfuehren(teile({ ...kopf, saat: 's2' }), ziele).kopf.inhalt, 'i1')
})

test('der Bericht nennt Sättigung und einen unsauberen Stand', () => {
  const ok = { ...ergebnis('FuzzA', fuzzlaufAuswerten(gruen), { prozent: '61.5', korpusdateien: 2, fehler: [] }), zuletztNeu: 6, verlauf: [[6, 2400, 3], [597, 240000, 3]] }
  const text = berichtText([ok], { commit: 'abc', zeit: 'x', go: 'go', budget: '10m', parallel: '5', speicher: '2GiB', geaendert: false, offen: 3 })
  assert.match(text, /\| 6 s \| gesättigt \|/)
  assert.match(text, /\*\*3 ungesicherte Änderung\(en\)\*\*/)
})

// **Am echten Register** — Zweitblick über Phase C, Befund 1: die Probe mit
// erfundenen Namen hatte die doppelten nicht. Zwei Hälften aus den Zielen in
// `RELEASE.md`, so geteilt wie die Kampagne sie fährt — gleichnamige Ziele
// in derselben Hälfte —, gehen zusammen.
test('zwei Hälften aus dem echten Register gehen zusammen', () => {
  const ziele = zieleAus(readFileSync(new URL('../docs/RELEASE.md', import.meta.url), 'utf8'))
  const namen = [...new Set(ziele.map((z) => z.name))]
  assert.ok(ziele.length > namen.length, `das Register hat ${ziele.length} Ziele und ${namen.length} Namen — die Probe braucht doppelte`)
  const vorn = new Set(namen.slice(0, Math.floor(namen.length / 2)))
  const kopf = { commit: 'abc1234', zeit: 'x', go: 'go1.27.2', budget: '10m', parallel: '5', speicher: '2GiB', geaendert: false, offen: 0 }
  const e = (z) => ({ ...z, ausgang: 'ohne Fund', fehler: [] })
  const ganz = zusammenfuehren([
    { ergebnisse: ziele.filter((z) => !vorn.has(z.name)).map(e), kopf },
    { ergebnisse: ziele.filter((z) => vorn.has(z.name)).map(e), kopf },
  ], ziele)
  assert.equal(ganz.ergebnisse.length, ziele.length)
  assert.deepEqual(ganz.ergebnisse.map((x) => `${x.paket} ${x.name}`), ziele.map((z) => `${z.paket} ${z.name}`))
})

// **Ein sauberer Baum heißt sauber** — die volle Kampagne vom 2026-10-11
// nannte `Inhalt e3b0c44298fc`: den Abdruck der leeren Zeichenkette, denn
// `inhalt` gilt dem Unterschied zum Commit. Richtig, aber nicht lesbar.
test('ein sauberer Stand nennt sich sauber und keinen leeren Abdruck', () => {
  const ok = { ...ergebnis('FuzzA', fuzzlaufAuswerten(gruen), { prozent: '61.5', korpusdateien: 2, fehler: [] }), zuletztNeu: 6, verlauf: [[6, 2400, 3], [597, 240000, 3]] }
  const kopf = { commit: 'abc', zeit: 'x', go: 'go', budget: '10m', parallel: '5', speicher: '2GiB', geaendert: false, offen: 0, inhalt: 'e3b0c44298fc', saat: 's1' }
  const text = berichtText([ok], kopf)
  assert.match(text, /Commit `abc` \(sauber gegen den Commit, Saat `s1`\)/)
  assert.doesNotMatch(text, /e3b0c44298fc/)

  const offen = berichtText([ok], { ...kopf, offen: 2, inhalt: 'i9' })
  assert.match(offen, /\(Inhalt `i9`, Saat `s1`\)/)
})
