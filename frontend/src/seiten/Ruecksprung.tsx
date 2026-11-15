// **Zurück ist nicht eine Handlung, sondern vier** — Werkstück O7 von Phase
// 5.2 (#1431).
//
// # Warum die vier Arten überhaupt dastehen
//
// Bis hierher gab es **einen** Knopf: *Sicherung zurückspielen*. Er tauscht
// die Daten — und wer ihn drückt, weil eine neue Fassung nicht tut, was sie
// soll, hat die Fassung damit nicht gewechselt. Was hier steht, ist die
// Trennung: vier Dinge können zurück, sie kosten Verschiedenes, und **zwei
// davon kann dieses Programm nicht selbst**. Ein Knopf, der sicher nein sagt,
// wäre schlimmer als ein Satz, der erklärt, wer es tut.
//
// # Und warum die verbuchten Rückkehren daneben stehen
//
// Ein historisches Backup kann bereits abgegebene Karten enthalten. Nach dem
// Zurückspielen stehen sie wieder da, und niemand weiß, ob sie noch jemandem
// gehören. **Diese Unsicherheit verjährt nicht** — sie endet, wenn ein Mensch
// nachgesehen und quittiert hat.
import { useId } from 'react'

import type { Ruecksprungstand } from '../haken/useRuecksprung'
import { alsZeitpunkt } from '../zeitpunkt'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useZeitzone, useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

const knopf =
  'rounded-md border border-kante px-3 py-1.5 text-sm font-medium ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60'

export function Ruecksprung({ stand }: { stand: Ruecksprungstand }) {
  const texte = useTexte()
  const worte = texte.seiten.ruecksprung
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const titelID = useId()

  return (
    <section aria-labelledby={titelID} className="max-w-xl">
      <h2 id={titelID} className="text-lg font-semibold tracking-tight">
        {worte.titel}
      </h2>

      <p className="mt-1 max-w-prose text-sm text-leise">{worte.hinweis}</p>

      <ul className="mt-4 flex flex-col gap-3">
        {stand.arten.map((art) => (
          <li key={art.art} className="flex flex-col gap-0.5">
            <span className="text-sm font-medium">
              {worte.arten[art.art] ?? art.art}
            </span>

            <span className="max-w-prose text-sm text-leise">
              {art.hindernis === ''
                ? worte.gehtHier
                : (worte.hindernisse[art.hindernis] ?? art.hindernis)}
            </span>
          </li>
        ))}
      </ul>

      <h3 className="mt-8 text-base font-medium">{worte.rueckkehrenTitel}</h3>

      {stand.rueckkehren.length === 0 ? (
        <p className="mt-1 max-w-prose text-sm text-leise">
          {worte.keineRueckkehren}
        </p>
      ) : (
        <ul className="mt-3 flex flex-col gap-4">
          {stand.rueckkehren.map((eine) => (
            <li key={eine.id} className="rounded-md border border-kante p-3">
              <p className="text-sm">
                {einsetzen(worte.rueckkehrZeile, {
                  zurueckgespielt: alsZeitpunkt(
                    eine.zurueckgespielt,
                    darstellung,
                    zeitzone,
                  ),
                  datenstand: alsZeitpunkt(
                    eine.datenstand,
                    darstellung,
                    zeitzone,
                  ),
                })}
              </p>

              <p className="mt-0.5 text-sm break-all text-leise">
                {einsetzen(worte.ausDatei, { quelle: eine.quelle })}
              </p>

              <p className="mt-1 text-sm text-leise">
                {eine.klaerung === 'quittiert'
                  ? einsetzen(worte.quittiert, {
                      geklaertAm: alsZeitpunkt(
                        eine.geklaertAm,
                        darstellung,
                        zeitzone,
                      ),
                    })
                  : worte.offen}
                {eine.mengen.length > 0 &&
                  ` · ${einsetzen(worte.betrifft, {
                    menge: String(
                      eine.mengen.reduce(
                        (summe, menge) => summe + menge.menge,
                        0,
                      ),
                    ),
                  })}`}
              </p>

              {/* **Quittieren steht nur bei einer offenen** — sonst wäre es
                  ein Knopf, der sicher nichts tut. */}
              {eine.klaerung !== 'quittiert' && (
                <div className="mt-3">
                  <button
                    type="button"
                    className={knopf}
                    disabled={stand.laeuft(eine.id)}
                    onClick={() => {
                      stand.quittieren(eine.id)
                    }}
                  >
                    {stand.laeuft(eine.id)
                      ? worte.quittierenLaeuft
                      : worte.quittieren}
                  </button>

                  <Erklaerungsabsatz
                    text={worte.quittierenHinweis}
                    klasse="mt-1 max-w-prose text-sm text-leise"
                  />
                </div>
              )}
            </li>
          ))}
        </ul>
      )}

      {stand.fehler !== undefined && (
        <p role="alert" className="mt-4 text-fehler">
          {texte.bausteine.zustaende.fehler} {stand.fehler}
        </p>
      )}
    </section>
  )
}
