// **Der Hinweis steht unter seinem Feld** — UXR-17
// (#1925),
// Befund 4 der Hüllenprobe `fmt02d`: im Formular *Neues Paket anlegen* brach
// der Hinweis zum Spiel bei Stil und Exportvorlage **in eine eigene Zeile
// links unter dem Formular** um, getrennt von seinem Feld. Hinweis und Feld
// standen als Geschwister in einer umbrechenden Zeile. Jetzt stehen sie
// zusammen, wie beim Kurznamen (`Feld`).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { Paketanlage } from './Paketanlage'

afterEach(() => {
  raeumeApi()
})

const t = kataloge.de.seiten.pakete

describe('Die Hinweise der Paketanlage', () => {
  it('stehen mit ihrem Feld zusammen und beschreiben es', async () => {
    belegeApi({
      PaketArten: () =>
        Promise.resolve([
          { art: 'mse-game', brauchtSpiel: false },
          { art: 'mse-style', brauchtSpiel: true },
        ]),
    })
    render(
      <Sprachraum sprache="de">
        <Paketanlage spiele={[]} angelegt={() => undefined} />
      </Sprachraum>,
    )
    await effekteNachholen()
    await screen.findByRole('option', { name: t.anlegenArtStil })

    await userEvent.selectOptions(
      screen.getByLabelText(t.anlegenArt),
      'mse-style',
    )

    for (const [feld, hinweis] of [
      [t.anlegenVollname, t.anlegenVollnameHinweis],
      [t.anlegenSpiel, t.anlegenSpielHinweis],
    ] as const) {
      const eingabe = screen.getByLabelText(feld)

      expect(eingabe).toHaveAccessibleDescription(hinweis)

      // **Beschriftung, Feld und Hinweis stehen in einem Kasten** — dem von
      // `Feld`. Das Feld der Spielwahl ist seit UXR-F1 (#2007) eine Auswahl
      // mit Suche, die ihr Eingabefeld in einen eigenen Behälter legt; der
      // Kasten ist deshalb der der Beschriftung und nicht der Elternknoten
      // des Feldes.
      const kasten = screen.getByText(feld, { selector: 'label' }).parentElement

      expect(kasten).toContainElement(eingabe)
      expect(kasten).toContainElement(screen.getByText(hinweis))
    }
  })
})
