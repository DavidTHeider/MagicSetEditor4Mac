// Ein Fehler, der einen **Schlüssel** trägt statt eines Satzes.
//
// # Warum er eine eigene Datei hat
//
// Weil `api.ts` ihn wirft und `meldung.ts` ihn liest — und `meldung.ts`
// zugleich `Bindungsfehler` aus `api.ts` braucht. Beides in eine Datei zu legen
// gäbe einen Ringschluss; ES-Module tragen ihn, aber ein Ring, den man beim
// nächsten Umbau nicht sieht, ist kein Entwurf, sondern ein Zufall.
//
// # Warum es ihn gibt
//
// #133: ein
// `new Error('Die Vorschau kam …')` bringt einen deutschen Satz in den
// Quelltext, den kein Katalog übersetzt und den `keintextimjsx.test.ts` nicht
// sieht — es liest JSX. Der Wächter `keinfehlertext.test.ts` weist ihn ab.
//
// **Die Meldung des `Error` ist der Schlüssel**, nicht ein Satz: ein
// Rückfalltext wäre wieder ein deutscher Satz im Quelltext, und die Regel hätte
// ein Loch in ihrer eigenen Umsetzung.
import type { Katalog } from './sprache/kataloge'

export type Meldungsschluessel = keyof Katalog['allgemein']['meldungen']

export class Katalogfehler extends Error {
  readonly schluessel: Meldungsschluessel
  readonly werte: Record<string, string>

  constructor(
    schluessel: Meldungsschluessel,
    werte: Record<string, string> = {},
  ) {
    super(schluessel)
    this.name = 'Katalogfehler'
    this.schluessel = schluessel
    this.werte = werte
  }
}
