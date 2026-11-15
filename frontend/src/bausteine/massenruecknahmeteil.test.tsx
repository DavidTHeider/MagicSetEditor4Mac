// **Eine Rücknahme, die auf halbem Weg scheitert** — Befund 4 des
// Zweitblicks über Phase 6.6 (DOM-03.AK-03,
// #1698).
//
// Je alter Lagerort ein Schreiben: scheitert das zweite, ist das erste
// geschrieben. **Geschrieben ist geschrieben** (§7, Regel 3) — die Sammlung
// wird nachgeladen, der Bericht nennt, was zurück ist, und *Zurücknehmen*
// bleibt für die Zeilen stehen, die noch nicht zurück sind.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Massenvorgaenge } from './Massenvorgaenge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'
import type { MassenstandDTO } from '../antwortensammlung'

const texte = kataloge.de.seiten.sammlung

afterEach(() => {
  raeumeApi()
})

function stand(id: number, lagerort: string): MassenstandDTO {
  return {
    id,
    fassung: 4,
    menge: 1,
    zustand: 'nearMint',
    lagerort,
    wunsch: false,
  }
}

it('lädt nach, nennt das Geschriebene und bietet den Rest wieder an', async () => {
  const aendern = vi.fn<Api['MasseAendern']>()
  aendern
    .mockResolvedValueOnce({
      bericht: {
        geaendert: 3,
        befunde: [],
        vorher: [stand(1, 'Kiste 4'), stand(2, 'Kiste 9'), stand(3, 'Kiste 4')],
      },
      fehler: null,
    })
    .mockResolvedValueOnce({
      bericht: {
        geaendert: 2,
        befunde: [],
        vorher: [stand(1, 'Ordner Rot'), stand(3, 'Ordner Rot')],
      },
      fehler: null,
    })
    .mockResolvedValueOnce({
      bericht: null,
      fehler: { code: 'unbekannt', schluessel: 'unbekannt', meldung: 'weg' },
    })
    .mockResolvedValueOnce({
      bericht: { geaendert: 1, befunde: [], vorher: [stand(2, 'Ordner Rot')] },
      fehler: null,
    })
  belegeApi({ MasseAendern: aendern })
  const aufGeaendert = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Massenvorgaenge
        gewaehlt={[
          { id: 1, fassung: 3 },
          { id: 2, fassung: 1 },
          { id: 3, fassung: 1 },
        ]}
        gesamt={3}
        aufGeaendert={aufGeaendert}
      />
    </Sprachraum>,
  )

  await userEvent.type(
    screen.getByRole('textbox', { name: texte.masseLagerortSetzen }),
    'Ordner Rot',
  )
  await userEvent.click(
    screen.getByRole('button', { name: texte.masseAendern }),
  )
  await userEvent.click(
    await screen.findByRole('button', {
      name: einsetzen(texte.masseZuruecknehmenVon, { anzahl: '3' }),
    }),
  )

  // **Gewartet wird auf die Wirkung** — Meldung, Bericht und Nachladen
  // entstehen im selben Zug.
  expect(await screen.findByRole('alert')).toBeInTheDocument()
  expect(
    screen.getByText(einsetzen(texte.masseZurueckgenommen, { anzahl: '2' })),
  ).toBeInTheDocument()
  expect(aufGeaendert).toHaveBeenCalledTimes(2)

  // **Der Rest bleibt rücknehmbar** — nur die Zeile, die noch nicht zurück
  // ist, und gegen ihre Fassung.
  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(texte.masseZuruecknehmenVon, { anzahl: '1' }),
    }),
  )

  await waitFor(() => {
    expect(aendern).toHaveBeenLastCalledWith({
      zeilen: [{ id: 2, fassung: 4 }],
      menge: null,
      zustand: null,
      lagerort: 'Kiste 9',
      wunsch: null,
    })
  })
  expect(
    await screen.findByText(
      einsetzen(texte.masseZurueckgenommen, { anzahl: '1' }),
    ),
  ).toBeInTheDocument()
  expect(screen.queryByRole('alert')).toBeNull()
})
