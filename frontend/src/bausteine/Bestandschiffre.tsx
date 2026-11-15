// **Die Verschlüsselung des Bestands, mit ihren Grenzen daneben** —
// Werkstück 14 von Phase 4.12
// (#1351),
// nach Vermerk 0072.
//
// # Warum ein eigenes Bauteil und kein Abschnitt in `Bestandspflege`
//
// **Weil jene sonst über die Zeilengrenze wüchse** (§2) — und weil der
// Schnitt der Frage folgt: *wie trage ich den Bestand weg* ist eine andere
// als *wer darf ihn lesen*.
//
// # Und warum die Grenzen mit im Bild stehen
//
// Der Vermerk sagt ausdrücklich, was Adiantum **nicht** leistet: keinen Schutz
// gegen Fälschung, und wer mehrere Abzüge vergleichen kann, sieht, welche
// Stellen sich geändert haben. Ein Versprechen ohne seine Grenze ist eine
// Falle — und der teuerste Satz steht daneben: **der Schlüssel ist nicht
// wiederherstellbar.**
import { useId, useState } from 'react'

import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { hauptknopf, nebenknopf as knopf } from './knopfklassen'
import { useTexte } from '../sprache/kontext'

interface Eigenschaften {
  // verschluesselt ist der Stand — `undefined`, solange er nicht feststeht.
  verschluesselt: boolean | undefined

  umschalten: (an: boolean) => void
  laeuft: boolean
  fehler: string | undefined
}

export function Bestandschiffre({
  verschluesselt,
  umschalten,
  laeuft,
  fehler,
}: Eigenschaften) {
  const texte = useTexte()
  const titelID = useId()
  const worte = texte.seiten.bestandspflege

  // **Die Rückfrage ist lokaler Zustand.** Sie gehört zu diesem Bauteil und zu
  // keinem Datensatz; wer sie nach oben reichte, machte aus einer Geste einen
  // Stand, den jemand aufräumen muss.
  const [gefragt, setzeGefragt] = useState(false)

  // **Solange der Stand nicht feststeht, steht hier nichts.** Ein Knopf, der
  // rät, welche Richtung er hat, schaltet irgendwann in die falsche.
  if (verschluesselt === undefined) {
    return (
      <section aria-labelledby={titelID} className="mt-10 max-w-xl">
        <h2 id={titelID} className="text-lg font-semibold tracking-tight">
          {worte.chiffreTitel}
        </h2>

        {fehler !== undefined && (
          <p role="alert" className="mt-4 text-fehler">
            {texte.bausteine.zustaende.fehler} {fehler}
          </p>
        )}
      </section>
    )
  }

  const ziel = !verschluesselt

  return (
    <section aria-labelledby={titelID} className="mt-10 max-w-xl">
      <h2 id={titelID} className="text-lg font-semibold tracking-tight">
        {worte.chiffreTitel}
      </h2>

      <p className="mt-1 max-w-prose text-sm text-leise">
        {verschluesselt ? worte.chiffreAnHinweis : worte.chiffreAusHinweis}
      </p>

      {/* **Die Grenzen im ersten Satz, die Gründe hinter dem Hilfesymbol**
          (UX-15, #1771: lang und ganz sichtbar, gemessen in WebKit am 2026-09-24) — was man vor dem Verschlüsseln wissen muss,
          bleibt sichtbar. */}
      <Erklaerungsabsatz
        text={worte.chiffreGrenzen}
        klasse="mt-2 max-w-prose text-sm text-leise"
      />

      {!gefragt && (
        <div className="mt-4">
          <button
            type="button"
            className={knopf}
            disabled={laeuft}
            onClick={() => {
              setzeGefragt(true)
            }}
          >
            {verschluesselt
              ? worte.chiffreAusschalten
              : worte.chiffreAnschalten}
          </button>
        </div>
      )}

      {gefragt && (
        <div className="mt-4 rounded-md border border-kante p-4">
          <p className="font-medium">
            {ziel ? worte.chiffreFrageAn : worte.chiffreFrageAus}
          </p>

          <p className="mt-2 max-w-prose text-sm text-leise">
            {ziel ? worte.chiffreWarnungAn : worte.chiffreWarnungAus}
          </p>

          <div className="mt-4 flex flex-wrap gap-3">
            <button
              type="button"
              className={hauptknopf}
              disabled={laeuft}
              onClick={() => {
                setzeGefragt(false)
                umschalten(ziel)
              }}
            >
              {ziel ? worte.chiffreAnschalten : worte.chiffreAusschalten}
            </button>

            <button
              type="button"
              className={knopf}
              disabled={laeuft}
              onClick={() => {
                setzeGefragt(false)
              }}
            >
              {worte.abbrechen}
            </button>
          </div>
        </div>
      )}

      {laeuft && (
        <p role="status" className="mt-4 text-sm text-leise">
          {worte.chiffreLaeuft}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="mt-4 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}
    </section>
  )
}
