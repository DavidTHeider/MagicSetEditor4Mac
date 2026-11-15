// **Die Vorschau der Karte, und was sich auf ihr tun lässt** —
// #246,
// #576.
//
// Eigene Datei, weil `Karteneditor.tsx` mit ihr über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort stehen die **Felder** der Karte, hier ihr
// **Bild** — und alles, was daran hängt: die Ansicht, die Befunde des
// Rasterers und der offene Zuschnitt. Die Karte selbst mit ihren
// anklickbaren Feldflächen steht seit UXR-06 in `Kartenbuehne.tsx`.

import type { AuswahlDTO, VorschauDTO } from '../api'
import type { OffenerZuschnitt } from '../haken/useBildfeld'
import { useTexte } from '../sprache/kontext'
import { Ladefeld } from './Ladefeld'
import { Ziehgriff } from '../bausteine/Ziehgriff'
import { Nennung } from '../bausteine/Nennung'
import { Objektkontextmenue } from '../bausteine/Objektkontextmenue'
import { Scryfallabgleich } from '../bausteine/Scryfallabgleich'
import type { Abgleichsstand } from '../haken/useScryfallabgleich'
import { Scryfallbild } from '../bausteine/Scryfallbild'
import { Scryfallregeln } from '../bausteine/Scryfallregeln'
import { Zuschnittfenster } from './Zuschnittfenster'
import { Massstab } from './Massstab'
import { Overlayschalter } from './Vorschauoverlays'
import type { Ansichtswahl } from './vorschauansicht'
import { Kartenbuehne } from './Kartenbuehne'
import { feldVon } from './Problemliste'

// Ladestand ist der Teil von `useLaden`, den diese Spalte braucht.
interface Ladestand {
  laedt: boolean
  fehler: string | undefined

  // abgeloest ist das Bild der vorigen Karte, solange die neue lädt
  // (`useKartenvorschau`).
  abgeloest?: VorschauDTO | undefined
}

// keineBefunde stehen am abgelösten Bild: seine gehören zur vorigen Karte.
const keineBefunde = new Set<string>()

// ladeschild steht über dem abgelösten Bild, oben in der Mitte.
const ladeschild =
  'absolute left-1/2 top-3 z-10 -translate-x-1/2 rounded-md border border-kante bg-flaeche px-2 py-1 text-sm text-text'

// **Und kein Bildfeld** — sonst ginge ein Bild, das während des Ladens darauf
// fällt, in das gleichnamige Feld der neuen Karte (§7, Regel 1).
const keinBildfeld = () => false

interface KartenvorschauEigenschaften {
  // karteID nennt die Karte für ihr Kontextmenü (GESTE, #2358): ein
  // Rechtsklick auf das Bild zeigt dieselben Handlungen wie „Diese Karte".
  karteID: string

  // ansicht ist die Größe, die der Editor gerade rastert (U5, #1416), samt
  // den Überlagerungen. Die Zeile hier wählt sie, gehalten wird sie über dem
  // Editor (UXR-06, `vorschauansicht.ts`) — er fragt damit den Go-Teil.
  // **Die Überlagerungen sind eine Wahl dieser Ansicht**, kein Gedächtnis:
  // wer sie einschaltet, prüft gerade etwas — und prüft weiter, wenn er
  // dabei ein Feld ändert.
  ansicht: Ansichtswahl

  // imZentrum sagt, dass die Vorschau allein im Zentrum des Arbeitsplatzes
  // steht und nicht neben dem Formular (UXR-06): dann nimmt sie die ganze
  // Breite, und die Karte passt sich auch der Höhe an (`Kartenbuehne`).
  imZentrum: boolean

  vorschau: Ladestand
  gezeigt: VorschauDTO | undefined
  zuschnitt: OffenerZuschnitt | undefined
  laeuft: boolean
  istBildfeld: (feld: string) => boolean
  bildWaehlen: (feld: string) => void

  // vonScryfall holt das Bild zum Kartennamen (#1005).
  vonScryfall: (feld: string, name: string, art: string) => void

