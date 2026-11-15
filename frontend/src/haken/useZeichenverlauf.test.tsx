// **Rücknahme je Handgriff im Symboleditor** —
// #1340,
// Werkstück F von Phase 4.11.
//
// # Was MSE2 hier tut
//
// `src/util/action_stack.cpp` Zeilen 21–51 (Fassung `f9d9356` vom
// 2021-07-06): jede Handlung wird auf einen Stapel gelegt, `undo()` nimmt die
// oberste herunter und legt sie auf den Wiederherstellungsstapel; eine neue
// Handlung leert diesen.
//
// **Verschmolzen wird selten und begründet.** `addAction` versucht es nur,
// wenn `last_was_add` gilt — nach einer Rücknahme also nicht —, und die
// Entscheidung trifft die Handlung selbst über `merge()`. Von allen
// Symbolhandlungen setzt genau **eine** es um: `SymbolPartNameAction`
// (`src/data/action/symbol.cpp` Zeilen 324–333), und auch die nur für
// **dasselbe** Teil.
//
// **Ein Zug ist trotzdem ein Schritt** — auf einem anderen Weg:
// `ControlPointMoveAction` erbt von `ExtendableAction`
// (`src/data/action/symbol_part.hpp` Zeilen 45–70) und wird während des Zuges
// über `move(delta)` **fortgeschrieben**, statt neu angelegt zu werden. Beim
// Aufsetzen entsteht sie, beim Loslassen endet ihr Leben.
//
// **Ein Tastendruck dagegen ist je ein Schritt.** `point_editor.cpp` Zeilen
// 355–380 ruft erst `resetActions()` und legt dann eine **neue** Handlung an;
// `ControlPointMoveAction` kennt kein `merge`.
import { act, renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import type { SymbolteilDTO } from '../antworten'
import { useZeichenverlauf } from './useZeichenverlauf'

// teil baut ein Teil mit einem erkennbaren Namen.
function teil(name: string): SymbolteilDTO {
  return {
    art: 'form',
    name,
    kombination: 'overlap',
    punkte: [],
    teile: [],
    symmetrie: null,
  } as unknown as SymbolteilDTO
}

const anfang = [teil('a')]

describe('useZeichenverlauf', () => {
  it('gibt am Anfang den geladenen Stand und nichts zurückzunehmen', () => {
    const { result } = renderHook(() => useZeichenverlauf(anfang))

    expect(result.current.teile).toEqual(anfang)
    expect(result.current.kannZuruecknehmen).toBe(false)
    expect(result.current.kannWiederherstellen).toBe(false)
  })

  it('nimmt genau den letzten Handgriff zurück, nicht die ganze Zeichnung', () => {
    const { result } = renderHook(() => useZeichenverlauf(anfang))

    act(() => {
      result.current.eintragen([teil('a'), teil('b')], { marke: 'anlegen:1' })
    })
    act(() => {
      result.current.eintragen([teil('a'), teil('b'), teil('c')], {
        marke: 'anlegen:2',
      })
    })

    act(() => {
      result.current.zuruecknehmen()
    })

    expect(result.current.teile.map((t) => t.name)).toEqual(['a', 'b'])
    expect(result.current.kannZuruecknehmen).toBe(true)
  })

  it('stellt wieder her, was zurückgenommen wurde', () => {
    const { result } = renderHook(() => useZeichenverlauf(anfang))

    act(() => {
      result.current.eintragen([teil('a'), teil('b')], { marke: 'anlegen:1' })
    })
    act(() => {
      result.current.zuruecknehmen()
    })

    expect(result.current.kannWiederherstellen).toBe(true)

    act(() => {
      result.current.wiederherstellen()
    })

    expect(result.current.teile.map((t) => t.name)).toEqual(['a', 'b'])
  })

  // **`redo_actions.clear()`**, `action_stack.cpp` Zeile 27.
  it('vergisst die Wiederherstellung, sobald etwas Neues geschieht', () => {
    const { result } = renderHook(() => useZeichenverlauf(anfang))

    act(() => {
      result.current.eintragen([teil('a'), teil('b')], { marke: 'anlegen:1' })
    })
    act(() => {
      result.current.zuruecknehmen()
    })
    act(() => {
      result.current.eintragen([teil('a'), teil('z')], { marke: 'anlegen:2' })
    })

    expect(result.current.kannWiederherstellen).toBe(false)
  })

  // **Ein Zug ist ein Schritt** — `ExtendableAction`.
  it('fasst einen Zug zu einem Schritt zusammen', () => {
    const { result } = renderHook(() => useZeichenverlauf(anfang))

    act(() => {
      result.current.eintragen([teil('b')], {
        fortschreibbar: true,
        marke: 'zug:1',
      })
    })
    act(() => {
      result.current.eintragen([teil('c')], {
        fortschreibbar: true,
        marke: 'zug:1',
      })
    })
    act(() => {
      result.current.eintragen([teil('d')], {
        fortschreibbar: true,
        marke: 'zug:1',
      })
    })

    act(() => {
      result.current.zuruecknehmen()
    })

    expect(result.current.teile.map((t) => t.name)).toEqual(['a'])
    expect(result.current.kannZuruecknehmen).toBe(false)
  })

  // **Und zwei Züge sind zwei Schritte** — die Handlung endet beim Loslassen.
  it('trennt zwei Züge, weil die Marke sich ändert', () => {
    const { result } = renderHook(() => useZeichenverlauf(anfang))

    act(() => {
      result.current.eintragen([teil('b')], {
        fortschreibbar: true,
        marke: 'zug:1',
      })
    })
    act(() => {
      result.current.eintragen([teil('c')], {
        fortschreibbar: true,
        marke: 'zug:2',
      })
    })

    act(() => {
      result.current.zuruecknehmen()
    })

    expect(result.current.teile.map((t) => t.name)).toEqual(['b'])
  })

  // **Ein Tastendruck verschmilzt nicht** — `point_editor.cpp` Zeilen 355–380:
  // `resetActions()`, dann eine neue Handlung ohne `merge`.
  it('legt für jeden Tastendruck einen eigenen Schritt an', () => {
    const { result } = renderHook(() => useZeichenverlauf(anfang))

    act(() => {
      result.current.eintragen([teil('b')], { marke: 'taste' })
    })
    act(() => {
      result.current.eintragen([teil('c')], { marke: 'taste' })
    })

    act(() => {
      result.current.zuruecknehmen()
    })

    expect(result.current.teile.map((t) => t.name)).toEqual(['b'])
  })

  // **Nach einer Rücknahme wird nicht verschmolzen** — `last_was_add`,
  // `action_stack.cpp` Zeile 30.
  //
  // **Es braucht zwei Schritte, sonst prüft der Fall nichts.** Der erste
  // Anlauf nahm den einzigen Schritt zurück und trug dann wieder ein: der
  // Stapel war leer, und dass es nichts zu verschmelzen gab, entschied schon
  // die Prüfung auf den obersten Schritt. Die Gegenprobe blieb grün, und der
  // Wächter darunter war nie betreten (§10, [[gegenprobe-am-schaden]]).
  //
  // **Und dieselbe Marke zweimal ist kein erfundener Fall**: wer sie aus der
  // Identität des gezogenen Punktes bildet, bekommt beim zweiten Zug an
  // demselben Punkt dieselbe.
  it('verschmilzt nicht mit einem Schritt, der gerade zurückgenommen war', () => {
    const { result } = renderHook(() => useZeichenverlauf(anfang))

    act(() => {
      result.current.eintragen([teil('b')], {
        fortschreibbar: true,
        marke: 'zug:punkt1',
      })
    })
    act(() => {
      result.current.eintragen([teil('c')], {
        fortschreibbar: true,
        marke: 'zug:punkt2',
      })
    })
    act(() => {
      result.current.zuruecknehmen()
    })

    // Jetzt liegt `zug:punkt1` obenauf, und der nächste Handgriff trägt
    // dieselbe Marke.
    act(() => {
      result.current.eintragen([teil('d')], {
        fortschreibbar: true,
        marke: 'zug:punkt1',
      })
    })

    act(() => {
      result.current.zuruecknehmen()
    })

    // Ohne die Regel hätte der neue Eintrag den obersten Schritt
    // fortgeschrieben, und die Rücknahme spränge auf `a` statt auf `b`.
    expect(result.current.teile.map((t) => t.name)).toEqual(['b'])
    expect(result.current.kannZuruecknehmen).toBe(true)
  })

  // **Der Name des Schritts** — MSEs `getName(bool to_undo)`.
  it('nennt den Handgriff, den die Rücknahme trifft', () => {
    const { result } = renderHook(() => useZeichenverlauf(anfang))

    act(() => {
      result.current.eintragen([teil('b')], {
        marke: 'zug:1',
        name: 'punkt verschieben',
      })
    })

    expect(result.current.naechsteRuecknahme).toBe('punkt verschieben')

    act(() => {
      result.current.zuruecknehmen()
    })

    expect(result.current.naechsteWiederherstellung).toBe('punkt verschieben')
    expect(result.current.naechsteRuecknahme).toBeUndefined()
  })

  // **Eine Rücknahme ohne Schritt tut nichts** — `assert(canUndo())`,
  // `action_stack.cpp` Zeile 42.
  it('tut nichts, wenn es nichts zurückzunehmen gibt', () => {
    const { result } = renderHook(() => useZeichenverlauf(anfang))

    act(() => {
      result.current.zuruecknehmen()
      result.current.wiederherstellen()
    })

    expect(result.current.teile).toEqual(anfang)
  })

  // **Der geladene Stand zählt, solange niemand etwas geändert hat.**
  //
  // Dieselbe Zusage wie bisher in `Symbol.tsx`: `teile` ist `undefined`, und
  // es gilt, was geladen wurde. Wer das letzte Teil löscht, bekommt es nicht
  // zurück, weil er eine **Änderung** vorgenommen hat.
  it('folgt dem geladenen Stand, bis jemand etwas ändert', () => {
    const { rerender, result } = renderHook(
      ({ geladen }) => useZeichenverlauf(geladen),
      { initialProps: { geladen: anfang } },
    )

    rerender({ geladen: [teil('x')] })
    expect(result.current.teile.map((t) => t.name)).toEqual(['x'])

    act(() => {
      result.current.eintragen([], { marke: 'loeschen' })
    })

    rerender({ geladen: [teil('x')] })
    expect(result.current.teile).toEqual([])
  })
})
