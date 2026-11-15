// **Eine Laufzeit, die Horcher hält und auf Zuruf meldet** — FMT-03 (#1692).
//
// Bisher baute jeder Lauf, der auf ein Wails-Ereignis hört, seine eigene;
// hier steht sie einmal, mit **mehreren** Horchern je Ereignis — der
// Laufstand einer Seite und die Übersicht hören dasselbe.
interface MitLaufzeit {
  runtime?: {
    EventsOn: (
      name: string,
      rueckruf: (...daten: unknown[]) => void,
    ) => () => void
  }
}

export interface Laufzeitprobe {
  melde: (name: string, ...daten: unknown[]) => void
  horchende: (name: string) => number
  raeume: () => void
}

export function laufzeitprobe(): Laufzeitprobe {
  const horcher = new Map<string, Set<(...daten: unknown[]) => void>>()

  ;(globalThis as MitLaufzeit).runtime = {
    EventsOn: (name, rueckruf) => {
      const alle = horcher.get(name) ?? new Set()
      alle.add(rueckruf)
      horcher.set(name, alle)

      return () => {
        alle.delete(rueckruf)
      }
    },
  }

  return {
    melde: (name, ...daten) => {
      for (const rueckruf of horcher.get(name) ?? []) {
        rueckruf(...daten)
      }
    },
    horchende: (name) => horcher.get(name)?.size ?? 0,
    raeume: () => {
      delete (globalThis as MitLaufzeit).runtime
    },
  }
}
