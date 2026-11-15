// **Die Vervollständigung im Feld** — Werkstück 1 von Phase 3.3
// (#1164).
//
// **Was ein Wort ist und was dazu passt, prüft `wortstelle.test.ts`** — reine
// Rechnung, ohne Vorrichtung. Hier geht es um die Bedienung: erscheinen die
// Vorschläge beim Tippen, und setzt ein Klick den Namen an die richtige
// Stelle?
import { useState } from 'react'

import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Quelltextfeld } from './Quelltextfeld'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.bausteine.vorschlaege

// Gefuehrt hält den Wert, wie eine echte Seite es tut.
//
// **Ohne diesen Halter prüft nichts.** Das Feld ist geführt: was jemand tippt,
// kommt erst zurück, wenn der Aufrufer es einträgt. Ein `vi.fn()` allein lässt
// den Wert leer — und dann gibt es kein angefangenes Wort, zu dem etwas
// vorzuschlagen wäre. Vier Läufe dieser Datei waren daran zuerst rot.
function Gefuehrt({
  anfang,
  namen,
  merken,
}: {
  anfang: string
  namen?: readonly string[]
  merken: (wert: string) => void
}) {
  const [wert, setzeWert] = useState(anfang)

  return (
    <Quelltextfeld
      bezeichnung="Ausdruck"
      wert={wert}
      setzeWert={(neuer) => {
        setzeWert(neuer)
        merken(neuer)
      }}
      sprache="keine"
      namen={namen}
    />
  )
}

// feld stellt ein geführtes Quelltextfeld mit Namen.
function feld(
  anfang: string,
  namen: readonly string[] | undefined = [
    'card',
    'card_style',
    'to_upper',
    'to_lower',
  ],
) {
  belegeApi({ Skriptfarben: () => Promise.resolve({ merkmale: [] }) })

  const gesetzt = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Gefuehrt anfang={anfang} namen={namen} merken={gesetzt} />
    </Sprachraum>,
  )

  return gesetzt
}

