// **Was ein Auftrag draußen kosten darf** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Warum vier Zahlen und nicht eine
//
// **Weil „noch 34,50 € übrig" die Frage nicht beantwortet, die jemand hier
// stellt.** Wer wissen will, ob er einen Auftrag anstoßen kann, will wissen,
// **was schon gebunden ist** — und ob etwas davon ungeklärt ist. Eine Zahl
// verschwiege den Unterschied zwischen „ausgegeben" und „wir wissen es nicht".
import { useId, useState } from 'react'

import { alsGeld } from '../geld'

import type { Kostenlauf } from '../haken/useKosten'
import { useKosten } from '../haken/useKosten'
import { useVorgang } from '../haken/useVorgang'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Feld } from '../bausteine/Feld'
import { Waehrungsfeld } from '../bausteine/Waehrungsfeld'
import { waehrungsbefund } from '../waehrungsgestalt'
import { ziffernbetrag } from '../ziffernbetrag'

const knopf =
  'rounded-md border border-kante px-3 py-1.5 text-sm font-medium ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60'

const feld =
  'rounded-md border border-kante bg-grund px-2 py-1.5 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

// Vorgabekonto ist das Konto, das die Seite ohne Zutun zeigt.
//
// **Eines und keine Liste**: es gibt heute keinen zahlenden Anbieter, und eine
// Auswahl über eine leere Liste wäre ein Bedienelement ohne Gegenstand. Wer
// eines einrichtet, tippt seinen Namen — dieselbe Gestalt wie beim Zugang.
const Vorgabekonto = 'anbieter'

export function Kosten() {
  const texte = useTexte()
  const worte = texte.seiten.kosten
  const kontoID = useId()
  const zeitraumID = useId()

  const [konto, setzeKonto] = useState(Vorgabekonto)
  const [zeitraum, setzeZeitraum] = useState('tag')

  const stand = useKosten(useVorgang(), konto, zeitraum)

  return (
    // **Ohne eigene Überschrift seit #2465**: die Kosten stehen in einem
    // Abschnitt der Einstellungen, und dessen Knopf trägt den Namen.
    <section className="max-w-xl">
      <p className="max-w-prose text-sm text-leise">{worte.hinweis}</p>

      <div className="mt-3 flex flex-wrap items-end gap-2">
        <label className="flex flex-col gap-1 text-sm" htmlFor={kontoID}>
          {worte.kontoWaehlen}

          <input
            id={kontoID}
            className={feld}
            value={konto}
            onChange={(ereignis) => {
              setzeKonto(ereignis.target.value)
            }}
          />
        </label>

        <label className="flex flex-col gap-1 text-sm" htmlFor={zeitraumID}>
          {worte.zeitraumWaehlen}

          <select
            id={zeitraumID}
            className={feld}
            value={zeitraum}
            onChange={(ereignis) => {
              setzeZeitraum(ereignis.target.value)
            }}
          >
            {/* **Die beiden kommen aus dem Katalog** und stehen nicht
                daneben: zwei Aufzählungen derselben Sache laufen an dem Tag
                auseinander, an dem eine dritte dazukommt. */}
            {Object.entries(worte.zeitraeume).map(([marke, name]) => (
              <option key={marke} value={marke}>
                {name}
              </option>
            ))}
          </select>
        </label>
      </div>

      {stand.stand !== undefined && <Zahlen stand={stand.stand} />}

      <Freigabe stand={stand} />

      <Posten stand={stand} />

      <Erklaerungsabsatz
        text={worte.lokalHinweis}
        klasse="mt-3 max-w-prose text-sm text-leise"
      />

      {stand.fehler !== undefined && (
        <p role="alert" className="mt-4 text-fehler">
          {texte.bausteine.zustaende.fehler} {stand.fehler}
        </p>
      )}
    </section>
  )
}

