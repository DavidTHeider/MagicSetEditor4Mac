// **Anlage und Ansicht in einer Zeile, als zwei Gruppen** — P03
// (#2174).
//
// Über dem Arbeitsplatz standen „Karte anlegen" und „Mehr", darunter in
// einer eigenen Zeile die Schalter für Navigator und Inspector. Jetzt bilden
// beide Paare je eine benannte Gruppe in **einer** Zeile — die Schalter
// gehören weiter dem Arbeitsplatz (`Arbeitsplatzschalter`, `schalterort`).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { Umschlag } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { Karten } from './Karten'

const texte = kataloge.de.seiten.karten

let raeumeBreite: () => void = () => undefined

beforeEach(() => {
  raeumeBreite = gibJsdomEineBreite(1400)
})

afterEach(() => {
  raeumeBreite()
  raeumeApi()
})

// bestand ist, was im Bestand steht — ein anderes Programm ändert ihn.
function mitBestand(bestand: Record<string, string>, scheitert = false) {
  const gesetzt = vi.fn<
    (
      setID: string,
      karteID: string,
      feld: string,
      wert: string,
    ) => Promise<Umschlag>
  >(() =>
    Promise.resolve({
      fehler: scheitert
        ? {
            code: 'konflikt',
            schluessel: 'konflikt',
            meldung: 'Dieses Feld hat sich inzwischen geändert.',
          }
        : null,
    }),
  )

  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Kartenzeilen: () =>
      Promise.resolve({
        zeilen: [{ id: '5', nummer: 0, name: 'Feuerball', felder: 2 }],
      }),
    Karte: () =>
      Promise.resolve({
        karte: {
          id: '5',
          nummer: 0,
          werte: Object.entries(bestand).map(([feld, wert]) => ({
            feld,
            wert,
          })),
        },
        fehler: null,
      }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
    Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
    KarteVerweis: () =>
      Promise.resolve({ verweis: null, befunde: [], fehler: null }),
    FeldSetzen: gesetzt,
  })

  return gesetzt
}

function seite(stand: number) {
  return (
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID="5"
        verlaufsstand={stand}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>
  )
}

describe('Anlage und Ansicht über den Karten', () => {
  it('stehen als zwei benannte Gruppen in einer Zeile', async () => {
    mitBestand({ name: 'Feuerball' })
    render(seite(0))

    const anlage = await screen.findByRole('group', {
      name: texte.gruppeAnlage,
    })
    expect(
      within(anlage).getByRole('button', { name: texte.anlegen }),
    ).toBeInTheDocument()
    expect(
      within(anlage).getByRole('button', { name: texte.mehr }),
    ).toBeInTheDocument()

    const ansicht = await screen.findByRole('group', {
      name: texte.gruppeAnsicht,
    })
    expect(
      await within(ansicht).findByRole('button', { name: /Navigator/ }),
    ).toBeInTheDocument()
    expect(
      within(ansicht).getByRole('button', { name: /Inspector/ }),
    ).toBeInTheDocument()

    // **Eine Zeile**: beide Gruppen haben denselben Elternknoten.
    expect(anlage.parentElement).toBe(ansicht.parentElement)
  })

  it('führen Tab von „Karte anlegen" über „Mehr" zu den Schaltern', async () => {
    mitBestand({ name: 'Feuerball' })
    render(seite(0))

    const ansicht = await screen.findByRole('group', {
      name: texte.gruppeAnsicht,
    })
    await within(ansicht).findByRole('button', { name: /Navigator/ })

    screen.getByRole('button', { name: texte.anlegen }).focus()
    await userEvent.tab()
    expect(document.activeElement).toHaveAccessibleName(texte.mehr)
    await userEvent.tab()
    expect(document.activeElement?.textContent).toMatch(/Navigator/)
    await userEvent.tab()
    expect(document.activeElement?.textContent).toMatch(/Inspector/)
  })
})
