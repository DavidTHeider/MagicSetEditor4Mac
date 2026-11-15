// **Die Kartenliste lässt sich wegklappen**
// (#697).
//
// Gemessen an der laufenden Anwendung bei 1.440 Punkten Fensterbreite: die
// Liste nimmt 224, und was bleibt, teilen sich Formular und Vorschau zu je
// 407. Für zwei Formularspalten reicht das nicht — wer die Liste wegklappt,
// bekommt sie.
//
// **Nicht gemerkt.** Der Zustand lebt so lange wie die Lasche. Wo er sonst
// hingehörte, sagt #245:
// in die Einstellungen des Bestands — nicht in einen `localStorage`, den keine
// Sicherung mitnimmt. Ein Schalter für eine Spalte ist dafür zu wenig.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
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

function zeige() {
  belegeApi({
    ...ausKarten(() =>
      Promise.resolve({
        karten: [karte('8', 0, 'Erste'), karte('9', 1, 'Zweite')],
      }),
    ),
    Karten: () =>
      Promise.resolve({
        karten: [karte('8', 0, 'Erste'), karte('9', 1, 'Zweite')],
      }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    Probevorschau: () => Promise.resolve({ vorschau: null }),
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
  })

  return render(
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
}

describe('Die Kartenliste', () => {
  it('steht offen und lässt sich wegklappen und wiederholen', async () => {
    const nutzer = userEvent.setup()
    zeige()

    const schalter = await screen.findByRole('button', {
      name: texte.bausteine.arbeitsplatz.navigatorZu,
    })

    expect(
      screen.getByRole('complementary', {
        name: texte.seiten.karten.listeBezeichnung,
      }),
    ).toBeInTheDocument()
    expect(schalter).toHaveAttribute('aria-expanded', 'true')

    await nutzer.click(schalter)

    expect(
      screen.queryByRole('complementary', {
        name: texte.seiten.karten.listeBezeichnung,
      }),
    ).not.toBeInTheDocument()

    // **Der Beleg, dass jemand zuhört** (§3): derselbe Schalter führt zurück.
    const zurueck = screen.getByRole('button', {
      name: texte.bausteine.arbeitsplatz.navigatorAuf,
    })
    expect(zurueck).toHaveAttribute('aria-expanded', 'false')

    await nutzer.click(zurueck)

    expect(
      screen.getByRole('complementary', {
        name: texte.seiten.karten.listeBezeichnung,
      }),
    ).toBeInTheDocument()
  })
})
