// **Was die Hilfeseite zeigt** (#806).
//
// # Warum es diesen Test gibt
//
// Die Seite war bis zum 2026-08-22 **abgedeckt und unzugesichert**: gerendert
// hat sie allein der axe-Durchlauf über `navigation.ts`, und der fragt nach
// Barrierefreiheit, nicht nach Inhalt. Die Abdeckung stand bei 100 %, während
// niemand prüfte, ob auch nur ein Wort dastand.
import userEvent from '@testing-library/user-event'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { seitenschluessel } from '../navigation'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Hilfe } from './Hilfe'
import { belegeApi } from '../test/api'

function zeige(sprache: 'de' | 'en' = 'de', einfuehrungZeigen = vi.fn()) {
  render(
    <Sprachraum sprache={sprache}>
      <Hilfe
        fassung="0.0.0-test"
        zuUeber={false}
        angekommen={vi.fn()}
        einfuehrungZeigen={einfuehrungZeigen}
      />
    </Sprachraum>,
  )
}

// **Die Übersicht der Seiten ist eine Lasche weiter** — seit UXR-22
// (#1930) stehen die Begriffe zuerst. Die Läufe, die die Übersicht meinen,
// schlagen sie deshalb auf, statt sie vorn zu erwarten.
async function seitenuebersicht(sprache: 'de' | 'en' = 'de') {
  await userEvent.click(
    screen.getByRole('tab', {
      name: kataloge[sprache].seiten.hilfe.seitenuebersicht,
    }),
  )
}

describe('Hilfe', () => {
  // **Jede Seite, in der Reihenfolge der Navigation.**
  //
  // Zwei Zusagen in einer, und die zweite ist die stärkere: eine Aufzählung,
  // die zwölf von dreizehn zeigt, sähe sonst genauso aus wie eine
  // vollständige — und ein `Object.entries` über den Hilfekatalog wäre
  // ebenfalls grün, während sich seine Folge beim nächsten Nachtragen
  // verschöbe.
  //
  // **Über die Begriffsrolle und nicht über den Text**: „Hilfe" und
  // „Druckbogen" stehen auch im Glossar darunter, und `getByText` fand beide.
  it('zeigt jede Seite in der Reihenfolge aus navigation.ts', async () => {
    zeige()
    await seitenuebersicht()

    const ueberschriften = [...seitenschluessel].map(
      (schluessel) => kataloge.de.hilfe.seiten[schluessel].ueberschrift,
    )

    const gezeigt = screen
      .getAllByRole('term')
      .map((knoten) => knoten.textContent)

    // Die Begriffsliste steht darunter; verglichen wird der Anfang.
    expect(gezeigt.slice(0, ueberschriften.length)).toEqual(ueberschriften)
  })

  // **Beide liegen hinter ihrer Lasche** — Werkstück I von Phase 4.10
  // (#1325).
  // Glossar und Schnittstellen sind zwei Gebiete; der Test geht damit
  // denselben Weg wie ein Anwender.
  it('zeigt die Begriffe und was hinausgeht', async () => {
    zeige()

    await userEvent.click(
      screen.getByRole('tab', { name: kataloge.de.seiten.hilfe.glossar }),
    )

    expect(
      screen.getByText(kataloge.de.hilfe.begriffe.vorbehalt.name),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('tab', {
        name: kataloge.de.hilfe.schnittstellen.ueberschrift,
      }),
    )

    expect(
      screen.getByText(kataloge.de.hilfe.schnittstellen.mcpName),
    ).toBeInTheDocument()
  })

  it('spricht englisch, wenn die Sitzung es tut', async () => {
    zeige('en')
    await seitenuebersicht('en')

    const englisch = [...seitenschluessel].map(
      (schluessel) => kataloge.en.hilfe.seiten[schluessel].ueberschrift,
    )

    const gezeigt = screen
      .getAllByRole('term')
      .map((knoten) => knoten.textContent)

    expect(gezeigt.slice(0, englisch.length)).toEqual(englisch)
  })
})

