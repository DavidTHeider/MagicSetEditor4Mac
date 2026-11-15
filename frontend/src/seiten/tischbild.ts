// **Welches Bild zu einem Objekt gehört, und wie eine Diagnose im Rahmen
// bleibt** — herausgelöst mit `M1`
// (#1603).
//
// **Eine eigene Datei, und der Schnitt folgt der Frage** (§2): `Partietisch.tsx`
// beantwortet, **was eine Partie gerade ist** — hier steht, **wie eine Karte
// zu ihrem Bild kommt**. Der Anlass war die Zeilengrenze: mit den zwei neuen
// Knöpfen der Zugleiste ging die eine Datei auf 405 Zeilen.

import type { Spielobjekt } from '../apipartie'
import type { Texturwunsch } from '../apitextur'

// bildwunsch sagt, welches Bild zu einem Objekt gehört.
//
// **Ein geheimes Objekt bekommt den Rücken seines Spiels** — und keine
// Kartenkennung: es trägt keine, und es darf keine bekommen. Der Rücken ist
// für jede Karte derselbe; wer ihn aus einer Karte baute, verriete sie über
// seine Maße (#1459).
//
// # Und `verdeckt` ist nicht die zweite Seite
//
// **Zwei Dinge, die der erste Anlauf verwechselt hat.** `verdeckt` heißt: es
// **liegt** umgedreht auf dem Tisch — `objekt.umdrehen` setzt genau das
// (`service.sandkasten.go`). Die **zweite öffentliche Seite** ist etwas
// anderes: eine Karte, die zwei Vorderseiten hat und von der eine gezeigt
// wird.
//
// **Der erste Anlauf band sie an `verdeckt`** und zeigte damit für jede
// umgedrehte Karte die zweite Seite — bei einer einseitigen Karte ist das
// ein Fehler, bei einer zweiseitigen die falsche Auskunft. Aufgefallen ist
// es an einem **unbetretenen Zweig** in der Abdeckung: die Bedingung hatte
// keinen Fall, weil es im Fenster keinen Weg dorthin gab (§3).
//
// **Wer die andere Seite sehen will, sagt es** — in der Betrachtung neben den
// Befehlen, und nur bei einer Karte, die eine hat.
export function bildwunsch(
  objekt: Spielobjekt,
  spiel: string,
  guete: string,
  seite = 'vorne',
): Texturwunsch {
  if (objekt.geheim) {
    return { art: 'ruecken', spiel, guete }
  }

  return {
    art: 'karte',
    karteID: objekt.karteID,
    bild: objekt.bild,
    objekt: objekt.kennung,
    seite,
    guete,
  }
}
