// **Der Verlauf zeigt die jüngsten zehn Schritte** — UXR-06
// (#1914).
//
// # Der Befund
//
// Am Probebestand steht der Verlauf leer unter dem Arbeitsplatz, und die
// Messung sah nichts. **Ein Set, an dem gearbeitet wird, hält 200 Schritte**
// (`GrenzeVerlauf` in `internal/storage/verlaufschreiben.go`) — jeder eine
// Tabellenzeile mit „Vergleichen", jeder dritte mit „Zurücknehmen". Das sind
// 200 Zeilen unter der Karte — gemessen bei 1.280 × 800 mit dieser Mischung
// eine Rolltiefe von 10,6 (R4 aus `docs/UIREVIEW.md`: 2,5) und 308
// Bedienelemente auf der Seite (R6: 200).
//
// # Was hier gehalten wird
//
// - **Die jüngsten zehn stehen da** — der oberste ist der, den ⌘Z bewegt.
// - **„10 von 200" und „Weitere 50 zeigen"** — dieselben Worte und dieselbe
//   Gestalt wie die Faltliste der Statistik (UXR-08).
// - **Wer die letzten holt, verliert den Fokus nicht**: verschwindet der
//   Knopf, geht er auf die Zahl.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import type { KartenzeileDTO, VerlaufsschrittDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Verlaufsliste } from './Verlaufsliste'

const worte = kataloge.de.bausteine.listenansicht
const texte = kataloge.de.seiten.karten.verlauf

const karten: KartenzeileDTO[] = [
  { id: '7', nummer: 0, name: 'Feuerball', felder: 1 },
]

function schritte(anzahl: number): VerlaufsschrittDTO[] {
  return Array.from({ length: anzahl }, (_, stelle) => ({
    id: String(anzahl - stelle),
    zeitpunkt: '2026-10-02T10:00:00Z',
    handlung: 'feld',
    karteID: '7',
    feld: 'name',
    vorher: `Stand ${String(stelle + 1)}`,
    gabVorher: true,
    nachher: `Stand ${String(stelle)}`,
    gabNachher: true,
    zurueckgenommen: false,
    urheber: 'fenster',
    herkunft: '',
    vorgang: stelle % 3 === 0 ? `v${String(stelle)}` : '',
    umfang: 1,
  }))
}

afterEach(() => {
  raeumeApi()
})

function zeige(anzahl: number) {
  belegeApi({
    Verlauf: () => Promise.resolve({ schritte: schritte(anzahl) }),
  })

  render(
    <Sprachraum sprache="de">
      <Verlaufsliste
        setID="1"
        karten={karten}
        verlaufsstand={0}
        zuruecknehmen={() => undefined}
      />
    </Sprachraum>,
  )
}

// zeilen zählt die Schritte in der Tabelle — ohne die Kopfzeile.
async function zeilen(): Promise<HTMLElement[]> {
  const tabelle = await screen.findByRole('table')

  return within(tabelle).getAllByRole('row').slice(1)
}

describe('Der Verlauf mit vielen Schritten', () => {
  it('zeigt bei 200 Schritten die jüngsten zehn und sagt, wie viele es sind', async () => {
    zeige(200)

    const gezeigt = await zeilen()
    expect(gezeigt).toHaveLength(10)

    // **Der jüngste zuerst** — so liefert ihn der Bestand, und ⌘Z bewegt ihn.
    expect(gezeigt[0]).toHaveTextContent('Stand 1')

    expect(
      screen.getByText(
        worte.zahl.replace('{gezeigt}', '10').replace('{gesamt}', '200'),
      ),
    ).toBeInTheDocument()

    // **Zehn Zeilen, nicht 200** — und damit höchstens zwanzig Knöpfe.
    expect(
      screen.getAllByRole('button', { name: texte.vergleichen }),
    ).toHaveLength(10)
  })

  it('holt die nächsten fünfzig und gibt am Ende den Fokus an die Zahl', async () => {
    zeige(70)
    await zeilen()

    await userEvent.click(
      screen.getByRole('button', {
        name: worte.weitere.replace('{anzahl}', '50'),
      }),
    )
    expect(await zeilen()).toHaveLength(60)

    // **Der Rest ist kleiner als ein Schritt** — der Knopf sagt, wie viele.
    await userEvent.click(
      screen.getByRole('button', {
        name: worte.weitere.replace('{anzahl}', '10'),
      }),
    )
    expect(await zeilen()).toHaveLength(70)

    const zahl = screen.getByText(
      worte.zahl.replace('{gezeigt}', '70').replace('{gesamt}', '70'),
    )
    expect(zahl).toHaveFocus()
    expect(
      screen.queryByRole('button', { name: /Weitere/ }),
    ).not.toBeInTheDocument()
  })

  it('zeigt bis zehn Schritte alle, ohne Zahl und ohne Knopf', async () => {
    zeige(10)

    expect(await zeilen()).toHaveLength(10)
    expect(
      screen.queryByText(
        worte.zahl.replace('{gezeigt}', '10').replace('{gesamt}', '10'),
      ),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: /Weitere/ }),
    ).not.toBeInTheDocument()
  })
})
