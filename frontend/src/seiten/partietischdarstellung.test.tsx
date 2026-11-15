// **Eine Darstellung zur Zeit** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Wogegen hier geprüft wird
//
// **Dass nie beide zugleich dastehen.** Der gezeichnete Tisch und die
// Zonenliste tragen dieselben Karten mit denselben Knöpfen; beide zugleich
// hieße, jede Karte zweimal in der Tabulatorreihenfolge zu haben — und wer
// mit der Tastatur bedient, ginge zweimal durch dasselbe Spiel.
//
// **Und dass der Tisch vorne steht.** Er ist das, wofür es #1460 gibt; die
// Liste ist die Sicht für Einzelheiten.
import { fireEvent, screen, waitFor } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { einLauf, einObjekt, einePartie } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { belegeApi } from '../test/api'
import { einNutzer } from '../test/anmeldebelegung'
import { eingepasst, nachPunkten } from '../tisch/kamera'
import { tischbilder } from '../tisch/tischbilder'
import { Tischbildschluessel } from './Partietisch'
import { tischbildquelle, Tischguete } from './tischbildquelle'
import { tischplan } from '../tisch/tischplan'
import { zeige } from './partietischprobe'

const eigene = kataloge.de.seiten.spielen

const nichts = () => undefined

// einTexturlauf gibt eine Naht, die immer dieselbe Adresse nennt.
const einTexturlauf = (belege: (wunsch: never) => void) =>
  ({
    belege,
    gib: nichts,
    adresse: () => '/spielbild/abc.png',
    zweiseitig: () => false,
    vorladen: nichts,
    raeume: nichts,
    befunde: [],
    fehler: undefined,
  }) as unknown as Parameters<typeof tischbildquelle>[1]

const mitWald = () =>
  einePartie({
    objekte: [
      einObjekt({
        kennung: 'w',
        zone: 'tisch',
        name: 'Wald',
        geheim: false,
        karteID: '7',
      }),
    ],
  })

it('stellt den gezeichneten Tisch voran', () => {
  zeige(einLauf(), mitWald(), false)

  expect(screen.getByRole('application')).toBeInTheDocument()
  expect(
    screen.queryByRole('heading', { name: eigene.zonen }),
  ).not.toBeInTheDocument()
})

it('tauscht die eine Darstellung gegen die andere', async () => {
  const hand = zeige(einLauf(), mitWald(), false)

  await hand.click(screen.getByRole('button', { name: eigene.tischListe }))

  expect(screen.queryByRole('application')).not.toBeInTheDocument()
  expect(
    screen.getByRole('heading', { name: eigene.zonen }),
  ).toBeInTheDocument()

  await hand.click(screen.getByRole('button', { name: eigene.tischGrafisch }))

  expect(screen.getByRole('application')).toBeInTheDocument()
})

// **Eine Wahl am gezeichneten Tisch ist dieselbe Wahl wie in der Liste** —
// sie steht am Knopf, und die Handlungsleiste darunter arbeitet mit ihr.
it('merkt sich, welche Karte am Tisch gewählt wurde', async () => {
  const hand = zeige(einLauf(), mitWald(), false)
  const knopf = screen.getByRole('button', { name: /Wald/ })

  expect(knopf).toHaveAttribute('aria-pressed', 'false')

  await hand.click(knopf)

  expect(screen.getByRole('button', { name: /Wald/ })).toHaveAttribute(
    'aria-pressed',
    'true',
  )
})

// **Ein Zug am Tisch ist derselbe Befehl wie über die Handlungsleiste** —
// eine andere Bedienung, keine andere Fachhandlung (#1460).
it('schickt einen Zug als Geste an den Wirt', async () => {
  const befehlen = vi.fn()
  const plan = tischplan(mitWald())
  const lage = plan.lagen.find((eine) => eine.kennung === 'w')
  const ziel = plan.felder.find(
    (eines) => eines.zone === 'friedhof' && eines.sitz === 'sitz1',
  )
  if (!lage || !ziel) throw new Error('Plan passt nicht zum Lauf')

  zeige(einLauf({ befehlen }), mitWald(), false)

  const flaeche = screen.getByRole('application')
  flaeche.getBoundingClientRect = () =>
    ({ left: 0, top: 0, width: 800, height: 600 }) as DOMRect

  // **Dieselbe Kamera wie der Haken** — und der hat ohne Renderer die Maße
  // null: unter jsdom gibt es kein WebGL, die Knöpfe stehen sichtbar da, und
  // die Fläche ist trotzdem bedienbar. Wer hier 800 × 600 annähme, rechnete
  // gegen eine Kamera, die es nicht gibt.
  const ohneFlaeche = { breite: 0, hoehe: 0 }
  const amOrt = (x: number, y: number) => {
    const punkt = nachPunkten(
      eingepasst(plan, ohneFlaeche, 16),
      { x, y },
      ohneFlaeche,
    )

    return { clientX: punkt.x, clientY: punkt.y }
  }

  fireEvent.pointerDown(flaeche, {
    button: 0,
    ...amOrt(lage.x + lage.breite / 2, lage.y + lage.hoehe / 2),
  })
  fireEvent.pointerUp(
    flaeche,
    amOrt(ziel.x + ziel.breite / 2, ziel.y + ziel.hoehe / 2),
  )

  await waitFor(() => {
    expect(befehlen).toHaveBeenCalledWith(
      'objekt.ziehen',
      expect.objectContaining({ objekt: 'w', zone: 'friedhof' }),
    )
  })
})

