// **Die Faltliste** — UXR-08
// (#1916).
//
// Geprüft wird, was sie zusagt: vorn so viele wie verlangt, der Rest
// stückweise, die Deckelung sichtbar — und der Fokus bleibt im Spiel, wenn der
// Knopf mit dem letzten Stück verschwindet.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Faltliste } from './Faltliste'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'

const worte = kataloge.de.bausteine.listenansicht

// zahlen sind `anzahl` Einträge — „Eintrag 1" bis „Eintrag anzahl".
function zahlen(anzahl: number): string[] {
  return Array.from({ length: anzahl }, (_, stelle) =>
    ['Eintrag', String(stelle + 1)].join(' '),
  )
}

function zeige(eintraege: string[], geordnet = false) {
  return render(
    <Sprachraum sprache="de">
      <Faltliste
        eintraege={eintraege}
        vorn={10}
        schritt={100}
        geordnet={geordnet}
        klasse="liste"
        zeile={(eintrag) => <li key={eintrag}>{eintrag}</li>}
      />
    </Sprachraum>,
  )
}

function weitere(anzahl: number) {
  return screen.getByRole('button', {
    name: einsetzen(worte.weitere, { anzahl: String(anzahl) }),
  })
}

function stand(gezeigt: number, gesamt: number) {
  return screen.getByText(
    einsetzen(worte.zahl, {
      gezeigt: String(gezeigt),
      gesamt: String(gesamt),
    }),
  )
}

describe('Die Faltliste', () => {
  it('zeigt vorn so viele wie verlangt und sagt, wie viele es sind', () => {
    zeige(zahlen(30))

    expect(screen.getAllByRole('listitem')).toHaveLength(10)
    expect(stand(10, 30)).toHaveAttribute('role', 'status')
    expect(weitere(20)).toBeInTheDocument()
  })

  // **Stückweise und nicht alles auf einmal** — 3.756 Balken auf einen Klick
  // wären dieselbe Wand, nur später.
  it('zeigt den Rest stückweise', async () => {
    zeige(zahlen(250))

    await userEvent.click(weitere(100))
    expect(screen.getAllByRole('listitem')).toHaveLength(110)

    await userEvent.click(weitere(100))
    expect(screen.getAllByRole('listitem')).toHaveLength(210)
    expect(stand(210, 250)).toBeInTheDocument()

    // **Der Fokus bleibt auf dem Knopf**, solange es ihn gibt.
    expect(weitere(40)).toHaveFocus()
  })

  // **Mit dem letzten Stück verschwindet der Knopf** — und der Fokus stünde
  // sonst auf dem Dokument.
  it('gibt den Fokus an die Zahl, wenn der Knopf verschwindet', async () => {
    zeige(zahlen(30))

    await userEvent.click(weitere(20))

    expect(screen.getAllByRole('listitem')).toHaveLength(30)
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
    expect(stand(30, 30)).toHaveFocus()
  })

  // **Eine kurze Liste bleibt, wie sie war** — ohne Zahl und ohne Knopf.
  it('zeigt eine kurze Liste ganz und ohne Zusatz', () => {
    zeige(zahlen(10))

    expect(screen.getAllByRole('listitem')).toHaveLength(10)
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('nummeriert, wenn sie geordnet ist', () => {
    const { container } = zeige(zahlen(3), true)

    const liste = within(container).getByRole('list')

    expect(liste.tagName).toBe('OL')
    expect(liste).toHaveClass('liste')
  })

  // **Was offen ist, bleibt im Bild** — #2035: die Liste der Gegenüber faltet,
  // und ein eben aufgenommener Fund, der hinten steht und mit seinem
  // Vergleich aufschlägt, darf nicht hinter der Falte verschwinden.
  it('zeigt auch hinter der Falte, was immer dastehen soll', async () => {
    render(
      <Sprachraum sprache="de">
        <Faltliste
          eintraege={zahlen(8)}
          vorn={3}
          schritt={10}
          immer={(eintrag) => eintrag === 'Eintrag 7'}
          klasse="liste"
          zeile={(eintrag) => <li key={eintrag}>{eintrag}</li>}
        />
      </Sprachraum>,
    )

    expect(
      within(screen.getByRole('list'))
        .getAllByRole('listitem')
        .map((zeile) => zeile.textContent),
    ).toEqual(['Eintrag 1', 'Eintrag 2', 'Eintrag 3', 'Eintrag 7'])
    expect(stand(4, 8)).toBeInTheDocument()

    await userEvent.click(weitere(4))

    expect(
      within(screen.getByRole('list')).getAllByRole('listitem'),
    ).toHaveLength(8)
  })
})
