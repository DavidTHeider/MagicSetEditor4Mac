// **Suchen, Ordnen und Filtern in der Partienliste** — PLAY-03.AK-05
// (#1711).
//
// # Im Fenster und nicht in Go
//
// **Weil die Liste schon ganz dasteht.** `App.Partien` gibt die Partien
// dieses Profils in einem Zug; eine zweite Abfrage je Tastendruck wäre
// Arbeit für nichts, und zwischen Ruf und Antwort läge jedes Mal die
// Wails-Naht. Dieselbe Überlegung wie bei der Einstellungssuche (U11,
// #1422).
//
// **Was daraus folgt und was nicht**: hier steht eine Sicht auf eine
// geladene Liste. Eine Suche über **alle** Partien aller Profile wäre eine
// andere Frage und gehörte in den Bestand.
//
// # Warum die Rechnung neben dem Bauteil steht
//
// **Weil sie ohne einen Baum zu prüfen ist** (§3). Was hier drinsteht — was
// ein Treffer ist, wonach geordnet wird —, ist die Fachregel; die Fläche
// daneben zeigt sie nur.
import type { Partiezeile } from '../apipartie'
import { falte, vergleicher } from '../sprache/vergleich'

// Ordnung sagt, wonach die Liste sortiert wird.
export type Ordnung = 'zuletzt' | 'name' | 'zug'

// Auswahl ist, was jemand an der Liste eingestellt hat.
export interface Auswahl {
  // frage ist die Sucheingabe — **leer heißt: alles**.
  frage: string

  // stand ist der gewählte Partiestand — **leer heißt: jeder**.
  stand: string

  ordnung: Ordnung
}

// lebenslauf ist die Reihenfolge, in der eine Partie ihre Stände durchläuft.
//
// **Sie steht hier und nicht in der Domäne**: sie ist eine Aussage über die
// Anzeige. Welche Stände es gibt, sagt `domain.Partiestaende`; welcher
// zuerst zur Wahl steht, ist eine Frage an die Fläche.
const lebenslauf = ['laeuft', 'pausiert', 'beendet']

// partienname nennt die Partie an ihren Sitzen.
//
// **An den Namen und nicht an der Kennung**: „David und Gegenüber" sagt
// einem Menschen, welche Partie das ist; `p1` sagt es niemandem. Wo kein
// Name steht, bleibt die Kennung — eine leere Zeile wäre keine Auskunft
// (§8).
export function partienname(partie: Partiezeile): string {
  const namen = partie.sitze
    .map((sitz) => sitz.name)
    .filter((name) => name !== '')

  return namen.length === 0 ? partie.id : namen.join(' · ')
}

// partienauswahl gibt die Partien, die zur Auswahl passen — geordnet.
export function partienauswahl(
  partien: readonly Partiezeile[],
  { frage, stand, ordnung }: Auswahl,
  sprache: string,
): Partiezeile[] {
  const gesucht = falte(frage.trim(), sprache)

  const uebrig = partien.filter(
    (partie) =>
      (stand === '' || partie.stand === stand) &&
      (gesucht === '' || heuhaufen(partie, sprache).includes(gesucht)),
  )

  if (ordnung === 'name') {
    const namen = vergleicher(sprache)

    return uebrig.sort((eine, andere) =>
      namen(partienname(eine), partienname(andere)),
    )
  }

  if (ordnung === 'zug') {
    // **Der weiteste Zug zuerst**, und bei Gleichstand die gegebene
    // Reihenfolge: `Array.prototype.sort` ist stabil, und zwei gleich weite
    // Partien sollen nicht bei jedem Zeichnen umeinanderspringen.
    return uebrig.sort((eine, andere) => andere.zug - eine.zug)
  }

  // **`zuletzt` ordnet nicht.** `Partiezeile` trägt kein Datum, und
  // `fassung` zählt Ereignisse und keine Zeit — eine Partie mit vielen Zügen
  // ist nicht die jüngere. Die Änderungszeit kennt der Bestand, und er gibt
  // die Liste schon in ihr ([app.Partien]).
  return uebrig
}

// heuhaufen ist, worin die Suche sucht.
//
// **Gesucht wird, was dasteht** — der Name der Partie und ihr Spiel. Eine
// Sitzkennung steht nur da, wo der Name fehlt, und `partienname` bringt sie
// dann von selbst mit; sonst wäre sie ein unsichtbarer Treffer, und niemand
// sähe, warum die Zeile bleibt.
//
// **Der Stand steht ausdrücklich nicht darin**: er hat seinen eigenen
// Filter, und er stünde hier in seiner Kennung — wer „läuft" tippte, fände
// `laeuft` nicht.
function heuhaufen(partie: Partiezeile, sprache: string): string {
  return falte(`${partienname(partie)} ${partie.spiel}`, sprache)
}

// staendeIn nennt die Stände, die in dieser Liste wirklich vorkommen.
//
// **Und nur die**: ein Filter auf einen Stand, den keine Partie hat, führte
// in eine leere Liste, und niemand sähe, warum.
export function staendeIn(partien: readonly Partiezeile[]): string[] {
  const gefunden = new Set(partien.map((partie) => partie.stand))

  const bekannt = lebenslauf.filter((stand) => gefunden.has(stand))

  // **Was der Lebenslauf nicht kennt, hängt hinten an** — fremde Daten sind
  // der Normalfall (§8), und ein vierter Stand soll nicht verschwinden.
  const fremd = [...gefunden].filter((stand) => !lebenslauf.includes(stand))

  return [...bekannt, ...fremd.sort()]
}
