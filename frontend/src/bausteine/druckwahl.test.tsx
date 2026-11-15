// **Drucke aus der eigenen Sammlung wählen** — UXR-F18
// (#2033).
//
// # Was hier gemessen wird
//
// **R7 aus `docs/UIREVIEW.md`**: bis vierzehn Drucke eine Liste mit Kästchen,
// ab fünfzehn ein Suchfeld davor — und die Liste zeichnet dann nie mehr als
// die gewählten und zwölf Treffer (R6: eine Sammlung mit tausend Drucken
// stellt keine tausend Kästchen).
//
// **Und die Regeln der Domäne**: ein Wunsch ist kein Besitz
// (`Exemplar.Besitzt`), und die Auswahl benennt den **Druck** —
// `Sammlungsfreigabe.Zeigt` vergleicht `Printing`, nicht die Zeile. Zwei
// Exemplare desselben Drucks sind deshalb eine Zeile mit zusammengezählter
// Menge.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Druckwahl } from './Druckwahl'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { exemplar } from '../test/sammlungsseite'

import type { ExemplarDTO } from '../api'

const worte = kataloge.de.bausteine.druckwahl

// drucke gibt `anzahl` verschiedene Drucke, je ein Exemplar.
function drucke(anzahl: number): ExemplarDTO[] {
  return Array.from({ length: anzahl }, (_, stelle) =>
    exemplar({
      id: stelle + 1,
      printing: `druck-${String(stelle + 1)}`,
      bezeichnung: `Karte ${String(stelle + 1).padStart(3, '0')}`,
      menge: 1,
    }),
  )
}

// zeige stellt die Wahl auf, mit der Sammlung, die die Bindung gibt.
function zeige(
  exemplare: ExemplarDTO[] | Error,
  eigenschaften: Partial<Parameters<typeof Druckwahl>[0]> = {},
) {
  const waehlen = vi.fn()

  belegeApi({
    Sammlung: () =>
      exemplare instanceof Error
        ? Promise.reject(exemplare)
        : Promise.resolve({ exemplare, fehler: null }),
  })

  const { container } = render(
    <Sprachraum sprache="de">
      <Druckwahl gewaehlt={[]} waehlen={waehlen} {...eigenschaften} />
    </Sprachraum>,
  )

  return { waehlen, hand: userEvent.setup(), container }
}

afterEach(() => {
  raeumeApi()
})

