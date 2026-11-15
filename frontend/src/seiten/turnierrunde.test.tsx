// **Die Werkbank zeigt, was diese Runde braucht** — UXR-12
// (#1920).
//
// # Der Befund
//
// Gemessen unter jsdom mit `test/turnierbestand.ts` — sechzehn
// Teilnehmende, vier Runden zu acht Tischen, die vierte läuft: die Lasche
// *Runden* zeichnete **alle vier Runden offen**, und jede Paarung trug drei
// Kästchen, *Eintragen*, *Zurücknehmen* und den Abschnitt *Hier gespielt*
// mit seinen zwei Absätzen. 96 Felder auf einer Fläche, die R3 aus
// `docs/UIREVIEW.md` bei zwölf deckelt, und kein Hauptknopf.
//
// # Was jetzt gilt
//
// **Wie im Artboard *Turniere***: eine Paarung ist eine Zeile — Namen,
// Stand und ein leiser Knopf *Ergebnis*, der das Eintragen an dieser Zeile
// aufklappt. Die laufende Runde steht offen, die früheren eingeklappt —
// eine Berichtigung bleibt dort, wo sie war (PROTOTYP-0.6.0.md, „Runden"
// statt „Runde n"). *Runde legen* führt.
import { waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { aufgeschlagen } from './turnierwerkbank'
import { raeumeApi } from '../test/api'
import { zaehle } from '../test/bedienzaehlung'
import { verstoesse } from '../test/axe'
import { abend, liste, rangliste } from '../test/turnierbestand'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { ohneFolgen } from './turnierprobe'
import type { Api } from '../api'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

const runde = (nummer: number) =>
  einsetzen(texte.runde, { nummer: String(nummer) })

// amAbend schlägt den Dämmerpokal bei seinen Runden auf.
async function amAbend(weitere: Partial<Api> = {}) {
  const werkbank = await aufgeschlagen(
    abend,
    {
      Turniere: () => Promise.resolve({ turniere: liste, fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste, fehler: null }),
      TurnierKorrekturfolgen: ohneFolgen,
      ...weitere,
    },
    'runden',
  )

  return { werkbank, feld: within(werkbank).getByRole('tabpanel') }
}

describe('die Runden eines Abends mit sechzehn', () => {
  it('zeigt die laufende Runde ohne ein Feld', async () => {
    const { feld } = await amAbend()

    const laufende = within(feld).getByRole('list', { name: runde(4) })
    expect(within(laufende).getAllByRole('listitem')).toHaveLength(8)

    // **Je Tisch ein Knopf** — und keines der Kästchen.
    const zaehlung = zaehle(feld)
    expect(zaehlung.felder).toBe(0)
    expect(zaehlung.bedien).toBeLessThanOrEqual(20)
    expect(within(laufende).getAllByRole('button')).toHaveLength(8)
  })

  it('klappt die früheren Runden ein', async () => {
    const { feld } = await amAbend()

    for (const nummer of [1, 2, 3]) {
      expect(
        within(feld).queryByRole('list', { name: runde(nummer) }),
      ).toBeNull()
      expect(
        within(feld).getByRole('button', { name: runde(nummer) }),
      ).toHaveAttribute('aria-expanded', 'false')
    }
  })

  it('führt mit »Runde legen«', async () => {
    const { feld } = await amAbend()

    const legen = within(feld).getByRole('button', { name: texte.rundeLegen })
    expect(legen).toHaveClass('bg-marke')
    expect(zaehle(feld).haupt).toBe(1)
  })

  it('sagt an jeder Zeile, wie es steht', async () => {
    const { feld } = await amAbend()
    const laufende = within(feld).getByRole('list', { name: runde(4) })

    expect(within(laufende).getAllByText('2:1:0')).toHaveLength(5)
    expect(within(laufende).getAllByText(texte.offen)).toHaveLength(3)
  })

  it('klappt das Eintragen an seiner Zeile auf', async () => {
    const ErgebnisSetzen = vi.fn<Api['TurnierErgebnisSetzen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const { werkbank, feld } = await amAbend({
      TurnierErgebnisSetzen: ErgebnisSetzen,
    })
    const wer = `Fenja ${texte.gegen} Ida`

    const ergebnis = within(feld).getByRole('button', {
      name: `${texte.ergebnis}: ${wer}`,
    })
    expect(ergebnis).toHaveAttribute('aria-expanded', 'false')

    await userEvent.click(ergebnis)

    expect(ergebnis).toHaveAttribute('aria-expanded', 'true')
    expect(zaehle(feld).felder).toBe(3)
    expect(await verstoesse(werkbank)).toEqual([])

    await userEvent.type(
      within(feld).getByLabelText(`${texte.spieleA} — ${wer}`),
      '{Backspace}2',
    )
    await userEvent.click(
      within(feld).getByRole('button', { name: `${texte.eintragen}: ${wer}` }),
    )

    await waitFor(() => {
      expect(ErgebnisSetzen).toHaveBeenCalledWith('7', '1', 'p4-6', {
        spieleA: 2,
        spieleB: 0,
        unentschieden: 0,
      })
    })
  })

  it('zeigt das Spiel hier erst am aufgeklappten Ergebnis', async () => {
    const { feld } = await amAbend()

    expect(within(feld).queryByRole('region', { name: texte.match })).toBeNull()

    await userEvent.click(
      within(feld).getByRole('button', {
        name: `${texte.ergebnis}: Fenja ${texte.gegen} Ida`,
      }),
    )

    expect(
      within(feld).getAllByRole('region', { name: texte.match }),
    ).toHaveLength(1)
  })

  // **Eine Berichtigung bleibt, wo sie war** — in ihrer Runde, die man dafür
  // aufklappt.
  it('berichtigt eine frühere Runde nach dem Aufklappen', async () => {
    const ErgebnisSetzen = vi.fn<Api['TurnierErgebnisSetzen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const { feld } = await amAbend({ TurnierErgebnisSetzen: ErgebnisSetzen })
    const wer = `Cleo ${texte.gegen} Lars`

    await userEvent.click(within(feld).getByRole('button', { name: runde(2) }))
    const zweite = within(feld).getByRole('list', { name: runde(2) })
    await userEvent.click(
      within(zweite).getByRole('button', {
        name: `${texte.ergebnis}: ${wer}`,
      }),
    )
    await userEvent.click(
      within(zweite).getByRole('button', {
        name: `${texte.zuruecknehmen}: ${wer}`,
      }),
    )

    await waitFor(() => {
      expect(ErgebnisSetzen).toHaveBeenCalledWith('7', '1', 'p2-3', null)
    })
  })
})
