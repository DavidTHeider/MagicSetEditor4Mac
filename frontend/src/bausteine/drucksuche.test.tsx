// **Ein Druck über seinen Namen** — AP-10
// (#2367),
// Befund B-20, Entwurf `Wege09`: „Nach Name, Set und Nummer — die
// Druckkennung steht nicht mehr im Hauptweg.“ Gesucht wird über dieselbe
// Deutung wie beim Einlesen einer Liste (`ListeDeuten`), auf Anweisung.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type {
  Listendeutungsantwort,
  Listentreffer,
} from '../antwortenlisteneinfuhr'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { Drucksuche } from './Drucksuche'

const texte = kataloge.de.seiten.sammlung.drucksuche

afterEach(raeumeApi)

function treffer(
  printing: string,
  setcode: string,
  nummer: string,
): Listentreffer {
  return {
    printing,
    name: 'Lightning Bolt',
    setcode,
    setname: setcode === 'm11' ? 'Magic 2011' : 'Magic 2010',
    nummer,
    seltenheit: 'common',
    bild: '',
  }
}

function antwort(
  liste: Listentreffer[],
  teil: { weitere?: boolean; befund?: string } = {},
): Listendeutungsantwort {
  return {
    deutung: {
      zeilen: [
        {
          roh: '1 Lightning Bolt',
          zeilennummer: 1,
          menge: 1,
          name: 'Lightning Bolt',
          setcode: '',
          nummer: '',
          finish: 'nonfoil',
          treffer: liste,
          drucke: liste.length,
          eindeutig: liste.length === 1,
          befund: teil.befund ?? '',
          weitere: teil.weitere ?? false,
          fehler: null,
        },
      ],
      befunde: [],
      alterSekunden: 0,
    },
    fehler: null,
  }
}

function zeige(ListeDeuten: (text: string) => Promise<Listendeutungsantwort>) {
  belegeApi({ ListeDeuten: vi.fn(ListeDeuten) })
  const waehle = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Drucksuche waehle={waehle} laeuft={false} />
    </Sprachraum>,
  )

  return waehle
}

