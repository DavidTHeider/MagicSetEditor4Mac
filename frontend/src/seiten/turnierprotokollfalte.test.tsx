// **Das Protokoll zeigt das Neueste, und davor stückweise** — UXR-12
// (#1920).
//
// # Der Befund
//
// Ein Abend mit sechzehn Teilnehmenden und vier Runden schreibt 51 Einträge
// ins Protokoll (`test/turnierbestand.ts`): Anlegen, sechzehn Anmeldungen,
// vier Runden, 29 Ergebnisse und eine Berichtigung. Untereinander zu je
// rund 24 Punkten sind das über 1.200 Punkte in einer Lasche — R4 aus
// `docs/UIREVIEW.md` erlaubt eine Rolltiefe von 2,5, und gesucht wird
// während des Abends fast immer das, was eben geschah.
//
// # Was jetzt gilt
//
// **Die Reihenfolge bleibt die, in der es geschah** (Vermerk 0139). Gezeigt
// werden die letzten zwanzig; darüber sagt eine Zeile, wie viele von wie
// vielen dastehen, und ein Knopf holt die früheren dazu — dieselben Worte wie
// die `Faltliste`, und wie dort bleibt der Fokus nicht im Leeren, wenn der
// Knopf mit dem letzten Stück verschwindet.
import { within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { aufgeschlagen } from './turnierwerkbank'
import { raeumeApi } from '../test/api'
import { abend, protokoll, rangliste } from '../test/turnierbestand'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.turniere
const worte = kataloge.de.bausteine.listenansicht

afterEach(() => {
  raeumeApi()
})

describe('das Protokoll eines Abends', () => {
  it('zeigt die letzten zwanzig und holt die früheren dazu', async () => {
    const werkbank = await aufgeschlagen(
      abend,
      {
        TurnierRangliste: () => Promise.resolve({ rangliste, fehler: null }),
        TurnierProtokoll: () =>
          Promise.resolve({ eintraege: protokoll, fehler: null }),
      },
      'protokoll',
    )

    const liste = await within(werkbank).findByRole('list', {
      name: texte.gebietProtokoll,
    })
    const zeilen = within(liste).getAllByRole('listitem')
    expect(protokoll).toHaveLength(51)
    expect(zeilen).toHaveLength(20)
    expect(zeilen.at(-1)).toHaveTextContent(
      `Ergebnis Cleo ${texte.gegen} Lars: 1:2:0`,
    )

    const stand = within(werkbank).getByRole('status')
    expect(stand).toHaveTextContent(
      einsetzen(worte.zahl, { gezeigt: '20', gesamt: '51' }),
    )

    await userEvent.click(
      within(werkbank).getByRole('button', {
        name: einsetzen(worte.weitere, { anzahl: '31' }),
      }),
    )

    const alle = within(liste).getAllByRole('listitem')
    expect(alle).toHaveLength(51)
    expect(alle[0]).toHaveTextContent('Angelegt als »Dämmerpokal«')
    expect(stand).toHaveTextContent(
      einsetzen(worte.zahl, { gezeigt: '51', gesamt: '51' }),
    )
    expect(stand).toHaveFocus()
  })

  // **Stückweise, nicht alles auf einmal** — ein Turnier über viele Abende
  // schreibt hunderte Einträge, und ein Klick holt die nächsten fünfzig.
  it('holt ein langes Protokoll stückweise und bleibt am Knopf', async () => {
    const lang = Array.from({ length: 100 }, (_, stelle) => ({
      zeit: '2026-10-01T18:00:00Z',
      vorgang: 'angemeldet',
      angaben: `Gast ${String(stelle + 1)}`,
      paarungID: '',
    }))
    const werkbank = await aufgeschlagen(
      abend,
      {
        TurnierRangliste: () => Promise.resolve({ rangliste, fehler: null }),
        TurnierProtokoll: () =>
          Promise.resolve({ eintraege: lang, fehler: null }),
      },
      'protokoll',
    )
    const liste = await within(werkbank).findByRole('list', {
      name: texte.gebietProtokoll,
    })

    const knopf = within(werkbank).getByRole('button', {
      name: einsetzen(worte.weitere, { anzahl: '50' }),
    })
    await userEvent.click(knopf)

    expect(within(liste).getAllByRole('listitem')).toHaveLength(70)
    expect(knopf).toHaveAccessibleName(
      einsetzen(worte.weitere, { anzahl: '30' }),
    )
    expect(knopf).toHaveFocus()
  })
})