it('lässt jede Karte genau einmal in der Tabulatorreihenfolge', () => {
  zeige(einLauf(), mitWald(), false)

  expect(screen.getAllByRole('button', { name: /Wald/ })).toHaveLength(1)
})

it('meldet den Bedarf an und gibt die Adresse, die daraufhin kommt', () => {
  const partie = mitWald()
  const belege = vi.fn()
  const bild = tischbildquelle(partie, einTexturlauf(belege))
  const adresse = bild(
    {
      kennung: 'w',
      seite: 'vorne',
    } as Parameters<typeof bild>[0],
    false,
  )

  expect(adresse).toBe('/spielbild/abc.png')
  expect(belege).toHaveBeenCalledWith({
    art: 'karte',
    karteID: '7',
    bild: '',
    objekt: 'w',
    seite: 'vorne',
    guete: Tischguete,
  })
})

// **Ein Rücken gehört dem Spiel und keiner Karte** — PLAY-03.AK-07
// (#1711).
//
// **Und die Kennung ist erfunden**: sie gehört zu keinem Objekt dieser
// Partie. Der Lauf sagt damit zweierlei zugleich — der Rücken kommt, **und**
// er kommt, ohne dass irgendeine Karte nachgeschlagen wurde. Wäre der Zweig
// hinter dem Objektfund, gäbe dieser Aufruf `undefined`.
it('fragt nach dem Rücken des Spiels, wo verdeckt gemeint ist', () => {
  const belege = vi.fn()
  const bild = tischbildquelle(mitWald(), einTexturlauf(belege))

  const adresse = bild(
    { kennung: 'gibtesnicht' } as Parameters<typeof bild>[0],
    true,
  )

  expect(adresse).toBe('/spielbild/abc.png')
  expect(belege).toHaveBeenCalledWith({
    art: 'ruecken',
    spiel: 'magic',
    guete: Tischguete,
  })
})

it('gibt nichts her, wo die Lage zu keinem Objekt gehört', () => {
  const bild = tischbildquelle(mitWald(), einTexturlauf(nichts))

  expect(
    bild({ kennung: 'gibtesnicht' } as Parameters<typeof bild>[0], false),
  ).toBeUndefined()
})

// **Das Tischbild gehört dem Menschen** — wer seinen Tisch als Verlies sieht,
// sieht ihn so in jedem seiner Profile.
it('bietet die sechs Tischbilder zur Wahl an', () => {
  zeige(einLauf(), mitWald(), false)

  const wahl = screen.getByRole('combobox', { name: eigene.tischbild })

  expect(
    [...wahl.querySelectorAll('option')].map((eine) => eine.value),
  ).toEqual(tischbilder.map((eines) => eines.name))
})

// **Die gespeicherte Wahl gilt beim Aufschlagen**, und sie ist zugleich das
// Abschlusssignal, auf das dieser Lauf wartet: erst wenn sie in der Auswahl
// steht, kennt der Haken den Menschen, dem sie gehört. Wer vorher wählte,
// schriebe ins Leere — und der Lauf wäre grün, solange die Maschine frei ist
// (§3: auf die Wirkung warten, nicht auf den Ruf).
it('merkt sich das gewählte Tischbild beim Menschen', async () => {
  const setze = vi.fn(() => Promise.resolve({ einstellungen: [] }))
  const profil = einNutzer().profile[0]
  belegeApi({
    GezeigtesProfil: () => Promise.resolve({ profil: profil ?? null }),
    Nutzereinstellungen: () =>
      Promise.resolve({
        einstellungen: [
          {
            schluessel: Tischbildschluessel,
            wert: 'taverne',
            werte: [],
          },
        ],
      }),
    NutzereinstellungSetzen: setze,
  })

  const hand = zeige(einLauf(), mitWald(), false)
  const wahl = screen.getByRole('combobox', { name: eigene.tischbild })

  await waitFor(() => {
    expect(wahl).toHaveValue('taverne')
  })

  await hand.selectOptions(wahl, 'dungeon')

  await waitFor(() => {
    expect(setze).toHaveBeenCalledWith(
      einNutzer().id,
      Tischbildschluessel,
      'dungeon',
    )
  })
})
