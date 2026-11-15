// **Das Kartenformular steht still, sobald es da ist**
// (#2301).
//
// Die Felder ordnet der Editor nach den **Flächen der Vorschau** — was der
// Stil zeichnet, in der Lage auf der Karte (#674). Vor der ersten Vorschau
// gab es keine Flächen: angeboten wurde, was die Karte führt, und kam die
// Vorschau, kam ein Bildfeld dazu, und alles darunter rückte. Am nativen
// Fenster sprang das Textfeld beim ersten Öffnen um 118 Punkte (`kartenruhe`),
// und ein Klick in diesen 50 ms traf den falschen Knopf.
//
// **Zwei Antworten**: beim Blättern ordnen die Flächen der vorigen Karte, bis
// die eigenen da sind — sie liegen in der Ansicht über dem Editor, die den
// Kartenwechsel übersteht. Und wo noch gar keine bekannt sind, steht das
// Formular unsichtbar da, bis die erste Vorschau antwortet oder scheitert.
import { act, render, screen, waitFor } from '@testing-library/react'
import { useState } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import type {
  FeldbeschreibungDTO,
  KarteDTO,
  VorschauAntwort,
  VorschauDTO,
} from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'
import { grundansicht, type Vorschauansicht } from './vorschauansicht'

afterEach(() => {
  raeumeApi()
})

function form(name: string, art = 'text'): FeldbeschreibungDTO {
  return {
    name,
    art,
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
    aufschrift: name,
  }
}

const formen = [form('name'), form('kunst', 'image'), form('text')]

// flaechen legt die Felder so, wie die Vorschau sie meldet: Name oben, das
// Bild darunter, der Text zuunterst.
const flaechen = [
  { feld: 'name', links: 0, oben: 0, breite: 10, hoehe: 5 },
  { feld: 'kunst', links: 0, oben: 10, breite: 10, hoehe: 5 },
  { feld: 'text', links: 0, oben: 20, breite: 10, hoehe: 5 },
]

const gezeichnet: VorschauDTO = {
  stil: '',
  gezeichnet: '',
  bild: '/vorschau/AAAA.png',
  breite: 10,
  hoehe: 14,
  dpi: 150,
  befunde: [],
  flaechen,
  nennungen: [],
  vorgaben: [],
}

const antwort: VorschauAntwort = { vorschau: gezeichnet }

function karte(id: string, stil = ''): KarteDTO {
  return {
    id,
    nummer: 1,
    fassung: 1,
    werte: [
      { feld: 'name', wert: `Karte ${id}` },
      { feld: 'text', wert: 'Regeltext' },
      ...(stil === '' ? [] : [{ feld: 'stylesheet', wert: stil }]),
    ],
  } as unknown as KarteDTO
}

// offeneVorschau gibt eine Vorschau, die erst antwortet, wenn der Test es sagt.
function offeneVorschau() {
  let antworte: (wert: VorschauAntwort) => void = () => undefined
  let scheitere: (grund: Error) => void = () => undefined
  const zusage = new Promise<VorschauAntwort>((ja, nein) => {
    antworte = ja
    scheitere = nein
  })

  return { holen: () => zusage, antworte, scheitere }
}

function Seite({
  id,
  anfang,
  stil = '',
}: {
  id: string
  anfang: Vorschauansicht
  stil?: string
}) {
  const ansicht = useState(anfang)

  return (
    <Sprachraum sprache="de">
      <Karteneditor
        key={id}
        setID="1"
        karte={karte(id, stil)}
        formen={formen}
        geaendert={() => undefined}
        ansicht={ansicht}
      />
    </Sprachraum>
  )
}

function formular(): HTMLFormElement {
  const gefunden = screen.getByRole('textbox', { name: 'name' }).closest('form')
  if (gefunden === null) {
    throw new Error('das Namensfeld steht in keinem Formular')
  }

  return gefunden
}

function reihenfolge(): string[] {
  return [...formular().querySelectorAll('[aria-label]')]
    .map((e) => e.getAttribute('aria-label') ?? '')
    .filter((name) => formen.some((f) => f.name === name))
}

