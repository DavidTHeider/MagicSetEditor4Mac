// **Der Verlauf eines Turniers** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Der Vorschlag steht neben dem Legen und nicht darin**: ein Turnierleiter
// soll hinsehen können, bevor er Tatsachen schafft (Vermerk 0062).
//
// # Die laufende Runde offen, die früheren eingeklappt
//
// **Seit UXR-12** (#1920).
// Während einer Runde braucht die Werkbank diese Runde und das Legen der
// nächsten; vier Runden zu acht Tischen untereinander waren eine Fläche, auf
// der man die laufende suchte. Die früheren stehen in der Reihenfolge, in der
// sie gespielt wurden, und klappen auf — **eine Berichtigung bleibt in ihrer
// Runde** (PROTOTYP-0.6.0.md, „Runden" statt „Runde n"). *Runde legen* führt.
import { useState } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { Aufklappkasten } from './Aufklappkasten'
import { Knopf } from './Knopf'
import { Turnierpaarung } from './Turnierpaarung'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import type { PaarungDTO, RundeDTO, TurnierDTO } from '../apiturniere'
import type { Turnierstand } from '../haken/useTurniere'

export function Turnierrunden({
  stand,
  turnier,
  nurLesen = false,
}: {
  stand: Turnierstand
  turnier: TurnierDTO

  // nurLesen heißt: das Turnier ist abgeschlossen — keine Runde, kein
  // Eintragen (DOM-04.AK-04, #1699).
  nurLesen?: boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  const [vorschlag, setzeVorschlag] = useState<PaarungDTO[] | null>(null)

  const namen = new Map(
    turnier.teilnehmer.map((teilnehmer) => [teilnehmer.id, teilnehmer.name]),
  )

  // **Ein Lesen und kein Schreiben.** `fuehre` lädt nach getaner Arbeit nach
  // und baut dabei dieses Bauteil ab — der Vorschlag verschwände mit ihm
  // (Befund 1 des Nachlaufs über Phase 4.7).
  const ansehen = async () => {
    await stand.lies(`turnier:${turnier.id}`, async () => {
      const antwort = await api().TurnierPaarungsvorschlag(turnier.id)

      setzeVorschlag(auspacken(antwort, () => antwort.paarungen))
    })
  }

  const legen = async () => {
    await stand.fuehre(`turnier:${turnier.id}`, async () => {
      auspacken(await api().TurnierRundeLegen(turnier.id), () => undefined)

      // **Der Vorschlag verschwindet, sobald er Tatsache ist.** Sonst stünde
      // er neben der Runde, die aus ihm entstand, und sähe wie eine zweite
      // aus.
      setzeVorschlag(null)
    })
  }

  const laeuft = stand.vorgang.laeuft(`turnier:${turnier.id}`)

  const titel = (runde: RundeDTO) =>
    eigene.runde.replace('{nummer}', String(runde.nummer))

  // **Eine geordnete Liste**, weil die Reihenfolge der Tische zählt — und
  // angesagt werden soll.
  const paarungen = (runde: RundeDTO) => (
    <ol aria-label={titel(runde)} className="flex flex-col gap-1">
      {runde.paarungen.map((paarung) => (
        <li key={paarung.id}>
          <Turnierpaarung
            stand={stand}
            turnierID={turnier.id}
            fassung={turnier.fassung}
            paarung={paarung}
            namen={namen}
            // **Über wie viele Spiele ein Match hier geht** —
            // `service.hoechstensSpiele`: bei einem Einzelspielturnier eines,
            // sonst **null**, und null heißt keine Grenze. Eine hier
            // erfundene Zahl wiese ein gültiges Ergebnis ab (§11).
            hoechstens={turnier.einzelspiel ? 1 : 0}
            // **Nach ihr schon gepaart** — dann ist jedes Ergebnis hier eine
            // Berichtigung (DOM-04.AK-02).
            spaeterGelegt={runde.nummer < turnier.runden.length}
            nurLesen={nurLesen}
          />
        </li>
      ))}
    </ol>
  )

  const laufende = turnier.runden.at(-1)

  return (
    <div className="flex flex-col gap-3">
      <h3 className="font-semibold">{eigene.runden}</h3>

      {laufende === undefined ? (
        <p>{eigene.rundenLeer}</p>
      ) : (
        <>
          {turnier.runden.slice(0, -1).map((runde) => (
            <Aufklappkasten key={runde.id} titel={titel(runde)}>
              {paarungen(runde)}
            </Aufklappkasten>
          ))}

          <div className="flex flex-col gap-1">
            <h4 className="text-sm font-semibold">{titel(laufende)}</h4>

            {paarungen(laufende)}
          </div>
        </>
      )}

      {/* **Die Empfehlung aus Appendix E** — sichtbar und überschreibbar
          (Befund 2 des Zweitblicks über Phase 4.7). Sie steht **neben** den
          Knöpfen und nicht in ihnen: wer weiterspielen will, kann. */}
      {turnier.empfohleneRunden > 0 && (
        <p className="text-sm text-leise">
          {eigene.empfehlung
            .replace('{gelegt}', String(turnier.runden.length))
            .replaceAll('{empfohlen}', String(turnier.empfohleneRunden))
            .replace(
              '{spieler}',
              String(
                turnier.teilnehmer.filter((einer) => !einer.ausgestiegen)
                  .length,
              ),
            )}
        </p>
      )}

      {vorschlag !== null && (
        <div className="flex flex-col gap-1 rounded-md border border-kante p-3">
          <h4 className="text-sm font-semibold">{eigene.vorschlagTitel}</h4>

          <ol
            aria-label={eigene.vorschlagTitel}
            className="flex flex-col gap-1"
          >
            {vorschlag.map((paarung, stelle) => (
              <li key={`${paarung.a}:${String(stelle)}`}>
                {paarung.b === ''
                  ? `${namen.get(paarung.a) ?? paarung.a} — ${eigene.bye}`
                  : `${namen.get(paarung.a) ?? paarung.a} ${eigene.gegen} ${
                      namen.get(paarung.b) ?? paarung.b
                    }`}
              </li>
            ))}
          </ol>

          <Erklaerungsabsatz
            text={eigene.vorschlagHinweis}
            klasse="text-sm text-leise"
          />
        </div>
      )}

      {!nurLesen && (
        <div className="flex flex-wrap gap-3">
          <Knopf
            angehalten={laeuft}
            aufKlick={() => {
              void ansehen()
            }}
          >
            {eigene.vorschlag}
          </Knopf>

          <Knopf
            gewicht="haupt"
            angehalten={laeuft}
            aufKlick={() => {
              void legen()
            }}
          >
            {eigene.rundeLegen}
          </Knopf>
        </div>
      )}
    </div>
  )
}
