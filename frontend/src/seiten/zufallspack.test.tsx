// **Das Zufallspack**
// (#1072).
//
// # Wofür
//
// `magic.mse-game/packs` führt **30** Packarten, und keine davon war zu
// erreichen. MSE hat dafür einen eigenen Bereich
// (`src/gui/set/random_pack_panel.cpp`, Fassung f9d9356 vom 2021-07-06).
//
// # Was hier geprüft wird
//
// Dass nur die **wählbaren** zur Wahl stehen — im echten Pack vier von
// dreißig —, dass der Same als **Text** reist und dass die Ziehreihenfolge
// steht.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO, PackartDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Zufallspack } from './Zufallspack'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.pack

function karte(id: string, name: string): KarteDTO {
  return { id, nummer: 0, werte: [{ feld: 'name', wert: name }] }
}

function zeige(
  arten: PackartDTO[],
  karten: KarteDTO[] = [],
  befunde: string[] = [],
) {
  const PackZiehen = vi.fn(() =>
    Promise.resolve({ karten, befunde: [], fehler: null }),
  )

  belegeApi({
    Packarten: () => Promise.resolve({ arten, befunde, fehler: null }),
    PackZiehen,
  })

  render(
    <Sprachraum sprache="de">
      <Zufallspack setID="3" />
    </Sprachraum>,
  )

  return PackZiehen
}

