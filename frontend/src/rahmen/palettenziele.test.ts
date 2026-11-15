// **Was die Befehlspalette tun und finden kann** — UX-02.AK-02
// (#1674).
import { describe, expect, it, vi } from 'vitest'

import {
  findziele,
  Hoechstens,
  passendeZiele,
  tunziele,
  type Palettendaten,
  type Palettenwege,
  type Palettenziel,
} from './palettenziele'
import { kataloge } from '../sprache/kataloge'

import type { DeckDTO } from '../apidecks'
import type { Peerkontakt } from '../apipeerkontakt'

const texte = kataloge.de
const eigene = texte.allgemein.palette

const deck = (id: string, name: string): DeckDTO =>
  ({ id, name, spiel: 'magic' }) as DeckDTO

const gegenueber = (kennung: string, eigenerName: string): Peerkontakt =>
  ({ kennung, eigenerName, selbstgewaehlterName: 'Sonst' }) as Peerkontakt

const daten: Palettendaten = {
  sets: [
    {
      id: '7',
      name: 'Werkstatt der Dämmerung',
      spiel: 'magic',
      stil: 'm15',
      herkunft: '',
      karten: 41,
      herkunftFehlt: false,
    },
  ],
  decks: [deck('3', 'Aschenrat')],
  gegenueber: [gegenueber('peer-1', 'Turfer')],
  karten: [
    {
      karteID: '12',
      setID: '7',
      nummer: 4,
      kartenname: 'Aschenwächter',
      feld: 'name',
      wert: 'Aschenwächter',
    },
  ],
}

// wege gibt die Wege der Palette — **mit offenem Set oder ohne**. Ein
// Vorgabeparameter `setID = '7'` hätte `undefined` still ersetzt.
function wege(mitSet = true) {
  return {
    setID: mitSet ? '7' : undefined,
    waehleSet: vi.fn(),
    oeffneKarte: vi.fn(),
    uebergib: vi.fn(),
    wechseln: vi.fn(),
  } satisfies Palettenwege
}

const text = (ziele: Palettenziel[]) => ziele.map((ziel) => ziel.text)

describe('Tun', () => {
  it('bietet mit offenem Set „Karte erstellen" an und legt sie auf der Kartenseite an', () => {
    const w = wege()
    const [erstes] = tunziele(daten, w, texte, false)

    expect(erstes?.text).toBe(eigene.karteErstellen)
    expect(erstes?.wo).toBe('Gestalten · Werkstatt der Dämmerung')

    erstes?.tun()
    expect(w.uebergib).toHaveBeenCalledWith({
      art: 'karte',
      name: 'Werkstatt der Dämmerung',
      setID: '7',
    })
  })

  // **Das offene Set kennt die Palette sofort, seinen Namen erst geladen** —
  // #2529: bis die Sets kamen, stand „Ein Set anlegen oder öffnen" da und
  // wurde danach zu „Karte erstellen"; wer „Sets" getippt hatte, sah die
  // Vorwahl springen.
  it('bietet „Karte erstellen" an, sobald ein Set offen ist, auch bevor die Sets geladen sind', () => {
    const w = wege()
    const [erstes] = tunziele({ ...daten, sets: [] }, w, texte, false)

    expect(erstes?.text).toBe(eigene.karteErstellen)
    expect(erstes?.wo).toBe('Gestalten')

    erstes?.tun()
    expect(w.uebergib).toHaveBeenCalledWith({
      art: 'karte',
      name: '',
      setID: '7',
    })
  })

  // **Ohne offenes Set gäbe es keinen Ort für eine Karte** — dann führt die
  // erste Zeile dorthin, wo eines entsteht.
  it('führt ohne offenes Set zu den Sets', () => {
    const w = wege(false)
    const [erstes] = tunziele(daten, w, texte, false)

    expect(erstes?.text).toBe(eigene.setWaehlen)
    erstes?.tun()
    expect(w.wechseln).toHaveBeenCalledWith('sets')
  })

  it('spielt ein Deck, tauscht mit einem Gegenüber und erstellt eine Sicherung', () => {
    const w = wege()
    const ziele = tunziele(daten, w, texte, false)

    expect(text(ziele)).toEqual([
      eigene.karteErstellen,
      'Schnell spielen mit Aschenrat',
      'Karten tauschen mit Turfer',
      eigene.sicherungErstellen,
      eigene.schnittstellenEinrichten,
    ])

    for (const ziel of ziele.slice(1)) ziel.tun()

    // **Schnell spielen** (P6, #2374): gleich an den Tisch.
    expect(w.uebergib).toHaveBeenNthCalledWith(1, {
      art: 'partie',
      name: 'Aschenrat',
      deckID: '3',
      sofort: true,
    })
    expect(w.uebergib).toHaveBeenNthCalledWith(2, {
      art: 'gegenueber',
      name: 'Turfer',
      kennung: 'peer-1',
    })
    expect(w.uebergib).toHaveBeenNthCalledWith(3, {
      art: 'einstellungsgruppe',
      name: texte.seiten.einstellungen.gruppen.daten,
      gruppe: 'daten',
    })
  })

  // **„Schnittstellen einrichten“ schlägt die Werkzeuge ohne Fenster auf** —
  // AP-10 (#2367), B-19: Kommandozeile, MCP und REST. **Nicht für einen
  // Leser**: die Einrichtung schreibt außerhalb des Bestands.
  it('schlägt für „Schnittstellen einrichten“ die Werkzeuge unter Verbindungen auf', () => {
    const w = wege()
    const ziel = tunziele(daten, w, texte, false).find(
      (eines) => eines.text === eigene.schnittstellenEinrichten,
    )

    expect(ziel?.wo).toBe(
      eigene.woEinstellungen.replace(
        '{gruppe}',
        texte.seiten.einstellungen.gruppen.verbindungen,
      ),
    )
    ziel?.tun()
    expect(w.uebergib).toHaveBeenCalledWith({
      art: 'einstellungsgruppe',
      name: texte.seiten.einstellungen.gruppen.verbindungen,
      gruppe: 'verbindungen',
      abschnitt: 'werkzeuge',
    })

    // **Der Leser bekommt die Sicherung, die Einrichtung nicht.**
    expect(text(tunziele(daten, wege(), texte, true))).toEqual([
      eigene.setWaehlen,
      eigene.sicherungErstellen,
    ])
  })

  // **Wie das Gegenüber heißt, sagt dieses Profil** — sonst sein
  // selbstgewählter Name, wie in `Partnerwahl`.
  it('nennt ein Gegenüber ohne eigenen Namen, wie es sich selbst nennt', () => {
    const ohneNamen = {
      ...daten,
      gegenueber: [gegenueber('peer-2', '')],
    }

    expect(text(tunziele(ohneNamen, wege(), texte, false))).toContain(
      'Karten tauschen mit Sonst',
    )
  })
})

