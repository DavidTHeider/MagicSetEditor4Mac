// **Ein Batch über mehrere Karten** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Warum dieser Kasten seine Auswahl selbst trägt
//
// **Weil die Kartenliste keine kennt.** Sie führt genau eine aufgeschlagene
// Karte (`kartenauswahl.ts`); eine Mehrfachauswahl dort wäre ein zweiter
// Halter derselben Sache und eine eigene Bedienung mit eigenen Tasten. Hier
// steht die Auswahl da, wo sie gebraucht wird — und sie zeigt dieselben
// Karten, weil sie dieselbe Liste liest.
//
// # Zwei Schritte, und der erste ändert nichts
//
// **Das Issue sagt es wörtlich**: „Batch … schreibt während Generierung
// nichts und zeigt Status je Ziel." Erst kommt je Karte ein Vorschlag oder
// ein Grund; geschrieben wird im zweiten Schritt, und dann **als Gruppe**.
//
// # Die vier Regeln aus §7
//
//   - **Regel 1 und 2** löst der `key` am Bauteil: ein Setwechsel räumt die
//     Auswahl und die Lagen mit ab. Wer ihn entfernt, nimmt beide weg.
//   - **Regel 3**: nach einer geglückten Übernahme wird gemeldet und nicht
//     zurückgenommen.
//   - **Regel 4** hält [useVorgang].
import { useId, useState } from 'react'

import { KIBatchgruppe } from './KIBatchgruppe'
import { KIKostenfreigabe } from './KIKostenfreigabe'
import { KIModellwahl } from './KIModellwahl'
import { aliasAus, anbieterAus, kostetGeld } from './kiwahl'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { kartenname } from '../seiten/kartenname'
import { meldungAus } from '../meldung'
import { useKIModelle } from '../haken/useKIModelle'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type { KIBatchAntwort } from '../antwortenbatch'

interface KIBatchkastenEigenschaften {
  setID: string

  // geaendert meldet dem Editor, dass die Karten neu zu lesen sind.
  geaendert: () => void
}

