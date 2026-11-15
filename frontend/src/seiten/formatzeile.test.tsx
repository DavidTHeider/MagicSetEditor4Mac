// **Was die Formatzeile über zwei Bedienschritte hinweg tut** — Befunde 1
// und 3 des Nachlaufs über Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `formatverwaltung.test.tsx` unter der
// Zeilengrenze aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, wie
// ein Format **entsteht**, hier, was beim **zweiten** Handgriff passiert —
// und in `formatfelder.test.tsx`, was die Zeile über eine **Eingabe** urteilt.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { baueFenster, warteAufDieFormatzeile } from './formatwerkbank'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { formatprobe } from './deckprobe'
import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

describe('die Formatzeile über zwei Handgriffe', () => {
  // **Befund 3 des Nachlaufs.** `Formatzeile` übernahm das DTO samt Fassung
  // **einmal** in lokalen Zustand, und der `key` der Liste hing an der
  // Kennung — die sich beim Speichern nicht ändert. Also blieb der Entwurf
  // stehen, und der zweite Speichervorgang schickte die alte Fassung.
  //
  // Der bisherige Test änderte und speicherte **einmal**; die Wiederholung
  // derselben Bedienfolge fehlte.
  it('schickt beim zweiten Speichern die neue Fassung', async () => {
    const FormatSetzen = await baueFenster(formatprobe())

    const bereich = await warteAufDieFormatzeile()

    const speichern = () =>
      userEvent.click(
        within(bereich).getByRole('button', {
          name: `${texte.formatSichern}: Konstruiert`,
        }),
      )

    const feld = within(bereich).getByLabelText(texte.mindestgroesse)

    await userEvent.clear(feld)
    await userEvent.type(feld, '40')
    await speichern()

    await waitFor(() => {
      expect(FormatSetzen).toHaveBeenCalledTimes(1)
    })

    // **Der zweite Handgriff.** Er geht gegen den Stand, den der erste
    // hinterlassen hat — Fassung 2.
    //
    // **Erst nachdem die Antwort da war** (#1321): sie zeichnet die Zeile
    // neu, und ein vorher gegriffener Knoten ist danach abgehängt. `clear`
    // sagt dazu „The element to be cleared could not be focused."
    await antwortAbwarten()

    const wieder = screen.getByLabelText(texte.mindestgroesse)

    await userEvent.clear(wieder)
    await userEvent.type(wieder, '41')
    await userEvent.click(
      screen.getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    await waitFor(() => {
      expect(FormatSetzen).toHaveBeenCalledTimes(2)
    })

    // **Der Index ist belegt** — das `waitFor` darüber hat auf den
    // zweiten Ruf gewartet.
    expect(FormatSetzen.mock.calls[1]?.[0].fassung).toBe('2')
    expect(screen.queryByText(/Fassung stimmt nicht/)).toBeNull()
  })

  // **Befund 1 des Nachlaufs.** Ein offenes Sideboard war an keiner Fläche zu
  // sehen — auch nicht am Fenster.
  it('zeigt ein offenes Sideboard und schickt es mit', async () => {
    const FormatSetzen = await baueFenster(
      formatprobe({ sideboardgroesse: 0, sideboardOffen: true }),
    )

    const bereich = await warteAufDieFormatzeile()

    const marke = within(bereich).getByLabelText(texte.sideboardOffen)

    expect(marke).toBeChecked()

    await userEvent.click(
      within(bereich).getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    await waitFor(() => {
      expect(FormatSetzen).toHaveBeenCalledTimes(1)
    })

    expect(FormatSetzen.mock.calls[0]?.[0].sideboardOffen).toBe(true)
  })

  // **Und wer sie abnimmt, stellt es ab.**
  it('nimmt das offene Sideboard wieder zurück', async () => {
    const FormatSetzen = await baueFenster(
      formatprobe({ sideboardgroesse: 0, sideboardOffen: true }),
    )

    const bereich = await warteAufDieFormatzeile()

    await userEvent.click(within(bereich).getByLabelText(texte.sideboardOffen))
    await userEvent.click(
      within(bereich).getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    await waitFor(() => {
      expect(FormatSetzen).toHaveBeenCalledTimes(1)
    })

    expect(FormatSetzen.mock.calls[0]?.[0].sideboardOffen).toBe(false)
  })
  // **Und die Gegenprobe zur Zusicherung darüber.**
  //
  // Der Fall oben endet mit „keine Meldung über die Fassung". Dieser Satz
  // sagt nur dann etwas, wenn die Vorrichtung eine solche Meldung überhaupt
  // geben **kann** — eine Gegenprobe, die nichts ändert, ist keine (§10).
  // Hier wird sie einmal ausgelöst: eine Fassung, die der Bestand nicht mehr
  // führt, bekommt den Konflikt.
  it('die Werkbank weist eine überholte Fassung ab', async () => {
    const FormatSetzen = await baueFenster(formatprobe())

    await warteAufDieFormatzeile()

    const antwort = await FormatSetzen(formatprobe({ fassung: '99' }))

    expect(antwort.fehler?.code).toBe('konflikt')
    expect(antwort.format).toBeNull()
  })
})

// **Ein Konflikt an der Zeile bietet den heutigen Stand an** — UX-08.AK-02
// (#1680).
// Jemand anderes hat das Format inzwischen gespeichert: der Entwurf steht,
// bis „Neu laden" den Stand holt — dann steht dessen Zahl da.
describe('ein Konflikt an der Formatzeile', () => {
  it('bietet „Neu laden" an und zeigt danach den heutigen Stand', async () => {
    let stand = formatprobe()
    const Formate = vi.fn(() =>
      Promise.resolve({ formate: [stand], fehler: null }),
    )

    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate,
      Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
      FormatSetzen: () => {
        stand = { ...stand, fassung: '2', mindestgroesse: 50 }

        return Promise.resolve({
          format: null,
          fehler: {
            code: 'konflikt',
            schluessel: 'konflikt',
            meldung: 'Der Eintrag wurde zwischenzeitlich geändert.',
          },
        })
      },
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
    const feld = within(bereich).getByLabelText(texte.mindestgroesse)

    await userEvent.clear(feld)
    await userEvent.type(feld, '40')
    await userEvent.click(
      within(bereich).getByRole('button', {
        name: `${texte.formatSichern}: Konstruiert`,
      }),
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.zustaende.neuLaden,
      }),
    )

    expect(await screen.findByDisplayValue('50')).toBeInTheDocument()
    expect(screen.queryByRole('alert')).toBeNull()
    expect(Formate.mock.calls.length).toBeGreaterThan(1)
  })

  // **Ein anderer Fehler ohne Feld bietet es nicht an** — ein Neuladen
  // änderte an ihm nichts.
  it('bietet es bei einem anderen Fehler nicht an', async () => {
    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
      FormatSetzen: () =>
        Promise.resolve({
          format: null,
          fehler: {
            code: 'belegt',
            schluessel: '',
            meldung: 'Der Bestand ist gesperrt.',
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

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Bestand ist gesperrt.',
    )
    expect(
      screen.queryByRole('button', {
        name: kataloge.de.bausteine.zustaende.neuLaden,
      }),
    ).toBeNull()
  })
})
