// **Eine Einfuhr, die Vorhandenes träfe, fragt vorher** — DOM-03.AK-01
// (#1698).
//
// Die Liste und das CSV: trifft eine Zeile eine, die die Sammlung schon
// führt, steht vor dem Schreiben die Vorschau — addieren, auslassen oder
// abbrechen. Trifft keine, geht es ohne Zwischenschritt durch (die übrigen
// Einfuhrtests belegen das mit `ohneDublette`).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Listeneinfuhr } from './Listeneinfuhr'
import { Sammlungsausfuhr } from './Sammlungsausfuhr'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'
import type {
  Listendeutungszeile,
  SammelvorschauDTO,
} from '../antwortenlisteneinfuhr'

const texte = kataloge.de.seiten.sammlung

afterEach(() => {
  raeumeApi()
})

const zeile: Listendeutungszeile = {
  roh: '4 Lightning Bolt (M11) 149',
  zeilennummer: 1,
  menge: 4,
  name: 'Lightning Bolt',
  setcode: 'm11',
  nummer: '149',
  finish: 'nonfoil',
  treffer: [
    {
      printing: 'druck-1',
      name: 'Lightning Bolt',
      setcode: 'm11',
      setname: 'Magic 2011',
      nummer: '149',
      seltenheit: 'common',
      bild: '',
    },
  ],
  drucke: 1,
  eindeutig: true,
  befund: '',
  weitere: false,
  fehler: null,
}

const dublette: SammelvorschauDTO = {
  posten: [
    {
      zeile: 1,
      bezeichnung: 'Lightning Bolt · M11 #149',
      menge: 4,
      zustand: 'nearMint',
      sprache: 'en',
      wunsch: false,
      art: 'zusammen',
      vorher: 2,
      nachher: 6,
    },
  ],
  befunde: [],
}

function bericht(ausgelassen: number) {
  return Promise.resolve({
    bericht: {
      angelegt: [],
      erhoeht: ausgelassen === 0 ? 1 : 0,
      ausgelassen,
      stueck: ausgelassen === 0 ? 4 : 0,
      befunde: [],
    },
    fehler: null,
  })
}

async function zeigeListe(einlesen: Api['ListeEinlesen'], vorschau = dublette) {
  belegeApi({
    ListeDeuten: () =>
      Promise.resolve({
        deutung: { zeilen: [zeile], befunde: [], alterSekunden: 0 },
        fehler: null,
      }),
    ListeVorschauen: () => Promise.resolve({ vorschau, fehler: null }),
    ListeEinlesen: einlesen,
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={vi.fn()} />
    </Sprachraum>,
  )

  await userEvent.type(
    screen.getByRole('textbox', { name: texte.listeFeld }),
    '4 Lightning Bolt (M11) 149',
  )
  await userEvent.click(screen.getByRole('button', { name: texte.listeDeuten }))
  await userEvent.click(
    await screen.findByRole('button', { name: texte.listeEinlesen }),
  )
}

describe('Die Listeneinfuhr mit einer Dublette', () => {
  it('fragt vorher und liest mit Auslassen ein', async () => {
    const einlesen = vi.fn(() => bericht(1))

    await zeigeListe(einlesen)

    expect(
      await screen.findByRole('region', { name: texte.vorschauTitel }),
    ).toBeInTheDocument()
    expect(einlesen).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('radio', { name: texte.vorschauAuslassen }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauEinlesen }),
    )

    await waitFor(() => {
      expect(einlesen).toHaveBeenCalledWith(
        [expect.objectContaining({ printing: 'druck-1', menge: 4 })],
        'auslassen',
      )
    })
    expect(
      await screen.findByText(
        einsetzen(texte.listeAusgelassen, { ausgelassen: '1' }),
      ),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('region', { name: texte.vorschauTitel }),
    ).toBeNull()
  })

  it('schreibt nichts, wenn abgebrochen wird', async () => {
    const einlesen = vi.fn<Api['ListeEinlesen']>()

    await zeigeListe(einlesen)

    await userEvent.click(
      await screen.findByRole('button', { name: texte.vorschauAbbrechen }),
    )

    expect(
      screen.queryByRole('region', { name: texte.vorschauTitel }),
    ).toBeNull()
    expect(einlesen).not.toHaveBeenCalled()
  })

  // **Eine Textänderung räumt die Vorschau mit** (§7, Regel 1): sie gehört
  // zu der Liste, aus der sie stammt.
  it('räumt die Vorschau, wenn sich der Text ändert', async () => {
    await zeigeListe(vi.fn<Api['ListeEinlesen']>())

    await screen.findByRole('region', { name: texte.vorschauTitel })
    await userEvent.type(
      screen.getByRole('textbox', { name: texte.listeFeld }),
      'x',
    )

    expect(
      screen.queryByRole('region', { name: texte.vorschauTitel }),
    ).toBeNull()
  })

  it('meldet einen Fehler der Vorschau und schreibt nichts', async () => {
    const einlesen = vi.fn<Api['ListeEinlesen']>()

    belegeApi({
      ListeDeuten: () =>
        Promise.resolve({
          deutung: { zeilen: [zeile], befunde: [], alterSekunden: 0 },
          fehler: null,
        }),
      ListeVorschauen: () =>
        Promise.resolve({
          vorschau: null,
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'Die Sammlung antwortet nicht.',
          },
        }),
      ListeEinlesen: einlesen,
    })

    render(
      <Sprachraum sprache="de">
        <Listeneinfuhr aufAufgenommen={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.type(
      screen.getByRole('textbox', { name: texte.listeFeld }),
      '4 Lightning Bolt (M11) 149',
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.listeDeuten }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: texte.listeEinlesen }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(einlesen).not.toHaveBeenCalled()
  })
})

