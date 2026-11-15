// **Der Maßstab der Vorschau ist eine Auswahl** — UXR-06
// (#1914).
//
// # Warum hier keine Knopfleiste mehr steht
//
// Bis UXR-06 wählte eine Leiste aus sechs Knöpfen die Größe (U5, #1416;
// die Läufe dazu standen in `zoomleiste.test.tsx`). Gemessen am nativen
// Fenster waren das sechs der 27 Bedienelemente im Bild der Kartenseite —
// R2 aus `docs/UIREVIEW.md` erlaubt zwanzig —, und die gedrückte Stufe trug
// das Hauptgewicht einer führenden Handlung. **R7 sieht für mehr als vier
// Werte eine Auswahl vor**, und das Artboard *Karten* zeichnet genau eine.
//
// Was die Leiste zusagte, sagt die Auswahl weiter zu: sechs Stufen, die
// Prozentzahl nach der Sprache, eine Tabstelle, und die Druckgröße nur, wenn
// die Vorlage eine Auflösung nennt.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Massstab } from './Massstab'

const worte = kataloge.de.seiten.karten.zoom

describe('Der Maßstab der Vorschau', () => {
  // **Die Stufen folgen der Sprache** — LOC-03.AK-01
  // (#1689):
  // das Englische schreibt kein Leerzeichen vor das Zeichen, das Deutsche
  // ein geschütztes.
  it.each([
    ['de', '150\u00a0%'],
    ['en', '150%'],
  ] as const)('%s beschriftet eine Stufe als %j', (sprache, name) => {
    render(
      <Sprachraum sprache={sprache}>
        <Massstab zoom="hundert" waehlen={vi.fn()} druckMoeglich />
      </Sprachraum>,
    )

    expect(screen.getByRole('option', { name })).toBeInTheDocument()
  })

  it('ist eine Auswahl mit sechs Stufen, die gewählte steht darin', async () => {
    const waehlen = vi.fn()
    render(
      <Sprachraum sprache="de">
        <Massstab zoom="hundert" waehlen={waehlen} druckMoeglich />
      </Sprachraum>,
    )

    const auswahl = screen.getByRole('combobox', { name: worte.bezeichnung })
    expect(within(auswahl).getAllByRole('option')).toHaveLength(6)
    expect(auswahl).toHaveDisplayValue('100\u00a0%')

    // **Keine Knöpfe mehr** — und damit kein Hauptgewicht für eine Stufe.
    expect(screen.queryAllByRole('button')).toHaveLength(0)

    // **Gewählt wird über den Wert**: `selectOptions` vergleicht Aufschriften
    // als HTML, und das geschützte Leerzeichen steht dort als `&nbsp;`.
    await userEvent.selectOptions(auswahl, '200')
    expect(waehlen).toHaveBeenLastCalledWith(200)

    await userEvent.selectOptions(auswahl, worte.druck)
    expect(waehlen).toHaveBeenLastCalledWith('druck')

    await userEvent.selectOptions(auswahl, worte.anpassen)
    expect(waehlen).toHaveBeenLastCalledWith('anpassen')

    await userEvent.selectOptions(auswahl, 'hundert')
    expect(waehlen).toHaveBeenLastCalledWith('hundert')
  })

  it('hält die Druckgröße an, wenn die Vorlage keine Auflösung nennt, und sagt warum', () => {
    render(
      <Sprachraum sprache="de">
        <Massstab zoom="anpassen" waehlen={vi.fn()} druckMoeglich={false} />
      </Sprachraum>,
    )

    expect(screen.getByRole('option', { name: worte.druck })).toBeDisabled()

    // **Der Grund steht an der Auswahl** — als ihre Beschreibung, damit ein
    // Vorleseprogramm ihn mit ihr liest.
    expect(
      screen.getByRole('combobox', { name: worte.bezeichnung }),
    ).toHaveAccessibleDescription(worte.ohneDruck)
  })

  it('beschreibt nichts, solange die Druckgröße möglich ist', () => {
    render(
      <Sprachraum sprache="de">
        <Massstab zoom="druck" waehlen={vi.fn()} druckMoeglich />
      </Sprachraum>,
    )

    const auswahl = screen.getByRole('combobox', { name: worte.bezeichnung })
    expect(auswahl).toHaveDisplayValue(worte.druck)
    expect(auswahl).not.toHaveAttribute('aria-describedby')
    expect(screen.queryByText(worte.ohneDruck)).not.toBeInTheDocument()
  })
})
