// **Ein Riegel je Handel, und eine späte Antwort ändert die Wahl nicht** —
// Befund Z4 des Zweitblicks über Phase 6.8
// (#1669).
//
// # Der Befund
//
// Kopieren, Löschen und Archivieren liefen unter einem **eigenen** Riegel
// neben den Schritten — zwei Schreibaufrufe an denselben Handel gingen
// gleichzeitig (§7, Regel 4). Und eine Antwort, die nach einem Wechsel der
// Wahl ankam, stellte die Wahl zurück — beim Kopieren auf die Kopie, bei
// Schritten und Vorschlägen auf den alten Handel (§7, Regel 2).
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { mitStandnachlauf } from '../test/standnachlauf'
import { kataloge } from '../sprache/kataloge'
import { eineListe } from './peerkontaktprobe'
import { einVorgang } from './handelslisteprobe'
import { bestaetigt, zeigeHandel, zuPhase } from './tauschprobe'

import { einJournal, einStand, eineFreigabe } from './buchungsprobe'

import type { Handelsvorgangantwort } from '../apihandelsleitung'
import type { Verbuchungsantwort } from '../apihandelsverbuchung'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

// zweiHandel stellt zwei Handel mit demselben Gegenüber bereit.
function zweiHandel(mehr: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi(
    mitStandnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: () =>
        Promise.resolve({
          vorgaenge: [
            einVorgang({ vorgang: 'h-1', revision: 1 }),
            einVorgang({ vorgang: 'h-2', revision: 2 }),
          ],
          fehler: null,
        }),
      ...mehr,
    }),
  )
}

// waehle nimmt einen Handel in „Mit wem".
async function waehle(
  hand: Awaited<ReturnType<typeof zeigeHandel>>,
  revision: number,
) {
  await zuPhase(hand, 'gegenueber')
  await hand.click(
    await screen.findByRole('radio', {
      name: new RegExp(`Revision ${String(revision)}`),
    }),
  )
}

// aufgeschoben ist ein Versprechen, das der Test selbst einlöst.
function aufgeschoben<T>() {
  let einloesen: (wert: T) => void = () => undefined
  const versprechen = new Promise<T>((fertig) => {
    einloesen = fertig
  })

  return { versprechen, einloesen }
}

// eintrag öffnet „Mit diesem Handel" und gibt einen Eintrag — **seit UXR-15
// stehen Kopieren, Löschen und Archiv in einem Menü** (#1923).
async function eintrag(
  hand: Awaited<ReturnType<typeof zeigeHandel>>,
  name: string,
) {
  await hand.click(
    await screen.findByRole('button', { name: eigene.liste.aktionen }),
  )

  return screen.findByRole('menuitem', { name })
}

it('hält Kopieren an, solange ein Schritt am Handel läuft', async () => {
  const schritt = aufgeschoben<Handelsvorgangantwort>()
  zweiHandel({ HandelSchritt: () => schritt.versprechen })

  const hand = await zeigeHandel()
  await waehle(hand, 1)

  const schritte = await screen.findByRole('group', {
    name: eigene.assistent.schritte,
  })
  await hand.click(
    within(schritte).getByRole('button', { name: eigene.schrittZustimmen }),
  )

  // **Derselbe Riegel**: solange der Schritt läuft, geht Kopieren nicht.
  expect(await eintrag(hand, eigene.liste.kopieren)).toHaveAttribute(
    'aria-disabled',
    'true',
  )
  await hand.keyboard('{Escape}')

  schritt.einloesen({ vorgang: einVorgang({ vorgang: 'h-1' }), fehler: null })
  await antwortAbwarten()
  expect(await eintrag(hand, eigene.liste.kopieren)).not.toHaveAttribute(
    'aria-disabled',
  )
})

it('lässt die Wahl, wenn die Kopie nach einem Wechsel ankommt', async () => {
  const kopie = aufgeschoben<Handelsvorgangantwort>()
  zweiHandel({ HandelKopieren: () => kopie.versprechen })

  const hand = await zeigeHandel()
  await waehle(hand, 1)
  await hand.click(await eintrag(hand, eigene.liste.kopieren))

  // **Jetzt wird gewechselt**, und danach kommt die Kopie an.
  await waehle(hand, 2)
  kopie.einloesen({
    vorgang: einVorgang({ vorgang: 'h-kopie', revision: 7 }),
    fehler: null,
  })

  await zuPhase(hand, 'gegenueber')
  expect(await screen.findByRole('radio', { name: /Revision 2/ })).toBeChecked()
  expect(screen.queryByText(eigene.liste.kopiert)).toBeNull()
})

it('lässt die Wahl, wenn ein Schritt nach einem Wechsel ankommt', async () => {
  const schritt = aufgeschoben<Handelsvorgangantwort>()
  zweiHandel({ HandelSchritt: () => schritt.versprechen })

  const hand = await zeigeHandel()
  await waehle(hand, 1)
  await hand.click(
    within(
      await screen.findByRole('group', { name: eigene.assistent.schritte }),
    ).getByRole('button', { name: eigene.schrittZustimmen }),
  )

  await waehle(hand, 2)
  schritt.einloesen({
    vorgang: einVorgang({ vorgang: 'h-1', zustand: 'angenommen' }),
    fehler: null,
  })

  await zuPhase(hand, 'gegenueber')
  expect(await screen.findByRole('radio', { name: /Revision 2/ })).toBeChecked()
})

