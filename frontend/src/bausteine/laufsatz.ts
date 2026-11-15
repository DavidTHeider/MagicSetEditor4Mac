// **Der gezählte Stand eines Datei-Laufs als Satz** — FMT-03.AK-02 (#1692):
// „12 von 300 Karten", „3 Bilder". Die Einheit kommt als Schlüssel aus Go;
// der Laufstand einer Seite und die Übersicht der Hintergrundaufgaben setzen
// denselben Satz.
import { einsetzen } from '../sprache/einsetzen'
import { zaehlwort } from '../sprache/zaehlwort'
import type { AuftragDTO } from '../apiaufmerksamkeit'
import type { Katalog } from '../sprache/kataloge'

export type Laufstandtexte = Katalog['bausteine']['laufstand']
type Einheit = keyof Laufstandtexte['einheiten']

function istEinheit(
  texte: Laufstandtexte,
  einheit: string,
): einheit is Einheit {
  return einheit in texte.einheiten
}

// satzAus setzt den gezählten Stand in der Sprache des Fensters.
export function laufsatz(texte: Laufstandtexte, auftrag: AuftragDTO): string {
  if (!istEinheit(texte, auftrag.einheit)) {
    return texte.laeuft
  }

  if (auftrag.gesamt > 0) {
    const form = texte.einheitenVon[auftrag.einheit]

    return einsetzen(texte.vonGesamt, {
      erledigt: String(auftrag.erledigt),
      gesamt: String(auftrag.gesamt),
      einheit: auftrag.gesamt === 1 ? form.eine : form.viele,
    })
  }

  return zaehlwort(auftrag.erledigt, texte.einheiten[auftrag.einheit])
}
