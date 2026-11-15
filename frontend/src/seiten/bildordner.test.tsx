// **Der zuletzt gewählte Ordner der Bildausgabe** — AP-10
// (#2367),
// **ported** aus MSE2: `settings.default_export_dir` wird nach jeder Wahl
// fortgeschrieben (`src/gui/images_export_window.cpp`, Zeilen 63–66, Fassung
// f9d9356 vom 2021-07-06) — **einer je Installation**, nicht je Set.
//
// Die Seite schlägt ihn vor, und der Lauf geht wirklich dorthin: ein
// vorbelegter Ordner, den das Ausgeben nicht nimmt, wäre eine Anzeige ohne
// Wirkung.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { Bilder } from './Bilder'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.bilder

// einstellungen gibt eine Antwort mit dem gemerkten Ordner.
function einstellungen(gemerkt: string) {
  return () =>
    Promise.resolve({
      einstellungen: [
        {
          schluessel: 'bildausfuhr.ordner',
          wert: gemerkt,
          wirksam: gemerkt,
          werte: [],
          bereich: 'bildausfuhr',
        },
      ],
    })
}

function zeige(gemerkt: string, belegung: Record<string, unknown> = {}) {
  belegeApi({
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Einstellungen: einstellungen(gemerkt),
    EinstellungSetzen: () => Promise.resolve({}),
    Bildnamen: () =>
      Promise.resolve({
        namen: [{ karteID: '1', nummer: 1, datei: '001-Karte.png' }],
      }),
    BildSchreiben: () =>
      Promise.resolve({
        pfad: '/x/001-Karte.png',
        breite: 10,
        hoehe: 14,
        befunde: [],
        uebersprungen: false,
      }),
    ...belegung,
  })

  render(
    <Sprachraum sprache="de">
      <Bilder setID="4" />
    </Sprachraum>,
  )
}

describe('Der zuletzt gewählte Bildordner', () => {
  it('steht vorbelegt da und sagt, dass er gemerkt ist', async () => {
    zeige('/Bilder/zuletzt')

    expect(await screen.findByText('/Bilder/zuletzt')).toBeInTheDocument()
    expect(screen.getByText(texte.gemerkt)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: texte.ausgeben })).toBeEnabled()
  })

  it('ist das Ziel des Laufs, ohne dass jemand wählt', async () => {
    const BildSchreiben = vi.fn(() =>
      Promise.resolve({
        pfad: '/Bilder/zuletzt/001-Karte.png',
        breite: 10,
        hoehe: 14,
        befunde: [],
        uebersprungen: false,
      }),
    )

    zeige('/Bilder/zuletzt', { BildSchreiben })

    await screen.findByText('/Bilder/zuletzt')
    await userEvent.click(screen.getByRole('button', { name: texte.ausgeben }))

    await waitFor(() => {
      expect(BildSchreiben).toHaveBeenCalledWith(
        '1',
        '/Bilder/zuletzt',
        '001-Karte.png',
        'png',
        expect.any(Number),
        expect.any(String),
        expect.any(String),
      )
    })
  })

  it('merkt sich einen neu gewählten Ordner', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({}))

    zeige('/Bilder/zuletzt', {
      EinstellungSetzen,
      Bildordner: () => Promise.resolve({ ordner: '/Bilder/neu' }),
    })

    await screen.findByText('/Bilder/zuletzt')
    await userEvent.click(
      screen.getByRole('button', { name: texte.ordnerWaehlen }),
    )

    expect(await screen.findByText('/Bilder/neu')).toBeInTheDocument()
    expect(EinstellungSetzen).toHaveBeenCalledWith(
      'bildausfuhr.ordner',
      '/Bilder/neu',
    )

    // **Frisch gewählt ist nicht „zuletzt gewählt".**
    expect(screen.queryByText(texte.gemerkt)).not.toBeInTheDocument()
  })

  // **Ein abgebrochener Dialog schreibt nichts** — belegt an einem zweiten
  // Dialog durch denselben Weg, der schreibt: genau ein Ruf, mit seinem Ordner.
  it('merkt nach einem Abbruch nichts', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({}))
    const Bildordner = vi
      .fn()
      .mockResolvedValueOnce({ abgebrochen: true })
      .mockResolvedValueOnce({ ordner: '/Bilder/zweiter' })

    zeige('/Bilder/zuletzt', { EinstellungSetzen, Bildordner })

    const waehlen = await screen.findByRole('button', {
      name: texte.ordnerWaehlen,
    })
    await userEvent.click(waehlen)
    await userEvent.click(waehlen)

    expect(await screen.findByText('/Bilder/zweiter')).toBeInTheDocument()
    expect(EinstellungSetzen).toHaveBeenCalledTimes(1)
    expect(EinstellungSetzen).toHaveBeenCalledWith(
      'bildausfuhr.ordner',
      '/Bilder/zweiter',
    )
  })

  // **Geschrieben ist geschrieben** (§7, Regel 3) gilt auch umgekehrt: ein
  // Merken, das scheitert, nimmt dem Lauf den Ordner nicht — es sagt aber,
  // dass er beim nächsten Mal nicht dasteht.
  it('behält den Ordner für den Lauf, wenn das Merken scheitert', async () => {
    zeige('', {
      EinstellungSetzen: () =>
        Promise.resolve({
          fehler: { code: 'validierung', meldung: 'Der Bestand klemmt.' },
        }),
      Bildordner: () => Promise.resolve({ ordner: '/Bilder/neu' }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.ordnerWaehlen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      texte.nichtGemerkt,
    )
    expect(screen.getByText('/Bilder/neu')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: texte.ausgeben })).toBeEnabled()
  })

  it('bleibt ohne gemerkten Ordner leer und hält das Ausgeben an', async () => {
    const Einstellungen = vi.fn(einstellungen(''))
    zeige('', { Einstellungen })

    // **Erst geladen, dann verneint** (§3): leer ist es auch davor.
    await waitFor(() => {
      expect(Einstellungen).toHaveBeenCalled()
    })
    await effekteNachholen()

    expect(await screen.findByText(texte.keinOrdner)).toBeInTheDocument()
    expect(screen.queryByText(texte.gemerkt)).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: texte.ausgeben })).toBeDisabled()
  })
})
