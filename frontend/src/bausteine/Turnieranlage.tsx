// **Die ersten zwei Schritte des Turnierassistenten** — Werkstück 5 von
// Phase 4.7 (#1294),
// zum Assistenten geworden mit DOM-04.AK-01
// (#1699):
// *Turnier* — Vorlage, Name, Spiel, Format, Notiz — und *Regeln*.
//
// **Beide Schritte in einem Bauteil**, damit *Zurück* nichts verliert: der
// Stand der Felder lebt, solange der Assistent vor dem Anlegen steht.
import { useId, useState } from 'react'

import { useSpielnamen } from '../haken/useSpielnamen'
import { Vorschlagsfeld } from './Vorschlagsfeld'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { Aktionsknopf } from './Aktionsknopf'
import { Auswahlfeld } from './Auswahlfeld'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Feld } from './Feld'
import { Knopf } from './Knopf'
import { useEinstellungen } from '../haken/useEinstellungen'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import type { FormatDTO } from '../apidecks'
import type { Turnierstand } from '../haken/useTurniere'
import { Pflichtfeld } from './Pflichtfeld'
import { feldbefund } from '../meldung'

const feldklassen =
  'rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

// spalte begrenzt ein Feld samt Hinweis auf eine lesbare Breite.
const spalte = 'w-full max-w-sm'

const kastenklassen =
  'size-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

// Befunde sind die Sätze der Anwendung an Name und Spiel.
interface Befunde {
  name: string | undefined
  spiel: string | undefined
}

const ohneBefunde: Befunde = { name: undefined, spiel: undefined }

interface AnlageEigenschaften {
  stand: Turnierstand

  // schritt ist 0 (*Turnier*) oder 1 (*Regeln*).
  schritt: number
  setzeSchritt: (schritt: number) => void

  // aufAngelegt bekommt die Kennung des neuen Turniers.
  aufAngelegt: (turnierID: string) => Promise<void>
}

// Turnieranlage holt die Vorgabe und baut das Formular darüber.
//
// **Zwei Bauteile aus demselben Grund wie beim Sammlungsformular**: die
// Vorgabe kommt aus dem Bestand und damit später als der erste Zeichenvorgang.
// Der `key` lässt das innere neu entstehen, sobald sie dasteht (§7).
export function Turnieranlage(eigenschaften: AnlageEigenschaften) {
  // **Womit ein neues Turnier anfängt** — Werkstück 9 von Phase 4.12
  // (#1351),
  // Punkt 2. Ohne Wahl steht der Schalter aus, wie er es immer tat.
  const vorgabe = useEinstellungen().wert('vorgabe.einzelspiel') === 'an'

  return <Anlage key={String(vorgabe)} {...eigenschaften} vorgabe={vorgabe} />
}

