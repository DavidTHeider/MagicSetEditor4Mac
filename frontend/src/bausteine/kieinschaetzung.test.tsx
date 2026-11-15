// **Die belegte Einschätzung im Editor** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass sichtbar ist, worauf sich eine Antwort stützt — und was daran nicht
// stimmt.** Das Issue verlangt beides: „sichtbare Belege öffnen den
// verwendeten Stand" und „niemals erfundenes Regelzitat". Ein Lauf, der nur
// den belegten Fall prüfte, ließe den zweiten offen.
//
// **Und dass zugeklappt nichts hinausgeht.** Das Aufschlagen einer Karte darf
// keinen Katalogruf auslösen (§16, Regel 6).
import { screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import {
  aufschlagenUndFragen,
  einBeleg,
  einBudget,
  eineEinschaetzung,
  eineFreigabe,
  einLokalerKatalog,
  hinweistexte as eigene,
  zeigeHinweiskasten as zeige,
} from '../test/kieinschaetzung'

afterEach(() => {
  raeumeApi()
})

// **Zugeklappt geht nichts hinaus.**
//
// **Mit [effekteNachholen] vor der Verneinung** (§3): der Ladehaken läuft in
// einem Effekt, und ein `not.toHaveBeenCalled()` davor sagte nur, dass React
// noch nicht so weit war. Ohne diese Zeile meldete der Lauf zusätzlich „an
// update … was not wrapped in act" — dasselbe Ereignis von der anderen Seite.
it('zugeklappt wird nichts gefragt', async () => {
  const modelle = vi.fn()
  const beurteilen = vi.fn()

  belegeApi({ KIModelle: modelle, KIEinschaetzen: beurteilen })

  zeige()

  await effekteNachholen()

  expect(modelle).not.toHaveBeenCalled()
  expect(beurteilen).not.toHaveBeenCalled()

  // **Der Beleg, dass jemand zuhört**: der Titel steht da, und der Lauf
  // darunter klickt ihn.
  expect(screen.getByText(eigene.kiEinschaetzungTitel)).toBeInTheDocument()
})

// **Eine belegte Einschätzung zeigt Satz, Vorbehalt, Belege und Stand.**
it('zeigt die Antwort mit ihren Belegen', async () => {
  const beurteilen = vi.fn(() => Promise.resolve(eineEinschaetzung()))

  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: beurteilen,
  })

  await aufschlagenUndFragen(zeige())

  expect(
    await screen.findByText('Der Block ist legal [B1].'),
  ).toBeInTheDocument()

  // **Der Vorbehalt steht dabei** — er kommt aus `domain` und reist mit.
  expect(screen.getByText(/keine amtliche Regelauskunft/)).toBeInTheDocument()

  expect(screen.getByText(eigene.kiEinschaetzungBelegt)).toBeInTheDocument()
  expect(screen.getByText('509.1 Blocks erklären')).toBeInTheDocument()
  expect(screen.getByText(eigene.kiEinschaetzungStaende)).toBeInTheDocument()
  expect(screen.getByText('Nur die mitgegebenen Stellen.')).toBeInTheDocument()

  // **Die Art steht im Ruf** — und der Nachweis ist leer, weil das Modell
  // auf diesem Rechner rechnet.
  expect(beurteilen).toHaveBeenCalledWith({
    art: 'regel',
    karteID: '4',
    anbieter: 'ollama',
    alias: 'gemma4:8b',
    frage: 'Wie wirkt sie beim Blocken?',
    format: '',
    freigegeben: [],
    nachweis: '',
  })
})

