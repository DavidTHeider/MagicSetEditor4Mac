// Die Ausfuhr über eine Exportvorlage, von der Setliste aus.
//
// # Was #209 verlangt
//
// `Vorlagenausfuehrer` stand seit #51 und wurde außerhalb von Tests **nirgends**
// gerufen — samt seiner Validierungs-, Options- und Skriptfehlerzweige. Diese
// Datei prüft die Naht dazwischen
// (#209).
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { SetVorlagenausfuhr } from './SetVorlagenausfuhr'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function probeset() {
  return {
    id: '7',
    name: 'Mein Probeset',
    spiel: 'probe',
    stil: 'probe-stil',
    herkunft: '/pfad/probe.mse-set',
    karten: 2,
    herkunftFehlt: false,
  }
}

// **Sie lädt, sobald sie dasteht** — seit UXR-04 (#1912) öffnet sie das
// Menü der Setkarte, und ein zweiter Knopf davor wäre ein Schritt ohne Frage.
// `schliessen` ist der Weg zurück, den die Karte ihr gibt.
function zeige(schliessen: () => void = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <SetVorlagenausfuhr satz={probeset()} schliessen={schliessen} />
    </Sprachraum>,
  )
}

function vorlage(teile: Record<string, unknown> = {}) {
  return {
    paket: 'probe-ausfuhr.mse-export-template',
    name: 'Ausfuhr der Probe',
    dateityp: '*.txt|*.txt',
    optionen: [],
    ...teile,
  }
}

describe('Die Ausfuhr über eine Exportvorlage', () => {
  it('bietet die Vorlagen des Sets an', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({ vorlagen: [vorlage()], befunde: [] }),
    })

    zeige()

    expect(
      await screen.findByRole('combobox', {
        name: texte.seiten.sets.vorlageWaehlen,
      }),
    ).toBeInTheDocument()

    expect(
      screen.getByRole('option', { name: 'Ausfuhr der Probe' }),
    ).toBeInTheDocument()
  })

  // **Die Wahlmöglichkeiten der Vorlage erscheinen als Eingaben** — der zweite
  // Haken aus #209.
  it('zeigt die Wahlmöglichkeiten als Eingaben', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({
          vorlagen: [
            vorlage({
              optionen: [
                {
                  art: 'boolean',
                  name: 'mit namen',
                  beschreibung: 'Namen mitschreiben',
                  vorgabe: 'yes',
                  auswahl: [],
                  wert: 'yes',
                },
                {
                  art: 'choice',
                  name: 'forum',
                  beschreibung: 'Forum',
                  vorgabe: 'mse',
                  auswahl: ['mse', 'salvation'],
                  wert: 'mse',
                },
                {
                  art: 'text',
                  name: 'kuerzel',
                  beschreibung: 'Kürzel',
                  vorgabe: 'PRB',
                  auswahl: [],
                  wert: 'PRB',
                },
              ],
            }),
          ],
          befunde: [],
        }),
    })

    zeige()

    expect(
      await screen.findByRole('checkbox', { name: 'Namen mitschreiben' }),
    ).toBeChecked()
    expect(screen.getByRole('combobox', { name: 'Forum' })).toHaveValue('mse')
    expect(screen.getByRole('textbox', { name: 'Kürzel' })).toHaveValue('PRB')
  })
})

// **Eine Option ohne Beschreibung zeigt ihren Namen.**
//
// Sie kommt so aus dem Paket: von den Optionen des bezogenen Packs führt
// nicht jede eine `description:`. Ein Feld ohne Aufschrift wäre für jemanden
// mit Bildschirmleser gar nicht zu deuten (§6).
describe('Eine Option ohne Beschreibung', () => {
  it('zeigt ihren Namen', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({
          vorlagen: [
            vorlage({
              optionen: [
                {
                  art: 'boolean',
                  name: 'ohne text',
                  beschreibung: '',
                  vorgabe: 'no',
                  auswahl: [],
                  wert: 'no',
                },
                {
                  art: 'choice',
                  name: 'die wahl',
                  beschreibung: '',
                  vorgabe: 'a',
                  auswahl: ['a'],
                  wert: 'a',
                },
                {
                  art: 'text',
                  name: 'das feld',
                  beschreibung: '',
                  vorgabe: '',
                  auswahl: [],
                  wert: '',
                },
              ],
            }),
          ],
          befunde: [],
        }),
    })

    zeige()

    expect(
      await screen.findByRole('checkbox', { name: 'ohne text' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('combobox', { name: 'die wahl' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('textbox', { name: 'das feld' }),
    ).toBeInTheDocument()
  })
})
