// **Die Belegung der Ziffern gilt im ganzen Fenster** — UX-02.AK-03
// (#1674):
// *„Cmd+0 bis Cmd+9, Menüs, Palette und sichtbare Navigation verwenden eine
// gemeinsame Aktionsregistrierung."* Eine Ziffer, die die Tastatur anders
// deutet, als die Leiste sie zeigt, wäre zweierlei Registrierung.
import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Rahmen } from './Rahmen'
import { belegungAus } from './ziffernbelegung'
import { leisteneintraege, mitBelegung } from '../leistenordnung'

function zeige(wert: string, wechseln = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Rahmen aktuell="start" wechseln={wechseln} ziffern={belegungAus(wert)}>
        <p>{kataloge.de.seiten.start.was}</p>
      </Rahmen>
    </Sprachraum>,
  )

  return wechseln
}

describe('die Belegung im Rahmen', () => {
  it('führt ⌘3 dorthin, wohin die Belegung sagt', () => {
    const wechseln = zeige('3=karten')

    fireEvent.keyDown(window, { key: '3', code: 'Digit3', metaKey: true })

    expect(wechseln).toHaveBeenCalledWith('karten')
  })

  // **Die Leiste zeigt, was die Taste tut**: ohne ⌘3 steht bei *Sammeln*
  // keine Ziffer, und die übrigen bleiben.
  it('zeigt an der Leiste die Ziffern, die gelten', () => {
    zeige('3=karten')

    const leiste = screen.getByRole('navigation', {
      name: kataloge.de.allgemein.navigation.bezeichnung,
    })
    const sammeln = within(leiste).getByRole('button', {
      name: new RegExp(kataloge.de.allgemein.bereiche.sammeln),
    })
    const gestalten = within(leiste).getByRole('button', {
      name: new RegExp(kataloge.de.allgemein.bereiche.gestalten),
    })

    expect(sammeln).not.toHaveTextContent('⌘')
    expect(gestalten).toHaveTextContent('⌘2')
  })
})

describe('mitBelegung', () => {
  it('lässt die Leiste mit der Vorgabe, wie sie ist', () => {
    expect(mitBelegung(leisteneintraege, belegungAus(''))).toEqual(
      leisteneintraege,
    )
  })

  // **Die Dienstzone bleibt**: ⌘, ⌘' und ⌘? sind keine Ziffern.
  it('gibt einem Bereich die Ziffer seiner ersten Seite und der Dienstzone ihre Zeichen', () => {
    const eintraege = mitBelegung(
      leisteneintraege,
      belegungAus('9=sammlung,3='),
    )
    const kuerzel = Object.fromEntries(
      eintraege.map((eintrag) => [eintrag.ziel, eintrag.kuerzel]),
    )

    expect(kuerzel).toMatchObject({
      start: '1',
      sammlung: '9',
      einstellungen: ',',
      hilfe: '?',
    })
  })
})
