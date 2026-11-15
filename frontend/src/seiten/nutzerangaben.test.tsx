// **Die Angaben zur Person im Fenster** — Werkstück 8 von Phase 4.12
// (#1351),
// Punkte 1 bis 5, 14 und 15.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nutzerseite } from './Nutzerseite'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { belegeAnmeldung } from '../test/anmeldebelegung'

import type { NutzerDTO } from '../api'
import { probeprofil } from '../test/profil'

afterEach(() => {
  raeumeApi()
})

const einer: NutzerDTO = {
  id: '1',
  name: 'Ich',
  bild: '',
  email: '',
  telefon: '',
  anonym: false,
  anrede: '',
  namenszusatz: '',
  vorname: 'Alex',
  weiterenamen: '',
  nachname: '',
  slogan: '',
  kurzebio: '',
  langebio: '',
  fassung: '3',
  profile: [
    probeprofil({ id: '10', nutzerID: '1', name: 'Mein Profil', fassung: '1' }),
  ],
}

const eigene = kataloge.de.seiten.nutzer

// imFormular grenzt die Suche auf den Abschnitt ein.
//
// **Nötig, weil auf dieser Seite mehrere Formulare stehen.** Eine Suche über
// die ganze Seite fände Felder aus allen, und ein Test, der das erste nähme,
// prüfte die falsche Stelle.
async function imFormular() {
  return within(await screen.findByRole('region', { name: eigene.angaben }))
}

function zeige(teile: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({
    ...belegeAnmeldung(),
    GezeigtesProfil: vi.fn(() =>
      Promise.resolve({ profil: einer.profile[0] ?? null }),
    ),
    Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Nutzerseite />
    </Sprachraum>,
  )
}

describe('Die Angaben zur Person', () => {
  it('zeigt jedes Feld mit seinem Wert', async () => {
    zeige()

    const formular = await imFormular()

    expect(formular.getByLabelText(eigene.felder.vorname)).toHaveValue('Alex')

    // **Alle neun** — acht freiwillige und der Anzeigename. Die Liste kommt
    // aus dem Katalog, damit ein weggefallenes Feld auffällt.
    for (const beschriftung of Object.values(eigene.felder)) {
      expect(formular.getByLabelText(beschriftung)).toBeInTheDocument()
    }
  })

  // **Alle Felder gehen zusammen hinaus — samt der Fassung.**
  //
  // **Beides ist die Zusage**: `NutzerAendern` setzt und verschmilzt nicht,
  // und ohne die Fassung würde abgewiesen, statt eine gleichzeitige Änderung
  // an einer anderen Fläche zu überschreiben.
  it('schickt jedes Feld und die Fassung mit', async () => {
    const aendern = vi.fn(() => Promise.resolve({ nutzer: einer }))

    zeige({ NutzerAendern: aendern })

    const formular = await imFormular()

    await userEvent.type(
      formular.getByLabelText(eigene.felder.slogan),
      'Immer im Dienst',
    )

    await userEvent.click(
      formular.getByRole('button', { name: eigene.angabenSichern }),
    )

    await screen.findByDisplayValue('Immer im Dienst')

    expect(aendern).toHaveBeenCalledWith(
      expect.objectContaining({
        id: '1',
        name: 'Ich',
        vorname: 'Alex',
        slogan: 'Immer im Dienst',
        fassung: '3',
      }),
    )
  })

  // **Ohne Anzeigenamen ruht der Knopf** — und das Feld sagt, warum.
  it('hält den Knopf an, solange der Anzeigename fehlt', async () => {
    zeige()

    const formular = await imFormular()

    await userEvent.clear(formular.getByLabelText(eigene.felder.name))

    expect(
      formular.getByRole('button', { name: eigene.angabenSichern }),
    ).toBeDisabled()
  })

  it('meldet, was beim Sichern schiefgeht', async () => {
    zeige({
      NutzerAendern: vi.fn(() =>
        Promise.resolve({
          nutzer: null,
          fehler: {
            code: 'konflikt' as const,
            schluessel: 'konflikt',
            meldung: 'Inzwischen geändert.',
          },
        }),
      ),
    })

    const formular = await imFormular()

    await userEvent.click(
      formular.getByRole('button', { name: eigene.angabenSichern }),
    )

    expect(await formular.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )
  })
})
