// **Das Menü einer Karte aus der Auskunft des Kerns** — P4
// (#2372):
// was ohne Feld geht, steht direkt darin; „Auf den Friedhof“ geht über die
// gedeutete Geste; „Weitere …“ führt in die Handlungsleiste — auch dann, wenn
// der Kern nichts anbietet.
import { describe, expect, test, vi } from 'vitest'

import { kartenmenue } from './kartenmenue'

import type { Objekthandlung } from '../apipartie'

const worte = {
  kartenmenueWeitere: 'Weitere …',
  kartenmenueFriedhof: 'Auf den Friedhof',
}
const namen = {
  handlungsname: { 'objekt.tappen': 'Tappen', 'objekt.opfern': 'Opfern' },
}

const baer: Objekthandlung = {
  objekt: 'sitz1/baer/0',
  handlungen: [
    'objekt.tappen',
    'objekt.zaehler',
    'objekt.opfern',
    'objekt.ziehen',
  ],
  primaer: 'objekt.tappen',
  zweit: '',
}

describe('das Kartenmenü', () => {
  test('nennt, was ohne Feld geht, den Friedhof und die Leiste', () => {
    const befehlen = vi.fn()
    const waehle = vi.fn()
    const menue = kartenmenue(
      worte,
      namen,
      { kennung: 'sitz1/baer/0', eigner: 'sitz1' },
      baer,
      befehlen,
      waehle,
    )

    expect(menue.map((eintrag) => eintrag.name)).toEqual([
      'Tappen',
      'Opfern',
      'Auf den Friedhof',
      'Weitere …',
    ])

    for (const eintrag of menue) {
      if ('aufWahl' in eintrag) {
        eintrag.aufWahl()
      }
    }

    expect(befehlen.mock.calls).toEqual([
      ['objekt.tappen', { objekt: 'sitz1/baer/0' }],
      ['objekt.opfern', { objekt: 'sitz1/baer/0' }],
      [
        'objekt.ziehen',
        { objekt: 'sitz1/baer/0', zone: 'friedhof', sitz: 'sitz1' },
      ],
    ])
    expect(waehle).toHaveBeenCalledWith('sitz1/baer/0')
  })

  // **Ein Menü ohne Eintrag wäre eine Sackgasse** — die Leiste sagt danach,
  // was gerade gilt.
  test('führt ohne Auskunft in die Leiste', () => {
    const menue = kartenmenue(
      worte,
      namen,
      { kennung: 'sitz2/hand/0', eigner: 'sitz2' },
      undefined,
      vi.fn(),
      vi.fn(),
    )

    expect(menue.map((eintrag) => eintrag.schluessel)).toEqual(['weitere'])
  })

  // **Was schreibt, sagt es** —
  // #2482:
  // für einen Leser stehen diese Zeilen gesperrt und erklärt da, in Menü und
  // Menüleiste gleich (`Menueliste`, `useHandlungsliste`). „Weitere …“ wählt
  // nur und schreibt nichts.
  test('kennzeichnet die schreibenden Einträge', () => {
    const menue = kartenmenue(
      worte,
      namen,
      { kennung: 'sitz1/baer/0', eigner: 'sitz1' },
      baer,
      vi.fn(),
      vi.fn(),
    )

    expect(
      menue.map((eintrag) => [eintrag.schluessel, eintrag.schreibt === true]),
    ).toEqual([
      ['objekt.tappen', true],
      ['objekt.opfern', true],
      ['friedhof', true],
      ['weitere', false],
    ])
  })
})
