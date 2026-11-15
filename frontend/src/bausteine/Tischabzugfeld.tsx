// **Die Fläche der Tischausgabe** — PLAY-05.AK-03 und AK-04
// (#1713).
//
// # Warum erst die Vorschau und dann das Sichern
//
// **Weil AK-03 sie verlangt** — „Screenshot, Bildexport und PDF besitzen
// Vorschau" — und weil es dasselbe Bild sein muss. Gesichert wird **genau**,
// was dasteht, und nicht ein zweiter Abzug, der zwischendurch anders aussähe:
// das ist AK-04, „aus einem konsistenten Snapshot, nicht aus halb laufender
// Animation".
//
// # Und warum der Abzug von außen kommt
//
// **Weil dieses Bauteil ohne Grafikkarte prüfbar bleiben soll** (§3). Wie ein
// Bild entsteht, steht in `tisch/tischabzug.ts` und darunter in `pixinaht.ts`;
// hier steht, wie man es anfordert und was danach dasteht.
import { useEffect, useId, useRef, useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'
import { api, auspacken } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zonenwort } from '../seiten/partiewoerter'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { Abzug, Abzugsfehler } from '../tisch/tischabzug'
import type { Ausschnittart, Ausschnittwahl } from '../tisch/tischausschnitt'
import type { Tischplan } from '../tisch/tischplan'
import { Abschnittstitel } from './Abschnittstitel'

// Blattformate sind die zwei, die zur Wahl stehen — dieselben Kennungen wie
// in `internal/domain/abzugsseite.go`.
//
// **Zwei Stellen, und das ist hier richtig so**: die Go-Seite entscheidet,
// was ein Blatt **ist**; diese Zeile entscheidet, was das Fenster
// **anbietet**. Ein unbekanntes Format weist der Dienst ab, statt still A4 zu
// nehmen — derselbe Satz wie bei der Commanderkennung in
// `spiel/partiezonenliste.ts`.
const Blattformate = ['a4', 'letter'] as const

// blattwort gibt den Namen eines Formats.
function blattwort(
  eigene: { abzugBlattA4: string; abzugBlattLetter: string },
  blatt: string,
): string {
  return blatt === 'a4' ? eigene.abzugBlattA4 : eigene.abzugBlattLetter
}

// Aufloesungen sind die drei, die zur Wahl stehen.
//
// **Dieselben wie am Druckbogen**: 150 für den Bildschirm, 300 für Papier,
// 600 für den Fall, dass jemand hineinzoomt. Die Zahlen sind gewählt und
// nicht gemessen.
const Aufloesungen = [150, 300, 600]

interface Abzugseigenschaften {
  plan: Tischplan

  // abzug macht das Bild — oder sagt, warum es keines gibt.
  abzug: (
    wahl: Ausschnittwahl,
    dpi: number,
  ) => Promise<Abzug | { fehler: Abzugsfehler }>

  // sicht ist der Sitz, für den der Tisch gerechnet ist — leer für die
  // Zuschauersicht. **Eine Vorschau gehört zu der Sicht, aus der sie stammt**
  // (PLAY-05.AK-02, Befund 1 des Zweitblicks über Phase 6.9): wer die Sicht
  // wechselt, darf die Hand der alten weder sehen noch sichern.
  sicht?: string
}

