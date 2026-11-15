// **Der Sprachraum trägt die Darstellung und die Zone** — LOC-03.AK-03
// (#1689).
//
// **Drei Angaben und nicht eine**: die Sprache der Texte, die Locale der
// Zahlen und die Zone der Uhr. Wer nur die erste nennt, bekommt für die
// zweite dieselbe und für die dritte die des Systems — so verhielt sich das
// Fenster bisher, und so bleibt es für jeden, der nichts einstellt.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Sprachraum, useDarstellung, useZeitzone } from './kontext'

function Probe() {
  return (
    <p>
      {useDarstellung()}|{useZeitzone() ?? 'System'}
    </p>
  )
}

it('reicht Darstellung und Zone weiter', () => {
  render(
    <Sprachraum sprache="de" darstellung="de-CH" zeitzone="Asia/Tokyo">
      <Probe />
    </Sprachraum>,
  )

  expect(screen.getByText('de-CH|Asia/Tokyo')).toBeInTheDocument()
})

it('nimmt ohne Angabe die Sprache und die Zone des Systems', () => {
  render(
    <Sprachraum sprache="en">
      <Probe />
    </Sprachraum>,
  )

  expect(screen.getByText('en|System')).toBeInTheDocument()
})
