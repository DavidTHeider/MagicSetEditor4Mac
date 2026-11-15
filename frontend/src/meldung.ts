// Die Meldung aus einer Ursache — **in der Sprache der Sitzung**.
//
// # Warum das an einer Stelle steht
//
// Der Ausdruck stand dreimal im Baum — im Ladehaken, im Karteneditor und in der
// Massenänderung —, und jede Kopie brachte ihren eigenen Zweig mit, den
// jemand abdecken musste. Abgedeckt wurde er dort, wo er am leichtesten
// herzustellen war: in einem Test, der eine Ablehnung ohne `Error` erfindet.
//
// **Wails lehnt nie ohne `Error` ab.** `calls.js` in v2.13.0 macht aus jeder
// Meldung `new Error(message.error)`, und auch der Zeitablauf lehnt mit einem
// `Error` ab. Der Zweig ist also nicht der Weg aus der Bindung, sondern der
// Rückfall für eine Ursache aus unserem eigenen Code — und der gehört einmal
// geprüft, nicht dreimal nachgestellt.
//
// # Warum ein Katalog dazukommt
//
// #133:
// `FehlerDTO` trug einen `schluessel`, und **niemand las ihn**. Sichtbar war
// der deutsche Satz aus Go — eine englische Sitzung zeigte deutsche
// Fehlermeldungen, und die englischen Kataloge waren produktiv nirgends
// verdrahtet. Der Vertragstest hält beide Kataloge gegen Go; dass sie jemand
// **liest**, hält er nicht.
import { Bindungsfehler } from './api'
import { Katalogfehler } from './katalogfehler'
import { einsetzen } from './sprache/einsetzen'
import { gemeldeteKennung, istFremd } from './unerwartet'
import type { Katalog } from './sprache/kataloge'

// **Weitergereicht**, damit eine Seite nur eine Anlaufstelle braucht: wer eine
// Meldung wirft, wirft sie in derselben Datei, aus der er sie liest.
export { Katalogfehler }

// meldungAus gibt den Satz, der einer Ursache entspricht.
//
// Vier Wege, und jeder hat seinen Grund:
//
//  1. **Katalogfehler** — im Frontend entstanden, trägt seinen Schlüssel.
//  2. **Bindungsfehler** — aus Go, trägt einen Schlüssel, den deutschen Satz
//     **und die Werte** für dessen Platzhalter (#940). Übersetzt wird über den
//     Schlüssel; der Satz bleibt der Rückfall für einen, den der Katalog nicht
//     kennt (§9). Eine unübersetzte Meldung ist besser als eine leere.
//  3. **Fehler** — aus Wails oder dem Browser, ohne Schlüssel. Ihn zu übersetzen
//     hieße zu raten, was er bedeutet.
//  4. **Alles andere** — als Text.
export function meldungAus(ursache: unknown, texte: Katalog): string {
  if (ursache instanceof Katalogfehler) {
    return einsetzen(
      texte.allgemein.meldungen[ursache.schluessel],
      ursache.werte,
    )
  }

  if (ursache instanceof Bindungsfehler) {
    const uebersetzt = ausKatalog(texte, ursache.schluessel)

    // **Platzhalter werden auch hier eingesetzt** (#940). Der Rückfall bleibt
    // die Meldung aus Go — sie trägt dieselben Angaben schon eingesetzt.
    return uebersetzt === undefined
      ? ursache.message
      : einsetzen(uebersetzt, mitSkriptsatz(ursache.werte, texte))
  }

  // **Ein Fehler mit eigenem Typ zeigt seinen Satz nicht** (F03, #2342) —
  // er kommt aus der Laufzeit; im Fenster steht der Satz für unerwartete
  // Fehler mit Kennung, im Protokoll der Rohtext ([gemeldeteKennung]).
  if (istFremd(ursache)) {
    return einsetzen(texte.fehler.unbekanntMitVorgang, {
      vorgang: gemeldeteKennung(ursache),
    })
  }

  return ursache instanceof Error ? ursache.message : String(ursache)
}

