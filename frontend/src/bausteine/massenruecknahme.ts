// **Die Rücknahme einer Massenänderung, Gruppe für Gruppe** — DOM-03.AK-03
// (#1698),
// aus `Massenvorgaenge.tsx` herausgelöst mit Befund 4 des Zweitblicks über
// Phase 6.6.
//
// **Je alter Lagerort ein Schreiben**, denn eine Massenänderung setzt einen
// Wert für alle ihre Zeilen. Scheitert eines, sind die davor geschrieben —
// und der Ausgang sagt es: was zurück ist, was noch aussteht, und ob
// überhaupt geschrieben wurde. Wer ihn liest, lädt dann nach (§7, Regel 3)
// und bietet den Rest wieder an, statt ihn zu verlieren.
import type {
  MassenberichtDTO,
  MassenstandDTO,
  MassenzeileDTO,
} from '../antwortensammlung'

// Ruecknahmeausgang ist, was nach der Rücknahme gilt.
export interface Ruecknahmeausgang {
  // zurueck zählt die Zeilen, die ihren alten Lagerort wieder haben.
  zurueck: number

  // befunde nennen, was ein Schreiben übergangen hat — meist eine Zeile,
  // die seither jemand anderes änderte.
  befunde: string[]

  // geschrieben sagt, ob mindestens ein Schreiben angekommen ist.
  geschrieben: boolean

  // offen sind die Zeilen, die nicht zurück sind — weil ihr Schreiben
  // nicht stattfand oder weil es sie übergangen hat.
  offen: MassenstandDTO[]

  // fehler ist der Grund, aus dem abgebrochen wurde — `undefined`, wenn
  // alles durchging.
  fehler: unknown
}

// ruecknahmeSchreiben schreibt Gruppe für Gruppe zurück und hört beim
// ersten Fehler auf.
export async function ruecknahmeSchreiben(
  stand: MassenstandDTO[],
  schreibe: (
    zeilen: MassenzeileDTO[],
    lagerort: string,
  ) => Promise<MassenberichtDTO | null>,
): Promise<Ruecknahmeausgang> {
  const gruppen = new Map<string, MassenstandDTO[]>()

  for (const zeile of stand) {
    gruppen.set(zeile.lagerort, [...(gruppen.get(zeile.lagerort) ?? []), zeile])
  }

  const ausgang: Ruecknahmeausgang = {
    zurueck: 0,
    befunde: [],
    geschrieben: false,
    offen: [],
    fehler: undefined,
  }

  for (const [ort, zeilen] of gruppen) {
    if (ausgang.fehler !== undefined) {
      ausgang.offen.push(...zeilen)

      continue
    }

    try {
      const bericht = await schreibe(
        zeilen.map((zeile) => ({ id: zeile.id, fassung: zeile.fassung })),
        ort,
      )

      ausgang.geschrieben = true
      ausgang.zurueck += bericht?.geaendert ?? 0
      ausgang.befunde.push(...(bericht?.befunde ?? []))

      // **Übergangen ist nicht erledigt** — Nachlauf über Phase 6.6. Der
      // Dienst schreibt „so viele wie möglich" und nennt in `vorher` nur, was
      // er schrieb; jede andere Zeile war ein Befund, und der kann auch
      // vorübergehend gewesen sein. Sie bleibt rücknehmbar: ein zweiter
      // Versuch überschreibt nichts, was seither jemand anderes schrieb — er
      // wird wieder zum Befund.
      const zurueck = new Set((bericht?.vorher ?? []).map((zeile) => zeile.id))
      ausgang.offen.push(...zeilen.filter((zeile) => !zurueck.has(zeile.id)))
    } catch (ursache: unknown) {
      ausgang.fehler = ursache
      ausgang.offen.push(...zeilen)
    }
  }

  return ausgang
}
