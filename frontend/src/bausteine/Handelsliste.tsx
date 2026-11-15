// **Was gerade läuft** — TRADE-01.AK-01/02/04
// (#1715).
//
// # Warum es diese Liste gibt
//
// **Weil man einen Handel bisher über sein Gegenüber fand.** Wer nicht mehr
// wusste, mit wem, fand ihn nicht — und der Handelsraum sagte selbst, dass
// eine Liste laufender Vorgänge fehlt. Hier stehen alle, mit dem, was als
// Nächstes zu tun ist; ein Klick schlägt Gegenüber und Vorgang auf.
//
// **Nach dem Artboard *Handeln***: die Spalte „Was gerade läuft" mit Zahl,
// je Zeile Gegenüber und „du bist dran". Suche und Auswahl darüber sind die
// Ergänzung, die AK-02 verlangt.
//
// # Warum die Auswahl ein Auswahlfeld ist und das Archiv ein Schalter
//
// **R7 aus `docs/UIREVIEW.md`** (UXR-15,
// #1923):
// fünf Werte sind mehr als vier, und was sofort wirkt und zwei Werte hat,
// ist ein Schalter. Hier standen fünf Umschalter, und der gedrückte sah aus
// wie ein Hauptknopf — die einzige führende Handlung der Spalte war ein
// Filter. **Die Ordnung bleibt ein Auswahlfeld mit drei Werten**: als
// Segmente brächen „Zuletzt geändert", „Was zu tun ist" und „Gegenüber" in
// 320 Punkten um, dieselbe Abwägung wie bei der Ausführung der Sammlung.
//
// # Und warum acht Zeilen
//
// **Weil die Liste mit dem Bestand wächst** (R2, R6): fünfundzwanzig
// abgeschlossene Handel waren fünfundzwanzig Knöpfe untereinander. Das
// Artboard zeigt vier; acht stehen in der Spalte über der Falz, der Rest
// kommt stückweise (`Faltliste`).
import { useEffect, useId, useState } from 'react'

import {
  handelsauswahlen,
  useHandelsliste,
  type Handelsauswahl,
} from '../haken/useHandelsliste'
import { useTexte } from '../sprache/kontext'
import { Handelslistenzeile } from './Handelslistenzeile'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Faltliste } from './Faltliste'
import { Knopf } from './Knopf'
import { Identitaetanlegen } from './Identitaetanlegen'
import { Ladefehler, Wirdgeladen } from './Ladehinweise'

import type { Handelsordnung, Handelspartner } from '../apihandelsverzeichnis'

interface Handelslisteneigenschaften {
  // gewaehlt ist der aufgeschlagene Vorgang — er wird markiert.
  gewaehlt: string
  // oeffnen bekommt das Gegenüber mit Namen — der Raum nennt es (UXR-15).
  oeffnen: (
    partner: Handelspartner,
    vorgang: string,
    archiviert: boolean,
  ) => void

  // stand zählt Änderungen an Vorgängen — jede lädt die Liste nach.
  stand: number

  // ohneIdentitaet meldet, ob diesem Profil die Identität fehlt (#2027) —
  // die Seite nimmt dann ihrem eigenen Hauptknopf das Gewicht (R1).
  ohneIdentitaet?: ((fehlt: boolean) => void) | undefined

  // vonSelbst bekommt den einen Handel, an dem du dran bist — nur gesetzt,
  // solange die Seite nichts aufgeschlagen hat (P13, #2381, H4).
  vonSelbst?: ((partner: Handelspartner, vorgang: string) => void) | undefined
}

const feldstil =
  'rounded-md border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

// dieZeilenVorn und derSchritt sind, wie viele Handel zu Anfang dastehen und
// wie viele ein Klick dazu zeigt.
const dieZeilenVorn = 8
const derSchritt = 20

