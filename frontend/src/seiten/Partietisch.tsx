// **Der Tisch einer Partie** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was hier zu sehen ist und was nicht
//
// **Die Sicht eines Sitzes.** Eine Bibliothek steht als Zahl da, eine fremde
// Hand als Zahl; was geheim ist, trägt kein Bild und keinen Namen. Es gibt
// hier keinen Weg zum Vollzustand — auch nicht für den Wirt, und der Grund
// steht in `internal/app/partie.go`.
//
// # Warum der Wirt genannt wird
//
// **Weil er den Vollzustand hält.** Das Issue verlangt es wörtlich: „diese
// Vertrauensgrenze wird vor Beitritt offengelegt". Ein Hinweis, der erst
// nach dem Mitspielen käme, wäre keiner.
//
// # Wie es angeordnet ist
//
// **Wie im Artboard *Tisch*** — seit dem 2026-09-24, Befund F09 der nativen
// Prüfung von 0.5.0 (UX-03, UX-08): links die Leinwand, so hoch, wie das
// Fenster unter dem Kopf ist, und darunter ihre Steuerung; rechts eine
// Spalte mit dem, was man während des Spiels braucht — der Zug, die
// Spieler, eine offene Wahl, die gewählte Karte.
//
// **Was man nicht während jedes Zuges braucht, steht darunter** — Anhalten
// und Beenden, der Verlauf, die Freigabe des Tisches und des Protokolls und
// der Stand der Leitung.
import { useId, useMemo, useRef, useState } from 'react'
import type { CSSProperties } from 'react'

import { Partiezonen } from './Partiezonen'
import { tischbildquelle } from './tischbildquelle'
import { useEigeneLagen } from '../haken/useEigeneLagen'
import { zugBehandeln } from './tischzug'
import { useKartenbedienung } from '../haken/useKartenbedienung'
import { useTischeinstellungen } from '../haken/useTischeinstellungen'
import { useTischbild } from '../haken/useTischbild'
import { Tischflaeche } from '../bausteine/Tischflaeche'
import { Tischspalte } from '../bausteine/Tischspalte'
import { Handlungsanmeldung } from '../bausteine/Handlungsanmeldung'
import { useTischruecknahme } from '../haken/useTischruecknahme'
import { useVerdeckt } from '../verdeckung'
import { Vorhang } from '../bausteine/Vorhang'
import { useVorhang } from '../haken/useVorhang'

import { tischmaler } from '../tisch/pixinaht'
import { tischplan } from '../tisch/tischplan'
import { Partiefuehrung } from './Partiefuehrung'
import { Tischdiagnose } from '../bausteine/Tischdiagnose'
import { AbschnitteUnter } from '../bausteine/Abschnittstitel'
import { Partiekopf } from './Partiekopf'
import { Sichtwechsel } from './Sichtwechsel'
import { Partieausgang } from '../bausteine/Partieausgang'
import { useTischwirkungen } from '../haken/useTischwirkungen'
import { useProfilnamen } from './sitzwort'
import { sitzkopfFuer, sitzname } from './partiewoerter'
import { useSpieltexturen } from '../haken/useSpieltexturen'
import { useVorladen } from '../haken/useVorladen'
import { useTexte } from '../sprache/kontext'

import type { Partie } from '../apipartie'
import { useFensterrest } from '../haken/useFensterrest'

import type { Partielauf } from '../haken/usePartie'

export {
  Autoanordnungschluessel,
  Beschriftungsschluessel,
  Tischbildschluessel,
} from '../haken/useTischeinstellungen'

interface Tischeigenschaften {
  lauf: Partielauf
  partie: Partie

  // ebene ist die Stufe der Überschrift — eins, wo der Tisch die Seite ist
  // (`Spielen.tsx`), zwei, wo er in einer Seite steht (der Tisch eines
  // Gastes).
  ebene?: 1 | 2

  // fremd heißt: der Tisch eines Wirts, an dem dieses Fenster Gast ist —
  // ohne die Führung, die dem Wirt gehört (`Partiefuehrung`).
  fremd?: boolean
}

// Tischrand ist der Innenrand unter dem Inhalt — `pb-8` an `main` in
// `rahmen/Rahmen.tsx`. Die Leinwand reicht bis an ihn und nicht darüber.
const Tischrand = 32

// Tischmindesthoehe ist die Höhe, unter die die Leinwand nicht fällt: in
// einem flachen Fenster scrollt die Seite, statt den Tisch zu zerdrücken.
const Tischmindesthoehe = 416

