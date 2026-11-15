// Die Wahlbilder einer Stiloption
// (#1225).
//
// **MSE zeigt zu jeder Wahl ein Bild, dieses Fenster zeigte nur Wörter.**
// Über `full-magic-pack/data` tragen 303 von 376 Stilpaketen `choice images:`
// (gezählt am 2026-08-28).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { type StiloptionDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Stiloptionen } from './Stiloptionen'

afterEach(() => {
  raeumeApi()
})

// tapsymbol ist die Option, an der MSE seine drei Symbole zeigt.
function tapsymbol(teile: Partial<StiloptionDTO> = {}): StiloptionDTO {
  return {
    name: 'tap symbol',
    art: 'choice',
    beschreibung: 'What tap symbols should be used?',
    wert: 'modern',
    moeglichkeiten: [
      { wert: 'modern', beschriftung: 'modern' },
      { wert: 'old', beschriftung: 'old' },
    ],
    eigeneFarbe: false,
    gesetzt: false,
    hatVorgabe: false,
    hatWahlbilder: true,
    vorgabe: '',
    ...teile,
  }
}

const punkt = 'data:image/png;base64,iVBORw0KGgo='

// **Zwei Wahlen, zwei Bilder — und die Wahl bleibt bedienbar.**
it('zeigt zu jeder Wahl ihr Bild und schreibt beim Umstellen', async () => {
  const setzen = vi.fn().mockResolvedValue({ fehler: null })

  belegeApi({
    Stiloptionen: () =>
      Promise.resolve({ optionen: [tapsymbol()], fehler: null }),
    Stilwahlbilder: () =>
      Promise.resolve({
        bilder: [
          { wahl: 'modern', bild: punkt },
          { wahl: 'old', bild: punkt },
        ],
        befunde: [],
        fehler: null,
      }),
    StiloptionSetzen: setzen,
  })

  render(
    <Sprachraum sprache="de">
      <Stiloptionen setID="1" geaendert={vi.fn()} />
    </Sprachraum>,
  )

  // **Optionsfelder und keine Auswahlliste** — ein `<option>` trägt kein Bild.
  const modern = await screen.findByRole('radio', { name: /modern/ })
  const alt = screen.getByRole('radio', { name: /old/ })

  expect(modern).toBeChecked()
  expect(screen.getAllByRole('presentation')).toHaveLength(2)

  await userEvent.click(alt)

  await waitFor(() => {
    expect(setzen).toHaveBeenCalledWith('1', 'tap symbol', 'old')
  })
})

// **Eine Wahl ohne Bild behält ihren Namen** — und zeigt kein zerbrochenes
// Symbol (§8).
it('lässt eine Wahl ohne Bild bei ihrem Namen', async () => {
  belegeApi({
    Stiloptionen: () =>
      Promise.resolve({ optionen: [tapsymbol()], fehler: null }),
    Stilwahlbilder: () =>
      Promise.resolve({
        bilder: [{ wahl: 'modern', bild: punkt }],
        befunde: ['old: die Datei fehlt'],
        fehler: null,
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Stiloptionen setID="1" geaendert={vi.fn()} />
    </Sprachraum>,
  )

  expect(await screen.findByRole('radio', { name: /old/ })).toBeInTheDocument()
  expect(screen.getAllByRole('presentation')).toHaveLength(1)
})

// **Ohne Bilder bleibt die Auswahlliste** — und es wird gar nicht gefragt.
//
// Der zweite Teil ist der wichtige: über das ganze Pack wären es 1.097
// Rasterläufe, und `magic-m15` erklärt dreizehn Optionen mit Bildern für vier.
it('fragt für eine Option ohne Bilder gar nicht erst', async () => {
  const holen = vi.fn()

  belegeApi({
    Stiloptionen: () =>
      Promise.resolve({
        optionen: [tapsymbol({ hatWahlbilder: false })],
        fehler: null,
      }),
    Stilwahlbilder: holen,
  })

  render(
    <Sprachraum sprache="de">
      <Stiloptionen setID="1" geaendert={vi.fn()} />
    </Sprachraum>,
  )

  expect(
    await screen.findByRole('combobox', { name: /tap symbol/ }),
  ).toBeInTheDocument()
  expect(holen).not.toHaveBeenCalled()
})

// **Der Vorgabeeintrag steht auch zwischen den Kacheln** — und setzt zurück.
//
// **Im Pack kommt der Fall nicht vor**: von den 1.096 Optionen mit
// `choice images:` nennt keine einzige ein `default:` (gezählt am
// 2026-08-28). Er steht trotzdem da, damit die Kacheln nicht können, was die
// Auswahlliste kann, minus das Zurücksetzen — eine Ungleichheit, die erst der
// ersten Vorlage auffiele, die beides führt.
it('wählt den Vorgabeeintrag, solange nichts gesetzt ist', async () => {
  belegeApi({
    Stiloptionen: () =>
      Promise.resolve({
        optionen: [tapsymbol({ hatVorgabe: true, gesetzt: false })],
        fehler: null,
      }),
    Stilwahlbilder: () =>
      Promise.resolve({
        bilder: [{ wahl: 'modern', bild: punkt }],
        befunde: [],
        fehler: null,
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Stiloptionen setID="1" geaendert={vi.fn()} />
    </Sprachraum>,
  )

  // **Gewählt ist die Vorgabe** — und nicht `modern`, obwohl der Wert darauf
  // steht. MSE hält es ebenso: `selection()` gibt 0, sobald
  // `value.isDefault()` (`src/gui/value/color.cpp`, Zeilen 104–108, Fassung
  // f9d9356 vom 2021-07-06).
  expect(await screen.findByRole('radio', { name: /Vorgabe/ })).toBeChecked()
  expect(screen.getByRole('radio', { name: /modern/ })).not.toBeChecked()
})

// **Und wer ihn wählt, setzt zurück** — es entfernt den gespeicherten Wert
// und schickt nicht die leere Zeichenkette (#764).
it('setzt über den Vorgabeeintrag zurück', async () => {
  const zuruecksetzen = vi.fn().mockResolvedValue({ fehler: null })

  belegeApi({
    Stiloptionen: () =>
      Promise.resolve({
        optionen: [tapsymbol({ hatVorgabe: true, gesetzt: true })],
        fehler: null,
      }),
    Stilwahlbilder: () =>
      Promise.resolve({
        bilder: [{ wahl: 'modern', bild: punkt }],
        befunde: [],
        fehler: null,
      }),
    StiloptionZuruecksetzen: zuruecksetzen,
  })

  render(
    <Sprachraum sprache="de">
      <Stiloptionen setID="1" geaendert={vi.fn()} />
    </Sprachraum>,
  )

  await userEvent.click(await screen.findByRole('radio', { name: /Vorgabe/ }))

  await waitFor(() => {
    expect(zuruecksetzen).toHaveBeenCalledWith('1', 'tap symbol')
  })
})
