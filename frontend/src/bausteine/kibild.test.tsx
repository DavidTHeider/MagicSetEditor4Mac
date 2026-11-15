// **Ein Modell erzeugt ein Kartenbild** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Dass zwischen »da ist eines« und »so steht es in der Karte« eine
// Entscheidung liegt.** Das Issue verlangt es: „Generierte oder bearbeitete
// Bilder bleiben Vorschläge, bis Nutzende sie mit vorhandener
// Bildverarbeitung übernehmen." Ein Lauf, der nur den Erfolgsweg prüfte,
// beantwortete die Frage nicht — deshalb steht das **Verwerfen** hier
// gleichberechtigt daneben.
//
// **Und dass nichts hinausgeht, was niemand wollte.** Zugeklappt wird kein
// Katalog gelesen und kein Bild erzeugt.
import { screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import {
  aufschlagenUndErzeugen,
  bildtexte as eigene,
  einKatalog,
  einVorschlag,
  gibFrei,
  mitFreigabe,
  derBildnachweis,
  ohneNachweis,
  zeigeBildkasten as zeige,
} from '../test/kibild'

afterEach(() => {
  raeumeApi()
})

// **Ohne Bildfeld gibt es den Kasten nicht.**
//
// **Und das ist kein Fehler**: eine Vorlage ohne Illustration ist eine
// Vorlage ohne Illustration.
it('ohne Bildfeld gibt es den Kasten nicht', () => {
  zeige(undefined, [])

  expect(screen.queryByText(eigene.kiBildTitel)).not.toBeInTheDocument()
})

// **Zugeklappt geht nichts hinaus.**
it('zugeklappt wird nichts gefragt', () => {
  const modelle = vi.fn()
  const vorschlagen = vi.fn()

  belegeApi(
    mitFreigabe({
      KIModelle: modelle,
      KIBildVorschlagen: vorschlagen,
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
    }),
  )

  zeige()

  expect(modelle).not.toHaveBeenCalled()
  expect(vorschlagen).not.toHaveBeenCalled()

  // **Der Beleg, dass jemand zuhört**: der Titel steht da, und der Lauf
  // darunter klickt ihn.
  expect(screen.getByText(eigene.kiBildTitel)).toBeInTheDocument()
})

// **Eine Angabe, die der Katalog nicht kennt, steht roh da.**
//
// **Ohne Umschrift steht keine da.**
//
// **»Zuschneiden und einsetzen« öffnet das vorhandene Fenster.**
//
// # Und der Vorschlag bleibt stehen
//
// **Ein Fehler steht als Fehler da.**
it('ein Fehler steht als Fehler da', async () => {
  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: () =>
        Promise.resolve(
          einVorschlag({
            fehler: {
              code: 'erlaubnis',
              schluessel: 'erlaubnis',
              meldung: '',
              feld: '',
              werte: {},
            },
          }),
        ),
    }),
  )

  const { hand } = zeige()

  await aufschlagenUndErzeugen(hand)

  expect(await screen.findByRole('alert')).toBeInTheDocument()
  expect(screen.queryByRole('img')).not.toBeInTheDocument()
})

// **Ohne Modell und ohne Aufgabe geht der Knopf nicht.**
//
// **Beide einzeln** — ein Lauf, der nur eines wegließe, ließe offen, ob das
// andere je geprüft wird.
it('ohne Modell oder Aufgabe bleibt der Knopf gesperrt', async () => {
  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
    }),
  )

  const { hand } = zeige()

  await hand.click(screen.getByText(eigene.kiBildTitel))

  const knopf = await screen.findByRole('button', {
    name: eigene.kiBildErzeugen,
  })

  expect(knopf).toBeDisabled()

  // **Nur das Modell reicht nicht.**
  await hand.selectOptions(
    screen.getByLabelText(eigene.kiVorschlagAnbieter),
    'openai/gpt-image-2.5-flare',
  )

  expect(knopf).toBeDisabled()

  // **Mit beidem geht er.**
  await hand.type(screen.getByLabelText(eigene.kiBildAufgabe), 'ein Blitz')

  expect(knopf).toBeEnabled()
})

