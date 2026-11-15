// **Das Hilfesymbol am Feld** — UX-15.AK-02
// (#1771).
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it } from 'vitest'

import { Feldhilfe } from './Feldhilfe'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.bausteine

function zeige(text = 'Der lange Teil der Erklärung.') {
  render(
    <Sprachraum sprache="de">
      <Feldhilfe zu="Sprache" text={text} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// **Der Knopf sagt, wozu er gehört** — WCAG 2.4.6: „Hilfe" allein sagt bei
// dreißig Feldern nicht, zu welchem.
it('nennt das Feld, zu dem die Hilfe gehört', () => {
  zeige()

  expect(
    screen.getByRole('button', {
      name: einsetzen(eigene.feldhilfe, { feld: 'Sprache' }),
    }),
  ).toBeInTheDocument()
})

// **Über die Tastatur und nicht nur über den Zeiger** — AK-02 wörtlich.
it('zeigt die Erklärung beim Fokussieren', async () => {
  const hand = zeige()

  await hand.tab()

  expect(await screen.findByRole('tooltip')).toHaveTextContent(
    'Der lange Teil der Erklärung.',
  )
})

// **Und eine Vorlesehilfe bekommt sie als Beschreibung des Knopfes** — sonst
// stünde der Text im Baum und niemand hörte ihn.
it('nennt die Erklärung als Beschreibung', async () => {
  const hand = zeige()

  await hand.tab()

  const knopf = screen.getByRole('button', {
    name: einsetzen(eigene.feldhilfe, { feld: 'Sprache' }),
  })

  expect(knopf.getAttribute('aria-describedby')).toBe(
    (await screen.findByRole('tooltip')).id,
  )
})
