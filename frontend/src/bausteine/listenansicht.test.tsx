// **Eine Liste, zwei Gestalten** — UX-11
// (#1683).
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Activity, useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { Listenansicht, seitengroesse, type Listenform } from './Listenansicht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { verstoesse } from '../test/axe'

const worte = kataloge.de.bausteine.listenansicht

interface Probe {
  id: string
  name: string
}

const drei: Probe[] = [
  { id: '1', name: 'Birke' },
  { id: '2', name: 'Ahorn' },
  { id: '3', name: 'Eiche' },
]

function Rahmen({
  eintraege = drei,
  anfang = 'liste',
  sortieren = true,
  eintrag = vi.fn(),
  festgehalten,
}: {
  eintraege?: Probe[]
  anfang?: Listenform
  sortieren?: boolean
  eintrag?: (probe: Probe, gestalt: string) => void
  festgehalten?: (probe: Probe) => boolean
}) {
  const [form, setzeForm] = useState<Listenform>(anfang)

  return (
    <Sprachraum sprache="de">
      <Listenansicht
        bezeichnung="Bäume"
        eintraege={eintraege}
        schluessel={(probe) => probe.id}
        suchtext={(probe) => probe.name}
        sortierungen={
          sortieren
            ? [
                {
                  name: 'bestand',
                  bezeichnung: 'Wie abgelegt',
                  vergleich: () => 0,
                },
                {
                  name: 'name',
                  bezeichnung: 'Nach Name',
                  vergleich: (a, b) => a.name.localeCompare(b.name, 'de'),
                },
              ]
            : []
        }
        form={form}
        setzeForm={setzeForm}
        leer={<p>Noch keine Bäume.</p>}
        festgehalten={festgehalten}
        eintrag={(probe, gestalt) => {
          eintrag(probe, gestalt)

          return (
            <button type="button" data-gestalt={gestalt}>
              {probe.name}
            </button>
          )
        }}
      />
    </Sprachraum>
  )
}

function namen() {
  return within(screen.getByRole('list'))
    .getAllByRole('button')
    .map((knopf) => knopf.textContent)
}

