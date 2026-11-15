// Prüfungen von scripts/bindungsreichweite.mjs — B01
// (#2220).
//
// **Der Nenner kommt aus dem Bindungsvertrag**, nicht aus den beobachteten
// Namen — sonst wäre jede Quote 100 %. Die Positivkontrolle steht hier: eine
// neue Bindung, die kein Drehbuch ruft, senkt die Quote.
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { gemischtAus, listeNach, nurBedienwege, reichweite } from './bindungsreichweite.mjs'

const schreibwege = `# Schreibwege

## Die 2 schreibenden Bindungen

\`SetAnlegen\` · \`KarteSetzen\`

## Die 2 lesenden Bindungen

**Erklärung**, die \`Fremdwort\` nennt.

\`Sets\` · \`Karte\`

---
`

test('die Liste ist der letzte Absatz unter der Überschrift', () => {
  assert.deepEqual(listeNach(schreibwege, 'schreibenden Bindungen'), ['SetAnlegen', 'KarteSetzen'])
  assert.deepEqual(listeNach(schreibwege, 'lesenden Bindungen'), ['Sets', 'Karte'])
})

test('eine fehlende Überschrift ist ein Fehler und keine leere Liste', () => {
  assert.throws(() => listeNach(schreibwege, 'gibt es nicht'), /gibt es nicht/)
})


const eingabe = () => ({
  vertrag: ['SetAnlegen', 'KarteSetzen', 'Sets', 'Karte'],
  schreibend: ['SetAnlegen', 'KarteSetzen'],
  lesend: ['Sets', 'Karte'],
  dialog: ['SetAnlegen'],
  ausnahmen: {},
  rufe: { SetAnlegen: { anzahl: 1, ausgaenge: { gelungen: 1 } }, Sets: { anzahl: 3, ausgaenge: { gelungen: 3 } } },
})

test('die Quote zählt erreichte gegen alle Bindungen der Klasse', () => {
  const bilanz = reichweite(eingabe())
  assert.equal(bilanz.gesamt.nenner, 4)
  assert.equal(bilanz.gesamt.erreicht, 2)
  assert.equal(bilanz.schreibend.quote, 0.5)
  assert.deepEqual(bilanz.schreibend.offen, ['KarteSetzen'])
  assert.equal(bilanz.dialog.erreicht, 1)
})

test('eine neue Bindung, die niemand ruft, senkt die Quote — die Positivkontrolle', () => {
  const vorher = reichweite(eingabe()).schreibend.quote
  const mehr = eingabe()
  mehr.vertrag.push('NeueBindung')
  mehr.schreibend.push('NeueBindung')
  assert.ok(reichweite(mehr).schreibend.quote < vorher)
})

test('eine begründete Ausnahme steht neben der Quote und nicht darin', () => {
  const mit = eingabe()
  mit.ausnahmen = { KarteSetzen: 'braucht einen fremden Dienst mit Zugang' }
  const bilanz = reichweite(mit)
  assert.equal(bilanz.schreibend.quote, 1)
  assert.deepEqual(bilanz.schreibend.ausgenommen, { KarteSetzen: 'braucht einen fremden Dienst mit Zugang' })
})

test('ein gerufener Name außerhalb des Vertrags wird genannt und nicht gezählt', () => {
  const mit = eingabe()
  mit.rufe.Erfunden = { anzahl: 1, ausgaenge: { abgewiesen: 1 } }
  const bilanz = reichweite(mit)
  assert.deepEqual(bilanz.fremd, ['Erfunden'])
  assert.equal(bilanz.gesamt.erreicht, 2)
})

test('eine Bindung ohne Klasse ist ein Fehler — der Vertrag und die Schreibwege laufen sonst auseinander', () => {
  const mit = eingabe()
  mit.vertrag.push('OhneKlasse')
  assert.throws(() => reichweite(mit), /OhneKlasse/)
})

test('nur gerufen zählt — eine Bindung mit null Rufen ist nicht erreicht', () => {
  const mit = eingabe()
  mit.rufe.Karte = { anzahl: 0, ausgaenge: {} }
  assert.equal(reichweite(mit).lesend.erreicht, 1)
})

test('ein Ruf, den das Tor abwies, hat die Bindung nicht erreicht', () => {
  const mit = eingabe()
  mit.rufe.KarteSetzen = { anzahl: 2, ausgaenge: { abgewiesen: 2 } }
  assert.deepEqual(reichweite(mit).schreibend.offen, ['KarteSetzen'])
  mit.rufe.KarteSetzen = { anzahl: 3, ausgaenge: { abgewiesen: 2, fachfehler: 1 } }
  assert.deepEqual(reichweite(mit).schreibend.offen, [])
})

// **Und einer, den die Störhülle selbst beantwortete, auch nicht** — sie
// führt die Bindung nicht aus (Zweitblick über Phase B, Befund 4).
test('ein gestörter Ruf hat die Bindung nicht erreicht', () => {
  const mit = eingabe()
  mit.rufe.KarteSetzen = { anzahl: 2, ausgaenge: { gestoert: 1, abgewiesen: 1 } }
  assert.deepEqual(reichweite(mit).schreibend.offen, ['KarteSetzen'])
  mit.rufe.KarteSetzen = { anzahl: 3, ausgaenge: { gestoert: 2, gelungen: 1 } }
  assert.deepEqual(reichweite(mit).schreibend.offen, [])
})

test('die Bedienwege stehen getrennt vom Vertragsrundgang', () => {
  const je = {
    kernwege: { SetAnlegen: { anzahl: 1, ausgaenge: { gelungen: 1 }, dauerMs: 1 } },
    vertrag: { KarteSetzen: { anzahl: 1, ausgaenge: { gelungen: 1 }, dauerMs: 1 } },
  }
  assert.deepEqual(Object.keys(nurBedienwege(je)), ['SetAnlegen'])
  assert.deepEqual(Object.keys(gemischtAus(je)).sort(), ['KarteSetzen', 'SetAnlegen'])
})

// **Erreicht ist nicht gelungen** — ein fachlich abgelehnter Ruf ging durch
// die Naht, schrieb aber nichts. Beide Zahlen stehen nebeneinander.
test('gelungen zählt nur Rufe, die ohne Fehler zurückkamen', () => {
  const mit = eingabe()
  mit.rufe.KarteSetzen = { anzahl: 1, ausgaenge: { fachfehler: 1 } }
  const bilanz = reichweite(mit)
  assert.equal(bilanz.schreibend.erreicht, 2)
  assert.equal(bilanz.schreibend.gelungen, 1)
})
