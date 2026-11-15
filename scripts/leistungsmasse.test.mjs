// Prüfungen von scripts/leistungsmasse.mjs — Z2 des Zweitblicks über Phase 6.11.
//
// **„Gleiche Auflösung" ist eine Zusage des Vergleichs** (`docs/LEISTUNGSVERGLEICH.md`
// §2): gleich ist sie, wenn je Karte Breite und Höhe gleich sind. Bis dahin
// wurden Breite und Höhe gespeichert und nirgends verglichen.

import { test } from 'node:test'
import assert from 'node:assert/strict'

import { mse2Dateiname, mse2Dateinamen, schrittmasseVergleichen, vergleicheBilderpaare } from './leistungsmasse.mjs'

test('MSE2 macht aus dem Kartennamen einen Dateinamen: erlaubt sind Buchstaben, Ziffern, Leerzeichen, _, - und .', () => {
  assert.equal(mse2Dateiname('All in Good Time'), 'All in Good Time.png')
  assert.equal(mse2Dateiname('"Ach! Hans, Run!"'), 'Ach Hans Run.png')
  assert.equal(mse2Dateiname("Mercadia's Downfall"), 'Mercadias Downfall.png')
  assert.equal(mse2Dateiname('Nicol Bolas, Dragon-God'), 'Nicol Bolas Dragon-God.png')
  assert.equal(mse2Dateiname('Åcht Æther'), 'Åcht Æther.png')
})

test('Auszeichnungen im Namen gehören nicht in den Dateinamen', () => {
  assert.equal(mse2Dateiname('Plain<b>er</b> Name'), 'Plainer Name.png')
})

test('vorn werden Leerzeichen und Punkte abgeschnitten — hinten schützt sie die Endung', () => {
  assert.equal(mse2Dateiname('  ..Anfang'), 'Anfang.png')
  assert.equal(mse2Dateiname('Ende . '), 'Ende . .png', 'bereinigt wird die ganze Zeichenkette samt „.png“')
})

test('ein Name ohne erlaubtes Zeichen verliert mit dem Punkt auch die Endung — wie bei MSE2', () => {
  // `clean_filename` läuft über `{card.name}.png` als Ganzes; der Punkt vorn
  // fällt, und übrig bleibt „png“. Nur eine Karte ohne Namen kommt hierher.
  assert.equal(mse2Dateiname('???'), 'png')
  assert.equal(mse2Dateiname(''), 'png')
})

test('gleiche Namen werden durchnummeriert: Name.png, Name.1.png, Name.2.png', () => {
  assert.deepEqual(mse2Dateinamen(['Plains', 'Island', 'Plains', 'Plains', 'Island']),
    ['Plains.png', 'Island.png', 'Plains.1.png', 'Plains.2.png', 'Island.1.png'])
})

test('Namen, die erst bereinigt gleich werden, zählen als gleich', () => {
  assert.deepEqual(mse2Dateinamen(['Ach! Hans', 'Ach Hans']), ['Ach Hans.png', 'Ach Hans.1.png'])
})

const namen = ['Eins', 'Zwei', 'Drei']
const dateienMse2 = { 'Eins.png': { breite: 435, hoehe: 621 }, 'Zwei.png': { breite: 435, hoehe: 621 }, 'Drei.png': { breite: 435, hoehe: 621 } }
const gleich = { '0.png': { breite: 435, hoehe: 621 }, '1.png': { breite: 435, hoehe: 621 }, '2.png': { breite: 435, hoehe: 621 } }

test('gleiche Maße je gepaarter Karte: kein Befund', () => {
  const v = vergleicheBilderpaare({ namen, referenz: dateienMse2, andere: gleich, andererName: (i) => `${i}.png` })
  assert.deepEqual(v, { gepaart: 3, abweichungen: [], ohneReferenz: [], ohneAndere: [] })
})

test('eine andere Breite oder Höhe bei einer Karte ist eine Abweichung — mit Stelle, Name und beiden Maßen', () => {
  const v = vergleicheBilderpaare({
    namen, referenz: dateienMse2, andererName: (i) => `${i}.png`,
    andere: { ...gleich, '1.png': { breite: 436, hoehe: 621 }, '2.png': { breite: 621, hoehe: 435 } },
  })
  assert.equal(v.gepaart, 3)
  assert.equal(v.abweichungen.length, 2)
  assert.match(v.abweichungen[0], /Karte 1 \(Zwei\): 436x621, MSE2 435x621/)
  assert.match(v.abweichungen[1], /Karte 2 \(Drei\): 621x435, MSE2 435x621/)
})

