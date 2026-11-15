// **An einem fremden Tisch mitspielen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass der Tisch gemeldet und nicht abgefragt wird.** Der Gast weiß nicht,
// wann etwas geschieht; er erfährt es über die Wails-Laufzeit. Ein Lauf, der
// nur den Beitritt prüfte, ließe offen, ob je ein Bild ankommt.
//
// **Und dass der Abriss den Tisch räumt.** Ein alter Stand wäre einer, an dem
// inzwischen gespielt worden sein kann — schlimmer als keiner.
import { screen, waitFor, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { einePartie } from './partieprobe'
import { ereignisTischbild } from '../haken/usePartiegast'
import { kataloge } from '../sprache/kataloge'
import {
  derWirt,
  gastApi,
  laufzeitbelegen,
  waehleWirt,
  zeigeGast,
} from './partiegastprobe'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { kurzform } from '../sprache/kurzform'

const eigene = kataloge.de.seiten.spielen

// **Ohne Tisch steht kein Platz da, nur der Beitritt** — seit UXR-13
// (#1921). Bis dahin sagte ein Satz „Du sitzt an keinem fremden Tisch" über
// der Wahl; der Beitritt steht jetzt in einer Schublade, die nur aufgeht, wer
// sich setzen will, und der Platz erscheint erst mit einem Tisch.
it('zeigt ohne Tisch keinen Platz, nur den Beitritt', () => {
  gastApi({ PartieBeitreten: vi.fn() })

  zeigeGast()

  expect(
    screen.queryByRole('region', { name: eigene.gastPlatz }),
  ).not.toBeInTheDocument()
  expect(screen.getByText(kurzform(eigene.gastErklaerung))).toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: eigene.gastBeitreten }),
  ).toBeInTheDocument()
})

// **Beitreten beim gewählten Wirt — ohne eine Kennung zu tippen**
// (UX-08.AK-04, #1680): die Partie nennt der Wirt, und sie geht leer hinaus.
it('tritt beim gewählten Wirt bei, ohne die Partie zu nennen', async () => {
  const beigetreten = vi.fn().mockResolvedValue({
    laeuft: true,
    partie: '',
    gegenueber: 'abcd-ef01',
    fehler: null,
  })

  gastApi({ PartieBeitreten: beigetreten })

  const hand = zeigeGast()

  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  await waitFor(() => {
    expect(
      screen.getByText(eigene.gastSitzt, { exact: false }),
    ).toBeInTheDocument()
  })

  expect(beigetreten).toHaveBeenCalledWith('bwirt', '', {
    deckID: '',
    leiheID: '',
  })
})

// **Für einen Wirt vor 0.6.0 lässt sich die Partie nennen** — unter einem
// Kasten, der sagt, wann das nötig ist.
it('nennt die Partie selbst, wenn es einer braucht', async () => {
  const beigetreten = vi.fn().mockResolvedValue({
    laeuft: true,
    partie: 'p1',
    gegenueber: 'abcd-ef01',
    fehler: null,
  })

  gastApi({ PartieBeitreten: beigetreten })

  const hand = zeigeGast()

  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastFachleute }))

  expect(
    screen.getByText(kurzform(eigene.gastFachleuteErklaerung)),
  ).toBeInTheDocument()

  await hand.type(screen.getByLabelText(eigene.gastPartie), 'p1')
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  expect(beigetreten).toHaveBeenCalledWith('bwirt', 'p1', {
    deckID: '',
    leiheID: '',
  })
})

// **Zur Wahl steht nur, bei wem mitgespielt werden darf** — bestätigt und
// mit „Eine Partie spielen". Jeder andere Eintrag endete mit „darf nicht".
it('nennt nur Gegenüber, bei denen mitgespielt werden darf', async () => {
  gastApi({
    PartieBeitreten: vi.fn(),
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([
          derWirt,
          einKontakt({
            kennung: 'bhandel',
            eigenerName: 'Nur Handel',
            stand: 'bestaetigt',
            rechte: ['handel'],
          }),
          einKontakt({
            kennung: 'bfremd',
            eigenerName: 'Unbestätigt',
            rechte: ['partie'],
          }),
        ]),
      ),
  })

  zeigeGast()

  const wahl = await screen.findByRole('group', { name: eigene.gastWahl })

  expect(
    await within(wahl).findByRole('radio', { name: /^Wirt · 1111-2222/ }),
  ).toBeInTheDocument()
  expect(within(wahl).getAllByRole('radio')).toHaveLength(1)
})

// **Ist niemand da, sagt die Wahl, wie jemand dazukommt.**
it('erklärt eine leere Wahl', async () => {
  gastApi({
    PartieBeitreten: vi.fn(),
    Peerkontakte: () => Promise.resolve(eineListe([])),
  })

  zeigeGast()

  expect(await screen.findByText(eigene.gastWahlLeer)).toBeInTheDocument()
})

// **Kommen die Kontakte nicht, steht der nächste Versuch da.**
it('bietet Wiederholen, wenn die Kontakte nicht kommen', async () => {
  const Peerkontakte = vi
    .fn()
    .mockRejectedValueOnce(new Error('Bestand gesperrt'))
    .mockResolvedValue(eineListe([derWirt]))

  gastApi({ PartieBeitreten: vi.fn(), Peerkontakte })

  const hand = zeigeGast()

  await hand.click(
    await screen.findByRole('button', {
      name: kataloge.de.bausteine.zustaende.wiederholen,
    }),
  )

  await waehleWirt(hand)

  expect(
    screen.getByRole('button', { name: eigene.gastBeitreten }),
  ).toBeEnabled()
})

