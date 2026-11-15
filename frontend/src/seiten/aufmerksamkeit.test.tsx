// **Was Aufmerksamkeit braucht** — Werkstück U11 von Phase 5.1 (#1422).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { AufmerksamkeitProbe } from '../test/aufmerksamkeitprobe'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi } from '../test/api'
import type { AufmerksamkeitAntwort, HinweisDTO } from '../apiaufmerksamkeit'
import type { Anschluss } from '../anschluss'
import type { Anschlussstelle } from '../anschlussraum'

// **Deck und Pakete reisen über den Anschluss** — AP-8 (#2365): ein Deck geht
// mit seiner Prüfung auf, die Pakete gefiltert.
const uebergeben: Anschluss[] = []
vi.mock('../anschlussraum', async (echt) => ({
  ...(await echt<typeof import('../anschlussraum')>()),
  useAnschluss: (): Anschlussstelle => ({
    offen: undefined,
    uebergib: (anschluss) => {
      uebergeben.push(anschluss)
    },
    verbrauche: () => undefined,
  }),
}))

const texte = kataloge.de.seiten.start.aufmerksamkeit

const gefuellt: AufmerksamkeitAntwort = {
  hinweise: [
    {
      art: 'setVorbehalte',
      objekt: { art: 'set', kennung: '7', name: 'Probeset', angeheftet: false },
      anzahl: 3,
      grund: 'style, Zeile 12',
      ort: '',
      schluessel: 'vorbehalt',
      werte: ['style', '12'],
    },
    {
      art: 'deckFehlend',
      objekt: {
        art: 'deck',
        kennung: '3',
        name: 'Mono Rot',
        angeheftet: false,
      },
      anzahl: 4,
      grund: '',
      ort: '',
      schluessel: '',
      werte: [],
    },
    {
      art: 'paketBefund',
      objekt: {
        art: 'paket',
        kennung: 'magic-m15',
        name: 'M15',
        angeheftet: false,
      },
      anzahl: 1,
      grund: 'style: Zeile 40',
      ort: '',
      schluessel: '',
      werte: [],
    },
  ],
  durchgesehen: { sets: 2, decks: 1, pakete: 5 },
  befunde: [],
  fehler: null,
}

function hinstellen(
  antwort: AufmerksamkeitAntwort = gefuellt,
  sprache: 'de' | 'en' = 'de',
) {
  const Aufmerksamkeitsruf = vi.fn().mockResolvedValue(antwort)
  belegeApi({ Aufmerksamkeit: Aufmerksamkeitsruf })
  const waehleSet = vi.fn()

  render(
    <Sprachraum sprache={sprache}>
      <AufmerksamkeitProbe waehleSet={waehleSet} stand={0} />
    </Sprachraum>,
  )

  return { Aufmerksamkeitsruf, waehleSet }
}

