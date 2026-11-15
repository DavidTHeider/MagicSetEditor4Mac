// **Der nächste Schritt ist ein Zug** — P13
// (#2381),
// Befund C-16, Vermerk 0200: der Hauptknopf geht, was seine Entscheidung
// umfasst, und bringt es zum Gegenüber; was unterwegs scheitert, steht am
// Knopf mit seinem eigenen Ausweg.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { du, einDienst, einStand } from '../test/handelszugprobe'
import { folge, vorbereitet, zeige } from '../test/zugbuehne'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

const texte = kataloge.de.seiten.handeln
const zug = texte.zug

afterEach(() => {
  raeumeApi()
})

describe('der Zug am Hauptknopf', () => {
  it('heißt, was er tut, und bringt es zum Gegenüber', async () => {
    const { api } = einDienst(folge, {}, vorbereitet)
    belegeApi(api as never)
    const { gebucht, geaendert } = zeige(vorbereitet)

    await userEvent.click(
      screen.getByRole('button', { name: zug.uebergebenVerbuchen }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent(
      einsetzen(zug.hinaus, { name: 'Bert' }),
    )
    // **Gebucht ist gebucht**: Journal und Beleg lesen nach.
    expect(gebucht).toHaveBeenCalledTimes(1)
    expect(geaendert).not.toHaveBeenCalled()
  })

  it.each([
    [
      'die Vorbereitung samt Übergabe',
      einStand({ zustand: 'angenommen', naechster: 'vorbereiten' }),
      zug.vorbereitenUebergeben,
    ],
    [
      'die Übergabe allein, solange das Gegenüber nicht bestätigt hat',
      einStand({ ...vorbereitet, zustimmungen: [] }),
      texte.schrittUebergeben,
    ],
    [
      'das Zustimmen',
      einStand({ zustand: 'angeboten', naechster: 'zustimmen' }),
      texte.schrittZustimmen,
    ],
  ])('nennt %s', (_fall, stand, name) => {
    belegeApi(einDienst({}).api as never)
    zeige({ ...stand, moegliche: [stand.naechster] })

    expect(screen.getByRole('button', { name })).toHaveClass('bg-marke')
  })

  it('sagt beim Anbieten, dass es zugleich zustimmt', () => {
    belegeApi(
      einDienst(
        {},
        {},
        einStand({ naechster: 'anbieten', moegliche: ['anbieten'] }),
      ).api as never,
    )
    zeige(einStand({ naechster: 'anbieten', moegliche: ['anbieten'] }))

    // **Im Namen und nicht daneben** (Vermerk 0200, Artboard
    // `Markttisch09`): eine Zustimmung, die niemand sieht, wäre keine.
    expect(
      screen.getByRole('button', { name: zug.anbietenZustimmen }),
    ).toHaveClass('bg-marke')
  })

  it('trägt den neuen Stand ein, wo nichts gebucht wurde', async () => {
    const angenommen = einStand({ zustand: 'angenommen', lage: 'wartet' })
    const angeboten = einStand({
      zustand: 'angeboten',
      beschluss: 'zustimmen',
      zustimmungen: [du],
      naechster: 'zustimmen',
      moegliche: ['zustimmen'],
    })
    belegeApi(einDienst({ zustimmen: angenommen }, {}, angeboten).api as never)
    const { geaendert, gebucht } = zeige(angeboten)

    await userEvent.click(
      screen.getByRole('button', { name: texte.schrittZustimmen }),
    )

    await vi.waitFor(() => {
      expect(geaendert).toHaveBeenCalledWith(angenommen)
    })
    expect(gebucht).not.toHaveBeenCalled()
  })

  it('verbucht mit „Vorbereiten und übergeben" nicht, auch wenn die Übergabe dabei schließt', async () => {
    // **Der Knopf geht so weit, wie sein Name sagt** (Befund 1 des
    // Zweitblicks über Phase 8).
    const angenommen = einStand({
      zustand: 'angenommen',
      naechster: 'vorbereiten',
      moegliche: ['vorbereiten', 'abbrechen'],
      phase: 'zusagen',
    })
    const geschlossen = einStand({
      zustand: 'uebergabeOffen',
      naechster: 'verbuchen',
      phase: 'verbuchen',
    })
    const dienst = einDienst(
      {
        vorbereiten: einStand({
          zustand: 'vorbereitet',
          beschluss: 'uebergeben',
          naechster: 'uebergeben',
        }),
        uebergeben: geschlossen,
      },
      {},
      angenommen,
    )
    belegeApi(dienst.api as never)
    const { geaendert } = zeige(angenommen)

    await userEvent.click(
      screen.getByRole('button', { name: zug.vorbereitenUebergeben }),
    )

    await vi.waitFor(() => {
      expect(geaendert).toHaveBeenCalledWith(geschlossen)
    })
    expect(dienst.rufe).not.toContain('verbuchen')
  })

  it('geht einen Nebenweg als einzelnen Schritt, wie bisher', async () => {
    const { api, rufe } = einDienst(
      {
        abbrechen: einStand({ zustand: 'abgebrochen', lage: 'abgebrochen' }),
      },
      {},
      vorbereitet,
    )
    belegeApi(api as never)
    const { geaendert } = zeige(vorbereitet)

    await userEvent.click(
      screen.getByRole('button', { name: texte.schrittAbbrechen }),
    )

    await vi.waitFor(() => {
      expect(geaendert).toHaveBeenCalled()
    })
    expect(rufe).toEqual(['abbrechen'])
  })
})
