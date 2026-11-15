// Die Meldungen, die aus der Fassade kommen.
//
// Die Schlüssel entsprechen denen, die Go setzt (AGENTS.md §9);
// `internal/app/fehlerschluessel_test.go` hält beide Seiten zusammen und wird
// rot, sobald hier einer fehlt oder einer zu viel steht.
//
// Der deutsche Satz aus Go reist als Rückfall mit: ein Schlüssel, den diese
// Datei nicht kennt, kostet eine unübersetzte Meldung, nicht eine leere.
export const fehler = {
  unbekannt:
    'Es ist ein unerwarteter Fehler aufgetreten. Was genau, steht im Protokoll — unter „Einstellungen › Diagnose“ siehst du sein Ende und kannst es sichern.',
  unbekanntMitVorgang:
    'Es ist ein unerwarteter Fehler aufgetreten. Was genau, steht im Protokoll — unter „Einstellungen › Diagnose“ siehst du sein Ende und kannst es sichern. Suche dort nach {vorgang}: unter dieser Kennung stehen alle Zeilen dieses Vorgangs.',
  nichtGefunden:
    'Der gesuchte Eintrag wurde nicht gefunden. Vielleicht hat ihn jemand anderes gelöscht — lade die Liste neu.',
  ohnePeeridentitaet:
    'Dieses Profil hat noch keine Identität. Lege sie im Kasten „Womit sich dieses Profil ausweist“ an — ohne sie kann sich niemand ausweisen.',
  peerNichtErreichbar:
    'Das Gegenüber ist gerade nicht zu erreichen. Horcht es, und stimmt seine Adresse? Sonst versuche es später noch einmal.',
  peerOhneAdresse:
    'Für dieses Gegenüber ist keine Adresse eingetragen. Trage sie an seinem Kontakt unter „Adresse und Notiz“ ein — ohne sie findet dieses Programm es nur im selben Netz.',
  ohneNetzerlaubnis:
    'Ins lokale Netz ging nichts hinaus. Hat macOS eben nach der Erlaubnis gefragt: erlauben und noch einmal versuchen. Sonst lässt sich der Grund von hier nicht sicher sagen: entweder fehlt die Erlaubnis — Systemeinstellungen → Datenschutz & Sicherheit → Lokales Netzwerk —, oder das Netz lässt Multicast nicht durch, etwa hinter einem VPN.',
  ohneNetzroute:
    'Ins lokale Netz gibt es gerade keinen Weg. Ist eine Netzwerkschnittstelle verbunden? Unter Einstellungen → Verbindungen lässt sich eine bestimmte wählen.',
  schluesselbundAblage:
    'Der Schlüssel dieses Profils ließ sich nicht im Schlüsselbund ablegen. Prüfe, ob der Anmeldungsschlüsselbund entsperrt ist, und versuche es dann noch einmal.',
  einladungZuNeu:
    'Diese Einladung stammt aus einer neueren Fassung dieses Programms. Bitte die Gegenseite um eine Einladung aus ihrer älteren Fassung — oder hole dir die neuere.',
  fernhorchenLaeuft:
    'Dieses Profil ist über fremde Netze schon erreichbar. Beende es zuerst, wenn du eine neue Einladung willst — die alte wird damit ungültig.',
  anmeldungAbgelaufen:
    'Die Anmeldung kam nicht binnen zehn Minuten zurück. Starte sie neu, wenn du im Browser so weit bist.',
  umschlagFassung:
    'Dieser Umschlag stammt aus einer anderen Fassung dieses Programms. Bitte die Gegenseite um einen aus ihrer älteren Fassung — oder hole dir die neuere.',
  umschlagUnterschrift:
    'Die Unterschrift dieses Umschlags trägt nicht. Er ist unterwegs verändert worden oder kommt von jemand anderem, als er sagt — nimm ihn nicht an und frage bei der Gegenseite nach.',
  umschlagInhalt:
    'Ein Stück dieses Umschlags passt nicht zu seinem Manifest. Bitte die Gegenseite, ihn noch einmal zu schicken — unterwegs ist etwas kaputtgegangen oder geändert worden.',
  fremderNutzer:
    'Das darf nur dieser Nutzer selbst. Wechsle zuerst zu diesem Nutzer — dabei fragt die Anwendung nach dem zweiten Faktor.',
  nichtFreigegeben:
    'Das Gegenüber gibt dir zum Handeln nichts frei. Bitte es, dir eine Sicht freizugeben — oder handelt über eine Datei.',
  zeileNichtVerbuchbar:
    'Die Zeile {zeile} ließe sich drüben nicht verbuchen: Zustand oder Sprache fehlen, oder eine Angabe gibt es nicht. Entferne sie und lege sie mit Zustand, Sprache und Ausführung neu an.',
  zusageUnpassend:
    'Dieses Exemplar erfüllt keine Zeile, die du gibst — Zustand, Sprache oder Ausführung weichen ab. Wähle ein passendes oder ergänze die Sammlung.',
  erlaubnis:
    'Dazu fehlt dir die Rolle. Wer dieses Profil verwaltet, kann sie dir geben.',
  unbekannteHerkunftsart:
    'Diese Herkunft kennt das Programm nicht. Wähle »aus einem Profil übernommen« oder »eingelesen«.',
  // **Die Kommandozentrale** (#1423).
  befehlBestaetigen:
    'Dieser Befehl ändert etwas, das sich nicht ohne Weiteres zurücknehmen lässt — lies die Vorschau und bestätige dann.',
  befehlArgumente:
    'Die Argumente sind kein JSON-Objekt. Gib ein Objekt in geschweiften Klammern ein — ein Feld je Parameter.',
  herkunftOhneProfil:
    'Eine Herkunft aus einem Profil nennt es, eine Einfuhr nennt keines — gib eines an oder lass es weg.',
  // **Der Satz sagt, was zu tun ist — und was nicht.** Wer über einer
  // Datei, die nicht aufgeht, „unerwarteter Fehler“ liest, ersetzt sie;
  // hier ist sie in Ordnung, und ein Ersetzen machte den Verlust endgültig.
  // **Der Satz sagt, was zu tun ist** — nicht „geht nicht“. Wer als Leser in
  // einem fremden Profil steht, hat zwei Wege: hinaus, oder um eine Rolle
  // bitten.
  nurLesend:
    'In diesem Profil darfst du nur lesen. Es gehört jemand anderem, und deine Rolle darin ist »Leser«. Wechsle in dein eigenes Profil, um zu arbeiten — oder bitte die Person, die es verwaltet, um die Rolle »Mitarbeiter«.',
  // **Was beim Start schiefging** — Werkstück 3 von Phase 4.13
  // (#1352). Zwei Lagen, zwei Ratschläge: die eine Datei ist unbrauchbar,
  // die andere ist in Ordnung und nur zu neu.
  bestandUnbrauchbar:
    'Diese Datei ist beschädigt oder kein Bestand von MagicSetEditor4Mac — sie lässt sich nicht öffnen. Spiele eine Sicherung zurück; die letzte liegt dort, wohin du sie gesichert hast.',
  bestandNeuereFassung:
    'Dieser Bestand stammt aus einer neueren Fassung von MagicSetEditor4Mac. Die Datei ist in Ordnung — ersetze sie nicht. Hole dir die neuere Fassung, oder spiele eine Sicherung von vorher zurück.',
  bestandAeltereFassung:
    'Dieser Bestand stammt aus einer Fassung vor 0.7.0, und diese Fassung zieht ihn nicht mehr nach. Die Datei ist in Ordnung — ersetze sie nicht. Öffne ihn einmal mit MagicSetEditor4Mac 0.7.0; danach öffnet ihn auch diese Fassung.',
  // **Drei weitere Lagen des Starts und die verworfene Sicherung** —
  // F05 (#2344). Keine der drei rät zum Ersetzen.
  bestandNurLesbar:
    'Der Bestand lässt sich lesen, aber nicht schreiben — der Datei oder ihrem Ordner fehlt das Schreibrecht. Die Datei ist in Ordnung — ersetze sie nicht. Gib ihr im Finder unter »Informationen« Schreibrecht und versuche es noch einmal.',
  bestandEA:
    'Beim Lesen des Bestands meldete der Datenträger einen Fehler. Ob die Datei heil ist, ist offen — ersetze sie nicht gleich. Prüfe den Datenträger, etwa mit dem Festplattendienstprogramm, und versuche es noch einmal.',
  kaumPlatz:
    'Der Bestand ließ sich nicht öffnen oder schreiben, und auf dem Datenträger ist kaum Platz frei. Die Datei ist vermutlich in Ordnung — ersetze sie nicht. Schaffe Platz und versuche es noch einmal; hilft das nicht, prüfe den Datenträger, etwa mit dem Festplattendienstprogramm, und das Schreibrecht.',
  startUngeklaert:
    'Der Bestand ließ sich nicht öffnen, und der Grund ist nicht geklärt — er steht darunter. Versuche es noch einmal. Spiele eine Sicherung erst zurück, wenn du sicher bist, dass diese Datei nicht mehr zu öffnen ist.',
  sicherungVerworfen:
    'Die Sicherung bestand ihre Prüfung nicht und wurde verworfen — es ist nichts gesichert. Versuche es noch einmal, oder sichere an einen anderen Ort.',
  sicherungUnvermerkt:
    'Die Sicherung liegt geprüft da, nur ihr Zeitpunkt ließ sich nicht festhalten. Behalte die Datei und versuche es noch einmal, wenn der Bestand frei ist — bis dahin zeigt das Alter den vorigen Stand.',
  schluesselFehlt:
    'Dieser Bestand ist verschlüsselt, und sein Schlüssel liegt nicht im Schlüsselbund dieses Rechners. Die Datei ist in Ordnung — ersetze sie nicht. Der Schlüssel entstand beim Einschalten der Verschlüsselung und lässt sich nicht neu berechnen; ohne ihn kommt niemand mehr an die Daten.',
  belegt:
    'Der Bestand ist gerade belegt — vermutlich schreibt ein anderes Fenster oder ein Befehl. Bitte gleich noch einmal versuchen.',
  bestandWirdGetauscht:
    'Der Bestand wird gerade ausgetauscht — durch das Zurückspielen einer Sicherung, das Umschalten der Verschlüsselung oder einen Neuanfang. Warte einen Augenblick und versuche es dann noch einmal.',
  anwendungEndet:
    'MagicSetEditor4Mac wird gerade beendet — dieser Schritt wurde nicht mehr ausgeführt. Starte die Anwendung neu, um weiterzuarbeiten.',
  vorgangAbgebrochen:
    'Der Vorgang wurde abgebrochen, bevor er fertig war. Starte ihn noch einmal, wenn du ihn noch brauchst.',
  konflikt: 'Der Eintrag wurde zwischenzeitlich geändert. Bitte neu laden.',
  // **Nur ein überholter Stand heißt »lade neu«** — ENG-05.AK-03 (#1730),
  // nativer Befund F08. Jeder andere Konflikt nennt seinen Grund, und eine
  // verletzte Eindeutigkeit sagt, was doppelt ist.
  konfliktgrund:
    'Das ging nicht: {grund}. Prüfe den Stand, bevor du es noch einmal versuchst.',
  doppelt:
    'Einen Eintrag mit diesen Angaben gibt es schon. Ändere eine Angabe — oder bearbeite den vorhandenen.',
  doppelterName: 'Diesen Namen gibt es hier schon — wähle einen anderen.',
  doppeltesExemplar:
    'Diesen Druck führt die Sammlung schon — im selben Zustand, in derselben Sprache und Ausführung. Erhöhe dort die Menge oder wähle eine andere Angabe.',
  // **Der volle Datenträger** (#1812): der Rat liegt außerhalb der Anwendung.
  gesperrt:
    'Auf diesen Ort darf dieses Programm nicht zugreifen — er ist gesperrt. Wähle einen anderen oder gib ihn frei, in den Rechten der Datei oder unter Systemeinstellungen › Datenschutz & Sicherheit › Dateien und Ordner, und versuche es noch einmal.',
  voll: 'Auf dem Datenträger ist kein Platz mehr. Schaffe Platz und versuche es noch einmal.',
  // **Ein fremder Dienst, der nicht antwortet** (D02, #2233): je Fall
  // der nächste Schritt; `{wirt}` ohne Pfad und Abfrage.
  netzName:
    'Der Name »{wirt}« ließ sich nicht auflösen. Prüfe die Internetverbindung und, falls einer eingerichtet ist, den Proxy unter „Einstellungen › Verbindungen“. Was die Verbindung meldete, steht unter „Einstellungen › Diagnose“ bei {vorgang}.',
  netzNichtErreichbar:
    'Unter »{wirt}« nimmt niemand Verbindungen an. Läuft der Dienst, und stimmt seine Adresse? Sonst versuche es später noch einmal. Was die Verbindung meldete, steht unter „Einstellungen › Diagnose“ bei {vorgang}.',
  netzKeineAntwort:
    '»{wirt}« hat nicht binnen {sekunden} Sekunden geantwortet. Versuche es später noch einmal. Was die Verbindung meldete, steht unter „Einstellungen › Diagnose“ bei {vorgang}.',
  netzAntwortStockt:
    'Die Antwort von »{wirt}« stand {sekunden} Sekunden lang still. Prüfe die Verbindung und versuche es noch einmal. Was die Verbindung meldete, steht unter „Einstellungen › Diagnose“ bei {vorgang}.',
  netzAbgerissen:
    'Die Verbindung zu »{wirt}« brach ab, bevor die Antwort ganz da war. Versuche es noch einmal; bleibt es dabei, liegt es am Dienst oder an einem Proxy dazwischen. Was die Verbindung meldete, steht unter „Einstellungen › Diagnose“ bei {vorgang}.',
  netzProxyAnmeldung:
    'Der Proxy »{wirt}« verlangt eine Anmeldung. Gib seine Zugangsdaten über HTTPS_PROXY in der Umgebung an — unter „Einstellungen › Verbindungen“ nimmt der Proxy bewusst kein Kennwort. Was die Verbindung meldete, steht unter „Einstellungen › Diagnose“ bei {vorgang}.',
  gedrosselt:
    'Die Gegenstelle meldet zu viele Anfragen. Warte einen Moment und versuche es dann noch einmal. Was die Verbindung meldete, steht unter „Einstellungen › Diagnose“ bei {vorgang}.',
  gedrosseltBis:
    '»{wirt}« nimmt gerade keine Anfrage an. Versuche es ab {uhrzeit} noch einmal — bis dahin fragt dieses Programm nicht erneut. Was die Verbindung meldete, steht unter „Einstellungen › Diagnose“ bei {vorgang}.',
  kontingent:
    'Zu viele Rufe in kurzer Zeit. Warte {sekunden} s und versuche es dann noch einmal.',
  prozessVeraltet:
    'Dieser Prozess stammt aus einer älteren Fassung als der Bestand und endet jetzt — geschrieben wurde nichts. Ein eingerichteter Dienst startet von selbst neu, mit dem Programm, auf das er eingerichtet ist; jeden anderen Prozess starte mit der neuen Fassung neu und versuche es dann noch einmal.',
  bestandVomDienstGehalten:
    'Der Dienst für Skripte hält den Bestand offen. Widerrufe ihn unter Einstellungen › Diagnose › »Kommandozeile und Werkzeugserver einrichten« und spiele dann noch einmal zurück.',
  // **Kein Konflikt im Sinne von »lade neu«** — sondern eine Entscheidung
  // (#1369).
  // **Die Befunde eines Ordners** — OPS-02 (#1719).
  ordnerangabe:
    '»{wert}« ist kein vollständiger Ort. Trage einen ein, der mit / oder mit ~/ beginnt, ohne .. dazwischen.',
  ordnerFehlt:
    'Der Ordner »{ordner}« ist nicht erreichbar. Schließe das Laufwerk an oder wähle einen anderen Ordner.',
  ordnerIstDatei:
    '»{ordner}« ist eine Datei und kein Ordner. Wähle den Ordner, in dem sie liegt, oder einen anderen.',
  ordnerGesperrt:
    'In »{ordner}« darf diese Anwendung nicht schreiben. Wähle einen anderen Ordner oder gib diesen frei.',
  ordnerVoll:
    'Auf dem Laufwerk von »{ordner}« sind nur noch {frei} frei. Schaffe Platz oder wähle einen anderen Ordner.',
  karteInDeck:
    'Diese Karte liegt in {anzahl} Deck(s) — {decks}. Das Löschen nimmt sie auch dort heraus; bestätige, wenn es so gemeint ist.',
  falschesSpiel:
    'Das Deck gehört zu einem anderen Spiel als dieses Turnier. Wähle ein Deck aus demselben Spiel.',
  turnierformatSpiel:
    'Das Format gehört zu einem anderen Spiel als dieses Turnier. Wähle ein Format aus demselben Spiel — oder keines.',
  paketGeaendert:
    'Dieses Paket weicht vom Stand beim Einbinden ab. Bestätige das Ersetzen — die Änderung bleibt als vorige Fassung erhalten.',
  // **Die Symbolbibliothek** (DOM-02, #1697).
  symbolZuVieleTeile:
    'Dieses Symbol hat {anzahl} Teile, erlaubt sind {grenze}. Fasse Teile zusammen oder wähle eine schlichtere Zeichnung.',
  symbolZuVielePunkte:
    'Dieses Symbol hat {anzahl} Punkte, erlaubt sind {grenze}. Vereinfache die Formen oder wähle eine schlichtere Zeichnung.',
  symbolZuTief:
    'Die Gruppen dieses Symbols stecken {anzahl} Ebenen tief, erlaubt sind {grenze}. Wähle eine flachere Zeichnung oder löse innere Gruppen auf.',
  symbolZuGross:
    'Diese Datei ist größer als {grenze} und damit keine Symboldatei. Wähle eine .mse-symbol-Datei.',
  symbolGespeichertZuGross:
    'Gespeichert wäre dieses Symbol größer als {grenze} — so viel fasst keine Symboldatei. Vereinfache die Formen oder nimm weniger Punkte.',
  bildZuGross:
    'Dieses Bild ist zu groß, um es auszupacken. Wähle ein kleineres oder verkleinere es vorher.',
  setOhneSymbol:
    'Dieses Set hat noch kein Symbol. Zeichne eines und lege es ins Set — dann lässt es sich aufnehmen.',
  // **Die eigenen Übersetzungen** — LOC-02 (#1688).
  sprachkennungUngueltig:
    '»{kennung}« ist keine Sprachkennung nach BCP 47. Nimm eine wie pt-BR oder zh-Hant.',
  sprachkennungEingebaut:
    '»{kennung}« ist eine eingebaute Sprache. Wähle eine eigene Kennung wie {kennung}-x-meine.',
  platzhalterAbweichend:
    'Diese Übersetzung nennt andere Platzhalter als der deutsche Satz. Trage genau diese ein: {erwartet}.',
  auszeichnungUnsicher:
    'Diese Übersetzung trägt, was kein Text ist, oder eine Auszeichnung, die nicht schließt. Entferne Tags und schließe ** und `.',
  sprachpaketNeuer:
    'Diese Übersetzung stammt aus einem neueren Programm ({format}). Hole dir die neue Fassung.',
  sprachpaketOhneLizenz:
    'Ohne Lizenz geht keine Übersetzung hinaus. Trage eine ein, etwa CC0-1.0.',
  bereichAbgeschaltet:
    'Dieser Bereich ist abgeschaltet ({schalter}). Wähle unter Einstellungen › Bereiche und Funktionen wieder „an“ — verloren ist nichts.',
  dienstAus:
    'Der Dienst {dienst} ist abgeschaltet. Wähle unter Einstellungen › Dienste „an“, wenn du ihn fragen willst.',
  fremdeEigeneKarten:
    'Eigene Karten des Absenders kamen nicht mit — ihre Kennung gilt nur in seinem Bestand: {karten}. Wenn du sie brauchst, lade ihr Set und lege sie hierher.',
  nichtGelegtePlaetze:
    'Diese Plätze ließen sich hier nicht legen: {plaetze}. Das übrige Deck ist angenommen; prüfe die Plätze und lege sie von Hand nach.',
  fremdeKartenUndNichtGelegte:
    'Eigene Karten des Absenders kamen nicht mit — ihre Kennung gilt nur in seinem Bestand: {karten}. Und diese Plätze ließen sich hier nicht legen: {plaetze}. Prüfe beides und lege es von Hand nach.',
  symbolVerwendet:
    'Diese Sets tragen das Symbol: {sets}. Bestätige das Löschen — sie behalten ihre Kopie.',
  symbolOhneVorige:
    'Dieses Symbol hat keine vorige Fassung. Sichere zuerst eine neue — dann lässt sich die jetzige zurückholen.',
  turnierAbgeschlossen:
    'Dieses Turnier ist abgeschlossen. Wähle »Wiedereröffnen«, wenn du etwas ändern willst.',
  sitzGegangen:
    'Dieser Sitz hat die Partie verlassen und kann nichts mehr tun. Wähle einen Sitz, der noch im Spiel ist.',
  fremderWert:
    'Dieses Feld hat inzwischen jemand anderes geändert. Prüfe den neuen Wert und entscheide, ob du deinen Entwurf behältst.',
  vorschlagNurGegenseite:
    'Ein offenes Angebot ändert nur die Gegenseite. Warte auf ihre Antwort, oder brich den Handel ab und biete neu an.',
  turnierNichtFertig:
    'Trage erst jedes Ergebnis ein: ein Turnier schließt ab, wenn es mindestens eine Runde hat und jede Paarung ein Ergebnis.',
  turniernameLeer: 'Ein Turnier braucht einen Namen — trage einen ein.',
  turnierspielLeer:
    'Ein Turnier gehört zu einem Spiel — sonst gibt es kein Format, gegen das sich prüfen ließe. Wähle eines aus.',
  validierungsgrund:
    'Das ging nicht: {grund}. Berichtige die Angabe und versuche es noch einmal.',

  validierung:
    'Die Eingabe ist nicht gültig. Der Hinweis am Feld sagt, woran es liegt — berichtige ihn dort.',

  // **Die Regel selbst ist die Abhilfe** (#1350). „Nicht gültig“ sagt nicht,
  // was zu tun ist; diese Zeile schon.
  // **Das Kennwort einer Sicherung** (OPS-06, #1723).
  sicherungskennwortKurz:
    'Ein Kennwort für eine Sicherung hat mindestens zwölf Zeichen. Wähle ein längeres — ein Satz aus vier Wörtern ist gut und leicht zu behalten.',
  sicherungskennwortFehlt:
    'Diese Sicherung ist verschlüsselt. Gib ihr Kennwort ein, um sie anzusehen oder zurückzuspielen.',
  sicherungskennwortFalsch:
    'Das Kennwort passt nicht zu dieser Sicherung. Prüfe es und gib es noch einmal ein.',
  // **Eine MTG-JSON-Datei, die sich nicht lesen lässt** (G01, #2248).
  mtgjsonKeinJSON:
    '{datei} ist kein lesbares JSON — an Byte {stelle} bricht die Datei ab oder ist falsch geformt. Wähle eine andere Datei.',
  mtgjsonKeinSet:
    '{datei} ist JSON, enthält aber kein Set im Format MTG JSON (an Byte {stelle} weicht sie ab). Wähle eine Datei mit einem Set — oder ein ganzes Verzeichnis.',
  sicherungImKlartext:
    'Der Bestand ist verschlüsselt — ohne Kennwort läge die Sicherung im Klartext. Setze unter Einstellungen › Daten und Sicherung ein Kennwort oder wähle dort ausdrücklich „ohne Kennwort sichern“.',

  planVerschluesselt:
    'Der Bestand ist verschlüsselt, und der Plan sichert ohne Kennwort — nach Plan wird er deshalb nicht gesichert. Sichere ihn von Hand mit Kennwort.',

  paketnameUngueltig:
    'Ein Paketname besteht aus Buchstaben, Ziffern, Bindestrichen und Unterstrichen — er wird zum Verzeichnisnamen. Wähle einen aus diesen Zeichen.',
  pflichtfeld: 'Dieses Feld darf nicht leer sein — trage etwas ein.',
  nichtEditierbar:
    'Dieses Feld lässt sich nicht bearbeiten — die Vorlage gibt es nur zum Anzeigen. Ändere den Wert, aus dem es entsteht.',
  berechnetesFeld:
    'Dieses Feld entsteht aus anderen und lässt sich nicht über viele Karten setzen. Die Teilfelder stehen einzeln in der Auswahl.',
  keinSetMitStil:
    'Kein offenes Set benutzt diesen Stil. Öffne ein Set, das ihn benutzt — oder wechsle den Stil des offenen.',
  vorlageFehlt:
    'Für diese Karte ist keine Vorlage eingebunden. Auf der Seite Pakete nachsehen, ob Spiel und Stilbogen des Sets dastehen — ohne sie hat eine Karte keine Felder und keine Flächen.',
  sicherungNeuereFassung:
    'Diese Sicherung stammt aus einer neueren Fassung von MagicSetEditor4Mac. Sie führt {unbekannt}, und diese Fassung reicht bis {stand}. Hole dir die neuere Fassung — oder wähle eine ältere Sicherung.',
  sicherungAeltereFassung:
    'Diese Sicherung stammt aus einer Fassung vor 0.7.0, und diese Fassung zieht sie nicht mehr nach. Ihr fehlen Schemaschritte bis {grundlinie}. Spiele sie mit MagicSetEditor4Mac 0.7.0 zurück; danach öffnet diese Fassung den Bestand.',
  sicherungUnbrauchbar:
    'Diese Datei ist keine Sicherung von MagicSetEditor4Mac. Wähle eine, die diese Anwendung geschrieben hat.',
  skriptFehler:
    'Der Ausdruck ist nicht auswertbar: {meldung} Berichtige ihn an der genannten Stelle.',

  // **Der fehlende zweite Faktor** — Werkstück 7 von Phase 4.12 (#1351).
  //
  // **Er ist keine Abweisung, sondern ein fehlender Schritt**: was dastand,
  // war richtig. Deshalb sagt der Satz, was zu tun ist, und nicht, was falsch
  // war (§6).
  // **Die lokale Sitzung** — ID-02 (#1701): jede Abweisung sagt, was nicht
  // stimmt, statt nur „nicht gültig".
  kennwortFalsch:
    'Dieses Kennwort stimmt nicht. Gib es noch einmal ein — oder setze über den Zettel ein neues.',
  kennwortLaenge:
    'Ein Kennwort hat mindestens {mindestens} und höchstens {hoechstens} Zeichen. Wähle eines in dieser Länge.',
  sperreOhneRueckweg:
    'Ohne Kennwort und ohne lokale Entsperrung ließe sich die Sperre nicht lösen. Setze zuerst ein Kennwort.',
  lokaleEntsperrungFehlt:
    'Lokale Entsperrung ist für diesen Nutzer auf diesem Rechner nicht eingerichtet. Gib stattdessen das Kennwort ein.',
  codeFalsch:
    'Dieser Code stimmt nicht — oder er ist schon verbraucht. Prüfe den Code auf dem Zettel — jeder gilt einmal.',
  zweiterFaktor:
    'Dieser Nutzer verlangt einen zweiten Faktor. Gib den Code aus deiner Authenticator-App ein oder bestätige mit Touch ID.',
  faktorGesperrt:
    'Zu viele falsche Codes. Warte einen Augenblick und versuche es dann noch einmal.',
  sicherungFremdeFassung:
    'Diese Sicherung passt nicht zu dieser Fassung von MagicSetEditor4Mac. Sie nennt {dort}, und diese Anwendung ist {hier}. Spiele sie mit der Fassung zurück, die sie geschrieben hat.',

  // **Die Sätze des Zerteilers und des Auswerters** —
  // #1178.
  //
  // **`durchgereicht` ist die Ausnahme, und sie hat einen Grund**: die 48
  // Sätze der Standardbibliothek nennen Funktions- und Argumentnamen. Ein
  // Eintrag je Funktion wäre eine zweite Abschrift der Bibliothek in zwei
  // Sprachen; dort steht der deutsche Satz, wie er kommt.
  skript: {
    fremdesZeichen: 'das Zeichen {0} gehört nicht zur Sprache',
    einzelnesGleich:
      'ein einzelnes `=`; zugewiesen wird mit `:=`, verglichen mit `==`',
    offeneZeichenkette: 'eine Zeichenkette ohne schließendes Anführungszeichen',
    zuTiefGeschachtelt: 'der Ausdruck ist tiefer geschachtelt als {0} Ebenen',
    keinSkript: 'es gibt kein Skript namens {0}',
    skriptkreis: 'diese Felder lesen sich im Kreis: {0}',
    keinVerweis:
      'mit `+` lässt sich kein Verweis auf ein anderes Skript verketten',
    keineZahl: '{0} ist keine Zahl',
    formUnbekannt: 'die Form {0} lässt sich nicht auswerten',
    nameUnbekannt: 'der Name {0} ist hier nicht bekannt',
    keineFunktion: 'hier wird {0} gerufen, das geht nur bei einer Funktion',
    ohneRumpf: 'diese Funktion hat keinen Rumpf',
    nichtWahrheitswert: '`not` braucht einen Wahrheitswert, hier steht {0}',
    vorzeichenZahl: 'ein Vorzeichen braucht eine Zahl, hier steht {0}',
    zaehlschleifeZahl:
      '{0} einer Zählschleife braucht eine Zahl, hier steht {1}',
    rechnenLinks: '`{0}` rechnet mit Zahlen, links {1}',
    rechnenRechts: '`{0}` rechnet mit Zahlen, rechts {1}',
    vergleichLinks: '`{0}` vergleicht Zahlen, links {1}',
    vergleichRechts: '`{0}` vergleicht Zahlen, rechts {1}',
    keineListe: 'sort_list ordnet eine Liste, hier steht {0}',
    zuTief:
      'die Auswertung geht mehr als {0} Stufen tief — ruft sich hier etwas selbst?',
    schleifeSteht:
      'eine Zählschleife kommt bei {0} nicht voran — die Zahl ist zu groß',
    funktionVerbinden:
      'eine Funktion lässt sich nur mit einer Funktion verbinden, hier steht {0} und {1}',
    fehltNach: 'nach {0} fehlt {1}',
    fehltHier: 'hier fehlt {0}',
    klammerOffen: 'diese {0} wird nicht geschlossen — {1} fehlt',
    ausdruckGehtWeiter: 'hier hört der Ausdruck auf, aber der Text geht weiter',
    zuweisungBrauchtNamen: 'zugewiesen wird an einen Namen',
    durchgereicht: '{0}',
    bedingungWahrheit:
      'eine Bedingung braucht einen Wahrheitswert, hier steht {0}',
    fuerJedenListe: '`for each` geht eine Liste durch, hier steht {0}',
    wahrheitLinks: '`{0}` braucht Wahrheitswerte, links steht {1}',
    wahrheitRechts: '`{0}` braucht Wahrheitswerte, rechts steht {1}',
    bindenFunktion: '`@` bindet Argumente an eine Funktion, hier steht {0}',
  },
} as const
