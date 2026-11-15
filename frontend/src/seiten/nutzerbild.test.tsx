// **Das Profilbild im Fenster** — Werkstück 8 von Phase 4.12
// (#1351).
//
// **Die Zusage lautet: ein Bild ist als Bild zu sehen.** Ein Dateiname im
// Bestand ist keines — geprüft wird deshalb am `<img>` und an seinem `alt`,
// nicht an einer Zeichenkette im Zustand.
import { render, screen } from '@testing-library/react'
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

// einBild ist ein winziges, gültiges PNG als Daten-URI.
const einBild =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=='

const einer: NutzerDTO = {
  id: '1',
  name: 'Ich',
  bild: '',
  email: '',
  telefon: '',
  anonym: false,
  anrede: '',
  namenszusatz: '',
  vorname: '',
  weiterenamen: '',
  nachname: '',
  slogan: '',
  kurzebio: '',
  langebio: '',
  fassung: '1',
  profile: [
    probeprofil({ id: '10', nutzerID: '1', name: 'Mein Profil', fassung: '1' }),
  ],
}

const gezeigtesProfil = () =>
  Promise.resolve({ profil: einer.profile[0] ?? null })

function belege(
  teile: Parameters<typeof belegeApi>[0] = {},
  nutzer: NutzerDTO = einer,
) {
  belegeApi({
    ...belegeAnmeldung(),
    GezeigtesProfil: vi.fn(gezeigtesProfil),
    Nutzer: vi.fn(() => Promise.resolve({ nutzer: [nutzer] })),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Nutzerseite />
    </Sprachraum>,
  )
}

// zeileOeffnen klappt die Zeile von „Ich" unter *Nutzer und Profile
// bearbeiten* auf — **und seit UXR-20 (#1928) zuerst den Kasten**: das
// Bearbeiten steht eingeklappt unter den eigenen Angaben.
async function zeileOeffnen() {
  await userEvent.click(
    await screen.findByRole('button', {
      name: 'Nutzer und Profile bearbeiten',
    }),
  )
  await userEvent.click(
    await screen.findByRole('button', { name: 'Ich bearbeiten' }),
  )
}

describe('Das Profilbild', () => {
  it('steht als Bild in der Liste', async () => {
    belege({
      Nutzerbilder: vi.fn(() =>
        Promise.resolve({ bilder: { '1': einBild }, badges: {} }),
      ),
    })

    const bild = await screen.findByRole('img', { name: 'Profilbild von Ich' })

    expect(bild).toHaveAttribute('src', einBild)
  })

  // **Wer keines hat, bekommt seinen Anfangsbuchstaben.**
  //
  // Eine Liste von leeren Kreisen unterscheidet keine Zeile von der anderen.
  it('wird durch den Anfangsbuchstaben ersetzt, wenn keines da ist', async () => {
    belege()

    const ersatz = await screen.findByLabelText('Ich hat kein Profilbild')

    expect(ersatz).toHaveTextContent('I')
    expect(screen.queryByRole('img')).toBeNull()
  })

  it('lässt sich wählen — mit Kennung und Fassung', async () => {
    const waehlen = vi.fn(() =>
      Promise.resolve({ nutzer: { ...einer, bild: 'nutzer-1.png' } }),
    )

    // **Der zweite Ruf gibt das Bild** — so wie der Bestand es nach dem
    // Ablegen täte.
    let gewaehlt = false

    belege({
      NutzerbildWaehlen: vi.fn(() => {
        gewaehlt = true

        return waehlen()
      }),
      Nutzerbilder: vi.fn(() =>
        Promise.resolve({
          bilder: gewaehlt ? { '1': einBild } : {},
          badges: {},
        }),
      ),
    })

    await zeileOeffnen()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Bild wählen …' }),
    )

    // **Gewartet wird auf die Wirkung und nicht auf den Ruf** (§3): das Bild
    // steht erst da, wenn die Antwort angekommen **und** die Liste neu
    // gelesen ist.
    await screen.findAllByRole('img', { name: 'Profilbild von Ich' })

    expect(waehlen).toHaveBeenCalled()
  })

  // **Der Knopf zum Wegnehmen steht nur da, wenn es etwas wegzunehmen gibt.**
  it('lässt sich entfernen, sobald eines da ist', async () => {
    let entfernt = false

    const entfernen = vi.fn(() => {
      entfernt = true

      return Promise.resolve({ nutzer: { ...einer, bild: '' } })
    })

    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() =>
        Promise.resolve({
          nutzer: [{ ...einer, bild: entfernt ? '' : 'nutzer-1.png' }],
        }),
      ),
      Nutzerbilder: vi.fn(() =>
        Promise.resolve({
          bilder: entfernt ? {} : { '1': einBild },
          badges: {},
        }),
      ),
      NutzerbildEntfernen: entfernen,
    })

    render(
      <Sprachraum sprache="de">
        <Nutzerseite />
      </Sprachraum>,
    )

    await zeileOeffnen()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Bild entfernen' }),
    )

    // **Die Wirkung ist der Ersatz an der Stelle des Bildes** — nicht der
    // Ruf, und nicht das Verschwinden des Knopfes allein.
    await screen.findAllByLabelText('Ich hat kein Profilbild')

    expect(entfernen).toHaveBeenCalledWith('1', '1')
  })

  // **Ohne Bild gibt es nichts wegzunehmen** — und keinen Knopf dafür.
  //
  // **Der Beleg, dass jemand zuhört, steht daneben**: der Knopf zum Wählen
  // ist da. Ohne ihn prüfte diese Verneinung nur, dass die Zeile gar nicht
  // aufgeklappt ist.
  it('zeigt keinen Knopf zum Wegnehmen, wenn keines da ist', async () => {
    belege()

    await zeileOeffnen()

    await screen.findByRole('button', { name: 'Bild wählen …' })

    expect(screen.queryByRole('button', { name: 'Bild entfernen' })).toBeNull()
  })

  it('meldet, was beim Wählen schiefgeht', async () => {
    belege({
      NutzerbildWaehlen: vi.fn(() =>
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

    await zeileOeffnen()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Bild wählen …' }),
    )

    // **Der Schlüssel gewinnt über den Satz aus Go** — die Meldung kommt aus
    // dem Katalog, und genau das ist die Zusage der Fehlerübersetzung.
    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )
  })
})
