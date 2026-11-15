// **Die Ansicht der Vorschau übersteht das Schreiben** — UXR-06
// (#1914).
//
// # Der Befund
//
// Maßstab und Überlagerungen lebten im Karteneditor, und der wird **nach
// jedem Schreiben** neu gebaut: sein Schlüssel trägt den Verlaufsstand, und
// der Bestandswächter zählt ihn nach jedem Feld hoch (`Karten.tsx`, #1075).
// **Auch eine Ebene höher reichte nicht**: mit dem Verlaufsstand lädt die
// Liste neu, und `useLaden` räumt sie dabei — der ganze Arbeitsplatz steht
// für einen Herzschlag nicht da. Gehalten wird die Ansicht deshalb in
// `Karten`, neben der Feldwahl.
// Wer die Karte auf 200 % stellte, die Bildgrenzen einschaltete und dann
// einen Namen änderte, sah nach dem ⇥ wieder „Anpassen" und keine Linie —
// dieselbe Klasse wie die Feldwahl in #1075, die deshalb eine Ebene höher
// liegt.
//
// **Gefahren wird die ganze Seite**, weil der Neuaufbau dort entsteht: ein
// Lauf gegen den Editor allein sähe den Schlüssel nicht.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { vorschau } from '../test/bildflaeche'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const karten: KarteDTO[] = [
  { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Feuerball' }] },
  { id: '2', nummer: 1, werte: [{ feld: 'name', wert: 'Mirri' }] },
]

function seite(verlaufsstand: number) {
  return (
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID={undefined}
        verlaufsstand={verlaufsstand}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>
  )
}

describe('Die Ansicht der Vorschau', () => {
  it('behält Maßstab und Überlagerungen, wenn ein Schreiben den Editor neu baut', async () => {
    const Vorschau = vi.fn(() => Promise.resolve(vorschau()))
    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      ...ausKarten(() => Promise.resolve({ karten })),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Verlauf: () => Promise.resolve({ schritte: [] }),
      Vorschau,
      Probevorschau: () => Promise.resolve(vorschau()),
    })

    const { rerender } = render(seite(0))

    await screen.findByRole('img', { name: 'Vorschau der Karte' })
    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: 'Größe der Vorschau' }),
      '200',
    )
    await userEvent.click(screen.getByRole('checkbox', { name: 'Bildgrenzen' }))

    await waitFor(() => {
      expect(Vorschau).toHaveBeenLastCalledWith('1', 2)
    })

    // **Das Schreiben**: der Bestandswächter zählt den Verlaufsstand hoch.
    rerender(seite(1))

    await screen.findByRole('img', { name: 'Vorschau der Karte' })
    expect(
      screen.getByRole('combobox', { name: 'Größe der Vorschau' }),
    ).toHaveDisplayValue('200\u00a0%')
    expect(screen.getByRole('checkbox', { name: 'Bildgrenzen' })).toBeChecked()
    expect(Vorschau).toHaveBeenLastCalledWith('1', 2)
  })

  // **Und sie bleibt stehen, während neu gerastert wird.** Ab 150 % fragt
  // der Editor den Go-Teil nach einem neuen Bild, und bis es da ist, gibt es
  // keines. Die Leiste hing am Bild: sie verschwand für einen Herzschlag, und
  // mit ihr der Fokus — wer mit der Tastatur den Maßstab wählte, stand danach
  // auf dem Dokument.
  it('bleibt stehen und behält den Fokus, während die Karte neu gerastert wird', async () => {
    let fertig: () => void = () => undefined
    const Vorschau = vi.fn((_karte: string, faktor: number) =>
      faktor === 1
        ? Promise.resolve(vorschau())
        : new Promise<ReturnType<typeof vorschau>>((erfuellen) => {
            fertig = () => {
              erfuellen(vorschau())
            }
          }),
    )
    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      ...ausKarten(() => Promise.resolve({ karten })),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Verlauf: () => Promise.resolve({ schritte: [] }),
      Vorschau,
      Probevorschau: () => Promise.resolve(vorschau()),
    })

    render(seite(0))
    await screen.findByRole('img', { name: 'Vorschau der Karte' })

    const massstab = screen.getByRole('combobox', {
      name: 'Größe der Vorschau',
    })
    await userEvent.selectOptions(massstab, '200')

    // **Das neue Bild ist unterwegs** — und die Auswahl steht, mit Fokus.
    await waitFor(() => {
      expect(Vorschau).toHaveBeenLastCalledWith('1', 2)
    })
    expect(massstab).toBeInTheDocument()
    expect(massstab).toHaveFocus()

    fertig()
    await screen.findByRole('img', { name: 'Vorschau der Karte' })
    expect(massstab).toHaveDisplayValue('200\u00a0%')
  })
})