describe('Das CSV mit einer Dublette', () => {
  async function zeigeCSV(
    einlesen: Api['SammlungCSVEinlesen'],
    vorschau: Promise<unknown>,
  ) {
    belegeApi({
      SammlungAusgeben: (form: string) =>
        Promise.resolve({
          form,
          text: 'printing,menge\ndruck-1,4\n',
          fehler: null,
        }),
      SammlungCSVVorschauen: () => vorschau as never,
      SammlungCSVEinlesen: einlesen,
    })

    render(
      <Sprachraum sprache="de">
        <Sammlungsausfuhr aufEingelesen={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: texte.ausfuhrAlsCSV }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: texte.ausfuhrZurueck }),
    )
  }

  it('fragt vorher und liest mit Addieren ein', async () => {
    const einlesen = vi.fn(() => bericht(0))

    await zeigeCSV(
      einlesen,
      Promise.resolve({ vorschau: dublette, fehler: null }),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.vorschauEinlesen }),
    )

    await waitFor(() => {
      expect(einlesen).toHaveBeenCalledWith(
        'printing,menge\ndruck-1,4\n',
        '',
        'addieren',
      )
    })
    expect(
      screen.queryByRole('region', { name: texte.vorschauTitel }),
    ).toBeNull()
  })

  // **Die eigene Ausfuhr ein zweites Mal, mit Auslassen** — nichts
  // verdoppelt sich, und der Bericht sagt, was stehen blieb.
  it('liest mit Auslassen ein und nennt, was stehen blieb', async () => {
    const einlesen = vi.fn(() => bericht(1))

    await zeigeCSV(
      einlesen,
      Promise.resolve({ vorschau: dublette, fehler: null }),
    )

    await userEvent.click(
      await screen.findByRole('radio', { name: texte.vorschauAuslassen }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauEinlesen }),
    )

    expect(
      await screen.findByText(
        einsetzen(texte.listeAusgelassen, { ausgelassen: '1' }),
      ),
    ).toBeInTheDocument()
    expect(einlesen).toHaveBeenCalledWith(
      'printing,menge\ndruck-1,4\n',
      '',
      'auslassen',
    )
  })

  it('schreibt nichts, wenn abgebrochen wird', async () => {
    const einlesen = vi.fn<Api['SammlungCSVEinlesen']>()

    await zeigeCSV(
      einlesen,
      Promise.resolve({ vorschau: dublette, fehler: null }),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.vorschauAbbrechen }),
    )

    expect(einlesen).not.toHaveBeenCalled()
  })

  it('meldet einen Fehler der Vorschau und schreibt nichts', async () => {
    const einlesen = vi.fn<Api['SammlungCSVEinlesen']>()

    await zeigeCSV(
      einlesen,
      Promise.resolve({
        vorschau: null,
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'Die Sammlung antwortet nicht.',
        },
      }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(einlesen).not.toHaveBeenCalled()
  })

  // **Eine Textänderung räumt die Vorschau mit** (§7, Regel 1).
  it('räumt die Vorschau, wenn sich der Text ändert', async () => {
    await zeigeCSV(
      vi.fn<Api['SammlungCSVEinlesen']>(),
      Promise.resolve({ vorschau: dublette, fehler: null }),
    )

    await screen.findByRole('region', { name: texte.vorschauTitel })
    await userEvent.type(
      screen.getByRole('textbox', { name: texte.ausfuhrErgebnis }),
      'x',
    )

    expect(
      screen.queryByRole('region', { name: texte.vorschauTitel }),
    ).toBeNull()
  })
})
