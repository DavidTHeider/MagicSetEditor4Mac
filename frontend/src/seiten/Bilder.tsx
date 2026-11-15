// Die Bildausgabe eines Sets.
//
// **Der Stapellauf läuft hier, nicht im Hintergrund.** Das ist keine
// Bequemlichkeit, sondern die einzige Bauart, die #49s zweite Zusage einlöst —
// „schreitet sichtbar fort und lässt sich abbrechen": ein Lauf in Go müsste dem
// Fenster sagen können, wo er steht, und diese Naht gibt es nicht (#162).
//
// Was hier **nicht** entsteht, sind die Dateinamen. Eindeutigkeit ist eine
// Aussage über den ganzen Lauf, und sie gehört dorthin, wo die Karten liegen.
import { useCallback, useRef, useState } from 'react'

import { api, auspacken, type BildnameDTO } from '../api'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { useEinstellungen } from '../haken/useEinstellungen'
import { Ausgabeauswahl } from '../bausteine/Ausgabeauswahl'
import { Fortschritt } from '../bausteine/Fortschritt'
import { Hinweiskasten } from '../bausteine/Hinweiskasten'
import { Knopf } from '../bausteine/Knopf'
import { Ausgabefuss } from '../bausteine/Ausgabefuss'
import { Bildereinstellungen } from './Bildereinstellungen'
import { Dateiverweis } from '../bausteine/Dateiverweis'

interface BilderEigenschaften {
  // setID ist **gesetzt** — der Arbeitsplatz rendert diese Lasche nur mit
  // offenem Set und zeigt sonst seinen eigenen leeren Zustand
  // (#288).
  // Ein Wächter „kein Set gewählt" wäre hier ein Zweig, den kein Aufruf
  // betreten kann (§3).
  setID: string
}

// Lauf ist der Stand eines Stapellaufs.
interface Lauf {
  geschrieben: number
  gesamt: number

  // uebersprungen sind die Karten, für die nichts geschrieben wurde —
  // `altbehalten` und eine Datei, die schon dalag (#491).
  uebersprungen: string[]
  ordner: string
  befunde: string[]
  abgebrochen: boolean
  fertig: boolean
}

