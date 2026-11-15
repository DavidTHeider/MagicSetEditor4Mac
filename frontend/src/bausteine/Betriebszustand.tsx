// **Der Betriebszustand im Fenster** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Die eine Regel, die hier zu sehen sein muss
//
// **Ein fremder Dienst, der nicht antwortet, macht diesen Prozess nicht
// unbereit** (§16, Satz 5). Deshalb stehen die zwei Dinge getrennt: oben, ob
// **er** arbeiten kann, darunter, was über **fremde** bekannt ist. Wer beides
// in eine Zeile schriebe, machte aus einem fremden Ausfall einen eigenen.
//
// # Warum er auf der Technikseite steht
//
// Weil er dieselbe Frage beantwortet wie die Diagnose daneben: *warum geht
// gerade etwas nicht?* — und weil er, anders als die Diagnose, **immer**
// dasteht und nichts kostet.
import { useLaden } from '../haken/useLaden'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

export function Betriebszustand() {
  const eigene = useTexte().seiten.technik

  const stand = useLaden(async () => {
    const antwort = await api().Gesundheit()

    return auspacken(antwort, () => antwort)
  }, [])

  const zustand = stand.daten

  return (
    <section className="flex flex-col gap-2 border-t border-kante pt-4">
      <h2 className="text-xl font-semibold">{eigene.gesundheit}</h2>

      <Erklaerungsabsatz
        text={eigene.gesundheitErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {zustand !== undefined && (
        <>
          {/* **Wer antwortet, lebt.** Ein Zweig für „lebt nicht" wäre einer,
              den kein Lauf betreten kann: die Antwort ist der Beweis. Das
              Feld reist trotzdem mit — die Kommandozeile und der Dienst
              bilden daraus ihren Rückgabewert und ihren Statuscode. */}
          <p>{eigene.lebt}</p>

          <p className={zustand.bereit ? '' : 'text-hinweis'}>
            {zustand.bereit ? eigene.bereit : eigene.nichtBereit}
          </p>

          {/* **Der Grund steht nur da, wenn es einen gibt** — „bereit, weil
              nichts dagegen spricht" wäre eine Zeile ohne Aussage. */}
          {zustand.grund !== '' && (
            <p className="text-sm">{satzZu(eigene, zustand.grund)}</p>
          )}

          {zustand.fremde.length > 0 && (
            <>
              <h3 className="mt-2 font-medium">{eigene.fremdeDienste}</h3>

              <ul
                aria-label={eigene.fremdeDienste}
                className="flex flex-col gap-1"
              >
                {zustand.fremde.map((fremd) => (
                  <li
                    key={fremd.dienst}
                    className="flex flex-wrap gap-2 text-sm"
                  >
                    <span className="w-32">{fremd.dienst}</span>

                    <span className="text-leise">
                      {fremd.erreichbar
                        ? eigene.erreichbar
                        : eigene.nichtErreichbar}
                    </span>

                    {fremd.grund !== '' && (
                      <span className="text-leise">{fremd.grund}</span>
                    )}
                  </li>
                ))}
              </ul>
            </>
          )}
        </>
      )}
    </section>
  )
}

// satzZu sucht den Satz zu einem Grund — **aus dem Katalog und nicht aus Go**.
//
// **Der Rückfall ist der Grund selbst.** Ein Grund, den dieses Fenster noch
// nicht kennt, ist immer noch eine Auskunft; eine leere Zeile wäre keine.
function satzZu(
  eigene: ReturnType<typeof useTexte>['seiten']['technik'],
  grund: string,
): string {
  const saetze: Record<string, string> = {
    bestandZu: eigene.grundBestandZu,
    wanderungLaeuft: eigene.grundWanderungLaeuft,
    bestandUnlesbar: eigene.grundBestandUnlesbar,
  }

  return saetze[grund] ?? grund
}
