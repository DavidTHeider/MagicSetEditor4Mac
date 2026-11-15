// **Ein Symbol in den Fassungen eines Sets und die vorige Fassung** —
// DOM-02.AK-02 (#1697):
// die Größen- und Farbvorschau und der Weg zurück.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { belegeApi, raeumeApi } from '../test/api'
import {
  krone,
  mitSymbolen,
  ohneMarken,
  zeigeBibliothek,
} from '../test/symbolbibliothek'

const texte = kataloge.de.seiten.karten.symbolbibliothek

afterEach(() => {
  raeumeApi()
})

const eintrag = (name: string) => screen.findByRole('listitem', { name })

describe('Die Fassungen eines Symbols', () => {
  it('zeigen das Symbol in jeder Fassung und Größe', async () => {
    const fassungen = vi.fn().mockResolvedValue({
      fassungen: [
        {
          name: 'common',
          bilder: [
            { groesse: 16, vorschau: 'data:image/png;base64,QQ==' },
            { groesse: 32, vorschau: '' },
          ],
        },
      ],
      fehler: null,
    })
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
      MarkenJeObjekt: ohneMarken,
      BibliothekssymbolFassungen: fassungen,
    })

    zeigeBibliothek()
    const zeile = await eintrag('Krone')
    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.fassungen }),
    )

    const liste = await within(zeile).findByRole('list', {
      name: einsetzen(texte.fassungsliste, { name: 'Krone' }),
    })
    expect(fassungen).toHaveBeenCalledWith('1', '7')
    expect(
      within(liste).getByRole('img', {
        name: einsetzen(texte.fassungsbild, {
          name: 'Krone',
          fassung: 'common',
          groesse: '16',
        }),
      }),
    ).toHaveAttribute('src', 'data:image/png;base64,QQ==')
    expect(within(liste).getAllByRole('img')).toHaveLength(1)

    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.fassungenZu }),
    )
    expect(within(zeile).queryByRole('list')).not.toBeInTheDocument()
  })

  it('sagt, wenn der Stil keine nennt, und zeigt einen Ladefehler', async () => {
    const fassungen = vi
      .fn()
      .mockResolvedValueOnce({ fassungen: [], fehler: null })
      .mockRejectedValueOnce(new Error('klemmt'))
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
      MarkenJeObjekt: ohneMarken,
      BibliothekssymbolFassungen: fassungen,
    })

    zeigeBibliothek()
    const zeile = await eintrag('Krone')
    const knopf = within(zeile).getByRole('button', { name: texte.fassungen })

    await userEvent.click(knopf)
    expect(
      await within(zeile).findByText(texte.keineFassungen),
    ).toBeInTheDocument()

    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.fassungenZu }),
    )
    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.fassungen }),
    )
    expect(await within(zeile).findByRole('alert')).toBeInTheDocument()
  })
})

describe('Die vorige Fassung', () => {
  it('lässt sich zurückholen, wenn es sie gibt', async () => {
    const zurueck = vi.fn().mockResolvedValue({
      id: '1',
      name: 'Krone',
      veraltet: [],
      fehler: null,
    })
    const liste = vi
      .fn()
      .mockResolvedValue(mitSymbolen({ ...krone, vorige: true }))
    belegeApi({
      Symbolbibliothek: liste,
      MarkenJeObjekt: ohneMarken,
      BibliothekssymbolZuruecknehmen: zurueck,
    })

    zeigeBibliothek()
    await userEvent.click(
      within(await eintrag('Krone')).getByRole('button', {
        name: texte.zuruecknehmen,
      }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent(
      einsetzen(texte.zurueckgeholt, { name: 'Krone' }),
    )
    expect(zurueck).toHaveBeenCalledWith('1')
    await waitFor(() => {
      expect(liste).toHaveBeenCalledTimes(2)
    })
  })

  it('fehlt, wenn es keine gibt', async () => {
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
      MarkenJeObjekt: ohneMarken,
    })

    zeigeBibliothek()

    expect(
      within(await eintrag('Krone')).queryByRole('button', {
        name: texte.zuruecknehmen,
      }),
    ).not.toBeInTheDocument()
  })
})
