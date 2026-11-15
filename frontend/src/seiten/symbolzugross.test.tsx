// **Ein Symbol, das zu groß ist — je Lage der Satz, der passt** —
// #1973.
//
// # Zwei Lagen, zwei Sätze
//
// **Eine gewählte Datei** über der Grenze ist keine Symboldatei, und der Rat
// ist eine andere Datei (`symbolZuGross`). **Ein Symbol, das gespeichert zu
// groß würde** — abgelegt, nachgezeichnet, in der Bibliothek gesichert —, ist
// eine Zeichnung, und der Rat ist, sie zu vereinfachen
// (`symbolGespeichertZuGross`). Bis #1973 bekamen alle den Satz der Datei,
// seit QA-06b (#1967) auch die Zeichnung, die beim Ablegen zu groß wurde.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { leererSetkopf } from '../test/setkopfprobe'
import {
  bibliothekAufklappen,
  krone,
  mitSymbolen,
  ohneMarken,
  symbolantwort,
  zeigeBibliothek,
} from '../test/symbolbibliothek'
import { Symbol } from './Symbol'
import { Symbolfach } from './Symbolfach'

afterEach(() => {
  raeumeApi()
})

const symboltexte = kataloge.de.seiten.karten.symbol
const bibliothekstexte = kataloge.de.seiten.karten.symbolbibliothek
const grenze = { grenze: '1 MiB' }

// zuGross ist die Absage des Dienstes in einer der beiden Lagen.
function zuGross(schluessel: string) {
  return {
    code: 'validierung',
    schluessel,
    meldung: 'Das Symbol ist zu groß.',
    feld: 'symbol',
    werte: grenze,
  }
}

// eineForm ist eine Zeichnung mit einem Teil, wie das Nachzeichnen sie gibt.
function eineForm() {
  return {
    teile: [
      {
        art: 'Form',
        name: 'Form 1',
        kombination: 'merge',
        punkte: [],
        spiegelart: '',
        kopien: 0,
        mitteX: 0,
        mitteY: 0,
        griffX: 0,
        griffY: 0,
        teile: [],
      },
    ],
    name: '',
    fehler: null,
  }
}

// zeigeSymbol setzt den Symboleditor für das Set 7 auf.
function zeigeSymbol(belegung: Parameters<typeof belegeApi>[0]) {
  belegeApi({
    Setangaben: vi.fn().mockResolvedValue({
      felder: [],
      werte: [],
      kopf: leererSetkopf,
      fehler: null,
    }),
    Grundform: vi.fn().mockResolvedValue(eineForm()),
    ...belegung,
  })

  render(
    <Sprachraum sprache="de">
      <Symbol setID="7" />
    </Sprachraum>,
  )
}

// zeichnungGenannt hält, dass die Meldung die Zeichnung nennt — und nicht
// zu einer anderen Datei rät.
function zeichnungGenannt(meldung: HTMLElement) {
  expect(meldung).toHaveTextContent(
    einsetzen(kataloge.de.fehler.symbolGespeichertZuGross, grenze),
  )
  expect(meldung).not.toHaveTextContent(
    einsetzen(kataloge.de.fehler.symbolZuGross, grenze),
  )
}

describe('Ein Symbol über der Grenze', () => {
  it('nennt bei einer gewählten Datei die Datei', async () => {
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
      MarkenJeObjekt: ohneMarken,
      BibliothekssymbolEinlesen: vi.fn().mockResolvedValue({
        ...symbolantwort({}),
        fehler: { ...zuGross('symbolZuGross'), feld: 'pfad' },
      }),
    })

    zeigeBibliothek()
    await screen.findByRole('listitem', { name: 'Krone' })
    await userEvent.click(
      screen.getByRole('button', { name: bibliothekstexte.einlesen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      einsetzen(kataloge.de.fehler.symbolZuGross, grenze),
    )
  })

  it('nennt beim Ablegen die Zeichnung', async () => {
    zeigeSymbol({
      SymbolSetzen: vi.fn().mockResolvedValue({
        teile: [],
        name: '',
        fehler: zuGross('symbolGespeichertZuGross'),
      }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: symboltexte.dreieck }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: symboltexte.sichern }),
    )

    zeichnungGenannt(await screen.findByRole('alert'))
  })

  it('nennt nach dem Nachzeichnen die Zeichnung', async () => {
    zeigeSymbol({
      SymbolNachzeichnen: vi.fn().mockResolvedValue(eineForm()),
      SymbolSetzen: vi.fn().mockResolvedValue({
        teile: [],
        name: '',
        fehler: zuGross('symbolGespeichertZuGross'),
      }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: symboltexte.nachzeichnen }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: symboltexte.sichern }),
    )

    zeichnungGenannt(await screen.findByRole('alert'))
  })

  it('nennt beim Sichern in der Bibliothek die Zeichnung', async () => {
    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
      MarkenJeObjekt: ohneMarken,
      BibliothekssymbolTeile: vi
        .fn()
        .mockResolvedValue({ ...eineForm(), name: 'Krone' }),
      BibliothekssymbolSichern: vi.fn().mockResolvedValue({
        id: '1',
        name: 'Krone',
        veraltet: [],
        fehler: zuGross('symbolGespeichertZuGross'),
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Symbolfach setID="7" ziel={undefined} />
      </Sprachraum>,
    )
    // **Die Bibliothek steht eingeklappt** — seit UXR-07 (#1915).
    await bibliothekAufklappen()

    await userEvent.click(
      within(await screen.findByRole('listitem', { name: 'Krone' })).getByRole(
        'button',
        { name: bibliothekstexte.bearbeiten },
      ),
    )
    await userEvent.click(
      await screen.findByRole('button', {
        name: bibliothekstexte.inBibliothekSichern,
      }),
    )

    zeichnungGenannt(await screen.findByRole('alert'))
  })
})
