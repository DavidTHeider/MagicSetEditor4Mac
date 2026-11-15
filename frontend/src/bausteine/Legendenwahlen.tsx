// **Die offene Wahl aus der Legendenregel** — Werkstück M12 der
// Regelmaschine
// (#1614).
//
// # Warum sie oben am Tisch steht und nicht bei den Befehlen
//
// **Weil sie das Spiel anhält.** CR 704.5j lässt die Partie stehen, bis
// jemand gewählt hat; ein Tisch, der die Frage unten in einer Leiste zeigte,
// sähe für den Wartenden aus wie ein Ausfall.
//
// # Und warum jeder Kandidat seinen eigenen Knopf hat
//
// **Weil die Wahl ein Objekt meint und keinen Namen.** Zwei Legenden heißen
// gleich — das ist der Anlass der Regel; wer den Namen wählte, hätte nichts
// entschieden. Der Befehl trägt deshalb die **Kennung**.
import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { kartenwort } from '../spiel/feldwerte'
import { sitzname } from '../seiten/partiewoerter'
import { useTexte } from '../sprache/kontext'

import type { Partie, Spielobjekt } from '../apipartie'
import { Abschnittstitel } from './Abschnittstitel'

export function Legendenwahlen({
  partie,
  befehlen,
  angehalten,
}: {
  partie: Partie
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void
  angehalten: boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  if (partie.legendenwahlen.length === 0) {
    return undefined
  }

  return (
    <div className="flex flex-col gap-2" role="status">
      <Abschnittstitel className="text-base font-semibold">
        {eigene.legendeTitel}
      </Abschnittstitel>

      <ul className="flex flex-col gap-2 text-sm">
        {partie.legendenwahlen.map((wahl) => (
          <li
            key={`${wahl.sitz}:${wahl.name}`}
            className="flex flex-wrap items-baseline gap-2"
          >
            <span>
              {einsetzen(eigene.legendeSatz, {
                wer: sitzname(partie, wahl.sitz),
                name: wahl.name,
              })}
            </span>

            {kandidaten(partie, wahl.sitz, wahl.name).map((objekt) => (
              <Knopf
                key={objekt.kennung}
                gewicht="leise"
                angehalten={angehalten}
                aufKlick={() => {
                  befehlen('spieler.legende', { objekt: objekt.kennung })
                }}
              >
                {eigene.legendeBehalten}
              </Knopf>
            ))}
          </li>
        ))}
      </ul>
    </div>
  )
}

// kandidaten sind die Objekte, zwischen denen gewählt wird.
//
// **Gerechnet und nicht mitgeschickt** — dieselbe Überlegung wie im Kern:
// welche Objekte gemeint sind, ergibt sich aus dem Tisch, und der ändert
// sich, während die Frage offen ist.
//
// **Der Name kommt aus [kartenwort]** und nicht aus einer zweiten Fassung
// daneben — zwei Antworten auf dieselbe Frage laufen auseinander (§2).
function kandidaten(partie: Partie, sitz: string, name: string): Spielobjekt[] {
  return partie.objekte.filter(
    (objekt) =>
      objekt.zone === 'tisch' &&
      objekt.beherrscher === sitz &&
      kartenwort(objekt) === name,
  )
}
