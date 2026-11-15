package main

// **Die zweite Hälfte der Hilfe** — Betrieb, Wartung und die Werkzeuge ohne
// Fenster.
//
// # Warum sie eine eigene Datei ist
//
// **Weil `hilfe.go` über die Zeilengrenze aus §2 ging.** Der Schnitt folgt
// der Frage: dort stehen die Befehle **am Bestand** — Sets, Karten, Decks,
// Sammlung —, hier die **am Betrieb**: Suchindex, Rücksetzung, Kosten,
// Diagnose.
//
// **Zusammengesetzt bleibt es ein Text.** Wer `mse hilfe` ruft, sieht keine
// Naht; die Datei ist ein Schnitt im Quelltext und keiner in der Ausgabe.

// hilfebetrieb ist die zweite Hälfte.
const hilfebetrieb = `  indexstand                        wogegen der lokale Suchindex sucht —
                                    Modell, Stände und ob er semantisch sucht
  indexsuche <frage> [grenze]       lokal suchen, Volltext und Vektoren —
                                    jeder Treffer nennt seine Herkunft
  indexaufbau <modell> <sammeldatei> <revision> [datum] [--delta]
                                    den Index neu bauen (Minuten; ^C verwirft);
                                    --delta übernimmt, was sich nicht änderte
  recherche <fall> <frage> [grenze]
                                    den Wissensbestand fragen — karten,
                                    aehnlich, regeln oder rulings
  aehnlich <karte> [grenze]         Karten finden, die einer eigenen ähneln
  kontextpaket <aufgabe> <fall> <frage> <anbieter> <modell> <fenster> <reserve> [--belege=n]
                                    zeigen, WAS an ein Modell ginge — gesendet
                                    wird nichts
  ruecksetzungsarten                die sechs Arten, zurückzusetzen — von der
                                    leichtesten zur schwersten
  ruecksetzung-vorschau <art> [profil] [nutzer]
                                    was diese Rücksetzung anfassen würde —
                                    ändert nichts, 1 wenn etwas im Weg steht
  preise-bulk <quelle> <umfang>     die Preise ALLER Drucke holen — mtgjson,
                                    heute oder neunzigTage; lädt viel
  markt <quelle> <druck> [--sprache=xx] [--finish=foil] [--zustand=…]
                                    die Angebote eines Marktplatzes zu einem
                                    Druck lesen — gekauft wird nichts
  fremdverlauf <quelle> <druck> [--finish=foil] [--zustand=…]
                                    den Preisverlauf eines Drucks holen und
                                    ablegen — »scrydex«
  anbieterstand <dienst>            das Kontingent an einem Anbieterkonto —
                                    beim Anbieter bis zu 30 Minuten alt
  ki-modelle                        nennen, was über KI-Modelle feststeht —
                                    Fähigkeiten, Herkunft, Ort
  ki-modell-pruefen <anbieter> <modell>
                                    messen, was ein Modell kann — ein kleiner
                                    Auftrag mit festem Schema; kostet Geld
  ki-modell-vergessen <anbieter> <modell>
                                    eine Messung zurücknehmen
  ki-vorschlag <karte> <anbieter> <modell> <auftrag> [--uebernehmen]
               [--felder=f1,f2] [--nur=f1,f2] [--ohne-geraetemodell]
                                    Felder vorschlagen lassen — ohne --uebernehmen
                                    NICHTS geschrieben; Anbieter »vorrang« wählt
  ki-einschaetzung <karte> <art> <anbieter> <modell> <frage> [format]
                                    eine Karte beurteilen lassen, mit Belegen —
                                    es wird NICHTS geschrieben
  ki-batch <karte,karte,…> <anbieter> <modell> <auftrag>
                                    mehrere Karten auf einmal — Cloudanbieter
                                    nur mit Freigabe am Fenster
  ki-gruppe <kennung=feld,feld> […]
                                    abgelegte Vorschläge als EINE Gruppe
                                    übernehmen — ganz oder gar nicht
  ki-erkennung <bilddatei> <anbieter> <modell>
                                    ein Kartenfoto lesen lassen und passende
                                    Drucke nennen — es legt NICHTS an
  zuruecksetzen <art> <name> [nachweis] [profil] [nutzer] [--revision <marke>]
                                    zurücksetzen — der Name bestätigt; wo
                                    gelöscht wird, ist die Marke Pflicht
  beispielsets                      die mitgelieferten Sets samt Pfad nennen
  oeffnen <pfad>                    ein .mse-set in den Bestand einfahren
  sichern <set> <ziel>              ein Set in eine Datei schreiben
  umschlag-schnueren <set> <ziel>   ein Set als signierten Umschlag ablegen
  umschlag-schauen <pfad>           zeigen, was in einem Umschlag steht —
                                    schreibt NICHTS in den Bestand
  umschlag-uebernehmen <pfad>       ihn hereinnehmen; die Kennungen sind neu
  partien                           die eigenen Partien nennen, jüngste
                                    zuerst; Sitze, Stand, eigener Sitz, dran
  partie-anlegen <spiel> <sitz>... [--sitzprofil=<sitz>:<profil>] [--ohne=<bereich>[,...]]
                                    eine Partie anfangen; Sitz: <kennung>:<name>:<deck>[:<commander>];
                                    --sitzprofil bindet einen Sitz, --ohne nennt selbst gespielte Regeln
  partie-zeigen <partie> <sitz>     zeigen, wie die Partie fuer DIESEN Sitz
                                    aussieht - fremde Haende bleiben zu
  partie-verlauf <partie> <sitz>    zeigen, was in der Partie geschehen ist,
                                    soweit DIESER Sitz es sehen darf
  partie-befehl <partie> <sitz> <kennung> <art> [json]
                                    einen Spielbefehl stellen; die Kennung
                                    schuetzt vor der doppelten Wirkung
  partie-stand <partie> <sitz> <pausieren|fortsetzen|beenden>
                                    anhalten, fortsetzen oder abschliessen
  partie-loeschen <partie>          einen Entwurf herausnehmen — beendete
                                    und eingewilligte Partien bleiben
  partie-kopieren <partie>          sie als neuen Uebungszustand danebenlegen
                                    - ohne Verlauf und ohne alte Zustimmungen
  partie-ausfuhr <partie> <ziel>    sie als Datei ablegen - mit dem VOLLEN
                                    Zustand, also auch fremden Haenden
  partie-einfuhr <pfad>             eine Datei als neue Partie einlesen
  nutzer-wechseln [<id> [nachweis]] zeigen, WER arbeitet - oder umstellen.
                                    Der Nachweis ist der zweite Faktor DIESES
                                    Nutzers, nicht der eigene
  kontext                           zeigen, wer in welchem Profil als was
                                    arbeitet - in einem Zug, mit Revision
  datei-ansehen <pfad>              zeigen, was in einer Datei liegt — nichts
                                    wird ausgepackt oder ausgeführt
  bestand-sichern <ziel> [--klartext]
                                    den GANZEN Bestand sichern — Kennwort aus
                                    MSE_SICHERUNGSKENNWORT (docs/SICHERUNG.md)
  bestand-verschluesselung [an|aus] zeigen oder umschalten, ob der Bestand
                                    verschluesselt auf der Platte liegt
  bestand-sicherungsstand           zeigen, wann der Bestand zuletzt gesichert
                                    wurde — und ob es über 14 Tage sind
  bestand-schauen <sicherung>       zeigen, was in einer Sicherung liegt —
                                    verschluesselt mit MSE_SICHERUNGSKENNWORT
  bestand-zurueckspielen <sicherung> <zeitpunkt>
                                    den Bestand ERSETZEN — der Zeitpunkt aus
                                    »bestand-schauen« bestätigt
  zeichnen <karte> <ziel>           eine Karte als Bild schreiben (.png/.jpg/.svg)
                                    --guete=fast|medium|high|ultra (Vorgabe: medium)
  bogen <set> <ziel.pdf>            die Karten eines Sets als Druckbogen legen
                                    --probe: nur durchrechnen, nichts schreiben
  symbol-nachzeichnen <set> <bild>  ein Bild als Setsymbol nachzeichnen
  vorlagen <set>                    die Exportvorlagen eines Sets zeigen
  vorlagen <set> ausfuehren <paket> <ziel> [feld=wert…]
                                    eine Exportvorlage ausführen
                                    --probe: nur auswerten, nichts schreiben
  vorlagen <set> option <paket> <feld> <wert>
                                    einen Optionswert behalten
  bild-setzen <set> <karte> <feld> <datei>
                                    ein Bild von der Platte in ein Feld setzen
  mtgjson-lesen <pfad> <spiel>      MTG JSON einfahren — Datei oder Verzeichnis
  mtgjson-schreiben <set> <spiel> <ziel.json>
                                    ein Set als MTG JSON schreiben
  mtgjson-spiele                    die Spiele nennen, für die es eine
                                    Zuordnung gibt
  skript <karte> <ausdruck>         einen MSE-Script-Ausdruck auswerten
  skriptnamen <karte>               die Namen der Umgebung aufzählen
  symbol <set> [name]               das abgelegte Symbol eines Sets als JSON
  symbolpalette <karte> <feld>      was ein Feld zum Einfügen anbietet
  symbol-setzen <set> [name]        eine Zeichnung von der Standardeingabe ablegen
  symbol-rechnen <name> [--feld wert]...
                                    eine Rechnung des Symboleditors; die Teile
                                    kommen von der Standardeingabe
  symbolbibliothek [liste [suche]]  die Symbole der Bibliothek samt Trägern
  symbolbibliothek aufnehmen <set> <name>
                                    das Symbol eines Sets aufnehmen
  symbolbibliothek duplizieren <id> <name>
                                    ein Symbol unter anderem Namen anlegen
  symbolbibliothek umbenennen <id> <name>
                                    einem Symbol einen anderen Namen geben
  symbolbibliothek loeschen <id> [--bestaetigt]
                                    löschen — ein getragenes nur bestätigt
  symbolbibliothek einsetzen <id> <set>
                                    eine Kopie zum Symbol eines Sets machen
  symbolbibliothek einlesen <datei> [name]
                                    eine .mse-symbol als neues Symbol
  symbolbibliothek ausgeben <id> [datei]
                                    als .mse-symbol — ohne Datei als Text
  symbolbibliothek teile <id>       die Zeichnung als Teile (JSON)
  symbolbibliothek sichern <id>     eine neue Fassung aus den Teilen von der
                                    Standardeingabe; die bisherige wird die vorige
  symbolbibliothek zuruecknehmen <id>
                                    die vorige Fassung zurückholen
  symbolbibliothek nachziehen <id>  die Sets mit der vorigen Fassung nachziehen
  sprachpaket [liste]               die eigenen Übersetzungen samt Stand
  sprachpaket zeige <kennung>       eine samt Sätzen und fehlenden Pfaden (JSON)
  sprachpaket anlegen <kennung> <name> <basis>
                                    eine über einer Fenstersprache anlegen
  sprachpaket setze <kennung> [--name=…] [--stand=…] [--lizenz=…] [--herkunft=…] [pfad=text]…
                                    Angaben und Sätze setzen; pfad= nimmt zurück
  sprachpaket loeschen <kennung>    eine entfernen
  sprachpaket ausgeben <kennung> [datei]
                                    als Datei — ohne Datei als Text
  sprachpaket einlesen <datei> [--ersetzen]
                                    eine aus einer Datei aufnehmen; benutzt wird
                                    sie mit »einstellungen setzen sprache.eigene«
  paketdateien <paket>              die bearbeitbaren Dateien eines Pakets (--json: mit Beiwerk)
  paketdatei <paket> <datei>        eine Paketdatei ausgeben (--json: mit Stand)
  paketdatei-pruefen <paket> <datei>
                                    ihre Skripte auf Syntaxfehler prüfen
  paketdatei-setzen <paket> <datei> <quelle> <stand>
                                    eine Paketdatei aus einer Datei setzen
  paketdatei-anlegen <paket> <datei> [quelle]
                                    eine neue Vorlage anlegen, leer oder aus einer Datei
  paketdatei-einfuegen <paket> <quelle> [ziel]
                                    ein Bild oder eine Schrift ins Paket legen
  paketdatei-umbenennen <paket> <datei> <neu> [stand]
                                    umbenennen — bei einer Vorlage mit Stand
  paketdatei-entfernen <paket> <datei> [stand]
                                    in die Sicherung neben dem Paket legen
  sammlung [druck]                  die eigene Sammlung zeigen
  sammlung [--suche <text>] [--nonfoil|--foil|--etched]
           [--nach <ordnung>] [--ab <n>] [--anzahl <n>]
                                    einen Ausschnitt davon
  masse-aendern <feld=wert…> <id:fassung…>
                                    viele Exemplare ändern — menge, zustand,
                                    lagerort, wunsch=ja|nein
  masse-loeschen <id:fassung…>      viele Exemplare auf einmal löschen
  preise-alle [stunden]             Preise für die Sammlung holen, die älter
                                    als <stunden> sind (Vorgabe: 24)
  kartenskripte <set>               die Kartenskripte des Spielpakets zeigen
  kartenskript <set> <name>         eines ausführen — legt mehrere Karten in
                                    EINEM Zug an
  auszaehlung <set> [dimension]     wonach dieses Spiel auszählen lässt —
                                    mit Dimension die Auszählung selbst
  autoersetzung <set>               die Autoersetzungsregeln dieses Spiels
  autoersetzung <set> setzen <muster> <ersatz> [teilwort] [aus]
                                    eine Regel setzen — teilwort: auch im
                                    Wort, aus: abgeschaltet
  autoersetzung <set> loeschen <muster>
                                    eine eigene Regel löschen
  autoersetzung <set> zuruecksetzen  die eigenen Regeln vergessen
  autoersetzung <set> an|aus        die Autoersetzung ganz schalten
  decks [spiel]                     die Decks auflisten
  deck <id>                         ein Deck samt seinen Plätzen zeigen
  deck neu <name> <spiel> [notiz]   ein leeres Deck anlegen
  deck umbenennen <id> <fassung> <name> [notiz]
                                    Name und Notiz eines Decks ändern
  deck kopieren <id> <name>         ein Deck unter neuem Namen kopieren
  deck aus-set <set> <menge> [name]
                                    aus den Karten eines Sets ein Deck anlegen
  deck wuenschen <id>               was fehlt, auf die Wunschliste setzen
  deck loeschen <id>                ein Deck samt seinen Plätzen löschen
  deck legen <id> <druck|--karte=<id>> <menge> [bereich]
                                    eine Karte ins Deck legen
                                    bereich: hauptdeck (Vorgabe) | sideboard
  deck nehmen <id> <platz>          einen Platz herausnehmen
  deck verschieben <id> <platz> <bereich>
                                    einen Platz in den anderen Bereich legen;
                                    gleiche Bezüge gehen zusammen
  deck format <id> [format]         auf ein Format stellen — ohne Angabe herunter
  deck pruefen <id>                 gegen das Format prüfen (1 = unzulässig)
  deck bestand <id>                 was davon in der Sammlung liegt (1 = es fehlt)
  deck zuordnungen <id>             Bedarf, freie Stücke und Zuordnungen (1 = es fehlt)
  deck zuordnen <id> <platz> <exemplar> <menge>   Exemplare für dieses Deck festlegen
  deck zuordnung-loesen <id> <zuordnung>          eine Zuordnung wieder lösen
  deck fehlkarten <id>              die fehlenden Karten als Liste (1 = es fehlt)
  deck liste <id>                   das Deck als Liste, wie deck-liste sie liest
  exemplar decks <exemplar>         welche Decks dieses Exemplar benutzen
  formate [spiel]                   die Formate auflisten
  format bauarten                   die drei nachgelesenen Bauarten zeigen
  format neu <bauart> <name> <spiel> [scryfallname]
                                    ein Format aus einer Bauart anlegen
  format setzen <id> <fassung> <name> <spiel>
                <mindest> <hoechst> <sideboard> <gleiche> [scryfallname]
                                    ein Format fortschreiben — 0: keine Grenze
                                    (Sideboard: keines), offen: unbegrenzt
  format loeschen <id>              ein Format löschen; die Decks bleiben
  turniere [spiel]                  die Turniere auflisten
  turnier <id>                      ein Turnier samt Verlauf zeigen
  turnier neu <name> <spiel> [notiz] [einzelspiel] [--format <id>]
                                    ein Turnier anlegen — das Format muss
                                    zum selben Spiel gehören
  turnier kopieren <id> <name> [--mit-teilnehmenden]
                                    ein Turnier nach einer Vorlage anlegen —
                                    Namen ja, Decks und Runden nein
  turnier loeschen <id>             ein Turnier samt Verlauf löschen
  turnier anmelden <id> <name> [deck]
                                    einen Teilnehmer eintragen — mit Deck
                                    wird seine Liste ABGEZOGEN
  turnier aussteigen <id> <teilnehmer>
                                    abmelden; er bleibt in der Rangliste
  turnier einsteigen <id> <teilnehmer>
                                    wieder anmelden
  turnier vorschlag <id>            zeigen, wie die nächste Runde ausfiele
  turnier runde <id>                die nächste Runde paaren und legen
  turnier ergebnis <id> <paarung> <a> [b] [unentschieden]
                                    einen Spielstand eintragen
                                    »weg« statt <a> nimmt ihn zurück
  turnier folgen <id> <paarung> <a> [b] [unentschieden]
                                    was eine Berichtigung verschöbe, ohne sie
                                    einzutragen — »weg« statt <a> für zurück
  turnier abschliessen <id>         abschließen — danach nimmt es nichts an
  turnier wiedereroeffnen <id>      den Abschluss zurücknehmen
  turnier protokoll <id>            was an einem Turnier geschah
  turnier rangliste <id>            den Stand nach den Turnierregeln
  turnier ausgeben <id> [datei]     als Datei ohne Kennungen dieses Bestands
  turnier einlesen <datei>          eine Ausgabe als neues Turnier samt Verlauf
  turnier bericht <id> [datei]      der Bericht als Markdown: Stand, Rangliste, Runden
  besitz <druck>                    was von einem Druck da ist — und was
                                    davon nur gewünscht
  lagerorte                         die Lagerorte samt dem, was dort liegt
  dubletten [behalten]              wovon mehr da ist als gewollt (Vorgabe: 1)
  wuensche                          die Wunschliste
  sammlung-ausgeben [liste|csv]     die Sammlung ausgeben (Vorgabe: csv)
                                    »liste« ist einfügbar und verlustbehaftet
  sammlung-einlesen [bezeichnung] [--auslassen]
                                    ein CSV von der Standardeingabe zurücklesen;
                                    --auslassen lässt vorhandene Zeilen stehen
  sammlung-vorschau                 dasselbe vorab zeigen — schreibt NICHTS
  liste-deuten                      eine Kartenliste von der Standardeingabe
                                    deuten — schreibt NICHTS
  liste-einlesen [zustand] [--auslassen]
                                    dasselbe, und das Eindeutige aufnehmen
  liste-vorschau [zustand]          zeigen, was liste-einlesen täte
  exemplar <druck> <menge> [zustand] [sprache] [nonfoil|foil|etched]
           [erwerbsdatum] [erwerbspreis] [waehrung] [notiz]
           [name] [setcode] [nummer] [lagerort] [wunsch|besitz]
           [niedrig|normal|hoch|sehrhoch] [genau|jederdruck]
                                    ein Exemplar aufnehmen; die letzten zwei
                                    gelten nur an einem Wunsch
  exemplar-aendern <id> <menge>     die Menge ändern; alles andere bleibt
  exemplar-loeschen <id>            ein Exemplar herausnehmen
  preise <druck>                    die abgelegten Preise eines Drucks
  preise-holen <druck>              Preise bei Scryfall holen und ablegen
  preisverlauf <druck> <waehrung> <finish>
                                    eine Preisreihe in der Zeit
  portfolio                         die Sammlung zusammenrechnen
  verweis <karte>                   auf welchen Druck eine Karte verweist
  verweis <karte> <druck>           die Karte auf einen Druck beziehen
  verweis <karte> --loesen          den Bezug lösen
  verweise <set>                    die Verweise aller Karten eines Sets
  marke <set> <karte>               die Marken einer Karte
  marke <set> <karte> <wort>        eine Marke setzen
  marke <set> <karte> <wort> --weg  eine Marke abnehmen
  marken <set>                      die Marken, die im Set vorkommen
  objektmarke <art> <kennung> [<wort> [--weg]]
                                    Marken eines Objekts lesen, setzen, abnehmen
  objektmarken <art>                die Marken jedes Objekts einer Art
  marken-filtern <art> <wort…> [--oder]
                                    die Objekte mit diesen Marken
  marke-vergeben <art> <wort> <kennung…>
                                    eine Marke an mehrere Objekte hängen
  markenkatalog                     alle Marken des Profils, gezählt
  marke-umbenennen <alt> <neu> [--zusammenfuehren]
                                    eine Marke umbenennen oder zusammenführen
  marke-tilgen <wort>               eine Marke von jedem Objekt nehmen
  ablage-kopieren <set> <karte…>    Karten als Text ausgeben
  ablage-ausschneiden <set> <karte…>
                                    dasselbe, und aus dem Set entfernen
  ablage-einfuegen <set>            Karten aus der Standardeingabe anlegen
  kartenliste <set>                 die Spalten und Zeilen der Kartenliste
  kartenfilter <set> <anfrage>
           [--spalten=<a,b>] [--sortieren=<feld>] [--absteigend]
                                    die Karten, die der Schnellfilter durchlässt;
                                    --spalten gilt für ALLE Sets des Spiels
  ersetzen <set> <text> <ersatz>
           [--tun] [--feld=<name>] [--gross-klein] [--ganzes-wort]
                                    zeigen, was sich ändern würde — --tun
                                    ersetzt wirklich
  schluesselwoerter <set>           die Schlüsselwörter eines Sets zeigen
  schluesselwort <set> anlegen      ein leeres Schlüsselwort anlegen
  schluesselwort <set> setzen <id> <feld> <wert>
                                    ein Feld eines Schlüsselworts setzen
  schluesselwort <set> loeschen <id>
                                    ein Schlüsselwort löschen
  schluesselwort <set> parameter <id>
                                    zeigen, was sich in Muster und
                                    Erinnerungstext einfügen lässt
  packarten <set>                   die Packarten eines Sets zeigen
  pack <set> <art> <same>           einen Booster ziehen — derselbe Same,
                                    dieselbe Ziehung
  wortlisten <karte> <feld>         die Auswahllisten eines Feldes zeigen
  wortwahl <set> <karte> <feld> <bereich> <weg>
                                    einen Eintrag daraus wählen
  scryfall <set> <karte> <name>     zeigen, was Scryfall anders hat
  scryfall-karte <set> <name>       eine Karte aus einem Scryfall-Namen anlegen
  scryfall-set <code> <spiel>       einen ganzen Satz von Scryfall einlesen
           [--uebernehmen f1,f2]    und die genannten Felder übernehmen
  scryfall-suche <frage> [grenze]   bei Scryfall suchen — schreibt NICHTS
  scryfall-regeln <name>            den Regeltext einer Karte holen
  scryfall-mana <kosten>            Manakosten deuten, etwa »{2}{R}«
  scryfall-bild <set> <karte> <feld> <name>
                                    ein Kartenbild von Scryfall in ein Feld
                                    setzen
  scryfall-sammeldaten <zeigen|holen>
                                    den Stand der Sammeldaten zeigen oder sie
                                    holen — »holen« lädt Hunderte Megabyte
  aktualisierung [--laden]          nach einer neueren Fassung sehen und sie
                                    ins Downloadverzeichnis holen
  platz                             was wie viel Platz belegt
  platz freigeben <posten>          einen räumbaren Posten wegräumen —
                                    die jüngste Sicherung bleibt stehen
  hilfe                             diese Übersicht

Beispiele:
  mse einstellungen setzen darstellung de-CH
                                    Zahlen und Daten wie in der Schweiz
  mse einstellungen --json          alle Einstellungen, für ein Skript
  mse deck neu Mono-Rot magic       ein leeres Deck anlegen
  mse decks --json                  und es wiederfinden
  mse sammlung --json               die eigene Sammlung, für ein Skript

Ausgabe:
  --json   maschinenlesbar statt für Menschen
  Ergebnisse gehen nach stdout, Meldungen und Absagen nach stderr.

Rückgabewerte:
  0        getan
  1        die Antwort ist nein — fachlich
  2        der Aufruf stimmt nicht
  130      abgebrochen (⌃C oder SIGTERM)
`
