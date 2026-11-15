// **Der Preisverlauf bei einer fremden Quelle** — Werkstück A5 von Phase 5.3
// (#1444).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Fremdverlauf } from './Fremdverlauf'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Fremdpreisantwort } from '../antworten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.sammlung

// bilanz baut eine Antwort.
function bilanz(teil = {}): Fremdpreisantwort {
  return {
    quelle: 'scrydex',
    fassung: 'mtg-m11-149',
    zugeordnet: true,
    staende: 4,
    befunde: [],
    fehler: null,
    ...teil,
  }
}

async function klickeHolen() {
  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.verlaufHolenBei, { quelle: 'scrydex' }),
    }),
  )
}

function zeige(aufGeholt = () => undefined) {
  render(
    <Sprachraum sprache="de">
      <Fremdverlauf printing="probe-druck" aufGeholt={aufGeholt} />
    </Sprachraum>,
  )
}

it('holt den Verlauf und meldet die Bilanz nach oben', async () => {
  const gerufen = vi.fn(() => Promise.resolve(bilanz()))

  belegeApi({ Fremdpreisverlauf: gerufen })

  const geholt = vi.fn()
  zeige(geholt)

  await klickeHolen()

  expect(gerufen).toHaveBeenCalledWith('scrydex', 'probe-druck', '', [])

  expect(
    await screen.findByText(
      einsetzen(eigene.verlaufBilanz, {
        staende: '4',
        fassung: 'mtg-m11-149',
      }),
    ),
  ).toBeInTheDocument()

  expect(geholt).toHaveBeenCalled()
})

it('zeigt die Befunde der Quelle', async () => {
  belegeApi({
    Fremdpreisverlauf: () =>
      Promise.resolve(bilanz({ befunde: ['ein Tag, den niemand lesen kann'] })),
  })
  zeige()

  await klickeHolen()

  expect(
    await screen.findByText('ein Tag, den niemand lesen kann'),
  ).toBeInTheDocument()
})

// **Geschrieben ist geschrieben** (§7, Regel 3): die Bilanz steht auch dann
// da, wenn die Antwort einen Fehler trägt — was abgelegt ist, bleibt abgelegt.
it('zeigt einen Fehler am Ort und meldet nichts nach oben', async () => {
  belegeApi({
    Fremdpreisverlauf: () =>
      Promise.resolve(
        bilanz({
          staende: 0,
          fehler: { code: 'konflikt', meldung: 'zwei Fassungen' },
        }),
      ),
  })

  const geholt = vi.fn()
  zeige(geholt)

  await klickeHolen()

  expect(await screen.findByRole('alert')).toHaveTextContent('zwei Fassungen')
  expect(geholt).not.toHaveBeenCalled()
})

it('lässt keinen zweiten Abruf zu, solange der erste läuft', async () => {
  let loese: (wert: Fremdpreisantwort) => void = () => undefined

  const gerufen = vi.fn(
    () =>
      new Promise<Fremdpreisantwort>((aufloesen) => {
        loese = aufloesen
      }),
  )

  belegeApi({ Fremdpreisverlauf: gerufen })
  zeige()

  await klickeHolen()

  expect(await screen.findByText(eigene.verlaufLaeuft)).toBeInTheDocument()

  // **Den Abrufknopf und nicht „den einen"** — seit UX-15
  // (#1771)
  // steht neben der Erläuterung ein Hilfeknopf, und `getByRole('button')`
  // ohne Namen war damit eine Aussage über die Fläche statt über den Abruf
  // (§10).
  await klickeHolen()

  expect(gerufen).toHaveBeenCalledTimes(1)

  loese(bilanz())

  expect(
    await screen.findByText(
      einsetzen(eigene.verlaufBilanz, {
        staende: '4',
        fassung: 'mtg-m11-149',
      }),
    ),
  ).toBeInTheDocument()
})
