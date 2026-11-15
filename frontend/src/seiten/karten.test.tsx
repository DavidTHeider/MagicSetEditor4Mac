// Die Kartenliste und der Kartenname.
// Die Kartenseite: Liste, Anhakung, Wahl und Editor.
//
// Von `seiten.test.tsx` getrennt, als die Datei über 400 Zeilen wuchs
// (AGENTS.md §2). Der Schnitt folgt der Frage, die eine Datei beantwortet:
// dort die Setliste, hier die Karten.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { type Api } from '../api'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'
import { ausKarten } from '../test/kartenproben'

// belegeSeite belegt zusätzlich, was die Kartenseite nebenbei lädt.
//
// **Ohne Beleg meldete die Fassade „die Methode ist nicht belegt"**, und ein
// Test über die Kartenliste fände zwei Warnungen statt einer — genau dafür ist
// die Vorrichtung streng (§3).
//
// `Kartenfelder` steht seit #560
// dabei: die Seite lädt die **Formen** der Felder, damit der Editor beim
// Kartenwechsel schon weiß, welches Steuerelement ein Feld bekommt.
function belegeSeite(teile: Partial<Api>) {
  return belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    ...teile,
  })
}

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function probeKarte(teile: Partial<KarteDTO> = {}): KarteDTO {
  return {
    id: '5',
    nummer: 0,
    werte: [{ feld: 'name', wert: 'Feuerball' }],
    ...teile,
  }
}

function zeigeKarten(setID: string) {
  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID={setID}
        karteID={undefined}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

describe('Kartenliste', () => {
  // **Der Fall „kein Set" steht hier nicht mehr.** Seit
  // #288
  // rendert der Arbeitsplatz diese Lasche nur mit offenem Set; der leere
  // Zustand steht in `setarbeitsplatz.test.tsx`.

  it('zeigt die Karten mit ihrer Nummer', async () => {
    belegeSeite({
      ...ausKarten(() =>
        Promise.resolve({
          karten: [probeKarte(), probeKarte({ id: '6', nummer: 1 })],
        }),
      ),
      Karten: () =>
        Promise.resolve({
          karten: [probeKarte(), probeKarte({ id: '6', nummer: 1 })],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    })

    zeigeKarten('1')

    await waitFor(() => {
      expect(screen.getAllByRole('row')).toHaveLength(2)
    })

    // **Die Nummer beginnt bei eins.** Im Bestand steht sie ab null, weil sie
    // dort eine Reihenfolge ist; auf der Karte ist sie eine Zählung.
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  // Eine Karte ohne Feldwerte hat Phase 2 eigens aufgehoben — sie soll auch zu
  // sehen sein.
  // **Der Schnellfilter verbirgt, was Go nicht nennt** — Werkstück U5 (#1416).
  it('verbirgt über den Schnellfilter die Karten, die Go nicht nennt', async () => {
    const KartenFiltern = vi.fn().mockResolvedValue({ kartenIDs: ['6'] })
    const zwei = [
      probeKarte(),
      probeKarte({
        id: '6',
        nummer: 1,
        werte: [{ feld: 'name', wert: 'Blitz' }],
      }),
    ]
    belegeSeite({
      ...ausKarten(() => Promise.resolve({ karten: zwei })),
      Karten: () => Promise.resolve({ karten: zwei }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      KartenFiltern,
    })
    zeigeKarten('1')
    await waitFor(() => {
      expect(screen.getAllByRole('row')).toHaveLength(2)
    })

    await userEvent.type(
      screen.getByLabelText(texte.seiten.karten.schnellfilter.bezeichnung),
      'blitz',
    )

    await waitFor(() => {
      expect(screen.getAllByRole('row')).toHaveLength(1)
    })
    expect(KartenFiltern).toHaveBeenLastCalledWith('1', 'blitz')
    expect(screen.getByRole('status', { name: '' })).toHaveTextContent(
      '1 von 2 Karten',
    )

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.seiten.karten.schnellfilter.leeren,
      }),
    )

    await waitFor(() => {
      expect(screen.getAllByRole('row')).toHaveLength(2)
    })
  })

  it('zeigt eine Karte ohne Feldwerte und sagt es dazu', async () => {
    belegeSeite({
      ...ausKarten(() =>
        Promise.resolve({ karten: [probeKarte({ werte: [] })] }),
      ),
      Karten: () => Promise.resolve({ karten: [probeKarte({ werte: [] })] }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    })

    zeigeKarten('1')

    await waitFor(() => {
      expect(screen.getByText(texte.seiten.karten.ohneName)).toBeInTheDocument()
    })

    expect(screen.getByText(texte.seiten.karten.ohneFelder)).toBeInTheDocument()
  })

  it('erklärt den nächsten Schritt, wenn das Set keine Karte hat', async () => {
    belegeSeite({
      ...ausKarten(() => Promise.resolve({ karten: [] })),
      Karten: () => Promise.resolve({ karten: [] }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    })

    zeigeKarten('1')

    await waitFor(() => {
      expect(screen.getByText(texte.seiten.karten.leer)).toBeInTheDocument()
    })
  })
})
