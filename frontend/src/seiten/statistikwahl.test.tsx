// **Die Wahl „Worüber" mit Suche** — UXR-08
// (#1916),
// Regel R7 aus `docs/UIREVIEW.md`: eine Auswahl **mit Suche ab fünfzehn**
// Werten.
//
// # Warum fünfzehn hier erreicht werden
//
// Am Probebestand nicht: er hat wenige Felder und wenige Dimensionen. Ein
// echtes Set schon — `magic.mse-game` erklärt **fünfzig** Dimensionen
// (`statistics_en`, gezählt am 2026-10-02), und `White.mse-set` trägt in gut
// vierzig Feldern einen Wert (aus der Setdatei gezählt am selben Tag). Die
// Fälle hier nehmen dieselben Mengen mit eigenen Namen; die des Pakets kommen
// nicht ins Repositorium (§11).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { Statistik } from './Statistik'
import { Statistikwahl, belegungswahl } from './Statistikwahl'

const texte = kataloge.de.seiten.statistik

// namen sind `anzahl` Namen mit Vorsatz — „Farbe 1", „Farbe 2" …
function namen(vorsatz: string, anzahl: number): string[] {
  return Array.from({ length: anzahl }, (_, stelle) =>
    [vorsatz, String(stelle + 1)].join(' '),
  )
}

// echt sind die Mengen eines echten Sets: fünfzig Dimensionen, vierzig
// Felder — darunter je eine mit „Mana" im Namen.
const echt = {
  dimensionen: ['Manawert', ...namen('Dimension', 49)],
  felder: ['mana cost', ...namen('Feld', 39)],
}

function Probe({
  felder,
  dimensionen,
  anfang = belegungswahl,
}: {
  felder: string[]
  dimensionen: string[]
  anfang?: string
}) {
  const [wahl, setzeWahl] = useState(anfang)

  return (
    <Sprachraum sprache="de">
      <Statistikwahl
        wahl={wahl}
        waehlen={setzeWahl}
        felder={felder}
        dimensionen={dimensionen}
      />
    </Sprachraum>
  )
}

// zeilen nennt, was die geöffnete Liste der Auswahl mit Suche zeichnet.
function zeilen(): string[] {
  return screen.getAllByRole('option').map((eine) => eine.textContent)
}

const wahlfeld = () => screen.getByRole('combobox', { name: texte.worueber })

