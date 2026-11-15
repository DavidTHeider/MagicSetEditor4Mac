// **Kommandozeile und Schnittstellen sehen einander** — OPS03
// (#2349):
// die Verknüpfungen in `~/.local/bin` stehen in beiden Blöcken, und wer sie
// im einen ändert, soll im anderen nicht den Stand von vorhin sehen.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Werkzeugeinrichtung } from './Werkzeugeinrichtung'

const technik = kataloge.de.seiten.technik

const vorschau = {
  verzeichnis: '/Users/jemand/.local/bin',
  imPfad: true,
  pfadzeile: '',
  pruefbefehl: 'mse meta',
  eintraege: [],
  fehler: null,
}

const befehle = (stand: string) => ({
  schnittstellen: [
    {
      name: 'befehle',
      stand,
      art: 'verknuepfung',
      datei: '/Users/jemand/.local/bin',
      zeilen: [],
    },
  ],
  nurEintrag: '',
  fehler: null,
})

afterEach(raeumeApi)

describe('Werkzeugeinrichtung', () => {
  it('liest die Kommandozeile nach, wenn alles zurückgenommen ist — und umgekehrt', async () => {
    const gelesen = { befehlszeile: 0, schnittstellen: 0 }

    belegeApi({
      Befehlszeile: (handlung: string) => {
        if (handlung === '') {
          gelesen.befehlszeile++
        }

        return Promise.resolve(vorschau)
      },
      Schnittstellen: () => {
        gelesen.schnittstellen++

        return Promise.resolve(befehle('eingerichtet'))
      },
      SchnittstellenAlleWiderrufen: () =>
        Promise.resolve({
          ...befehle('offen'),
          ausgaenge: [{ name: 'befehle', ausgang: 'widerrufen', fehler: null }],
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Werkzeugeinrichtung />
      </Sprachraum>,
    )

    await screen.findByText(technik.schnittstellen.namen.befehle)
    await waitFor(() => {
      expect(gelesen).toEqual({ befehlszeile: 1, schnittstellen: 1 })
    })

    await userEvent.click(
      screen.getByRole('button', {
        name: technik.schnittstellen.alle.widerrufen,
      }),
    )
    await waitFor(() => {
      expect(gelesen.befehlszeile).toBe(2)
    })
    expect(
      await screen.findByText(technik.schnittstellen.alle.ausgaenge.widerrufen),
    ).toBeTruthy()

    await userEvent.click(
      screen.getByRole('button', { name: technik.befehlszeileHerausnehmen }),
    )
    await waitFor(() => {
      expect(gelesen.schnittstellen).toBe(2)
    })
  })
})
