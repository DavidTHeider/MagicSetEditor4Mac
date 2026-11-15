// **Die Eingaben der Bilderseite** — Ordner, Vorlage, Format, Güte, Breite.
//
// Eigene Datei, weil `bilder.test.tsx` mit ihnen über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort steht, **was ein Lauf tut**, hier, **was
// die Seite entgegennimmt**.
// Die Bildausgabe eines Sets.
//
// **Der Stapel läuft hier und nicht im Hintergrund** — deshalb stehen
// Fortschritt und Abbruch in diesen Tests und nicht in einem Go-Test: sie sind
// eine Eigenschaft der Oberfläche.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

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

describe('Die Eingaben', () => {
  // **Alle fünf Formate stehen zur Wahl** (#852). SVG lag bis zum 2026-08-22
  // als fertiges Paket da, und die Auswahl bot es nicht an — ein Format, das
  // niemand wählen kann, gibt es für Anwendende nicht.
  it('bietet jedes Format an, das der Dienst kennt', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    })

    zeige()

    const auswahl = screen.getByLabelText(texte.seiten.bilder.format)

    for (const format of ['png', 'jpg', 'bmp', 'tiff', 'svg']) {
      await userEvent.selectOptions(auswahl, format)
      expect((auswahl as HTMLSelectElement).value).toBe(format)
    }
  })

  // **Was eingestellt wird, kommt beim Dienst an** — sonst gäbe es drei Felder,
  // die aussehen, als täten sie etwas.
  // **Die gewählte Güte reist mit** — Ω.4 aus
  // #1029.
  //
  // Eine Wahl, die nur im Fenster steht und den Aufruf nicht erreicht, sieht
  // aus wie eine Wahl. Der Test greift deshalb den Wert am Aufruf ab und
  // nicht am Bedienelement.
  it('gibt die gewählte Güte weiter', async () => {
    const BildSchreiben = vi.fn(() => geschrieben())

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () =>
        Promise.resolve({ ordner: '/tmp/bilder', abgebrochen: false }),
      Bildnamen: () => Promise.resolve(namen(1)),
      BildSchreiben,
    })

    zeige()

    await druecke(texte.seiten.bilder.ordnerWaehlen)

    await userEvent.selectOptions(
      screen.getByLabelText(texte.seiten.bilder.guete),
      'ultra',
    )

    await druecke(texte.seiten.bilder.ausgeben)

    await waitFor(() => {
      expect(BildSchreiben).toHaveBeenCalled()
    })

    expect(BildSchreiben).toHaveBeenCalledWith(
      '1',
      '/tmp/bilder',
      '001-Karte.png',
      'png',
      750,
      'nummerierenneu',
      'ultra',
    )
  })

  it('gibt Vorlage, Format und Breite weiter', async () => {
    const Bildnamen = vi.fn(() => Promise.resolve(namen(1)))
    const BildSchreiben = vi.fn(() => geschrieben())

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () =>
        Promise.resolve({ ordner: '/tmp/bilder', abgebrochen: false }),
      Bildnamen,
      BildSchreiben,
    })

    zeige()

    await druecke(texte.seiten.bilder.ordnerWaehlen)

    const vorlage = screen.getByLabelText(texte.seiten.bilder.vorlage)
    await userEvent.clear(vorlage)
    await userEvent.type(vorlage, '{{name}')

    await userEvent.selectOptions(
      screen.getByLabelText(texte.seiten.bilder.format),
      'jpg',
    )

    const breite = screen.getByLabelText(texte.seiten.bilder.breite)
    await userEvent.clear(breite)
    await userEvent.type(breite, '1500')

    await druecke(texte.seiten.bilder.ausgeben)

    await waitFor(() => {
      expect(BildSchreiben).toHaveBeenCalled()
    })

    expect(Bildnamen).toHaveBeenCalledWith(
      '4',
      '{name}',
      'jpg',
      'nummerierenneu',
      [],
    )
    expect(BildSchreiben).toHaveBeenCalledWith(
      '1',
      '/tmp/bilder',
      '001-Karte.png',
      'jpg',
      1500,
      'nummerierenneu',

      // **Und die Güte** — Ω.4: sie reist mit jedem Aufruf, damit die
      // Wahl nicht an einer Stelle hängen bleibt.
      'medium',
    )
  })

  it('meldet einen Fehler beim Ordnerdialog', async () => {
    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () =>
        Promise.resolve({
          ordner: '',
          abgebrochen: false,
          fehler: {
            code: 'unbekannt',
            schluessel: 'fehler.unbekannt',
            meldung: 'der Dialog ging nicht auf',
            feld: '',
          },
        }),
    })

    zeige()

    await druecke(texte.seiten.bilder.ordnerWaehlen)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'der Dialog ging nicht auf',
    )
  })

  // Ein abgebrochener Dialog lässt den Ordner, wie er war.
  it('behält den Ordner, wenn niemand einen wählt', async () => {
    let erstesMal = true

    belegeApi({
      // **Die Ausgabeauswahl lädt die Kartenliste** (#1167).
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () => {
        const antwort = erstesMal
          ? { ordner: '/tmp/bilder', abgebrochen: false }
          : { ordner: '', abgebrochen: true }
        erstesMal = false

        return Promise.resolve(antwort)
      },
    })

    zeige()

    await druecke(texte.seiten.bilder.ordnerWaehlen)
    await druecke(texte.seiten.bilder.ordnerWaehlen)

    expect(await screen.findByText('/tmp/bilder')).toBeInTheDocument()
  })
})