const auswahlstil =
  'h-8 rounded-md border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Handelsliste({
  gewaehlt,
  oeffnen,
  stand,
  ohneIdentitaet,
  vonSelbst,
}: Handelslisteneigenschaften) {
  const eigene = useTexte().seiten.handeln.liste
  const ueberschrift = useId()
  const suchfeld = useId()
  const auswahlfeld = useId()
  const ordnungsfeld = useId()

  const [suche, setzeSuche] = useState('')
  const [auswahl, setzeAuswahl] = useState<Handelsauswahl>('offen')
  const [archiv, setzeArchiv] = useState(false)
  const [ordnung, setzeOrdnung] = useState<Handelsordnung>('')

  const liste = useHandelsliste({ suche, auswahl, archiv, ordnung }, stand)

  const fehlt = liste.daten === null
  useEffect(() => {
    ohneIdentitaet?.(fehlt)
  }, [fehlt, ohneIdentitaet])

  const gefiltert = suche !== '' || auswahl !== 'offen' || archiv

  // **Der eine Handel, an dem du dran bist, steht offen** (P13, #2381, H4):
  // wer ⌘5 drückt, weil ein Angebot kam, will zustimmen und nicht suchen.
  // Bei mehreren wählt ein Mensch — und nur in der Liste, wie sie ohne
  // Filter steht.
  const dran = gefiltert
    ? []
    : (liste.daten ?? []).filter((zeile) => zeile.lage === 'dran')
  const einziger = dran.length === 1 ? dran[0] : undefined
  const partner = einziger?.gegenueber.at(0)
  const vorgang = einziger?.vorgang.vorgang

  useEffect(() => {
    if (partner !== undefined && vorgang !== undefined) {
      vonSelbst?.(partner, vorgang)
    }
  }, [partner, vorgang, vonSelbst])

  const namen: Record<Handelsauswahl, string> = {
    offen: eigene.auswahlOffen,
    dran: eigene.auswahlDran,
    wartet: eigene.auswahlWartet,
    abgeschlossen: eigene.auswahlAbgeschlossen,
    abgebrochen: eigene.auswahlAbgebrochen,
  }

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-2">
      <h2
        id={ueberschrift}
        className="flex items-center gap-2 text-base font-semibold"
      >
        {eigene.titel}
        {liste.daten !== undefined && liste.daten !== null && (
          <span className="rounded-full bg-marke px-2 text-xs text-marketext">
            {liste.daten.length}
          </span>
        )}
      </h2>

      <Erklaerungsabsatz text={eigene.erklaerung} klasse="text-sm text-leise" />

      {/* **Ohne Identität gibt es nichts zu filtern** (#2027): der Leerzustand
          steht dann gleich unter der Überschrift und nicht unter der Falz. */}
      {liste.daten !== null && (
        <>
          <label className="flex flex-col gap-1 text-sm" htmlFor={suchfeld}>
            {eigene.suche}
            <input
              id={suchfeld}
              type="search"
              value={suche}
              placeholder={eigene.sucheHinweis}
              onChange={(ereignis) => {
                setzeSuche(ereignis.target.value)
              }}
              className={feldstil}
            />
          </label>

          <div className="flex flex-wrap items-end gap-3 text-sm">
            <label className="flex flex-col gap-1" htmlFor={auswahlfeld}>
              {eigene.auswahl}
              <select
                id={auswahlfeld}
                value={auswahl}
                onChange={(ereignis) => {
                  setzeAuswahl(ereignis.target.value as Handelsauswahl)
                }}
                className={auswahlstil}
              >
                {handelsauswahlen.map((art) => (
                  <option key={art} value={art}>
                    {namen[art]}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1" htmlFor={ordnungsfeld}>
              {eigene.ordnung}
              <select
                id={ordnungsfeld}
                value={ordnung}
                onChange={(ereignis) => {
                  setzeOrdnung(ereignis.target.value as Handelsordnung)
                }}
                className={auswahlstil}
              >
                <option value="">{eigene.ordnungGeaendert}</option>
                <option value="lage">{eigene.ordnungLage}</option>
                <option value="partner">{eigene.ordnungPartner}</option>
              </select>
            </label>
          </div>

          <div className="text-sm">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                role="switch"
                checked={archiv}
                onChange={(ereignis) => {
                  setzeArchiv(ereignis.target.checked)
                }}
                className="accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              />
              {eigene.archiv}
            </label>
          </div>
        </>
      )}

      {liste.daten === undefined ? (
        liste.fehler === undefined ? (
          <Wirdgeladen />
        ) : (
          <Ladefehler fehler={liste.fehler} wiederholen={liste.nachladen} />
        )
      ) : liste.daten === null ? (
        <Identitaetanlegen angelegt={liste.nachladen} />
      ) : liste.daten.length === 0 ? (
        <div className="flex flex-col items-start gap-2 text-sm text-leise">
          <p>{gefiltert ? eigene.leerGefiltert : eigene.leer}</p>
          {gefiltert && (
            <Knopf
              klein
              aufKlick={() => {
                setzeSuche('')
                setzeAuswahl('offen')
                setzeArchiv(false)
              }}
            >
              {eigene.zuruecksetzen}
            </Knopf>
          )}
        </div>
      ) : (
        <Faltliste
          eintraege={liste.daten}
          vorn={dieZeilenVorn}
          schritt={derSchritt}
          klasse="flex flex-col gap-1.5"
          zeile={(zeile) => {
            // **Ein Handel ohne Gegenüber lässt sich nicht aufschlagen** —
            // der Raum gehört zu einem; nach der Matrix gibt es ihn nicht,
            // in einem von Hand geänderten Bestand schon.
            const partner = zeile.gegenueber.at(0)

            return (
              <li key={zeile.vorgang.vorgang}>
                <Handelslistenzeile
                  zeile={zeile}
                  gewaehlt={zeile.vorgang.vorgang === gewaehlt}
                  oeffnen={
                    partner === undefined
                      ? undefined
                      : () => {
                          oeffnen(
                            partner,
                            zeile.vorgang.vorgang,
                            zeile.archiviert,
                          )
                        }
                  }
                />
              </li>
            )
          }}
        />
      )}
    </section>
  )
}
