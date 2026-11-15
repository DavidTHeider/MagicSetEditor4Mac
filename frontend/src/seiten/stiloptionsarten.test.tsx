// **Die Bedienelemente der Stiloptionen** — je Art eines
// (#630).
//
// **Eine zweite Datei**, weil `stiloptionen.test.tsx` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort steht, wie die Liste geladen
// und geschrieben wird, hier, was eine einzelne Art anzeigt.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Stiloptionen } from './Stiloptionen'

afterEach(() => {
  raeumeApi()
})

// paketwahl gibt eine Stiloption der Art `package choice`
// (#632).
//
// **Wert und Beschriftung sind hier nicht dasselbe**: gespeichert wird der
// Paketdateiname, dasteht der Kurzname.
function paketwahl() {
  return {
    optionen: [
      {
        name: 'overlay',
        art: 'package choice',
        beschreibung: 'Ob eine Auflage über die Karte gelegt wird.',
        wert: '',
        moeglichkeiten: [
          { wert: '', beschriftung: 'Keines' },
          { wert: 'probe-overlay-folie.mse-include', beschriftung: 'Folie' },
        ],
        eigeneFarbe: false,
        gesetzt: false,
        hatVorgabe: false,
        hatWahlbilder: false,
        vorgabe: '',
      },
    ],
  }
}

describe('Eine Paketwahl', () => {
  it('zeigt die Kurznamen und speichert die Dateinamen', async () => {
    const StiloptionSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeApi({
      Stiloptionen: () => Promise.resolve(paketwahl()),
      StiloptionSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Stiloptionen setID="1" geaendert={vi.fn()} />
      </Sprachraum>,
    )

    const liste = await screen.findByRole('combobox', { name: /overlay/ })

    // **Der leere Eintrag steht vorn und heißt, wie das Paket ihn nennt.**
    expect((liste as HTMLSelectElement).options[0]?.textContent).toBe('Keines')

    await userEvent.selectOptions(liste, 'probe-overlay-folie.mse-include')

    expect(StiloptionSetzen).toHaveBeenCalledWith(
      '1',
      'overlay',
      'probe-overlay-folie.mse-include',
    )
  })
})

describe('Eine Farbstiloption', () => {
  // farboption gibt eine Farbe mit Wahlen **und** eigener Farbe — so sieht
  // die Mehrheit aus: 81 der 105 im Pack
  // (#630).
  function farboption(wert = 'rgb(0,0,0)', moeglichkeiten = true) {
    return {
      optionen: [
        {
          name: 'randfarbe',
          art: 'color',
          beschreibung: 'Die Farbe des Kartenrands.',
          wert,
          moeglichkeiten: moeglichkeiten
            ? [
                { wert: 'rgb(0,0,0)', beschriftung: 'schwarz' },
                { wert: 'rgb(255,255,255)', beschriftung: 'weiß' },
              ]
            : [],
          eigeneFarbe: true,
          gesetzt: false,
          hatVorgabe: false,
          hatWahlbilder: false,
          vorgabe: '',
        },
      ],
    }
  }

  function zeigeFarbe(antwort: ReturnType<typeof farboption>) {
    const StiloptionSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeApi({
      Stiloptionen: () => Promise.resolve(antwort),
      StiloptionSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Stiloptionen setID="1" geaendert={vi.fn()} />
      </Sprachraum>,
    )

    return StiloptionSetzen
  }

  it('zeigt die benannten Farben und speichert die Farbe', async () => {
    const StiloptionSetzen = zeigeFarbe(farboption())

    const liste = await screen.findByRole('combobox', { name: /randfarbe/ })

    await userEvent.selectOptions(liste, 'rgb(255,255,255)')

    expect(StiloptionSetzen).toHaveBeenCalledWith(
      '1',
      'randfarbe',
      'rgb(255,255,255)',
    )
  })

  // **Der Farbwähler steht daneben**, weil `allow custom` ohne Angabe wahr
  // ist — über das ganze Pack ist er nirgends abgeschaltet.
  it('schreibt eine eigene Farbe in MSEs Form', async () => {
    const StiloptionSetzen = zeigeFarbe(farboption())

    const waehler = await screen.findByLabelText(/eigene Farbe/)

    expect(waehler).toHaveValue('#000000')

    fireEvent.input(waehler, { target: { value: '#0c2238' } })

    await waitFor(() => {
      expect(StiloptionSetzen).toHaveBeenCalledWith(
        '1',
        'randfarbe',
        'rgb(12,34,56)',
      )
    })
  })

  // **Ohne Wahlen bleibt nur der Farbwähler** — 16 der 105 im Pack nennen
  // keine, und ein leeres Auswahlfeld wäre eine Liste ohne Inhalt.
  it('zeigt ohne Wahlen keine Liste', async () => {
    zeigeFarbe(farboption('rgb(12,34,56)', false))

    expect(await screen.findByLabelText(/eigene Farbe/)).toHaveValue('#0c2238')

    expect(
      screen.queryByRole('combobox', { name: /randfarbe/ }),
    ).not.toBeInTheDocument()
  })

  // **Und ein unlesbarer Wert zeigt Schwarz, ohne ihn zu überschreiben.**
  // Ein wx-Farbname wie `papayawhip` kommt so an; ihn zu ersetzen hieße, eine
  // Angabe der Vorlage stillschweigend zu verlieren.
  it('zeigt bei einem unlesbaren Wert Schwarz und schreibt nichts', async () => {
    const StiloptionSetzen = zeigeFarbe(farboption('papayawhip', false))

    expect(await screen.findByLabelText(/eigene Farbe/)).toHaveValue('#000000')

    expect(StiloptionSetzen).not.toHaveBeenCalled()
  })
})

describe('Eine Mehrfachwahl', () => {
  // mehrfachwahl gibt eine Option mit drei Wahlen, zwei davon gesetzt.
  function mehrfachwahl(wert = 'nyx, fahrzeug') {
    return {
      optionen: [
        {
          name: 'rahmen',
          art: 'multiple choice',
          beschreibung: 'Welche Rahmen benutzt werden.',
          wert,
          moeglichkeiten: [
            { wert: 'nyx', beschriftung: 'nyx' },
            { wert: 'schnee', beschriftung: 'schnee' },
            { wert: 'fahrzeug', beschriftung: 'fahrzeug' },
          ],
          eigeneFarbe: false,
          gesetzt: false,
          hatVorgabe: false,
          hatWahlbilder: false,
          vorgabe: '',
        },
      ],
    }
  }

  function zeigeMehrfach(antwort: ReturnType<typeof mehrfachwahl>) {
    const StiloptionSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeApi({
      Stiloptionen: () => Promise.resolve(antwort),
      StiloptionSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Stiloptionen setID="1" geaendert={vi.fn()} />
      </Sprachraum>,
    )

    return StiloptionSetzen
  }

  it('hakt an, was im Wert steht', async () => {
    zeigeMehrfach(mehrfachwahl())

    expect(await screen.findByRole('checkbox', { name: 'nyx' })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'fahrzeug' })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'schnee' })).not.toBeChecked()
  })

  // **Geschickt wird die Reihenfolge der Erklärung**, nicht die des
  // Anklickens — der Dienst ordnet sie ohnehin, und eine Seite, die etwas
  // anderes schickt, als sie zeigt, ist schwerer zu lesen.
  it('nimmt eine Wahl dazu', async () => {
    const StiloptionSetzen = zeigeMehrfach(mehrfachwahl())

    await userEvent.click(
      await screen.findByRole('checkbox', { name: 'schnee' }),
    )

    expect(StiloptionSetzen).toHaveBeenCalledWith(
      '1',
      'rahmen',
      'nyx, schnee, fahrzeug',
    )
  })

  it('nimmt eine Wahl weg', async () => {
    const StiloptionSetzen = zeigeMehrfach(mehrfachwahl())

    await userEvent.click(await screen.findByRole('checkbox', { name: 'nyx' }))

    expect(StiloptionSetzen).toHaveBeenCalledWith('1', 'rahmen', 'fahrzeug')
  })

  // **Die letzte wegzunehmen schickt nichts** — was daraus wird, entscheidet
  // `empty choice:` im Paket, und das weiß nur der Dienst.
  it('schickt die leere Wahl, wenn nichts bleibt', async () => {
    const StiloptionSetzen = zeigeMehrfach(mehrfachwahl('schnee'))

    await userEvent.click(
      await screen.findByRole('checkbox', { name: 'schnee' }),
    )

    expect(StiloptionSetzen).toHaveBeenCalledWith('1', 'rahmen', '')
  })
})

