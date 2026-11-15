# 0137 — Kontakte kommen als vCard

**Stand:** entschieden am 2026-09-27 · ID-04
(#1703)

## Die Frage

ID-04 verlangt, **ausgewählte** Kontakte aus Apple Kontakte **mit minimaler
Berechtigung** als lokale Kontakte zu übernehmen — mit Vorschau, Auswahl
einzelner Kontakte und Wege, Dublettenabgleich, und ohne dass eine Adresse
dabei als geprüft gilt.

## Die Entscheidung

**Kontakte kommen als vCard-Datei, die ein Mensch wählt — nicht über die
Kontakte-Berechtigung von macOS.**

In Apple Kontakte wählt man die Kontakte aus und gibt sie über *Ablage ›
Exportieren › vCard exportieren* aus (oder zieht die Auswahl auf den
Schreibtisch). Diese Anwendung liest genau diese Datei, gewählt im
Dateidialog. Der Dialog ist die Berechtigung: er gibt **eine** Datei frei,
nicht das Adressbuch.

| AK | Wie es eingelöst wird | Beleg |
| --- | --- | --- |
| **AK-01** — Datenumfang, Auswahl einzelner Kontakte | Die Seite erklärt vor dem Wählen, was gelesen wird (Name, Bild, E-Mail, Telefon) und was nicht (alles andere). **Die Auswahl entsteht zweimal**: in Kontakte, was ausgegeben wird, und in der Vorschau, welche Karte und welcher Weg übernommen wird. Das ganze Adressbuch ist nie Voraussetzung | `TestKontakteAusEinerVCardUebernehmen`, `kontakte.test.tsx` |
| **AK-02** — Name, Bild, Wege, Dublettenabgleich | `domain.KontaktAbgleichen`: zuerst über den **Ursprung** (die `UID` der Karte), dann über Name und einen gemeinsamen Weg — und dann als *vermutlich*. **Eine Karte, die es schon gibt — auch vermutlich —, lässt sich nicht noch einmal anlegen, nur ersetzen**; eine Datei, die dieselbe Karte zweimal trägt, zeigt sie einmal, und eine Auswahl legt nichts doppelt an. Die Übernahme ist **ein** Geschäft | `TestDerAbgleichFindetDenVorhandenen`, `TestEineVorhandeneKarteWirdNichtNochEinmalAngelegt`, `TestEineDateiVerdoppeltSichNichtSelbst`, `TestEineVermutungWirdNichtNeuAngelegt`, `TestEinKontaktGehtHineinUndKommtWieder` |
| **AK-03** — Ablehnung, Entzug, Änderungen am Ursprung, einmalig gegen Synchronisierung | **Es gibt keine Berechtigung, die abgelehnt oder entzogen werden könnte** — und das ist die Antwort, nicht eine Lücke. Wer den Dialog schließt, hat nichts gewählt, und nichts geschieht. **Ein Import ist einmalig**: eine Synchronisierung gibt es nicht. Wer dieselbe Karte später noch einmal einliest, sieht sie als *geändert* und entscheidet, ob er ersetzt | `TestDasEinlesenKenntAbbruchUndFehler`, die Vorschau mit Abgleich |
| **AK-04** — keine Prüfung, keine Paarung, keine Kontaktaufnahme | `domain.Kontakt` trägt **kein Feld**, das eine Prüfung behaupten könnte, und ist kein `Peerkontakt`; ein Lauf hält die Felder fest. Übernehmen schickt nichts nach draußen: der Weg kennt kein Netz | `TestEinKontaktBehauptetKeinePruefung`, `internal/schichten/netz_test.go` |

## Verworfen

| Verworfen | Weil |
| --- | --- |
| **`CNContactStore` mit der Kontakte-Berechtigung** | **Zurückgestellt, nicht widerlegt.** Sie ist eine TCC-Freigabe, und eine TCC-Freigabe lässt sich an einer ad-hoc signierten Anwendung ohne Apple-Developer-Konto (Vermerk 0003) in dieser Arbeitsumgebung nicht abnehmen — die Freigabe erteilt ein Mensch in den Systemeinstellungen. **Ob macOS dabei einen begrenzten Zugriff anbietet**, wie Apple ihn für iOS beschreibt, ist hier **nicht geprüft**; der Dateiweg braucht keine Freigabe und ist deshalb der erste. *Hier stand bis zum Zweitblick „sie gibt das ganze Adressbuch frei" — eine Behauptung ohne Beleg für macOS (Befund 4)* |
| **`CNContactPicker`** (die Auswahl ohne Berechtigung) | Er käme dem Ziel am nächsten, braucht aber ein natives Popover an einer AppKit-Ansicht. Wails 3 gibt keine her, und eine zweite native Brücke neben `internal/biometrie` wäre für einen P2-Weg ein hoher Preis. **Offen, nicht abgelehnt**: kommt eine native Ansicht dazu, ist er der nächste Schritt |
| **Eine Synchronisierung** | ID-04.AK-03 verlangt die Trennung; eine stille Aktualisierung aus einem fremden Adressbuch wäre genau die Übernahme, die §16 Regel 6 ausschließt |

## Zweitblick am 2026-09-27

**Vier Befunde, alle angenommen.** Zwei gleiche Karten ohne UID in einer
Datei kamen beide als neu durch, und eine Vermutung ließ sich als neuer
Kontakt anlegen — jetzt gleicht die Vorschau auch innerhalb der Datei ab,
die Übernahme auch innerhalb der Auswahl, und eine Vermutung ersetzt.
Einlesen und Übernehmen liefen unter zwei Vorgangsschlüsseln nebeneinander;
jetzt ist es einer. Und die Ablehnung der Kontakte-Berechtigung war stärker
formuliert als belegt (oben).

## Nachlauf am 2026-09-27

**Drei Befunde, alle behoben.** Zwei Karten, die nur einen Namen trugen,
teilten keinen Weg und kamen doppelt durch — jetzt sind **gleiche Angaben
ohne Ursprung** dieselbe Karte. Innerhalb einer Datei verschwand eine
*vermutlich* gleiche Karte aus der Vorschau, obwohl es zwei Menschen sein
können — jetzt wird nur verworfen, was **sicher** dieselbe ist (gleiche
UID oder gleiche Angaben), und in einer Auswahl dürfen zwei Vermutungen
nebeneinander stehen. Und die Eindeutigkeit der Sammlung aus 0086 hielt
für eine Zeile ohne Druckzeile nichts — Migration 0087.

## Was die Belege nicht halten

- **Ob Apple Kontakte heute genau so ausgibt wie die Proben.** Die Proben
  sind selbst geschrieben (§3) — nach RFC 2426 und der Gestalt, die
  Kontakte schreibt (Gruppen, `X-ABLabel`, `ENCODING=b`); eine echte Ausgabe
  liegt im Repositorium nicht, sie enthielte Personendaten.
- **vCard 2.1 mit Quoted-Printable** wird genannt und nicht gelesen.
- **Die Befunde des Lesers kommen auf Deutsch**, wie die der Listeneinfuhr.
