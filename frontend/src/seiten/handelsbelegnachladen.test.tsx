// **Der Beleg steht da, ohne dass jemand klickt** — gefunden mit zwei
// Fenstern (Hüllenprobe `nat04handel`, S12.NAT-04,
// #1672).
//
// Bert bestätigte zuerst, Annas Beleg kam danach über die Leitung, und sein
// Fenster sagte weiter „Nur du hast unterschrieben": der Kasten zeigte die
// Antwort auf den eigenen letzten Klick, und einen Weg, den Beleg ohne
// Schritt zu lesen, gab es nicht. Wer einen abgeschlossenen Handel wieder
// aufschlug, sah gar keinen.
import { render, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { mitStandnachlauf } from '../test/standnachlauf'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { alsZeitpunkt } from '../zeitpunkt'
import { Sprachraum } from '../sprache/kontext'
import { Handeln } from './Handeln'
import { eineListe } from './peerkontaktprobe'
import { waehleGegenueber } from './tauschprobe'
import {
  bestaetigt,
  buchung,
  dieVorgaenge,
  einJournal,
  einStand,
  imAbschluss,
} from './buchungsprobe'

import type { Verbuchungsantwort } from '../apihandelsverbuchung'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

function seite(bestandsstand: number) {
  return (
    <Sprachraum sprache="de">
      <Handeln bestandsstand={bestandsstand} />
    </Sprachraum>
  )
}

// abgeschlossen ist der Handel nach beiden Stimmen.
const abgeschlossen = imAbschluss({
  zustand: 'beidseitigVerbucht',
  lage: 'abgeschlossen',
  naechster: '',
  moegliche: [],
  ende: true,
})

// mitBeleg belegt den Raum mit einem Handel und einem Beleg, den der Lauf
// auswechselt — wie drüben die Leitung.
function mitBeleg(anfang: Verbuchungsantwort) {
  const bestand = { beleg: anfang }

  belegeApi(
    mitStandnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: dieVorgaenge(abgeschlossen),
      Handelsbeleg: () => Promise.resolve(bestand.beleg),
      Handelsjournal: () => Promise.resolve(einJournal()),
    }),
  )

  return bestand
}

const einseitig = einStand({ zustand: 'beidseitigVerbucht' })

const beidseitig = einStand({
  zustand: 'beidseitigVerbucht',
  beleg: {
    ...einseitig.beleg,
    beidseitig: true,
    gegenzeichner: 'bgegenueber',
    gegengezeichnet: '2026-10-04T16:45:19Z',
  },
})

it('zeigt den Beleg beim Aufschlagen, ohne einen Schritt', async () => {
  mitBeleg(beidseitig)
  const hand = userEvent.setup()
  render(seite(0))
  await waehleGegenueber(hand, 'Fremder Rechner')

  const kasten = await buchung(hand)

  expect(
    await within(kasten).findByText(eigene.buchungBeidseitig, {
      exact: false,
    }),
  ).toBeVisible()
})

it('zeigt einen Gegenbeleg, der ohne Klick ankam', async () => {
  const bestand = mitBeleg(einseitig)
  const hand = userEvent.setup()
  const { rerender } = render(seite(0))
  await waehleGegenueber(hand, 'Fremder Rechner')

  const kasten = await buchung(hand)
  expect(
    await within(kasten).findByText(eigene.buchungOhneGegenbeleg),
  ).toBeVisible()

  bestand.beleg = beidseitig
  rerender(seite(1))

  expect(
    await within(kasten).findByText(
      einsetzen(eigene.buchungGegenbeleg, {
        wer: 'Fremder Rechner',
        // **Lesbar und nicht roh** — wie „zuletzt" in der Liste.
        wann: alsZeitpunkt('2026-10-04T16:45:19Z', 'de'),
      }),
    ),
  ).toBeVisible()
})

// **Und „Verbuchen" steht nur, solange es etwas zu verbuchen gibt** — bis
// dahin war ein zweiter Druck der einzige Weg, den Beleg zu sehen.
it('bietet Verbuchen nach dem Verbuchen nicht mehr an', async () => {
  mitBeleg(beidseitig)
  const hand = userEvent.setup()
  render(seite(0))
  await waehleGegenueber(hand, 'Fremder Rechner')

  const kasten = await buchung(hand)
  await within(kasten).findByText(eigene.buchungBeidseitig, { exact: false })

  expect(
    within(kasten).queryByRole('button', { name: eigene.buchungVerbuchen }),
  ).toBeNull()
})
