// **Wie hoch ein Bereich sein darf, damit er bis an den Fensterrand reicht**
// — Befund F09 aus der nativen Prüfung von 0.5.0
// (#1675,
// #1680).
import { act, render, screen } from '@testing-library/react'
import { useRef } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useFensterrest } from './useFensterrest'

// Probe zeigt die gemessene Höhe als Text — gemessen wird am echten Knoten.
function Probe({ unten, mindestens }: { unten: number; mindestens: number }) {
  const knoten = useRef<HTMLDivElement>(null)
  const hoehe = useFensterrest(knoten, { unten, mindestens })

  return (
    <div ref={knoten} data-testid="bereich">
      {hoehe}
    </div>
  )
}

// obenBei legt den oberen Rand des Bereichs fest, relativ zum Fenster.
function obenBei(oben: number) {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
    top: oben,
  } as DOMRect)
}

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

describe('useFensterrest', () => {
  it('reicht vom oberen Rand des Bereichs bis an den unteren des Fensters', async () => {
    vi.stubGlobal('innerHeight', 900)
    obenBei(200)

    render(<Probe unten={32} mindestens={400} />)

    expect(await screen.findByText('668')).toBeInTheDocument()
  })

  // **Gescrollt ändert nichts**: gemessen wird der Abstand zum Anfang des
  // Dokuments, nicht zum Fensterrand von jetzt.
  it('rechnet einen gescrollten Stand auf den Anfang des Dokuments zurück', async () => {
    vi.stubGlobal('innerHeight', 900)
    vi.stubGlobal('scrollY', 150)
    obenBei(50)

    render(<Probe unten={32} mindestens={400} />)

    expect(await screen.findByText('668')).toBeInTheDocument()
  })

  it('wird nie kleiner als die Mindesthöhe', async () => {
    vi.stubGlobal('innerHeight', 500)
    obenBei(200)

    render(<Probe unten={32} mindestens={400} />)

    expect(await screen.findByText('400')).toBeInTheDocument()
  })

  it('misst nach, wenn das Fenster sich ändert', async () => {
    vi.stubGlobal('innerHeight', 900)
    obenBei(200)

    render(<Probe unten={32} mindestens={400} />)
    await screen.findByText('668')

    vi.stubGlobal('innerHeight', 1000)
    act(() => {
      window.dispatchEvent(new Event('resize'))
    })

    expect(await screen.findByText('768')).toBeInTheDocument()
  })

  // **Und wenn darüber etwas erscheint** — ein Meldungsband, das Band einer
  // neuen Fassung: das Fenster bleibt gleich groß, das Dokument nicht.
  it('misst nach, wenn das Dokument seine Größe ändert', async () => {
    let melden: (() => void) | undefined

    vi.stubGlobal(
      'ResizeObserver',
      class {
        constructor(rueckruf: () => void) {
          melden = rueckruf
        }

        observe() {
          return undefined
        }

        disconnect() {
          melden = undefined
        }
      },
    )
    vi.stubGlobal('innerHeight', 900)
    obenBei(200)

    const { unmount } = render(<Probe unten={32} mindestens={400} />)
    await screen.findByText('668')

    obenBei(260)
    act(() => {
      melden?.()
    })

    expect(await screen.findByText('608')).toBeInTheDocument()

    unmount()
    expect(melden).toBeUndefined()
  })

  // **Im nächsten Bild, nicht im Rückruf** (QA01,
  // #2382):
  // wer im Rückruf neu zeichnet, ändert die Größe, die ihm gerade gemeldet
  // wurde, und WebKit wirft »ResizeObserver loop completed with undelivered
  // notifications«.
  it('zeichnet nach einer Meldung des Dokuments erst im nächsten Bild neu', async () => {
    let melden: (() => void) | undefined
    const bilder = new Map<number, FrameRequestCallback>()
    let naechstes = 0

    vi.stubGlobal(
      'ResizeObserver',
      class {
        constructor(rueckruf: () => void) {
          melden = rueckruf
        }

        observe() {
          return undefined
        }

        disconnect() {
          return undefined
        }
      },
    )
    vi.stubGlobal('requestAnimationFrame', (rueckruf: FrameRequestCallback) => {
      naechstes += 1
      bilder.set(naechstes, rueckruf)
      return naechstes
    })
    vi.stubGlobal('cancelAnimationFrame', (nummer: number) => {
      bilder.delete(nummer)
    })
    vi.stubGlobal('innerHeight', 900)
    obenBei(200)

    const { unmount } = render(<Probe unten={32} mindestens={400} />)
    await screen.findByText('668')

    obenBei(260)
    act(() => {
      melden?.()
      melden?.()
    })

    expect(screen.getByTestId('bereich')).toHaveTextContent('668')
    expect(bilder.size).toBe(1)

    act(() => {
      for (const rueckruf of bilder.values()) rueckruf(0)
    })

    expect(screen.getByTestId('bereich')).toHaveTextContent('608')

    melden?.()
    unmount()

    expect(bilder.size).toBe(0)
  })
})
