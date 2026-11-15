// **Die Symbolschriften, nur zur Ansicht** — DOM-02.AK-01
// (#1697):
// „klare Unterscheidung zwischen Setsymbol und Symbolschrift".
//
// Eine Symbolschrift ist ein **Paket** — sie gehört dem Rechner, wird unter
// *Pakete* eingebunden und von einem Stil in Textfelder gesetzt. Hier steht
// sie neben der Bibliothek, damit niemand sie dort sucht, und ohne Knopf,
// damit niemand glaubt, sie ließe sich hier einsetzen.
//
// **Die Liste steht eingeklappt** — UXR-07
// (#1915).
// Das Full-Magic-Pack bringt 76 Symbolschriften mit; offen machten sie die
// Lasche *Symbol* anderthalb Fenster höher (R4 aus `docs/UIREVIEW.md`).
// Überschrift und Erklärung bleiben stehen, damit man sie hier findet; der
// einzige Knopf öffnet die Liste und setzt nichts ein.
import { useId } from 'react'

import { api } from '../api'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'

// symbolschriftart ist die Paketart einer Symbolschrift.
const symbolschriftart = 'mse-symbol-font'

export function Symbolschriften() {
  const eigene = useTexte().seiten.karten.symbolbibliothek
  const ueberschrift = useId()

  const schriften = useLaden(async () => {
    const antwort = await api().Pakete()

    return auspacken(antwort, () =>
      antwort.pakete.filter((paket) => paket.art === symbolschriftart),
    )
  }, [])

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-2">
      <h4 id={ueberschrift} className="text-sm font-medium">
        {eigene.schriften}
      </h4>
      <Erklaerungsabsatz text={eigene.schriftenErklaerung} />
      {schriften.daten !== undefined &&
        (schriften.daten.length === 0 ? (
          <p className="max-w-prose text-sm text-leise">
            {eigene.keineSchriften}
          </p>
        ) : (
          <Aufklappkasten
            titel={einsetzen(eigene.schriftenZeigen, {
              zahl: String(schriften.daten.length),
            })}
          >
            <ul aria-label={eigene.schriftenListe} className="text-sm">
              {schriften.daten.map((schrift) => (
                <li key={schrift.name}>
                  {schrift.version === ''
                    ? schrift.anzeigename
                    : einsetzen(eigene.schriftMitFassung, {
                        name: schrift.anzeigename,
                        fassung: schrift.version,
                      })}
                </li>
              ))}
            </ul>
          </Aufklappkasten>
        ))}
    </section>
  )
}
