// **Beim Kartenwechsel blitzt die Vorschau nicht leer** — AP-4
// (#2361),
// Abnahme „Vorschauwechsel: Lage vorher/nachher konstant, kein Leerblitz“.
//
// Eine andere Karte ist eine neue Frage, und `useKartenvorschau` räumt —
// das bleibt so, denn §7 Regel 1 verlangt es: kein Feld, kein Befund und
// keine Vorgabe der vorigen Karte darf für die neue gelten. **Bis die neue
// Antwort da war, stand aber gar kein Bild da**, und die Spalte fiel
// zusammen. Jetzt bleibt das Bild der vorigen Karte stehen — **nur zum
// Ansehen**: gedämpft, `inert`, `aria-busy`, ohne Kontextmenü.
import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KartenlisteAntwort, VorschauAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { form, vorschau } from '../test/bildflaeche'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(raeumeApi)

const zustaende = kataloge.de.bausteine.zustaende

const karten = [
  { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Eins' }] },
  { id: '2', nummer: 1, werte: [{ feld: 'name', wert: 'Zwei' }] },
]

function aufbau(): KartenlisteAntwort {
  return {
    spalten: [],
    zeilen: karten.map((karte) => ({
      karteID: karte.id,
      nummer: karte.nummer + 1,
      werte: { name: karte.werte[0]?.wert ?? '' },
      marken: [],
    })),
    sortierfeld: '',
    aufsteigend: true,
    ohneVorlage: false,
    fehler: null,
  }
}

function zeige(karteID: string) {
  return (
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID={karteID}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={vi.fn()}
      />
    </Sprachraum>
  )
}

const bildname = 'Vorschau der Karte'

describe('der Vorschauwechsel', () => {
  it('lässt das vorige Bild stehen, nur zum Ansehen, bis das neue da ist', async () => {
    let zweite: (antwort: VorschauAntwort) => void = () => undefined
    const erste = vorschau()
    const neue = vorschau()
    if (neue.vorschau !== null) {
      neue.vorschau.bild = '/vorschau/BBBB.png'
    }

    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
      ...ausKarten(() => Promise.resolve({ karten })),
      Karten: () => Promise.resolve({ karten }),
      Kartenlistenaufbau: () => Promise.resolve(aufbau()),
      // **Ein Bildfeld**, damit das abgelöste Bild eines haben könnte.
      Kartenfelder: () =>
        Promise.resolve({
          felder: [form('kunst', 'image'), form('name', 'text')],
          fehler: null,
        }),
      Vorschau: (karteID: string) =>
        karteID === '1'
          ? Promise.resolve(erste)
          : new Promise<VorschauAntwort>((los) => {
              zweite = los
            }),
    })

    const { rerender } = render(zeige('1'))
    const vorher = await screen.findByRole('img', { name: bildname })
    expect(vorher).toHaveAttribute('src', '/vorschau/AAAA.png')
    expect(vorher.closest('[inert]')).toBeNull()
    await screen.findByRole('button', { name: /^kunst — / })

    rerender(zeige('2'))

    const zwischen = await screen.findByRole('img', {
      name: bildname,
      hidden: true,
    })
    expect(zwischen).toHaveAttribute('src', '/vorschau/AAAA.png')
    const huelle = zwischen.closest('[inert]')
    expect(huelle).toHaveAttribute('aria-busy', 'true')
    // **Kein Bildfeld im abgelösten Bild** — ein Bild, das jetzt darauf
    // fiele, ginge sonst in das gleichnamige Feld der neuen Karte.
    expect(huelle?.querySelector('[data-bildfeld]')).toBeNull()
    // **Und man sieht, dass es lädt** — gedämpft allein hielte man das Bild
    // für das der neuen Karte (keine falsche Illusion).
    expect(huelle).toHaveTextContent(zustaende.laedt)

    zweite(neue)

    // **Gewartet wird auf das neue Bild** — das abgelöste ist bis dahin
    // auch eines mit diesem Namen.
    await waitFor(() => {
      expect(
        screen.getAllByRole('img', { name: bildname, hidden: true }),
      ).toHaveLength(1)
    })
    const nachher = screen.getByRole('img', { name: bildname })
    expect(nachher).toHaveAttribute('src', '/vorschau/BBBB.png')
    expect(nachher.closest('[inert]')).toBeNull()
  })

  // **Und auf dem Rückweg ebenso** — Befund 7 des Zweitblicks über Phase 6
  // von 0.9.0: A steht, B lädt noch, zurück zu A. Das letzte Bild gehört A
  // selbst, und eine Bedingung „nur das einer **anderen** Karte“ ließ die
  // Vorschau leer, bis A neu gerechnet war.
  it('lässt auf dem schnellen Rückweg das letzte Bild stehen', async () => {
    let aufrufeEins = 0
    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
      ...ausKarten(() => Promise.resolve({ karten })),
      Karten: () => Promise.resolve({ karten }),
      Kartenlistenaufbau: () => Promise.resolve(aufbau()),
      Kartenfelder: () =>
        Promise.resolve({ felder: [form('name', 'text')], fehler: null }),
      Vorschau: (karteID: string) => {
        if (karteID === '1') {
          aufrufeEins += 1
          if (aufrufeEins === 1) {
            return Promise.resolve(vorschau())
          }
        }

        return new Promise<VorschauAntwort>(() => undefined)
      },
    })

    const { rerender } = render(zeige('1'))
    await screen.findByRole('img', { name: bildname })
    rerender(zeige('2'))
    await screen.findByRole('img', { name: bildname, hidden: true })
    rerender(zeige('1'))

    await waitFor(() => {
      expect(aufrufeEins).toBe(2)
    })
    const zurueck = screen.getByRole('img', { name: bildname, hidden: true })
    expect(zurueck).toHaveAttribute('src', '/vorschau/AAAA.png')
    expect(zurueck.closest('[inert]')).toHaveAttribute('aria-busy', 'true')
  })
})
