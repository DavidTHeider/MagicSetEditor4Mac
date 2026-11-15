// **Die Wege einer Partie im Fenster** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum sie neben [usePartie] stehen und nicht darin
//
// **Weil sie eine andere Frage beantworten** (§2): der Haken sagt, *was* das
// Fenster mit einer Partie tut; hier steht, *wie* eine Handlung benannt und
// adressiert wird — die Sperre, die Bindung und die Kennung. Der Schnitt
// folgt der Frage und nicht der Zeilenzahl; dass die Zeilengrenze ihn
// erzwungen hat, ändert nichts daran, dass es der richtige ist.
import { api, auspacken, type PartieAntwort } from '../api'
import type { Partie } from '../apipartie'

// Standwechsel sind die drei Wege, die eine laufende Partie anhalten.
export type Standwechsel = 'pausieren' | 'fortsetzen' | 'beenden'

// partievorgang ist die Sperre **einer Partie** — §7, Regel 4.
//
// # Warum ein Schlüssel und nicht zwei
//
// **Weil der Datensatz die Partie ist und nicht die Art des Schreibens.**
// [useVorgang] sperrt exakt gleiche Schlüssel; mit einem Schlüssel je Art
// könnten ein Spielbefehl und ein Anhalten derselben Partie
// **nebeneinander** starten.
//
// **Die Fassungskontrolle im Kern mildert den Schaden**, sie ersetzt die
// Regel aber nicht: der zweite Ruf ginge gegen einen Stand, den das Backend
// noch nicht kennt.
export function partievorgang(kennung: string): string {
  return `partie-${kennung}`
}

// standwegFuer gibt die Bindung zu einem Wort.
//
// **Drei Bindungen und ein Wort davor** — am Fenster sind es drei Knöpfe, an
// den anderen Flächen ein Argument. Die Fachregel ist dieselbe (§16).
//
// **Als Verbund und nicht als `switch`.** Ein `return api().PartieBeenden`
// am Zeilenende ist für `internal/werkzeuge/bindungsaufrufer_test.go` **kein
// Aufruf**: er verlangt die Gestalt, in der ein Aufruf aussieht — ein `(`,
// ein `,` oder ein `)` hinter dem Namen. Hier steht das Komma, und damit ist
// die Weitergabe auch für ihn eine.
export function standwegFuer(
  wohin: Standwechsel,
): (kennung: string, fuer: string) => Promise<PartieAntwort> {
  const wege = {
    pausieren: api().PartiePausieren,
    fortsetzen: api().PartieFortsetzen,
    beenden: api().PartieBeenden,
  }

  return wege[wohin]
}

// neueBefehlskennung gibt eine Kennung für **eine** Handlung.
//
// **`crypto.randomUUID` und kein Zähler**: ein Zähler finge nach einem
// Neustart des Fensters wieder bei eins an, und der Kern hielte den ersten
// Befehl für eine Wiederholung des ersten von vorhin.
export function neueBefehlskennung(): string {
  return crypto.randomUUID()
}

// starthaendeZiehen zieht für jeden Sitz im Beginn die Starthand — P6
// (#2374),
// Artboard `PartieStart09`. **Ziehen ist keine Entscheidung**: CR 103.5
// (Fassung 2026-09-25) lässt jeden Spieler ziehen, behalten oder Mulligan
// bleibt seine Wahl. **Jeder Sitz zieht selbst**, nacheinander, gegen die
// Fassung, die der vorige hinterließ; der erste Fehler hält an und kommt
// zurück — angelegt ist die Partie trotzdem (§7, Regel 3). **Es sagt, ob es
// gezogen hat**: ohne Sitz im Beginn gibt es nichts nachzulesen.
export async function starthaendeZiehen(partie: Partie): Promise<boolean> {
  let fassung = partie.fassung
  let gezogen = false

  for (const sitz of partie.sitze) {
    // **Ein wartender Sitz zieht nicht** (P9b, #2377): sein Deck bringt der
    // Gast erst mit, und ziehen tut er dann selbst.
    if (sitz.eroeffnung !== 'ausstehend' || sitz.wartet) {
      continue
    }

    const antwort = await api().PartieBefehl(
      partie.id,
      neueBefehlskennung(),
      sitz.kennung,
      'spieler.starthand',
      {},
      fassung,
    )
    auspacken(antwort, () => undefined)
    fassung = antwort.fassung
    gezogen = true
  }

  return gezogen
}

// tischFuerEingeladeneOeffnen gibt einen Tisch frei, an dem jemand erwartet
// wird — P9c (#2377),
// Artboard `Netzpartie09`: **wer einlädt, will den Tisch offen haben**, und
// ein zweiter Klick auf „Tisch öffnen“ wäre einer, den niemand vergessen
// darf. Freigegeben werden nur die wartenden Sitze; ohne einen geschieht nichts.
export async function tischFuerEingeladeneOeffnen(
  partie: Partie,
): Promise<void> {
  const wartend = partie.sitze
    .filter((sitz) => sitz.wartet)
    .map((sitz) => sitz.kennung)

  if (wartend.length === 0) {
    return
  }

  const antwort = await api().PartieFreigeben(partie.id, partie.fuer, wartend)
  auspacken(antwort, () => undefined)
}
