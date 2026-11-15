// Welche Karte gewählt ist — und welche es nach einem Löschen wird.
//
// **Eine eigene Datei**, weil `Karten.tsx` an der Zeilengrenze steht (§2) — und
// weil der Schnitt einer Frage folgt: hier steht, *wer* aufgeschlagen ist, dort,
// *was* mit ihm geschieht. Beides sind reine Funktionen und lassen sich
// einzeln prüfen.
import type { KarteDTO, KartenzeileDTO } from '../api'

// gewaehlteKarte sucht die gewählte Karte in der geladenen Liste.
//
// **Aus der Liste und nicht aus einem eigenen Zustand.** Sonst führte der
// Editor eine zweite Fassung derselben Karte, und nach einem Nachladen zeigten
// beide etwas anderes.
export function gewaehlteKarte(
  karten: KartenzeileDTO[] | undefined,
  gewaehlt: string | undefined,
): KartenzeileDTO | undefined {
  return karten?.find((karte) => karte.id === gewaehlt)
}

// nachrueckendeKarte sagt, was nach dem Löschen aufgeschlagen wird.
//
// **ported.** MSE2 merkt sich die *Stelle* der gelöschten Karte, baut die Liste
// neu und wählt dann, was an dieser Stelle nun steht
// (`src/gui/control/card_list.cpp:78–89`, Fassung f9d9356 vom 2021-07-06). Das
// ist die nachrückende Karte. Steht dort nichts mehr — die gelöschte war die
// letzte —, greift `selectItemPos` auf `sorted_list.back()` zurück, also auf
// die letzte verbliebene (`src/gui/control/item_list.cpp:87–95`, dieselbe
// Fassung). Ist die Liste leer, bleibt nichts gewählt.
//
// **Die Stelle und nicht der Nachbar.** Beides fällt in der Mitte zusammen; am
// Ende nicht, und dort ist MSEs Antwort die vorige Karte.
//
// **Auch für die Schlüsselwörter** (UXR-F3,
// #2009):
// `KeywordList` und `CardList` erben dieselbe `ItemList`, und das Entfernen
// wählt in beiden an derselben Stelle. Die Funktion braucht von einer Zeile
// nur die Kennung — deshalb nimmt sie diesen Ausschnitt und nicht
// `KartenzeileDTO`.
export function nachrueckendeKarte(
  karten: readonly { readonly id: string }[],
  geloescht: string,
): string | undefined {
  const stelle = karten.findIndex((karte) => karte.id === geloescht)
  const uebrig = karten.filter((karte) => karte.id !== geloescht)

  // **Ohne Wächter auf die leere Liste**, weil der Zugriff sie schon
  // beantwortet: bleibt keine Karte übrig, ist beides `undefined`, und
  // aufgeschlagen wird nichts. Ein Wächter daneben wäre ein Zweig, den kein
  // Test je betritt — und einer, den niemand mehr auf seine Wirkung prüft.
  return (uebrig[stelle] ?? uebrig[uebrig.length - 1])?.id
}

// leereKarte ist die Karte, solange ihre Feldwerte noch laden.
//
// **Und nicht `null`.** Der Editor zeichnet dann seine Felder aus den Formen
// des Sets — leer, aber vollständig. Ein Herzschlag ohne Editor wäre ein
// Springen, und ein `undefined` wäre ein Zweig in jeder Zeile darunter.
export function leereKarte(zeile: KartenzeileDTO): KarteDTO {
  return { id: zeile.id, nummer: zeile.nummer, werte: [] }
}
