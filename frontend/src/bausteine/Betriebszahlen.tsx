// **Laufzeiten und Verbrauch** — OPS-05.AK-01/04
// (#1722).
//
// # Die Einheit steht an jeder Zahl
//
// **Aus Go kommen Millisekunden, Mikrosekunden und Bytes**, und der Name des
// Feldes sagt, welche. Hier werden sie in die Form gebracht, in der jemand
// sie liest — Megabyte statt Bytes, Sekunden ab tausend Millisekunden — und
// die Einheit steht dabei.
//
// # Und was fehlt, steht auch da
//
// **Die Bindungen des Fensters misst niemand**, und eine leere Laufzeittafel
// am Fenster sähe sonst aus wie ein Fenster, das nichts getan hat. Die
// Antwort sagt es selbst (`ungemessen`), und hier steht es in Worten.
import { alsGroesse } from '../seiten/Diagnosepaket'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { alsZeitpunkt } from '../zeitpunkt'
import { alsDauer } from './dauer'
import { Laufzeitentafel } from './Laufzeitentafel'

import type { MetrikenAntwort, ProzessDTO } from '../apimetriken'

export function Betriebszahlen({ zahlen }: { zahlen: MetrikenAntwort }) {
  const technik = useTexte().seiten.technik
  const eigene = technik.betrieb
  const darstellung = useDarstellung()
  const zahl = (wert: number) => wert.toLocaleString(darstellung)

  const namen = eigene.ungemessenNamen as Record<string, string>
  const ungemessen = zahlen.ungemessen.map((name) => namen[name] ?? name)

  return (
    <div className="flex flex-col gap-2">
      <h3 className="text-lg font-semibold">{eigene.ueberschrift}</h3>

      {/* **Der Aufwand der Erhebung als eigener Satz** (AP-11, #2368, B-27)
          und nicht als Anhang an den Zeitraum. Sichtbar bleibt er, weil
          OPS-05.AK-04 ihn verlangt — ein Titel ist mit der Tastatur nicht
          zu erreichen. */}
      <p className="text-sm text-leise">
        <span>
          {einsetzen(eigene.zeitraum, {
            seit: alsZeitpunkt(zahlen.seit, darstellung),
            bis: alsZeitpunkt(zahlen.bis, darstellung),
          })}
        </span>{' '}
        <span>
          {einsetzen(eigene.aufwand, { mikros: zahl(zahlen.aufwandMikros) })}
        </span>
      </p>

      <dl className="grid max-w-2xl grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
        <dt className="text-leise">{eigene.speicher}</dt>
        <dd>
          {einsetzen(
            zahlen.prozess.fussabdruckBytes > 0
              ? eigene.speicherWert
              : eigene.speicherOhneFussabdruck,
            {
              fussabdruck: alsGroesse(
                zahlen.prozess.fussabdruckBytes,
                darstellung,
              ),
              heap: alsGroesse(zahlen.prozess.heapBytes, darstellung),
              system: alsGroesse(zahlen.prozess.systemBytes, darstellung),
            },
          )}
        </dd>

        <dt className="text-leise">{eigene.web}</dt>
        <dd>
          <Webspeicher prozess={zahlen.prozess} />
        </dd>

        <dt className="text-leise">{eigene.goroutinen}</dt>
        <dd>{zahl(zahlen.prozess.goroutinen)}</dd>

        <dt className="text-leise">{eigene.rechenzeit}</dt>
        <dd>{alsDauer(zahlen.prozess.cpuMs, darstellung)}</dd>

        <dt className="text-leise">{eigene.datenbank}</dt>
        <dd>
          {zahlen.datenbank === null
            ? eigene.ohneBestand
            : einsetzen(eigene.datenbankWert, {
                offen: zahl(zahlen.datenbank.offen),
                inBenutzung: zahl(zahlen.datenbank.inBenutzung),
                gewartet: zahl(zahlen.datenbank.gewartet),
                dauer: alsDauer(zahlen.datenbank.wartezeitMs, darstellung),
              })}
        </dd>

        <dt className="text-leise">{eigene.zwischenspeicher}</dt>
        <dd>
          {einsetzen(eigene.zwischenspeicherWert, {
            treffer: zahl(zahlen.speicher.treffer),
            geholt: zahl(zahlen.speicher.geholt),
            veraltet: zahl(zahlen.speicher.veraltet),
            fehlgeschlagen: zahl(zahlen.speicher.fehlgeschlagen),
          })}
        </dd>
      </dl>

      <Laufzeitentafel zahlen={zahlen} />

      <p className="text-sm text-leise">
        {einsetzen(eigene.ungemessen, { liste: ungemessen.join('; ') })}
      </p>

      <p className="text-sm text-leise">
        {zahlen.spur.an
          ? einsetzen(eigene.spurAn, {
              aufgenommen: zahl(zahlen.spur.aufgenommen),
              gesehen: zahl(zahlen.spur.gesehen),
              abtastung: zahl(zahlen.spur.abtastung),
              grenze: zahl(zahlen.spur.grenze),
              minuten: zahl(zahlen.spur.aufbewahrungS / 60),
              verdraengt: zahl(zahlen.spur.verdraengt),
              abgelaufen: zahl(zahlen.spur.abgelaufen),
            })
          : eigene.spurAus}
      </p>
    </div>
  )
}

// Webspeicher nennt den Fußabdruck der WebKit-Prozesse (OPS02, #2348) — oder
// warum er fehlt. **Leer und ohne Grund heißt: keine Webansicht**, wie am
// Dienst und am Werkzeugserver.
function Webspeicher({ prozess }: { prozess: ProzessDTO }) {
  const eigene = useTexte().seiten.technik.betrieb
  const darstellung = useDarstellung()

  if (prozess.webUngemessen !== '') {
    return einsetzen(eigene.webUngemessen, { grund: prozess.webUngemessen })
  }

  if (prozess.web.length === 0) {
    return eigene.ohneWeb
  }

  return einsetzen(eigene.webWert, {
    summe: alsGroesse(
      prozess.web.reduce((summe, web) => summe + web.fussabdruckBytes, 0),
      darstellung,
    ),
    teile: prozess.web
      .map((web) =>
        einsetzen(eigene.webTeil, {
          art: eigene.webArten[web.art],
          groesse: alsGroesse(web.fussabdruckBytes, darstellung),
        }),
      )
      .join(', '),
  })
}
