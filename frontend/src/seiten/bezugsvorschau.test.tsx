// **Eine Aktualisierung und ihr Rückweg** — G04
// (#2251).
//
// Sie zeigt vorher, was sie berührt, und verlangt eine Bestätigung, wo etwas
// wegfallen kann; der bisherige Stand bleibt zum Zurückholen.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import {
  aufschlagen,
  bezogen,
  bezug,
  spitze,
  t,
  vorschau,
} from '../test/bezuege'

import type { Api } from '../api'

afterEach(() => {
  raeumeApi()
})

describe('Eine Aktualisierung', () => {
  function belegt(teile: Partial<Api> = {}) {
    let runde = 0
    const bezuege = vi.fn(() => {
      runde += 1

      return Promise.resolve({
        bezuege: [
          bezug({
            lage: 'neuer',
            spitze,
            spitzeZeit: '2026-10-07T02:02:04Z',
            vorsprung: 63,
            stand: runde > 1 ? spitze : bezogen,
          }),
        ],
      })
    })

    return { PaketBezuege: bezuege, ...teile }
  }

  it('zeigt vorher, was sie berührt, und führt dann aus', async () => {
    const aktualisieren = vi.fn(() =>
      Promise.resolve({
        ort: '/x',
        stand: spitze,
        pruefsumme: '',
        uebergangen: [],
      }),
    )
    const vorschauen = vi.fn(() => Promise.resolve(vorschau()))
    const belegung = belegt({
      PaketBezugVorschau: vorschauen,
      PaketBezugAktualisieren: aktualisieren,
    })
    await aufschlagen(belegung)

    await userEvent.click(
      await screen.findByRole('button', {
        name: /Aktualisierung ansehen … \(M15-Magic-Pack\)/,
      }),
    )
    expect(vorschauen).toHaveBeenCalledWith('M15-Magic-Pack', spitze)

    const kasten = await screen.findByRole('region', {
      name: 'Aktualisierung von M15-Magic-Pack',
    })
    expect(kasten).toHaveTextContent(
      'magic-m15.mse-style: 1 neu, 2 geändert, 0 entfernt',
    )
    expect(kasten).toHaveTextContent(
      '4 weitere Dateien gehören zu keinem Paket.',
    )
    expect(kasten).toHaveTextContent(
      'Diese Sets nennen ein berührtes Paket: Mein Set',
    )
    expect(kasten).toHaveTextContent(t.aufheben)

    await userEvent.click(
      within(kasten).getByRole('button', { name: 'Auf c8e38c9 aktualisieren' }),
    )
    expect(aktualisieren).toHaveBeenCalledWith('M15-Magic-Pack', spitze, false)

    await vi.waitFor(() => {
      expect(belegung.PaketBezuege).toHaveBeenCalledTimes(2)
    })
    expect(
      screen.queryByRole('region', {
        name: 'Aktualisierung von M15-Magic-Pack',
      }),
    ).not.toBeInTheDocument()
  })

  it('verlangt eine Bestätigung, wo etwas wegfallen kann', async () => {
    const aktualisieren = vi.fn(() =>
      Promise.resolve({
        ort: '/x',
        stand: spitze,
        pruefsumme: '',
        uebergangen: [],
      }),
    )
    await aufschlagen(
      belegt({
        PaketBezugVorschau: () =>
          Promise.resolve(
            vorschau({
              pakete: [],
              uebrige: 0,
              sets: [],
              bestaetigungNoetig: true,
              gruende: ['setVerliert', 'gekuerzt'],
              vorigeErsetzt: true,
            }),
          ),
        PaketBezugAktualisieren: aktualisieren,
      }),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: /Aktualisierung ansehen/ }),
    )
    const kasten = await screen.findByRole('region', {
      name: 'Aktualisierung von M15-Magic-Pack',
    })
    expect(kasten).toHaveTextContent(t.keinePakete)
    expect(kasten).toHaveTextContent(t.keineSets)
    expect(kasten).toHaveTextContent(t.gruende.setVerliert)
    expect(kasten).toHaveTextContent(/Der schon aufgehobene Stand/)

    const knopf = within(kasten).getByRole('button', {
      name: 'Auf c8e38c9 aktualisieren',
    })
    expect(knopf).toBeDisabled()

    await userEvent.click(
      within(kasten).getByRole('checkbox', { name: t.bestaetigen }),
    )
    await userEvent.click(knopf)
    expect(aktualisieren).toHaveBeenCalledWith('M15-Magic-Pack', spitze, true)
  })

  it('meldet eine gescheiterte Vorschau und eine gescheiterte Aktualisierung am Baum', async () => {
    const vorschauen = vi
      .fn()
      .mockRejectedValueOnce(new Error('weg'))
      .mockResolvedValue(vorschau())
    const belegung = belegt({
      PaketBezugVorschau: vorschauen,
      PaketBezugAktualisieren: () => Promise.reject(new Error('fort')),
    })
    await aufschlagen(belegung)

    const ansehen = await screen.findByRole('button', {
      name: /Aktualisierung ansehen/,
    })
    await userEvent.click(ansehen)
    expect(await screen.findByRole('alert')).toBeInTheDocument()

    await userEvent.click(ansehen)
    const kasten = await screen.findByRole('region', {
      name: 'Aktualisierung von M15-Magic-Pack',
    })
    await userEvent.click(
      within(kasten).getByRole('button', { name: 'Auf c8e38c9 aktualisieren' }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    // **Nach einem gescheiterten Schreiben wird nicht nachgeladen** (§7, Regel 3).
    expect(belegung.PaketBezuege).toHaveBeenCalledTimes(1)
  })

  it('verwirft die Vorschau auf Wunsch', async () => {
    await aufschlagen(
      belegt({ PaketBezugVorschau: () => Promise.resolve(vorschau()) }),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: /Aktualisierung ansehen/ }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: t.verwerfen }),
    )

    expect(
      screen.queryByRole('region', {
        name: 'Aktualisierung von M15-Magic-Pack',
      }),
    ).not.toBeInTheDocument()
  })
})

describe('Der Rückweg', () => {
  it('tauscht mit dem aufgehobenen Stand und lädt nach', async () => {
    const zurueck = vi.fn(() =>
      Promise.resolve({
        ort: '/x',
        stand: bezogen,
        pruefsumme: '',
        uebergangen: [],
      }),
    )
    const bezuege = vi.fn(() =>
      Promise.resolve({
        bezuege: [bezug({ vorige: true, vorigerStand: spitze })],
      }),
    )
    await aufschlagen({ PaketBezuege: bezuege, PaketBezugZurueck: zurueck })

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${t.zurueck} (M15-Magic-Pack, c8e38c9)`,
      }),
    )

    expect(zurueck).toHaveBeenCalledWith('M15-Magic-Pack')
    await vi.waitFor(() => {
      expect(bezuege).toHaveBeenCalledTimes(2)
    })
  })

  it('meldet einen gescheiterten Rückweg am Baum', async () => {
    const bezuege = vi.fn(() =>
      Promise.resolve({
        bezuege: [bezug({ vorige: true, vorigerStand: spitze })],
      }),
    )
    await aufschlagen({
      PaketBezuege: bezuege,
      PaketBezugZurueck: () => Promise.reject(new Error('weg')),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: /zurück \(M15-Magic-Pack/ }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(bezuege).toHaveBeenCalledTimes(1)
  })
})