describe('Die Wahl „Worüber"', () => {
  // **Unter fünfzehn bleibt es das native Feld** mit seinen zwei Gruppen —
  // eine Suche wäre ein Bedienelement mehr für eine Liste, die man auf einen
  // Blick überschaut.
  it('bleibt unter fünfzehn Einträgen ein natives Feld mit Gruppen', () => {
    render(
      <Probe felder={namen('Feld', 6)} dimensionen={namen('Dimension', 7)} />,
    )

    expect(wahlfeld().tagName).toBe('SELECT')
    expect(within(wahlfeld()).getAllByRole('option')).toHaveLength(14)
    expect(
      within(wahlfeld()).getByRole('group', { name: texte.gruppeFelder }),
    ).toBeInTheDocument()
  })

  it('lässt eine leere Gruppe weg', () => {
    render(<Probe felder={[]} dimensionen={namen('Dimension', 3)} />)

    expect(
      within(wahlfeld()).queryByRole('group', { name: texte.gruppeFelder }),
    ).toBeNull()
    expect(
      within(wahlfeld()).getByRole('group', { name: texte.gruppeSpielpaket }),
    ).toBeInTheDocument()
  })

  it('wählt im nativen Feld', async () => {
    render(
      <Probe felder={namen('Feld', 2)} dimensionen={namen('Dimension', 2)} />,
    )

    await userEvent.selectOptions(wahlfeld(), 'feld:Feld 2')

    expect(wahlfeld()).toHaveValue('feld:Feld 2')
  })

  // **Ab fünfzehn eine Auswahl mit Suche** (UXR-F1, #2007).
  it('wird ab fünfzehn Einträgen eine Auswahl mit Suche', async () => {
    render(
      <Probe felder={namen('Feld', 7)} dimensionen={namen('Dimension', 7)} />,
    )

    expect(wahlfeld().tagName).toBe('INPUT')
    expect(wahlfeld()).toHaveValue(texte.belegungWahl)

    await userEvent.click(wahlfeld())

    expect(zeilen()).toHaveLength(15)
  })

  // **Dieselbe Bezeichnung in zwei Gruppen** (`rarity` ist Dimension und
  // Feld): die Gruppe steht blass hinter dem Namen und wird mitgesucht.
  it('nennt in jeder Zeile die Gruppe, auch bei gleichem Namen', async () => {
    render(
      <Probe
        felder={['rarity', ...namen('Feld', 13)]}
        dimensionen={['rarity', ...namen('Dimension', 2)]}
      />,
    )

    await userEvent.click(wahlfeld())
    await userEvent.keyboard('rarity')

    expect(zeilen()).toEqual([
      `rarity ${texte.gruppeSpielpaket}`,
      `rarity ${texte.gruppeFelder}`,
    ])
  })

  // **Ohne Rücksicht auf Groß und Klein**, über Dimensionen und Felder
  // zugleich — wer „mana" sucht, weiß nicht, in welcher Gruppe es steht.
  it('grenzt beide Gruppen zugleich ein und wählt', async () => {
    render(<Probe {...echt} />)

    await userEvent.click(wahlfeld())

    expect(zeilen()).toHaveLength(91)

    await userEvent.keyboard('MANA')

    expect(zeilen()).toEqual([
      `Manawert ${texte.gruppeSpielpaket}`,
      `mana cost ${texte.gruppeFelder}`,
    ])

    await userEvent.click(screen.getByRole('option', { name: /^Manawert/ }))

    expect(wahlfeld()).toHaveValue('Manawert')
  })

  it('wählt ein Feld und eine Dimension unter ihrem Vorsatz', async () => {
    const waehlen = vi.fn()
    render(
      <Sprachraum sprache="de">
        <Statistikwahl
          wahl={belegungswahl}
          waehlen={waehlen}
          felder={echt.felder}
          dimensionen={echt.dimensionen}
        />
      </Sprachraum>,
    )

    await userEvent.click(wahlfeld())
    await userEvent.click(screen.getByRole('option', { name: /^Feld 3 / }))
    await userEvent.click(wahlfeld())
    await userEvent.click(screen.getByRole('option', { name: /^Dimension 4 / }))

    expect(waehlen).toHaveBeenNthCalledWith(1, 'feld:Feld 3')
    expect(waehlen).toHaveBeenNthCalledWith(2, 'dimension:Dimension 4')
  })

  // **Die gewählte steht im Feld**, auch wenn die Suche sie ausblendet — und
  // es steht da, dass nichts passt. Escape stellt die Wahl her.
  it('behält die gewählte und sagt, wenn nichts passt', async () => {
    const { container } = render(<Probe {...echt} anfang="feld:Feld 3" />)

    expect(wahlfeld()).toHaveValue('Feld 3')

    await userEvent.click(wahlfeld())
    await userEvent.keyboard('gibtsnicht')

    expect(screen.queryAllByRole('option')).toHaveLength(0)
    expect(
      screen.getAllByText(
        kataloge.de.bausteine.suchauswahl.keiner.replace(
          '{suche}',
          'gibtsnicht',
        ),
      ),
    ).not.toHaveLength(0)

    await userEvent.keyboard('{Escape}')

    expect(wahlfeld()).toHaveValue('Feld 3')

    // **Mit Suche und leerem Treffer hält die Wahl WCAG ein.**
    expect(await verstoesse(container)).toEqual([])
  })

  // **Ein Setwechsel räumt auch die Eingabe** (§7, Regel 1) — sie grenzte
  // sonst die Felder des neuen Sets nach dem Wort für das alte ein.
  it('räumt die Eingabe, wenn das Set wechselt', async () => {
    belegeApi({
      Statistik: () =>
        Promise.resolve({
          statistik: {
            karten: 40,
            felder: 40,
            leere: 0,
            belegung: echt.felder.map((name) => ({ name, karten: 1 })),
          },
        }),
      Auszaehldimensionen: () =>
        Promise.resolve({ dimensionen: [], fehler: null }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Statistik setID="7" />
      </Sprachraum>,
    )

    await userEvent.click(await screen.findByRole('combobox'))
    await userEvent.keyboard('mana')
    expect(zeilen()).toHaveLength(1)

    rerender(
      <Sprachraum sprache="de">
        <Statistik setID="8" />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    })
    expect(await screen.findByRole('combobox')).toHaveValue(texte.belegungWahl)
  })
})

afterEach(() => {
  raeumeApi()
})