describe('die Wahl eines Drucks', () => {
  it('sucht nach dem Namen und lässt unter den Drucken wählen', async () => {
    const deuten = vi.fn<(text: string) => Promise<Listendeutungsantwort>>(() =>
      Promise.resolve(
        antwort([treffer('a', 'm11', '149'), treffer('b', 'm10', '146')]),
      ),
    )
    const waehle = zeige(deuten)

    await userEvent.type(screen.getByLabelText(texte.karte), 'Lightning Bolt')
    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    await userEvent.click(
      await screen.findByRole('radio', {
        name: 'Lightning Bolt · Magic 2010 (M10) · Nr. 146',
      }),
    )
    expect(waehle).toHaveBeenLastCalledWith(treffer('b', 'm10', '146'))
    expect(deuten).toHaveBeenCalledWith('1 Lightning Bolt')
  })

  // **Ein einziger Treffer ist gewählt** — aber nicht still übernommen:
  // aufgenommen wird erst mit „Aufnehmen“ (§16, Regel 6).
  it('wählt einen einzigen Treffer vor, und ↩ sucht', async () => {
    const waehle = zeige(() =>
      Promise.resolve(antwort([treffer('a', 'm11', '149')])),
    )

    await userEvent.type(
      screen.getByLabelText(texte.karte),
      'Lightning Bolt (M11) 149{Enter}',
    )

    expect(
      await screen.findByRole('radio', { checked: true }),
    ).toBeInTheDocument()
    expect(waehle).toHaveBeenLastCalledWith(treffer('a', 'm11', '149'))
  })

  it('sagt am Feld, wenn nichts gefunden wurde, und wenn es mehr gibt', async () => {
    zeige(() => Promise.resolve(antwort([])))
    await userEvent.type(
      screen.getByLabelText(texte.karte),
      'Gibtesnicht{Enter}',
    )
    expect(await screen.findByText(texte.keinTreffer)).toBeInTheDocument()
    expect(screen.getByLabelText(texte.karte)).toHaveAttribute(
      'aria-invalid',
      'true',
    )
  })

  it('nennt, dass es weitere Drucke gibt', async () => {
    zeige(() =>
      Promise.resolve(antwort([treffer('a', 'm11', '149')], { weitere: true })),
    )
    await userEvent.type(screen.getByLabelText(texte.karte), 'Bolt{Enter}')
    expect(await screen.findByText(texte.weitere)).toBeInTheDocument()
  })

  it('zeigt einen Fehler der Suche am Feld', async () => {
    zeige(() =>
      Promise.resolve({
        deutung: null,
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'kaputt',
          feld: '',
          werte: {},
        },
      }),
    )
    await userEvent.type(screen.getByLabelText(texte.karte), 'Bolt{Enter}')
    expect(
      await screen.findByText(kataloge.de.fehler.unbekannt),
    ).toBeInTheDocument()
  })

  it('sucht ohne Text nicht', async () => {
    const deuten = vi.fn(() => Promise.resolve(antwort([])))
    zeige(deuten)
    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))
    expect(deuten).not.toHaveBeenCalled()
  })

  it('sagt „nichts gefunden“ auch für eine leere Deutung', async () => {
    zeige(() =>
      Promise.resolve({
        deutung: { zeilen: [], befunde: [], alterSekunden: 0 },
        fehler: null,
      }),
    )
    await userEvent.type(screen.getByLabelText(texte.karte), 'Bolt{Enter}')
    expect(await screen.findByText(texte.keinTreffer)).toBeInTheDocument()
  })

  it('sagt „nichts gefunden“ ohne Deutung', async () => {
    zeige(() => Promise.resolve({ deutung: null, fehler: null }))
    await userEvent.type(screen.getByLabelText(texte.karte), 'Bolt{Enter}')
    expect(await screen.findByText(texte.keinTreffer)).toBeInTheDocument()
  })

  // **Eine alte Antwort übermalt keine neue** (§7, Regel 2): die erste
  // Suche antwortet erst, nachdem die zweite schon dasteht.
  it('lässt eine späte Antwort einer alten Suche fallen', async () => {
    let erste: (wert: Listendeutungsantwort) => void = () => undefined
    const deuten = vi
      .fn<(text: string) => Promise<Listendeutungsantwort>>()
      .mockImplementationOnce(
        () =>
          new Promise((fertig) => {
            erste = fertig
          }),
      )
      .mockImplementationOnce(() =>
        Promise.resolve(antwort([treffer('b', 'm10', '146')])),
      )
    zeige(deuten)
    const feld = screen.getByLabelText(texte.karte)

    await userEvent.type(feld, 'Alt{Enter}')
    await userEvent.clear(feld)
    await userEvent.type(feld, 'Neu{Enter}')
    expect(
      await screen.findByRole('radio', {
        name: 'Lightning Bolt · Magic 2010 (M10) · Nr. 146',
      }),
    ).toBeInTheDocument()

    erste(antwort([treffer('a', 'm11', '149')]))
    await effekteNachholen()

    expect(screen.getAllByRole('radio')).toHaveLength(1)
    expect(
      screen.getByRole('radio', {
        name: 'Lightning Bolt · Magic 2010 (M10) · Nr. 146',
      }),
    ).toBeInTheDocument()
  })

  // **Eine neue Suche hebt die Wahl sofort auf** (Zweitblick 6b, Befund 1):
  // solange sie läuft, gilt kein alter Druck mehr, und ohne Treffer bleibt
  // es dabei.
  it('meldet eine aufgehobene Wahl, sobald neu gesucht wird', async () => {
    let zweite: (wert: Listendeutungsantwort) => void = () => undefined
    const deuten = vi
      .fn<(text: string) => Promise<Listendeutungsantwort>>()
      .mockImplementationOnce(() =>
        Promise.resolve(antwort([treffer('a', 'm11', '149')])),
      )
      .mockImplementationOnce(
        () =>
          new Promise((fertig) => {
            zweite = fertig
          }),
      )
    const waehle = zeige(deuten)
    const feld = screen.getByLabelText(texte.karte)

    await userEvent.type(feld, 'Bolt{Enter}')
    expect(await screen.findByRole('radio', { checked: true })).toBeVisible()
    expect(waehle).toHaveBeenLastCalledWith(treffer('a', 'm11', '149'))

    await userEvent.clear(feld)
    await userEvent.type(feld, 'Gibtesnicht{Enter}')
    expect(waehle).toHaveBeenLastCalledWith(undefined)
    expect(screen.queryByRole('radio')).toBeNull()

    zweite(antwort([]))
    expect(await screen.findByText(texte.keinTreffer)).toBeInTheDocument()
    expect(waehle).toHaveBeenLastCalledWith(undefined)
  })

  // **Weitertippen verwirft die laufende Suche** (Zweitblick 6b, Befund 1):
  // ihre Antwort gehört zu einem Text, der nicht mehr im Feld steht.
  it('lässt die Antwort fallen, wenn während der Suche weitergetippt wird', async () => {
    let erste: (wert: Listendeutungsantwort) => void = () => undefined
    const deuten = vi
      .fn<(text: string) => Promise<Listendeutungsantwort>>()
      .mockImplementationOnce(
        () =>
          new Promise((fertig) => {
            erste = fertig
          }),
      )
    const waehle = zeige(deuten)
    const feld = screen.getByLabelText(texte.karte)

    await userEvent.type(feld, 'Bolt{Enter}')
    await userEvent.type(feld, 's')

    // **Ein Beleg, dass jemand zuhört**: der Knopf sucht wieder.
    expect(screen.getByRole('button', { name: texte.suchen })).toBeEnabled()

    erste(antwort([treffer('a', 'm11', '149')]))
    await effekteNachholen()

    expect(screen.queryByRole('radio')).toBeNull()
    expect(waehle).not.toHaveBeenCalledWith(treffer('a', 'm11', '149'))
  })

  // **Während das Formular aufnimmt, sucht auch ↩ nicht** — Nachlauf zum
  // Zweitblick über Phase 6, zweiter Teil, R1: der Knopf war angehalten,
  // der Weg über das Feld nicht.
  it('sucht nicht, solange das Formular aufnimmt', async () => {
    const deuten = vi.fn(() => Promise.resolve(antwort([])))
    belegeApi({ ListeDeuten: deuten })
    const zeichne = (laeuft: boolean) => (
      <Sprachraum sprache="de">
        <Drucksuche waehle={vi.fn()} laeuft={laeuft} />
      </Sprachraum>
    )
    const { rerender } = render(zeichne(true))
    const feld = screen.getByLabelText(texte.karte)

    await userEvent.type(feld, 'Bolt{Enter}')
    expect(deuten).not.toHaveBeenCalled()

    // **Ein Beleg, dass jemand zuhört**: derselbe Weg sucht danach.
    rerender(zeichne(false))
    await userEvent.type(feld, '{Enter}')
    expect(deuten).toHaveBeenCalledWith('1 Bolt')
  })
})