function Anlage({
  stand,
  schritt,
  setzeSchritt,
  aufAngelegt,
  vorgabe,
}: AnlageEigenschaften & { vorgabe: boolean }) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  const einzelspielhinweis = useId()

  const [vorlageID, setzeVorlageID] = useState('')
  const [mitTeilnehmenden, setzeMitTeilnehmenden] = useState(true)
  const [name, setzeName] = useState('')
  const [spiel, setzeSpiel] = useState('magic')
  const spielnamen = useSpielnamen()
  const [formatID, setzeFormatID] = useState('')
  const [notiz, setzeNotiz] = useState('')
  const [einzelspiel, setzeEinzelspiel] = useState(vorgabe)
  const [befunde, setzeBefunde] = useState<Befunde>(ohneBefunde)

  const formate = useLaden<FormatDTO[]>(async () => {
    const antwort = await api().Formate('')

    return auspacken(antwort, () => antwort.formate)
  }, [])

  // **Aus einer Vorlage gelten ihre Einstellungen** — die Kopie übernimmt
  // sie (`service.Turnierleitung.Kopieren`), und ein Feld, das anderes
  // zeigte, versprach etwas, das nicht geschieht.
  const vorlage = stand.turniere.find((turnier) => turnier.id === vorlageID)
  const wirksam = vorlage ?? { spiel, formatID, einzelspiel, notiz }
  const passende = (formate.daten ?? []).filter(
    (format) => format.spiel === wirksam.spiel,
  )

  // **Ein Format eines anderen Spiels geht nicht mit** — wer das Spiel nach
  // der Formatwahl ändert, legt sonst ein Turnier an, das Go abweist.
  const gewaehlt = passende.some((format) => format.id === wirksam.formatID)
    ? wirksam.formatID
    : ''

  const anlegen = async () => {
    setzeBefunde(ohneBefunde)

    let angelegt: string | undefined

    await stand.fuehre(
      'turnier:neu',
      async () => {
        const antwort =
          vorlage === undefined
            ? await api().TurnierAnlegen(
                name,
                spiel,
                notiz,
                einzelspiel,
                gewaehlt,
              )
            : await api().TurnierKopieren(vorlage.id, name, mitTeilnehmenden)

        angelegt = auspacken(antwort, () => antwort.turnier)?.id
      },
      // **Ein Befund am Namen oder Spiel führt dorthin zurück** — ENG-05.AK-03
      // (#1730): er steht am Feld, und das Feld steht im ersten Schritt.
      (ursache) => {
        const neu = {
          name: feldbefund(ursache, 'name', texte),
          spiel: feldbefund(ursache, 'spiel', texte),
        }

        if (neu.name === undefined && neu.spiel === undefined) {
          return false
        }

        setzeBefunde(neu)
        setzeSchritt(0)

        return true
      },
    )

    if (angelegt !== undefined) {
      await aufAngelegt(angelegt)
    }
  }

  if (schritt === 1) {
    return (
      <div className="flex flex-col gap-3">
        {vorlage !== undefined && <p>{eigene.regelnAusVorlage}</p>}

        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={wirksam.einzelspiel}
            disabled={vorlage !== undefined}
            aria-describedby={einzelspielhinweis}
            onChange={(ereignis) => {
              setzeEinzelspiel(ereignis.target.checked)
            }}
            className={kastenklassen}
          />

          <span className="text-sm">{eigene.einzelspiel}</span>
        </label>

        <p id={einzelspielhinweis} className="text-sm text-leise">
          {eigene.einzelspielHinweis}
        </p>
        <Erklaerungsabsatz
          text={eigene.schweizerSystem}
          klasse="max-w-3xl text-sm text-leise"
        />

        <div className="flex flex-wrap gap-2">
          <Knopf
            gewicht="leise"
            aufKlick={() => {
              setzeSchritt(0)
            }}
          >
            {eigene.zurueck}
          </Knopf>
          {/* **Anlegen führt** (R1) — seit der Assistent in einer Schublade
              steht (UXR-12, #1920), ist sie die Region dieses Knopfs. */}
          <Aktionsknopf
            aktion="anlegen"
            gewicht="haupt"
            angehalten={stand.vorgang.laeuft('turnier:neu')}
            aufKlick={() => {
              void anlegen()
            }}
          >
            {eigene.anlegen}
          </Aktionsknopf>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-start gap-3">
        {/* **Eine begrenzte Spalte je Feld** — der Hinweis darunter machte
            sie sonst so breit wie sein Satz (Sichtprüfung, Hüllenprobe
            `dom04ak01`). */}
        <div className={spalte}>
          <Feld beschriftung={eigene.vorlage} hinweis={eigene.vorlageHinweis}>
            {(verbindung) => (
              <Auswahlfeld
                id={verbindung.id}
                bezeichnung={eigene.vorlage}
                eintraege={[
                  { wert: '', name: eigene.ohneVorlage },
                  ...stand.turniere.map((turnier) => ({
                    wert: turnier.id,
                    name: turnier.name,
                  })),
                ]}
                wert={vorlageID}
                aendern={setzeVorlageID}
                beschriebenVon={verbindung['aria-describedby']}
                ungueltig={verbindung['aria-invalid']}
                klasse={feldklassen}
              />
            )}
          </Feld>
        </div>

        {vorlage !== undefined && (
          <label className="flex items-center gap-2 pt-7">
            <input
              type="checkbox"
              checked={mitTeilnehmenden}
              onChange={(ereignis) => {
                setzeMitTeilnehmenden(ereignis.target.checked)
              }}
              className={kastenklassen}
            />
            <span className="text-sm">{eigene.mitTeilnehmenden}</span>
          </label>
        )}
      </div>

      <div className="flex flex-wrap items-start gap-3">
        {/* **Der Name ist Pflicht, und das Feld sagt, warum** — DOM-04.AK-01. */}
        <div className={spalte}>
          <Pflichtfeld
            beschriftung={eigene.name}
            wert={name}
            befund={befunde.name}
            hinweis={eigene.nameWarum}
            geaendert={(wert) => {
              setzeName(wert)
              setzeBefunde((bisher) => ({ ...bisher, name: undefined }))
            }}
          />
        </div>

        {/* **Vorschläge aus den installierten Spielen, Freitext bleibt** —
            UX-06.AK-01 (#1678), der Grund steht in `useSpielnamen`. */}
        <div className={spalte}>
          <Vorschlagsfeld
            beschriftung={eigene.spiel}
            wert={wirksam.spiel}
            setzeWert={(wert) => {
              setzeSpiel(wert)
              setzeBefunde((bisher) => ({ ...bisher, spiel: undefined }))
            }}
            vorschlaege={spielnamen}
            befund={befunde.spiel}
            hinweis={eigene.spielWarum}
            angehalten={vorlage !== undefined}
            breite={12}
          />
        </div>

        <div className={spalte}>
          <Feld
            beschriftung={eigene.format}
            hinweis={eigene.formatWarum}
            {...(formate.fehler === undefined
              ? {}
              : { befund: formate.fehler })}
          >
            {(verbindung) => (
              <Auswahlfeld
                id={verbindung.id}
                bezeichnung={eigene.format}
                eintraege={[
                  { wert: '', name: eigene.ohneFormat },
                  ...passende.map((format) => ({
                    wert: format.id,
                    name: format.name,
                  })),
                ]}
                wert={gewaehlt}
                angehalten={vorlage !== undefined}
                aendern={setzeFormatID}
                beschriebenVon={verbindung['aria-describedby']}
                ungueltig={verbindung['aria-invalid']}
                klasse={feldklassen}
              />
            )}
          </Feld>
        </div>

        <label className="flex flex-col gap-1">
          <span className="text-sm">{eigene.notiz}</span>

          <input
            value={wirksam.notiz}
            disabled={vorlage !== undefined}
            onChange={(ereignis) => {
              setzeNotiz(ereignis.target.value)
            }}
            className={feldklassen}
          />
        </label>
      </div>

      <div>
        <Knopf
          gewicht="haupt"
          aufKlick={() => {
            setzeSchritt(1)
          }}
        >
          {eigene.weiter}
        </Knopf>
      </div>
    </div>
  )
}
