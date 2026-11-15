// **Was im Bestand herumliegt** — Werkstück O8 von Phase 5.2
// (#1432).
// Die Gestalt kommt aus `internal/app/datenpflege.go`.
import type { Umschlag } from './umschlag'

// PflegebefundDTO ist eine Auffälligkeit im Bestand.
export interface PflegebefundDTO {
  // art und klasse sind Marken — der Satz dazu steht im Katalog.
  art: string
  klasse: string

  anzahl: number

  // revision ist die Prüfsumme über die betroffenen Zeilen. Sie reist mit
  // zurück: wer sie mitschickt, sagt, welche Zeilen er gesehen hat.
  revision: string

  beispiele: string[]
}

export interface PflegevorschauAntwort extends Umschlag {
  erstellt: string
  gueltigbis: string
  befunde: PflegebefundDTO[]
}

export interface PflegewahlDTO {
  art: string
  revision: string
}

export interface PflegezeileDTO {
  art: string
  ausgang: string
  anzahl: number
  grund: string
}

export interface PflegebilanzAntwort extends Umschlag {
  // gesichert ist der Ort der Sicherung — leer, wenn nichts zu tun war.
  gesichert: string

  zeilen: PflegezeileDTO[]

  // vorher und nachher sind die Befunde vor und nach der Pflege.
  vorher: PflegebefundDTO[]
  nachher: PflegebefundDTO[]
}

export interface Pflegeapi {
  Pflegevorschau: () => Promise<PflegevorschauAntwort>
  Pflegen: (
    erstellt: string,
    wahl: PflegewahlDTO[],
  ) => Promise<PflegebilanzAntwort>
}
