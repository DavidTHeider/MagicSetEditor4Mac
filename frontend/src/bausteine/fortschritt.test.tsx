// **Ein Lauf, den man hören kann** — Werkstück U3 von Phase 5.1
// (#1414).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Fortschritt } from './Fortschritt'

it('sagt seinen Satz in einer höflichen Live-Region', () => {
  render(<Fortschritt satz="3 geschrieben von 5" erledigt={3} gesamt={5} />)

  const bereich = screen.getByRole('status')
  expect(bereich).toHaveTextContent('3 geschrieben von 5')

  // **Höflich und nicht dringend**: ein Stapellauf meldet sich alle paar
  // Sekunden, und ein `assertive` unterbräche jedes Mal, was gerade
  // vorgelesen wird.
  expect(bereich).toHaveAttribute('aria-live', 'polite')
})

// **Der Balken trägt dieselbe Aussage wie der Satz darüber** — für die, die
// ihn sehen statt ihn zu hören. Er ist deshalb ein `progressbar` mit Zahlen
// und kein `role="presentation"`: ein Vorleseprogramm konnte bis heute die
// Stelle im Lauf nicht abfragen.
it('gibt den Stand als Balken mit Zahlen und nennt ihn beim Satz', () => {
  render(<Fortschritt satz="3 geschrieben von 5" erledigt={3} gesamt={5} />)

  const balken = screen.getByRole('progressbar', {
    name: '3 geschrieben von 5',
  })

  expect(balken).toHaveAttribute('aria-valuenow', '3')
  expect(balken).toHaveAttribute('aria-valuemin', '0')
  expect(balken).toHaveAttribute('aria-valuemax', '5')
})

// **Ein Lauf ohne ein einziges Stück ist fertig, nicht kaputt.** Er darf keine
// Division durch null hinterlassen — und `aria-valuemax` muss über dem
// Minimum liegen, sonst ist der Balken für ein Vorleseprogramm unlesbar.
it('kommt mit einem Lauf ohne Stücke zurecht', () => {
  render(<Fortschritt satz="0 geschrieben von 0" erledigt={0} gesamt={0} />)

  const balken = screen.getByRole('progressbar')
  expect(balken).toHaveAttribute('aria-valuenow', '0')
  expect(balken).toHaveAttribute('aria-valuemax', '1')
})

// **Ohne Rückruf kein Knopf** (§10): einer, der nichts anhält, wäre eine
// Zusage ohne Rumpf.
it('zeigt keinen Anhalteknopf, solange niemand einen bekommt', () => {
  render(<Fortschritt satz="fertig" erledigt={5} gesamt={5} />)

  expect(screen.queryByRole('button')).toBeNull()
})

it('hält an, wenn jemand es verlangt', async () => {
  const aufKlick = vi.fn()

  render(
    <Fortschritt
      satz="3 geschrieben von 5"
      erledigt={3}
      gesamt={5}
      anhalten={{ beschriftung: 'Anhalten', aufKlick }}
    />,
  )

  await userEvent.click(screen.getByRole('button', { name: 'Anhalten' }))

  expect(aufKlick).toHaveBeenCalledTimes(1)
})

// **Nach dem Lauf ruht der Knopf** — er steht weiter da, damit die Zeile
// nicht springt, aber es gibt nichts mehr anzuhalten.
it('hält den Knopf angehalten, wenn nichts mehr läuft', () => {
  render(
    <Fortschritt
      satz="5 Bilder geschrieben nach /tmp"
      erledigt={5}
      gesamt={5}
      anhalten={{
        beschriftung: 'Anhalten',
        angehalten: true,
        aufKlick: vi.fn(),
      }}
    />,
  )

  expect(screen.getByRole('button', { name: 'Anhalten' })).toBeDisabled()
})
