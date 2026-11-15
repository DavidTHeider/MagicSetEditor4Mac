// **Die Leiste oben** — UX-14.AK-02
// (#1686):
// *„horizontale Anordnung ist ebenfalls möglich und erhält dieselben
// Aktionen, Kürzel, Fokuspfade und Aktivmarkierungen."* Geprüft wird, was
// davon im DOM steht; wie es aussieht, zeigt das Drehbuch `ux14-leistenlage`.
import { fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Rahmen } from './Rahmen'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { leistenname } from '../test/leistenweg'

afterEach(() => {
  raeumeApi()
})

function zeige(waagerecht: boolean, offen = true) {
  belegeApi({})
  const wechseln = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Rahmen
        aktuell="sets"
        wechseln={wechseln}
        seitenleiste={{ offen, umschalten: vi.fn(), waagerecht }}
      >
        <p>{kataloge.de.seiten.start.was}</p>
      </Rahmen>
    </Sprachraum>,
  )

  const navigation = screen.getByRole('navigation', {
    name: kataloge.de.allgemein.navigation.bezeichnung,
  })

  return {
    wechseln,
    navigation,
    eintraege: () => within(navigation).getAllByRole('button'),
  }
}

describe('Die Leiste oben', () => {
  it('steht als Zeile über dem Inhalt und hat nichts zum Einklappen', () => {
    const { navigation } = zeige(true)

    expect(navigation.closest('[data-lage]')).toHaveAttribute(
      'data-lage',
      'waagerecht',
    )
    expect(
      screen.queryByRole('button', {
        name: kataloge.de.allgemein.navigation.seitenleisteAusblenden,
      }),
    ).toBeNull()
  })

  // **Senkrecht bleibt, wie es war** — mit Knopf zum Einklappen.
  it('steht links, wo nichts anderes gewählt ist', () => {
    const { navigation } = zeige(false)

    expect(navigation.closest('[data-lage]')).toHaveAttribute(
      'data-lage',
      'senkrecht',
    )
    expect(
      screen.getByRole('button', {
        name: kataloge.de.allgemein.navigation.seitenleisteAusblenden,
      }),
    ).toBeInTheDocument()
  })

  // **Oben steht sie immer** — auch, wenn sie links eingeklappt war: eine
  // Zeile nimmt dem Inhalt keine Breite, und ohne sie gäbe es keinen Weg.
  it('steht oben auch dann, wenn sie links eingeklappt war', () => {
    const { navigation } = zeige(true, false)

    expect(navigation.closest('[data-lage]')).toBeVisible()
  })

  it('wandert mit → und ←, und ↓ lässt sie stehen', () => {
    const { eintraege } = zeige(true)
    const [erster, zweiter] = eintraege()
    if (erster === undefined || zweiter === undefined) {
      throw new Error('zu wenige Einträge')
    }

    erster.focus()
    fireEvent.keyDown(erster, { key: 'ArrowRight' })
    expect(zweiter).toHaveFocus()

    fireEvent.keyDown(zweiter, { key: 'ArrowLeft' })
    expect(erster).toHaveFocus()

    fireEvent.keyDown(erster, { key: 'ArrowDown' })
    expect(erster).toHaveFocus()

    // **Ende und Pos1 bleiben**, wie in der Spalte.
    const alle = eintraege()
    fireEvent.keyDown(erster, { key: 'End' })
    expect(alle[alle.length - 1]).toHaveFocus()
  })

  // **Oben steht das Kürzel im Titel** — die Zeile braucht ihre Breite für
  // die Namen; das Kürzel bleibt am Eintrag und in der Kürzeltafel.
  it('nennt das Kürzel im Titel des Eintrags statt daneben', () => {
    const { eintraege } = zeige(true)

    const sammeln = eintraege().find((knopf) =>
      knopf.textContent.includes(leistenname('sammlung')),
    )
    expect(sammeln).toHaveAttribute('title', '⌘3')
    expect(sammeln).not.toHaveTextContent('⌘3')
  })

  // **Oben zeigt die Palette nur ihr Kürzel** — und ihr Name enthält es, damit
  // wer „⌘K" sieht, den Knopf mit dem Namen findet, den er liest (WCAG 2.5.3).
  it('zeigt die Palette oben als Kürzel, mit vollem Namen', () => {
    zeige(true)

    expect(
      screen.getByRole('button', {
        name: new RegExp(`^${kataloge.de.allgemein.palette.oeffnen}\\s*⌘K$`),
      }),
    ).toBeInTheDocument()
  })

  it('markiert die aufgeschlagene Seite wie links', () => {
    const { eintraege } = zeige(true)

    const markiert = eintraege().filter(
      (knopf) => knopf.getAttribute('aria-current') === 'page',
    )
    expect(markiert).toHaveLength(1)
    expect(markiert[0]).toHaveTextContent(leistenname('sets'))
  })

  it('hält die Regeln der Barrierefreiheit', async () => {
    zeige(true)

    expect(await verstoesse(document.body)).toEqual([])
  })
})
