// **Gegenüber finden, wo gehandelt und gespielt wird** — UXR-16
// (#1924).
//
// # Der Befund
//
// **Das Artboard *Gegenüber* beginnt mit „In deiner Nähe"** und darunter
// „Weiter weg? Dann tauscht ihr einen Einladungscode aus". Im Kasten
// „Gegenüber aufnehmen oder bestätigen" stand nur der zweite Teil: die Suche
// im Netz gab es allein unter *Einstellungen › Verbindungen*. Wer beim
// Handeln jemanden am selben Tisch aufnehmen wollte, ging dorthin, nahm ihn
// auf, kam zurück, schlug den Kasten auf und suchte ihn in der Liste.
//
// # Was jetzt gilt
//
// Die Suche steht im Kasten vorn, **gefragt wird weiter erst auf
// Knopfdruck**, und wer einen Fund aufnimmt, sieht gleich darunter dessen
// Vergleich — drei Klicks vom Kasten bis „Stimmt überein".
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Peerkontakte } from './Peerkontakte'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'
import { fund, gegenueber, zwecke } from '../test/gegenueberbestand'

import type { Peerkontakt } from '../apipeerkontakt'

const eigene = kataloge.de.seiten.technik

afterEach(() => {
  raeumeApi()
})

function liste(kontakte: Peerkontakt[]) {
  return Promise.resolve({ kontakte, zwecke, fehler: null })
}

describe('In deiner Nähe', () => {
  it('steht im Kasten vorn und fragt erst auf Knopfdruck', async () => {
    const gesucht = vi.fn(() =>
      Promise.resolve({ nachbarn: [fund('Ida')], verworfen: 0, fehler: null }),
    )

    belegeApi({ Peerkontakte: () => liste([]), Nachbarn: gesucht })

    render(
      <Sprachraum sprache="de">
        <Peerkontakte eingebettet />
      </Sprachraum>,
    )

    const nah = await screen.findByRole('region', {
      name: eigene.nachbarnTitel,
    })

    expect(gesucht).not.toHaveBeenCalled()

    await userEvent.click(
      within(nah).getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    expect(await within(nah).findByText('Idas Mac')).toBeInTheDocument()
  })

  it('steht in den Einstellungen nicht ein zweites Mal', async () => {
    belegeApi({ Peerkontakte: () => liste([]) })

    render(
      <Sprachraum sprache="de">
        <Peerkontakte />
      </Sprachraum>,
    )

    // **Erst die Liste, dann das Nein** — sonst wäre es vor der ersten
    // Antwort gesagt und hieße nichts.
    await userEvent.click(
      screen.getByRole('button', { name: eigene.kontakteZeigen }),
    )
    await screen.findByText(eigene.kontakteLeer)

    expect(
      screen.queryByRole('region', { name: eigene.nachbarnTitel }),
    ).toBeNull()
  })

  it('schlägt einen eben aufgenommenen Fund mit seinem Vergleich auf', async () => {
    let stand: Peerkontakt[] = [gegenueber('Anna')]
    const geaendert = vi.fn()

    belegeApi({
      Peerkontakte: () => liste(stand),
      Nachbarn: () =>
        Promise.resolve({
          nachbarn: [fund('Ida')],
          verworfen: 0,
          fehler: null,
        }),
      PeerkontaktAufnehmen: (kennung: string, name: string) => {
        stand = [
          ...stand,
          gegenueber('Ida', {
            kennung,
            eigenerName: '',
            selbstgewaehlterName: name,
            vergleichscode: '12345 67890 12345 67890 12345 67890',
          }),
        ]

        return liste(stand)
      },
      PeerkontaktAdresse: () => liste(stand),
    })

    render(
      <Sprachraum sprache="de">
        <Peerkontakte eingebettet aufGeaendert={geaendert} />
      </Sprachraum>,
    )

    const hand = userEvent.setup()

    await hand.click(
      await screen.findByRole('button', { name: eigene.nachbarnSuchen }),
    )
    await hand.click(
      await screen.findByRole('button', { name: eigene.nachbarnAufnehmen }),
    )

    // **Ida und nicht Anna** steht offen, obwohl Anna als Erste wartet —
    // wer eben aufgenommen hat, vergleicht als Nächstes mit ihm.
    await waitFor(() => {
      expect(
        screen.getByRole('button', {
          name: `${eigene.gegenueber.vergleichen}: Idas Mac`,
        }),
      ).toHaveAttribute('aria-expanded', 'true')
    })
    expect(
      screen.getByRole('button', {
        name: `${eigene.gegenueber.vergleichen}: Anna`,
      }),
    ).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.getByText('12345 67890 12345 67890 12345 67890'),
    ).toBeVisible()
    expect(geaendert).toHaveBeenCalledTimes(1)
  })

  // **Die Erklärung der Liste steht an der Liste** — im Kasten stand sie
  // obenan, über der Suche und der Einladung, und erklärte etwas, das erst
  // drei Abschnitte tiefer kam.
  it('erklärt die Liste dort, wo sie steht', async () => {
    belegeApi({ Peerkontakte: () => liste([gegenueber('Anna')]) })

    render(
      <Sprachraum sprache="de">
        <Peerkontakte eingebettet />
      </Sprachraum>,
    )

    await screen.findByText('Anna', { selector: 'li span' })

    const erklaerung = screen
      .getByText(kurzform(eigene.kontakteErklaerung))
      .closest('p') as HTMLElement
    const einladung = screen.getByRole('region', {
      name: eigene.einladung.titel,
    })

    expect(
      einladung.compareDocumentPosition(erklaerung) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  })
})
