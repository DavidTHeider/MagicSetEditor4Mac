// **Das Kontingent an einem Anbieterkonto** — Werkstück A5 von Phase 5.3
// (#1444).
//
// # Warum es einen Knopf braucht
//
// **Weil die Frage ins Netz geht.** Wer die Einstellungen aufschlägt, löste
// sonst einen Abruf aus — und der kostet einen Kredit von genau dem
// Kontingent, nach dem er fragt.
//
// # Und warum die Frist dabeisteht
//
// **Nachgelesen am 2026-09-11**: „Usage info is updated every 20-30 minutes."
// Eine Zahl ohne diesen Satz sähe aus wie eine Sperre. Die Grenze, die hier
// gilt, ist das lokale Budget
// ([0096](../../docs/entscheidungen/0096-welche-optionalen-datenquellen-zugaenglich-sind.md)).
import { useState } from 'react'

import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { Anbieterstandantwort } from '../antworten'

// derDienst ist der einzige, der heute ein Kontingent nennt.
const derDienst = 'scrydex'

export function Anbieterstand() {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  const [stand, setzeStand] = useState<Anbieterstandantwort>()
  const [meldung, setzeMeldung] = useState<string>()

  const vorgang = useVorgang()
  const laeuft = vorgang.laeuft(derDienst)

  const holen = () => {
    void vorgang.starte(derDienst, async () => {
      setzeMeldung(undefined)

      try {
        const geholt = await api().Anbieterstand(derDienst)

        setzeStand(geholt)
        auspacken(geholt, () => undefined)
      } catch (ursache: unknown) {
        setzeMeldung(meldungAus(ursache, texte))
      }
    })
  }

  return (
    <section className="flex flex-col gap-2">
      <h3 className="text-sm font-medium">{eigene.kontingentTitel}</h3>

      <Erklaerungsabsatz
        text={eigene.kontingentErklaerung}
        klasse="text-sm text-leise"
      />

      <Knopf
        bezeichnung={einsetzen(eigene.kontingentHolenBei, {
          dienst: derDienst,
        })}
        aufKlick={holen}
        angehalten={laeuft}
      >
        {laeuft ? eigene.kontingentLaeuft : eigene.kontingentHolen}
      </Knopf>

      {meldung !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {meldung}
        </p>
      )}

      {stand?.fehler === null && (
        <div className="flex flex-col gap-1 text-sm">
          <p>
            {einsetzen(eigene.kontingentZahlen, {
              rest: String(stand.rest),
              gesamt: String(stand.gesamt),
            })}
          </p>

          {/* **Die Frist steht immer dabei** — ohne sie sähe der Rest aus wie
              eine Sperre. */}
          <p className="text-leise">
            {einsetzen(eigene.kontingentFrist, {
              minuten: String(stand.hoechstensAltMinuten),
            })}
          </p>

          <p className="text-leise">
            {einsetzen(eigene.kontingentUeberzug, {
              satz: stand.ueberzugssatz,
            })}
          </p>
        </div>
      )}
    </section>
  )
}
