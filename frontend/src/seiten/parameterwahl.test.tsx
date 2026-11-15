// **Die Parameterhilfen der Schlüsselwörter** — Werkstück D von Phase 4.11
// (#1342).
//
// # Was MSE2 hier tut
//
// Zwei Knöpfe mit Aufklappmenüs (`src/gui/set/keywords_panel.cpp`, Zeilen
// 190–232, Fassung f9d9356 vom 2021-07-06). Hier sind es zwei Auswahllisten —
// ein Menü über einem Knopf wäre ein Menü in einem Menü, und §6 verbietet
// verschachtelte Dialoge.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Schluesselworteditor } from './Schluesselworteditor'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.schluesselwoerter

const hilfen = {
  // **Die zweite ohne Beschreibung** — `description:` ist optional, und dann
  // steht nur der Name da.
  arten: [
    {
      name: 'zahl',
      beschreibung: 'eine Zahl',
      einfuegen: '<atom-param>zahl</atom-param>',
    },
    {
      name: 'wort',
      beschreibung: '',
      einfuegen: '<atom-param>wort</atom-param>',
    },
  ],
  verweise: [
    {
      nummer: 1,
      art: 'mana',
      name: '',
      beschreibung: '',
      einfuegen: '{param1}',
    },
    {
      nummer: 1,
      art: 'mana',
      name: 'manawert',
      beschreibung: 'als Zahl',
      einfuegen: '{cmc(param1)}',
    },
  ],
  befunde: [],
  fehler: null,
}

function zeige(schreiben = vi.fn(), zusatz = {}) {
  belegeApi({ Parameterhilfen: () => Promise.resolve(hilfen), ...zusatz })

  render(
    <Sprachraum sprache="de">
      <Schluesselworteditor
        wort={{
          id: '7',
          wort: 'Kicker',
          muster: 'Kicker ',
          erinnerung: 'Du darfst ',
          regeln: '',
          modus: 'custom',
          fest: false,
          gueltig: true,
        }}
        modi={['custom']}
        angehalten={false}
        setID="1"
        stand={0}
        bearbeiten={{ schreiben, loeschen: vi.fn() }}
      />
    </Sprachraum>,
  )

  return schreiben
}

// **Die Parameterarten stehen unter dem Muster.**
it('zeigt die Parameterarten des Spiels', async () => {
  zeige()

  const wahl = await screen.findByRole('combobox', {
    name: texte.parameterEinfuegen,
  })

  expect(wahl).toBeInTheDocument()
  expect(
    screen.getByRole('option', { name: 'zahl — eine Zahl' }),
  ).toBeInTheDocument()

  // **Und eine ohne Beschreibung steht mit ihrem Namen allein da** — ohne den
  // Gedankenstrich, der ins Leere zeigte.
  expect(screen.getByRole('option', { name: 'wort' })).toBeInTheDocument()
})

// **Eine Wahl schreibt in das Feld darüber — an der Textmarke.**
//
// Die Marke steht nach dem Zeichnen am Ende; geprüft wird deshalb, dass der
// Text **angehängt** ankommt und der bisherige Inhalt stehen bleibt.
it('fügt die Marke ins Muster ein und schreibt sie fort', async () => {
  const schreiben = zeige()

  await userEvent.selectOptions(
    await screen.findByRole('combobox', { name: texte.parameterEinfuegen }),
    'zahl',
  )

  // **Und es steht auch da.** Das Feld ist ungesteuert: ohne den Griff in den
  // Knoten schriebe die Zeile den neuen Wert fort und zeigte den alten.
  //
  // **Gewartet wird darauf und nicht auf den Ruf** (§3, #1321).
  await waitFor(() => {
    expect(screen.getByRole('textbox', { name: texte.muster })).toHaveValue(
      'Kicker <atom-param>zahl</atom-param>',
    )
  })

  expect(schreiben).toHaveBeenCalledWith(
    'muster',
    'Kicker <atom-param>zahl</atom-param>',
    expect.any(Array),
  )
})

// **Die Verweise stehen unter dem Erinnerungstext — mit Nummer und Art.**
//
// MSEs Aufschrift ist `1. <name>`; die spitzen Klammern sind seine eigenen
// (`keywords_panel.cpp`, Zeile 204).
it('zeigt die Verweise auf die Parameter dieses Schlüsselworts', async () => {
  zeige()

  await screen.findByRole('combobox', { name: texte.parameterVerweisen })

  expect(screen.getByRole('option', { name: '1. ⟨mana⟩' })).toBeInTheDocument()
  expect(
    screen.getByRole('option', { name: '1. ⟨mana⟩ — manawert' }),
  ).toBeInTheDocument()
})

// **Eine Wahl schreibt in den Erinnerungstext.**
it('fügt den Verweis in den Erinnerungstext ein', async () => {
  const schreiben = zeige()

  await userEvent.selectOptions(
    await screen.findByRole('combobox', { name: texte.parameterVerweisen }),
    '1',
  )

  await waitFor(() => {
    expect(schreiben).toHaveBeenCalledWith(
      'erinnerung',
      'Du darfst {cmc(param1)}',
      expect.any(Array),
    )
  })
})

// **Ein Befund steht an der Zeile und nimmt das Menü nicht mit.**
it('zeigt einen Befund der Verweisskripte', async () => {
  zeige(vi.fn(), {
    Parameterhilfen: () =>
      Promise.resolve({
        ...hilfen,
        befunde: ['das Verweisskript "manawert" lief nicht'],
      }),
  })

  expect(
    await screen.findByText(/Verweisskript "manawert" lief nicht/),
  ).toBeInTheDocument()

  expect(
    screen.getByRole('combobox', { name: texte.parameterVerweisen }),
  ).toBeInTheDocument()
})

// **Und wirklich an der Textmarke, nicht am Ende.**
//
// # Warum dieser Fall zusätzlich dasteht
//
// Weil der Fall darüber ihn nicht enthält: nach dem Zeichnen steht die Marke
// am Ende, und dort sind Einfügen und Anhängen dasselbe. Eine Gegenprobe hat
// es gezeigt — den Einschub durch ein `wert + text` ersetzt, und der Lauf
// blieb grün (§10).
it('fügt an der Textmarke ein und nicht am Ende', async () => {
  const schreiben = vi.fn()

  zeige(schreiben)

  const feld = await screen.findByRole('textbox', { name: texte.muster })

  // **Die Marke vor das Leerzeichen** — `Kicker| `.
  ;(feld as HTMLInputElement).setSelectionRange(6, 6)

  // **Gewartet, nicht gegriffen**: die Auswahl entsteht erst, wenn die
  // Parameter der Vorlage da sind (§3, #1321).
  await userEvent.selectOptions(
    await screen.findByRole('combobox', { name: texte.parameterEinfuegen }),
    'zahl',
  )

  await waitFor(() => {
    expect(schreiben).toHaveBeenCalledWith(
      'muster',
      'Kicker<atom-param>zahl</atom-param> ',
      expect.any(Array),
    )
  })
})