// **Bearbeiten ist eine eigene Entscheidung.**
//
// Das Issue trennt Generierung und Bearbeitung; wer nichts anhakt, bekommt
// ein neues Bild und nicht ein verändertes.
it('bearbeiten reist als eigene Entscheidung mit', async () => {
  const vorschlagen = vi.fn(() => Promise.resolve(einVorschlag()))

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: vorschlagen,
    }),
  )

  const { hand } = zeige()

  await hand.click(screen.getByText(eigene.kiBildTitel))

  await hand.selectOptions(
    await screen.findByLabelText(eigene.kiVorschlagAnbieter),
    'openai/gpt-image-2.5-flare',
  )

  await hand.type(screen.getByLabelText(eigene.kiBildAufgabe), 'dunkler')
  await hand.click(screen.getByLabelText(eigene.kiBildBearbeiten))
  await hand.selectOptions(screen.getByLabelText(eigene.kiBildGuete), 'high')
  await hand.selectOptions(
    screen.getByLabelText(eigene.kiBildGroesse),
    '1024x1024',
  )

  await hand.click(screen.getByRole('button', { name: eigene.kiBildErzeugen }))

  await gibFrei(hand)

  expect(vorschlagen).toHaveBeenCalledWith(
    '4',
    'image',
    'openai',
    'gpt-image-2.5-flare',
    'dunkler',
    '1024x1024',
    'high',
    true,
    derBildnachweis,
  )
})

// **Scheitert das Verwerfen, bleibt der Vorschlag stehen.**
//
// **Bei zwei Bildfeldern wird gewählt, in welches es geht.**
//
// **Sonst landete jedes erzeugte Bild im ersten** — und eine Karte mit
// Illustration und Wasserzeichen hat zwei.
it('bei zwei Bildfeldern wird gewählt', async () => {
  const vorschlagen = vi.fn(() => Promise.resolve(einVorschlag()))

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: vorschlagen,
    }),
  )

  const { hand } = zeige(undefined, ['image', 'watermark'])

  await hand.click(screen.getByText(eigene.kiBildTitel))

  await hand.selectOptions(
    await screen.findByLabelText(eigene.kiVorschlagAnbieter),
    'openai/gpt-image-2.5-flare',
  )

  await hand.selectOptions(
    screen.getByLabelText(eigene.kiBildFeld),
    'watermark',
  )

  await hand.type(screen.getByLabelText(eigene.kiBildAufgabe), 'ein Zeichen')
  await hand.click(screen.getByRole('button', { name: eigene.kiBildErzeugen }))

  await gibFrei(hand)

  await screen.findByRole('figure')

  expect(vorschlagen).toHaveBeenCalledWith(
    '4',
    'watermark',
    'openai',
    'gpt-image-2.5-flare',
    'ein Zeichen',
    '1024x1536',
    'medium',
    false,
    derBildnachweis,
  )
})

// **Zweimal aufklappen fragt nicht zweimal.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil `KIModelle` nach draußen geht.** Ein Katalog, der beim zweiten
// Aufklappen erneut gelesen würde, wäre ein Ruf, den niemand ausgelöst hat
// (§16, Regel 6) — und bei einem Cloudanbieter kostet er Zeit und Aufsehen.
//
// **Genau das ist beim ersten Anlauf passiert**, als die Modellwahl ihren
// Katalog selbst hielt: sie fällt beim Zuklappen aus dem DOM, und mit ihr
// fiel die Erinnerung, dass schon gefragt wurde. Seitdem hängt er in
// [useKIModelle] beim Aufrufer.
it('zweimal aufklappen fragt nicht zweimal', async () => {
  const modelle = vi.fn(() => Promise.resolve(einKatalog()))

  belegeApi(
    mitFreigabe({
      KIModelle: modelle,
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
    }),
  )

  const { hand } = zeige()

  await hand.click(screen.getByText(eigene.kiBildTitel))

  await screen.findByLabelText(eigene.kiVorschlagAnbieter)

  await hand.click(screen.getByText(eigene.kiBildTitel))
  await hand.click(screen.getByText(eigene.kiBildTitel))

  expect(await screen.findByLabelText(eigene.kiVorschlagAnbieter)).toBeVisible()
  expect(modelle).toHaveBeenCalledTimes(1)
})
