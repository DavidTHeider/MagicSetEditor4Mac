// **Ein gelöschtes Set kann nicht das offene bleiben** — §7, Regel 1
// (#1218).
//
// # Der Fall
//
// Wer aus der Setliste das Set löscht, das er gerade aufgeschlagen hat, hätte
// sonst ein Fenster, dessen Laschen auf ein Set zeigen, das es nicht mehr
// gibt: jede Karte, jede Vorschau, jede Stilliste liefe gegen eine Kennung
// ohne Zeile.
//
// # Und die Gegenrichtung gehört dazu
//
// **Wer ein *anderes* Set löscht, verliert seine Arbeit nicht.** Ein
// bedingungsloses Schließen wäre die bequeme Fassung derselben Regel und
// falsch — deshalb steht der zweite Fall hier neben dem ersten, und er ist es,
// der die Bedingung prüft.
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { leisteneintrag } from './test/leistenweg'
import { waehleSethandlung } from './test/sethandlung'

const texte = kataloge.de

function satz(id: string, name: string) {
  return {
    id,
    name,
    spiel: 'probe',
    stil: 'schlicht',
    herkunft: `/pfad/${name}.mse-set`,
    karten: 1,
    herkunftFehlt: false,
  }
}

beforeEach(() => {
  belegeApi({ ...belegeFenster('de') })
})

afterEach(() => {
  raeumeApi()
})

async function zurSetliste() {
  await userEvent.click(await leisteneintrag('sets'))
}

async function oeffne(name: string) {
  await zurSetliste()
  await userEvent.click(
    await screen.findByRole('button', {
      name: `${name} — ${texte.seiten.sets.oeffnen}`,
    }),
  )
}

// loesche tippt den Namen ab und bestätigt — der Weg, den jemand geht. Seit
// UXR-04 (#1912) beginnt er im Menü der Setkarte; vorher trug jede Karte
// einen Knopf „Löschen …".
async function loesche(name: string) {
  await waehleSethandlung(texte, name, texte.seiten.sets.loeschen)

  // **Die Frage des gemeinten Sets** — die Liste hat mehrere.
  const zeile = screen
    .getByLabelText(texte.seiten.sets.loeschenFeld)
    .closest('li')

  if (!zeile) {
    throw new Error(`keine Zeile für ${name}`)
  }

  await userEvent.type(
    within(zeile).getByLabelText(texte.seiten.sets.loeschenFeld),
    name,
  )

  await userEvent.click(
    within(zeile).getByRole('button', { name: texte.seiten.sets.loeschenJa }),
  )
}

// zurKarten geht auf eine **Setseite** — dort entscheidet sich, ob noch ein
// Set offen ist.
//
// **Über ⌥⌘1 und nicht über die Seitenleiste.** Die führt nur
// Anwendungsseiten; zu einer Setseite kommt man über die Laschen — und die
// stehen im Arbeitsplatz, den es ohne offenes Set gerade nicht gibt. Das
// Kürzel ist der Weg, der in beiden Fällen derselbe ist.
//
// **⌥⌘1 seit dem 2026-09-03**, vorher ⌘4: die erste Reihe gehört seither ganz
// der Seitenleiste (Werkstück C von Phase 4.10), die Laschen sitzen auf ⌥⌘.
//
// **Über `code` und nicht über `key`**, denn ⌥ schreibt die Taste um: ⌥1 gibt
// auf einer Mac-Tastatur `"¡"`, nicht `"1"` (`rahmen/zweitereihe.test.ts`).
function zurKarten() {
  fireEvent.keyDown(window, {
    key: '¡',
    code: 'Digit1',
    metaKey: true,
    altKey: true,
  })
}

function belegeMitZweiSets(
  SetLoeschen: (
    setID: string,
    name: string,
  ) => Promise<{
    befunde: string[]
    fehler: null
  }>,
) {
  belegeApi({
    ...belegeFenster('de'),
    Sets: () =>
      Promise.resolve({ sets: [satz('7', 'Erstes'), satz('8', 'Zweites')] }),
    Karten: () => Promise.resolve({ karten: [] }),
    Kartenzeilen: () => Promise.resolve({ zeilen: [], gesamt: 0 }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    SetLoeschen,
  })
}

it('schließt das offene Set, wenn es das gelöschte war', async () => {
  const SetLoeschen = vi.fn(() =>
    Promise.resolve({ befunde: [], fehler: null }),
  )
  belegeMitZweiSets(SetLoeschen)

  render(<App />)

  await oeffne('Erstes')

  // **Vorher steht der Arbeitsplatz** — sonst sagte der Vergleich danach
  // nichts.
  expect(await screen.findByRole('tablist')).toBeInTheDocument()

  await zurSetliste()
  await loesche('Erstes')

  await waitFor(() => {
    expect(SetLoeschen).toHaveBeenCalledWith('7', 'Erstes')
  })

  // **Auf einer Setseite steht danach der Leerzustand.** Nicht auf der
  // Setliste: die ist eine Anwendungsseite und hat ohnehin keinen
  // Arbeitsplatz.
  zurKarten()

  expect(
    await screen.findByText(texte.allgemein.arbeitsplatz.keinSet),
  ).toBeInTheDocument()
})

it('lässt das offene Set stehen, wenn ein anderes gelöscht wird', async () => {
  const SetLoeschen = vi.fn(() =>
    Promise.resolve({ befunde: [], fehler: null }),
  )
  belegeMitZweiSets(SetLoeschen)

  render(<App />)

  await oeffne('Erstes')
  await zurSetliste()
  await loesche('Zweites')

  await waitFor(() => {
    expect(SetLoeschen).toHaveBeenCalledWith('8', 'Zweites')
  })

  // **Derselbe Weg wie oben, und das andere Ergebnis**: das Set ist noch
  // offen, also stehen seine Laschen.
  zurKarten()

  expect(await screen.findByRole('tablist')).toBeInTheDocument()
  expect(
    screen.queryByText(texte.allgemein.arbeitsplatz.keinSet),
  ).not.toBeInTheDocument()
})

// **Und eine späte Antwort trifft nur das Set, das sie meint** (Nachlauf über
// die Folgen vor 0.8.0, Befund 1): wer das offene Set löscht und, während
// die Antwort aussteht, ein anderes aufschlägt, behält das andere. Der
// Rückruf verglich mit dem Set, das beim **Beginn** des Löschens offen war.
it('lässt ein inzwischen geöffnetes Set stehen, wenn das Löschen spät antwortet', async () => {
  let antworte: () => void = () => undefined
  const SetLoeschen = vi.fn(
    () =>
      new Promise<{ befunde: string[]; fehler: null }>((ja) => {
        antworte = () => {
          ja({ befunde: [], fehler: null })
        }
      }),
  )
  belegeMitZweiSets(SetLoeschen)

  render(<App />)

  await oeffne('Erstes')
  await zurSetliste()
  await loesche('Erstes')

  await waitFor(() => {
    expect(SetLoeschen).toHaveBeenCalledWith('7', 'Erstes')
  })

  await oeffne('Zweites')
  expect(await screen.findByRole('tablist')).toBeInTheDocument()

  await act(async () => {
    antworte()
    await Promise.resolve()
  })

  zurKarten()

  expect(await screen.findByRole('tablist')).toBeInTheDocument()
  expect(
    screen.queryByText(texte.allgemein.arbeitsplatz.keinSet),
  ).not.toBeInTheDocument()
})
