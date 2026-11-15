// **Ob die Reihenfolge in einer Zone etwas bedeutet** — Befund **F04**
// (#1748).
//
// # Warum das Fenster es überhaupt wissen muss
//
// **Weil eine Stelle in einer ungeordneten Zone ein Irrtum ist und kein
// Wunsch.** Der Kern weist sie ab (`internal/service/sandkastenschieben.go`,
// `gemeinteStelle`) — und das ist richtig. Falsch war, dass das Formular sie
// überhaupt anbot: die Vorgabe der Zone ist `hand`, die Vorgabe der Stelle war
// `oben`, und die Auswahlliste kannte keinen neutralen Wert. Der
// Auslieferungszustand des Befehls war damit ein Fehler, und es gab nichts,
// womit man ihn behob.
//
// # Warum eine eigene Datei und keine Liste im Bauteil
//
// **Weil es eine Regelaussage ist und keine Darstellung.** `tisch/zonenfeld.ts`
// trägt `geordnet` schon, aber nur für die Felder, die ein **Sitz** hat — der
// gemeinsame Stapel kommt dort nicht vor. Wer die Frage dort nachschlüge,
// bekäme für `stapel` keine Antwort.
//
// # Gehalten wird sie gegen den Kern
//
// `internal/werkzeuge/zonenordnung_test.go` liest diese Liste und hält sie
// gegen `domain.Zone.Geordnet()`. Ohne ihn wären es zwei Antworten auf
// dieselbe Frage — und genau das war schon einmal der Fall: bis zum
// 2026-09-17 galt der Friedhof in `internal/domain` als ungeordnet, während
// der gezeichnete Tisch daneben CR 400.5 richtig wiedergab.

// geordneteZonen sind die drei, deren Reihenfolge sich nur ändern lässt, wenn
// eine Regel oder ein Effekt es erlaubt — CR 400.5 (Fassung 2026-08-19).
export const geordneteZonen = ['bibliothek', 'friedhof', 'stapel'] as const

// istGeordnet sagt, ob die Reihenfolge in dieser Zone etwas bedeutet.
//
// **Eine unbekannte Zone ist keine geordnete** — derselbe vorsichtige Satz
// wie in `domain.Zone.Geordnet()`: was niemand kennt, bekommt die Antwort,
// die nichts behauptet.
export function istGeordnet(zone: string): boolean {
  return (geordneteZonen as readonly string[]).includes(zone)
}
