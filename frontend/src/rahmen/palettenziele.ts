// **Was die Befehlspalette tun und finden kann** — UX-02.AK-02
// (#1674):
// *„Cmd+K findet Objekte und Aktionen kontextabhängig, etwa ‚Karte
// erstellen', ‚Deck spielen', ‚Karten tauschen' und ‚Sicherung erstellen', und
// führt ohne Kenntnis technischer Seitennamen zum nächsten sinnvollen
// Schritt."*
//
// # Zwei Abschnitte, wie im Artboard *Befehle*
//
// **Tun** sind Handlungen — sie hängen an dem, was offen ist und was es
// gibt: „Karte erstellen" nur mit offenem Set, „Deck spielen" je Deck,
// „Karten tauschen" je bestätigtem Gegenüber. **Finden** sind Objekte beim
// Namen: Sets, Karten, Decks, Gegenüber.
//
// # Warum hier keine Handlung selbst steht
//
// **Jede geht über einen Weg, den es schon gibt**: ein Set über `waehleSet`,
// eine Karte über `oeffneKarte`, alles andere über einen Anschluss
// (`anschluss.ts`) — dort nimmt die Zielfläche auf, mit ihrem eigenen
// Vorgang und ihrem eigenen Fehler. Eine Palette, die selbst schriebe, wäre
// ein zweiter Schreibweg neben jedem ersten.
import { einsetzen } from '../sprache/einsetzen'
import { zaehlwort } from '../sprache/zaehlwort'

import type { Anschluss } from '../anschluss'
import type { DeckDTO } from '../apidecks'
import type { Peerkontakt } from '../apipeerkontakt'
import type { SetDTO, TrefferDTO } from '../formen'
import type { Seitenschluessel } from '../navigation'
import type { Katalog } from '../sprache/kataloge'

// Palettenziel ist eine Zeile der Palette.
export interface Palettenziel {
  // schluessel unterscheidet die Zeile — auch zwei gleichnamige Decks.
  schluessel: string

  text: string

  // wo sagt, wohin die Zeile führt — **der Bereich und nicht die Seite**:
  // „Spielen · Partie beginnen", nicht `spielen`.
  wo: string

  tun: () => void
}

// Palettendaten ist, was die Palette über den Bestand weiß.
export interface Palettendaten {
  sets: readonly SetDTO[]
  decks: readonly DeckDTO[]
  gegenueber: readonly Peerkontakt[]
  karten: readonly TrefferDTO[]
}

// Palettenwege sind die Wege, über die eine Zeile wirkt.
export interface Palettenwege {
  // setID ist das offene Set — leer, wenn keines offen ist.
  setID: string | undefined

  waehleSet: (setID: string) => void
  oeffneKarte: (setID: string, karteID: string) => void
  uebergib: (anschluss: Anschluss) => void
  wechseln: (seite: Seitenschluessel) => void
}

// gegenuebername ist, wie dieses Profil ein Gegenüber nennt — sonst, wie es
// sich selbst nennt (dieselbe Regel wie in `Partnerwahl`).
function gegenuebername(kontakt: Peerkontakt): string {
  return kontakt.eigenerName || kontakt.selbstgewaehlterName
}

// tunziele gibt die Handlungen, die gerade gehen.
//
// **`lesend` lässt weg, was schreibt** (G07, #2254) — eine Karte anlegen,
// eine Partie beginnen, einen Tausch anfangen. Statt der Karte steht dann
// die Wahl eines Sets vorn, wie ohne offenes Set.
export function tunziele(
  daten: Palettendaten,
  wege: Palettenwege,
  texte: Katalog,
  lesend: boolean,
): Palettenziel[] {
  const eigene = texte.allgemein.palette
  const offen = daten.sets.find((eines) => eines.id === wege.setID)

  // **Das offene Set kennt die Palette sofort, seinen Namen erst geladen** —
  // #2529: entschieden wurde am geladenen Set, und bis die Sets kamen, stand
  // „Ein Set anlegen oder öffnen" da. Wer „Sets" tippte, sah die Vorwahl
  // springen, als die Zeile zu „Karte erstellen" wurde. Bis der Name da ist,
  // nennt die Zeile nur den Arbeitsbereich.
  const offeneID = wege.setID
  const ort =
    offen === undefined
      ? einsetzen(eigene.woKarte, { set: '' }).replace(/\s*·\s*$/, '')
      : einsetzen(eigene.woKarte, { set: offen.name })

  const erstes: Palettenziel =
    offeneID === undefined || lesend
      ? {
          schluessel: 'tun:set',
          text: eigene.setWaehlen,
          wo: eigene.woSets,
          tun: () => {
            wege.wechseln('sets')
          },
        }
      : {
          schluessel: 'tun:karte',
          text: eigene.karteErstellen,
          wo: ort,
          tun: () => {
            wege.uebergib({
              art: 'karte',
              name: offen?.name ?? '',
              setID: offeneID,
            })
          },
        }

  const gruppe = texte.seiten.einstellungen.gruppen.daten
  const verbindungen = texte.seiten.einstellungen.gruppen.verbindungen

  return [
    erstes,
    ...(lesend ? [] : schreibendeZiele(daten, wege, texte)),
    {
      schluessel: 'tun:sicherung',
      text: eigene.sicherungErstellen,
      wo: einsetzen(eigene.woEinstellungen, { gruppe }),
      tun: () => {
        wege.uebergib({
          art: 'einstellungsgruppe',
          name: gruppe,
          gruppe: 'daten',
        })
      },
    },

    // **Die Schnittstellen** — AP-10 (#2367), B-19: Kommandozeile, MCP und
    // REST unter *Verbindungen*. **Nicht für einen Leser**: die Einrichtung
    // schreibt außerhalb des Bestands, in `~/.local/bin` und LaunchAgents.
    ...(lesend
      ? []
      : [
          {
            schluessel: 'tun:schnittstellen',
            text: eigene.schnittstellenEinrichten,
            wo: einsetzen(eigene.woEinstellungen, { gruppe: verbindungen }),
            tun: () => {
              wege.uebergib({
                art: 'einstellungsgruppe',
                name: verbindungen,
                gruppe: 'verbindungen',
                abschnitt: 'werkzeuge',
              })
            },
          },
        ]),
  ]
}