// **Ohne Gegenüber geht der Knopf nicht.**
//
// **Am Knopf und nicht erst an der Antwort** (§6): ein Fehler am Feld ist
// billiger als eine Absage aus Go.
it('hält den Knopf ohne Gegenüber an', () => {
  gastApi({ PartieBeitreten: vi.fn() })

  zeigeGast()

  expect(
    screen.getByRole('button', { name: eigene.gastBeitreten }),
  ).toBeDisabled()
})

// **Ein gemeldetes Tischbild wird gezeigt.**
//
// **Das ist die Zusage**: der Gast rechnet nichts, er bekommt die Sicht — und
// zeichnet sie mit demselben Bauteil wie einen eigenen Tisch.
it('zeigt den Tisch, den der Wirt gemeldet hat', async () => {
  const horcher = laufzeitbelegen()

  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      gegenueber: 'abcd',
      fehler: null,
    }),
  })

  const hand = zeigeGast()

  // **Ohne Bild und ohne Sitz steht kein Tisch da** — auch kein leerer.
  expect(
    screen.queryByRole('region', { name: eigene.gastPlatz }),
  ).not.toBeInTheDocument()

  const melde = horcher.get(ereignisTischbild)
  expect(melde).toBeDefined()

  // **Erst beitreten** (Z1, Zweitblick über 6.11): ein Bild, das ohne Platz
  // kommt, gehört nicht diesem Gast — das Fenster des Wirts bekommt dieselben
  // Ereignisse für seine eigene Partie.
  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))
  expect(
    await screen.findByText(eigene.gastSitzt, { exact: false }),
  ).toBeInTheDocument()

  await waitFor(() => {
    melde?.('p1', {
      partie: einePartie(),
      erlaubt: ['objekt.tappen'],
      handlungen: [],
    })
  })

  expect(
    await screen.findByRole('heading', { name: eigene.tischTitel }),
  ).toBeInTheDocument()
})

// **Ohne Laufzeit bricht nichts.**
//
// **Die Gegenprobe** (§10): unter `vite preview` und in jedem Lauf ohne
// Vorrichtung gibt es keine — und ein Fenster ohne Meldeweg zeigt eben keinen
// fremden Tisch, statt umzufallen.
it('kommt ohne Wails-Laufzeit aus', () => {
  ;(globalThis as { runtime?: unknown }).runtime = undefined

  gastApi({ PartieBeitreten: vi.fn() })

  zeigeGast()

  expect(
    screen.getByRole('button', { name: eigene.gastBeitreten }),
  ).toBeInTheDocument()
})

// **Aufstehen räumt den Tisch.**
it('räumt den Tisch beim Aufstehen', async () => {
  const horcher = laufzeitbelegen()

  const aufgestanden = vi.fn().mockResolvedValue({
    laeuft: false,
    partie: '',
    gegenueber: '',
    fehler: null,
  })

  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      gegenueber: 'abcd',
      fehler: null,
    }),
    PartieBeitrittBeenden: aufgestanden,
  })

  const hand = zeigeGast()

  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  await waitFor(() => {
    horcher.get(ereignisTischbild)?.('p1', {
      partie: einePartie(),
      erlaubt: [],
      handlungen: [],
    })
  })

  await screen.findByRole('heading', { name: eigene.tischTitel })

  await hand.click(screen.getByRole('button', { name: eigene.gastAufstehen }))

  // **Der Platz ist fort und der Beitritt wieder da** — seit UXR-13 steht
  // nach dem Aufstehen kein leerer Tisch mehr, sondern der Weg an den nächsten.
  await waitFor(() => {
    expect(
      screen.queryByRole('region', { name: eigene.gastPlatz }),
    ).not.toBeInTheDocument()
  })
  expect(
    screen.getByRole('button', { name: eigene.gastBeitreten }),
  ).toBeInTheDocument()

  expect(aufgestanden).toHaveBeenCalled()
})

// **Ein Fehler steht am Bauteil.**
it('meldet einen Fehler am Bauteil', async () => {
  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: false,
      partie: '',
      gegenueber: '',
      fehler: { code: 'validierung', meldung: 'Keine Rolle.', feld: '' },
    }),
  })

  const hand = zeigeGast()

  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  expect(await screen.findByRole('alert')).toHaveTextContent('Keine Rolle.')
})

// **Und am Platz steht er auch** — wer sitzt, sieht die Schublade nicht
// mehr; ein Nachsehen, das scheitert, meldet sich dort, wo geklickt wurde
// (§6). Der Platz bleibt: gescheitert ist das Nachsehen, nicht das Sitzen.
it('meldet einen Fehler am Platz', async () => {
  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      gegenueber: 'abcd',
      fehler: null,
    }),
    PartieBeitrittStand: vi.fn().mockResolvedValue({
      laeuft: false,
      partie: '',
      gegenueber: '',
      fehler: { code: 'intern', meldung: 'Die Leitung klemmt.', feld: '' },
    }),
  })

  const hand = zeigeGast()

  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  const platz = await screen.findByRole('region', { name: eigene.gastPlatz })

  await hand.click(
    within(platz).getByRole('button', { name: eigene.gastNachsehen }),
  )

  expect(await within(platz).findByRole('alert')).toHaveTextContent(
    'Die Leitung klemmt.',
  )
  expect(
    within(platz).getByRole('button', { name: eigene.gastAufstehen }),
  ).toBeInTheDocument()
})
