// **Die Einführung ist ein Fenster, keine Seitenhälfte** —
// #1325.
//
// # Was vorher war
//
// Die Startseite zeigte *Erste Schritte* **oder** *Dein Überblick*, je
// nachdem, ob der Bestand ein Set führt. Solange die Zahlen unterwegs waren,
// galt „kein Set" — und die Seite **sprang um**, sobald die Antwort kam. Der
// Kommentar darüber sagte ausdrücklich das Gegenteil: *„eine Startseite, die
// erst »Erste Schritte« zeigt und dann etwas anderes, wäre ein Sprung vor den
// Augen dessen, der gerade liest."*
//
// Und sie war an dem Tag weg, an dem man sie zum zweiten Mal gebraucht hätte.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Einfuehrung } from './Einfuehrung'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { Umschlag } from '../umschlag'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.start

function zeige(zusatz: Record<string, unknown> = {}) {
  const schliessen = vi.fn()
  const wechseln = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Einfuehrung
        setID={undefined}
        wechseln={wechseln}
        schliessen={schliessen}
        {...zusatz}
      />
    </Sprachraum>,
  )

  return { schliessen, wechseln }
}

it('steht als Fenster mit den drei Schritten', () => {
  zeige()

  const fenster = screen.getByRole('dialog', {
    name: texte.einfuehrung.bezeichnung,
  })

  expect(fenster).toBeInTheDocument()
  expect(screen.getByText(texte.ersteSchritte)).toBeInTheDocument()

  for (const schritt of [
    'setOeffnen',
    'kartenBearbeiten',
    'paketeHolen',
  ] as const) {
    expect(screen.getByText(texte.schritte[schritt].titel)).toBeInTheDocument()
  }
})

// **Ein Schritt, den man noch nicht gehen kann, bleibt stehen** (#673) — die
// Zusage zieht mit um.
it('lässt den mittleren Schritt stehen, wenn kein Set offen ist', () => {
  const { wechseln } = zeige()

  expect(
    screen.getByText(texte.schritte.kartenBearbeiten.ersteinSet),
  ).toBeInTheDocument()

  screen.getByRole('button', {
    name: texte.schritte.kartenBearbeiten.stattdessen,
  })

  expect(wechseln).not.toHaveBeenCalled()
})

it('führt zu den Karten, wenn ein Set offen ist', async () => {
  const { wechseln, schliessen } = zeige({ setID: '1' })

  await userEvent.click(
    screen.getByRole('button', {
      name: texte.schritte.kartenBearbeiten.handlung,
    }),
  )

  expect(wechseln).toHaveBeenCalledWith('karten')

  // **Wer einem Schritt folgt, will das Fenster nicht mehr sehen.** Ein
  // Fenster, das über dem Ziel stehen bliebe, wäre der zweite Klick, den
  // niemand verlangt hat.
  expect(schliessen).toHaveBeenCalled()
})

// **Der Schalter steht im Fenster selbst** — und schreibt in den Bestand.
//
// §6: niemand sucht die Einstellungen, um etwas loszuwerden, das gerade vor
// ihm steht.
it('schaltet das Wiederkehren im Fenster selbst ab', async () => {
  const EinstellungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  belegeApi({ EinstellungSetzen })

  const { schliessen } = zeige()

  const schalter = screen.getByRole('checkbox', {
    name: texte.einfuehrung.wiederzeigen,
  })

  expect(schalter).toBeChecked()

  await userEvent.click(schalter)

  await waitFor(() => {
    expect(EinstellungSetzen).toHaveBeenCalledWith('ersteschritte', 'aus')
  })

  // **Und der Haken ist weg.**
  //
  // **Das ist die Zusicherung, die gefehlt hat** — gemeldet vom Auftraggeber
  // am 2026-09-06: „Der Haken dort geht nicht weg." Das Feld war
  // **kontrolliert** über eine Eigenschaft, die sich nie ändert; der Klick
  // schrieb die Einstellung, und React zeichnete den Haken sofort wieder hin.
  //
  // **Die Läufe hier sahen es nicht**, weil sie `wiederZeigen` selbst
  // mitgaben und danach nur den **Ruf** prüften, nicht die **Wirkung** (§3:
  // gewartet wird auf die Wirkung, nicht auf den Ruf).
  expect(schalter).not.toBeChecked()

  // **Das Fenster bleibt offen.** Wer den Haken wegnimmt, sagt etwas über das
  // **nächste** Mal — nicht über dieses.
  expect(schliessen).not.toHaveBeenCalled()
})

// **Und wenn das Schreiben scheitert, springt der Haken zurück.**
//
// **Optimistisch mit Rücknahme** (§6). Ohne die Rücknahme stünde ein Haken da,
// den der Bestand nicht kennt — und beim nächsten Start käme die Einführung
// doch wieder, ohne dass jemand versteht, warum.
it('nimmt den Haken zurück, wenn das Schreiben scheitert', async () => {
  belegeApi({
    // **Der Typ steht dran**, weil `vi.fn` die Rückgabe sonst zu
    // `code: string` verbreitert und `FehlerDTO` eine Aufzählung verlangt.
    EinstellungSetzen: vi.fn((): Promise<Umschlag> =>
      Promise.resolve({
        fehler: { code: 'konflikt', schluessel: 'konflikt', meldung: '' },
      }),
    ),
  })

  zeige()

  const schalter = screen.getByRole('checkbox', {
    name: texte.einfuehrung.wiederzeigen,
  })

  await userEvent.click(schalter)

  expect(await screen.findByRole('alert')).toHaveTextContent(
    kataloge.de.fehler.konflikt,
  )

  expect(schalter).toBeChecked()
})

it('schließt sich über den Knopf', async () => {
  const { schliessen } = zeige()

  await userEvent.click(
    screen.getByRole('button', { name: texte.einfuehrung.schliessen }),
  )

  expect(schliessen).toHaveBeenCalled()
})

// **Und über Escape** — dieselbe Zusage wie bei der Befehlspalette.
it('schließt sich über Escape', async () => {
  const { schliessen } = zeige()

  await userEvent.keyboard('{Escape}')

  expect(schliessen).toHaveBeenCalled()
})

// **Und wieder an** — der Schalter geht in beide Richtungen.
//
// Ein Schalter, der nur in eine Richtung geht, ist eine Sackgasse; wer ihn
// versehentlich wegnimmt, soll ihn im selben Fenster zurückholen können,
// statt die Einstellungen suchen zu müssen.
it('schaltet das Wiederkehren auch wieder an', async () => {
  const EinstellungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  belegeApi({ EinstellungSetzen })

  zeige({ wiederZeigen: false })

  const schalter = screen.getByRole('checkbox', {
    name: texte.einfuehrung.wiederzeigen,
  })

  expect(schalter).not.toBeChecked()

  await userEvent.click(schalter)

  await waitFor(() => {
    expect(EinstellungSetzen).toHaveBeenCalledWith('ersteschritte', 'an')
  })
})