// vorgangAus gibt die Vorgangskennung eines Fehlers — oder nichts. Unter ihr
// stehen im Protokoll alle Zeilen dieses Vorgangs (NET-02, S13.NAT-04, #1706).
export function vorgangAus(ursache: unknown): string | undefined {
  if (istFremd(ursache)) {
    return gemeldeteKennung(ursache)
  }

  if (!(ursache instanceof Bindungsfehler)) {
    return undefined
  }

  const vorgang = ursache.werte.vorgang

  return vorgang === undefined || vorgang === '' ? undefined : vorgang
}

// istKonflikt sagt, ob eine Ursache ein Konflikt über den Stand ist —
// UX-08.AK-02 (#1680).
//
// **Am Code und nicht am Satz** (§8): wer „zwischenzeitlich" im Text suchte,
// fände es in der nächsten Sprache nicht mehr.
export function istKonflikt(ursache: unknown): boolean {
  return ursache instanceof Bindungsfehler && ursache.code === 'konflikt'
}

// istAbgebrochen sagt, ob eine Ursache ein gewollter Abbruch ist — FMT-03:
// wer einen Lauf anhält, bekommt keine Fehlermeldung dafür.
export function istAbgebrochen(ursache: unknown): boolean {
  return ursache instanceof Bindungsfehler && ursache.code === 'abgebrochen'
}

// ausKatalog sucht einen Go-Schlüssel im Fehlerkatalog.
//
// **Über eine Nachschlagetabelle und nicht über einen Typ.** Der Schlüssel kommt
// aus Go und ist zur Bauzeit eine gewöhnliche Zeichenkette; ein Typ, der ihn
// einschränkte, behauptete eine Zusage, die erst der Vertragstest in
// `internal/app/fehlerschluessel_test.go` einlöst.
export function ausKatalog(
  texte: Katalog,
  schluessel: string,
): string | undefined {
  // **Nur die flachen Einträge.** `fehler.skript` ist ein Unterkatalog — die
  // Sätze des Zerteilers (#1178) —, und ein Fehlerschlüssel zeigt nie
  // dorthin. Wer ihn mitnähme, bekäme ein Objekt statt eines Satzes.
  const eintraege = texte.fehler as unknown as Record<string, unknown>
  const eintrag = eintraege[schluessel]

  return typeof eintrag === 'string' ? eintrag : undefined
}

// mitSkriptsatz ersetzt `{meldung}` durch den übersetzten Satz des Zerteilers.
//
// **Nur wo einer dasteht.** Jeder andere Fehler geht unverändert durch; ein
// `meldung` ohne `skriptschluessel` ist keine Meldung des Zerteilers.
function mitSkriptsatz(
  werte: Record<string, string>,
  texte: Katalog,
): Record<string, string> {
  const schluessel = werte.skriptschluessel
  if (!schluessel) {
    return werte
  }

  return { ...werte, meldung: skriptsatz(werte, schluessel, texte) }
}

// skriptsatz gibt den Satz zu einer Meldung des Zerteilers —
// #1178.
//
// # Warum ein zweiter Nachschlag
//
// Der Fehler selbst heißt `skriptFehler` und lautet „Der Ausdruck ist nicht
// auswertbar: {meldung}". Was in `{meldung}` steht, kommt aus
// `internal/msescript` und ist **eine eigene Meldung mit eigenem Schlüssel**.
//
// # Der Rückfall ist der deutsche Satz
//
// Für die 48 Sätze der Standardbibliothek gibt es keinen Schlüssel; sie kommen
// als `durchgereicht` und tragen ihren Satz als Wert. Und ein Schlüssel, den
// dieser Katalog noch nicht kennt, ist besser unübersetzt als leer (§9).
// **Ohne eigene Wache auf den Schlüssel.** [mitSkriptsatz] ruft sie nur, wenn
// einer dasteht; eine zweite Prüfung hier wäre ein Zweig, den kein Aufruf
// betritt (§3).
function skriptsatz(
  werte: Record<string, string>,
  schluessel: string,
  texte: Katalog,
): string {
  const rueckfall = werte.meldung ?? ''

  const saetze = texte.fehler.skript as Record<string, string>
  const satz = saetze[schluessel]
  if (satz === undefined) {
    return rueckfall
  }

  // **Die Teile stehen einzeln da** — `nach {0} fehlt {1}`. Sie im Satz zu
  // lassen hieße, ihn nicht übersetzen zu können.
  return satz.replace(/\{(\d+)\}/g, (ganz: string, stelle: string) => {
    return werte['skriptwert' + stelle] ?? ganz
  })
}