// **Eine unbelegte Antwort steht als solche da — und bleibt lesbar.**
//
// **Das ist der Satz, für den es diesen Kasten gibt.** Wer die erfundene
// Nummer stillschweigend striche, ließe den Rest als geprüft aussehen.
it('nennt, was in keinem Beleg steht — und zeigt den Satz trotzdem', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: () =>
      Promise.resolve(
        eineEinschaetzung({
          satz: 'Das steht in 702.19b und in [B1].',
          belegt: false,
          befunde: [
            {
              art: 'regelnummer',
              stelle: '702.19b',
              meldung: 'Diese Regelnummer steht in keinem Beleg.',
            },
          ],
        }),
      ),
  })

  await aufschlagenUndFragen(zeige())

  expect(
    await screen.findByText('Das steht in 702.19b und in [B1].'),
  ).toBeInTheDocument()

  expect(screen.getByRole('alert')).toHaveTextContent(
    eigene.kiEinschaetzungUnbelegt,
  )

  expect(
    screen.getByText(/702\.19b — Diese Regelnummer steht in keinem Beleg\./),
  ).toBeInTheDocument()
})

// **Eine Antwort ohne jede Fußnote nennt ihre Stelle nicht.**
//
// **Der Befund `ohnebeleg` trägt keine** — und dann steht nur die Meldung da,
// ohne einen Gedankenstrich davor.
it('zeigt einen Befund ohne Stelle als bloße Meldung', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: () =>
      Promise.resolve(
        eineEinschaetzung({
          satz: 'Dazu steht nichts in den Belegen.',
          belegt: false,
          befunde: [
            {
              art: 'ohnebeleg',
              stelle: '',
              meldung: 'Diese Einschätzung beruft sich auf keinen Beleg.',
            },
          ],
          einschraenkungen: [],
          belege: [],
          korpusstaende: [],
        }),
      ),
  })

  await aufschlagenUndFragen(zeige())

  expect(
    await screen.findByText('Diese Einschätzung beruft sich auf keinen Beleg.'),
  ).toBeInTheDocument()

  // **Und ohne Belege steht keine Belegliste da.**
  expect(
    screen.queryByText(eigene.kiEinschaetzungBelege),
  ).not.toBeInTheDocument()
  expect(
    screen.queryByText(eigene.kiEinschaetzungEinschraenkungen),
  ).not.toBeInTheDocument()
  expect(
    screen.queryByText(eigene.kiEinschaetzungStaende),
  ).not.toBeInTheDocument()
})

// **Ein Fehler steht als Fehler da.**
it('zeigt einen Fehler und keine Antwort', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: () =>
      Promise.resolve(
        eineEinschaetzung({
          // **Ohne `schluessel`** — dann zeigt das Fenster die Meldung, die
          // die Go-Seite geschrieben hat. Mit einem Schlüssel läge der Satz
          // im Sprachkatalog, und der Lauf prüfte dessen Eintrag statt des
          // Wegs (§9).
          fehler: {
            code: 'validierung',
            schluessel: '',
            meldung: 'kein Index',
          },
        }),
      ),
  })

  await aufschlagenUndFragen(zeige())

  expect(await screen.findByRole('alert')).toHaveTextContent('kein Index')
  // **Mit dem Weg zur Einrichtung** (AI-01.AK-04, #1742).
  expect(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.kieinrichtung.oeffnen,
    }),
  ).toBeInTheDocument()
  expect(screen.getByText(eigene.kiEinschaetzungLeer)).toBeInTheDocument()
})

// **Ein Modell, das Geld kostet, wird vorher freigegeben.**
it('fragt vor einem bezahlten Modell nach der Freigabe', async () => {
  const beurteilen = vi.fn(() => Promise.resolve(eineEinschaetzung()))

  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog('cloud')),
    Kostenstand: einBudget,
    KIFreigabeErteilen: eineFreigabe,
    KIEinschaetzen: beurteilen,
  })

  const hand = zeige()

  await aufschlagenUndFragen(hand, 'openai/gemma4:8b')

  // **Noch ist nichts hinausgegangen** — erst die Freigabe löst den Auftrag
  // aus (#1539).
  expect(beurteilen).not.toHaveBeenCalled()

  await hand.type(await screen.findByLabelText(eigene.kiFreigabeBetrag), '1200')
  await hand.click(
    screen.getByRole('button', { name: eigene.kiFreigabeErteilen }),
  )

  expect(
    await screen.findByText('Der Block ist legal [B1].'),
  ).toBeInTheDocument()

  expect(beurteilen).toHaveBeenCalledWith({
    art: 'regel',
    karteID: '4',
    anbieter: 'openai',
    alias: 'gemma4:8b',
    frage: 'Wie wirkt sie beim Blocken?',
    format: '',
    freigegeben: [],
    nachweis: 'nw-1',
  })
})

