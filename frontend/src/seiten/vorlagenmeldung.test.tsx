// Was eine Exportvorlage meldet, wenn sie nicht durchläuft.
//
// **Eine dritte Datei**, und der Schnitt folgt wieder der Frage (AGENTS.md §2):
// dort *was lässt sich wählen*, dort *was geschieht beim Ausführen*, hier *was
// steht da, wenn es schiefgeht*. Das ist der vierte Haken aus
// #209 — und
// er ist der wichtigste: von neunzehn bezogenen Vorlagen laufen drei.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
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

describe('Was eine Exportvorlage meldet', () => {
  // **Eine Vorlage, die nicht durchläuft, meldet warum** — der vierte Haken.
  // Sie ist der Normalfall, nicht die Ausnahme: von neunzehn bezogenen
  // Vorlagen laufen drei.
  it('sagt, warum eine Vorlage nicht durchläuft', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({ vorlagen: [vorlage()], befunde: [] }),
      VorlageAusfuehren: () =>
        Promise.resolve({
          pfad: '',
          abgebrochen: false,
          befunde: ['der Name "write_image_file" ist hier nicht bekannt'],
          fehler: {
            code: 'validierung',
            schluessel: 'validierung',
            meldung: 'Die Eingabe ist nicht gültig.',
            feld: '',
          },
        }),
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.vorlageAusfuehren,
      }),
    )

    const meldung = await screen.findByRole('alert')
    expect(meldung).toHaveTextContent(texte.seiten.sets.vorlageGescheitert)
    expect(meldung).toHaveTextContent('write_image_file')
  })

  it('sagt es, wenn keine Vorlage installiert ist', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () => Promise.resolve({ vorlagen: [], befunde: [] }),
    })

    zeige()

    expect(
      await screen.findByText(texte.seiten.sets.vorlageKeine),
    ).toBeInTheDocument()
  })

  it('zeigt die Befunde beim Lesen der Vorlagen', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({
          vorlagen: [vorlage()],
          befunde: [
            'leer.mse-export-template: die Exportvorlage führt kein Skript',
          ],
        }),
    })

    zeige()

    expect(
      await screen.findByText(
        'leer.mse-export-template: die Exportvorlage führt kein Skript',
      ),
    ).toBeInTheDocument()
  })

  it('meldet einen Fehler beim Lesen der Vorlagen', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({
          vorlagen: [],
          befunde: [],
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'der Bestand ist nicht lesbar',
            feld: '',
          },
        }),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      texte.fehler.unbekannt,
    )
  })

  // **Ein abgebrochener Dialog ist keine Meldung wert** — wer selbst abbricht,
  // weiß, dass nichts geschrieben wurde.
  it('sagt nichts, wenn niemand einen Ort gewählt hat', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({ vorlagen: [vorlage()], befunde: [] }),
      VorlageAusfuehren: () =>
        Promise.resolve({ pfad: '', befunde: [], abgebrochen: true }),
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.vorlageAusfuehren,
      }),
    )

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('zeigt die Befunde eines geglückten Laufs', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({ vorlagen: [vorlage()], befunde: [] }),
      VorlageAusfuehren: () =>
        Promise.resolve({
          pfad: '/pfad/a.txt',
          befunde: ['options.unbekannt: diese Wahl führt die Vorlage nicht'],
          abgebrochen: false,
        }),
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.vorlageAusfuehren,
      }),
    )

    expect(
      await screen.findByText(
        'options.unbekannt: diese Wahl führt die Vorlage nicht',
      ),
    ).toBeInTheDocument()
  })

  it('wechselt die Vorlage samt ihren Vorgaben', async () => {
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
              optionen: [
                {
                  art: 'text',
                  name: 'kuerzel',
                  beschreibung: 'Kürzel',
                  vorgabe: 'ZWO',
                  auswahl: [],
                  wert: 'ZWO',
                },
              ],
            }),
          ],
          befunde: [],
        }),
    })

    zeige()

    await userEvent.selectOptions(
      await screen.findByRole('combobox', {
        name: texte.seiten.sets.vorlageWaehlen,
      }),
      'zweite.mse-export-template',
    )

    expect(screen.getByRole('textbox', { name: 'Kürzel' })).toHaveValue('ZWO')
  })

  // **Schließen gibt die Karte zurück** — seit UXR-04 (#1912) hängt sie die
  // Fläche aus; vorher stand danach wieder der Knopf „Über eine
  // Exportvorlage …", und den gibt es jetzt nur im Menü.
  it('lässt sich wieder schließen', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen: () =>
        Promise.resolve({ vorlagen: [vorlage()], befunde: [] }),
    })

    const schliessen = vi.fn()
    zeige(schliessen)

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.vorlageSchliessen,
      }),
    )

    expect(schliessen).toHaveBeenCalledTimes(1)
  })

  // **Ein Lesefehler lässt einen zweiten Versuch** — der Knopf steht nur
  // dann, als Wiederholung.
  it('lädt nach einem Lesefehler auf Knopfdruck noch einmal', async () => {
    const Exportvorlagen = vi
      .fn()
      .mockResolvedValueOnce({
        vorlagen: [],
        befunde: [],
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'der Bestand ist nicht lesbar',
          feld: '',
        },
      })
      .mockResolvedValue({ vorlagen: [vorlage()], befunde: [] })
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Exportvorlagen,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.seiten.sets.vorlageAusfuhr} — Mein Probeset`,
      }),
    )

    expect(
      await screen.findByRole('combobox', {
        name: texte.seiten.sets.vorlageWaehlen,
      }),
    ).toBeInTheDocument()
    expect(Exportvorlagen).toHaveBeenCalledTimes(2)
  })

  // **Auch eine Auswahl und ein Text reisen mit.** Ohne diese Fälle blieben die
  // beiden Eingabearten neben dem Kästchen ungeprüft — und `option.auswahl`
  // ist genau die Stelle, an der ein `null` aus Go das Frontend zerrisse (§3).
  it('schickt eine Auswahl und einen Text mit', async () => {
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
      VorlageAusfuehren,
    })

    zeige()

    await userEvent.selectOptions(
      await screen.findByRole('combobox', { name: 'Forum' }),
      'salvation',
    )

    const feld = screen.getByRole('textbox', { name: 'Kürzel' })
    await userEvent.clear(feld)
    await userEvent.type(feld, 'ZWO')

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.sets.vorlageAusfuehren }),
    )

    expect(VorlageAusfuehren).toHaveBeenCalledWith(
      '7',
      'probe-ausfuhr.mse-export-template',
      '*.txt|*.txt',
      { forum: 'salvation', kuerzel: 'ZWO' },
      [],
    )
  })
})
