// **Die Menge im Inspector ist eine ganze Zahl ab eins** — F03
// (#2245).
//
// Bis dahin las `Number(wert)` das Feld bei jedem Tastendruck: ein leeres
// stand als 0 da, „2.5" ging als Kommazahl an ein `int` in Go. Das
// Aufnahmeformular derselben Seite hielt es schon so wie hier
// (`Sammlungsformular`, UX-06.AK-03).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { keineHerkunft } from '../apiherkunft'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Sammlungsinspector } from './Sammlungsinspector'

import type { ExemplarDTO } from '../api'

const texte = kataloge.de.seiten.sammlung

const dasExemplar: ExemplarDTO = {
  id: 1,
  fassung: 1,
  printing: 'probe-druck',
  druckID: '',
  vorlaeufig: false,
  menge: 3,
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
  lagerort: 'Kiste A',
  wunsch: false,
  bezeichnung: 'probe-druck',
  herkunft: keineHerkunft,
}

async function bearbeite() {
  belegeApi({
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
  })
  const aufAendern = vi.fn(() => Promise.resolve(true))

  render(
    <Sprachraum sprache="de">
      <Sammlungsinspector
        exemplar={dasExemplar}
        inSicht
        aufAendern={aufAendern}
        aufLoeschen={vi.fn()}
        aufGeholt={vi.fn()}
        laeuft={false}
      />
    </Sprachraum>,
  )

  await userEvent.click(screen.getByRole('button', { name: texte.bearbeiten }))

  return aufAendern
}

afterEach(() => {
  raeumeApi()
})

it.each(['0', '2.5', '-1'])(
  'hält das Sichern bei %j am Feld an',
  async (roh) => {
    const aufAendern = await bearbeite()
    const menge = screen.getByLabelText(texte.menge)

    await userEvent.clear(menge)
    await userEvent.type(menge, roh)

    expect(menge).toHaveAttribute('aria-invalid', 'true')
    expect(menge).toHaveAccessibleDescription(
      einsetzen(kataloge.de.allgemein.ganzeZahlAb, { kleinstens: '1' }),
    )
    expect(
      screen.getByRole('button', { name: texte.aenderungSichern }),
    ).toBeDisabled()
    expect(aufAendern).not.toHaveBeenCalled()
  },
)

it('lässt ein geleertes Feld leer und hält das Sichern an', async () => {
  await bearbeite()
  const menge = screen.getByLabelText(texte.menge)

  await userEvent.clear(menge)

  // **Leer und nicht 0** — vorher stand nach dem Löschen sofort eine Null da.
  expect(menge).toHaveValue(null)
  expect(
    screen.getByRole('button', { name: texte.aenderungSichern }),
  ).toBeDisabled()
})

// **Wer abbricht und neu öffnet, fängt beim Exemplar an** — auch die Sperre
// des Knopfs, die zum verworfenen Entwurf gehörte.
it('gibt das Sichern nach Abbrechen und neu Öffnen wieder frei', async () => {
  await bearbeite()

  await userEvent.clear(screen.getByLabelText(texte.menge))
  await userEvent.type(screen.getByLabelText(texte.menge), '0')
  await userEvent.click(screen.getByRole('button', { name: texte.abbrechen }))
  await userEvent.click(screen.getByRole('button', { name: texte.bearbeiten }))

  expect(screen.getByLabelText(texte.menge)).toHaveValue(3)
  expect(
    screen.getByRole('button', { name: texte.aenderungSichern }),
  ).toBeEnabled()
})
