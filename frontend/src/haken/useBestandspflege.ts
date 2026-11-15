// Sichern und Zurückspielen als **ein** Vorgang
// (#940).
//
// # Warum die beiden hier zusammenkommen
//
// Regel 4 aus §7 sagt: ein Vorgang je Datensatz. Der Datensatz ist der
// **Bestand**, und es gibt genau einen — die beiden Wege darauf sind zwei
// Handlungen an derselben Datei. Zwei Haken mit je eigenem `useVorgang`
// sperrten einander nicht, und ein Sichern während eines Zurückspielens läse
// eine Datei, die gerade ausgetauscht wird.
//
// **Deshalb hier ein `useVorgang` und derselbe Schlüssel für beide.**
import { useState } from 'react'

import { useBestandSichern, type Sicherungslauf } from './useBestandSichern'
import {
  useSicherungsstand,
  type Sicherungsstandlauf,
} from './useSicherungsstand'
import {
  useSicherungZurueckspielen,
  type Rueckspiellauf,
} from './useSicherungZurueckspielen'
import { useVorgang } from './useVorgang'
import {
  useVerschluesselung,
  type Verschluesselungslauf,
} from './useVerschluesselung'

export interface Bestandspflege {
  sicherung: Sicherungslauf
  rueckspiel: Rueckspiellauf

  // chiffre schaltet die Verschlüsselung um — Werkstück 14 von Phase 4.12.
  //
  // **Unter demselben Vorgang wie die anderen beiden**, und der Grund im Kopf
  // dieser Datei gilt hier am schärfsten: das Umschalten **tauscht** die
  // Datei. Ein Sichern, das dabei liest, läse eine, die gerade verschwindet.
  chiffre: Verschluesselungslauf

  // alter ist, wann zuletzt gesichert wurde, und der Plan — OPS01 (#2347).
  // **Unter demselben Vorgang**: der Plan sichert denselben Bestand.
  alter: Sicherungsstandlauf

  // ersetzt zählt, wie oft der Bestand in dieser Sitzung ersetzt wurde —
  // Befund M9 (a) des Zweitblicks über Phase 6.3 (#1666). **Für die
  // aufgehobenen Seiten** (`useAufgehobeneSeiten`): sie räumt `danach` nicht,
  // weil es sie nicht kennt, und ohne diese Zahl blieben sie mit dem Zustand
  // des ersetzten Bestands stehen.
  ersetzt: number

  // gewechselt meldet, dass der Bestand auf einem **anderen** Weg gewechselt
  // hat — ein Neuanfang über die Rücksetzungsseite (#2292). Derselbe Weg wie
  // nach dem Zurückspielen: Zahl hoch, `danach` räumt und lädt neu.
  gewechselt: () => void

  // veraendert meldet, dass sich **im** Bestand geändert hat, was das Fenster
  // gelesen hat — zurückgesetzte Einstellungen (#2292). Es liest neu und
  // räumt nichts: Set und Karte gibt es danach noch.
  veraendert: () => void
}

// **`danach` läuft, wenn der Bestand ersetzt wurde.**
//
// Ein Zurückspielen tauscht **alles** — Sets, Karten, Einstellungen, Verlauf.
// Das ist Regel 1 aus AGENTS.md §7 in ihrer größten Form: das Bezugsobjekt hat
// gewechselt, und der offene Kontext gehört geräumt. Wer das nicht täte, zeigte
// Kennungen, die es nicht mehr gibt (Zweitblick über Phase S).
export function useBestandspflege(
  danach: () => void,
  neuLesen: () => void,
): Bestandspflege {
  const vorgang = useVorgang()
  const [ersetzt, setzeErsetzt] = useState(0)

  const gewechselt = () => {
    setzeErsetzt((bisher) => bisher + 1)
    danach()
  }

  const sicherung = useBestandSichern(vorgang)

  return {
    sicherung,
    rueckspiel: useSicherungZurueckspielen(vorgang, gewechselt),
    chiffre: useVerschluesselung(vorgang),

    // **Nachgelesen nach jeder eigenen Sicherung und jedem Tausch** — der
    // Pfad ist neu, oder der Bestand ein anderer.
    alter: useSicherungsstand(vorgang, [sicherung.pfad, ersetzt]),
    ersetzt,
    gewechselt,
    veraendert: neuLesen,
  }
}
