// Der Haken hinter ⌘Z und ⇧⌘Z.
//
// Er hält alle vier Regeln aus AGENTS.md §7 — und jede bekommt hier ihren Fall.
import { act, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { type BewegungAntwort } from '../api'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { useVerlauf } from './useVerlauf'

afterEach(() => {
  raeumeApi()
})

// bewegt baut eine geglückte Antwort.
function bewegt(karteID: string): BewegungAntwort {
  return {
    bewegt: true,
    karteID,
    feld: 'name',
    umfang: 1,
    urheber: 'fenster',
    fehler: null,
  }
}

// kaputt ist ein Fehler, wie ihn `internal/app` schickt.
function kaputt(meldung: string): BewegungAntwort {
  return {
    ...nichts,
    fehler: { code: 'unbekannt', schluessel: 'fehler.unbekannt', meldung },
  }
}

// nichts ist die Antwort auf einen leeren Verlauf.
const nichts: BewegungAntwort = {
  bewegt: false,
  karteID: '0',
  feld: '',
  umfang: 0,
  urheber: '',
  fehler: null,
}

// Probe zeigt den Stand und bietet beide Richtungen als Schaltflächen.
function Probe({ setID }: { setID: string | undefined }) {
  const verlauf = useVerlauf(setID)

  return (
    <>
      <span data-testid="stand">{verlauf.stand}</span>
      <span data-testid="fehler">{verlauf.fehler ?? ''}</span>
      <button
        type="button"
        onClick={() => {
          verlauf.zuruecknehmen()
        }}
      >
        zurück
      </button>
      <button
        type="button"
        onClick={() => {
          verlauf.wiederherstellen()
        }}
      >
        vor
      </button>
    </>
  )
}

// abwarten lässt die angefangenen Zusagen zu Ende laufen.
//
// **Eine Makroaufgabe und keine Mikroaufgabe** (#1321): `test/api.ts`
// antwortet seit dem 2026-09-02 eine Makroaufgabe später, so wie die IPC im
// Betrieb. `Promise.resolve()` leert nur die Mikroaufgaben davor — die
// Antwort steht dann noch in der Schlange, und der Stand darunter ist der von
// vorher.
const abwarten = antwortAbwarten

// druecke klickt eine der beiden Schaltflächen und wartet die Zusage ab.
async function druecke(beschriftung: string) {
  await act(async () => {
    screen.getByRole('button', { name: beschriftung }).click()
    await abwarten()
  })
}

const stand = () => screen.getByTestId('stand').textContent
const fehler = () => screen.getByTestId('fehler').textContent

describe('useVerlauf', () => {
  it('hebt den Stand nach einer geglückten Rücknahme', async () => {
    belegeApi({
      Zuruecknehmen: () => Promise.resolve(bewegt('7')),
      Wiederherstellen: () => Promise.resolve(bewegt('7')),
    })

    render(<Probe setID="1" />)
    expect(stand()).toBe('0')

    await druecke('zurück')
    expect(stand()).toBe('1')

    await druecke('vor')
    expect(stand()).toBe('2')
  })

  // **Ein leerer Verlauf verschiebt nichts.** ⌘Z am Anfang ist eine gewöhnliche
  // Lage; wer den Stand trotzdem höbe, ließe die ganze Kartenliste neu laden,
  // ohne dass sich etwas geändert hat.
  it('hebt den Stand nicht, wenn nichts zurückzunehmen war', async () => {
    belegeApi({ Zuruecknehmen: () => Promise.resolve(nichts) })

    render(<Probe setID="1" />)
    await druecke('zurück')

    expect(stand()).toBe('0')
    expect(fehler()).toBe('')
  })

  // **Ohne Set gibt es nichts zurückzunehmen** — und der Bestand wird nicht
  // gefragt.
  it('fragt ohne Set gar nicht', async () => {
    const gerufen = vi.fn(() => Promise.resolve(nichts))
    belegeApi({ Zuruecknehmen: gerufen })

    render(<Probe setID={undefined} />)
    await druecke('zurück')

    expect(gerufen).not.toHaveBeenCalled()
    expect(stand()).toBe('0')
  })

  it('reicht die Setkennung durch', async () => {
    const gerufen = vi.fn(() => Promise.resolve(bewegt('7')))
    belegeApi({ Zuruecknehmen: gerufen })

    render(<Probe setID="42" />)
    await druecke('zurück')

    expect(gerufen).toHaveBeenCalledWith('42', '')
  })

  // **Ein Fehler erscheint und verschiebt den Stand nicht** (§7, Regel 3):
  // geschrieben ist geschrieben, und was nicht geschrieben wurde, lädt auch
  // nichts nach.
  it('meldet einen Fehler, ohne den Stand zu heben', async () => {
    belegeApi({
      Zuruecknehmen: () => Promise.resolve(kaputt('die Platte ist voll')),
    })

    render(<Probe setID="1" />)
    await druecke('zurück')

    expect(stand()).toBe('0')
    expect(fehler()).toBe('die Platte ist voll')
  })

  // **Eine geglückte Bewegung räumt die vorige Meldung weg.**
  it('räumt die Meldung nach einer geglückten Bewegung', async () => {
    let scheitern = true
    belegeApi({
      Zuruecknehmen: () =>
        Promise.resolve(scheitern ? kaputt('kaputt') : bewegt('7')),
    })

    render(<Probe setID="1" />)
    await druecke('zurück')
    expect(fehler()).toBe('kaputt')

    scheitern = false
    await druecke('zurück')
    expect(fehler()).toBe('')
  })

  // **Ein Wechsel des Sets räumt den offenen Kontext** (§7, Regel 1). Eine
  // Meldung über das vorige Set stünde sonst über dem neuen.
  it('räumt die Meldung beim Setwechsel', async () => {
    belegeApi({
      Zuruecknehmen: () => Promise.resolve(kaputt('kaputt')),
    })

    const { rerender } = render(<Probe setID="1" />)
    await druecke('zurück')
    expect(fehler()).toBe('kaputt')

    rerender(<Probe setID="2" />)
    expect(fehler()).toBe('')
  })

  // **Ein Vorgang je Set** (§7, Regel 4). Ein zweiter Aufruf, während der erste
  // läuft, ginge gegen einen Stand, den das Backend noch nicht kennt — und der
  // Verlauf wanderte zwei Schritte, wo jemand einen gemeint hat.
  it('lässt keinen zweiten Aufruf zu, solange der erste läuft', async () => {
    let loesen: ((antwort: BewegungAntwort) => void) | undefined
    const gerufen = vi.fn(
      () =>
        new Promise<BewegungAntwort>((erfuellen) => {
          loesen = erfuellen
        }),
    )

    belegeApi({ Zuruecknehmen: gerufen })
    render(<Probe setID="1" />)

    await druecke('zurück')
    await druecke('zurück')

    expect(gerufen).toHaveBeenCalledTimes(1)

    await act(async () => {
      loesen?.(bewegt('7'))
      await abwarten()
    })

    expect(stand()).toBe('1')
  })

  // **Eine Antwort aus einem vorigen Set wird verworfen** (§7, Regel 2).
  //
  // Wer während einer laufenden Rücknahme das Set wechselt, bekäme sonst deren
  // Meldung über dem neuen Set zu sehen — und einen Stand, der eine Liste
  // nachlädt, die mit der Antwort nichts zu tun hat.
  it('verwirft die Antwort, wenn das Set inzwischen gewechselt hat', async () => {
    let loesen: ((antwort: BewegungAntwort) => void) | undefined
    belegeApi({
      Zuruecknehmen: () =>
        new Promise<BewegungAntwort>((erfuellen) => {
          loesen = erfuellen
        }),
    })

    const { rerender } = render(<Probe setID="1" />)
    await druecke('zurück')

    rerender(<Probe setID="2" />)

    await act(async () => {
      loesen?.(bewegt('7'))
      await abwarten()
    })

    expect(stand()).toBe('0')
    expect(fehler()).toBe('')
  })

  // **Und ein Fehler ebenso.** Er nimmt einen anderen Zweig als der Erfolg, und
  // ein Wächter, der nur die eine Hälfte hält, macht die andere unsichtbar (§10).
  it('verwirft auch einen Fehler aus einem vorigen Set', async () => {
    let scheitern: ((ursache: unknown) => void) | undefined
    belegeApi({
      Zuruecknehmen: () =>
        new Promise<BewegungAntwort>((_, ablehnen) => {
          scheitern = ablehnen
        }),
    })

    const { rerender } = render(<Probe setID="1" />)
    await druecke('zurück')

    rerender(<Probe setID="2" />)

    await act(async () => {
      scheitern?.(new Error('kaputt'))
      await abwarten()
    })

    expect(fehler()).toBe('')
  })

  // **Auch in der Gegenrichtung.** Sie nimmt einen eigenen Aufruf; ein Wächter,
  // der nur die eine Hälfte hält, macht die andere unsichtbar (§10).
  it('verwirft auch eine Wiederherstellung aus einem vorigen Set', async () => {
    let loesen: ((antwort: BewegungAntwort) => void) | undefined
    belegeApi({
      Wiederherstellen: () =>
        new Promise<BewegungAntwort>((erfuellen) => {
          loesen = erfuellen
        }),
    })

    const { rerender } = render(<Probe setID="1" />)
    await druecke('vor')

    rerender(<Probe setID="2" />)

    await act(async () => {
      loesen?.(bewegt('7'))
      await abwarten()
    })

    expect(stand()).toBe('0')
  })
})
