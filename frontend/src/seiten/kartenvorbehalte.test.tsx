// **Die Vorbehalte stehen nicht mehr auf der Kartenlasche** — Werkstück I von
// Phase 4.10
// (#1325).
//
// # Warum sie umgezogen sind
//
// Sie handeln vom **Set**, und die Kartenlasche handelt von Karten. Der Kasten
// stand dort am Ende einer 1.728 Punkte hohen Seite, unter drei ungleich
// langen Spalten — ein Element, das thematisch nicht dazugehört.
//
// Was **hier** noch geprüft wird: dass die Kartenlasche ihn nicht mehr zeigt.
// Dass die Setangaben ihn zeigen, prüft `setvorbehalte.test.tsx`.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { type Api } from '../api'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'
import { kartenname } from './kartenname'
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

describe('die Vorbehalte der Kartenseite', () => {
  // **Die Kartenlasche zeigt sie nicht mehr** (Werkstück I von Phase 4.10).
  //
  // Sie fragt auch nicht mehr danach: eine Bindung, die niemand anzeigt,
  // kostet einen Weg zur Datenbank und eine Fehlerquelle. `belegeApi` ist
  // streng — belegte man `Vorbehalte` nicht und die Seite riefe es doch,
  // scheiterte der Lauf mit dem Namen der Methode (§3).
  it('fragt nicht mehr nach den Vorbehalten', async () => {
    const Vorbehalte = vi.fn(() =>
      Promise.resolve({ vorbehalte: [], fehler: null }),
    )

    belegeSeite({
      ...ausKarten(() => Promise.resolve({ karten: [probeKarte()] })),
      Karten: () => Promise.resolve({ karten: [probeKarte()] }),
      Vorbehalte,
    })

    zeigeKarten('1')

    // **Gewartet wird auf die Wirkung** (§3): die Karte steht da, also ist die
    // Seite fertig — und erst dann sagt die Verneinung etwas.
    expect(await screen.findByText('Feuerball')).toBeInTheDocument()

    expect(Vorbehalte).not.toHaveBeenCalled()
  })
})

// **Die aufgeschlagene Karte kommt aus ihrer eigenen Bindung**
// (#891).
//
// Bis zum 2026-08-25 trug die Liste jeden Feldwert jeder Karte mit; seitdem
// gibt `Kartenzeilen` Namen und Feldzahl, und der Editor holt sich die eine
// Karte, die offen ist.
describe('Die aufgeschlagene Karte', () => {
  it('holt ihre Feldwerte über `Karte`, nicht über die Liste', async () => {
    const Karte = vi.fn(() =>
      Promise.resolve({
        karte: {
          id: '5',
          nummer: 0,
          werte: [{ feld: 'name', wert: 'Aus der Einzelbindung' }],
        },
      }),
    )

    belegeSeite({
      Kartenzeilen: () =>
        Promise.resolve({
          zeilen: [{ id: '5', nummer: 0, name: 'Aus der Liste', felder: 1 }],
        }),
      Karte,
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
    })

    zeigeKarten('1')

    // **Der Wert im Formular stammt aus `Karte`.** Stünde dort „Aus der
    // Liste", käme er aus der Zeile — und die trägt ihn gar nicht mehr.
    expect(
      await screen.findByDisplayValue('Aus der Einzelbindung'),
    ).toBeInTheDocument()

    expect(Karte).toHaveBeenCalledWith('5')
  })

  // **Nach einer Änderung wird beides nachgeladen** — die Liste, weil der Name
  // in ihr steht, und die Karte, weil ein Schreibvorgang auch **andere** Felder
  // setzen kann (Scryfall-Übernahme, Bildwahl).
  //
  // Ohne das Nachladen der Karte zeigte das Formular den Stand von vor der
  // Übernahme, während im Bestand der neue steht.
  it('lädt nach einer Feldänderung Liste **und** Karte nach', async () => {
    const Kartenzeilen = vi.fn(() =>
      Promise.resolve({
        zeilen: [{ id: '5', nummer: 0, name: 'Feuerball', felder: 1 }],
      }),
    )
    const Karte = vi.fn(() => Promise.resolve({ karte: probeKarte() }))

    belegeSeite({
      Kartenzeilen,
      Karte,
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
      FeldSetzen: () => Promise.resolve({}),
    })

    zeigeKarten('1')

    const feld = await screen.findByRole('textbox', { name: 'name' })

    const vorher = {
      zeilen: Kartenzeilen.mock.calls.length,
      karte: Karte.mock.calls.length,
    }

    await userEvent.clear(feld)
    await userEvent.type(feld, 'Eisblitz')
    await userEvent.tab()

    await waitFor(() => {
      expect(Karte.mock.calls.length).toBeGreaterThan(vorher.karte)
    })

    expect(Kartenzeilen.mock.calls.length).toBeGreaterThan(vorher.zeilen)
  })
})

describe('kartenname', () => {
  it('nimmt das Feld `name`', () => {
    expect(kartenname(probeKarte(), 'Rückfall')).toBe('Feuerball')
  })

  it('fällt zurück, wenn es keinen Namen gibt', () => {
    expect(kartenname(probeKarte({ werte: [] }), 'Rückfall')).toBe('Rückfall')
  })

  // Ein Name aus lauter Leerzeichen ist keiner — er ergäbe eine Zeile, auf der
  // nichts steht.
  it('fällt auch bei einem leeren Namen zurück', () => {
    expect(
      kartenname(
        probeKarte({ werte: [{ feld: 'name', wert: '   ' }] }),
        'Rückfall',
      ),
    ).toBe('Rückfall')
  })

  // **Die Meldung einer gescheiterten Rücknahme steht hier** — dort, wo ⌘Z
  // wirkt, und nicht als anonymer Hinweis über dem Fenster (§6).
  it('zeigt die Meldung einer gescheiterten Rücknahme', async () => {
    belegeSeite({
      ...ausKarten(() =>
        Promise.resolve({ karten: [probeKarte()], fehler: null }),
      ),
      Karten: () => Promise.resolve({ karten: [probeKarte()], fehler: null }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler="die Platte ist voll"
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'die Platte ist voll',
    )
  })

  // **Ein höherer Verlaufsstand lädt die Liste neu.**
  //
  // Nicht, weil die Anzeige sich selbst zurechtrückte: was die Rücknahme
  // geändert hat, steht im Bestand, und gelesen wird es von dort.
  it('lädt die Karten neu, wenn der Verlaufsstand steigt', async () => {
    let name = 'Eisball'
    belegeSeite({
      ...ausKarten(() =>
        Promise.resolve({
          karten: [probeKarte({ werte: [{ feld: 'name', wert: name }] })],
          fehler: null,
        }),
      ),
      Karten: () =>
        Promise.resolve({
          karten: [probeKarte({ werte: [{ feld: 'name', wert: name }] })],
          fehler: null,
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    expect(
      await screen.findByRole('button', { name: /Eisball/ }),
    ).toBeInTheDocument()

    name = 'Feuerball'

    rerender(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={1}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: /Feuerball/ }),
      ).toBeInTheDocument()
    })
  })
})
