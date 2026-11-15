// Prüfungen der Bewertung des Dauerbetriebs (B04, #2223).
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { abschnitte, aufteilen, bewertung, enthaelt, freigegeben, fussabdruckAus } from './huellenprobe/dauerbetrieb.mjs'

// reihe legt je Wert einen ganzen Abschnitt an — **nach der Aufwärmstunde**
// (Vermerk 0204).
const reihe = (werte) => werte.flatMap((wert, abschnitt) => Array.from({ length: 40 }, (_, i) => ({ minute: 60 + abschnitt * 20 + i * 0.5, gesamt: wert })))

test('die Aufwärmzeit zählt nicht, danach je 20 Minuten ein Median', () => {
  const proben = [{ minute: 5, gesamt: 9999 }, ...reihe([100, 101, 100])]
  assert.deepEqual(abschnitte(proben).map((t) => t.median), [100, 101, 100])
  assert.deepEqual(abschnitte(proben).map((t) => t.vollstaendig), [true, true, true])
})

// **Ein angebrochener Abschnitt ist kein Abschnitt** — Zweitblick über Phase B,
// Befund 7: der 125-Minuten-Lauf endete fünf Minuten in seinen sechsten, und
// der zählte wie ein ganzer.
test('ein angebrochener letzter Abschnitt zählt weder für die Rate noch für den Anstieg', () => {
  const angebrochen = Array.from({ length: 10 }, (_, i) => ({ minute: 120 + i * 0.5, gesamt: 900 }))
  const urteil = bewertung([...reihe([500, 501, 499]), ...angebrochen])
  assert.deepEqual(urteil.unvollstaendig, [3])
  assert.equal(urteil.haelt, true)
  assert.ok(Math.abs(urteil.wachstumJeStunde) < 0.01)
})

// **Die Rate kommt aus der Zeit und nicht aus der Zahl der Abschnitte**, und
// ein stetiger Anstieg verlangt drei **benachbarte**.
test('eine Lücke zwischen Abschnitten verbindet sie nicht zu einem Anstieg', () => {
  const ohneDritten = reihe([500, 510, 9999, 520]).filter((x) => x.minute < 100 || x.minute >= 120)
  const urteil = bewertung(ohneDritten)
  assert.deepEqual(urteil.luecken, [2])
  assert.equal(urteil.stetig, false)
  // 500 → 520 über 60 Minuten von Mitte zu Mitte: 4 % je Stunde.
  assert.ok(Math.abs(urteil.wachstumJeStunde - 0.04) < 0.002, String(urteil.wachstumJeStunde))
})

test('flacher Speicher hält', () => {
  const urteil = bewertung(reihe([500, 505, 498, 502, 501]))
  assert.equal(urteil.haelt, true)
})

// **Ein Anstieg zählt erst ab 1 % über dem vorigen Abschnitt** (Vermerk 0204,
// #2527): das Restwachstum der Anwendung aus #2312 liegt bei rund 0,6 % je
// Abschnitt und brach die Regel bis dahin allein.
test('ein Anstieg unter 1 % je Abschnitt ist kein Trend — das Restwachstum aus #2312', () => {
  const urteil = bewertung(reihe([550, 553.3, 556.6]))
  assert.equal(urteil.stetig, false)
  assert.equal(urteil.haelt, true)
})

test('ab 1 % je Abschnitt über drei Abschnitte ist er einer', () => {
  const urteil = bewertung(reihe([500, 505, 510.05]))
  assert.equal(urteil.stetig, true)
  assert.equal(urteil.haelt, false)
})

// **Die erste Stunde zählt nicht** — in ihr wärmt WebKit auf (zweimal gemessen,
// Vermerk 0204).
test('was in der ersten Stunde steigt, zählt nicht', () => {
  const aufwaermen = Array.from({ length: 80 }, (_, i) => ({ minute: 20 + i * 0.5, gesamt: 400 + i * 2 }))
  const urteil = bewertung([...aufwaermen, ...reihe([600, 601, 600])])
  assert.equal(urteil.gueltig, true)
  assert.deepEqual(urteil.teile, [600, 601, 600])
  assert.equal(urteil.haelt, true)
})

