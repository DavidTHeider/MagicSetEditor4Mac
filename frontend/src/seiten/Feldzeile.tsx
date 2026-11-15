// Eine **Zeile des Karteneditors** — Aufschrift, Eingabe, Palette, Erklärung,
// Fehler.
//
// # Warum als eigene Datei
//
// `Karteneditor.tsx` beantwortet die Frage „was tut dieses Fenster": laden,
// abbrechen, schreiben, den Stand halten. Wie **ein** Feld aussieht, ist eine
// andere Frage (AGENTS.md §2), und mit der Symbolpalette
// (#564) ist
// sie groß genug geworden, um sie zu stellen.
import type { ReactNode } from 'react'

import type {
  FeldbeschreibungDTO,
  PalettenEintragDTO,
  WortlistenbereichDTO,
} from '../api'
import { Entwurfskonflikt } from '../bausteine/Entwurfskonflikt'
import { Knopf } from '../bausteine/Knopf'
import { useEntwurf } from '../haken/useEntwurf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Feldeingabe } from './Feldeingabe'
import { fehlerkennung, hilfekennung } from './feldkennung'
import { ohneAuszeichnung } from './ohneauszeichnung'
import { Symbolpalette } from './Symbolpalette'
import { Wortlistenwahl } from './Wortlistenwahl'

interface FeldzeileEigenschaften {
  feld: string
  wert: string

  // form ist die Beschreibung aus dem Spielpaket — `undefined`, wenn das Spiel
  // dieses Feld nicht erklärt.
  form: FeldbeschreibungDTO | undefined

  fehler: string | undefined
  laeuft: boolean

  // palette ist gesetzt, solange **dieses** Feld gewählt ist und seine Palette
  // geladen wurde.
  palette: { eintraege: PalettenEintragDTO[]; roh: boolean } | undefined

  // wortlisten sind die Auswahlbereiche **dieses** Feldes — leer, solange ein
  // anderes gewählt ist oder das Feld keine hat
  // (#1074).
  wortlisten: WortlistenbereichDTO[]

  // hergeleitet ist der **gerechnete** Wert dieses Feldes, wenn die Karte keinen
  // eigenen führt (#563).
  //
  // **Er ist keine Eingabe, sondern eine Herleitung.** MSE hält denselben
  // Unterschied am Wert selbst (`Defaultable<String>`): das Feld trägt ihn, ist
  // aber als „default" markiert und wird beim ersten Tippen zu einem eigenen.
  hergeleitet: string | undefined

  // vorschlag steht unter der Eingabe — der KI-Vorschlag an Name und
  // Flavortext (AI-05, #2042); fehlt er, hat das Feld keinen.
  vorschlag?: ReactNode | undefined

  gewaehlt: () => void

  // getippt meldet den Entwurf für die Vorschau; `undefined` verwirft ihn.
  getippt: (wert: string | undefined) => void
  // schreiben sagt mit einem Versprechen, ob es glückte (`useEntwurf`).
  schreiben: (wert: string, bekannt: readonly string[]) => unknown
  // einfuegen setzt ein Symbol ins Feld und gibt den neuen Wert — **den
  // schreibt der Entwurf** (`uebernehmen`), nicht der Aufrufer: im Konflikt
  // ginge der fremde Wert sonst mit einem Klick verloren.
  einfuegen: (code: string, roh: boolean) => string | undefined

  // bildWaehlen öffnet den Dateidialog für ein Feld der Art `image`.
  bildWaehlen: () => void

  // symbolZeichnen öffnet den Symboleditor für ein Feld der Art `symbol`.
  //
  // **Wahlfrei**, und ohne ihn steht kein Knopf da: eine Fläche, die keinen
  // Editor hat, soll keinen anbieten — ein Knopf, der nichts tut, ist
  // schlimmer als keiner
  // (#1339).
  symbolZeichnen?: (() => void) | undefined

  // wortWaehlen wählt einen Eintrag aus einer Auswahlliste — mit den Werten,
  // die der Entwurf kennt.
  wortWaehlen: (
    bereich: number,
    weg: number[],
    bekannt: readonly string[],
  ) => void

  // autoErsetzen reicht die `auto replace:`-Regeln des Spiels durch —
  // #897.
  autoErsetzen?:
    | ((
        text: string,
        stelle: number,
      ) => Promise<{ text: string; stelle: number }>)
    | undefined
}

