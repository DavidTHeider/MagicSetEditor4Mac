// **Was ein Zug selbst festhält** — P13
// (#2381),
// herausgelöst aus `handelszug.ts`, als er an die Zeilengrenze kam (§2).
import { api } from './fassade'
import { auspacken } from './umschlag'
import { erfuellt } from './zusagepassung'

import type { Handelsvorgang as Vorgangsstand } from './apihandelsleitung'

// gedeckt hält fest, was eindeutig festzuhalten ist, und sagt, ob danach
// jede eigene Zeile gedeckt ist.
//
// **Eindeutig heißt**: derselbe Druck steht nur in einer eigenen Zeile, und
// im Bestand gibt es genau eine Zeile dieses Drucks, die kein Wunsch ist,
// reicht und **zur Absprache passt** — Zustand, Sprache und Ausführung
// (`zusagepassung.ts`, dieselbe Regel wie im Dienst, #2492; eine offene
// Angabe erfüllt nichts). Sonst ginge ein anderes Exemplar weg, als
// drüben ankommt, und die Belege passten nicht zusammen: die Gegenzeichnung
// bliebe still aus (gefunden mit `nat04handel`). Was schon zum Teil
// festgehalten ist, entscheidet ein Mensch.
export async function gedeckt(stand: Vorgangsstand): Promise<boolean> {
  const eigene = stand.positionen.filter(
    (position) => position.von === stand.ich,
  )

  if (eigene.length === 0) {
    return true
  }

  const drucke = new Set(eigene.map((position) => position.printing))

  if (drucke.size !== eigene.length) {
    return false
  }

  const gelesen = await api().Handelszusagen(stand.vorgang)
  const zusagen = auspacken(gelesen, () => gelesen.zusagen)

  for (const position of eigene) {
    const antwort = await api().SammlungBesitz(position.printing)
    const zeilen = (
      auspacken(antwort, () => antwort.besitz)?.zeilen ?? []
    ).filter((zeile) => !zeile.wunsch)
    const passende = zeilen.filter((zeile) => erfuellt(zeile, position))

    // **Gezählt wird, was zur Absprache passt** (Befund 4 des Zweitblicks
    // über Phase 8): eine Zusage auf einer anderen Zeile desselben Drucks
    // deckt nichts — sie ginge weg, und drüben käme die abgesprochene an.
    // Steht eine da, entscheidet ein Mensch.
    const halten = (auf: typeof zeilen) =>
      zusagen
        .filter((zusage) => auf.some((zeile) => zeile.id === zusage.exemplarId))
        .reduce((summe, zusage) => summe + zusage.menge, 0)
    const gehalten = halten(passende)

    if (halten(zeilen) > gehalten) {
      return false
    }

    if (gehalten >= position.menge) {
      continue
    }

    const [einzige] = passende

    if (
      gehalten > 0 ||
      passende.length !== 1 ||
      einzige === undefined ||
      einzige.menge < position.menge
    ) {
      return false
    }

    const festgehalten = await api().HandelFesthalten(
      stand.vorgang,
      einzige.id,
      position.menge,
      einzige.fassung,
    )
    auspacken(festgehalten, () => festgehalten)
  }

  return true
}