  // kartenname ist der Name, mit dem der Kasten vorbelegt wird.
  kartenname: string

  // manakosten ist der Wert des Kostenfelds — für die Prüfung gegen Scryfall.
  manakosten: string

  // gewaehltesBildfeld ist das Bildfeld, für das der Kasten gilt — `undefined`,
  // solange keines gewählt ist.
  gewaehltesBildfeld: string | undefined

  // abgleich ist der Feldabgleich mit Scryfall (#1013).
  abgleich: Abgleichsstand
  feldWaehlen: (feld: string) => void
  merkeFeld: (feld: string | undefined) => void
  uebernehmen: (auswahl: AuswahlDTO) => void
  abbrechen: () => void
}

export function Kartenvorschau({
  karteID,
  ansicht: [{ zoom, overlays }, setzeAnsicht],
  imZentrum,
  vorschau,
  gezeigt,
  zuschnitt,
  laeuft,
  istBildfeld,
  bildWaehlen,
  vonScryfall,
  kartenname,
  manakosten,
  gewaehltesBildfeld,
  abgleich,
  feldWaehlen,
  merkeFeld,
  uebernehmen,
  abbrechen,
}: KartenvorschauEigenschaften) {
  const texte = useTexte()

  // **Der Stand des gezeigten Vorschlags** (#1241) — leer, solange keiner
  // dasteht.
  //
  // **Hier und nicht im Rückruf**: dort liefe nur der eine Zweig, denn der
  // Knopf steht nur da, wo ein Vorschlag steht. Ein Zweig, den kein Aufruf
  // betritt, ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
  const abgleichsstand = abgleich.vorschlag?.stand ?? ''

  // **Ob die Vorlage eine Auflösung nennt** — solange das Bild unterwegs
  // ist, gilt sie als genannt: sonst flackerte der Satz dazu bei jeder Stufe.
  const aufloesend = (gezeigt?.dpi ?? 1) > 0
  const befundfelder = new Set(
    (gezeigt?.befunde ?? [])
      .map(feldVon)
      .filter((feld): feld is string => feld !== undefined),
  )

  return (
    // **Die Karte bleibt im Bild** (§6,
    // #667).
    //
    // Bei `magic-m15` ist das Formular daneben fast 5.000 Punkte hoch — 25
    // Felder mit ihren Beschreibungen aus dem Paket. Ohne Haftung war die
    // Vorschau fort, sobald jemand `Text` bearbeitete, und „eine Änderung ist
    // in der Vorschau zu sehen" galt nur für die ersten drei Felder.
    //
    // **Nur nebeneinander.** Bricht die Spalte um, steht die Vorschau unter
    // dem Formular; ein haftender Block fräße dort die halbe Höhe.
    //
    // # Und sie brach um, genau da, wo jemand arbeitet
    //
    // **Gemessen am 2026-08-28** bei 1280 × 800, dem kleinsten Fenster, das
    // ein MacBook Air voreingestellt hat
    // (#1271):
    // die Editorspalte ist dort **590 px** breit. Formular und Vorschau
    // trugen beide `min-w-72` — 288 + 288 + 32 Abstand sind **608**, also
    // mehr als 590. Die Spalte brach um, die Haftung fiel weg, und die
    // Vorschau stand bei y = 743, während das Namensfeld bei y = 395 steht.
    // Bei 800 Punkten Fenster heißt das: **wer tippt, sieht die Karte nicht.**
    //
    // Genau das verbietet §6, und es stand als Zusage schon hier — nur galt
    // sie erst ab einer Breite, die niemand nachgemessen hatte. Eine Zusage
    // mit einer Bedingung, die keiner kennt, ist keine.
    //
    // **Die Vorschau nimmt jetzt zwei Fünftel und hat eine kleinere
    // Untergrenze.** Bei 590 px sind das 236 für die Karte und 330 für die
    // Felder; beides ist benutzbar, und nichts bricht um. Die Karte selbst
    // ist über `max-w-full` (unten am `img`) so breit wie ihr Platz — ein
    // kleineres Bild beantwortet „ist meine Änderung angekommen" genauso.
    //
    // **Im Arbeitsplatz gilt das nicht mehr** (UXR-06): dort steht das
    // Formular im Inspector, und die Vorschau hat das Zentrum für sich.
    <div
      className={
        (imZentrum
          ? 'w-full'
          : 'w-2/5 max-w-[23.5rem] min-w-52 shrink-0 lg:sticky lg:top-0 lg:self-start') +
        (zoom === 'anpassen' ? '' : ' overflow-auto')
      }
    >
      {/* **Die Ansicht in einer Zeile** (UXR-06, #1914): der Maßstab als
          Auswahl und die drei Überlagerungen daneben — vorher sechs Knöpfe
          in bis zu drei Zeilen und die Kästchen in einer vierten. **Über der
          Karte und nicht darunter**, wo das Artboard die Kästchen zeichnet:
          bei 200 % stünden sie sonst unter einer Karte, die höher ist als
          das Fenster.

          **Und sie steht, während neu gerastert wird** — ab 150 % kommt ein
          neues Bild, und bis es da ist, gibt es keines. Hinge die Zeile am
          Bild, verschwände sie für einen Herzschlag und mit ihr der Fokus. */}
      {(gezeigt !== undefined || vorschau.laedt) && (
        <div className="mb-2 flex flex-wrap items-center gap-x-4 gap-y-2">
          <Massstab
            zoom={zoom}
            waehlen={(neu) => {
              setzeAnsicht((bisher) => ({ ...bisher, zoom: neu }))
            }}
            druckMoeglich={aufloesend}
          />
          <Overlayschalter
            wahl={overlays}
            umschalten={(art) => {
              setzeAnsicht((bisher) => ({
                ...bisher,
                overlays: {
                  ...bisher.overlays,
                  [art]: !bisher.overlays[art],
                },
              }))
            }}
            beschnittMoeglich={aufloesend}
          />
          {/* **Das Bild in den Finder ziehen** (AP-12, #2369, B-24) — erst,
              wenn es eines gibt. */}
          {gezeigt !== undefined && (
            <Ziehgriff bild={gezeigt.bild} kartenname={kartenname} />
          )}
        </div>
      )}

      <Ladefeld
        laedt={vorschau.laedt}
        fehler={vorschau.fehler}
        leer={false}
        vorhanden={gezeigt !== undefined || vorschau.abgeloest !== undefined}
      >
        <></>
      </Ladefeld>

      {/* **Beim Kartenwechsel blitzt sie nicht leer** — AP-4 (#2361): bis
          das neue Bild da ist, steht das der vorigen Karte, **nur zum
          Ansehen** — gedämpft, `inert` (kein Feld, kein Kontextmenü),
          `aria-busy`. Ohne es fiel die Spalte bis zur Antwort zusammen. Daraus
          gilt nichts für die neue Karte (§7, Regel 1). */}
      {gezeigt === undefined && vorschau.abgeloest !== undefined && (
        <div inert aria-busy="true" className="relative">
          {/* **Man sieht, dass es lädt** — gedämpft allein hielte man das
              Bild für das der neuen Karte. Für ein Vorleseprogramm sagt es
              `Ladefeld` oben; hier steht es nur für die Augen. */}
          <p aria-hidden="true" className={ladeschild}>
            {texte.bausteine.zustaende.laedt}
          </p>
          <div className="opacity-60">
            <Kartenbuehne
              gezeigt={vorschau.abgeloest}
              zoom={zoom}
              overlays={overlays}
              imZentrum={imZentrum}
              befundfelder={keineBefunde}
              istBildfeld={keinBildfeld}
              bildWaehlen={bildWaehlen}
              feldWaehlen={feldWaehlen}
              merkeFeld={merkeFeld}
            />
          </div>
        </div>
      )}

      {gezeigt && (
        <>
          <Objektkontextmenue art="karte" kennung={karteID}>
            <Kartenbuehne
              gezeigt={gezeigt}
              zoom={zoom}
              overlays={overlays}
              imZentrum={imZentrum}
              befundfelder={befundfelder}
              istBildfeld={istBildfeld}
              bildWaehlen={bildWaehlen}
              feldWaehlen={feldWaehlen}
              merkeFeld={merkeFeld}
            />
          </Objektkontextmenue>
          {/* **Die Nennung steht hier und nicht am Zuschnitt** (#1025).
              Diese Fläche zeigt den Bildausschnitt, auch wenn der Kasten
              längst zu ist — und Scryfall verlangt Künstler und Copyright
              „in the same interface presenting the art crop". */}
          <Nennung nennungen={gezeigt.nennungen} />
          {/* **Die Befunde stehen in der Problemliste** (U5): dort führen sie
              zum Feld, hier stünden sie ein zweites Mal ohne Weg. */}
        </>
      )}

      {/* **Der Ausschnitt steht im Editor, nicht in einem Fenster darüber**
            (§6: keine verschachtelten Dialoge). Er nimmt den Platz neben der
            Vorschau ein, solange eine Wahl offen ist
            (#576). */}
      {/* **Der Weg zu Scryfall steht am gewählten Bildfeld** (#1005) — nicht
          als eigener Reiter: er füllt dieses Feld und nichts sonst. */}
      {zuschnitt === undefined && gewaehltesBildfeld !== undefined && (
        <Scryfallbild
          feld={gewaehltesBildfeld}
          holen={(name, art) => {
            vonScryfall(gewaehltesBildfeld, name, art)
          }}
          laeuft={laeuft}
          vorschlag={kartenname}
        />
      )}

      {/* **Die Auskünfte hängen an keinem Feld** (#888) — sie ändern nichts
          an der Karte und stehen deshalb unabhängig von der Feldwahl. */}
      {/* **Der Feldabgleich hängt an keinem Feld** (#1013) — er betrifft die
          ganze Karte, nicht das gewählte Feld. Er steht **vor** den
          Auskünften, weil er schreibt und sie nur lesen. */}
      <Scryfallabgleich
        holen={() => {
          abgleich.holen(kartenname)
        }}
        kartenname={kartenname}
        laeuft={laeuft}
        meldung={abgleich.meldung}
        uebernehmen={(felder) => {
          // **Der Stand des gezeigten Vorschlags geht mit** (#1241).
          abgleich.uebernehmen(kartenname, felder, abgleichsstand)
        }}
        vorschlag={abgleich.vorschlag}
      />

      <Scryfallregeln kartenname={kartenname} manakosten={manakosten} />

      {zuschnitt !== undefined && (
        <div className="mt-3">
          <Zuschnittfenster
            zuschnitt={zuschnitt.antwort}
            laeuft={laeuft}
            uebernehmen={uebernehmen}
            abbrechen={abbrechen}
          />

          {/* **Die Nennung steht neben dem Bild, solange es hier liegt.**
              Scryfall verlangt Künstler und Copyright „in the same interface
              presenting the art crop" — diese Fläche ist sie. */}
          {zuschnitt.scryfall !== undefined && (
            <p className="mt-2 text-xs text-leise">
              {zuschnitt.scryfall.nennung}{' '}
              <a
                className="underline focus-visible:outline-2"
                href={zuschnitt.scryfall.adresse}
                rel="noreferrer"
                target="_blank"
              >
                {texte.seiten.karten.scryfall.quelle}
              </a>
            </p>
          )}

          {/* **Und was die Auskunft tolerant hingenommen hat, steht dabei.**
              §16: nichts wird still übernommen. Ein Feld, das der Anbieter
              umbenannt hat, hält niemanden auf — verschwiegen wird es
              trotzdem nicht (Nachlauf über Phase 4.5, Befund 10). */}
          {zuschnitt.scryfall !== undefined &&
            zuschnitt.scryfall.befunde.length > 0 && (
              <ul className="mt-2 text-xs text-hinweis">
                {zuschnitt.scryfall.befunde.map((befund) => (
                  <li key={befund}>{befund}</li>
                ))}
              </ul>
            )}
        </div>
      )}
    </div>
  )
}