// Zahlen zeigt die vier Zahlen samt Grenze.
function Zahlen({ stand }: { stand: NonNullable<Kostenlauf['stand']> }) {
  const worte = useTexte().seiten.kosten
  const darstellung = useDarstellung()

  const zeilen = [
    { name: worte.grenze, betrag: stand.grenze },
    { name: worte.reserviert, betrag: stand.reserviert },
    { name: worte.verbraucht, betrag: stand.verbraucht },
    { name: worte.unklar, betrag: stand.unklar },
    { name: worte.verfuegbar, betrag: stand.verfuegbar },
  ]

  return (
    <div className="mt-4">
      {stand.grenze.wert === 0 && (
        <p className="max-w-prose text-sm text-leise">{worte.ohneGrenze}</p>
      )}

      <dl className="mt-2 grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
        {zeilen.map((zeile) => (
          <div key={zeile.name} className="contents">
            <dt className="text-leise">{zeile.name}</dt>
            <dd className="text-right tabular-nums">
              {alsGeld(zeile.betrag.wert, zeile.betrag.waehrung, darstellung)}
            </dd>
          </div>
        ))}
      </dl>

      {stand.unklar.wert > 0 && (
        <Erklaerungsabsatz
          text={worte.unklarHinweis}
          klasse="mt-2 max-w-prose text-sm text-leise"
        />
      )}
    </div>
  )
}

// Freigabe nimmt eine neue Grenze entgegen.
function Freigabe({ stand }: { stand: Kostenlauf }) {
  const texte = useTexte()
  const worte = texte.seiten.kosten

  const [betrag, setzeBetrag] = useState('')
  const [waehrung, setzeWaehrung] = useState('EUR')

  // **Nur Ziffern** — »1.000« wäre für `Number` ein Cent (N5).
  const zahl = ziffernbetrag(betrag)
  // **Eine Währung ohne Gestalt geht nicht hinüber** (UX-06.AK-03): der
  // Dienst wiese sie ab, und das Feld hat den Grund schon genannt.
  const brauchbar =
    zahl !== undefined &&
    waehrung.trim() !== '' &&
    waehrungsbefund(waehrung, false) === undefined

  return (
    <div className="mt-5">
      <h3 className="text-base font-medium">{worte.grenzeSetzen}</h3>

      <div className="mt-2 flex flex-wrap items-end gap-2">
        <Feld
          beschriftung={worte.betragTippen}
          {...(betrag === '' || zahl !== undefined
            ? {}
            : { befund: texte.allgemein.nurZiffern })}
        >
          {(verbindung) => (
            <input
              {...verbindung}
              className={feld}
              inputMode="numeric"
              value={betrag}
              onChange={(ereignis) => {
                setzeBetrag(ereignis.target.value)
              }}
            />
          )}
        </Feld>

        {/* **Groß, und das wird gesagt statt still gemacht** — `Betrag.Pruefe`
            weist `eur` ab; das Feld nennt vorher, wie es heißt (UX-06). */}
        <Waehrungsfeld
          beschriftung={worte.waehrungTippen}
          wert={waehrung}
          setzeWert={setzeWaehrung}
          kleinErlaubt={false}
        />

        <button
          type="button"
          className={knopf}
          disabled={stand.laeuft || !brauchbar}
          onClick={
            zahl === undefined
              ? undefined
              : () => {
                  stand.setzeGrenze(zahl, waehrung)
                }
          }
        >
          {stand.laeuft ? worte.setztGerade : worte.setzen}
        </button>
      </div>

      <p className="mt-1 text-sm text-leise">{worte.nullNimmtWeg}</p>
    </div>
  )
}

// Posten zeigt die letzten Aufträge.
function Posten({ stand }: { stand: Kostenlauf }) {
  const worte = useTexte().seiten.kosten
  const darstellung = useDarstellung()

  const posten = stand.posten?.posten ?? []

  return (
    <div className="mt-5">
      <h3 className="text-base font-medium">{worte.postenTitel}</h3>

      {posten.length === 0 ? (
        <p className="mt-1 text-sm text-leise">{worte.keinePosten}</p>
      ) : (
        <ul className="mt-2 flex flex-col gap-1 text-sm">
          {posten.map((einer) => (
            <li key={einer.kennung}>
              {einsetzen(worte.postenZeile, {
                aufgabe: einer.aufgabe,
                stand: worte.staende[einer.stand] ?? einer.stand,
                betrag: alsGeld(
                  einer.stand === 'abgerechnet'
                    ? einer.betrag.wert
                    : einer.hoechstbetrag.wert,
                  einer.hoechstbetrag.waehrung,
                  darstellung,
                ),
              })}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
