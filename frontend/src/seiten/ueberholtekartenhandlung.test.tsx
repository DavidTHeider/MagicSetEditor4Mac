// **Eine Antwort gehört zu dem Stand, aus dem sie stammt** — §7 Regel 2,
// #716.
//
// # Der Befund
//
// Aus dem Zweitblick durch Codex vom 2026-08-21: `Karten` bleibt beim
// Setwechsel **dieselbe** React-Komponente. Der Wechselblock räumt Auswahl und
// Anhakung, besitzt aber keinen Laufzähler.
//
// Während `KarteAnlegen` für Set A läuft, kann auf Set B gewechselt werden. Die
// alte Antwort setzte danach ungeprüft die neue Auswahl, lud die aktuelle Liste
// nach und meldete den Bestand als geändert — eine Karte aus einem Set, das
// niemand mehr offen hat, stand aufgeschlagen da.
//
// Dieselbe Klasse bei Kopieren, Ordnen und Löschen; beim Anlegefehler
// zusätzlich Regel 1, denn der Setwechsel räumte ihn überhaupt nicht.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { oeffneKartenhandlungen } from '../test/kartenlasche'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'
import { ausKarten } from '../test/kartenproben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function karte(id: string, nummer: number, name: string): KarteDTO {
  return { id, nummer, werte: [{ feld: 'name', wert: name }] }
}

