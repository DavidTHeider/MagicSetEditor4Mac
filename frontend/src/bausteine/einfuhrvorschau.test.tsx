// **Eine Einfuhr zeigt vorher, was sie zusammenführt** — DOM-03.AK-01
// (#1698).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Einfuhrvorschau, braucheVorschau } from './Einfuhrvorschau'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type {
  SammelvorschauDTO,
  VorschaupostenDTO,
} from '../antwortenlisteneinfuhr'

const texte = kataloge.de.seiten.sammlung

function posten(teil: Partial<VorschaupostenDTO> = {}): VorschaupostenDTO {
  return {
    zeile: 1,
    bezeichnung: 'Lightning Bolt · M11 #149',
    menge: 4,
    zustand: 'nearMint',
    sprache: 'en',
    wunsch: false,
    art: 'neu',
    vorher: 0,
    nachher: 4,
    ...teil,
  }
}

const vorschau: SammelvorschauDTO = {
  posten: [
    posten({ art: 'zusammen', vorher: 2, nachher: 6 }),
    posten({ zeile: 2, bezeichnung: 'Glut', menge: 1, nachher: 1 }),
    posten({
      zeile: 3,
      bezeichnung: 'Glut',
      menge: 2,
      art: 'wiederholt',
      vorher: 1,
      nachher: 3,
    }),
  ],
  befunde: ['Zeile 4: die Menge fehlt'],
}

function zeige(aufEinlesen = vi.fn(), aufAbbrechen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Einfuhrvorschau
        vorschau={vorschau}
        laeuft={false}
        aufEinlesen={aufEinlesen}
        aufAbbrechen={aufAbbrechen}
      />
    </Sprachraum>,
  )

  return { aufEinlesen, aufAbbrechen }
}

describe('Die Einfuhrvorschau', () => {
  it('zählt und nennt, was eine vorhandene Zeile trifft', () => {
    zeige()

    expect(
      screen.getByText(
        einsetzen(texte.vorschauZaehlung, {
          neu: '1',
          zusammen: '1',
          wiederholt: '1',
        }),
      ),
    ).toBeInTheDocument()

    const liste = screen.getByRole('list', { name: texte.vorschauTitel })
    expect(liste).toHaveTextContent(
      einsetzen(texte.vorschauZusammen, {
        bezeichnung: 'Lightning Bolt · M11 #149',
        zustand: 'Near Mint',
        vorher: '2',
        nachher: '6',
      }),
    )
    // **Die neue Zeile steht nicht in der Liste** — sie trifft nichts.
    expect(liste.querySelectorAll('li')).toHaveLength(2)
    expect(screen.getByText('Zeile 4: die Menge fehlt')).toBeInTheDocument()
  })

  // **Addieren ist vorgewählt** — das Verhalten von bisher.
  it('liest mit Addieren ein, wenn niemand etwas anderes wählt', async () => {
    const { aufEinlesen } = zeige()

    expect(
      screen.getByRole('radio', { name: texte.vorschauAddieren }),
    ).toBeChecked()

    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauEinlesen }),
    )

    expect(aufEinlesen).toHaveBeenCalledWith('addieren')
  })

  it('liest mit Auslassen ein, wenn es gewählt ist', async () => {
    const { aufEinlesen } = zeige()

    await userEvent.click(
      screen.getByRole('radio', { name: texte.vorschauAuslassen }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauEinlesen }),
    )

    expect(aufEinlesen).toHaveBeenCalledWith('auslassen')
  })

  // **Wer umentscheidet, bekommt, was er zuletzt gewählt hat.**
  it('liest mit Addieren ein, wenn nach Auslassen zurückgewechselt wird', async () => {
    const { aufEinlesen } = zeige()

    await userEvent.click(
      screen.getByRole('radio', { name: texte.vorschauAuslassen }),
    )
    await userEvent.click(
      screen.getByRole('radio', { name: texte.vorschauAddieren }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauEinlesen }),
    )

    expect(aufEinlesen).toHaveBeenCalledWith('addieren')
  })

  it('lässt abbrechen, ohne einzulesen', async () => {
    const { aufEinlesen, aufAbbrechen } = zeige()

    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauAbbrechen }),
    )

    expect(aufAbbrechen).toHaveBeenCalled()
    expect(aufEinlesen).not.toHaveBeenCalled()
  })

  // **Ein Zustand, den der Katalog nicht kennt, steht, wie er kam.**
  it('nennt einen unbekannten Zustand roh', () => {
    render(
      <Sprachraum sprache="de">
        <Einfuhrvorschau
          vorschau={{
            posten: [posten({ art: 'zusammen', zustand: 'erfunden' })],
            befunde: [],
          }}
          laeuft={false}
          aufEinlesen={vi.fn()}
          aufAbbrechen={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(
      screen.getByRole('list', { name: texte.vorschauTitel }),
    ).toHaveTextContent('(erfunden)')
  })

  // **Eine Vorschau braucht nur, was etwas trifft** — eine Einfuhr aus
  // lauter neuen Zeilen geht ohne Zwischenschritt durch.
  it('wird nur gebraucht, wenn etwas zusammengeführt würde', () => {
    expect(braucheVorschau(vorschau)).toBe(true)
    expect(braucheVorschau(null)).toBe(false)
    expect(braucheVorschau({ posten: [posten()], befunde: [] })).toBe(false)
    expect(
      braucheVorschau({ posten: [posten({ art: 'wiederholt' })], befunde: [] }),
    ).toBe(true)
  })
})
