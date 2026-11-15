// **Zwei Symbolhandlungen zugleich** — Befund 2 des Zweitblicks über Phase
// 4.11.
//
// # Der Befund
//
// `Grundform`, `Nachzeichnen` und die Wege der Teilliste, der Punkte und der
// Symmetrie schicken jeweils den **vollständigen alten Zeichenstand** an eine
// asynchrone Bindung und setzen die Antwort danach ganz ein. Ohne Wächter
// laufen zwei Handlungen parallel; antwortet die ältere zuletzt, ersetzt sie
// den neueren Stand — und legt dabei noch einen irreführenden
// Rücknahmeschritt an.
//
// **Das verletzt §7 Regel 2 und Regel 4.**
import { act, renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { useZeichenschutz } from './useZeichenschutz'

describe('useZeichenschutz', () => {
  // **Regel 4: ein Vorgang je Datensatz.**
  //
  // Die Zeichnung ist **ein** Datensatz. Zwei Umformungen zugleich gehen
  // beide gegen den Stand von vorher — die zweite verlöre, was die erste
  // wollte, gleich in welcher Reihenfolge die Antworten kommen.
  it('lässt nur eine Umformung zugleich zu', async () => {
    const { result } = renderHook(() => useZeichenschutz())

    let ersteFertig: () => void = () => undefined
    const erste = vi.fn(
      () =>
        new Promise<void>((aufloesen) => {
          ersteFertig = aufloesen
        }),
    )
    const zweite = vi.fn(() => Promise.resolve())

    act(() => {
      void result.current.starte(erste)
      void result.current.starte(zweite)
    })

    expect(erste).toHaveBeenCalledTimes(1)
    expect(zweite).not.toHaveBeenCalled()

    // **Danach geht wieder etwas** — der Wächter hält an, er sperrt nicht.
    await act(async () => {
      ersteFertig()
      await Promise.resolve()
    })

    act(() => {
      void result.current.starte(zweite)
    })

    expect(zweite).toHaveBeenCalledTimes(1)
  })

  // **Regel 2: eine Antwort gehört zu dem Stand, aus dem sie stammt.**
  //
  // Der Zähler wird beim Start genommen; wer danach abräumt, verwirft die
  // Antwort.
  it('verwirft eine Antwort, deren Stand vorbei ist', async () => {
    const { result } = renderHook(() => useZeichenschutz())

    const eingetragen = vi.fn()

    await act(async () => {
      await result.current.starte((nochAktuell) => {
        result.current.raeume()

        if (nochAktuell()) {
          eingetragen()
        }

        return Promise.resolve()
      })
    })

    expect(eingetragen).not.toHaveBeenCalled()
  })

  // **Und ohne Abräumen trägt sie sich ein** — der Beleg, dass der Fall
  // darüber am Wächter scheitert und nicht daran, dass nie etwas eingetragen
  // wird.
  it('trägt eine Antwort ein, deren Stand noch gilt', async () => {
    const { result } = renderHook(() => useZeichenschutz())

    const eingetragen = vi.fn()

    await act(async () => {
      await result.current.starte((nochAktuell) => {
        if (nochAktuell()) {
          eingetragen()
        }

        return Promise.resolve()
      })
    })

    expect(eingetragen).toHaveBeenCalledTimes(1)
  })

  // **Ein Fehler gibt den Vorgang frei — und wird weitergereicht.**
  //
  // Frei, sonst bliebe der Editor nach einem einzigen gescheiterten Ruf für
  // immer stumm (§7, Regel 3). Weitergereicht, weil nur der Aufrufer weiß, was
  // eine gescheiterte Umformung für ihn heißt: die sechs Stellen fangen ihn
  // selbst und melden ihn am Editor.
  it('gibt nach einem Fehler wieder frei und reicht ihn weiter', async () => {
    const { result } = renderHook(() => useZeichenschutz())

    const danach = vi.fn(() => Promise.resolve())

    let gefangen: unknown

    await act(async () => {
      await result.current
        .starte(() => Promise.reject(new Error('kaputt')))
        .catch((ursache: unknown) => {
          gefangen = ursache
        })
    })

    expect(gefangen).toBeInstanceOf(Error)

    act(() => {
      void result.current.starte(danach)
    })

    expect(danach).toHaveBeenCalledTimes(1)
  })

  // **Ein Abräumen gibt den Vorgang frei** — Nachlauf über Phase 4.11.
  //
  // # Der Befund
  //
  // `raeume()` zählte nur den Stand hoch; `unterwegs` blieb bis zum Ende des
  // **alten** Rufs belegt. Hängt die Naht, war damit auch das neu
  // aufgeschlagene Symbol für jede Handlung gesperrt.
  it('gibt beim Abräumen frei, ohne auf den alten Ruf zu warten', () => {
    const { result } = renderHook(() => useZeichenschutz())

    const haengt = vi.fn(() => new Promise<void>(() => undefined))
    const danach = vi.fn(() => Promise.resolve())

    act(() => {
      void result.current.starte(haengt)
    })

    expect(haengt).toHaveBeenCalledTimes(1)

    // **Das Bezugsobjekt wechselt, während der Ruf hängt.**
    act(() => {
      result.current.raeume()
      void result.current.starte(danach)
    })

    expect(danach).toHaveBeenCalledTimes(1)
  })

  // **Und der alte Ruf gibt danach nichts mehr frei.**
  //
  // Sonst liefen zwei zugleich: der abgeräumte kommt zurück, setzt
  // `unterwegs` auf falsch — und der neue, der noch läuft, wäre nicht mehr
  // geschützt.
  it('lässt einen abgeräumten Ruf den neuen nicht freigeben', async () => {
    const { result } = renderHook(() => useZeichenschutz())

    let alterFertig: () => void = () => undefined

    const alter = vi.fn(
      () =>
        new Promise<void>((aufloesen) => {
          alterFertig = aufloesen
        }),
    )
    const neuer = vi.fn(() => new Promise<void>(() => undefined))
    const dritter = vi.fn(() => Promise.resolve())

    act(() => {
      void result.current.starte(alter)
    })

    act(() => {
      result.current.raeume()
      void result.current.starte(neuer)
    })

    // **Der alte kommt zurück** — er darf den neuen nicht freigeben.
    await act(async () => {
      alterFertig()
      await Promise.resolve()
    })

    act(() => {
      void result.current.starte(dritter)
    })

    expect(dritter).not.toHaveBeenCalled()
  })
})
