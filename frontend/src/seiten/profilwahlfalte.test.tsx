// **Die Profilwahl faltet ab vielen Profilen** — #2022, Folge aus UXR-20
// (#1928).
//
// Mit einem Haushalt (drei Menschen, sieben Profile) stand in der Spalte neben
// den Laschen je Profil ein „Hierhin wechseln", und die Lasche *Konto* trug
// bei 1280 × 800 25 Bedienelemente im Bild (R2: höchstens zwanzig). Seitdem
// steht das gezeigte Profil vorn, dazu drei weitere, der Rest auf Wunsch.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { vielBestand } from '../test/nutzerseite'
import { Profilwahl } from './Profilwahl'

const eigene = kataloge.de.seiten.nutzer

function zeige(gezeigt: string) {
  render(
    <Sprachraum sprache="de">
      <Profilwahl nutzer={vielBestand} gezeigt={gezeigt} />
    </Sprachraum>,
  )

  return screen.getByRole('region', { name: eigene.wahl })
}

const zeilen = (bereich: HTMLElement) =>
  within(bereich).getAllByRole('listitem')

describe('Die Profilwahl mit sieben Profilen', () => {
  it('zeigt vier und den Rest auf Wunsch', async () => {
    const bereich = zeige('10')

    expect(zeilen(bereich)).toHaveLength(4)

    await userEvent.click(
      within(bereich).getByRole('button', {
        name: /Weitere 3/,
      }),
    )

    expect(zeilen(bereich)).toHaveLength(7)
  })

  it('stellt das gezeigte Profil vorn hin, auch wenn es spät käme', () => {
    const bereich = zeige('31')

    const erste = zeilen(bereich)[0]
    expect(erste).toHaveTextContent('Draft')
    expect(erste).toHaveTextContent(eigene.hier)
    expect(zeilen(bereich)).toHaveLength(4)
  })
})
