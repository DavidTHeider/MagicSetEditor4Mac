// **Wie weit ein Zug ist** — der Balken unter der Zeile des Zuges, Artboard
// *Tisch*.
//
// # Warum eine feste Liste
//
// **Weil das Regelwerk die Reihenfolge festlegt** und nicht der Kern: CR 500.1
// nennt die fünf Phasen, CR 501.1 die drei Schritte des Anfangs, CR 506.1 die
// fünf des Kampfs und CR 512.1 die zwei des Endes (Fassung 2026-09-25). Die
// Hauptphasen haben keinen Schritt. Die Kennungen sind die des Kerns
// (`internal/domain`), dieselben, die `Zugleiste` in Worte fasst.

// abschnitte sind Phase und Schritt in der Reihenfolge eines Zuges.
const abschnitte: readonly (readonly [string, string])[] = [
  ['anfang', 'enttappen'],
  ['anfang', 'versorgung'],
  ['anfang', 'ziehen'],
  ['hauptVorKampf', ''],
  ['kampf', 'kampfbeginn'],
  ['kampf', 'angreifer'],
  ['kampf', 'blocker'],
  ['kampf', 'kampfschaden'],
  ['kampf', 'kampfende'],
  ['hauptNachKampf', ''],
  ['ende', 'endeschritt'],
  ['ende', 'aufraeumen'],
]

// Zugstand ist die Stelle eines Abschnitts unter allen.
export interface Zugstand {
  stelle: number
  von: number
}

// zugfortschritt gibt die Stelle eines Abschnitts — oder nichts, wo er
// unbekannt ist. **Ein unbekannter Abschnitt ist kein Fehler** (§8), aber
// auch kein Stand, den man zeichnen könnte.
export function zugfortschritt(
  phase: string,
  schritt: string,
): Zugstand | undefined {
  const stelle = abschnitte.findIndex(
    ([einePhase, einSchritt]) => einePhase === phase && einSchritt === schritt,
  )

  return stelle < 0 ? undefined : { stelle: stelle + 1, von: abschnitte.length }
}
