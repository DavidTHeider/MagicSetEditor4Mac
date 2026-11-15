// Die Stilauswahl eines Sets (#41).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { type Api } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { waehleZeile } from '../test/suchwahl'
import { Stilwahl } from './Stilwahl'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

// stile gibt zwei Angebote: den eigenen und einen fremden.
function stile() {
  return {
    stile: [
      {
        paket: 'probe-schlicht',
        anzeigename: 'Probe-schlicht',
        spiel: 'probe',
        version: '2026-08-15',
        eigenesSpiel: true,
        gewaehlt: true,
      },
      {
        paket: 'fremd',
        anzeigename: 'Fremd',
        spiel: 'anderes',
        version: '2026-01-01',
        eigenesSpiel: false,
        gewaehlt: false,
      },
    ],
  }
}

function belegeStil(teile: Partial<Api>) {
  return belegeApi({
    Stile: () => Promise.resolve(stile()),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),

    // **Die Vorschau der Stilseite lädt mit** (#643). Ohne diese beiden
    // meldete sie ihren eigenen Befund, und jede Prüfung auf `alert` fände
    // zwei.
    Karten: () => Promise.resolve({ karten: [] }),
    ...teile,
  })
}

// waehle nimmt einen Stil aus der Auswahl.
//
// **Über die Auswahl und nicht über ein Optionsfeld** — UXR-07 (#1915): bis
// dahin stand je Stil ein Optionsfeld da, und dieser Lauf klickte es über
// seinen Wert an. Jetzt ist es **ein** Auswahlfeld.
async function waehle(zeile: string) {
  await waehleZeile(
    await screen.findByRole('combobox', {
      name: texte.seiten.karten.stilWaehlen,
    }),
    zeile,
  )
}

// fremd ist die Zeile des Stils eines anderen Spiels.
const fremd = 'Fremd fremd · anderes'
function zeige(geaendert = vi.fn()) {
  const stand = render(
    <Sprachraum sprache="de">
      <Stilwahl setID="1" geaendert={geaendert} />
    </Sprachraum>,
  )

  return { geaendert, stand }
}

