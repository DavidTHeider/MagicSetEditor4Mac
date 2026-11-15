# 0131 — Ein Abkömmling fragt nach seiner Wurzel, in der Ablage

**Stand:** entschieden am 2026-09-23
(#1815) —
schreibt [0066](0066-nutzer-profile-und-die-vier-entscheidungen.md),
Entscheidung 2, für die Abkömmlinge fort.

## Was galt

0066 trennt Profile ohne RLS: zehn Wurzeln tragen `profil_id`, jede Abfrage
auf eine Wurzel klemmt darauf, und ein Wächter hält das. **Die Abkömmlinge —
Karten, Kartenwerte, Deckplätze, Runden, Partieereignisse — erben ihr Profil
über den Fremdschlüssel** und stehen ausdrücklich nicht unter dem Wächter:
eine Zeile ist nur so sicher wie die Abfrage, die ihre Kennung geliefert hat.

**Geliefert hat sie oft der Aufrufer.** Am 2026-09-23 schrieb ein Nutzer im
eigenen Profil an der Kommandozeile mit `mse feld 1 1 name …` den Kartennamen
eines anderen um. Nachgezählt an einer Vorrichtung, in der Profil 2 von jeder
Art ein Objekt besitzt: **37 von 55** Rufen der Karten- und Setablagen gingen
aus Profil 1 durch, dazu 8 der Deck- und Sammlungsablagen, 14 der Turnier-,
Match- und Partieablagen und die Gruppenschreibung der KI.

## Was gilt

**Wer eine Kennung eines Abkömmlings annimmt, fragt zuerst, ob seine Wurzel
dem gezeigten Profil gehört — beim Schreiben im selben Zug.** Die Fragen
stehen an **einer** Stelle
([`internal/storage/besitz.go`](../../internal/storage/besitz.go)), je Art
eine indizierte Abfrage.

- **Eine fremde Kennung ist „nicht gefunden"** — dieselbe Antwort wie für eine,
  die es nicht gibt. Eine andere verriete, dass es sie woanders gibt.
- **Wo eine Liste die Antwort ist** — eine Suche, ein Nachschlagen über
  mehrere Kennungen —, ist sie für eine fremde Kennung **leer**, wie für eine
  fehlende.
- **Auch eine Verbindung bleibt im Profil**: ein Deckplatz nimmt keine fremde
  Karte auf, eine Zuordnung keinen Platz aus einem anderen Deck, eine Paarung
  keinen Teilnehmer eines anderen Turniers, eine Matchbindung keine fremde
  Partie. Der Bestand hält keine dieser Bedingungen selbst.
- **Eine Leseabfrage aus einer Zeile klemmt in sich selbst** (`NameZurKarte`,
  `FeldZurKarte`, `KarteLesen`): eine Abfrage statt zwei, und kein Zweig, der
  nur unter einem Wettlauf zu betreten wäre.

## Warum in der Ablage

- **Nicht an den Flächen.** Fenster, Kommandozeile, MCP und Dienst rufen
  dieselben Ablagen; eine Prüfung an jeder wären vier Stellen, an denen sie
  fehlen kann. Gefehlt hat sie an einer — der Kommandozeile —, und dort sah
  es niemand.
- **Nicht im Dienst.** Der Dienst prüfte an manchen Wegen schon (die
  Matchbindung holte ihre Partie, die Zuordnung ihr Deck) und an anderen
  nicht (die Meldung eines Ergebnisses, die Gruppenschreibung). Eine Regel,
  die jeder Anwendungsfall selbst einlöst, löst der nächste nicht ein.
- **Nicht über eine `profil_id` in jeder Tafel.** 0066 hat das schon verworfen:
  eine zweite Wahrheit neben dem Fremdschlüssel läuft auseinander, sobald sich
  eine Karte bewegt. Die Frage nach der Wurzel kostet eine indizierte Abfrage.

## Was es hält

- **Die Läufe** `internal/storage/besitz*_intern_test.go` rufen jede Methode
  aus Profil 1 mit den Kennungen von Profil 2 und zählen Profil 2 danach nach;
  die Gegenprobe ruft als Eigner.
- **Zwei Wächter**, und sie halten Verschiedenes:
  `TestJedeAbkoemmlingsmethodeWirdAusEinemFremdenProfilGerufen` verlangt für
  jede Methode der fünfzehn Abkömmlingsablagen einen dieser Läufe;
  `TestJedeAblagemethodeFragtNachDemProfil` verlangt von **jeder** exportierten
  Methode der Ablage, dass sie eine Klemme erreicht — oder begründet
  ausgenommen ist.
- **Der Lauf an der Fläche**: `TestEinNutzerSchreibtKeineFremdeKarteUm`.

## Was es nicht hält

- **Den Sinn einer Klemme.** Der zweite Wächter sieht, dass gefragt wird,
  nicht wonach. Die Turnierablage kam an ihm vorbei, weil sie `ProfilAus` für
  die Fassung des Turniers rief, während der Teilnehmer, um den es ging, aus
  einem fremden Turnier ausstieg. Deshalb gibt es den ersten.
- **Kennungen, die eine Bindung ausdrücklich für ein anderes Profil
  annimmt** — Zugriffstoken, Vorgaben, Freigaben. Dort ist die Frage eine
  andere: nicht „gehört es dem gezeigten Profil", sondern „welche Rolle hat
  der Mensch am **Zielprofil**". Sie steht in der Anwendung
  (`internal/app/zielprofil.go`) und unter einem eigenen Wächter
  (`TestJedeBindungMitZielprofilFragtNachDerRolle`) —
  #1817,
  gefunden beim Bau dieses Wächters und im selben Zug behoben.
- **Dateien neben einem Set.** Die Symbolzeichnung liegt auf der Platte und
  nicht im Bestand; die Ablage sieht sie nicht. Ihre drei Wege — lesen,
  ablegen, nachzeichnen — fragen deshalb an ihrer Naht (`SetDaSein`), und MCP
  und Dienst reichen dafür den Kontext des Rufs durch: bis dahin riefen sie die
  Werkstatt mit `context.Background()`, also **ohne** Profil. Gefunden hat es
  die Symbolpalette, die seit der Frage in `KarteLesen` an beiden Flächen
  scheiterte. **Kein Wächter hält eine neue Dateiablage dieser Art**; die Läufe
  `TestDasSymbolEinesFremdenSetsTrifftNichts` und
  `TestDasFensterFasstKeinFremdesSymbolAn` halten die bestehende.
- **Ein fremdes Werkzeug** — dieselbe Grenze wie in 0066.

## Verworfen

- **Eine Sicht je Profil.** SQLite kennt keine Sicht mit Parameter; eine je
  Profil hieße, Sichten beim Anlegen eines Profils zu erzeugen, und jede
  Abfrage müsste ihren Namen zusammensetzen.
- **Die Prüfung nur an den schreibenden Wegen.** Ein Lesen über eine fremde
  Kennung geht an einem zweiten Faktor vorbei, der das Hineinwechseln schützt
  (`docs/AUTORISIERUNG.md`, Bedrohungsmodell).
