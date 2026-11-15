// **Wer unter *Verbindungen* jemanden aufnimmt, sieht ihn in der Liste
// darüber** — Befund aus UXR-16
// (#1924).
//
// Unter *Einstellungen › Verbindungen* stehen die Liste der Gegenüber
// (`Peerkontakte`) und „Wer in der Nähe ist" (`Nachbarschaft`) nebeneinander.
// „Als Kontakt aufnehmen" schrieb, und die Liste darüber blieb, wie sie war —
// gegen §7, Regel 3: nach einem geglückten Schreiben wird nachgeladen. Im
// Kasten „Gegenüber aufnehmen oder bestätigen" war das gelöst, weil dort
// beides **ein** Bauteil ist; hier sind es zwei Geschwister.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Netzbereich } from './Technik'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { fund, gegenueber, zwecke } from '../test/gegenueberbestand'

import type { Peerkontakt } from '../apipeerkontakt'

const eigene = kataloge.de.seiten.technik

afterEach(() => {
  raeumeApi()
})

function liste(kontakte: Peerkontakt[]) {
  return Promise.resolve({ kontakte, zwecke, fehler: null })
}

describe('die Verbindungen nach einer Aufnahme', () => {
  it('laden die Liste der Gegenüber nach', async () => {
    let stand: Peerkontakt[] = [gegenueber('Anna')]

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
          }),
        ]

        return liste(stand)
      },
      PeerkontaktAdresse: () => liste(stand),
    })

    render(
      <Sprachraum sprache="de">
        <Netzbereich />
      </Sprachraum>,
    )

    const hand = userEvent.setup()

    await hand.click(
      await screen.findByRole('button', { name: eigene.kontakteZeigen }),
    )
    const kontakte = (
      await screen.findByText('Anna', { selector: 'li span' })
    ).closest('ul') as HTMLElement

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )
    await hand.click(
      await screen.findByRole('button', { name: eigene.nachbarnAufnehmen }),
    )

    expect(
      await within(kontakte).findByText('Idas Mac', { selector: 'li span' }),
    ).toBeInTheDocument()
  })
})
