// **Eine Karte als Bild am Spieltisch** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Warum das Anmelden hier steht
//
// **Weil React schon weiß, wann ein Ding erscheint und verschwindet.** Ein
// Bauteil, das sich beim Erscheinen anmeldet und beim Verschwinden abmeldet,
// führt die Referenzzählung von selbst — und sieben gleiche Karten melden
// sich siebenmal an und einmal geholt.
//
// # Und warum ein Platzhalter bleibt, solange nichts da ist
//
// **Weil eine Karte eine Größe hat, bevor sie ein Bild hat.** Wer sie erst
// mit dem Bild einfügte, ließe den Tisch bei jedem Nachladen springen.
//
// # Warum das Bild für ein Vorleseprogramm nicht da ist
//
// **Weil daneben schon steht, welche Karte es ist.** Es sitzt in einem Knopf,
// dessen Aufschrift die Karte nennt; ein `alt` mit demselben Text sagte sie
// **zweimal** — und für jemanden, der zuhört, ist das kein Bild, sondern eine
// Wiederholung. Ein Bild, das nichts hinzufügt, ist schmückend und wird
// verborgen (WCAG 1.1.1).
import { useEffect } from 'react'

import type { Texturwunsch } from '../apitextur'
import type { Texturlauf } from '../haken/useSpieltexturen'

interface Kartenbildeigenschaften {
  texturen: Texturlauf

  wunsch: Texturwunsch

  // ausweichguete ist die Stufe, die gezeigt wird, solange die gewünschte
  // fehlt — **das Vorgeladene** (#1459: „Hohe Auflösung bei Inspektion/Zoom
  // nachladen").
  //
  // **Sie wird nicht angemeldet**: sie liegt schon da oder eben nicht. Wer
  // sie holte, machte aus jedem Bild zwei.
  //
  // **Pflicht und nicht wahlfrei.** Der erste Anlauf ließ sie aus — und der
  // Zweig für „keine" hatte keinen Aufrufer, weil jede Stelle, die eine
  // Karte zeigt, etwas hat, worauf sie ausweichen kann. Ein Zweig ohne
  // Aufrufer ist keine Anwendungsfunktion (§2), und ein Test dafür prüfte
  // eine Welt, die es nicht gibt.
  ausweichguete: string
}

export function Kartenbild({
  texturen,
  wunsch,
  ausweichguete,
}: Kartenbildeigenschaften) {
  const { belege, gib } = texturen

  useEffect(() => {
    belege(wunsch)

    return () => {
      gib(wunsch)
    }
    // **Der Wunsch wird über seine Felder verglichen und nicht als Objekt.**
    // Ein frisch gebautes Objekt ist bei jedem Zeichnen ein neues, und der
    // Effekt liefe jedes Mal — ab- und wieder anmelden, für dasselbe Bild.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [belege, gib, ...merkmale(wunsch)])

  const adresse =
    texturen.adresse(wunsch) ||
    texturen.adresse({ ...wunsch, guete: ausweichguete })

  if (adresse === '') {
    // **Ein Platzhalter mit dem Seitenverhältnis einer Karte** — 375 : 523,
    // das Maß jeder `magic-m15`-Vorlage. Ohne es sprängen die Zeilen, sobald
    // ein Bild ankommt.
    return (
      <span aria-hidden="true" className={platzhalterklassen(wunsch.guete)} />
    )
  }

  return <img src={adresse} alt="" className={bildklassen(wunsch.guete)} />
}

// merkmale gibt die Felder eines Wunsches als Vergleichsliste.
function merkmale(wunsch: Texturwunsch): string[] {
  if (wunsch.art === 'ruecken') {
    return [wunsch.art, wunsch.spiel, wunsch.guete]
  }

  return [wunsch.art, wunsch.karteID, wunsch.seite, wunsch.guete]
}

// bildklassen und platzhalterklassen geben die Klassen eines Kartenbildes.
//
// **Außerhalb des JSX**, weil `keintextimjsx.test.ts` jede Zeichenkette mit
// Platzhalter im Baum abweist — und zwar zu Recht: ein Textknoten mit `${…}`
// ist genau die Gestalt, in der ein Oberflächentext durchrutscht.
// Klassennamen sind keine Oberflächentexte, also stehen sie hier.
function bildklassen(guete: string): string {
  return grundklassen + weiteZu(guete)
}

function platzhalterklassen(guete: string): string {
  return grundklassen + 'bg-betont ' + weiteZu(guete)
}

// grundklassen sind, was beide gemeinsam haben — **das Seitenverhältnis einer
// Karte**, 375 : 523, das Maß jeder `magic-m15`-Vorlage.
const grundklassen = 'block aspect-[375/523] rounded-sm border border-kante '

// weiteZu gibt die Breite im Fenster zur gewünschten Güte.
//
// **Die Güte *ist* die Größe.** Sie in einer eigenen Eigenschaft noch einmal
// zu nennen hieße, zwei Angaben zu führen, die dasselbe bedeuten — und am
// Tag, an dem sie auseinanderlaufen, zeigt ein 1024 Punkte breites Bild in
// einem 80 Punkte breiten Kästchen.
//
// **Zwei Größen und nicht fünf.** `miniatur` und `ausfuhr` haben im Fenster
// heute keinen Aufrufer — eine Zeile dafür wäre ein Zweig, den niemand
// betritt (§2). Wer sie einführt, führt hier eine Größe dazu ein, und ein
// roter Abdeckungslauf sagt es ihm.
function weiteZu(guete: string): string {
  if (guete === 'betrachten') {
    return 'w-56'
  }

  return 'w-20'
}
