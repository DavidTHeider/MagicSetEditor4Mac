# 0080 — Maschinenzugänge, Tokenrotation und widerrufbare MCP-Sitzungen

**Stand:** angenommen
**Datum:** 2026-09-08
**Anlass:** Werkstück O2 von Phase 5.2
(#1426) —
Maschinenzugänge sollen begrenzt, widerrufbar und ohne wieder auslesbares
Geheimnis bleiben; ein Widerruf soll **in der laufenden Sitzung** wirken.

## Lage

**Was am 2026-09-08 galt, ist nachgelesen** (Stand `c95d625b`):

| Frage | Antwort |
| --- | --- |
| Was ein Token ist | Kennung, Profil, Bezeichnung, Scope `lesen`/`schreiben`, Ablauf; im Bestand nur SHA-256 des Geheimnisses ([0068](0068-technische-profile-und-zugriffstoken.md)) |
| Wo es geprüft wird | am Dienst je Anfrage (`Tokenwache`), sonst nirgends |
| Wie `mse-mcp` sich ausweist | **gar nicht**: der Prozess läuft über stdio im gezeigten Profil — lokaler Vertrauensbetrieb |
| Was ein Wechsel des Geheimnisses kostet | ein neues Token und das Löschen des alten: dazwischen eine Lücke oder zwei Zugänge, und wer den zweiten Schritt vergisst, hat sie für immer |
| Was niemand sieht | ob ein Token noch benutzt wird, und für wen es arbeitet |

**Und was der Anhang des Issues falsch annahm**, steht dort selbst: eine
Startauthentisierung genügt nicht. Ein Token, das beim Start geprüft und dann
zurückgezogen wird, trüge die Sitzung bis zu ihrem Ende.

## Entscheidung

### Rotation am selben Datensatz, mit genau einem Vorgänger

**Ein Token wechselt sein Geheimnis und bleibt dasselbe** — Kennung,
Bezeichnung, Platz in jeder Liste. Das bisherige rückt in die Spalte des
Vorgängers und gilt noch **eine Stunde** ([domain.Tokenuebergang]): lang genug
für einen laufenden Massenlauf und für die Hand, die das neue Geheimnis in die
Einrichtung des Gegenübers trägt; kurz genug, dass es keine zweite Tür bleibt.
Eine zweite Rotation wirft den Vorgänger sofort weg; ein Widerruf sperrt beide.
Wanderung 0048 legt `principal`, `letzte_nutzung`, `vorgaenger_gehuetet` und
`vorgaenger_bis` an — die alten Zeilen bekommen ihre Bezeichnung als Principal,
denn das ist das Einzige, was nachweislich ein Mensch getippt hat.

**Der Principal ist Pflicht**: `Bezeichnung` sagt, wofür ein Token da ist,
`Principal`, wer dahintersteht — ein Mensch oder eine Maschine. Er ist ein
Name und kein Fremdschlüssel, weil ein Baurechner kein Nutzer dieser
Anwendung ist.

### Vier Ablaufwahlen, gerechnet in Go

Ein Tag, dreißig Tage, ein Jahr, nie — kalendarisch gerechnet aus der Uhr des
Dienstes, damit alle vier Flächen dieselben Zeitpunkte nennen. »Nie« steht
auch am Fenster: es zu verschweigen triebe niemanden zu kürzeren Fristen,
sondern an die Kommandozeile.

### Der MCP-Prozess weist sich aus — und jeder Ruf prüft

`mse-mcp` liest sein Token aus einer **Datei mit den Rechten 0600**
(`MSE_TOKEN_DATEI`) oder aus `MSE_TOKEN`, **nie aus einem Argument**:
Argumente stehen in der Prozessliste. Eine Datei, die andere lesen dürfen,
wird abgewiesen, nicht gelesen.

Mit Token setzt der Server eine **Wache je Ruf** (`mcp.Server.MitWache`):
vor jedem Werkzeug wird das Token gegen den Bestand geprüft — Widerruf,
Ablauf, Umfang gegen die Art des Werkzeugs, und das Profil des Tokens
überschreibt das gezeigte. Der Widerruf wirkt damit beim **nächsten** Ruf
derselben Sitzung; belegt an den gebauten Programmen
(`internal/flaechen/tokenrotation_test.go`). Jede angenommene Prüfung vermerkt
die letzte Nutzung, ohne die Fassung zu heben.

**Ohne Token bleibt der Altbetrieb** — das gezeigte Profil, ohne Wache —,
denn dieser Prozess lief bis hierher so, und er läuft über stdio auf
demselben Rechner. Der Vertrag steht in `docs/ADMINISTRATION.md`.

### Was MCP nicht bekommt

**Keine Rotation** — sie gäbe ein Geheimnis zurück, und das stünde im
Verlauf eines Sprachmodells: derselbe Grund, aus dem `KontoEintragen` fehlt.
Der Sammelwiderruf steht dort, auch wenn er die eigene Sitzung trifft: ein
Zugang, den ein Modell nicht schließen kann, wäre der gefährlichere.

## Verworfen

- **Ein zweites Token je Rotation.** Zwei Zeilen sind zwei Buchführungen,
  und „Wiederholung erzeugt keine unbemerkten Extra-Token" ließe sich nur
  über Aufräumen halten, das jemand vergisst.
- **Eine Kette von Vorgängern.** Nach fünf Rotationen stünde ein Zugang
  offen, den niemand mehr auf dem Zettel hat.
- **Das Token als Argument von `mse-mcp`.** Bequem für eine
  Client-Konfiguration — und für jeden, der `ps` ruft.
- **Nur beim Start prüfen.** Das war die Annahme des Anhangs; sie hält den
  Widerruf bis zum Neustart zurück.
- **Die letzte Nutzung nur gedrosselt vermerken.** Ein `UPDATE` je
  angenommenem Ruf ist billig; eine Drossel wäre eine zweite Wahrheit
  darüber, wann etwas zuletzt geschah.

## Folgen

- **Bestehende Aufrufer, die ein Token anlegen, nennen einen Principal** —
  an der Zeile, über MCP und am Dienst. Ein leerer wird abgewiesen.
- **Ein Token wird beim Prüfen geschrieben** (die Nutzung) — die Wache am
  Dienst und die Wache im MCP-Prozess laufen deshalb über die schreibende
  Verbindung.
- **Zwei neue Fähigkeiten an den Flächen** und eine erweiterte; die
  Rotation fehlt bei MCP begründet (`docs/FAEHIGKEITEN.md`).
- **Die Einrichtung eines Clients steht am Fenster** (Zugänge, Aufklappkasten)
  und in `docs/ADMINISTRATION.md`.