export function Tischabzugfeld({
  plan,
  abzug,
  sicht = '',
}: Abzugseigenschaften) {
  const eigene = useTexte().seiten.spielen

  const ausschnittID = useId()
  const zoneID = useId()
  const dpiID = useId()
  const blattID = useId()

  const [art, setzeArt] = useState<Ausschnittart>('sichtbar')
  const [feld, setzeFeld] = useState(0)
  const [dpi, setzeDpi] = useState(150)

  const [blatt, setzeBlatt] = useState<string>(Blattformate[0])
  const [ausgabe, setzeAusgabe] = useState<Ausgabe>({ fuer: sicht })

  // **Nur, was zur heutigen Sicht gehört** — eine ältere Vorschau steht weder
  // da, noch geht sie in eine Datei; ihr Befund und ihr Pfad ebenso wenig
  // (§7, Regel 1).
  const { bild, befund, pfad } = ausgabe.fuer === sicht ? ausgabe : {}

  // **Und eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7,
  // Regel 2): was nach einem Wechsel ankommt, wird verworfen.
  const heutigeSicht = useRef(sicht)
  useEffect(() => {
    heutigeSicht.current = sicht
  }, [sicht])

  // **Und nur die jüngste Anfrage zählt** — zwei Abzüge in derselben Sicht
  // konnten in umgekehrter Reihenfolge antworten (Befund 3 des Nachlaufs
  // über Phase 6.9).
  const juengste = useRef(0)

  const gewaehltesFeld = plan.felder[feld]

  const erzeuge = async () => {
    const fuer = sicht
    juengste.current += 1
    const nummer = juengste.current

    // **Ein neuer Abzug nimmt den gesicherten Ort weg**, Vorschau und Befund
    // bleiben bis zur Antwort stehen.
    setzeAusgabe((vorher) => ({ ...vorher, pfad: undefined }))

    const ergebnis = await abzug(wahlAus(art, gewaehltesFeld), dpi)

    if (fuer !== heutigeSicht.current || nummer !== juengste.current) {
      return
    }

    setzeAusgabe(
      'fehler' in ergebnis
        ? { fuer, befund: befundsatz(eigene, ergebnis.fehler) }
        : { fuer, bild: ergebnis },
    )
  }

  // **Gesichert wird das Bild der Vorschau** und nicht ein zweiter Abzug.
  const sichere = async (fertiges: Abzug, alsPDF: boolean) => {
    const antwort = alsPDF
      ? await api().TischabzugPDFSpeichern(fertiges.bild, blatt)
      : await api().TischabzugSpeichern(fertiges.bild)

    auspacken(antwort, () => {
      // **Ein Abbruch ist kein Erfolg und kein Fehler**: wer den Dialog
      // wegklickt, hat sich entschieden, und die Vorschau bleibt stehen.
      // **Der Ort gehört zu dem Bild, das gesichert wurde** — steht
      // inzwischen ein anderes da, aus einer anderen Sicht oder einem
      // neueren Abzug, bleibt er weg.
      if (!antwort.abgebrochen) {
        setzeAusgabe((vorher) =>
          vorher.bild === fertiges ? { ...vorher, pfad: antwort.pfad } : vorher,
        )
      }

      return undefined
    })
  }

  return (
    <div className="flex flex-col gap-3">
      <Abschnittstitel className="text-base font-semibold">
        {eigene.abzugTitel}
      </Abschnittstitel>

      <Erklaerungsabsatz
        text={eigene.abzugErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <div className="flex flex-wrap items-end gap-3">
        <label htmlFor={ausschnittID} className="flex flex-col gap-1">
          <span className="text-sm font-medium">{eigene.abzugAusschnitt}</span>
          <select
            id={ausschnittID}
            value={art}
            className={feldklassen}
            onChange={(ereignis) => {
              setzeArt(ereignis.target.value as Ausschnittart)
            }}
          >
            <option value="sichtbar">{eigene.abzugSichtbar}</option>
            <option value="tisch">{eigene.abzugGanzerTisch}</option>
            <option value="zone">{eigene.abzugEineZone}</option>
          </select>
        </label>

        {/* **Die Zonenwahl steht nur da, wo sie gemeint ist** — drei
            Bedienelemente für einen Ausschnitt, der keine Zone ist, wären
            Platz für nichts (#1771). */}
        {art === 'zone' && (
          <label htmlFor={zoneID} className="flex flex-col gap-1">
            <span className="text-sm font-medium">{eigene.abzugZonenwahl}</span>
            <select
              id={zoneID}
              value={feld}
              className={feldklassen}
              onChange={(ereignis) => {
                setzeFeld(Number(ereignis.target.value))
              }}
            >
              {plan.felder.map((eines, stelle) => (
                <option key={`${eines.sitz}/${eines.zone}`} value={stelle}>
                  {zonenfeldwort(eigene, eines)}
                </option>
              ))}
            </select>
          </label>
        )}

        <label htmlFor={dpiID} className="flex flex-col gap-1">
          <span className="text-sm font-medium">{eigene.abzugAufloesung}</span>
          <select
            id={dpiID}
            value={dpi}
            className={feldklassen}
            onChange={(ereignis) => {
              setzeDpi(Number(ereignis.target.value))
            }}
          >
            {Aufloesungen.map((eine) => (
              <option key={eine} value={eine}>
                {eine}
              </option>
            ))}
          </select>
        </label>

        <Knopf
          gewicht="leise"
          aufKlick={() => {
            void erzeuge()
          }}
        >
          {eigene.abzugErzeugen}
        </Knopf>

        {/* **Sichern gibt es erst mit einem Bild** — ein Knopf ohne Wirkung
            ist eine Zusage, die niemand einlöst (§10). */}
        {bild !== undefined && (
          <>
            <Aktionsknopf
              aktion="sichern"
              aufKlick={() => {
                void sichere(bild, false)
              }}
            >
              {eigene.abzugSichern}
            </Aktionsknopf>

            {/* **Das Seitenformat steht beim PDF-Knopf** — ohne Vorschau
                wäre es eine Wahl ohne Gegenstand (#1771). */}
            <label htmlFor={blattID} className="flex flex-col gap-1">
              <span className="text-sm font-medium">{eigene.abzugBlatt}</span>
              <select
                id={blattID}
                value={blatt}
                className={feldklassen}
                onChange={(ereignis) => {
                  setzeBlatt(ereignis.target.value)
                }}
              >
                {Blattformate.map((eines) => (
                  <option key={eines} value={eines}>
                    {blattwort(eigene, eines)}
                  </option>
                ))}
              </select>
            </label>

            <Knopf
              gewicht="leise"
              aufKlick={() => {
                void sichere(bild, true)
              }}
            >
              {eigene.abzugPDF}
            </Knopf>
          </>
        )}
      </div>

      {befund !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {befund}
        </p>
      )}

      {pfad !== undefined && (
        <p role="status" className="text-sm text-leise">
          {einsetzen(eigene.abzugGesichert, { pfad })}
        </p>
      )}

      {bild !== undefined && (
        <figure className="flex flex-col gap-1">
          <img
            src={bild.bild}
            alt={eigene.abzugVorschau}
            className="max-h-64 max-w-full rounded-sm border border-kante"
          />
          <figcaption className="text-xs text-leise">
            {einsetzen(eigene.abzugMasse, {
              breite: String(bild.breite),
              hoehe: String(bild.hoehe),
            })}
          </figcaption>
        </figure>
      )}
    </div>
  )
}

