// **Die Seitenleiste ist eine Tabstelle** — #1488: genau ein Eintrag trägt
// `tabIndex 0` (die aufgeschlagene Seite), ↑ und ↓ wandern, Home und End
// springen, und die Kürzel ⌘1–⌘8 bleiben. Derselbe Vertrag wie die Paketliste
// (#1290) und der Vorlagenbaum (U7).
import { fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Rahmen } from './Rahmen'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { leistenname } from '../test/leistenweg'

function zeige(aktuell: Parameters<typeof Rahmen>[0]['aktuell'] = 'start') {
  belegeApi({})
  const wechseln = vi.fn()

  const ansicht = render(
    <Sprachraum sprache="de">
      <Rahmen aktuell={aktuell} wechseln={wechseln}>
        <p>{kataloge.de.seiten.start.was}</p>
      </Rahmen>
    </Sprachraum>,
  )

  // **Mit Namen, seit es zwei gibt** — UX-03: über dem Kopf steht der Weg,
  // und der ist ebenfalls ein Orientierungspunkt. Eine Abfrage ohne Namen
  // fände beide.
  const navigation = within(
    screen.getByRole('navigation', {
      name: kataloge.de.allgemein.navigation.bezeichnung,
    }),
  )
  const eintraege = () => navigation.getAllByRole('button')
  // eintrag gibt den n-ten Eintrag — oder scheitert laut, statt still.
  const eintrag = (n: number): HTMLElement => {
    const gefunden = eintraege()[n]
    if (gefunden === undefined) {
      throw new Error(`kein Eintrag ${String(n)}`)
    }

    return gefunden
  }

  return { ...ansicht, wechseln, navigation, eintraege, eintrag }
}

afterEach(() => {
  raeumeApi()
})

describe('Die Seitenleiste als eine Tabstelle', () => {
  it('lässt genau einen Eintrag mit Tab erreichen — die aufgeschlagene Seite', () => {
    const { eintraege } = zeige('sets')

    const halte = eintraege().filter((knopf) => knopf.tabIndex === 0)
    expect(halte).toHaveLength(1)
    expect(halte[0]).toHaveTextContent(leistenname('sets'))
    for (const knopf of eintraege()) {
      if (knopf !== halte[0]) {
        expect(knopf.tabIndex).toBe(-1)
      }
    }
  })

  it('wandert mit ↓ und ↑ und nimmt den Halt mit', () => {
    const { eintrag } = zeige('start')

    const erster = eintrag(0)
    const zweiter = eintrag(1)
    erster.focus()
    fireEvent.keyDown(erster, { key: 'ArrowDown' })

    expect(zweiter).toHaveFocus()
    expect(zweiter.tabIndex).toBe(0)
    expect(erster.tabIndex).toBe(-1)

    fireEvent.keyDown(zweiter, { key: 'ArrowUp' })
    expect(erster).toHaveFocus()

    // Am Anfang bleibt ↑ stehen — kein Umlauf.
    fireEvent.keyDown(erster, { key: 'ArrowUp' })
    expect(erster).toHaveFocus()
  })

  it('springt mit End ans Ende und mit Home an den Anfang', () => {
    const { eintraege, eintrag } = zeige('start')

    const alle = eintraege()
    const erster = eintrag(0)
    const letzter = eintrag(alle.length - 1)
    erster.focus()

    fireEvent.keyDown(erster, { key: 'End' })
    expect(letzter).toHaveFocus()

    fireEvent.keyDown(letzter, { key: 'Home' })
    expect(erster).toHaveFocus()
  })

  // **Der Halt folgt der Seite**: wer über ⌘4 zu den Paketen wechselt, findet
  // beim nächsten Tab die Pakete, nicht die Seite von vorhin.
  it('folgt der aufgeschlagenen Seite', () => {
    const { rerender, eintraege } = zeige('start')

    rerender(
      <Sprachraum sprache="de">
        <Rahmen aktuell="pakete" wechseln={vi.fn()}>
          <p>{kataloge.de.seiten.start.was}</p>
        </Rahmen>
      </Sprachraum>,
    )

    const halte = eintraege().filter((knopf) => knopf.tabIndex === 0)
    expect(halte).toHaveLength(1)
    expect(halte[0]).toHaveTextContent(leistenname('pakete'))
  })

  it('lässt die Kürzel und den Klick in Ruhe', () => {
    const { wechseln, eintrag } = zeige('start')

    fireEvent.keyDown(window, { key: '2', metaKey: true })
    expect(wechseln).toHaveBeenCalledWith('sets')

    // Eine Taste mit ⌘ ist kein Wandern in der Liste.
    const erster = eintrag(0)
    const zweiter = eintrag(1)
    erster.focus()
    fireEvent.keyDown(erster, {
      key: 'ArrowDown',
      metaKey: true,
    })
    expect(erster).toHaveFocus()
    expect(zweiter.tabIndex).toBe(-1)

    fireEvent.click(zweiter)
    expect(wechseln).toHaveBeenCalledWith('sets')
  })

  // **Andere Tasten gehen durch** — Enter klickt, ein Buchstabe tut nichts.
  it('lässt andere Tasten in Ruhe', () => {
    const { eintrag } = zeige('start')

    const erster = eintrag(0)
    const zweiter = eintrag(1)
    erster.focus()
    fireEvent.keyDown(erster, { key: 'a' })

    expect(erster).toHaveFocus()
    expect(zweiter.tabIndex).toBe(-1)
  })

  // **Eine Setlasche ist keine Leistenseite**: der Halt bleibt beim ersten
  // Eintrag, statt ins Leere zu zeigen.
  it('hält beim ersten Eintrag, wenn eine Setlasche offen ist', () => {
    const { eintraege } = zeige('konsole')

    const halte = eintraege().filter((knopf) => knopf.tabIndex === 0)
    expect(halte).toHaveLength(1)
    expect(halte[0]).toBe(eintraege()[0])
  })

  it('hält die Regeln der Barrierefreiheit', async () => {
    const { container } = zeige('start')
    expect(await verstoesse(container)).toEqual([])
  })
})
