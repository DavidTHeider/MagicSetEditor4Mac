// **Ein Set als MTG JSON schreiben** —
// #677.
//
// Aus `Sets.tsx` herausgelöst, als die Ausgänge einen gemeinsamen Ort bekamen:
// dieselbe Handlung stand an der Setzeile und wird jetzt auf der Lasche
// **Ausgeben** gebraucht. Der Schnitt folgt der Frage — dort steht, *welche*
// Sets es gibt, hier, was aus **einem** wird (§2).
import { useState } from 'react'

import { api, auspacken, type AusfuhrAntwort, type SetDTO } from '../api'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Ausgabefuss } from '../bausteine/Ausgabefuss'
import { type Vorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { Dateiverweis } from '../bausteine/Dateiverweis'

interface Eigenschaften {
  satz: SetDTO
  vorgang: Vorgang
}

export function MTGJSONSichern({ satz, vorgang }: Eigenschaften) {
  const texte = useTexte()
  const [ausfuhr, setzeAusfuhr] = useState<AusfuhrAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  // **Ein Vorgang je Datensatz** (§7, Regel 4): der Datensatz ist das Set. Zwei
  // Ausfuhren desselben Sets gleichzeitig gingen gegen einen Stand, den das
  // Backend noch nicht kennt.
  const sichern = async () => {
    await vorgang.starte(`mtgjson-${satz.id}`, async () => {
      setzeFehler(undefined)
      setzeAusfuhr(undefined)

      try {
        const antwort = await api().MTGJSONSchreiben(satz.id, satz.spiel)

        auspacken(antwort, () => undefined)

        // **Ein abgebrochener Dialog ist keine Meldung wert** — wer den
        // Dateidialog schließt, hat nichts falsch gemacht.
        if (antwort.abgebrochen) {
          return
        }

        setzeAusfuhr(antwort)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  return (
    <>
      {/* **Er führt** (UXR-09, R1): auf der Fläche des Ziels ist er die eine
          Handlung — in derselben Fußleiste wie bei den anderen Zielen
          (AP-10, B-20). */}
      <Ausgabefuss>
        <Aktionsknopf
          aktion="sichern"
          gewicht="haupt"
          bezeichnung={`${texte.seiten.sets.mtgjsonSichern} — ${satz.name}`}
          angehalten={vorgang.laeuft(`mtgjson-${satz.id}`)}
          aufKlick={() => {
            void sichern()
          }}
        >
          {texte.seiten.sets.mtgjsonSichern}
        </Aktionsknopf>
      </Ausgabefuss>

      {fehler !== undefined && (
        <p role="alert" className="mt-3 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {ausfuhr && (
        <div role="status" className="mt-3">
          <p className="text-leise">
            {`${zaehlwort(ausfuhr.karten, texte.allgemein.zaehlwoerter.karten)} ${texte.seiten.sets.mtgjsonGesichert} ${ausfuhr.pfad}`}
          </p>

          <Dateiverweis pfad={ausfuhr.pfad} />

          {/* **Die Verluste stehen beim Erfolg** (§8): die Datei ist da, und
              was nicht abbildbar war, steht daneben. */}
          {ausfuhr.befunde.length > 0 && (
            <>
              <p className="mt-2 max-w-prose text-leise">
                {texte.seiten.sets.mtgjsonVerluste}
              </p>
              <ul className="mt-1 space-y-1 font-mono text-sm">
                {ausfuhr.befunde.map((verlust) => (
                  <li key={verlust}>{verlust}</li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}
    </>
  )
}
