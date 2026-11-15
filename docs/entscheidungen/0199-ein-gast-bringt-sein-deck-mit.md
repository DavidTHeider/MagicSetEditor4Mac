# 0199 — Ein eingeladener Gast bringt sein Deck mit; eine Leihe nur bei ihrer Kennung

**Stand:** angenommen am 2026-10-18 · #2377 (P9b, Befund C-11). Baut auf [0198](0198-kartenbilder-am-fremden-tisch.md) auf. Umgesetzt in `internal/domain/platznahme.go`, `internal/domain/sichtbauen.go`, `internal/service/partieplatznahme.go`, `internal/service/partiewirtplatz.go`, `internal/service/deckleiheamtisch.go`, `internal/service/partierundebild.go`, `internal/service/gastbildantwort.go`, `internal/app/partiemitbringen.go`.

## Lage

Bis P9a spielte ein Gast immer ein Deck **aus dem Bestand des Wirts**: der
Wirt legte die Partie mit allen Decks an, und der Gast setzte sich an einen
freien Sitz. Das Abnahmekriterium von P9 verlangt, dass der Gast **sein
eigenes oder ein angenommenes Leihdeck** mitbringt — und dass ein Gast, der
das nicht kann, **vor dem Anfang** erklärt abgewiesen wird.

## Entscheidung

1. **Ein Sitz kann eingeladen sein** (`Sitz.Eingeladen`, eine Peerkennung):
   er entsteht ohne Deck und ohne Bibliothek und **wartet**. Nur ein solcher
   Sitz darf ohne Deck dastehen; der Wirt wird nie eingeladen. **Wen er
   eingeladen hat, sieht nur der Wirt** — alle anderen lesen nur, dass der
   Sitz wartet. Die Freigabe gibt einen eingeladenen Sitz **nur** dem
   Eingeladenen.
2. **Der Gast bringt beim ersten Rückkehrwunsch mit, womit er sich setzt**
   (`Wiederaufnahmewunsch.Mitgebracht`), genau eines von beiden:
   - **ein eigenes Deck** als Abzug, wie am eigenen Tisch eingefroren, samt
     seinem Spiel. Die Quelle setzt der Wirt — sie ist der Peer am anderen
     Ende der Leitung —, und den Abdruck rechnet er nach;
   - **eine Leihe dieses Wirts**, nur bei ihrer Kennung. Der Wirt nimmt dann
     **seinen eigenen**, ausgestellten Abzug und prüft wie beim Vorlegen:
     von ihm ausgestellt, nicht zurückgenommen, diesem Gast für diese Partie,
     jetzt.
3. **Wartet der Sitz und bringt der Gast nichts mit** — ein Gast mit älterer
   Fassung kennt das Feld nicht —, antwortet der Wirt mit einem **Abbruch und
   seinem Grund**, bevor jemand gespielt hat. Ebenso bei einem Deck eines
   anderen Spiels, einer ungültigen Leihe oder einem zweiten Hinsetzen an
   einen fremden Sitz.
4. **Das Hinsetzen ist ein Eintrag im Strom** (`sitz.platznehmen`, öffentlich
   nur Sitz, Kartenzahl und Abdruck) — kein Kernbefehl, denn kein Sitz stellt
   ihn, und ein Gast, der ihn als Wunsch schicken könnte, setzte sich mit
   einem Deck seiner Wahl an einen Sitz seiner Wahl. **Einmal je Sitz**: ein
   zweites Mal ist dieselbe Antwort.
