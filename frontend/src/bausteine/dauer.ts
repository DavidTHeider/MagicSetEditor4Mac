// **Eine Dauer, wie jemand sie liest** — OPS-05
// (#1722).
// Eine eigene Datei, weil `Betriebszahlen` und `Laufzeitentafel` sie beide
// brauchen und einander sonst im Kreis einführten.

// alsDauer schreibt Millisekunden, ab einer Sekunde in Sekunden.
export function alsDauer(ms: number, darstellung: string): string {
  if (ms < 1000) {
    return `${ms.toLocaleString(darstellung, { maximumFractionDigits: 1 })} ms`
  }

  return `${(ms / 1000).toLocaleString(darstellung, { maximumFractionDigits: 1 })} s`
}
