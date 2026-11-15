import { renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  ereignisEinstellungen,
  ereignisSetSichern,
  ereignisSetSichernUnter,
  ereignisBestandSichern,
  ereignisSetNeu,
  ereignisSetZuletzt,
  ereignisSetOeffnen,
  ereignisSetSchliessen,
  ereignisDrucken,
  ereignisHilfe,
  ereignisSeitenleiste,
  ereignisNavigator,
  ereignisInspector,
  ereignisSeiteZurueck,
  ereignisSeiteVor,
  ereignisDateiGeoeffnet,
  ereignisDateiFehler,
  useFensterereignisse,
} from './fenster'

interface MitLaufzeit {
  runtime?: { EventsOn: (name: string, rueckruf: () => void) => () => void }
}

afterEach(() => {
  delete (globalThis as MitLaufzeit).runtime
})

describe('die Ereignisse für Navigator und Inspector', () => {
  // **Dieselben Namen wie in Go** (`app.EreignisNavigator`,
  // `app.EreignisInspector`) — UXR-F14 (#2029). Gehört werden sie im
  // `Arbeitsplatz`, nicht in `useFensterereignisse`; der Name allein ist die
  // Naht, und ein Tippfehler ergäbe einen Menüeintrag, der nichts tut.
  it('heißen so, wie Go sie meldet', () => {
    expect(ereignisNavigator).toBe('menue:navigator')
    expect(ereignisInspector).toBe('menue:inspector')
  })
})

describe('useFensterereignisse', () => {
  // **Ohne Laufzeit tut es nichts, und das ist kein Fehler**: in der
  // Browservorschau gibt es keine native Menüleiste.
  it('kommt ohne Wails-Laufzeit aus', () => {
    expect(() => {
      renderHook(() => {
        useFensterereignisse({
          einstellungen: vi.fn(),
          sichern: vi.fn(),
          sichernUnter: vi.fn(),
          bestandSichern: vi.fn(),
          setNeu: vi.fn(),
          setZuletzt: vi.fn(),
          setOeffnen: vi.fn(),
          setSchliessen: vi.fn(),
          drucken: vi.fn(),
          hilfe: vi.fn(),
          rueckgaengig: vi.fn(),
          wiederherstellen: vi.fn(),
          seite: vi.fn(),
          ueber: vi.fn(),
          seitenleiste: vi.fn(),
          seiteZurueck: vi.fn(),
          seiteVor: vi.fn(),
          dateiGeoeffnet: vi.fn(),
          dateiFehler: vi.fn(),
        })
      })
    }).not.toThrow()
  })

  it('meldet sich für jeden Eintrag an und wieder ab', () => {
    const horcher = new Map<string, (...daten: unknown[]) => void>()
    const abgemeldet: string[] = []

    ;(globalThis as MitLaufzeit).runtime = {
      EventsOn: (name, rueckruf) => {
        horcher.set(name, rueckruf)

        return () => {
          abgemeldet.push(name)
        }
      },
    }

    const handlungen = {
      einstellungen: vi.fn(),
      sichern: vi.fn(),
      sichernUnter: vi.fn(),
      bestandSichern: vi.fn(),
      setNeu: vi.fn(),
      setZuletzt: vi.fn(),
      setOeffnen: vi.fn(),
      setSchliessen: vi.fn(),
      drucken: vi.fn(),
      hilfe: vi.fn(),
      rueckgaengig: vi.fn(),
      wiederherstellen: vi.fn(),
      seite: vi.fn(),
      ueber: vi.fn(),
      seitenleiste: vi.fn(),
      seiteZurueck: vi.fn(),
      seiteVor: vi.fn(),
      dateiGeoeffnet: vi.fn(),
      dateiFehler: vi.fn(),
    }

    const { unmount } = renderHook(() => {
      useFensterereignisse(handlungen)
    })

    horcher.get(ereignisEinstellungen)?.()
    horcher.get(ereignisSetSichern)?.()
    horcher.get(ereignisSetSichernUnter)?.()
    horcher.get(ereignisBestandSichern)?.()
    horcher.get(ereignisSetNeu)?.()
    horcher.get(ereignisSetZuletzt)?.('7')
    horcher.get(ereignisSetOeffnen)?.()
    horcher.get(ereignisSetSchliessen)?.()
    horcher.get(ereignisDrucken)?.()
    horcher.get(ereignisHilfe)?.()
    horcher.get(ereignisSeitenleiste)?.()

    // **Zurück und Vorwärts** unter *Darstellung* — UX-03.AK-02 (#1675).
    horcher.get(ereignisSeiteZurueck)?.()
    horcher.get(ereignisSeiteVor)?.()

    // **Die beiden mit Nutzlast** (#891): sie prüfen ihre Werte, statt sie zu
    // glauben — ein blindes `as` machte aus einer geänderten Go-Seite eine
    // Seite, die stumm nichts tut.
    horcher.get(ereignisDateiGeoeffnet)?.('7')
    // **Mit Epoche, und mit einer, die keine Zahl ist** (J05, #2269): die
    // Zahl geht weiter, alles andere zählt als „ohne".
    horcher.get(ereignisDateiGeoeffnet)?.('8', 3)
    horcher.get(ereignisDateiGeoeffnet)?.('9', 'drei')
    horcher.get(ereignisDateiFehler)?.('/pfad/Probe.mse-set', {
      schluessel: 'nichtGefunden',
      meldung: 'Es ist ein unerwarteter Fehler aufgetreten.',
      werte: null,
    })

    expect(handlungen.einstellungen).toHaveBeenCalledOnce()
    expect(handlungen.sichern).toHaveBeenCalledOnce()
    expect(handlungen.sichernUnter).toHaveBeenCalledOnce()
    expect(handlungen.bestandSichern).toHaveBeenCalledOnce()
    expect(handlungen.setNeu).toHaveBeenCalledOnce()
    expect(handlungen.setZuletzt).toHaveBeenCalledWith('7')
    expect(handlungen.setOeffnen).toHaveBeenCalledOnce()
    expect(handlungen.setSchliessen).toHaveBeenCalledOnce()
    expect(handlungen.drucken).toHaveBeenCalledOnce()
    expect(handlungen.hilfe).toHaveBeenCalledOnce()
    expect(handlungen.seitenleiste).toHaveBeenCalledOnce()
    expect(handlungen.seiteZurueck).toHaveBeenCalledOnce()
    expect(handlungen.seiteVor).toHaveBeenCalledOnce()
    expect(handlungen.dateiGeoeffnet).toHaveBeenCalledWith('7', undefined)
    expect(handlungen.dateiGeoeffnet).toHaveBeenCalledWith('8', 3)
    expect(handlungen.dateiGeoeffnet).toHaveBeenCalledWith('9', undefined)
    expect(handlungen.dateiFehler).toHaveBeenCalledWith(
      '/pfad/Probe.mse-set',
      expect.objectContaining({ schluessel: 'nichtGefunden' }),
    )

    // **Abmelden gehört dazu.** Ein Horcher, der einen Neuaufbau überlebt,
    // ruft danach in eine Komponente, die es nicht mehr gibt — und der Fehler
    // zeigt sich irgendwo anders.
    unmount()

    // **Die Zahl steht für die Aufzählung darüber.** Sie ist gezählt, nicht
    // gesetzt: kommt ein Eintrag dazu und hier nicht, wird der Lauf rot.
    expect(abgemeldet).toHaveLength(19)
  })
})
