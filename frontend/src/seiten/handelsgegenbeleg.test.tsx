// **Der Gegenbeleg einer Buchung** — Werkstück T6 von Phase 5.6
// (#1471),
// #1650.
//
// **Die Vorrichtung teilt es mit `handelsbuchung.test.tsx`**
// (`buchungsprobe.ts`) — seit UXR-15 hängen die Knöpfe am Stand des Handels.
import { screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import { zeigeHandel } from './tauschprobe'
import {
  bestaetigt,
  dieVorgaenge,
  einJournal,
  einStand,
  eineFreigabe,
  imAbschluss,
  trageEin,
  mitBelegnachlauf,
} from './buchungsprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln

// **Was den beidseitigen Abschluss deckt, steht dabei** —
// #1650.
//
// **Und wo nichts steht, sagt es der Satz.** „Beide haben unterschrieben"
// ohne den Namen dessen, der es getan hat, ist eine Behauptung — und eine
// leere Zeile wäre eine offene Frage.
describe('Der Gegenbeleg', () => {
  afterEach(() => {
    raeumeApi()
  })

  it('nennt ihn mit Kennung, wenn einer da ist', async () => {
    belegeApi(
      mitBelegnachlauf({
        Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
        HandelVorgaenge: dieVorgaenge(imAbschluss()),
        Handelsverbuchung: () =>
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
        HandelFreigabeErteilen: eineFreigabe,
        Handelsjournal: () => Promise.resolve(einJournal()),
      }),
    )

    await trageEin(await zeigeHandel(), eigene.buchungVerbuchen)

    expect(
      await screen.findByText(/peer-bert/, { exact: false }),
    ).toBeInTheDocument()
  })

  it('sagt ausdrücklich, wenn keiner da ist', async () => {
    belegeApi(
      mitBelegnachlauf({
        Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
        HandelVorgaenge: dieVorgaenge(imAbschluss()),
        Handelsverbuchung: () => Promise.resolve(einStand()),
        HandelFreigabeErteilen: eineFreigabe,
        Handelsjournal: () => Promise.resolve(einJournal()),
      }),
    )

    await trageEin(await zeigeHandel(), eigene.buchungVerbuchen)

    expect(
      await screen.findByText(eigene.buchungOhneGegenbeleg),
    ).toBeInTheDocument()
  })
})
