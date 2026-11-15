// **Welches Exemplar weggeht — gewählt aus der Sammlung** — UX-08.AK-04
// (#1680).
//
// # Was hier gemessen wird
//
// **Dass niemand Kennung oder Fassung einer Besitzzeile tippt**: je eigener
// Position stehen die Zeilen der Sammlung mit diesem Druck zur Wahl, und
// beide Zahlen kommen aus der gewählten. **Ein Wunsch ist kein Besitz**, und
// die Positionen des Gegenübers fragen die Sammlung gar nicht erst.
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { keineHerkunft } from '../apiherkunft'
import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { zeigeHandel, zuPhase } from './tauschprobe'
import { kataloge } from '../sprache/kataloge'

import type { ExemplarDTO } from '../api'
import type { Handelsposition, Handelsvorgang } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

// einePosition ist eine Zeile eines Angebots.
function einePosition(anders: Partial<Handelsposition>): Handelsposition {
  return {
    von: 'bich',
    printing: 'bolt-m10',
    druckId: 0,
    menge: 2,
    guete: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    abbild: '',
    bewertung: 0,
    ...anders,
  }
}

// einHandel ist der gewählte Handel mit einer eigenen und einer fremden Zeile.
function einHandel(eigeneZeile: Partial<Handelsposition> = {}): Handelsvorgang {
  return {
    vorgang: 'h-1',
    revision: 2,
    eltern: 'ab',
    abdruck: 'abcdef0123456789',
    art: 'tausch',
    initiator: 'bich',
    parteien: ['bich', 'bgegenueber'],
    ich: 'bich',
    zustand: 'angeboten',
    positionen: [
      einePosition(eigeneZeile),
      einePosition({ von: 'bgegenueber', printing: 'shivan-lea', menge: 1 }),
    ],
    bewertungGeteilt: false,
    zustimmungen: [],
    ende: false,
    beschluss: '',
    lage: 'dran',
    naechster: 'zustimmen',
    moegliche: ['zustimmen', 'abbrechen', 'ablaufen'],
    phase: 'angebot',
  }
}

// einExemplar ist eine Zeile der eigenen Sammlung.
function einExemplar(teil: Partial<ExemplarDTO> = {}): ExemplarDTO {
  return {
    id: 7,
    fassung: 3,
    printing: 'bolt-m10',
    druckID: '',
    vorlaeufig: false,
    menge: 4,
    zustand: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: 'Blitzschlag',
    setcode: 'm10',
    nummer: '146',
    lagerort: 'Ordner A',
    wunsch: false,
    bezeichnung: 'Blitzschlag',
    herkunft: keineHerkunft,
    ...teil,
  }
}

// zeigeMitBesitz stellt den Handel auf und wählt ihn.
async function zeigeMitBesitz(
  zeilen: ExemplarDTO[] | Error,
  mehr: Record<string, unknown> = {},
  eigeneZeile: Partial<Handelsposition> = {},
) {
  const SammlungBesitz = vi.fn(() =>
    zeilen instanceof Error
      ? Promise.reject(zeilen)
      : Promise.resolve({
          besitz: {
            printing: 'bolt-m10',
            stueck: 4,
            gewuenscht: 0,
            besitzt: true,
            zeilen,
          },
          fehler: null,
        }),
  )

  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })]),
      ),
    HandelVorgaenge: () =>
      Promise.resolve({ vorgaenge: [einHandel(eigeneZeile)], fehler: null }),
    SammlungBesitz,
    ...mehr,
  })

  const hand = await zeigeHandel()

  // **Gewählt wird in „Mit wem", festgehalten in „Zusagen"** (TRADE-02,
  // #1716).
  await zuPhase(hand, 'gegenueber')
  await hand.click(
    await screen.findByRole('radio', { name: /Angeboten · Revision 2/ }),
  )
  await zuPhase(hand, 'zusagen')

  return { hand, SammlungBesitz }
}

