// **Das Geheimnis eines Anmeldeanbieters** —
// #1401.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anbietergeheimnisse } from './Anbietergeheimnisse'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.einstellungen

afterEach(() => {
  raeumeApi()
})

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Anbietergeheimnisse />
    </Sprachraum>,
  )
}

// alle gibt eine Antwort mit den vier Anbietern.
function alle(hinterlegt: string[] = []) {
  return {
    anbieter: [
      {
        anbieter: 'github',
        pflicht: true,
        verboten: false,
        hinterlegt: hinterlegt.includes('github'),
      },
      {
        anbieter: 'google',
        pflicht: false,
        verboten: false,
        hinterlegt: hinterlegt.includes('google'),
      },
      {
        anbieter: 'microsoft',
        pflicht: false,
        verboten: true,
        hinterlegt: false,
      },
      {
        anbieter: 'x',
        pflicht: false,
        verboten: false,
        hinterlegt: hinterlegt.includes('x'),
      },
    ],
    fehler: null,
  }
}

describe('Die Anbietergeheimnisse', () => {
  // **Der Fall, für den es die Seite gibt**: GitHub verlangt eines, und es
  // fehlt. Bis #1401 stand dafür nur ein Verweis auf den Schlüsselbund.
  it('sagt, wo eines verlangt wird und fehlt', async () => {
    belegeApi({ Anbietergeheimnisse: vi.fn().mockResolvedValue(alle()) })

    zeige()

    expect(
      await screen.findByText(`GitHub: ${texte.geheimnisPflicht}`),
    ).toBeInTheDocument()
  })

  // **Wer keines darf, bekommt kein Feld** — ein ausfüllbares Feld, das die
  // Anmeldung danach scheitern ließe, ist schlimmer als keines.
  it('gibt Microsoft kein Feld, sondern den Grund', async () => {
    belegeApi({ Anbietergeheimnisse: vi.fn().mockResolvedValue(alle()) })

    zeige()

    expect(
      await screen.findByText(`Microsoft: ${texte.geheimnisVerboten}`),
    ).toBeInTheDocument()

    expect(
      screen.queryByLabelText(`${texte.geheimnis}: Microsoft`),
    ).not.toBeInTheDocument()
  })

  // **Und es lässt sich eintragen** — die Abnahme des Issues.
  it('trägt ein Geheimnis ein und lädt danach nach', async () => {
    const gefragt = vi
      .fn()
      .mockResolvedValueOnce(alle())
      .mockResolvedValue(alle(['github']))

    const gesetzt = vi.fn().mockResolvedValue({ fehler: null })

    belegeApi({
      Anbietergeheimnisse: gefragt,
      AnbietergeheimnisSetzen: gesetzt,
    })

    zeige()

    const feld = await screen.findByLabelText(`${texte.geheimnis}: GitHub`)

    await userEvent.type(feld, 'sehr geheim')
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.zugangSetzen}: GitHub` }),
    )

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledWith('github', 'sehr geheim')
    })

    // **Geschrieben ist geschrieben, und danach wird nachgeladen** (§7,
    // Regel 3) — sonst stünde weiter „wird verlangt und fehlt".
    expect(
      await screen.findByText(`GitHub: ${texte.zugangLiegt}`),
    ).toBeInTheDocument()
  })

  // **Ein hinterlegtes lässt sich wieder herausnehmen.**
  it('nimmt ein Geheimnis wieder heraus', async () => {
    const geloescht = vi.fn().mockResolvedValue({ fehler: null })

    belegeApi({
      Anbietergeheimnisse: vi.fn().mockResolvedValue(alle(['github'])),
      AnbietergeheimnisLoeschen: geloescht,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.zugangLoeschen}: GitHub`,
      }),
    )

    // **Erst nach der Frage** (UX-10.AK-04, #1682): zurück gibt das
    // Geheimnis nur der Anbieter. **Abbrechen nimmt nichts heraus.**
    expect(geloescht).not.toHaveBeenCalled()
    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.aktionen.abbrechen,
      }),
    )
    expect(geloescht).not.toHaveBeenCalled()
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.zugangLoeschen}: GitHub` }),
    )
    await userEvent.click(
      within(
        screen.getByRole('group', {
          name: `${texte.geheimnis} für GitHub löschen?`,
        }),
      ).getByRole('button', { name: texte.zugangLoeschenJa }),
    )

    await waitFor(() => {
      expect(geloescht).toHaveBeenCalledWith('github')
    })
  })

  // **Und ein Fehler steht am Feld** (§6) und nicht als anonymer Hinweis.
  //
  // **Zugesichert wird der Satz aus dem Katalog und nicht der aus Go.** Das
  // Fenster übersetzt über den Schlüssel; wer hier den Go-Satz erwartete,
  // prüfte eine Durchreiche, die es nicht gibt (§3, „zwischen den
  // Prüfebenen").
  it('zeigt einen Fehler am verursachenden Feld', async () => {
    belegeApi({
      Anbietergeheimnisse: vi.fn().mockResolvedValue(alle()),
      AnbietergeheimnisSetzen: vi.fn().mockResolvedValue({
        fehler: {
          code: 'validierung',
          schluessel: 'unbekannt',
          meldung: 'der Bund klemmt',
          feld: '',
          werte: {},
        },
      }),
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.zugangSetzen}: GitHub`,
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )
  })

  // **Ein Anbieter, den der Katalog noch nicht kennt, steht mit seinem
  // Schlüssel da** — und nicht ohne Namen.
  //
  // **Der Fall ist echt und nicht gestellt**: die Liste kommt aus Go, und ein
  // fünfter Anbieter dort ist einen Katalogeintrag früher da als hier. Ein
  // leerer Name wäre eine Zeile, die niemand zuordnen kann (§6).
  it('nennt einen unbekannten Anbieter bei seinem Schlüssel', async () => {
    belegeApi({
      Anbietergeheimnisse: vi.fn().mockResolvedValue({
        anbieter: [
          {
            anbieter: 'apple',
            pflicht: false,
            verboten: false,
            hinterlegt: false,
          },
        ],
        fehler: null,
      }),
    })

    zeige()

    expect(
      await screen.findByText(`apple: ${texte.zugangFehlt}`),
    ).toBeInTheDocument()
  })

  // **Und ein Lesefehler steht nicht im Nichts.**
  it('meldet einen Lesefehler', async () => {
    belegeApi({
      Anbietergeheimnisse: vi.fn().mockResolvedValue({
        anbieter: [],
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'nichts geht',
          feld: '',
          werte: {},
        },
      }),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )
  })
})
