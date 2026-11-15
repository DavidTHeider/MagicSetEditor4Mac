// **Das Urteil einer Zelle der Zustandsprobe** — G11
// (#2277).
//
// B03 verzögerte bei jeder Setseite auch die Bindungen, die das Öffnen des
// Sets ruft, und Ersetzen und Ausgeben standen als „ohne Ladehinweis" da —
// obwohl sie nichts davon zeigen. Ob eine Seite von einer Bindung abhängt,
// wird nicht behauptet, sondern gemessen: steht ihr Inhalt bei gestörten
// Bindungen wie im befüllten Zustand, hängt er nicht an ihnen.
import { strict as assert } from 'node:assert'
import { test } from 'node:test'

import { urteil } from './huellenprobe/zustaende.mjs'

const lage = (teil) => ({ laedt: false, alarm: [], ganz: 'Ersetzen Text suchen', bedienung: 'input:text:abc', ...teil })

// befuellt ist, wie die Seite im befüllten Durchgang stand: Text **und**
// Bedienzustand (Befund 6 des Zweitblicks über Phase G).
const befuellt = { ganz: 'Ersetzen Text suchen', bedienung: 'input:text:abc' }

test('laden: wer sagt, dass er lädt, sagt es', () => {
  assert.equal(urteil('laden', lage({ laedt: true, ganz: 'anders' }), befuellt), 'sagt, dass sie lädt')
})

test('laden: steht die Seite wie befüllt, hängt ihr Inhalt nicht an den Bindungen', () => {
  assert.equal(urteil('laden', lage(), befuellt), 'steht wie befüllt')
})

test('laden: fehlt Inhalt ohne Hinweis, ist das ein Befund', () => {
  assert.equal(urteil('laden', lage({ ganz: 'Ersetzen' }), befuellt), 'ohne Ladehinweis')
})

test('Fehler: eine Meldung ist eine Meldung', () => {
  assert.equal(urteil('Fehler', lage({ alarm: ['weg'], ganz: 'x' }), befuellt), 'meldet')
})

test('Fehler: steht die Seite wie befüllt, hat das Scheitern ihr nichts genommen', () => {
  assert.equal(urteil('Fehler', lage(), befuellt), 'steht wie befüllt')
})

test('Fehler: fehlt Inhalt ohne Meldung, ist das ein Befund', () => {
  assert.equal(urteil('Fehler', lage({ ganz: '' }), befuellt), 'ohne Meldung')
})

// **Derselbe Text mit einem anderen Feldwert oder einem gesperrten Knopf ist
// nicht wie befüllt** — Befund 6 des Zweitblicks über Phase G: der Vergleich
// las nur `innerText`, und eine andere Auswahl gab denselben Text.
test('gleicher Text, anderer Bedienzustand: ein Befund', () => {
  const anders = lage({ bedienung: 'input:text:' })
  assert.equal(urteil('laden', anders, befuellt), 'ohne Ladehinweis')
  assert.equal(urteil('Fehler', anders, befuellt), 'ohne Meldung')
})

// **Ohne befüllten Vergleich kein „wie befüllt"** — eine Seite, die im
// befüllten Durchgang nicht erreicht wurde, hat nichts, woran sie sich messen
// ließe.
test('ohne Vergleich bleibt es beim Befund', () => {
  assert.equal(urteil('laden', lage({ ganz: '' }), undefined), 'ohne Ladehinweis')
  assert.equal(urteil('Fehler', lage({ ganz: '' }), undefined), 'ohne Meldung')
})
