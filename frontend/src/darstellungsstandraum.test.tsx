// **Der Raum, in dem die Teile des Fensters sich melden** — UXR-F14
// (#2029).
//
// Geprüft wird, was nach Go geht und in welcher Reihenfolge — die Fassade ist
// die Naht, und `test/api.ts` antwortet eine Makroaufgabe später, wie die
// Wails-IPC (§3).
import { render, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  anfangsstand,
  fehlt,
  useFensterteilmeldung,
  type Darstellungsstand,
  type Fensterteilname,
} from './darstellungsstand'
import { Darstellungsstandraum } from './darstellungsstandraum'
import { antwortAbwarten, belegeApi, raeumeApi } from './test/api'

afterEach(() => {
  raeumeApi()
})

const steht = { vorhanden: true, offen: true }
const zu = { vorhanden: true, offen: false }

function Teil({
  teil,
  vorhanden,
  offen,
}: {
  teil: Fensterteilname
  vorhanden: boolean
  offen: boolean
}) {
  useFensterteilmeldung(teil, vorhanden, offen)

  return null
}

describe('Darstellungsstandraum', () => {
  // **Immer beim Einhängen**, auch den Anfangsstand: der Raum entsteht neu,
  // wenn die Sitzung aufgeht oder das Profil wechselt, und Go hat dann noch
  // den Stand von vorher.
  it('meldet beim Einhängen den Anfangsstand', async () => {
    const DarstellungsstandMelden = vi.fn(() =>
      Promise.resolve({ fehler: null }),
    )
    belegeApi({ DarstellungsstandMelden })

    render(<Darstellungsstandraum>{null}</Darstellungsstandraum>)

    await waitFor(() => {
      expect(DarstellungsstandMelden).toHaveBeenCalledTimes(1)
    })
    expect(DarstellungsstandMelden).toHaveBeenCalledWith(anfangsstand)
  })

  it('sammelt die Teile und meldet den ganzen Stand', async () => {
    const DarstellungsstandMelden = vi.fn(() =>
      Promise.resolve({ fehler: null }),
    )
    belegeApi({ DarstellungsstandMelden })

    render(
      <Darstellungsstandraum>
        <Teil teil="seitenleiste" vorhanden offen />
        <Teil teil="navigator" vorhanden offen={false} />
      </Darstellungsstandraum>,
    )

    const erwartet: Darstellungsstand = {
      seitenleiste: steht,
      navigator: zu,
      inspector: fehlt,
    }
    await waitFor(() => {
      expect(DarstellungsstandMelden).toHaveBeenLastCalledWith(erwartet)
    })
  })

  // **Abmelden heißt „fehlt"**: der Arbeitsplatz einer Seite, die man
  // verlässt, ist weg, und die Menüeinträge darauf sind es auch.
  it('nimmt ein Teil zurück, wenn sein Bauteil geht', async () => {
    const DarstellungsstandMelden = vi.fn(() =>
      Promise.resolve({ fehler: null }),
    )
    belegeApi({ DarstellungsstandMelden })

    const mit = (
      <Darstellungsstandraum>
        <Teil teil="inspector" vorhanden offen />
      </Darstellungsstandraum>
    )
    const { rerender } = render(mit)

    await waitFor(() => {
      expect(DarstellungsstandMelden).toHaveBeenLastCalledWith({
        ...anfangsstand,
        inspector: steht,
      })
    })

    rerender(<Darstellungsstandraum>{null}</Darstellungsstandraum>)

    await waitFor(() => {
      expect(DarstellungsstandMelden).toHaveBeenLastCalledWith(anfangsstand)
    })
  })

  // **Ein Wechsel zwischen zwei Arbeitsplätzen ändert den Stand nicht**: das
  // eine Bauteil meldet ab, das nächste an, im selben Zug.
  it('meldet einen unveränderten Stand nicht noch einmal', async () => {
    const DarstellungsstandMelden = vi.fn(() =>
      Promise.resolve({ fehler: null }),
    )
    belegeApi({ DarstellungsstandMelden })

    const { rerender } = render(
      <Darstellungsstandraum>
        <Teil key="a" teil="navigator" vorhanden offen />
      </Darstellungsstandraum>,
    )
    await waitFor(() => {
      expect(DarstellungsstandMelden).toHaveBeenCalledTimes(2)
    })

    // Ein anderes Bauteil mit demselben Stand — Abmelden und Anmelden
    // gleichzeitig.
    rerender(
      <Darstellungsstandraum>
        <Teil key="b" teil="navigator" vorhanden offen />
      </Darstellungsstandraum>,
    )
    await antwortAbwarten()

    expect(DarstellungsstandMelden).toHaveBeenCalledTimes(2)

    // **Und wer sich wirklich ändert, wird gemeldet** — sonst wäre die
    // Prüfung darüber auch bei einem Raum grün, der nie etwas sendet.
    rerender(
      <Darstellungsstandraum>
        <Teil key="b" teil="navigator" vorhanden offen={false} />
      </Darstellungsstandraum>,
    )
    await waitFor(() => {
      expect(DarstellungsstandMelden).toHaveBeenCalledTimes(3)
    })
  })

  // **Der Reihe nach** — dieselbe Zusage wie `sprache/menue.ts`: zwei
  // schnelle Wechsel dürfen in Go nicht in umgekehrter Reihenfolge ankommen,
  // sonst steht die Leiste im Stand von vorhin.
  it('schickt den nächsten Stand erst, wenn der vorige angekommen ist', async () => {
    const angefangen: Darstellungsstand[] = []
    let loslassen: (() => void) | undefined

    belegeApi({
      DarstellungsstandMelden: vi.fn((stand: Darstellungsstand) => {
        angefangen.push(stand)

        return angefangen.length === 1
          ? new Promise<{ fehler: null }>((fertig) => {
              loslassen = () => {
                fertig({ fehler: null })
              }
            })
          : Promise.resolve({ fehler: null })
      }),
    })

    const { rerender } = render(
      <Darstellungsstandraum>{null}</Darstellungsstandraum>,
    )
    await waitFor(() => {
      expect(angefangen).toHaveLength(1)
    })

    rerender(
      <Darstellungsstandraum>
        <Teil teil="navigator" vorhanden offen />
      </Darstellungsstandraum>,
    )
    await antwortAbwarten()

    // **Der zweite wartet**, solange der erste in Go unterwegs ist.
    expect(angefangen).toHaveLength(1)

    loslassen?.()

    await waitFor(() => {
      expect(angefangen).toHaveLength(2)
    })
    expect(angefangen[1]).toEqual({ ...anfangsstand, navigator: steht })
  })

  // **Ein gescheiterter Ruf wird nicht gemeldet und hält den nächsten nicht
  // auf** — gesperrt hält Go jede Bindung an, und hinter der Sperre gibt es
  // diesen Raum nicht; was bleibt, ist Bestenfalls-Arbeit. Vitest weist eine
  // unbehandelte Ablehnung selbst ab.
  it('lässt einen gescheiterten Ruf den nächsten nicht aufhalten', async () => {
    const DarstellungsstandMelden = vi
      .fn()
      .mockRejectedValueOnce(new Error('die Brücke ist weg'))
      .mockResolvedValueOnce({
        fehler: {
          code: 'unbekannt',
          schluessel: 'fehler.unbekannt',
          meldung: 'abgewiesen',
          werte: null,
        },
      })
      .mockResolvedValue({ fehler: null })
    belegeApi({ DarstellungsstandMelden })

    const { rerender } = render(
      <Darstellungsstandraum>{null}</Darstellungsstandraum>,
    )
    await waitFor(() => {
      expect(DarstellungsstandMelden).toHaveBeenCalledTimes(1)
    })

    // Der zweite Ruf wird mit einem Fehlerumschlag beantwortet (`auspacken`
    // wirft), der dritte geht durch.
    rerender(
      <Darstellungsstandraum>
        <Teil teil="navigator" vorhanden offen />
      </Darstellungsstandraum>,
    )
    await waitFor(() => {
      expect(DarstellungsstandMelden).toHaveBeenCalledTimes(2)
    })

    rerender(
      <Darstellungsstandraum>
        <Teil teil="navigator" vorhanden offen={false} />
      </Darstellungsstandraum>,
    )
    await waitFor(() => {
      expect(DarstellungsstandMelden).toHaveBeenCalledTimes(3)
    })
  })
})

describe('useFensterteilmeldung', () => {
  // **Ohne Raum geschieht nichts, und das ist kein Fehler**: jeder Test des
  // Arbeitsplatzes steht ohne einen.
  it('meldet ohne Raum nichts und wirft nicht', () => {
    const DarstellungsstandMelden = vi.fn(() =>
      Promise.resolve({ fehler: null }),
    )
    belegeApi({ DarstellungsstandMelden })

    const { unmount } = render(<Teil teil="navigator" vorhanden offen />)
    unmount()

    expect(DarstellungsstandMelden).not.toHaveBeenCalled()
  })
})
