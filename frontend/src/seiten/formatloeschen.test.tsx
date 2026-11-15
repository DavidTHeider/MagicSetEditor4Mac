// **Ein Format zu löschen fragt vorher** — UX-10.AK-04
// (#1682).
//
// Aus `formatverwaltung.test.tsx` herausgelöst, als die Datei über 400 Zeilen
// wuchs (§2). Der Schnitt folgt der Frage: dort steht, **was ein Format ist**
// — anlegen, ändern, die vier Zahlen; hier, **was vor dem Löschen steht**.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe, formatprobe } from './deckprobe'
import { bearbeite } from './formatwerkbank'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

// zurFormatlasche schlägt die Formate auf.
async function zurFormatlasche() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.formateKnopf }),
  )

  return await screen.findByRole('region', { name: texte.formate })
}

describe('Ein Format löschen', () => {
  // zeigeFormate stellt die Formatlasche mit einem Format hin.
  const zeigeFormate = async (
    FormatLoeschen: (formatID: string) => Promise<{ fehler: null }>,
  ) => {
    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      FormatLoeschen,
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await zurFormatlasche()
    // **Löschen steht im Formular des Formats** — seit UXR-11 (#1919) erst
    // auf Wunsch.
    await bearbeite()
  }

  it('fragt, bevor es löscht', async () => {
    const FormatLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))

    await zeigeFormate(FormatLoeschen)

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.formatLoeschen}: Konstruiert`,
      }),
    )

    expect(
      await screen.findByText(
        einsetzen(texte.formatLoeschenFrage, { name: 'Konstruiert' }),
      ),
    ).toBeVisible()

    expect(FormatLoeschen).not.toHaveBeenCalled()
  })

  it('löscht erst nach dem Bestätigen', async () => {
    const FormatLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))

    await zeigeFormate(FormatLoeschen)

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.formatLoeschen}: Konstruiert`,
      }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: texte.formatLoeschenJa }),
    )

    await waitFor(() => {
      expect(FormatLoeschen).toHaveBeenCalledWith('5')
    })
  })

  // **Abbrechen lässt das Format stehen** und räumt die Frage weg (§7,
  // Regel 1). Ohne diesen Lauf bliebe der Zweig unbetreten — und ein
  // unbetretener Zweig ist entweder ein fehlender Test oder toter Code.
  it('lässt das Format stehen, wenn jemand abbricht', async () => {
    const FormatLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))

    await zeigeFormate(FormatLoeschen)

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.formatLoeschen}: Konstruiert`,
      }),
    )
    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.aktionen.abbrechen,
      }),
    )

    expect(
      await screen.findByRole('button', {
        name: `${texte.formatLoeschen}: Konstruiert`,
      }),
    ).toBeVisible()

    expect(FormatLoeschen).not.toHaveBeenCalled()
  })
})
