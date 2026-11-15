// **Der Symboleditor im größeren Zustand** — AP-12
// (#2369),
// Befund B-26: „Die Schublade „Symbol“ darf breiter aufgehen, oder sie
// bekommt einen Knopf „Größer“.“ Die Zeichenfläche maß 320 Punkte in einer
// Schublade von 460.
//
// **adapted** aus MSE2: dort ist der Symboleditor ein eigenes Fenster,
// 650 × 600 groß und frei zu ziehen, und die Zeichenfläche wächst mit
// (`src/gui/symbol/window.cpp` 68, Fassung f9d9356 vom 2021-07-06); gemerkt
// wird die Größe nicht. Hier ist es ein Umschalter in der Schublade.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { SchubladenweiteKontext } from '../haken/useSchubladenweite'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

function zeige(weite?: { weit: boolean; umschalten: () => void }) {
  belegeApi({
    Setangaben: vi
      .fn()
      .mockResolvedValue({ felder: [], werte: [], fehler: null }),
  })

  const seite = <Symbol setID="7" />

  render(
    <Sprachraum sprache="de">
      {weite === undefined ? (
        seite
      ) : (
        <SchubladenweiteKontext value={weite}>{seite}</SchubladenweiteKontext>
      )}
    </Sprachraum>,
  )
}

describe('der Symboleditor, größer', () => {
  it('bietet in der Schublade „Größer“ als Umschalter an', async () => {
    const umschalten = vi.fn()
    zeige({ weit: false, umschalten })

    const knopf = screen.getByRole('button', { name: texte.groesser })
    expect(knopf).toHaveAttribute('aria-pressed', 'false')
    expect(screen.getByRole('group', { name: texte.zeichnung })).toHaveClass(
      'w-80',
    )

    await userEvent.click(knopf)

    expect(umschalten).toHaveBeenCalledOnce()
  })

  // **Der Name bleibt, der Zustand wechselt** — ein Umschalter, der seinen
  // Namen tauscht, sagte zweimal das Gegenteil.
  it('zeigt im größeren Zustand eine größere Zeichenfläche', () => {
    zeige({ weit: true, umschalten: vi.fn() })

    expect(
      screen.getByRole('button', { name: texte.groesser }),
    ).toHaveAttribute('aria-pressed', 'true')
    const flaeche = screen.getByRole('group', { name: texte.zeichnung })
    expect(flaeche).not.toHaveClass('w-80')
    expect(flaeche).toHaveClass('w-[32rem]')
  })

  it('hat außerhalb einer Schublade keinen Umschalter', () => {
    zeige()

    expect(screen.getByRole('group', { name: texte.zeichnung })).toHaveClass(
      'w-80',
    )
    expect(screen.queryByRole('button', { name: texte.groesser })).toBeNull()
  })

  // **Grundformen mit Zeichen und Namen** — B-26: die Werkzeuge standen als
  // leise Textknöpfe da. Das Zeichen ist Schmuck; der Name trägt den Knopf.
  it('zeigt die Grundformen mit ihrem Zeichen, benannt nach dem Namen', () => {
    zeige()

    const kreis = screen.getByRole('button', { name: texte.kreis })
    expect(kreis).toHaveTextContent('○')
    expect(kreis.querySelector('[aria-hidden="true"]')).toHaveTextContent('○')
  })
})