describe('Eine Kartenhandlung, die vom Setwechsel überholt wird', () => {
  it('trägt ihre Antwort nicht in das neue Set ein', async () => {
    const nutzer = userEvent.setup()

    // **Die Antwort kommt erst, wenn der Test sie freigibt** — so lässt sich
    // der Wechsel dazwischenschieben.
    let freigeben: (() => void) | undefined
    const angelegt = new Promise<void>((aufloesen) => {
      freigeben = aufloesen
    })

    const bestandGeaendert = vi.fn()

    belegeApi({
      ...ausKarten((setID: string) =>
        Promise.resolve({
          karten:
            setID === '7' ? [karte('8', 0, 'Aus A')] : [karte('9', 0, 'Aus B')],
        }),
      ),
      Karten: (setID: string) =>
        Promise.resolve({
          karten:
            setID === '7' ? [karte('8', 0, 'Aus A')] : [karte('9', 0, 'Aus B')],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
      Probevorschau: () => Promise.resolve({ vorschau: null }),
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      KarteAnlegen: () =>
        angelegt.then(() => ({ karteID: '99', fehler: null })),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="7"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={bestandGeaendert}
        />
      </Sprachraum>,
    )

    await screen.findByText('Aus A')

    await nutzer.click(
      screen.getByRole('button', { name: texte.seiten.karten.anlegen }),
    )

    // **Der Setwechsel, während der Aufruf läuft.**
    rerender(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="12"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={bestandGeaendert}
        />
      </Sprachraum>,
    )

    await screen.findByText('Aus B')

    // **Nach dem Setwechsel steht die Lasche neu** — auch das Menü der Karte
    // (bis UXR-06 der Aufklapper „Ablage und Bezug").
    await oeffneKartenhandlungen()

    bestandGeaendert.mockClear()

    freigeben?.()

    // **Gewartet wird auf ein Abschlusssignal** (§3): der Knopf ist wieder
    // bedienbar, sobald der Vorgang beendet ist.
    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: texte.seiten.karten.anlegen }),
      ).not.toBeDisabled()
    })

    // **Die Karte des alten Sets steht nicht da** — und der Bestand des neuen
    // wurde nicht als geändert gemeldet.
    expect(screen.getByText('Aus B')).toBeInTheDocument()
    expect(bestandGeaendert).not.toHaveBeenCalled()
  })

  // **Und ein Fehler des alten Sets erscheint nicht am neuen.**
  //
  // „Die Karte ließ sich nicht anlegen" zu einem Set, das niemand mehr offen
  // hat, sieht aus wie ein Fehler des aktuellen.
  it('meldet ihren Fehler nicht am neuen Set', async () => {
    const nutzer = userEvent.setup()

    let scheitern: (() => void) | undefined
    const angelegt = new Promise<void>((_, ablehnen) => {
      scheitern = () => {
        ablehnen(new Error('geht nicht'))
      }
    })

    belegeApi({
      ...ausKarten((setID: string) =>
        Promise.resolve({
          karten:
            setID === '7' ? [karte('8', 0, 'Aus A')] : [karte('9', 0, 'Aus B')],
        }),
      ),
      Karten: (setID: string) =>
        Promise.resolve({
          karten:
            setID === '7' ? [karte('8', 0, 'Aus A')] : [karte('9', 0, 'Aus B')],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
      Probevorschau: () => Promise.resolve({ vorschau: null }),
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      KarteAnlegen: () =>
        angelegt.then(() => ({ karteID: '99', fehler: null })),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="7"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={vi.fn()}
        />
      </Sprachraum>,
    )

    await screen.findByText('Aus A')

    await nutzer.click(
      screen.getByRole('button', { name: texte.seiten.karten.anlegen }),
    )

    rerender(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="12"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={vi.fn()}
        />
      </Sprachraum>,
    )

    await screen.findByText('Aus B')

    scheitern?.()

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: texte.seiten.karten.anlegen }),
      ).not.toBeDisabled()
    })

    // **Gesucht wird in der Werkzeugzeile**, nicht auf der ganzen Seite: die
    // bringt eigene Meldungen mit (die Vorschau der Attrappe kommt ohne Bild),
    // und eine Zählung über alle wäre ein Wächter, der beim nächsten Hinweis
    // umkippt.
    const zeile = screen
      .getByRole('button', { name: texte.seiten.karten.anlegen })
      .closest('div')

    expect(zeile).not.toBeNull()
    expect(within(zeile as HTMLElement).queryByRole('alert')).toBeNull()
  })

  // **Dieselbe Klasse bei den Werkzeugen.** `Kartenkopieren`, `Kartenordnen`
  // und `Kartenloeschen` rufen ihren Rückruf aus dem Zeichenvorgang, in dem
  // geklickt wurde — und der trug bis hierher ungeprüft in das neue Set ein.
  it('trägt auch eine Kopie nicht in das neue Set ein', async () => {
    const nutzer = userEvent.setup()

    let freigeben: (() => void) | undefined
    const kopiert = new Promise<void>((aufloesen) => {
      freigeben = aufloesen
    })

    const bestandGeaendert = vi.fn()

    belegeApi({
      ...ausKarten((setID: string) =>
        Promise.resolve({
          karten:
            setID === '7' ? [karte('8', 0, 'Aus A')] : [karte('9', 0, 'Aus B')],
        }),
      ),
      Karten: (setID: string) =>
        Promise.resolve({
          karten:
            setID === '7' ? [karte('8', 0, 'Aus A')] : [karte('9', 0, 'Aus B')],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
      Probevorschau: () => Promise.resolve({ vorschau: null }),
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      KarteKopieren: () =>
        kopiert.then(() => ({ karteID: '99', fehler: null })),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="7"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={bestandGeaendert}
        />
      </Sprachraum>,
    )

    await screen.findByText('Aus A')

    // **Die Handlungen der Karte stehen in ihrem Menü** (#1240, #1413).
    await oeffneKartenhandlungen()

    await nutzer.click(
      screen.getByRole('menuitem', { name: texte.seiten.karten.duplizieren }),
    )

    rerender(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="12"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={bestandGeaendert}
        />
      </Sprachraum>,
    )

    await screen.findByText('Aus B')

    // **Nach dem Setwechsel steht die Lasche neu** — auch das Menü.
    await oeffneKartenhandlungen()

    bestandGeaendert.mockClear()

    freigeben?.()

    await waitFor(() => {
      expect(
        screen.getByRole('menuitem', { name: texte.seiten.karten.duplizieren }),
      ).not.toHaveAttribute('aria-disabled')
    })

    expect(bestandGeaendert).not.toHaveBeenCalled()
  })
})
