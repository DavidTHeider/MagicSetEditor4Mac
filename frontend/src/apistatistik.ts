// **Was ein Set zählt** — Statistik und Verteilung.
//
// **Eine eigene Datei**, weil `api.ts` mit ihnen über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort steht, **was man mit einem Set tut**, hier,
// **was man über eines erfährt**, ohne es zu ändern.
import type { StatistikAntwort, VerteilungAntwort } from './antwortenstatistik'

export interface Statistikapi {
  // Statistik zählt, was in einem Set steht.
  Statistik: (setID: string) => Promise<StatistikAntwort>

  // Verteilung zählt es je Wert **eines** Feldes.
  Verteilung: (setID: string, feld: string) => Promise<VerteilungAntwort>
}
