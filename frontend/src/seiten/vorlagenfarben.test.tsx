// **Der Vorlagenbereich färbt nach dem Paketformat** —
// #1181.
//
// # Der Befund
//
// `Quelltextfeld` nimmt eine Sprache, und dort stand `keine` — Zeilennummern
// ja, Farbe nein. Das war richtig herum geirrt: der Skript-Lexer liest aus
// `game: magic` Bezeichner und Doppelpunkte und färbt nichts Sinnvolles, und
// keine Farbe ist besser als eine falsche.
//
// # Was hier geprüft wird
//
// Dass die **Paketbindung** gefragt wird und nicht die Skriptbindung — die
// Naht, an der die beiden Sprachen auseinandergehen.
import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Quelltextfeld } from '../bausteine/Quelltextfeld'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(raeumeApi)

describe('das Quelltextfeld im Paketformat', () => {
  it('fragt die Paketbindung und nicht die Skriptbindung', async () => {
    const Paketfarben = vi.fn().mockResolvedValue({
      merkmale: [{ art: 'Schlüssel', zeile: 1, spalte: 1, breite: 4 }],
    })
    const Skriptfarben = vi.fn().mockResolvedValue({ merkmale: [] })

    belegeApi({ Paketfarben, Skriptfarben })

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          sprache="paket"
          bezeichnung="Inhalt"
          wert="game: magic"
          setzeWert={vi.fn()}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(Paketfarben).toHaveBeenCalledWith('game: magic')
    })

    expect(Skriptfarben).not.toHaveBeenCalled()
  })

  it('färbt den Schlüssel', async () => {
    belegeApi({
      Paketfarben: () =>
        Promise.resolve({
          merkmale: [{ art: 'Schlüssel', zeile: 1, spalte: 1, breite: 4 }],
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          sprache="paket"
          bezeichnung="Inhalt"
          wert="game: magic"
          setzeWert={vi.fn()}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(
        screen.getByTestId('quelltextanzeige').querySelector('span')?.className,
      ).toContain('marke')
    })
  })
})