export function Bilder({ setID }: BilderEigenschaften) {
  // **Die Wahl beim Namenskonflikt liegt im Bestand** (#491, #245): sie
  // überlebt den Neustart, und MSE hält sie ebenso — nur je Spiel statt je
  // Anwendung. Das ist eine benannte Abweichung: dieses Programm führt keine
  // Einstellungen je Spiel, und eine Tabelle dafür zu bauen kostet mehr, als
  // die Unterscheidung wert ist.
  const einstellungen = useEinstellungen()

  const konfliktwahl = einstellungen
    .imBereich('bildausfuhr')
    .find((eintrag) => eintrag.schluessel === 'bildkonflikt')

  // **Ohne Antwort gilt MSEs Vorgabe.** Der Lauf soll nicht daran hängen, ob
  // der Bestand gerade lesbar ist; was dann gilt, ist dasselbe wie auf einem
  // frischen Rechner.
  const konflikt = konfliktwahl?.wert ?? 'nummerierenneu'

  const texte = useTexte()
  const vorgang = useVorgang()

  // **Der zuletzt gewählte Ordner steht vorbelegt da** — AP-10 (#2367),
  // **ported** aus MSE2: `settings.default_export_dir`, nach jeder Wahl
  // fortgeschrieben (`src/gui/images_export_window.cpp`, Zeilen 63–66,
  // Fassung f9d9356 vom 2021-07-06), **einer für alle Sets** wie dort.
  //
  // **Abgeleitet und nicht abgeschrieben**: solange hier niemand wählt, gilt
  // der gemerkte — auch wenn die Einstellungen erst nach dem ersten Bild
  // ankommen. Leer heißt: noch keiner, und das Ausgeben bleibt angehalten.
  const gemerkt = einstellungen.wert('bildausfuhr.ordner')
  const [gewaehlt, setzeGewaehlt] = useState<string>()
  const ordner = gewaehlt ?? gemerkt
  const [vorlage, setzeVorlage] = useState('{nummer}-{name}')
  const [format, setzeFormat] = useState('png')
  const [breite, setzeBreite] = useState(750)

  // **Die Vorgabe ist die Parität zu MSE**, nicht die schnellste Güte: wer
  // Bilder schreibt, will in der Regel die, die MSE auch zeichnet.
  const [guete, setzeGuete] = useState('medium')

  // **Welche Karten hinausgehen** — Befund 2 des Zweitblicks über Phase 3.6.
  // Leer heißt: das ganze Set.
  const [auswahl, setzeAuswahl] = useState<readonly string[]>([])

  const [lauf, setzeLauf] = useState<Lauf>()
  const [fehler, setzeFehler] = useState<string>()

  // **Der Abbruch steht in einer Referenz, nicht im Zustand.** Die Schleife
  // liest ihn zwischen zwei Bildern; ein Zustandswert wäre in ihr für immer der,
  // mit dem sie angefangen hat.
  const anhalten = useRef(false)

  const waehleOrdner = useCallback(async () => {
    setzeFehler(undefined)

    try {
      const antwort = await api().Bildordner()

      auspacken(antwort, () => undefined)

      if (!antwort.abgebrochen) {
        setzeGewaehlt(antwort.ordner)
        einstellungen.setze('bildausfuhr.ordner', antwort.ordner)
      }
    } catch (ursache: unknown) {
      setzeFehler(meldungAus(ursache, texte))
    }

    // **`texte` gehört dazu** (#1170):
    // die Meldung kommt aus dem Katalog, und wer die Sprache wechselt, bekäme
    // sonst die vorige. Kein sichtbarer Fehler heute — der Katalog wechselt
    // selten —, aber eine Zusage, die nicht gilt.
  }, [texte, einstellungen])

  // **Ein Vorgang je Set** (§7, Regel 4): zwei Läufe über dasselbe Set
  // schrieben dieselben Dateien, und wer zuletzt fertig würde, entschiede.
  const ausgeben = async (ziel: string, wessen: string) => {
    await vorgang.starte(`bilder-${wessen}`, async () => {
      anhalten.current = false
      setzeFehler(undefined)

      try {
        const liste = await api().Bildnamen(wessen, vorlage, format, konflikt, [
          ...auswahl,
        ])
        const namen = auspacken(liste, () => liste.namen)

        const stand: Lauf = {
          geschrieben: 0,
          gesamt: namen.length,
          ordner: ziel,
          befunde: [],
          uebersprungen: [],
          abgebrochen: false,
          fertig: namen.length === 0,
        }

        setzeLauf(stand)

        await schreibe(namen, ziel, stand)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  // schreibe geht die Namen der Reihe nach durch.
  //
  // **Der Reihe nach und nicht nebenläufig**: jedes Bild kostet Arbeitsspeicher
  // in der Größe des Bildes, und vier gleichzeitig in Druckauflösung sind ein
  // Vielfaches dessen, was eines kostet.
  const schreibe = async (namen: BildnameDTO[], ziel: string, stand: Lauf) => {
    const befunde: string[] = []
    const uebersprungen: string[] = []

    for (const [stelle, name] of namen.entries()) {
      if (anhalten.current) {
        setzeLauf({
          ...stand,
          uebersprungen: [...uebersprungen],
          geschrieben: stelle,
          befunde,
          abgebrochen: true,
          fertig: true,
        })

        return
      }

      const antwort = await api().BildSchreiben(
        name.karteID,
        ziel,
        name.datei,
        format,
        breite,
        konflikt,
        guete,
      )

      // **Ein Fehler hält den Lauf an** (§8): wer weiterschriebe, sammelte
      // dreihundert Meldungen derselben Ursache — die Platte ist voll, das
      // Verzeichnis ist fort —, und niemand liest sie.
      auspacken(antwort, () => undefined)

      // **Übersprungen wird gezählt und benannt** (#491). Wer `altbehalten`
      // wählt und danach „312 Bilder geschrieben" liest, hat keine Auskunft,
      // sondern eine Zahl, die stimmt und nichts sagt.
      if (antwort.uebersprungen) {
        uebersprungen.push(name.datei)
      }

      befunde.push(...antwort.befunde)

      setzeLauf({
        ...stand,
        uebersprungen: [...uebersprungen],
        geschrieben: stelle + 1,
        befunde: [...befunde],
        fertig: stelle + 1 === namen.length,
      })
    }
  }

  const laeuft = vorgang.laeuft(`bilder-${setID}`)

  // **Die Sätze stehen oberhalb des JSX.** `keintextimjsx` weist eine
  // Vorlagenzeichenkette mit Platzhalter im Ausdruck ab — sie ist von einem
  // Oberflächentext nicht zu unterscheiden. Der Wortlaut ist unverändert.
  const fehlersatz =
    fehler === undefined
      ? undefined
      : `${texte.bausteine.zustaende.fehler} ${fehler}`

  // **Ein Merken, das scheitert, steht am Ordner** (§6) — der Lauf nimmt ihn
  // trotzdem, aber beim nächsten Mal stünde er nicht da.
  const fehlschlag = einstellungen.fehlschlag
  const merkfehler =
    fehlschlag?.schluessel === 'bildausfuhr.ordner'
      ? `${texte.seiten.bilder.nichtGemerkt} ${meldungAus(fehlschlag.ursache, texte)}`
      : undefined

  const laufsatz =
    lauf === undefined
      ? ''
      : lauf.abgebrochen
        ? `${texte.seiten.bilder.abgebrochen} ${String(lauf.geschrieben)} ${texte.seiten.bilder.laeuft} ${String(lauf.gesamt)}`
        : lauf.fertig
          ? `${String(lauf.geschrieben)} ${texte.seiten.bilder.fertig} ${lauf.ordner}`
          : `${String(lauf.geschrieben)} ${texte.seiten.bilder.laeuft} ${String(lauf.gesamt)}`

  return (
    <>
      {/* **Überschrift und Erklärung trägt die Fläche des Assistenten**
          (UXR-09, #1917) — `Ausgabeflaeche`, unter dem Namen, den die Wahl
          „Wohin?" daneben trägt. */}
      <Bildereinstellungen
        ordner={ordner}
        gemerkt={gewaehlt === undefined && gemerkt !== ''}
        {...(merkfehler === undefined ? {} : { merkfehler })}
        vorlage={vorlage}
        format={format}
        guete={guete}
        breite={breite}
        laeuft={laeuft}
        konflikt={
          konfliktwahl === undefined
            ? undefined
            : { wert: konfliktwahl.wert, werte: konfliktwahl.werte }
        }
        setzeVorlage={setzeVorlage}
        setzeFormat={setzeFormat}
        setzeGuete={setzeGuete}
        setzeBreite={setzeBreite}
        setzeKonflikt={(wert) => {
          einstellungen.setze('bildkonflikt', wert)
        }}
        waehleOrdner={() => {
          void waehleOrdner()
        }}
      />

      {/* **Welche Karten hinausgehen** — Befund 2 des Zweitblicks über
          Phase 3.6. MSE2 stellt die Wahl vor der Bildausgabe genauso
          (`SetWindow::onFileExportImages`, `src/gui/set/window.cpp`, Zeilen
          668–673, Fassung f9d9356 vom 2021-07-06). */}
      <div className="mt-4">
        <Ausgabeauswahl
          setID={setID}
          gewaehlt={auswahl}
          setzeGewaehlt={setzeAuswahl}
          angehalten={laeuft}
        />
      </div>

      {/* **Der Knopf, der schreibt, führt** (UXR-09, R1 aus
          `docs/UIREVIEW.md`): er ist die Handlung, derentwegen jemand hier
          ist, und „Ordner wählen …" darüber bereitet sie nur vor. Er steht in
          der Fußleiste, die beim Rollen stehen bleibt (AP-10, B-20). */}
      <Ausgabefuss>
        <Knopf
          gewicht="haupt"
          angehalten={ordner === '' || laeuft}
          aufKlick={() => {
            void ausgeben(ordner, setID)
          }}
        >
          {texte.seiten.bilder.ausgeben}
        </Knopf>
      </Ausgabefuss>

      {fehlersatz !== undefined && (
        <div className="mt-3">
          <Hinweiskasten art="fehler" meldung={fehlersatz} />
        </div>
      )}

      {lauf && (
        <div className="mt-4">
          {/* **Der Lauf steht in `Fortschritt`** — Werkstück U3 von Phase 5.1
              (#1414).
              Ein `role="status"` neben einem Balken mit `role="presentation"`
              sagte den Satz an, und der Balken wäre für ein Vorleseprogramm
              nicht da.

              **Der Anhalteknopf steht an dem Lauf, den er anhält.** Das
              kostet die Sekunden, in denen die Namensliste noch geholt wird —
              dort gibt es noch keinen Lauf und damit keinen Knopf. */}
          <Fortschritt
            satz={laufsatz}
            erledigt={lauf.geschrieben}
            gesamt={lauf.gesamt}
            anhalten={{
              beschriftung: texte.seiten.bilder.abbrechen,
              angehalten: !laeuft,
              aufKlick: () => {
                anhalten.current = true
              },
            }}
          />

          {lauf.fertig && !lauf.abgebrochen && (
            <Dateiverweis pfad={lauf.ordner} />
          )}

          {/* **Übersprungen ist kein Befund** (#491): niemand hat sich
              verrechnet, es war die Wahl. Es steht trotzdem da — „312 Bilder
              geschrieben" ist eine Zahl, die stimmt und nichts sagt. */}
          {lauf.uebersprungen.length > 0 && (
            <>
              <p className="mt-3 max-w-prose text-leise">
                {einsetzen(texte.seiten.bilder.uebersprungen, {
                  anzahl: String(lauf.uebersprungen.length),
                })}
              </p>
              <ul className="mt-1 space-y-1 font-mono text-sm">
                {lauf.uebersprungen.map((datei) => (
                  <li key={datei}>{datei}</li>
                ))}
              </ul>
            </>
          )}

          {lauf.befunde.length > 0 && (
            <>
              <p className="mt-3 max-w-prose text-leise">
                {texte.seiten.bilder.befunde}
              </p>
              <ul className="mt-1 space-y-1 font-mono text-sm">
                {lauf.befunde.map((befund) => (
                  <li key={befund}>{befund}</li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}
    </>
  )
}
