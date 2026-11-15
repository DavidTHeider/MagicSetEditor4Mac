// **Der Fokus folgt der Auswahl in die Kartenliste** —
// #1290.
//
// Der Fall, für den es diesen Haken gibt, ist der **späte**: die Zeile am Ende
// einer langen Liste ist nicht gezeichnet, wenn die Auswahl auf sie fällt. Sie
// entsteht erst nach dem Rollen — und dann muss der Fokus ihr folgen.

import { render } from '@testing-library/react'
import { useRef, useState } from 'react'
import { describe, expect, it } from 'vitest'

import { useZeilenfokus } from './useZeilenfokus'
import { effekteNachholen } from '../test/effekte'

function Probe({
  ziel,
  gezeichnet,
}: {
  ziel: string | undefined
  gezeichnet: readonly string[]
}) {
  const gitter = useRef<HTMLDivElement | null>(null)

  const [fenster] = useState<readonly [number, number]>([0, gezeichnet.length])

  useZeilenfokus(gitter, ziel, () => undefined, ziel, fenster)

  return (
    <div ref={gitter} role="grid">
      {gezeichnet.map((karte) => (
        <div
          key={karte}
          role="row"
          tabIndex={-1}
          data-karte={karte}
          aria-label={karte}
        />
      ))}
    </div>
  )
}

describe('useZeilenfokus', () => {
  it('setzt den Fokus auf die Zeile der gewählten Karte', async () => {
    const { container } = render(
      <Probe ziel="b" gezeichnet={['a', 'b', 'c']} />,
    )
    await effekteNachholen()

    expect(document.activeElement).toBe(
      container.querySelector('[data-karte="b"]'),
    )
  })

  it('tut nichts ohne Wunsch', async () => {
    render(<Probe ziel={undefined} gezeichnet={['a', 'b']} />)
    await effekteNachholen()

    expect(document.activeElement).toBe(document.body)
  })

  it('tut nichts, wenn der Wunsch auf eine andere Karte zeigt', async () => {
    // **Der Wunsch stammt aus der Taste, die Auswahl aus dem Fenster** — sie
    // laufen einen Herzschlag lang auseinander, und dann darf nichts
    // geschehen.
    function Auseinander() {
      const gitter = useRef<HTMLDivElement | null>(null)

      useZeilenfokus(gitter, 'c', () => undefined, 'a', [0, 2])

      return (
        <div ref={gitter} role="grid">
          <div role="row" tabIndex={-1} data-karte="a" aria-label="a" />
        </div>
      )
    }

    render(<Auseinander />)
    await effekteNachholen()

    expect(document.activeElement).toBe(document.body)
  })

  it('nimmt einen Fokus nicht weg, den jemand außerhalb gesetzt hat', async () => {
    // **Befund 6 des Zweitblicks über Phase 4.3.** Wer während des Rollens
    // etwas anderes anklickt, bekam den Fokus rückwirkend in die Zeile
    // gezogen.
    //
    // **Innerhalb des Gitters zählt nicht**: dort hat der Tastenweg begonnen.
    function MitFremdemFokus() {
      const gitter = useRef<HTMLDivElement | null>(null)
      const [da, setzeDa] = useState(false)

      useZeilenfokus(gitter, 'b', () => undefined, 'b', [0, da ? 1 : 0])

      return (
        <>
          <button
            type="button"
            onClick={() => {
              setzeDa(true)
            }}
          >
            woanders
          </button>
          <div ref={gitter} role="grid">
            {da ? (
              <div role="row" tabIndex={-1} data-karte="b" aria-label="b" />
            ) : null}
          </div>
        </>
      )
    }

    const { getByText, container } = render(<MitFremdemFokus />)
    await effekteNachholen()

    // **Der Fokus steht auf dem Knopf, und erst danach kommt die Zeile.** Ein
    // Klick setzt den Fokus dorthin und zeichnet sie im selben Zug.
    getByText('woanders').focus()
    getByText('woanders').click()
    await effekteNachholen()
    await new Promise((fertig) => setTimeout(fertig, 0))

    expect(container.querySelector('[data-karte="b"]')).toBeInTheDocument()
    expect(document.activeElement).toBe(getByText('woanders'))
  })

  it('wartet, bis die Zeile gezeichnet ist', async () => {
    // **Der Fall, für den es den Haken gibt.** „c" ist gewählt und steht noch
    // nicht im Baum; erst das Rollen bringt sie.
    function Spaeter() {
      const gitter = useRef<HTMLDivElement | null>(null)

      const [gezeichnet, setzeGezeichnet] = useState(['a', 'b'])

      useZeilenfokus(gitter, 'c', () => undefined, 'c', [0, gezeichnet.length])

      return (
        <div ref={gitter} role="grid">
          {gezeichnet.map((karte) => (
            <div
              key={karte}
              role="row"
              tabIndex={-1}
              data-karte={karte}
              aria-label={karte}
            />
          ))}
          <button
            type="button"
            onClick={() => {
              setzeGezeichnet(['a', 'b', 'c'])
            }}
          >
            rollen
          </button>
        </div>
      )
    }

    const { container, getByText } = render(<Spaeter />)
    await effekteNachholen()

    expect(document.activeElement).toBe(document.body)

    getByText('rollen').click()
    await effekteNachholen()

    expect(document.activeElement).toBe(
      container.querySelector('[data-karte="c"]'),
    )
  })
})
