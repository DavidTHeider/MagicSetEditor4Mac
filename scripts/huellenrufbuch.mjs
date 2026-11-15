// **Das Rufbuch einer Hüllenprobe** — B01
// (#2220).
//
// Der Abnahmebau schreibt jeden Bindungsruf des Fensters mit
// (`internal/fenster/rufbuch.go`); die Probe holt den Abzug vor jedem
// Beenden über die Leitung und mischt ihn über alle Starts eines Drehbuchs.
// Ausgewertet wird er in `scripts/bindungsreichweite.mjs`.
//
// **Und die Maus der Leitung** steht hier mit, seit `huellenprobe.mjs` an der
// Zeilengrenze steht: ziehen, loslassen, rechts klicken — Ereignisse ab
// AppKit, kein `dispatchEvent` in der Seite (A04, #2218).

// mischen legt einen Abzug in ein Rufbuch — je Bindung Anzahl, Ausgänge
// und Dauer zusammengezählt.
export function mischen(buch, abzug) {
  for (const [name, zahl] of Object.entries(abzug ?? {})) {
    const eintrag = buch[name] ?? { anzahl: 0, ausgaenge: {}, dauerMs: 0 }
    eintrag.anzahl += zahl.anzahl
    eintrag.dauerMs += zahl.dauerMs
    for (const [art, anzahl] of Object.entries(zahl.ausgaenge ?? {})) eintrag.ausgaenge[art] = (eintrag.ausgaenge[art] ?? 0) + anzahl
    buch[name] = eintrag
  }

  return buch
}

// rufbuchHolen holt den Abzug der laufenden Anwendung und leert ihn dort —
// **ein Fehler kostet nur die Mitschrift**, nicht das Drehbuch.
export async function rufbuchHolen(probe) {
  probe.rufbuch ??= {}
  try {
    const antwort = await probe.leitung.frage({ was: 'rufbuch', leeren: true })
    if (antwort.ok) mischen(probe.rufbuch, antwort.daten)
  } catch {
    // die Leitung ist schon zu — dann gibt es nichts mehr zu holen
  }

  return probe.rufbuch
}

export const ziehe = (p, weg, halten = false) => p.frage({ was: 'ziehe', weg, halten })
export const loslassen = (p, x, y) => p.frage({ was: 'loslassen', x, y })
export const rechtsklick = (p, x, y) => p.frage({ was: 'rechtsklick', x, y })

// stoere verzögert Bindungen oder lässt sie fachlich scheitern — B03
// (#2222),
// `internal/fenster/rufstoerung.go`. `[]` hebt alle auf.
export const stoere = (p, stoerungen) => p.frage({ was: 'stoere', stoerungen })
