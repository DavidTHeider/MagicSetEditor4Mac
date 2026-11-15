// **Ein Entwurf im Schlüsselworteditor überlebt fremde Änderungen** — U06
// (#2122).
//
// Der Editor trug den Stand der angekommenen Antwort im `key`
// (`geladenBei`), und jede Änderung irgendwo baute ihn neu auf. Dieselbe
// Klasse wie U01; der Vertrag steht in `useEntwurf`.
import { act, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import { texte, waehle, wort, zeige } from '../test/schluesselwoerter'

const entwurf = kataloge.de.bausteine.entwurf

afterEach(() => {
  raeumeApi()
})

// zwei sind ein Wort, das ein anderes Programm ändert, und eines daneben.
// **Dieselben Gegenstände wie im Bestand der Probe** — er kopiert die Liste,
// nicht die Wörter.
function zwei() {
  const beben = wort({
    id: '7',
    wort: 'Beben',
    muster: 'Beben',
    erinnerung: 'alt',
    modus: 'custom',
  })

  return {
    beben,
    woerter: [
      beben,
      wort({ id: '8', wort: 'Sog', muster: 'Sog', modus: 'custom' }),
    ],
  }
}

describe('ein Entwurf im Schlüsselworteditor', () => {
  it('bleibt mit Fokus stehen, wenn ein anderes Feld fremd geändert wird', async () => {
    const { beben, woerter } = zwei()
    const probe = zeige(woerter)

    await waehle('Beben')
    const muster = await screen.findByLabelText(texte.muster)
    await userEvent.type(muster, ' groß')

    beben.erinnerung = 'neu'
    probe.zeichneNeu(1)

    await waitFor(() => {
      expect(screen.getByLabelText(texte.erinnerung)).toHaveValue('neu')
    })
    expect(screen.getByLabelText(texte.muster)).toBe(muster)
    expect(muster).toHaveValue('Beben groß')
    expect(document.activeElement).toBe(muster)
  })

  it('sagt, wenn dasselbe Feld fremd geändert wurde, und schreibt nicht still darüber', async () => {
    const { beben, woerter } = zwei()
    const probe = zeige(woerter)

    await waehle('Beben')
    const muster = await screen.findByLabelText(texte.muster)
    await userEvent.type(muster, ' groß')

    beben.muster = 'Erdbeben'
    probe.zeichneNeu(1)

    expect(
      await screen.findByText(
        einsetzen(entwurf.fremdGeaendert, { wert: 'Erdbeben' }),
      ),
    ).toBeInTheDocument()

    await act(async () => {
      muster.blur()
      await Promise.resolve()
    })
    expect(probe.SchluesselwortSetzen).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('button', { name: entwurf.behalten }),
    )

    await waitFor(() => {
      expect(probe.SchluesselwortSetzen).toHaveBeenCalledWith(
        '3',
        '7',
        'muster',
        'Beben groß',
        expect.any(Array),
      )
    })
  })

  it('zeigt nach einer Rücknahme den zurückgeholten Modus', async () => {
    const { beben, woerter } = zwei()
    const probe = zeige(woerter)

    await waehle('Beben')
    expect(await screen.findByLabelText(texte.modus)).toHaveValue('custom')

    beben.modus = 'core'
    probe.zeichneNeu(1)

    await waitFor(() => {
      expect(screen.getByLabelText(texte.modus)).toHaveValue('core')
    })
  })

  // **Ein Einschub aus dem Menü ist kein fremder Wert** — er schreibt an
  // der Eingabe vorbei, und ohne `uebernehmen` erschiene er nach dem
  // Nachlesen als Konflikt.
  it('meldet nach einem Einschub aus dem Menü keinen Konflikt und baut das Feld nicht neu', async () => {
    const { woerter } = zwei()
    const probe = zeige(woerter, {
      belegen: {
        Parameterhilfen: () =>
          Promise.resolve({
            arten: [
              {
                name: 'zahl',
                beschreibung: '',
                einfuegen: '<atom-param>zahl</atom-param>',
              },
            ],
            verweise: [],
            befunde: [],
            fehler: null,
          }),
      },
    })

    await waehle('Beben')
    const muster = await screen.findByLabelText(texte.muster)
    await userEvent.type(muster, ' groß')
    await userEvent.tab()

    // **Wie ein Mensch: erst, wenn das Menü wieder geht** — während eines
    // Schreibens ist der Editor angehalten (§7, Regel 4).
    const menue = await screen.findByRole('combobox', {
      name: texte.parameterEinfuegen,
    })
    await waitFor(() => {
      expect(menue).toBeEnabled()
    })
    await userEvent.selectOptions(menue, 'zahl')

    await waitFor(() => {
      expect(probe.SchluesselwortSetzen).toHaveBeenLastCalledWith(
        '3',
        '7',
        'muster',
        'Beben groß<atom-param>zahl</atom-param>',
        expect.any(Array),
      )
    })
    probe.zeichneNeu(1)

    await waitFor(() => {
      expect(screen.getByLabelText(texte.muster)).toHaveValue(
        'Beben groß<atom-param>zahl</atom-param>',
      )
    })
    expect(
      screen.queryByRole('button', { name: entwurf.behalten }),
    ).not.toBeInTheDocument()
    expect(screen.getByLabelText(texte.muster)).toBe(muster)
  })
})