// befundsatz gibt den Satz zu einem Befund der Paketprüfung in der Sprache
// des Fensters — `undefined`, wenn kein Katalog den Schlüssel kennt; dann
// bleibt der deutsche Satz aus Go stehen (§9).
//
// **Zwei Herkünfte, ein Satz**: ein Syntaxfehler trägt einen Schlüssel des
// Zerteilers (`fehler.skript`), ein Feldartbefund einen der Werkstatt
// (`internal/service/paketpruefung.go`), und dessen Satz steht unter
// `seiten.vorlagen`. Die Werte kommen als Liste — so antwortet die Prüfung
// einer Paketdatei.
//
// **An einer Stelle und nicht an zweien**: der Befund steht unter dem Editor
// (`Vorlagenbefund`) und als Grund auf der Startseite (`Aufmerksamkeit`,
// #1963).
// Bis dahin stand er nur unter dem Editor übersetzt da.
export function befundsatz(
  schluessel: string,
  werte: readonly string[],
  texte: Katalog,
): string | undefined {
  const { feldartFehlt, feldartUnbekannt } = texte.seiten.vorlagen
  const werkstatt = new Map([
    ['feldartFehlt', feldartFehlt],
    ['feldartUnbekannt', feldartUnbekannt],
  ]).get(schluessel)

  if (werkstatt !== undefined) {
    return einsetzen(
      werkstatt,
      Object.fromEntries(werte.map((wert, stelle) => [stelle, wert])),
    )
  }

  if (!Object.hasOwn(texte.fehler.skript, schluessel)) {
    return undefined
  }

  const benannt = Object.fromEntries(
    werte.map((wert, stelle) => ['skriptwert' + String(stelle), wert]),
  )

  return skriptsatz(benannt, schluessel, texte)
}

// meldungAusAngabe gibt den Satz zu einem Fehler, der über die Ereignisnaht
// kam.
//
// **Dieselbe Übersetzung wie bei einer abgelehnten Bindung** — nur die Quelle
// ist eine andere: Go meldet ein Ereignis samt Fehlerverbund, statt einen Ruf
// abzulehnen (#891).
//
// **Der Rückfall bleibt die Meldung aus Go**, wie in [meldungAus]: eine
// unübersetzte Meldung ist besser als eine leere (§9).
export function meldungAusAngabe(
  angabe: {
    schluessel: string
    meldung: string
    werte: Record<string, string> | null
  },
  texte: Katalog,
): string {
  const uebersetzt = ausKatalog(texte, angabe.schluessel)

  return uebersetzt === undefined
    ? angabe.meldung
    : einsetzen(uebersetzt, mitSkriptsatz(angabe.werte ?? {}, texte))
}

// verursachendesFeld sagt, an welchem Feld eine Meldung erscheint.
//
// **Sagt die Bindung eines, gilt das** (§6): eine Domänenprüfung kennt das
// Feld, an dem sie gescheitert ist, und das muss nicht das eben bearbeitete
// sein. Sonst das, an dem gerade gearbeitet wurde.
//
// **Hier und nicht in der Seite**, weil drei Wege dieselbe Frage stellen — das
// Schreiben, das Bildsetzen und das Ablegen. Drei Ausdrücke wären drei
// Stellen, an denen die Antwort auseinanderlaufen kann.
export function verursachendesFeld(ursache: unknown, feld: string): string {
  return ursache instanceof Bindungsfehler ? (ursache.feld ?? feld) : feld
}

// feldbefund gibt die Meldung einer Ursache, **wenn sie zu diesem Feld
// gehört** — sonst nichts. ENG-05.AK-03
// (#1730):
// ein doppelter Name steht am Namensfeld und nicht über der Seite (§6).
export function feldbefund(
  ursache: unknown,
  feld: string,
  texte: Katalog,
): string | undefined {
  return ursache instanceof Bindungsfehler && ursache.feld === feld
    ? meldungAus(ursache, texte)
    : undefined
}
