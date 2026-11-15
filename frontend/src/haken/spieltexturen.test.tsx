// **Die Spieltexturen im Fenster** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass der Haken die Zahlen wirklich führt.** Der Speicher daneben ist
// reine Buchführung; hier zählt, dass zwischen Anmelden und Abmelden
// **einmal** geholt wird — und dass ein Ding ohne Kartenkennung gar nicht
// erst fragt.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { useSpieltexturen } from './useSpieltexturen'

import type { ReactNode } from 'react'
import type { TexturAntwort } from '../apitextur'

afterEach(() => {
  raeumeApi()
})

// raum ist der Sprachkontext, den der Haken für seine Meldungen braucht.
function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

// eineAntwort gibt eine Textur, wie die Naht sie gibt.
function eineAntwort(adresse = '/spielbild/abc.png'): TexturAntwort {
  return {
    textur: {
      schluessel: 'abc',
      adresse,
      guete: 'tisch',
      seite: 'vorne',
      breite: 512,
      hoehe: 714,
      befunde: [],
      zweiseitig: false,
    },
    fehler: null,
  }
}

// **Sieben Anmeldungen, ein Ruf.**
it('holt für sieben gleiche Karten einmal', async () => {
  const geholt = vi.fn(() => Promise.resolve(eineAntwort()))

  belegeApi({ Spieltextur: geholt })

  const { result } = renderHook(() => useSpieltexturen(), { wrapper: raum })

  act(() => {
    for (let i = 0; i < 7; i += 1) {
      result.current.belege({
        art: 'karte',
        karteID: '7',
        seite: 'vorne',
        guete: 'tisch',
      })
    }
  })

  await waitFor(() => {
    expect(
      result.current.adresse({
        art: 'karte',
        karteID: '7',
        seite: 'vorne',
        guete: 'tisch',
      }),
    ).toBe('/spielbild/abc.png')
  })

  expect(geholt).toHaveBeenCalledTimes(1)
  expect(geholt).toHaveBeenCalledWith('7', 'vorne', 'tisch')
})

// **Ohne Kartenkennung wird gar nicht gefragt.**
//
// **Ein verdecktes Objekt und ein echter Druck tragen keine** — ein Ruf mit
// leerer Kennung wäre eine Abweisung, die niemand braucht.
it('fragt ohne Kartenkennung nicht', async () => {
  const geholt = vi.fn()

  belegeApi({ Spieltextur: geholt })

  const { result } = renderHook(() => useSpieltexturen(), { wrapper: raum })

  act(() => {
    result.current.belege({
      art: 'karte',
      karteID: '',
      seite: 'vorne',
      guete: 'tisch',
    })
    result.current.gib({
      art: 'karte',
      karteID: '',
      seite: 'vorne',
      guete: 'tisch',
    })
  })

  await antwortAbwarten()

  expect(geholt).not.toHaveBeenCalled()
  expect(
    result.current.adresse({
      art: 'karte',
      karteID: '',
      seite: 'vorne',
      guete: 'tisch',
    }),
  ).toBe('')
})

// **Was klagt, steht als Meldung da.**
it('meldet, was beim Holen schiefging', async () => {
  belegeApi({
    Spieltextur: () => Promise.reject(new Error('Der Stil fehlt.')),
  })

  const { result } = renderHook(() => useSpieltexturen(), { wrapper: raum })

  act(() => {
    result.current.belege({
      art: 'karte',
      karteID: '7',
      seite: 'vorne',
      guete: 'tisch',
    })
  })

  await waitFor(() => {
    expect(result.current.fehler).toBe('Der Stil fehlt.')
  })
})

// **Und das Räumen nimmt die Adressen und die Meldung mit.**
it('räumt Adressen und Meldung weg', async () => {
  belegeApi({ Spieltextur: () => Promise.resolve(eineAntwort()) })

  const { result } = renderHook(() => useSpieltexturen(), { wrapper: raum })

  act(() => {
    result.current.belege({
      art: 'karte',
      karteID: '7',
      seite: 'vorne',
      guete: 'tisch',
    })
  })

  await waitFor(() => {
    expect(
      result.current.adresse({
        art: 'karte',
        karteID: '7',
        seite: 'vorne',
        guete: 'tisch',
      }),
    ).not.toBe('')
  })

  act(() => {
    result.current.raeume()
  })

  expect(
    result.current.adresse({
      art: 'karte',
      karteID: '7',
      seite: 'vorne',
      guete: 'tisch',
    }),
  ).toBe('')
  expect(result.current.fehler).toBeUndefined()
})

