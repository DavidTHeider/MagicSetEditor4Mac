// **Die Vorbehalte des Imports stehen bei den Setangaben** — Werkstück I von
// Phase 4.10
// (#1325).
//
// # Warum sie umgezogen sind
//
// Sie handeln vom **Set**, und die Kartenlasche handelt von Karten. Der Kasten
// stand dort am Ende einer 1.728 Punkte hohen Seite, unter drei ungleich
// langen Spalten — ein Element, das thematisch nicht dazugehört.
//
// **Die Zusage selbst ist älter** (#39): was der Import nicht gedeutet hat,
// ist zu sehen und reist unverändert mit. Seit Werkstück H steht die **Zahl**
// offen und die Zeilen liegen hinter einem Schalter.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Setangaben } from './Setangaben'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { leererSetkopf } from '../test/setkopfprobe'

import type { Api, VorbehaltDTO } from '../api'

const texte = kataloge.de

afterEach(() => {
  raeumeApi()
})

// zeigeAngaben zeichnet die Setangaben über den gegebenen Belegen.
function zeigeAngaben(teile: Partial<Api> = {}) {
  belegeApi({
    Setangaben: () =>
      Promise.resolve({
        felder: [],
        werte: [],
        kopf: leererSetkopf,
        fehler: null,
      }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Setangaben setID="1" verlaufsstand={0} />
    </Sprachraum>,
  )
}

describe('die Vorbehalte bei den Setangaben', () => {
  // **Die Vorbehalte sind sichtbar** — das ist die Zusage aus #39.
  //
  // **Seit Werkstück H von Phase 4.10 erst nach einem Klick**
  // (#1325).
  // Der Block stand offen am Ende der Kartenseite und zeigte bei `Vanguard`
  // fünf Zeilen Rohdaten — für die meisten uninteressant, und die Seite ist
  // 2.190 Punkte hoch. **Die Zahl bleibt aber oben stehen**: dass etwas nicht
  // gedeutet wurde, ist die Nachricht; welche Zeile es war, ist die Antwort
  // auf eine Frage, die man erst stellt.
  it('zeigt, was der Import nicht gedeutet hat, nach einem Klick', async () => {
    const vorbehalt: VorbehaltDTO = {
      datei: 'set',
      zeile: 12,
      inhalt: 'unverstanden: wert',
    }

    zeigeAngaben({
      Vorbehalte: () => Promise.resolve({ vorbehalte: [vorbehalt] }),
    })

    // **Die Zahl steht da, bevor jemand klickt.**
    const schalter = await screen.findByRole('button', {
      name: einsetzen(texte.seiten.sets.vorbehalte.mitZahl, { anzahl: '1' }),
    })

    expect(screen.queryByText('unverstanden: wert')).not.toBeInTheDocument()

    await userEvent.click(schalter)

    expect(screen.getByText('unverstanden: wert')).toBeInTheDocument()
    expect(screen.getByText('set:12')).toBeInTheDocument()
  })

  // **Ohne Vorbehalte steht kein Schalter** — nur der Satz, dass alles
  // gelesen wurde. Ein Aufklappkasten, hinter dem nichts liegt, ist ein
  // Bedienelement, das etwas verspricht, was nicht kommt.
  it('sagt es, wenn ein Set vollständig gelesen wurde', async () => {
    zeigeAngaben({
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    })

    await waitFor(() => {
      expect(
        screen.getByText(texte.seiten.sets.vorbehalte.keine),
      ).toBeInTheDocument()
    })
  })
})
