// **Die Budgetzahlen einer Kontextfreigabe sind ganze Zahlen** — F03
// (#2245).
//
// Bis dahin las `Number(...)` beide Felder: ein leeres wurde null, und
// „8192.5" ging als Kommazahl an ein `int` in Go — die Absage kam als Satz
// über JSON statt am Feld. **Was ein zulässiges Budget ist, sagt weiter Go**
// (`service.pruefeKontextwunsch`); hier geht es nur darum, dass eine Zahl
// ankommt, die jemand getippt hat.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Kontextfreigabe } from './Kontextfreigabe'

const texte = kataloge.de.seiten.recherche

afterEach(() => {
  raeumeApi()
})

function zeige() {
  const gerufen = vi.fn()
  belegeApi({ KontextZusammenstellen: gerufen })

  render(
    <Sprachraum sprache="de">
      <Kontextfreigabe freigegeben={[]} fall="regeln" frage="trampelschaden" />
    </Sprachraum>,
  )

  return gerufen
}

it.each([
  ['Kontextfenster', '8192.5', texte.freigabeFenster],
  ['Ausgabereserve', '-1', texte.freigabeReserve],
  ['Kontextfenster', '', texte.freigabeFenster],
])(
  'hält das Zusammenstellen an, wenn im %s %j steht',
  async (_name, roh, etikett) => {
    const gerufen = zeige()
    const nutzer = userEvent.setup()

    await nutzer.type(
      screen.getByLabelText(texte.freigabeAufgabe),
      'Erkläre Trampelschaden',
    )

    const feld = screen.getByLabelText(etikett)
    await nutzer.clear(feld)
    if (roh !== '') {
      await nutzer.type(feld, roh)

      expect(feld).toHaveAttribute('aria-invalid', 'true')
      expect(feld).toHaveAccessibleDescription(
        einsetzen(kataloge.de.allgemein.ganzeZahlAb, { kleinstens: '0' }),
      )
    }

    expect(
      screen.getByRole('button', { name: texte.freigabeZeigen }),
    ).toBeDisabled()
    expect(gerufen).not.toHaveBeenCalled()
  },
)