describe('Die Vervollständigung', () => {
  afterEach(raeumeApi)

  it('schlägt vor, was das angefangene Wort fortsetzt', async () => {
    feld('')

    await userEvent.type(screen.getByLabelText('Ausdruck'), 'to_')

    expect(
      await screen.findByRole('button', { name: 'to_upper' }),
    ).toBeInTheDocument()

    expect(screen.getByRole('button', { name: 'to_lower' })).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'card' }),
    ).not.toBeInTheDocument()
  })

  // **Ohne angefangenes Wort steht dort nichts.**
  //
  // Die Gegenprobe: eine Liste, die immer offensteht, ist keine
  // Vervollständigung, sondern eine zweite Namensliste im Weg.
  it('schlägt ohne angefangenes Wort nichts vor', async () => {
    feld('')

    await userEvent.type(screen.getByLabelText('Ausdruck'), '1 + ')

    await waitFor(() => {
      expect(screen.queryByText(texte.bezeichnung)).not.toBeInTheDocument()
    })
  })

  it('setzt einen Vorschlag an die Stelle des Wortes', async () => {
    const gesetzt = feld('')

    await userEvent.type(screen.getByLabelText('Ausdruck'), 'to_u')
    await userEvent.click(
      await screen.findByRole('button', { name: 'to_upper' }),
    )

    expect(gesetzt).toHaveBeenCalledWith('to_upper')

    // **Und nur einmal.** Der Knopf hört auf `mousedown` **und** auf `click`;
    // eine Maus löst beides aus. Griffe beides, stünde der Name zweimal da —
    // die Zusicherung darüber sähe das nicht, weil sie nur nach *einem* Aufruf
    // mit diesem Wert fragt.
    expect(screen.getByLabelText('Ausdruck')).toHaveValue('to_upper')
  })

  // **Das Wort wird ersetzt, nicht angehängt.**
  //
  // Ohne diese Zusicherung sagte der Lauf darüber auch dann ja, wenn der Name
  // hinten drangehängt würde — `to_uto_upper` sähe im Fehlerfall ähnlich aus.
  it('ersetzt das angefangene Wort statt anzuhängen', async () => {
    const gesetzt = feld('1 + ca')

    const eingabe = screen.getByLabelText<HTMLTextAreaElement>('Ausdruck')

    await userEvent.click(eingabe)
    eingabe.setSelectionRange(6, 6)
    await userEvent.type(eingabe, 'r')

    await userEvent.click(await screen.findByRole('button', { name: 'card' }))

    expect(gesetzt).toHaveBeenCalledWith('1 + card')
  })

  // **Ohne Namen gibt es keine Vorschläge.**
  //
  // Auf einem Rechner ohne eingebundene Vorlage kennt die Umgebung `card`
  // nicht; eine Liste, die ihn trotzdem anböte, schlüge etwas vor, das nicht
  // läuft.
  it('schlägt ohne Namen nichts vor', async () => {
    // **Eine leere Liste und nicht `undefined`.** Ein weggelassenes Argument
    // nähme den Vorgabewert des Helfers — und der trägt Namen.
    feld('', [])

    await userEvent.type(screen.getByLabelText('Ausdruck'), 'to_')

    expect(screen.queryByText(texte.bezeichnung)).not.toBeInTheDocument()
  })

  // **Mit der Tastatur geht es auch.**
  //
  // # Warum das ein eigener Lauf ist
  //
  // Der Knopf greift beim **Drücken** statt beim Klick — sonst nähme er dem
  // Feld den Fokus, und dessen `onBlur` räumte die Liste, bevor der Knopf
  // gedrückt ist. Ein `mousedown` gibt es an der Tastatur aber nicht: ohne
  // `aufKlick` daneben wäre die Vervollständigung nur mit der Maus zu bedienen
  // (§6). Dieser Lauf hält den zweiten Weg fest.
  it('nimmt einen Vorschlag auch über die Tastatur an', async () => {
    const gesetzt = feld('')

    await userEvent.type(screen.getByLabelText('Ausdruck'), 'to_u')

    const vorschlag = await screen.findByRole('button', { name: 'to_upper' })

    vorschlag.focus()
    await userEvent.keyboard('{Enter}')

    expect(gesetzt).toHaveBeenCalledWith('to_upper')
  })

  // **Die Tabulatortaste führt vom Feld in die Vorschläge.**
  //
  // # Warum das eine eigene Zusicherung ist
  //
  // Die Leiste steht im Bild **über** dem Feld. Läge sie auch im Baum davor,
  // führte `Tab` vom Feld am nächsten Knopf vorbei und **nie** in die
  // Vorschläge — mit der Tastatur wären sie unerreichbar (§6). Im laufenden
  // Fenster gemessen: `Tab` landete auf »Auswerten«.
  //
  // Diese Zusicherung hält die Reihenfolge im Baum fest; die im Bild macht
  // eine CSS-Klasse. Wer sie vertauscht, sieht hier rot.
  it('führt mit der Tabulatortaste vom Feld in die Vorschläge', async () => {
    feld('')

    await userEvent.type(screen.getByLabelText('Ausdruck'), 'to_u')

    expect(await screen.findByText(texte.bezeichnung)).toBeInTheDocument()

    await userEvent.tab()

    expect(screen.getByRole('button', { name: 'to_upper' })).toHaveFocus()
  })

  // **Verlässt der Fokus die ganze Gruppe, gehen die Vorschläge.**
  //
  // Eine Liste neben einem Feld, das niemand mehr bearbeitet, ist Beiwerk.
  // **Die ganze Gruppe** — der Schritt in einen Vorschlag ist kein Weggehen,
  // und der Lauf darüber hält das fest.
  it('räumt die Vorschläge, wenn der Fokus die Gruppe verlässt', async () => {
    feld('')

    await userEvent.type(screen.getByLabelText('Ausdruck'), 'to_')

    expect(await screen.findByText(texte.bezeichnung)).toBeInTheDocument()

    await userEvent.click(document.body)

    await waitFor(() => {
      expect(screen.queryByText(texte.bezeichnung)).not.toBeInTheDocument()
    })
  })
})
