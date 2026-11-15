// **Sichern und Zurückspielen mit Kennwort, von der Seite aus** — OPS-06.AK-02
// (#1723).
//
// Über die Einstellungsseite und nicht über die Haken: geprüft wird der Weg,
// den jemand geht — Kennwort zweimal, sichern; eine verschlüsselte Sicherung
// wählen, ihr Kennwort geben, sehen, bestätigen.
import type { Sicherungsauswahl } from './api'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { zuGebiet } from './test/gebiete'
import { kataloge } from './sprache/kataloge'
import {
  antwortAbwarten,
  belegeApi,
  belegeFenster,
  raeumeApi,
} from './test/api'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.bestandspflege
const kennwort = 'korrekt pferd batterie heftklammer'

const verschluesselte: Sicherungsauswahl = {
  pfad: '/Users/probe/geheim.msesicherung',
  angelegtAm: '2026-09-26T09:12:33.000Z',
  schemastand: '0090_messung.sql',
  sets: 2,
  karten: 40,
  bilder: 'ersetzt',
  abgebrochen: false,
  verschluesselt: true,
  paketeEnthalten: ['eigen.mse-game'],
  paketeFehlen: ['magic-m15.mse-style'],
}

const fehltKennwort = {
  ...verschluesselte,
  sets: 0,
  karten: 0,
  fehler: {
    code: 'validierung' as const,
    schluessel: 'sicherungskennwortFehlt',
    meldung: 'Diese Sicherung ist verschlüsselt — sie braucht ihr Kennwort.',
    feld: 'kennwort',
  },
}

async function zumBestand() {
  const nutzer = userEvent.setup()

  render(<App />)

  await nutzer.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.einstellungen.titel,
    }),
  )
  await zuGebiet(kataloge.de.seiten.einstellungen.gruppen.daten)
  await screen.findByRole('heading', { name: worte.titel })

  return nutzer
}

describe('Eine Sicherung mit Kennwort', () => {
  it('geht mit dem zweimal gleichen Kennwort hinaus', async () => {
    const BestandSichern = vi.fn(() =>
      Promise.resolve({ pfad: '/tmp/s.msesicherung' }),
    )

    belegeApi({ ...belegeFenster('de'), BestandSichern })

    const nutzer = await zumBestand()
    const knopf = screen.getByRole('button', { name: worte.sichern })

    await nutzer.type(screen.getByLabelText(worte.kennwort), kennwort)
    await nutzer.type(
      screen.getByLabelText(worte.kennwortWiederholen),
      'anders',
    )

    // **Verschieden hält an** — und sagt es am Feld.
    expect(screen.getByText(worte.kennwortVerschieden)).toBeInTheDocument()
    expect(knopf).toBeDisabled()

    await nutzer.clear(screen.getByLabelText(worte.kennwortWiederholen))
    await nutzer.type(
      screen.getByLabelText(worte.kennwortWiederholen),
      kennwort,
    )

    expect(screen.queryByText(worte.kennwortVerschieden)).toBeNull()

    await nutzer.click(knopf)

    await waitFor(() => {
      expect(BestandSichern).toHaveBeenCalledWith(kennwort, false)
    })
  })

  it('verlangt bei einem verschlüsselten Bestand ein ausdrückliches Ohne', async () => {
    const BestandSichern = vi.fn(() => Promise.resolve({ pfad: '' }))

    belegeApi({
      ...belegeFenster('de'),
      Verschluesselungsstand: () =>
        Promise.resolve({ verschluesselt: true, fehler: null }),
      BestandSichern,
    })

    const nutzer = await zumBestand()
    const knopf = screen.getByRole('button', { name: worte.sichern })

    const ohne = await screen.findByRole('checkbox', {
      name: new RegExp(worte.klartext),
    })

    expect(knopf).toBeDisabled()

    await nutzer.click(ohne)
    await nutzer.click(knopf)

    await waitFor(() => {
      expect(BestandSichern).toHaveBeenCalledWith('', true)
    })
    await antwortAbwarten()

    // **Mit Kennwort ist das Ohne keine Frage mehr.**
    await nutzer.type(screen.getByLabelText(worte.kennwort), 'x')

    expect(
      screen.queryByRole('checkbox', { name: new RegExp(worte.klartext) }),
    ).toBeNull()
  })
})

describe('Eine verschlüsselte Sicherung zurückspielen', () => {
  it('fragt nach dem Kennwort und benutzt es bis zum Tausch', async () => {
    const SicherungBeschauen = vi.fn((_pfad: string, eingabe: string) =>
      Promise.resolve(
        eingabe === kennwort
          ? verschluesselte
          : {
              ...fehltKennwort,
              fehler: {
                ...fehltKennwort.fehler,
                schluessel: 'sicherungskennwortFalsch',
              },
            },
      ),
    )
    const Ruecksprunglage = vi.fn(() =>
      Promise.resolve({
        datenstand: '',
        zeilen: [],
        unklar: [],
        handpruefungNoetig: false,
        fehler: null,
      }),
    )
    const SicherungZurueckspielen = vi.fn(() =>
      Promise.resolve({
        ersetzt: true,
        beiseite: '/Users/probe/vorher.db',
        unveraendert: false,
      }),
    )

    belegeApi({
      ...belegeFenster('de'),
      SicherungWaehlen: () => Promise.resolve(fehltKennwort),
      SicherungBeschauen,
      Ruecksprunglage,
      SicherungZurueckspielen,
    })

    const nutzer = await zumBestand()

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))

    const feld = await screen.findByLabelText(worte.gesperrtKennwort)

    // **Ein falsches Kennwort steht am Feld** — und das Feld bleibt.
    await nutzer.type(feld, 'falsch{Enter}')

    expect(
      await screen.findByText(kataloge.de.fehler.sicherungskennwortFalsch),
    ).toBeInTheDocument()

    await nutzer.clear(feld)
    await nutzer.type(feld, kennwort)
    await nutzer.click(screen.getByRole('button', { name: worte.entsperren }))

    await screen.findByText(worte.rueckfrage)

    expect(SicherungBeschauen).toHaveBeenLastCalledWith(
      verschluesselte.pfad,
      kennwort,
    )
    expect(screen.getByText(worte.ja)).toBeInTheDocument()

    // **Was fehlt, steht vor dem Bestätigen da** (OPS-06.AK-03).
    expect(screen.getByText('eigen.mse-game')).toBeInTheDocument()
    expect(screen.getByText('magic-m15.mse-style')).toBeInTheDocument()

    await waitFor(() => {
      expect(Ruecksprunglage).toHaveBeenCalledWith(
        verschluesselte.pfad,
        kennwort,
      )
    })
    await antwortAbwarten()

    await nutzer.click(screen.getByRole('button', { name: worte.bestaetigen }))

    await waitFor(() => {
      expect(SicherungZurueckspielen).toHaveBeenCalledWith(
        verschluesselte.pfad,
        kennwort,
      )
    })
  })

  it('lässt sich ohne Kennwort wieder schließen', async () => {
    belegeApi({
      ...belegeFenster('de'),
      SicherungWaehlen: () => Promise.resolve(fehltKennwort),
    })

    const nutzer = await zumBestand()

    await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))
    await screen.findByLabelText(worte.gesperrtKennwort)

    // **Ohne Eingabe gibt es nichts zu öffnen.**
    expect(
      screen.getByRole('button', { name: worte.entsperren }),
    ).toBeDisabled()

    await nutzer.click(screen.getByRole('button', { name: worte.abbrechen }))

    expect(screen.queryByLabelText(worte.gesperrtKennwort)).toBeNull()
  })
})
