// **Beispiele, die laufen** — Werkstück 1 von Phase 3.3
// (#1164).
//
// # Warum es sie gibt
//
// Das Issue sagt es: „Für jemanden, der MSE-Script **nicht** kennt, ist ein
// leeres Feld keine Einladung." Eine Liste verfügbarer Namen hilft dem, der
// weiß, wonach er sucht; wer gar nicht anfängt, braucht einen ersten Ausdruck,
// der etwas tut.
//
// # Warum sie in zwei Gruppen stehen
//
// **Weil die eine Hälfte ohne Vorlagenpaket gar nicht läuft.** Am 2026-08-26
// nachgemessen, an einer Karte aus `testdaten/probe.mse-set` auf einem Rechner
// ohne eingebundenen Stil:
//
//	card.name                 → der Name "card" ist hier nicht bekannt
//	to_upper(card.name)       → der Name "card" ist hier nicht bekannt
//	set.title                 → der Name "set" ist hier nicht bekannt
//	"Karte " + (1 + 1)        → "Karte 2"
//
// `card` und `set` legt die **Vorlage** in die Umgebung; ohne Stilpaket bleibt
// die Bibliothek (§8). Ein Beispiel, das beim ersten Klick einen Fehler gibt,
// ist das Gegenteil einer Einladung — und genau der erste Klick trifft den,
// der auf einem frischen Rechner sitzt.
//
// Die erste Gruppe läuft **immer**, die zweite braucht eine eingebundene
// Vorlage. Der Satz darüber sagt es, statt es jemanden herausfinden zu lassen.
//
// # Warum genau diese
//
// Sie bauen aufeinander auf und decken zusammen ab, was die Sprache ausmacht:
// Rechnen, ein Funktionsaufruf, eine benannte Angabe, eine Bedingung, eine
// Verkettung — und danach der Zugriff auf Karte und Set.

// Beispielschluessel sind die Namen der Sätze in den Sprachkatalogen (§9).
//
// **Eine Aufzählung und keine Zeichenkette.** Damit ist der Nachschlag im
// Katalog **vollständig**: ein Schlüssel ohne Satz ist ein Übersetzungsfehler
// und kein Rückfall zur Laufzeit. Der erste Anlauf hatte ein `?? schluessel`
// daneben — einen Zweig, den kein Lauf betreten kann (§3).
export type Beispielschluessel =
  | 'rechnung'
  | 'grossSchreiben'
  | 'laenge'
  | 'ausschnitt'
  | 'bedingung'
  | 'name'
  | 'grossName'
  | 'regeltext'
  | 'setangabe'

// Beispiel ist ein Ausdruck samt dem Schlüssel seines Satzes.
export interface Beispiel {
  schluessel: Beispielschluessel
  ausdruck: string
}

// immerLaufende brauchen nichts als die Standardbibliothek.
export const immerLaufende: readonly Beispiel[] = [
  { schluessel: 'rechnung', ausdruck: '1 + 1' },
  { schluessel: 'grossSchreiben', ausdruck: 'to_upper("magic")' },
  { schluessel: 'laenge', ausdruck: 'length("Lightning Bolt")' },
  {
    schluessel: 'ausschnitt',
    ausdruck: 'substring("Lightning Bolt", begin: 0, end: 9)',
  },
  {
    schluessel: 'bedingung',
    ausdruck: 'if 1 == 1 then "ja" else "nein"',
  },
]

// mitVorlage brauchen ein eingebundenes Stilpaket — dort kommen `card` und
// `set` in die Umgebung.
export const mitVorlage: readonly Beispiel[] = [
  { schluessel: 'name', ausdruck: 'card.name' },
  { schluessel: 'grossName', ausdruck: 'to_upper(card.name)' },
  { schluessel: 'regeltext', ausdruck: 'length(card.rule_text)' },
  { schluessel: 'setangabe', ausdruck: 'set.title' },
]
