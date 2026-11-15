// **Fünf Arbeitsbereiche, ein Einstieg und eine Dienstzone** — UX-02
// (#1674),
// gezeichnet in `PROTOTYP-0.6.0.md` (Entwicklungsarchiv).
//
// # Wofür es sie gibt
//
// **Weil dreizehn gleichrangige Einträge keine Ordnung sind.** Die
// Seitenleiste führte Start, Sets, Karten, Pakete, Vorlagen, Sammlung, Decks,
// Turniere, Spielen, Handeln, Hilfe, Einstellungen und Nutzer — nebeneinander,
// ohne Gruppe, und wer „ein Deck bauen" wollte, musste wissen, dass das unter
// *Decks* steht und nicht unter *Sammlung*.
//
// UX-02.AK-01 verlangt „ungefähr fünf primäre Arbeitsbereiche" und sagt dazu:
// *„die endgültigen Namen sind Ergebnis dieser Prüfung."* Der Entwurf ist ihr
// Gegenstand; diese Datei baut ihn.
//
// # Warum eine eigene Datei
//
// **Weil `navigation.ts` mit ihr über die Zeilengrenze ginge** (§2). Der
// Schnitt folgt der Frage: dort steht, **welche Seiten** es gibt und wie man
// sie erreicht — hier, **wie sie geordnet** sind.
//
// # Und warum die Dienstzone keinen der fünf Plätze verbraucht
//
// **UX-02.AK-04 sagt es wörtlich**: *„Einstellungen, Hilfe und Profilwechsel
// bleiben direkt zugänglich, ohne neue fachliche Hauptbereiche zu
// verbrauchen."* Sie steht deshalb als eigene Zone unter den fünf, und ihre
// drei Einträge behalten ihre macOS-Kürzel — ⌘, ⌘' ⌘? —, die §6 ohnehin
// verlangt.

import { anwendungsschluessel, setschluessel } from './navigation'

import type { Seitenschluessel } from './navigation'

// Bereichsschluessel benennt einen Arbeitsbereich.
export const bereichsschluessel = [
  'start',
  'gestalten',
  'sammeln',
  'spielen',
  'handeln',
  'ressourcen',
  'programm',
] as const

export type Bereichsschluessel = (typeof bereichsschluessel)[number]

// Bereich ist ein Arbeitsbereich samt den Seiten, die zu ihm gehören.
export interface Bereich {
  schluessel: Bereichsschluessel

  // kuerzel ist die Ziffer, die zusammen mit ⌘ in diesen Bereich führt —
  // leer bei der Dienstzone.
  //
  // **Sie führt auf die erste Seite des Bereichs.** Ein Kürzel, das einen
  // Bereich „öffnet", ohne irgendwo zu landen, wäre eine Taste ohne Wirkung.
  kuerzel: string

  // seiten sind die Seiten dieses Bereichs, in der Reihenfolge, in der die
  // Leiste sie zeigt.
  seiten: readonly Seitenschluessel[]

  // dienst sagt, ob dieser Bereich die Dienstzone ist.
  //
  // **Als Feld und nicht als Vergleich mit `'programm'`**: die Leiste zeichnet
  // sie mit einem Trenner darüber und ohne Gruppenüberschrift, und wer das an
  // einen Schlüsselnamen hinge, hätte die Gestalt an einem Wort festgemacht.
  dienst: boolean
}

// dienstseiten sind die drei Einträge der Dienstzone.
//
// **Eine eigene Aufzählung und nicht bloß ein Feld im Bereich.** Sie sind die
// einzigen Einträge, die eine **Seite** benennen statt eines Bereichs, und die
// Leiste braucht zu jedem ein eigenes Sinnbild. Mit diesem Typ ist die
// Zuordnung dort vollständig — ohne ihn stünde daneben ein Rückfall, den kein
// Aufruf betritt.
export const dienstseiten = ['einstellungen', 'nutzer', 'hilfe'] as const

export type Dienstseite = (typeof dienstseiten)[number]

