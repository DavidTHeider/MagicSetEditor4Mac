// **Wo eine Vorgabe wirkt** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// **Eine Vorgabe ohne Wirkung wäre eine Zeile im Bestand.** Diese Datei prüft
// die vier Stellen, an denen sie wirkt — und jede davon **auch ohne** Vorgabe,
// weil ein Lauf, der nur den einen Fall kennt, nicht sagt, dass die Vorgabe
// den Unterschied macht (§10).
import { keineHerkunft } from '../apiherkunft'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Deckanlage } from './Deckanlage'
import { Deckliste } from './Deckliste'
import { useDeckhandlung } from '../haken/useDeckhandlung'
import { Sammlungsformular } from './Sammlungsformular'
import { SetNeu } from '../seiten/SetNeu'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { DeckDTO } from '../apidecks'
import type { PaketDTO } from '../pakete'

afterEach(() => {
  raeumeApi()
})

// belegeVorgabe belegt die zwei Rufe, aus denen eine Vorgabe kommt.
function belegeVorgabe(art: string, werte: string[]): Record<string, unknown> {
  return {
    GezeigtesProfil: vi.fn(() => Promise.resolve({ profil: { id: '10' } })),
    Profilvorgaben: vi.fn(() =>
      Promise.resolve({
        vorgaben: [
          { art, mehrwertig: false, werte, moeglich: [], verwaist: [] },
        ],
      }),
    ),
  }
}

// einPaket ist die Vorlage, aus der die Paketfälle ihre Abweichung bauen.
function einPaket(art: string, name: string): PaketDTO {
  return {
    art,
    name,
    anzeigename: name,
    version: '',
    pfad: '',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang: 0,
    artrang: 0,
    herkunftsart: 'mitgeliefert' as const,
    hauptdatei: 'style',
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
  }
}

// einDeck ist die Vorlage, aus der die Deckfälle ihre Abweichung bauen.
function einDeck(id: string, name: string): DeckDTO {
  return {
    id,
    fassung: '1',
    name,
    spiel: 'magic',
    notiz: '',
    formatID: '',
    plaetze: [],
    hauptdeck: 0,
    sideboard: 0,
    herkunft: keineHerkunft,
  }
}

// deckstand ist der kleinste Stand, den Liste und Anlage brauchen.
//
// **Von Hand und nicht über `useDecks`**: was hier geprüft wird, ist die
// Wirkung der Vorgabe — der Haken dahinter hat seine eigenen Läufe.
function deckstand(): Parameters<typeof Deckliste>[0]['stand'] {
  return {
    decks: ['Erstes', 'Zweites', 'Drittes'].map((name, stelle) =>
      einDeck(String(stelle + 1), name),
    ),
    formate: [],
    offen: null,
    geladen: true,
    ladefehler: undefined,
    fehler: undefined,
    konflikt: false,
    nachKonflikt: vi.fn(),
    vorgang: { laeuft: () => false, starte: vi.fn() },
    waehle: vi.fn(),
    liesNach: vi.fn(),
    wahl: () => 0,
    wechsleVon: vi.fn(),
    vergiss: vi.fn(),
    neuLaden: vi.fn(),
    // **Das Doppel fängt, wie der Betrieb fängt** (§3): `useDecks.fuehre`
    // schluckt einen Fehler der Arbeit und meldet ihn — es reicht ihn **nicht**
    // an den Aufrufer weiter. Ein Doppel, das weiterreicht, ist strenger als
    // die Naht, und ein Aufrufer, der danach weiterarbeitet, sähe hier einen
    // Fehler, den es im Betrieb nicht gibt (Befund 4 des Nachlaufs zum
    // Zweitblick über Phase 5.1).
    fuehre: async (_marke: string, arbeit: () => Promise<void>) => {
      try {
        await arbeit()
      } catch {
        // gemeldet wird im Betrieb; hier zählt, dass nichts hinausgeht
      }
    },
    melde: vi.fn(),
  }
}

