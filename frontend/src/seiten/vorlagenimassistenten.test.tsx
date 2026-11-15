// **Die Exportvorlage als Ziel des Ausgabeassistenten** — UXR-09
// (#1917).
//
// Auf der Setliste lädt erst ein Knopf die Vorlagen; im Assistenten hat die
// Wahl „Exportvorlage" das schon gesagt. Was dann zwischen Wahl und Liste
// steht — das Laden und ein Fehler dabei —, steht hier, samt der Fläche, die
// auf ihr Set wartet.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Ausgeben } from './Ausgeben'
import { SetVorlagenausfuhr } from './SetVorlagenausfuhr'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

const probeset = {
  id: '7',
  name: 'Mein Probeset',
  spiel: 'magic',
  stil: 'magic-m15',
  herkunft: '/pfad/probe.mse-set',
  karten: 2,
  herkunftFehlt: false,
}

const vorlage = {
  paket: 'magic-spoiler.mse-export-template',
  name: 'Spoiler',
  dateityp: '*.html|*.html',
  optionen: [],
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <SetVorlagenausfuhr satz={probeset} imAssistenten />
    </Sprachraum>,
  )
}

describe('Im Assistenten', () => {
  // **Solange geladen wird, sagt es ein Satz** — kein Knopf, der „Über eine
  // Exportvorlage …" anbietet, was eben schon gewählt wurde.
  it('sagt, dass die Vorlagen geladen werden', async () => {
    belegeApi({ Exportvorlagen: () => new Promise(() => undefined) })

    zeige()

    expect(await screen.findByRole('status')).toHaveTextContent(
      texte.bausteine.zustaende.laedt,
    )
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  // **Ein Fehler beim Laden steht am Ort — und der Knopf ist der zweite
  // Versuch.**
  it('nennt einen Lesefehler und lädt auf Knopfdruck noch einmal', async () => {
    const Exportvorlagen = vi
      .fn()
      .mockResolvedValueOnce({
        vorlagen: [],
        befunde: [],
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'der Bestand ist belegt',
        },
      })
      .mockResolvedValueOnce({ vorlagen: [vorlage], befunde: [] })

    belegeApi({
      // **Der Assistent merkt sich das Ziel** (AP-10, #2367).
      EinstellungSetzen: () => Promise.resolve({}),
      Exportvorlagen,
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      texte.fehler.unbekannt,
    )

    await userEvent.click(
      screen.getByRole('button', {
        name: `${texte.seiten.sets.vorlageAusfuhr} — Mein Probeset`,
      }),
    )

    expect(
      await screen.findByRole('combobox', {
        name: texte.seiten.sets.vorlageWaehlen,
      }),
    ).toHaveValue(vorlage.paket)
    expect(Exportvorlagen).toHaveBeenCalledTimes(2)
  })
})

describe('Die Fläche eines Ziels', () => {
  // **Exportvorlage und MTG JSON brauchen das Set ganz** — bis es da ist,
  // steht der Ladesatz an ihrer Stelle.
  it('wartet auf das Set', async () => {
    belegeApi({
      // **Der Assistent merkt sich das Ziel** (AP-10, #2367).
      EinstellungSetzen: () => Promise.resolve({}),
      Sets: () => new Promise(() => undefined),
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Ausgeben setID="7" seite="ausgeben" wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('radio', {
        name: texte.seiten.sets.mtgjsonUeberschrift,
      }),
    )

    expect(
      screen.getByRole('region', {
        name: texte.seiten.sets.mtgjsonUeberschrift,
      }),
    ).toHaveTextContent(texte.bausteine.zustaende.laedt)
  })
})
