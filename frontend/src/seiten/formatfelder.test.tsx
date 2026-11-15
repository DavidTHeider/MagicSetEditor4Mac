// **Was die Formatzeile über eine Eingabe urteilt** — Werkstück U3 von Phase
// 5.1 (#1414).
//
// # Die Regeln stehen in `internal/domain/format.go` und nicht hier
//
//	{"mindestgroesse", f.Mindestgroesse}, … if zahl.wert < 0 { … }
//	if f.Hoechstgroesse > 0 && f.Hoechstgroesse < f.Mindestgroesse { … }
//
// **Eine Obergrenze gibt es nicht**, und **null ist keine Grenze** — beides
// nachgelesen und beides hier festgehalten.
//
// **Eine eigene Datei**, damit `formatzeile.test.tsx` unter der Zeilengrenze
// aus §2 bleibt.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { baueFenster, warteAufDieFormatzeile } from './formatwerkbank'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { formatprobe } from './deckprobe'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

// **Die Zahlen eines Formats melden sich am Feld** — Werkstück U3 von Phase
// 5.1 (#1414).
//
// Die Regel steht in `domain.Format.zahlenPruefen`: jede der vier Zahlen wird
// gegen `< 0` geprüft („Die Zahl ist %d — negativ gibt es nicht."). **Eine
// Obergrenze gibt es nicht**, und `Formatzeile` hatte bis heute ein
// `inputMode="numeric"` ohne jede Meldung: `Number.parseInt('-5')` gab `-5`,
// und der Befund kam als anonymer Satz über der Liste zurück.
describe('was die Formatzeile am Feld prüft', () => {
  it('meldet eine negative Zahl am Feld und schickt sie nicht', async () => {
    const FormatSetzen = await baueFenster(formatprobe())

    const bereich = await warteAufDieFormatzeile()

    const feld = within(bereich).getByLabelText(texte.mindestgroesse)

    await userEvent.clear(feld)
    await userEvent.type(feld, '-5')

    expect(feld).toHaveAttribute('aria-invalid', 'true')

    await userEvent.click(
      within(bereich).getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    await waitFor(() => {
      expect(FormatSetzen).toHaveBeenCalledTimes(1)
    })

    // **Der abgewiesene Wert erreicht den Dienst nicht** (§7, Regel 3).
    expect(FormatSetzen.mock.calls[0]?.[0].mindestgroesse).toBe(60)
  })

  // **Null ist keine Grenze und deshalb ein gültiger Wert** — bei drei der
  // vier Zahlen (`domain.PruefeGroessen`: `format.Mindestgroesse > 0 && …`),
  // beim Sideboard heißt sie „keines" (CR 903.5e). Eine neue Prüfung, die sie
  // abwiese, machte jedes Commander-Format unspeicherbar.
  it('lässt die Null durch — sie heißt „keine Grenze“', async () => {
    const FormatSetzen = await baueFenster(formatprobe())

    const bereich = await warteAufDieFormatzeile()

    const feld = within(bereich).getByLabelText(texte.hoechstgroesse)

    await userEvent.clear(feld)
    await userEvent.type(feld, '0')

    expect(feld).toHaveAttribute('aria-invalid', 'false')

    await userEvent.click(
      within(bereich).getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    await waitFor(() => {
      expect(FormatSetzen).toHaveBeenCalledTimes(1)
    })

    expect(FormatSetzen.mock.calls[0]?.[0].hoechstgroesse).toBe(0)
  })

  // **Ein Format hat einen Namen** — `domain.Format.Validate`, Feld `name`,
  // Schlüssel `SchluesselPflichtfeld`. Gemeldet wird, wer das Feld anfasst
  // und leer verlässt; ein Formular, das beim Aufschlagen anklagt, stumpft ab.
  it('meldet den leeren Namen am Feld', async () => {
    await baueFenster(formatprobe())
    await warteAufDieFormatzeile()

    // **Die Zeile und nicht der Bereich**: das Anlegeformular daneben hat
    // ebenfalls ein Feld „Name", und eine Abfrage über beide träfe zwei.
    const zeile = screen
      .getByRole('button', { name: `${texte.formatSichern}: Konstruiert` })
      .closest('div')

    if (zeile === null) {
      throw new Error('die Formatzeile fehlt')
    }

    const feld = within(zeile).getByLabelText(texte.name)

    await userEvent.clear(feld)
    await userEvent.tab()

    expect(feld).toHaveAttribute('aria-invalid', 'true')
    expect(within(zeile).getByRole('alert')).toBeInTheDocument()
  })
  // **Ein Feldfehler aus Go steht am genannten Feld** (§6). Der Befund über
  // **zwei** Felder — `Hoechstgroesse > 0 && < Mindestgroesse` — lässt sich
  // am einzelnen Feld nicht entscheiden; Go nennt `hoechstgroesse`, und
  // `verursachendesFeld` liest das.
  it('hängt den Feldfehler aus Go an das genannte Feld', async () => {
    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
      FormatSetzen: () =>
        Promise.resolve({
          format: null,
          fehler: {
            code: 'validierung' as const,
            schluessel: 'fehler.validierung',
            feld: 'hoechstgroesse',
            meldung: 'Die Obergrenze 40 liegt unter der Untergrenze 60.',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.formateKnopf }),
    )

    const bereich = await warteAufDieFormatzeile()

    await userEvent.click(
      within(bereich).getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    const feld = within(bereich).getByLabelText(texte.hoechstgroesse)

    await waitFor(() => {
      expect(feld).toHaveAttribute('aria-invalid', 'true')
    })

    expect(feld).toHaveAccessibleDescription(
      'Die Obergrenze 40 liegt unter der Untergrenze 60.',
    )
  })

  // **Was zu keinem Feld gehört, steht in der Zeile** — ein Konflikt über die
  // Fassung nennt keines, und ein Satz ohne Adresse gehört wenigstens an den
  // Datensatz, um den es geht (§6).
  it('zeigt einen Fehler ohne Feld in der Zeile', async () => {
    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
      FormatSetzen: () =>
        Promise.resolve({
          format: null,
          fehler: {
            code: 'konflikt' as const,
            schluessel: 'fehler.konflikt',
            meldung: 'Die Fassung stimmt nicht.',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.formateKnopf }),
    )

    const bereich = await warteAufDieFormatzeile()

    await userEvent.click(
      within(bereich).getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    expect(await within(bereich).findByRole('alert')).toHaveTextContent(
      'Die Fassung stimmt nicht.',
    )

    // **Und kein Feld trägt ihn** — ein Befund an jedem Feld wäre dieselbe
    // Auskunft wie gar keiner.
    expect(
      within(bereich).getByLabelText(texte.hoechstgroesse),
    ).toHaveAttribute('aria-invalid', 'false')
  })
})
