// **Jede Einstellung hat das Bedienelement ihrer Werte** — UXR-F2
// (#2008),
// Regel R7 aus `docs/UIREVIEW.md`: *Schalter für Sofortwirkung, Segmente bis
// vier Werte, Auswahl darüber.*
//
// # Was er hält
//
// **Jede Zeile des Katalogs, nicht eine Auswahl davon.** Die Zeilen kommen
// aus `docs/EINSTELLUNGEN.md` (`test/einstellungskatalog.ts`) — der Tafel, die
// `internal/werkzeuge/einstellungsschema_test.go` gegen `internal/domain`
// hält. Eine neue Einstellung steht damit hier, ohne dass jemand diese Datei
// anfasst, und bekommt ihr Bedienelement von der Regel statt von der
// Gewohnheit.
//
// Je Zeile gilt dreierlei:
//
// 1. **Das Element passt zur Zahl der Werte**: ohne Werteliste ein Textfeld;
//    zwei Werte, die an und aus sind, ein Schalter; zwei bis vier sonst
//    Segmente; mehr eine Auswahl.
// 2. **Es gibt genau eines** — gezählt am Schlüssel, nicht an der Aufschrift
//    (`einstellungsdoppelung.test.tsx`).
// 3. **Kein Wert geht verloren**: wer `erben` wählen konnte, kann es
//    weiterhin, und jeder Wert ist von jedem anderen aus erreichbar.
//
// # Was er nicht hält
//
// **Ob das Element gut aussieht** und ob es im engen Fenster umbricht — das
// sieht nur das Fenster. Und die Regel selbst steht hier ein zweites Mal, in
// den Worten des Issues; ändert jemand sie, ändert er zwei Stellen mit
// Absicht.
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import {
  alsEintrag,
  elementVon,
  einstellungskatalog,
  erwartetesElement,
  type Rolle,
} from '../test/einstellungskatalog'
import { Wahl } from './Wahl'

import type { Katalogzeile } from '../test/einstellungskatalog'

const katalog = einstellungskatalog()

// bezeichnung ist, wie die Zeile im Test heißt.
const bezeichnung = (zeile: Katalogzeile) =>
  `${zeile.schluessel} (${zeile.tafel}, ${String(zeile.werte.length)} Werte)`

function zeige(zeile: Katalogzeile, wert?: string) {
  const aendern = vi.fn()

  const { container } = render(
    <Sprachraum sprache="de">
      <Wahl
        eintrag={alsEintrag(zeile, wert)}
        angehalten={false}
        aendern={aendern}
      />
    </Sprachraum>,
  )

  return { aendern, container }
}

// element nennt, was die Zeile gezeichnet hat — am Schlüssel gefunden.
function element(container: HTMLElement, schluessel: string): Rolle {
  const treffer = container.querySelectorAll(
    `[data-einstellung="${schluessel}"]`,
  )

  expect(treffer, `${schluessel} steht genau einmal da`).toHaveLength(1)

  return elementVon(treffer[0] as HTMLElement)
}

describe('der Katalog, den dieser Lauf liest', () => {
  // **Eine Gegenprobe, dass er etwas sieht**: ohne Zeilen wäre jeder Lauf
  // darunter grün und leer (§3).
  it('hat Zeilen aus beiden Tafeln und aus allen vier Arten', () => {
    expect(katalog.length).toBeGreaterThan(100)
    expect(new Set(katalog.map((zeile) => zeile.tafel))).toEqual(
      new Set(['profil', 'mensch']),
    )
    expect(
      new Set(
        katalog.map((zeile) =>
          erwartetesElement(zeile.werte, zeile.schluessel),
        ),
      ),
    ).toEqual(new Set(['textbox', 'switch', 'radiogroup', 'combobox']))
    expect(katalog.some((zeile) => zeile.werte.includes('erben'))).toBe(true)
  })

  it('führt den Erbwert in Segmenten wie in Auswahllisten', () => {
    const mitErben = katalog.filter((zeile) => zeile.werte.includes('erben'))

    expect(
      new Set(
        mitErben.map((zeile) =>
          erwartetesElement(zeile.werte, zeile.schluessel),
        ),
      ),
    ).toEqual(new Set(['radiogroup', 'combobox']))
  })
})

describe('jede Einstellung des Katalogs', () => {
  it.each(katalog.map((zeile) => [bezeichnung(zeile), zeile] as const))(
    '%s hat das Bedienelement ihrer Werte',
    (_name, zeile) => {
      const { container } = zeige(zeile)

      expect(element(container, zeile.schluessel)).toBe(
        erwartetesElement(zeile.werte, zeile.schluessel),
      )
    },
  )

  it.each(katalog.map((zeile) => [bezeichnung(zeile), zeile] as const))(
    '%s trägt einen Namen',
    (_name, zeile) => {
      const { container } = zeige(zeile)
      const art = erwartetesElement(zeile.werte, zeile.schluessel)
      const knoten = container.querySelector(
        `[data-einstellung="${zeile.schluessel}"]`,
      )

      // **Der Name steht am Bedienelement** und nicht nur neben ihm: ein
      // Vorleseprogramm liest ihn beim Eintreten.
      expect(
        screen.getByRole(art, { name: /\S/ }),
        `${zeile.schluessel} hat einen Namen`,
      ).toBe(knoten)
    },
  )

  // **Jeder Wert ist erreichbar** — auch `erben`, auch der letzte. Von einem
  // anderen Wert aus gewählt, denn der gewählte meldet nichts.
  it.each(
    katalog
      .filter((zeile) => zeile.werte.length > 0)
      .map((zeile) => [bezeichnung(zeile), zeile] as const),
  )('%s lässt jeden Wert wählen', async (_name, zeile) => {
    const art = erwartetesElement(zeile.werte, zeile.schluessel)

    for (const ziel of zeile.werte) {
      const von = zeile.werte.find((wert) => wert !== ziel) ?? ziel
      const { aendern } = zeige(zeile, von)

      if (art === 'switch') {
        expect(zeile.werte, 'ein Schalter hat zwei Werte').toHaveLength(2)
        await userEvent.click(screen.getByRole('switch'))
      } else if (art === 'radiogroup') {
        const feld = screen
          .getAllByRole('radio')
          .find((kandidat) => kandidat.getAttribute('value') === ziel)

        if (feld === undefined) {
          throw new Error(`${zeile.schluessel}: kein Segment für „${ziel}“`)
        }

        await userEvent.click(feld)
      } else {
        await userEvent.selectOptions(screen.getByRole('combobox'), ziel)
      }

      expect(aendern).toHaveBeenCalledExactlyOnceWith(ziel)
      cleanup()
    }
  })

  // **Und es sind genau die Werte des Katalogs** — nicht mehr, nicht weniger,
  // in der Reihenfolge, in der Go sie nennt.
  it.each(
    katalog
      .filter(
        (zeile) =>
          erwartetesElement(zeile.werte, zeile.schluessel) !== 'switch',
      )
      .filter((zeile) => zeile.werte.length > 0)
      .map((zeile) => [bezeichnung(zeile), zeile] as const),
  )('%s zeigt genau ihre Werte', (_name, zeile) => {
    const { container } = zeige(zeile)
    const felder =
      erwartetesElement(zeile.werte, zeile.schluessel) === 'radiogroup'
        ? screen.getAllByRole('radio').map((feld) => feld.getAttribute('value'))
        : [...container.querySelectorAll('option')].map(
            (option) => option.value,
          )

    expect(felder).toEqual(zeile.werte)
  })
})
