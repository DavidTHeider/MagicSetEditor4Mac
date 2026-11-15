// **Führungen, und die Bereiche, die keine Seite sind** — Werkstück U11 von
// Phase 5.1 (#1422).
//
// # Zwei Gebiete der Hilfe, eine Gestalt
//
// Beide sind Beschreibungslisten mit einer Handlung am Eintrag, wo es eine
// gibt: die ersten Schritte lassen sich **jetzt** zeigen, und der KI-Bereich
// führt zu den Schnittstellen, wo steht, was hinausgeht. Der KI-Bereich hat
// als einziger keine Seite in der Leiste: er ist kein Ort, sondern ein Weg
// von außen herein. **Spielen und Handeln stehen hier nicht**: beide sind
// Seiten der Leiste
// (#1474).
// Wer sie hier nennte, nennte sie zweimal — und der Satz daneben wäre falsch.
//
// # Eine Führung, drei Erklärungen
//
// Ehrlich gezählt gibt es **eine** Führung: die ersten Schritte. Was daneben
// steht — Arbeitsplatz, Befehlspalette, Kontextleiste —, erklärt den Aufbau
// in je einem Satz. Es steht hier, weil es sonst nirgends steht.
import { Flaeche } from '../bausteine/Flaeche'
import { Knopf } from '../bausteine/Knopf'
import { Satz } from '../bausteine/Satz'
import { useTexte } from '../sprache/kontext'

interface Punkt {
  name: string
  satz: string
  handlung?: { text: string; tue: () => void }
}

function Punkte({ punkte }: { punkte: Punkt[] }) {
  return (
    <dl className="grid gap-3 sm:grid-cols-[10rem_minmax(0,1fr)]">
      {punkte.map((punkt) => (
        <div key={punkt.name} className="contents">
          <dt className="font-semibold">{punkt.name}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={punkt.satz} />
            {punkt.handlung !== undefined && (
              <div className="mt-2">
                <Knopf gewicht="neben" aufKlick={punkt.handlung.tue}>
                  {punkt.handlung.text}
                </Knopf>
              </div>
            )}
          </dd>
        </div>
      ))}
    </dl>
  )
}

export function Hilfebereiche({
  zuSchnittstellen,
}: {
  zuSchnittstellen: () => void
}) {
  const { bereiche } = useTexte().hilfe

  return (
    <div className="mt-6">
      <h3 className="mb-1 font-semibold">{bereiche.ueberschrift}</h3>
      <p className="mb-3 max-w-prose text-leise">
        <Satz text={bereiche.einleitung} />
      </p>
      <Punkte
        punkte={[
          {
            name: bereiche.kiName,
            satz: bereiche.kiSatz,
            handlung: { text: bereiche.kiKnopf, tue: zuSchnittstellen },
          },
        ]}
      />
    </div>
  )
}

export function Hilfefuehrungen({
  einfuehrungZeigen,
}: {
  einfuehrungZeigen: () => void
}) {
  const { fuehrungen } = useTexte().hilfe

  return (
    <Flaeche
      ueberschrift={fuehrungen.ueberschrift}
      sprungmarke="hilfe-fuehrungen"
    >
      <p className="mb-4 max-w-prose text-leise">
        <Satz text={fuehrungen.einleitung} />
      </p>
      <Punkte
        punkte={[
          {
            name: fuehrungen.startName,
            satz: fuehrungen.startSatz,
            handlung: { text: fuehrungen.startZeigen, tue: einfuehrungZeigen },
          },
          {
            name: fuehrungen.arbeitsplatzName,
            satz: fuehrungen.arbeitsplatzSatz,
          },
          { name: fuehrungen.paletteName, satz: fuehrungen.paletteSatz },
          { name: fuehrungen.leisteName, satz: fuehrungen.leisteSatz },
        ]}
      />
    </Flaeche>
  )
}
