// **Kontakte aus Apple Kontakte** — ID-04
// (#1703),
// Vermerk 0137. Was die Seite tut, steht in `haken/useAdressbuch.ts` (§2).
//
// **Eine Ergänzung am Artboard *Gegenüber***: der Entwurf zeigt das Pairing;
// die Kontakte stehen daneben in einem eigenen Kasten, weil ein Kontakt
// **kein** Gegenüber ist (AK-04).
//
// # Lange Listen gefaltet
//
// **UXR-16** (#1924).
// Apple Kontakte gibt „alle Kontakte" als **eine** vCard aus. Mit 150
// übernommenen Kontakten und einer Datei von 200 Karten gezählt am
// 2026-10-02: **154 Bedienelemente** in der Liste und in der Vorschau
// **600 Kästchen** — je Karte eines und je Weg eines, auch für die 150, die
// unverändert schon dastehen und gar nicht wählbar waren (R2, R3, R6 in
// `docs/UIREVIEW.md`). Jetzt zeigen Liste und Vorschau die ersten und den
// Rest stückweise (`Faltliste`); was unverändert ist, steht als **eine**
// Zeile da; und die Wege einer Karte werden erst einzeln wählbar, wenn
// jemand es will — gelesen werden sie immer.
import { useId, useState } from 'react'

import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Entscheidung } from '../bausteine/Entscheidung'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Faltliste } from '../bausteine/Faltliste'
import { Knopf } from '../bausteine/Knopf'
import type { KontaktDTO } from '../apikontakte'
import {
  adressbuchvorgang,
  useAdressbuch,
  type Karteinwahl,
} from '../haken/useAdressbuch'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

const haken =
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Adressbuch() {
  const eigene = useTexte().seiten.adressbuch
  const buch = useAdressbuch()
  const meldung = buch.fehler ?? buch.ladefehler

  return (
    <Aufklappkasten titel={eigene.titel}>
      <div className="flex flex-col gap-3">
        <Erklaerungsabsatz text={eigene.erklaerung} />
        <Erklaerungsabsatz text={eigene.keinePruefung} />

        <div>
          <Knopf
            angehalten={buch.vorgang.laeuft(adressbuchvorgang)}
            aufKlick={buch.einlesen}
          >
            {eigene.waehlen}
          </Knopf>
        </div>

        {meldung !== undefined && (
          <p role="alert" className="text-sm text-fehler">
            {meldung}
          </p>
        )}

        {buch.meldung !== undefined && (
          <p role="status" className="text-sm">
            {buch.meldung}
          </p>
        )}

        {buch.befunde.map((befund) => (
          <p key={befund} className="text-sm text-hinweis">
            {befund}
          </p>
        ))}

        {buch.auswahl !== undefined && (
          <Vorschau auswahl={buch.auswahl} buch={buch} />
        )}

        <Kontaktliste kontakte={buch.kontakte} buch={buch} />
      </div>
    </Aufklappkasten>
  )
}

type Buch = ReturnType<typeof useAdressbuch>

// vorn und schritt sind, wie viele Karten und Kontakte zu Anfang und je
// Klick dazukommen: zehn Karten sind zehn Kästchen, und R3 erlaubt zwölf
// Felder; zwanzig Kontakte tragen zwanzig Knöpfe und füllen den Kasten.
const kartenvorn = 10
const kontaktevorn = 20

function Vorschau({
  auswahl,
  buch,
}: {
  auswahl: readonly Karteinwahl[]
  buch: Buch
}) {
  const eigene = useTexte().seiten.adressbuch
  const anzahl = auswahl.filter((wahl) => wahl.an).length
  const titel = useId()

  // **Die Stelle reist mit**: `karteWaehlen` und `wegWaehlen` zählen in der
  // ganzen Auswahl, und die Liste zeigt nur, was zu wählen ist.
  const stellen = auswahl.map((wahl, stelle) => ({ wahl, stelle }))
  const waehlbar = stellen.filter(
    ({ wahl }) => wahl.karte.abgleich !== 'gleich',
  )
  const gleich = stellen.filter(({ wahl }) => wahl.karte.abgleich === 'gleich')

  return (
    <section
      aria-labelledby={titel}
      className="flex flex-col gap-2 rounded-sm border border-kante p-3"
    >
      <h3 id={titel} className="text-sm font-medium">
        {eigene.vorschauTitel}
      </h3>

      <Faltliste
        eintraege={waehlbar}
        vorn={kartenvorn}
        schritt={kartenvorn}
        klasse="flex flex-col gap-2"
        zeile={({ wahl, stelle }) => (
          <Vorschauzeile
            key={`${wahl.karte.kontakt.ursprung}:${String(stelle)}`}
            wahl={wahl}
            stelle={stelle}
            buch={buch}
          />
        )}
      />

      {/* **Unverändert ist keine Wahl** — diese Karten stehen schon so da,
          und übernommen werden sie nicht. Ihre Namen stehen auf Wunsch. */}
      {gleich.length > 0 && (
        <Aufklappkasten
          titel={einsetzen(eigene.unveraendert, {
            anzahl: String(gleich.length),
          })}
        >
          <Faltliste
            eintraege={gleich}
            vorn={kontaktevorn}
            schritt={kontaktevorn}
            klasse="flex flex-col gap-1 text-sm"
            zeile={({ wahl, stelle }) => (
              <li key={`${wahl.karte.kontakt.ursprung}:${String(stelle)}`}>
                {wahl.karte.kontakt.name}
              </li>
            )}
          />
        </Aufklappkasten>
      )}

      <div className="flex flex-wrap gap-2">
        <Knopf
          gewicht="haupt"
          angehalten={anzahl === 0 || buch.vorgang.laeuft(adressbuchvorgang)}
          aufKlick={() => {
            buch.uebernehmen(auswahl)
          }}
        >
          {einsetzen(eigene.uebernehmen, { anzahl: String(anzahl) })}
        </Knopf>
        <Knopf aufKlick={buch.verwerfen}>{eigene.verwerfen}</Knopf>
      </div>
    </section>
  )
}

