// Was bei einem Namenskonflikt geschieht —
// #491.
//
// Die Bildausgabe schrieb eine Datei je Karte und **überschrieb**, wenn eine
// dalag. Das Muster `{card.name}` erzeugt für zwei Länder desselben Namens
// denselben Dateinamen; eines blieb übrig, und niemand erfuhr es.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Bilder } from './Bilder'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

const werte = ['altbehalten', 'ueberschreiben', 'nummerieren', 'nummerierenneu']

// einstellungen gibt eine Antwort mit der gegebenen Wahl.
function einstellungen(wahl = 'nummerierenneu') {
  return () =>
    Promise.resolve({
      einstellungen: [
        {
          schluessel: 'bildkonflikt',
          wert: wahl,
          wirksam: wahl,
          werte,
          bereich: 'bildausfuhr',
        },
      ],
    })
}

function zeige(belegung: Record<string, unknown> = {}) {
  belegeApi({
    // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Einstellungen: einstellungen(),
    Bildnamen: vi.fn(() =>
      Promise.resolve({
        namen: [{ karteID: '1', nummer: 1, datei: '001-Karte.png' }],
      }),
    ),
    ...belegung,
  })

  render(
    <Sprachraum sprache="de">
      <Bilder setID="4" />
    </Sprachraum>,
  )
}

describe('Die Wahl beim Namenskonflikt', () => {
  it('steht an der Bildausgabe, mit allen vier Werten', async () => {
    zeige()

    const wahl = await screen.findByRole('combobox', {
      name: texte.seiten.bilder.konflikt,
    })

    expect(wahl).toHaveDisplayValue(
      texte.seiten.bilder.konflikte.nummerierenneu,
    )

    for (const wert of werte) {
      expect(
        screen.getByRole('option', {
          name: texte.seiten.bilder.konflikte[
            wert as keyof typeof texte.seiten.bilder.konflikte
          ],
        }),
      ).toBeInTheDocument()
    }
  })

  // **Die Vorgabe ist nicht `ueberschreiben`** — der letzte Abnahmekasten aus
  // #491, und die Antwort ist nachgelesen: `images_export_conflicts` steht bei
  // MSE auf `CONFLICT_NUMBER_OVERWRITE` (`src/data/settings.cpp`, Zeile 74,
  // Fassung f9d9356 vom 2021-07-06).
  it('steht auf MSEs Vorgabe, nicht auf Überschreiben', async () => {
    zeige()

    expect(
      await screen.findByRole('combobox', {
        name: texte.seiten.bilder.konflikt,
      }),
    ).toHaveDisplayValue(texte.seiten.bilder.konflikte.nummerierenneu)
  })

  // **Die Wahl wird behalten** — sie geht in den Bestand, nicht in den
  // Zustand dieser Seite (#245).
  it('schreibt eine Änderung in den Bestand', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({}))

    zeige({ EinstellungSetzen })

    await userEvent.selectOptions(
      await screen.findByRole('combobox', {
        name: texte.seiten.bilder.konflikt,
      }),
      'altbehalten',
    )

    expect(EinstellungSetzen).toHaveBeenCalledWith(
      'bildkonflikt',
      'altbehalten',
    )
  })

  // **Und sie reist mit dem Lauf** — in beide Aufrufe. `Bildnamen` löst den
  // laufenden Durchgang auf, `BildSchreiben` die Platte; dieselbe Regel, zwei
  // Aufrufstellen.
  it('reist mit in beide Aufrufe', async () => {
    const Bildnamen = vi.fn(() =>
      Promise.resolve({
        namen: [{ karteID: '1', nummer: 1, datei: '001-Karte.png' }],
      }),
    )
    const BildSchreiben = vi.fn(() =>
      Promise.resolve({
        pfad: '/tmp/bilder/001-Karte.png',
        breite: 10,
        hoehe: 14,
        befunde: [],
        uebersprungen: false,
      }),
    )

    zeige({
      Einstellungen: einstellungen('altbehalten'),
      Bildnamen,
      BildSchreiben,
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () => Promise.resolve({ ordner: '/tmp/bilder' }),
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.bilder.ordnerWaehlen,
      }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.bilder.ausgeben }),
    )

    await waitFor(() => {
      expect(Bildnamen).toHaveBeenCalledWith(
        '4',
        expect.any(String),
        'png',
        'altbehalten',
        [],
      )
    })

    await waitFor(() => {
      expect(BildSchreiben).toHaveBeenCalledWith(
        '1',
        '/tmp/bilder',
        '001-Karte.png',
        'png',
        expect.any(Number),
        'altbehalten',

        // **Und die Güte** — Ω.4: sie reist mit jedem Aufruf, damit die
        // Wahl nicht an einer Stelle hängen bleibt.
        'medium',
      )
    })
  })

  // **Übersprungen wird gezählt und benannt** (#491).
  //
  // Wer `altbehalten` wählt und danach „312 Bilder geschrieben" liest, hat
  // keine Auskunft, sondern eine Zahl, die stimmt und nichts sagt.
  it('nennt die übersprungenen Karten', async () => {
    zeige({
      Einstellungen: einstellungen('altbehalten'),
      BildSchreiben: () =>
        Promise.resolve({
          pfad: '',
          breite: 0,
          hoehe: 0,
          befunde: [],
          uebersprungen: true,
        }),
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () => Promise.resolve({ ordner: '/tmp/bilder' }),
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.bilder.ordnerWaehlen,
      }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.bilder.ausgeben }),
    )

    expect(await screen.findByText(/1 Karten übersprungen/)).toBeInTheDocument()
    expect(screen.getByText('001-Karte.png')).toBeInTheDocument()
  })
})
