// **Die Ausgabeauswahl sagt, dass sie lädt, und was scheiterte** — G11
// (#2277).
//
// B03 verzögerte `Kartenzeilen` auf der Seite Bilder um vier Sekunden: die
// Auswahl stand da wie fertig und nannte „Das ganze Set (0 Karten)" — als
// hätte das Set keine. Ein Scheitern stand ohne Rückweg da.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Ausgabeauswahl } from './Ausgabeauswahl'

const texte = kataloge.de.bausteine.ausgabeauswahl
const zustaende = kataloge.de.bausteine.zustaende

const fehlschlag = {
  code: 'unbekannt' as const,
  schluessel: 'unbekannt',
  meldung: 'roh',
  werte: {},
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Ausgabeauswahl
        setID="1"
        gewaehlt={[]}
        setzeGewaehlt={() => undefined}
        angehalten={false}
      />
    </Sprachraum>,
  )
}

afterEach(raeumeApi)

describe('Die Ausgabeauswahl beim Laden', () => {
  it('sagt, dass sie lädt, und nennt solange keine Kartenzahl', async () => {
    belegeApi({ Kartenzeilen: () => new Promise(() => undefined) })

    zeige()

    expect(await screen.findByRole('status')).toHaveTextContent(zustaende.laedt)
    expect(
      screen.getByRole('radio', { name: texte.ganzesSetOhneZahl }),
    ).toBeChecked()
    expect(screen.queryByText(/0 Karten/)).toBeNull()
  })

  it('meldet ein Scheitern und lädt erneut', async () => {
    let gefragt = 0
    belegeApi({
      Kartenzeilen: () => {
        gefragt++

        return Promise.resolve(
          gefragt === 1
            ? { zeilen: [], fehler: fehlschlag }
            : {
                zeilen: [{ id: '1', nummer: 0, name: 'Erste', felder: 1 }],
                fehler: null,
              },
        )
      },
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )

    await userEvent.click(
      screen.getByRole('button', { name: zustaende.wiederholen }),
    )

    expect(
      await screen.findByRole('radio', { name: /Das ganze Set \(1 Karten\)/ }),
    ).toBeChecked()
    expect(screen.queryByRole('alert')).toBeNull()
  })
})