// **Was wächst, muss freigebbar sein** (Vermerk 0204): nach Ruhe und
// Speicherwarnung liegt WebKit unter seinem Median im ersten Abschnitt nach
// der Aufwärmstunde.
test('nach der Speicherwarnung liegt WebKit unter dem ersten Abschnitt — freigegeben', () => {
  const proben = [...reihe([600, 610, 605]).map((x) => ({ ...x, web: x.gesamt - 100 })), { minute: 127, aufgabe: 'nach der Speicherwarnung', web: 300, gesamt: 400 }]
  assert.deepEqual(freigegeben(proben), { gueltig: true, vorher: 500, nachher: 300, haelt: true })
})

test('gibt WebKit nach der Warnung nichts frei, hält es nicht — und ohne Warnung gibt es kein Urteil', () => {
  const proben = reihe([600, 610, 605]).map((x) => ({ ...x, web: x.gesamt - 100 }))
  assert.equal(freigegeben([...proben, { minute: 127, aufgabe: 'nach der Speicherwarnung', web: 520, gesamt: 620 }]).haelt, false)
  assert.equal(freigegeben(proben).gueltig, false)
})

test('mehr als 10 % je Stunde hält nicht', () => {
  const urteil = bewertung(reihe([500, 480, 600, 560, 700]))
  assert.ok(urteil.wachstumJeStunde > 0.1)
  assert.equal(urteil.haelt, false)
})

test('weniger als drei Abschnitte sind kein Urteil', () => {
  assert.equal(bewertung(reihe([500, 501])).gueltig, false)
})

// **Der physische Fußabdruck statt RSS** (E06, #2278): RSS zählt unter macOS
// Seiten mit, die Go schon zurückgegeben hat, und den Grafikspeicher des
// Webinhalts nicht. `footprint -f bytes --noCategories` nennt je Prozess
// eine Zeile; gelesen wird sie je Kennung, in MB.
test('liest den Fußabdruck je Prozess aus der Ausgabe von footprint', () => {
  const ausgabe = [
    '======================================================================',
    'MagicSetEditor4Mac [8063]: 64-bit    Footprint: 92274688 B (16384 bytes per page)',
    '======================================================================',
    '',
    'Auxiliary data:',
    '    phys_footprint: 92274688 B',
    '======================================================================',
    'com.apple.WebKit.WebContent [8079]: 64-bit    Footprint: 566231040 B (16384 bytes per page)',
  ].join('\n')
  assert.deepEqual(fussabdruckAus(ausgabe), new Map([['8063', 88], ['8079', 540]]))
})

test('ein Prozess ohne Zeile fehlt in der Antwort, statt null zu sein', () => {
  assert.deepEqual(fussabdruckAus('footprint: no process with pid 1'), new Map())
})

// **Eine fehlende Messung ist keine Null** — Zweitblick über die Folgen vor
// 0.8.0, Befund 12: `footprint` nannte einen Prozess nicht, und die Summe
// zählte ihn mit 0. Am Ende des Laufs sähe das wie ein sinkender Speicher aus.
test('eine Probe ohne vollständige Messwerte macht das Urteil ungültig', () => {
  const proben = reihe([500, 500, 500])
  proben[50].gesamt = null
  const urteil = bewertung(proben)
  assert.equal(urteil.gueltig, false)
  assert.match(urteil.grund, /1 Probe/)
})

// **Und fehlt ein Prozess der Anwendung, gilt das Urteil nicht** — ein neu
// gestarteter Webinhalt ist ein anderer Prozess mit anderem Speicher.
test('ein Wechsel der eigenen WebKit-Prozesse macht das Urteil ungültig', () => {
  const proben = reihe([500, 500, 500]).map((x) => ({ ...x, webPids: ['1', '2'] }))
  proben[70].webPids = ['1', '3']
  const urteil = bewertung(proben)
  assert.equal(urteil.gueltig, false)
  assert.match(urteil.grund, /WebKit/)
})

// **Ein WebKit-Prozess, der erst später kommt, zählt nicht mit** — im Lauf
// vom 2026-10-12 startete das System etwa alle 35 Minuten einen
// `WebContent.EnhancedSecurity`, der die Anwendung überlebte.
test('teilt WebKit-Prozesse in die der Anwendung und fremde', () => {
  const alle = [{ pid: '1' }, { pid: '2' }, { pid: '9' }]
  const { eigen, fremd } = aufteilen(alle, new Set(['1', '2', '3']))
  assert.deepEqual(eigen.map((x) => x.pid), ['1', '2'])
  assert.deepEqual(fremd.map((x) => x.pid), ['9'])
})

