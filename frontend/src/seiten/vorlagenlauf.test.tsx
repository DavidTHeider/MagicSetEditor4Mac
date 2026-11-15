// Das Ausführen einer Exportvorlage.
//
// **Eine zweite Datei**, weil die erste sonst über die Zeilengrenze wüchse
// (AGENTS.md §2). Der Schnitt folgt der Frage: dort *welche Vorlagen gibt es
// und was lässt sich einstellen*, hier *was geschieht beim Ausführen*
// (#209).
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { abgebrochen } from '../test/laufauftrag'
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

describe('Das Ausführen einer Exportvorlage', () => {
  // **Das Ergebnis geht über den Dateidialog hinaus, mit `file type:` als
  // Filter** — der dritte Haken. Die Oberfläche zeigt den Filter nicht; sie
  // reicht ihn an die Bindung zurück, und die baut den Dialog daraus.
  it('reicht den Dateityp der Vorlage an die Bindung zurück', async () => {
    const VorlageAusfuehren = vi.fn(() =>
      Promise.resolve({
        pfad: '/pfad/ausfuhr.txt',
        befunde: [],
        abgebrochen: false,
      }),
    )

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({ vorlagen: [vorlage()], befunde: [] }),
      VorlageAusfuehren,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.vorlageAusfuehren,
      }),
    )

    expect(VorlageAusfuehren).toHaveBeenCalledWith(
      '7',
      'probe-ausfuhr.mse-export-template',
      '*.txt|*.txt',
      {},
      [],
    )

    expect(await screen.findByRole('status')).toHaveTextContent(
      `${texte.seiten.sets.vorlageGeschrieben} /pfad/ausfuhr.txt`,
    )
  })

  it('schickt die gewählten Werte mit', async () => {
    const VorlageAusfuehren = vi.fn(() =>
      Promise.resolve({ pfad: '/pfad/a.txt', befunde: [], abgebrochen: false }),
    )

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
              ],
            }),
          ],
          befunde: [],
        }),
      VorlageAusfuehren,
    })

    zeige()

    // **Zweimal geklickt: aus und wieder an.** Ein Kästchen, das sich nur
    // ausschalten lässt, ist eine halbe Zusage — und die andere Hälfte wäre
    // ungeprüft.
    const kasten = await screen.findByRole('checkbox', {
      name: 'Namen mitschreiben',
    })

    await userEvent.click(kasten)
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.sets.vorlageAusfuehren }),
    )

    expect(VorlageAusfuehren).toHaveBeenLastCalledWith(
      '7',
      'probe-ausfuhr.mse-export-template',
      '*.txt|*.txt',
      { 'mit namen': 'no' },
      [],
    )

    await userEvent.click(kasten)
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.sets.vorlageAusfuehren }),
    )

    expect(VorlageAusfuehren).toHaveBeenLastCalledWith(
      '7',
      'probe-ausfuhr.mse-export-template',
      '*.txt|*.txt',
      { 'mit namen': 'yes' },
      [],
    )
  })

  it('trägt den Erfolg eines alten Laufs nicht an einer neuen Vorlage ein', async () => {
    let antworten!: (antwort: {
      pfad: string
      befunde: string[]
      abgebrochen: boolean
    }) => void

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({
          vorlagen: [
            vorlage(),
            vorlage({
              paket: 'zweite.mse-export-template',
              name: 'Zweite Vorlage',
            }),
          ],
          befunde: [],
        }),
      VorlageAusfuehren: () =>
        new Promise((erfuellen) => {
          antworten = erfuellen
        }),
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.vorlageAusfuehren,
      }),
    )
    await userEvent.selectOptions(
      screen.getByRole('combobox', {
        name: texte.seiten.sets.vorlageWaehlen,
      }),
      'zweite.mse-export-template',
    )

    await act(async () => {
      antworten({ pfad: '/pfad/erste.txt', befunde: [], abgebrochen: false })
      await Promise.resolve()
    })

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('trägt den Fehler eines alten Laufs nicht an einer neuen Vorlage ein', async () => {
    let abweisen!: (ursache: unknown) => void

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({
          vorlagen: [
            vorlage(),
            vorlage({
              paket: 'zweite.mse-export-template',
              name: 'Zweite Vorlage',
            }),
          ],
          befunde: [],
        }),
      VorlageAusfuehren: () =>
        new Promise((_erfuellen, zurueckweisen) => {
          abweisen = zurueckweisen
        }),
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.vorlageAusfuehren,
      }),
    )
    await userEvent.selectOptions(
      screen.getByRole('combobox', {
        name: texte.seiten.sets.vorlageWaehlen,
      }),
      'zweite.mse-export-template',
    )

    await act(async () => {
      abweisen({
        code: 'validierung',
        schluessel: 'validierung',
        meldung: 'Die Eingabe ist nicht gültig.',
        feld: '',
      })
      await Promise.resolve()
    })

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('trägt eine Laufantwort nicht an geänderten Optionen ein', async () => {
    let antworten!: (antwort: {
      pfad: string
      befunde: string[]
      abgebrochen: boolean
    }) => void

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({
          vorlagen: [
            vorlage({
              optionen: [
                {
                  art: 'text',
                  name: 'kuerzel',
                  beschreibung: 'Kürzel',
                  vorgabe: 'ALT',
                  auswahl: [],
                  wert: 'ALT',
                },
              ],
            }),
          ],
          befunde: [],
        }),
      VorlageAusfuehren: () =>
        new Promise((erfuellen) => {
          antworten = erfuellen
        }),
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.vorlageAusfuehren,
      }),
    )

    const feld = screen.getByRole('textbox', { name: 'Kürzel' })
    await userEvent.clear(feld)
    await userEvent.type(feld, 'NEU')

    await act(async () => {
      antworten({ pfad: '/pfad/alt.txt', befunde: [], abgebrochen: false })
      await Promise.resolve()
    })

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })
})

// **Ein angehaltener Lauf ist keine Fehlermeldung** (FMT-03, #1692) —
// gewartet wird darauf, dass der Knopf nach dem Lauf wieder frei ist.
it('zeigt ein Anhalten nicht als Fehler', async () => {
  let antworten: (wert: unknown) => void = () => undefined
  belegeApi({
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Exportvorlagen: () =>
      Promise.resolve({ vorlagen: [vorlage()], befunde: [] }),
    VorlageAusfuehren: () =>
      new Promise((weiter) => {
        antworten = weiter as (wert: unknown) => void
      }),
  })

  zeige()
  const knopf = await screen.findByRole('button', {
    name: texte.seiten.sets.vorlageAusfuehren,
  })
  await userEvent.click(knopf)
  expect(knopf).toBeDisabled()

  act(() => {
    antworten({
      pfad: '',
      befunde: [],
      abgebrochen: false,
      fehler: abgebrochen,
    })
  })

  await waitFor(() => {
    expect(knopf).toBeEnabled()
  })
  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
})
