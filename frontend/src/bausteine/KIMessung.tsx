// **Messen, was ein Modell kann** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Warum es diesen Kasten gibt
//
// **Weil nicht jeder Anbieter seine Modelle beschreibt.** OpenAI nennt an
// `/v1/models` **keine** Fähigkeiten, und ohne Beleg gehört ein Modell nicht
// in den Katalog. Ohne einen dritten Weg wäre der Anbieter ein Schalter, der
// nichts tut.
//
// **Hier steht der dritte Weg**: die Modelle, die er nennt, ohne etwas über
// sie zu sagen — jedes mit einem Knopf, der einen kleinen Auftrag mit festem
// Schema hinausschickt.
//
// # Die vier Regeln aus §7
//
//   - **Regel 1 und 2** löst der `key` am Bauteil: ein Anbieterwechsel räumt
//     die offene Meldung mit ab.
//   - **Regel 3**: nach einer geglückten Messung wird der Katalog
//     nachgeladen — ein Fehler dabei nimmt die Messung nicht zurück.
//   - **Regel 4** hält [useVorgang]: ein zweiter Ruf, während der erste
//     läuft, ginge gegen einen Stand, den niemand kennt — und kostete ein
//     zweites Mal Geld.
import { useState } from 'react'

import { KIKostenfreigabe } from './KIKostenfreigabe'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

interface KIMessungEigenschaften {
  anbieter: string

  // kostet sagt, ob dieser Anbieter Geld nimmt (#1539).
  //
  // **Vom Aufrufer und nicht aus dem Namen**: wo ein Anbieter rechnet, sagt
  // er selbst (`KIAnbieterDTO.ort`). Eine Liste von Anbieternamen hier wäre
  // eine zweite Wahrheit, die beim nächsten Anbieter ausbliebe.
  kostet: boolean

  // messbar sind die Modelle, über die nichts feststeht.
  messbar: string[]

  // gemessen sind die, für die eine Messung liegt.
  gemessen: string[]

  // gemeldet sagt, dass der Katalog neu zu lesen ist.
  gemeldet: () => void
}

export function KIMessung({
  anbieter,
  kostet,
  messbar,
  gemessen,
  gemeldet,
}: KIMessungEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  const vorgang = useVorgang()

  const [fehler, setzeFehler] = useState<string>()
  const [geglueckt, setzeGeglueckt] = useState<string>()

  // gefragt hält das Modell, für das gerade eine Freigabe erbeten wird —
  // leer heißt: keine offene Frage (#1539).
  //
  // **Der Alias und kein Schalter**: es gibt mehrere Knöpfe, und eine
  // Freigabe gilt dem, für den jemand gedrückt hat.
  const [gefragt, setzeGefragt] = useState('')

  // ausloesen entscheidet, ob vorher zu fragen ist.
  //
  // **Eine Messung ist eine Anfrage und keine Ausnahme** (A9): sie schickt
  // einen kleinen Auftrag hinaus, und bei einem Cloudanbieter kostet der.
  function ausloesen(alias: string) {
    if (kostet) {
      setzeGefragt(alias)

      return
    }

    void messen(alias, '')
  }

  async function messen(alias: string, nachweis: string) {
    setzeFehler(undefined)
    setzeGeglueckt(undefined)

    await vorgang.starte(`messung/${anbieter}/${alias}`, async () => {
      try {
        const antwort = await api().KIModellPruefen(anbieter, alias, nachweis)
        const proben = auspacken(antwort, () => antwort.proben)

        setzeGeglueckt(
          einsetzen(eigene.kiMessungGeglueckt, {
            alias,
            faehigkeiten: proben.map((probe) => probe.faehigkeit).join(', '),
          }),
        )

        // **Erst melden, dann nachladen** (§7, Regel 3): gemessen ist
        // gemessen, und ein Lesefehler danach nimmt es nicht zurück.
        gemeldet()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  async function vergessen(alias: string) {
    setzeFehler(undefined)
    setzeGeglueckt(undefined)

    await vorgang.starte(`vergessen/${anbieter}/${alias}`, async () => {
      try {
        auspacken(await api().KIModellVergessen(anbieter, alias), () => null)

        gemeldet()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  if (messbar.length === 0 && gemessen.length === 0) {
    return null
  }

  return (
    <section className="flex flex-col gap-2">
      <h4 className="text-sm font-medium">{eigene.kiMessbarTitel}</h4>

      {messbar.length > 0 && (
        <p className="max-w-prose text-sm text-leise">
          {einsetzen(eigene.kiMessbarSatz, { anbieter })}
        </p>
      )}

      <ul className="flex flex-col gap-1 text-sm">
        {messbar.map((alias) => (
          <li key={alias} className="flex items-center gap-2">
            <span className="text-leise">{alias}</span>

            <button
              type="button"
              className="rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
              disabled={vorgang.laeuft(`messung/${anbieter}/${alias}`)}
              onClick={() => {
                ausloesen(alias)
              }}
            >
              {eigene.kiMessen}
            </button>

            {/* **Die Meldung steht bei dem Modell, das gemessen wird** —
                nicht am Kasten (§6): wer drei Knöpfe sieht, will wissen,
                welcher gerade arbeitet. */}
            {vorgang.laeuft(`messung/${anbieter}/${alias}`) && (
              <span role="status" className="text-leise">
                {eigene.kiMessungLaeuft}
              </span>
            )}

            {/* **Was Geld kostet, wird vorher freigegeben** (#1539) — und
                die Freigabe steht bei dem Modell, für das sie gilt. */}
            {gefragt === alias && (
              <KIKostenfreigabe
                anbieter={anbieter}
                freigegeben={(nachweis) => {
                  setzeGefragt('')
                  void messen(alias, nachweis)
                }}
                verzichtet={() => {
                  setzeGefragt('')
                }}
              />
            )}
          </li>
        ))}
      </ul>

      {gemessen.length > 0 && (
        <ul className="flex flex-col gap-1 text-sm">
          {gemessen.map((alias) => (
            <li key={alias} className="flex items-center gap-2">
              <span className="text-leise">{alias}</span>

              <button
                type="button"
                className="rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
                disabled={vorgang.laeuft(`vergessen/${anbieter}/${alias}`)}
                onClick={() => {
                  void vergessen(alias)
                }}
              >
                {eigene.kiMessungVergessen}
              </button>
            </li>
          ))}
        </ul>
      )}

      {geglueckt !== undefined && (
        <p role="status" className="text-sm text-leise">
          {geglueckt}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}
