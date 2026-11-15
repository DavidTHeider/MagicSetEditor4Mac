// Die Übersicht, die `mse` ohne Argumente und auf `mse hilfe` ausgibt.
//
// **Eine eigene Datei**, seit `lauf.go` an der Zeilengrenze aus §2 stand. Sie
// ist ein Text und keine Ablaufsteuerung; beide zusammenzulassen hieße, jede
// neue Zeile Hilfe gegen eine Zeile Code zu tauschen.

package main

// hilfe ist, was `mse` ohne Argumente und auf Verlangen sagt.
//
// **Deutsch wie jede andere Meldung** (§9). Die Befehlsnamen selbst bleiben,
// wie sie sind: sie stehen in Skripten.
const hilfe = `mse — MagicSetEditor4Mac auf der Kommandozeile

Aufruf:
  mse <befehl> [argumente] [--json]

Befehle:
  sets                              die Sets im Bestand auflisten
  set-neu <name> <spiel> [stil]     ein leeres Set anlegen
  set-loeschen <id> <name>          ein Set ENDGÜLTIG löschen — der Name bestätigt
  karten <set>                      die Karten eines Sets auflisten
  feld <set> <karte> <feld> <wert>  einen Feldwert setzen
  felder <set>                      die Feldnamen eines Sets nennen
  felder-setzen <set> <feld> <wert> <karte>...
                                    ein Feld auf mehreren Karten setzen, in
                                    EINEM Zug; --erwartet=<json> prüft vorher
  felder-vorschau <set> <feld> <wert> <karte>...
                                    zeigen, was felder-setzen je Karte änderte
  karte <set> anlegen               eine leere Karte anlegen
  karte <set> kopieren <karte>      eine Karte samt Werten kopieren
  karte <set> loeschen <karte>      eine Karte löschen
           [--trotzdem]             auch, wenn sie in einem Deck liegt
  karte <set> tauschen <a> <b>      zwei Karten in der Reihenfolge tauschen
  suche <text> [--set=<set>]        im Volltext über alle Kartenfelder suchen
  setangaben <set>                  die Angaben eines Sets zeigen
  setangaben <set> setzen <f> <w>   eine Angabe setzen (Titel, Setcode, …)
  vorbehalte <set>                  zeigen, was beim Einfahren nicht gedeutet wurde
  probleme <set>                    jede Karte zeichnen und die mit Problemen nennen
  stile <set>                       die wählbaren Stile eines Sets zeigen
  stile <set> folgen <paket>        zeigen, was ein Wechsel kostet
  stile <set> wechseln <paket>      den Stil wechseln
  stiloption <set>                  die Optionen des gewählten Stils zeigen
  stiloption <set> setzen <f> <w>   eine Stiloption setzen
  stiloption <set> zuruecksetzen <f>  eine Stiloption auf die Vorgabe stellen
  zusatzfeld <set> <karte>          die Zusatzfelder des Stils einer Karte zeigen
  zusatzfeld <set> <karte> setzen <f> <w>
                                    den Wert eines Zusatzfelds setzen
  pakete                            die installierten Pakete auflisten
  paket-beziehen <adresse> <stand> <unterverzeichnis> <name>
                                    ein Vorlagenpaket beziehen
  paket-installieren <quelle>       ein Paket von der Platte einbinden
  nutzer                            die Nutzer samt ihren Profilen
  nutzer-neu <name> [bild]          einen Nutzer samt erstem Profil anlegen
  profil-neu <nutzer> <name>        ein weiteres Profil für einen Nutzer
  profil [kennung] [code]           das Profil nennen, in dem gearbeitet wird — oder
                                    wechseln; der Code ist der zweite Faktor
  anbieter                          die Anbieter nennen, über die man sich anmelden kann
  anmelden <anbieter>               über einen Anbieter anmelden — öffnet den Browser
  konten <nutzer> [anbieter]        die Anbieterkonten nennen — oder eines lösen
  konto-neu <kennung> [name]        ein Konto auf eine Adresse oder Rufnummer —
                                    das Kennwort kommt von der Standardeingabe
  konto-anmelden <kennung> [code]   unter einer Kennung anmelden — das Kennwort
                                    kommt von der Standardeingabe
  konto-anonym                      ein Konto ohne jede Angabe
  sitzung                           der festgehaltene Stand der Sitzung
  kennwort-wechseln                 ein neues Kennwort für den, der arbeitet —
                                    altes und neues je Zeile auf der Eingabe
  kennwort-wiederherstellen <nutzer>
                                    ein vergessenes Kennwort über den Zettel —
                                    Code und neues je Zeile auf der Eingabe
  faktoren <nutzer>                 die zweiten Faktoren nennen
  faktor-neu <nutzer> totp|biometrie
                                    einen zweiten Faktor einrichten — bei totp
                                    stehen Geheimnis und Codes nur EINMAL da
  faktor-bestaetigen <nutzer> <code>
                                    den totp-Faktor eintragen; erst danach gilt er
  faktor-loeschen <nutzer> <art>    einen zweiten Faktor herausnehmen
  nutzer-aendern <id>               die Angaben eines Nutzers setzen — JSON
                                    von der Standardeingabe, wie »mse nutzer«
  nutzer-bild <id> <pfad> <fassung> ein Profilbild setzen — »--keines« als Pfad
                                    nimmt es weg. Das Bild landet als PNG mit
                                    höchstens 256 Punkten Kantenlänge
  nutzereinstellung <id> [<schlüssel> <wert>]
                                    die Einstellungen eines MENSCHEN nennen
                                    oder setzen — sie gelten in jedem Profil
  personendaten <id>                ausgeben, was über einen Menschen
                                    gespeichert ist — ohne Kennwort, als JSON
  nutzer-loeschen <id> <name>       einen Nutzer ENDGÜLTIG löschen — der Name bestätigt
  profil-aendern <id>               die Angaben eines Profils setzen — JSON
                                    von der Standardeingabe, wie »mse nutzer«
  profil-badge <id> <pfad> <fassung>
                                    das Abzeichen eines Profils setzen —
                                    »--keines« als Pfad nimmt es weg
  profil-loeschen <id> <name>       ein Profil ENDGÜLTIG löschen — der Name bestätigt
  token <profilID>                  die Zugriffstoken eines Profils nennen —
                                    ohne ihre Geheimnisse
  token-anlegen <profilID>          ein Zugriffstoken anlegen (JSON von der
                                    Standardeingabe); der KLARTEXT kommt EINMAL
  token-aendern <id>                Bezeichnung, Erlaubnis und Ablauf setzen —
                                    das Geheimnis bleibt dasselbe
  token-loeschen <id>               ein Zugriffstoken zurückziehen
  token-rotieren <id>               ein neues Geheimnis geben — der KLARTEXT
                                    geht nach stdout und ist danach fort
  token-alle-widerrufen <profilID>  jedes Token eines Profils zurückziehen
  token-ablaufwahlen                die Fristen nennen, die zur Wahl stehen
  vorgaben <profilID>               die Vorgaben eines Profils nennen — Format,
                                    Spiel, Vorlage, Decks und Lagerorte
  vorgabe-setzen <profilID> <art>   die Werte EINER Art setzen — JSON von der
                                    Standardeingabe; leer räumt sie ab
  deck-uebernehmen <profilID> <deckID>
                                    ein Deck aus einem fremden Profil holen
  decks-uebernehmbar <profilID>     die Decks dort nennen — zur Wahl
  freigaben <profilID>              nennen, wer an diesem Profil welche Rolle
                                    hat — samt der eigenen
  freigabe-setzen <profilID> <nutzerID> <rolle>
                                    eine Rolle vergeben — leser, mitarbeiter
                                    oder verwalter; NUR ein Verwalter darf das
  freigabe-loeschen <profilID> <nutzerID>
                                    eine Freigabe zurückziehen
  paket-anlegen mse-game <name> [vollname]
                                    ein leeres Spielpaket anlegen
  paket-anlegen mse-style <name> <spiel> [vollname]
                                    ein leeres Stilpaket für dieses Spiel
  paket-anlegen mse-include <name> [vollname]
                                    ein leeres Include
  paket-anlegen mse-export-template <name> <spiel> [vollname]
                                    eine Exportvorlage, die je Karte den Namen ausgibt
  paket-entfernen <art> <name>      ein Paket wieder herausnehmen
  paket-kopieren <art> <name> [neu] eine eigene Kopie anlegen — das Original
                                    bleibt, wie es ist
  paket-weitergeben <ziel> <paket>...
                                    Pakete als .mse-installer schreiben, etwa
                                    probe.mse-game; ein Ordner nimmt den ersten
  installer-einlesen <datei>        die Pakete eines .mse-installer einlesen;
                                    vorhandene bleiben
  paket-plan <vorgang> <art> <name>
                                    zeigen, was entfernen oder zurueck bewirkt
  paket-plan einbinden <quelle>     zeigen, was ein Einbinden bewirkt
  paket-ersetzen <quelle> [--bestaetigt]
                                    eine neue Fassung einlegen; die bisherige
                                    bleibt als vorige
  paket-zurueck <art> <name>        die vorige Fassung zurückholen
  paketbezuege [--fragen]           die bezogenen Bäume und ihre Stände — mit
                                    --fragen samt dem neuesten der Quelle
  paketbezug-vorschau <name> <commit>
                                    zeigen, was eine Aktualisierung täte
  paketbezug-aktualisieren <name> <commit> [--bestaetigt]
                                    auf einen Commit aktualisieren; der
                                    bisherige Baum bleibt aufgehoben
  paketbezug-zurueck <name>         den aufgehobenen Stand zurückholen
  paketbedarf <set>                 zeigen, welche Pakete einem Set fehlen
  set-vorschau <pfad>               zeigen, was das Einfahren einer Setdatei
                                    täte — ohne zu schreiben
  mse2-orte                         die üblichen Orte einer MSE2-Installation
  mse2-uebernehmen <verzeichnis>    die Vorlagen von dort übernehmen
  statistik <set> [feld]            ein Set auszählen — ohne Feld die Kennzahlen
  verlauf <set>                     die Schritte des Sets nennen — der jüngste
                                    zuerst, mit dem Vorgang für zuruecknehmen
  zuruecknehmen <set> [vorgang]     eine Änderung rückgängig machen
  wiederherstellen <set> [vorgang]  sie wieder ausführen
  einstellungen                     die Einstellungen zeigen
  einstellungen setzen <s> <w>      eine Einstellung setzen
  einstellungen ausfuhr [ebene…]    sie als JSON ausgeben
  einstellungen vorschau <datei>    zeigen, was eine Datei änderte
  einstellungen einfuhr <d> <rev>   sie übernehmen
  dienstkatalog                     die Fähigkeiten der fremden Dienste samt Bereitschaft
  grenzen                           die Grenzen dieses Prozesses: Kosten, Guthaben, Parallelität
  peer                              womit sich dieses Profil gegenüber anderen ausweist
  peer-anlegen                      die erste Identität anlegen — nur, wenn es keine gibt
  peer-rotieren                     sie ablösen; jeder Kontakt muss neu prüfen
  kontakte                          wer bekannt ist und was er darf
  kontakt-aufnehmen <kennung> <name>  ein Gegenüber eintragen (unbestätigt)
  kontakt-bestaetigen <kennung> <abdruck>  den Abdruck vergleichen
  kontakt-sperren <kennung>         ausschließen, samt allem, was er durfte
  einladung-erstellen <name> <zweck> [--adresse <adresse>]
                                    eine unterschriebene Einladung ausstellen;
                                    mit der Adresse von horchen einlösbar
  einladung-pruefen <code>          sehen, was eine Einladung sagt
  einladung-annehmen <code> <name> [<vergleichscode> [freigeben]] [--mein-name <name>]
                                    sie annehmen; mit Code bestätigt
  kontakt-entsperren <kennung>      auf unbestätigt zurück
  kontakt-adresse <kennung> <adr>   wo es erreichbar ist (leer löscht)
  kontakt-notiz <kennung> <text>    was man sich notiert (leer löscht)
  kontakt-freigaben <kennung>       was ein Gegenüber sehen darf
  kontakt-freigeben <kennung> <zweck> <umfang> [behalten|druck…]
                                    eine Sicht erteilen (behalten: Vorgabe 1)
  kontakt-widerrufen <kennung> <zweck> <umfang>
                                    sie zurücknehmen
  kontakt-sicht <kennung> <zweck>   was es wirklich sähe
  sicht-anfragen <kennung>          was es dir zum Handeln freigibt —
                                    über die Leitung
  horchen <adresse>                für andere erreichbar sein; bleibt, bis
                                   ⌃C kommt, und nennt die wirkliche Adresse
  fernhorchen                      über fremde Netze erreichbar sein; bleibt,
                                   bis ⌃C kommt, und nennt die Einladung
  handelszustellung <kennung>      den Ausgang über eine Leitung hinübertragen
  handel-festhalten <vorgang> <besitzzeile> <menge> <fassung>
                                   welches Exemplar weggeht
  handelszusagen <vorgang>         was dafür festgehalten ist
  handelspost <kennung>            was zu einem Gegenüber noch aussteht
  handelspost-schicken <vorgang> <kennung>
                                   den heutigen Stand in den Ausgang legen
  handelspost-bestaetigen <nachrichtID>
                                   eine Nachricht wegräumen, die ankam
  handelspost-wiederaufnahme <vorgang> <kennung> <revision> <abdruck> <folge>
                                   was nach einem Abriss noch einmal hinausmuss
  handel eroeffnen <vorgang> <art> <kennung>
                                   einen Vorgang anlegen, Positionen von der
                                   Standardeingabe; "" vergibt die Kennung
  handel vorschlagen <vorgang>     die nächste Revision eintragen, ebenso
  handel stand <vorgang>           wo ein Vorgang steht
  handel protokoll <vorgang>       wie er dorthin kam, Wechsel für Wechsel
  handel anhalt <vorgang>          was die Seiten ungefähr wert sind — Summen
                                   und Differenz je Währung, in Cent
  handel vorgaenge <kennung>       die Handel mit einem Gegenüber
  handel liste [--lage=dran,wartet] [--suche=…] [--partner=…] [--marke=…]
         [--seit=…] [--bis=…] [--ordnung=geaendert|angelegt|lage|partner] [--archiv]
                                   die eigenen Handel — mit Lage und nächstem Schritt
  handel entwurf <vorgang> <art> <kennung>
                                   einen Entwurf neu schreiben, Positionen als JSON
  handel kopieren <vorgang>        aus einem Handel einen neuen Entwurf
  handel verwerfen <vorgang>       einen Entwurf löschen — nur solange er einer ist
  handel archivieren|hervorholen <vorgang>
                                   einen beendeten Handel ins Archiv und zurück
  handel anbieten|gegenangebot|zustimmen|vorbereiten|uebergeben <vorgang> [wer]
         [--gesehen=<abdruck>]     den Weg bis zur Übergabe gehen — mit
                                   --gesehen nur von dieser Revision aus
  handel ablehnen|abbrechen|ablaufen|klaeren <vorgang> [wer]
                                   ihn beenden oder herausnehmen
  handel abgleichen <vorgang>      melden, dass hier fertig ist
  handel verbuchen <vorgang> <freigabe>
                                   im eigenen Bestand verbuchen — die
                                   Freigabe erteilt ein Mensch am Fenster
  handel bestaetigen <vorgang> <wer> <freigabe>
                                   die Stimme einer Partei entgegennehmen
  handelsjournal [vorgang]         die Bewegungen im eigenen Bestand
  handelsbeleg <vorgang>           Stand und Beleg — wer unterschrieben hat
  markt-links <printing> [name]    wohin eine Karte außerhalb führt
  markt-oeffnen <anbieter> <adr>   einen geprüften Weg im Browser öffnen
  markt-konto <quelle> [seite] [menge]
                                   Angebote und Bestellungen, nur gelesen;
                                   Preise in der kleinsten Einheit (EUR: Cent)
  deck-liste <deckID>              eine Deckliste von der Standardeingabe
                                   deuten — schreibt nichts, nennt den Stand
  deck-liste-uebernehmen <deckID> <stand>
                                   dasselbe, und das Eindeutige ins Deck
                                   legen; »-« nimmt den geltenden Stand
  tausch-abgleich <kennung> <datei> <behalten>
                                   Tauschkandidaten gegen einen
                                   freigegebenen fremden Stand
  nachbarn                          wer sich im selben Netz meldet
  schnittstellen                    ob Shellprofil und Codex mse kennen —
                                    eingerichtet wird am Fenster
  schnittstellen widerrufen [ziel]  nimmt zurück, was diese Anwendung dort
                                    eintrug — fremde Einträge bleiben
  schalter                          welche Bereiche und Funktionen an sind —
                                    geschaltet über einstellungen setzen
  netzweg                           wie dieses Programm hinausgeht — Proxy,
                                    Fristen und woher jeder Wert stammt
  regelfreigabe <spiel> <strenge> <sitze>
                                    zeigen, welche Regeln durchgesetzt werden
                                    und ob es so losgehen darf (0 ja, 1 nein)
  zweck-erlauben <kennung> <zweck>  einen Zweck freigeben
  zweck-entziehen <kennung> <zweck>
                                    einen Zweck zurücknehmen
  gesundheit                        sagt, ob dieser Prozess lebt und arbeiten
                                    kann — Rückgabewert 0 bereit, 1 nicht bereit
  meta                              Fassung, Plattform, Go, Fläche, Bau und
                                    Bindungsvertrag — auch ohne Bestand
  vertraeglichkeit                  ob die neueste Freigabe DIESEN Bestand
                                    öffnen kann — 0 ja, 1 nein, 2 ungeklärt
  ruecksprung-arten                 die vier Arten, zurückzugehen, und was
                                    jeder im Weg steht
  ruecksprung-lage <datei>          was ein Zurückspielen dieser Sicherung
                                    kostet — 1, wenn etwas unklar bliebe
  rueckkehren                       jede verbuchte Rückkehr dieses Bestands
  rueckkehr-quittieren <id>         eine Rückkehr für nachgesehen erklären
  pflege-vorschau                   nachsehen, was im Bestand herumliegt —
                                    ändert nichts, 1 wenn etwas dasteht
  pflege <zeitpunkt> <art:revision…>
                                    die gewählten Maßnahmen ausführen —
                                    sichert vorher und sieht danach nach
  kosten <konto> [zeitraum]         was an einem Anbieterkonto freigegeben,
                                    gebunden und ungeklärt ist (1: ungeklärt)
  kostenposten <konto> [zeitraum]   die letzten Aufträge an diesem Konto
  kostenbudget <konto> <zeitraum> <betrag> <waehrung>
                                    die Grenze in der kleinsten Einheit setzen
                                    (5000 EUR = 50,00 €); 0 nimmt sie weg
` + hilfebetrieb
