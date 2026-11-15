// **Der Verweis einer Karte auf einen echten Druck** — Werkstück 2 von
// Phase 3.6
// (#1167).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KartenverweisDTO } from '../apiscryfall'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { oeffneAnlegen, oeffneBezug } from '../test/kartenlasche'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'
import { Kartenverweis } from './Kartenverweis'

const texte = kataloge.de.seiten.karten.verweis

function verweis(teile: Partial<KartenverweisDTO> = {}): KartenverweisDTO {
  return {
    karteID: '7',
    printing: '516e137a-d9d9-4ba7-89fb-e680ddb2c2d1',
    druckID: '',
    vorlaeufig: false,
    name: 'Lightning Bolt',
    setcode: 'lea',
    nummer: '161',
    adresse: 'https://scryfall.com/card/lea/161',
    gesetzt: '2026-08-27T12:00:00Z',
    ...teile,
  }
}

const treffer = {
  kennung: 'e3285e6b-3e79-4d7c-bf96-d920f973b0d5',
  name: 'Lightning Bolt',
  setcode: '2ed',
  setname: 'Unlimited Edition',
  nummer: '161',
  seltenheit: 'common',
  manakosten: '{R}',
  typzeile: 'Instant',
  bild: '',
}

function zeige(vorschlag?: typeof treffer, aufFehler = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Kartenverweis
        karteID="7"
        vorschlag={vorschlag}
        vorgang={{ starte: (_, was) => was(), laeuft: () => false }}
        aufFehler={aufFehler}
      />
    </Sprachraum>,
  )

  return aufFehler
}

afterEach(() => {
  raeumeApi()
})

describe('der Verweis auf einen echten Druck', () => {
  // **Ohne Verweis erklärt der leere Zustand die Folge** (§6): der Abgleich
  // nimmt dann irgendeinen Druck.
  it('sagt, wenn die Karte auf keinen Druck bezogen ist', async () => {
    belegeApi({
      KarteVerweis: vi.fn().mockResolvedValue({
        verweis: null,
        befunde: [],
        fehler: null,
      }),
    })

    zeige()

    expect(await screen.findByText(texte.keiner)).toBeInTheDocument()
  })

  // **Mit Verweis stehen Name, Druck und Zeitpunkt da** — und der Name führt
  // zur Kartenseite.
  it('zeigt den bezogenen Druck samt Adresse und Zeitpunkt', async () => {
    belegeApi({
      KarteVerweis: vi.fn().mockResolvedValue({
        verweis: verweis(),
        befunde: [],
        fehler: null,
      }),
    })

    zeige()

    const link = await screen.findByRole('link', { name: 'Lightning Bolt' })

    expect(link).toHaveAttribute('href', 'https://scryfall.com/card/lea/161')
    expect(screen.getByText(/aus LEA 161/)).toBeInTheDocument()

    // **Das Alter steht dabei** (§16, Regel 5): Name und Nummer stammen aus
    // der Antwort von damals, nicht von jetzt.
    expect(screen.getByText(/bezogen am/)).toBeInTheDocument()
  })

  // **Ein angeklickter Treffer lässt sich beziehen** — mit seiner Kennung,
  // nicht mit seinem Namen.
  it('bezieht die Karte auf die Kennung des angeklickten Treffers', async () => {
    const gesetzt = vi.fn().mockResolvedValue({
      verweis: verweis({ printing: treffer.kennung, setcode: '2ed' }),
      befunde: [],
      fehler: null,
    })

    belegeApi({
      KarteVerweis: vi
        .fn()
        .mockResolvedValue({ verweis: null, befunde: [], fehler: null }),
      KarteVerweisSetzen: gesetzt,
    })

    zeige(treffer)

    await userEvent.click(
      await screen.findByRole('button', { name: /Lightning Bolt/ }),
    )

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledWith('7', treffer.kennung)
    })
  })

  // **Ohne angeklickten Treffer gibt es nichts zu beziehen.** Ein Knopf ohne
  // Gegenstand wäre eine Frage ohne Antwort (§6).
  it('bietet das Beziehen erst an, wenn ein Treffer gewählt ist', async () => {
    belegeApi({
      KarteVerweis: vi
        .fn()
        .mockResolvedValue({ verweis: null, befunde: [], fehler: null }),
    })

    zeige()

    expect(await screen.findByText(texte.keiner)).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: /beziehen/ }),
    ).not.toBeInTheDocument()
  })

  // **Und lösen geht nur, wo einer ist.**
  it('löst den Bezug und lädt danach nach', async () => {
    const geloest = vi.fn().mockResolvedValue({ fehler: null })
    const gelesen = vi
      .fn()
      .mockResolvedValueOnce({ verweis: verweis(), befunde: [], fehler: null })
      .mockResolvedValue({ verweis: null, befunde: [], fehler: null })

    belegeApi({ KarteVerweis: gelesen, KarteVerweisLoesen: geloest })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.loesen }),
    )

    await waitFor(() => {
      expect(geloest).toHaveBeenCalledWith('7')
    })

    // **Regel 3**: geschrieben ist geschrieben — nachgeladen wird danach, und
    // der leere Zustand steht wieder da.
    expect(await screen.findByText(texte.keiner)).toBeInTheDocument()
  })

  // **Ein Schreibfehler geht nach oben**, weil der Kasten den Kartenwechsel
  // nicht überlebt (Befund 3 des Nachlaufs über Phase 3.5).
  it('meldet einen Schreibfehler nach oben', async () => {
    belegeApi({
      KarteVerweis: vi
        .fn()
        .mockResolvedValue({ verweis: null, befunde: [], fehler: null }),
      KarteVerweisSetzen: vi.fn().mockResolvedValue({
        verweis: null,
        befunde: [],
        fehler: {
          code: 'nichtGefunden',
          schluessel: 'nichtGefunden',
          meldung: 'Scryfall kennt diesen Druck nicht',
        },
      }),
    })

    const aufFehler = zeige(treffer)

    await userEvent.click(
      await screen.findByRole('button', { name: /Lightning Bolt/ }),
    )

    // **Der Satz kommt aus dem Katalog, nicht aus Go** (§9): die Go-Seite
    // schickt den Schlüssel daneben, und die Oberfläche spricht die Sprache
    // des Fensters.
    await waitFor(() => {
      expect(aufFehler).toHaveBeenCalledWith(kataloge.de.fehler.nichtGefunden)
    })
  })

  // **Und ein Lesefehler steht am Kasten**, weil ihn niemand anders zeigt.
  it('zeigt einen Lesefehler am Kasten', async () => {
    belegeApi({
      KarteVerweis: vi.fn().mockResolvedValue({
        verweis: null,
        befunde: [],
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'Der Bestand ist belegt',
        },
      }),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )
  })
})

