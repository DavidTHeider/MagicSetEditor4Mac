// **Ein Kasten für die KI unter dem Kartenformular** — UXR-21
// (#1929).
//
// Das Artboard *Karten* zeigt im Inspector Felder und sonst nichts; darunter
// standen sechs KI-Teile offen untereinander — Vorschlag, Bild, Einschätzung,
// mehrere Karten, Herkunft, Ähnliche —, gemessen am Fenster neun
// Bedienelemente nach dem Formular, und der Inspector rollte eng 2,8. Was
// am Feld hilft, steht am Feld (AI-05); der Rest steht hinter **einem**
// Aufklapper.
//
// **Offen bleibt offen**: der Editor wird je Karte neu aufgebaut, und wer mit
// der KI arbeitet, schlüge den Kasten sonst an jeder Karte neu auf.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it } from 'vitest'

import { KIKasten, KIKastenraum } from './KIKasten'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

const titel = kataloge.de.bausteine.kikasten.titel

function Karte({ nummer }: { nummer: number }) {
  return (
    <KIKasten key={nummer}>
      <button type="button">Teil von Karte {nummer}</button>
    </KIKasten>
  )
}

it('steht zugeklappt da und nennt sich', () => {
  render(
    <Sprachraum sprache="de">
      <KIKasten>
        <button type="button">ein Teil</button>
      </KIKasten>
    </Sprachraum>,
  )

  expect(screen.getByText(titel)).toBeVisible()
  expect(screen.getByRole('button', { name: 'ein Teil' })).not.toBeVisible()
})

it('zeigt seine Teile auf einen Klick', async () => {
  render(
    <Sprachraum sprache="de">
      <KIKasten>
        <button type="button">ein Teil</button>
      </KIKasten>
    </Sprachraum>,
  )

  await userEvent.click(screen.getByText(titel))

  expect(screen.getByRole('button', { name: 'ein Teil' })).toBeVisible()
})

it('bleibt über einen Kartenwechsel offen', async () => {
  const { rerender } = render(
    <Sprachraum sprache="de">
      <KIKastenraum>
        <Karte nummer={1} />
      </KIKastenraum>
    </Sprachraum>,
  )

  await userEvent.click(screen.getByText(titel))

  rerender(
    <Sprachraum sprache="de">
      <KIKastenraum>
        <Karte nummer={2} />
      </KIKastenraum>
    </Sprachraum>,
  )

  expect(screen.getByRole('button', { name: 'Teil von Karte 2' })).toBeVisible()
})

// **Und zugeklappt bleibt zugeklappt** — die Gegenrichtung.
it('bleibt über einen Kartenwechsel zu, wenn er zu war', async () => {
  const { rerender } = render(
    <Sprachraum sprache="de">
      <KIKastenraum>
        <Karte nummer={1} />
      </KIKastenraum>
    </Sprachraum>,
  )

  await userEvent.click(screen.getByText(titel))
  await userEvent.click(screen.getByText(titel))

  rerender(
    <Sprachraum sprache="de">
      <KIKastenraum>
        <Karte nummer={2} />
      </KIKastenraum>
    </Sprachraum>,
  )

  expect(
    screen.getByRole('button', { name: 'Teil von Karte 2' }),
  ).not.toBeVisible()
})