test('es genügt, dass nur die Höhe oder nur die Breite abweicht', () => {
  const nurHoehe = vergleicheBilderpaare({ namen, referenz: dateienMse2, andererName: (i) => `${i}.png`, andere: { ...gleich, '0.png': { breite: 435, hoehe: 620 } } })
  const nurBreite = vergleicheBilderpaare({ namen, referenz: dateienMse2, andererName: (i) => `${i}.png`, andere: { ...gleich, '0.png': { breite: 434, hoehe: 621 } } })
  assert.equal(nurHoehe.abweichungen.length, 1)
  assert.equal(nurBreite.abweichungen.length, 1)
})

test('ein Bild, das bei einem der beiden fehlt, ist ungepaart — und keine stillschweigende Lücke', () => {
  const { '1.png': _weg, ...ohneEins } = gleich
  const { 'Drei.png': _auch, ...ohneDrei } = dateienMse2
  const v = vergleicheBilderpaare({ namen, referenz: ohneDrei, andere: ohneEins, andererName: (i) => `${i}.png` })
  assert.equal(v.gepaart, 1)
  assert.equal(v.ohneAndere.length, 1)
  assert.equal(v.ohneReferenz.length, 1)
  assert.match(v.ohneAndere[0], /Karte 1 \(Zwei\): kein Bild bei den anderen \(1\.png\)/)
  assert.match(v.ohneReferenz[0], /Karte 2 \(Drei\): kein Bild bei MSE2 \(Drei\.png\)/)
})

const schritt = (programm, satz, stelle, breite, hoehe, zusatz = {}) =>
  ({ programm, satz, stelle, zustand: 'warm', wiederholung: 0, ok: true, breite, hoehe, ...zusatz })

test('die Schritte der Sitzungen werden je Karte zwischen den Programmen verglichen', () => {
  const abweichungen = schrittmasseVergleichen([
    schritt('mse2', 'Multicolor', 5, 435, 621), schritt('heute', 'Multicolor', 5, 435, 621), schritt('v050', 'Multicolor', 5, 435, 621),
    schritt('mse2', 'Multicolor', 9, 435, 621), schritt('heute', 'Multicolor', 9, 400, 621), schritt('v050', 'Multicolor', 9, 435, 621),
  ])
  assert.equal(abweichungen.length, 1)
  assert.deepEqual([abweichungen[0].programm, abweichungen[0].satz, abweichungen[0].stelle], ['heute', 'Multicolor', 9])
  assert.match(abweichungen[0].text, /Multicolor\[9\]: heute 400x621, MSE2 435x621/)
})

test('dieselbe Karte in einem Programm mit zwei Größen ist ebenfalls eine Abweichung', () => {
  const abweichungen = schrittmasseVergleichen([
    schritt('mse2', 'S', 1, 435, 621), schritt('heute', 'S', 1, 435, 621),
    schritt('heute', 'S', 1, 435, 620, { zustand: 'wiederholung' }),
  ])
  assert.equal(abweichungen.length, 1)
  assert.equal(abweichungen[0].programm, 'heute')
  assert.match(abweichungen[0].text, /435x620/)
})

test('auch die Referenz selbst, die eine Karte in zwei Größen zeichnet, ist eine Abweichung', () => {
  const abweichungen = schrittmasseVergleichen([
    schritt('mse2', 'S', 1, 435, 621), schritt('mse2', 'S', 1, 435, 620, { wiederholung: 1 }), schritt('heute', 'S', 1, 435, 621),
  ])
  assert.deepEqual(abweichungen.map((a) => a.programm), ['mse2', 'heute'])
})

test('zwei Größen bei einem Programm zählen auch dann, wenn die Referenz kein Maß hat', () => {
  const abweichungen = schrittmasseVergleichen([
    schritt('mse2', 'S', 1, null, null, { ok: false }),
    schritt('heute', 'S', 1, 435, 621), schritt('heute', 'S', 1, 400, 621, { wiederholung: 1 }),
  ])
  assert.equal(abweichungen.length, 1)
  assert.doesNotMatch(abweichungen[0].text, /MSE2/)
})

test('ein Schritt ohne Bild — oder ohne Referenz — wird nicht verglichen, sondern ist als Fehler schon gezählt', () => {
  assert.deepEqual(schrittmasseVergleichen([
    schritt('mse2', 'S', 1, null, null, { ok: false }), schritt('heute', 'S', 1, 435, 621),
    schritt('mse2', 'S', 2, 435, 621), schritt('heute', 'S', 2, null, null, { ok: false }),
  ]), [])
})

test('ein Schritt ohne Maße (Skript, Bereitschaft) ist kein Bildschritt', () => {
  assert.deepEqual(schrittmasseVergleichen([{ programm: 'mse2', schicht: 'Skript: Rechnung', ok: true }, { programm: 'heute', schicht: 'bereit', ok: true }]), [])
})