// **Eine Probe misst den Teil der Runde, in den sie fällt** (#2527): im Lauf
// vom 2026-10-20 lagen die Mediane nach Teil zwischen 676 MiB (Bearbeiten)
// und 893 MiB (Tisch), und weil eine Runde rund fünf Sekunden dauert, traf
// die Probe alle 30 s über lange Strecken denselben Teil. Abschnitte aus
// verschiedenen Teilen vergleichen verschiedene Zustände, nicht Zeiten.
test('Proben aus verschiedenen Teilen der Runde machen das Urteil ungültig', () => {
  const proben = reihe([500, 500, 500]).map((x, i) => ({ ...x, aufgabe: `Dauer — Runde ${String(i)}: ${i % 3 === 0 ? 'tisch' : 'ausgabe'}` }))
  const urteil = bewertung(proben)
  assert.equal(urteil.gueltig, false)
  assert.match(urteil.grund, /Teilen der Runde/)
  assert.match(urteil.grund, /ausgabe 80, tisch 40/)
})

test('Proben am selben Punkt der Runde lassen das Urteil gelten — Anfang und Ende zählen nicht als Teil', () => {
  const proben = [
    { minute: 0, gesamt: 300, aufgabe: 'Start' },
    ...reihe([500, 500, 500]).map((x, i) => ({ ...x, aufgabe: `Dauer — Runde ${String(i)}: zwischen den Runden` })),
    { minute: 80.2, gesamt: 500, aufgabe: 'Ende' },
  ]
  assert.equal(bewertung(proben).gueltig, true)
})

test('dieselben WebKit-Prozesse über den ganzen Lauf lassen das Urteil gelten', () => {
  const proben = reihe([500, 500, 500]).map((x) => ({ ...x, webPids: ['1', '2'] }))
  assert.equal(bewertung(proben).gueltig, true)
})

// **Eine Marke ist ein ganzes Wort** — `d1` steckt sonst auch in `d12`, und
// eine verlorene Änderung fiele nicht auf.
test('eine Marke wird als ganzes Wort gefunden', () => {
  assert.equal(enthaelt('Regeltext\nd12', 'd1'), false)
  assert.equal(enthaelt('Regeltext\nd1', 'd1'), true)
  assert.equal(enthaelt('d1', 'd1'), true)
})

// **Wem ein WebKit-Prozess gehört, legt die erste Probe fest — und sie muss
// eindeutig sein** (Nachlauf über die Folgen vor 0.8.0, Befund 4). Startet
// ein fremder Prozess, während das Fenster aufgeht, zählte er sonst den
// ganzen Lauf über mit. Die Anwendung bringt genau einen GPU- und einen
// Netzprozess und einen oder zwei Webinhalte mit.
test('ein fremder WebKit-Prozess in der ersten Probe macht das Urteil ungültig', () => {
  const proben = reihe([500, 500, 500]).map((x) => ({ ...x, webPids: ['1', '2', '3', '4'], webArten: ['GPU', 'WebContent', 'Networking', 'EnhancedSecurity'] }))
  const urteil = bewertung(proben)
  assert.equal(urteil.gueltig, false)
  assert.match(urteil.grund, /eindeutig/)
})

test('zwei GPU-Prozesse in der ersten Probe sind nicht eindeutig', () => {
  const proben = reihe([500, 500, 500]).map((x) => ({ ...x, webPids: ['1', '2', '3', '4'], webArten: ['GPU', 'GPU', 'WebContent', 'Networking'] }))
  assert.equal(bewertung(proben).gueltig, false)
})

test('die Prozesse der Anwendung lassen das Urteil gelten', () => {
  for (const arten of [['GPU', 'WebContent', 'Networking'], ['GPU', 'WebContent', 'WebContent', 'Networking']]) {
    const proben = reihe([500, 500, 500]).map((x) => ({ ...x, webPids: arten.map((_, i) => String(i)), webArten: arten }))
    assert.equal(bewertung(proben).gueltig, true, arten.join(','))
  }
})
