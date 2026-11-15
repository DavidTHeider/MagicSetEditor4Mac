// **Der Speicher der Spieltexturen** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wofür es ihn gibt
//
// **Damit sieben gleiche Karten eine Textur benutzen und sieben Dinge
// bleiben.** Das ist die Abnahme des Issues, wörtlich — und sie hat zwei
// Hälften: das Bild wird **einmal** geholt, und es wird weggeworfen, wenn das
// **letzte** Objekt es losgelassen hat.
//
// # Warum eine Referenzzählung und kein Wegwerfen beim Verschwinden
//
// **Weil ein Objekt verschwindet, während ein anderes dieselbe Textur noch
// zeigt.** Wer beim Abräumen einer Karte ihre Textur verwürfe, nähme sie den
// übrigen sechs weg — und die müssten sie neu holen, obwohl sie dastand. Das
// Issue sagt es als Zusage: „ein noch verwendetes Shared Asset wird nicht
// zerstört."
//
// # Und warum er kein React kennt
//
// **Weil er sich so prüfen lässt.** Was hier steht, ist Buchführung über
// Zahlen; dass ein Bauteil sie beim Erscheinen und Verschwinden führt, ist
// eine andere Frage und steht in `useSpieltexturen`.
import type { Textur, Texturwunsch } from '../apitextur'

// Texturbeleg ist, was der Speicher zu einem Schlüssel hält.
interface Texturbeleg {
  // zahl ist, wie viele Dinge diese Textur gerade zeigen — **null bei
  // einem vorgeladenen Bild**, das noch niemand zeigt.
  zahl: number

  // textur ist leer, solange sie noch geholt wird.
  textur?: Textur
}

// texturschluessel macht aus einem Wunsch einen Schlüssel.
//
// **Die Art steht vorn**, damit ein Spiel namens `7` und eine Karte namens
// `7` nicht denselben Schlüssel bekommen.
//
// **Der *inhaltliche* Schlüssel kommt aus der Antwort** und steht in
// `Textur.schluessel`; dieser hier ist der Griff, mit dem das Fenster fragt.
export function texturschluessel(wunsch: Texturwunsch): string {
  if (wunsch.art === 'ruecken') {
    return ['ruecken', wunsch.spiel, wunsch.guete].join(':')
  }

  // **Ohne eigene Zeile die Bildkennung** (P9, #2377) — sie erkennt dieselbe
  // Karte an jedem Ende.
  if (wunsch.karteID === '') {
    return ['bild', wunsch.bild ?? '', wunsch.seite, wunsch.guete].join(':')
  }

  return ['karte', wunsch.karteID, wunsch.seite, wunsch.guete].join(':')
}

// Texturspeicher zählt, wer welche Textur gerade zeigt.
export class Texturspeicher {
  private belege = new Map<string, Texturbeleg>()

  // gezaehlteRunde steigt bei jedem Räumen.
  //
  // # Wofür sie da ist
  //
  // **Weil der Beleg allein Regel 2 nicht hält.** `setze` prüfte, ob der
  // Beleg noch steht — und ein Beleg kann zwischenzeitlich **neu**
  // entstanden sein: räumen, dieselbe Karte wieder anzeigen, und die
  // Antwort von vorhin trägt den Schlüssel von jetzt. Nach einem
  // Profilwechsel wäre das die Karte eines anderen Profils unter derselben
  // Kennung.
  //
  // **Eine Zahl, die beim Räumen steigt, trennt die beiden Fälle.**
  private gezaehlteRunde = 0

  // runde gibt den Stand, zu dem eine Anfrage gehört.
  runde(): number {
    return this.gezaehlteRunde
  }

  // kennt sagt, ob zu diesem Schlüssel schon etwas läuft oder liegt.
  kennt(schluessel: string): boolean {
    return this.belege.has(schluessel)
  }

  // merke legt einen Schlüssel an, **ohne** ihn zu belegen.
  //
  // **Für das Vorladen**: das Bild wird geholt, aber niemand zeigt es. Ein
  // Beleg mit der Zahl eins bliebe für immer stehen, weil kein `gib` dazu
  // käme.
  merke(schluessel: string): void {
    if (!this.belege.has(schluessel)) {
      this.belege.set(schluessel, { zahl: 0 })
    }
  }

  // belegen meldet ein Ding an und sagt, ob die Textur **neu** ist.
  //
  // **Der Rückgabewert ist die ganze Ersparnis**: nur beim ersten Anmelden
  // wird geholt. Beim zweiten bis siebten steht sie schon da oder ist
  // unterwegs.
  belegen(schluessel: string): boolean {
    const beleg = this.belege.get(schluessel)

    if (beleg === undefined) {
      this.belege.set(schluessel, { zahl: 1 })

      return true
    }

    beleg.zahl += 1

    // **Auch ein vorgeladener Schlüssel wird nicht noch einmal geholt.**
    // Er trägt die Zahl null und trotzdem einen Eintrag: das Bild ist
    // unterwegs oder liegt schon da.
    return false
  }

  // gib meldet ein Ding ab und wirft die Textur weg, wenn es das letzte war.
  gib(schluessel: string): void {
    const beleg = this.belege.get(schluessel)
    if (beleg === undefined) {
      return
    }

    beleg.zahl -= 1

    if (beleg.zahl <= 0) {
      this.belege.delete(schluessel)
    }
  }

  // setze trägt eine geholte Textur ein.
  //
  // **Nur, wenn der Beleg noch steht.** Wer eine Karte wegräumt, während ihr
  // Bild unterwegs ist, bekommt es nicht nachträglich eingetragen — sonst
  // hielte der Speicher eine Textur, die niemand zeigt, und niemand gäbe sie
  // je zurück.
  setze(schluessel: string, textur: Textur, runde: number): boolean {
    // **Eine Antwort aus einer alten Runde wird verworfen** (§7, Regel 2).
    if (runde !== this.gezaehlteRunde) {
      return false
    }

    const beleg = this.belege.get(schluessel)
    if (beleg === undefined) {
      return false
    }

    beleg.textur = textur

    return true
  }

  // hole gibt die Textur zu einem Schlüssel — undefined, solange keine da ist.
  hole(schluessel: string): Textur | undefined {
    return this.belege.get(schluessel)?.textur
  }

  // zahl sagt, wie viele Dinge eine Textur zeigen — null, wenn keines.
  //
  // **Für den Lauf über die Zählung.** Ohne ihn wäre ein Speicher, der nicht
  // mehr wegwirft, von außen nicht von einem zu unterscheiden, der es tut.
  zahl(schluessel: string): number {
    return this.belege.get(schluessel)?.zahl ?? 0
  }

  // groesse ist die Zahl der gehaltenen Texturen.
  groesse(): number {
    return this.belege.size
  }

  // raeume wirft alles weg — für den Profilwechsel und das Ende einer Partie.
  //
  // **Weil die Bilder zu einem Profil gehören.** Wer wechselt, sieht die
  // Karten eines anderen; ein Speicher, der sie noch hielte, zeigte sie
  // weiter.
  raeume(): void {
    this.belege.clear()
    this.gezaehlteRunde += 1
  }
}
