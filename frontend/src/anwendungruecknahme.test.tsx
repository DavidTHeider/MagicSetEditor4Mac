// **⌘Z trifft die offene Zeichnung** —
// #1340,
// Werkstück F von Phase 4.11.
//
// # Warum dieser Lauf und nicht nur der am Haken
//
// `ruecknahme.test.tsx` prüft die Stelle, `symbolruecknahme.test.tsx` die
// Knöpfe der Seite. **Keiner der beiden prüft den Weg vom Menü dorthin** —
// und der ist die Zusage aus dem Issue: nicht „es gibt Knöpfe", sondern ⌘Z.
//
// Zwei grüne Seiten einer Grenze sind kein Nachweis für die Grenze.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import {
  Ruecknahmeraum,
  useRuecknahmeweg,
  useRuecknahmeziel,
} from './ruecknahme'
import { ruecknahmeweg } from './ruecknahmeweg'

// Zeichner ist die Seite, die einen eigenen Verlauf mitbringt.
function Zeichner({ nimm }: { nimm: () => void }) {
  useRuecknahmeweg({ zuruecknehmen: nimm, wiederherstellen: () => undefined })

  return <p>zeichnet</p>
}

// Rahmen ist, was `Anwendung` tut: den Menübefehl auf den richtigen Verlauf
// legen.
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
  const handlungen = ruecknahmeweg(ziel, {
    zuruecknehmen: amBestand,
    wiederherstellen: () => undefined,
  })

  return (
    <div>
      {offen ? <Zeichner nimm={nimm} /> : null}
      <button type="button" onClick={handlungen.zuruecknehmen}>
        menü
      </button>
    </div>
  )
}

describe('Der Menübefehl Rückgängig', () => {
  it('trifft den Bestand, solange keine Zeichnung offen ist', async () => {
    const amBestand = vi.fn()
    const nimm = vi.fn()

    render(
      <Ruecknahmeraum>
        <Rahmen amBestand={amBestand} offen={false} nimm={nimm} />
      </Ruecknahmeraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'menü' }))

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

    await userEvent.click(screen.getByRole('button', { name: 'menü' }))

    expect(nimm).toHaveBeenCalledTimes(1)
    expect(amBestand).not.toHaveBeenCalled()
  })

  // **Die Entscheidung fällt beim Aufruf, nicht beim Zeichnen.**
  //
  // Ein `ziel.gemeldetes()` im Rumpf der Komponente läse den Stand von
  // **vorhin**: der Zeichner meldet sich beim Zeichnen an, und der Rahmen ist
  // über ihm — er zeichnet zuerst. Wer die Entscheidung dort trifft, greift
  // beim ersten Anstrich noch am Bestand vorbei.
  it('fragt bei jedem Aufruf neu, wohin er zielt', async () => {
    const amBestand = vi.fn()
    const nimm = vi.fn()

    const { rerender } = render(
      <Ruecknahmeraum>
        <Rahmen amBestand={amBestand} offen nimm={nimm} />
      </Ruecknahmeraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'menü' }))

    rerender(
      <Ruecknahmeraum>
        <Rahmen amBestand={amBestand} offen={false} nimm={nimm} />
      </Ruecknahmeraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'menü' }))

    expect(nimm).toHaveBeenCalledTimes(1)
    expect(amBestand).toHaveBeenCalledTimes(1)
  })

  it('gilt für die Wiederherstellung genauso', async () => {
    const zurueck = vi.fn()
    const vor = vi.fn()

    function Zweiter() {
      useRuecknahmeweg({ zuruecknehmen: zurueck, wiederherstellen: vor })

      return null
    }

    function Rahmen2() {
      const ziel = useRuecknahmeziel()
      const handlungen = ruecknahmeweg(ziel, {
        zuruecknehmen: () => undefined,
        wiederherstellen: () => undefined,
      })

      return (
        <div>
          <Zweiter />
          <button type="button" onClick={handlungen.wiederherstellen}>
            vor
          </button>
        </div>
      )
    }

    render(
      <Ruecknahmeraum>
        <Rahmen2 />
      </Ruecknahmeraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'vor' }))

    expect(vor).toHaveBeenCalledTimes(1)
    expect(zurueck).not.toHaveBeenCalled()
  })
})
