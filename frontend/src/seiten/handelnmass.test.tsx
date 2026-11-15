// **Das Prüfmaß an *Handeln*, mit echten Mengen** — UXR-15
// (#1923).
//
// # Was vorher war
//
// Am Probebestand maß die Hüllenprobe dreizehn Bedienelemente — ohne einen
// Handel und ohne ein bestätigtes Gegenüber. Mit zehn Gegenübern und zwölf
// offenen Handeln zeichnete die Spalte links **31 Bedienelemente**: die
// Liste mit fünf Umschaltern über ihr, darunter jedes Gegenüber mit seinem
// Knopf, dann Aufnehmen und Kontakte (R2 in `docs/UIREVIEW.md`). Und der
// einzige Knopf im Hauptgewicht war ein **Filter** — der gedrückte
// Umschalter „Offen" —, nicht die Handlung, um die es geht (R1).
//
// # Was jetzt gilt
//
// **Eine führende Handlung**: „Neuer Tausch …" — wie im Artboard *Handeln*.
// Die Wahl des Gegenübers, das Aufnehmen und die Kontakte stehen in der
// Schublade dahinter; die Liste zeigt acht und den Rest stückweise.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { zaehle } from '../test/bedienzaehlung'
import { handelsuebersicht, zehnGegenueber } from '../test/handelsbestand'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln

// zeigeBestand stellt *Handeln* mit dem ganzen Bestand auf.
async function zeigeBestand() {
  belegeApi({
    Peerkontakte: () => Promise.resolve(zehnGegenueber()),
    HandelUebersicht: handelsuebersicht,
  })

  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )

  // **Gewartet wird auf die Wirkung**: die erste Zeile der Liste steht.
  await screen.findByRole('button', { name: /^Anna.*Du bist dran/ })

  return screen.getByRole('tabpanel')
}

describe('Handeln mit zehn Gegenübern und zwölf offenen Handeln', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('führt mit einem Knopf und zeichnet höchstens zwölf Bedienelemente', async () => {
    const gebiet = await zeigeBestand()
    const gezaehlt = zaehle(gebiet)

    expect(gezaehlt.bedien).toBeLessThanOrEqual(12)

    // **Der einzige Knopf im Hauptgewicht ist die Handlung**, kein Filter
    // (R1) — bis hierher sah der gedrückte Umschalter „Offen" so aus.
    expect(gezaehlt.haupt).toBe(1)
    expect(
      screen.getByRole('button', { name: eigene.neuerTausch }),
    ).toHaveClass('bg-marke')
  })

  it('stellt die Gegenüber erst in der Schublade zur Wahl', async () => {
    await zeigeBestand()

    expect(
      screen.queryByRole('heading', { name: eigene.partnerTitel }),
    ).not.toBeInTheDocument()
  })
})
