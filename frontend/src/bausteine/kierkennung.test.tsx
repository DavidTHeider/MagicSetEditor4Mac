// **Ein Kartenfoto lesen lassen** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass der Weg steht und nichts angelegt wird.** Das Issue nennt es als
// eigene Abnahme: „Bild einer mehrdeutigen/falschen/keiner Karte erzeugt
// **keine** falsche Printing-Bindung oder Sammlungsmengen."
//
// **Und dass zugeklappt nichts hinausgeht** (§16, Regel 6).
import { screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { effekteNachholen } from '../test/effekte'
import {
  aufschlagenUndLesen,
  einBudget,
  einKatalog,
  eineErkennung,
  erkennungstexte as eigene,
  zeigeErkennungskasten as zeige,
} from '../test/kierkennung'

afterEach(() => {
  raeumeApi()
})

// **Zugeklappt geht nichts hinaus.**
//
// **Mit [effekteNachholen] vor der Verneinung** (§3): der Ladehaken läuft in
// einem Effekt, und ein `not.toHaveBeenCalled()` davor sagte nur, dass React
// noch nicht so weit war.
it('zugeklappt wird nichts gefragt', async () => {
  const modelle = vi.fn()
  const erkennen = vi.fn()

  belegeApi({ KIModelle: modelle, KIKarteErkennen: erkennen })

  zeige()

  await effekteNachholen()

  expect(modelle).not.toHaveBeenCalled()
  expect(erkennen).not.toHaveBeenCalled()

  // **Der Beleg, dass jemand zuhört**: der Titel steht da, und die Läufe
  // darunter klicken ihn.
  expect(screen.getByText(eigene.kiErkennungTitel)).toBeInTheDocument()
})

// **Der gewöhnliche Weg: lesen, zeigen, nichts anlegen.**
//
// **Der Pfad geht leer hinaus** — die Bindung fragt dann nach der Datei. Ein
// Pfad aus dem Fenster wäre einer, den niemand gewählt hat.
it('liest das Foto und zeigt, was daraus wurde', async () => {
  const erkennen = vi.fn(() => Promise.resolve(eineErkennung()))

  belegeApi({
    KIModelle: () => Promise.resolve(einKatalog()),
    KIKarteErkennen: erkennen,
  })

  await aufschlagenUndLesen(zeige())

  // **Zweimal**: als gelesene Stelle und als Name des gefundenen Drucks.
  expect(await screen.findAllByText('Blitzschlag')).toHaveLength(2)

  expect(erkennen).toHaveBeenCalledWith({
    pfad: '',
    anbieter: 'ollama',
    alias: 'gemma4:8b',
    nachweis: '',
  })

  // **Und der Satz, dass nichts aufgenommen wurde, steht dabei** — gerade
  // bei genau einem Treffer.
  expect(screen.getByText(eigene.kiErkennungEinDruck)).toBeInTheDocument()
  expect(screen.getByText(eigene.kiErkennungNichtsAngelegt)).toBeInTheDocument()

  // **Es gibt keinen Knopf, der übernähme.** Das ist die Abnahme des Issues,
  // und sie ist eine Aussage über die Abwesenheit.
  expect(
    screen.queryByRole('button', { name: /übernehmen|aufnehmen/i }),
  ).not.toBeInTheDocument()
})

// **Ein geschlossener Dialog ist keine Antwort.**
//
// # Warum das ein eigener Zustand ist
//
// **Weil eine leere Antwort sonst wie »kein Druck passt« aussieht.** Wer den
// Dateidialog schließt, hat nichts über sein Foto erfahren — und bekommt
// hier auch keine Aussage darüber.
it('sagt nach einem abgebrochenen Dialog nichts über das Bild', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(einKatalog()),
    KIKarteErkennen: () =>
      Promise.resolve(
        eineErkennung({
          beobachtungen: [],
          nichtGelesen: [],
          kandidaten: [],
          drucke: 0,
          eindeutig: false,
          alter: '',
          anbieter: '',
          modell: '',
          abgebrochen: true,
        }),
      ),
  })

  await aufschlagenUndLesen(zeige())

  expect(await screen.findByText(eigene.kiErkennungLeer)).toBeInTheDocument()

  expect(
    screen.queryByText(eigene.kiErkennungOhneKandidat),
  ).not.toBeInTheDocument()
})

