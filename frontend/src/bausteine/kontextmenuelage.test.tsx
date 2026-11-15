// **Wo das Kontextmenü aufgeht und wann** — AP-5
// (#2362):
// der Anstoß von außen, die Lage am Fensterrand und im Fenster, und die
// Liste, die nur über ihre Zeilen öffnet. Herausgelöst aus
// `kontextmenue.test.tsx`, als beide zusammen über 400 Zeilen gingen (§2);
// dort steht, wie das Menü sich bedient.
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { Kontextmenue } from './Kontextmenue'

// **Ein Anstoß von außen öffnet das Menü an einer Stelle** — AP-5 (#2362):
// ein Rechtsklick auf eine andere Zeile wählt sie zuerst; geöffnet wird erst,
// wenn die Einträge der neu gewählten dastehen, und genau einmal.
describe('der Anstoß', () => {
  it('öffnet einmal, sobald es Einträge gibt, und meldet sich verbraucht', () => {
    const verbraucht = vi.fn()
    const { rerender } = render(
      <Kontextmenue
        bezeichnung="Handlungen"
        eintraege={[]}
        anstoss={{ x: 10, y: 10, nummer: 1 }}
        anstossVerbraucht={verbraucht}
      >
        <p>Zeile</p>
      </Kontextmenue>,
    )
    expect(screen.queryByRole('menu')).toBeNull()

    rerender(
      <Kontextmenue
        bezeichnung="Handlungen"
        eintraege={[{ schluessel: 'a', name: 'Öffnen', aufWahl: vi.fn() }]}
        anstoss={{ x: 10, y: 10, nummer: 1 }}
        anstossVerbraucht={verbraucht}
      >
        <p>Zeile</p>
      </Kontextmenue>,
    )

    expect(screen.getByRole('menu')).toBeInTheDocument()
    expect(verbraucht).toHaveBeenCalledTimes(1)
  })

  it('schließt, wenn ein anderes Objekt gewählt wird', () => {
    const eintraege = [{ schluessel: 'a', name: 'Öffnen', aufWahl: vi.fn() }]
    const { rerender } = render(
      <Kontextmenue bezeichnung="Karte" eintraege={eintraege} objekt="karte/1">
        <p>Zeile</p>
      </Kontextmenue>,
    )
    fireEvent.contextMenu(screen.getByText('Zeile'), { clientX: 4, clientY: 4 })
    expect(screen.getByRole('menu')).toBeInTheDocument()

    rerender(
      <Kontextmenue bezeichnung="Karte" eintraege={eintraege} objekt="karte/1">
        <p>Zeile</p>
      </Kontextmenue>,
    )
    expect(screen.getByRole('menu')).toBeInTheDocument()

    rerender(
      <Kontextmenue bezeichnung="Karte" eintraege={eintraege} objekt="karte/2">
        <p>Zeile</p>
      </Kontextmenue>,
    )
    expect(screen.queryByRole('menu')).toBeNull()
  })
})

// **Das Menü bleibt im Fenster** — AP-5 (#2362). Am Zeiger geöffnet, ragte
// es an einer unteren Zeile über den Rand und war dort nicht zu erreichen.
describe('die Lage am Fensterrand', () => {
  it('rückt nach oben und nach links, bis es ins Fenster passt', () => {
    const hoehe = window.innerHeight
    const breite = window.innerWidth
    const messen = vi
      .spyOn(HTMLElement.prototype, 'getBoundingClientRect')
      .mockImplementation(function (this: HTMLElement) {
        if (this.getAttribute('role') !== 'menu') {
          return new DOMRect(0, 0, 0, 0)
        }
        const oben = Number.parseFloat(this.style.top)
        const links = Number.parseFloat(this.style.left)

        return new DOMRect(links, oben, 200, 300)
      })

    render(
      <Kontextmenue
        bezeichnung="Handlungen"
        eintraege={[{ schluessel: 'a', name: 'Öffnen', aufWahl: vi.fn() }]}
      >
        <p>Zeile</p>
      </Kontextmenue>,
    )
    fireEvent.contextMenu(screen.getByText('Zeile'), {
      clientX: breite - 50,
      clientY: hoehe - 40,
    })

    const menue = screen.getByRole('menu')
    expect(Number.parseFloat(menue.style.top)).toBe(hoehe - 300 - 8)
    expect(Number.parseFloat(menue.style.left)).toBe(breite - 200 - 8)
    messen.mockRestore()
  })
})

