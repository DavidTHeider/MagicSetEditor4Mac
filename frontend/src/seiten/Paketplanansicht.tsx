// **Der Plan vor einer Paketänderung** — DOM-01.AK-02 und AK-03
// (#1696).
//
// **Was sie tut, steht in `usePaketplan`; wie sie aussieht, hier** (§2).
//
// # Warum eine Entscheidung und kein eigener Kasten
//
// Der Plan **ist** die Rückfrage: er steht an der Klickstelle, der Fokus
// liegt auf dem Abbrechen, und Escape nimmt ihn zurück — die Zusage von
// `Entscheidung` (UX-10). Vor dem Löschen trägt er deshalb die Aktion
// `loeschen`; die anderen drei Vorgänge nehmen nichts weg — Ersetzen und
// Zurück heben die bisherige Fassung als vorige auf.
//
// # Was nicht geht, wird nicht angeboten
//
// Ein Plan, der `moeglich: false` sagt, bekommt keinen Knopf zum Bestätigen —
// nur den Satz, warum, und den Weg hinaus. Ein Knopf, der mit einer Absage
// endet, wäre eine Frage, deren Antwort schon feststeht.
import { useState } from 'react'

import type { PaketplanDTO } from '../api'
import { Entscheidung } from '../bausteine/Entscheidung'
import { Knopf } from '../bausteine/Knopf'
import type { Paketplanung } from '../haken/usePaketplan'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { abhaengigkeitssatz } from './abhaengigkeitssatz'

interface Eigenschaften {
  planung: Paketplanung
}

export function Paketplanansicht({ planung }: Eigenschaften) {
  const texte = useTexte()
  const t = texte.seiten.pakete.plan
  const { plan, fehler } = planung

  // **Der Haken gilt dem Plan, zu dem er gesetzt wurde** (§7, Regel 1): ein
  // neuer Plan ist ein anderes Objekt, und der Haken ist damit wieder aus —
  // ohne dass jemand ihn zurücksetzen müsste.
  const [ueberschriebenFuer, setzeUeberschriebenFuer] = useState<PaketplanDTO>()

  const meldung = fehler !== undefined && (
    <p role="alert" className="text-fehler">
      {texte.bausteine.zustaende.fehler} {fehler}
    </p>
  )

  if (plan === undefined) {
    return meldung === false ? null : meldung
  }

  const ueberschreiben = ueberschriebenFuer === plan
  const schluessel = plan.vorgang === 'entfernen' ? 'loeschen' : plan.vorgang
  const frage = einsetzen(t.titel[schluessel], { paket: plan.paket })

  const inhalt = (
    <>
      {plan.vorgang !== 'entfernen' && (
        <p className="text-sm">
          {einsetzen(t.fassung, {
            von: plan.von === '' ? t.ohneFassung : plan.von,
            zu: plan.zu === '' ? t.ohneFassung : plan.zu,
          })}
        </p>
      )}

      {/* **Woher und was es darf** — DOM-01.AK-04: „nicht vertrauenswürdige
          Quellen erhalten keine stillen zusätzlichen Rechte". Die Quelle steht
          vor dem Einbinden da, und der Satz darunter sagt, was ein Paket
          überhaupt kann. */}
      {plan.quelle !== '' && (
        <>
          <p className="break-all text-sm">
            {einsetzen(t.quelle, { quelle: plan.quelle })}
          </p>
          <p className="max-w-prose text-sm text-leise">{t.keineRechte}</p>
        </>
      )}

      <Planliste
        ueberschrift={t.sets}
        leer={t.keineSets}
        eintraege={plan.sets}
      />

      <Planliste
        ueberschrift={t.konflikte}
        leer={t.keineKonflikte}
        eintraege={plan.konflikte.map((befund) =>
          abhaengigkeitssatz(befund, texte),
        )}
      />

      {plan.geaendert && <p className="text-sm">{t.planGeaendert}</p>}

      {plan.geaendert && plan.vorgang === 'ersetzen' && (
        <label className="flex items-start gap-2 text-sm">
          <input
            type="checkbox"
            checked={ueberschreiben}
            onChange={(ereignis) => {
              setzeUeberschriebenFuer(
                ereignis.target.checked ? plan : undefined,
              )
            }}
            className="mt-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
          <span>{t.ueberschreiben}</span>
        </label>
      )}

      {!plan.moeglich && (
        <p className="max-w-prose text-sm">{t.nichtMoeglich[schluessel]}</p>
      )}

      {/* **Der Fehler steht am Plan** — dort, wo der Haken sitzt, den eine
          Absage wegen einer lokalen Änderung verlangt (§6). */}
      {meldung}
    </>
  )

  if (!plan.moeglich) {
    return (
      <div
        role="group"
        aria-label={frage}
        className="flex flex-col gap-2 rounded-sm border border-kante bg-flaeche p-3"
      >
        <p className="font-medium">{frage}</p>
        {inhalt}
        <div>
          <Knopf aufKlick={planung.verwerfen}>{t.verwerfen}</Knopf>
        </div>
      </div>
    )
  }

  const gemeinsam = {
    frage,
    abbrechen: t.verwerfen,
    laeuft: planung.laeuft,
    aufBestaetigen: () => {
      void planung.ausfuehren(ueberschreiben)
    },
    aufAbbrechen: planung.verwerfen,
    children: inhalt,
  }

  return plan.vorgang === 'entfernen' ? (
    <Entscheidung
      {...gemeinsam}
      aktion="loeschen"
      bestaetigen={texte.seiten.pakete.loeschen}
    />
  ) : (
    <Entscheidung {...gemeinsam} bestaetigen={t.bestaetigen[schluessel]} />
  )
}

interface Listeneigenschaften {
  ueberschrift: string
  leer: string
  eintraege: readonly string[]
}

// Planliste ist ein Abschnitt des Plans — mit einem Satz statt einer leeren
// Liste, denn „nichts betroffen" ist eine Aussage (§6).
function Planliste({ ueberschrift, leer, eintraege }: Listeneigenschaften) {
  return (
    <div className="text-sm">
      <p className="font-medium">{ueberschrift}</p>
      {eintraege.length === 0 ? (
        <p className="text-leise">{leer}</p>
      ) : (
        <ul className="list-disc pl-5">
          {eintraege.map((eintrag) => (
            <li key={eintrag}>{eintrag}</li>
          ))}
        </ul>
      )}
    </div>
  )
}
