// **Die Bilder am Tisch** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass eine Adresse ankommt und keine Bytes.** 182 KB kosten als Base64
// über die Wails-Brücke 554 ms, und ein Tisch zeigt sechzig Karten.
//
// **Dass eine verdeckte Karte den Rücken ihres Spiels zeigt** — ohne
// Kartenkennung, und ohne dass ihre Maße etwas über sie verraten.
//
// **Dass vorgeladen wird**, was sicher gebraucht wird: der Rücken und das
// eigene Deck.
//
// **Und dass eine fehlende Ressource eine sichtbare Diagnose hat.**
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einLauf, einObjekt, einePartie } from './partieprobe'
import { zeige } from './partietischprobe'

afterEach(() => {
  raeumeApi()
})

// **Eine offene Karte bekommt ihr Bild.**
//
// # Die Abnahme aus #1459 am Fenster
//
// **Eine Adresse und keine Bytes**: was ankommt, ist ein `src` — und der
// Browser holt es über dieselbe Leitung wie jede andere Datei. Ein
// `data:`-Bild hier wäre die Gestalt, die für die Vorschau gemessen 554 ms
// gekostet hat.
it('zeigt für eine offene Karte ihr Bild', async () => {
  belegeApi({
    Spieltextur: () =>
      Promise.resolve({
        textur: {
          schluessel: 'abc',
          adresse: '/spielbild/abc.png',
          guete: 'tisch',
          seite: 'vorne',
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
    expect(document.querySelector('img')).toHaveAttribute(
      'src',
      '/spielbild/abc.png',
    )
  })

  // **Und das Bild ist für ein Vorleseprogramm nicht da** — der Knopf daneben
  // nennt die Karte schon, und ein `alt` mit demselben Text sagte sie zweimal.
  expect(document.querySelector('img')).toHaveAttribute('alt', '')
})

// **Und was beim Zeichnen klagt, steht am Tisch.**
//
// # Warum hier **zwei** Wege belegt sind
//
// **Weil der Tisch zwei Bilder anfordert** — die Karte und den Rücken (er
// wird vorgeladen). Blieb der Rücken unbelegt, gab es **zwei** Fehlerquellen,
// und `texturen.fehler` hält nur eine: welche dasteht, entschied dann die
// Reihenfolge der Antworten. Allein lief der Lauf grün, unter Last in
// `make ci` fiel er um — gefunden am 2026-09-16 mit G3.
it('meldet, was beim Zeichnen schiefging', async () => {
  belegeApi({
    Spieltextur: () => Promise.reject(new Error('Der Stil fehlt.')),
    Rueckentextur: () =>
      Promise.resolve({
        textur: {
          schluessel: 'ruecken',
          adresse: '/spielbild/ruecken.png',
          guete: 'tisch',
          seite: 'vorne',
          breite: 375,
          hoehe: 523,
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

  // **Auf diesen Satz warten und nicht auf „einen Alert"** (§3): eine
  // Zusicherung über die Rolle allein trifft auch die Meldung, die zufällig
  // zuerst da ist.
  expect(await screen.findByText(/Der Stil fehlt\./)).toBeInTheDocument()
})

// **Eine verdeckte Karte zeigt den Rücken ihres Spiels.**
//
// # Der Satz, für den es diesen Lauf gibt
//
// **Die Projektion gibt für ein geheimes Objekt keine Kennung heraus**, und
// das ist Absicht (G1): eine Bibliothek ist auch für ihren Eigner verdeckt.
// Der Rücken hängt deshalb am **Spiel** und geht seinen eigenen Weg.
//
// **Und er ist für alle derselbe**: zwei verdeckte Karten, ein Ruf, ein
// Bild.
it('zeigt für eine verdeckte Karte den Rücken', async () => {
  const gefragt: string[][] = []

  belegeApi({
    Spieltextur: () =>
      Promise.reject(new Error('hier wird keine Karte gefragt')),
    Rueckentextur: (spiel: string, guete: string) => {
      gefragt.push([spiel, guete])

      return Promise.resolve({
        textur: {
          schluessel: 'ruecken1',
          adresse: `/spielbild/ruecken-${guete}.png`,
          guete,
          seite: 'ruecken',
          breite: 512,
          hoehe: 714,
          befunde: [],
          zweiseitig: false,
        },
        fehler: null,
      })
    },
  })

  zeige(
    einLauf(),
    einePartie({
      spiel: 'magic',
      objekte: [
        einObjekt({ zone: 'tisch', name: '', geheim: true, karteID: '' }),
        einObjekt({
          kennung: 'o2',
          zone: 'tisch',
          name: '',
          geheim: true,
          karteID: '',
        }),
      ],
    }),
  )

  await waitFor(() => {
    expect(document.querySelector('img')).toHaveAttribute(
      'src',
      '/spielbild/ruecken-tisch.png',
    )
  })

  // **Zwei verdeckte Karten, ein Ruf je Güte.**
  //
  // **Zwei Güten, weil vorgeladen wird**: `tisch-niedrig` beim Öffnen,
  // `tisch` für das, was wirklich daliegt.
  expect(gefragt.map(([spiel]) => spiel)).toEqual(['magic', 'magic'])
  expect(new Set(gefragt.map(([, guete]) => guete))).toEqual(
    new Set(['tisch-niedrig', 'tisch']),
  )
})

// **Das eigene Deck wird vorgeladen.**
//
// **Das Issue verlangt es wörtlich**: „Eigene Decks, Rückseite … vorladen."
// Geholt wird in der **niedrigen** Tischgüte — der Stufe, die es genau dafür
// gibt.
it('lädt den Rücken und das eigene Deck vor', async () => {
  const gefragt: string[][] = []

  belegeApi({
    Spieltextur: (karteID: string, seite: string, guete: string) => {
      gefragt.push([karteID, seite, guete])

      return Promise.resolve({
        textur: {
          schluessel: `k${karteID}-${guete}`,
          adresse: `/spielbild/k${karteID}-${guete}.png`,
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
      vorrat: ['11', '12'],
      objekte: [einObjekt({ zone: 'bibliothek', geheim: true, karteID: '' })],
    }),
  )

  await waitFor(() => {
    expect(gefragt).toHaveLength(2)
  })

  expect(gefragt).toEqual([
    ['11', 'vorne', 'tisch-niedrig'],
    ['12', 'vorne', 'tisch-niedrig'],
  ])
})

// **Und was beim Zeichnen auffiel, steht am Tisch.**
//
// **Eine fehlende Ressource hat eine sichtbare Diagnose** (#1459). Sie steht
// als `status` und nicht als `alert`: die Karte ist da, sie sieht nur
// behilflich aus — ein `alert` unterbräche eine Vorlesehilfe mitten im Satz.
it('zeigt die Befunde des Zeichnens', async () => {
  belegeApi({
    Spieltextur: () =>
      Promise.resolve({
        textur: {
          schluessel: 'abc',
          adresse: '/spielbild/abc.png',
          guete: 'tisch',
          seite: 'vorne',
          breite: 512,
          hoehe: 714,
          befunde: ['Die Schrift MPlantin fehlt.'],
          zweiseitig: false,
        },
        fehler: null,
      }),
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

  expect(await screen.findByRole('status')).toHaveTextContent(
    'Die Schrift MPlantin fehlt.',
  )
})