export function Partietisch({
  lauf,
  partie,
  ebene = 2,
  fremd = false,
}: Tischeigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen
  const ueberschrift = useId()

  const [gewaehlt, setzeGewaehlt] = useState<string>()

  // **Eine Darstellung zur Zeit** — Werkstück G3 von Phase 5.5
  // (#1460).
  //
  // Der gezeichnete Tisch und die Zonenliste zeigen **dieselben** Karten mit
  // **denselben** Knöpfen. Beide zugleich hieße: jede Karte zweimal in der
  // Tabulatorreihenfolge, und wer mit der Tastatur bedient, ginge zweimal
  // durch dasselbe Spiel. Genau das meint #1460 mit „keine zweite
  // Karten-Fachimplementierung".
  //
  // **Ohne Grafikkarte bleibt trotzdem alles erreichbar**: die Tischfläche
  // setzt ihre Knöpfe dann sichtbar statt durchsichtig.
  const [alsListe, setzeAlsListe] = useState(false)
  // gastsitze meldet die Freigabe — wer Gast ist, zu dem folgt die Sicht nie.
  const [gastsitze, setzeGastsitze] = useState<readonly string[]>()

  // **Der Vorhang am gemeinsamen Gerät** — P7 (#2375), in `useVorhang`.
  // „Offene Hände“ gilt für **eine** Partie und ist eine ausdrückliche Wahl;
  // am fremden Tisch gibt es keinen Vorhang, dort sitzt man an einem Sitz.
  const [offeneHaende, setzeOffeneHaende] = useState<string>()
  const vorhang = useVorhang({
    partie,
    ansehen: lauf.ansehen,
    gastsitze: fremd ? undefined : gastsitze,
    offen: offeneHaende === partie.id,
  })
  const zu = vorhang.vorhang !== undefined

  // **⌘Z meint den Tisch, solange er dasteht** — P8 (#2376), Befund 9 des
  // Zweitblicks über Phase 7: auch beim Sichtwechsel und hinter dem Vorhang,
  // dann wirkungslos.
  //
  // **Und unter einer Schublade auch nicht** (Nachlauf, R5): der Tisch bleibt
  // dort stehen, verdeckt und `inert`.
  const verdeckt = useVerdeckt()
  const bedienbar = !lauf.wechselt && !zu && !verdeckt
  const kannZuruecknehmen = useTischruecknahme(lauf, bedienbar)

  // **Ein Klick löst die Primärhandlung aus, ein Doppelklick die zweite, ein
  // Rechtsklick öffnet das Menü** — P4 (#2372), aus der Auskunft des Kerns;
  // ⌘-Klick wählt. **Ein wartender Klick verfällt** beim Sichtwechsel und
  // hinter dem Vorhang (Befund 10).
  const bedienung = useKartenbedienung({
    partie,
    lauf,
    waehle: setzeGewaehlt,
    bedienbar,
  })

  // **Hinter dem Vorhang ist nichts gewählt** — eine gewählte Handkarte
  // nennte ihren Namen in der Leiste.
  if (zu && gewaehlt !== undefined) {
    setzeGewaehlt(undefined)
  }

  // zurueckZumKopf gibt den Fokus an die Überschrift, wenn der Vorhang aufgeht
  // — sein Knopf ist dann fort.
  const zurueckZumKopf = () => {
    document.getElementById(ueberschrift)?.focus()
  }

  const tisch = useTischeinstellungen()

  // **Und das eigene Bild, wenn es eines gibt** — es liegt neben der Wahl und
  // nicht in ihr: wer eines ablegt, hat trotzdem einen Satz gewählt, auf den
  // er zurückfällt.
  const eigenes = useTischbild()

  // **Die Texturen gehören dem Tisch** (G2,
  // #1459):
  // ein Speicher je Tisch, und er wird geräumt, wenn der Tisch verschwindet.
  // Sieben gleiche Karten melden sich siebenmal an — geholt wird einmal.
  const texturen = useSpieltexturen(fremd, partie.id)

  // **Der Rücken und das eigene Deck werden vorgeladen** (#1459: „Eigene
  // Decks, Rückseite … vorladen"). In der **niedrigen** Tischgüte: sie ist
  // die Stufe, die es genau dafür gibt, und sie steht da, bis die volle
  // ankommt.
  useVorladen(texturen, partie)

  // **Der Plan hängt an der Partie und sonst nichts.** Er wird bei jeder
  // Antwort neu gerechnet — das ist billig — und der Maler gleicht ab, statt
  // die Szene neu zu bauen.

  const eigeneLagen = useEigeneLagen(partie.id, tisch.ordnetSelbst)

  const plan = useMemo(
    () => tischplan(partie, { gewendet: [], eigene: eigeneLagen.lagen }),
    [partie, eigeneLagen.lagen],
  )
  const bild = useMemo(
    () => tischbildquelle(partie, texturen),
    [partie, texturen],
  )

  const objekt = partie.objekte.find((eines) => eines.kennung === gewaehlt)

  // **Die Profile des Menschen, der arbeitet** (G8, #1465) — sie trennen zwei
  // Sitze, die denselben Namen tragen.
  // **Die Profilnamen werden neu geholt, sobald ein Befund dasteht** (G8,
  // #1465): einer von ihnen sagt, dass eine Bindung ins Leere zeigt — und
  // dann darf die Sichtwahl den Namen nicht mehr führen. Warum der Anlass
  // die Befunde sind und nicht die Fassung, steht bei [useEigeneProfile].
  const profilnamen = useProfilnamen(lauf.befunde)

  // **Was seit dem letzten Bild geschehen ist** — PLAY-04.AK-02
  // (#1712).
  // Aus dem Vergleich zweier Stände und nicht aus dem Ereignisstrom; der
  // Grund steht bei [useTischwirkungen].
  const wirkungen = useTischwirkungen(partie)

  const tischort = useRef<HTMLDivElement>(null)
  const hoehe = useFensterrest(tischort, {
    unten: Tischrand,
    mindestens: Tischmindesthoehe,
  })

  // **Die gewählte Karte steht in der Menüleiste** (#2482) — dieselbe Liste
  // wie Rechtsklick und ⇧F10, nur solange der Tisch bedienbar ist.
  const objektmenue =
    bedienbar && gewaehlt !== undefined
      ? bedienung.menueFuer(gewaehlt)
      : undefined

  // **Die Abschnitte eine Stufe unter dem Kopf** (#1992): h2 unter der h1
  // der eigenen Seite, h3 unter der h2 beim Gast.
  return (
    <AbschnitteUnter ueber={ebene}>
      {objektmenue !== undefined && gewaehlt !== undefined && (
        <Handlungsanmeldung
          titel={texte.seiten.karten.kartenmenue}
          art="spielobjekt"
          kennung={gewaehlt}
          eintraege={objektmenue.eintraege}
        />
      )}
      <section
        aria-labelledby={ueberschrift}
        className="@container flex flex-col gap-4"
      >
        <Partiekopf
          partie={partie}
          lauf={lauf}
          fremd={fremd}
          ueberschrift={ueberschrift}
          profilnamen={profilnamen}
          setzeGewaehlt={setzeGewaehlt}
          alsListe={alsListe}
          umschalten={() => {
            setzeAlsListe((bisher) => !bisher)
          }}
          offeneHaende={offeneHaende === partie.id}
          setzeOffeneHaende={(offen) => {
            setzeOffeneHaende(offen ? partie.id : undefined)
          }}
          ebene={ebene}
        />

        {/* **Das Ergebnis steht ganz oben** (M17, #1618) — es ist die
          Antwort auf die einzige Frage, die dann noch zählt. Solange die
          Partie läuft, zeichnet es nichts. */}
        <Partieausgang partie={partie} />

        {/* **Zwei Spalten, sobald der Inhalt breit genug ist** — gemessen am
          Abschnitt und nicht am Fenster (`@container`): eine offene
          Seitenleiste nimmt dem Inhalt 15rem, gleich wie breit das Fenster
          ist. Darunter steht die Spalte unter dem Tisch, und **der Tisch
          lässt ihr 13rem**, damit der Zug im Blick bleibt — F09 verlangt
          Tisch **und** aktuellen Zug sichtbar. **13 und nicht 11** (#2526):
          der Zug ist rund 162 Punkte hoch und steht zwei Abstände unter dem
          Tisch, zusammen 194; mit 11rem lag er bei 1024 × 700 unter der
          Falz. **Und mindestens 17rem statt 20**: die Leiste darunter ist
          dort eine Zeile, und die Leinwand bleibt bei 222 statt 234
          Punkten (`play06-tisch`).

          **Die Höhe als Merkmal und nicht als Stil der Spalten**: die
          rechte Spalte nimmt sie nur, wo sie neben dem Tisch steht; darunter
          wächst sie mit ihrem Inhalt, statt ein zweites Mal zu scrollen. */}
        <div
          className="flex flex-col gap-4 @4xl:flex-row @4xl:items-start"
          style={{ '--tischhoehe': `${String(hoehe)}px` } as CSSProperties}
        >
          <div
            ref={tischort}
            // **`flex-1` nur nebeneinander**: untereinander ist die Höhe die
            // Hauptachse, und eine Basis von null ginge dort vor `height` —
            // die Leinwand nähme ihre Inhaltshöhe und schöbe den Zug aus dem
            // Blick (gemessen in WebKit bei 1024 × 700: 734 statt 321 Punkte).
            className="flex h-[max(17rem,calc(var(--tischhoehe)-13rem))] min-w-0 flex-col @4xl:h-(--tischhoehe) @4xl:flex-1"
          >
            {vorhang.vorhang !== undefined ? (
              <Vorhang
                wer={sitzname(partie, vorhang.vorhang)}
                bisher={sitzname(partie, partie.fuer)}
                bestaetigen={() => {
                  vorhang.bestaetigen()
                  zurueckZumKopf()
                }}
                bleiben={() => {
                  vorhang.bleiben()
                  zurueckZumKopf()
                }}
              />
            ) : lauf.wechselt ? (
              <Sichtwechsel fehler={lauf.fehler} />
            ) : alsListe ? (
              <div className="min-h-0 flex-1 overflow-y-auto">
                <Partiezonen
                  partie={partie}
                  texturen={texturen}
                  gewaehlt={gewaehlt}
                  waehle={setzeGewaehlt}
                  wirkungen={wirkungen.jeDing}
                  erlaubt={lauf.erlaubt}
                  befehlen={lauf.befehlen}
                  handlungen={lauf.handlungen}
                  greife={bedienung.greife}
                  tastenFuer={bedienung.tastenFuer}
                  taste={bedienung.taste}
                />
              </div>
            ) : (
              <Tischflaeche
                plan={plan}
                bild={bild}
                malerei={tischmaler}
                aufKarte={bedienung.greife}
                menueFuer={bedienung.menueFuer}
                tastenFuer={bedienung.tastenFuer}
                taste={bedienung.taste}
                tischbild={tisch.bildname}
                aufTischbild={tisch.setzeBild}
                sitzkopf={sitzkopfFuer(partie, eigene, profilnamen)}
                beschriftungsstufe={tisch.beschriftungsstufe}
                wirkungen={wirkungen.jeDing}
                hintergrund={eigenes.adresse}
                eigenesBild={eigenes.waehlen}
                eigenesBildWeg={eigenes.entfernen}
                ablegen={!tisch.ordnetSelbst}
                sicht={partie.fuer}
                aufZug={(ziel) => {
                  zugBehandeln(ziel, {
                    lege: eigeneLagen.lege,
                    befehle: lauf.befehlen,
                    waehle: setzeGewaehlt,
                  })
                }}
                {...(gewaehlt === undefined ? {} : { gewaehlt })}
              />
            )}
          </div>

          {/* **Die Spalte gehört zur Sicht** und geht beim Wechsel und
            hinter dem Vorhang mit (`Sichtwechsel.tsx`, `Vorhang.tsx`). */}
          {!lauf.wechselt && !zu && (
            <Tischspalte
              lauf={lauf}
              kannZuruecknehmen={kannZuruecknehmen}
              partie={partie}
              objekt={objekt}
              texturen={texturen}
              entwaehle={() => {
                setzeGewaehlt(undefined)
              }}
            />
          )}
        </div>

        {/* **Am fremden Tisch nicht** — gefunden mit zwei Fenstern
          (`nat04`, S12.NAT-04): Anhalten, Freigabe, Rückschau und Protokoll
          gehören dem Wirt und fragten den eigenen Bestand nach einer Partie,
          die dort nicht liegt. */}
        {/* **Hinter dem Vorhang ruht auch die Führung** — die Rückschau
          trägt, was nur der gezeigte Sitz sehen darf. Verborgen und nicht
          abgebaut: die Freigabe darin meldet, wer Gast ist. */}
        {!fremd && (
          <div hidden={zu}>
            <Partiefuehrung
              lauf={lauf}
              partie={partie}
              meldeGastsitze={setzeGastsitze}
            />
          </div>
        )}

        {/* **Und wie die Leitung steht** — einmal je Fenster und nicht je
          Sitz: es gibt eine Leitung, gleich wessen Sicht gerade gerechnet
          wird. */}
        <Tischdiagnose />
      </section>
    </AbschnitteUnter>
  )
}
