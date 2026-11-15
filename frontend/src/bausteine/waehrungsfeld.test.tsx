// **Ein Währungsfeld schlägt vor, urteilt wie Go und lässt den Rohtext stehen**
// — UX-06.AK-01 bis AK-04
// (#1678).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { expect, it } from 'vitest'

import { Waehrungsfeld } from './Waehrungsfeld'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.bausteine.waehrung

// Feld hält den Wert wie ein Aufrufer — der Rohtext soll stehen bleiben.
function Feld({
  kleinErlaubt,
  befund,
}: {
  kleinErlaubt: boolean
  befund?: string
}) {
  const [wert, setzeWert] = useState('')

  return (
    <Waehrungsfeld
      beschriftung="Währung"
      wert={wert}
      setzeWert={setzeWert}
      kleinErlaubt={kleinErlaubt}
      befund={befund}
    />
  )
}

function zeige(kleinErlaubt: boolean, befund?: string) {
  render(
    <Sprachraum sprache="de">
      <Feld
        kleinErlaubt={kleinErlaubt}
        {...(befund === undefined ? {} : { befund })}
      />
    </Sprachraum>,
  )

  return screen.getByLabelText('Währung')
}

it('schlägt die Währungen der Plattform mit ihrem Namen vor', () => {
  const feld = zeige(true)

  const liste = document.getElementById(feld.getAttribute('list') ?? '')
  const euro = liste?.querySelector('option[value="EUR"]')

  expect(euro?.getAttribute('label')).toBe('Euro')
})

it('urteilt erst, wenn das Feld verlassen ist — und lässt den Rohtext stehen', async () => {
  const feld = zeige(true)

  await userEvent.type(feld, 'EURO')
  expect(screen.queryByRole('alert')).toBeNull()

  await userEvent.tab()

  expect(screen.getByRole('alert')).toHaveTextContent(
    einsetzen(texte.keineGestalt, { wert: 'EURO' }),
  )
  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(feld).toHaveValue('EURO')
})

it('nimmt einen gültigen Code ohne Befund', async () => {
  const feld = zeige(true)

  await userEvent.type(feld, 'eur')
  await userEvent.tab()

  expect(screen.queryByRole('alert')).toBeNull()
  expect(feld).toHaveAttribute('aria-invalid', 'false')
})

it('sagt, wie der Code heißt, wo klein nicht angenommen wird', async () => {
  const feld = zeige(false)

  await userEvent.type(feld, 'eur')
  await userEvent.tab()

  expect(screen.getByRole('alert')).toHaveTextContent(
    einsetzen(texte.grossSchreiben, { wert: 'eur', gross: 'EUR' }),
  )
})

// **Wer wieder tippt, bekommt keinen alten Satz** — das Urteil gehört zu
// dem, was beim Verlassen dastand.
it('nimmt den Satz beim Weitertippen zurück', async () => {
  const feld = zeige(true)

  await userEvent.type(feld, '€')
  await userEvent.tab()
  expect(screen.getByRole('alert')).toBeInTheDocument()

  await userEvent.type(feld, 'X')

  expect(screen.queryByRole('alert')).toBeNull()
})

it('lässt den Befund der Anwendung vor dem eigenen stehen', async () => {
  const feld = zeige(true, 'Zu einem Preis gehört seine Währung.')

  await userEvent.type(feld, '€')
  await userEvent.tab()

  expect(screen.getByRole('alert')).toHaveTextContent(
    'Zu einem Preis gehört seine Währung.',
  )
})
