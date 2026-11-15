// **Rechtsklick und ⇧F10 an einer Zeile der Sammlung** — AP-6, Teil 2
// (#2363):
// dieselben Handlungen wie im Inspector — Bearbeiten (dort steht die Menge
// zuerst), Löschen mit seiner Rückfrage, Zum Tausch und Ins Deck —, an der
// Zeile selbst. Eine andere Zeile wird zuerst aufgeschlagen, und das Menü
// gehört dann ihr (`useAnstoss`, wie in der Kartenliste).
import { keineHerkunft } from '../apiherkunft'
import { act, fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { ExemplarDTO } from '../api'
import { Objekthandlungsraum } from '../objekthandlungsraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Sammlung } from './Sammlung'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.sammlung

function exemplar(teile: Partial<ExemplarDTO>): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: '8a1b0d5e',
    druckID: '',
    vorlaeufig: false,
    menge: 1,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: '',
    setcode: '',
    nummer: '',
    lagerort: '',
    wunsch: false,
    bezeichnung: '8a1b0d5e',
    herkunft: keineHerkunft,
    ...teile,
  }
}

const bolt = exemplar({
  id: 1,
  name: 'Lightning Bolt',
  bezeichnung: 'Lightning Bolt · M11 #149',
})
const golem = exemplar({
  id: 2,
  printing: 'c0ffee',
  name: 'Iron Golem',
  bezeichnung: 'Iron Golem · M11 #210',
})

function zeige() {
  const ExemplarLoeschen = vi.fn(() => Promise.resolve(null))
  const exemplare = [bolt, golem]
  belegeApi({
    Sammlung: () => Promise.resolve({ exemplare, fehler: null }),
    Sammlungsfenster: () =>
      Promise.resolve({
        exemplare,
        gesamt: exemplare.length,
        insgesamt: exemplare.length,
        stueck: exemplare.length,
        fehler: null,
      }),
    Portfolio: () =>
      Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    ObjekthandlungenMelden: () => Promise.resolve({ fehler: null }),
    ExemplarLoeschen,
  })

  render(
    <Sprachraum sprache="de">
      <Objekthandlungsraum>
        <Sammlung bestandGeaendert={vi.fn()} />
      </Objekthandlungsraum>
    </Sprachraum>,
  )

  return { ExemplarLoeschen }
}

const zeile = (bezeichnung: string) =>
  screen.getByRole('row', { name: new RegExp(bezeichnung) })

const inspector = () =>
  within(screen.getByRole('complementary', { name: eigene.inspector }))

describe('der Rechtsklick an einer Zeile der Sammlung', () => {
  it('schlägt eine andere Zeile zuerst auf und zeigt dann ihre Handlungen', async () => {
    zeige()
    await screen.findByText(golem.bezeichnung)

    fireEvent.contextMenu(zeile(golem.bezeichnung), {
      clientX: 20,
      clientY: 20,
    })

    const menue = await screen.findByRole('menu', {
      name: eigene.exemplarmenue,
    })
    expect(inspector().getByText(golem.bezeichnung)).toBeInTheDocument()
    const eintraege = within(menue).getAllByRole('menuitem')
    expect(eintraege).toHaveLength(4)
    ;[
      eigene.bearbeiten,
      eigene.loeschen,
      eigene.zumTausch,
      eigene.insDeck,
    ].forEach((name, stelle) => {
      expect(eintraege[stelle]).toHaveAccessibleName(name)
    })
  })

  // **Löschen fragt im Inspector nach**, wie sein Knopf (UX-10.AK-04).
  it('lässt Löschen fragen, ohne zu löschen', async () => {
    const { ExemplarLoeschen } = zeige()
    await screen.findByText(bolt.bezeichnung)

    fireEvent.contextMenu(zeile(bolt.bezeichnung), {
      clientX: 20,
      clientY: 20,
    })
    await userEvent.click(
      await screen.findByRole('menuitem', { name: eigene.loeschen }),
    )

    expect(inspector().getByText(eigene.loeschenFrage)).toBeInTheDocument()
    expect(ExemplarLoeschen).not.toHaveBeenCalled()
  })

  // **Bearbeiten öffnet das Formular, die Menge zuerst** (UXR-10).
  it('öffnet mit Bearbeiten das Formular samt Menge', async () => {
    zeige()
    await screen.findByText(bolt.bezeichnung)

    fireEvent.contextMenu(zeile(bolt.bezeichnung), {
      clientX: 20,
      clientY: 20,
    })
    await userEvent.click(
      await screen.findByRole('menuitem', { name: eigene.bearbeiten }),
    )

    expect(inspector().getByLabelText(eigene.menge)).toHaveValue(1)
  })

  it('geht mit ⇧F10 an der gehaltenen Zeile auf', async () => {
    zeige()
    await screen.findByText(bolt.bezeichnung)

    act(() => {
      zeile(bolt.bezeichnung).focus()
    })
    fireEvent.keyDown(zeile(bolt.bezeichnung), { key: 'F10', shiftKey: true })

    expect(
      await screen.findByRole('menu', { name: eigene.exemplarmenue }),
    ).toBeInTheDocument()
  })
})
