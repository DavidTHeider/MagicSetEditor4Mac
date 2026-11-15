// **Bearbeiten im Inspector in zwei Gruppen** — UXR-10
// (#1918),
// nach dem Artboard `Sammlung`: der Inspector zeigt Menge, Zustand und
// Lagerort. Nach *Bearbeiten* standen dreizehn Felder in einer Spalte von
// 320 Punkten untereinander (R3 in `docs/UIREVIEW.md` lässt zwölf zu) —
// dieselbe Teilung wie im Formular zum Aufnehmen hält sie klein.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { exemplar, felder } from '../test/sammlungsseite'
import { Sammlungsinspector } from './Sammlungsinspector'

const texte = kataloge.de.seiten.sammlung

afterEach(() => {
  raeumeApi()
})

async function bearbeite(waehrung = '') {
  belegeApi({
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
  })
  const aufAendern = vi.fn(() => Promise.resolve(true))

  render(
    <Sprachraum sprache="de">
      <Sammlungsinspector
        exemplar={exemplar({ lagerort: 'Kiste A', erwerbswaehrung: waehrung })}
        inSicht
        aufAendern={aufAendern}
        aufLoeschen={vi.fn(() => Promise.resolve(true))}
        aufGeholt={vi.fn()}
        laeuft={false}
      />
    </Sprachraum>,
  )
  await userEvent.click(screen.getByRole('button', { name: texte.bearbeiten }))

  return aufAendern
}

describe('Bearbeiten im Inspector', () => {
  it('zeigt Menge, Zustand und Lagerort zuerst und klappt das Übrige ein', async () => {
    await bearbeite()

    const offen = [
      texte.menge,
      texte.zustand,
      texte.lagerort,
      texte.druck,
      texte.sprache,
      texte.ausfuehrung,
    ]
    for (const feld of offen) {
      expect(screen.getByLabelText(feld)).toBeInTheDocument()
    }
    expect(felder()).toBe(offen.length)

    // **In der Reihenfolge des Artboards**: Menge, Zustand, Lagerort — und
    // erst dann, was den Druck beschreibt.
    const [menge, zustand, lagerort, druck] = [
      texte.menge,
      texte.zustand,
      texte.lagerort,
      texte.druck,
    ].map((feld) => screen.getByLabelText(feld))
    const vor = (eines: Element | undefined, anderes: Element | undefined) =>
      anderes !== undefined &&
      (eines?.compareDocumentPosition(anderes) ?? 0) &
        Node.DOCUMENT_POSITION_FOLLOWING
    expect(vor(menge, zustand)).toBeTruthy()
    expect(vor(zustand, lagerort)).toBeTruthy()
    expect(vor(lagerort, druck)).toBeTruthy()

    const knopf = screen.getByRole('button', { name: texte.weitereAngaben })
    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    await userEvent.click(knopf)

    for (const feld of [
      texte.name,
      texte.setcode,
      texte.nummer,
      texte.erwerbspreis,
      texte.erwerbswaehrung,
      texte.erwerbsdatum,
      texte.notiz,
    ]) {
      expect(screen.getByLabelText(feld)).toBeVisible()
    }
  })

  // **Ein Befund in einer zugeklappten Gruppe sähe niemand.**
  it('öffnet die weiteren Angaben, wenn der Preis zwei Lesarten hat', async () => {
    const aufAendern = await bearbeite('BHD')

    const knopf = screen.getByRole('button', { name: texte.weitereAngaben })
    await userEvent.click(knopf)
    await userEvent.type(screen.getByLabelText(texte.erwerbspreis), '1.234')
    await userEvent.click(knopf)

    await userEvent.click(
      screen.getByRole('button', { name: texte.aenderungSichern }),
    )

    expect(await screen.findByText(/1\.234,000/)).toBeVisible()
    expect(knopf).toHaveAttribute('aria-expanded', 'true')
    await waitFor(() => {
      expect(aufAendern).not.toHaveBeenCalled()
    })
  })
})
