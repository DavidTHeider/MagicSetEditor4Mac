// **Ein Feld für eine Kennung lässt macOS nichts berichtigen** — #2389.
//
// Gefunden bei der Ausgangsmessung P0 (#2336): getippt war `p0-blitz`, im
// Handel stand `P0-blitz`, und das eigene Exemplar passte nicht mehr. Ein
// Feld, dessen Wert genau so ankommen muss, wie er getippt wurde, trägt
// deshalb `autocapitalize`, `autocorrect` und `spellcheck` aus
// ([`kennungseingabe`](./kennungseingabe.ts)); welche das sind, führt
// `docs/FELDPRUEFUNG.md`, Abschnitt „Kennungsfelder", und
// `kennungsfelder.test.ts` hält die Liste.
import { render, screen } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { dieLeereZeile, Handelspositionsform } from './Handelspositionsform'

function istKennungsfeld(feld: HTMLElement) {
  expect(feld).toHaveAttribute('autocapitalize', 'off')
  expect(feld).toHaveAttribute('autocorrect', 'off')
  expect(feld).toHaveAttribute('spellcheck', 'false')
}

describe('ein Kennungsfeld', () => {
  test('der Druck einer Position', () => {
    // **Mit einer Vorlage, die eine Kennung trägt, steht der Kasten offen**
    // — seit P10 (#2378) steht das Feld nicht mehr im Hauptweg.
    render(
      <Handelspositionsform
        vorlage={{ ...dieLeereZeile, printing: 'p-1' }}
        uebernehmen={() => undefined}
        verwerfen={() => undefined}
      />,
    )

    istKennungsfeld(screen.getByLabelText('Druck'))
  })

  test('die Sprache einer Position', () => {
    render(
      <Handelspositionsform
        vorlage={dieLeereZeile}
        uebernehmen={() => undefined}
        verwerfen={() => undefined}
      />,
    )

    istKennungsfeld(screen.getByRole('combobox', { name: 'Sprache' }))
  })
})
