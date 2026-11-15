// **Die Vergleichskarten einer Balanceanalyse** — Werkstück A13.2b von Phase
// 5.3 (#1452).
//
// # Was hier gemessen wird
//
// **Dass die Strategie dasteht, bevor der Vergleich kommt** — sie beantwortet
// die Frage, die jemand vor den Karten stellt.
//
// **Dass ein Vergleich seine Unterschiede zeigt** — „ähnlich" allein sagt
// nichts, und das Issue verlangt die konkreten.
//
// **Und dass nirgends eine Zahl steht.** Es gibt kein Feld dafür; der Lauf
// hält, dass auch keines nachwächst.
import { screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import {
  aufschlagenUndFragen,
  eineBalanceantwort,
  eineEinschaetzung,
  einLokalerKatalog,
  hinweistexte as eigene,
  zeigeHinweiskasten as zeige,
} from '../test/kieinschaetzung'

afterEach(() => {
  raeumeApi()
})

// **Strategie, Format, Vergleich und Synergien stehen da.**
it('zeigt Strategie, Vergleich und was nicht betrachtet wurde', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: () => Promise.resolve(eineBalanceantwort()),
  })

  await aufschlagenUndFragen(zeige())

  expect(await screen.findByText(/Ähnlichkeitssuche/)).toBeInTheDocument()
  expect(screen.getByText(/Commander/)).toBeInTheDocument()
  expect(screen.getByText('Funkenwurf')).toBeInTheDocument()
  expect(screen.getByText('ein Mana billiger')).toBeInTheDocument()
  expect(screen.getByText(eigene.kiEinschaetzungSynergien)).toBeInTheDocument()
  expect(screen.getByText('Zusammenspiel mit Manaflut')).toBeInTheDocument()

  // **Und die beiden Überschriften sind zwei.**
  //
  // Sie trugen bis zum 2026-09-13 **denselben** Satz — „Was nicht betrachtet
  // wurde" stand einmal über den Einschränkungen der Antwort und einmal über
  // den ausgelassenen Synergien, in beiden Sprachen. Im Fenster las man ihn
  // zweimal mit verschiedenem Inhalt darunter; gezeigt hat es dieser Lauf,
  // weil `getByText` bei zwei Treffern scheitert.
  expect(eigene.kiEinschaetzungSynergien).not.toBe(
    eigene.kiEinschaetzungEinschraenkungen,
  )
})

// **Ohne Format steht da, dass keines genannt wurde.**
//
// **Und nicht ein leeres Feld**: ein Doppelpunkt, hinter dem nichts steht,
// sieht aus wie ein Fehler der Anzeige und nicht wie eine Auskunft.
it('sagt es, wenn kein Format genannt wurde', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: () => Promise.resolve(eineBalanceantwort({ format: '' })),
  })

  await aufschlagenUndFragen(zeige())

  expect(
    await screen.findByText(new RegExp(eigene.kiEinschaetzungOhneFormat)),
  ).toBeInTheDocument()
})

// **Ohne Vergleich steht der Satz, der sagt warum.**
it('sagt es, wenn keine Vergleichskarte genannt wurde', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: () =>
      Promise.resolve(
        eineBalanceantwort({ vergleiche: [], nichtBetrachtet: [] }),
      ),
  })

  await aufschlagenUndFragen(zeige())

  expect(
    await screen.findByText(eigene.kiEinschaetzungOhneVergleich),
  ).toBeInTheDocument()

  expect(
    screen.queryByText(eigene.kiEinschaetzungSynergien),
  ).not.toBeInTheDocument()
})

// **Und das Format geht mit, wenn es eines gibt.**
it('schickt das Format beim Vergleich mit', async () => {
  const beurteilen = vi.fn(() => Promise.resolve(eineBalanceantwort()))

  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: beurteilen,
  })

  const hand = zeige()

  await hand.click(screen.getByText(eigene.kiEinschaetzungTitel))
  await hand.selectOptions(
    await screen.findByLabelText(eigene.kiEinschaetzungArt),
    'balance',
  )
  await hand.type(
    screen.getByLabelText(eigene.kiEinschaetzungFormat),
    'Commander',
  )
  await hand.selectOptions(
    screen.getByLabelText(eigene.kiVorschlagAnbieter),
    'ollama/gemma4:8b',
  )
  await hand.type(
    screen.getByLabelText(eigene.kiEinschaetzungFrage),
    'Wie steht sie da?',
  )
  await hand.click(
    screen.getByRole('button', { name: eigene.kiEinschaetzungFragen }),
  )

  expect(await screen.findByText('Funkenwurf')).toBeInTheDocument()
  expect(beurteilen).toHaveBeenCalledWith({
    art: 'balance',
    karteID: '4',
    anbieter: 'ollama',
    alias: 'gemma4:8b',
    frage: 'Wie steht sie da?',
    format: 'Commander',
    freigegeben: [],
    nachweis: '',
  })
})

// **Eine Regelantwort zeigt keinen Vergleichsblock.**
//
// **Der Kopf stand hier über einem Rumpf, der das Gegenteil prüfte** — er
// schickte `eineBalanceantwort` hinein und sicherte zu, dass „Funkenwurf"
// dasteht. Grün war er, und über die Regelantwort sagte er nichts (§10).
it('zeigt bei einer Regelantwort keinen Vergleich', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: () => Promise.resolve(eineEinschaetzung()),
  })

  await aufschlagenUndFragen(zeige())

  // **Erst der Beleg, dass die Antwort da ist** — sonst spräche die
  // Verneinung darunter über einen Kasten, der noch gar nichts erhalten hat.
  expect(await screen.findByText(/Der Block ist legal/)).toBeInTheDocument()

  for (const satz of [
    eigene.kiEinschaetzungVergleiche,
    eigene.kiEinschaetzungStrategie,
    eigene.kiEinschaetzungSynergien,
  ]) {
    expect(screen.queryByText(new RegExp(satz))).not.toBeInTheDocument()
  }
})

// **Der eigene Bestand geht auch in einen Vergleich nur mit Kreuz** —
// Werkstück A13.4 von Phase 5.3
// (#1452).
//
// # Warum gerade hier
//
// **Weil eine Balanceanalyse über `aehnlich` sucht**, und der liest die
// eigenen Karten. Bis A13.4 tat er es von selbst: wer eine Karte vergleichen
// ließ, schickte dabei seine eigenen mit, ohne es zu sagen.
it('schickt die eigenen Karten erst nach dem Kreuz mit', async () => {
  const beurteilen = vi.fn(() => Promise.resolve(eineBalanceantwort()))

  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: beurteilen,
  })

  const hand = zeige()

  await hand.click(screen.getByText(eigene.kiEinschaetzungTitel))

  await hand.selectOptions(
    await screen.findByLabelText(eigene.kiEinschaetzungArt),
    'balance',
  )

  await hand.click(
    screen.getByLabelText(kataloge.de.seiten.recherche.freigabeEigeneKarten),
  )

  await hand.selectOptions(
    screen.getByLabelText(eigene.kiVorschlagAnbieter),
    'ollama/gemma4:8b',
  )
  await hand.type(
    screen.getByLabelText(eigene.kiEinschaetzungFrage),
    'Wie steht sie da?',
  )
  await hand.click(
    screen.getByRole('button', { name: eigene.kiEinschaetzungFragen }),
  )

  await screen.findByText('Sie kostet eins mehr als [B1].')

  expect(beurteilen).toHaveBeenCalledWith(
    expect.objectContaining({ freigegeben: ['eigenekarte'] }),
  )
})