describe('Aufmerksamkeit', () => {
  it('führt zum Set, zum Deck und zum Paket', async () => {
    const { waehleSet } = hinstellen()
    const bereich = await screen.findByRole('region', { name: texte.titel })

    await userEvent.click(
      within(bereich).getByRole('button', {
        name: /Probeset Set · 3 nicht gedeutet/,
      }),
    )
    expect(waehleSet).toHaveBeenCalledWith('7')

    // **Am Objekt, nicht auf der Seite** — AP-8 (#2365), B-14: das Deck geht
    // mit seiner Prüfung auf, die Pakete stehen gefiltert da.
    await userEvent.click(
      within(bereich).getByRole('button', { name: /Mono Rot Deck · 4 fehlen/ }),
    )
    expect(uebergeben.at(-1)).toEqual({
      art: 'deckansicht',
      name: 'Mono Rot',
      deckID: '3',
      lasche: 'pruefung',
    })

    await userEvent.click(
      within(bereich).getByRole('button', {
        name: /1 eigenes Paket mit Befund im Skript/,
      }),
    )
    expect(uebergeben.at(-1)).toEqual(
      expect.objectContaining({ art: 'paketauswahl', namen: ['magic-m15'] }),
    )

    // Der Grund steht am Hinweis, und die Zählung darunter.
    expect(within(bereich).getByText('style, Zeile 12')).toBeInTheDocument()
    expect(
      within(bereich).getByText(
        '2 Sets, 1 Deck und 5 eigene Pakete durchgesehen.',
      ),
    ).toBeInTheDocument()
  })

  // **Die ersten fünf, der Rest auf Wunsch** — UX-15 (#1771). Gemessen am
  // 2026-09-24 in WebKit mit dem echten Bestand: 536 Pakete mit einem
  // Skriptbefund machten die Startseite 24.827 Punkte hoch, und was darunter
  // stand, sah niemand. Das Artboard *Main* zeigt drei Hinweise und ihre Zahl.
  it('zeigt fünf Hinweise mit ihrer Zahl und die übrigen auf Wunsch', async () => {
    // **Decks und nicht Pakete** — Pakete stehen seit AP-8 als eine Zeile.
    const viele = Array.from({ length: 12 }, (_, stelle) => ({
      art: 'deckFehlend' as const,
      objekt: {
        art: 'deck' as const,
        kennung: `deck-${String(stelle)}`,
        name: `Deck ${String(stelle)}`,
        angeheftet: false,
      },
      anzahl: 1,
      grund: '',
      ort: '',
      schluessel: '',
      werte: [],
    }))
    hinstellen({ ...gefuellt, hinweise: viele })

    const bereich = await screen.findByRole('region', { name: texte.titel })
    const hinweise = () =>
      within(bereich).getAllByRole('button', { name: /^Deck \d+/ })

    expect(hinweise()).toHaveLength(5)
    expect(
      within(bereich).getByRole('heading', { name: `${texte.titel} 12` }),
    ).toBeInTheDocument()

    await userEvent.click(
      within(bereich).getByRole('button', { name: 'Alle 12 zeigen' }),
    )
    expect(hinweise()).toHaveLength(12)

    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weniger }),
    )
    expect(hinweise()).toHaveLength(5)
  })

  // **»eigene« Pakete** (G01, `g01einsteiger`): die mitgelieferten liegen im
  // Bündel und werden nicht durchgesehen — »0 Pakete« neben sieben Paketen
  // unter Ressourcen las sich wie ein Fehler.
  it('sagt im leeren Zustand, was durchgesehen wurde', async () => {
    hinstellen({
      hinweise: [],
      durchgesehen: { sets: 1, decks: 0, pakete: 1 },
      befunde: [],
      fehler: null,
    })

    expect(
      await screen.findByText(
        'Nichts — 1 Set, 0 Decks und 1 eigenes Paket durchgesehen.',
      ),
    ).toBeInTheDocument()
    expect(screen.queryByText(texte.befunde)).not.toBeInTheDocument()
  })

  it('nennt, was nicht durchgesehen werden konnte', async () => {
    hinstellen({ ...gefuellt, befunde: ['Decks: kaputt', 'Sauber: unlesbar'] })

    const bereich = await screen.findByRole('region', { name: texte.titel })
    expect(within(bereich).getByText(texte.befunde)).toBeInTheDocument()
    expect(within(bereich).getByText('Decks: kaputt')).toBeInTheDocument()
    expect(within(bereich).getByText('Sauber: unlesbar')).toBeInTheDocument()
  })

  it('zeigt einen Fehler der Bindung', async () => {
    hinstellen({
      hinweise: [],
      durchgesehen: { sets: 0, decks: 0, pakete: 0 },
      befunde: [],
      fehler: {
        code: 'unbekannt',
        meldung: 'Bestand fehlt',
        feld: '',
        schluessel: '',
      },
    })

    expect(await screen.findByRole('alert')).toHaveTextContent('Bestand fehlt')
  })

  // **Der Grund in der Sprache des Fensters** — LOC
  // (#1963).
  // Gefunden an der Hüllenprobe `leisteenglisch`: unter „Needs attention"
  // stand der deutsche Satz aus Go. Ohne Schlüssel, oder mit einem, den der
  // Katalog nicht kennt, bleibt er stehen (§9).
  it('setzt den Grund in der Sprache des Fensters ein', async () => {
    const paket = (
      kennung: string,
      grund: Pick<HinweisDTO, 'grund' | 'ort' | 'schluessel' | 'werte'>,
    ): HinweisDTO => ({
      art: 'paketBefund',
      objekt: { art: 'paket', kennung, name: kennung, angeheftet: false },
      anzahl: 1,
      ...grund,
    })
    const [setHinweis, deckHinweis] = gefuellt.hinweise as [
      HinweisDTO,
      HinweisDTO,
    ]

    hinstellen(
      {
        ...gefuellt,
        hinweise: [
          setHinweis,
          { ...deckHinweis, grund: 'Lightning Bolt' },
          paket('magic-modules', {
            grund: 'loyalty/init_script: ein einzelnes `=`',
            ort: 'loyalty/init_script',
            schluessel: 'einzelnesGleich',
            werte: [],
          }),
          paket('magic-m15', {
            grund: 'style: diese `(` wird nicht geschlossen — `)` fehlt',
            ort: 'style',
            schluessel: 'klammerOffen',
            werte: ['`(`', '`)`'],
          }),
          paket('magic-fremd', {
            grund: 'card: MSE2 kennt die Feldart „bild“ nicht',
            ort: 'card',
            schluessel: 'feldartUnbekannt',
            werte: ['bild', 'text, choice'],
          }),
          paket('magic-neu', {
            grund: 'extra: ein Satz, den es nur deutsch gibt',
            ort: 'extra',
            schluessel: 'gibtEsHierNicht',
            werte: [],
          }),
        ],
      },
      'en',
    )

    const englisch = kataloge.en.seiten.start.aufmerksamkeit
    const bereich = await screen.findByRole('region', { name: englisch.titel })
    // **Die Gründe der Pakete stehen unter ihrer Sammelzeile** (AP-8, #2365)
    // — zugeklappt, und auf Wunsch weiter in der Sprache des Fensters.
    await userEvent.click(
      within(bereich).getByRole('button', { name: englisch.gruendeZeigen }),
    )

    for (const satz of [
      'style, line 12',
      'Lightning Bolt',
      'loyalty/init_script: a single `=`; assign with `:=`, compare with `==`',
      'style: this `(` is never closed — `)` is missing',
      'card: ' +
        kataloge.en.seiten.vorlagen.feldartUnbekannt
          .replace('{0}', 'bild')
          .replace('{1}', 'text, choice'),
      'extra: ein Satz, den es nur deutsch gibt',
    ]) {
      expect(within(bereich).getByText(satz)).toBeInTheDocument()
    }
  })

  // **Zweimal abgebrochen, ohne vorherigen Stand, zeigt sie nichts** — G11
  // (#2277): kein Ladehinweis, der nie endet, und keine Meldung über einen
  // Abbruch, der die Absicht war (`useLaden` holt genau einmal nach).
  it('steht nach zwei Abbrüchen ohne Stand still da', async () => {
    const abgebrochen = {
      code: 'abgebrochen' as const,
      schluessel: '',
      meldung: 'abgebrochen',
    }
    const ruf = vi.fn(() =>
      Promise.resolve({ ...gefuellt, hinweise: [], fehler: abgebrochen }),
    )
    belegeApi({ Aufmerksamkeit: ruf })

    render(
      <Sprachraum sprache="de">
        <AufmerksamkeitProbe waehleSet={vi.fn()} stand={0} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(ruf).toHaveBeenCalledTimes(2)
    })
    await waitFor(() => {
      expect(screen.queryByRole('status')).toBeNull()
    })
    expect(screen.queryByRole('alert')).toBeNull()
  })
})
