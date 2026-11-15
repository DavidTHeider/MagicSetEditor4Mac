import { describe, expect, it } from 'vitest'

import { einsetzen } from './einsetzen'

describe('einsetzen', () => {
  it('setzt jeden Platzhalter ein', () => {
    expect(einsetzen('{a} und {b}', { a: 'eins', b: 'zwei' })).toBe(
      'eins und zwei',
    )
  })

  // **Der Fall, um den es geht** (Nachlauf zum Zweitblick über Phase 9): ein
  // eingesetzter Wert, der wie ein Platzhalter aussieht, wird nicht noch einmal
  // ausgewertet. Zwei `replace` nacheinander täten genau das.
  it('wertet einen eingesetzten Wert nicht noch einmal aus', () => {
    expect(einsetzen('{a} und {b}', { a: '{b}', b: 'zwei' })).toBe(
      '{b} und zwei',
    )
  })

  // Ein Platzhalter ohne Wert bleibt stehen, statt zu verschwinden: ein Satz
  // mit einer Lücke ist eine Auskunft, ein Satz ohne die Angabe ist keine.
  it('lässt einen Platzhalter ohne Wert stehen', () => {
    expect(einsetzen('{a} und {b}', { a: 'eins' })).toBe('eins und {b}')
  })
})
