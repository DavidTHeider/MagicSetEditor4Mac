// **Die Wahl eines Exemplars für einen Platz** — Werkstück U9 von Phase 5.1
// (#1420).
//
// **Eine eigene Datei**, damit `deckzuordnung.test.tsx` unter der
// Zeilengrenze aus §2 bleibt. Hier stehen die Fälle der Wahl: zwei Bereiche,
// eine geänderte Menge, ein vorläufiger Druck und die Zeile ohne Druck.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Deckzuordnung } from './Deckzuordnung'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { deckprobe, platzprobe } from '../seiten/deckprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { keineHerkunft } from '../apiherkunft'

import type { DeckDTO, ZuordnungszeileDTO } from '../apidecks'
import type { ExemplarDTO } from '../api'

const texte = kataloge.de.seiten.decks

function exemplar(teil: Partial<ExemplarDTO> = {}): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: 'a1b2c3d4-0000-4000-8000-00000000000a',
    druckID: '4',
    vorlaeufig: false,
    menge: 2,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: 'Blitz',
    setcode: 'm11',
    nummer: '149',
    lagerort: '',
    wunsch: false,
    bezeichnung: 'Blitz · M11 #149',
    herkunft: keineHerkunft,
    ...teil,
  }
}

function zeile(teil: Partial<ZuordnungszeileDTO> = {}): ZuordnungszeileDTO {
  return {
    printing: 'a1b2c3d4-0000-4000-8000-00000000000a',
    druckID: '4',
    vorlaeufig: false,
    bezeichnung: 'Blitz · M11 #149',
    gebraucht: 4,
    vorhanden: 2,
    fehlend: 2,
    grund: '',
    reserviert: 0,
    zugeordnet: 0,
    frei: 2,
    zuordnungen: [],
    ...teil,
  }
}

function zeige(
  deck: DeckDTO,
  zeilen: ZuordnungszeileDTO[],
  exemplare: ExemplarDTO[],
  zuordnen = vi.fn().mockResolvedValue({
    zuordnung: { id: '1', platzID: '3', exemplarID: '1', menge: 1 },
    fehler: null,
  }),
) {
  belegeApi({
    DeckZuordnungen: () =>
      Promise.resolve({ zeilen, fehlend: 2, unbekannt: 0, fehler: null }),
    Sammlung: () => Promise.resolve({ exemplare, fehler: null }),
    DeckZuordnen: zuordnen,
  })

  render(
    <Sprachraum sprache="de">
      <Deckzuordnung
        stand={
          {
            fuehre: async (_s: string, was: () => Promise<void>) => {
              await was()
            },
            vorgang: { laeuft: () => false, starte: vi.fn() },
          } as never
        }
        deck={deck}
      />
    </Sprachraum>,
  )

  return zuordnen
}

afterEach(() => {
  raeumeApi()
})

// kastenAufklappen öffnet die Zeilen von Bedarf und Zuordnung — seit UXR-11
// (#1919)
// stehen sie eingeklappt hinter dem Satz, was fehlt.
async function kastenAufklappen() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.zuordnungen }),
  )
}

// zeileAufklappen öffnet dazu das Formular einer Zeile — es steht erst auf
// Wunsch, und immer nur eines.
async function zeileAufklappen(name: string) {
  await kastenAufklappen()
  await userEvent.click(
    screen.getByRole('button', {
      name: `${texte.zuordnungExemplarWaehlen}: ${name}`,
    }),
  )
}

