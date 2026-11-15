// **Die Karte selbst** — ihr Bild, die Linien darüber und die Flächen ihrer
// Felder (#246).
//
// # Warum eine eigene Datei
//
// Seit UXR-06 (#1914)
// rechnet die Karte ihre Breite selbst, und `Kartenvorschau.tsx` ginge damit
// über die Zeilengrenze (§2). Der Schnitt folgt der Frage: dort steht, **was
// um die Karte herum** steht — Ansicht, Nennung, Scryfall —, hier die Karte.
//
// # „Anpassen" heißt: die Karte passt
//
// Im Arbeitsplatz steht die Vorschau allein im Zentrum. Gemessen bei
// 1.280 × 800 mit dem Probebestand nahm sie dort zwei Fünftel der Breite —
// die Regel stammt aus der Zeit, als Formular und Vorschau nebeneinander
// standen (#1271) —, die Karte war 256 Punkte breit, begann bei y = 627 und
// stand zu sechzig Prozent unter dem Fensterrand. Jetzt nimmt sie die Breite
// des Zentrums, **aber nie mehr, als die Höhe unter ihr zulässt**
// (`passendeBreite`); gemessen wird diese Höhe wie die der Spalten
// (`useFensterrest`, derselbe Rand). Eine Stufe gibt ihre eigene Breite, und
// wer größer stellt, rollt.
import { useRef } from 'react'

import type { VorschauDTO } from '../api'
import { Spaltenrand } from '../bausteine/arbeitsplatzmasse'
import { useFensterrest } from '../haken/useFensterrest'
import { useTexte } from '../sprache/kontext'
import { Overlaylinien, type Overlaywahl } from './Vorschauoverlays'
import { cssBreite, passendeBreite, type Zoomwahl } from './zoom'

// Kartenmindesthoehe ist die Höhe, unter die eine passende Karte nicht
// fällt: in einem flachen Fenster rollt die Spalte, statt die Karte zu einer
// Briefmarke zu machen.
const Kartenmindesthoehe = 240

interface BuehnenEigenschaften {
  gezeigt: VorschauDTO
  zoom: Zoomwahl
  overlays: Overlaywahl

  // imZentrum sagt, dass die Karte allein im Zentrum des Arbeitsplatzes
  // steht — dann passt sie sich auch der Höhe an.
  imZentrum: boolean

  befundfelder: ReadonlySet<string>
  istBildfeld: (feld: string) => boolean
  bildWaehlen: (feld: string) => void
  feldWaehlen: (feld: string) => void
  merkeFeld: (feld: string | undefined) => void
}