describe('Eine Auskunft', () => {
  // **Sie trägt keinen Wert und ist nicht bearbeitbar** — `InfoField()
  // { editable = false; }` (`src/data/field/information.hpp:26–28`, Fassung
  // f9d9356 vom 2021-07-06). Über das Pack sind alle 13 Vorkommen
  // Zwischenüberschriften in langen Optionslisten
  // (#639).
  function auskunft(beschreibung = 'Was darunter steht, gehört zusammen.') {
    return {
      optionen: [
        {
          name: 'Weitere Einstellungen',
          art: 'info',
          beschreibung,
          wert: '',
          moeglichkeiten: [],
          eigeneFarbe: false,
          gesetzt: false,
          hatVorgabe: false,
          hatWahlbilder: false,
          vorgabe: '',
        },
      ],
    }
  }

  function zeigeAuskunft(antwort: ReturnType<typeof auskunft>) {
    belegeApi({
      Stiloptionen: () => Promise.resolve(antwort),
      StiloptionSetzen: vi.fn(() => Promise.resolve({ fehler: null })),
    })

    render(
      <Sprachraum sprache="de">
        <Stiloptionen setID="1" geaendert={vi.fn()} />
      </Sprachraum>,
    )
  }

  it('steht als Überschrift da, nicht als Feld', async () => {
    zeigeAuskunft(auskunft())

    expect(
      await screen.findByRole('heading', { name: 'Weitere Einstellungen' }),
    ).toBeInTheDocument()

    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    expect(screen.queryByRole('combobox')).not.toBeInTheDocument()
    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()
  })

  it('zeigt ihre Erklärung darunter', async () => {
    zeigeAuskunft(auskunft())

    expect(
      await screen.findByText('Was darunter steht, gehört zusammen.'),
    ).toBeInTheDocument()
  })

  // **Sechs der 13 tragen keine** — dann steht nur die Überschrift da.
  it('kommt ohne Erklärung aus', async () => {
    zeigeAuskunft(auskunft(''))

    expect(
      await screen.findByRole('heading', { name: 'Weitere Einstellungen' }),
    ).toBeInTheDocument()
  })
})