describe('Das Vorgabespiel eines Profils', () => {
  const worte = kataloge.de.seiten.sets

  const probepakete = [
    einPaket('mse-game', 'magic'),
    einPaket('mse-game', 'space'),
    einPaket('mse-style', 'magic-m15'),
    einPaket('mse-style', 'magic-schlicht'),

    // **Auch `space` hat einen Stil** — ohne ihn gäbe es nur ein nutzbares
    // Spiel, und die Auswahl entfiele (UX-05.AK-04).
    einPaket('mse-style', 'space-weit'),
  ]

  function zeigeSetNeu(teile: Record<string, unknown>) {
    belegeApi({
      Pakete: vi.fn(() =>
        Promise.resolve({ pakete: probepakete, befunde: [], fehler: null }),
      ),
      ...teile,
    })

    render(
      <Sprachraum sprache="de">
        <SetNeu angelegt={vi.fn()} zuDenPaketen={vi.fn()} />
      </Sprachraum>,
    )
  }

  it('ist im Formular für ein neues Set vorgewählt', async () => {
    zeigeSetNeu(belegeVorgabe('spiel', ['space']))

    await waitFor(() => {
      expect(screen.getByDisplayValue('space')).toBeChecked()
    })
  })

  it('und ohne sie steht das erste da', async () => {
    // **Der Beleg, dass die Vorgabe den Unterschied macht** — ohne diesen
    // Fall stünde der Lauf darüber auch dann grün da, wenn `space` ohnehin
    // das erste wäre (§10).
    zeigeSetNeu(belegeVorgabe('spiel', []))

    await waitFor(() => {
      expect(screen.getByDisplayValue('magic')).toBeChecked()
    })
  })

  it('und eine, die es nicht mehr gibt, wird übergangen', async () => {
    // **Sonst stünde das Formular auf einem Wert, den die Liste nicht kennt**
    // — und der Knopf wäre tot, ohne dass jemand sähe, warum.
    zeigeSetNeu(belegeVorgabe('spiel', ['kartenrücken']))

    await waitFor(() => {
      expect(screen.getByDisplayValue('magic')).toBeChecked()
    })
  })

  it('und die Vorgabevorlage ebenso', async () => {
    zeigeSetNeu(belegeVorgabe('vorlage', ['magic-schlicht']))

    // **Der Stil steht im zweiten Schritt** des Assistenten (#1677).
    await userEvent.click(
      await screen.findByRole('button', { name: worte.neuWeiter }),
    )

    // **Vorgewählt in der Auswahl** seit UXR-04 (#1912) — vorher als
    // gesetztes Optionsfeld.
    await waitFor(() => {
      expect(screen.getByRole('combobox', { name: worte.neuStil })).toHaveValue(
        'magic-schlicht',
      )
    })
  })
})

describe('Der zugeordnete Lagerort', () => {
  const eigene = kataloge.de.seiten.sammlung

  function zeigeSammlung(teile: Record<string, unknown>) {
    belegeApi({
      Einstellungen: () => Promise.resolve({ einstellungen: [], fehler: null }),
      ...teile,
    })

    render(
      <Sprachraum sprache="de">
        <Sammlungsformular aufAufnehmen={vi.fn()} laeuft={false} />
      </Sprachraum>,
    )
  }

  it('steht in einer neuen Sammlungszeile', async () => {
    zeigeSammlung(belegeVorgabe('lagerort', ['Schrank']))

    await waitFor(() => {
      expect(screen.getByLabelText(eigene.lagerort)).toHaveValue('Schrank')
    })
  })

  it('und ohne ihn bleibt das Feld leer', async () => {
    zeigeSammlung(belegeVorgabe('lagerort', []))

    await waitFor(() => {
      expect(screen.getByLabelText(eigene.lagerort)).toHaveValue('')
    })
  })
})

// **Die Handlungen stehen seit #2469 über der Liste** — die Seite gibt sie
// ihr mit, und hier tut es dieser Rahmen.
function DecklisteMitHandlung({
  stand,
}: {
  stand: Parameters<typeof Deckliste>[0]['stand']
}) {
  return <Deckliste stand={stand} handlung={useDeckhandlung(stand)} />
}