// **In einer Liste öffnet nur die Zeile eines Objekts** — AP-5 (#2362): ein
// Rechtsklick auf eine Gruppe, einen Abstandhalter oder die leere Fläche
// zeigte sonst das Menü des gewählten Objekts, also eines anderen.
describe('nur angestoßen', () => {
  it('öffnet weder mit dem Rechtsklick noch mit ⇧F10 von selbst', () => {
    render(
      <Kontextmenue
        bezeichnung="Handlungen"
        eintraege={[{ schluessel: 'a', name: 'Öffnen', aufWahl: vi.fn() }]}
        nurAngestossen
      >
        <button type="button">Gruppe</button>
      </Kontextmenue>,
    )
    const gruppe = screen.getByRole('button', { name: 'Gruppe' })

    expect(fireEvent.contextMenu(gruppe)).toBe(true)
    fireEvent.keyDown(gruppe, { key: 'F10', shiftKey: true })

    expect(screen.queryByRole('menu')).toBeNull()
  })
})

// **Kein Vorfahr schneidet das Menü ab** — AP-5 (#2362), gemessen am
// 2026-10-16 mit `ap5kontext09`: in der Kartenliste schnitt die Spalte
// (`overflow`) das Menü auf »Karte anleg…« ab, die Kürzel standen
// unsichtbar dahinter. Es steht deshalb fest am Fenster; ob es wirklich
// ganz zu sehen ist, misst die Probe am Bild, jsdom kennt kein Abschneiden.
// kleinesMenue gibt dem Menü eine wirkliche Größe — die Vorrichtung aus
// `test/listenmass.ts` macht jede Liste 4.000 hoch, auch diese, und ein so
// hohes Menü rückte zu Recht an den oberen Rand.
function kleinesMenue() {
  return vi
    .spyOn(Element.prototype, 'getBoundingClientRect')
    .mockImplementation(function (this: Element) {
      return this.getAttribute('role') === 'menu'
        ? new DOMRect(0, 0, 200, 120)
        : new DOMRect()
    })
}

describe('die Lage im Fenster', () => {
  it('steht fest am Zeiger und bleibt bei seinem Objekt', () => {
    const messen = kleinesMenue()
    render(
      <div style={{ overflow: 'hidden' }}>
        <Kontextmenue
          bezeichnung="Karte"
          eintraege={[{ schluessel: 'a', name: 'Öffnen', aufWahl: vi.fn() }]}
          objekt="karte/7"
        >
          <p>Zeile</p>
        </Kontextmenue>
      </div>,
    )

    fireEvent.contextMenu(screen.getByText('Zeile'), {
      clientX: 120,
      clientY: 80,
    })

    const menue = screen.getByRole('menu')
    expect(menue.style.position).toBe('fixed')
    expect(menue.style.left).toBe('120px')
    expect(menue.style.top).toBe('80px')
    expect(menue.closest('[data-objekt]')).toHaveAttribute(
      'data-objekt',
      'karte/7',
    )
    messen.mockRestore()
  })

  it('schließt mit einem Klick daneben, nicht mit einem darin', async () => {
    render(
      <Kontextmenue
        bezeichnung="Karte"
        eintraege={[
          {
            schluessel: 'a',
            name: 'Öffnen',
            aufWahl: vi.fn(),
            angehalten: true,
          },
        ]}
      >
        <p>Zeile</p>
      </Kontextmenue>,
    )
    fireEvent.contextMenu(screen.getByText('Zeile'), { clientX: 5, clientY: 5 })

    fireEvent.mouseDown(screen.getByRole('menuitem', { name: 'Öffnen' }))
    expect(screen.getByRole('menu')).toBeInTheDocument()

    fireEvent.mouseDown(document.body)
    await waitFor(() => {
      expect(screen.queryByRole('menu')).toBeNull()
    })
  })
})

// **Der Rahmen gibt seine Ereignisse zuerst dem Anstoß** — AP-6 (#2363):
// eine Deckzeile ist ein frei gezeichneter Kasten, und Rechtsklick wie ⇧F10
// auf einer anderen Zeile sollen sie erst öffnen.
describe('anstossen am Rahmen', () => {
  it('bekommt Rechtsklick und Taste zuerst und hält das eigene Menü an', () => {
    const anstossen = vi.fn((ereignis: { preventDefault: () => void }) => {
      ereignis.preventDefault()
    })
    render(
      <Kontextmenue
        bezeichnung="Deck"
        eintraege={[{ schluessel: 'a', name: 'Öffnen', aufWahl: vi.fn() }]}
        anstossen={anstossen}
      >
        <button type="button">Zeile</button>
      </Kontextmenue>,
    )
    const zeile = screen.getByRole('button', { name: 'Zeile' })

    fireEvent.contextMenu(zeile)
    fireEvent.keyDown(zeile, { key: 'F10', shiftKey: true })

    expect(anstossen).toHaveBeenCalledTimes(2)
    expect(screen.queryByRole('menu')).toBeNull()
  })
})

