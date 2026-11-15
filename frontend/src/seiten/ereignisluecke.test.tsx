// **Ein Verlauf mit Lücke sagt es** — Z3 aus dem Zweitblick über Phase 6.11
// (#1672). War die Seite so lange verborgen, dass über der Nachholgrenze Züge
// wegfielen, sähe die Liste sonst vollständig aus.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Ereignisliste } from './Ereignisliste'
import { einLauf } from './partieprobe'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen

function zeige(ereignisluecke: boolean) {
  render(
    <Sprachraum sprache="de">
      <Ereignisliste lauf={einLauf({ ereignisluecke })} />
    </Sprachraum>,
  )
}

it('nennt die Lücke über der Liste', () => {
  zeige(true)

  expect(screen.getByText(eigene.ereignisseLuecke)).toBeInTheDocument()
})

it('schweigt ohne Lücke', () => {
  zeige(false)

  // **Und die Liste steht da** — dass gezeichnet wurde (§3).
  expect(screen.getByText(eigene.ereignisseLeer)).toBeInTheDocument()
  expect(screen.queryByText(eigene.ereignisseLuecke)).not.toBeInTheDocument()
})
