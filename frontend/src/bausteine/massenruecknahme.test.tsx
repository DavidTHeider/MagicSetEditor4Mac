// **Eine Massenänderung lässt sich zurücknehmen** — DOM-03.AK-03
// (#1698).
//
// Ein Lagerort für viele Zeilen überschreibt viele Lagerorte, und die
// Sammlung hat keinen Verlauf. Der Bericht der Änderung nennt je Zeile den
// Lagerort vorher und die neue Fassung; *Zurücknehmen* schreibt jeden alten
// Lagerort an seine Zeilen zurück — gegen die neue Fassung, damit nichts
// überschrieben wird, was seither jemand anderes schrieb (§7, Regel 2).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Massenvorgaenge } from './Massenvorgaenge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'
import type { MassenstandDTO, MassenzeileDTO } from '../antwortensammlung'

const texte = kataloge.de.seiten.sammlung

afterEach(() => {
  raeumeApi()
})

const drei: MassenzeileDTO[] = [
  { id: 1, fassung: 3 },
  { id: 2, fassung: 1 },
  { id: 3, fassung: 1 },
]

function stand(id: number, lagerort: string): MassenstandDTO {
  return {
    id,
    fassung: id === 1 ? 4 : 2,
    menge: 1,
    zustand: 'nearMint',
    lagerort,
    wunsch: false,
  }
}

const vorher = [stand(1, 'Kiste 4'), stand(2, 'Kiste 9'), stand(3, 'Kiste 4')]

function zeige(aendern: Api['MasseAendern'], aufGeaendert = vi.fn()) {
  belegeApi({
    MasseAendern: aendern,
    MasseLoeschen: () =>
      Promise.resolve({
        bericht: { geaendert: 3, befunde: [], vorher: [] },
        fehler: null,
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Massenvorgaenge gewaehlt={drei} gesamt={3} aufGeaendert={aufGeaendert} />
    </Sprachraum>,
  )

  return aufGeaendert
}

async function setzeLagerort() {
  await userEvent.type(
    screen.getByRole('textbox', { name: texte.masseLagerortSetzen }),
    'Ordner Rot',
  )
  await userEvent.click(
    screen.getByRole('button', { name: texte.masseAendern }),
  )
}

describe('Die Rücknahme einer Massenänderung', () => {
  it('schreibt jeden alten Lagerort an seine Zeilen zurück', async () => {
    // **Jeder Ruf zählt, was er bekam** — und nennt in `vorher`, was er
    // schrieb, wie der Dienst es tut (`Massenpflege.Aendern`).
    const aendern = vi.fn<Api['MasseAendern']>((aenderung) =>
      Promise.resolve({
        bericht: {
          geaendert: aenderung.zeilen.length,
          befunde: [],
          vorher:
            aenderung.lagerort === 'Ordner Rot'
              ? vorher
              : aenderung.zeilen.map((zeile) =>
                  stand(zeile.id, aenderung.lagerort ?? ''),
                ),
        },
        fehler: null,
      }),
    )
    const aufGeaendert = zeige(aendern)

    await setzeLagerort()
    await userEvent.click(
      await screen.findByRole('button', {
        name: einsetzen(texte.masseZuruecknehmenVon, { anzahl: '3' }),
      }),
    )

    await waitFor(() => {
      expect(aendern).toHaveBeenCalledTimes(3)
    })
    expect(aendern).toHaveBeenCalledWith(
      expect.objectContaining({
        zeilen: [
          { id: 1, fassung: 4 },
          { id: 3, fassung: 2 },
        ],
        lagerort: 'Kiste 4',
      }),
    )
    expect(aendern).toHaveBeenCalledWith(
      expect.objectContaining({
        zeilen: [{ id: 2, fassung: 2 }],
        lagerort: 'Kiste 9',
      }),
    )
    expect(
      await screen.findByText(
        einsetzen(texte.masseZurueckgenommen, { anzahl: '3' }),
      ),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', {
        name: einsetzen(texte.masseZuruecknehmenVon, { anzahl: '3' }),
      }),
    ).toBeNull()
    expect(aufGeaendert).toHaveBeenCalledTimes(2)
  })

  // **Was seither jemand anderes schrieb, bleibt** — und steht als Befund
  // da, statt überschrieben zu werden.
  it('nennt, was sich nicht zurücknehmen ließ', async () => {
    // **Jede Zeile der Rücknahme ist seither überholt** — und der Dienst
    // nennt je Zeile ihre Kennung.
    const aendern = vi.fn<Api['MasseAendern']>((aenderung) =>
      Promise.resolve({
        bericht:
          aenderung.lagerort === 'Ordner Rot'
            ? { geaendert: 3, befunde: [], vorher }
            : {
                geaendert: 0,
                befunde: aenderung.zeilen.map(
                  (zeile) =>
                    `Exemplar ${String(zeile.id)}: hat sich seit dem Lesen geändert`,
                ),
                vorher: [],
              },
        fehler: null,
      }),
    )

    zeige(aendern)

    await setzeLagerort()
    await userEvent.click(
      await screen.findByRole('button', {
        name: einsetzen(texte.masseZuruecknehmenVon, { anzahl: '3' }),
      }),
    )

    expect(
      await screen.findByText(
        einsetzen(texte.masseZurueckgenommen, { anzahl: '0' }),
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByText('Exemplar 2: hat sich seit dem Lesen geändert'),
    ).toBeInTheDocument()
  })

  // **Ein Bericht, der `null` ist, zählt nichts** — wie beim Ändern selbst.
  it('kommt mit einem leeren Bericht der Rücknahme zurecht', async () => {
    const aendern = vi.fn<Api['MasseAendern']>()
    aendern
      .mockResolvedValueOnce({
        bericht: { geaendert: 3, befunde: [], vorher },
        fehler: null,
      })
      .mockResolvedValue({ bericht: null, fehler: null })

    zeige(aendern)

    await setzeLagerort()
    await userEvent.click(
      await screen.findByRole('button', {
        name: einsetzen(texte.masseZuruecknehmenVon, { anzahl: '3' }),
      }),
    )

    expect(
      await screen.findByText(
        einsetzen(texte.masseZurueckgenommen, { anzahl: '0' }),
      ),
    ).toBeInTheDocument()
  })

  it('meldet einen Fehler der Rücknahme', async () => {
    const aendern = vi.fn<Api['MasseAendern']>()
    aendern
      .mockResolvedValueOnce({
        bericht: { geaendert: 3, befunde: [], vorher },
        fehler: null,
      })
      .mockResolvedValue({
        bericht: null,
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'Die Sammlung antwortet nicht.',
        },
      })

    zeige(aendern)

    await setzeLagerort()
    await userEvent.click(
      await screen.findByRole('button', {
        name: einsetzen(texte.masseZuruecknehmenVon, { anzahl: '3' }),
      }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  // **Ohne Stand vorher gibt es nichts zurückzunehmen.**
  it('bietet keine Rücknahme an, wenn nichts geschrieben wurde', async () => {
    zeige(() =>
      Promise.resolve({
        bericht: { geaendert: 0, befunde: [], vorher: [] },
        fehler: null,
      }),
    )

    await setzeLagerort()

    expect(
      await screen.findByText(
        einsetzen(texte.masseBericht, { geaendert: '0' }),
      ),
    ).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /zurücknehmen/i })).toBeNull()
  })
})
