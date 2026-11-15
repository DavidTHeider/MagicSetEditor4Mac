// **Der Hinweis, dass hier nur gelesen wird** — Werkstück 16 von Phase 4.12
// (#1351).
//
// **Gefragt wird seit G10 nicht mehr hier** (#2276): der Raum `Schreibrecht`
// fragt den Arbeitskontext, und Streifen wie Knöpfe lesen ihn. Wie gefragt
// wird — Fehlschlag, späte Antwort, Abbau —, steht in
// `haken/schreibrecht.test.tsx`.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Schreibrechtraum } from '../haken/schreibrecht'
import { Sprachraum } from '../sprache/kontext'
import { Nurlesehinweis } from './Nurlesehinweis'

function zeige(nurLesend: boolean) {
  render(
    <Sprachraum sprache="de">
      <Schreibrechtraum nurLesend={nurLesend}>
        <Nurlesehinweis />
        <p>daneben</p>
      </Schreibrechtraum>
    </Sprachraum>,
  )
}

describe('Nurlesehinweis', () => {
  it('sagt es, wenn nur gelesen werden darf', () => {
    zeige(true)

    expect(screen.getByRole('status')).toHaveTextContent(
      /als Leser: du kannst alles ansehen und nichts ändern/,
    )
  })

  // **Und schweigt sonst** (§6): ein Streifen, der immer dasteht, sagt nichts.
  it('schweigt, wenn geschrieben werden darf', () => {
    zeige(false)

    expect(screen.getByText('daneben')).toBeInTheDocument()
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })
})
