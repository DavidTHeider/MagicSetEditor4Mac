// Der Bestand als Gegenstand: sichern und zurückspielen
// (#940).
//
// **Eine eigene Datei, weil es eine eigene Frage ist** (§2): die Einstellungen
// daneben ändern, wie die Anwendung aussieht und spricht; hier geht es um die
// Daten selbst.
//
// # Warum die Rückfrage im Seitenbereich steht und nicht in einem Dialog
//
// §6: keine verschachtelten Dialoge. Der native Dateidialog ist schon einer —
// eine Bestätigung darüber wäre der zweite. Was gewählt wurde, steht deshalb
// **auf der Seite**, mit dem, was darin liegt, und daneben die zwei Knöpfe.
import { useId } from 'react'

import { Bestandschiffre } from '../bausteine/Bestandschiffre'
import { Sicherungsalter } from '../bausteine/Sicherungsalter'
import { Knopf } from '../bausteine/Knopf'
import {
  Kennwortabfrage,
  Sicherungsformular,
} from '../bausteine/Sicherungskennwort'
import type { Verschluesselungslauf } from '../haken/useVerschluesselung'
import type { Sicherungsstandlauf } from '../haken/useSicherungsstand'
import type { Bilderlage, Sicherungsauswahl } from '../api'
import type { RuecksprunglageAntwort } from '../apirueckkehr'
import { alsZeitpunkt } from '../zeitpunkt'
import { Ruecksprunglage } from '../bausteine/Ruecksprunglage'
import { useDarstellung, useZeitzone, useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { hauptknopf, nebenknopf as knopf } from '../bausteine/knopfklassen'

interface Bestandspflegeeigenschaften {
  sichern: (kennwort: string, klartext: boolean) => void
  sichertGerade: boolean

  waehlen: () => void

  // gesperrt ist der Pfad einer gewählten Sicherung, die ihr Kennwort
  // braucht (OPS-06, #1723); entsperren öffnet sie damit.
  gesperrt: string | undefined
  entsperren: (pfad: string, kennwort: string) => void
  zurueckspielen: (pfad: string) => void
  spieltGerade: boolean

  auswahl: Sicherungsauswahl | undefined

  // lage ist der Vergleich zur gewählten Sicherung (O7, #1431) — was das
  // Zurückspielen kostet und was es unklar ließe.
  lage: RuecksprunglageAntwort | undefined

  beiseite: string | undefined
  fehler: string | undefined
  abbrechen: () => void

  // chiffre schaltet die Verschlüsselung — fehlt dort, wo der Bestand gar
  // nicht offen ist (der Wiederherstellungsweg beim Start).
  chiffre?: Verschluesselungslauf

  // alter ist, wann zuletzt gesichert wurde, und der Plan (OPS01, #2347) —
  // fehlt dort, wo kein Bestand offen ist.
  alter?: Sicherungsstandlauf

  // nurZurueckspielen lässt den Sicherungsknopf weg.
  //
  // **Für die Wiederherstellung** (#962): steht kein Bestand offen, gibt es
  // nichts zu sichern, und ein Knopf, der nur scheitern kann, ist ein
  // falsches Versprechen. Die Seite ist im Übrigen **dieselbe** — der Weg
  // führt über dieselben Bindungen wie sonst und nicht über eine zweite
  // Fassung derselben Frage.
  nurZurueckspielen?: boolean
}

export function Bestandspflege({
  sichern,
  sichertGerade,
  waehlen,
  gesperrt,
  entsperren,
  zurueckspielen,
  spieltGerade,
  auswahl,
  lage,
  beiseite,
  fehler,
  abbrechen,
  chiffre,
  alter,
  nurZurueckspielen = false,
}: Bestandspflegeeigenschaften) {
  const texte = useTexte()
  const titelID = useId()

  const worte = texte.seiten.bestandspflege

  return (
    <section aria-labelledby={titelID} className="mt-10 max-w-xl">
      <h2 id={titelID} className="text-lg font-semibold tracking-tight">
        {worte.titel}
      </h2>

      <p className="mt-1 max-w-prose text-sm text-leise">{worte.hinweis}</p>

      {alter !== undefined && <Sicherungsalter lauf={alter} />}

      {/* **Das Kennwort vor dem Dialog** (OPS-06, #1723): ein zu kurzes
          oder ein fehlendes „ohne" gehört ans Feld und nicht hinter einen
          Dialog, den jemand erst bestätigt. */}
      {!nurZurueckspielen && (
        <Sicherungsformular
          sichern={sichern}
          angehalten={sichertGerade || spieltGerade}
          verschluesselt={chiffre?.verschluesselt === true}
        />
      )}

      <div className="mt-4 flex flex-wrap gap-3">
        <Knopf
          klein
          angehalten={sichertGerade || spieltGerade}
          aufKlick={waehlen}
        >
          {worte.waehlen}
        </Knopf>
      </div>

      {gesperrt !== undefined && (
        <Kennwortabfrage
          key={gesperrt}
          pfad={gesperrt}
          entsperren={(kennwort) => {
            entsperren(gesperrt, kennwort)
          }}
          abbrechen={abbrechen}
          angehalten={spieltGerade}
          fehler={fehler}
        />
      )}

      {/* **Die Rückfrage steht erst da, wenn etwas geprüft ist.** Eine
          Bestätigung ohne Gegenstand wäre eine Frage ins Leere. */}
      {auswahl !== undefined && (
        <Rueckfrage
          auswahl={auswahl}
          lage={lage}
          laeuft={spieltGerade}
          bestaetigen={() => {
            zurueckspielen(auswahl.pfad)
          }}
          abbrechen={abbrechen}
        />
      )}

      {beiseite !== undefined && (
        <p role="status" className="mt-4 text-sm text-leise">
          {einsetzen(worte.zurueckgespielt, { beiseite })}
        </p>
      )}

      {/* **Beim Entsperren steht der Fehler am Feld** (§6) und nicht
          noch einmal hier. */}
      {fehler !== undefined && gesperrt === undefined && (
        <p role="alert" className="mt-4 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {/* **Die Verschlüsselung unter der Sicherung und nicht darüber.**
          Die Reihenfolge ist der Rat: wer verschlüsselt, soll vorher
          gesichert haben — der Schlüssel ist nicht wiederherstellbar. */}
      {chiffre !== undefined && (
        <Bestandschiffre
          verschluesselt={chiffre.verschluesselt}
          umschalten={chiffre.umschalten}
          laeuft={chiffre.laeuft}
          fehler={chiffre.fehler}
        />
      )}
    </section>
  )
}

interface Rueckfrageeigenschaften {
  auswahl: Sicherungsauswahl
  lage: RuecksprunglageAntwort | undefined
  laeuft: boolean
  bestaetigen: () => void
  abbrechen: () => void
}

// Rueckfrage zeigt, was in der gewählten Sicherung steht — und fragt.
//
// **Sie nennt die Zahlen, bevor sie fragt.** „Wirklich zurückspielen?" ist
// keine Frage, die jemand beantworten kann; „107 Karten aus einem Bestand vom
// 3. August" ist eine.
function Rueckfrage({
  auswahl,
  lage,
  laeuft,
  bestaetigen,
  abbrechen,
}: Rueckfrageeigenschaften) {
  const texte = useTexte()
  const worte = texte.seiten.bestandspflege
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  // **Was mit den Bildern geschieht, hängt am Format der gewählten Datei**
  // (#1162).
  // Ein Satz für alle drei Fälle sagte für eine nackte `.db` das Gegenteil
  // dessen, was geschieht.
  const bilderworte: Record<Bilderlage, string> = {
    ersetzt: worte.bilderErsetzt,
    geleert: worte.bilderGeleert,
    unberuehrt: worte.bilderUnberuehrt,
  }

  /// **−1 heißt „diese Sicherung kennt die Tabelle nicht"** und ist etwas
  // anderes als 0. Eine Zahl daneben, die nicht stimmt, wäre schlimmer als ein
  // Strich.
  const zahl = (wert: number) => (wert < 0 ? worte.unbekannt : String(wert))

  return (
    <div className="mt-4 rounded-md border border-kante p-4">
      <p className="font-medium">{worte.rueckfrage}</p>

      <dl className="mt-3 grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
        <dt className="text-leise">{worte.datei}</dt>
        <dd className="break-all">{auswahl.pfad}</dd>

        <dt className="text-leise">{worte.angelegtAm}</dt>
        {/* **Lesbar und nicht roh** — das fand die Sichtprüfung zu O7
            (#1431): daneben steht seither der wiederhergestellte Stand, und
            zwei Zeitangaben in zwei Formen nebeneinander sehen aus wie ein
            Fehler. Der Grund, warum das im Browser geschieht und nicht in Go,
            steht in `zeitpunkt.ts`. */}
        <dd>{alsZeitpunkt(auswahl.angelegtAm, darstellung, zeitzone)}</dd>

        <dt className="text-leise">{worte.sets}</dt>
        <dd>{zahl(auswahl.sets)}</dd>

        <dt className="text-leise">{worte.karten}</dt>
        <dd>{zahl(auswahl.karten)}</dd>

        <dt className="text-leise">{worte.bilderLage}</dt>
        <dd className="max-w-prose">{bilderworte[auswahl.bilder]}</dd>

        <dt className="text-leise">{worte.verschluesselt}</dt>
        <dd>{auswahl.verschluesselt ? worte.ja : worte.nein}</dd>

        {/* **Was fehlt, steht vor dem Bestätigen da** (OPS-06.AK-03): ein
            bezogenes Paket kommt nicht mit und muss neu bezogen werden. */}
        <dt className="text-leise">{worte.paketeEnthalten}</dt>
        <dd className="break-words">
          {auswahl.paketeEnthalten.join(', ') || worte.keine}
        </dd>

        <dt className="text-leise">{worte.paketeFehlen}</dt>
        <dd className="break-words">
          {auswahl.paketeFehlen.join(', ') || worte.keine}
        </dd>
      </dl>

      <p className="mt-3 max-w-prose text-sm text-leise">{worte.warnung}</p>

      {/* **Was es kostet, steht vor den Knöpfen** (O7, #1431). Danach hätte
          es niemand gelesen, der schon geklickt hat — dieselbe Überlegung wie
          beim Satz über die Shell im technischen Bereich. */}
      {lage !== undefined && <Ruecksprunglage lage={lage} />}

      <div className="mt-4 flex gap-3">
        <button
          type="button"
          className={hauptknopf}
          disabled={laeuft}
          onClick={bestaetigen}
        >
          {worte.bestaetigen}
        </button>

        <button
          type="button"
          className={knopf}
          disabled={laeuft}
          onClick={abbrechen}
        >
          {worte.abbrechen}
        </button>
      </div>
    </div>
  )
}
