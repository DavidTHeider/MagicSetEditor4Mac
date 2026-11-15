// **Die Verschlüsselung des Bestands am Fenster** — Werkstück 14 von Phase
// 4.12 (#1351),
// nach Vermerk 0072.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Bestandschiffre } from './Bestandschiffre'
import { Sprachraum } from '../sprache/kontext'
import { alsSatz } from '../test/auszeichnung'
import { kurzform, langform, ohneAuszeichnung } from '../sprache/kurzform'
import { seiten } from '../sprache/de/seiten'

function zeige(teile: Partial<Parameters<typeof Bestandschiffre>[0]> = {}) {
  const umschalten = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Bestandschiffre
        verschluesselt={false}
        umschalten={umschalten}
        laeuft={false}
        fehler={undefined}
        {...teile}
      />
    </Sprachraum>,
  )

  return { umschalten }
}

describe('Bestandschiffre', () => {
  // **Solange der Stand nicht feststeht, steht kein Knopf da.** Ein Knopf, der
  // rät, welche Richtung er hat, schaltet irgendwann in die falsche.
  it('zeigt keinen Schalter, solange der Stand offen ist', () => {
    zeige({ verschluesselt: undefined })

    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  // **Und ein Fehler beim Lesen erscheint trotzdem** — sonst stünde eine
  // Überschrift ohne alles da, und niemand wüsste, warum.
  it('meldet einen Fehler auch ohne Stand', () => {
    zeige({ verschluesselt: undefined, fehler: 'nichts zu machen' })

    expect(screen.getByRole('alert')).toHaveTextContent('nichts zu machen')
  })

  // **Die Grenzen stehen neben dem Versprechen.** Der Vermerk verlangt es
  // ausdrücklich: eine Zusage ohne ihre Grenze ist eine Falle.
  // **Die Grenzen im ersten Satz, die Gründe hinter dem Hilfesymbol** —
  // UX-15 (#1771): was man vor dem Einschalten wissen muss, bleibt sichtbar.
  it('nennt, was die Verschlüsselung nicht leistet', async () => {
    zeige()

    // **Gegen den Katalogtext selbst**, nicht gegen ein Bruchstück davon:
    // `Satz` zerlegt `**…**` in eigene Knoten, und ein Textsucher fände dann
    // nur die Hälfte.
    const grenzen = seiten.bestandspflege.chiffreGrenzen
    expect(screen.getByText(alsSatz(kurzform(grenzen)))).toBeInTheDocument()

    await userEvent.hover(
      screen.getByRole('button', { name: /Mehr zu „Die Verschlüsselung/ }),
    )
    expect(await screen.findByRole('tooltip')).toHaveTextContent(
      langform(ohneAuszeichnung(grenzen)) ?? '',
    )
  })

  it('fragt vor dem Einschalten und nennt den Schlüsselbund', async () => {
    const { umschalten } = zeige()

    await userEvent.click(
      screen.getByRole('button', { name: /Verschlüsselung einschalten/ }),
    )

    // **Die Rückfrage sagt, was auf dem Spiel steht** — nicht „wirklich?".
    expect(screen.getByText(/Schlüsselbund dieses/)).toBeInTheDocument()
    expect(umschalten).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('button', { name: 'Verschlüsselung einschalten …' }),
    )

    expect(umschalten).toHaveBeenCalledWith(true)
  })

  it('fragt vor dem Ausschalten und schaltet in die andere Richtung', async () => {
    const { umschalten } = zeige({ verschluesselt: true })

    await userEvent.click(
      screen.getByRole('button', { name: /Verschlüsselung ausschalten/ }),
    )

    await userEvent.click(
      screen.getByRole('button', { name: 'Verschlüsselung ausschalten …' }),
    )

    expect(umschalten).toHaveBeenCalledWith(false)
  })

  // **Ein Abbruch schaltet nichts** — und die Rückfrage geht weg.
  it('nimmt die Rückfrage zurück, ohne zu schalten', async () => {
    const { umschalten } = zeige()

    await userEvent.click(
      screen.getByRole('button', { name: /Verschlüsselung einschalten/ }),
    )

    await userEvent.click(screen.getByRole('button', { name: 'Abbrechen' }))

    expect(umschalten).not.toHaveBeenCalled()
    expect(screen.queryByText(/Schlüsselbund dieses/)).not.toBeInTheDocument()
  })

  it('sagt, dass umgeschrieben wird, und sperrt dabei', () => {
    zeige({ laeuft: true })

    expect(screen.getByRole('status')).toHaveTextContent(
      /Bestand wird umgeschrieben/,
    )
    expect(
      screen.getByRole('button', { name: /Verschlüsselung einschalten/ }),
    ).toBeDisabled()
  })

  it('meldet einen Fehler beim Schalten', () => {
    zeige({ fehler: 'der Schlüsselbund war zu' })

    expect(screen.getByRole('alert')).toHaveTextContent(
      'der Schlüsselbund war zu',
    )
  })
})
