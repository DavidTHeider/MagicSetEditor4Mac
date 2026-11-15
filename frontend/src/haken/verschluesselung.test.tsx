// **Der Haken zur Verschlüsselung des Bestands** — Werkstück 14 von Phase 4.12
// (#1351).
//
// # Wonach hier gefragt wird
//
// Nach dem **Stand**, nicht nach der Chiffre: dass er gelesen wird, dass er
// aus der **Antwort** kommt und nicht aus dem, was gewünscht war, und dass ein
// Fehler stehen bleibt (§7, Regel 3).
import { renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { Sprachraum } from '../sprache/kontext'
import { useVerschluesselung } from './useVerschluesselung'
import { useVorgang } from './useVorgang'

afterEach(() => {
  raeumeApi()
})

function rahmen({ children }: { children: React.ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function zeige() {
  return renderHook(() => useVerschluesselung(useVorgang()), {
    wrapper: rahmen,
  })
}

describe('useVerschluesselung', () => {
  it('liest den Stand beim Aufbau', async () => {
    belegeApi({
      Verschluesselungsstand: vi
        .fn()
        .mockResolvedValue({ verschluesselt: true, fehler: null }),
    })

    const { result } = zeige()

    await waitFor(() => {
      expect(result.current.verschluesselt).toBe(true)
    })
  })

  // **Der Stand kommt aus der Antwort und nicht aus dem Wunsch.**
  //
  // Hier liegt der Fall, den ein optimistisches Setzen verschluckte: die
  // Bindung liest nach dem Tausch von der Platte. Sagt sie etwas anderes als
  // das Verlangte, gilt sie — sonst zeigte die Seite einen Stand, den es
  // nicht gibt.
  it('übernimmt den Stand aus der Antwort, nicht den verlangten', async () => {
    const umschalten = vi
      .fn()
      .mockResolvedValue({ verschluesselt: false, fehler: null })

    belegeApi({
      Verschluesselungsstand: vi
        .fn()
        .mockResolvedValue({ verschluesselt: false, fehler: null }),
      VerschluesselungUmschalten: umschalten,
    })

    const { result } = zeige()

    await waitFor(() => {
      expect(result.current.verschluesselt).toBe(false)
    })

    result.current.umschalten(true)

    await antwortAbwarten()

    expect(umschalten).toHaveBeenCalledWith(true)
    expect(result.current.verschluesselt).toBe(false)
  })

  // **Die Meldung wird genannt und nicht nur ihr Dasein geprüft.**
  //
  // Beim Schreiben dieses Laufs stand hier `toBeDefined()`, und er war grün,
  // **bevor** die Methode in der Fassadenliste stand — eine unbelegte Methode
  // lehnt ebenso ab („Die Methode … ist nicht belegt"). Ein Lauf, der zwei
  // ganz verschiedene Ursachen nicht unterscheidet, prüft keine davon (§10).
  it('meldet einen Fehler beim Lesen — und zwar den vom Bestand', async () => {
    belegeApi({
      Verschluesselungsstand: vi.fn().mockResolvedValue({
        verschluesselt: false,
        fehler: {
          code: 'schluesselFehlt',
          schluessel: 'schluesselFehlt',
          meldung: 'der Schlüssel fehlt',
        },
      }),
    })

    const { result } = zeige()

    await waitFor(() => {
      expect(result.current.fehler).toMatch(/verschlüsselt/)
    })

    expect(result.current.fehler).not.toMatch(/nicht belegt/)
  })

  it('meldet einen Fehler beim Schalten und lässt den Stand stehen', async () => {
    belegeApi({
      Verschluesselungsstand: vi
        .fn()
        .mockResolvedValue({ verschluesselt: false, fehler: null }),
      VerschluesselungUmschalten: vi.fn().mockResolvedValue({
        verschluesselt: false,
        fehler: {
          code: 'belegt',
          schluessel: 'belegt',
          meldung: 'jemand anders',
        },
      }),
    })

    const { result } = zeige()

    await waitFor(() => {
      expect(result.current.verschluesselt).toBe(false)
    })

    result.current.umschalten(true)

    await waitFor(() => {
      expect(result.current.fehler).toBeDefined()
    })

    expect(result.current.verschluesselt).toBe(false)
  })

  // **Eine Antwort, die nach dem Abbau kommt, trägt nichts mehr ein.**
  //
  // Ohne den Wächter meldete React „state update on an unmounted component",
  // und unter Last wäre es ein Eintrag in einen Baum, den es nicht gibt.
  //
  // **Beide Wege einzeln**, denn es sind zwei Wächter: einer im Erfolgsweg,
  // einer im Fehlerweg. Der zweite blieb beim ersten Anlauf unbetreten, und
  // die Abdeckung hat ihn angestrichen — ein Zweig, den kein Lauf betritt,
  // ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
  it.each([
    ['eine Antwort', { verschluesselt: true, fehler: null }, undefined],
    [
      'einen Fehler',
      undefined,
      new Error('der Bestand war nicht zu erreichen'),
    ],
  ])(
    'trägt nach dem Abbau %s nicht mehr ein',
    async (_name, antwort, grund) => {
      belegeApi({
        Verschluesselungsstand: vi
          .fn()
          .mockImplementation(() =>
            grund === undefined
              ? Promise.resolve(antwort)
              : Promise.reject(grund),
          ),
      })

      const { result, unmount } = zeige()

      unmount()

      await antwortAbwarten()

      // **Und der Beleg, dass der Wächter wirklich griff**: nichts steht drin.
      // Ohne diese Zeile wäre der Lauf auch dann grün, wenn er eingetragen
      // hätte — `unmount` allein sagt nichts über den Zustand.
      expect(result.current.verschluesselt).toBeUndefined()
      expect(result.current.fehler).toBeUndefined()
    },
  )
})