// schreibendeZiele sind Deck spielen und Tauschen — beide legen etwas an.
function schreibendeZiele(
  daten: Palettendaten,
  wege: Palettenwege,
  texte: Katalog,
): Palettenziel[] {
  const eigene = texte.allgemein.palette

  return [
    ...daten.decks.map((deck) => ({
      schluessel: `tun:deck:${deck.id}`,
      text: einsetzen(eigene.deckSpielen, { name: deck.name }),
      wo: eigene.woPartie,
      tun: () => {
        // **Schnell spielen** (P6, #2374): mit der Vorbelegung gleich an
        // den Tisch — derselbe Anfang wie aus dem Menü des Decks.
        wege.uebergib({
          art: 'partie',
          name: deck.name,
          deckID: deck.id,
          sofort: true,
        })
      },
    })),
    ...daten.gegenueber.map((kontakt) => ({
      schluessel: `tun:gegenueber:${kontakt.kennung}`,
      text: einsetzen(eigene.tauschenMit, { name: gegenuebername(kontakt) }),
      wo: eigene.woHandel,
      tun: () => {
        wege.uebergib({
          art: 'gegenueber',
          name: gegenuebername(kontakt),
          kennung: kontakt.kennung,
        })
      },
    })),
  ]
}

// findziele gibt die Objekte, beim Namen.
export function findziele(
  daten: Palettendaten,
  wege: Palettenwege,
  texte: Katalog,
): Palettenziel[] {
  const eigene = texte.allgemein.palette
  const woerter = texte.allgemein.zaehlwoerter
  const setname = (setID: string) =>
    daten.sets.find((eines) => eines.id === setID)?.name ?? ''

  return [
    ...daten.karten.map((treffer) => ({
      schluessel: `finden:karte:${treffer.karteID}`,
      text: treffer.kartenname,
      wo: einsetzen(eigene.alsKarte, { set: setname(treffer.setID) }),
      tun: () => {
        wege.oeffneKarte(treffer.setID, treffer.karteID)
      },
    })),
    ...daten.sets.map((set) => ({
      schluessel: `finden:set:${set.id}`,
      text: set.name,
      wo: einsetzen(eigene.alsSet, {
        karten: zaehlwort(set.karten, woerter.karten),
      }),
      tun: () => {
        wege.waehleSet(set.id)
      },
    })),
    ...daten.decks.map((deck) => ({
      schluessel: `finden:deck:${deck.id}`,
      text: deck.name,
      wo: einsetzen(eigene.alsDeck, { spiel: deck.spiel }),
      tun: () => {
        wege.uebergib({ art: 'deckansicht', name: deck.name, deckID: deck.id })
      },
    })),
    ...daten.gegenueber.map((kontakt) => ({
      schluessel: `finden:gegenueber:${kontakt.kennung}`,
      text: gegenuebername(kontakt),
      wo: eigene.alsGegenueber,
      tun: () => {
        wege.uebergib({
          art: 'gegenueber',
          name: gegenuebername(kontakt),
          kennung: kontakt.kennung,
        })
      },
    })),
  ]
}

// Ohneeingabe ist, wie viele Handlungen ohne Eingabe dastehen — die ersten,
// und die erste ist die, die zum offenen Set gehört.
export const Ohneeingabe = 4

// Hoechstens ist, wie viele Zeilen ein Abschnitt zeigt.
export const Hoechstens = 8

// passendeZiele filtert über Text und Ort; **was mit der Eingabe anfängt,
// steht vorn** — wer „Asch" tippt, meint den Aschenwächter und nicht jede
// Karte, deren Set so heißt.
export function passendeZiele(
  ziele: readonly Palettenziel[],
  eingabe: string,
): Palettenziel[] {
  const gesucht = eingabe.trim().toLowerCase()

  const rang = (ziel: Palettenziel): number => {
    const text = ziel.text.toLowerCase()

    if (text.startsWith(gesucht)) return 0
    if (text.includes(gesucht)) return 1
    if (ziel.wo.toLowerCase().includes(gesucht)) return 2

    return -1
  }

  return ziele
    .map((ziel, stelle) => ({ ziel, stelle, rang: rang(ziel) }))
    .filter((eintrag) => eintrag.rang >= 0)
    .sort((a, b) => a.rang - b.rang || a.stelle - b.stelle)
    .map((eintrag) => eintrag.ziel)
    .slice(0, Hoechstens)
}
