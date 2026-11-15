// **Eine Handlungsliste, auch in der Menüleiste** — GESTE (#2358).
//
// Geprüft wird, was nach Go geht — die Fassade ist die Naht, und
// `test/api.ts` antwortet eine Makroaufgabe später (§3) — und was ein Klick
// aus der Leiste bewirkt: **nur das, was zu Objekt und Epoche gehört**.
import { act, cleanup, render, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { Menueeintrag } from './bausteine/Menue'
import { ereignisObjekthandlung } from './fenster'
import { Schreibrechtraum } from './haken/schreibrecht'
import { useHandlungsliste } from './haken/useHandlungsliste'
import { keinObjekt } from './objekthandlungen'
import { Objekthandlungsraum } from './objekthandlungsraum'
import { belegeApi, raeumeApi } from './test/api'
import { laufzeitprobe, type Laufzeitprobe } from './test/laufzeitprobe'

let laufzeit: Laufzeitprobe

beforeEach(() => {
  laufzeit = laufzeitprobe()
})

afterEach(() => {
  cleanup()
  laufzeit.raeume()
  raeumeApi()
})

function Objekt({
  kennung,
  eintraege,
}: {
  kennung: string
  eintraege: readonly Menueeintrag[]
}) {
  useHandlungsliste({ titel: 'Karte', art: 'karte', kennung, eintraege })

  return null
}

function melder() {
  const ObjekthandlungenMelden = vi.fn(() => Promise.resolve({ fehler: null }))
  belegeApi({ ObjekthandlungenMelden })

  return ObjekthandlungenMelden
}

const klick = (kennung: string, epoche: number, schluessel: string) => {
  act(() => {
    laufzeit.melde(ereignisObjekthandlung, {
      art: 'karte',
      kennung,
      epoche,
      schluessel,
    })
  })
}

describe('Objekthandlungsraum', () => {
  it('meldet beim Einhängen, dass nichts gewählt ist', async () => {
    const melden = melder()

    render(<Objekthandlungsraum>{null}</Objekthandlungsraum>)

    await waitFor(() => {
      expect(melden).toHaveBeenCalledWith(keinObjekt)
    })
  })

  it('meldet die Liste des Objekts — angehalten und für Leser Schreibendes gesperrt', async () => {
    const melden = melder()
    const eintraege: Menueeintrag[] = [
      {
        schluessel: 'duplizieren',
        name: 'Duplizieren',
        schreibt: true,
        kuerzel: 'befehl+d',
        aufWahl: vi.fn(),
      },
      { schluessel: 'kopieren', name: 'Kopieren', aufWahl: vi.fn() },
      { schluessel: 'hoch', name: 'Nach oben', angehalten: true },
    ]

    render(
      <Objekthandlungsraum>
        <Schreibrechtraum nurLesend>
          <Objekt kennung="7" eintraege={eintraege} />
        </Schreibrechtraum>
      </Objekthandlungsraum>,
    )

    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith({
        titel: 'Karte',
        art: 'karte',
        kennung: '7',
        epoche: 1,
        eintraege: [
          {
            schluessel: 'duplizieren',
            name: 'Duplizieren',
            gesperrt: true,
            kuerzel: 'befehl+d',
          },
          { schluessel: 'kopieren', name: 'Kopieren', gesperrt: false },
          { schluessel: 'hoch', name: 'Nach oben', gesperrt: true },
        ],
      })
    })
  })

  it('führt einen Klick aus der Leiste aus, der zu Objekt und Epoche gehört', async () => {
    melder()
    const duplizieren = vi.fn()

    render(
      <Objekthandlungsraum>
        <Objekt
          kennung="7"
          eintraege={[
            {
              schluessel: 'duplizieren',
              name: 'Duplizieren',
              aufWahl: duplizieren,
            },
          ]}
        />
      </Objekthandlungsraum>,
    )
    await waitFor(() => {
      expect(laufzeit.horchende(ereignisObjekthandlung)).toBe(1)
    })

    klick('7', 1, 'duplizieren')

    expect(duplizieren).toHaveBeenCalledTimes(1)
  })

  it('verwirft einen Klick aus einer vorigen Epoche, für ein anderes Objekt oder auf Gesperrtes', async () => {
    const melden = melder()
    const vorher = vi.fn()
    const nachher = vi.fn()
    const gesperrt = vi.fn()

    // **Derselbe Baum, nur die Rolle wechselt** — sonst hängte React das
    // Objekt neu ein, und es begänne wieder bei Epoche 1.
    const { rerender } = render(
      <Objekthandlungsraum>
        <Schreibrechtraum nurLesend={false}>
          <Objekt
            kennung="7"
            eintraege={[
              {
                schluessel: 'duplizieren',
                name: 'Duplizieren',
                aufWahl: vorher,
              },
            ]}
          />
        </Schreibrechtraum>
      </Objekthandlungsraum>,
    )
    rerender(
      <Objekthandlungsraum>
        <Schreibrechtraum nurLesend>
          <Objekt
            kennung="7"
            eintraege={[
              {
                schluessel: 'duplizieren',
                name: 'Duplizieren!',
                aufWahl: nachher,
              },
              {
                schluessel: 'loeschen',
                name: 'Löschen',
                schreibt: true,
                aufWahl: gesperrt,
              },
            ]}
          />
        </Schreibrechtraum>
      </Objekthandlungsraum>,
    )
    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(
        expect.objectContaining({ epoche: 2 }),
      )
    })

    // **Ein Beleg, dass jemand zuhört** (§3): derselbe Weg mit der heutigen
    // Epoche wirkt.
    klick('7', 2, 'duplizieren')
    expect(nachher).toHaveBeenCalledTimes(1)

    klick('7', 1, 'duplizieren')
    klick('8', 2, 'duplizieren')
    klick('7', 2, 'loeschen')
    klick('7', 2, 'unbekannt')
    act(() => {
      laufzeit.melde(ereignisObjekthandlung, 'kein Verbund')
    })

    expect(vorher).not.toHaveBeenCalled()
    expect(nachher).toHaveBeenCalledTimes(1)
    expect(gesperrt).not.toHaveBeenCalled()
  })

  it('gibt dem jüngsten Objekt die Leiste und dem vorigen sie zurück', async () => {
    const melden = melder()

    const { rerender } = render(
      <Objekthandlungsraum>
        <Objekt
          kennung="7"
          eintraege={[{ schluessel: 'a', name: 'A', aufWahl: vi.fn() }]}
        />
        <Objekt
          kennung="9"
          eintraege={[{ schluessel: 'b', name: 'B', aufWahl: vi.fn() }]}
        />
      </Objekthandlungsraum>,
    )
    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(
        expect.objectContaining({ kennung: '9' }),
      )
    })

    // **Eine Aktualisierung behält ihren Platz**: das vordere Objekt zeichnet
    // neu, und die Leiste bleibt beim jüngsten.
    rerender(
      <Objekthandlungsraum>
        <Objekt
          kennung="7"
          eintraege={[{ schluessel: 'a', name: 'A neu', aufWahl: vi.fn() }]}
        />
        <Objekt
          kennung="9"
          eintraege={[{ schluessel: 'b', name: 'B', aufWahl: vi.fn() }]}
        />
      </Objekthandlungsraum>,
    )
    await act(async () => {
      await new Promise((fertig) => setTimeout(fertig, 0))
    })
    expect(melden).toHaveBeenLastCalledWith(
      expect.objectContaining({ kennung: '9' }),
    )

    rerender(
      <Objekthandlungsraum>
        <Objekt
          kennung="7"
          eintraege={[{ schluessel: 'a', name: 'A', aufWahl: vi.fn() }]}
        />
      </Objekthandlungsraum>,
    )
    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(
        expect.objectContaining({ kennung: '7' }),
      )
    })

    rerender(<Objekthandlungsraum>{null}</Objekthandlungsraum>)
    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(keinObjekt)
    })
  })

  it('meldet dieselbe Liste nicht zweimal und übersteht einen gescheiterten Ruf', async () => {
    const ObjekthandlungenMelden = vi.fn(() =>
      Promise.reject(new Error('gesperrt')),
    )
    belegeApi({ ObjekthandlungenMelden })
    const eintraege: Menueeintrag[] = [
      { schluessel: 'a', name: 'A', aufWahl: vi.fn() },
    ]

    const { rerender } = render(
      <Objekthandlungsraum>
        <Objekt kennung="7" eintraege={eintraege} />
      </Objekthandlungsraum>,
    )
    await waitFor(() => {
      expect(ObjekthandlungenMelden).toHaveBeenCalledTimes(2)
    })
    rerender(
      <Objekthandlungsraum>
        <Objekt kennung="7" eintraege={[...eintraege]} />
      </Objekthandlungsraum>,
    )
    await act(async () => {
      await new Promise((fertig) => setTimeout(fertig, 0))
    })

    expect(ObjekthandlungenMelden).toHaveBeenCalledTimes(2)
  })

  it('tut außerhalb eines Raums nichts', () => {
    expect(() => render(<Objekt kennung="7" eintraege={[]} />)).not.toThrow()
  })
})
