// **Der Stil eines neuen Sets, mit dem ganzen Pack** — UXR-04
// (#1912).
//
// # Was gemessen war
//
// Am Probebestand stehen im Schritt *Aussehen* eine Handvoll Stile. Mit dem
// Full-Magic-Pack bringt `magic` **355** mit (gezählt am 2026-10-02 über
// `data/magic-*.mse-style`), und die `Stilliste` zeichnete je Stil ein
// Optionsfeld: 355 Felder und 355 Tabstopps, R3, R6 und R7 aus
// `docs/UIREVIEW.md` zugleich. Dieselbe Frage hat UXR-07 in der Schublade
// *Eigenschaften* mit der `Stilauswahl` beantwortet — hier steht sie jetzt
// auch.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sets } from './Sets'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { kataloge } from '../sprache/kataloge'
import { zeilengrenze } from '../bausteine/suchwahl'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.sets

function paket(art: string, name: string, anzeigename: string, rang = 0) {
  return {
    art,
    name,
    anzeigename,
    version: '',
    pfad: '',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang,
    artrang: 0,
    herkunftsart: 'mitgeliefert' as const,
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
  }
}

// vollerPack ist Magic mit so vielen Stilen, wie der Full-Magic-Pack bringt.
function vollerPack() {
  const stile = Array.from({ length: 355 }, (_, stelle) =>
    paket(
      'mse-style',
      `magic-stil${String(stelle).padStart(3, '0')}`,
      `Stil ${String(stelle)}`,
      stelle + 1,
    ),
  )

  return {
    pakete: [paket('mse-game', 'magic', 'Magic'), ...stile],
    befunde: [],
    fehler: null,
  }
}

function zeige(SetAnlegen = vi.fn().mockResolvedValue({ id: '9' })) {
  belegeApi({
    Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    Pakete: vi.fn().mockResolvedValue(vollerPack()),
    SetAnlegen,
  })

  render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        waehlen={vi.fn()}
        speichern={stillerSpeicherlauf()}
      />
    </Sprachraum>,
  )

  return SetAnlegen
}

describe('Der Stil eines neuen Sets', () => {
  it('steht mit 355 Stilen in einem Auswahlfeld und nicht in 355 Optionsfeldern', async () => {
    zeige()

    // **Ein Spiel ist kein Schritt** — der erste ist das Aussehen.
    await screen.findByRole('heading', {
      name: worte.neuSchrittAussehen,
      level: 3,
    })

    const neu = screen.getByRole('region', { name: worte.neuTitel })
    expect(within(neu).queryAllByRole('radio')).toHaveLength(0)

    const auswahl = within(neu).getByRole('combobox', { name: worte.neuStil })

    // **Die Liste steht erst, wenn das Feld offen ist** (UXR-F1, #2007):
    // 355 Zeilen im Baum, solange niemand sie sucht, wären 355 Knoten für
    // nichts. **Und auch offen zeichnet sie höchstens zweihundert** und sagt,
    // dass es mehr gibt.
    expect(screen.queryAllByRole('option')).toHaveLength(0)

    await userEvent.click(auswahl)
    expect(screen.getAllByRole('option')).toHaveLength(zeilengrenze)
    expect(
      screen.getAllByText(
        kataloge.de.bausteine.suchauswahl.begrenzt
          .replace('{gezeigt}', String(zeilengrenze))
          .replace('{anzahl}', '355'),
      ),
    ).not.toHaveLength(0)

    // **Vorgewählt ist der erste in MSE2s Ordnung**, wie vorher — bei
    // offener Liste steht die Wahl als Platzhalter da (UXR-F1, `fmt02d`).
    expect(auswahl).toHaveAttribute('placeholder', 'Stil 0')
  })

  it('grenzt über die Suche ein und legt mit dem gewählten Stil an', async () => {
    const SetAnlegen = zeige()

    const auswahl = await screen.findByRole('combobox', { name: worte.neuStil })

    // **Die Suche steht im Feld** und nicht darüber (UXR-F1, #2007): Tippen
    // ersetzt den Text und grenzt die Liste ein.
    await userEvent.click(auswahl)
    await userEvent.keyboard('stil354')

    expect(
      screen.getAllByRole('option').map((zeile) => zeile.textContent),
    ).toEqual(['Stil 354 magic-stil354'])

    await userEvent.click(screen.getByRole('option'))
    expect(auswahl).toHaveValue('Stil 354')
    await userEvent.click(screen.getByRole('button', { name: worte.neuWeiter }))
    await userEvent.type(await screen.findByLabelText(worte.neuName), 'Groß')
    await userEvent.click(
      screen.getByRole('button', { name: worte.neuAnlegen }),
    )

    expect(SetAnlegen).toHaveBeenCalledWith('Groß', 'magic', 'magic-stil354')
  })
})