// Vorschauzeile ist eine Karte, die zu wählen ist: ihr Kästchen, wie sie zu
// den vorhandenen steht, und ihre Wege — **gelesen** immer, **einzeln
// gewählt** erst auf Wunsch (AK-01). Ein abgewählter Weg steht durchgestrichen
// da, auch wenn die Wahl wieder zu ist.
function Vorschauzeile({
  wahl,
  stelle,
  buch,
}: {
  wahl: Karteinwahl
  stelle: number
  buch: Buch
}) {
  const eigene = useTexte().seiten.adressbuch
  const [wege, setzeWege] = useState(false)
  const wegeID = useId()
  const name = wahl.karte.kontakt.name

  return (
    <li className="flex flex-col gap-1">
      <div className="flex flex-wrap items-center gap-2">
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={wahl.an}
            onChange={(ereignis) => {
              buch.karteWaehlen(stelle, ereignis.target.checked)
            }}
            className={haken}
          />
          <span className="font-medium">{name}</span>
          <span className="text-leise">
            {einsetzen(eigene.abgleich[wahl.karte.abgleich], {
              name: wahl.karte.vorhanden?.name ?? '',
            })}
          </span>
        </label>

        {wahl.karte.kontakt.wege.length > 0 && (
          <Knopf
            klein
            gewicht="leise"
            bezeichnung={`${eigene.wegeWaehlen}: ${name}`}
            ausgeklappt={wege}
            steuert={wegeID}
            aufKlick={() => {
              setzeWege(!wege)
            }}
          >
            {eigene.wegeWaehlen}
          </Knopf>
        )}
      </div>

      <ul id={wegeID} className="ml-6 flex flex-col">
        {wahl.karte.kontakt.wege.map((weg, nummer) => {
          const zeile = einsetzen(eigene.wegZeile, {
            art: eigene.weg[weg.art],
            wert: weg.wert,
          })

          return (
            <li key={`${weg.art}:${weg.wert}`} className="text-sm">
              {wege ? (
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={wahl.wege[nummer]}
                    disabled={!wahl.an}
                    onChange={(ereignis) => {
                      buch.wegWaehlen(stelle, nummer, ereignis.target.checked)
                    }}
                    className={haken}
                  />
                  {zeile}
                </label>
              ) : (
                <span
                  className={
                    wahl.wege[nummer] === true
                      ? 'text-leise'
                      : 'text-leise line-through'
                  }
                >
                  {zeile}
                </span>
              )}
            </li>
          )
        })}
      </ul>
    </li>
  )
}

function Kontaktliste({
  kontakte,
  buch,
}: {
  kontakte: readonly KontaktDTO[]
  buch: Buch
}) {
  const eigene = useTexte().seiten.adressbuch
  const [fragt, setzeFragt] = useState<KontaktDTO>()

  return (
    <section className="flex flex-col gap-2">
      <h3 className="text-sm font-medium">{eigene.listeTitel}</h3>

      {kontakte.length === 0 && !buch.laedt && (
        <p className="text-sm text-leise">{eigene.leer}</p>
      )}

      <Faltliste
        eintraege={kontakte}
        vorn={kontaktevorn}
        schritt={kontaktevorn}
        klasse="flex flex-col gap-2"
        zeile={(kontakt) => (
          <li key={kontakt.id} className="flex items-start gap-3">
            {kontakt.bild !== null && kontakt.bild !== '' && (
              <img
                src={`data:${kontakt.bildart};base64,${kontakt.bild}`}
                alt={einsetzen(eigene.bild, { name: kontakt.name })}
                className="h-10 w-10 rounded-full object-cover"
              />
            )}
            <div className="flex min-w-0 flex-1 flex-col text-sm">
              <span className="font-medium">{kontakt.name}</span>
              {kontakt.wege.map((weg) => (
                <span key={`${weg.art}:${weg.wert}`} className="text-leise">
                  {einsetzen(eigene.wegZeile, {
                    art: eigene.weg[weg.art],
                    wert: weg.wert,
                  })}
                </span>
              ))}
            </div>
            <Aktionsknopf
              aktion="entfernen"
              bezeichnung={einsetzen(eigene.entfernenKontakt, {
                name: kontakt.name,
              })}
              angehalten={buch.vorgang.laeuft(kontakt.id)}
              aufKlick={() => {
                setzeFragt(kontakt)
              }}
            >
              {eigene.entfernen}
            </Aktionsknopf>
          </li>
        )}
      />

      {fragt !== undefined && (
        <Entscheidung
          frage={einsetzen(eigene.entfernenFrage, { name: fragt.name })}
          erklaerung={eigene.entfernenErklaerung}
          bestaetigen={eigene.entfernen}
          abbrechen={eigene.abbrechen}
          aktion="entfernen"
          aufBestaetigen={() => {
            buch.loeschen(fragt)
            setzeFragt(undefined)
          }}
          aufAbbrechen={() => {
            setzeFragt(undefined)
          }}
        />
      )}
    </section>
  )
}
