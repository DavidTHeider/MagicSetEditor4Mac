// **Woher die Werte einer Karte stammen** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Warum es diese Anzeige gibt
//
// **Weil eine Herkunft, die niemand sieht, keine ist.** Das Issue verlangt
// sie „bei erster Übernahme verpflichtend" — und das heißt nicht, dass sie
// im Bestand liegt, sondern dass jemand sie lesen kann.
//
// # Und sie sagt, wenn der Wert nicht mehr derselbe ist
//
// **Weil manuelle Nachbearbeitung erlaubt ist** (das Issue sagt es
// ausdrücklich). Wer den Wert danach ändert, darf das — und dann soll hier
// nicht stehen, der Inhalt stamme von der KI.
//
// # Zugeklappt, bis jemand fragt
//
// **Dieselbe Gestalt wie der Vorschlagskasten daneben** (A8, #1447), und aus
// verwandtem Grund: die Herkunft ist eine **Frage**, die jemand stellt, und
// nicht etwas, das jede geöffnete Karte beantwortet. Der Ruf geht zwar nur in
// den Bestand und nicht nach draußen — aber ein Kartenformular, das bei jedem
// Blättern eine zweite Abfrage auslöst, bezahlt für eine Auskunft, die selten
// jemand will.
import { useState } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { api } from '../fassade'
import { auspacken } from '../umschlag'

interface KIHerkunftEigenschaften {
  karteID: string

  // stand ändert sich bei jeder Übernahme — dann wird neu gelesen.
  stand: number
}

// KIHerkunft zeigt, welches Feld aus welchem Vorschlag stammt.
export function KIHerkunft({ karteID, stand }: KIHerkunftEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  // offen sagt, ob der Kasten aufgeklappt ist — **und ob gefragt wird**.
  const [offen, setzeOffen] = useState(false)

  const gelesen = useLaden(async () => {
    if (!offen) {
      return undefined
    }

    const antwort = await api().KIFeldherkunft(karteID)

    return auspacken(antwort, () => antwort.herkunft)
  }, [karteID, stand, offen])

  const herkunft = gelesen.daten ?? []

  return (
    <details
      className="rounded-sm border border-kante px-3 py-2"
      onToggle={(ereignis) => {
        setzeOffen(ereignis.currentTarget.open)
      }}
    >
      <summary className="cursor-pointer text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring">
        {eigene.kiHerkunftTitel}
      </summary>

      {/* **Was zugeklappt ist, steht nicht da** — sonst fände jede Abfrage
          über die Rolle etwas, das niemand sieht. */}
      {offen && (
        <div className="mt-3 flex flex-col gap-1">
          {gelesen.fehler !== undefined && (
            <p role="alert" className="text-sm text-fehler">
              {gelesen.fehler}
            </p>
          )}

          {/* **Solange nichts gelesen ist, steht nichts da.** Ein „keine
              Herkunft" vor der Antwort wäre eine Auskunft, die noch niemand
              hat. */}
          {!gelesen.laedt &&
            gelesen.fehler === undefined &&
            (herkunft.length === 0 ? (
              <p className="text-sm text-leise">{eigene.kiHerkunftKeine}</p>
            ) : (
              <ul className="flex flex-col gap-1 text-sm text-leise">
                {herkunft.map((einer) => (
                  <li key={einer.feld} className="flex flex-col">
                    <span>
                      {einsetzen(eigene.kiHerkunftZeile, {
                        feld: einer.feld,
                        anbieter: einer.anbieter,
                        modell: einer.modell,
                        zeitpunkt: einer.zeitpunkt,
                      })}
                    </span>

                    {/* **Die Aufgabe und nicht der Rohprompt** — das Issue zieht
                  die Grenze, und sie steht schon im Bestand. */}
                    <span>
                      {einsetzen(eigene.kiHerkunftAufgabe, {
                        aufgabe: einer.aufgabe,
                      })}
                    </span>

                    {einer.veraendert && (
                      <span>{eigene.kiHerkunftVeraendert}</span>
                    )}
                  </li>
                ))}
              </ul>
            ))}
        </div>
      )}
    </details>
  )
}
