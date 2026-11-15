// **Die Güten am Tisch** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass die gewählte Karte in höherer Güte nachgeladen wird** — „Hohe
// Auflösung bei Inspektion/Zoom nachladen", und erst **nach** der Wahl.
//
// **Und dass bis dahin die vorgeladene Güte dasteht.** Ein leeres Kästchen,
// das sich nach einer Sekunde füllt, wäre schlechter als ein kleineres Bild,
// das sofort da ist.
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'
import nutzerereignis from '@testing-library/user-event'

import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { einLauf, einObjekt, einePartie } from './partieprobe'
import { effekteNachholen } from '../test/effekte'
import { zeige } from './partietischprobe'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// **Die gewählte Karte wird in der Betrachtungsgüte nachgeladen.**
//
// # Der Abnahmepunkt, für den es diesen Lauf gibt
//
// **„Hohe Auflösung bei Inspektion/Zoom nachladen"** (#1459). Am Tisch liegt
// jede Karte in `tisch`; wer eine wählt, sieht sie daneben klein, und wer
// sie **groß ansieht**, bekommt die Betrachtungsgüte — erst dann wird dafür
// gezeichnet. Bis zum 2026-09-24 genügte schon die Wahl; seitdem steht die
// Karte klein da wie im Artboard *Tisch* (UX-08, #1680).
//
// **Und bis sie da ist, steht die Tischgüte da.** Ein leeres Kästchen, das
// sich nach einer Sekunde füllt, wäre schlechter als ein kleineres Bild,
// das sofort da ist.
it('lädt die gewählte Karte in höherer Güte nach', async () => {
  const gefragt: string[][] = []

  belegeApi({
    Spieltextur: (karteID: string, seite: string, guete: string) => {
      gefragt.push([karteID, seite, guete])

      return Promise.resolve({
        textur: {
          schluessel: `k-${guete}`,
          adresse: `/spielbild/k-${guete}.png`,
          guete,
          seite,
          breite: 512,
          hoehe: 714,
          befunde: [],
          zweiseitig: false,
        },
        fehler: null,
      })
    },
    Rueckentextur: () =>
      Promise.resolve({
        textur: {
          schluessel: 'r',
          adresse: '/spielbild/r.png',
          guete: 'tisch-niedrig',
          seite: 'ruecken',
          breite: 512,
          hoehe: 714,
          befunde: [],
          zweiseitig: false,
        },
        fehler: null,
      }),
  })

  zeige(
    einLauf(),
    einePartie({
      objekte: [
        einObjekt({
          zone: 'tisch',
          name: 'Wald',
          geheim: false,
          karteID: '77',
        }),
      ],
    }),
  )

  await waitFor(() => {
    expect(gefragt).toContainEqual(['77', 'vorne', 'tisch'])
  })

  // **Vor der Wahl wird nichts in der Betrachtungsgüte geholt.**
  expect(gefragt).not.toContainEqual(['77', 'vorne', 'betrachten'])

  // **Erst die Effekte nachholen, dann klicken** (#761): `waitFor` sagt zu,
  // dass der Ruf geschehen ist — nicht, dass Reacts passive Effekte gelaufen
  // sind. Ein Klick auf einen Baum, den ein Effekt gerade erst aufbaut, geht
  // unter Last ins Leere, und `waitFor` danach heilt das nicht: es
  // wiederholt die Zusicherung, nicht das Ereignis.
  await effekteNachholen()

  const hand = nutzerereignis.setup()

  await hand.click(screen.getByRole('button', { name: /Wald/ }))

  // **Gewählt ist noch nicht groß** — und dass die Tafel dasteht, belegt
  // ihr Knopf, bevor die Verneinung etwas sagt (§3).
  const gross = await screen.findByRole('button', { name: eigene.grossAnsehen })
  await effekteNachholen()

  expect(gefragt).not.toContainEqual(['77', 'vorne', 'betrachten'])

  await hand.click(gross)

  await waitFor(() => {
    expect(gefragt).toContainEqual(['77', 'vorne', 'betrachten'])
  })

  // **Und das große Bild steht am Ende da.**
  await waitFor(() => {
    expect(
      document.querySelector('img[src="/spielbild/k-betrachten.png"]'),
    ).not.toBeNull()
  })
})

