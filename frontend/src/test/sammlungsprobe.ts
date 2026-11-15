// **Was die Läufe über das Sammlungsfenster gemeinsam brauchen.**
//
// Aus `sammlungsfenster.test.tsx` herausgelöst, als es über 400 Zeilen ging
// (§2) — und damit die zweite Datei daneben dieselbe Probe benutzt statt einer
// zweiten, die anders aussieht.
import type { ExemplarDTO } from '../api'
import { platzhalter } from '../haken/sammlungszeilen'

export const bestand = 1000

// exemplar nimmt die Felder, die dieser Test nicht braucht, vom Platzhalter —
// geprüft wird die Zuordnung einer Antwort zu ihrer Sicht, nicht ein Feld.
export function exemplar(stelle: number, name: string): ExemplarDTO {
  return { ...platzhalter(stelle), id: stelle + 1, name, bezeichnung: name }
}

// belegeFenster antwortet auf das erste Fenster sofort und hält jedes weitere
// zurück, bis der Test es freigibt.
//
// **Eine andere Suche gibt andere Zeilen.** Gefiltert wird in Go, und das hier
// nachzubauen wäre eine zweite Wahrheit (§3) — aber eine Sicht, die dieselben
// Zeilen gäbe, prüfte nicht, was mit einer Wahl geschieht, die in der neuen
// Sicht fehlt. Deshalb hängt hier nur die **Kennung** an der Suche.
