// **Ein Symbol der Bibliothek bearbeiten** — DOM-02.AK-02
// (#1697):
// der Editor öffnet es, sichert eine neue Fassung, nennt die Sets mit der
// vorigen und übernimmt sie auf Wunsch.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type {
  SymbolaenderungAntwort,
  SymbolfolgenAntwort,
} from '../apisymbolbibliothek'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { leererSetkopf } from '../test/setkopfprobe'
import {
  bibliothekAufklappen,
  krone,
  mitSymbolen,
  ohneMarken,
  zacke,
} from '../test/symbolbibliothek'
import { Symbolfach } from './Symbolfach'

const texte = kataloge.de.seiten.karten.symbolbibliothek

afterEach(() => {
  raeumeApi()
})

// rechteck ist eine Zeichnung mit einem Teil.
const rechteck = [
  {
    art: 'Form',
    name: 'Rechteck',
    kombination: 'overlap',
    punkte: [],
    spiegelart: '',
    kopien: 0,
    mitteX: 0,
    mitteY: 0,
    griffX: 0,
    griffY: 0,
    teile: [],
  },
]

// zeigeFach setzt das Fach für das Set 7 auf — mit der Bibliothek darunter,
// aufgeklappt (seit UXR-07, #1915, steht sie eingeklappt).
async function zeigeFach(belegung: Parameters<typeof belegeApi>[0]) {
  const angaben = vi.fn().mockResolvedValue({
    felder: [],
    werte: [],
    kopf: leererSetkopf,
    fehler: null,
  })
  const teile = vi
    .fn()
    .mockResolvedValue({ teile: rechteck, name: 'Krone', fehler: null })
  belegeApi({
    Setangaben: angaben,
    Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
    MarkenJeObjekt: ohneMarken,
    BibliothekssymbolTeile: teile,
    ...belegung,
  })

  render(
    <Sprachraum sprache="de">
      <Symbolfach setID="7" ziel={undefined} />
    </Sprachraum>,
  )
  await bibliothekAufklappen()

  return { teile, angaben }
}

async function bearbeiteKrone() {
  await userEvent.click(
    within(await screen.findByRole('listitem', { name: 'Krone' })).getByRole(
      'button',
      { name: texte.bearbeiten },
    ),
  )
}

