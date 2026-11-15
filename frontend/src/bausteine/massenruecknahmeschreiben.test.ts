// **Die Rücknahme hört beim ersten Fehler auf** — und was danach käme,
// bleibt offen statt verloren (Befund 4 des Zweitblicks über Phase 6.6).
import { expect, it, vi } from 'vitest'

import { ruecknahmeSchreiben } from './massenruecknahme'

import type { MassenstandDTO } from '../antwortensammlung'

function stand(id: number, lagerort: string): MassenstandDTO {
  return {
    id,
    fassung: 2,
    menge: 1,
    zustand: 'nearMint',
    lagerort,
    wunsch: false,
  }
}

it('schreibt nach einem Fehler nichts mehr und führt den Rest als offen', async () => {
  const schreibe = vi
    .fn()
    // **Der Dienst nennt in `vorher` nur, was er schrieb** — Zeile 4 wurde
    // übergangen (Nachlauf über Phase 6.6) und bleibt offen.
    .mockResolvedValueOnce({
      geaendert: 1,
      befunde: ['übergangen'],
      vorher: [{ ...stand(1, 'Ordner'), fassung: 3 }],
    })
    .mockRejectedValueOnce(new Error('weg'))

  const ausgang = await ruecknahmeSchreiben(
    [stand(1, 'A'), stand(2, 'B'), stand(3, 'C'), stand(4, 'A')],
    schreibe,
  )

  expect(schreibe).toHaveBeenCalledTimes(2)
  expect(schreibe).toHaveBeenNthCalledWith(
    1,
    [
      { id: 1, fassung: 2 },
      { id: 4, fassung: 2 },
    ],
    'A',
  )
  expect(ausgang.zurueck).toBe(1)
  expect(ausgang.befunde).toEqual(['übergangen'])
  expect(ausgang.geschrieben).toBe(true)
  expect(ausgang.offen.map((zeile) => zeile.id)).toEqual([4, 2, 3])
  expect(ausgang.fehler).toBeInstanceOf(Error)
})
