// **Wohin ⌘Z zielt, wenn eine Zeichnung offen ist** —
// #1340.
//
// # Warum es diese Frage gibt
//
// MSE2 hat für das Symbol ein **eigenes Fenster** mit eigenem `ActionStack`
// (`src/gui/symbol/window.cpp`, Fassung `f9d9356` vom 2021-07-06); das ⌘Z des
// Setfensters erreicht es nie. Hier ist der Editor eine **Seite** im selben
// Fenster, und beide Verläufe hängen an demselben Menüeintrag.
//
// **Die offene Zeichnung geht vor**, solange es sie gibt. Der Verlauf des
// Bestands nimmt zurück, was **geschrieben** ist — und eine Zeichnung, an der
// gerade gearbeitet wird, ist es noch nicht.
import { Activity } from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import {
  Ruecknahmeraum,
  useRuecknahmeweg,
  useRuecknahmeziel,
} from './ruecknahme'

// Zeichner meldet ein Ziel an, solange er dasteht.
function Zeichner({ nimm }: { nimm: () => void }) {
  useRuecknahmeweg({ zuruecknehmen: nimm, wiederherstellen: () => undefined })

  return <p>zeichnet</p>
}

// Rahmen fragt das Ziel, so wie `Anwendung` es tut.
function Rahmen({
  amBestand,
  offen,
  nimm,
}: {
  amBestand: () => void
  offen: boolean
  nimm: () => void
}) {
  const ziel = useRuecknahmeziel()

  return (
    <div>
      {offen ? <Zeichner nimm={nimm} /> : null}
      <button
        type="button"
        onClick={() => {
          ;(ziel.gemeldetes()?.zuruecknehmen ?? amBestand)()
        }}
      >
        rückgängig
      </button>
    </div>
  )
}

describe('Ruecknahmeziel', () => {
  it('trifft den Bestand, solange keine Zeichnung offen ist', async () => {
    const amBestand = vi.fn()
    const nimm = vi.fn()

    render(
      <Ruecknahmeraum>
        <Rahmen amBestand={amBestand} offen={false} nimm={nimm} />
      </Ruecknahmeraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'rückgängig' }))

    expect(amBestand).toHaveBeenCalledTimes(1)
    expect(nimm).not.toHaveBeenCalled()
  })

  it('trifft die offene Zeichnung, sobald es eine gibt', async () => {
    const amBestand = vi.fn()
    const nimm = vi.fn()

    render(
      <Ruecknahmeraum>
        <Rahmen amBestand={amBestand} offen nimm={nimm} />
      </Ruecknahmeraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'rückgängig' }))

    expect(nimm).toHaveBeenCalledTimes(1)
    expect(amBestand).not.toHaveBeenCalled()
  })

  // **§7 Regel 1: ein Wechsel räumt den offenen Kontext.**
  //
  // Wer die Seite verlässt, lässt kein Ziel zurück, auf das später noch
  // gezeigt würde — sonst nähme ⌘Z auf der Kartenseite einen Handgriff an
  // einer Zeichnung zurück, die niemand mehr sieht.
  it('vergisst das Ziel, wenn die Zeichnung geschlossen wird', async () => {
    const amBestand = vi.fn()
    const nimm = vi.fn()

    const { rerender } = render(
      <Ruecknahmeraum>
        <Rahmen amBestand={amBestand} offen nimm={nimm} />
      </Ruecknahmeraum>,
    )

    rerender(
      <Ruecknahmeraum>
        <Rahmen amBestand={amBestand} offen={false} nimm={nimm} />
      </Ruecknahmeraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'rückgängig' }))

    expect(amBestand).toHaveBeenCalledTimes(1)
    expect(nimm).not.toHaveBeenCalled()
  })

  // **Eine verborgene Seite zieht ⌘Z nicht an sich** — P8 (#2376). `<Activity>`
  // hält besuchte Seiten im Baum und zeichnet sie verborgen neu; meldete sich
  // ein Weg dabei an, nähme ⌘Z auf einer anderen Seite am Tisch oder an einer
  // Zeichnung zurück, die niemand sieht.
  it('lässt eine verborgene Seite kein Ziel melden', async () => {
    const amBestand = vi.fn()
    const nimm = vi.fn()

    const bild = (sichtbar: boolean, weg: () => void) => (
      <Ruecknahmeraum>
        <Activity mode={sichtbar ? 'visible' : 'hidden'}>
          <Zeichner nimm={weg} />
        </Activity>
        <Rahmen amBestand={amBestand} offen={false} nimm={vi.fn()} />
      </Ruecknahmeraum>
    )

    const { rerender } = render(bild(true, nimm))
    rerender(bild(false, nimm))
    rerender(
      bild(false, () => {
        nimm()
      }),
    )

    await userEvent.click(screen.getByRole('button', { name: 'rückgängig' }))

    expect(amBestand).toHaveBeenCalledTimes(1)
    expect(nimm).not.toHaveBeenCalled()
  })

  // **Ohne Raum ein Ziel, das niemand fragt.** Dieselbe Zusage wie bei
  // `useAblage`: ein Bauteil, das außerhalb geprüft wird, soll nicht daran
  // scheitern.
  it('gibt auch ohne Raum ein Ziel', async () => {
    const amBestand = vi.fn()

    render(<Rahmen amBestand={amBestand} offen={false} nimm={vi.fn()} />)

    await userEvent.click(screen.getByRole('button', { name: 'rückgängig' }))

    expect(amBestand).toHaveBeenCalledTimes(1)
  })
})
