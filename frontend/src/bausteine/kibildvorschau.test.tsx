// **Was ein Bildvorschlag zeigt, bevor ihn jemand nimmt** — Werkstück A12
// von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Der Vorschaukasten und die zwei Wege aus ihm heraus.** Die Läufe über
// das Formular stehen in `kibild.test.tsx` daneben; hier steht, was danach
// kommt: was dasteht, was **nicht** dasteht, und was Verwerfen und
// Übernehmen tun.
//
// **Das Verwerfen steht gleichberechtigt neben dem Übernehmen.** Ein Lauf,
// der nur den Erfolgsweg prüfte, beantwortete die Frage des Issues nicht:
// „Verworfenes/abgebrochenes/abgelaufenes Bild lässt Karte unverändert."
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import {
  aufschlagenUndErzeugen,
  bildtexte as eigene,
  derBildnachweis,
  einKatalog,
  mitFreigabe,
  einVorschlag,
  ohneNachweis,
  zeigeBildkasten as zeige,
} from '../test/kibild'

afterEach(() => {
  raeumeApi()
})

// **Erzeugt steht der Vorschlag da — mit allem, was feststeht.**
it('ein erzeugtes Bild steht mit seinen Angaben da', async () => {
  const vorschlagen = vi.fn(() => Promise.resolve(einVorschlag()))

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: vorschlagen,
    }),
  )

  const { hand } = zeige()

  await aufschlagenUndErzeugen(hand)

  expect(await screen.findByRole('img')).toHaveAttribute(
    'src',
    'data:image/png;base64,AAAA',
  )

  // **Der Ruf trägt, was gewählt wurde** — Feld, Format und Güte gehen in
  // den Preis ein, und ein falscher Wert wäre ein Fehler, für den jemand
  // bezahlt hat.
  expect(vorschlagen).toHaveBeenCalledWith(
    '4',
    'image',
    'openai',
    'gpt-image-2.5-flare',
    'ein Blitz',
    '1024x1536',
    'medium',
    false,

    // **Und die Freigabe** (#1539): ohne sie weist die Kostenwache ab, und
    // ein Lauf ohne sie prüfte einen Weg, den es am Fenster nicht gibt.
    derBildnachweis,
  )

  // **Maße, Frist und Umschrift stehen dabei.**
  expect(screen.getByText(/1024 × 1536/)).toBeInTheDocument()
  expect(screen.getByText(/2026-09-12T10:00:00Z/)).toBeInTheDocument()
  expect(screen.getByText(/Ölgemälde/)).toBeInTheDocument()

  // **Und die Angaben mit ihrer Beschriftung.**
  //
  // **Im Vorschaukasten gesucht und nicht auf der Seite**: »Format« steht
  // auch als Beschriftung der Formatwahl da, und `getByText` fände beide.
  const kasten = within(screen.getByRole('figure'))

  expect(kasten.getByText(eigene.kiBildAngabeGroesse)).toBeInTheDocument()
  expect(kasten.getByText('1024x1536')).toBeInTheDocument()
})

// **Und verschwindet nicht**: ein Anbieter darf mehr nennen, als dieses
// Programm kennt, und eine verschluckte Auskunft ist eine verlorene (§16).
it('eine unbekannte Angabe steht roh da', async () => {
  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: () =>
        Promise.resolve(
          einVorschlag({ angaben: [{ name: 'seed', wert: '4711' }] }),
        ),
    }),
  )

  const { hand } = zeige()

  await aufschlagenUndErzeugen(hand)

  expect(await screen.findByText('seed')).toBeInTheDocument()
  expect(screen.getByText('4711')).toBeInTheDocument()
})

// **Ein leerer Satz »Gezeichnet wurde:« wäre eine Auskunft über nichts.**
it('ohne Umschrift steht keine da', async () => {
  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: () => Promise.resolve(einVorschlag({ umschrift: '' })),
    }),
  )

  const { hand } = zeige()

  await aufschlagenUndErzeugen(hand)

  await screen.findByRole('img')

  expect(screen.queryByText(/Gezeichnet wurde/)).not.toBeInTheDocument()
})

// **Verworfen ist weg — und der Vorrat erfährt es.**
it('ein verworfener Vorschlag ist weg', async () => {
  const verwerfen = vi.fn(() => Promise.resolve({ fehler: null }))

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: () => Promise.resolve(einVorschlag()),
      KIBildVerwerfen: verwerfen,
    }),
  )

  const { hand } = zeige()

  await aufschlagenUndErzeugen(hand)

  await hand.click(
    await screen.findByRole('button', { name: eigene.kiBildVerwerfen }),
  )

  expect(await screen.findByText(eigene.kiBildVerworfen)).toBeInTheDocument()
  expect(verwerfen).toHaveBeenCalledWith('bv-1')
  expect(screen.queryByRole('img')).not.toBeInTheDocument()
})

// **Weil sich ein Zuschnitt abbrechen lässt** (§7, Regel 3). Wer ihn hier
// schon wegräumte, hätte nach einem Abbruch nichts mehr, womit er es erneut
// versuchen könnte.
it('zuschneiden öffnet das vorhandene Fenster', async () => {
  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: () => Promise.resolve(einVorschlag()),
    }),
  )

  const zuschneiden = vi.fn(() => Promise.resolve(true))

  const { hand } = zeige(zuschneiden)

  await aufschlagenUndErzeugen(hand)

  await hand.click(
    await screen.findByRole('button', { name: eigene.kiBildUebernehmen }),
  )

  expect(zuschneiden).toHaveBeenCalledWith('image', 'bv-1')
  expect(screen.getByRole('img')).toBeInTheDocument()
})

// **Und die Meldung sagt, was los ist.** Ein Vorschlag, der auf einen
// Fehler hin verschwände, wäre für den, der es erneut versuchen will,
// unerreichbar — und im Vorrat läge er weiter.
it('ein gescheitertes Verwerfen lässt den Vorschlag stehen', async () => {
  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: () => Promise.resolve(einVorschlag()),
      KIBildVerwerfen: () =>
        Promise.reject(new Error('Der Vorrat antwortet nicht.')),
    }),
  )

  const { hand } = zeige()

  await aufschlagenUndErzeugen(hand)

  await hand.click(
    await screen.findByRole('button', { name: eigene.kiBildVerwerfen }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Der Vorrat antwortet nicht.',
  )

  expect(screen.getByRole('img')).toBeInTheDocument()
})
