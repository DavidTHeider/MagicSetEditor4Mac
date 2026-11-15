// **Die Epoche des Bestands, die das Fenster zuletzt erfuhr** — J05
// (#2269,
// zu #2292).
//
// Go zählt sie bei jedem Tausch des Bestands hoch und nennt sie in der
// Antwort von Neuanfang und Zurückspielen und in jeder Finder-Meldung. Wails
// stellt Ereignisse in eine eigene Schlange, und eine Antwort kann sie
// überholen; eine Kennung aus einer älteren Epoche gehört zu einem Bestand,
// den das Fenster nicht mehr zeigt.
//
// **An einer Stelle und nicht in einem Haken**: sie kommt aus drei Wegen —
// zwei Antworten, auch gescheiterten, und dem Ereignis selbst —, und ein
// Wert, der durch drei Haken gereicht würde, ginge an einem verloren. Hier
// stand bis zum Nachlauf über J05 eine Referenz in `Anwendung`, die nur eine
// geglückte Antwort erreichte.

let bekannt = 0

// geloescht sind die Sets, die in **diesem** Bestand gelöscht wurden (#2306).
// `sets.id` trägt `AUTOINCREMENT` und wurde seit Wanderung 0002 nie neu
// angelegt: eine Kennung kehrt in einem Bestand nicht wieder, in einem neuen
// schon — deshalb gilt die Liste bis zur nächsten Epoche.
//
// **Eine Kennung statt einer Epoche**, weil ein Löschen nicht den ganzen
// Bestand ungültig macht: eine Finder-Meldung für ein **anderes** Set, die
// kurz davor eingelesen wurde, gilt weiter. Ein Hochzählen beim Löschen
// verwarf genau sie (Nachlauf über J05, Befund 3).
const geloescht = new Set<string>()

// epocheGesehen merkt sich eine Epoche, wenn sie höher ist als die bekannte.
export function epocheGesehen(epoche: number | undefined): void {
  if (epoche !== undefined && epoche > bekannt) {
    bekannt = epoche
    geloescht.clear()
  }
}

// epocheBekannt nennt die Epoche, die das Fenster zuletzt erfuhr — ein
// Empfänger, der sich Fassungen merkt, beginnt mit ihr (#2307).
export function epocheBekannt(): number {
  return bekannt
}

// kennungGeloescht merkt sich ein Set, das gelöscht ist — **nur in der
// Epoche, in der das Löschen begann** (Zweitblick über die Folgen vor 0.8.0,
// Befund 5). Wurde inzwischen getauscht, kann dieselbe Kennung ein anderes
// Set meinen, und die Liste gehört dem neuen Bestand.
export function kennungGeloescht(setID: string, epoche: number): void {
  if (epoche === bekannt) {
    geloescht.add(setID)
  }
}

// kennungIstGeloescht sagt, ob ein Set in diesem Bestand gelöscht wurde.
export function kennungIstGeloescht(setID: string): boolean {
  return geloescht.has(setID)
}

// epocheVeraltet sagt, ob eine Epoche älter ist als die bekannte. **Ohne
// Angabe nicht** — eine Meldung ohne Epoche stammt aus einem Go, das sie
// noch nicht kannte, und ein Verwerfen hieße, gar nichts mehr zu öffnen.
export function epocheVeraltet(epoche: number | undefined): boolean {
  return epoche !== undefined && epoche < bekannt
}

// epocheVergessen setzt zurück — für Tests, die ein frisches Fenster bauen.
export function epocheVergessen(): void {
  bekannt = 0
  geloescht.clear()
}
