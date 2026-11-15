// **Die Partieverwaltung an der Naht** — PLAY-03.AK-05
// (#1711).
//
// **Eine eigene Datei**, weil `apipartie.ts` mit ihr über die Zeilengrenze
// aus §2 ginge. Der Schnitt folgt der Frage: dort steht die **Sicht** einer
// laufenden Partie, hier, was mit ihr als Ganzem geschieht.
import type { Umschlag } from './umschlag'

export interface Partieverwaltungapi {
  // PartieLoeschen nimmt eine Partie aus dem Bestand.
  //
  // **Geht nur bei einem Entwurf**: eine beendete bleibt nachvollziehbar,
  // und eine, für die jemand eine Einwilligung gegeben hat, bleibt auch.
  // Warum, sagt der Fehler.
  PartieLoeschen: (kennung: string) => Promise<PartieverwaltungAntwort>

  // PartieKopieren legt sie als neuen Übungszustand daneben.
  //
  // **Ohne Verlauf und ohne alte Zustimmungen** — und die Kopie läuft, auch
  // wenn die Vorlage beendet ist.
  PartieKopieren: (kennung: string) => Promise<PartieverwaltungAntwort>

  // PartieAusfuhrSpeichern fragt nach einem Ort und legt sie dorthin.
  //
  // **Darin steht der volle Zustand** — auch fremde Hände. Eine Lage ohne
  // ihre verdeckten Teile wäre eine andere.
  PartieAusfuhrSpeichern: (kennung: string) => Promise<PartieverwaltungAntwort>

  // PartieEinfuhrOeffnen fragt nach einer Datei und legt sie als neue Partie
  // an.
  PartieEinfuhrOeffnen: () => Promise<PartieverwaltungAntwort>
}

// PartieverwaltungAntwort ist der Ausgang eines dieser vier Wege.
export interface PartieverwaltungAntwort extends Umschlag {
  // kennung ist die Partie, die dabei entstanden ist — leer beim Löschen.
  kennung: string

  // pfad ist der Ort der Datei — leer, wo keine geschrieben wurde.
  pfad: string

  // inhalt ist die Partie als Text — im Fenster immer leer, denn hier geht
  // sie über einen Ort und nicht über den Arbeitsspeicher.
  inhalt: string

  // abgebrochen sagt, dass niemand einen Ort gewählt hat — **kein Fehler**.
  abgebrochen: boolean

  // befunde sind Auffälligkeiten der eingelesenen Datei — **ein Befund ist
  // kein Fehler** (§8).
  befunde: string[]
}