export function KIBatchkasten({
  setID,
  geaendert,
}: KIBatchkastenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  const vorgang = useVorgang()

  // **Die Kennung der Überschrift kommt von React** und steht nicht fest:
  // ein fester Wert kollidiert, sobald zwei Kästen im selben Dokument stehen
  // — und doppelte Kennungen sind ein axe-Befund, bevor sie ein sichtbarer
  // Fehler sind.
  const ueberschrift = useId()

  const [offen, setzeOffen] = useState(false)
  const [gefragt, setzeGefragt] = useState(false)
  const katalog = useKIModelle(gefragt)

  const [gewaehltesModell, setzeModell] = useState('')
  const [auftrag, setzeAuftrag] = useState('')
  const [ziele, setzeZiele] = useState<string[]>([])
  const [batch, setzeBatch] = useState<KIBatchAntwort>()
  const [fehler, setzeFehler] = useState<string>()
  const [freigabe, setzeFreigabe] = useState(false)

  // derLauf ist der Schlüssel, unter dem [useVorgang] mitzählt.
  //
  // **Einmal benannt und nicht zweimal getippt**: der erste Anlauf startete
  // unter `batch/${setID}` und fragte nach `setID` — der laufende Zustand war
  // damit **nie** sichtbar, und der Knopf nahm einen zweiten Ruf entgegen
  // (§7, Regel 4). Gefunden hat es der Lauf, der auf »Läuft …« wartet.
  const derLauf = `batch/${setID}`

  // **Die Karten werden erst beim Aufklappen gelesen** — dieselbe Regel wie
  // beim Katalog nebenan: das Aufschlagen einer Karte soll keine Liste laden,
  // die niemand sehen will.
  const liste = useLaden(
    () => (offen ? api().Karten(setID) : Promise.resolve(undefined)),
    [setID, offen],
  )

  // **Die Namen kommen aus den Werten** — eine Karte hat kein Namensfeld im
  // DTO, sondern Feldwerte, und welcher der Name ist, weiß [kartenname]. Er
  // nimmt auch den Rückfall entgegen: eine Karte ohne Namen ist keine ohne
  // Zeile.
  const karten = liste.daten?.karten ?? []

  function umschalten(id: string) {
    setzeZiele((bisher) =>
      bisher.includes(id)
        ? bisher.filter((eine) => eine !== id)
        : [...bisher, id],
    )
  }

  function ausloesen() {
    if (kostetGeld(katalog.modelle, gewaehltesModell)) {
      setzeFreigabe(true)

      return
    }

    void fahren('')
  }

  async function fahren(nachweis: string) {
    setzeFehler(undefined)

    await vorgang.starte(derLauf, async () => {
      try {
        const antwort = await api().KIBatchVorschlagen(
          ziele,
          anbieterAus(gewaehltesModell),
          aliasAus(gewaehltesModell),
          auftrag,
          nachweis,
        )

        setzeBatch(auspacken(antwort, () => antwort))
      } catch (ursache) {
        setzeBatch(undefined)
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const brauchbar =
    ziele.length > 0 && gewaehltesModell !== '' && auftrag.trim() !== ''

  return (
    <section
      className="flex flex-col gap-2 border-t border-kante pt-3"
      aria-labelledby={ueberschrift}
    >
      <button
        type="button"
        id={ueberschrift}
        className="self-start text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        aria-expanded={offen}
        onClick={() => {
          setzeOffen((bisher) => !bisher)
          setzeGefragt(true)
        }}
      >
        {eigene.kiBatchTitel}
      </button>

      {offen && (
        <>
          <p className="max-w-prose text-sm text-leise">{eigene.kiBatchSatz}</p>

          <fieldset className="flex flex-col gap-1">
            <legend className="text-sm">
              {einsetzen(eigene.kiBatchZiele, {
                gewaehlt: String(ziele.length),
                gesamt: String(karten.length),
              })}
            </legend>

            {liste.fehler !== undefined && (
              <p role="alert" className="text-sm text-fehler">
                {liste.fehler}
              </p>
            )}

            {karten.length === 0 && liste.fehler === undefined && (
              <p className="text-sm text-leise">{eigene.kiBatchOhneKarten}</p>
            )}

            <ul className="flex max-h-48 flex-col gap-1 overflow-y-auto">
              {karten.map((karte) => (
                <li key={karte.id}>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={ziele.includes(karte.id)}
                      onChange={() => {
                        umschalten(karte.id)
                      }}
                    />
                    {kartenname(karte, eigene.kiBatchOhneName)}
                  </label>
                </li>
              ))}
            </ul>
          </fieldset>

          <KIModellwahl
            gefragt={gefragt}
            modelle={katalog.modelle}
            laedt={katalog.laedt}
            fehler={katalog.fehler}
            gewaehlt={gewaehltesModell}
            waehlen={setzeModell}
          />

          <label className="flex flex-col gap-1 text-sm">
            {eigene.kiBatchAuftrag}
            <textarea
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              rows={2}
              value={auftrag}
              onChange={(ereignis) => {
                setzeAuftrag(ereignis.target.value)
              }}
            />
          </label>

          {freigabe ? (
            <KIKostenfreigabe
              anbieter={anbieterAus(gewaehltesModell)}
              freigegeben={(nachweis) => {
                setzeFreigabe(false)
                void fahren(nachweis)
              }}
              verzichtet={() => {
                setzeFreigabe(false)
              }}
            />
          ) : (
            <button
              type="button"
              className="self-start rounded-sm bg-marke px-3 py-1 text-sm text-marketext focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
              disabled={!brauchbar || vorgang.laeuft(derLauf)}
              onClick={ausloesen}
            >
              {vorgang.laeuft(derLauf)
                ? eigene.kiBatchLaeuft
                : eigene.kiBatchFahren}
            </button>
          )}

          {fehler !== undefined && (
            <p role="alert" className="text-sm text-fehler">
              {fehler}
            </p>
          )}

          {batch !== undefined && (
            <KIBatchgruppe batch={batch} geaendert={geaendert} />
          )}
        </>
      )}
    </section>
  )
}
