// Die Optionen einer Exportvorlage werden behalten —
// #492.
//
// Über `full-magic-pack` führen **21 von 22** Exportvorlagen zusammen **96**
// Optionsfelder (44 boolean, 24 choice, 23 text, 4 color, 1 multiple; gezählt
// am 2026-08-19). Angeboten wurden sie schon; **behalten** wurden sie nicht —
// sie lagen im Zustand der Seite und waren beim nächsten Öffnen fort.
//
//	IndexMap<FieldP,ValueP>& exportOptionsFor(const ExportTemplate& export_template);
//
// `src/data/settings.hpp`, Zeilen 180–186, Fassung f9d9356 vom 2021-07-06 —
// **je Vorlage**, nicht je Set.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { SetVorlagenausfuhr } from './SetVorlagenausfuhr'

afterEach(() => {
  raeumeApi()
})

const vorlage = {
  paket: 'magic-forum.mse-export-template',
  name: 'Forum',
  dateityp: '*.txt|*.txt',
  optionen: [
    {
      art: 'text',
      name: 'creator',
      // **Die Beschriftung ist die Beschreibung, sonst der Name** — sie kommt
      // aus dem Paket und wird nicht übersetzt (§9).
      beschreibung: 'The creator of the set.',
      vorgabe: '',

      // **Der behaltene Wert** — er kommt mit der Vorlage.
      wert: 'Anna',
      auswahl: [],
    },
  ],
}

const probeset = {
  id: '7',
  name: 'Mein Probeset',
  spiel: 'probe',
  stil: 'probe-stil',
  herkunft: '/pfad/probe.mse-set',
  karten: 2,
  herkunftFehlt: false,
}

// zeige belegt die Fassade und schlägt die Ausfuhr auf.
//
// **Geladen wird beim Öffnen, nicht beim Bau der Seite** — die Setliste führte
// sonst je Zeile einen Paketlauf über die Platte. Geöffnet wird sie seit
// UXR-04 (#1912) über das Menü der Setkarte, und dann lädt sie gleich.
function zeige(belegung: Record<string, unknown> = {}) {
  belegeApi({
    // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Exportvorlagen: () => Promise.resolve({ vorlagen: [vorlage], befunde: [] }),
    ...belegung,
  })

  render(
    <Sprachraum sprache="de">
      <SetVorlagenausfuhr satz={probeset} schliessen={vi.fn()} />
    </Sprachraum>,
  )
}

describe('Die Optionen einer Exportvorlage', () => {
  // **Der behaltene Wert steht da, nicht die Vorgabe.**
  it('zeigt den behaltenen Wert', async () => {
    zeige()

    expect(await screen.findByLabelText(/The creator of the set/)).toHaveValue(
      'Anna',
    )
  })

  // **Und eine Änderung wird behalten — je Vorlage.**
  it('schreibt eine Änderung mit dem Paketnamen', async () => {
    const ExportoptionSetzen = vi.fn(() => Promise.resolve({}))

    zeige({ ExportoptionSetzen })

    const feld = await screen.findByLabelText(/The creator of the set/)
    await userEvent.clear(feld)
    await userEvent.type(feld, 'B')

    expect(ExportoptionSetzen).toHaveBeenCalledWith(
      'magic-forum.mse-export-template',
      'creator',
      'B',
    )
  })

  // **Ein gescheitertes Behalten hält den Lauf nicht an — und wird gemeldet.**
  //
  // Der Wert steht im Zustand der Seite und geht mit; wer deswegen die Ausgabe
  // abbräche, nähme jemandem sein Ergebnis wegen einer Bequemlichkeit.
  // **Verschweigen darf man ihn trotzdem nicht** (§8): die Oberfläche
  // behauptet sonst, die Option werde behalten, und beim nächsten Öffnen steht
  // der alte Wert da.
  //
  // **Der Fehler steckt im Umschlag, nicht in einer Ablehnung.** Der erste
  // Anlauf hängte ein `.catch()` an den Aufruf — der Zweitblick über Phase G
  // hat gefunden, dass die Bindung sich **erfolgreich** auflöst und den Fehler
  // *in* der Antwort trägt. Ein `.catch()` daran fängt nie etwas, und dieser
  // Test hielt genau das fest: er verlangte, dass **keine** Meldung erscheint.
  it('meldet ein gescheitertes Behalten und behält den Wert im Feld', async () => {
    zeige({
      ExportoptionSetzen: () =>
        Promise.resolve({
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'fehler.unbekannt',
            meldung: 'Die Platte ist voll.',
          },
        }),
    })

    const feld = await screen.findByLabelText(/The creator of the set/)
    await userEvent.clear(feld)
    await userEvent.type(feld, 'B')

    // **Der Wert bleibt** — die laufende Ausgabe nimmt ihn mit.
    expect(feld).toHaveValue('B')

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Die Platte ist voll.',
    )
  })

  // **Und eine Ablehnung ebenso** — der Weg, den es vorher zu fangen glaubte.
  it('meldet auch eine abgelehnte Zusage', async () => {
    zeige({
      ExportoptionSetzen: () => Promise.reject(new Error('weg')),
    })

    const feld = await screen.findByLabelText(/The creator of the set/)
    await userEvent.clear(feld)
    await userEvent.type(feld, 'B')

    expect(feld).toHaveValue('B')
    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})

// **Eine überholte Meldung gehört nicht zum heutigen Stand** (§7, Regel 2).
//
// Der Nachlauf über Phase G hat gefunden, dass die Korrektur von oben ohne
// Standwächter arbeitet: `lauf.naechster()` wird gerufen, die Nummer aber nicht
// behalten. Ein Fehler aus einem **überholten** `ExportoptionSetzen` trägt sich
// danach ein — nach einer neueren, geglückten Wahl.
//
// Das ist die unauffälligere Hälfte der Regel: die Oberfläche behauptet, das
// Behalten sei gescheitert, während der zuletzt gewählte Wert längst im
// Bestand steht.
describe('Ein überholtes Behalten', () => {
  it('meldet sich nicht mehr', async () => {
    const antworten: ((wert: unknown) => void)[] = []

    zeige({
      ExportoptionSetzen: () =>
        new Promise((aufloesen) => {
          antworten.push(aufloesen)
        }),
    })

    const feld = await screen.findByLabelText(/The creator of the set/)
    await userEvent.clear(feld)
    await userEvent.type(feld, 'B')

    await waitFor(() => {
      expect(antworten.length).toBeGreaterThan(1)
    })

    // **Der letzte gewinnt.** Die früheren sind überholt — ihre Antwort gehört
    // zu einem Stand, den niemand mehr sieht.
    const letzter = antworten.length - 1

    antworten[letzter]?.({})

    for (let i = 0; i < letzter; i += 1) {
      antworten[i]?.({
        fehler: {
          code: 'unbekannt' as const,
          schluessel: 'fehler.unbekannt',
          meldung: 'Die Platte ist voll.',
        },
      })
    }

    await waitFor(() => {
      expect(feld).toHaveValue('B')
    })

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