// **Eine Antwort aus einer alten Runde wird nicht eingetragen.**
//
// # Der Weg, den der Speicherlauf allein nicht deckt
//
// **Der Speicher prüft die Runde, der Haken muss sie mitgeben.** Ein Lauf
// über den Speicher sagt, dass er eine alte Antwort abweist; dass der Haken
// die Runde **vom Anfang der Anfrage** nimmt und nicht die aktuelle, sagt er
// nicht — und genau das ist §7, Regel 2.
it('trägt eine Antwort aus einer alten Runde nicht ein', async () => {
  let loslassen: ((antwort: TexturAntwort) => void) | undefined

  belegeApi({
    Spieltextur: () =>
      new Promise<TexturAntwort>((aufloesen) => {
        loslassen = aufloesen
      }),
  })

  const { result } = renderHook(() => useSpieltexturen(), { wrapper: raum })

  const wunsch = {
    art: 'karte' as const,
    karteID: '7',
    seite: 'vorne',
    guete: 'tisch',
  }

  act(() => {
    result.current.belege(wunsch)
  })

  // **Dazwischen wird geräumt** — ein Profilwechsel, ein neuer Tisch.
  act(() => {
    result.current.raeume()
  })

  // **Und derselbe Schlüssel wird wieder belegt**, sodass der Beleg allein
  // die Antwort durchließe.
  act(() => {
    result.current.belege(wunsch)
  })

  act(() => {
    loslassen?.({
      textur: {
        schluessel: 'alt',
        adresse: '/spielbild/alt.png',
        guete: 'tisch',
        seite: 'vorne',
        breite: 512,
        hoehe: 714,
        befunde: ['aus der alten Runde'],
        zweiseitig: false,
      },
      fehler: null,
    })
  })

  await act(async () => {
    await Promise.resolve()
  })

  expect(result.current.adresse(wunsch)).toBe('')
  expect(result.current.befunde).toEqual([])
})

// **Was schon bekannt ist, wird nicht noch einmal vorgeladen.**
//
// **Und ein Wunsch ohne Bild gar nicht** — ein verdecktes Objekt trägt keine
// Kartenkennung, und ein Spiel ohne Namen keinen Rücken.
it('lädt nichts zweimal und nichts ohne Bild vor', async () => {
  const gefragt: string[] = []

  belegeApi({
    Spieltextur: (karteID: string) => {
      gefragt.push(karteID)

      return Promise.resolve({
        textur: {
          schluessel: karteID,
          adresse: `/spielbild/${karteID}.png`,
          guete: 'tisch-niedrig',
          seite: 'vorne',
          breite: 256,
          hoehe: 357,
          befunde: [],
          zweiseitig: false,
        },
        fehler: null,
      })
    },
    Rueckentextur: () => Promise.reject(new Error('kein Rücken gefragt')),
  })

  const { result } = renderHook(() => useSpieltexturen(), { wrapper: raum })

  const eines = {
    art: 'karte' as const,
    karteID: '7',
    seite: 'vorne',
    guete: 'tisch-niedrig',
  }

  await act(async () => {
    result.current.vorladen([
      eines,
      // **Zweimal derselbe** — und einmal geholt.
      eines,
      // **Ohne Kennung** — es gibt kein Bild.
      { art: 'karte' as const, karteID: '', seite: 'vorne', guete: 'tisch' },
      // **Ohne Spiel** — es gibt keinen Rücken.
      { art: 'ruecken' as const, spiel: '', guete: 'tisch' },
    ])

    await Promise.resolve()
  })

  await waitFor(() => {
    expect(result.current.adresse(eines)).toBe('/spielbild/7.png')
  })

  expect(gefragt).toEqual(['7'])

  // **Und wer ihn danach anmeldet, holt ihn auch nicht noch einmal.**
  act(() => {
    result.current.belege(eines)
  })

  expect(gefragt).toEqual(['7'])
})

// **Derselbe Befund steht einmal da, nicht sechzigmal.**
//
// **Die Zusage steht im Kopf des Hakens** — „sechzig Karten mit derselben
// fehlenden Schrift sind ein Befund und nicht sechzig" — und war ungeprüft.
// Ein Satz im Kommentar, den kein Lauf hält, ist eine Behauptung (§10).
it('nennt denselben Befund nur einmal', async () => {
  belegeApi({
    Spieltextur: (karteID: string) =>
      Promise.resolve({
        textur: {
          schluessel: karteID,
          adresse: `/spielbild/${karteID}.png`,
          guete: 'tisch',
          seite: 'vorne',
          breite: 512,
          hoehe: 714,
          befunde: ['Die Schrift MPlantin fehlt.', `Karte ${karteID}`],
          zweiseitig: false,
        },
        fehler: null,
      }),
  })

  const { result } = renderHook(() => useSpieltexturen(), { wrapper: raum })

  for (const karteID of ['7', '8']) {
    act(() => {
      result.current.belege({
        art: 'karte',
        karteID,
        seite: 'vorne',
        guete: 'tisch',
      })
    })
  }

  await waitFor(() => {
    expect(result.current.befunde).toHaveLength(3)
  })

  // **Der gemeinsame Satz einmal, die eigenen je einmal.**
  expect(result.current.befunde).toEqual([
    'Die Schrift MPlantin fehlt.',
    'Karte 7',
    'Karte 8',
  ])
})