describe('das Kartenformular beim Öffnen', () => {
  it('steht unsichtbar da, bis die erste Vorschau antwortet', async () => {
    const vorschau = offeneVorschau()
    belegeApi({ Vorschau: vorschau.holen })

    render(<Seite id="1" anfang={grundansicht} />)

    expect(formular()).toHaveAttribute('aria-busy', 'true')

    await act(async () => {
      vorschau.antworte(antwort)
      await Promise.resolve()
    })

    await waitFor(() => {
      expect(formular()).not.toHaveAttribute('aria-busy')
    })
    expect(reihenfolge()).toEqual(['name', 'kunst', 'text'])
  })

  it('zeigt sich auch, wenn die Vorschau scheitert', async () => {
    const vorschau = offeneVorschau()
    belegeApi({ Vorschau: vorschau.holen })

    render(<Seite id="1" anfang={grundansicht} />)

    await act(async () => {
      vorschau.scheitere(new Error('kein Bild'))
      await Promise.resolve()
    })

    await waitFor(() => {
      expect(formular()).not.toHaveAttribute('aria-busy')
    })
  })

  // **Beim Blättern** kennt die Ansicht die Flächen schon — die Felder stehen
  // sofort in der Lage der vorigen Karte, und nichts rückt, wenn die eigene
  // Vorschau kommt.
  it('ordnet beim Blättern nach den Flächen, die schon bekannt sind', async () => {
    const vorschau = offeneVorschau()
    belegeApi({ Vorschau: vorschau.holen })

    render(
      <Seite
        id="2"
        anfang={{ ...grundansicht, flaechenJeStil: { '': flaechen } }}
      />,
    )

    expect(formular()).not.toHaveAttribute('aria-busy')
    expect(reihenfolge()).toEqual(['name', 'kunst', 'text'])

    // **Und die eigene Vorschau kommt wirklich** (J05, #2269, Befund 5 des
    // nachgeholten Zweitblicks über die Folgepakete) — ohne sie belegte der
    // Fall nur die Lage davor.
    await act(async () => {
      vorschau.antworte(antwort)
      await Promise.resolve()
    })

    expect(formular()).not.toHaveAttribute('aria-busy')
    expect(reihenfolge()).toEqual(['name', 'kunst', 'text'])
  })

  // **Eine Karte mit anderem Stil wartet auf ihre eigenen Flächen** (J05,
  // #2269): ihre Felder können anders liegen, und die der vorigen Karte
  // ordneten das sichtbare Formular um, sobald die eigenen kamen.
  it('hält das Formular beim Blättern in einen anderen Stil zurück', async () => {
    const vorschau = offeneVorschau()
    belegeApi({ Vorschau: vorschau.holen })

    render(
      <Seite
        id="3"
        stil="m15-planeswalker"
        anfang={{ ...grundansicht, flaechenJeStil: { '': flaechen } }}
      />,
    )

    expect(formular()).toHaveAttribute('aria-busy', 'true')

    // **Mit ihrem Stil**, wie Go ihn nennt (`VorschauDTO.stil`) — die Flächen
    // gehören dann diesem und nicht dem des Sets.
    await act(async () => {
      vorschau.antworte({
        vorschau: { ...gezeichnet, stil: 'm15-planeswalker' },
      })
      await Promise.resolve()
    })

    await waitFor(() => {
      expect(formular()).not.toHaveAttribute('aria-busy')
    })
    expect(reihenfolge()).toEqual(['name', 'kunst', 'text'])
  })

  it('merkt sich die Flächen der Vorschau in der Ansicht', async () => {
    belegeApi({ Vorschau: () => Promise.resolve(antwort) })
    const gemerkt: Vorschauansicht[] = []

    function Beobachter() {
      const ansicht = useState(grundansicht)
      gemerkt.push(ansicht[0])

      return (
        <Sprachraum sprache="de">
          <Karteneditor
            setID="1"
            karte={karte('1')}
            formen={formen}
            geaendert={() => undefined}
            ansicht={ansicht}
          />
        </Sprachraum>
      )
    }

    render(<Beobachter />)

    await waitFor(() => {
      expect(gemerkt.at(-1)?.flaechenJeStil?.['']).toEqual(flaechen)
    })
  })
})
