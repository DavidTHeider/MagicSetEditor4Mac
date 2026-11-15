// **Die Einzeltasten an einer Karte** — P4
// (#2372),
// Befund C-09: T für die Primärhandlung, ⌫ auf den Friedhof, E ins Exil —
// nur, was der Kern an dieser Karte anbietet, und angesagt über
// `aria-keyshortcuts`.
import { renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import type { ReactNode } from 'react'

import { einObjekt, einePartie } from '../seiten/partieprobe'
import { Schreibrechtraum } from './schreibrecht'
import { useKartenbedienung } from './useKartenbedienung'

import type { Objekthandlung } from '../apipartie'

const baer: Objekthandlung = {
  objekt: 'sitz1/baer/0',
  handlungen: ['objekt.tappen', 'objekt.exilieren', 'objekt.ziehen'],
  primaer: 'objekt.tappen',
  zweit: '',
}

function zeige(handlungen: Objekthandlung[] = [baer], nurLesend = false) {
  const befehlen = vi.fn()
  const waehle = vi.fn()
  const partie = einePartie({
    objekte: [
      einObjekt({ kennung: 'sitz1/baer/0', eigner: 'sitz1', zone: 'tisch' }),
      einObjekt({ kennung: 'sitz2/hand/0', eigner: 'sitz2', zone: 'hand' }),
    ],
  })
  const { result } = renderHook(
    () =>
      useKartenbedienung({ partie, lauf: { handlungen, befehlen }, waehle }),
    {
      wrapper: ({ children }: { children: ReactNode }) => (
        <Schreibrechtraum nurLesend={nurLesend}>{children}</Schreibrechtraum>
      ),
    },
  )

  return { befehlen, waehle, bedienung: result.current }
}

describe('die Einzeltasten an einer Karte', () => {
  it('nennt die Tasten, die an dieser Karte gelten', () => {
    const { bedienung } = zeige()

    expect(bedienung.tastenFuer('sitz1/baer/0')).toBe('T Backspace E')
    expect(bedienung.tastenFuer('sitz2/hand/0')).toBe('')
  })

  it.each([
    ['t', 'objekt.tappen', { objekt: 'sitz1/baer/0' }],
    [
      'Backspace',
      'objekt.ziehen',
      { objekt: 'sitz1/baer/0', zone: 'friedhof', sitz: 'sitz1' },
    ],
    [
      'Delete',
      'objekt.ziehen',
      { objekt: 'sitz1/baer/0', zone: 'friedhof', sitz: 'sitz1' },
    ],
    ['e', 'objekt.exilieren', { objekt: 'sitz1/baer/0' }],
  ])('stellt mit %s %s', (taste, art, nutzlast) => {
    const { befehlen, bedienung } = zeige()

    expect(bedienung.taste('sitz1/baer/0', taste)).toBe(true)
    expect(befehlen).toHaveBeenCalledWith(art, nutzlast)
  })

  it('lässt fremde Tasten und Karten ohne Auskunft durch', () => {
    const { befehlen, bedienung } = zeige()

    expect(bedienung.taste('sitz1/baer/0', 'x')).toBe(false)
    expect(bedienung.taste('sitz2/hand/0', 't')).toBe(false)
    expect(befehlen).not.toHaveBeenCalled()
  })

  // **Ohne Primärhandlung tut T nichts** — es wählt auch nicht, denn das
  // tut die Leertaste.
  it('tut ohne Primärhandlung mit T nichts', () => {
    const { befehlen, bedienung } = zeige([
      { ...baer, primaer: '', handlungen: [] },
    ])

    expect(bedienung.tastenFuer('sitz1/baer/0')).toBe('')
    expect(bedienung.taste('sitz1/baer/0', 't')).toBe(false)
    expect(befehlen).not.toHaveBeenCalled()
  })

  // **Eine Auskunft zu einem Objekt, das der Abzug nicht kennt** — die
  // Geste geht ohne Ziel-Sitz hinaus, und der Kern sagt warum
  // (`internal/service/zugdeutung.go`); ein Menü gibt es dazu nicht.
  it('rät keinen Sitz und kein Menü zu einem unbekannten Objekt', () => {
    const { befehlen, bedienung } = zeige([{ ...baer, objekt: 'weg/0' }])

    expect(bedienung.taste('weg/0', 'Backspace')).toBe(true)
    expect(befehlen).toHaveBeenCalledWith('objekt.ziehen', {
      objekt: 'weg/0',
      zone: 'friedhof',
      sitz: '',
    })
    expect(bedienung.menueFuer('weg/0')).toBeUndefined()
  })
})

// **Ein Leser stellt nichts** —
// #2482:
// Einzeltasten und Klick gingen bis dahin an der Leserprüfung vorbei, und der
// Kern wies erst danach ab. Eine Taste gilt dann nicht, ein Klick wählt nur.
describe('für einen Leser', () => {
  it('stellt keine Einzeltaste und wählt mit einem Klick nur', () => {
    const { befehlen, waehle, bedienung } = zeige([baer], true)

    expect(bedienung.taste('sitz1/baer/0', 't')).toBe(false)
    bedienung.greife('sitz1/baer/0', { detail: 1, metaKey: false })

    expect(befehlen).not.toHaveBeenCalled()
    expect(waehle).toHaveBeenCalledWith('sitz1/baer/0')
  })
})
