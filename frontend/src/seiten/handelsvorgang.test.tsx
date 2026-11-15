// **Ein Handel von der Eröffnung bis zur Übergabe** —
// #1637.
//
// # Was hier gemessen wird
//
// **Dass eine Position mit der Seite hinausgeht, die gewählt wurde.** Die
// eigene Kennung kommt aus dem Ausweis und nicht aus einem Feld; ginge sie
// leer hinaus, wiese der Vertrag sie ab — und niemand wüsste, warum.
//
// **Dass ohne gewählten Handel nur das Eröffnen dasteht** — Stand,
// Vorschlag und Schritte gehören zu einem Handel, den es gibt (UX-08.AK-04:
// gewählt und nicht getippt).
//
// **Dass ein gescheiterter Ruf keinen Stand von vorhin stehen lässt** (§7,
// Regel 1): wer „Zustimmen" drückt und eine Absage bekommt, soll nicht die
// Revision von eben lesen.
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { mitStandnachlauf } from '../test/standnachlauf'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { derAusweis, einVorgang, meineKennung } from './vorgangsprobe'
import {
  keinAnhalt,
  kennungsfeld,
  oeffneHandel,
  oeffneZeile,
  zeigeHandel,
  zuPhase,
} from './tauschprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln

// belege belegt die Fassade samt dem Nachlesen des Stands, das ein Zug vor
// seinem ersten Schritt macht (Befund 1 des Zweitblicks über Phase 8).
const belege = (teile: Parameters<typeof belegeApi>[0]) =>
  belegeApi(mitStandnachlauf(teile))

// bestaetigt ist ein Kontakt, an dem der Kasten erscheint.
function bestaetigt() {
  return einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })
}

// trageEin wählt den Vorgang — ein leerer ist ein neuer Handel — und drückt
// den genannten Knopf.
async function trageEin(
  hand: Awaited<ReturnType<typeof zeigeHandel>>,
  vorgang: string,
  knopf: string,
) {
  if (vorgang !== '') {
    await oeffneHandel(hand, vorgang)
  }

  // **Der Vorgang steht in „Was tauscht ihr"** (TRADE-02, #1716).
  await zuPhase(hand, 'angebot')

  const raum = await screen.findByRole('region', { name: eigene.vorgangTitel })

  // **„Stand ansehen" steht seit UXR-15 im Kasten für Fachleute** (#1923) —
  // aufgeklappt wird er nur, wenn er zu ist.
  const fachleute = within(raum).queryByRole('button', {
    name: eigene.vorgangFachleute,
  })
  if (
    knopf === eigene.vorgangAnsehen &&
    fachleute?.getAttribute('aria-expanded') === 'false'
  ) {
    await hand.click(fachleute)
  }

  await hand.click(within(raum).getByRole('button', { name: knopf }))
}