export function Feldzeile({
  feld,
  wert,
  form,
  fehler,
  laeuft,
  palette,
  wortlisten,
  hergeleitet,
  vorschlag,
  gewaehlt,
  getippt,
  schreiben,
  einfuegen,
  bildWaehlen,
  symbolZeichnen,
  wortWaehlen,
  autoErsetzen,
}: FeldzeileEigenschaften) {
  const texte = useTexte()
  const beschreibung = form?.beschreibung ?? ''

  // **Nur, solange das Feld leer ist.** Kommt eine Vorgabe für ein Feld, das
  // die Karte führt, ist das ein Fehler weiter unten; die Zeile zeigt dann den
  // eigenen Wert und keinen Hinweis, statt beides zu behaupten.
  const gezeigteHerleitung = wert === '' ? hergeleitet : undefined

  // **Ein Entwurf übersteht fremde Änderungen** (U01, U02, #2119): die Regel
  // steht in `useEntwurf`.
  const entwurf = useEntwurf(wert, schreiben, getippt)

  return (
    <label
      // **Ein mehrzeiliges Feld nimmt beide Spalten** (#674): der Regeltext
      // einer Karte in einer halben Breite wäre eine Spalte aus Silben.
      // **`group`, damit die Erklärung am Fokus hängt** (#1271) — sie steht
      // unten und wird sichtbar, sobald jemand in diesem Feld steht.
      className={
        'group flex flex-col gap-1' +
        (form?.mehrzeilig === true ? ' xl:col-span-2' : '')
      }
      // **Am Umschlag und nicht an der Eingabe.** [Feldeingabe] gibt je nach
      // Art ein `input`, ein `textarea`, ein `select` oder eine Ankreuzfläche
      // her; ein `onFocus` an jeder von ihnen wäre dieselbe Zeile viermal — und
      // beim fünften Steuerelement vergessen.
      onFocusCapture={gewaehlt}
    >
      {/* **Die Aufschrift und nicht der Name** — MSE beschriftet so
          (`src/data/field.cpp`, Zeile 69, Fassung f9d9356 vom 2021-07-06). Der
          Name steht im `aria-label` der Eingabe: er ist die Kennung, an der das
          Schreiben hängt. */}
      <span className="text-sm font-medium">{form?.aufschrift ?? feld}</span>

      {/* **Ein Bildfeld wird gewählt, nicht getippt.** MSE öffnet dafür einen
          Dateidialog (`ImageValueEditor::onLeftDClick`,
          `src/gui/value/image.cpp`, Zeilen 22–35, Fassung f9d9356 vom
          2021-07-06); der Dateiname, den das Set danach trägt, entsteht beim
          Ablegen und ist nichts, was jemand kennen kann. */}
      {form?.art === 'image' && (
        <Knopf
          gewicht="neben"
          bezeichnung={`${form.aufschrift} — ${texte.seiten.karten.bildWaehlen}`}
          angehalten={laeuft}
          aufKlick={bildWaehlen}
        >
          {texte.seiten.karten.bildWaehlen}
        </Knopf>
      )}

      {/* **Ein Symbolfeld wird gezeichnet, nicht getippt.** Sein Wert ist ein
          **Dateiname im Paket** (`src/data/field/symbol.hpp`, Zeilen 63–70,
          Fassung f9d9356 vom 2021-07-06), und MSE2 öffnet dafür seinen
          Symboleditor (`src/gui/value/symbol.cpp`) — derselbe Grund wie beim
          Bildfeld darüber: ein Dateiname entsteht beim Ablegen und ist nichts,
          was jemand kennen kann
          (#1339). */}
      {form?.art === 'symbol' && symbolZeichnen !== undefined && (
        <Knopf
          gewicht="neben"
          bezeichnung={`${form.aufschrift} — ${texte.seiten.karten.symbolZeichnen}`}
          angehalten={laeuft}
          aufKlick={symbolZeichnen}
        >
          {texte.seiten.karten.symbolZeichnen}
        </Knopf>
      )}

      {/* **Der gespeicherte Wert ist der Schlüssel** — gefunden an der
          Hüllenprobe zu AI-05 (#2042): die Eingabe steht auf `defaultValue`,
          und ein von außen geschriebener Wert (KI-Übernahme, Wortwahl,
          Abgleich) stand im Bestand, aber nicht im Feld. **Solange ein
          Entwurf lebt, bleibt der Schlüssel der von damals** (U01): eine
          fremde Änderung baut die Eingabe nicht neu, und wer tippt, behält
          Text und Fokus. */}
      <Feldeingabe
        key={entwurf.schluessel}
        feld={feld}
        wert={wert}
        form={form}
        angehalten={laeuft}
        fehlerhaft={fehler !== undefined}
        beschriebenVon={hinweis(feld, fehler, beschreibung)}
        hergeleitet={gezeigteHerleitung}
        autoErsetzen={autoErsetzen}
        aendern={entwurf.aendern}
        verlassen={entwurf.verlassen}
      />

      <Entwurfskonflikt
        griff={entwurf}
        fremd={ohneAuszeichnung(wert)}
        angehalten={laeuft}
      />

      {vorschlag}

      {/* **Die Herleitung steht unter dem Feld, benannt als solche** (#563).
          Der Platzhalter im Feld zeigt sie **dort**, wo der Wert stünde; diese
          Zeile sagt, **dass** es eine ist — und ein Platzhalter allein sagt
          das nicht, er sieht aus wie ein Hinweis, was man tippen könnte. */}
      {gezeigteHerleitung !== undefined && (
        <span className="text-sm text-leise">
          {einsetzen(texte.seiten.karten.hergeleitetWert, {
            // **Ohne Auszeichnungen** (#687): der Hinweis ist ein Satz an
            // jemanden, kein Wert. Was geschrieben wird, bleibt roh.
            wert: ohneAuszeichnung(gezeigteHerleitung),
          })}
        </span>
      )}

      {/* **Die Palette steht unter dem Feld, zu dem sie gehört.** MSE hängt
          dieselben Einträge in ein Kontextmenü, weil ein wxWidgets-Textfeld
          keine Kacheln neben sich stellen kann; hier sind sie sichtbar, ohne
          dass jemand sucht. */}
      {palette !== undefined && (
        <Symbolpalette
          eintraege={palette.eintraege}
          angehalten={laeuft}
          einfuegen={(code) => {
            const neu = einfuegen(code, palette.roh)

            if (neu !== undefined) {
              getippt(neu)
              entwurf.uebernehmen(neu)
            }
          }}
        />
      )}

      {/* **Die Auswahl steht unter dem Feld, zu dem sie gehört** (#1074).
          MSE setzt einen Pfeil **in** den Text; warum es hier anders aussieht,
          steht in [Wortlistenwahl]. */}
      <Wortlistenwahl
        bereiche={wortlisten}
        angehalten={laeuft}
        waehlen={(bereich, weg) => {
          wortWaehlen(bereich, weg, entwurf.bekannt)
        }}
      />

      {/* **Die Erklärung aus dem Paket** — unübersetzt (§9), und über
          `aria-describedby` am Feld statt als Absatz daneben.

          # Sichtbar, sobald jemand in dem Feld steht

          **Nachgelesen, nicht entworfen** (#1271):

              LocalizedString description;  ///< Description, used in status bar

          `src/data/field.hpp`, Zeile 48, und `card_editor.cpp`, Zeile 358:
          `SetStatusText(hovered_viewer->getField()->description.get())`. MSE2
          zeigt sie **in der Statuszeile beim Überfahren** — nicht dauerhaft
          unter jedem Feld.

          Unter allen Feldern zugleich wären das — gemessen am 2026-08-28 —
          **1.683 px für 16 Felder** in einem 900-px-Fenster, zwei Bildschirme
          für eine Karte. Die Sätze sind lang; `illustrator` allein hat
          zweiundzwanzig Wörter.

          **`sr-only` und nicht `hidden`**: eine Vorlesehilfe liest sie
          weiterhin — `aria-describedby` zeigt auf genau diesen Knoten, und ein
          entfernter Knoten wäre eine Zusage, die ins Leere geht. */}
      {beschreibung !== '' && (
        <span
          id={hilfekennung(feld)}
          className="sr-only text-sm text-leise group-focus-within:not-sr-only"
        >
          {beschreibung}
        </span>
      )}

      {fehler !== undefined && (
        <span role="alert">
          <span id={fehlerkennung(feld)} className="text-sm text-fehler">
            {fehler}
          </span>
        </span>
      )}
    </label>
  )
}

// hinweis sagt, welche Erklärung an einem Feld hängt.
//
// **Der Fehler geht vor.** Wo beides dasteht, ist die Meldung das, was jemand
// jetzt braucht; die Beschreibung des Feldes hat er schon gelesen.
function hinweis(
  feld: string,
  fehler: string | undefined,
  beschreibung: string,
): string | undefined {
  if (fehler !== undefined) {
    return fehlerkennung(feld)
  }

  if (beschreibung !== '') {
    return hilfekennung(feld)
  }

  return undefined
}