it('lässt die Wahl, wenn das Löschen eines Entwurfs nach einem Wechsel ankommt', async () => {
  const geloescht = aufgeschoben<{ fehler: null }>()
  belegeApi(
    mitStandnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: () =>
        Promise.resolve({
          vorgaenge: [
            einVorgang({ vorgang: 'h-1', revision: 1, zustand: 'entwurf' }),
            einVorgang({ vorgang: 'h-2', revision: 2 }),
          ],
          fehler: null,
        }),
      HandelEntwurfLoeschen: () => geloescht.versprechen,
    }),
  )

  const hand = await zeigeHandel()
  await waehle(hand, 1)
  await hand.click(await eintrag(hand, eigene.liste.entwurfLoeschen))
  await hand.click(
    screen.getByRole('button', { name: eigene.liste.entwurfLoeschenJa }),
  )

  await waehle(hand, 2)
  geloescht.einloesen({ fehler: null })

  await zuPhase(hand, 'gegenueber')
  expect(await screen.findByRole('radio', { name: /Revision 2/ })).toBeChecked()
})

// **Auch Buchung und Zusagen hängen am Riegel des Raums** — Befund N3 des
// Nachlaufs über Phase 6.8: sie hatten je einen eigenen, und ein Schritt
// und ein Verbuchen konnten gleichzeitig starten.
//
// **Der Handel steht nach der Übergabe** — seit UXR-15 (#1923) steht
// „Verbuchen" erst in dieser Phase, und der Schritt, der läuft, ist die
// Klärung.
it('hält Verbuchen und Zusagen an, solange ein Schritt läuft', async () => {
  const schritt = aufgeschoben<Handelsvorgangantwort>()
  belegeApi(
    mitStandnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: () =>
        Promise.resolve({
          vorgaenge: [
            einVorgang({
              vorgang: 'h-1',
              zustand: 'uebergabeOffen',
              phase: 'verbuchen',
              naechster: 'verbuchen',
              moegliche: ['verbuchen', 'klaeren'],
            }),
          ],
          fehler: null,
        }),
      HandelSchritt: () => schritt.versprechen,
    }),
  )

  const hand = await zeigeHandel()
  await waehle(hand, 1)
  await hand.click(
    within(
      await screen.findByRole('group', { name: eigene.assistent.schritte }),
    ).getByRole('button', { name: eigene.schrittKlaeren }),
  )

  await zuPhase(hand, 'verbuchen')
  expect(
    screen.getByRole('button', { name: eigene.buchungVerbuchen }),
  ).toBeDisabled()

  await zuPhase(hand, 'zusagen')
  expect(
    screen.getByRole('button', { name: eigene.zusagenNachsehen }),
  ).toBeDisabled()

  schritt.einloesen({ vorgang: einVorgang({ vorgang: 'h-1' }), fehler: null })
})

// **Auch eine späte Buchung lässt den Stand des anderen** — UXR-15
// (#1923): nach einem Schritt der Buchung lädt der Raum nach und vergisst
// die Antwort von vorhin, aber nur für den Handel, aus dem gebucht wurde.
it('lässt den Stand des gewählten, wenn eine Buchung nach einem Wechsel ankommt', async () => {
  const verbucht = aufgeschoben<Verbuchungsantwort>()
  belegeApi(
    mitStandnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: () =>
        Promise.resolve({
          vorgaenge: [
            einVorgang({
              vorgang: 'h-1',
              zustand: 'uebergabeOffen',
              phase: 'verbuchen',
              naechster: 'verbuchen',
              moegliche: ['verbuchen', 'klaeren'],
            }),
            einVorgang({ vorgang: 'h-2', revision: 2 }),
          ],
          fehler: null,
        }),
      HandelSchritt: () =>
        Promise.resolve({
          vorgang: einVorgang({
            vorgang: 'h-2',
            revision: 2,
            zustand: 'angenommen',
            lage: 'wartet',
            naechster: '',
            moegliche: [],
          }),
          fehler: null,
        }),
      HandelFreigabeErteilen: eineFreigabe,
      Handelsverbuchung: () => verbucht.versprechen,
      Handelsjournal: () => Promise.resolve(einJournal()),
      // **Ein Zug legt die Übergabe vor dem Buchen in den Ausgang** (P13).
      HandelSchicken: () =>
        Promise.resolve({ nachrichtId: 'n-1', fehler: null }),
      HandelZustellen: () =>
        Promise.resolve({
          gegenueber: 'bgegenueber',
          getragen: 1,
          fehler: null,
        }),
    }),
  )

  const hand = await zeigeHandel()
  await waehle(hand, 2)
  await hand.click(
    within(
      await screen.findByRole('group', { name: eigene.assistent.schritte }),
    ).getByRole('button', { name: eigene.schrittZustimmen }),
  )
  expect(
    await screen.findByText('Wartet auf Fremder Rechner'),
  ).toBeInTheDocument()

  await waehle(hand, 1)
  await zuPhase(hand, 'verbuchen')
  await hand.click(
    await screen.findByRole('button', { name: eigene.buchungVerbuchen }),
  )

  await waehle(hand, 2)
  verbucht.einloesen(einStand())

  // **Die Antwort von vorhin gilt weiter** — ohne den Riegel stünde der
  // Stand aus der Liste da, „Du bist dran: zustimmen".
  await antwortAbwarten()
  expect(screen.getByText('Wartet auf Fremder Rechner')).toBeInTheDocument()
})