// **Und wer nicht freigibt, löst nichts aus.**
it('ohne Freigabe geht kein Auftrag hinaus', async () => {
  const beurteilen = vi.fn()

  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog('cloud')),
    Kostenstand: einBudget,
    KIEinschaetzen: beurteilen,
  })

  const hand = zeige()

  await aufschlagenUndFragen(hand, 'openai/gemma4:8b')

  await hand.click(
    await screen.findByRole('button', {
      name: eigene.kiFreigabeVerzichten,
    }),
  )

  expect(beurteilen).not.toHaveBeenCalled()
  expect(screen.getByText(eigene.kiEinschaetzungLeer)).toBeInTheDocument()
})

// **Die Balancefrage geht als solche hinaus.**
it('schickt die gewählte Art mit', async () => {
  const beurteilen = vi.fn(() =>
    Promise.resolve(eineEinschaetzung({ art: 'balance' })),
  )

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

  expect(await screen.findByText(/Der Block ist legal/)).toBeInTheDocument()
  expect(beurteilen).toHaveBeenCalledWith({
    art: 'balance',
    karteID: '4',
    anbieter: 'ollama',
    alias: 'gemma4:8b',
    frage: 'Wie steht sie da?',
    format: '',
    freigegeben: [],
    nachweis: '',
  })
})

// **Ohne Datum steht keines da — und der Punkt davor auch nicht.**
//
// **Zwei Stellen, eine Regel** (Beleg und Korpusstand): eine Quelle ohne
// Gültigkeitsdatum gibt es wirklich — das Regelwerk hat eines, ein
// Oracle-Text nicht. Ein „ · " ohne etwas dahinter sähe aus wie ein Datum,
// das verlorenging.
it('lässt ein fehlendes Datum ganz weg', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(einLokalerKatalog()),
    KIEinschaetzen: () =>
      Promise.resolve(
        eineEinschaetzung({
          belege: [einBeleg({ quelldatum: '', titel: 'Oracle: Blitzschlag' })],
          korpusstaende: [
            { quelle: 'oracle', revision: '2026-09-01', datum: '' },
          ],
        }),
      ),
  })

  await aufschlagenUndFragen(zeige())

  expect(await screen.findByText('Oracle: Blitzschlag')).toBeInTheDocument()
  expect(screen.getByText(/· regelwerk$/)).toBeInTheDocument()
  expect(screen.getByText('oracle · 2026-09-01')).toBeInTheDocument()
})

// **Und Zuklappen räumt den Inhalt ab.**
//
// **Der Katalog bleibt gefragt**: wer einmal aufgeklappt hat, soll beim
// zweiten Mal nicht noch einmal nach draußen fragen.
it('zugeklappt steht der Inhalt nicht mehr da', async () => {
  const modelle = vi.fn(() => Promise.resolve(einLokalerKatalog()))

  belegeApi({ KIModelle: modelle })

  const hand = zeige()

  await hand.click(screen.getByText(eigene.kiEinschaetzungTitel))
  expect(
    await screen.findByLabelText(eigene.kiEinschaetzungFrage),
  ).toBeInTheDocument()

  await hand.click(screen.getByText(eigene.kiEinschaetzungTitel))
  expect(
    screen.queryByLabelText(eigene.kiEinschaetzungFrage),
  ).not.toBeInTheDocument()

  await hand.click(screen.getByText(eigene.kiEinschaetzungTitel))
  await screen.findByLabelText(eigene.kiEinschaetzungFrage)

  expect(modelle).toHaveBeenCalledTimes(1)
})
