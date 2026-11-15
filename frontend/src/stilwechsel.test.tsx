// **Ein Stilwechsel ist sofort im Kopf des Arbeitsplatzes zu sehen** (#288).
//
// # Warum das durch die ganze Anwendung geprüft wird
//
// Seit die Stilwahl nicht mehr über den Karten steht, liegt zwischen ihr und
// allem anderen der Zustand von `App`. Ein Rückruf, der dort verpufft, sähe im
// Einzeltest der Stilwahl richtig aus: sie ruft ihn ja. Erst hier zeigt sich,
// ob jemand zuhört.
//
// **Der Kopf ist der Abnehmer.** Er steht neben jeder Lasche und nennt Spiel,
// Kartenzahl und Stilpaket; wer den Stil wechselt, sieht es dort, ohne die
// Lasche zu verlassen.
//
// **`useBestandsstand` genügt nicht.** Er fragt in Abständen nach; wer den
// Stil wechselt, weiß es sofort und soll nicht auf den nächsten Abruf warten.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { ausKarten } from './test/kartenproben'
import { leisteneintrag } from './test/leistenweg'
import { waehleZeile } from './test/suchwahl'
import { verborgen } from './test/verborgen'

const texte = kataloge.de

beforeEach(() => {
  belegeApi({ ...belegeFenster('de') })
})

afterEach(() => {
  raeumeApi()
})

describe('Stilwechsel', () => {
  it('steht sofort im Kopf des Arbeitsplatzes', async () => {
    const Karten = vi.fn(() => Promise.resolve({ karten: [] }))

    // **Gewartet wird auf die Liste, die die Seite wirklich holt** (#891). Seit
    // dem 2026-08-25 ist das `Kartenzeilen`; `Karten` ruft die Kartenseite gar
    // nicht mehr, und ein `waitFor` darauf liefe in seine Frist.
    const wege = ausKarten(Karten)
    const Kartenzeilen = vi.fn(wege.Kartenzeilen)

    // **Die Liste antwortet beim zweiten Mal anders** — so wie der Bestand es
    // täte, nachdem der Stil gewechselt ist.
    let stil = 'magic-m15'

    belegeApi({
      ...belegeFenster('de'),
      Sets: () =>
        Promise.resolve({
          sets: [
            {
              id: '7',
              name: 'Probeset',
              spiel: 'magic',
              stil,
              karten: 0,
              herkunft: '/a.mse-set',
              herkunftFehlt: false,
            },
          ],
        }),
      ...wege,
      Kartenzeilen,
      Karten,
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Stile: () =>
        Promise.resolve({
          stile: [
            {
              paket: 'magic-m15',
              anzeigename: 'Magic-m15',
              spiel: 'magic',
              version: '1',
              eigenesSpiel: true,
              gewaehlt: true,
            },
            {
              paket: 'magic-m15-altered',
              anzeigename: 'Magic-m15-altered',
              spiel: 'magic',
              version: '1',
              eigenesSpiel: true,
              gewaehlt: false,
            },
          ],
        }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      // **Der Dienst nennt immer Folgen** — hier keine, aber die Rückfrage
      // steht, und erst die Zustimmung wechselt (§7 und #41).
      Stilfolgen: () =>
        Promise.resolve({
          folgen: {
            ungedeuteteFelder: [],
            verloreneOptionen: [],
            spielwechsel: false,
          },
        }),
      StilWechseln: () => {
        stil = 'magic-m15-altered'

        return Promise.resolve({})
      },
    })

    render(<App />)

    await userEvent.click(await leisteneintrag('sets'))
    await userEvent.click(
      await screen.findByRole('button', {
        name: `Probeset — ${texte.seiten.sets.oeffnen}`,
      }),
    )

    await waitFor(() => {
      expect(Kartenzeilen).toHaveBeenCalled()
    })

    // **Der ganze Satz und kein Teilstück.** `magic-m15` steht auch als
    // Eintrag im Auswahlfeld der Stillasche; wer darauf prüfte, prüfte das
    // Auswahlfeld und nicht den Kopf.
    const kopfzeile = (stilpaket: string) => `magic · ${stilpaket}`

    // **Nur, was zu sehen ist** (UX-03.AK-02, #1675): die Setliste, von der
    // aus das Set aufging, bleibt verborgen stehen und nennt es auch.
    expect(
      await screen.findByText(kopfzeile('magic-m15'), { ignore: verborgen }),
    ).toBeInTheDocument()

    // **Der Stil steht in der Schublade „Eigenschaften"** (UX-08, #1680),
    // die bei ihm aufgeht; der Kopf bleibt darunter stehen.
    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.allgemein.arbeitsplatz.schubladen.eigenschaften.knopf,
      }),
    )
    await screen.findByRole('tab', {
      name: texte.seiten.stil.titel,
      selected: true,
    })

    // **Über die Auswahl** — seit UXR-07 (#1915) kein Optionsfeld je Stil.
    await waehleZeile(
      await screen.findByRole('combobox', {
        name: texte.seiten.karten.stilWaehlen,
      }),
      'Magic-m15-altered magic-m15-altered',
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.karten.stilUebernehmen,
      }),
    )

    // **Ohne die Lasche zu verlassen**: der Kopf steht daneben und nennt jetzt
    // den neuen.
    expect(
      await screen.findByText(kopfzeile('magic-m15-altered'), {
        ignore: verborgen,
      }),
    ).toBeInTheDocument()
  })
})
