// **War der Bildschirm aus, während ein Drehbuch lief?** — C03
// (#2228).
//
// Der zweite volle Lauf vom 2026-10-10 lief ganz bei ausgeschaltetem
// Bildschirm: zwanzig Drehbücher wurden rot, fast alle an Fristen, und ihre
// Belege sahen aus wie Befunde an der Anwendung. Nur `readmebilder` sagte, dass
// der Bildschirm schlief — weil `screencapture` scheiterte. Eine Webansicht,
// die niemand sieht, zeichnet nicht, und wer auf ein gezeichnetes Bild wartet,
// wartet bis zur Frist.
import { strict as assert } from 'node:assert'
import { test } from 'node:test'

import { ausphasen, bildschirmAus, ueberschneidungen, vermerkeImBeleg, vermerkeLauf } from './bildschirmzeit.mjs'

// Zwei Zeilen aus `pmset -g log` dieses Rechners, wie sie dastehen — die
// Lücke zwischen Zeit und Art ist ein Tabulator.
const echt = [
  '2026-10-10 03:00:13 +0200 Assertions          \tPID 42499(caffeinate) Summary PreventUserIdleDisplaySleep "caffeinate command-line tool"',
  '2026-10-10 04:48:35 +0200 Notification        \tDisplay is turned off                                                      \t          ',
  '2026-10-10 06:31:33 +0200 Notification        \tDisplay is turned on                                                       \t          ',
].join('\n')

// echtBisNachDemLauf reicht über das Ende der Läufe unten hinaus — ein
// Protokoll, das vorher endet, sagt nichts (#2284).
const echtBisNachDemLauf = `${echt}\n2026-10-10 07:10:00 +0200 Assertions          \tPID 1(x) Summary`

test('liest eine Ausphase aus dem Energieprotokoll', () => {
  assert.deepEqual(ausphasen(echt), [
    { von: '2026-10-10T02:48:35.000Z', bis: '2026-10-10T04:31:33.000Z' },
  ])
})

test('eine Ausphase ohne Ende dauert noch an', () => {
  const text = '2026-10-10 04:48:35 +0200 Notification        \tDisplay is turned off'

  assert.deepEqual(ausphasen(text), [{ von: '2026-10-10T02:48:35.000Z', bis: null }])
})

// **Ein Einschalten ohne vorheriges Ausschalten** heißt: aus seit vor dem
// Beginn des Protokolls — nicht „nie aus".
test('ein Einschalten ohne Ausschalten davor war eine Ausphase seit unbekannt', () => {
  const text = '2026-10-10 06:31:33 +0200 Notification        \tDisplay is turned on'

  assert.deepEqual(ausphasen(text), [{ von: null, bis: '2026-10-10T04:31:33.000Z' }])
})

test('findet die Phasen, die einen Zeitraum berühren, und nur sie', () => {
  const phasen = ausphasen(echt)

  // Lauf 2 begann um 03:50 UTC, mitten in der Phase.
  assert.equal(ueberschneidungen(phasen, '2026-10-10T03:50:42Z', '2026-10-10T05:05:00Z').length, 1)
  // Lauf 1 endete vorher.
  assert.equal(ueberschneidungen(phasen, '2026-10-10T01:43:50Z', '2026-10-10T02:40:00Z').length, 0)
  // Ein Drehbuch nach dem Einschalten.
  assert.equal(ueberschneidungen(phasen, '2026-10-10T04:31:34Z', '2026-10-10T04:40:00Z').length, 0)
  // Eine offene Phase reicht bis heute.
  assert.equal(ueberschneidungen([{ von: '2026-10-10T02:48:35.000Z', bis: null }], '2026-10-10T09:00:00Z', '2026-10-10T09:01:00Z').length, 1)
})

// **Ein unlesbares Protokoll ist kein „war an"** (AGENTS.md §10, das Nein
// gehört zum Ergebnis): wer es nicht weiß, sagt `undefined`.
test('sagt nichts, wenn sich das Protokoll nicht lesen lässt', () => {
  const nichtLesbar = () => {
    throw new Error('pmset fehlt')
  }

  assert.equal(bildschirmAus('2026-10-10T03:50:42Z', '2026-10-10T05:05:00Z', nichtLesbar), undefined)
  assert.deepEqual(bildschirmAus('2026-10-10T03:50:42Z', '2026-10-10T05:05:00Z', () => echtBisNachDemLauf), [
    { von: '2026-10-10T02:48:35.000Z', bis: '2026-10-10T04:31:33.000Z' },
  ])
})

