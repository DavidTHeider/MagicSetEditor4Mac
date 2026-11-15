// **Ein neuer Tausch mit einem Klick** — P12
// (#2380),
// Befund C-18 und H1 ≤ 5: die Gegenüber, mit denen gehandelt werden darf,
// stehen als eigene Knöpfe unter „Neuer Tausch …" — die Schublade bleibt für
// alles Übrige (Aufnehmen, Bestätigen, die ganze Liste).
//
// **Nur wer bestätigt ist und den Zweck `handel` hat**: wer es nicht ist,
// käme am Tisch ohnehin nicht weiter, und ein Knopf, der dann abweist, wäre
// ein Versprechen ohne Einlösung.
//
// **Bis zu drei, und sonst keiner** — nicht die ersten drei. Mit zehn
// Gegenübern stünden links fünf Knöpfe mehr, als das Maß aus UXR-15 zulässt
// (`handelnmass.test.tsx`, gefunden von der Vorprüfung), und welche es
// wären, hinge an der Reihenfolge der Kontakte. Wer viele hat, wählt in der
// Schublade aus der ganzen Liste.
import { api } from '../fassade'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Knopf } from './Knopf'

const hoechstens = 3

export function Tauschpartnerleiste({
  waehlen,
}: {
  waehlen: (kennung: string, name: string) => void
}) {
  const tisch = useTexte().seiten.handeln.markttisch

  const partner = useLaden(async () => {
    const antwort = await api().Peerkontakte()

    const moegliche = auspacken(antwort, () => antwort.kontakte).filter(
      (kontakt) =>
        kontakt.stand === 'bestaetigt' && kontakt.rechte.includes('handel'),
    )

    return moegliche.length > hoechstens ? [] : moegliche
  }, [])

  if (partner.daten === undefined || partner.daten.length === 0) {
    return null
  }

  return (
    <div
      role="group"
      aria-label={tisch.neuTitel}
      className="flex flex-wrap gap-1"
    >
      {partner.daten.map((kontakt) => {
        const name = kontakt.eigenerName || kontakt.selbstgewaehlterName

        return (
          <Knopf
            key={kontakt.kennung}
            gewicht="leise"
            bezeichnung={einsetzen(tisch.neuMit, { name })}
            aufKlick={() => {
              waehlen(kontakt.kennung, name)
            }}
          >
            {name}
          </Knopf>
        )
      })}
    </div>
  )
}
