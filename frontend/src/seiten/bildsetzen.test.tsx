// **Ein Bildfeld lässt sich füllen** —
// #246.
//
// # Wofür
//
// Ein Kartenfeld der Art `image` war bis hierher ein Textfeld: wer eine
// Illustration einsetzen wollte, musste den Dateinamen kennen, den das Set
// später vergeben wird. MSE öffnet dafür einen Dateidialog
// (`ImageValueEditor::onLeftDClick`, `src/gui/value/image.cpp`, Zeilen 22–35,
// Fassung f9d9356 vom 2021-07-06).
//
// # Der Dialog steht in Go
//
// Ein Fenster kann keine Datei vom Dateisystem lesen — es bekäme einen Namen
// und keinen Pfad. Die Oberfläche ruft `BildSetzen`, und was dabei geschieht,
// entscheidet `internal/app`.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO, KarteDTO, VorschauAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { zuschnittantwort } from '../test/zuschnitt'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function vorschau(): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: [{ feld: 'kunst', links: 0, oben: 0, breite: 10, hoehe: 5 }],
      nennungen: [],
      vorgaben: [],
    },
  }
}

function form(art: string): FeldbeschreibungDTO {
  return {
    name: 'kunst',
    aufschrift: 'Kunst',
    art,
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
  }
}

function karte(werte: KarteDTO['werte'] = []): KarteDTO {
  return { id: '5', nummer: 0, werte }
}

function zeige(art: string, teile = {}) {
  const BildSetzen = vi.fn(() => Promise.resolve(zuschnittantwort()))

  belegeApi({
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
    BildSetzen,
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte()}
        formen={[form(art)]}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )

  return BildSetzen
}

describe('Ein Bildfeld', () => {
  it('bietet einen Knopf zum Wählen einer Datei', async () => {
    const BildSetzen = zeige('image')

    await userEvent.click(
      await screen.findByRole('button', {
        name: `Kunst — ${texte.seiten.karten.bildWaehlen}`,
      }),
    )

    await waitFor(() => {
      expect(BildSetzen).toHaveBeenCalledWith('5', 'kunst')
    })
  })

  // **Nur ein Bildfeld.** Ein Textfeld bekommt keinen — der Knopf öffnet einen
  // Dateidialog, und in ein Textfeld gehört kein Dateiname.
  it('gibt einem Textfeld keinen', async () => {
    zeige('text')

    await screen.findByLabelText('kunst')

    expect(
      screen.queryByRole('button', {
        name: `Kunst — ${texte.seiten.karten.bildWaehlen}`,
      }),
    ).not.toBeInTheDocument()
  })

  // **Ein Fehler erscheint am Feld** (§6), nicht als anonymer Hinweis.
  it('meldet einen Fehler am Feld', async () => {
    zeige('image', {
      BildSetzen: () => Promise.reject(new Error('Die Datei ist kein Bild.')),
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: `Kunst — ${texte.seiten.karten.bildWaehlen}`,
      }),
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Die Datei ist kein Bild.',
      )
    })
  })
})

// **Ein Bildfeld lässt sich wieder leeren** — der erste Abnahmekasten aus
// #246
// nennt beides: „füllen **und wieder leeren**".
//
// # Wie, und warum das genügt
//
// Ein Bildfeld trägt im Set einen **Dateinamen**. Der steht im Editor in einem
// Textfeld — dieselbe Eingabe wie bei jedem anderen Feld —, und wer ihn löscht,
// leert das Feld. Ein eigener Knopf „Bild entfernen" wäre ein zweiter Weg zu
// derselben Handlung; MSE hat auch keinen (ein Bildfeld dort wird über sein
// Kontextmenü geleert).
//
// **Und er ist mit der Tastatur erreichbar**, weil es ein gewöhnliches
// Eingabefeld ist (§6).
describe('Ein Bildfeld leeren', () => {
  it('schreibt den leeren Namen, wenn jemand das Feld leert', async () => {
    const FeldSetzen = vi.fn(() => Promise.resolve(zuschnittantwort()))

    belegeApi({
      Vorschau: () => Promise.resolve(vorschau()),
      Probevorschau: () => Promise.resolve(vorschau()),
      FeldSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte([{ feld: 'kunst', wert: 'bild1.png' }])}
          formen={[form('image')]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    const feld = await screen.findByRole('textbox', { name: 'kunst' })
    expect(feld).toHaveValue('bild1.png')

    await userEvent.clear(feld)
    await userEvent.tab()

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith(
        '1',
        '5',
        'kunst',
        '',
        expect.any(Array),
      )
    })
  })
})
