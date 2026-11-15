// **Das Menü „Exemplar“ der Leiste** — AP-6, Teil 2 (#2363): dieselben vier
// Handlungen wie im Inspector; Zum Tausch und Ins Deck reichen den Druck
// weiter wie die Knöpfe am Fuß, und für einen Leser sind Bearbeiten und
// Löschen grau (`docs/AUTORISIERUNG.md`: ein Exemplar ändern verlangt
// »Mitarbeiter«).
import { act, cleanup, render, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { Anschluss } from '../anschluss'
import type { Anschlussstelle } from '../anschlussraum'
import { ereignisObjekthandlung } from '../fenster'
import { Objekthandlungsraum } from '../objekthandlungsraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { platzhalter } from './sammlungszeilen'
import { Schreibrechtraum } from './schreibrecht'
import { useExemplarhandlungen } from './useExemplarhandlungen'

const eigene = kataloge.de.seiten.sammlung

let uebergeben: Anschluss | undefined

vi.mock('../anschlussraum', async (echt) => ({
  ...(await echt<typeof import('../anschlussraum')>()),
  useAnschluss: (): Anschlussstelle => ({
    offen: undefined,
    uebergib: (anschluss) => {
      uebergeben = anschluss
    },
    verbrauche: () => undefined,
  }),
}))

const exemplar = {
  ...platzhalter(0),
  id: 4,
  name: 'Aschenwächter',
  printing: 'p-4',
  zustand: 'played',
  sprache: 'de',
  finish: 'foil' as const,
}

function Inspector() {
  useExemplarhandlungen({
    exemplar,
    laeuft: false,
    offen: false,
    fragt: false,
    bearbeiten: vi.fn(),
    loeschen: vi.fn(),
  })

  return null
}

let laufzeit: Laufzeitprobe

beforeEach(() => {
  laufzeit = laufzeitprobe()
  uebergeben = undefined
})

afterEach(() => {
  cleanup()
  laufzeit.raeume()
  raeumeApi()
})

function zeige(nurLesend = false) {
  const melden = vi.fn(() => Promise.resolve({ fehler: null }))
  belegeApi({ ObjekthandlungenMelden: melden })
  render(
    <Sprachraum sprache="de">
      <Objekthandlungsraum>
        <Schreibrechtraum nurLesend={nurLesend}>
          <Inspector />
        </Schreibrechtraum>
      </Objekthandlungsraum>
    </Sprachraum>,
  )

  return melden
}

const klick = (schluessel: string) => {
  act(() => {
    laufzeit.melde(ereignisObjekthandlung, {
      art: 'exemplar',
      kennung: '4',
      epoche: 1,
      schluessel,
    })
  })
}

describe('das Menü „Exemplar“', () => {
  it('reicht den Druck zum Tausch weiter, samt Zustand, Sprache und Ausführung', async () => {
    zeige()
    await waitFor(() => {
      expect(laufzeit.horchende(ereignisObjekthandlung)).toBe(1)
    })

    klick('zumTausch')

    expect(uebergeben).toEqual({
      art: 'handel',
      name: 'Aschenwächter',
      printing: 'p-4',
      guete: 'played',
      sprache: 'de',
      finish: 'foil',
    })
  })

  it('reicht den Druck ins Deck weiter', async () => {
    zeige()
    await waitFor(() => {
      expect(laufzeit.horchende(ereignisObjekthandlung)).toBe(1)
    })

    klick('insDeck')

    expect(uebergeben).toEqual({
      art: 'deck',
      name: 'Aschenwächter',
      printing: 'p-4',
    })
  })

  it('meldet Bearbeiten und Löschen für einen Leser gesperrt', async () => {
    const melden = zeige(true)

    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(
        expect.objectContaining({
          titel: eigene.exemplarmenue,
          eintraege: [
            expect.objectContaining({
              schluessel: 'bearbeiten',
              gesperrt: true,
            }),
            expect.objectContaining({ schluessel: 'loeschen', gesperrt: true }),
            expect.objectContaining({
              schluessel: 'zumTausch',
              gesperrt: false,
            }),
            expect.objectContaining({ schluessel: 'insDeck', gesperrt: false }),
          ],
        }),
      )
    })
  })
})