describe('Ein Handel von der Eröffnung bis zur Übergabe', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('eröffnet mit der eigenen Kennung an der Position', async () => {
    const eroeffnet = vi.fn(() => Promise.resolve(einVorgang()))

    belege({
      ...keinAnhalt,
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Peeridentitaet: derAusweis,
      HandelEroeffnen: eroeffnet,
    })

    const hand = await zeigeHandel()

    // **Erst wenn der Ausweis da ist** — die eigene Spalte trägt seine
    // Kennung, und ohne sie ginge eine Position mit leerer Seite hinaus.
    await oeffneZeile(hand)

    const raum = screen.getByRole('region', { name: eigene.vorgangTitel })

    await hand.click(await kennungsfeld(hand, raum))
    await hand.paste('bolt-m10')

    await hand.click(
      within(raum).getByRole('button', {
        name: eigene.positionHinzufuegen,
      }),
    )

    // **Ein neuer Handel nennt keine Kennung** — die vergibt die Leitung.
    await trageEin(hand, '', eigene.vorgangEroeffnen)

    expect(eroeffnet).toHaveBeenCalledWith({
      vorgang: '',
      art: 'tausch',
      gegenueber: ['bgegenueber'],
      positionen: [
        {
          von: meineKennung,
          printing: 'bolt-m10',
          druckId: 0,
          menge: 1,
          guete: 'nearMint',
          sprache: 'de',
          finish: 'nonfoil',
          abbild: '',
          bewertung: 0,
        },
      ],
      bewertungGeteilt: false,
    })

    expect(
      await screen.findByText(
        new RegExp(`${eigene.buchungZustand}: ${eigene.buchungStandEntwurf}`),
      ),
    ).toBeInTheDocument()
  })

  it('geht einen Schritt und zeigt, wer zugestimmt hat', async () => {
    const geschritten = vi.fn(() =>
      Promise.resolve(
        einVorgang({
          zustand: 'angeboten',
          zustimmungen: [meineKennung],
          beschluss: 'zustimmen',
        }),
      ),
    )

    belege({
      ...keinAnhalt,
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Peeridentitaet: derAusweis,
      HandelSchritt: geschritten,
      // **Die Schritte stehen erst, wenn der Stand bekannt ist** (TRADE-02,
      // #1716): welche gehen, sagt er.
      HandelVorgaenge: () =>
        Promise.resolve({ vorgaenge: [einVorgang().vorgang], fehler: null }),
    })

    const hand = await zeigeHandel()
    await oeffneHandel(hand, 'v-1')

    // **Nur die Schritte, die gehen** — anbieten und abbrechen, der übliche
    // als Hauptknopf.
    const schritte = await screen.findByRole('group', {
      name: eigene.assistent.schritte,
    })
    expect(
      within(schritte)
        .getAllByRole('button')
        .map((knopf) => knopf.textContent),
    ).toEqual([eigene.zug.anbietenZustimmen, eigene.schrittAbbrechen])

    const anbieten = eigene.zug.anbietenZustimmen
    await hand.click(within(schritte).getByRole('button', { name: anbieten }))

    // **Mit dem Abdruck, den das Fenster zeigte** (#2496).
    expect(geschritten).toHaveBeenCalledWith(
      'v-1',
      'anbieten',
      '',
      'abcdef0123456789',
    )

    const raum = await screen.findByRole('region', {
      name: eigene.vorgangTitel,
    })

    expect(
      await within(raum).findByText(
        // **„du" und nicht die eigene Kennung** (UXR-15, #1923).
        new RegExp(`${eigene.vorgangZustimmungen}: ${eigene.vorgangDu}`),
      ),
    ).toBeInTheDocument()
  })

  it('schlägt eine neue Revision vor — erst, wenn die Menge eine ganze Zahl ist', async () => {
    const vorgeschlagen = vi.fn(() =>
      Promise.resolve(einVorgang({ revision: 2, zustand: 'verhandlung' })),
    )

    belege({
      ...keinAnhalt,
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Peeridentitaet: derAusweis,
      HandelVorschlagen: vorgeschlagen,
      // **Vorschlagen geht, wo ein Gegenangebot geht** (TRADE-02, #1716).
      HandelVorgaenge: () =>
        Promise.resolve({
          vorgaenge: [
            einVorgang({
              zustand: 'verhandlung',
              moegliche: ['gegenangebot', 'zustimmen', 'ablehnen'],
            }).vorgang,
          ],
          fehler: null,
        }),
    })

    const hand = await zeigeHandel()

    // **Erst wählen, dann zusammenstellen** — wer den Handel wechselt, räumt
    // die Zeilen mit (§7, Regel 1).
    await oeffneHandel(hand, 'v-1')

    await oeffneZeile(hand)

    const raum = screen.getByRole('region', { name: eigene.vorgangTitel })

    await hand.click(await kennungsfeld(hand, raum))
    await hand.paste('bolt-m10')

    // **Eine Menge, die keine ganze Zahl ab eins ist, bleibt am Feld** —
    // F03 (#2245).
    // Bis dahin wurde „viele" null und „2.5" ging als Kommazahl an ein `int`;
    // jetzt kommt die Zeile gar nicht erst in den Handel.
    const menge = within(raum).getByLabelText(eigene.positionMenge)
    const hinzufuegen = within(raum).getByRole('button', {
      name: eigene.positionHinzufuegen,
    })

    for (const roh of ['viele', '2.5', '0']) {
      await hand.clear(menge)
      await hand.click(menge)
      await hand.paste(roh)
      await hand.click(hinzufuegen)

      expect(menge).toHaveAttribute('aria-invalid', 'true')
      expect(menge).toHaveAccessibleDescription(
        einsetzen(kataloge.de.allgemein.ganzeZahlAb, { kleinstens: '1' }),
      )
    }

    await hand.clear(menge)
    await hand.click(menge)
    await hand.paste('2')
    await hand.click(hinzufuegen)

    await trageEin(hand, '', eigene.vorgangVorschlagen)

    expect(vorgeschlagen).toHaveBeenCalledWith('v-1', [
      {
        von: meineKennung,
        printing: 'bolt-m10',
        druckId: 0,
        menge: 2,
        guete: 'nearMint',
        sprache: 'de',
        finish: 'nonfoil',
        abbild: '',
        bewertung: 0,
      },
    ])
  })

  it('zeigt ohne gewählten Handel nur das Eröffnen', async () => {
    belege({
      ...keinAnhalt,
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Peeridentitaet: derAusweis,
    })

    await zeigeHandel()

    const raum = await screen.findByRole('region', {
      name: eigene.vorgangTitel,
    })

    expect(
      within(raum).getByRole('button', { name: eigene.vorgangEroeffnen }),
    ).toBeInTheDocument()
    expect(
      within(raum).queryByRole('button', { name: eigene.vorgangAnsehen }),
    ).toBeNull()
    expect(
      within(raum).queryByRole('group', { name: eigene.vorgangSchritte }),
    ).toBeNull()
  })

  // **Vom Vertrag der erste Satz, der Rest hinter dem Fragezeichen** —
  // UX-15, QA01 (#2382): seit *Handeln* einen Handel von selbst aufschlägt
  // (P13), steht der Absatz im ersten Blick, und er hatte 189 Zeichen.
  it('fasst die Erklärung des Vertrags kurz', async () => {
    belege({
      ...keinAnhalt,
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Peeridentitaet: derAusweis,
    })

    await zeigeHandel()

    const raum = await screen.findByRole('region', {
      name: eigene.vorgangTitel,
    })
    expect(within(raum).queryByText('kein Bestand')).toBeNull()
    expect(
      within(raum).getByRole('button', {
        name: einsetzen(kataloge.de.bausteine.mehrZu, {
          satz: 'Hier läuft der Vertrag: eröffnen, anbieten, zustimmen, vorbereiten, übergeben.',
        }),
      }),
    ).toBeInTheDocument()
  })

  it('lässt nach einer Absage keinen Stand von vorhin stehen', async () => {
    let ansehen = 0

    belege({
      ...keinAnhalt,
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Peeridentitaet: derAusweis,
      // **Das erste Ansehen glückt, das zweite nicht** — danach darf der
      // Stand des ersten nicht mehr dastehen.
      HandelStand: () => {
        ansehen += 1

        return ansehen === 1
          ? Promise.resolve(einVorgang({ ende: true }))
          : Promise.reject(new Error('geht nicht'))
      },
    })

    const hand = await zeigeHandel()

    await trageEin(hand, 'v-1', eigene.vorgangAnsehen)

    expect(await screen.findByText(eigene.vorgangEnde)).toBeInTheDocument()

    await trageEin(hand, 'v-1', eigene.vorgangAnsehen)

    const raum = screen.getByRole('region', { name: eigene.vorgangTitel })

    expect(await within(raum).findByRole('alert')).toBeInTheDocument()
    expect(within(raum).queryByText(eigene.vorgangEnde)).not.toBeInTheDocument()
  })

  it('nimmt eine Zeile wieder heraus', async () => {
    belege({
      ...keinAnhalt,
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Peeridentitaet: derAusweis,
    })

    const hand = await zeigeHandel()

    await oeffneZeile(hand)

    const raum = screen.getByRole('region', { name: eigene.vorgangTitel })

    await hand.click(await kennungsfeld(hand, raum))
    await hand.paste('bolt-m10')
    await hand.click(
      within(raum).getByRole('button', { name: eigene.positionHinzufuegen }),
    )

    expect(within(raum).getByText(/bolt-m10/)).toBeInTheDocument()

    await hand.click(
      within(raum).getByRole('button', {
        name: eigene.positionEntfernenName.replace('{nummer}', '1'),
      }),
    )

    expect(within(raum).getByText(eigene.positionLeer)).toBeInTheDocument()
  })
})
