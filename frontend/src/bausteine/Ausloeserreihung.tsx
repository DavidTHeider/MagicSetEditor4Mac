// **Die offene Wahl aus CR 603.3b** — Werkstück M9 der Regelmaschine
// (#1611).
//
// # Warum sie oben am Tisch steht und nicht bei den Befehlen
//
// **Weil sie das Spiel anhält** — derselbe Grund wie bei der Legendenwahl
// nebenan, und hier sogar härter: CR 603.3 legt die Fähigkeiten auf den
// Stapel, **bevor** das nächste Mal jemand Priorität bekäme. Solange geordnet
// wird, hat **niemand** eine, und ein Tisch, der die Frage unten in einer
// Leiste zeigte, sähe für alle wie ein Ausfall aus.
//
// # Und warum jede wartende Fähigkeit ihren eigenen Knopf hat
//
// **Weil die Wahl eine Fähigkeit meint und keine Art.** Zwei Kreaturen, die
// dem Monarchen Kampfschaden zufügen, lösen dieselbe Regel zweimal aus
// (CR 725.2) — wer die Art wählte, hätte nichts entschieden. Der Befehl trägt
// deshalb die **Kennung**.
import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { handlungswort, sitzname } from '../seiten/partiewoerter'
import { useTexte } from '../sprache/kontext'

import type { Partie } from '../apipartie'
import { Abschnittstitel } from './Abschnittstitel'

export function Ausloeserreihung({
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

  if (partie.ausloeserreihung === '') {
    return undefined
  }

  return (
    <div
      className="flex flex-col gap-2"
      role="status"
      aria-label={eigene.reihungTitel}
    >
      <Abschnittstitel className="text-base font-semibold">
        {eigene.reihungTitel}
      </Abschnittstitel>

      <p className="text-sm">
        {einsetzen(eigene.reihungSatz, {
          wer: sitzname(partie, partie.ausloeserreihung),
        })}
      </p>

      <ul className="flex flex-wrap gap-2 text-sm">
        {partie.ausgeloest
          .filter((einer) => einer.beherrscher === partie.ausloeserreihung)
          .map((einer) => (
            <li key={einer.kennung}>
              <Knopf
                gewicht="leise"
                angehalten={angehalten}
                aufKlick={() => {
                  befehlen('spieler.reihung', { kennung: einer.kennung })
                }}
              >
                {`${eigene.reihungLegen} ${handlungswort(
                  texte.seiten.handlungen.ausloesername,
                  einer.art,
                )}`}
              </Knopf>
            </li>
          ))}
      </ul>
    </div>
  )
}
