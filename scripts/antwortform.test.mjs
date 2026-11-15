// Prüfungen der Antwortformprüfung — C01
// (#2226).
//
// **Die Positivkontrolle steht hier**: ein falsch serialisiertes Feld — ein
// umbenanntes, ein `null` statt einer Liste, eine Zahl statt eines Textes —
// muss rot werden.
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { pruefeForm } from './antwortform.mjs'

const setForm = {
  art: 'objekt',
  felder: {
    sets: { form: { art: 'liste', von: { art: 'objekt', felder: { id: { form: { art: 'string' } }, karten: { form: { art: 'number' } } } } } },
    fehler: { form: { art: 'oder', von: [{ art: 'objekt', felder: { code: { form: { art: 'string' } } } }, { art: 'null' }] } },
    hinweis: { form: { art: 'string' }, optional: true },
  },
}

test('eine richtige Antwort hält ihre Form', () => {
  assert.deepEqual(pruefeForm({ sets: [{ id: '1', karten: 3 }], fehler: null }, setForm).abweichungen, [])
})

test('ein umbenanntes Feld ist rot — fehlt und steht zusätzlich da', () => {
  const ergebnis = pruefeForm({ Sets: [], fehler: null }, setForm)
  assert.deepEqual(ergebnis.abweichungen, ['sets: fehlt'])
  assert.deepEqual(ergebnis.zusaetzlich, ['Sets'])
})

test('null statt einer Liste ist rot', () => {
  assert.deepEqual(pruefeForm({ sets: null, fehler: null }, setForm).abweichungen, ['sets: null statt liste'])
})

test('eine Zahl statt eines Textes ist rot — auch tief in der Liste', () => {
  assert.deepEqual(pruefeForm({ sets: [{ id: 7, karten: 3 }], fehler: null }, setForm).abweichungen, ['sets[0].id: number statt string'])
})

test('ein optionales Feld darf fehlen, ein Fehler darf stehen', () => {
  assert.deepEqual(pruefeForm({ sets: [], fehler: { code: 'x' } }, setForm).abweichungen, [])
})

test('ein Literal und eine Abbildung werden geprüft', () => {
  const form = { art: 'objekt', felder: { art: { form: { art: 'oder', von: [{ art: 'literal', wert: 'a' }, { art: 'literal', wert: 'b' }] } }, texte: { form: { art: 'abbildung', von: { art: 'string' } } } } }
  assert.deepEqual(pruefeForm({ art: 'a', texte: { x: 'y' } }, form).abweichungen, [])
  assert.deepEqual(pruefeForm({ art: 'c', texte: { x: 1 } }, form).abweichungen, ['art: "c" passt zu keiner Möglichkeit', 'texte.x: number statt string'])
})

test('unbekannt lässt alles durch und wird gezählt', () => {
  const ergebnis = pruefeForm({ x: 1 }, { art: 'objekt', felder: { x: { form: { art: 'unbekannt' } } } })
  assert.deepEqual(ergebnis.abweichungen, [])
  assert.equal(ergebnis.unbekannt, 1)
})

// **Die Formen kommen aus den Typen des Fensters** — eine je Bindung, und die
// von `Sets` nennt, was `SetDTO` verlangt.
test('die Formen stehen für jede Methode von Api da', async () => {
  const { antwortformen } = await import('./antwortformen.mjs')
  const formen = antwortformen()
  assert.ok(Object.keys(formen).length > 500)
  const set = formen.Sets.felder.sets.form.von.felder
  assert.deepEqual(set.id.form, { art: 'string' })
  assert.deepEqual(set.karten.form, { art: 'number' })
  assert.equal(pruefeForm({ sets: [{ id: 1 }] }, formen.Sets).abweichungen.length > 0, true)
})
