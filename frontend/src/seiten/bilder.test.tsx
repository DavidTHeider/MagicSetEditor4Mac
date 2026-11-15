// Die Bildausgabe eines Sets.
//
// **Der Stapel läuft hier und nicht im Hintergrund** — deshalb stehen
// Fortschritt und Abbruch in diesen Tests und nicht in einem Go-Test: sie sind
// eine Eigenschaft der Oberfläche.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { BildAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Bilder } from './Bilder'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function zeige(setID = '4') {
  render(
    <Sprachraum sprache="de">
      <Bilder setID={setID} />
    </Sprachraum>,
  )
}

function namen(anzahl: number) {
  return {
    namen: Array.from({ length: anzahl }, (_, stelle) => ({
      karteID: String(stelle + 1),
      nummer: stelle + 1,
      datei: `00${String(stelle + 1)}-Karte.png`,
    })),
  }
}

function geschrieben(teile: Record<string, unknown> = {}) {
  return Promise.resolve({
    pfad: '/tmp/bilder/001-Karte.png',
    breite: 750,
    hoehe: 1046,
    befunde: [],
    uebersprungen: false,
    ...teile,
  })
}

async function druecke(name: string) {
  await userEvent.click(await screen.findByRole('button', { name }))
}

// **Der Fall „kein Set" steht hier nicht mehr.** Seit
// #288 rendert
// der Arbeitsplatz diese Lasche nur mit offenem Set; `setID` ist eine
// Zeichenkette, und der Wächter wäre ein Zweig, den kein Aufruf betreten kann
// (§3). Der leere Zustand steht in `setarbeitsplatz.test.tsx`.