describe('Drucke wählen', () => {
  it('zeigt je Druck eine Zeile mit der zusammengezählten Menge', async () => {
    zeige([
      exemplar({
        id: 1,
        printing: 'bolt',
        bezeichnung: 'Blitzschlag · M10 #146',
        menge: 2,
      }),
      // **Derselbe Druck in einer anderen Ausführung** — eine Zeile.
      exemplar({
        id: 2,
        printing: 'bolt',
        bezeichnung: 'Blitzschlag · M10 #146',
        menge: 1,
        finish: 'foil',
      }),
      exemplar({
        id: 3,
        printing: 'giant',
        bezeichnung: 'Riesenwuchs · M10 #200',
        menge: 4,
      }),
    ])

    expect(await screen.findAllByRole('checkbox')).toHaveLength(2)
    expect(
      screen.getByRole('checkbox', { name: /3 × Blitzschlag · M10 #146/ }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('checkbox', { name: /4 × Riesenwuchs/ }),
    ).toBeInTheDocument()
  })

  // **Die Bindung gibt die Sammlung nach der Kennung des Drucks** — eine
  // Folge von Namen in dieser Ordnung läse sich wie zufällig.
  it('ordnet die Drucke nach ihrem Namen', async () => {
    zeige([
      exemplar({
        id: 1,
        printing: 'a-zuletzt',
        bezeichnung: 'Zitrone',
        menge: 1,
      }),
      exemplar({
        id: 2,
        printing: 'm-mitte',
        bezeichnung: 'Birne',
        menge: 1,
      }),
      exemplar({
        id: 3,
        printing: 'z-zuerst',
        bezeichnung: 'Apfel',
        menge: 1,
      }),
    ])

    const kaestchen = await screen.findAllByRole('checkbox')

    expect(
      kaestchen.map((eines) => eines.closest('label')?.textContent),
    ).toEqual(['1 × Apfel', '1 × Birne', '1 × Zitrone'])
  })

  it('führt keinen Wunsch auf — er ist kein Besitz', async () => {
    zeige([
      exemplar({ id: 1, printing: 'bolt', bezeichnung: 'Blitzschlag' }),
      exemplar({
        id: 2,
        printing: 'wunsch',
        bezeichnung: 'Nur gewünscht',
        wunsch: true,
      }),
    ])

    expect(await screen.findAllByRole('checkbox')).toHaveLength(1)
    expect(screen.queryByText(/Nur gewünscht/)).toBeNull()
  })

  it('wählt einen Druck und nimmt die Wahl zurück', async () => {
    const { waehlen, hand } = zeige(drucke(3), {
      gewaehlt: ['druck-1'],
    })

    const zweiter = await screen.findByRole('checkbox', { name: /Karte 002/ })
    expect(screen.getByRole('checkbox', { name: /Karte 001/ })).toBeChecked()
    expect(zweiter).not.toBeChecked()

    await hand.click(zweiter)
    expect(waehlen).toHaveBeenLastCalledWith(['druck-1', 'druck-2'])

    await hand.click(screen.getByRole('checkbox', { name: /Karte 001/ }))
    expect(waehlen).toHaveBeenLastCalledWith([])
  })

  it('bietet bis vierzehn Drucke ohne Suche an', async () => {
    zeige(drucke(14))

    expect(await screen.findAllByRole('checkbox')).toHaveLength(14)
    expect(screen.queryByRole('searchbox')).toBeNull()
  })

  // **„Ab 15" heißt fünfzehn** (R7 in `docs/UIREVIEW.md`), nicht sechzehn.
  it('stellt schon bei genau fünfzehn Drucken ein Suchfeld davor', async () => {
    zeige(drucke(15))

    expect(
      await screen.findByRole('searchbox', { name: worte.suchen }),
    ).toBeInTheDocument()
  })

  it('stellt ab fünfzehn ein Suchfeld davor und zeigt zunächst nur die Wahl', async () => {
    zeige(drucke(20), { gewaehlt: ['druck-3'] })

    const suche = await screen.findByRole('searchbox', { name: worte.suchen })

    // **Gezeichnet wird die Wahl und sonst nichts** (R6): wer sucht, sieht
    // Treffer; wer nicht sucht, sieht, was er schon hat.
    expect(screen.getAllByRole('checkbox')).toHaveLength(1)
    expect(screen.getByRole('checkbox', { name: /Karte 003/ })).toBeChecked()
    expect(
      screen.getByText(worte.nurSuche.replace('{anzahl}', '20 Drucke')),
    ).toBeInTheDocument()

    // **Das Suchfeld steht vor der Liste.**
    expect(
      suche.compareDocumentPosition(screen.getByRole('checkbox')) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  })

  it('engt mit der Suche ein und zeigt höchstens zwölf Treffer', async () => {
    const { hand } = zeige(drucke(40))

    const suche = await screen.findByRole('searchbox', { name: worte.suchen })

    await hand.type(suche, 'karte')

    expect(screen.getAllByRole('checkbox')).toHaveLength(12)
    expect(screen.getByText(worte.genauer)).toBeInTheDocument()

    await hand.clear(suche)
    await hand.type(suche, 'Karte 007')

    expect(screen.getAllByRole('checkbox')).toHaveLength(1)
    expect(screen.queryByText(worte.genauer)).toBeNull()
  })

  it('sucht auch nach der Kennung des Drucks', async () => {
    const { hand } = zeige(drucke(20))

    await hand.type(
      await screen.findByRole('searchbox', { name: worte.suchen }),
      'druck-17',
    )

    expect(screen.getAllByRole('checkbox')).toHaveLength(1)
    expect(screen.getByRole('checkbox', { name: /Karte 017/ })).toBeVisible()
  })

  it('sagt, wenn nichts zur Suche passt', async () => {
    const { hand } = zeige(drucke(20))

    await hand.type(
      await screen.findByRole('searchbox', { name: worte.suchen }),
      'zzz',
    )

    expect(screen.queryAllByRole('checkbox')).toHaveLength(0)
    expect(
      screen.getByText(worte.keiner.replace('{suche}', 'zzz')),
    ).toBeInTheDocument()
  })

  it('lässt die Wahl stehen, auch wenn die Suche sie nicht trifft', async () => {
    const { hand } = zeige(drucke(20), { gewaehlt: ['druck-3'] })

    await hand.type(
      await screen.findByRole('searchbox', { name: worte.suchen }),
      'Karte 017',
    )

    // **Zwei Kästchen: der Treffer und die Wahl** — sonst fiele die Wahl aus
    // dem Bild, und niemand könnte sie zurücknehmen.
    expect(screen.getAllByRole('checkbox')).toHaveLength(2)
    expect(screen.getByRole('checkbox', { name: /Karte 003/ })).toBeChecked()
    expect(screen.getByText(/1 Druck gewählt/)).toBeInTheDocument()
  })

  it('erklärt den nächsten Schritt, wenn nichts zu wählen ist', async () => {
    zeige([])

    expect(await screen.findByText(worte.leer)).toBeInTheDocument()
    expect(screen.queryAllByRole('checkbox')).toHaveLength(0)
  })

  it('meldet, wenn die Sammlung nicht zu lesen war', async () => {
    zeige(new Error('klemmt'))

    expect(await screen.findByRole('alert')).toHaveTextContent('klemmt')
  })

  it('nennt den Befund am Feld und beschreibt die Gruppe damit', async () => {
    zeige(drucke(3), { befund: 'Wähle etwas.' })

    const gruppe = await screen.findByRole('group', { name: worte.drucke })
    const befund = within(gruppe).getByRole('alert')

    expect(befund).toHaveTextContent('Wähle etwas.')
    expect(gruppe).toHaveAccessibleDescription(
      expect.stringContaining('Wähle etwas.'),
    )
  })

  it('hält jedes Bedienelement an, solange ein Vorgang läuft', async () => {
    zeige(drucke(20), { angehalten: true })

    expect(
      await screen.findByRole('searchbox', { name: worte.suchen }),
    ).toBeDisabled()
  })

  it('besteht die Prüfung auf Barrierefreiheit', async () => {
    const { container } = zeige(drucke(20), {
      gewaehlt: ['druck-2'],
      befund: 'Wähle etwas.',
    })

    await screen.findByRole('searchbox', { name: worte.suchen })

    expect(await verstoesse(container)).toEqual([])
  })
})
