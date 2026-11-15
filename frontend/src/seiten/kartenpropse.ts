// **Was die Kartenlasche von außen bekommt**
// (#288).
//
// **Eine eigene Datei**, weil `Karten.tsx` mit `karteGewaehlt` über die
// Zeilengrenze aus §2 wuchs. Der Schnitt folgt der Frage: dort, **was** die
// Lasche tut; hier, **was** sie dafür braucht — und warum jede dieser Angaben
// da ist.

export interface KartenEigenschaften {
  // setID ist **gesetzt** — der Arbeitsplatz rendert diese Lasche nur mit
  // offenem Set und zeigt sonst seinen eigenen leeren Zustand
  // (#288).
  // Ein Wächter „kein Set gewählt" wäre hier ein Zweig, den kein Aufruf
  // betreten kann (§3).
  setID: string

  // karteID ist die Vorgabe von außen — die Karte, die ein Treffer
  // aufgeschlagen hat. **Vorgabe und nicht Steuerung**: wer danach in der Liste
  // eine andere wählt, behält sie, bis eine neue Vorgabe kommt.
  karteID: string | undefined

  // karteGewaehlt sagt dem Fenster, welche Karte jetzt offen ist.
  //
  // # Warum die Seite nach oben meldet
  //
  // Weil andere Laschen dieselbe Karte meinen. Die **Skriptkonsole** wertet
  // einen Ausdruck in ihrem Kontext aus und holt sich die Kennung aus dem
  // Zustand von `App` — und dorthin sagte bis zum 2026-08-25 niemand Bescheid.
  // Wer eine Karte in der Liste anklickte und dann die Konsole aufschlug, las
  // „Dafür muss eine Karte aufgeschlagen sein", während sie daneben im Editor
  // stand. Über die **Suche** geöffnet ging es, über die Liste nicht
  // (#892).
  //
  // **Gefunden bei der Sichtprüfung im laufenden Fenster**, nicht von einem
  // Test — dieselbe Gestalt wie
  // #1013.
  karteGewaehlt: (karteID: string) => void

  // symbolZeichnen führt zum Symboleditor für ein Kartenfeld der Art `symbol`
  // (#1339).
  //
  // **Wahlfrei**: ohne ihn steht kein Knopf da — einer, der nichts tut, wäre
  // schlimmer als keiner.
  symbolZeichnen?: ((karteID: string, feld: string) => void) | undefined

  // verlaufsstand steigt nach jeder geglückten Rücknahme (#52).
  //
  // **Er steht in den Ladeabhängigkeiten**, nicht in einem Rückruf, den der
  // Rahmen hielte: was sich geändert hat, steht im Bestand, und gelesen wird es
  // von dort. Ein ⌘Z, das die Anzeige selbst zurechtrückte, führte eine zweite
  // Fassung derselben Karte.
  verlaufsstand: number

  verlaufsfehler: string | undefined

  // zuruecknehmen nimmt einen genannten Schritt zurück — für die Verlaufsliste
  // (U6, #1417); wahlfrei, siehe `Seiteninhalt`.
  zuruecknehmen?: ((vorgang: string) => void) | undefined

  // bestandGeaendert meldet dem Fenster, dass sich am Bestand etwas geändert
  // hat — eine neue Karte etwa.
  //
  // **Der Kopf des Arbeitsplatzes liest es**: er nennt die Kartenzahl, und
  // ohne diese Meldung stand dort „1 Karten", während daneben zwei lagen
  // (#562).
  bestandGeaendert: () => void
}