describe('Mit einem gewählten Set', () => {
  it('nimmt einen Ordner entgegen und zeigt ihn', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () =>
        Promise.resolve({ ordner: '/tmp/bilder', abgebrochen: false }),
    })

    zeige()

    await druecke(texte.seiten.bilder.ordnerWaehlen)

    expect(await screen.findByText('/tmp/bilder')).toBeInTheDocument()
  })

  // **Der Lauf schreibt jede Karte einzeln** — mit dem Namen, den das Backend
  // vergeben hat, nicht mit einem, den die Oberfläche sich ausdenkt.
  it('schreibt jede Karte und zeigt, wie weit es ist', async () => {
    const BildSchreiben = vi.fn(() => geschrieben())

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () =>
        Promise.resolve({ ordner: '/tmp/bilder', abgebrochen: false }),
      Bildnamen: () => Promise.resolve(namen(3)),
      BildSchreiben,
    })

    zeige()

    await druecke(texte.seiten.bilder.ordnerWaehlen)
    await druecke(texte.seiten.bilder.ausgeben)

    await waitFor(() => {
      expect(BildSchreiben).toHaveBeenCalledTimes(3)
    })

    expect(BildSchreiben).toHaveBeenCalledWith(
      '1',
      '/tmp/bilder',
      '001-Karte.png',
      'png',
      750,

      // **Und das Verhalten bei einem Namenskonflikt** (#491). Ohne Beleg der
      // Einstellungen gilt MSEs Vorgabe — der Lauf hängt nicht daran, ob der
      // Bestand gerade lesbar ist.
      'nummerierenneu',

      // **Und die Güte** — Ω.4: sie reist mit jedem Aufruf, damit die
      // Wahl nicht an einer Stelle hängen bleibt.
      'medium',
    )

    expect(await screen.findByRole('status')).toHaveTextContent('3')
  })

  // **Ein Stapellauf, der sich nicht abbrechen lässt, ist einer, den man nicht
  // anfängt.** Dreihundert Karten in Druckauflösung dauern.
  it('bricht ab, wenn jemand es verlangt', async () => {
    let laeuft: (() => void) | undefined
    const BildSchreiben = vi.fn(
      () =>
        new Promise<BildAntwort>((erfuellen) => {
          laeuft = () => {
            erfuellen({
              pfad: '/tmp/bilder/001-Karte.png',
              breite: 1,
              hoehe: 1,
              befunde: [],
              uebersprungen: false,
            })
          }
        }),
    )

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () =>
        Promise.resolve({ ordner: '/tmp/bilder', abgebrochen: false }),
      Bildnamen: () => Promise.resolve(namen(5)),
      BildSchreiben,
    })

    zeige()

    await druecke(texte.seiten.bilder.ordnerWaehlen)
    await druecke(texte.seiten.bilder.ausgeben)

    await waitFor(() => {
      expect(laeuft).toBeDefined()
    })

    await druecke(texte.seiten.bilder.abbrechen)

    laeuft?.()

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent(
        texte.seiten.bilder.abgebrochen,
      )
    })

    // **Nach dem Abbruch kein weiteres Bild.** Das erste lief schon; das zweite
    // darf nicht mehr angefangen werden.
    expect(BildSchreiben).toHaveBeenCalledTimes(1)
  })

  // **Ein Fehler hält den Lauf an** (§8): wer weiterschriebe, sammelte
  // dreihundert Meldungen derselben Ursache, und niemand liest sie.
  it('hält beim ersten Fehler an und sagt, wie weit es kam', async () => {
    const BildSchreiben = vi
      .fn()
      .mockImplementationOnce(() => geschrieben())
      .mockImplementationOnce(() =>
        geschrieben({
          fehler: {
            code: 'validierung',
            schluessel: 'fehler.validierung',
            meldung: 'kein Platz mehr',
            feld: '',
          },
        }),
      )

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () =>
        Promise.resolve({ ordner: '/tmp/bilder', abgebrochen: false }),
      Bildnamen: () => Promise.resolve(namen(4)),
      BildSchreiben,
    })

    zeige()

    await druecke(texte.seiten.bilder.ordnerWaehlen)
    await druecke(texte.seiten.bilder.ausgeben)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'kein Platz mehr',
    )
    expect(BildSchreiben).toHaveBeenCalledTimes(2)
  })

  // **Was dem Bild fehlt, steht beim Erfolg** (§8).
  it('nennt, was den Bildern fehlte', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () =>
        Promise.resolve({ ordner: '/tmp/bilder', abgebrochen: false }),
      Bildnamen: () => Promise.resolve(namen(1)),
      BildSchreiben: () =>
        geschrieben({ befunde: ['001-Karte.png → Ebene 3: keine Schrift'] }),
    })

    zeige()

    await druecke(texte.seiten.bilder.ordnerWaehlen)
    await druecke(texte.seiten.bilder.ausgeben)

    expect(
      await screen.findByText('001-Karte.png → Ebene 3: keine Schrift'),
    ).toBeInTheDocument()
  })

  // Ohne Ordner läuft nichts — und der Knopf sagt es, statt es zu verschweigen.
  it('gibt ohne Ordner nichts aus', async () => {
    const Bildnamen = vi.fn(() => Promise.resolve(namen(3)))

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Bildnamen,
    })

    zeige()

    expect(
      await screen.findByRole('button', { name: texte.seiten.bilder.ausgeben }),
    ).toBeDisabled()
    expect(Bildnamen).not.toHaveBeenCalled()
  })
})

// Ein Set ohne Karten ist kein Fehler — und kein Lauf, der bei null hängt.
it('meldet einen Lauf über null Karten als fertig', async () => {
  belegeApi({
    // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    EinstellungSetzen: () => Promise.resolve({}),
    Bildordner: () =>
      Promise.resolve({ ordner: '/tmp/bilder', abgebrochen: false }),
    Bildnamen: () => Promise.resolve(namen(0)),
    BildSchreiben: () => geschrieben(),
  })

  zeige()

  await druecke(texte.seiten.bilder.ordnerWaehlen)
  await druecke(texte.seiten.bilder.ausgeben)

  expect(await screen.findByRole('status')).toHaveTextContent(
    texte.seiten.bilder.fertig,
  )
})