// bereiche ist die Ordnung, wie der Entwurf sie zeichnet.
//
// # Zwei Abweichungen von der Beispielliste in UX-02.AK-01
//
// Beide stehen mit Grund in `PROTOTYP-0.6.0.md`:
//
//   - **„Ressourcen" statt „Werkzeuge".** Pakete und Vorlagen sind Material,
//     mit dem gearbeitet wird, keine Werkzeuge.
//   - **Aus sechs Bereichen werden fünf.** Start ist der Einstieg und Hilfe
//     Teil der Dienstzone; sonst verbrauchten beide je einen der fünf Plätze
//     — was AK-04 gerade verhindern will.
//
// # Und eine Abweichung vom Entwurf selbst
//
// **Der Entwurf legt `karten` unter *Gestalten* als „Kartenarbeitsplatz" an**,
// und das bleibt so. Was er **nicht** vorwegnimmt: die vier Nebenflächen
// (*Eigenschaften des Sets*, *Werkzeuge*, *Ausgeben*, *Einlesen*) entstehen
// als Drawer und Wizards mit den Issues, die sie bauen — UX-05, UX-13, FMT-03.
// Solange es sie nicht gibt, bleiben ihre Seiten **Laschen des offenen Sets**
// und damit dort, wo sie heute sind. Das ist die Reihenfolge aus §13: eine
// Fläche wandert, wenn die neue steht, nicht vorher.
export const bereiche: readonly Bereich[] = [
  { schluessel: 'start', kuerzel: '1', seiten: ['start'], dienst: false },

  {
    schluessel: 'gestalten',
    kuerzel: '2',
    seiten: ['sets', 'karten'],
    dienst: false,
  },

  { schluessel: 'sammeln', kuerzel: '3', seiten: ['sammlung'], dienst: false },

  {
    schluessel: 'spielen',
    kuerzel: '4',
    seiten: ['decks', 'turniere', 'spielen'],
    dienst: false,
  },

  { schluessel: 'handeln', kuerzel: '5', seiten: ['handeln'], dienst: false },

  {
    schluessel: 'ressourcen',
    kuerzel: '6',
    seiten: ['pakete', 'vorlagen'],
    dienst: false,
  },

  {
    schluessel: 'programm',
    kuerzel: '',
    seiten: dienstseiten,
    dienst: true,
  },
]

// bereichsseiten sind alle Seiten, die in einem Bereich stehen.
export const bereichsseiten: readonly Seitenschluessel[] = bereiche.flatMap(
  (bereich) => bereich.seiten,
)

// bereichVon gibt den Bereich, in dem eine Seite steht — oder keinen.
//
// **Keiner ist die richtige Antwort für eine Lasche des Sets.** Sie steht in
// der Laschenleiste und nicht in der Hauptnavigation; ihr einen Bereich
// anzudichten hieße, sie zweimal zu führen.
export function bereichVon(seite: Seitenschluessel): Bereich | undefined {
  return bereiche.find((bereich) => bereich.seiten.includes(seite))
}

// ersteSeiteVon gibt die Seite, auf die das Kürzel eines Bereichs führt.
//
// **Die erste und nicht eine gemerkte.** Ein Kürzel, das dorthin zurückführt,
// wo man zuletzt war, ist bequem und unvorhersehbar: dieselbe Taste landet an
// zwei Tagen an zwei Orten. Wer weiterarbeiten will, nimmt „Weiterarbeiten"
// auf der Startseite.
export function ersteSeiteVon(bereich: Bereich): Seitenschluessel {
  const [erste] = bereich.seiten

  /* v8 ignore next -- jeder Bereich hat mindestens eine Seite, und
     `bereichsordnung.test.ts` hält das; TypeScript sieht es nicht. */
  return erste ?? 'start'
}

// ausserhalbDerBereiche sind die registrierten Seiten, die in keinem stehen.
//
// **Sie sind die Laschen des offenen Sets** — und die Liste steht hier, damit
// der Wächter daneben in **beide** Richtungen zählen kann: keine Seite ohne
// Bereich, die nicht hier steht, und keine hier, die es nicht gibt.
export const ausserhalbDerBereiche: readonly Seitenschluessel[] = [
  ...setschluessel,
].filter((schluessel) => !bereichsseiten.includes(schluessel))

// alleSeitenschluessel ist die Menge, gegen die der Wächter zählt.
export const alleSeitenschluessel: readonly Seitenschluessel[] = [
  ...anwendungsschluessel,
  ...setschluessel,
]

// bereichFuer gibt den Bereich, den die Leiste hervorhebt.
//
// **Er ist nicht dasselbe wie `bereichVon`.** Jener sagt, welcher Bereich eine
// Seite **führt** — und für eine Lasche des offenen Sets ist das keiner, denn
// sie steht in der Laschenleiste. Dieser sagt, welcher Eintrag **leuchtet**,
// und da ist die Antwort *Gestalten*: wer ein Set offen hat, arbeitet dort.
//
// **Der Entwurf zeichnet es so.** Die Artboards *Karten*, *Eigenschaften*,
// *Werkzeuge* und *Ausgeben* tragen alle `aktiv: 'gestalten'`
// (`prototyp/teile/` (Entwicklungsarchiv)). Ohne diesen Rückfall
// leuchtete kein Eintrag, sobald jemand eine Karte bearbeitet — und die
// Leiste sagte auf der meistbenutzten Fläche der Anwendung nichts.
export function bereichFuer(seite: Seitenschluessel): Bereich | undefined {
  return bereichVon(seite) ?? bereichVon('sets')
}
