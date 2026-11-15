// **Das eigene Tischbild an der Naht** — Werkstück G3 von Phase 5.5
// (#1460).
//
// **Eine Adresse und keine Bytes** — dieselbe Entscheidung wie bei den
// Spieltexturen (G2): ein Hintergrund von 1920 Punkten käme als Base64 über
// die Wails-Brücke bei jedem Öffnen des Tisches erneut, und 182 KB kosten
// dort gemessen 554 ms.
import type { Umschlag } from './umschlag'

// TischbildAntwort sagt, ob ein eigenes Bild vorliegt.
export interface TischbildAntwort extends Umschlag {
  // adresse ist leer, solange keines abgelegt ist.
  adresse: string
}

// Tischbildapi sind die drei Wege.
//
// **Kein Weg mit einem Pfad**, und das ist dieselbe Überlegung wie beim
// Profilbild: das Bild kommt über den Dateidialog herein, und der gehört nach
// Go — ein Fenster bekäme einen Namen und keinen Pfad.
export interface Tischbildapi {
  // Tischbild sagt, ob dieser Mensch eines hat.
  Tischbild: (nutzerID: string) => Promise<TischbildAntwort>

  // TischbildWaehlen öffnet den Dateidialog und legt das Bild ab.
  //
  // **Ein Abbruch ist kein Fehler**: wer den Dialog schließt, bekommt zurück,
  // was vorher galt.
  TischbildWaehlen: (nutzerID: string) => Promise<TischbildAntwort>

  // TischbildEntfernen nimmt es wieder weg — danach gilt wieder einer der
  // sechs Sätze.
  TischbildEntfernen: (nutzerID: string) => Promise<TischbildAntwort>

  // **Das Fensterbild** — UX-09.AK-01
  // (#1681).
  // Dieselben drei Wege und dieselbe Antwort; **ob** es erscheint und wie
  // groß, sagt die Einstellung `hintergrund`.
  Fensterbild: (nutzerID: string) => Promise<TischbildAntwort>
  FensterbildWaehlen: (nutzerID: string) => Promise<TischbildAntwort>
  FensterbildEntfernen: (nutzerID: string) => Promise<TischbildAntwort>

  // TischabzugSpeichern fragt nach einem Ort und legt einen Abzug dorthin —
  // PLAY-05.AK-03
  // (#1713).
  //
  // **Hier gehen Bytes hinaus und keine Adresse**, anders als bei den drei
  // Wegen darüber: der Abzug entsteht auf der Leinwand und liegt nirgends,
  // bevor er geschrieben ist. Er geht **einmal** über die Brücke, auf einen
  // Klick hin — nicht bei jedem Öffnen des Tisches.
  TischabzugSpeichern: (datenURL: string) => Promise<TischabzugAntwort>

  // TischabzugPDFSpeichern legt dasselbe Bild auf ein Blatt Papier —
  // PLAY-05.AK-03.
  //
  // **Das Blatt geht mit**: `a4` oder `letter`. Ein unbekanntes Format ist
  // ein Befund und kein stilles A4 (§8).
  TischabzugPDFSpeichern: (
    datenURL: string,
    blatt: string,
  ) => Promise<TischabzugAntwort>
}

// TischabzugAntwort ist der Ausgang eines Abzugs.
export interface TischabzugAntwort extends Umschlag {
  pfad: string

  // abgebrochen sagt, dass niemand einen Ort gewählt hat — **kein Fehler**.
  abgebrochen: boolean
}
