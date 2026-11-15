// **Einen Handel verbuchen** — Werkstück T6 von Phase 5.6
// (#1471).
//
// # Was hier gemessen wird
//
// **Welche Knöpfe stehen, sagt der Stand des Handels** — seit UXR-15
// (#1923)
// in `handelsbuchungknoepfe.test.tsx`. Die Läufe hier legen ihn deshalb über
// `HandelVorgaenge` fest (`buchungsprobe.ts`).
//
// **Dass nach einem gescheiterten Schritt kein Stand von vorhin stehen
// bleibt** (§7, Regel 1) — seit der Hüllenprobe `nat04handel` (#1672) liest
// der Kasten ihn aus dem Bestand (`mitBelegnachlauf`), und was dasteht, ist
// dessen Stand. Das Journal steht seit UXR-15 in einem eigenen Kasten und
// hat seinen eigenen Lauf (`handelsjournal.test.tsx`).
//
// **Und dass die zweite Unterschrift eine eigene Tat ist**: solange eine
// Stimme fehlt, steht „nur du hast unterschrieben" da.
import { screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import { zeigeHandel } from './tauschprobe'
import {
  bestaetigt,
  buchungsstand,
  dieVorgaenge,
  einJournal,
  einStand,
  eineFreigabe,
  imAbschluss,
  mitBelegnachlauf,
  standSagt,
  trageEin,
} from './buchungsprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln

// gebucht ist der Stand nach dem Verbuchen: als Nächstes der Abgleich.
const gebucht = imAbschluss({
  zustand: 'lokalVerbucht',
  naechster: 'abgleichen',
  moegliche: ['abgleichen', 'klaeren'],
})

describe('Einen Handel verbuchen', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  // **Das Journal steht seit UXR-15 im eigenen Kasten** (#1923) — was es
  // zeigt, misst `handelsjournal.test.tsx`.
  it('bucht, meldet den Abgleich, bestätigt den eigenen Abschluss und sagt, dass der Beleg einseitig ist', async () => {
    // **Der ganze Zug, wie im Betrieb** (Befund 4 des Nachlaufs über Phase
    // 8): der Bestand schreibt nach jedem Schritt seinen Stand fort, und Liste
    // und Nachlesen lesen ihn — sonst endete der Zug nach dem Verbuchen, und
    // Abgleich und eigene Stimme blieben ungeprüft.
    let bestand = imAbschluss()
    const verbucht = vi.fn(() => {
      bestand = gebucht

      return Promise.resolve(einStand())
    })
    const abgeglichen = vi.fn(() => {
      bestand = imAbschluss({
        zustand: 'abgleichOffen',
        beschluss: 'bestaetigen',
        naechster: 'bestaetigen',
        moegliche: ['bestaetigen', 'klaeren'],
      })

      return Promise.resolve(einStand({ zustand: 'abgleichOffen' }))
    })
    const gestimmt = vi.fn(() => {
      bestand = imAbschluss({
        zustand: 'abgleichOffen',
        beschluss: 'bestaetigen',
        zustimmungen: ['bich'],
        lage: 'wartet',
        naechster: '',
        moegliche: ['klaeren'],
      })

      return Promise.resolve(einStand({ zustand: 'abgleichOffen' }))
    })
    const gelesen = vi.fn(() => Promise.resolve(einJournal()))

    belegeApi(
      mitBelegnachlauf({
        Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
        HandelVorgaenge: () =>
          Promise.resolve({ vorgaenge: [bestand], fehler: null }),
        HandelStand: () => Promise.resolve({ vorgang: bestand, fehler: null }),
        HandelFreigabeErteilen: eineFreigabe,
        Handelsverbuchung: verbucht,
        Handelsabgleich: abgeglichen,
        Handelsbestaetigung: gestimmt,
        Handelsjournal: gelesen,
      }),
    )

    await trageEin(await zeigeHandel(), eigene.buchungVerbuchen)

    // **Der Zug liest vorher nach, legt die Übergabe in den Ausgang** und geht
    // dann bis zur eigenen Abschlussstimme.
    await vi.waitFor(() => {
      expect(gestimmt).toHaveBeenCalledWith('v-1', 'bich', 'f-1')
    })
    expect(verbucht).toHaveBeenCalledWith('v-1', 'f-1')
    expect(abgeglichen).toHaveBeenCalledWith('v-1')

    // **Und der Beleg ist einseitig.**
    await standSagt(eigene.buchungEinseitig)

    // **Das Journal fragt erst, wer es aufklappt** — eine Buchung liest es
    // nicht mehr mit.
    expect(gelesen).not.toHaveBeenCalled()
  })

  // **Ein vierzehnter Zustand ist immer noch eine Auskunft** — der Rückfall
  // zeigt ihn, statt das Feld leer zu lassen (§10: ein Zweig, den kein Lauf
  // betritt, ist eine Zusage, die niemand hält).
  it('zeigt einen Zustand, den dieses Fenster nicht kennt', async () => {
    belegeApi(
      mitBelegnachlauf({
        Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
        HandelVorgaenge: dieVorgaenge(imAbschluss(), gebucht),
        HandelFreigabeErteilen: eineFreigabe,
        Handelsverbuchung: () =>
          Promise.resolve(einStand({ zustand: 'schwebend', hatBeleg: false })),
        Handelsjournal: () => Promise.resolve(einJournal()),
      }),
    )

    await trageEin(await zeigeHandel(), eigene.buchungVerbuchen)

    await standSagt('schwebend')
  })

  // **Nach einem gescheiterten Schritt steht der Stand des Bestands** — und
  // nicht, was vorher geantwortet wurde (§7, Regel 1). Bis zur Hüllenprobe
  // `nat04handel` (#1672) räumte der Kasten die Zeile; seither liest er sie,
  // und ein gescheiterter Schritt liest gar nicht erst nach (Regel 3).
  it('meldet einen gescheiterten Schritt und zeigt den Stand des Bestands', async () => {
    belegeApi(
      mitBelegnachlauf({
        Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
        HandelVorgaenge: dieVorgaenge(imAbschluss()),
        HandelFreigabeErteilen: eineFreigabe,
        Handelsverbuchung: () =>
          Promise.resolve(
            einStand({
              fehler: {
                code: 'konflikt' as const,
                schluessel: 'handel',
                meldung: 'so nicht',
              },
            }),
          ),
        Handelsjournal: () => Promise.resolve(einJournal()),
      }),
    )

    await trageEin(await zeigeHandel(), eigene.buchungVerbuchen)

    // **Der Zug nennt den Schritt dazu** (Vermerk 0200).
    expect(
      await screen.findByText(
        `${eigene.buchungVerbuchen} ging nicht: so nicht`,
      ),
    ).toBeInTheDocument()

    await standSagt(eigene.buchungStandUebergabeOffen)
    expect(await buchungsstand()).not.toHaveTextContent(
      eigene.buchungStandLokalVerbucht,
    )
  })

  it('meldet den Abgleich und nimmt die Stimme der Gegenseite', async () => {
    // **Die Freigabe wird hier gezählt** — an ihr hängt Befund 5 des
    // Zweitblicks über Phase 5.6.
    const erteilt = vi.fn(eineFreigabe)

    const abgeglichen = vi.fn(() =>
      Promise.resolve(einStand({ zustand: 'abgleichOffen' })),
    )

    const gestimmt = vi.fn(() =>
      Promise.resolve(
        einStand({
          zustand: 'beidseitigVerbucht',
          beleg: {
            vorgang: 'v-1',
            abdruck: 'abdruck-1',
            unterschrift: 'deadbeef',
            beidseitig: true,
            gegenzeichner: 'peer-bert',
            gegengezeichnet: '2026-09-19T12:00:00Z',
            gebucht: '2026-09-18T12:00:00Z',
            befund: null,
          },
        }),
      ),
    )

    belegeApi(
      mitBelegnachlauf({
        Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
        // **Erst der Abgleich, dann die Stimme des Gegenübers** — welcher
        // Knopf steht, sagt der Stand, den der Raum nach jedem Schritt lädt.
        HandelVorgaenge: dieVorgaenge(
          gebucht,
          imAbschluss({
            zustand: 'abgleichOffen',
            naechster: 'bestaetigen',
            moegliche: ['bestaetigen', 'klaeren'],
            beschluss: 'bestaetigen',
          }),
        ),
        HandelFreigabeErteilen: erteilt,
        Handelsabgleich: abgeglichen,
        Handelsbestaetigung: gestimmt,
        Handelsjournal: () => Promise.resolve(einJournal()),
      }),
    )

    const hand = await zeigeHandel()

    await trageEin(hand, eigene.buchungAbgleichen)

    expect(abgeglichen).toHaveBeenCalledWith('v-1')
    await standSagt(eigene.buchungStandAbgleichOffen)

    await trageEin(hand, eigene.buchungBestaetigen)

    // **Die Stimme kommt vom Kontakt, den man ansieht.**
    expect(gestimmt).toHaveBeenCalledWith('v-1', 'bgegenueber', 'f-1')

    // **Und die Freigabe gilt genau dieser Stimme** — Befund 5 des
    // Zweitblicks über Phase 5.6: eine, die nur den Vorgang nennte, gälte
    // für **jede** Unterschrift.
    expect(erteilt).toHaveBeenCalledWith('v-1', 'bestaetigen', 'bgegenueber')

    await standSagt(eigene.buchungBeidseitig)
  })

  // **Ohne Zustimmung wird nicht gebucht** — Werkstück T10 von Phase 5.6
  // (#1475).
  //
  // **Der Riegel sitzt im Dienst und nicht hier**; was dieser Lauf hält, ist
  // die Reihenfolge: scheitert das Erteilen, wird die Verbuchung **gar nicht
  // gerufen**. Ein Fenster, das trotzdem riefe, verließe sich darauf, dass
  // drüben jemand nein sagt — und genau das ist die Gestalt, die §10 meint.
  it('bucht nicht, wenn die Freigabe scheitert', async () => {
    const verbucht = vi.fn()

    belegeApi(
      mitBelegnachlauf({
        Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
        HandelVorgaenge: dieVorgaenge(imAbschluss(), gebucht),
        HandelFreigabeErteilen: () =>
          Promise.resolve({
            kennung: '',
            laeuftab: '',
            fehler: {
              code: 'erlaubnis' as const,
              schluessel: 'handel',
              meldung: 'den Handelsvorgang gibt es nicht',
            },
          }),
        Handelsverbuchung: verbucht,
      }),
    )

    const hand = await zeigeHandel()

    await trageEin(hand, eigene.buchungVerbuchen)

    expect(
      await screen.findByText(
        `${eigene.buchungVerbuchen} ging nicht: den Handelsvorgang gibt es nicht`,
      ),
    ).toBeInTheDocument()

    expect(verbucht).not.toHaveBeenCalled()
  })

  it('bucht nicht, wenn sich der Handel seit dem Ansehen geändert hat', async () => {
    // **Befund 1 des Zweitblicks über Phase 8**: der Zug liest nach und geht
    // nichts, wenn nicht mehr gilt, was das Fenster zeigt — der Raum liest
    // den Handel neu.
    const verbucht = vi.fn(() => Promise.resolve(einStand()))

    belegeApi(
      mitBelegnachlauf({
        Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
        HandelVorgaenge: () =>
          Promise.resolve({ vorgaenge: [imAbschluss()], fehler: null }),
        HandelStand: () =>
          Promise.resolve({
            vorgang: imAbschluss({ zustand: 'klaerung', revision: 2 }),
            fehler: null,
          }),
        HandelFreigabeErteilen: eineFreigabe,
        Handelsverbuchung: verbucht,
      }),
    )

    await trageEin(await zeigeHandel(), eigene.buchungVerbuchen)

    expect(await screen.findByText(eigene.zug.veraltet)).toBeInTheDocument()
    expect(verbucht).not.toHaveBeenCalled()
  })
})