describe('Ein Symbol der Bibliothek im Editor', () => {
  it('öffnet es, sichert eine neue Fassung und zieht die Sets nach', async () => {
    const sichern = vi.fn().mockResolvedValue({
      id: '1',
      name: 'Krone',
      veraltet: ['Kronenset'],
      fehler: null,
    })
    const nachziehen = vi
      .fn()
      .mockResolvedValue({ sets: ['Kronenset'], fehler: null })
    const { teile } = await zeigeFach({
      BibliothekssymbolSichern: sichern,
      BibliothekssymbolNachziehen: nachziehen,
    })

    await bearbeiteKrone()

    expect(
      await screen.findByRole('heading', {
        name: einsetzen(texte.bearbeitungsueberschrift, { name: 'Krone' }),
      }),
    ).toBeInTheDocument()
    await waitFor(() => {
      expect(teile).toHaveBeenCalledWith('1')
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.inBibliothekSichern }),
    )

    expect(
      await screen.findByText(einsetzen(texte.veraltet, { sets: 'Kronenset' })),
    ).toBeInTheDocument()
    expect(sichern).toHaveBeenCalledWith('1', rechteck)

    await userEvent.click(
      screen.getByRole('button', { name: texte.nachziehen }),
    )

    expect(
      await screen.findByText(
        einsetzen(texte.nachgezogen, { sets: 'Kronenset' }),
      ),
    ).toBeInTheDocument()
    expect(nachziehen).toHaveBeenCalledWith('1')
  })

  it('sagt, wenn kein Set die vorige Fassung trägt', async () => {
    await zeigeFach({
      BibliothekssymbolSichern: vi.fn().mockResolvedValue({
        id: '1',
        name: 'Krone',
        veraltet: [],
        fehler: null,
      }),
    })

    await bearbeiteKrone()
    await userEvent.click(
      await screen.findByRole('button', { name: texte.inBibliothekSichern }),
    )

    expect(await screen.findByText(texte.keineVeraltet)).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: texte.nachziehen }),
    ).not.toBeInTheDocument()
  })

  it('zeigt einen Fehler beim Nachziehen und kehrt zum Setsymbol zurück', async () => {
    const { angaben } = await zeigeFach({
      BibliothekssymbolSichern: vi.fn().mockResolvedValue({
        id: '1',
        name: 'Krone',
        veraltet: ['Kronenset'],
        fehler: null,
      }),
      BibliothekssymbolNachziehen: vi.fn().mockResolvedValue({
        sets: [],
        fehler: {
          code: 'validierung',
          schluessel: '',
          meldung: 'Der Ordner ist gesperrt.',
          werte: {},
        },
      }),
    })

    await bearbeiteKrone()
    await userEvent.click(
      await screen.findByRole('button', { name: texte.inBibliothekSichern }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: texte.nachziehen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Ordner ist gesperrt.',
    )

    const vorher = angaben.mock.calls.length
    await userEvent.click(
      screen.getByRole('button', { name: texte.zumSetsymbol }),
    )

    await waitFor(() => {
      expect(angaben.mock.calls.length).toBeGreaterThan(vorher)
    })
    expect(
      screen.queryByRole('heading', {
        name: einsetzen(texte.bearbeitungsueberschrift, { name: 'Krone' }),
      }),
    ).not.toBeInTheDocument()
  })

  it('liest das bearbeitete Symbol neu, wenn seine vorige Fassung zurückkommt', async () => {
    const zurueck = vi.fn().mockResolvedValue({
      id: '1',
      name: 'Krone',
      veraltet: [],
      fehler: null,
    })
    const { teile, angaben } = await zeigeFach({
      Symbolbibliothek: vi
        .fn()
        .mockResolvedValue(mitSymbolen({ ...krone, vorige: true })),
      BibliothekssymbolZuruecknehmen: zurueck,
    })

    // **Ohne Bearbeitung** bleibt der Editor des Setsymbols, wie er ist.
    const zeile = await screen.findByRole('listitem', { name: 'Krone' })
    const vorher = angaben.mock.calls.length
    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.zuruecknehmen }),
    )
    await screen.findByRole('status')
    expect(angaben.mock.calls.length).toBe(vorher)

    await bearbeiteKrone()
    await waitFor(() => {
      expect(teile).toHaveBeenCalledTimes(1)
    })

    await userEvent.click(
      within(await screen.findByRole('listitem', { name: 'Krone' })).getByRole(
        'button',
        { name: texte.zuruecknehmen },
      ),
    )

    await waitFor(() => {
      expect(teile).toHaveBeenCalledTimes(2)
    })
    expect(zurueck).toHaveBeenCalledTimes(2)
  })

  // **Eine Antwort gehört zu dem Symbol, für das sie kam** (§7, Regel 2):
  // wer mitten im Sichern oder Nachziehen ein anderes öffnet, sieht deren
  // Ergebnis nicht bei diesem.
  it('trägt eine späte Antwort nicht bei einem anderen Symbol ein', async () => {
    let sichern: (antwort: SymbolaenderungAntwort) => void = () => undefined
    let nachziehen: (antwort: SymbolfolgenAntwort) => void = () => undefined
    await zeigeFach({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone, zacke)),
      BibliothekssymbolSichern: vi.fn(
        () =>
          new Promise<SymbolaenderungAntwort>((erfuellen) => {
            sichern = erfuellen
          }),
      ),
      BibliothekssymbolNachziehen: vi.fn(
        () =>
          new Promise<SymbolfolgenAntwort>((erfuellen) => {
            nachziehen = erfuellen
          }),
      ),
    })
    const oeffne = async (name: string) => {
      await userEvent.click(
        within(await screen.findByRole('listitem', { name })).getByRole(
          'button',
          { name: texte.bearbeiten },
        ),
      )
      await screen.findByRole('heading', {
        name: einsetzen(texte.bearbeitungsueberschrift, { name }),
      })
    }
    const sichere = async () => {
      await userEvent.click(
        await screen.findByRole('button', { name: texte.inBibliothekSichern }),
      )
    }
    const veraltet = (sets: string) => einsetzen(texte.veraltet, { sets })
    const nachgezogen = (sets: string) => einsetzen(texte.nachgezogen, { sets })

    await oeffne('Krone')
    await sichere()
    await oeffne('Zacke')
    sichern({ id: '1', name: 'Krone', veraltet: ['Kronenset'], fehler: null })
    await antwortAbwarten()
    expect(screen.queryByText(veraltet('Kronenset'))).not.toBeInTheDocument()

    // **Der Beleg, dass jemand zuhört**: dasselbe Sichern für Zacke kommt an.
    await sichere()
    sichern({ id: '2', name: 'Zacke', veraltet: ['Zackenset'], fehler: null })
    expect(await screen.findByText(veraltet('Zackenset'))).toBeInTheDocument()

    // **Und das Nachziehen**: Zacke zieht nach, Krone wird geöffnet.
    await userEvent.click(
      screen.getByRole('button', { name: texte.nachziehen }),
    )
    await oeffne('Krone')
    nachziehen({ sets: ['Zackenset'], fehler: null })
    await antwortAbwarten()
    expect(screen.queryByText(nachgezogen('Zackenset'))).not.toBeInTheDocument()

    await sichere()
    sichern({ id: '1', name: 'Krone', veraltet: ['Kronenset'], fehler: null })
    await userEvent.click(
      await screen.findByRole('button', { name: texte.nachziehen }),
    )
    nachziehen({ sets: ['Kronenset'], fehler: null })
    expect(
      await screen.findByText(nachgezogen('Kronenset')),
    ).toBeInTheDocument()

    // **Und ein Fehler**: Krone scheitert beim Nachziehen, Zacke ist offen.
    const gesperrt: SymbolfolgenAntwort = {
      sets: [],
      fehler: {
        code: 'validierung',
        schluessel: '',
        meldung: 'Der Ordner ist gesperrt.',
        werte: {},
      },
    }
    const scheitere = async (id: string, name: string, anderes: string) => {
      await sichere()
      sichern({ id, name, veraltet: ['Kronenset'], fehler: null })
      await userEvent.click(
        await screen.findByRole('button', { name: texte.nachziehen }),
      )
      if (anderes !== name) {
        await oeffne(anderes)
      }
      nachziehen(gesperrt)
    }

    await scheitere('1', 'Krone', 'Zacke')
    await antwortAbwarten()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()

    await scheitere('2', 'Zacke', 'Zacke')
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Ordner ist gesperrt.',
    )
  })
})
