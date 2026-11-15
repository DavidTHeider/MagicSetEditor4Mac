// **Welche Besitzzeilen weggehen** —
// #1643.
//
// # Was hier gemessen wird
//
// **Dass ein Mensch eine Zeile benennt und sie danach dasteht.** Bis zu
// diesem Werkstück konnte dieses Programm kein einziges Exemplar für einen
// Handel festhalten — `storage.Festhalten` stand ohne Aufrufer da.
//
// **Und dass eine Zahl, die keine ist, am Feld stehen bleibt** (F03): sie
// wird weder `NaN` — das ginge als `null` durch die Serialisierung — noch
// eine Zahl, die niemand getippt hat.
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { oeffneHandel, zeigeHandel, zuPhase } from './tauschprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Handelszusage } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

// bestaetigt ist ein Kontakt, an dem der Kasten erscheint.
function bestaetigt() {
  return einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })
}

// imKasten gibt den Kasten der Zusagen.
function imKasten() {
  return screen.getByRole('region', { name: eigene.zusagenTitel })
}

// eineZusage baut eine festgehaltene Zeile.
function eineZusage(anders: Partial<Handelszusage> = {}): Handelszusage {
  return {
    exemplarId: 7,
    menge: 2,
    fassung: 1,
    revision: 1,
    uebergeben: false,
    ...anders,
  }
}

// fuelleUndDruecke wählt den Handel, trägt die Felder für Fachleute ein und
// drückt einen Knopf. **Über die Felder und nicht über die Wahl** — die
// steht in `handelszusagenwahl.test.tsx`; hier geht es um den Weg von Hand.
async function fuelleUndDruecke(
  hand: Awaited<ReturnType<typeof zeigeHandel>>,
  vorgang: string,
  werte: Record<string, string>,
  knopf: string,
) {
  await oeffneHandel(hand, vorgang)

  // **Die Zusagen stehen in „Zusagen"** (TRADE-02, #1716).
  await zuPhase(hand, 'zusagen')

  const kasten = await screen.findByRole('region', {
    name: eigene.zusagenTitel,
  })

  const fachleute = within(kasten).getByRole('button', {
    name: eigene.zusagenFachleute,
  })

  if (fachleute.getAttribute('aria-expanded') !== 'true') {
    await hand.click(fachleute)
  }

  for (const [feld, wert] of Object.entries(werte)) {
    const eingabe = within(kasten).getByLabelText(feld)

    await hand.tripleClick(eingabe)
    await hand.paste(wert)
  }

  await hand.click(within(kasten).getByRole('button', { name: knopf }))
}

