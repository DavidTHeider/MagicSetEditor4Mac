// **Ein Erläuterungsabsatz, der sich kurz fasst** — UX-15.AK-01 bis AK-03
// (#1771).
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it } from 'vitest'

import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.bausteine

const lang =
  'Der erste Satz hilft. Der zweite begründet, und den liest man einmal.'

function zeige(text: string) {
  render(
    <Sprachraum sprache="de">
      <Erklaerungsabsatz text={text} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

it('zeigt den ersten Satz', () => {
  zeige(lang)

  expect(screen.getByText('Der erste Satz hilft.')).toBeInTheDocument()
})

it('hält den Rest hinter einem Fragezeichen', async () => {
  const hand = zeige(lang)

  await hand.tab()

  expect(await screen.findByRole('tooltip')).toHaveTextContent(
    'Der zweite begründet, und den liest man einmal.',
  )

  // **Der Knopf heißt nach dem Satz, den er fortsetzt** — bei dreißig
  // solchen Absätzen wäre „Mehr dazu" dreißigmal derselbe Name (WCAG 2.4.6).
  expect(
    screen.getByRole('button', {
      name: einsetzen(eigene.mehrZu, { satz: 'Der erste Satz hilft.' }),
    }),
  ).toBeInTheDocument()
})

// **Ein kurzer Absatz bekommt kein Fragezeichen** — es öffnete eine leere
// Erklärung (§10).
it('zeigt kein Fragezeichen für einen einzelnen Satz', () => {
  zeige('Nur ein Satz.')

  expect(screen.queryByRole('button')).not.toBeInTheDocument()
})

// **Die Auszeichnung der Kataloge wird aufgelöst** — sonst stünden die
// Sternchen da, wie unter „Nach neuen Fassungen sehen" (#1100).
it('zeichnet aus, statt Sternchen zu zeigen', () => {
  zeige('Das ist **wichtig**. Und das hier begründet es.')

  expect(screen.getByText('wichtig', { selector: 'strong' })).toBeTruthy()
  expect(screen.queryByText(/\*\*/)).toBeNull()
})
