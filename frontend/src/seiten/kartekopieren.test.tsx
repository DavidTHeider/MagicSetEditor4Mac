// **Eine Karte kopieren**
// (#625).
//
// **Ohne Rückfrage, anders als beim Löschen.** Eine Kopie zu viel ist kein
// Verlust — sie steht da, und ⌘Z nimmt sie zurück. Wer hier fragte, fragte vor
// jeder folgenlosen Handlung.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO } from '../api'
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

function karte(id: string, name: string): KarteDTO {
  return { id, nummer: Number(id) - 1, werte: [{ feld: 'name', wert: name }] }
}

function zeige(
  teile: Partial<Api>,
  bestand: () => KarteDTO[] = () => [karte('1', 'Erste')],
) {
  belegeApi({
    ...ausKarten(() => Promise.resolve({ karten: bestand() })),
    Karten: () => Promise.resolve({ karten: bestand() }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    Probevorschau: () => Promise.resolve({ vorschau: null }),
    // **Ein eigener `Karten`-Beleg wird mitprojiziert** (#891). Die Seite holt
    // `Kartenzeilen` und `Karte`; ohne diese Zeile bliebe die Vorgabe darüber
    // stehen, und der Test prüfte Karten, die er nicht beschrieben hat.
    ...(teile.Karten ? ausKarten(teile.Karten, '7') : {}),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="7"
        karteID="1"
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

describe('Karte kopieren', () => {
  it('kopiert und schlägt die Kopie auf', async () => {
    const KarteKopieren = vi.fn(() =>
      Promise.resolve({ karteID: '2', fehler: null }),
    )

    let bestand = [karte('1', 'Erste')]

    zeige({ KarteKopieren }, () => bestand)

    await oeffneKartenhandlungen()

    const kopieren = await screen.findByRole('menuitem', {
      name: texte.seiten.karten.duplizieren,
    })

    bestand = [karte('1', 'Erste'), karte('2', 'Erste')]

    await userEvent.click(kopieren)

    expect(KarteKopieren).toHaveBeenCalledWith('7', '1')

    // **Aufgeschlagen ist die Kopie**, nicht das Original — so hält es MSE2:
    // nach dem Einfügen wählt es die neue Karte
    // (`src/gui/control/card_list.cpp:72–77`, Fassung f9d9356 vom 2021-07-06).
    await waitFor(() => {
      expect(
        screen.getByRole('button', {
          name: `Erste — ${texte.seiten.karten.bearbeiten}`,
          current: true,
        }),
      ).toBeInTheDocument()
    })
  })

  it('meldet einen Befund an der Stelle, an der geklickt wurde', async () => {
    const KarteKopieren = vi.fn(() =>
      Promise.resolve({
        karteID: '',
        fehler: {
          code: 'nichtGefunden' as const,
          schluessel: 'nichtGefunden',
          meldung: 'Die Karte gibt es nicht.',
        },
      }),
    )

    zeige({ KarteKopieren })

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.duplizieren,
      }),
    )

    expect(
      await screen.findByText(texte.fehler.nichtGefunden),
    ).toBeInTheDocument()
  })

  // **Regel 4 aus §7: ein Vorgang je Datensatz.** Ein zweiter Klick, während
  // der erste läuft, legte zwei Kopien an, wo jemand eine wollte.
  it('legt bei zwei Klicks nur eine Kopie an', async () => {
    let loesen = () => {
      // bis der Aufruf sie ersetzt: die Zusage steht noch nicht
    }

    const KarteKopieren = vi.fn(
      () =>
        new Promise<{ karteID: string; fehler: null }>((erfuellen) => {
          loesen = () => {
            erfuellen({ karteID: '2', fehler: null })
          }
        }),
    )

    zeige({ KarteKopieren })

    await oeffneKartenhandlungen()

    const kopieren = await screen.findByRole('menuitem', {
      name: texte.seiten.karten.duplizieren,
    })

    await userEvent.click(kopieren)
    await userEvent.click(kopieren)

    expect(KarteKopieren).toHaveBeenCalledTimes(1)

    loesen()
  })

  // **Regel 1 aus §7 über den Verlauf**: wer die angelegte Karte aufgeschlagen
  // hat und sie mit ⌘Z entfernt, bekommt den offenen Kontext geräumt
  // (#626).
  // Der Weg dahin ist der Verlaufsstand: er steigt nach jeder geglückten
  // Rücknahme, die Liste wird neu geladen, und eine Wahl auf eine Karte, die
  // es nicht mehr gibt, findet niemanden.
  it('räumt den Editor, wenn ⌘Z die aufgeschlagene Karte entfernt', async () => {
    let bestand = [karte('1', 'Erste'), karte('2', 'Zweite')]

    belegeApi({
      ...ausKarten(() => Promise.resolve({ karten: bestand })),
      Karten: () => Promise.resolve({ karten: bestand }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
      Probevorschau: () => Promise.resolve({ vorschau: null }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="7"
          karteID="2"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    expect(await screen.findByDisplayValue('Zweite')).toBeInTheDocument()

    bestand = [karte('1', 'Erste')]

    rerender(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="7"
          karteID="2"
          verlaufsstand={1}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.queryByDisplayValue('Zweite')).not.toBeInTheDocument()
    })

    // **Gewartet und nicht gegriffen** (#1321). Zwischen „das Feld ist weg"
    // und „der leere Zustand steht da" liegt eine Antwort: die Liste wird neu
    // geholt. Bis zum 2026-09-02 antwortete `test/api.ts` synchron, und die
    // beiden fielen zusammen.
    expect(
      await screen.findByText(texte.seiten.karten.keineKarte),
    ).toBeInTheDocument()
  })

  it('steht nicht da, solange keine Karte aufgeschlagen ist', async () => {
    zeige({ Karten: () => Promise.resolve({ karten: [] }) })

    expect(
      await screen.findByText(texte.seiten.karten.leerTitel),
    ).toBeInTheDocument()

    // **Ohne Karte gibt es auch den Aufklapp nicht** — er gehört zu ihr
    // (#1240).
    expect(
      screen.queryByRole('button', { name: texte.seiten.karten.dieseKarte }),
    ).not.toBeInTheDocument()

    expect(
      screen.queryByRole('button', { name: texte.seiten.karten.duplizieren }),
    ).not.toBeInTheDocument()
  })
})
