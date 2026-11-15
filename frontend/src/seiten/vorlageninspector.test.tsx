import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Vorlageninspector } from './Vorlageninspector'
import { kurzform } from '../sprache/kurzform'

import type { ComponentProps } from 'react'

const texte = kataloge.de.seiten.vorlagen

type Eigenschaften = ComponentProps<typeof Vorlageninspector>

function zeige(teile: Partial<Eigenschaften> = {}) {
  belegeApi({
    Vorlagenvorschau: () =>
      Promise.resolve({
        vorschau: null,
        fehler: {
          code: 'nichtGefunden',
          schluessel: 'keinSetMitStil',
          meldung: 'kein Set',
        },
      }),
  })

  const eigenschaften: Eigenschaften = {
    paket: '/pakete/probe.mse-style',
    datei: 'style',
    gueltig: 'mse version: 2.0.0',
    veraltet: null,
    karteID: '',
    setzeKarteID: vi.fn(),
    karten: [
      { id: '7', name: 'Langtext', nummer: 1 },
      { id: '9', name: '', nummer: 2 },
    ],
    kartenFehler: undefined,
    symbol: null,
    verwendungen: [],
    verwendungenLaden: false,
    oeffnen: vi.fn(),
    ...teile,
  }

  render(
    <Sprachraum sprache="de">
      <Vorlageninspector {...eigenschaften} />
    </Sprachraum>,
  )

  return eigenschaften
}

describe('Der Inspector der Werkstatt', () => {
  afterEach(raeumeApi)

  // **Die Testkarte ist eine Auswahl mit Suche** — UXR-F1
  // (#2007),
  // R7: ein Set trägt hunderte Karten. Die Zeilen stehen da, sobald das Feld
  // offen ist.
  it('bietet die Testkarten an und reicht die Wahl weiter', async () => {
    const { setzeKarteID } = zeige()

    const wahl = screen.getByRole('combobox', { name: texte.testkarte })

    // **Die Vorgabe hat einen Namen**: der leere Wert wäre ein leeres Feld.
    expect(wahl).toHaveValue(texte.testkarteVorgabe)
    expect(screen.queryAllByRole('option')).toHaveLength(0)

    await userEvent.click(wahl)

    expect(
      screen.getAllByRole('option').map((zeile) => zeile.textContent),
    ).toEqual([texte.testkarteVorgabe, 'Langtext', 'Karte 2'])

    await userEvent.click(screen.getByRole('option', { name: 'Karte 2' }))

    expect(setzeKarteID).toHaveBeenCalledExactlyOnceWith('9')
  })

  it('findet eine Karte über ein Stück ihres Namens', async () => {
    const { setzeKarteID } = zeige()

    await userEvent.click(
      screen.getByRole('combobox', { name: texte.testkarte }),
    )
    await userEvent.keyboard('angte{Enter}')

    expect(setzeKarteID).toHaveBeenCalledExactlyOnceWith('7')
  })

  it('zeigt die gewählte Karte im Feld', () => {
    zeige({ karteID: '9' })

    expect(screen.getByRole('combobox', { name: texte.testkarte })).toHaveValue(
      'Karte 2',
    )
  })

  // **Die Vorgabe wählen heißt: keine Karte** — der leere Wert, wie das
  // native Feld ihn meldete.
  it('meldet den leeren Wert, wenn die Vorgabe gewählt wird', async () => {
    const { setzeKarteID } = zeige({ karteID: '7' })

    await userEvent.click(
      screen.getByRole('combobox', { name: texte.testkarte }),
    )
    await userEvent.click(
      screen.getByRole('option', { name: texte.testkarteVorgabe }),
    )

    expect(setzeKarteID).toHaveBeenCalledExactlyOnceWith('')
  })

  // **Die Vorschau steht oben, die Testkarte darunter** — wie im Artboard
  // `Werkstatt` (UXR-18, #1926): die Karte ist das Hauptobjekt (§6), und
  // die Wahl, womit geprobt wird, ist ihre Unterschrift. Vorher schob die
  // Wahl samt Erklärung das Bild im 320 Punkte breiten Inspector unter den
  // Rand.
  it('zeigt die Vorschau vor der Wahl der Testkarte', () => {
    zeige()

    const vorschau = screen.getByRole('heading', { name: texte.vorschau })
    const wahl = screen.getByRole('combobox', { name: texte.testkarte })

    expect(
      vorschau.compareDocumentPosition(wahl) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  })

  // **Die Diagnose steht seit FMT-02 unter dem Editor** — ihre Läufe stehen
  // in `vorlagenbefund.test.tsx` (#1691).

  it('zeigt keine Vorschau, solange es keinen gültigen Stand gab', () => {
    zeige({
      gueltig: undefined,
    })

    expect(
      screen.queryByText(kurzform(texte.vorschauErklaerung)),
    ).not.toBeInTheDocument()
  })

  it('erklärt, woher ein Name kommt, und springt zur Definition', async () => {
    const { oeffnen } = zeige({
      symbol: {
        name: 'casting cost',
        symbol: {
          name: 'casting cost',
          art: 'card field',
          datei: 'game',
          zeile: 4,
          spalte: 2,
        },
      },
      verwendungen: [{ datei: 'style', zeile: 12, text: 'casting cost:' }],
    })

    expect(screen.getByText('Kartenfeld in game, Zeile 4')).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.zurDefinition }),
    )

    expect(oeffnen).toHaveBeenCalledWith('game', 4, 2)

    await userEvent.click(
      screen.getByRole('button', { name: 'style:12 casting cost:' }),
    )

    expect(oeffnen).toHaveBeenCalledWith('style', 12, 1)
  })

  it('sagt, wenn das Paket einen Namen nicht definiert — und wenn keiner unter dem Zeiger steht', () => {
    const { unmount } = render(
      <Sprachraum sprache="de">
        <Vorlageninspector
          paket="/p"
          datei="style"
          gueltig="x"
          veraltet={null}
          karteID=""
          setzeKarteID={vi.fn()}
          karten={[]}
          kartenFehler={undefined}
          symbol={{ name: 'to_upper', symbol: undefined }}
          verwendungen={[]}
          verwendungenLaden
          oeffnen={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(
      screen.getByText(
        '„to_upper“ definiert dieses Paket nicht — der Name kommt aus der Bibliothek oder aus einem eingebundenen Paket.',
      ),
    ).toBeInTheDocument()
    expect(screen.getByText(texte.verwendungLaden)).toBeInTheDocument()
    expect(
      screen.getByText(kurzform(texte.testkarteErklaerung)),
    ).toBeInTheDocument()

    unmount()
    raeumeApi()
    zeige({ symbol: { name: 'x', symbol: undefined }, verwendungen: [] })

    expect(screen.getByText(texte.verwendungenKeine)).toBeInTheDocument()
  })

  it('nennt eine unbekannte Art schlicht Definition', () => {
    zeige({
      symbol: {
        name: 'x',
        symbol: { name: 'x', art: 'sonst', datei: 'game', zeile: 9, spalte: 1 },
      },
    })

    expect(screen.getByText('Definition in game, Zeile 9')).toBeInTheDocument()
  })

  it('nennt den Fehler der Testkarten', () => {
    zeige({ kartenFehler: 'der Bestand ist zu', symbol: null })

    expect(screen.getByRole('alert')).toHaveTextContent('der Bestand ist zu')
    expect(screen.getByText(texte.herkunftKeine)).toBeInTheDocument()
  })
})
