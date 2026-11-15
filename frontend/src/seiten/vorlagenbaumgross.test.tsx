// **Der große Vorlagenbaum** — UXR-18
// (#1926),
// Regel R6 aus `docs/UIREVIEW.md`: mit den 526 Paketen des Full-Magic-Pack
// zeichnete die Werkstatt 35.769 Knöpfe. Ein großer Baum beginnt zugeklappt,
// sein Beiwerk steht hinter einer Zeile, und keine Liste zeigt mehr als ein
// Stück — **und alles bleibt mit Maus und Tastatur erreichbar**. Der kleine
// Baum steht in `vorlagenbaum.test.tsx` und bleibt, wie er war.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Stueck } from './vorlagenbaumfaltung'
import { Vorlagenbaum } from './Vorlagenbaum'

const texte = kataloge.de.seiten.vorlagen

function namen(anzahl: number, vorne: string, hinten = ''): string[] {
  return Array.from(
    { length: anzahl },
    (_, stelle) => `${vorne}${String(stelle)}${hinten}`,
  )
}

// zeige legt `anzahl` Pakete an, jedes mit `style` — und das erste mit so
// viel Beiwerk, wie verlangt.
function zeige({
  anzahl = Stueck + 50,
  bilder = 0,
  gewaehlt = null,
}: {
  anzahl?: number
  bilder?: number
  gewaehlt?: { paket: string; datei: string } | null
} = {}) {
  const pakete = namen(anzahl, '/pakete/p', '.mse-style')
  belegeApi({
    Paketdateien: (paket: string) =>
      Promise.resolve({
        dateien: ['style'],
        beiwerk: paket === pakete[0] ? namen(bilder, 'b', '.png') : [],
        fehler: null,
      }),
  })

  const waehlen = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Vorlagenbaum
        geladen
        fehlschlag={false}
        pakete={pakete}
        gewaehlt={gewaehlt}
        waehlen={waehlen}
        geaendert={false}
      />
    </Sprachraum>,
  )

  return waehlen
}

function weiter(gezeigt: number, gesamt: number) {
  return einsetzen(texte.weitereZeigen, {
    gezeigt: String(gezeigt),
    gesamt: String(gesamt),
  })
}

function baum() {
  return within(screen.getByRole('navigation'))
}

describe('Der große Vorlagenbaum', () => {
  afterEach(raeumeApi)

  it('beginnt zugeklappt und zeigt die Pakete stückweise', async () => {
    zeige()

    await screen.findByRole('button', { name: 'p0.mse-style' })
    const pakete = baum().getAllByRole('button', { expanded: false })

    expect(pakete).toHaveLength(Stueck)
    expect(baum().queryByRole('button', { name: 'style' })).toBeNull()
    expect(
      baum().getByRole('button', { name: weiter(Stueck, Stueck + 50) }),
    ).toBeInTheDocument()
    // **Das Maß aus R6**: gezeichnet wird höchstens die Grenze.
    expect(baum().getAllByRole('button').length).toBeLessThanOrEqual(200)
  })

  it('öffnet mit „Weitere zeigen" das nächste Stück und setzt den Fokus darauf', async () => {
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: weiter(Stueck, Stueck + 50) }),
    )

    expect(
      baum().getByRole('button', { name: `p${String(Stueck)}.mse-style` }),
    ).toHaveFocus()
    expect(baum().getAllByRole('button', { expanded: false })).toHaveLength(
      Stueck + 50,
    )
    expect(baum().queryByRole('button', { name: /^Weitere/ })).toBeNull()
  })

  it('klappt ein Paket von Hand auf und lässt die anderen zu', async () => {
    const waehlen = zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: 'p3.mse-style' }),
    )
    await userEvent.click(baum().getByRole('button', { name: 'style' }))

    expect(waehlen).toHaveBeenCalledWith('/pakete/p3.mse-style', 'style')
    expect(baum().getAllByRole('button', { name: 'style' })).toHaveLength(1)
  })

  it('stellt das Beiwerk hinter eine Zeile und erreicht es mit der Tastatur', async () => {
    const waehlen = zeige({ anzahl: 1, bilder: Stueck + 150 })

    const paket = await screen.findByRole('button', { name: 'p0.mse-style' })
    expect(paket).toHaveAttribute('aria-expanded', 'false')

    paket.focus()
    await userEvent.keyboard('{Enter}{ArrowDown}')
    expect(baum().getByRole('button', { name: 'style' })).toHaveFocus()

    await userEvent.keyboard('{ArrowDown}')
    const zeile = baum().getByRole('button', {
      name: einsetzen(texte.beiwerkZeile, { anzahl: String(Stueck + 150) }),
    })
    expect(zeile).toHaveFocus()
    expect(zeile).toHaveAttribute('aria-expanded', 'false')
    expect(
      baum().queryByRole('button', { name: `b0.png — ${texte.beiwerk}` }),
    ).toBeNull()

    await userEvent.keyboard('{Enter}{End}')
    expect(
      baum().getByRole('button', { name: weiter(Stueck, Stueck + 150) }),
    ).toHaveFocus()

    await userEvent.keyboard('{Enter}')
    const neu = baum().getByRole('button', {
      name: `b${String(Stueck)}.png — ${texte.beiwerk}`,
    })
    expect(neu).toHaveFocus()

    await userEvent.keyboard('{Enter}')
    expect(waehlen).toHaveBeenCalledWith(
      '/pakete/p0.mse-style',
      `b${String(Stueck)}.png`,
      true,
    )
  })

  it('zeigt die Wahl offen, auch hinter dem ersten Stück, und macht sie zur Tabstelle', async () => {
    zeige({
      gewaehlt: {
        paket: `/pakete/p${String(Stueck + 20)}.mse-style`,
        datei: 'style',
      },
    })

    const datei = await screen.findByRole('button', { name: 'style' })
    expect(datei).toHaveAttribute('aria-current', 'true')

    screen.getByLabelText(texte.baumSuche).focus()
    await userEvent.tab()
    expect(datei).toHaveFocus()
  })

  // **Auch ein kleiner Baum kürzt eine lange Liste** — `magic-modules`
  // bringt 143 Textdateien mit (gezählt am 2026-10-02 über das
  // Full-Magic-Pack).
  it('zeigt die Dateien eines Pakets stückweise', async () => {
    belegeApi({
      Paketdateien: () =>
        Promise.resolve({
          dateien: namen(Stueck + 40, 'd'),
          beiwerk: [],
          fehler: null,
        }),
    })
    render(
      <Sprachraum sprache="de">
        <Vorlagenbaum
          geladen
          fehlschlag={false}
          pakete={['/pakete/modules.mse-include']}
          gewaehlt={null}
          waehlen={vi.fn()}
          geaendert={false}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: weiter(Stueck, Stueck + 40) }),
    )

    expect(
      baum().getByRole('button', { name: `d${String(Stueck)}` }),
    ).toHaveFocus()
    expect(
      baum().getByRole('button', { name: `d${String(Stueck + 39)}` }),
    ).toBeInTheDocument()
  })

  it('klappt auf, was der Filter übrig lässt, solange es wenig ist', async () => {
    zeige()

    await screen.findByRole('button', { name: 'p0.mse-style' })
    await userEvent.type(screen.getByLabelText(texte.baumSuche), 'p12')

    // p12 und p120 bis p129 — elf Pakete, je mit `style`.
    expect(baum().getAllByRole('button', { name: 'style' })).toHaveLength(11)
    expect(baum().getAllByRole('button', { expanded: true })).toHaveLength(11)
  })
})