describe('Stilauswahl', () => {
  it('zeigt die installierten Stile und nennt beim fremden sein Spiel', async () => {
    belegeStil({})
    zeige()

    // **Jeder mit seinem Namen und seiner Kennung** — UX-05, S07.NAT-02
    // (#1677); seit UXR-07 als Eintrag einer Auswahl (#1915), seit UXR-F1
    // (#2007) einer mit Suche: die Zeilen stehen da, sobald sie offen ist.
    await userEvent.click(
      await screen.findByRole('combobox', {
        name: texte.seiten.karten.stilWaehlen,
      }),
    )

    expect(
      await screen.findByRole('option', { name: /^Probe-schlicht probe-/ }),
    ).toBeInTheDocument()

    // **Der fremde wird gezeigt, nicht weggefiltert** — mit seinem Spiel
    // hinter der Kennung, damit niemand ihn für einen eigenen hält.
    expect(screen.getByRole('option', { name: fremd })).toBeInTheDocument()
  })

  it('erklärt den leeren Zustand', async () => {
    belegeStil({ Stile: () => Promise.resolve({ stile: [] }) })
    zeige()

    expect(
      await screen.findByText(texte.seiten.karten.stilLeer),
    ).toBeInTheDocument()
  })

  // **Die Folgen stehen vor dem Wechsel** — die Zusage aus #41.
  it('fragt erst nach den Folgen und wechselt erst nach der Zustimmung', async () => {
    const Stilfolgen = vi.fn(() =>
      Promise.resolve({
        folgen: {
          ungedeuteteFelder: ['flavor'],
          verloreneOptionen: ['hintergrund'],
          spielwechsel: true,
        },
      }),
    )
    const StilWechseln = vi.fn(() => Promise.resolve({}))

    belegeStil({ Stilfolgen, StilWechseln })
    const { geaendert } = zeige()

    await waehle(fremd)

    expect(Stilfolgen).toHaveBeenCalledWith('1', 'fremd')
    expect(await screen.findByText('flavor')).toBeInTheDocument()
    expect(screen.getByText('hintergrund')).toBeInTheDocument()
    expect(
      screen.getByText(texte.seiten.karten.stilSpielwechsel),
    ).toBeInTheDocument()

    // **Noch ist nichts geschehen.**
    expect(StilWechseln).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.stilUebernehmen }),
    )

    expect(StilWechseln).toHaveBeenCalledWith('1', 'fremd')
    await waitFor(() => {
      expect(geaendert).toHaveBeenCalled()
    })
  })

  // **Eine führende Handlung je Region** — R1 aus `docs/UIREVIEW.md` und das
  // Artboard *Eigenschaften*: in der Rückfrage ist es das Übernehmen, das
  // Behalten steht daneben im Nebengewicht (UXR-07, #1915).
  it('führt die Rückfrage mit dem Übernehmen', async () => {
    belegeStil({
      Stilfolgen: () =>
        Promise.resolve({
          folgen: {
            ungedeuteteFelder: [],
            verloreneOptionen: [],
            spielwechsel: false,
          },
        }),
    })
    zeige()

    await waehle(fremd)

    expect(
      await screen.findByRole('button', {
        name: texte.seiten.karten.stilUebernehmen,
      }),
    ).toHaveClass('bg-marke')
    expect(
      screen.getByRole('button', { name: texte.seiten.karten.stilAbbrechen }),
    ).not.toHaveClass('bg-marke')
  })

  it('wechselt nicht, wenn jemand die Rückfrage verneint', async () => {
    const StilWechseln = vi.fn(() => Promise.resolve({}))

    belegeStil({
      Stilfolgen: () =>
        Promise.resolve({
          folgen: {
            ungedeuteteFelder: [],
            verloreneOptionen: [],
            spielwechsel: false,
          },
        }),
      StilWechseln,
    })
    zeige()

    await waehle(fremd)

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.karten.stilAbbrechen,
      }),
    )

    expect(StilWechseln).not.toHaveBeenCalled()
    expect(
      screen.queryByText(texte.seiten.karten.stilFrage),
    ).not.toBeInTheDocument()
  })

  // **Kein Stil ist gewählt** — ein Set, dessen `stylesheet:` auf ein Paket
  // zeigt, das nicht installiert ist. Die Auswahl steht dann auf nichts, statt
  // stillschweigend den ersten zu behaupten.
  it('steht auf nichts, wenn kein Angebot gewählt ist', async () => {
    belegeStil({
      Stile: () =>
        Promise.resolve({
          stile: [
            {
              paket: 'probe-schlicht',
              anzeigename: 'Probe-schlicht',
              spiel: 'probe',
              version: '',
              eigenesSpiel: true,
              gewaehlt: false,
            },
          ],
        }),
    })
    zeige()

    // **Nichts ist gewählt, und der Satz darüber sagt, warum** — früher ein
    // leerer Eintrag in der Auswahlliste.
    expect(
      await screen.findByRole('combobox', {
        name: texte.seiten.karten.stilWaehlen,
      }),
    ).toHaveValue('')
    expect(screen.getByText(texte.seiten.karten.stilKeiner)).toBeInTheDocument()
  })

  // **Eine Antwort ohne Folgen und ohne Fehler zeigt keine Rückfrage.**
  //
  // Sie sollte nicht vorkommen; käme sie doch, wäre eine Rückfrage über nichts
  // die schlechtere Antwort.
  it('zeigt keine Rückfrage, wenn die Folgen fehlen', async () => {
    belegeStil({ Stilfolgen: () => Promise.resolve({ folgen: null }) })
    zeige()

    await waehle(fremd)

    await waitFor(() => {
      expect(
        screen.queryByText(texte.seiten.karten.stilFrage),
      ).not.toBeInTheDocument()
    })
  })

  // **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext** (§7, Regel 1).
  //
  // Wer während einer offenen Rückfrage das Set wechselt, hätte sonst eine
  // Zustimmung über ein Set, das er gar nicht mehr ansieht.
  it('räumt die Rückfrage, wenn das Set wechselt', async () => {
    belegeStil({
      Stilfolgen: () =>
        Promise.resolve({
          folgen: {
            ungedeuteteFelder: [],
            verloreneOptionen: [],
            spielwechsel: false,
          },
        }),
    })
    const { stand } = zeige()

    await waehle(fremd)

    await screen.findByText(texte.seiten.karten.stilFrage)

    stand.rerender(
      <Sprachraum sprache="de">
        <Stilwahl setID="2" geaendert={vi.fn()} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(
        screen.queryByText(texte.seiten.karten.stilFrage),
      ).not.toBeInTheDocument()
    })
  })

  // **Die Eingabe in der Auswahl gehört zu dem Set, in dem sie entstand** —
  // §7, Regel 1. Seit UXR-F1 (#2007) hält die `Suchauswahl` den Entwurf
  // selbst; mit den `daten` des alten Sets geht auch sie.
  it('räumt die Eingabe und die offene Liste, wenn das Set wechselt', async () => {
    belegeStil({})
    const { stand } = zeige()

    const feld = await screen.findByRole('combobox', {
      name: texte.seiten.karten.stilWaehlen,
    })

    await userEvent.click(feld)
    await userEvent.keyboard('fremd')

    expect(feld).toHaveValue('fremd')
    expect(screen.getByRole('listbox')).toBeInTheDocument()

    stand.rerender(
      <Sprachraum sprache="de">
        <Stilwahl setID="2" geaendert={vi.fn()} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    })
    expect(
      await screen.findByRole('combobox', {
        name: texte.seiten.karten.stilWaehlen,
      }),
    ).toHaveValue('Probe-schlicht')
  })

  it('meldet einen Fehler beim Laden', async () => {
    belegeStil({
      Stile: () =>
        Promise.resolve({
          stile: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'fehler.unbekannt',
            meldung: 'Der Bestand schweigt.',
          },
        }),
    })
    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Bestand schweigt.',
    )
  })

  it('meldet einen Fehler beim Fragen und beim Wechseln', async () => {
    belegeStil({
      Stilfolgen: () =>
        Promise.resolve({
          folgen: null,
          fehler: {
            code: 'nichtGefunden' as const,
            schluessel: 'fehler.nichtGefunden',
            meldung: 'Diesen Stil gibt es nicht.',
          },
        }),
    })
    zeige()

    await waehle(fremd)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Diesen Stil gibt es nicht.',
    )
  })

  it('meldet einen Fehler beim Übernehmen', async () => {
    belegeStil({
      Stilfolgen: () =>
        Promise.resolve({
          folgen: {
            ungedeuteteFelder: [],
            verloreneOptionen: [],
            spielwechsel: false,
          },
        }),
      StilWechseln: () =>
        Promise.resolve({
          fehler: {
            code: 'konflikt' as const,
            schluessel: 'fehler.konflikt',
            meldung: 'Das ging schief.',
          },
        }),
    })
    zeige()

    await waehle(fremd)

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.karten.stilUebernehmen,
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Das ging schief.',
    )
  })
})
