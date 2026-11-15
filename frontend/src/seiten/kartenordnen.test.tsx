// **Karten umordnen**
// (#628).
//
// Getauscht wird mit dem Nachbarn — MSEs Elementarschritt beim Ziehen. Warum
// die Referenz ihn nicht ausführt, steht bei `storage.KartenTauschen`.
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

function zeige(teile: Partial<Api>, karteID = '2') {
  belegeApi({
    ...ausKarten(() =>
      Promise.resolve({
        karten: [
          karte('1', 'Erste'),
          karte('2', 'Zweite'),
          karte('3', 'Dritte'),
        ],
      }),
    ),
    Karten: () =>
      Promise.resolve({
        karten: [
          karte('1', 'Erste'),
          karte('2', 'Zweite'),
          karte('3', 'Dritte'),
        ],
      }),
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
        karteID={karteID}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

describe('Karten umordnen', () => {
  it('tauscht mit der Karte darüber', async () => {
    const KartenTauschen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({ KartenTauschen })

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.nachOben,
      }),
    )

    expect(KartenTauschen).toHaveBeenCalledWith('7', '2', '1')
  })

  it('tauscht mit der Karte darunter', async () => {
    const KartenTauschen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({ KartenTauschen })

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.nachUnten,
      }),
    )

    expect(KartenTauschen).toHaveBeenCalledWith('7', '2', '3')
  })

  // **An den Rändern sagt der Knopf nein, statt es zu tun.** Angehalten und
  // nicht verschwunden: eine Knopfreihe, die je nach gewählter Karte springt,
  // ist schwerer zu bedienen als eine, die stillsteht.
  it('hält an, wo es nicht weitergeht', async () => {
    const KartenTauschen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({ KartenTauschen }, '1')

    await oeffneKartenhandlungen()

    expect(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.nachOben,
      }),
    ).toHaveAttribute('aria-disabled', 'true')

    expect(
      screen.getByRole('menuitem', { name: texte.seiten.karten.nachUnten }),
    ).not.toHaveAttribute('aria-disabled')
  })

  it('meldet einen Befund an der Stelle, an der geklickt wurde', async () => {
    const KartenTauschen = vi.fn(() =>
      Promise.resolve({
        fehler: {
          code: 'validierung' as const,
          schluessel: 'validierung',
          meldung: 'Das geht so nicht.',
        },
      }),
    )

    zeige({ KartenTauschen })

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.nachOben,
      }),
    )

    expect(
      await screen.findByText(texte.fehler.validierung),
    ).toBeInTheDocument()
  })

  // **Regel 4 aus §7**: ein zweiter Klick, während der erste läuft, ginge
  // gegen eine Ordnung, die der Bestand noch nicht kennt.
  it('tauscht bei zwei Klicks nur einmal', async () => {
    let loesen = () => {
      // bis der Aufruf sie ersetzt: die Zusage steht noch nicht
    }

    const KartenTauschen = vi.fn(
      () =>
        new Promise<{ fehler: null }>((erfuellen) => {
          loesen = () => {
            erfuellen({ fehler: null })
          }
        }),
    )

    zeige({ KartenTauschen })

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.nachOben,
      }),
    )

    // **Das Menü ist mit der Wahl zugegangen** — und beim zweiten Öffnen
    // ruht der Eintrag, solange der Tausch läuft (§7, Regel 4).
    await oeffneKartenhandlungen()
    const oben = screen.getByRole('menuitem', {
      name: texte.seiten.karten.nachOben,
    })
    expect(oben).toHaveAttribute('aria-disabled', 'true')
    await userEvent.click(oben)
    expect(KartenTauschen).toHaveBeenCalledTimes(1)

    loesen()

    await waitFor(() => {
      expect(oben).not.toHaveAttribute('aria-disabled')
    })
  })
})