describe('Das Zufallspack', () => {
  // **Nur die wählbaren stehen zur Wahl.** Im echten Pack sind es vier von
  // dreißig; die übrigen sind Bausteine, die von jenen aus gezogen werden.
  it('bietet nur die wählbaren Packarten an', async () => {
    zeige([
      { name: 'booster pack', waehlbar: true, karten: 12 },
      { name: 'common', waehlbar: false, karten: 8 },
    ])

    const wahl = await screen.findByLabelText(texte.art)

    expect(
      screen.getByRole('option', { name: 'booster pack (12 Karten)' }),
    ).toBeInTheDocument()

    expect(
      screen.queryByRole('option', { name: /common/ }),
    ).not.toBeInTheDocument()
    expect(wahl).toHaveValue('booster pack')
  })

  it('zieht mit der gewählten Art und dem angegebenen Samen', async () => {
    const PackZiehen = zeige(
      [{ name: 'booster pack', waehlbar: true, karten: 12 }],
      [karte('1', 'Eisregen'), karte('2', 'Feuerball')],
    )

    const samenfeld = await screen.findByLabelText(texte.same)
    await userEvent.clear(samenfeld)
    await userEvent.type(samenfeld, '4711')
    await userEvent.click(screen.getByRole('button', { name: texte.ziehen }))

    await waitFor(() => {
      expect(PackZiehen).toHaveBeenCalledWith('3', 'booster pack', '4711')
    })

    // **In ihrer Ziehreihenfolge** — so legt MSE sie hin.
    const gezogen = await screen.findAllByRole('listitem')
    expect(gezogen[0]).toHaveTextContent('Eisregen')
    expect(gezogen[1]).toHaveTextContent('Feuerball')
  })

  // **Ein neuer Same ist ein Knopf und kein Zufall im Hintergrund.** Wer
  // zieht, soll die Zahl sehen, mit der gezogen wurde — sonst ist das Ergebnis
  // nicht wiederherstellbar.
  it('zählt den Samen sichtbar hoch', async () => {
    zeige([{ name: 'booster pack', waehlbar: true, karten: 12 }])

    await userEvent.click(
      await screen.findByRole('button', { name: texte.neuerSame }),
    )

    expect(screen.getByLabelText(texte.same)).toHaveValue('2')
  })

  it('sagt es, wenn das Spiel keine Packarten mitbringt', async () => {
    zeige([])

    expect(await screen.findByText(texte.keineArten)).toBeInTheDocument()
  })

  it('sagt es, wenn eine Packart nichts hergibt', async () => {
    zeige([{ name: 'leer', waehlbar: true, karten: 0 }], [])

    await userEvent.click(
      await screen.findByRole('button', { name: texte.ziehen }),
    )

    expect(await screen.findByText(texte.keineKarten)).toBeInTheDocument()
  })

  it('nennt die Packarten, deren Filter scheiterte', async () => {
    zeige(
      [{ name: 'booster pack', waehlbar: true, karten: 1 }],
      [],
      ['kaputt: der Filter ließ sich nicht ausrechnen'],
    )

    expect(
      await screen.findByText('kaputt: der Filter ließ sich nicht ausrechnen'),
    ).toBeInTheDocument()
  })

  it('lässt eine andere Packart wählen', async () => {
    const PackZiehen = zeige([
      { name: 'booster pack', waehlbar: true, karten: 12 },
      { name: 'fat pack', waehlbar: true, karten: 30 },
    ])

    await userEvent.selectOptions(
      await screen.findByLabelText(texte.art),
      'fat pack',
    )

    await userEvent.click(screen.getByRole('button', { name: texte.ziehen }))

    await waitFor(() => {
      expect(PackZiehen).toHaveBeenCalledWith('3', 'fat pack', '1')
    })
  })

  it('meldet einen Fehler beim Ziehen', async () => {
    belegeApi({
      Packarten: () =>
        Promise.resolve({
          arten: [{ name: 'booster pack', waehlbar: true, karten: 1 }],
          befunde: [],
          fehler: null,
        }),
      PackZiehen: () =>
        Promise.resolve({
          karten: [],
          befunde: [],
          fehler: {
            code: 'validierung' as const,
            schluessel: 'probe',
            meldung: 'Der Same ist keine ganze Zahl ab null.',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Zufallspack setID="3" />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.ziehen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Same ist keine ganze Zahl ab null.',
    )
  })
  // **Was ein Same sein darf, steht in `app.samenAus`** — Werkstück U3 von
  // Phase 5.1 (#1414):
  //
  //	samen, err := strconv.ParseUint(text, 10, 64)
  //
  // Also Ziffern und höchstens 2⁶⁴−1; ein leerer Text heißt null und ist
  // gültig. Hier stand bis heute „jede Zeichenfolge ist gültig" — und das
  // war falsch: `abc` reiste bis nach Go und kam als Feldfehler zurück, den
  // niemand einem Feld zuordnete.
  it('meldet einen Samen, der keine Zahl ist, am Feld', async () => {
    const PackZiehen = zeige([
      { name: 'booster pack', waehlbar: true, karten: 1 },
    ])

    const feld = await screen.findByLabelText(texte.same)

    await userEvent.clear(feld)
    await userEvent.type(feld, 'abc')

    expect(feld).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getByRole('alert')).toHaveTextContent(texte.sameZahl)

    await userEvent.click(screen.getByRole('button', { name: texte.ziehen }))

    expect(PackZiehen).not.toHaveBeenCalled()
  })

  // **Und ein Same jenseits von 2⁶⁴−1 ebenso** — `ParseUint(…, 64)` weist ihn
  // ab, und eine Zahl, die im Fenster durchginge, käme als Fehler zurück.
  it('meldet einen zu großen Samen am Feld', async () => {
    zeige([{ name: 'booster pack', waehlbar: true, karten: 1 }])

    const feld = await screen.findByLabelText(texte.same)

    await userEvent.clear(feld)
    await userEvent.type(feld, '18446744073709551616')

    expect(feld).toHaveAttribute('aria-invalid', 'true')
  })

  // **Der Grenzfall gilt**: 2⁶⁴−1 ist der größte Same, den Go annimmt.
  it('lässt den größten Samen durch', async () => {
    const PackZiehen = zeige([
      { name: 'booster pack', waehlbar: true, karten: 1 },
    ])

    const feld = await screen.findByLabelText(texte.same)

    await userEvent.clear(feld)
    await userEvent.type(feld, '18446744073709551615')

    expect(feld).toHaveAttribute('aria-invalid', 'false')

    await userEvent.click(screen.getByRole('button', { name: texte.ziehen }))

    await waitFor(() => {
      expect(PackZiehen).toHaveBeenCalledWith(
        '3',
        'booster pack',
        '18446744073709551615',
      )
    })
  })

  // **Und ein leerer Same ist gültig** — `samenAus` gibt für ihn null zurück.
  it('lässt einen leeren Samen durch', async () => {
    const PackZiehen = zeige([
      { name: 'booster pack', waehlbar: true, karten: 1 },
    ])

    const feld = await screen.findByLabelText(texte.same)

    await userEvent.clear(feld)

    expect(feld).toHaveAttribute('aria-invalid', 'false')

    await userEvent.click(screen.getByRole('button', { name: texte.ziehen }))

    await waitFor(() => {
      expect(PackZiehen).toHaveBeenCalledWith('3', 'booster pack', '')
    })
  })

  // **Ein Feldfehler aus Go steht am Feld** (§6): `samenAus` gibt
  // `NewFieldError("same", …)`, und `verursachendesFeld` liest ihn.
  it('hängt den Feldfehler aus Go an das Samenfeld', async () => {
    belegeApi({
      Packarten: () =>
        Promise.resolve({
          arten: [{ name: 'booster pack', waehlbar: true, karten: 1 }],
          befunde: [],
          fehler: null,
        }),
      PackZiehen: () =>
        Promise.resolve({
          karten: [],
          befunde: [],
          fehler: {
            code: 'validierung' as const,
            schluessel: 'probe',
            feld: 'same',
            meldung: 'Der Same "9" ist keine ganze Zahl ab null.',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Zufallspack setID="3" />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.ziehen }),
    )

    const feld = await screen.findByLabelText(texte.same)

    await waitFor(() => {
      expect(feld).toHaveAttribute('aria-invalid', 'true')
    })

    expect(feld).toHaveAccessibleDescription(
      'Der Same "9" ist keine ganze Zahl ab null.',
    )
  })
})

// **Ein leerer Same zählt von null hoch** — `samenAus` liest ihn als null,
// und `BigInt('')` gäbe es nicht. Werkstück U3 von Phase 5.1 (#1414).
it('zählt auch von einem leeren Samen hoch', async () => {
  zeige([{ name: 'booster pack', waehlbar: true, karten: 1 }])

  const feld = await screen.findByLabelText(texte.same)

  await userEvent.clear(feld)
  await userEvent.click(screen.getByRole('button', { name: texte.neuerSame }))

  expect(feld).toHaveValue('1')
})