// **Die Geometrie ändert sich, während das Menü offen ist oder wartet** —
// Befund 5 des Zweitblicks über Phase 6 von 0.9.0: beim Rollen und beim
// Verkleinern stand ein festes Menü getrennt von seinem Objekt, und ein
// wartender Anstoß ging an der alten Zeigerstelle auf, auch wenn seine Zeile
// inzwischen woanders stand (die Deckliste sortiert „zuletzt benutzt“).
describe('die Lage, wenn sich etwas bewegt', () => {
  const eintraege = [{ schluessel: 'a', name: 'Öffnen', aufWahl: vi.fn() }]

  // **Folgen, nicht schließen** — gemessen am nativen Fenster mit
  // `ap5kontext09`: ein Rechtsklick auf eine andere Karte rollt die Liste zu
  // ihr, und ein Menü, das beim Rollen schloss, war weg, bevor man es sah.
  it('folgt beim Rollen und Verkleinern seiner Zeile', () => {
    let oben = 100
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(
      function (this: HTMLElement) {
        return this.getAttribute('role') === 'menu'
          ? new DOMRect(0, 0, 0, 0)
          : new DOMRect(20, oben, 200, 40)
      },
    )
    render(
      <Kontextmenue bezeichnung="Deck" eintraege={eintraege}>
        <p>Zeile</p>
      </Kontextmenue>,
    )
    fireEvent.contextMenu(screen.getByText('Zeile'), {
      clientX: 30,
      clientY: 110,
    })
    expect(screen.getByRole('menu').style.top).toBe('110px')

    // **Rollen im Menü bewegt es nicht.**
    oben = 10
    act(() => {
      screen.getByRole('menu').dispatchEvent(new Event('scroll'))
    })
    expect(screen.getByRole('menu').style.top).toBe('110px')

    oben = 60
    act(() => {
      document.dispatchEvent(new Event('scroll'))
    })
    expect(screen.getByRole('menu').style.top).toBe('70px')

    oben = 80
    act(() => {
      window.dispatchEvent(new Event('resize'))
    })
    expect(screen.getByRole('menu').style.top).toBe('90px')
    vi.restoreAllMocks()
  })

  // **Und ohne Rollen** — Rest R2 des Nachlaufs: die Deckliste sortiert
  // „zuletzt benutzt“ und lädt nach, und die Zeile rückte unter dem offenen
  // Menü weg.
  it('folgt seiner Zeile, wenn die Liste sich umordnet', async () => {
    let oben = 100
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(
      function (this: HTMLElement) {
        return this.getAttribute('role') === 'menu'
          ? new DOMRect(0, 0, 0, 0)
          : new DOMRect(20, oben, 200, 40)
      },
    )
    render(
      <Kontextmenue bezeichnung="Deck" eintraege={eintraege}>
        <p>Zeile</p>
      </Kontextmenue>,
    )
    fireEvent.contextMenu(screen.getByText('Zeile'), {
      clientX: 30,
      clientY: 110,
    })

    oben = 240

    await waitFor(() => {
      expect(screen.getByRole('menu').style.top).toBe('250px')
    })

    vi.restoreAllMocks()
  })

  it('geht bei einem Anstoß an seiner Zeile auf, wo sie jetzt steht', () => {
    const messen = kleinesMenue()
    const zeile = document.createElement('div')
    zeile.getBoundingClientRect = () => new DOMRect(200, 300, 100, 40)
    document.body.append(zeile)

    render(
      <Kontextmenue
        bezeichnung="Deck"
        eintraege={eintraege}
        anstoss={{
          x: 50,
          y: 60,
          nummer: 1,
          anker: zeile,
          versatz: { x: 10, y: 20 },
        }}
      >
        <p>Zeile</p>
      </Kontextmenue>,
    )

    const menue = screen.getByRole('menu')
    expect(menue.style.left).toBe('210px')
    expect(menue.style.top).toBe('320px')
    zeile.remove()
    messen.mockRestore()
  })

  it('schließt beim Rollen, wenn es keinen Anker mehr gibt', () => {
    render(
      <Kontextmenue
        bezeichnung="Deck"
        eintraege={eintraege}
        anstoss={{ x: 5, y: 5, nummer: 1 }}
      >
        <p>Zeile</p>
      </Kontextmenue>,
    )
    expect(screen.getByRole('menu')).toBeInTheDocument()

    act(() => {
      document.dispatchEvent(new Event('scroll'))
    })

    expect(screen.queryByRole('menu')).toBeNull()
  })
})
