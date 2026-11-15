// **Das Gebiet „Bestand" der Einstellungsseite** — Platz, Sicherung,
// Verschlüsselung und die Rückkehren.
//
// **Seit #2465 in zwei Teilen**: offen steht, was Platz belegt und was den
// Bestand sichert (`Bestandsgebiet`); Zurückgehen, Aufräumen und
// Zurücksetzen stehen in einem Abschnitt darunter (`Bestandseingriffe`),
// und was ein Auftrag kosten darf, in einem eigenen
// (`seiten/einstellungsbeiwerk.tsx`). Am Fenster rollte die Gruppe 3,7
// Bildschirme, eng 5,0.
//
// **Eine eigene Datei**, weil `Einstellungen.tsx` mit dem Rücksprung (O7,
// #1431) über die Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort
// steht, **welche Gebiete es gibt**, hier, **was in diesem einen steht**.
import type { Bestandspflege as Bestandspflegestand } from '../haken/useBestandspflege'
import type { Pflegelauf } from '../haken/useDatenpflege'
import type { Ruecksprungstand } from '../haken/useRuecksprung'
import { Bestandspflege } from '../seiten/Bestandspflege'
import { Platz } from './Platz'
import { Datenpflege } from '../seiten/Datenpflege'
import { Ruecksetzung } from '../seiten/Ruecksetzung'
import { Ruecksprung } from '../seiten/Ruecksprung'
import { useTexte } from '../sprache/kontext'

export function Bestandsgebiet({
  bestandspflege,
}: {
  bestandspflege: Bestandspflegestand
}) {
  const texte = useTexte()

  return (
    <div className="flex flex-col gap-6">
      {/* **Was Platz belegt, über der Sicherung** — Werkstück 4 von Phase
          4.13 (#1352). Wer hierherkommt, weil die Platte voll ist, sieht
          zuerst, was sie füllt. */}
      <section className="flex flex-col gap-2">
        <h3 className="text-base font-medium">
          {texte.seiten.bestandspflege.platzUeberschrift}
        </h3>

        <Platz />
      </section>

      <Bestandspflege
        sichern={bestandspflege.sicherung.sichern}
        sichertGerade={bestandspflege.sicherung.laeuft}
        waehlen={bestandspflege.rueckspiel.waehlen}
        gesperrt={bestandspflege.rueckspiel.gesperrt}
        entsperren={bestandspflege.rueckspiel.entsperren}
        zurueckspielen={bestandspflege.rueckspiel.zurueckspielen}
        spieltGerade={bestandspflege.rueckspiel.laeuft}
        auswahl={bestandspflege.rueckspiel.auswahl}
        lage={bestandspflege.rueckspiel.lage}
        beiseite={bestandspflege.rueckspiel.beiseite}
        fehler={bestandspflege.rueckspiel.fehler}
        abbrechen={bestandspflege.rueckspiel.raeumen}
        chiffre={bestandspflege.chiffre}
        alter={bestandspflege.alter}
      />
    </div>
  )
}

// Bestandseingriffe ist, was in den Bestand eingreift — der Abschnitt
// „Zurückgehen, aufräumen, zurücksetzen" (#2465).
export function Bestandseingriffe({
  bestandspflege,
  ruecksprung,
  pflege,
}: {
  bestandspflege: Bestandspflegestand
  ruecksprung: Ruecksprungstand
  pflege: Pflegelauf
}) {
  return (
    <div className="flex flex-col gap-6">
      {/* **Zuerst, was zurückgehen kann** (O7, #1431). Bei der Sicherung
          und nicht auf einer eigenen Seite: „zurück" ist die Frage, die
          jemand hier stellt — und die Antwort ist, dass es vier
          verschiedene sind. */}
      <Ruecksprung stand={ruecksprung} />

      {/* **Und zuletzt das Aufräumen** (O8, #1432). Unter dem Zurückgehen,
          weil es dieselbe Sorte Eingriff ist und die seltenere: wer hier
          landet, sucht nicht als Erstes nach verwaisten Zeilen. */}
      <Datenpflege stand={pflege} />

      {/* **Und ganz unten das Zurücksetzen** (O4, #1428). Unter dem
          Aufräumen, weil es dieselbe Sorte Eingriff ist und die schwerste:
          was hier steht, nimmt am meisten weg. */}
      <Ruecksetzung
        gewechselt={bestandspflege.gewechselt}
        veraendert={bestandspflege.veraendert}
      />
    </div>
  )
}