// **Solange die volle Güte fehlt, steht die vorgeladene da.**
//
// # Wofür dieser Lauf da ist
//
// **Weil die Ausweichgüte sonst ungeprüft wäre.** Gemessen am 2026-09-15:
// ein Eingriff, der sie ausbaut, ließ die ganze Seitensuite grün — jeder
// andere Lauf beantwortet beide Güten sofort, und dann ist der Ausweich nie
// nötig. Eine Gegenprobe, die nichts ändert, ist keine (§10).
//
// **Hier antwortet die Tischgüte gar nicht**, und genau das ist der Fall im
// Betrieb: sechzig Karten werden nacheinander gezeichnet, und bis dahin
// soll nicht ein leeres Kästchen dastehen.
it('zeigt die vorgeladene Güte, solange die volle fehlt', async () => {
  belegeApi({
    Spieltextur: (_karteID: string, seite: string, guete: string) => {
      // **Die volle Güte kommt nie** — sie wird gerade gezeichnet.
      if (guete === 'tisch') {
        return new Promise<never>(() => undefined)
      }

      return Promise.resolve({
        textur: {
          schluessel: `k-${guete}`,
          adresse: `/spielbild/k-${guete}.png`,
          guete,
          seite,
          breite: 256,
          hoehe: 357,
          befunde: [],
          zweiseitig: false,
        },
        fehler: null,
      })
    },
    Rueckentextur: () =>
      Promise.resolve({
        textur: {
          schluessel: 'r',
          adresse: '/spielbild/r.png',
          guete: 'tisch-niedrig',
          seite: 'ruecken',
          breite: 256,
          hoehe: 357,
          befunde: [],
          zweiseitig: false,
        },
        fehler: null,
      }),
  })

  zeige(
    einLauf(),
    einePartie({
      vorrat: ['77'],
      objekte: [
        einObjekt({
          zone: 'tisch',
          name: 'Wald',
          geheim: false,
          karteID: '77',
        }),
      ],
    }),
  )

  await waitFor(() => {
    expect(document.querySelector('img')).toHaveAttribute(
      'src',
      '/spielbild/k-tisch-niedrig.png',
    )
  })
})

// **Die zweite Seite gibt es nur, wo die Karte eine hat — und sie wird
// gezeigt, wenn jemand sie verlangt.**
//
// # Der Abnahmepunkt, für den es diesen Lauf gibt
//
// **„Mehrseitige Karte wechselt korrekt zwischen Faces"** (#1459). Der erste
// Anlauf band das an `verdeckt` — und das ist etwas anderes: `verdeckt`
// heißt, die Karte **liegt** umgedreht, `objekt.umdrehen` setzt genau das.
// Aufgefallen ist die Verwechslung an einem unbetretenen Zweig in der
// Abdeckung: die Bedingung hatte keinen Fall, weil es im Fenster keinen Weg
// dorthin gab (§3).
it('bietet die andere Seite nur bei einer zweiseitigen Karte', async () => {
  const gefragt: string[][] = []

  belegeApi({
    Spieltextur: (karteID: string, seite: string, guete: string) => {
      gefragt.push([karteID, seite, guete])

      return Promise.resolve({
        textur: {
          schluessel: `k-${seite}-${guete}`,
          adresse: `/spielbild/k-${seite}-${guete}.png`,
          guete,
          seite,
          breite: 512,
          hoehe: 714,
          befunde: [],
          zweiseitig: karteID === '77',
        },
        fehler: null,
      })
    },
    Rueckentextur: () =>
      Promise.resolve({
        textur: {
          schluessel: 'r',
          adresse: '/spielbild/r.png',
          guete: 'tisch-niedrig',
          seite: 'ruecken',
          breite: 512,
          hoehe: 714,
          befunde: [],
          zweiseitig: false,
        },
        fehler: null,
      }),
  })

  zeige(
    einLauf(),
    einePartie({
      objekte: [
        einObjekt({
          zone: 'tisch',
          name: 'Delver',
          geheim: false,
          karteID: '77',
        }),
        einObjekt({
          kennung: 'o2',
          zone: 'tisch',
          name: 'Wald',
          geheim: false,
          karteID: '78',
        }),
      ],
    }),
  )

  await waitFor(() => {
    expect(gefragt).toContainEqual(['77', 'vorne', 'tisch'])
  })

  await effekteNachholen()

  const hand = nutzerereignis.setup()

  // **Die einseitige Karte bietet nichts an** — groß angesehen, und das
  // bleibt für die nächste Karte so.
  await hand.click(screen.getByRole('button', { name: /Wald/ }))
  await hand.click(
    await screen.findByRole('button', { name: eigene.grossAnsehen }),
  )

  await waitFor(() => {
    expect(gefragt).toContainEqual(['78', 'vorne', 'betrachten'])
  })

  expect(screen.queryByRole('button', { name: 'Andere Seite' })).toBeNull()

  // **Die zweiseitige schon.**
  await hand.click(screen.getByRole('button', { name: /Delver/ }))

  const umdrehen = await screen.findByRole('button', { name: 'Andere Seite' })

  await hand.click(umdrehen)

  await waitFor(() => {
    expect(gefragt).toContainEqual(['77', 'zweite', 'betrachten'])
  })

  // **Und zurück geht es auch** — nicht nur die Aufschrift wechselt.
  await hand.click(await screen.findByRole('button', { name: 'Erste Seite' }))

  await screen.findByRole('button', { name: 'Andere Seite' })

  await waitFor(() => {
    expect(
      document.querySelector('img[src="/spielbild/k-vorne-betrachten.png"]'),
    ).not.toBeNull()
  })
})
