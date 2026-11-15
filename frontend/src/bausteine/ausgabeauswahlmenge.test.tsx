// **Welche Karten — bei einem Set, wie es wirklich ist** — UXR-09
// (#1917),
// R6 und R3 aus `docs/UIREVIEW.md`.
//
// # Der Befund
//
// Die Messung am Probebestand fand nichts: `saat/probe-werkstatt.mse-set`
// hat eine Handvoll Karten, und „Nur ausgewählte" zeichnete eine Handvoll
// Kästchen. **Ein Set hat dreihundert** — ein Grundset von Magic 250 bis 280,
// dazu Marken und Rückseiten. Die Teilmenge zeichnete dann je Karte ein
// Kästchen: dreihundert Felder, dreihundert Tabstopps, und R6 erlaubt auf
// einer Fläche zweihundert Bedienelemente.
//
// **Was an ihre Stelle tritt**: ein Feld, das eingrenzt, und höchstens
// sechzig Zeilen auf einmal — weitere auf Knopfdruck. Gewählt bleibt, was
// gewählt ist, auch wenn die Eingrenzung es ausblendet.
//
// **Die Gestalt der Sets- und Deckliste** (UX-11.AK-03): dieselbe
// Seitengröße, dieselben Sätze — „60 von 300", „Weitere 60 zeigen", „Nichts
// passt zu …". Eine zweite Fassung derselben Bedienung wäre eine, die
// auseinanderläuft.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { belegeApi, raeumeApi } from '../test/api'
import { Ausgabeauswahl } from './Ausgabeauswahl'

const texte = kataloge.de.bausteine.ausgabeauswahl
const liste60 = kataloge.de.bausteine.listenansicht

// set gibt so viele Kartenzeilen, wie ein Set hat.
function set(anzahl: number) {
  return Array.from({ length: anzahl }, (_, stelle) => ({
    id: String(stelle + 1),
    nummer: stelle,
    name: `Karte ${String(stelle + 1)}`,
    felder: 2,
  }))
}

function zeige(anzahl: number, gewaehlt: readonly string[] = ['1']) {
  const setzeGewaehlt = vi.fn()

  belegeApi({
    Kartenzeilen: () => Promise.resolve({ zeilen: set(anzahl), fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Ausgabeauswahl
        setID="1"
        gewaehlt={gewaehlt}
        setzeGewaehlt={setzeGewaehlt}
        angehalten={false}
      />
    </Sprachraum>,
  )

  return setzeGewaehlt
}

// liste ist die Liste der Kästchen — die Teilmenge.
async function liste() {
  return screen.findByRole('list', { name: texte.aufschrift })
}

afterEach(() => {
  raeumeApi()
})

describe('die Teilmenge eines großen Sets', () => {
  it('zeichnet höchstens sechzig Kästchen', async () => {
    zeige(300)

    expect(within(await liste()).getAllByRole('checkbox')).toHaveLength(60)
    expect(
      screen.getByText(
        einsetzen(liste60.zahl, { gezeigt: '60', gesamt: '300' }),
      ),
    ).toBeInTheDocument()
  })

  it('zeigt auf Knopfdruck sechzig weitere', async () => {
    zeige(300)

    await userEvent.click(
      await screen.findByRole('button', {
        name: einsetzen(liste60.weitere, { anzahl: '60' }),
      }),
    )

    expect(within(await liste()).getAllByRole('checkbox')).toHaveLength(120)
  })

  // **Eingegrenzt wird über den Namen**, groß oder klein geschrieben.
  it('grenzt über den Namen ein', async () => {
    zeige(300)

    await userEvent.type(
      await screen.findByRole('searchbox', { name: texte.eingrenzen }),
      'karte 27',
    )

    const kaestchen = within(await liste()).getAllByRole('checkbox')

    // 27 und 270 bis 279.
    expect(kaestchen).toHaveLength(11)
    expect(kaestchen[0]).toHaveAccessibleName('Karte 27')
  })

  // **Eine gewählte Karte bleibt gewählt**, auch wenn die Eingrenzung sie
  // ausblendet — die Zahl an „Nur ausgewählte" sagt es.
  it('behält eine ausgeblendete Wahl', async () => {
    const setzeGewaehlt = zeige(300, ['1', '2'])

    await userEvent.type(
      await screen.findByRole('searchbox', { name: texte.eingrenzen }),
      'Karte 280',
    )
    await userEvent.click(screen.getByRole('checkbox', { name: 'Karte 280' }))

    expect(setzeGewaehlt).toHaveBeenLastCalledWith(['1', '2', '280'])
  })

  it('sagt, wenn nichts passt', async () => {
    zeige(300)

    await userEvent.type(
      await screen.findByRole('searchbox', { name: texte.eingrenzen }),
      'Drache',
    )

    expect(
      screen.getByText(einsetzen(liste60.nichtsPasst, { suche: 'Drache' })),
    ).toBeInTheDocument()
  })

  // **Ein kleines Set braucht kein Suchfeld** — dieselbe Grenze wie in R7:
  // ab fünfzehn Werten. Die Kästchen daneben belegen, dass die Liste steht.
  it('fragt bei wenigen Karten nicht nach einer Eingrenzung', async () => {
    zeige(14)

    expect(within(await liste()).getAllByRole('checkbox')).toHaveLength(14)
    expect(screen.queryByRole('searchbox')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /60/ })).toBeNull()
  })
})