describe('Die Listenansicht', () => {
  it('sagt, dass es nichts gibt, wenn es nichts gibt', () => {
    render(<Rahmen eintraege={[]} />)

    expect(screen.getByText('Noch keine Bäume.')).toBeInTheDocument()
    expect(screen.queryByRole('list')).toBeNull()
  })

  // **Gefiltert leer ist nicht leer** (AK-04) — und der nächste Schritt ist
  // ein anderer.
  it('filtert, zählt und unterscheidet »nichts passt« von »nichts da«', async () => {
    render(<Rahmen />)

    const suche = screen.getByLabelText(worte.suche)
    await userEvent.type(suche, 'ei')
    expect(namen()).toEqual(['Eiche'])
    expect(screen.getByRole('status')).toHaveTextContent('1 von 3')

    await userEvent.type(suche, 'x')
    expect(screen.getByText('Nichts passt zu »eix«.')).toBeInTheDocument()
    await userEvent.click(
      screen.getByRole('button', { name: worte.sucheLeeren }),
    )
    expect(namen()).toEqual(['Birke', 'Ahorn', 'Eiche'])
  })

  // **Was festgehalten ist, bleibt stehen** — #2469: das offene Deck
  // verschwand aus der Liste, wenn die Suche es nicht traf, und mit ihm das
  // Menü „Deck“ in der Menüleiste.
  it('hält einen festgehaltenen Eintrag, auch wenn Suche und Seite ihn nicht treffen', async () => {
    const viele = Array.from({ length: seitengroesse + 5 }, (_, stelle) => ({
      id: String(stelle),
      name: `Baum ${String(stelle).padStart(3, '0')}`,
    }))
    const hinten = `Baum ${String(seitengroesse + 2).padStart(3, '0')}`
    render(
      <Rahmen
        eintraege={viele}
        festgehalten={(probe) => probe.name === hinten}
      />,
    )

    expect(namen()).toContain(hinten)
    expect(namen()).toHaveLength(seitengroesse + 1)

    await userEvent.type(screen.getByLabelText(worte.suche), 'Baum 001')
    expect(namen()).toEqual(['Baum 001', hinten])
  })

  it('ordnet um, wenn eine Ordnung gewählt wird', async () => {
    render(<Rahmen />)

    await userEvent.selectOptions(
      screen.getByLabelText(worte.sortierung),
      'name',
    )

    expect(namen()).toEqual(['Ahorn', 'Birke', 'Eiche'])
  })

  // **Dieselbe Handlung in beiden Gestalten** (AK-02) — und der Wechsel
  // lässt Suche und Ordnung stehen (AK-01).
  it('zeichnet denselben Eintrag in beiden Gestalten und behält Suche und Ordnung', async () => {
    const eintrag = vi.fn()
    render(<Rahmen eintrag={eintrag} />)

    await userEvent.type(screen.getByLabelText(worte.suche), 'e')
    await userEvent.selectOptions(
      screen.getByLabelText(worte.sortierung),
      'name',
    )
    const kacheln = screen.getByRole('radio', { name: worte.kacheln })
    expect(kacheln).not.toBeChecked()

    await userEvent.click(kacheln)

    expect(kacheln).toBeChecked()
    expect(screen.getByRole('list')).toHaveAttribute('data-form', 'kacheln')
    expect(screen.getByLabelText(worte.suche)).toHaveValue('e')
    expect(namen()).toEqual(['Birke', 'Eiche'])
    expect(eintrag).toHaveBeenCalledWith(drei[0], 'kachel')
    expect(eintrag).toHaveBeenCalledWith(drei[0], 'zeile')
  })

  // **Der Eintrag, an dem man war, bleibt es** (AK-01) — ein Klick ins
  // Segment fokussiert in WebKit nicht, der Fokus steht also noch im Eintrag,
  // und der wird beim Wechsel neu gezeichnet. `fireEvent` statt `userEvent`,
  // weil jsdom beim Klick fokussiert und WebKit nicht.
  it('gibt dem Eintrag, an dem man war, nach dem Wechsel wieder den Fokus', () => {
    render(<Rahmen />)

    screen.getByRole('button', { name: 'Eiche' }).focus()
    fireEvent.click(screen.getByRole('radio', { name: worte.kacheln }))

    expect(screen.getByRole('button', { name: 'Eiche' })).toHaveFocus()
  })

  // **Ein Fokus neben den Einträgen ändert nicht, an welchem man war** —
  // etwa auf der Liste selbst.
  it('merkt sich den Eintrag auch über einen Fokus neben ihm hinweg', () => {
    render(<Rahmen />)

    screen.getByRole('button', { name: 'Eiche' }).focus()
    fireEvent.focus(screen.getByRole('list'))
    fireEvent.click(screen.getByRole('radio', { name: worte.kacheln }))

    expect(screen.getByRole('button', { name: 'Eiche' })).toHaveFocus()
  })

  // **Passt nichts, gibt es keine Liste** — und ein Wechsel der Gestalt
  // sucht den Eintrag dann nicht.
  it('wechselt die Gestalt auch, wenn gerade nichts passt', async () => {
    render(<Rahmen />)

    screen.getByRole('button', { name: 'Eiche' }).focus()
    await userEvent.type(screen.getByLabelText(worte.suche), 'x')
    await userEvent.click(screen.getByRole('radio', { name: worte.kacheln }))

    expect(screen.getByText('Nichts passt zu »x«.')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: worte.kacheln })).toBeChecked()
  })

  // **Eine aufgehobene Seite nimmt beim Einblenden keinen Fokus** — gefunden
  // von der WebKit-Probe `ux10-tastatur` am 2026-09-26: wer auf die Decks
  // zurückkam, fand den Fokus im zuletzt geöffneten Deck statt im Inhalt,
  // denn `<Activity>` lässt die Effekte beim Zeigen neu laufen.
  it('nimmt beim Wiedereinblenden keinen Fokus', () => {
    const { rerender } = render(
      <Activity mode="visible">
        <Rahmen />
      </Activity>,
    )
    screen.getByRole('button', { name: 'Eiche' }).focus()
    const draussen = document.createElement('button')
    document.body.append(draussen)

    rerender(
      <Activity mode="hidden">
        <Rahmen />
      </Activity>,
    )
    draussen.focus()
    rerender(
      <Activity mode="visible">
        <Rahmen />
      </Activity>,
    )

    expect(draussen).toHaveFocus()
    draussen.remove()
  })

  // **Ein Wechsel ohne Eintrag lässt keine Marke liegen** — Nachlauf über
  // Phase 6.3 (N20): stand der Fokus beim Wechsel in keinem Eintrag, blieb die
  // Marke stehen, und das nächste Einblenden nahm ihn doch.
  it('nimmt auch nach einem Wechsel ohne Eintrag beim Einblenden keinen Fokus', async () => {
    const { rerender } = render(
      <Activity mode="visible">
        <Rahmen />
      </Activity>,
    )
    await userEvent.click(screen.getByRole('radio', { name: worte.kacheln }))
    screen.getByRole('button', { name: 'Eiche' }).focus()
    const draussen = document.createElement('button')
    document.body.append(draussen)

    rerender(
      <Activity mode="hidden">
        <Rahmen />
      </Activity>,
    )
    draussen.focus()
    rerender(
      <Activity mode="visible">
        <Rahmen />
      </Activity>,
    )

    expect(draussen).toHaveFocus()
    draussen.remove()
  })

  // **Ein Druck auf die Gestalt, die schon gilt, wechselt nichts** — und
  // der Fokus bleibt, wo er war: seit QA01 (#2382) nimmt ein Mausdruck auf
  // das Segment ihn nicht mehr (`fokusBleibt`), wie in WebKit ein Knopf.
  it('lässt die Gestalt, die schon gilt, wie sie ist', async () => {
    render(<Rahmen />)

    screen.getByRole('button', { name: 'Eiche' }).focus()
    await userEvent.click(screen.getByRole('radio', { name: worte.liste }))

    expect(screen.getByRole('list')).toHaveAttribute('data-form', 'liste')
    expect(screen.getByRole('button', { name: 'Eiche' })).toHaveFocus()
  })

  it('lässt den Fokus, wo er ist, wenn vorher keiner in der Liste stand', async () => {
    render(<Rahmen />)

    await userEvent.click(screen.getByRole('radio', { name: worte.kacheln }))

    expect(screen.getByRole('list')).toHaveAttribute('data-form', 'kacheln')
    expect(document.body).toHaveFocus()
  })

  // **Die Ansicht ist ein Segment** — AP-6 (#2363), Regel R7 aus
  // `docs/UIREVIEW.md`: zwei Werte einer Frage sind eine Optionsgruppe mit
  // **einer** Tabstelle, kein Paar Knöpfe. Ein Paar Knöpfe zählte zweimal
  // zu den Bedienelementen im Bild (R2), und das offene Deck lag darüber.
  it('stellt die Ansicht als Optionsgruppe mit Namen dar', () => {
    render(<Rahmen />)

    const gruppe = screen.getByRole('radiogroup', { name: worte.ansicht })
    expect(within(gruppe).getAllByRole('radio')).toHaveLength(2)
    expect(
      within(gruppe).getByRole('radio', { name: worte.liste }),
    ).toBeChecked()
    expect(screen.queryByRole('button', { name: worte.kacheln })).toBeNull()
  })

  it('wechselt mit dem Segment auch zurück zur Liste', async () => {
    render(<Rahmen />)

    await userEvent.click(screen.getByRole('radio', { name: worte.kacheln }))
    await userEvent.click(screen.getByRole('radio', { name: worte.liste }))

    expect(screen.getByRole('list')).toHaveAttribute('data-form', 'liste')
    expect(screen.getByRole('radio', { name: worte.liste })).toBeChecked()
  })

  // **Mit den Pfeilen bleibt der Fokus in der Gruppe** — auch wenn vorher
  // ein Eintrag ihn hatte: sonst ginge der zweite Pfeil ins Leere.
  it('lässt den Fokus beim Wählen mit den Pfeilen in der Gruppe', async () => {
    render(<Rahmen />)

    screen.getByRole('button', { name: 'Eiche' }).focus()
    screen.getByRole('radio', { name: worte.liste }).focus()
    await userEvent.keyboard('{ArrowRight}')

    expect(screen.getByRole('list')).toHaveAttribute('data-form', 'kacheln')
    expect(screen.getByRole('radio', { name: worte.kacheln })).toHaveFocus()
  })

  // **Seitenweise** (AK-03): große Bestände zeichnen nicht alles auf einmal.
  it('zeigt große Bestände seitenweise', async () => {
    const viele = Array.from({ length: seitengroesse + 1 }, (_, stelle) => ({
      id: String(stelle),
      name: `Baum ${String(stelle)}`,
    }))
    render(<Rahmen eintraege={viele} />)

    expect(namen()).toHaveLength(seitengroesse)
    await userEvent.click(
      screen.getByRole('button', { name: 'Weitere 1 zeigen' }),
    )
    expect(namen()).toHaveLength(seitengroesse + 1)
    expect(screen.queryByRole('button', { name: /^Weitere/ })).toBeNull()
  })

  it('kommt ohne Ordnungen aus und behält dann die Reihenfolge', () => {
    render(<Rahmen sortieren={false} />)

    expect(screen.queryByLabelText(worte.sortierung)).toBeNull()
    expect(namen()).toEqual(['Birke', 'Ahorn', 'Eiche'])
  })

  it('hält die Regeln der Barrierefreiheit', async () => {
    const { container } = render(<Rahmen anfang="kacheln" />)

    expect(await verstoesse(container)).toEqual([])
  })
})