export function Kartenbuehne({
  gezeigt,
  zoom,
  overlays,
  imZentrum,
  befundfelder,
  istBildfeld,
  bildWaehlen,
  feldWaehlen,
  merkeFeld,
}: BuehnenEigenschaften) {
  const texte = useTexte()
  const rahmen = useRef<HTMLDivElement>(null)
  const hoehe = useFensterrest(rahmen, {
    unten: Spaltenrand,
    mindestens: Kartenmindesthoehe,
  })

  const passend = imZentrum && zoom === 'anpassen'

  return (
    // **Die Karte ist das Hauptobjekt** (§6): über dem Bild liegen die
    // Flächen ihrer Felder, und ein Klick darauf setzt den Fokus ins
    // zugehörige Eingabefeld.
    //
    // **`relative`**, damit die Flächen am Bild hängen und nicht an der
    // Spalte: das Bild ist schmaler als seine Punktzahl, sobald das Fenster
    // es ist. **Ein Block mit `mx-auto`**, wenn sie passt — sie steht dann
    // in der Mitte —, **sonst `inline-block`**: eine gezoomte Karte, die
    // breiter ist als die Spalte, beginnt links, sonst läge ihr Anfang
    // unerreichbar außerhalb.
    <div
      ref={rahmen}
      className={
        passend
          ? 'relative mx-auto block'
          : 'relative inline-block ' +
            (zoom === 'anpassen' ? 'max-w-full' : 'max-w-none')
      }
      style={{
        width: passend
          ? passendeBreite(gezeigt, hoehe)
          : cssBreite(zoom, gezeigt),
      }}
    >
      <img
        src={gezeigt.bild}
        alt={texte.seiten.karten.vorschau}
        width={gezeigt.breite}
        height={gezeigt.hoehe}
        className="block w-full max-w-full rounded-sm border border-kante"
      />
      <Overlaylinien
        wahl={overlays}
        gezeigt={gezeigt}
        istBildfeld={istBildfeld}
        befundfelder={befundfelder}
      />

      {gezeigt.flaechen.map((flaeche, stelle) => (
        <button
          // **Der Schlüssel trägt die Stelle.** Ein Feld kann mehrere
          // Flächen haben — `magic-m15` zeichnet für `card color`
          // vier —, und zwei gleiche Schlüssel ließen React die
          // falsche wiederverwenden.
          key={`${flaeche.feld}/${String(stelle)}`}
          type="button"
          // **Nicht in der Tabfolge** (U5, #1416). Neunzehn Flächen
          // standen zwischen Zoomleiste und Namensfeld; wer mit der
          // Tastatur ins Formular will, hat dort dieselben Felder. Die
          // Fläche ist der Weg der Maus — und ein Vorleseprogramm
          // erreicht sie weiter, denn ein Knopf bleibt sie.
          tabIndex={-1}
          // **Ein Bildfeld nimmt ein Bild entgegen, jedes andere den
          // Fokus.** MSE öffnet den Dialog beim Doppelklick
          // (`ImageValueEditor::onLeftDClick`, `src/gui/value/image.cpp`,
          // Zeilen 22–35, Fassung f9d9356 vom 2021-07-06) und wählt
          // beim einfachen. Hier tut der einfache es: „wählen" heißt
          // bei uns, den Fokus ins Eingabefeld zu setzen, und ein
          // Bildfeld hat keines, in das sich tippen ließe.
          aria-label={`${flaeche.feld} — ${
            istBildfeld(flaeche.feld)
              ? texte.seiten.karten.bildEinsetzen
              : texte.seiten.karten.feldWaehlen
          }`}
          onClick={() => {
            if (istBildfeld(flaeche.feld)) {
              bildWaehlen(flaeche.feld)

              return
            }

            feldWaehlen(flaeche.feld)
          }}
          // **Und es nimmt eine gezogene Datei entgegen.** Der Pfad
          // kommt nicht von hier: der Browser sieht nur einen Namen.
          // Wails meldet ihn über `runtime.OnFileDrop`
          // ([abgelegteBilder]); hier wird nur festgehalten, über
          // welchem Feld losgelassen wurde.
          // **Die Marke für das Ablegen unter Wails 3** (ENG-02): dort
          // kommt kein `dragover` an, und `bildfeldAn` sucht das Feld
          // an der Stelle, die Go meldet.
          data-bildfeld={istBildfeld(flaeche.feld) ? flaeche.feld : undefined}
          // **Und ein Ablageziel für Wails** (AP-4, #2361, B-24): steht eine
          // Datei darüber, markiert die Laufzeit das Feld, es zeigt eine
          // gestrichelte Kante, und `Ablagehinweis` nennt es.
          data-file-drop-target={istBildfeld(flaeche.feld) ? '' : undefined}
          onDragOver={(ereignis) => {
            if (istBildfeld(flaeche.feld)) {
              ereignis.preventDefault()
              merkeFeld(flaeche.feld)
            }
          }}
          // **Und wer das Feld wieder verlässt, meint es nicht mehr.**
          //
          // Ohne diese Zeile blieb das Feld gemerkt, bis irgendwo etwas
          // abgelegt wurde — auch neben der Karte, auch auf einer
          // anderen. Der Zweitblick über Phase Y hat es als Befund 2
          // angestrichen: `nimmFeld` vergisst beim Ablegen, nicht beim
          // Weggehen.
          onDragLeave={() => {
            if (istBildfeld(flaeche.feld)) {
              merkeFeld(undefined)
            }
          }}
          // **Und wer den Zug beendet, ohne die Fläche zu verlassen** —
          // #1112.
          //
          // `dragleave` kommt nicht in jedem Fall: Escape über der
          // Fläche, ein Zug aus dem Fenster heraus, überlappende
          // Bildflächen. Ob dann noch eines kommt, entscheidet die
          // Plattform — `dragend` kommt immer.
          //
          // **Ohne diese Zeile traf das nächste Ablegen irgendwo das
          // Feld von vorhin**, wie im Fall darüber, nur seltener.
          onDragEnd={() => {
            if (istBildfeld(flaeche.feld)) {
              merkeFeld(undefined)
            }
          }}
          style={{
            left: `${String((flaeche.links / gezeigt.breite) * 100)}%`,
            top: `${String((flaeche.oben / gezeigt.hoehe) * 100)}%`,
            width: `${String((flaeche.breite / gezeigt.breite) * 100)}%`,
            height: `${String((flaeche.hoehe / gezeigt.hoehe) * 100)}%`,
          }}
          className="absolute rounded-sm hover:bg-marke/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-ring [&.file-drop-target-active]:bg-marke/10 [&.file-drop-target-active]:outline-2 [&.file-drop-target-active]:outline-dashed [&.file-drop-target-active]:outline-marke"
        />
      ))}
    </div>
  )
}
