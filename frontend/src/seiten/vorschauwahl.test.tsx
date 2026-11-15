// **Ein Klick in die Vorschau wählt das Feld darunter**
// (#246).
//
// AGENTS.md §6 sagt zu: die Karte ist das Hauptobjekt. Eine Vorschau, in der
// sich nichts anfassen lässt, ist eine **Abbildung** des Hauptobjekts, nicht
// das Hauptobjekt.
//
// **Und per Tastatur.** Die Flächen sind Knöpfe, keine Klickfänger auf einem
// Bild; wer mit Tab durch die Karte geht, kommt an jedes Feld.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import type { KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

const karte: KarteDTO = {
  id: '5',
  nummer: 0,
  werte: [
    { feld: 'name', wert: 'Feuerball' },
    { feld: 'rule text', wert: 'Fliegend.' },
  ],
}

function belegeMitFlaechen() {
  belegeApi({
    Vorschau: () =>
      Promise.resolve({
        vorschau: {
          stil: '',
          gezeichnet: '',
          bild: '/vorschau/A.png',
          breite: 100,
          hoehe: 140,
          dpi: 150,
          befunde: [],
          flaechen: [
            { feld: 'name', links: 10, oben: 8, breite: 80, hoehe: 12 },
            { feld: 'rule text', links: 10, oben: 60, breite: 80, hoehe: 40 },
          ],
          nennungen: [],
          vorgaben: [],
        },
      }),
    Probevorschau: () => Promise.resolve({ vorschau: null }),
    FeldSetzen: () => Promise.resolve({}),
  })
}

describe('Wahl in der Vorschau', () => {
  it('gibt jeder Fläche einen Knopf mit dem Feldnamen', async () => {
    belegeMitFlaechen()

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    const flaeche = await screen.findByRole('button', {
      name: `rule text — ${texte.seiten.karten.feldWaehlen}`,
    })
    expect(flaeche).toBeInTheDocument()

    // **Und sie liegt nicht in der Tabfolge** (U5, #1416). Neunzehn Flächen
    // standen zwischen der Zoomleiste und dem Namensfeld; wer mit der
    // Tastatur ins Formular will, hat dort dieselben Felder — die Fläche ist
    // der Weg der Maus, und ein Vorleseprogramm erreicht sie weiter.
    expect(flaeche).toHaveAttribute('tabindex', '-1')
  })

  // **Der Klick setzt den Fokus ins Feld.** Wer auf den Regeltext der Karte
  // tippt, will ihn ändern — nicht erst noch das Eingabefeld suchen.
  it('setzt den Fokus in das gewählte Feld', async () => {
    belegeMitFlaechen()

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: `rule text — ${texte.seiten.karten.feldWaehlen}`,
      }),
    )

    expect(screen.getByLabelText('rule text')).toHaveFocus()
  })

  // **Eine Fläche ohne Eingabefeld tut nichts.** Ein Feldstil kann eine Ebene
  // für etwas zeichnen, das die Karte gar nicht führt — dann gibt es nichts zu
  // fokussieren, und ein Fehler wäre die falsche Antwort.
  it('bleibt ruhig, wenn es kein Eingabefeld gibt', async () => {
    belegeApi({
      Vorschau: () =>
        Promise.resolve({
          vorschau: {
            stil: '',
            gezeichnet: '',
            bild: '/vorschau/A.png',
            breite: 100,
            hoehe: 140,
            dpi: 150,
            befunde: [],
            flaechen: [
              { feld: 'gibtsnicht', links: 0, oben: 0, breite: 10, hoehe: 10 },
            ],
            nennungen: [],
            vorgaben: [],
          },
        }),
      Probevorschau: () => Promise.resolve({ vorschau: null }),
      FeldSetzen: () => Promise.resolve({}),
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: `gibtsnicht — ${texte.seiten.karten.feldWaehlen}`,
      }),
    )

    expect(screen.getByLabelText('name')).not.toHaveFocus()
  })

  // **Ohne Flächen bleibt es ein Bild.** Ein Set ohne Stil zeichnet den
  // Behelfsplan; auch der nennt seine Felder, aber ein leeres Feld gibt es
  // trotzdem — dann steht kein Knopf da.
  it('zeigt keine Knöpfe ohne Flächen', async () => {
    belegeApi({
      Vorschau: () =>
        Promise.resolve({
          vorschau: {
            stil: '',
            gezeichnet: '',
            bild: '/vorschau/A.png',
            breite: 100,
            hoehe: 140,
            dpi: 150,
            befunde: [],
            flaechen: [],
            nennungen: [],
            vorgaben: [],
          },
        }),
      Probevorschau: () => Promise.resolve({ vorschau: null }),
      FeldSetzen: () => Promise.resolve({}),
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await screen.findByRole('img')

    expect(
      screen.queryByRole('button', {
        name: new RegExp(texte.seiten.karten.feldWaehlen),
      }),
    ).not.toBeInTheDocument()
  })
})
