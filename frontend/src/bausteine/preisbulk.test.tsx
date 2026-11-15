// **Preise einer ganzen Quelle holen** — Werkstück A3 von Phase 5.3
// (#1442).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Preisbulk } from './Preisbulk'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.sammlung

// zeige zeichnet das Bauteil.
function zeige(aufGeholt = () => undefined) {
  render(
    <Sprachraum sprache="de">
      <Preisbulk aufGeholt={aufGeholt} />
    </Sprachraum>,
  )
}

// bilanz baut eine Antwort.
function bilanz(teil = {}) {
  return {
    quelle: 'mtgjson',
    umfang: 'heute',
    drucke: 3,
    staende: 12,
    befunde: [],
    fehler: null,
    ...teil,
  }
}

it('holt für die gewählte Quelle und den gewählten Umfang', async () => {
  const gerufen = vi.fn(() => Promise.resolve(bilanz()))

  belegeApi({ Preisbulk: gerufen })

  const geholt = vi.fn()
  zeige(geholt)

  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.bulkHolenBei, { quelle: 'mtgjson' }),
    }),
  )

  await waitFor(() => {
    expect(gerufen).toHaveBeenCalledWith('mtgjson', 'heute')
  })

  // **Und das Portfolio rechnet neu** — sonst stünde die alte Summe neben
  // frischen Preisen.
  await waitFor(() => {
    expect(geholt).toHaveBeenCalled()
  })

  expect(await screen.findByText(/12 Preise zu 3 Drucken/)).toBeInTheDocument()
})

it('nimmt den Umfang, den jemand wählt', async () => {
  const gerufen = vi.fn(() =>
    Promise.resolve(bilanz({ umfang: 'neunzigTage' })),
  )

  belegeApi({ Preisbulk: gerufen })
  zeige()

  await userEvent.selectOptions(
    screen.getByLabelText(eigene.bulkUmfang),
    'neunzigTage',
  )

  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.bulkHolenBei, { quelle: 'mtgjson' }),
    }),
  )

  await waitFor(() => {
    expect(gerufen).toHaveBeenCalledWith('mtgjson', 'neunzigTage')
  })
})

// **Ohne Druck mit MTGJSON-Kennung sagt es das — und nicht »0 Preise«.**
it('sagt es, wenn kein Druck gefragt werden konnte', async () => {
  belegeApi({
    Preisbulk: () => Promise.resolve(bilanz({ drucke: 0, staende: 0 })),
  })

  zeige()

  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.bulkHolenBei, { quelle: 'mtgjson' }),
    }),
  )

  expect(await screen.findByText(eigene.bulkOhneDrucke)).toBeInTheDocument()
})

// **Ein Fehler steht am Bauteil und nicht im Nichts** (§6).
it('zeigt einen Fehler des Bezugs', async () => {
  belegeApi({
    Preisbulk: () =>
      Promise.resolve(
        bilanz({
          fehler: { code: 'validierung', meldung: 'der Dienst ist aus' },
        }),
      ),
  })

  zeige()

  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.bulkHolenBei, { quelle: 'mtgjson' }),
    }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'der Dienst ist aus',
  )
})

// **Die Befunde der Quelle stehen da** (§8).
it('nennt die Befunde der Quelle', async () => {
  belegeApi({
    Preisbulk: () =>
      Promise.resolve(bilanz({ befunde: ['einen Markt kennt sie nicht'] })),
  })

  zeige()

  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.bulkHolenBei, { quelle: 'mtgjson' }),
    }),
  )

  expect(
    await screen.findByText('einen Markt kennt sie nicht'),
  ).toBeInTheDocument()
})
