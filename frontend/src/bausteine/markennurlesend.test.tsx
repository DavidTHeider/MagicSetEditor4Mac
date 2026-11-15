// **Marken für einen Leser** — G10
// (#2276):
// Marken schreiben ins Profil, und ein Leser sieht sie, setzt und entfernt
// aber keine. Feld und Knöpfe sagen es, statt erst am Tor zu scheitern.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Markenzeile } from './Markenzeile'
import { Objektmarken } from './Objektmarken'

const texte = kataloge.de.bausteine.objektmarken
const satz = kataloge.de.bausteine.nurlesen.knopf

afterEach(raeumeApi)

describe('Marken für einen Leser', () => {
  it('lässt das Feld lesen und sperrt Setzen und Entfernen', async () => {
    const setzen = vi.fn()
    const entfernen = vi.fn()
    belegeApi({ ObjektmarkeSetzen: setzen, ObjektmarkeEntfernen: entfernen })

    render(
      <Sprachraum sprache="de">
        <Schreibrechtraum nurLesend>
          <Objektmarken
            art="exemplar"
            kennung="31"
            marken={['tausch']}
            geaendert={vi.fn()}
          />
        </Schreibrechtraum>
      </Sprachraum>,
    )

    const feld = screen.getByRole('textbox', { name: texte.setzenFuer })
    await userEvent.type(feld, 'Binder 2{Enter}')
    expect(feld).toHaveValue('')
    // **Das Feld sagt, warum** (H01, #2255) — nicht nur seine Knöpfe.
    expect(feld).toHaveAccessibleDescription(satz)

    const absenden = screen.getByRole('button', { name: texte.setzen })
    expect(absenden).toHaveAttribute('aria-disabled', 'true')
    expect(absenden).toHaveAccessibleDescription(satz)

    const weg = screen.getByRole('button', { name: '»tausch« entfernen' })
    expect(weg).toHaveAttribute('aria-disabled', 'true')
    expect(weg).toHaveAccessibleDescription(satz)
    expect(weg).toHaveAttribute('title', satz)
    await userEvent.click(weg)

    // **Und sie sehen gesperrt aus** wie jeder `Knopf`, der schreibt.
    for (const knopf of [absenden, weg]) {
      expect(knopf).toHaveClass('aria-disabled:opacity-50')
    }

    expect(setzen).not.toHaveBeenCalled()
    expect(entfernen).not.toHaveBeenCalled()
  })

  it('öffnet einem Leser die Zeile nicht und einem Mitarbeiter schon', async () => {
    const zeile = (nurLesend: boolean) => (
      <Sprachraum sprache="de">
        <Schreibrechtraum nurLesend={nurLesend}>
          <Markenzeile
            art="deck"
            kennung="7"
            name="Mono-Grün"
            marken={['turnier']}
            nachladen={vi.fn()}
          />
        </Schreibrechtraum>
      </Sprachraum>
    )
    const { rerender } = render(zeile(true))

    const schalter = screen.getByRole('button', {
      name: texte.zeileBearbeiten.replace('{name}', 'Mono-Grün'),
    })
    expect(schalter).toHaveAttribute('aria-disabled', 'true')
    expect(schalter).toHaveAccessibleDescription(satz)
    expect(schalter).toHaveClass('aria-disabled:opacity-50')
    await userEvent.click(schalter)
    expect(schalter).toHaveAttribute('aria-expanded', 'false')

    // **Derselbe Klick öffnet sie, wo geschrieben werden darf** — der Beleg,
    // dass der Klick ankommt.
    rerender(zeile(false))
    expect(schalter).not.toHaveAttribute('aria-disabled')
    await userEvent.click(schalter)
    expect(
      screen.getByRole('textbox', { name: texte.setzenFuer }),
    ).toBeVisible()
  })
})
