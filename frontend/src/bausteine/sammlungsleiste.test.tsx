// **Die Leiste über der Liste** — UXR-10
// (#1918),
// nach dem Artboard `Sammlung`: der Ansichtswechsel steht über der Liste,
// als Segmente. Bis hierher war er eine Auswahlliste mit drei Werten im
// Navigator — R7 in `docs/UIREVIEW.md` sagt Segmente bis vier Werte, und
// Sets und Decks tragen ihn längst so (`Listenansicht`).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sammlungsleiste } from './Sammlungsleiste'
import { leereSicht, type Sicht } from '../seiten/sammlungssicht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'

const texte = kataloge.de.seiten.sammlung

function zeige(teil: Partial<Sicht> = {}) {
  const setze = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Sammlungsleiste sicht={{ ...leereSicht, ...teil }} setze={setze} />
    </Sprachraum>,
  )

  return setze
}

describe('Die Leiste über der Liste', () => {
  it('wählt die Ansicht mit einem Segment', async () => {
    const setze = zeige()

    const gruppe = screen.getByRole('group', { name: texte.ansicht })
    expect(
      screen.getByRole('button', { name: texte.ansichtTabelle }),
    ).toHaveAttribute('aria-pressed', 'true')
    expect(gruppe).toContainElement(
      screen.getByRole('button', { name: texte.ansichtGruppen }),
    )

    await userEvent.click(
      screen.getByRole('button', { name: texte.ansichtGruppen }),
    )

    expect(setze).toHaveBeenCalledWith({ ansicht: 'gruppen' })
  })

  // **Ein zweiter Druck auf die gewählte Ansicht ändert nichts** — eine neue
  // Sicht räumt die Haken (`useSammlung`), und wer zweimal klickt, verlöre
  // sonst seine Auswahl.
  it('lässt die gewählte Ansicht stehen, statt sie neu zu setzen', async () => {
    const setze = zeige({ ansicht: 'galerie' })

    await userEvent.click(
      screen.getByRole('button', { name: texte.ansichtGalerie }),
    )

    expect(setze).not.toHaveBeenCalled()
  })

  it('sortiert über eine Auswahl', async () => {
    const setze = zeige()

    await userEvent.selectOptions(
      screen.getByLabelText(texte.sortierung),
      'menge',
    )

    expect(setze).toHaveBeenCalledWith({ sortierung: 'menge' })
  })

  it('lässt die Gruppierung außerhalb der Gruppenansicht ruhen', () => {
    zeige({ ansicht: 'tabelle' })

    expect(screen.getByLabelText(texte.gruppierung)).toBeDisabled()
  })

  it('wählt die Gruppierung in der Gruppenansicht', async () => {
    const setze = zeige({ ansicht: 'gruppen' })

    const feld = screen.getByLabelText(texte.gruppierung)
    expect(feld).toBeEnabled()
    await userEvent.selectOptions(feld, 'lagerort')

    expect(setze).toHaveBeenCalledWith({ gruppierung: 'lagerort' })
  })
})