// **Ein Fehler steht am Kasten und nicht als leere Liste da.**
//
// # Warum der gezeigte Satz der allgemeine ist
//
// **Weil übersetzt wird, was einen Schlüssel hat** (§9, #133). Die Absagen
// dieses Wegs kommen aus `domain.ErrValidation` ohne eigenen Schlüssel, und
// der Katalog antwortet dann mit seinem Satz für `unbekannt`. Das ist keine
// Lücke dieses Kastens, sondern die Regel — und der Lauf schreibt sie hin,
// statt einen deutschen Go-Satz zu erwarten, den eine englische Sitzung nie
// sähe.
it('zeigt den Fehler statt einer leeren Antwort', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(einKatalog()),
    KIKarteErkennen: () =>
      Promise.resolve(
        eineErkennung({
          fehler: {
            code: 'validierung',
            schluessel: 'unbekannt',
            meldung: 'openai liest keine Bilder.',
            feld: '',
          },
        }),
      ),
  })

  await aufschlagenUndLesen(zeige())

  expect(await screen.findByRole('alert')).toHaveTextContent(
    kataloge.de.fehler.unbekannt,
  )

  expect(screen.getByText(eigene.kiErkennungLeer)).toBeInTheDocument()
})

// **Ein bezahltes Modell wird vorher freigegeben** — dieselbe Regel wie bei
// jedem anderen Auftrag (0078).
it('fragt vor einem bezahlten Lauf nach der Freigabe', async () => {
  const erkennen = vi.fn(() => Promise.resolve(eineErkennung()))

  belegeApi({
    KIModelle: () => Promise.resolve(einKatalog('cloud')),
    Kostenstand: () => einBudget(),
    KIFreigabeErteilen: () =>
      Promise.resolve({
        kennung: 'n-1',
        laeuftab: '2026-09-14T10:05:00Z',
        fehler: null,
      }),
    KIKarteErkennen: erkennen,
  })

  const hand = zeige()

  await aufschlagenUndLesen(hand, 'openai/gemma4:8b')

  // **Erst der Beleg, dass der Kasten steht** — die Verneinung darunter
  // träfe sonst auch zu, wenn der Knopf gar nichts ausgelöst hätte (§3).
  const feld = await screen.findByLabelText(eigene.kiFreigabeBetrag)

  expect(erkennen).not.toHaveBeenCalled()

  await hand.type(feld, '1200')
  await hand.click(
    screen.getByRole('button', { name: eigene.kiFreigabeErteilen }),
  )

  // **Zweimal**: als gelesene Stelle und als Name des gefundenen Drucks.
  expect(await screen.findAllByText('Blitzschlag')).toHaveLength(2)

  expect(erkennen).toHaveBeenCalledWith({
    pfad: '',
    anbieter: 'openai',
    alias: 'gemma4:8b',
    nachweis: 'n-1',
  })
})

// **Und wer die Freigabe nicht gibt, löst nichts aus.**
it('geht ohne Freigabe nicht hinaus', async () => {
  const erkennen = vi.fn(() => Promise.resolve(eineErkennung()))

  belegeApi({
    KIModelle: () => Promise.resolve(einKatalog('cloud')),
    Kostenstand: () => einBudget(),
    KIKarteErkennen: erkennen,
  })

  const hand = zeige()

  await aufschlagenUndLesen(hand, 'openai/gemma4:8b')

  await hand.click(
    screen.getByRole('button', { name: eigene.kiFreigabeVerzichten }),
  )

  // **Der Beleg, dass der Kasten wirklich weg ist** — er ist die sichtbare
  // Wirkung, auf die sich die Verneinung stützt (§3).
  expect(
    screen.queryByLabelText(eigene.kiFreigabeBetrag),
  ).not.toBeInTheDocument()

  expect(erkennen).not.toHaveBeenCalled()
})

// **Zuklappen räumt ab und fragt nicht neu.**
//
// **Der Katalog hängt in [useKIModelle] und nicht im Kasten** — genau
// deshalb: ein Katalog, der im aufgeklappten Teil hinge, fiele beim Zuklappen
// aus dem DOM und fragte beim nächsten Aufschlagen erneut nach draußen (§16,
// Regel 6).
it('räumt beim Zuklappen ab, ohne erneut zu fragen', async () => {
  const modelle = vi.fn(() => Promise.resolve(einKatalog()))

  belegeApi({ KIModelle: modelle, KIKarteErkennen: vi.fn() })

  const hand = zeige()

  await hand.click(screen.getByText(eigene.kiErkennungTitel))

  expect(await screen.findByText(eigene.kiErkennungSatz)).toBeInTheDocument()

  await hand.click(screen.getByText(eigene.kiErkennungTitel))

  // **Die sichtbare Wirkung zuerst** — sie ist der Beleg, dass der zweite
  // Klick angekommen ist (§3).
  expect(screen.queryByText(eigene.kiErkennungSatz)).not.toBeInTheDocument()

  await effekteNachholen()

  expect(modelle).toHaveBeenCalledTimes(1)
})