// Ausgabe ist, was ein Abzug hinterlässt — und für welche Sicht.
interface Ausgabe {
  fuer: string
  bild?: Abzug
  befund?: string
  pfad?: string | undefined
}

// wahlAus baut die Ausschnittwahl aus Art und gewähltem Feld.
//
// **Ohne Feld bleibt es der sichtbare Bereich** — ein Plan ohne Felder gibt
// es nicht, aber ein Zugriff, der davon ausginge, wäre eine Annahme.
function wahlAus(
  art: Ausschnittart,
  feld: { zone: string; sitz: string } | undefined,
): Ausschnittwahl {
  if (art !== 'zone' || feld === undefined) {
    return { art }
  }

  return { art, zone: feld.zone, sitz: feld.sitz }
}

// zonenfeldwort nennt ein Feld beim Namen seiner Zone und seines Sitzes.
//
// **Die Sitzkennung und nicht der Name**: die steht am Feld, und der Name
// hinge an der Partie, die dieses Bauteil nicht kennt. Der gemeinsame Stapel
// trägt keine (CR 405.1) und steht dann allein da.
function zonenfeldwort(
  eigene: Record<string, unknown>,
  feld: { zone: string; sitz: string },
): string {
  const wort = zonenwort(eigene, feld.zone)

  return feld.sitz === '' ? wort : `${wort} · ${feld.sitz}`
}

const feldklassen =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

// Abzugssaetze sind die fünf Befunde, je einer je Grund.
interface Abzugssaetze {
  abzugKeinMaler: string
  abzugKeinAusschnitt: string
  abzugDpiFehler: string
  abzugZuGross: string
  abzugLeer: string
  abzugZieht: string
  abzugAufbau: string
}

// befundsatz übersetzt einen Abzugsfehler.
//
// **Als Verbund und nicht über einen Indexzugriff**: der Katalog trägt seit
// PLAY-03.AK-04 auch Verbünde, und `Record<string, string>` wäre eine
// Behauptung über etwas, das wächst (derselbe Satz wie bei `wortAus` in
// `partiewoerter.ts`).
function befundsatz(eigene: Abzugssaetze, fehler: Abzugsfehler): string {
  const saetze: Record<Abzugsfehler, string> = {
    keinMaler: eigene.abzugKeinMaler,
    keinAusschnitt: eigene.abzugKeinAusschnitt,
    dpi: eigene.abzugDpiFehler,
    zuGross: eigene.abzugZuGross,
    leer: eigene.abzugLeer,
    zieht: eigene.abzugZieht,
    aufbau: eigene.abzugAufbau,
  }

  return saetze[fehler]
}