describe('Finden', () => {
  it('findet Karte, Set, Deck und Gegenüber beim Namen', () => {
    const w = wege()
    const ziele = findziele(daten, w, texte)

    expect(ziele.map((ziel) => [ziel.text, ziel.wo])).toEqual([
      ['Aschenwächter', 'Karte · Werkstatt der Dämmerung'],
      ['Werkstatt der Dämmerung', 'Set · 41 Karten'],
      ['Aschenrat', 'Deck · magic'],
      ['Turfer', eigene.alsGegenueber],
    ])

    for (const ziel of ziele) ziel.tun()

    expect(w.oeffneKarte).toHaveBeenCalledWith('7', '12')
    expect(w.waehleSet).toHaveBeenCalledWith('7')
    expect(w.uebergib).toHaveBeenCalledWith({
      art: 'deckansicht',
      name: 'Aschenrat',
      deckID: '3',
    })
    expect(w.uebergib).toHaveBeenCalledWith({
      art: 'gegenueber',
      name: 'Turfer',
      kennung: 'peer-1',
    })
  })

  // **Eine Karte aus einem Set, das die Palette nicht kennt**, steht ohne
  // Setnamen da — statt gar nicht.
  it('lässt den Setnamen leer, wenn das Set nicht in der Liste steht', () => {
    const fremd = { ...daten, sets: [] }

    expect(findziele(fremd, wege(), texte)[0]?.wo).toBe('Karte · ')
  })
})

describe('passendeZiele', () => {
  const ziel = (text: string, wo = ''): Palettenziel => ({
    schluessel: text,
    text,
    wo,
    tun: () => undefined,
  })

  it('stellt voran, was mit der Eingabe anfängt, dann was sie enthält, dann den Ort', () => {
    const ziele = [
      ziel('Glutschmiede', 'Karte · Asche'),
      ziel('Große Asche'),
      ziel('Aschenwächter'),
      ziel('Wald'),
    ]

    expect(text(passendeZiele(ziele, ' asch '))).toEqual([
      'Aschenwächter',
      'Große Asche',
      'Glutschmiede',
    ])
  })

  it('zeigt höchstens einen Abschnitt voll', () => {
    const viele = Array.from({ length: Hoechstens + 3 }, (_, stelle) =>
      ziel(`Karte ${String(stelle)}`),
    )

    expect(passendeZiele(viele, 'karte')).toHaveLength(Hoechstens)
  })
})