describe('Welche Exemplare weggehen', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('hält eine benannte Besitzzeile fest und zeigt sie', async () => {
    const gehalten = vi.fn(() =>
      Promise.resolve({
        vorgang: 'v-1',
        zusagen: [eineZusage()],
        fehler: null,
      }),
    )

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelFesthalten: gehalten,
    })

    await fuelleUndDruecke(
      await zeigeHandel(),
      'v-1',
      {
        [eigene.zusagenExemplar]: '7',
        [eigene.zusagenMenge]: '2',
        [eigene.zusagenFassung]: '1',
      },
      eigene.zusagenFesthalten,
    )

    expect(gehalten).toHaveBeenCalledWith('v-1', 7, 2, 1)

    expect(await within(imKasten()).findByText(/#7/)).toBeInTheDocument()
  })

  // **Festgehalten ist festgehalten** (§7, Regel 3; #1970): kam die Liste
  // danach nicht, sagt der Kasten das an ihrer Stelle — und nicht „nichts
  // festgehalten", was die leere Liste behauptete.
  it('nimmt ein Festhalten mit Nachladefehler als geschehen', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelFesthalten: () =>
        Promise.resolve({
          vorgang: 'v-1',
          zusagen: [],
          nachladefehler: {
            code: 'validierung' as const,
            schluessel: '',
            meldung: 'Die Liste klemmt',
          },
          fehler: null,
        }),
    })

    await fuelleUndDruecke(
      await zeigeHandel(),
      'v-1',
      {
        [eigene.zusagenExemplar]: '7',
        [eigene.zusagenMenge]: '2',
        [eigene.zusagenFassung]: '1',
      },
      eigene.zusagenFesthalten,
    )

    expect(
      await within(imKasten()).findByText('Die Liste klemmt'),
    ).toBeInTheDocument()
    expect(within(imKasten()).queryByText(eigene.zusagenLeer)).toBeNull()
    expect(within(imKasten()).getAllByRole('alert')).toHaveLength(1)
  })

  // **F03** (#2245):
  // bis dahin las `parseInt` „12a" als 12 und hielt damit ein **anderes**
  // Exemplar fest, als jemand getippt hatte; „2.5" Stück wurden zwei.
  it('weist eine Zahl, die keine ist, am Feld ab und hält nichts fest', async () => {
    const gehalten = vi.fn(() =>
      Promise.resolve({ vorgang: 'v-1', zusagen: [], fehler: null }),
    )

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelFesthalten: gehalten,
    })

    await fuelleUndDruecke(
      await zeigeHandel(),
      'v-1',
      {
        [eigene.zusagenExemplar]: '12a',
        [eigene.zusagenMenge]: '2.5',
        [eigene.zusagenFassung]: '1',
      },
      eigene.zusagenFesthalten,
    )

    const satz = einsetzen(kataloge.de.allgemein.ganzeZahlAb, {
      kleinstens: '0',
    })

    for (const feld of [eigene.zusagenExemplar, eigene.zusagenMenge]) {
      const eingabe = within(imKasten()).getByLabelText(feld)

      expect(eingabe).toHaveAttribute('aria-invalid', 'true')
      expect(eingabe).toHaveAccessibleDescription(satz)
    }

    expect(
      within(imKasten()).getByLabelText(eigene.zusagenFassung),
    ).not.toHaveAttribute('aria-invalid', 'true')
    expect(gehalten).not.toHaveBeenCalled()
  })

  // **Die Gegenprobe zum Nein**: dieselben Felder, richtig getippt, gehen
  // durch — und eine Null bleibt eine Null, über die Go urteilt.
  it('gibt eine richtig getippte Zahl weiter, auch eine Null', async () => {
    const gehalten = vi.fn(() =>
      Promise.resolve({ vorgang: 'v-1', zusagen: [], fehler: null }),
    )

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelFesthalten: gehalten,
    })

    await fuelleUndDruecke(
      await zeigeHandel(),
      'v-1',
      {
        [eigene.zusagenExemplar]: ' 12 ',
        [eigene.zusagenMenge]: '0',
        [eigene.zusagenFassung]: '3',
      },
      eigene.zusagenFesthalten,
    )

    expect(gehalten).toHaveBeenCalledWith('v-1', 12, 0, 3)
  })

  it('zeigt eine übergebene Zusage als übergeben', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Handelszusagen: () =>
        Promise.resolve({
          vorgang: 'v-1',
          zusagen: [eineZusage({ uebergeben: true })],
          fehler: null,
        }),
    })

    await fuelleUndDruecke(
      await zeigeHandel(),
      'v-1',
      {},
      eigene.zusagenNachsehen,
    )

    expect(
      await within(imKasten()).findByText(new RegExp(eigene.zusagenUebergeben)),
    ).toBeInTheDocument()
  })

  it('sagt, wenn nichts festgehalten ist', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Handelszusagen: () =>
        Promise.resolve({ vorgang: 'v-1', zusagen: [], fehler: null }),
    })

    await fuelleUndDruecke(
      await zeigeHandel(),
      'v-1',
      {},
      eigene.zusagenNachsehen,
    )

    expect(
      await within(imKasten()).findByText(eigene.zusagenLeer),
    ).toBeInTheDocument()
  })

  it('sagt ohne gewählten Handel, was zuerst kommt', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
    })

    await zuPhase(await zeigeHandel(), 'zusagen')

    const kasten = await screen.findByRole('region', {
      name: eigene.zusagenTitel,
    })

    expect(
      within(kasten).getByText(eigene.zusagenOhneVorgang),
    ).toBeInTheDocument()
    expect(
      within(kasten).queryByRole('button', { name: eigene.zusagenFesthalten }),
    ).toBeNull()
  })

  it('räumt die Liste, wenn der Ruf scheitert', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Handelszusagen: () =>
        Promise.resolve({
          vorgang: 'v-1',
          zusagen: [eineZusage()],
          fehler: null,
        }),
      HandelFesthalten: () => Promise.reject(new Error('geht nicht')),
    })

    const hand = await zeigeHandel()

    await fuelleUndDruecke(hand, 'v-1', {}, eigene.zusagenNachsehen)

    expect(await within(imKasten()).findByText(/#7/)).toBeInTheDocument()

    // **Mit einer Kennung**, damit der Ruf wirklich geht — ein leeres Feld
    // hält seit F03 am Feld an, statt als Exemplar null hinauszugehen.
    await fuelleUndDruecke(
      hand,
      'v-1',
      { [eigene.zusagenExemplar]: '7' },
      eigene.zusagenFesthalten,
    )

    expect(
      await within(imKasten()).findByText('geht nicht'),
    ).toBeInTheDocument()
    expect(within(imKasten()).queryByText(/#7/)).not.toBeInTheDocument()
  })
})
