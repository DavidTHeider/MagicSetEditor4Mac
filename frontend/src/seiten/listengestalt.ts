// **Die Gestalt der Kartenliste, als ein Bündel** — Werkstück 3 von Phase 3.7
// (#1168).
//
// **Ein Verbund und keine sechs Eigenschaften**: sie gehören zusammen und
// werden zusammen durchgereicht — durch `Karten.tsx` bis in die Listenspalte.
// Sechs Stellen nebeneinander sagen am Aufrufort nichts mehr, und wer eine
// vergisst, merkt es erst am Typfehler drei Ebenen tiefer.
import type { ListenspalteDTO } from '../api'
import type { Vorgang } from '../haken/useVorgang'
import type { Listenzusatz } from './Kartenliste'

export interface Listengestalt {
  // spalten sind **alle** angebotenen Spalten — auch die unsichtbaren, denn
  // der Wahlkasten zeigt sie.
  spalten: readonly ListenspalteDTO[]

  // zusaetze sind die **sichtbaren** neben dem Namen, samt ihren Werten.
  zusaetze: readonly Listenzusatz[]

  sortierfeld: string
  aufsteigend: boolean

  // ordnung ist die Reihenfolge der Kartenkennungen, die der Dienst berechnet
  // hat — leer, solange nichts geladen ist.
  //
  // # Warum sie hier steht
  //
  // **Weil sie sonst nirgends ankam.** Der Dienst sortiert seine Zeilen, die
  // Bindung gibt sie so zurück — und dieser Haken machte daraus nur eine
  // Zuordnung von Werten nach Kennung. Die Liste bekam ihre Karten weiter aus
  // `Kartenzeilen`, und das ordnet nach `cards.reihenfolge`: **die gewählte
  // Sortierung war unsichtbar** (Befund 1 des Zweitblicks über Phase 3.7).
  //
  // **Eine Kennungsliste und nicht die Karten selbst.** Die Zeilen des
  // Aufbaus tragen keine Feldwerte für die Liste — die kommen aus
  // `useKartendaten`, und zwei Quellen für dieselbe Karte liefen auseinander.
  ordnung: readonly string[]

  // marken ist Kartenkennung → ihre Marken, alphabetisch
  // (#1208).
  //
  // **Mit dem Aufbau und nicht in einem zweiten Ruf**: sie werden je Zeile
  // gezeigt und je Zeile gefiltert, und eine Abfrage je Zeile wären bei 3.788
  // Karten 3.788 Abfragen.
  marken: Readonly<Record<string, readonly string[]>>

  // fehler ist die Meldung des **Ladens** — `undefined`, solange es glückt.
  //
  // **Sie steht im Verbund und nicht in einem Rückruf nach oben.** Wer die
  // Gestalt nicht laden kann, sieht keine Spalten und keine Ordnung; die
  // Meldung gehört an denselben Ort, an dem sie fehlen. Ohne sie wäre der
  // Kasten leer, und §8 sagt, dass das die schlechteste Antwort ist.
  fehler: string | undefined

  // ohneVorlage sagt, dass die Spalten aus den Karten stammen und nicht aus
  // dem Spielpaket — dann fehlen Aufschrift und Reihenfolge.
  ohneVorlage: boolean

  vorgang: Vorgang
  geaendert: () => void
}