test('vermerkt die Ausphase im Kopf des Belegs, nicht in seinen Zeilen', () => {
  const beleg = [
    '# Hüllenprobe: start',
    '',
    '**1 Befund(e)** — Commit `0326d3d7e`,',
    'Lauf `x`, Bauabdruck `y`.',
    '',
    '- ✗ das Fenster füllt die Arbeitsfläche',
    '',
  ].join('\n')

  const neu = vermerkeImBeleg(beleg, [{ von: '2026-10-10T02:48:35.000Z', bis: '2026-10-10T04:31:33.000Z' }])

  assert.equal(neu, [
    '# Hüllenprobe: start',
    '',
    '**1 Befund(e)** — Commit `0326d3d7e`,',
    'Lauf `x`, Bauabdruck `y`.',
    '**Der Bildschirm war aus** (2026-10-10T02:48:35.000Z bis 2026-10-10T04:31:33.000Z), während das Drehbuch lief — ein Rot hier ist zuerst die Umgebung: eine Webansicht, die niemand sieht, zeichnet nicht.',
    '',
    '- ✗ das Fenster füllt die Arbeitsfläche',
    '',
  ].join('\n'))
  assert.match(vermerkeImBeleg(beleg, [{ von: null, bis: null }]), /\(unbekannt bis andauernd\)/)
})

// **Je Drehbuch, nach seiner eigenen Zeit** — und in jeder Kopie seines
// Belegs: der unter `docs/belege` wird gelesen, die im Laufverzeichnis
// verglichen.
test('vermerkt nur die Drehbücher, die im Dunkeln liefen', () => {
  const kopf = '# Hüllenprobe: x\n\n**bestanden** — Commit `c`.\n\n- ✓ a\n'
  const dateien = new Map([['a.md', kopf], ['a-kopie.md', kopf], ['b.md', kopf]])
  const faelle = [
    { name: 'a', beginn: '2026-10-10T04:30:00Z', ende: '2026-10-10T04:32:00Z', belege: ['a.md', 'a-kopie.md'] },
    { name: 'b', beginn: '2026-10-10T04:32:00Z', ende: '2026-10-10T04:35:00Z', belege: ['b.md'] },
  ]

  const imDunkeln = vermerkeLauf(faelle, ausphasen(echt), {
    lesen: (pfad) => dateien.get(pfad),
    schreiben: (pfad, text) => dateien.set(pfad, text),
  })

  assert.deepEqual(imDunkeln, ['a'])
  assert.deepEqual(faelle[0].bildschirmAus, [{ von: '2026-10-10T02:48:35.000Z', bis: '2026-10-10T04:31:33.000Z' }])
  assert.equal(faelle[1].bildschirmAus, undefined)
  assert.match(dateien.get('a.md'), /Der Bildschirm war aus/)
  assert.match(dateien.get('a-kopie.md'), /Der Bildschirm war aus/)
  assert.equal(dateien.get('b.md'), kopf)
})

// **Was das Protokoll nicht abdeckt, ist unbekannt** — Zweitblick über Phase C,
// Befund 7: ein leeres Protokoll oder eines ohne Bildschirmzeilen ergab „keine
// Ausphase", also scheinbar „war an". Bekannt ist der Zustand nur, wenn ein
// Wechsel **vor** dem Beginn des Laufs im Protokoll steht.
test('ein leeres, fremdes oder zu spät beginnendes Protokoll sagt nichts', () => {
  const lauf = ['2026-10-10T03:50:42Z', '2026-10-10T05:05:00Z']
  assert.equal(bildschirmAus(...lauf, () => ''), undefined)
  assert.equal(bildschirmAus(...lauf, () => 'kein Protokoll\nnur Text'), undefined)
  assert.equal(bildschirmAus(...lauf, () => '2026-10-10 03:00:13 +0200 Assertions          \tPID 1(x) Summary'), undefined)
  // Der erste Wechsel kommt erst nach dem Beginn — wie es davor stand, weiß
  // niemand.
  assert.equal(bildschirmAus(...lauf, () => '2026-10-10 06:31:33 +0200 Notification        \tDisplay is turned on'), undefined)
})

test('ein Wechsel vor dem Beginn macht den Zustand bekannt', () => {
  const an = '2026-10-10 01:00:00 +0200 Notification        \tDisplay is turned on\n2026-10-10 05:00:00 +0200 Assertions          \tPID 1(x) Summary'
  assert.deepEqual(bildschirmAus('2026-10-10T01:43:50Z', '2026-10-10T02:35:34Z', () => an), [])
  assert.deepEqual(bildschirmAus('2026-10-10T03:50:42Z', '2026-10-10T05:05:00Z', () => echtBisNachDemLauf), [
    { von: '2026-10-10T02:48:35.000Z', bis: '2026-10-10T04:31:33.000Z' },
  ])
})

// **Und das Protokoll muss bis zum Ende des Laufs reichen** — Nachlauf über
// Phase C (#2284): ein abgeschnittenes mit einer alten Einschaltzeile galt
// als „nie aus“. Was nach seiner letzten Zeile geschah, weiß es nicht.
test('ein Protokoll, das vor dem Laufende abbricht, sagt nichts', () => {
  const lauf = ['2026-10-10T03:50:42Z', '2026-10-10T05:05:00Z']
  const an = '2026-10-10 01:00:00 +0200 Notification        \tDisplay is turned on'
  const vorEnde = '2026-10-10 06:00:00 +0200 Assertions          \tPID 1(x) Summary'
  const nachEnde = '2026-10-10 07:10:00 +0200 Assertions          \tPID 1(x) Summary'
  assert.equal(bildschirmAus(...lauf, () => `${an}\n${vorEnde}`), undefined)
  assert.deepEqual(bildschirmAus(...lauf, () => `${an}\n${nachEnde}`), [])
})
