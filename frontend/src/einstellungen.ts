// Die Anwendungseinstellungen, wie das Fenster sie sieht —
// #245.
import type { Umschlag } from './api'

// EinstellungDTO ist eine Einstellung samt ihrer möglichen Werte.
//
// **Die Werte kommen mit.** Ohne sie baute das Fenster die Auswahlliste selbst
// — und dann stünden die erlaubten Werte an zwei Stellen, von denen die hier
// die ungeprüfte wäre.
export interface EinstellungDTO {
  schluessel: string

  // wert ist, was **gewählt** ist — bei den vier Erscheinungseinstellungen
  // vielleicht `erben`.
  wert: string

  // wirksam ist, was **gilt** — Werkstück 9 von Phase 4.12
  // (#1351),
  // Punkt 9.
  //
  // **Er weicht von `wert` genau dann ab, wenn dort `erben` steht**: dann
  // gilt, was beim Menschen gewählt ist. Die Auswahlliste zeigt `wert`, die
  // Anwendung richtet sich nach `wirksam`.
  wirksam: string

  werte: string[]

  // bereich sagt, **wo** die Einstellung steht — `anwendung` oder
  // `bildausfuhr` (#491).
  //
  // Nicht jede Einstellung gehört auf die Einstellungsseite: die Wahl beim
  // Namenskonflikt gehört zu dem Ausgabelauf, den man gerade startet.
  bereich: string

  // **Ebene, Herkunft und Vorgabe** — Werkstück U11 von Phase 5.1
  // (#1422):
  // die Go-Seite schickt sie immer; wahlweise sind sie nur, damit die
  // Vorrichtungen der Tests, die sie nicht meinen, sie weglassen dürfen.
  ebene?: string
  herkunft?: string
  // wirkung sagt, **wann** ein geänderter Wert gilt — `sofort` oder
  // `neustart` (UX-04.AK-02,
  // #1676).
  //
  // **Wahlweise wie die drei darüber**: ein Bestand von morgen kann einen
  // Schlüssel tragen, für den die Go-Seite nichts sagt, und dann steht hier
  // nichts statt einer Behauptung.
  wirkung?: string

  vorgabe?: string
}

// EinstellungenAntwort trägt die Einstellungen.
export interface EinstellungenAntwort extends Umschlag {
  einstellungen: EinstellungDTO[]
}