// **Und der Weg nach oben trägt bis in die Handlungszeile** — die Meldung
// muss den Kartenwechsel überleben, und die Zeile überlebt ihn.
describe('ein Verweisfehler auf der Kartenseite', () => {
  it('steht in der Handlungszeile und nicht im Kasten', async () => {
    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Kartenzeilen: () =>
        Promise.resolve({
          zeilen: [{ id: '5', nummer: 0, name: 'Feuerball', felder: 1 }],
        }),
      Karte: (karteID: string) =>
        Promise.resolve({
          karte: { id: karteID, nummer: 0, werte: [] },
          fehler: null,
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
      KarteVerweis: () =>
        Promise.resolve({ verweis: verweis(), befunde: [], fehler: null }),
      KarteVerweisLoesen: () =>
        Promise.resolve({
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'der Bestand ist belegt',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID="5"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await oeffneBezug()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.loesen }),
    )

    // **In der Handlungszeile gesucht und nicht irgendwo** — die Seite trägt
    // weitere Meldungen (die Vorschau hat kein Bild), und eine Prüfung, die
    // „irgendein Alarm" nähme, wäre grün, ohne den Weg geprüft zu haben.
    expect(
      await screen.findByText(kataloge.de.fehler.unbekannt),
    ).toBeInTheDocument()
  })

  // **Ein Fehlschlag beim Lösen wird gemeldet, nicht verschluckt.**
  it('meldet einen abgewiesenen Lösevorgang nach oben', async () => {
    belegeApi({
      KarteVerweis: vi
        .fn()
        .mockResolvedValue({ verweis: verweis(), befunde: [], fehler: null }),
      KarteVerweisLoesen: vi.fn().mockResolvedValue({
        fehler: {
          code: 'belegt',
          schluessel: 'belegt',
          meldung: 'der Bestand ist belegt',
        },
      }),
    })

    const aufFehler = zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.loesen }),
    )

    await waitFor(() => {
      expect(aufFehler).toHaveBeenCalledWith(kataloge.de.fehler.belegt)
    })
  })
})

// **Ein Kartenwechsel räumt die Wahl, er blendet sie nicht aus** — Befund 1
// des Nachlaufs über Phase 3.6.
//
// # Warum das eine eigene Zusage ist
//
// Der erste Durchgang **filterte**: der Treffer blieb im Zustand und wurde
// nur nicht gezeigt. Karte A → B → **A** brachte damit denselben Vorschlag
// wieder — die Wahl war nie geräumt, sie war unsichtbar. §7, Regel 1 meint
// räumen.
describe('die Wahl beim Kartenwechsel', () => {
  it('kommt bei der Rückkehr zur ersten Karte nicht wieder', async () => {
    belegeApi({
      ...ausKarten(() =>
        Promise.resolve({
          karten: [
            { id: '5', nummer: 0, werte: [{ feld: 'name', wert: 'Erste' }] },
            { id: '6', nummer: 1, werte: [{ feld: 'name', wert: 'Zweite' }] },
          ],
        }),
      ),
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
      ScryfallSuche: () =>
        Promise.resolve({
          treffer: [treffer],
          weitere: false,
          alterStunden: 0,
          befunde: [],
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID="5"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    // **Beide Kästen** (#1240): die Suche steht im Anlegekasten, der Bezug
    // bei der Karte — seit UXR-06 (#1914) aus dem Menü „Diese Karte".
    await oeffneAnlegen()
    await oeffneBezug()

    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.seiten.karten.scryfallsuche.suchen,
      }),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: /Unlimited Edition/ }),
    )

    expect(
      await screen.findByRole('button', { name: /beziehen/ }),
    ).toBeInTheDocument()

    // **Zur zweiten Karte und zurück.**
    await userEvent.click(screen.getByRole('button', { name: /Zweite/ }))
    await userEvent.click(screen.getByRole('button', { name: /Erste/ }))

    // **Der Wechsel hat den Kasten geschlossen** (§7, Regel 1); wieder
    // geöffnet, gehört er der Karte, die jetzt offen ist.
    await oeffneBezug()

    expect(
      screen.queryByRole('button', { name: /beziehen/ }),
    ).not.toBeInTheDocument()
  })
})
