# Autorisierung

**Wer was darf — je Fähigkeit, als Schnittmenge.** Die Regeln stehen in
[0078](entscheidungen/0078-der-wirksame-akteur.md); hier steht, was sie für
**jede** Fähigkeit aus [`FAEHIGKEITEN.md`](FAEHIGKEITEN.md) heißen.

**Die Zahl steht weiter unten und nicht hier** — sie wächst mit jeder neuen
Fähigkeit, und ein Satz im Fließtext wächst nicht mit. Hier stand **166**,
während die gezählte Tabelle **201** sagte, und Vermerk 0078 nannte **164**;
angestrichen vom Zweitblick über Phase 5.2 (Hinweis 14). Der Wächter hält die
Tabelle, nicht diesen Absatz — also nennt dieser Absatz keine Zahl mehr.

## Wie zu lesen

| Spalte | Heißt |
| --- | --- |
| **Fähigkeit** | der Name aus `FAEHIGKEITEN.md`, Zeichen für Zeichen — der Wächter hält beide Listen gegeneinander |
| **Träger** | wem die Sache gehört: einem **Profil** (dann zählt die Rolle darin), dem **Nutzer** (dem Menschen, der arbeitet — ohne Profil), dem **Bestand** (der Datei auf der Platte) oder dem **Rechner** (Pakete, Vorlagen, die Anwendung selbst) |
| **Wirkung** | eine der **sechs** Klassen aus `domain.Wirkung`, von innen nach außen: `lokalLesen`, `lokalSchreiben`, `externEmpfangen`, `externSenden`, `externSchreiben`, `verbindlich`. **Die dritte kam mit N7** (#1568) — „macht diesen Rechner von außen erreichbar" traf keine der fünf. **Sie steht vor dem Senden**, weil diese Reihe die Reichweite der **Daten** misst: ein offener Port bewegt kein Byte. Was er schafft, ist eine Gelegenheit, und die hält das zweite Tor — bestätigter Kontakt und freigegebener Zweck |
| **Mindestrolle** | was die Klasse an einem **Profil** verlangt — `leser`, `mitarbeiter`, `verwalter + Nachweis`. Für die anderen Träger steht ein Strich: dort gibt es keine Rolle |
| **Durchsetzung** | **wo** das Nein entsteht. Zutritt und `query_only` sind die zwei Stellen von vorher; der **Riegel** ist die dritte aus 0078 |

**Was für jede Zeile gleich gilt, steht nicht in ihr**: ein Token reicht nie
weiter als die Rolle und nie nach draußen (`lesen` bis `lokalLesen`,
`schreiben` bis `lokalSchreiben`); ein Nachweis wird nur am Fenster
ausgestellt; der Akteur entsteht **je Aufruf**, und ein Entzug zwischen zwei
Rufen ist beim zweiten wirksam. Das ist `domain.Akteur.Darf`, und die Tests
dazu stehen in `internal/domain/akteur_test.go`.

**Und seit ID-03 (#1702) gilt es auch für die
Verbindung.** Bis dahin war „beim zweiten wirksam" nur für die Riegel wahr:
ob das Fenster über die schreibende oder die nur lesende Verbindung
arbeitet, entschied sich beim Wechsel, und MCP-Server und Dienst entschieden
es beim Start. Jetzt vergleicht das Fenster bei jedem Ruf die Revision der
Rechte und richtet neu, wenn sie sich bewegt hat
(`TestEinEntzugWirktBeimNaechstenAufruf`); MCP-Server und Dienst fragen vor
jedem schreibenden Werkzeug und jeder schreibenden Anfrage den frisch gebauten
Akteur (`TestDerServerFragtJedenSchreibendenRufNeu` in `cmd/mse-mcp`,
`TestDerDienstFragtJedeSchreibendeAnfrageNeu` in `cmd/mse-rest`). Ausgenommen sind nur die zwei
Wechsel von Profil und Nutzer — wer nur noch lesen darf, muss hinauskommen.

**Und wo der erste Schritt nicht schreibt, fragt die Bindung selbst.** Die
KI-Aufträge stehen unter `lokalSchreiben`, Mindestrolle Mitarbeiter — ein
Modellruf kostet Zeit, Strom und bei einem Anbieter Geld —, aber erst die
Übernahme schreibt. Die nur lesende Verbindung hielt deshalb nur die
Übernahme auf, und ein Leser ließ das Modell arbeiten. Seit ID-03 fragen die
acht Einstiege den Akteur vor dem Modell (`TestEinLeserLaesstKeinModellArbeiten`),
und an MCP und Dienst die Schreibwache.

**„Zutritt: ohne Rolle kein Profil" hat eine Ausnahme**, und sie steht hier,
weil sie in keiner Zeile steht: ein Eigner **ohne** zweiten Faktor lässt jeden
herein (`service.Nutzerverwaltung.Zutritt`, 0074). Wer so kommt, hat dort
keine Rolle — er **liest**, und seit ID-03 schreibt er dort nichts mehr
(nativer Befund F01: vorher sagte das Banner „nur lesen", und die Verbindung
blieb schreibend).

**„Fremde Nutzer verlangen deren zweiten Faktor" gilt seit ID-03 wirklich.**
Die Zeilen am Träger Nutzer sagten es seit 0078, und kein Dienst hielt es:
jeder konnte den Faktor eines anderen entfernen, ihn umbenennen oder löschen.
Jetzt fragt jede Bindung, die einen Menschen oder ein Profil nennt,
`Nutzerverwaltung.UeberNutzer` bzw. `UeberProfil` — über einen anderen
schreibt, wer ihn werden kann, ohne dass sein Faktor fragt. Ein Wächter hält,
dass keine Bindung die Frage vergisst (`internal/werkzeuge/uebernutzer_test.go`).

## Gezählt

| | Zahl |
| --- | ---: |
| Fähigkeiten | **379** |
| `lokalLesen` | **150** |
| `lokalSchreiben` | **199** |
| `externEmpfangen` | **6** — und hier stehen alle sechs: **Horchposten**, **Ferneinladung** und **Ankündigung** aus Werkstück N7 (#1568), der **Dienst für Skripte** und der **Werkzeugserver über HTTP** aus N6 (#1572), dazu die **Partiefreigabe** aus G5 (#1462). Die ersten drei standen vorher unter `lokalSchreiben`, und die Zahl dort ist um genau diese drei gesunken.<br>**Die Aufzählung stand zweimal und war einmal zu kurz** — die Zahl sagte fünf, die Namen nannten drei (Befund 6 des Zweitblicks über Phase 5.4). Eine Kurzbilanz, die weniger nennt, als sie zählt, lässt eine Prüfung an ihr vorbeilaufen (§10) |
| `externSenden` | **24** |
| `externSchreiben` | **0** — die Klasse ist besetzt, aber leer; die erste Fähigkeit darin kommt mit Phase 5.3 |
| `verbindlich` | **0** — dito, Phase 5.6 |
| am Träger **Profil** | **291** |
| am Träger **Nutzer** | **23** |
| am Träger **Bestand** | **14** |
| am Träger **Rechner** | **51** |

**Diese Zahlen zählt ein Wächter nach** (`internal/werkzeuge/autorisierung_test.go`),
und er hält dazu: jede Fähigkeit aus `FAEHIGKEITEN.md` steht hier genau
einmal, jede Wirkung ist eine der fünf, jede Mindestrolle ist die, die
`domain` für die Klasse nennt — und am Träger Profil steht nie ein Strich.
**Was er nicht hält**: ob der Träger stimmt und ob die Durchsetzung wirklich
dort sitzt, wo die Zeile es sagt. Für die Riegel belegen das die Läufe
`internal/app/leserausfuhr_test.go`, `internal/app/leseraussen_test.go`,
`internal/app/leserbibliothek_test.go` und
`internal/flaechen/leserausfuhr_test.go`.

## Die Matrix

| Fähigkeit | Träger | Wirkung | Mindestrolle | Durchsetzung |
| --- | --- | --- | --- | --- |
| Sets auflisten | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Ein Set öffnen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Set vorab ansehen | Profil | `lokalLesen` | leser | liest die Datei und fragt den Bestand dieses Profils, ob dieselbe Quelle schon darin liegt — geschrieben wird nichts (FMT-03, #1692). **Am Dienst und über MCP-HTTP nur, was der Pfadvertrag zulässt**, wie jeder Pfad |
| Ein leeres Set anlegen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Set löschen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Zuletzt benutzt merken | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Die Menüleiste nach dem Stand des Fensters führen — Titel und Aktivierung von Seitenleiste, Navigator und Inspector | Rechner | `lokalSchreiben` | — | sie ändert die Menüleiste dieses Fensters und sonst nichts — kein Profil, kein Bestand, keine Daten; die Titel kommen aus dem Katalog, der Stand aus dem Fenster (UXR-F14, #2029). **Gesperrt hält das Tor die Bindung an**, und die Leiste graut die drei Einträge dann selbst |
| Die Handlungen des gewählten Objekts in die Menüleiste bringen — dieselbe Liste wie im Rechtsklick und im Überlauf | Rechner | `lokalSchreiben` | — | sie ändert die Menüleiste dieses Fensters und sonst nichts — kein Profil, kein Bestand, keine Daten; die Einträge kommen aus derselben Liste wie im Rechtsklick, ein Klick meldet nur zurück, und **was er tut, entscheidet das Fenster** samt der Rolle (GESTE, #2358). **Gesperrt hält das Tor die Bindung an**, und die Leiste zeigt dann kein Objektmenü |
| Weiterarbeiten — die Übersicht mit Angeheftetem | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) — die Übersicht liest, Anheften und Lösen schreiben je Profil |
| Was Aufmerksamkeit braucht — Sets mit Vorbehalten, Decks mit Fehlkarten, Pakete mit Skriptbefunden, und was der Betrieb meldet | Profil | `lokalLesen` | leser | liest nur, und nur, was das Profil sieht — dieselben Quellen wie `Vorbehalte`, `DeckAbgleichen` und `PaketdateiPruefen`. **Das Betriebsband (OPS02) schreibt auch nichts**: es fragt `statfs`, den Sicherungsstand und `launchctl print` und zählt, wie oft der Webprozess neu geladen wurde |
| Hintergrundaufgaben — was gerade geholt wird, mit Stand, Fortschritt und Dauer | Rechner | `lokalLesen` | — | liest das Auftragsbuch dieses Fensters oder dieses Dienstes — es gehört dem Prozess, nicht einem Profil. **Am Dienst gehört ein Auftrag dem Principal, der ihn anstieß** (FMT-03, #1692); ein anderer bekommt `404`. Abbrechen (`DELETE`) und Wiederholen (`POST`) gehen wie jeder schreibende Weg durch Schreibwache und Tokenumfang, und die Wiederholung fragt den Pfadvertrag neu — die **Wirkung** des Auftrags selbst trägt die Zeile seines Weges. **Am Fenster hält `AuftragAbbrechen` einen eigenen Datei-Lauf an** — es gibt dort nur die eigenen, und angehalten wird, was der Mensch davor angestoßen hat |
| Eine Ausgabe des Dienstes abholen | Rechner | `lokalLesen` | — | liest den Ausgang des Dienstes — er gehört dem Prozess, nicht einem Profil (FMT-03.AK-03, #1692). **Und dem Principal, der die Ausgabe anstieß** (SEC-P6, #2002), wie ein Auftrag: ein anderer — auch die Prozessmarke, auch ein Token desselben Profils — bekommt `404` wie für einen Namen, den es nicht gibt. Der Name bleibt 80 Bit gewürfelt, nur in der Antwort der Ausgabe genannt, eine Stunde gültig; **er ist nicht mehr das einzige Recht**. Vergessen (`DELETE`) geht wie jeder schreibende Weg durch Schreibwache und Tokenumfang und fragt denselben Principal |
| Mitgelieferte Sets nennen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Ein Set sichern | Profil | `lokalSchreiben` | mitarbeiter | **Riegel im Setschreiber** (0078) — vor `query_only`, denn die Kopie liegt sonst schon da, wenn der Treiber die Herkunftszeile abweist |
| Setangaben lesen und setzen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Vorbehalte eines Sets lesen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die Karten eines Sets auf Probleme prüfen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) — die Prüfung zeichnet und schreibt nichts; **ihr Stand liegt im Prozess**, und `Setstand` fragt das Profil, bevor einer genannt wird (PERF02, #2352) |
| Karten auflisten | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die **Gestalt** der Kartenliste — Spalten und Ordnung | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Der **Schnellfilter** der Kartenliste | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Eine Diagnose für einen Fehlerbericht holen | Rechner | `lokalSchreiben` | — | sie schreibt eine Datei über die Anwendung, nicht über ein Profil — der Dateidialog ist der Riegel |
| Ein Diagnosepaket zusammenstellen, sichern und wieder löschen | Rechner | `lokalSchreiben` | — | es schreibt eine Datei über die Anwendung, nicht über ein Profil — der Dateidialog ist der Riegel. **Der Protokollteil hängt zusätzlich an der Rolle** (O9): ein Gast in einem fremden Profil bekommt ihn nicht, und der Grund steht statt seiner im Paket |
| Einen Fehler, den das Fenster beim Zeichnen gefangen hat, mit seiner Kennung ins Protokoll schreiben | Rechner | `lokalSchreiben` | — | **Es schreibt eine Zeile ins Protokoll dieser Anwendung**, nicht in ein Profil und nicht in den Bestand — dieselbe Klasse wie das Diagnosepaket darüber. **Keine Rolle**, weil jede Fehlerfläche melden können muss, auch die eines Lesers; was in die Zeile kommt, ist geprüft und gekürzt (F01, #2340) |
| Sehen, dass die letzte Sitzung nicht sauber endete, und ob eine Absturzspur vorliegt | Rechner | `lokalLesen` | — | sie gibt aus, was der Start beim Lesen der liegengebliebenen Marken fand — Beginn, Fassung und ob eine Spur da ist, **nicht die Spur selbst**. **Keine Rolle**, weil ein Abbruch das Fenster betraf und nicht ein Profil; die Spur geht erst mit dem Diagnosepaket hinaus, geschwärzt (F04, #2343) |
| Ein Leistungsprofil dieses Fensters aufzeichnen | Rechner | `lokalSchreiben` | — | es schreibt eine Datei neben den Bestand und misst den Prozess, nicht ein Profil — wie die Diagnose. **Eines zur Zeit** je Prozess, und die Dauer ist auf 60 Sekunden begrenzt |
| Die Zwischenspeicher sehen und leeren | Rechner | `lokalSchreiben` | — | er wirft weg, was sich neu holen lässt, und hält keine Daten eines Profils — Vorschauen und Spielbilder leert ohnehin jeder Profilwechsel |
| Betrieb und Vertrauen sehen — Fassung, Datenstand, Signatur, Aktualisierung, Orte und was geprüft ist | Rechner | `lokalLesen` | — | sie liest, was installiert ist, und nennt Orte als Tilde — kein Inhalt eines Profils |
| Die Karten mit ihrem Bild sehen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Marken an einer Karte führen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Marken an jedem Fachobjekt führen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser (0073); **eine Marke ist nie ein Recht** — keine Prüfung liest sie ([0135](entscheidungen/0135-marken-an-jedem-fachobjekt.md)) |
| Nach Marken filtern | Profil | `lokalLesen` | leser | nur die Marken des eigenen Profils (UX-12.AK-04) |
| Marken pflegen — umbenennen, zusammenführen, überall entfernen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser (0073); nur die Marken des eigenen Profils |
| Eine Karte auf einen **echten Druck** beziehen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Karten bei Scryfall **suchen** | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — jede Anfrage nach draußen fragt den Akteur ([0078](entscheidungen/0078-der-wirksame-akteur.md)). **Was der Zwischenspeicher des Rechners schon hat, bekommt auch ein Leser**: öffentliche Kartendaten, keine Anfrage (SEC01, `rollenwechsel_test.go`) |
| Regelhinweise von Scryfall lesen | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — jede Anfrage nach draußen fragt den Akteur ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Einen Manakosten gegen Scryfall prüfen | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — jede Anfrage nach draußen fragt den Akteur ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Sammeldaten von Scryfall vorhalten | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — jede Anfrage nach draußen fragt den Akteur ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Kartendaten von Scryfall übernehmen | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — jede Anfrage nach draußen fragt den Akteur ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Kartenfelder beschreiben | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Eine Karte anlegen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Die Kartenskripte eines Spielpakets zeigen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Ein Kartenskript ausführen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Wonach ein Spielpaket auszählen lässt** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Ein Set nach einer Dimension auszählen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Die Parameterhilfen eines Schlüsselworts** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Die Decks sehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Ein Deck samt seinen Plätzen sehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Ein Deck anlegen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Ein Deck umbenennen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Ein Deck kopieren** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Aus einem Set ein Deck anlegen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)); **das Set muss im selben Profil liegen** — `Deckablage.Uebernehmen` prüft jede Karte gegen das gezeigte Profil (#1815) |
| **Ein Deck löschen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Eine Karte ins Deck legen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Einen Platz aus dem Deck nehmen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Einen Platz in den anderen Bereich legen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Sehen, ob dieser Prozess lebt und arbeiten kann — der Betriebszustand | Bestand | `lokalLesen` | — | Der Weg fragt den Bestand nur, **ob** er antwortet — kein Inhalt, kein Profil, keine Zeile. `GET /gesundheit` und `GET /bereit` stehen trotzdem **hinter der Marke**: „wandert gerade" ist eine Auskunft über diese Datei, und der Dienst hört ohnehin nur auf `127.0.0.1` |
| Sehen, was dieser Prozess gewirkt hat — Zahlen mit endlich vielen Reihen | Bestand | `lokalLesen` | — | Gezählt wird im Arbeitsspeicher **dieses** Prozesses; der Bestand kommt nicht vor. Die Reihen tragen drei Aufzählungen und keine Kennung — sie sagen, **wie oft** etwas gewirkt hat, nie **woran**. Hinter der Marke wie der Betriebszustand: wie oft jemand nach draußen gesendet hat, ist eine Auskunft über seine Arbeit |
| Sehen, welches Programm antwortet — Fassung, Plattform, Fläche | Bestand | `lokalLesen` | — | Alles davon steht in der Binärdatei; der Bestand kommt nicht vor. Hinter der Marke wie der Betriebszustand: eine Fassung sagt, welche Lücken dieser Stand hat |
| Sehen, ob die neueste Freigabe **diesen** Bestand öffnen kann | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — die Freigabe und ihr Manifest kommen von außen ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Einen Fassungswechsel vorbereiten — Platz, laufende Arbeit, eine zurückgelesene Sicherung | Rechner | `lokalSchreiben` | — | es schreibt eine Sicherung und liest sie zurück; der Ort ist der Sicherungsordner, nicht ein Profil |
| Sehen, was zurückgehen kann und was davon diese Anwendung selbst tut | Rechner | `lokalLesen` | — | Vier Marken, die im Programm stehen — kein Bestand, kein Profil, kein Netz |
| Sehen, was ein Zurückspielen dieser Sicherung kostet | Bestand | `lokalLesen` | — | Der Vergleich geht über **jedes** Profil: eine Rückkehr tauscht den ganzen Bestand, und wer nur das eigene zählte, verschwiege die Unsicherheit der anderen. Deshalb Träger **Bestand** und keine Rolle — gelesen wird eine Datei, die jemand ausdrücklich benennt |
| Sehen, worüber nach einer Rückkehr nichts feststeht | Profil | `lokalLesen` | leser | Die Epoche sieht jedes Profil — der Bestand ist gemeinsam zurückgegangen. **Ihre Mengen sieht nur, wem sie gehören** ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Nachsehen, was im Bestand herumliegt | Bestand | `lokalLesen` | — | Die Analyse zählt **verwaiste** Zeilen — also gerade die, die zu keinem Profil mehr gehören. Sie ist eine Frage nach der Gesundheit der **Datei** und nicht nach dem Besitz; dieselbe Lage wie beim Waisenzähler der Diagnose |
| Aufräumen, was im Bestand herumliegt | Bestand | `lokalSchreiben` | — | Sie räumt weg, was zu keinem Profil gehört, und sichert vorher. Der Träger ist deshalb der **Bestand**; die Sammlung wird dabei nicht angefasst — was nach Dubletten aussieht, steht als Auskunft da ([0088](entscheidungen/0088-die-epoche-einer-rueckkehr.md) für die Rückkehren, die ebenso stehen bleiben) |
| Sehen, wie viel Geld an einem Konto freigegeben ist | Profil | `lokalLesen` | leser | Ein Budget gehört einem **Profil** — es ist die Freigabe seines Eigners, und ein fremdes geht niemanden an ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die letzten Aufträge eines Kontos ansehen | Profil | `lokalLesen` | leser | dito. **Ohne Prompt und ohne Zugangsdaten** — was dasteht, ist die Aufgabe als Marke |
| Die Kostengrenze eines Kontos setzen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Kein Nachweis**, und das ist die Aussage: eine Grenze gibt nichts aus, sie begrenzt — was darüber hinaus geht, verlangt danach ohnehin je Auftrag eine Freigabe |
| Einen bezahlten KI-Auftrag freigeben | Profil | `lokalSchreiben` | mitarbeiter | **Der Nachweis wird hier ausgestellt und nicht eingelöst** (#1539) — geschrieben wird eine Zeile im Bestand, und die Klasse sagt genau das. **Durchgesetzt wird sie aber nicht über `Pruefe`**, sondern in `Nachweise.Ausstellen`, und zwar **enger**: Modus `fenster`, kein Token, nicht nur lesend, mindestens Mitarbeiter. Ein Modell stellt sich keine Freigabe aus ([0078](entscheidungen/0078-der-wirksame-akteur.md)). **`verbindlich` steht hier nicht**, obwohl es nach der stärkeren Klasse aussieht: die verlangt selbst einen Nachweis, und ein Nachweis, der einen Nachweis braucht, ist ein Kreis. Die Klasse steht am **Auftrag**, den diese Freigabe trägt, und dort wird sie eingelöst |
| Die Arten, zurückzusetzen, ansehen | Rechner | `lokalLesen` | — | Sechs Marken, die im Programm stehen — kein Bestand, kein Profil, kein Netz. Dieselbe Lage wie bei den Rücksprungarten |
| Sehen, was eine Rücksetzung anfassen würde | Profil | `lokalLesen` | leser | **Gezählt wird nur, was dem Rufenden gehört**: ein fremdes Profil kommt als Hindernis zurück, und zwar **ohne Zahlen** — wer es nicht anfassen darf, soll auch nicht erfahren, wie viele Karten darin liegen ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Zurücksetzen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)), dazu die Prüfung, wem das Ziel gehört. **Die ganze Anwendung verlangt zusätzlich einen [Nachweis](entscheidungen/0078-der-wirksame-akteur.md)**: eine Rolle reicht dafür nicht, denn sie nähme anderen Eigentümern ihre Daten weg — und ein Nachweis entsteht am Fenster, von einem Menschen. **Ihre Vorschau sieht jedes Profil des Bestands** (Befund 1 des Zweitblicks über Phase G, #2282): ein fremdes wird mitgezählt und nicht benannt, und seine offene Rückkehr sperrt |
| Den Neuanfang freigeben | Profil | `lokalSchreiben` | mitarbeiter | **Dieselbe Gestalt wie die KI-Freigabe** (G12, #2282): ausgestellt und nicht eingelöst, durchgesetzt in `Nachweise.Ausstellen` — Modus `fenster`, kein Token, nicht nur lesend, mindestens Mitarbeiter. Eingelöst wird der Nachweis beim **Zurücksetzen** darüber, und dort steht, warum eine Rolle allein nicht reicht |
| Einen laufenden Bildauftrag abbrechen | Profil | `lokalSchreiben` | mitarbeiter | Dieselbe Klasse wie der Auftrag, den er beendet — wer keinen anstoßen darf, hält auch keinen fremden an. **Er trifft nur einen Auftrag in diesem Prozess**; das ist die Reichweite eines Kontexts und keine Lücke der Prüfung, wörtlich wie beim Indexabbruch |
| Sehen, wogegen der lokale Suchindex sucht | Profil | `lokalLesen` | leser | Der Index liegt im Bestand und wird über die Profilbindung gelesen; ein Profil sieht die öffentlichen Dokumente und seine eigenen — die Sicht steht schon in der Kandidatenauswahl im SQL und wird vor der Ausgabe **ein zweites Mal** geprüft |
| Lokal suchen — Volltext und Vektoren zusammengeführt | Profil | `lokalLesen` | leser | Dieselbe Sichtprüfung, und sie ist hier die Zusage: **fremde Treffer dürfen auch nicht in Anzahl, Ausschnitt oder Vorschlag auftauchen**. Ein Tombstone sperrt einen Inhalt sofort, auch in einem Index, der vor dem Widerruf gebaut wurde |
| Den Suchindex neu aufbauen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). Dazu ein Platz der Klasse `indexbau` und **eine** Schreibinstanz je Index: ein zweiter Aufbau wird abgewiesen, statt zu warten |
| Einen laufenden Indexaufbau abbrechen | Profil | `lokalSchreiben` | mitarbeiter | Dieselbe Klasse wie der Aufbau — wer nicht bauen darf, hält auch keinen fremden Bau an. Der Abbruch erreicht nur einen Aufbau **in diesem Prozess**; das ist die Reichweite des Riegels und keine Lücke der Rechteprüfung |
| Den Wissensbestand fragen — Karten, Regeln, Rulings | Profil | `lokalLesen` | leser | Dieselbe Sichtprüfung wie bei der Indexsuche, über die sie läuft — **und derselbe Fortsetzungszeiger bindet das Profil**: wer die zweite Seite einer Liste holen will, die für jemand anderen entstand, bekommt eine Absage statt der Seite ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Karten finden, die einer eigenen ähneln | Profil | `lokalLesen` | leser | Dazu die Ausgangskarte: sie wird über den Setleser des Profils gelesen, und eine fremde ist damit keine. **Der Stand der Karte bindet die Antwort** — eine Karte, die sich seit der Anfrage geändert hat, gibt einen Konflikt statt einer Antwort auf eine andere Frage |
| Sehen, was an ein Modell ginge | Profil | `lokalLesen` | leser | **Lesend, weil nichts hinausgeht**: das Paket entsteht und bleibt liegen. Die Freigabe ist die Stelle, an der ein Mensch sieht, **was** hinausginge — Anbieter, Ort der Berechnung, Datenkategorien und die Ausschnitte selbst. **`externSenden` beginnt erst danach**, beim Ruf an das Modell (A6/A11); wer diese Zeile für die Freigabe hielte, verwechselte das Zeigen mit dem Senden |
| Eine Rückkehr für nachgesehen erklären | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Ein Mensch quittiert, kein Modell**: nachsehen heißt in die Kiste greifen und zählen — deshalb gibt es das Werkzeug über MCP nicht |
| **Ein Deck auf ein Format stellen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Ein Deck gegen sein Format prüfen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Nachsehen, was von einem Deck da ist** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Sehen, welche Exemplare für ein Deck vorgesehen sind** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Die Fehlkarten eines Decks als Liste bekommen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil (0070) |
| **Ein Deck als Liste ausgeben** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil (0070) |
| **Etwas über die Ziele des Systems teilen** — eine Einladung, ein Kartenbild, ein Set, ein Deck, die Sammlung | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil (0070). **Was hinausgeht, schickt der gewählte Dienst** und nicht dieses Programm; es legt nur eine Kopie ab (0155) |
| **Die Fehlkarten eines Decks auf die Wunschliste setzen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Sehen, welche Decks ein Exemplar benutzen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil (0070) |
| **Exemplare einem Deckplatz zuordnen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Eine Zuordnung wieder lösen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser (0073) |
| **Die Formate sehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Die nachgelesenen Deckbauarten sehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Ein Format aus einer Bauart anlegen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Ein Format anlegen oder fortschreiben** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Ein Format löschen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Die Turniere sehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Ein Turnier samt Verlauf sehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Ein Turnier anlegen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Ein Turnier nach einer Vorlage anlegen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)); die Vorlage wird im Profil gelesen |
| **Ein Turnier löschen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Einen Teilnehmer anmelden** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Einen Teilnehmer abmelden** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Sehen, wie die nächste Runde ausfiele** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)). **MCP nennt es `schreiben.turnier.vorschlag`, obwohl es nichts schreibt** — eine Entscheidung über die Nachbarschaft zum Legen (`internal/mcp/turniereschreiben.go`), keine über die Wirkung |
| **Die nächste Runde legen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Ein Ergebnis eintragen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Die Rangliste sehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Sehen, was eine Berichtigung verschöbe** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)); schreibt nichts |
| **Ein Turnier abschließen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Ein Turnier wiedereröffnen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Das Protokoll eines Turniers sehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Ein Turnier als Datei ausgeben** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)); **die Datei entsteht auf diesem Rechner** und trägt keine Kennung des Bestands (Vermerk [0140](entscheidungen/0140-ein-turnier-als-datei.md)) |
| **Ein Turnier aus einer Datei einlesen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)); das Turnier entsteht im gezeigten Profil |
| **Den Bericht eines Turniers sichern** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Die Autoersetzungen eines Spiels sehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Eine Autoersetzungsregel setzen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Eine eigene Autoersetzungsregel löschen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Die eigenen Autoersetzungen vergessen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Die Autoersetzung ganz schalten** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Eine Karte kopieren | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Karten über die **Zwischenablage** bewegen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Eine Karte löschen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Karten umordnen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Feld setzen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Feld auf vielen Karten setzen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Eine Massenänderung vorher sehen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) — sie schreibt nichts, auch mit `erwartet` nicht |
| Ein Symbol zeichnen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Ein Bild als Symbol nachzeichnen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Das Fenster rechnet nur** (`SymbolNachzeichnen` ist lesend) — der Weg ohne Fenster legt das Ergebnis ins Set, und das ist die Wirkung dieser Fähigkeit |
| Ein gezeichnetes Symbol ins Set legen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Einen Bézierhandgriff ziehen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Einem Punkt seine Art geben | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Einen Punkt einer Form einfügen oder löschen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Das Symbol eines Sets lesen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die Packarten eines Spiels zeigen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Einen Booster ziehen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die Schlüsselwörter eines Sets zeigen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Ein Schlüsselwort anlegen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Schlüsselwort ändern | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Schlüsselwort löschen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Die Auswahl eines Feldes zeigen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Aus der Auswahl wählen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Beim Tippen ersetzen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Im Volltext suchen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Im Text **ersetzen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Set auszählen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Einen Ausdruck auswerten | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die Namen der Umgebung lesen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Einen Ausdruck einfärben | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Eine Paketdatei einfärben | Rechner | `lokalLesen` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Die Dateien einer Vorlage aufzählen | Rechner | `lokalLesen` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Die Skripte einer Vorlage prüfen | Rechner | `lokalLesen` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Eine Vorlage lesen | Rechner | `lokalLesen` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Eine Vorlage schreiben | Rechner | `lokalSchreiben` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Eine Vorlage im Paket anlegen | Rechner | `lokalSchreiben` | — | wie das Schreiben einer Vorlage (FMT-02, #1691) |
| Ein Bild oder eine Schrift ins Paket legen | Rechner | `lokalSchreiben` | — | **nur Fenster und Kommandozeile** — sie liest eine Datei mit den Rechten dessen, der davor sitzt; über REST oder MCP wäre sie ein Weg, beliebige Dateien des Rechners hinauszulesen |
| Eine Datei im Paket umbenennen | Rechner | `lokalSchreiben` | — | eine Vorlage nur mit ihrem Stand |
| Eine Datei aus dem Paket entfernen | Rechner | `lokalSchreiben` | — | in die Sicherung neben dem Paket; eine Vorlage nur mit ihrem Stand |
| Eine ungespeicherte Vorlage sehen | Rechner | `lokalLesen` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Die Testkarten zu einer Vorlage nennen | Rechner | `lokalLesen` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Die Symbole einer Vorlage nennen | Rechner | `lokalLesen` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Die Verwendungen eines Namens nennen | Rechner | `lokalLesen` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Die eigene Sammlung aufzählen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Nachsehen, was von einem Druck da ist | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Ein Exemplar aufnehmen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Exemplar ändern | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Exemplar herausnehmen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Die abgelegten Preise eines Drucks sehen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Den Preisverlauf sehen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Einen verbindlichen Handelsschritt freigeben | Nutzer | `lokalSchreiben` | — | **Der Nachweis ist die Zustimmung selbst**, und er entsteht nur am Fenster: `Nachweise.Ausstellen` weist jeden anderen Modus und jedes Token ab. **Träger ist der Nutzer**, denn was hier festgehalten wird, ist, dass **ein Mensch** dieses eine wollte — nicht, was einem Profil gehört. **Mindestens mitarbeitend**, und das prüft `Ausstellen` selbst; ein Strich steht hier, weil die Spalte die Rolle am **Profil** meint. **Fünf Minuten, einmal, an Vorgang und Stand gebunden** (T10, #1475) |
| Eine **Deckliste** deuten | Profil | `externSenden` | mitarbeiter | **Nicht `lokalLesen`**: jede unbekannte Zeile fragt Scryfall nach den passenden Drucken — derselbe Weg und derselbe Riegel wie beim Deuten einer Kartenliste. Riegel an der Netzgrenze ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Eine **Deckliste** ins Deck legen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Es schreibt Plätze und keinen Besitz** — ein Deck verweist auf Karten; die Sammlung bleibt unberührt. **Und der Stand aus der Vorschau ist die zweite Bedingung**, die keine Rolle ersetzt: er hält die Änderung eines anderen ab, nicht den Falschen (T8, #1473) |
| Das **eigene** Konto bei einem Marktplatz ansehen | Profil | `externSenden` | mitarbeiter | **Nicht `lokalLesen`**: der Abruf geht ins Netz und nennt dabei die Zugangsmarke des Kontos. Riegel an der Netzgrenze ([0078](entscheidungen/0078-der-wirksame-akteur.md)) — **und davor ein zweiter Schalter**, die `konto.schau`, die ausgeschaltet anfängt: wer den Marktplatz eingeschaltet hat, hat damit nicht sein eigenes Konto freigegeben (T11, #1476). **Geschrieben wird dort nichts** — es gibt keinen Weg dafür |
| Sehen, wo es eine Karte außerhalb gibt | Profil | `lokalLesen` | leser | **Die Adressen entstehen auf diesem Rechner** — aus den Kennungen eines Drucks, die schon im Bestand liegen. Es geht kein Byte hinaus, und deshalb ist es `lokalLesen` und nicht `externSenden`: Zutritt, ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Einen dieser Wege im Browser öffnen | Nutzer | `externSenden` | — | **Dieselbe Gestalt wie die Anmeldung über einen Anbieter**: was hinausgeht, geht aus dem Browser des Menschen, und der Browser ist der Riegel. **Träger ist deshalb der Nutzer und nicht das Profil** — der Anbieter erfährt, welche Karte jemand aufgeschlagen hat, und nichts aus dem Bestand. **Die Durchsetzung dieses Programms ist die Positivliste**: https, der Wirt Zeichen für Zeichen, keine Benutzerangabe, kein eigener Port, kein Geheimnis in der Abfrage — und sie wird **beim Öffnen** geprüft und nicht nur beim Bauen, denn dazwischen ging die Adresse durch fremde Hände (T8, #1473) |
| **Die Angebote eines Marktplatzes sehen** | Profil | `externSenden` | mitarbeiter | **Nicht `lokalLesen`** (A4, #1443): der Abruf geht ins Netz, nennt dabei die Zugangsmarke des Kontos — und merkt sich die gefundene Zuordnung im Bestand. Beides zusammen ist mehr als Lesen |
| **Den Preisverlauf einer anderen Quelle holen** | Profil | `externSenden` | mitarbeiter | **Wie der Marktabruf** (A5, #1444): der Abruf geht ins Netz, nennt dabei die Zugangsdaten des Kontos, verbraucht Kontingent — und legt Preisstände ab |
| **Das Kontingent an einem Anbieterkonto sehen** | Profil | `externSenden` | mitarbeiter | **Auch das Fragen kostet** (A5, #1444): Scrydex rechnet die Abfrage als Kredit. Sie ist deshalb kein `lokalLesen`, obwohl sie am Bestand nichts ändert |
| **Sehen, woher ein Kartenwert stammt** | Profil | `lokalLesen` | leser | **Eine Auskunft aus dem Bestand** (A11, #1450) — hier geht kein Ruf nach draußen und es entsteht kein Auftrag. **Je Profil**: die Herkunft steht unter der Profilkennung, und wer ein anderes Profil führt, sieht sie nicht |
| **Sehen, was über KI-Modelle feststeht** | Profil | `lokalLesen` | leser | **Es fragt eine **lokale** Laufzeit, und nur, wenn `ki.laufzeit` auf `an` steht** (A7, #1446); die Vorgabe ist `aus`. Gelesen wird, was installiert ist — nichts wird geladen, nichts geändert. Die Ortsregel kommt aus den Einstellungen. **Wohin gefragt wird, hat ein Mensch gesetzt** — `ki.adresse` setzt kein Token (SEC-P4, #2000) —, und die Grundsperre in `internal/netz` hält `0.0.0.0`, `169.254.0.0/16` und ihresgleichen auch dann fern |
| **Ein Modell Felder einer Karte vorschlagen lassen** | Profil | `lokalSchreiben` | mitarbeiter | **Zwei Schritte, und nur der zweite schreibt** (A7, #1446). Der erste fragt ein Modell und ändert nichts — er kostet trotzdem Zeit und Strom und gehört deshalb nicht unter `lokalLesen`. Der zweite setzt Felder der Karte, mit Herkunft `ki`. **An den drei kopflosen Flächen ist es ein Ruf**, und `uebernehmen` sagt vorher, was am Fenster der zweite Klick sagt |
| **Mehrere Karten auf einmal vorschlagen lassen** | Profil | `lokalSchreiben` | mitarbeiter | **Er schreibt an keiner Karte** (A13.1, #1452) — und steht trotzdem unter `lokalSchreiben`: er bucht **einen** Kostenposten für den ganzen Batch und schließt ihn wieder ([0105](entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)). **Durchgesetzt wird die Freigabe an der Kostenwache**, wie bei jedem einzelnen Auftrag; was der Batch ändert, ist ihr **Zuschnitt** und nicht ihre Zusage |
| **Mehrere Vorschläge als eine Gruppe übernehmen** | Profil | `lokalSchreiben` | mitarbeiter | **Sie schreibt an Karten** (A13.1b, #1452) — anders als der Batch darüber, und deshalb unter `schreiben` und nicht unter `lokalSchreiben`. **Die Aufsicht fragt für jede Karte einzeln**, und zwar **vor** dem ersten Schreiben: wer eine Karte in der Gruppe nicht ändern darf, ändert **keine** ([0109](entscheidungen/0109-eine-gruppe-darf-ueber-setgrenzen-gehen.md)). **Kostet nichts** — das Modell hat schon geantwortet; hier wird nur geschrieben |
| **Ein Modell ein Kartenbild erzeugen oder bearbeiten lassen** | Profil | `lokalSchreiben` | mitarbeiter | **Vier Schritte, und nur der letzte schreibt** (A12, #1451). Erzeugen legt in den Vorrat, Zuschnitt und Vorschau lesen von dort, Verwerfen räumt weg — erst Einsetzen berührt die Karte. **Die Stufe richtet sich nach dem Ende**: wer einen Vorschlag erzeugen darf, darf ihn auch einsetzen, und eine mildere Stufe für die ersten drei hieße, Geld ausgeben zu dürfen, ohne das Ergebnis benutzen zu können. **Die Auskunft über die Herkunft liegt darunter** und geht als Lesen durch — sie fragt niemanden nach draußen |
| **Eine Karte beurteilen lassen — mit Belegen** | Profil | `lokalSchreiben` | mitarbeiter | **Sie ändert weder Karte noch Sammlung noch Spielstand** (A13.2a, #1452) — und steht trotzdem unter `lokalSchreiben`, aus demselben Grund wie der Batch darüber: sie bucht einen Kostenposten, schließt ihn wieder und fragt ein Modell, das anderswo rechnen kann. **Eine mildere Stufe wäre die stillste Art, Geld auszugeben**: `lokalLesen` verspricht, dass ein Ruf nichts kostet. **Und es gibt keinen zweiten Schritt** — anders als beim Vorschlag wird eine Einschätzung nicht übernommen; es gibt kein Werkzeug, das sie in eine Karte trüge |
| **Eine Karte auf einem Foto erkennen** | Profil | `lokalSchreiben` | mitarbeiter | **Sie legt nichts an** (A13.3, #1452) — weder Druckbindung noch Sammlungsmenge — und steht trotzdem unter `lokalSchreiben`, aus demselben Grund wie die Einschätzung darüber: sie bucht einen Kostenposten, schließt ihn wieder und schickt ein **Bild** an ein Modell, das anderswo rechnen kann. **Ein Foto ist dabei die größere Ausgabe von beiden**: ein Bild kostet bei jedem Anbieter ein Vielfaches eines Satzes, und `lokalLesen` verspricht, dass ein Ruf nichts kostet. **Und es gibt keinen zweiten Schritt**: kein Werkzeug trägt einen Kandidaten in die Sammlung |
| **Apples Gerätemodell für eine Sitzung abschalten** | Profil | `lokalSchreiben` | mitarbeiter | **Dieselbe Rolle wie ein Vorschlag** (AI-05, #2042): wer ein Modell fragen darf, darf es für seine Sitzung abschalten. Er ändert **nichts im Bestand** — der Schalter lebt im Arbeitsspeicher des Prozesses und beendet ein laufendes `fm serve`; die Einstellung `ki.geraetemodell` bleibt, wie sie ist |
| **Messen, was ein Modell kann** | Profil | `lokalSchreiben` | mitarbeiter | **Sie schreibt** (A9, #1448) — eine Aussage über ein Modell, je Profil, Anbieter, Alias und Fähigkeit eine. **Und sie kostet Geld**, wenn der Anbieter anderswo rechnet: derselbe Riegel wie beim Vorschlag hält sie auf — Schalter, »nur lokal«, Journal und die Freigabe eines Menschen. `vergessen` nimmt sie zurück und steht auf derselben Stufe: wer messen darf, darf auch verwerfen |
| Preise holen | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — jede Anfrage nach draußen fragt den Akteur ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Preise einer ganzen Quelle holen | Profil | `externSenden` | mitarbeiter | Derselbe Riegel, und dazu der **Dienstschalter**: die Vorgabe für MTGJSON ist „aus" ([0063](entscheidungen/0063-neue-dienste-sind-aus.md)). Ein Ruf lädt eine große Datei — bis zu einem Gigabyte (A3, #1442) |
| Eine Kartenliste **deuten** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Eine Kartenliste **einlesen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Eine Kartenliste **vorab ansehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Viele Exemplare auf einmal ändern** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Viele Exemplare auf einmal löschen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| **Preise für die ganze Sammlung** | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — jede Anfrage nach draußen fragt den Akteur ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| **Besitze ich diese Karte?** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die **Lagerorte** zeigen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| **Dubletten** finden | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die **Wunschliste** zeigen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die Sammlung **ausgeben** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Eine CSV-Sammlung **zurücklesen** | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Eine CSV-Sammlung **vorab ansehen** | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die Sammlung zusammenrechnen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die Werkzeuge ohne Fenster finden | Rechner | `lokalLesen` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Die Kommandozeile einrichten — `mse` und `mse-mcp` in `~/.local/bin` | Rechner | `lokalSchreiben` | — | Verknüpfungen im Zuhause des Menschen vor dem Fenster, keinem Profil; Fremdes bleibt (OPS-03) |
| Nachsehen, was **Kommandozeile, Codex, Claude und Dienst** von dieser Anwendung wissen | Rechner | `lokalLesen` | — | liest fünf Dateien im Zuhause des Menschen vor dem Fenster, keines Profils — Shellprofil, Codex, Claude Desktop, `~/.claude.json` und die plist des Dienstes (OPS-09, #1961) |
| Kommandozeile, Codex, Claude und Dienst **einrichten und widerrufen** — je Ziel oder **alle mit einem Ruf**, und **wählen, wo der Dienst lauscht** | Rechner | `lokalSchreiben` | — | ein markierter Block, ein JSON-Eintrag mit Gedächtnis, `claude mcp` oder `launchctl` — Dateien und Dienste des Menschen vor dem Fenster, keines Profils; Fremdes bleibt (OPS-09, [0157](entscheidungen/0157-ein-markierter-block.md), [0159](entscheidungen/0159-drei-ziele-ohne-block.md)) |
| Stile nennen und wechseln | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Stiloptionen lesen und setzen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Die Felder des Stils an einer Karte lesen und setzen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Die Wahlen einer Stiloption zeigen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Eine Karte zeichnen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Ein Kartenbild einsetzen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Bild zuschneiden | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)). **Eine Rechnung über das Rechteck**, sie schreibt nichts — geschrieben wird erst beim Einsetzen. **Bis zum 2026-10-01 fehlte diese Zeile**, und kein Wächter sah es: der Zähler hielt ihre Zeile für eine Kopfzeile, weil ihre Bemerkung auf das Wort „Fähigkeit“ und einen Tabellenstrich endet (OPS-10, #1908) |
| Ein Kartenbild von Scryfall holen | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — jede Anfrage nach draußen fragt den Akteur ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Eine Karte aus einem Scryfall-Namen anlegen | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — jede Anfrage nach draußen fragt den Akteur ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Ein ganzes Set von Scryfall einlesen | Profil | `externSenden` | mitarbeiter | Riegel an der Netzgrenze — jede Anfrage nach draußen fragt den Akteur ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Teile gruppieren, auflösen und umordnen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Eine Symmetrie bedienen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Die Symbolbibliothek sehen und durchsuchen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) — die Bibliothek gehört dem Profil (DOM-02, #1697) |
| Ein Setsymbol in die Bibliothek aufnehmen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Symbol der Bibliothek duplizieren und umbenennen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Symbol der Bibliothek löschen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) — und ein getragenes nur bestätigt |
| Ein Symbol der Bibliothek in ein Set einsetzen | Profil | `lokalSchreiben` | mitarbeiter | **Riegel** ([0078](entscheidungen/0078-der-wirksame-akteur.md)) vor der ersten Datei, denn die lesende Verbindung hält keine Datei auf; dazu `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Die Kopie liegt dann schon neben dem Set** — erst die Datei, dann der Wert —, aber kein Set zeigt auf sie; dieselbe Reihenfolge wie beim gezeichneten Symbol |
| Ein Symbol ein- und ausgeben | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) fürs Einlesen; das Ausgeben liest die Bibliothek und schreibt eine Datei — davor steht der **Riegel** ([0078](entscheidungen/0078-der-wirksame-akteur.md)) vor der ersten Datei |
| Ein Symbol der Bibliothek bearbeiten und zurücknehmen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) — das Lesen der Teile geht auch als Leser |
| Die Sets mit der vorigen Fassung nachziehen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) für das Lesen der Träger; **die Dateien schreibt die Ablage, und kein Wert im Bestand ändert sich** — ohne den **Riegel** ([0078](entscheidungen/0078-der-wirksame-akteur.md)) vor der ersten Datei überschriebe ein Leser die Symbole fremder Sets (`service/symbolbibliothekriegel_test.go`, `app/leserbibliothek_test.go`) |
| Ein Symbol in den Fassungen eines Sets sehen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Symbole zur Auswahl stellen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Kartenbilder schreiben | Profil | `lokalSchreiben` | mitarbeiter | **Riegel in der Bildausfuhr** (0078) — ein Bild ist eine Datei, und die Platte fragt den Bestand nicht |
| Ein Profilbild setzen und wegnehmen | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Die Einstellungen eines **Menschen** lesen und setzen | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Die eigenen Daten ausgeben | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, schreibt seine **eigenen** Daten in eine Datei; MCP und Dienst geben das JSON zurück, statt zu schreiben |
| Ein Profil-Abzeichen setzen und wegnehmen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Die Nutzer mit ihrem Bild sehen | Nutzer | `lokalLesen` | — | kein Profil im Spiel — wer arbeitet, darf über sich lesen |
| Einen Druckbogen legen | Profil | `lokalSchreiben` | mitarbeiter | **Riegel im Bogenleger** (0078). **Der Probelauf davor liest nur** (UXR-F7, #2015): er rechnet Karten, Seiten und Raster durch denselben Plan, fragt den Riegel nicht und schreibt nichts — ein Leser darf wissen, wie ein Bogen würde |
| Über eine Exportvorlage ausgeben | Profil | `lokalSchreiben` | mitarbeiter | **Riegel im Fenster** vor dem Lauf (0078) — die Vorlage schreibt das Fenster selbst; `query_only` für die Exportoption. **Der Probelauf davor liest nur** (UXR-F7, #2015): er wertet die Vorlage durch denselben Weg aus, legt ihre Beigaben in eine Ablage statt auf die Platte und fragt den Riegel nicht |
| MTG JSON lesen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| MTG JSON schreiben | Profil | `lokalSchreiben` | mitarbeiter | **Riegel im Exporteur** (0078) |
| Pakete auflisten | Rechner | `lokalLesen` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Ein Paket beziehen | Rechner | `lokalSchreiben` | — | liest aus dem Netz und schreibt Pakete — für den Rechner, nicht für ein Profil. Gesendet wird nichts als die Adresse. **Und nur von einer öffentlich gerichteten** (SEC-P4, #2000): die Adresse nennt der Rufer, und ein Token darf das — deshalb wählt der Bezug weder die Schleife noch das eigene Netz noch `169.254.0.0/16`, geprüft nach der Auflösung und bei jeder Weiterleitung (`internal/netz/ziel.go`) |
| Die eigenen Übersetzungen sehen | Rechner | `lokalLesen` | — | eine eigene Übersetzung gehört dem Rechner wie ein installiertes Paket (LOC-02, #1688, Vermerk [0147](entscheidungen/0147-eigene-uebersetzungen.md)); welche gilt, ist eine Einstellung des Menschen |
| Eine eigene Übersetzung anlegen | Rechner | `lokalSchreiben` | — | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) — für den Rechner, nicht für ein Profil |
| Eine eigene Übersetzung bearbeiten | Rechner | `lokalSchreiben` | — | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)); geändert wird in einem Zug |
| Eine eigene Übersetzung löschen | Rechner | `lokalSchreiben` | — | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) — wer sie benutzt hat, fällt auf ihre Basis zurück |
| Eine eigene Übersetzung ein- und ausgeben | Rechner | `lokalSchreiben` | — | `query_only` für den Leser fürs Einlesen; das Ausgeben liest und schreibt eine Datei und fragt deshalb den **Riegel** ([0078](entscheidungen/0078-der-wirksame-akteur.md)) |
| Nutzer und Profile führen | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Das Profil wechseln | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Über einen Anbieter anmelden | Nutzer | `externSenden` | — | der Anmeldefluss schickt den Code an den Anbieter — für den Menschen, ohne Profil; der Browser ist der Riegel |
| Anbieterkonten führen | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Nutzer und Profile bearbeiten | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Ein eigenes Konto anlegen und sich damit anmelden | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Einen zweiten Faktor einrichten und verlangen | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Das Fenster sperren, entsperren, abmelden und lokal anmelden | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — die Sitzung gehört dem Fenster und dem Menschen, der arbeitet. **Gesperrt und abgemeldet bekommt keine Bindung ein Profil**, damit auch keine Rolle; anmelden verlangt das Kennwort und den zweiten Faktor des Ziels (ID-02, #1701) |
| Das eigene Kennwort wechseln oder über den Zettel wiederherstellen | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — **und kein Recht dazu**: die Wiederherstellung setzt nur das Kennwort, keine Rolle, keine Freigabe, kein Gegenüber (ID-02.AK-04) |
| Zugriffstoken für ein technisches Profil führen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) — **und die Rolle am Zielprofil** (#1817): wer ein Profil oder ein Token nennt, braucht sie an **dessen** Profil, auch zum Lesen der Liste. `query_only` gehört zum gezeigten, und aus dem eigenen Profil heraus entstand ein schreibendes Token für ein fremdes. **Rotation und Sammelwiderruf ebenso** (O2, [0080](entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md)); die Rotation gibt ihr Geheimnis nur am Fenster, an der Kommandozeile und am Dienst — und eine MCP-Sitzung mit Token prüft es **je Ruf** |
| Vorgaben eines Profils, die auf **Zeilen** zeigen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) — und die Rolle am **genannten** Profil (#1817). Lesen darf sie, wer im Profil steht; an einem anderen braucht es dieselbe Rolle wie das Setzen |
| Wer außer dem Eigner an einem Profil etwas darf | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Die Klasse verlangt `mitarbeiter`, das Vergeben mehr**: Setzen und Löschen gehen nur mit `verwalter` am **Zielprofil** (Werkstück 11, #1351; #1817) — der Eigner vergibt ohne eigene Zeile. **Das Lesen ist die mildere Wirkung**: die Liste liest im gezeigten Profil jede Rolle, an einem anderen braucht es `mitarbeiter` (#1817) — sie nennt Menschen und ihre Rollen. **Hier stand nur das Lesen**, und die Zeile las sich, als genüge `mitarbeiter` zum Vergeben (Zweitblick über Phase 6.1, Befund 7) |
| Ein Deck aus einem fremden Profil holen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Die Decks eines fremden Profils sehen | Profil | `lokalLesen` | leser | **Zwei Riegel**: der eigene Zutritt wie bei jeder lesenden Fähigkeit, und dazu `Nutzerverwaltung.Zutritt` am **fremden** Profil, bevor gelesen wird — dieselbe Prüfung wie die Übernahme ([0071](entscheidungen/0071-woher-ein-deck-kommt.md), UX-08.AK-04) |
| Ein Paket aus einem Gerüst anlegen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Ein Paket einbinden oder entfernen | Rechner | `lokalSchreiben` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Ein Paket als eigene Kopie anpassen | Rechner | `lokalSchreiben` | — | dito — die Kopie landet im verwalteten Verzeichnis des Rechners, das Original bleibt (FMT-02.AK-04, #1691) |
| Ein Paket als `.mse-installer` weitergeben und einlesen | Rechner | `lokalSchreiben` | — | dito — der Installer liegt, wo er gewählt wurde; eingelesen wird ins verwaltete Verzeichnis des Rechners, und was dort liegt, bleibt (FMT-02.AK-04, #1691) |
| Vor einer Paketänderung sehen, was sie bewirkt | Rechner | `lokalLesen` | — | dito; **die Sets, die er nennt, sind die des Profils**, das fragt (DOM-01) |
| Eine neue Fassung über ein Paket legen | Rechner | `lokalSchreiben` | — | dito; ein lokal geändertes Paket nur mit Bestätigung |
| Die vorige Fassung eines Pakets zurückholen | Rechner | `lokalSchreiben` | — | dito |
| Die Stände der bezogenen Pakete sehen und eine Aktualisierung prüfen | Rechner | `lokalLesen` | — | liest den Katalog; **ins Netz geht es nur auf Wunsch** (`fragen`, die Vorschau), und gesendet wird nichts als die Adresse des Repositoriums und zwei Commits — an GitHubs Auskunft, über denselben Klienten wie der Bezug (G04, #2251) |
| Ein bezogenes Paket aktualisieren und zurückholen | Rechner | `lokalSchreiben` | — | wie ein Bezug, dazu hebt es den bisherigen Baum auf und tauscht ihn zurück; wo die Vorschau es verlangt, nur mit Bestätigung |
| Sehen, welche Pakete einem Set fehlen | Profil | `lokalLesen` | leser | liest den Kopf **eines Sets dieses Profils** und den Katalog des Rechners; ändert nichts (DOM-01.AK-04) |
| Aus einer MSE2-Installation übernehmen | Rechner | `lokalSchreiben` | — | Pakete und Vorlagen gehören dem Rechner, nicht einem Profil |
| Eine Datei ansehen | Rechner | `lokalLesen` | — | sie liest eine Datei auf der Platte und nicht den Bestand — **am Dienst und über MCP-HTTP nur, was der Pfadvertrag zulässt**; im Finder zeigen und öffnen nur am Fenster, und geöffnet wird nur, was nichts ausführen kann (UX-13, #1685) |
| Den Bestand sichern | Bestand | `lokalSchreiben` | — | der Bestand gehört dem Rechner — jeder, der ihn öffnen kann; die Sicherung liegt an der schreibenden Verbindung (0073) |
| Sehen, wann der Bestand zuletzt gesichert wurde — und den freiwilligen Plan | Bestand | `lokalLesen` | — | der Bestand gehört dem Rechner — jeder, der ihn öffnen kann; gelesen wird nur der Vermerk, ein Pfad steht darin nicht (OPS01) |
| Den freiwilligen Plan einrichten und beenden — wöchentlich in einen Ordner, sieben eigene | Bestand | `lokalSchreiben` | — | nur am Fenster: der Ordner kommt aus dem Dialog, und gesichert wird der ganze Bestand; weggeräumt werden nur Sicherungen, die ein Vermerk wiedererkennt (OPS01) |
| Eine Sicherung zurückspielen | Bestand | `lokalSchreiben` | — | der Bestand gehört dem Rechner — jeder, der ihn öffnen kann; die Sicherung liegt an der schreibenden Verbindung (0073) |
| Einen gescheiterten Start erneut versuchen | Bestand | `lokalSchreiben` | — | er öffnet den Bestand und hebt ihn, wie der Start es täte — **keine Rolle**, weil es ohne offenen Bestand keine Sitzung gibt, deren Rolle gälte; dieselbe Lage wie beim Zurückspielen (F05, #2344) |
| Wer wo als was arbeitet — der **Arbeitskontext** | Nutzer | `lokalLesen` | — | kein Profil im Spiel — wer arbeitet, darf über sich lesen |
| Wer arbeitet, ansehen | Nutzer | `lokalLesen` | — | kein Profil im Spiel — wer arbeitet, darf über sich lesen |
| Ein anderer Nutzer werden | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Den Bestand verschlüsseln | Bestand | `lokalSchreiben` | — | nur am Fenster und an der Kommandozeile; der Bestand gehört dem Rechner, nicht einem Profil |
| Eine Änderung zurücknehmen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) |
| Den Verlauf eines Sets sehen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)) |
| Den Stand des Bestands lesen | Bestand | `lokalLesen` | — | der Bestand gehört dem Rechner — jeder, der ihn öffnen kann |
| Sehen, was die fremden Dienste hier können — der Dienstkatalog | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)). **Die Auskunft schickt nichts** — sie sagt, ob eine Anfrage ginge, und ob der Akteur nach draußen dürfte, steht als Status `gesperrt` **in** der Antwort statt sie zu verhindern (#1440) |
| Sehen, was dieser Prozess an Rufen und Parallelität zulässt — die Grenzen | Rechner | `lokalLesen` | — | eine Auskunft über das Programm, an keinem Bestand und keinem Profil — wer den Prozess erreicht, darf sie lesen; sie verrät nichts über einen Bestand |
| Die Kommandozentrale — einen Fachbefehl strukturiert wählen, mit Vorschau ausführen, abbrechen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) — der Katalog und ein lesender Befehl stehen jedem Leser offen; ein schreibender fragt `Pruefe(WirkungSchreiben)` **vor** der Bestätigung, und das Werkzeug selbst prüft danach noch einmal, wie es das über MCP täte |
| Die Skriptkonsole — einen laufenden Ausdruck abbrechen | Profil | `lokalSchreiben` | mitarbeiter | Der Abbruch beendet einen Lauf, den derselbe Akteur begonnen hat; er ändert nichts am Bestand, aber er greift in einen Vorgang ein — deshalb `lokalSchreiben` wie `BefehlAbbrechen`, und ein Leser, der nur liest, hat keinen Lauf, den er abbrechen müsste (U13, #1424) |
| Zugangsdaten fremder Dienste verwalten | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Kontakte aus einem Adressbuch übernehmen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Nur am Fenster**, und nichts geht nach draußen: gelesen wird die Datei, die ein Mensch im Dialog wählt, und eine übernommene Adresse wird weder geprüft noch angeschrieben (ID-04, [0137](entscheidungen/0137-kontakte-kommen-als-vcard.md)) |
| Das Geheimnis eines Anmeldeanbieters hinterlegen | Nutzer | `lokalSchreiben` | — | kein Profil im Spiel — wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor ([0074](entscheidungen/0074-wer-arbeitet-steht-fest.md)) |
| Sehen, was Platz belegt — und alte Sicherungen wegräumen | Bestand | `lokalSchreiben` | — | der Bestand gehört dem Rechner — jeder, der ihn öffnen kann; die Sicherung liegt an der schreibenden Verbindung (0073) |
| Sehen, welche Bereiche und Funktionen an sind — die Schalter | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)). **Geschaltet wird über die Einstellungen** und damit unter `lokalSchreiben`; **ein Schalter ist keine Autorisierung** ([0158](entscheidungen/0158-ein-schalter-verbirgt-er-sichert-nicht.md)) |
| Einstellungen lesen und setzen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Wohin diese Anwendung sendet, setzt nur, wer senden darf** (SEC-P4, #2000): `ki.adresse`, die Adresse eines KI-Anbieters und `netz.proxy.*` fragen den Riegel im Einstellungsverwalter nach `externSenden` — **ein Token setzt sie nie**, denn es reicht nie dorthin (0078). Bis dahin setzte ein Token `schreiben` das Ziel, das danach ein lesender Weg anwählte |
| **Einstellungen ausführen** | Profil | `lokalLesen` | leser | **Riegel im Austauschdienst** ([0079](entscheidungen/0079-einstellungsebenen-und-portabler-import.md)) — er fragt vor dem Lesen. **Kein `lokalSchreiben`, obwohl das Fenster eine Datei anlegt**: ohne Fenster schreibt dieser Weg **keine** Datei, sondern gibt Text; am Fenster ist der Speicherdialog der zweite Riegel — dieselbe Einordnung wie bei der Diagnose weiter oben |
| **Eine Einfuhr von Einstellungen vorschauen** | Profil | `lokalLesen` | leser | **Riegel im Austauschdienst** ([0079](entscheidungen/0079-einstellungsebenen-und-portabler-import.md)) — sie ändert nichts und darf deshalb, wer lesen darf |
| **Eine Einfuhr von Einstellungen übernehmen** | Profil | `lokalSchreiben` | mitarbeiter | **Riegel im Austauschdienst** ([0079](entscheidungen/0079-einstellungsebenen-und-portabler-import.md)) — **vor** dem Lesen des Stands und vor jeder Zeile, wie beim Setschreiber (0078). Daneben hält `query_only` den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Trägt die Auswahl eine Adresse, wohin gesendet wird, fragt er zusätzlich nach `externSenden`** (SEC-P4, #2000) — eine Einfuhr ist kein Umweg um den Riegel der Einstellung |
| Nach einer neueren Fassung sehen und sie holen | Rechner | `lokalSchreiben` | — | liest aus dem Netz und schreibt eine Datei; gesendet wird nichts als die Anfrage |
| Sehen, womit sich dieses Profil gegenüber anderen ausweist | Profil | `lokalLesen` | leser | liest den Bestand und den Schlüsselbund; der **private** Schlüssel verlässt die Naht nie — hinaus geht nur die öffentliche Kennung und ihr Abdruck (N1, #1454) |
| Die erste Identität anlegen | Profil | `lokalSchreiben` | mitarbeiter | schreibt den Schlüsselbund und den Bestand. **Nur, wenn es keine gibt** — bei Verlust und Widerspruch ist es ein Konflikt, und die Handlung heißt dann Rotation |
| Die Identität ablösen | Profil | `lokalSchreiben` | mitarbeiter | dasselbe, und zusätzlich entwertet es **jede** Bestätigung, die ein Kontakt gegeben hat. Deshalb steht es an drei Flächen und nicht an vier |
| Sehen, wen dieses Profil kennt | Profil | `lokalLesen` | leser | liest die Kontakte und ihre Freigaben; nach außen geht nichts (N1, #1454) |
| Ein Gegenüber von Hand eintragen | Profil | `lokalSchreiben` | mitarbeiter | trägt eine fremde Kennung als **unbestätigten** Kontakt ein; sie wird gegen die Form geprüft, und erlaubt wird damit nichts (N1, #1454) |
| Ein Gegenüber bestätigen | Profil | `lokalSchreiben` | mitarbeiter | trägt ein, dass ein **Mensch** einen Fingerabdruck verglichen hat. Der verglichene Abdruck reist mit und wird gegen die Kennung gehalten; er gibt **keinen** Zweck frei (N1, #1454) |
| Eine Einladung ausstellen | Profil | `lokalLesen` | leser | unterschreibt mit dem Schlüssel des Profils und **schreibt nichts**; die Einladung erlaubt niemandem etwas, bis ein Mensch auf der anderen Seite sie annimmt und vergleicht (NET-01, #1705) |
| Eine Einladung prüfen | Profil | `lokalLesen` | leser | liest Kontakte und angenommene Einladungen; **schreibt nichts** — die Vorschau ist eine Auskunft (NET-01, #1705) |
| Eine Einladung annehmen | Profil | `lokalSchreiben` | mitarbeiter | legt einen Kontakt an und verbraucht die Nonce **in einer Transaktion**; ein Recht nur mit passendem Vergleichscode, eine Sperre wird nicht übergangen — die Bedingung steht im `WHERE` (NET-01, #1705). **Eine einlösbare Einladung meldet sich danach beim Einladenden** (NET-01b) — an der Adresse, die **er** hineingeschrieben hat, mit dem eigenen Namen und der eigenen Adresse. Dafür fragt das Einladungswerk **vor** der Annahme nach `externSenden` (SEC-P4, #2000): ganz oder gar nicht, und **ein Token nimmt eine solche Einladung nicht an**. Bis dahin löste ein Token `schreiben` damit eine Verbindung zu beliebigem `host:port` aus |
| Ein Gegenüber sperren | Profil | `lokalSchreiben` | mitarbeiter | schließt es aus und nimmt ihm alle Rechte. **Die einzige Vertrauensänderung mit MCP-Werkzeug** — Sperren ist die Richtung, in der ein Irrtum nichts kostet (N1, #1454) |
| Eine Sperre zurücknehmen | Profil | `lokalSchreiben` | mitarbeiter | setzt auf `unbestaetigt` zurück, nicht auf den Stand von vorher: wer gesperrt war, wird neu geprüft (N1, #1454) |
| Einen Zweck freigeben | Profil | `lokalSchreiben` | mitarbeiter | erlaubt einem **bestätigten** Gegenüber genau einen Zweck. Die Bedingung steht **in** der Schreibtransaktion und nicht davor (N1, #1454) |
| Einen Zweck zurücknehmen | Profil | `lokalSchreiben` | mitarbeiter | nimmt ihn zurück; ein Zweck, der gar nicht freigegeben war, ist kein Fehler (N1, #1454) |
| Eintragen, wo ein Gegenüber erreichbar ist | Profil | `lokalSchreiben` | mitarbeiter | trägt eine Adresse ein, unter der gewählt wird. **Sie erlaubt nichts** — wer dort sitzt, sagt der Handschlag, und was er darf, die Rechte; deshalb setzt sie **keine** Bestätigung voraus. Eine leere löscht die eingetragene (N2, #1455) |
| Sich zu einem Gegenüber etwas notieren | Profil | `lokalSchreiben` | mitarbeiter | Freitext für den, der ihn schreibt. **Er erlaubt nichts und zeigt nichts**: eine Sichtkarte trägt ihn nicht, und eine Freigabe erst recht nicht. Eine leere Notiz löscht die vorhandene (T3, #1468) |
| Sehen, was ein Gegenüber sehen darf | Profil | `lokalLesen` | leser | nennt die erteilten Sichten eines Gegenübers samt ihren Umfängen. **Lesend, und nur über die eigenen**: eine Abfrage klemmt auf ein Profil (T3, #1468) |
| Einem Gegenüber etwas freigeben | Profil | `lokalSchreiben` | mitarbeiter | **Eine Zustimmung.** Standard ist geschlossen; fünf Umfänge werden einzeln erteilt und gelten nebeneinander, der **weitere** gewinnt. **Sie überstimmt keine Sperre**: ein gesperrter Kontakt sieht nichts, gleich welche Freigaben danebenliegen — das ist die Zusage aus N1, und hier wird sie eingelöst (T3, #1468) |
| Eine Freigabe zurücknehmen | Profil | `lokalSchreiben` | mitarbeiter | beendet **künftige** Antworten. Was ein Gegenüber schon kopiert hat, liegt bei ihm, und kein Widerruf holt es zurück; die Oberfläche sagt das, statt es offen zu lassen (T3, #1468) |
| Sehen, was ein Gegenüber wirklich sähe | Profil | `lokalLesen` | leser | die Vorschau auf die eigene Sammlung durch die Augen eines anderen — **über denselben Weg wie die Antwort an ihn**. Sie ist der Blick auf eine Zusage, der niemand glauben muss (T3, #1468) |
| Fragen, was ein Gegenüber zum Handeln freigibt — über die Leitung | Profil | `externSenden` | mitarbeiter | **`externSenden` und nicht `lokalLesen`**, obwohl hier nichts geschrieben wird: die Anfrage baut eine Leitung zu einem fremden Rechner und geht mit der Identität dieses Profils unterschrieben hinaus — dieselbe Wirkung wie `HandelZustellen` (P11, #2379) |
| Sehen, was zwei Sammlungen einander zu bieten haben | Profil | `lokalLesen` | leser | der Abgleich der **eigenen** Dubletten und Wünsche gegen einen Stand, den ein Gegenüber freigegeben hat. **Er liest und schreibt nichts**; der fremde Stand kommt als Eingabe herein und nicht aus dem Bestand (T4, #1469) |
| Sehen, was zu einem Gegenüber noch aussteht | Profil | `lokalLesen` | leser | Zähler über den eigenen Handelsausgang — **keine Nachrichten und keine Inhalte**. Er sagt, ob etwas klemmt, und nicht, worum es geht (T5, #1470) |
| Einen Handel eröffnen | Profil | `lokalSchreiben` | mitarbeiter | die erste Revision eines Vorgangs — **es bewegt keinen Bestand**, und der Initiator kommt aus dem Ausweis dieses Profils und nicht aus dem Ruf (T1, #1466, angeschlossen mit #1637) |
| Eine neue Revision vorschlagen | Profil | `lokalSchreiben` | mitarbeiter | die nächste Revision gegen den Stand, den der Bestand **jetzt** führt; ist er weitergezogen, kommt ein Konflikt. **Kein Bestand bewegt sich** (T1, #1466, #1637) |
| Einen Handel bis zur Übergabe führen | Profil | `lokalSchreiben` | mitarbeiter | die neun Schritte vor dem Abschluss — Zustandswechsel am eigenen Vorgang und sonst nichts. **Die drei verbindlichen gehen hier nicht**: sie stehen drei Zeilen tiefer und verlangen zusätzlich einen Nachweis (T1, #1466, #1637) |
| Sehen, wo ein Handel steht | Profil | `lokalLesen` | leser | die letzte Revision eines eigenen Vorgangs samt ihrem Abdruck — sie verlässt den Rechner nicht (T1, #1466, #1637) |
| Die Handel mit einem Gegenüber sehen | Profil | `lokalLesen` | leser | je eigenem Vorgang mit diesem Gegenüber die letzte Revision — dieselbe Auskunft wie eine Zeile darüber, nur ohne dass jemand die Kennung kennen muss (UX-08.AK-04, #1680) |
| Alle eigenen Handel sehen | Profil | `lokalLesen` | leser | je eigenem Vorgang die letzte Revision mit abgeleiteter Lage — nur Vorgänge, an denen dieses Profil beteiligt ist; sie verlässt den Rechner nicht (TRADE-01, #1715) |
| Sehen, wie ein Handel dorthin kam | Profil | `lokalLesen` | leser | die Zustandswechsel eines eigenen Vorgangs mit Revision und Zeitpunkt — geschrieben vom Bestand selbst und auf keinem Weg zu ändern; sie verlassen den Rechner nicht (TRADE-03.AK-03, #1717) |
| Sehen, wer einen Handel unterschrieben hat | Profil | `lokalLesen` | leser | Zustand und Beleg eines eigenen Vorgangs — Unterschrift und **wessen** geprüfter Gegenbeleg hier liegt; gelesen, ohne einen Schritt zu gehen, und sie verlassen den Rechner nicht (Hüllenprobe `nat04handel`, #1672) |
| Sehen, was die Seiten eines Tauschs ungefähr wert sind | Profil | `lokalLesen` | leser | Preise aus dem eigenen Preisbestand gegen die Zeilen eines eigenen Vorgangs oder Entwurfs gerechnet — nichts wird gespeichert, nichts verlässt den Rechner (TRADE-02b, #1951) |
| Einen Entwurf neu schreiben | Profil | `lokalSchreiben` | mitarbeiter | nur ein Entwurf, der nie hinausging — die Bedingung steht in der Anweisung der Ablage; **er bewegt keinen Bestand** (TRADE-01, #1715) |
| Einen Handel als neuen Entwurf kopieren | Profil | `lokalSchreiben` | mitarbeiter | ein neuer Entwurf mit diesem Profil als Initiator, ohne Stimmen — **keine Nachricht geht hinaus** (TRADE-01, #1715) |
| Einen Entwurf löschen | Profil | `lokalSchreiben` | mitarbeiter | nur ein Entwurf, der nie hinausging; was hinausging, bleibt als Beleg (TRADE-01, #1715) |
| Einen beendeten Handel archivieren | Profil | `lokalSchreiben` | mitarbeiter | nur ein Handel, der zu Ende ist, und nur eine Ansicht: **gelöscht wird nichts** (TRADE-01, #1715) |
| Den Stand eines Handels an ein Gegenüber schicken | Profil | `lokalSchreiben` | mitarbeiter | **Sie legt in den Ausgang und geht nicht hinaus** — deshalb `lokalSchreiben` und nicht `externSenden`: über die Leitung trägt sie niemand, und wer den Ausgang leert, ist ein anderer Weg (#1634). **Unterschrieben wird trotzdem hier**, mit der Identität dieses Profils (T5, #1470, #1637) |
| Welche Besitzzeilen für einen Handel weggehen | Profil | `lokalSchreiben` | mitarbeiter | **Es bindet Bestand und bewegt keinen** (#1643): eine Zusage sagt, **welches** Exemplar weggeht, und nimmt es niemandem weg — das tut erst die Verbuchung, und die verlangt `verbindlich`. **Deshalb reicht ein Token**: über MCP gibt es trotzdem nur den lesenden Weg, und zwar aus einem anderen Grund — die Wahl gehört einem Menschen |
| Den Ausgang über eine Leitung zu einem Gegenüber tragen | Profil | `externSenden` | mitarbeiter | **Und hier `externSenden` und nicht `lokalSchreiben`** — anders als die Zeile darüber, und das ist der ganze Unterschied (#1634): jene legt in den **eigenen** Ausgang, diese baut eine Verbindung zu einem fremden Rechner und schickt. **Geschrieben wird dabei nichts** — der Ausgang bleibt, wie er war; weg ist eine Zeile erst mit der Bestätigung von drüben |
| Eine angekommene Nachricht wegräumen | Profil | `lokalSchreiben` | mitarbeiter | eine Zeile aus dem eigenen Ausgang; eine Bestätigung ins Leere ist ein Nein (T5, #1470, #1637) |
| Sagen, was nach einem Abriss noch einmal hinausmuss | Profil | `lokalLesen` | leser | ein Vergleich zweier Fortschritte über dem eigenen Ausgang — **es schreibt nichts** und schickt nichts; es nennt Kennungen (T5, #1470, #1637) |
| Einen übergebenen Handel im eigenen Bestand verbuchen | Profil | `lokalSchreiben` | mitarbeiter | **Sie bewegt echtes Papier.** Abgang, Zugang, Journal, Zusagenabschluss und Beleg gehen in **einer** Transaktion; ein Abgang nur gegen eine bestätigte Übergabe passender Fassung. Sie gilt nur für den **eigenen** Bestand — was beim Gegenüber geschieht, entscheidet dessen Programm (T6, #1471) |
| Der Gegenseite melden, dass hier fertig ist | Profil | `lokalSchreiben` | mitarbeiter | ein Zustandswechsel am eigenen Vorgang und sonst nichts — kein Bestand, keine Zusage (T6, #1471) |
| Die zweite Unterschrift entgegennehmen | Profil | `lokalSchreiben` | mitarbeiter | **Erst wenn alle Parteien gestimmt haben**, wird der Beleg beidseitig. Eine Stimme wechselt nichts; zwei Unterschriften vorzutäuschen ist ausgeschlossen (T6, #1471) |
| Sehen, was sich im eigenen Bestand bewegt hat | Profil | `lokalLesen` | leser | das eigene Journal — mit Lagerort und Besitzzeile, weil es das eigene Buch ist und nicht der Beleg, der hinausgeht. **Es wird nichts herausgenommen**: eine falsche Buchung bekommt eine Gegenbewegung mit Bezug und Grund (T6, #1471) |
| Ein Deck aus einem Umschlag annehmen | Profil | `lokalSchreiben` | mitarbeiter | es entsteht ein **neues** Deck; überschrieben wird nichts, und ein schon vergebener Name bekommt einen sichtbaren Zusatz (T7, #1472) |
| Eine Karte aus einem Umschlag annehmen | Profil | `lokalSchreiben` | mitarbeiter | es entsteht eine **neue** Karte im **gewählten** Set; die Kennung aus dem Umschlag stammt aus einem fremden Bestand und wird nicht übernommen (T7, #1472, angeschlossen mit #1639) |
| Eine Vorlage aus einem Umschlag annehmen | Profil | `lokalSchreiben` | mitarbeiter | es entsteht ein **neues** Paket unter dem Namen, den ein Mensch wählt — ein vorhandenes ist ein **Konflikt** und kein Überschreiben (T7, #1472, #1639) |
| Eine empfangene Sammlungssicht ansehen | Profil | `lokalLesen` | leser | sie wird **gelesen und nicht gebucht** — was ein Gegenüber hat, ist kein Besitz dieses Profils (T7, #1472) |
| Sehen, wer an einem Tisch saß | Profil | `lokalLesen` | leser | ein Fingerabdruck aus der Sitzvergabe gegen die eigene Kontaktliste — es öffnet nichts und gibt nichts frei (T7, #1472) |
| Für andere erreichbar sein | Profil | `externEmpfangen` | mitarbeiter | öffnet einen Port auf diesem Rechner, unter der Identität **dieses** Profils. **Nur auf Aufruf**, nie von selbst, und ein Profilwechsel beendet es. **Die Klasse ist seit N7 die richtige** (#1568); hier stand `lokalSchreiben` mit dem Satz, dass sie weniger sagt, als geschieht. **Die Rolle bleibt dieselbe**: `externEmpfangen` steht in der Reihe **vor** `externSenden`, und ein Mitarbeiter reicht bis dorthin — das Issue verlangt es, und ein Lauf hält es fest. **Und sie wird jetzt auch gefragt**: bis N7 stand hier eine Mindestrolle, die **kein Ruf** prüfte; ein Leser in einem fremden Profil konnte einen Port öffnen. Das war keine Lücke, die N7 geschaffen hat, sondern eine, die es benannt hat |
| Eine Partie für Gäste öffnen | Profil | `externEmpfangen` | mitarbeiter | **Dieselbe Klasse wie eine Zeile höher, und der Grund liegt eine Stufe weiter**: dort wird dieser Rechner ansprechbar, hier darf ein fremder Rechner den **Stand einer Partie ändern**. Erreichbar sein und mitspielen lassen sind zwei Entscheidungen, und die zweite wiegt schwerer. **Nur das Freigeben trägt sie**: Beenden ist `lokalSchreiben` (es nimmt etwas zurück und öffnet nichts), Ablesen ist `lokalLesen`. **Welcher Sitz einem Gegenüber gehört, entscheidet der Wirt** — eine Sitzung belegt einen Rechner und keinen Sitz, und wer beides gleichsetzte, ließe einen Gast im Namen eines anderen ziehen (G5, #1462) |
| An einem fremden Tisch mitspielen | Profil | `externSenden` | mitarbeiter | **Die Gegenrichtung zur Zeile darüber, und eine andere Klasse**: dieser Rechner wird dabei **nicht erreichbar**, er **greift hinaus**. Wer beides gleich einordnete, verspräche etwas anderes, als geschieht (§10). **Nur Beitreten und Wünschen tragen sie**: Aufstehen ist `lokalSchreiben` (es nimmt zurück und greift nirgends hin), Ablesen ist `lokalLesen`. **Ein entzogener Zweck hält den Wunsch auf**, auch mitten in einer offenen Sitzung — das ist die Zusage aus N1, und hier wird sie eingelöst (G5, #1462) |
| Am Tisch reden und zeigen | Profil | `externSenden` | mitarbeiter | **Derselbe Riegel wie beim Wünschen**: ein Satz und ein Zeiger verlassen diesen Rechner. **Und dieselbe Klasse für beide Seiten des Tisches** — ob dieser Rechner rundruft oder an seinen Wirt schickt, ändert an der Reichweite der Daten nichts (G5, #1462) |
| Was von einer Partie bleibt | Profil | `lokalSchreiben` | mitarbeiter | **Schreibend, obwohl nichts gespielt wird.** Eine Einwilligung greift in keine Partie ein und erzeugt kein Ereignis; was sie ändert, ist, **wer ein Protokoll lesen darf** — und das ist eine Zusage über Daten. Das Löschen ebenso: es räumt Zustand, Strom und Freigaben weg. **Nicht `externSenden`**, denn nichts davon verlässt diesen Rechner: die Einwilligung wirkt erst, wenn jemand den Verlauf hier abruft (G5, #1462) |
| Wie die Leitung steht | Profil | `lokalLesen` | leser | **Lesend, obwohl es um eine Leitung geht.** Sie fragt den eigenen Stand ab und schickt dafür nichts hinaus — die Zahlen liegen ohnehin auf diesem Rechner. **Und sie trägt kein Geheimnis**: weder den Abdruck des Decks noch den des Stroms, damit ein Betrachter sie sehen darf, ohne dass daraus etwas über Karten folgt (G5, #1462) |
| Ein Deck verleihen und leihen | Profil | `externSenden` | mitarbeiter | **Die höchste der drei Wirkungen, die hier vorkommen, und sie entscheidet.** Das Anbieten legt ein Dokument ab, das einem fremden Peer etwas erlaubt — es verlässt diesen Rechner in dem Augenblick, in dem es geschrieben wird. Das Annehmen ist `externEmpfangen`, Widerruf und Rückgabe sind `lokalSchreiben`, das Lesen ist `lokalLesen`; **jeder Weg fragt nach seiner eigenen**, und ein Lauf hält das je Weg. Die Zeile nennt die stärkste, weil eine Fähigkeit eine Zeile hat (G6, #1463) |
| Eine Paarung an eine Partie binden | Profil | `lokalSchreiben` | mitarbeiter | **Schreibend, und nichts davon geht hinaus.** Binden, melden und übernehmen ändern den Bestand dieses Rechners; was über die Leitung kommt, kommt als Partie und hat dort seine eigene Wirkung. Das Lesen ist `lokalLesen`; **jeder Weg fragt nach seiner eigenen**, und ein Lauf hält das je Weg. Die Zeile nennt die stärkste, weil eine Fähigkeit eine Zeile hat (G7, #1464) |
| Über fremde Netze erreichbar sein | Profil | `externEmpfangen` | mitarbeiter | gibt eine **Einladung** aus, mit der jeder eine Verbindung aufbauen kann, der sie bekommt — über ein Relais, das jemand anderes betreibt. **Nur auf Aufruf**, nie von selbst, und ein Profilwechsel beendet es. **Dieselbe Klasse wie die Zeile darüber** — und hier wiegt sie am schwersten: eine Einladung kann jeder einlösen, der sie bekommt, und das Relais betreibt jemand anderes. **Was sie trotzdem nicht ist, ist `externSenden`**: es geht kein Byte hinaus, bis jemand anklopft (N7, #1568). **Jede Einladung stirbt mit dem Horchen**: der Serverschlüssel wird nicht aufbewahrt |
| In der Nähe gefunden werden | Profil | `externEmpfangen` | mitarbeiter | kündigt dieses Profil im selben Netz an — Peerkennung und selbstgewählter Name gehen an jeden, der zuhört. **Nur auf Aufruf**, nur mit laufendem Horchposten, und ein Profilwechsel beendet es. **Dieselbe Klasse wie beim Horchen** (N7, #1568) — und der Grenzfall, an dem sie sich bewähren muss: hier geht **doch** etwas hinaus, nämlich die Ankündigung. Sie nennt Peerkennung und selbstgewählten Namen und sonst nichts, und ihr Zweck ist, erreichbar zu **werden** — deshalb hier und nicht bei `externSenden` (N2, #1455) |
| Sehen, wer in der Nähe ist | Profil | `lokalLesen` | leser | fragt einmal ins Netz und wartet auf Antworten. **Es ändert nichts an diesem Rechner** — deshalb `lokalLesen`, obwohl ein Paket hinausgeht; wer den Bestand lesen darf, darf auch fragen, wer nebenan steht (N2, #1455) |
| Einen Gegenstand als Umschlag ablegen | Profil | `lokalSchreiben` | mitarbeiter | liest ein Set aus **diesem** Profil und schreibt eine Datei daneben. **`lokalSchreiben` und nicht `externSenden`**: die Datei liegt auf dieser Platte, und wer sie weitergibt, entscheidet danach — ein Umschlag ist ein Gegenstand, kein Versand (N4, #1457) |
| Sehen, was in einem Umschlag steht | Profil | `lokalLesen` | leser | packt in einen Wegwerfort aus, prüft und räumt ihn wieder weg. **Am Bestand ändert sich nichts** — das ist die Zusage, die diese Fähigkeit von der nächsten trennt (N4, #1457) |
| Einen Umschlag übernehmen | Profil | `lokalSchreiben` | mitarbeiter | legt ein **neues** Set in diesem Profil an. Es überschreibt nichts; die Kennungen sind neu, und was hereinkommt, ist vorher geprüft (N4, #1457) |
| Sehen, wie dieses Programm hinausgeht | Rechner | `lokalLesen` | — | liest die Einstellungen dieses Geräts und die Umgebung dieses Prozesses. **Träger ist der Rechner**, nicht das Profil: ein Proxy beschreibt das Netz dieser Maschine, und drüben beschriebe er eines, das es nicht gibt (N5, #1571) |
| Den Dienst für Skripte schalten | Profil | `externEmpfangen` | mitarbeiter | macht einen Port auf, unter dem ein Fremder anklopfen kann. **Träger ist das Profil und nicht der Rechner**: der Dienst bedient **dessen** Bestandsausschnitt, sein Zeugnis trägt **dessen** Identität, und ein Profilwechsel beendet ihn ([0115](entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)). **Dieselbe Klasse wie der Horchposten** (N7, #1568), und dieselbe Mindestrolle: wer Daten hinausschicken darf, darf auch einen Port öffnen |
| Den Werkzeugserver über HTTP schalten | Profil | `externEmpfangen` | mitarbeiter | dieselbe Klasse und derselbe Grund wie in der Zeile darüber (N6, #1572) |
| Eine Partie anfangen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Und die Rolle wird gefragt, bevor die Decks gelesen werden**: der erste Anlauf baute erst den Tisch und fragte danach — ein Leser bekam damit sein Nein, nachdem das Programm zwei fremde Decks eingelesen hatte. **`lokalSchreiben` und nicht `externEmpfangen`**: eine Partie bleibt auf diesem Rechner; wer **mit anderen** spielt, öffnet einen Port, und das ist der Horchposten aus N2 mit seiner eigenen Zeile (G1, #1458) |
| Sehen, welche Partien es gibt | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)). **Eine Liste ist keine Sicht** (#1600): sie nennt Sitze, Stand und die **Zahl** der Objekte — keinen Kartennamen. Wer sie liest, erfährt, **dass** es eine Partie gibt; was darin liegt, sagt die Zeile darüber, und die rechnet je Sitz |
| Sehen, wie eine Partie für einen Sitz aussieht | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)). **Die Rolle ist hier die äußere Schale und nicht die Sichtgrenze**: *ob* jemand die Partie lesen darf, entscheidet das Profil; *wie viel* er sieht, entscheidet die Projektion `domain.Partie.Sicht` — und die kennt keine Rollen, sondern Sitze. **Eine Rolle macht keine Sicht größer**: ein Verwalter sieht nicht mehr als ein Leser (G1, #1458). **Welchen Sitz, nennt der Rufer** — an allen vier Flächen, denn ein Profil ist kein Nachweis über einen Menschen (G8, #1465). Wer die Partie lesen darf, darf also **jede** Sitzsicht wählen, auch die, in der eine fremde Hand offen liegt. **Das ist keine Grenze gegen den Menschen an diesem Rechner**: hier hält der Wirt den vollen Zustand ([0116](entscheidungen/0116-eine-partie-zeigt-eine-sicht-und-nie-den-zustand.md) §2), und die Ausfuhr trägt ihn auch. **Die Grenze, die hält, ist die zum Gast**: über die Leitung geht nur die Projektion **seines** Sitzes (`tischbildFuer`, `Ereignis.Fuer`). **Hier stand bis zum Zweitblick über Phase 6.1 nur der Satz über die Rolle** — und er las sich wie eine Zusage über die Sitzwahl (Befund 2). **Ein Werkzeug, das an einem Sitz spielt, bekommt ein an diesen Sitz gebundenes Token** (#1819): Partie und Sitz entstehen mit dem Token und ändern sich nicht. Mit ihm liefern Sicht, Verlauf, Wirtsstand und Befehle **nur** diesen Sitz, jeder andere gibt `erlaubnis` (403); eine andere Partie ist zu, und den **ganzen Zustand** — Ausfuhr, Kopie, Einfuhr, Löschen, Vergessen — bekommt es nie. Gesetzt wird die Bindung von der Tokenwache, also an MCP und am Dienst; am Fenster und an der Kommandozeile gibt es kein Token und damit keine Bindung. **Angelegt** wird ein gebundenes Token an der Kommandozeile (`mse token-anlegen`, `partie` und `sitz` im JSON) oder am Dienst (`POST /profile/{id}/token`); Ändern und Rotieren lassen die Bindung stehen. Die Tokenliste des Fensters führt sie mit, **zeigt sie aber nicht an**. Ein ungebundenes Token verhält sich wie zuvor |
| Sehen, was in einer Partie geschehen ist | Profil | `lokalLesen` | leser | **Dieselbe Aufteilung wie eine Zeile höher, und sie greift hier später**: die Rolle sagt, *ob* jemand den Verlauf lesen darf; *wie viel* er sieht, entscheidet `domain.RueckschauFuer` — und die kennt keine Rollen, sondern Sitze. **Ein Verwalter sieht nachträglich nicht mehr als ein Leser**, und das ist der Punkt: am Tisch filtert die Projektion, im Protokoll danach die Rückschau. Ohne sie wäre jede Rolle, die lesen darf, nachträglich allwissend (G5, #1462). **Den Sitz nennt auch hier der Rufer** — dieselbe Einschränkung wie eine Zeile höher: die Rückschau begrenzt je Sitz, nicht je Mensch |
| Einen Spielbefehl stellen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)), und **danach** die zweite Frage, die keine Rollenfrage ist: ob der Absender überhaupt an diesem Tisch sitzt. Beide Neins sind verschieden, und beide werden gebraucht (G1, #1458) |
| Eine Partie anhalten, fortsetzen oder beenden | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Am Fenster drei Bindungen, an den anderen Flächen ein Weg mit drei Worten** — dieselbe Klasse, dieselbe Rolle, dieselbe Prüfung; der Unterschied ist die Gestalt (G1, #1458) |
| Eine Partie löschen oder kopieren | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Beides schreibt**, auch das Kopieren: es legt eine Partie an. **Welche gelöscht werden darf, ist keine Frage der Rolle, sondern der Partie** — `domain.Loeschbar` weist eine beendete und eine mit Einwilligung ab, und zwar für jeden (PLAY-03.AK-05, #1711) |
| Eine Partie ausführen und einlesen | Profil | `lokalSchreiben` | mitarbeiter | `query_only` für den Leser ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)). **Zwei Wirkungen an einer Fähigkeit, und die strengere gilt**: die Ausfuhr ist `lokalLesen`, die Einfuhr legt eine Partie an. Eine Zeile, die die mildere nennte, ließe eine Prüfung an der anderen vorbeilaufen (§10). **Die Ausfuhr trägt den vollen Zustand** — auch fremde Hände; wer sie bekommt, bekommt mehr als jede Sicht, und deshalb steht sie unter derselben Erlaubnis wie der Bestand selbst (PLAY-03.AK-05, #1711) |
| Sehen, was von den Spielregeln durchgesetzt wird | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)). **Rein lesend, und trotzdem eine Sperre**: sie sagt, dass `streng` nicht zur Wahl steht, und dieselbe Antwort kommt an allen vier Flächen — eine Sperre, die nur das Fenster kennt, ist keine (G4, #1461) |
| Einen eigenen Tischhintergrund mitbringen | Nutzer | `lokalSchreiben` | — | **Es gehört dem Menschen und nicht dem Profil** — wer seinen Tisch so sehen will, sieht ihn so in jedem seiner Profile. **Kein Profil im Spiel** — wer arbeitet, darf über sich schreiben. **Und keine Rolle**: ein Hintergrund ändert nichts am Spiel, an keinem Bestand und an keiner Karte; ihn an eine Rolle zu binden hieße, Gestaltung mit Verantwortung zu verwechseln (G3, #1460) |
| Ein eigenes Fensterbild mitbringen | Nutzer | `lokalSchreiben` | — | **Dieselbe Klasse wie der Tischhintergrund darüber**, und aus demselben Grund: das Bild gehört dem Menschen, und es ändert nichts am Spiel, an keinem Bestand und an keiner Karte. **Ob es erscheint**, sagt die Einstellung `hintergrund` des Profils — die schreibt, wer das Profil schreiben darf (UX-09, #1681) |
| Einen Abzug des Tisches sichern | Nutzer | `lokalSchreiben` | — | **Geschrieben wird an einen Ort, den ein Mensch im Dialog gewählt hat** — nicht in den Bestand und nicht in eine Partie. **Kein Profil im Spiel**: ein Bild des eigenen Tisches gehört dem, der davorsitzt. **Und keine Rolle** — derselbe Satz wie eine Zeile darüber: ein Abzug ändert nichts am Spiel, an keinem Bestand und an keiner Karte. **Was er zeigt, zeigt er ohnehin**: gezeichnet wird die Projektion dieses Sitzes, und eine verdeckte Karte bleibt auch im Bild verdeckt (PLAY-05.AK-02, #1713) |
| Eine Karte am Spieltisch zeichnen | Profil | `lokalLesen` | leser | Zutritt: ohne Rolle kein Profil ([0070](entscheidungen/0070-sichtbarkeit-und-rollen.md)). **Dieselbe Klasse wie *Eine Karte zeichnen***: es entsteht ein Bild im Arbeitsspeicher, und auf der Platte ändert sich nichts. **Auch der Eintrag im Texturfach ist keiner**, sondern ein Zwischenspeicher mit Byte-Budget — wer ihn für Schreiben hielte, müsste jeden Puffer dafür halten. **Drei Bindungen, eine Fähigkeit**: `Spieltextur` fragt nach einer Seite **dieser** Karte, `Rueckentextur` nach dem Rücken **dieses** Spiels, `Fremdkartenbild` den Wirt nach dem Bild eines Objekts am fremden Tisch (P9, #2377); alle lesen, alle brauchen dieselbe Rolle (G2, #1459) |

## Die Rollenmatrix, gefahren

**ID-03.AK-01**: Rolle × Eigentümer/Fremdprofil × Fläche × Wirkung, **an der
Wirkung** geprüft — ob ein Deck entstand, ob eine Datei daliegt, ob eine
Gegenstelle gefragt wurde. Jedes Nein hat sein Ja daneben: das eigene Profil
und der Mitarbeiter fahren dieselben Wege und kommen durch.

| Lage | Wo gefahren |
| --- | --- |
| Kommandozeile, MCP, Dienst × eigenes Profil, Mitarbeiter, Leser, ohne Rolle × lesen, Bestand schreiben, Datei schreiben, hinausschicken | `internal/flaechen/rollenmatrix_test.go` |
| Fenster × ohne Rolle × Deck, Set, Turnier, Profiltext, Kartenname, Ersetzen, Schlüsselwort | `internal/app/ohnerolle_test.go` |
| Fenster × Leser × fünf Wege, die eine Datei schreiben | `internal/app/leserausfuhr_test.go` |
| Fenster × Leser × was hinausgeht | `internal/app/leseraussen_test.go` |
| Fenster × fremder Mensch mit Faktor × Faktor, Nutzer, Profil | `internal/app/fremdernutzer_test.go` |
| alle vier Flächen × Entzug im laufenden Betrieb | `internal/app/entzug_test.go`, `cmd/mse-mcp/entzug_test.go`, `cmd/mse-rest/entzug_test.go` |
| Dienst × Token mit Lese- und Schreibumfang | `internal/flaechen/autorisierungsmatrix_test.go` |

**Was das Issue daneben nennt, und wo es steht:**

- **Vorschau** — liest und steht damit in jeder Lage offen; die
  Ersetzungsvorschau neben dem Ersetzen, das nur schreibt, wer darf
  (`ohnerolle_test.go`).
- **Suche** — die Scryfallsuche als `externSenden` in der Matrix; die lokale
  Suche liest.
- **Tags** — **gibt es in dieser Anwendung nicht**; in `FAEHIGKEITEN.md`
  nachgesehen am 2026-09-23.
- **Exporte** — MTG JSON in der Matrix, dazu die fünf Dateiwege am Fenster.
- **Logs und Diagnosepakete** — den Protokollteil bekommt nur ein Verwalter
  (`darfDasProtokollSehen`, `TestEinGastBekommtDasProtokollNichtInsPaket`).
- **Sicherungen** — Träger ist der Bestand: sichern kann, wer ihn öffnen
  kann, und die Sicherung trägt alle Profile — ohne Kennwort im Klartext,
  mit Kennwort verschlüsselt (unten, [`SICHERUNG.md`](SICHERUNG.md)).

## Bedrohungsmodell — wovor die Rollen schützen und wovor nicht

**ID-03.AK-03** (#1702), Stand 2026-09-23. **Profile und
Rollen trennen Arbeit — sie trennen nicht Menschen, die dieselbe Anmeldung am
Mac teilen.** Was darüber hinausgeht, leisten der zweite Faktor, die
Dateirechte des Systems und die Verschlüsselung, und jede nur zum Teil.

| Wer | Was er kann | Was ihn aufhält |
| --- | --- | --- |
| Ein zweiter Mensch am selben Fenster, ohne Freigabe, in einem Profil, dessen **Eigner keinen Faktor** hat | hineinwechseln und alles lesen, was das Profil zeigt; mit einem Nutzerwechsel sogar dieser Eigner **werden** | beim Schreiben als er selbst: `query_only` und die Riegel. Gegen den Nutzerwechsel **nichts** — ohne Faktor gibt es keinen Ausweis (0074) |
| Derselbe, wenn der **Eigner einen Faktor** hat | ohne Code oder Freigabe nicht hinein, und weder dessen Faktor noch Profile, Bild, Einstellungen, Konten oder Personendaten ändern | der Faktor am Wechsel; `UeberNutzer` an jeder Bindung über einen Menschen |
| Ein **Leser** (Freigabe) | lesen; nicht schreiben, nicht exportieren, nicht hinausschicken | `query_only` für den Bestand, die Riegel für Dateien und Netz |
| Einer, dem die **Freigabe entzogen** wird | ab dem **nächsten** Aufruf nichts mehr, was die Freigabe gab — ein Ruf, der schon läuft, führt zu Ende (0078) | die Revision der Rechte am Fenster, die Schreibwache je Ruf an MCP und Dienst, der Akteur je Aufruf an den Riegeln |
| Ein **Token** (Dienst, MCP über HTTP) | nie mehr als die Rolle seines Profils und nie nach draußen | `Akteur.Darf` je Aufruf |
| **Derselbe macOS-Nutzer mit einem anderen Programm** — `sqlite3`, ein Skript, der Finder | **alles**: jedes Profil lesen und ändern, Freigaben setzen, den Bestand kopieren | **nichts in dieser Anwendung** (unten) |
| Ein anderer macOS-Nutzer, ein anderer Rechner | die Datei nicht lesen, wenn sie verschlüsselt ist; sonst, was die Rechte erlauben | das Verzeichnis mit 0700 (die Datei selbst trägt 0644) und die Verschlüsselung der **ruhenden** Datei (0072) |
| Wer eine **Sicherung** bekommt | ohne Kennwort den ganzen Bestand im Klartext, alle Profile; mit Kennwort nichts | das Kennwort (age, scrypt) — und ein verschlüsselter Bestand gibt eine Sicherung ohne Kennwort nur auf ausdrücklichen Wunsch her (OPS-06, `SICHERUNG.md`) |
| Wer ein **Diagnosepaket** bekommt | den Protokollteil nur, wenn die Rolle reichte | die Rolle am Protokollteil (O9) |

### `query_only` ist ein zusätzlicher Schreibschutz

**Die erste Linie sind der Akteur und die Riegel**: jeder Weg, der eine Datei
schreibt oder etwas hinausschickt, fragt `Akteur.Darf`, und seit ID-03 fragen
MCP-Server und Dienst es vor jedem schreibenden Ruf. **`query_only` fängt, was
davor vergessen wurde** — eine Bindung ohne Riegel schreibt über eine
Verbindung, die nichts annimmt (0073).

**Was es nicht ist:**

- **Keine Zeilenrechte.** SQLite kennt keine; eine Verbindung mit `query_only`
  liest jede Zeile der Datei. Dass ein Profil nur seine eigenen Sets sieht,
  sagen die Abfragen dieses Programms (`WHERE profil_id = ?`), nicht die
  Datenbank — **und seit #1815 auch für jede Karte, jeden Deckplatz, jede
  Runde und jedes Partieereignis**, die über eine Kennung vom Aufrufer
  kommen: die Ablage fragt, ob ihre Wurzel dem gezeigten Profil gehört
  ([0131](entscheidungen/0131-ein-abkoemmling-fragt-nach-seiner-wurzel.md)).
  Bis dahin schrieb ein Nutzer aus dem eigenen Profil den Kartennamen eines
  anderen um, an der Rolle vorbei, die `query_only` für das **gezeigte**
  Profil setzt.
- **Kein Schutz gegen denselben macOS-Nutzer.** `query_only` ist eine
  Eigenschaft einer Verbindung **dieses** Programms. Wer die Datei mit einem
  anderen öffnet, öffnet sie ohne, und das System lässt ihn: sie gehört ihm.
  (Nachgesehen am 2026-09-23: `bestand.db` trägt 0644; was andere Nutzer
  aussperrt, ist das Verzeichnis darüber mit 0700.) **Auch die
  Verschlüsselung hilft hier nicht**: sein Schlüsselbund gibt ihm den
  Schlüssel, genau wie diesem Programm.

### Was ein Gast lesen kann

**Ein Mensch, der in einem fremden Profil nur liest** — als Leser oder ohne
Rolle —, sieht, was die Seiten zeigen, und dazu zwei Dinge über den Eigner,
die nicht vom Profil abhängen: die **Anbieterkonten** (die Nutzerseite zeigt
sie für den Eigner des gezeigten Profils) und die **Nutzereinstellungen**, die
kein Geheimnis tragen. Beide stehen als benannte Ausnahme im Wächter
`uebernutzer_test.go`.

### Was „der nächste Aufruf" heißt

**Laufende Arbeit behält ihren Ursprung** (0078). Ein Auftrag, der vor dem
Entzug begann — ein langer Import, ein KI-Auftrag —, schreibt zu Ende; ein
Formular, das danach abgeschickt wird, ist ein neuer Aufruf und scheitert.