describe('Welches Exemplar weggeht, aus der Sammlung gewählt', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('hält eine gewählte Zeile fest, ohne Kennung oder Fassung zu tippen', async () => {
    const HandelFesthalten = vi.fn(() =>
      Promise.resolve({ vorgang: 'h-1', zusagen: [], fehler: null }),
    )

    const { hand, SammlungBesitz } = await zeigeMitBesitz(
      [einExemplar(), einExemplar({ id: 8, wunsch: true, lagerort: 'Wunsch' })],
      { HandelFesthalten },
    )

    const kasten = screen.getByRole('region', { name: eigene.zusagenTitel })
    const wahl = await within(kasten).findByRole('group', {
      name: '2 × Blitzschlag',
    })

    // **Nur die eigene Zeile fragt die Sammlung** — und ein Wunsch steht
    // nicht zur Wahl.
    expect(SammlungBesitz).toHaveBeenCalledTimes(1)
    expect(SammlungBesitz).toHaveBeenCalledWith('bolt-m10')
    expect(within(wahl).getAllByRole('radio')).toHaveLength(1)

    expect(
      within(wahl).getByRole('button', { name: eigene.zusagenFesthalten }),
    ).toBeDisabled()

    await hand.click(
      within(wahl).getByRole('radio', {
        name: 'Ordner A · Near Mint · Deutsch · Nonfoil · 4 Exemplare',
      }),
    )
    await hand.click(
      within(wahl).getByRole('button', { name: eigene.zusagenFesthalten }),
    )

    expect(HandelFesthalten).toHaveBeenCalledWith('h-1', 7, 2, 3)
  })

  it('sagt, wenn davon nichts in der Sammlung ist', async () => {
    await zeigeMitBesitz([])

    expect(
      await screen.findByText(eigene.zusagenKeinExemplar),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('group', { name: '2 × bolt-m10' }),
    ).toBeInTheDocument()
  })

  // **Ein Wert, den das Fenster nicht kennt, ist immer noch eine Auskunft** —
  // und eine Karte ohne Namen heißt nach ihrem Druck. Gesehen in der
  // WebKit-Probe: die Legende stand als „1 ×" da.
  it('zeigt den rohen Wert, wo es kein Wort dafür gibt', async () => {
    await zeigeMitBesitz(
      [
        einExemplar({
          lagerort: '',
          zustand: 'fremd',
          sprache: 'xx',
          finish: 'glanz' as ExemplarDTO['finish'],
          menge: 1,
          name: '',
          bezeichnung: '',
        }),
      ],
      {},
      { guete: 'fremd', sprache: 'xx', finish: 'glanz' },
    )

    expect(
      await screen.findByRole('group', { name: '2 × bolt-m10' }),
    ).toBeInTheDocument()
    expect(
      await screen.findByRole('radio', {
        name: `${eigene.zusagenOhneLagerort} · fremd · xx · glanz · 1 Exemplar`,
      }),
    ).toBeInTheDocument()
  })

  // **Nur was zur Absprache passt** (#2492): unter „Zusagen" stand jede
  // Zeile des Drucks zur Wahl, auch eine in anderem Zustand — sie ging weg,
  // drüben kam die abgesprochene an, und die Gegenzeichnung fiel still aus.
  it('bietet nur Zeilen an, die Zustand, Sprache und Ausführung erfüllen', async () => {
    await zeigeMitBesitz([
      einExemplar({ id: 7 }),
      einExemplar({ id: 8, zustand: 'excellent', lagerort: 'Ordner B' }),
      einExemplar({
        id: 9,
        finish: '' as ExemplarDTO['finish'],
        lagerort: 'Ordner C',
      }),
    ])

    const gruppe = await screen.findByRole('group', { name: '2 × Blitzschlag' })
    const wahl = await within(gruppe).findAllByRole('radio')

    expect(wahl.map((knopf) => (knopf as HTMLInputElement).value)).toEqual([
      '7',
      '9',
    ])
  })

  it('sagt, wenn keine Zeile zur Absprache passt, und was zu tun ist', async () => {
    await zeigeMitBesitz([einExemplar({ zustand: 'excellent' })])

    expect(
      await screen.findByText(eigene.zusagenKeinePassende),
    ).toBeInTheDocument()
    expect(screen.queryByRole('radio', { name: /Ordner A/ })).toBeNull()
  })

  it('bietet Wiederholen, wenn die Sammlung nicht antwortet', async () => {
    await zeigeMitBesitz(new Error('Bestand gesperrt'))

    expect(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.zustaende.wiederholen,
      }),
    ).toBeInTheDocument()
  })
})
