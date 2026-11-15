// **Die eigene Übersetzung wird unter der Sprache gewählt** — LOC-02
// (#1688).
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { eintraege, zeigeEinstellungen } from '../test/einstellungsseite'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { sprachpaketprobe } from '../test/sprachprobe'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

afterEach(() => {
  raeumeApi()
})

const eigeneSprache: EinstellungDTO = {
  schluessel: 'sprache.eigene',
  wert: '',
  wirksam: '',
  werte: [],
  bereich: 'anwendung',
  ebene: 'nutzer',
  herkunft: 'nutzer',
  vorgabe: '',
  wirkung: 'sofort',
}

describe('die eigene Übersetzung auf der Einstellungsseite', () => {
  it('steht als Wahl über ihrer Liste und schreibt, was gewählt wird', async () => {
    belegeApi({
      Sprachpakete: () =>
        Promise.resolve({ pakete: [sprachpaketprobe()], fehler: null }),
    })

    const setze = zeigeEinstellungen({
      alleEinstellungen: [...eintraege, eigeneSprache],
    })

    // **Kein Textfeld daneben**: die Wahl steht im Bereich.
    expect(
      screen.queryByRole('textbox', { name: eigene.namen['sprache.eigene'] }),
    ).toBeNull()

    // **Seit UXR-19 hinter dem Aufklapper „Eigene Übersetzungen"** (#1927):
    // anlegen will das nur, wer eine Sprache vermisst.
    await userEvent.click(
      screen.getByRole('button', { name: eigene.eigenuebersetzung.titel }),
    )

    await screen.findByText('Português')
    await userEvent.selectOptions(
      screen.getByRole('combobox', {
        name: eigene.eigenuebersetzung.verwenden,
      }),
      'pt-BR',
    )

    expect(setze).toHaveBeenCalledWith('sprache.eigene', 'pt-BR')
  })
})