// **Der Weg zurück aus »Über« wird gegangen** — und nicht nur der hin.
//
// # Warum dieser Lauf nachgetragen wurde
//
// **Weil die Hälfte des Wechsels unbetreten war.** `Hilfe` merkt sich, ob sie
// zuletzt wegen des Menüeintrags aufgeschlagen wurde; bei jedem **Wechsel**
// dieses Werts wird die Erinnerung nachgezogen, und nur beim Wechsel **nach**
// „Über" wird zusätzlich gerollt. Getestet war bis zum 2026-09-06 allein
// `zuUeber={false}` — der Wechsel fand nie statt.
//
// **Was das kostet, hat die Pipeline gezeigt**: der Zweig flackerte in der
// Abdeckung, weil ihn kein Lauf ausdrücklich meinte.
describe('Der Wechsel nach »Über« und zurück', () => {
  // **Führungen, Kürzel und die Bereiche ohne Seite** — Werkstück U11 von
  // Phase 5.1 (#1422).
  it('holt die ersten Schritte aus den Führungen zurück', async () => {
    const einfuehrungZeigen = vi.fn()
    zeige('de', einfuehrungZeigen)

    await userEvent.click(
      screen.getByRole('tab', {
        name: kataloge.de.hilfe.fuehrungen.ueberschrift,
      }),
    )
    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.hilfe.fuehrungen.startZeigen,
      }),
    )

    expect(einfuehrungZeigen).toHaveBeenCalledTimes(1)
  })

  it('zeigt die Tastenkürzel hinter ihrer Lasche', async () => {
    zeige()

    await userEvent.click(
      screen.getByRole('tab', {
        name: kataloge.de.seiten.einstellungen.kuerzelUeberschrift,
      }),
    )

    // ⌘K steht in der Tafel — der Beleg, dass es die Tafel ist und nicht ihr Name.
    expect(await screen.findByText('⌘K')).toBeInTheDocument()
  })

  it('führt vom KI-Bereich zu den Schnittstellen', async () => {
    zeige()
    await seitenuebersicht()
    const { bereiche, schnittstellen } = kataloge.de.hilfe

    // **Spielen und Handeln stehen hier nicht mehr** — seit Werkstück G1 von
    // Phase 5.5 (#1458)
    // und `T9` von Phase 5.6
    // (#1474)
    // sind beide Seiten der Leiste und keine Bereiche ohne eine. Wer sie hier
    // ließe, nennte sie zweimal, und der Satz daneben wäre falsch.
    //
    // **Die KI bleibt, und sie ist der Grund für den Abschnitt**: sie ist kein
    // Ort, sondern ein Weg von außen herein.
    expect(screen.getByText(bereiche.kiName)).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: bereiche.kiKnopf }),
    )

    expect(
      await screen.findByRole('tab', {
        name: schnittstellen.ueberschrift,
        selected: true,
      }),
    ).toBeInTheDocument()
  })

  it('zieht die Erinnerung in beide Richtungen nach', async () => {
    const angekommen = vi.fn()

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Hilfe
          fassung="0.0.0-test"
          zuUeber
          angekommen={angekommen}
          einfuehrungZeigen={vi.fn()}
        />
      </Sprachraum>,
    )

    // **Erst der Beleg, dass der Hinweg wirkt** — ohne ihn sagte der Rückweg
    // darunter nur, dass nichts abstürzt (§10).
    await waitFor(() => {
      expect(angekommen).toHaveBeenCalled()
    })

    angekommen.mockClear()

    rerender(
      <Sprachraum sprache="de">
        <Hilfe
          fassung="0.0.0-test"
          zuUeber={false}
          angekommen={angekommen}
          einfuehrungZeigen={vi.fn()}
        />
      </Sprachraum>,
    )

    // **Und die Seite steht weiter da.** Der Rückweg zieht die Erinnerung
    // nach und rollt nicht — er darf die Seite nicht leeren.
    expect(
      screen.getByRole('heading', {
        name: kataloge.de.hilfe.seiten.hilfe.ueberschrift,
      }),
    ).toBeInTheDocument()
  })

  // **Die Grenzen stehen bei den Schnittstellen** — Werkstück O5 von Phase
  // 5.2 (#1429), gelesen aus derselben Auskunft wie auf den anderen Flächen.
  it('nennt die Grenzen dieses Prozesses', async () => {
    belegeApi({})
    zeige()

    await userEvent.click(
      screen.getByRole('tab', {
        name: kataloge.de.hilfe.schnittstellen.ueberschrift,
      }),
    )

    expect(
      await screen.findByText(/120 Einheiten am Stück/),
    ).toBeInTheDocument()
    expect(screen.getByText(/höchstens 2 Rasterungen/)).toBeInTheDocument()

    // **Mit Kontingenten sagt sie nichts darüber** — sie gelten, wie die
    // Zahlen davor es sagen.
    expect(
      screen.queryByText(/Kontingente sind abgeschaltet/),
    ).not.toBeInTheDocument()
  })

  // **Abgeschaltete Kontingente stehen da** — QA-08.AK-03 (#1739): wer kein
  // Nein bekommt, soll wissen, warum, und was trotzdem gilt.
  it('sagt, wenn die Kontingente abgeschaltet sind', async () => {
    belegeApi({
      Grenzen: () =>
        Promise.resolve({
          grenzen: {
            parallel: { rendern: 2 },
            antwortBytes: 1,
            principal: { burst: 120, nachfuellen: 120, fensterSekunden: 60 },
            global: { burst: 600, nachfuellen: 600, fensterSekunden: 60 },
            reserve: 120,
            kosten: {},
            kontingente: 'aus',
            verbindungen: 16,
            auftraege: { laufend: 4, behalten: 16, aufbewahrungSekunden: 3600 },
            kostentafel: [],
          },
          fehler: null,
        }),
    })
    zeige()

    await userEvent.click(
      screen.getByRole('tab', {
        name: kataloge.de.hilfe.schnittstellen.ueberschrift,
      }),
    )

    const satz = await screen.findByText(/Kontingente sind abgeschaltet/)
    expect(satz.closest('[data-kontingente]')).toHaveTextContent(
      /Zugang und TLS gelten weiter/,
    )
  })

  // **Eine Auskunft ohne Tabellen fällt auf die Vorgaben zurück, ein Fehler
  // steht am Ort** — die beiden Ränder derselben Zeile.
  it('trägt die Grenzen auch ohne Tabellen und zeigt einen Fehler am Ort', async () => {
    belegeApi({
      Grenzen: () =>
        Promise.resolve({
          grenzen: {
            parallel: {},
            antwortBytes: 1,
            principal: { burst: 7, nachfuellen: 9, fensterSekunden: 60 },
            global: { burst: 7, nachfuellen: 9, fensterSekunden: 60 },
            reserve: 3,
            kosten: {},
            kontingente: 'an',
            verbindungen: 16,
            auftraege: { laufend: 4, behalten: 16, aufbewahrungSekunden: 3600 },
            kostentafel: [],
          },
          fehler: null,
        }),
    })
    zeige()

    await userEvent.click(
      screen.getByRole('tab', {
        name: kataloge.de.hilfe.schnittstellen.ueberschrift,
      }),
    )

    expect(
      await screen.findByText(
        /7 Einheiten am Stück, 9 je Minute; Lesen kostet 1, Schreiben 2/,
      ),
    ).toBeInTheDocument()
    cleanup()

    belegeApi({
      Grenzen: () =>
        Promise.resolve({
          grenzen: {
            parallel: {},
            antwortBytes: 0,
            principal: { burst: 0, nachfuellen: 0, fensterSekunden: 0 },
            global: { burst: 0, nachfuellen: 0, fensterSekunden: 0 },
            reserve: 0,
            kosten: {},
            kontingente: 'an',
            verbindungen: 16,
            auftraege: { laufend: 4, behalten: 16, aufbewahrungSekunden: 3600 },
            kostentafel: [],
          },
          fehler: {
            code: 'unbekannt',
            meldung: 'kein Haushalt',
            feld: '',
            schluessel: '',
          },
        }),
    })
    zeige()

    await userEvent.click(
      screen.getByRole('tab', {
        name: kataloge.de.hilfe.schnittstellen.ueberschrift,
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('kein Haushalt')
  })
})
