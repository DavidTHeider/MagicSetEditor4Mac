// **Vom Symbolfeld einer Karte zum Symboleditor** — Werkstück A von Phase
// 4.11 (#1339).
//
// # Warum dieser Lauf durch die ganze Anwendung geht
//
// Weil die Zusage über **vier** Bauteile läuft: die Feldzeile zeigt den Knopf,
// die Kartenarbeitsfläche hängt die Kartenkennung dran, die Anwendung merkt
// sich das Ziel und wechselt die Seite, und der Symboleditor liest daraufhin
// den **Feldwert der Karte** statt der Setangabe.
//
// Jedes Stück hat seinen eigenen Lauf. **Keiner von ihnen sagt, dass der Weg
// trägt** — und genau das ist die Abnahme aus #1339: „die Karte zeigt ihre
// eigene Zeichnung".
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { ausKarten } from './test/kartenproben'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { Sprachraum } from './sprache/kontext'
import { leisteneintrag } from './test/leistenweg'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

// symbolform ist ein Kartenfeld der Art `symbol`.
function symbolform() {
  return {
    name: 'kartensymbol',
    aufschrift: 'Kartensymbol',
    art: 'symbol',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
  }
}

it('führt vom Symbolfeld einer Karte in den Symboleditor', async () => {
  const Symbolstand = vi.fn().mockResolvedValue({
    teile: [],
    name: 'eigen.mse-symbol',
    fehler: null,
  })

  const karte = {
    id: '1',
    nummer: 0,
    werte: [
      { feld: 'name', wert: 'Prüfstein' },
      { feld: 'kartensymbol', wert: 'eigen.mse-symbol' },
    ],
  }

  belegeApi({
    ...belegeFenster('de'),
    Sets: () =>
      Promise.resolve({
        sets: [
          {
            id: '42',
            name: 'Gewähltes Set',
            spiel: 'magic',
            stil: 'm15',
            herkunft: '',
            karten: 1,
            herkunftFehlt: false,
          },
        ],
      }),
    ...ausKarten(() => Promise.resolve({ karten: [karte] })),
    Kartenfelder: () => Promise.resolve({ felder: [symbolform()] }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Symbol: Symbolstand,
  })

  render(
    <Sprachraum sprache="de">
      <App />
    </Sprachraum>,
  )

  await userEvent.click(await leisteneintrag('sets'))

  await userEvent.click(
    await screen.findByRole('button', { name: /^Gewähltes Set —/ }),
  )

  await userEvent.click(
    await screen.findByRole('button', {
      name: `Kartensymbol — ${texte.seiten.karten.symbolZeichnen}`,
    }),
  )

  // **Der Editor lädt die Zeichnung des Kartenfeldes** — nicht die des Sets.
  await waitFor(() => {
    expect(Symbolstand).toHaveBeenCalledWith('42', 'eigen.mse-symbol')
  })
})
