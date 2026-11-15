// **Die Formate am Fenster** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `deckurteil.test.tsx` unter der Zeilengrenze
// aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, was man über ein
// Deck **erfährt**, hier, wie ein **Format** entsteht und sich ändert.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { bauartprobe, formatprobe } from './deckprobe'
import { bearbeite } from './formatwerkbank'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

// zurFormatlasche wählt das Gebiet »Formate« und gibt seinen Bereich.
//
// **Seit Werkstück I von Phase 4.10 liegen die Formate hinter einer Lasche**
// (#1325).
// Der Test geht damit denselben Weg wie ein Anwender — vorher stand alles
// zugleich auf einer 1.614 Punkte hohen Seite.
async function zurFormatlasche() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.formateKnopf }),
  )

  return await screen.findByRole('region', { name: texte.formate })
}

describe('die Formate am Fenster', () => {
  it('legt ein Format aus einer Bauart an', async () => {
    const FormatAusBauart = vi.fn(() =>
      Promise.resolve({ format: formatprobe(), fehler: null }),
    )

    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () => Promise.resolve({ formate: [], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      FormatAusBauart,
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    const bereich = await zurFormatlasche()

    await userEvent.type(
      within(bereich).getByLabelText(texte.name),
      'Konstruiert',
    )
    await userEvent.type(
      within(bereich).getByLabelText(texte.scryfallname),
      'standard',
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.formatNeu }),
    )

    await waitFor(() => {
      expect(FormatAusBauart).toHaveBeenCalledWith(
        'konstruiert',
        'Konstruiert',
        'magic',
        'standard',
      )
    })
  })

  it('nimmt eine andere Bauart und ein anderes Spiel', async () => {
    const FormatAusBauart = vi.fn(() =>
      Promise.resolve({ format: formatprobe(), fehler: null }),
    )

    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () => Promise.resolve({ formate: [], fehler: null }),
      Bauarten: () =>
        Promise.resolve({
          bauarten: [
            bauartprobe(),
            { bauart: 'commander', format: formatprobe() },
          ],
          fehler: null,
        }),
      FormatAusBauart,
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    const bereich = await zurFormatlasche()

    // **Erst muss die Bauartenliste da sein.** Der Bereich trägt seinen Namen
    // schon, während sie noch unterwegs ist — die Auswahl steht dann mit
    // ihrem Platzhalter da, und `selectOptions` scheitert an „Value
    // \"commander\" not found in options" (§3, #1321).
    const bauart = within(bereich).getByLabelText(texte.bauart)

    await waitFor(() => {
      expect(within(bauart).getAllByRole('option').length).toBeGreaterThan(1)
    })

    await userEvent.selectOptions(bauart, 'commander')
    await userEvent.type(within(bereich).getByLabelText(texte.name), 'C')
    await userEvent.clear(within(bereich).getByLabelText(texte.spiel))
    await userEvent.type(within(bereich).getByLabelText(texte.spiel), 'vs')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.formatNeu }),
    )

    await waitFor(() => {
      expect(FormatAusBauart).toHaveBeenCalledWith('commander', 'C', 'vs', '')
    })
  })

  it('speichert und löscht ein Format', async () => {
    const FormatSetzen = vi.fn(() =>
      Promise.resolve({ format: formatprobe(), fehler: null }),
    )
    const FormatLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      FormatSetzen,
      FormatLoeschen,
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await zurFormatlasche()
    await bearbeite()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    await waitFor(() => {
      expect(FormatSetzen).toHaveBeenCalledWith(formatprobe())
    })

    // **Auf den bedienbaren Knopf gewartet, nicht auf den Ruf** (#1809):
    // solange das Speichern läuft, ruht die Zeile.
    await waitFor(() => {
      expect(
        screen.getByRole('button', {
          name: `${texte.formatLoeschen}: Konstruiert`,
        }),
      ).toBeEnabled()
    })

    // **Seit UX-10.AK-04 liegt eine Rückfrage dazwischen.**
    await userEvent.click(
      screen.getByRole('button', {
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

  it('ändert die vier Zahlen und schickt sie mit', async () => {
    // **Befund 8 des Zweitblicks über Phase 4.6.** Vorher zeigte die Liste
    // die Zahlen nur an, und „Speichern" reichte dasselbe DTO unverändert
    // zurück — eine Fähigkeit, die die Dokumentation behauptete und das
    // Fenster nicht hatte.
    const FormatSetzen = vi.fn(() =>
      Promise.resolve({ format: formatprobe(), fehler: null }),
    )

    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      FormatSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await zurFormatlasche()
    await bearbeite()

    const zeile = (
      await screen.findByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      })
    ).closest('div')

    if (zeile === null) {
      throw new Error('die Formatzeile fehlt')
    }

    const untergrenze = within(zeile).getByLabelText(texte.mindestgroesse)

    await userEvent.clear(untergrenze)
    await userEvent.type(untergrenze, '40')

    // **Und eine Null, die etwas bedeutet**: kein Sideboard.
    const sideboard = within(zeile).getByLabelText(texte.sideboardgroesse)

    await userEvent.clear(sideboard)
    await userEvent.type(sideboard, '0')

    await userEvent.click(
      within(zeile).getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    await waitFor(() => {
      expect(FormatSetzen).toHaveBeenCalledWith(
        formatprobe({ mindestgroesse: 40, sideboardgroesse: 0 }),
      )
    })
  })

  // **Eine unbrauchbare Eingabe wird nicht zur Null** — berichtigt in
  // Werkstück U3 von Phase 5.1
  // (#1414).
  //
  // # Was hier vorher stand
  //
  // „macht aus einer unbrauchbaren Eingabe eine Null" — `mitZahl` setzte für
  // alles, was `Number.parseInt` nicht las, eine `0`. Und **null bedeutet bei
  // diesen vier Feldern etwas**: keine Grenze, kein Sideboard, beliebig oft
  // (`domain.PruefeGroessen`). Wer sich vertippt, verschob damit stillschweigend
  // die Regeln eines Formats.
  //
  // Seit dem Umstieg auf `Zahlenfeld` gilt dieselbe Zusage wie überall sonst:
  // **was nicht gilt, erreicht den Aufrufer nicht** (§7, Regel 3) — und ein
  // leeres Feld ist kein Fehler, sondern ein halber Gedanke.
  it('macht aus einer unbrauchbaren Eingabe keine Null', async () => {
    // **Und zeigt sie auch nicht als eine.**
    const FormatSetzen = vi.fn(() =>
      Promise.resolve({ format: formatprobe(), fehler: null }),
    )

    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      FormatSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await zurFormatlasche()
    await bearbeite()

    const feld = await screen.findByLabelText(texte.hoechstzahlGleich)

    await userEvent.clear(feld)
    await userEvent.type(feld, 'viele')

    // **Ein `type="number"` nimmt keinen Text an**: das Feld bleibt leer, und
    // leer ist kein Fehler.
    expect(feld).toHaveValue(null)
    expect(feld).toHaveAttribute('aria-invalid', 'false')

    await userEvent.click(
      screen.getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    // **Und der Entwurf trägt weiter die alte Zahl** — nicht die Null.
    await waitFor(() => {
      expect(FormatSetzen).toHaveBeenCalledWith(formatprobe())
    })
  })

  it('ändert Name und Scryfall-Namen eines vorhandenen Formats', async () => {
    const FormatSetzen = vi.fn(() =>
      Promise.resolve({ format: formatprobe(), fehler: null }),
    )

    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      FormatSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await zurFormatlasche()
    await bearbeite()

    const zeile = (
      await screen.findByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      })
    ).closest('div')

    if (zeile === null) {
      throw new Error('die Formatzeile fehlt')
    }

    const name = within(zeile).getByLabelText(texte.name)

    await userEvent.clear(name)
    await userEvent.type(name, 'Pioneer')

    const scryfall = within(zeile).getByLabelText(texte.scryfallname)

    await userEvent.clear(scryfall)
    await userEvent.type(scryfall, 'pioneer')

    await userEvent.click(
      within(zeile).getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    await waitFor(() => {
      expect(FormatSetzen).toHaveBeenCalledWith(
        formatprobe({ name: 'Pioneer', scryfallname: 'pioneer' }),
      )
    })
  })
})