describe('Die Präferenzdecks', () => {
  function zeigeDecks(werte: string[]) {
    belegeApi(belegeVorgabe('deck', werte))

    render(
      <Sprachraum sprache="de">
        <DecklisteMitHandlung stand={deckstand()} />
      </Sprachraum>,
    )
  }

  it('stehen zuerst — in ihrer eigenen Reihenfolge', async () => {
    zeigeDecks(['3', '1'])

    await waitFor(() => {
      const namen = screen
        .getAllByRole('listitem')
        .map((zeile) => zeile.textContent)

      // **Die Reihenfolge ist die Aussage** — nicht, dass alle drei da sind.
      expect(namen[0]).toContain('Drittes')
      expect(namen[1]).toContain('Erstes')
      expect(namen[2]).toContain('Zweites')
    })
  })

  it('und ohne sie bleibt die Reihenfolge des Bestands', async () => {
    // **Der Beleg, dass die Vorgabe den Unterschied macht** (§10).
    zeigeDecks([])

    await waitFor(() => {
      expect(screen.getAllByRole('listitem')[0]).toHaveTextContent('Erstes')
    })
  })
})

describe('Das Vorgabeformat', () => {
  const eigene = kataloge.de.seiten.decks

  function zeigeAnlage(werte: string[], teile: Record<string, unknown> = {}) {
    belegeApi({
      ...belegeVorgabe('format', werte),
      DeckAnlegen: vi.fn(() =>
        Promise.resolve({ deck: einDeck('7', 'Neu'), fehler: null }),
      ),
      ...teile,
    })

    render(
      <Sprachraum sprache="de">
        <Deckanlage stand={deckstand()} />
      </Sprachraum>,
    )
  }

  async function legeAn() {
    await userEvent.type(await screen.findByLabelText(eigene.name), 'Neu')
    await userEvent.click(screen.getByRole('button', { name: eigene.anlegen }))
  }

  it('wird einem neuen Deck gleich gesetzt', async () => {
    const waehlen = vi.fn(() => Promise.resolve({ deck: null, fehler: null }))

    zeigeAnlage(['5'], { DeckFormatWaehlen: waehlen })

    await legeAn()

    await waitFor(() => {
      expect(waehlen).toHaveBeenCalledWith('7', '5')
    })
  })

  it('und ohne sie bleibt das Deck ohne Format', async () => {
    // **Der Beleg, dass die Vorgabe den zweiten Ruf auslöst** — ohne diesen
    // Fall stünde der Lauf darüber auch dann grün da, wenn er immer käme
    // (§10).
    const waehlen = vi.fn(() => Promise.resolve({ deck: null, fehler: null }))

    zeigeAnlage([], { DeckFormatWaehlen: waehlen })

    await legeAn()

    // **Auf die Wirkung des Anlegens gewartet** — die Verneinung darunter
    // trifft vorher ohnehin zu (§3).
    await waitFor(() => {
      expect(screen.getByLabelText(eigene.name)).toHaveValue('')
    })

    expect(waehlen).not.toHaveBeenCalled()
  })
})

// **Ein gescheitertes Format nimmt das Deck nicht mit** — Befund 4 des
// Nachlaufs zum Zweitblick über Phase 5.1: der Merker fürs Aufschlagen stand
// **hinter** dem zweiten Ruf, und damit hing das Aufschlagen am Format. Ging
// es schief, war das Deck da und nirgends zu sehen — `fuehre` lädt nach einem
// Fehler nicht nach (§7, Regel 3).
describe('Das Vorgabeformat, das nicht gilt', () => {
  const eigene = kataloge.de.seiten.decks

  it('schlägt das Deck trotzdem auf', async () => {
    const waehle = vi.fn(() => Promise.resolve(undefined))

    belegeApi({
      ...belegeVorgabe('format', ['5']),
      DeckAnlegen: vi.fn(() =>
        Promise.resolve({ deck: einDeck('7', 'Neu'), fehler: null }),
      ),
      DeckFormatWaehlen: vi.fn(() =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'validierung' as const,
            schluessel: 'validierung',
            meldung: 'Das Format gehört zu einem anderen Spiel.',
            feld: 'formatID',
          },
        }),
      ),
    })

    render(
      <Sprachraum sprache="de">
        <Deckanlage stand={{ ...deckstand(), waehle }} />
      </Sprachraum>,
    )

    await userEvent.type(await screen.findByLabelText(eigene.name), 'Neu')
    await userEvent.click(screen.getByRole('button', { name: eigene.anlegen }))

    await waitFor(() => {
      expect(waehle).toHaveBeenCalledWith('7')
    })
  })
})
