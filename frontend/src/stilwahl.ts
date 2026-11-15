// Die Typen der Stilauswahl (#41).
//
// **Eine eigene Datei**, wie `pakete.ts` und `vorlagen.ts`: die Fassade nennt
// die Methoden, die Typen stehen bei ihrem Gegenstand (§2).
import type { Umschlag } from './api'
import type { FeldbeschreibungDTO } from './formen'

// StilDTO ist ein wählbarer Stil.
export interface StilDTO {
  paket: string

  // anzeigename ist, was über dem Stil steht — wie bei `PaketDTO`.
  anzeigename: string
  spiel: string
  version: string

  // eigenesSpiel sagt, ob der Stil zum Spiel des Sets gehört. Ein fremder wird
  // **gezeigt** und nicht weggefiltert — er ist eine Möglichkeit mit Preis.
  eigenesSpiel: boolean
  gewaehlt: boolean
}

export interface StileAntwort extends Umschlag {
  stile: StilDTO[]
}

// FolgenDTO sagt, was ein Wechsel kostet.
export interface FolgenDTO {
  // ungedeuteteFelder gehen **nicht** verloren: der Bestand führt jeden
  // Feldwert mit. Was verlorengeht, ist ihre Darstellung.
  ungedeuteteFelder: string[]

  // verloreneOptionen gehen wirklich verloren.
  verloreneOptionen: string[]
  spielwechsel: boolean
}

export interface FolgenAntwort extends Umschlag {
  folgen: FolgenDTO | null
}

// StiloptionDTO ist eine Stiloption samt Wert.
export interface StiloptionDTO {
  name: string
  art: string

  // beschreibung kommt aus dem Paket und ist deshalb unübersetzt (§9).
  beschreibung: string
  wert: string

  // moeglichkeiten sind die Wahlen der Option — Wert und Beschriftung.
  //
  // **Eine Liste für jede Art, die eine Auswahl ist**: `choice`, `boolean`,
  // `package choice` und `color`. Bei den ersten beiden sind Wert und
  // Beschriftung dasselbe; bei einer Paketwahl steht
  // `probe-overlay-folie.mse-include` gegen „Folie", bei einer Farbe
  // `rgb(0,0,0)` gegen „schwarz" (#632, #634).
  moeglichkeiten: Wahlmoeglichkeit[]

  // eigeneFarbe sagt, ob eine Farbe außerhalb der Liste erlaubt ist — bei
  // jeder anderen Art als `color` falsch (#634).
  eigeneFarbe: boolean

  // gesetzt sagt, ob `wert` aus dem Bestand kommt oder aus der Vorgabe.
  // vorgabe ist der **gerechnete** Wert aus `default:` — leer, wenn die
  // Vorlage keinen nennt.
  //
  // **Zwei Angaben für einen Eintrag** (#764): `vorgabe` sagt, ob es ihn gibt
  // und welche Farbe er zeigt, `gesetzt`, ob er gewählt ist. MSE blendet ihn
  // ohne `default_script` ganz aus.
  gesetzt: boolean

  // hatVorgabe sagt, ob die Vorlage überhaupt ein `default:` nennt — **nicht
  // dasselbe wie ein nichtleeres `vorgabe`**. Ein Skript darf scheitern;
  // `magic-m15-bigtext` schreibt `default: black` ohne Anführungszeichen, und
  // MSE zeigt seinen Eintrag trotzdem.
  hatVorgabe: boolean
  vorgabe: string

  // hatWahlbilder sagt, ob es zu den Wahlen Bilder gibt
  // (#1225).
  //
  // **Damit die Seite nur dort nachfragt, wo es etwas zu holen gibt.** Ohne
  // die Auskunft riefe sie `Stilwahlbilder` für jede Option — und bekäme bei
  // `magic-m15` neunmal von dreizehn eine leere Liste.
  hatWahlbilder: boolean
}

// WahlbildDTO ist das Bild einer Wahl.
export interface WahlbildDTO {
  // wahl ist der Name der Wahl, wie die Option ihn erklärt — bei einer Farbe
  // also `schwarz` und nicht `rgb(0,0,0)`. Zugeordnet wird über die
  // **Beschriftung** einer `Wahlmoeglichkeit`.
  wahl: string

  // bild ist ein PNG als Data-URI — unmittelbar in ein `src` einsetzbar.
  bild: string
}

export interface WahlbilderAntwort extends Umschlag {
  bilder: WahlbildDTO[]

  // befunde sind die Wahlen, deren Bild nicht entstanden ist. Sie stehen
  // **neben** den Bildern: zwei von drei Tapsymbolen sind besser als keines.
  befunde: string[]
}

// Wahlmoeglichkeit ist eine Möglichkeit einer Stiloption.
export interface Wahlmoeglichkeit {
  wert: string
  beschriftung: string
}

export interface StiloptionenAntwort extends Umschlag {
  optionen: StiloptionDTO[]
}

// ZusatzfeldDTO ist ein Zusatzfeld des Stils einer Karte samt ihrem Wert
// (#2298).
//
// **Die Form ist die eines Kartenfelds** — das Formular zeichnet es mit
// denselben Bauteilen.
export interface ZusatzfeldDTO {
  form: FeldbeschreibungDTO
  wert: string

  // gespeichert sagt, ob die Karte überhaupt einen Wert trägt — sonst gilt
  // die Vorgabe des Stils.
  gespeichert: boolean
}

export interface ZusatzfelderAntwort extends Umschlag {
  felder: ZusatzfeldDTO[]
}
