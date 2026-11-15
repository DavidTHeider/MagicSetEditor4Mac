// **Angebote und fremder Verlauf eingeklappt** — UXR-10
// (#1918).
//
// Mit einem gehaltenen Exemplar standen im Inspector zwölf Bedienelemente,
// fünf davon für zwei Abrufe ins Netz und ihre Erklärungen — die Angebote
// eines Marktplatzes und der Verlauf einer fremden Quelle. Das Artboard
// `Sammlung` zeigt im Inspector keinen davon. Der eigene Preis bleibt
// stehen; was hinaus fragt, steht unter einem Knopf.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Preise } from './Preise'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

const texte = kataloge.de.seiten.sammlung

afterEach(raeumeApi)

it('klappt Angebote und fremden Verlauf ein und fragt erst offen', async () => {
  const links = vi.fn(() => Promise.resolve({ links: [], fehler: null }))

  belegeApi({
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    Marktlinks: links,
  })

  render(
    <Sprachraum sprache="de">
      <Preise printing="probe-druck" aufGeholt={vi.fn()} />
    </Sprachraum>,
  )

  // **Der eigene Preis steht da** — samt dem Knopf, der ihn holt.
  expect(await screen.findByText(texte.keinePreise)).toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: `${texte.preiseHolen} probe-druck` }),
  ).toBeInTheDocument()

  const knopf = screen.getByRole('button', { name: texte.marktUndVerlauf })
  expect(knopf).toHaveAttribute('aria-expanded', 'false')
  expect(screen.queryByRole('button', { name: /Angebote bei/ })).toBeNull()
  expect(screen.queryByRole('button', { name: /Verlauf bei/ })).toBeNull()
  expect(links).not.toHaveBeenCalled()

  await userEvent.click(knopf)

  expect(
    screen.getByRole('button', { name: /Angebote bei/ }),
  ).toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: /Verlauf bei/ }),
  ).toBeInTheDocument()
  await waitFor(() => {
    expect(links).toHaveBeenCalled()
  })
})
