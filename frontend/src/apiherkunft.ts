// **Woher ein Deck und woher eine Sammlungszeile kam** — Werkstück 12 von
// Phase 4.12
// (#1351).
//
// # Warum es dafür keine eigene Bindung gibt
//
// **Weil die Herkunft mit dem Ding reist, dessen Herkunft sie ist.** Ein
// zweiter Ruf je Datensatz wäre eine Stelle, an die jede Seite denken müsste —
// und wer sie vergäße, zeigte ein übernommenes Deck als eigenes
// ([0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md)).

// HerkunftDTO sagt, woher ein Datensatz kam.
export interface HerkunftDTO {
  // `profil` oder `einfuhr` — **leer heißt: hier entstanden**, und dann zeigt
  // die Seite nichts an. Eine Zeile „Herkunft: —" wäre eine Auskunft über
  // nichts.
  art: string

  // Die lesbare Bezeichnung, abgeschrieben zum Zeitpunkt der Übernahme: ein
  // Profilname, ein Dateiname.
  quelle: string

  // Leer, wenn es keines gab.
  quellProfil: string

  uebernommen: string
}

// keineHerkunft ist die leere — „hier entstanden".
//
// **Exportiert, damit sie nicht an elf Stellen nachgebaut wird.** Eine
// Vorrichtung, die sie je für sich zusammensetzt, prüft ihre eigene
// Zusammensetzung und nicht den Weg (§3).
export const keineHerkunft: HerkunftDTO = {
  art: '',
  quelle: '',
  quellProfil: '',
  uebernommen: '',
}
