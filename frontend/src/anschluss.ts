// **Was von einer Fläche zur nächsten mitreist** — UX-03.AK-01
// (#1675).
//
// # Was das Kriterium verlangt
//
// *„Karte, Set, Sammlung, Deck, Turnier, Partie und Handel erhalten
// konsistente Objektverweise; ‚Zum Deck‘, ‚In Sammlung‘, ‚Spielen‘, ‚Teilen‘
// und ‚Tauschen‘ übernehmen Auswahl und zulässigen Kontext."*
//
// Der Entwurf sagt es in vier Wörtern, und sie stehen im Artboard `Sammlung`
// neben dem Weg zum Deck: **„Die Auswahl reist mit."**
//
// # Warum eine eigene Datei ohne React
//
// **Weil die Frage „wohin gehört das?" ohne einen Baum zu beantworten ist.**
// Was ein Anschluss trägt und wo er landet, ist eine Aussage über diese
// Anwendung; *wie* er ankommt, steht in `anschlussraum.tsx`.
import type { Seitenschluessel } from './navigation'

// anschlussarten ist die Aufzählung — **zur Laufzeit, nicht nur im Typ**.
//
// Ein Wächter, der jede Art gegen ihre Fläche hält, braucht sie als Wert;
// ein `type` allein ließe sich nicht durchgehen.
export const anschlussarten = [
  'deck',
  'handel',
  'partie',
  // **Die vier aus der Befehlspalette** — UX-02.AK-02
  // (#1674):
  // „Karte erstellen", ein Deck aufschlagen, mit einem Gegenüber tauschen,
  // eine Sicherung erstellen. Sie führen **zum nächsten Schritt**, ohne dass
  // jemand wissen muss, wie die Seite dazu heißt.
  'karte',
  'deckansicht',
  'gegenueber',
  'einstellungsgruppe',
  // **Und die zwei mit einer Auswahl** — UX-03.AK-01
  // (#1675):
  // „3 Exemplare → Deck" aus der Sammlung und „Die 7 eintauschen" aus dem
  // Deck, wie die Artboards *Sammlung* und *Decks* sie zeichnen.
  'deckstapel',
  'tauschwunsch',
  // **Und eine Partie aus einem Turnier** — DOM-04.AK-02
  // (#1699):
  // eine an eine Paarung gebundene Partie wird am Tisch aufgeschlagen.
  'partieansicht',
  // **Und ein Paket, das einem Set fehlt** — DOM-01.AK-04
  // (#1696):
  // der Bedarf eines Sets führt zur Paketseite, die Suche schon gefüllt.
  'paketsuche',
  // **Und die Pakete mit Befund aus der Aufmerksamkeit** — AP-8
  // (#2365):
  // die Paketliste zeigt genau sie, gefiltert.
  'paketauswahl',
  // **Und die Sets mit Problemen aus dem Überblick** — AP-8, Teil 2 (#2365).
  'setauswahl',
  // **Und die Karten mit Befund eines Sets** — AP-8, Teil 2b: die
  // Kartenliste zeigt genau sie.
  'kartenauswahl',
  // **Und die Decks mit fehlenden Karten** — AP-8, Teil 2b: die Deckliste
  // zeigt genau sie.
  'deckauswahl',
  // **Und der Beitritt an einen fremden Tisch** — P9c
  // (#2377):
  // „Bei Ada mitspielen“ von der Deckseite, S13 in zwei Handlungen.
  'mitspielen',
] as const

// Stueck ist ein Druck mit seiner Menge — das, was eine Auswahl trägt.
export interface Stueck {
  printing: string

  // name ist, was ein Mensch sieht — der Kartenname, nicht die Kennung.
  name: string

  menge: number
}

// Anschlussart sagt, **was** übergeben wird — und damit auch, wohin.
//
// **Die Art und nicht das Ziel ist das Unterscheidungsmerkmal.** Ein Druck,
// der ins Deck reist, trägt anderes bei sich als einer, der in ein Angebot
// geht — die Güte etwa gehört zum Handel und nicht zum Deckplatz. Stünde nur
// das Ziel da, müsste jede Fläche raten, welche Felder gefüllt sind.
export type Anschlussart = (typeof anschlussarten)[number]

