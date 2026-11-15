// **Der Weg vom Wählen bis zum Ablegen** —
// #576.
//
// Bis hierher war das Einsetzen **ein** Aufruf: wählen, mit MSEs Vorgabewahl
// zuschneiden, ablegen. Jetzt liegt das Zuschnittfenster dazwischen, und
// dieser Weg hat drei Stellen, an denen §7 gilt:
//
//   - Regel 1: ein Kartenwechsel **räumt** den offenen Zuschnitt.
//   - Regel 3: nach einem **gescheiterten** Einsetzen bleibt die Wahl stehen —
//     wer die Meldung liest, soll sehen, worauf sie sich bezieht.
//   - Regel 4: ein Vorgang je Datensatz, und der Schlüssel ist derselbe wie
//     beim Schreiben desselben Feldes.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO, VorschauAntwort } from '../api'
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
      // **Die Fläche muss dastehen**: der Editor bietet an, was auf der
      // Karte **liegt** (`angeboteneFelder`) — eine leere Liste ergäbe ein
      // leeres Formular.
      flaechen: [{ feld: 'kunst', links: 0, oben: 0, breite: 10, hoehe: 5 }],
      nennungen: [],
      vorgaben: [],
    },
  }
}

const bildform = {
  name: 'kunst',
  aufschrift: 'Kunst',
  art: 'image',
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

function karte(id = '5'): KarteDTO {
  return { id, nummer: 0, werte: [] }
}

function zeige(teile = {}) {
  belegeApi({
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
    BildSetzen: () => Promise.resolve(zuschnittantwort()),
    Zuschnittgrenze: () =>
      Promise.resolve({ x: 0, y: 0, breite: 125, hoehe: 100 }),
    ...teile,
  })

  const gezeichnet = render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte()}
        formen={[bildform]}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )

  return gezeichnet
}

// waehle drückt „Bild wählen" und wartet auf das Zuschnittfenster.
async function waehle() {
  await userEvent.click(
    await screen.findByRole('button', {
      name: `Kunst — ${texte.seiten.karten.bildWaehlen}`,
    }),
  )

  return screen.findByRole('button', {
    name: texte.seiten.karten.zuschnitt.uebernehmen,
  })
}

describe('Der Weg über das Zuschnittfenster', () => {
  it('legt ab, was gewählt wurde', async () => {
    const BildEinsetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({ BildEinsetzen })

    await userEvent.click(await waehle())

    await waitFor(() => {
      expect(BildEinsetzen).toHaveBeenCalledWith(
        '1',
        '5',
        'kunst',
        '/pfad/urlaub.jpg',
        {
          x: 0,
          y: 0,
          breite: 125,
          hoehe: 100,
        },
      )
    })

    // **Und danach ist es fort** — die Wahl ist eingelöst.
    await waitFor(() => {
      expect(
        screen.queryByRole('button', {
          name: texte.seiten.karten.zuschnitt.uebernehmen,
        }),
      ).not.toBeInTheDocument()
    })
  })

  // **Regel 3: nach einem gescheiterten Einsetzen bleibt die Wahl stehen.**
  it('behält die Wahl, wenn das Ablegen scheitert', async () => {
    zeige({
      BildEinsetzen: () => Promise.reject(new Error('Die Platte ist voll.')),
    })

    await userEvent.click(await waehle())

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Die Platte ist voll.',
    )

    expect(
      screen.getByRole('button', {
        name: texte.seiten.karten.zuschnitt.uebernehmen,
      }),
    ).toBeInTheDocument()
  })

  it('lässt sich abbrechen, ohne etwas abzulegen', async () => {
    const BildEinsetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({ BildEinsetzen })
    await waehle()

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.seiten.karten.zuschnitt.abbrechen,
      }),
    )

    expect(
      screen.queryByRole('button', {
        name: texte.seiten.karten.zuschnitt.uebernehmen,
      }),
    ).not.toBeInTheDocument()
    expect(BildEinsetzen).not.toHaveBeenCalled()
  })

  // **Regel 1: ein Kartenwechsel räumt den offenen Kontext.**
  //
  // Ohne das stünde nach dem Wechsel ein Fenster da, dessen „Einsetzen" ein
  // Feld der **vorigen** Karte träfe — und niemand sähe es.
  it('schließt sich beim Kartenwechsel', async () => {
    const { rerender } = zeige()

    await waehle()

    rerender(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte('9')}
          formen={[bildform]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(
        screen.queryByRole('button', {
          name: texte.seiten.karten.zuschnitt.uebernehmen,
        }),
      ).not.toBeInTheDocument()
    })
  })

  // **Eine abgebrochene Dateiwahl macht kein Fenster auf.**
  //
  // Ein leerer Pfad heißt „niemand hat gewählt" — dieselbe Übereinkunft wie
  // beim Öffnen eines Sets.
  it('bleibt bei einer abgebrochenen Dateiwahl zu', async () => {
    zeige({
      BildSetzen: () => Promise.resolve(zuschnittantwort({ pfad: '' })),
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: `Kunst — ${texte.seiten.karten.bildWaehlen}`,
      }),
    )

    await waitFor(() => {
      expect(
        screen.queryByRole('button', {
          name: texte.seiten.karten.zuschnitt.uebernehmen,
        }),
      ).not.toBeInTheDocument()
    })
  })
})