describe('die Wahl für einen Platz', () => {
  it('lässt den Bereich wählen, wenn derselbe Druck zweimal liegt', async () => {
    const zuordnen = zeige(
      deckprobe({
        plaetze: [
          platzprobe({ id: '3', bereich: 'hauptdeck' }),
          platzprobe({ id: '4', bereich: 'sideboard', menge: 1 }),
        ],
      }),
      [zeile()],
      [exemplar()],
    )
    await zeileAufklappen('Blitz · M11 #149')

    const bereich = await screen.findByLabelText(texte.zuordnungPlatz)
    await userEvent.selectOptions(bereich, '4')

    await userEvent.selectOptions(
      screen.getByLabelText(texte.zuordnungExemplar),
      '1',
    )

    const menge = screen.getByLabelText(texte.zuordnungMenge)
    const knopf = screen.getByRole('button', {
      name: `${texte.zuordnen}: Blitz · M11 #149`,
    })

    // **Ein leeres Mengenfeld ist keine Menge** — F03
    // (#2245).
    // Bis dahin war es eine Eins, und dieselbe Regel machte aus „0" eine Eins
    // und gab „2.5" an ein `int` weiter. Jetzt gilt, was `Zahlenfeld` für ein
    // Formular sagt: wer absendet, meint einen Wert.
    await userEvent.clear(menge)
    expect(knopf).toBeDisabled()

    await userEvent.type(menge, '0')
    expect(menge).toHaveAttribute('aria-invalid', 'true')
    expect(menge).toHaveAccessibleDescription(
      einsetzen(kataloge.de.allgemein.ganzeZahlAb, { kleinstens: '1' }),
    )
    expect(knopf).toBeDisabled()

    await userEvent.clear(menge)
    await userEvent.type(menge, '2.5')
    expect(menge).toHaveAttribute('aria-invalid', 'true')
    expect(knopf).toBeDisabled()
    expect(zuordnen).not.toHaveBeenCalled()

    await userEvent.clear(menge)
    await userEvent.type(menge, '1')
    await userEvent.click(knopf)

    await waitFor(() => {
      expect(zuordnen).toHaveBeenCalledWith('1', '4', '1', 1)
    })

    // **Auf den bedienbaren Knopf gewartet, nicht auf den Ruf** (#1809):
    // nach der Zuordnung lädt die Liste nach, und im Betrieb ruht der Knopf,
    // solange sie läuft.
    await waitFor(() => {
      expect(
        screen.getByRole('button', {
          name: `${texte.zuordnen}: Blitz · M11 #149`,
        }),
      ).toBeEnabled()
    })

    await userEvent.clear(screen.getByLabelText(texte.zuordnungMenge))
    await userEvent.type(screen.getByLabelText(texte.zuordnungMenge), '2')

    await userEvent.click(
      screen.getByRole('button', {
        name: `${texte.zuordnen}: Blitz · M11 #149`,
      }),
    )

    await waitFor(() => {
      expect(zuordnen).toHaveBeenCalledWith('1', '4', '1', 2)
    })
  })

  it('findet den Platz und das Exemplar eines Drucks ohne Alias', async () => {
    const zuordnen = zeige(
      deckprobe({
        plaetze: [platzprobe({ id: '5', printing: '', druckID: '9' })],
      }),
      [
        zeile({
          printing: '',
          druckID: '9',
          vorlaeufig: true,
          bezeichnung: 'Eigenbau',
        }),
      ],
      [
        exemplar({
          id: 2,
          printing: '',
          druckID: '9',
          vorlaeufig: true,
          bezeichnung: 'Eigenbau',
        }),
        // **Ein Wunsch steht nicht zur Wahl** — er ist kein Besitz (0057).
        exemplar({ id: 3, printing: '', druckID: '9', wunsch: true }),
      ],
    )
    await zeileAufklappen('Eigenbau')

    await userEvent.selectOptions(
      await screen.findByLabelText(texte.zuordnungExemplar),
      '2',
    )

    expect(screen.getByLabelText(texte.zuordnungExemplar)).toHaveTextContent(
      'Eigenbau',
    )

    await userEvent.click(
      screen.getByRole('button', { name: `${texte.zuordnen}: Eigenbau` }),
    )

    await waitFor(() => {
      expect(zuordnen).toHaveBeenCalledWith('1', '5', '2', 1)
    })
  })

  it('zeigt für eine Zeile ohne Druck den Grund und keine Wahl', async () => {
    zeige(
      deckprobe({ plaetze: [platzprobe({ karteID: '7', printing: '' })] }),
      [
        zeile({
          printing: '',
          druckID: '',
          bezeichnung: '',
          grund:
            'Diese Karten sind eigene Entwürfe ohne Verweis auf einen Druck.',
        }),
      ],
      [exemplar()],
    )
    await kastenAufklappen()

    expect(
      await screen.findByText(
        'Diese Karten sind eigene Entwürfe ohne Verweis auf einen Druck.',
      ),
    ).toBeInTheDocument()

    expect(
      screen.queryByLabelText(texte.zuordnungExemplar),
    ).not.toBeInTheDocument()
  })
})
