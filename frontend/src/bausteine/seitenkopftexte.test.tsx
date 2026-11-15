// **Der Seitenkopf sagt kurz, wofür die Seite da ist** — UX-15.AK-01 bis
// AK-03 (#1771).
//
// # Der Befund
//
// **Gemessen am 2026-09-22** über die Seitenkataloge: die Kopferklärungen
// reichen bis **340** Zeichen — vier bis fünf Zeilen über dem, was man
// tatsächlich sucht. Unter *Einstellungen* ist der Kopf länger als das erste
// Feld darunter.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it } from 'vitest'

import { Seitenkopf } from './Seitenkopf'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.bausteine

const lang =
  'Der erste Satz hilft beim Entscheiden. Der zweite begründet, und den liest man einmal.'

function zeige(erklaerung: string) {
  render(
    <Sprachraum sprache="de">
      <Seitenkopf titel="Sets" erklaerung={erklaerung} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// **Sichtbar steht der erste Satz.**
it('zeigt den ersten Satz und nicht den ganzen Absatz', () => {
  zeige(lang)

  expect(
    screen.getByText('Der erste Satz hilft beim Entscheiden.'),
  ).toBeInTheDocument()

  expect(
    screen.queryByText(/Der zweite begründet/, { selector: 'p' }),
  ).not.toBeInTheDocument()
})

// **Und der Rest hinter dem Fragezeichen — an dieser Seite, nicht an einer
// beliebigen.**
it('hält den Rest hinter einem Fragezeichen zur Seite', async () => {
  const hand = zeige(lang)

  const knopf = screen.getByRole('button', {
    name: einsetzen(eigene.feldhilfe, { feld: 'Sets' }),
  })

  await hand.tab()

  expect(await screen.findByRole('tooltip')).toHaveTextContent(
    'Der zweite begründet, und den liest man einmal.',
  )

  expect(knopf).toHaveAttribute(
    'aria-describedby',
    screen.getByRole('tooltip').id,
  )
})

// **Ein kurzer Kopf bekommt kein Fragezeichen** — es öffnete eine leere
// Erklärung (§10).
it('zeigt kein Fragezeichen, wo es nichts zu zeigen gibt', () => {
  zeige('Nur ein Satz.')

  expect(screen.queryByRole('button')).not.toBeInTheDocument()
  expect(screen.getByText('Nur ein Satz.')).toBeInTheDocument()
})

// **Und ohne Erklärung steht gar nichts** — ein leerer Absatz schöbe den
// Inhalt nach unten.
it('zeigt nichts, wo es keine Erklärung gibt', () => {
  render(
    <Sprachraum sprache="de">
      <Seitenkopf titel="Sets" />
    </Sprachraum>,
  )

  expect(screen.queryByRole('button')).not.toBeInTheDocument()
  expect(screen.getByRole('heading', { name: 'Sets' })).toBeInTheDocument()
})
