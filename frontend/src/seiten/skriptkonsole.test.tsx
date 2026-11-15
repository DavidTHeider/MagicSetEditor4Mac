// **Die Skriptkonsole**
// (#1070).
//
// # Wofür
//
// §15 nennt MSEs eigene Kommandozeile als **Orakel** dieses Projekts. Eine
// Konsole im Fenster beantwortet dieselben Fragen über **dieses** Programm und
// macht die beiden vergleichbar — ohne dass jemand ein Terminal aufmacht.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import {
  stillePflege,
  stilleDatenpflege,
  stillerRuecksprung,
} from '../test/bestandspflegeprobe'
import { einstellungsprobe } from '../test/seitenvorrichtung'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { Seiteninhalt } from './Seiteninhalt'
import { Skriptkonsole } from './Skriptkonsole'

// seiteninhaltMitKarte baut die Konsolenseite so, wie das Fenster sie baut —
// über `Seiteninhalt`, weil dort der Schlüssel steht.
function seiteninhaltMitKarte(karteID: string) {
  return (
    <Sprachraum sprache="de">
      <Seiteninhalt
        bestandGeaendert={vi.fn()}
        bestandspflege={stillePflege()}
        ruecksprung={stillerRuecksprung()}
        pflege={stilleDatenpflege()}
        einstellungen={einstellungsprobe}
        karteID={karteID}
        meldefehler={undefined}
        oeffnungsanstoss={false}
        oeffnungAngestossen={vi.fn()}
        neuanstoss={false}
        neuAngestossen={vi.fn()}
        bestandsstand={0}
        oeffneKarte={vi.fn()}
        oeffneSymbol={vi.fn()}
        seite="konsole"
        setID="4"
        speichern={stillerSpeicherlauf()}
        verlaufsfehler={undefined}
        verlaufsstand={0}
        waehleSet={vi.fn()}
        setEntfernt={vi.fn()}
        wechseln={vi.fn()}
        fassung="0.0.0-test"
        zuUeber={false}
        einfuehrungZeigen={vi.fn()}
        ueberAngekommen={vi.fn()}
      />
    </Sprachraum>
  )
}

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.konsole

function zeige(karteID = '7', zuDenKarten = vi.fn()) {
  const Skript = vi.fn(() =>
    Promise.resolve({
      quelltext: '"ab"',
      text: 'ab',
      art: 'text',
      befunde: [],
      stelle: null,
      dauerMs: 0,
      abgebrochen: false,
      eintraege: [],
      fehler: null,
    }),
  )

  belegeApi({ Skript })

  render(
    <Sprachraum sprache="de">
      <Skriptkonsole karteID={karteID} zuDenKarten={zuDenKarten} />
    </Sprachraum>,
  )

  return Skript
}

