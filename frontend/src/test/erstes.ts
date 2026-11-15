// **Der Kopf einer Liste, von der der Lauf weiß, dass sie einen hat** —
// Werkstück G3 von Phase 5.5
// (#1460).
//
// **`noUncheckedIndexedAccess` weiß es nicht**, und ein `!` an jeder Stelle
// wäre dieselbe Behauptung ohne Satz. Hier steht sie einmal, mit einem Wurf,
// der sagt, was schiefging — dieselbe Gestalt wie `element()` in
// `partieprobe.ts`.

export function erstes<T>(liste: readonly T[]): T {
  const [eines] = liste
  if (eines === undefined) {
    // keinfehlertext: eine Vorrichtung für Läufe, nie im Fenster
    throw new Error('die Liste ist leer')
  }

  return eines
}

// an ist dasselbe für eine bestimmte Stelle.
export function an<T>(liste: readonly T[], stelle: number): T {
  const eines = liste[stelle]
  if (eines === undefined) {
    // keinfehlertext: eine Vorrichtung für Läufe, nie im Fenster
    throw new Error(`keine Stelle ${String(stelle)}`)
  }

  return eines
}