// Anschluss ist ein Objekt auf dem Weg zu der Fläche, die etwas damit tut.
//
// **Ein unterscheidender Verbund und kein Sack mit optionalen Feldern.** Wer
// `art === 'partie'` liest, hat eine Deckkennung und keinen Druck; der
// Übersetzer hält das, und keine Fläche prüft es noch einmal.
export type Anschluss =
  | {
      art: 'deck'

      // name ist, was ein Mensch sieht — der Kartenname, nicht die Kennung.
      name: string

      // printing ist der Druck, den die Kartenwahl des Decks übernimmt.
      printing: string
    }
  | {
      art: 'handel'
      name: string
      printing: string

      // guete, sprache und finish reisen beim Handel mit: eine Position
      // nennt sie, und wer sie hier fallen ließe, ließe den Menschen
      // dieselben vier Felder noch einmal ausfüllen.
      guete: string
      sprache: string
      finish: string
    }
  | {
      art: 'partie'
      name: string

      // deckID ist das Deck, mit dem gespielt werden soll.
      deckID: string

      // sofort fängt mit der Vorbelegung gleich an — „Schnell spielen“ (P6,
      // #2374). Ohne bleibt das Formular stehen.
      sofort?: true

      // eingeladen setzt den zweiten Sitz auf einen Gast, der sein Deck
      // mitbringt — „Bo einladen“ (P9c, #2377).
      eingeladen?: { kennung: string; name: string }
    }
  | {
      // mitspielen setzt sich an den Tisch eines Gegenübers — mit dem Deck,
      // das mitgeht (P9c, #2377); leer heißt: der Wirt stellt eines.
      art: 'mitspielen'
      name: string
      gegenueber: string
      deckID: string
    }
  | {
      // karte legt im offenen Set eine leere Karte an — **auf der
      // Kartenseite selbst**, mit deren Vorgang und deren Fehler am Feld.
      art: 'karte'
      name: string

      // setID ist das Set, in das sie soll. Ist inzwischen ein anderes offen,
      // tut die Seite nichts: eine Karte im falschen Set wäre schlimmer als
      // keine (§7, Regel 2).
      setID: string
    }
  | {
      // deckansicht schlägt ein Deck auf — anders als `deck`, das einen
      // Druck in eines legen will.
      art: 'deckansicht'
      name: string
      deckID: string

      // lasche schlägt im Deck eine Lasche auf — die Prüfung, wenn die
      // Aufmerksamkeit wegen fehlender Karten hinführt (AP-8, B-14).
      lasche?: 'pruefung'
    }
  | {
      art: 'gegenueber'
      name: string

      // kennung ist die Peerkennung des bestätigten Gegenübers.
      kennung: string
    }
  | {
      art: 'einstellungsgruppe'
      name: string

      // gruppe ist eine der neun aus `einstellungsgruppen.ts`.
      gruppe: string

      // abschnitt wird darin aufgeschlagen und bekommt den Fokus — einer aus
      // `einstellungsabschnitte.ts` (AP-10, #2367: „Schnittstellen
      // einrichten“ endete sonst vor einem zugeklappten Kasten).
      abschnitt?: string

      // vorgang ist eine Vorgangskennung, auf die die Diagnose filtert —
      // „Diagnose öffnen" an einem unerwarteten Fehler (NET-02, S13.NAT-04,
      // #1706).
      vorgang?: string
    }
  | {
      // deckstapel legt mehrere Drucke in das aufgeschlagene Deck.
      art: 'deckstapel'
      name: string
      stuecke: readonly Stueck[]
    }
  | {
      // partieansicht schlägt eine vorhandene Partie auf — anders als
      // `partie`, das mit einem Deck eine neue anfängt.
      art: 'partieansicht'
      name: string
      partieID: string

      // sitz ist der Sitz, aus dessen Sicht sie aufgeht — die Sicht gehört
      // einem Sitz, und eine ohne ihn gibt es nicht.
      sitz: string
    }
  | {
      // tauschwunsch trägt, was einem Deck fehlt, als Wunsch in einen Handel
      // — **vom Gegenüber**, denn es soll hereinkommen.
      art: 'tauschwunsch'
      name: string
      stuecke: readonly Stueck[]
    }
  | {
      art: 'paketsuche'
      name: string

      // paket ist der Dateiname, wie der Bedarf ihn nennt — `magic-m15.mse-style`.
      paket: string
    }
  | {
      // paketauswahl zeigt in der Paketliste genau diese Pakete — **die
      // Namen aus derselben Antwort wie die Zahl**, damit Wert und Filter
      // zum selben Stand gehören (`PLAN-0.9.0.md`, Abhängigkeiten).
      art: 'paketauswahl'

      // name ist der Satz, der die Auswahl nennt — er steht am Filter.
      name: string
      namen: readonly string[]
    }
  | {
      // setauswahl zeigt in der Setliste genau diese Sets — wie
      // `paketauswahl`, die Kennungen aus derselben Antwort wie die Zahl.
      art: 'setauswahl'
      name: string
      setIDs: readonly string[]
    }
  | {
      // kartenauswahl zeigt in der Kartenliste eines Sets genau diese Karten
      // — die mit Befund aus seinem Prüfstand. **An ein Set gebunden**: die
      // Kartenseite nimmt sie nur für dieses Set auf.
      art: 'kartenauswahl'
      name: string
      setID: string
      karteIDs: readonly string[]
    }
  | {
      // deckauswahl zeigt in der Deckliste genau diese Decks — die mit
      // fehlenden Karten, aus derselben Antwort wie die Zahl.
      art: 'deckauswahl'
      name: string
      deckIDs: readonly string[]
    }

// anschlussseite sagt, wo eine Art aufgenommen wird.
//
// **Ein vollständiger Verbund und keine Abbildung mit Rückfall.** Eine neue
// Art ohne Fläche wäre ein Anschluss ins Leere; so wird sie vom Übersetzer
// verlangt, statt still bei `undefined` zu landen.
export const anschlussseite: Readonly<Record<Anschlussart, Seitenschluessel>> =
  {
    deck: 'decks',
    handel: 'handeln',
    partie: 'spielen',
    karte: 'karten',
    deckansicht: 'decks',
    gegenueber: 'handeln',
    einstellungsgruppe: 'einstellungen',
    deckstapel: 'decks',
    tauschwunsch: 'handeln',
    partieansicht: 'spielen',
    paketsuche: 'pakete',
    paketauswahl: 'pakete',
    setauswahl: 'sets',
    kartenauswahl: 'karten',
    deckauswahl: 'decks',
    mitspielen: 'spielen',
  }

// setgebunden sind die Arten, die ein Set tragen und deshalb auf eine
// Setlasche führen dürfen — die Befehlspalette bietet sie nur an, wenn ein
// Set offen ist, und die Fläche prüft beim Aufnehmen, ob es noch dasselbe
// ist.
export const setgebunden: readonly Anschlussart[] = ['karte', 'kartenauswahl']
