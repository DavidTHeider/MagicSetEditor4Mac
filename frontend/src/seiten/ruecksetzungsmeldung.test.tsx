// **Was eine Rücksetzung nach oben meldet, hängt an ihrer Art** — #2292
// (#2292).
//
// Drei Arten nehmen das Bezugsobjekt — Neuanfang, Profil, Mensch —, und nach
// ihnen räumt die Anwendung offenes Set und Karte wie nach dem Zurückspielen
// (`gewechselt`). Die drei leichteren lassen es stehen, ändern aber, was das
// Fenster gelesen hat; nach ihnen wird nur neu gelesen (`veraendert`). Ein
// Layout zurückzusetzen darf das offene Set nicht schließen.
//
// **Jede Verneinung hat ihren Zuhörer**: in jedem Fall wird der andere Rückruf
// gerufen, und damit steht fest, dass der Weg gelaufen ist (§3).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Ruecksetzung } from './Ruecksetzung'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const worte = kataloge.de.seiten.ruecksetzung

afterEach(() => {
  raeumeApi()
})

const faelle: [string, 'gewechselt' | 'veraendert'][] = [
  ['layout', 'veraendert'],
  ['profileinstellungen', 'veraendert'],
  ['nutzereinstellungen', 'veraendert'],
  ['profil', 'gewechselt'],
  ['nutzer', 'gewechselt'],
  ['anwendung', 'gewechselt'],
]

describe('Die Meldung nach einer Rücksetzung', () => {
  for (const [art, erwartet] of faelle) {
    it(`meldet nach »${art}« ${erwartet}`, async () => {
      belegeApi({
        Ruecksetzungsarten: () => Promise.resolve([art]),
        Ruecksetzungsvorschau: () =>
          Promise.resolve({
            art,
            betroffen: [],
            hindernisse: [],
            bereit: true,
            nameNoetig: false,
            nachweisNoetig: false,
            sicherungEmpfohlen: false,
            revision: 'marke',
            fehler: null,
          }),
        Zuruecksetzen: () =>
          Promise.resolve({
            art,
            zurueckgesetzt: 1,
            beiseite: '',
            fehler: null,
          }),
      })

      const rueckrufe = { gewechselt: vi.fn(), veraendert: vi.fn() }
      const nutzer = userEvent.setup()

      render(
        <Sprachraum sprache="de">
          <Ruecksetzung {...rueckrufe} />
        </Sprachraum>,
      )

      await nutzer.click(
        await screen.findByRole('button', { name: worte.nachsehen }),
      )
      await nutzer.click(await screen.findByRole('button', { name: worte.tun }))

      const anderer = erwartet === 'gewechselt' ? 'veraendert' : 'gewechselt'

      await waitFor(() => {
        expect(rueckrufe[erwartet]).toHaveBeenCalledTimes(1)
      })
      expect(rueckrufe[anderer]).not.toHaveBeenCalled()
    })
  }
})