5. **Wer die Karte zeichnet, folgt der Quelle** (0198) — jetzt in **beiden**
   Richtungen: der Gast zeichnet sein eigenes Deck selbst und beantwortet die
   Bildfragen des Wirts (nur für Objekte, die seine letzte Sicht offen und mit
   eigener Zeile nennt); fragt ein **dritter** Gast nach einer solchen Karte,
   reicht der Wirt die Frage an den Eigner weiter, mit kürzerer Frist als die
   des Fragenden, und nie an den Fragenden selbst. Am Fenster des Wirts fragt
   `Fremdkartenbild` den Eigner über die Runde — **mit der Partie**, zu deren
   Tisch das Bild gehört (Zweitblick über Phase 7, Befund 4). **Gefragt wird
   nur ein Eigner, dessen eigener Bestand die Zeile hält** (`zeichnetSelbst`,
   Befund 6), und **das Weiterreichen wartet nicht in der Empfangsschleife**
   (Befund 5): fragten zwei Gäste zugleich nach der Karte des jeweils anderen,
   warteten sonst beide Schleifen bis zur Frist.
6. **Zwei Schreiber an derselben Partie warten aufeinander** (`Partieleitung`):
   das Fenster des Wirts und die Schleife jedes Gastes schreiben im selben
   Prozess. Gefunden mit zwei Anwendungen — ohne das Schloss scheiterte der
   zweite an der Fassungsprüfung, und der Konflikt beendete die Sitzung des
   Gastes.

## Verworfen

- **Den Abzug einer Leihe mitschicken**: ein Abzug, der behauptete, aus dem
  Bestand des Wirts zu stammen, brächte den Wirt dazu, beliebige Zeilen
  seines Bestands zu zeichnen und dem Gast zu zeigen.
- **Ein Kernbefehl „Platz nehmen"**: siehe Punkt 4.
- **Die Partie erst anlegen, wenn der Gast da ist**: der Wirt hätte keinen
  Tisch zu öffnen, und die Freigabe keine Partie, die sie nennt.

## Grenzen

- **Eine Leihe eines Dritten** (weder vom Wirt noch vom Gast) kann an diesem
  Tisch niemand zeichnen; ihre Karten stehen ohne Bild da.
- **Der Wirt prüft die Gestalt eines mitgebrachten Abzugs, nicht seine
  Herkunft**: was ein Gast als sein Deck schickt, ist sein Deck. Das ist
  dieselbe Vertrauensgrenze wie am Küchentisch.
- **Ein abgewiesener Wunsch eines Gastes beendete die Sitzung** — älter als
  P9; seit #2478
  geht ein Nein des Kerns als `Wunschabsage` am Tischbild an den Gast, und
  die Sitzung bleibt. Was kein Urteil des Kerns ist, beendet sie weiter.
- **Ein Abzug hat höchstens tausend Plätze und Karten** (`Abzuggrenze`,
  Befund 2): je Exemplar entsteht ein Objekt, und die Grenze der Leitung
  begrenzt die Nachricht, nicht die Menge darin.

## Rückweg

Ein Wirt vor 0.9.0 liest `Mitgebracht` nicht und setzt den Gast wie bisher an
einen Sitz mit einem Deck aus seinem Bestand. Ein Gast vor 0.9.0 an einem
eingeladenen Sitz wird vor dem Anfang mit Grund abgewiesen. **Gefragt wird ein
Gast nur, wenn er es erklärt hat und sein eigener Bestand die Zeile hält**:
jeder Rückkehrwunsch dieser Fassung trägt `Kann: ["bild"]`, ein älterer Gast
schickt das Feld nicht, und der Wirt merkt es am Platz. Zweimal stand hier ein
schwächerer Satz, und zweimal hielt er nicht: bis zum Zweitblick über Phase 7
fragte der Wirt jeden Sitzinhaber, dessen Karte er nicht zeichnete (Befund 6);
danach jeden, dessen Bestand die Zeile hielt — und das hält auch ein älterer
Gast, dem der Wirt eine Leihe **von ihm** vorgelegt hat (Nachlauf, R4).
**Bestandseigentum ist kein Fähigkeitsnachweis.**

**Die Absage eines Wunsches** reist als Zusatzfeld am Tischbild: ein Gast vor
0.9.0 liest es nicht und bekommt ein frisches Bild; eine eigene Kontrollart
beendete gerade seine Sitzung (§16, erweiternd).