describe('Die Skriptkonsole', () => {
  it('wertet einen Ausdruck aus und zeigt den Quelltext', async () => {
    const Skript = zeige()

    await userEvent.type(screen.getByLabelText(texte.ausdruck), '"a" + "b"')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))

    await waitFor(() => {
      expect(Skript).toHaveBeenCalledWith('7', '"a" + "b"')
    })

    // **Der Quelltext und nicht der Text** — so zeigt MSE es: wer eine Konsole
    // benutzt, will eine Zeichenkette von einem Namen unterscheiden.
    expect(await screen.findByText('"ab"')).toBeInTheDocument()
    expect(screen.getByText('Art: text')).toBeInTheDocument()
  })

  // **Ohne aufgeschlagene Karte gibt es kein `card`** — und die Seite sagt es,
  // statt ein Eingabefeld anzubieten, dessen Ergebnis niemand nachvollziehen
  // kann.
  it('sagt es, wenn keine Karte offen ist', () => {
    zeige('')

    expect(screen.getByText(texte.keineKarte)).toBeInTheDocument()
    expect(screen.queryByLabelText(texte.ausdruck)).not.toBeInTheDocument()
  })

  // **Und sie sagt, wie man dorthin kommt** — Werkstück C von Phase 4.10
  // (#1325).
  //
  // Der Satz nannte, **was fehlt**, und ließ offen, **was zu tun ist**. §6
  // verlangt das Zweite; wer die Konsole sucht, muss sonst selbst wissen, dass
  // der Weg über die Kartenlasche geht.
  it('führt von hier zu einer Karte', async () => {
    const gewechselt = vi.fn()

    zeige('', gewechselt)

    await userEvent.click(
      screen.getByRole('button', { name: texte.zuDenKarten }),
    )

    expect(gewechselt).toHaveBeenCalledWith('karten')
  })

  it('zeigt die Vorbehalte beim Aufbau der Umgebung', async () => {
    belegeApi({
      Skript: () =>
        Promise.resolve({
          quelltext: '2',
          text: '2',
          art: 'zahl',
          befunde: ['das Stilpaket liegt nicht vor'],
          stelle: null,
          dauerMs: 0,
          abgebrochen: false,
          eintraege: [],
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Skriptkonsole karteID="7" zuDenKarten={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.type(screen.getByLabelText(texte.ausdruck), '1 + 1')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))

    expect(
      await screen.findByText('das Stilpaket liegt nicht vor'),
    ).toBeInTheDocument()

    expect(screen.getByText(texte.vorbehalte)).toBeInTheDocument()
  })

  // **Ein Zerteilfehler steht am Ausdruck** und nicht als anonymer Hinweis
  // (§6) — mit seiner Stelle, wie überall sonst.
  it('meldet einen gescheiterten Ausdruck', async () => {
    belegeApi({
      Skript: () =>
        Promise.resolve({
          quelltext: '',
          text: '',
          art: '',
          befunde: [],
          stelle: null,
          dauerMs: 0,
          abgebrochen: false,
          eintraege: [],
          fehler: {
            code: 'validierung' as const,
            schluessel: 'probe',
            meldung: 'Zeile 1, Spalte 3: der Ausdruck ist unvollständig.',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Skriptkonsole karteID="7" zuDenKarten={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.type(screen.getByLabelText(texte.ausdruck), '1 +')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Zeile 1, Spalte 3',
    )
  })
})

// **Ein Kartenwechsel räumt die Konsole** — Befund 6 des Zweitblicks durch
// Codex über Phase W.
//
// # Der Befund
//
// `Seiteninhalt` schlüsselte die Konsole über `setID`. Die aufgeschlagene
// **Karte** ist aber ihr Bezugsobjekt: sie steht in der Überschrift, sie geht
// in jeden Aufruf, und `card` im Ausdruck meint sie. Ein Kartenwechsel ließ
// Ausdruck, Ergebnis und Fehler stehen — §7, Regel 1.
//
// Und §7, Regel 2 dazu: eine verspätete Antwort für Karte A wurde unter Karte
// B eingetragen. `useVorgang` verhindert nur zwei gleichzeitige Aufrufe
// **derselben** Kennung.
//
// # Warum dieser Lauf durch `Seiteninhalt` geht
//
// Weil der Schlüssel **dort** steht. Ein Lauf, der die Konsole selbst
// aufbaut, schriebe den richtigen Schlüssel hin und prüfte die eigene
// Testzeile — der erste Anlauf tat genau das und war grün, während der Fehler
// unverändert dastand. `setwechselraeumt.test.tsx` nennt dieselbe Grenze für
// seine fünf Seiten und prüft sie deshalb ausdrücklich **nicht**.
//
// # Was ein Schlüssel löst
//
// `docs/SCHREIBWEGE.md` sagt es: ein `key`, der das Bauteil beim Wechsel
// abräumt, löst Regel 1 und Regel 2 **zugleich** — die verspätete Antwort
// trägt in einen Zustand ein, den es nicht mehr gibt.
describe('Ein Kartenwechsel', () => {
  it('räumt Ausdruck und Ergebnis der Konsole', async () => {
    const Skript = vi.fn(() =>
      Promise.resolve({
        quelltext: '"ab"',
        text: 'ab',
        art: 'text',
        befunde: [],
        stelle: null,
        dauerMs: 0,
        abgebrochen: false,
        eintraege: [],
        fehler: null,
      }),
    )

    belegeApi({ Skript })

    const { rerender } = render(seiteninhaltMitKarte('7'))

    const feld = await screen.findByLabelText(texte.ausdruck)
    await userEvent.type(feld, 'card.name')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))

    await waitFor(() => {
      expect(Skript).toHaveBeenCalledWith('7', 'card.name')
    })

    expect(await screen.findByText('"ab"')).toBeInTheDocument()

    rerender(seiteninhaltMitKarte('9'))

    // **Der Ausdruck ist weg, und das Ergebnis auch.**
    expect(await screen.findByLabelText(texte.ausdruck)).toHaveValue('')
    expect(screen.queryByText('"ab"')).not.toBeInTheDocument()
  })
})

// **Und eine verspätete Antwort erreicht die neue Karte nicht** — §7, Regel 2.
//
// Der Lauf darüber prüft das **sichtbare** Räumen. Diesen zweiten Teil prüfte
// er nicht: eine Antwort, die erst nach dem Kartenwechsel eintrifft, trägt in
// einen Zustand ein, den es nicht mehr gibt. Angestrichen vom Nachlauf des
// Zweitblicks über Phase W.
//
// **Die Antwort wird eigens angehalten** und erst nach dem Wechsel
// freigegeben — ohne das käme sie zu früh, und der Lauf prüfte wieder nur das
// Räumen.
describe('Eine verspätete Antwort', () => {
  it('erreicht die neue Karte nicht', async () => {
    let freigeben: (() => void) | undefined
    const angehalten = new Promise<void>((aufloesen) => {
      freigeben = aufloesen
    })

    const Skript = vi.fn(async () => {
      await angehalten

      return {
        quelltext: '"von Karte 7"',
        text: 'von Karte 7',
        art: 'text',
        befunde: [],
        stelle: null,
        dauerMs: 0,
        abgebrochen: false,
        eintraege: [],
        fehler: null,
      }
    })

    belegeApi({ Skript })

    const { rerender } = render(seiteninhaltMitKarte('7'))

    const feld = await screen.findByLabelText(texte.ausdruck)
    await userEvent.type(feld, 'card.name')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))

    await waitFor(() => {
      expect(Skript).toHaveBeenCalledWith('7', 'card.name')
    })

    // **Gewechselt, während die Antwort noch aussteht.**
    rerender(seiteninhaltMitKarte('9'))
    expect(await screen.findByLabelText(texte.ausdruck)).toHaveValue('')

    freigeben?.()
    await angehalten

    // **Und sie taucht nirgends auf.**
    await waitFor(() => {
      expect(screen.queryByText('"von Karte 7"')).not.toBeInTheDocument()
    })

    expect(await screen.findByLabelText(texte.ausdruck)).toHaveValue('')
  })
})
