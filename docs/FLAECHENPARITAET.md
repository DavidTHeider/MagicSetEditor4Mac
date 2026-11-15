# Flächenparität: jede schreibende Objektklasse über die Flächen

**C04** (#2229).
`FAEHIGKEITEN.md` sagt, **dass** eine Fähigkeit an einer Fläche da ist; ein
Wächter hält das. Diese Datei sagt, **ob dieselbe Sache dasselbe ergibt** —
je schreibender Objektklasse, an Fenster, Kommandozeile, MCP und Dienst:

- **Ergebnis** — ein Lauf in `internal/flaechen`, der denselben Vorgang über
  mehr als eine Fläche fährt und das Ergebnis vergleicht.
- **Fehlerklasse** — ein Lauf, der dieselbe Abweisung an jeder Fläche
  verlangt, mit derselben Klasse und demselben Kern
  (`abweisungGleich`, seit C04 **samt Fenster**, oder von Hand ebenso).
  **Seit F03** (#2342)
  verlangt `abweisungGleich` den Code auch **maschinenlesbar**: an der
  Kommandozeile mit `--json` (`{"fehler":{"code":…}}`), an MCP in
  `structuredContent.fehler` — Fenster und Dienst nannten ihn schon.
- **Rolle** — ein Lauf, der als Leser oder ohne Rolle schreibt und an jeder
  Fläche das Nein **an der Wirkung** prüft. **Sie entfällt, wo kein Profil
  Träger ist**: gehört jede Bindung einer Klasse dem Nutzer oder dem Rechner
  (`AUTORISIERUNG.md`), gibt es keine Rolle, die abzuweisen wäre. Das
  entscheidet nicht diese Tabelle, sondern die Kataloge — Bindung →
  Fähigkeit in `FAEHIGKEITEN.md`, Fähigkeit → Träger in `AUTORISIERUNG.md` —,
  und der Wächter hält beide Richtungen.

**Die Zuordnung Bindung → Klasse** steht in der zweiten Spalte, und der
Wächter `internal/werkzeuge/flaechenparitaet_test.go` hält sie gegen die
293 schreibenden Bindungen aus `SCHREIBWEGE.md`: **eine neue schreibende
Bindung ohne Klasse macht die Pipeline rot**. Er hält auch, dass jeder
genannte Lauf existiert und die Bilanz unten stimmt.

**Was er nicht hält**: ob ein Lauf vergleicht, was er verspricht — das liest
ein Mensch; und welche Flächen eine Klasse haben **sollte** — das steht in
`FAEHIGKEITEN.md` mit Grund.

## Bilanz der Zuordnung (2026-10-09)

Von **36** Klassen gibt es **4** nur am Fenster; der Nenner
für den Vergleich über Flächen ist **32**. **Für die Rolle sind es 29**:
Nutzer, Sitzung/Anmeldung und Sprachpaket haben keinen Profilträger.

| Gegenstand | Klassen mit zugeordnetem Lauf | Quote |
| --- | --- | --- |
| Ergebnis | 32 von 32 | 100 % |
| Fehlerklasse | 32 von 32 | 100 % |
| Rolle | 29 von 29 | 100 % |

**Jede Klasse hat ihre drei Läufe — zugeordnet.** Welche Flächen sie fahren,
steht unter „Gemessene Flächen". Wo eine Fläche die Fähigkeit nicht
hat, steht der Grund in der letzten Spalte und im Lauf selbst (`ohneMCP`,
`ohneKommandozeile`, `ohneFenster`) — ein Fall ohne Lauf und ohne Grund ist
rot. Das Fenster fährt die Rollenlagen in `internal/app`
(`AUTORISIERUNG.md`), weil die Rollenläufe hier parallel laufen und das
Zuhause eines Prozesses nur eines ist.

**Was die Vollständigkeit gekostet hat, war ein Befund**: der Rollenlauf
über die übrigen Klassen fand, dass `mse symbol-setzen` und
`mse symbol-nachzeichnen` als Leser — und ganz ohne Rolle — in ein fremdes
Profil schrieben. Die Kommandozeile schützt einen Leser über eine
Verbindung, die nichts annimmt; eine Zeichnung ist aber eine Datei. Beide
Nähte fragen jetzt die Rolle, bevor sie ablegen
(`internal/app/symbolrolle_intern_test.go`).

## Gemessene Flächen

**Die Bilanz oben zählt eine Zuordnung, keine Parität** — Zweitblick über
Phase C, Befund 3. Sie sagt, dass einer Klasse ein Lauf zugeordnet ist; ob
er alle Flächen fährt, sagt sie nicht. Gelesen fuhr **kein Ergebnislauf das
Fenster**: es kam nur in den Abweisungen vor (`abweisungGleich`).

**Diese Bilanz ist gemessen.** Der Wächter
`internal/werkzeuge/flaechenparitaetflaechen_test.go` liest je Lauf aus dem
Quelltext von `internal/flaechen`, welche Flächen er anspricht — von dem Lauf
aus jede Funktion des Pakets, die er ruft oder weiterreicht, und darin die
Helfer `fensterAus` (Fenster), `kommandozeile`, `mcp…` und `starteDienst…`,
`dienstGelang`, `.ruf…` (Dienst). Je Klasse und Spalte steht die Vereinigung
ihrer Läufe; eine Zeile, die nicht stimmt, macht ihn rot, und er nennt sie,
wie sie lauten soll.

**Was er nicht hält**: ob eine angesprochene Fläche **verglichen** wird — eine,
die nur zum Einrichten gerufen wird, zählt mit. Und `app.App` direkt
(`PeerflaecheUeber`) ist kein Fenster: ohne `NeueBindungen` fehlen Tore,
Rufkontext und die Abbildung von Wails. **Die Rolle des Fensters** prüft
`internal/app` je Rollenlage (`AUTORISIERUNG.md`); sie steht hier nicht, weil
dieser Wächter nur `internal/flaechen` liest.

**Die Klasse `validierung` trennt `abweisungGleich` an Kommandozeile und MCP
nicht** von anderen: dort steht kein Vorsatz wie „nicht gefunden:", und MCP
nennt keinen Code — die Zusage trägt dann der Kern. Fenster und Dienst nennen
die Klasse.

**Das Fenster im Ergebnis fährt eine Kohorte nach Risiko** (C04): die Klassen,
in denen ein Mensch etwas schreibt, das er verlieren kann — Karte/Feld,
Exemplar/Sammlung, Deck, Set, Handel, Einstellung, Stil und die Sicht auf die
Sammlung — **und seit P1a (#2335) die Partie**: die Abwahl beim Anfangen ging
am Fenster verloren, während die drei übrigen Flächen sie gar nicht kannten.
Die übrigen vergleichen Kommandozeile, MCP und Dienst; das Fenster
prüfen dort Läufe in `internal/app` oder keiner, und ihre Köpfe sagen es.

Klassen je Fläche, von 32 Klassen mit mehr als einer Fläche (Rolle: von 29):

| Fläche | Ergebnis | Fehlerklasse | Rolle |
| --- | --- | --- | --- |
| Fenster | 10 | 32 | 0 |
| Kommandozeile | 32 | 32 | 29 |
| MCP | 30 | 31 | 28 |
| Dienst | 32 | 32 | 29 |

| Klasse | Ergebnis | Fehlerklasse | Rolle |
| --- | --- | --- | --- |
| Partie/Tisch | F K M R | F K M R | K M R |
| Kontakt/Peer | K M R | F K M R | K M R |
| Sitzung/Anmeldung | K M R | F K M R | entfällt |
| Turnier | K M R | F K M R | K M R |
| Handel | F K M R | F K M R | K M R |
| Karte/Feld | F K M R | F K M R | K M R |
| Paket | F K M R | F K M R | K M R |
| Deck | F K M R | F K M R | K M R |
| KI | K M R | F K M R | K M R |
| Symbol | K M R | F K M R | K M R |
| Bestand | K M R | F K M R | K M R |
| Profil/Freigabe | K M R | F K M R | K M R |
| Exemplar/Sammlung | F K M R | F K M R | K M R |
| Nutzer | K M R | F K M R | entfällt |
| Marke | K M R | F K M R | K M R |
| Set | F K M R | F K M R | K M R |
| Ausgabe | K M R | F K M R | K M R |
| Sprachpaket | K M R | F K M R | entfällt |
| Token | K R | F K M R | K R |
| Umschlag | K M R | F K M R | K M R |
| Autoersetzung | K M R | F K M R | K M R |
| Betrieb | K M R | F K M R | K M R |
| Einstellung | F K M R | F K M R | K M R |
| Preise/Markt | K M R | F K M R | K M R |
| Scryfall | K M R | F K M R | K M R |
| Format | K M R | F K M R | K M R |
| Kartenbild | K M R | F K M R | K M R |
| Schlüsselwort | K M R | F K M R | K M R |
| Stil | F K M R | F K M R | K M R |
| Sicht/Freigabe Sammlung | F K R | F K R | K M R |
| Verlauf | K M R | F K M R | K M R |
| Kosten | K M R | F K M R | K M R |

## Je Klasse

| Klasse | Schreibende Bindungen | Ergebnis | Fehlerklasse | Rolle | Ohne die Fähigkeit |
| --- | --- | --- | --- | --- | --- |
| Partie/Tisch | 21: `PartieAnlegen`, `PartieAusfuhrSpeichern`, `PartieBeenden`, `PartieBefehl`, `PartieBeitreten`, `PartieBeitrittBeenden`, `PartieEinfuhrOeffnen`, `PartieFortsetzen`, `PartieFreigabeBeenden`, `PartieFreigeben`, `PartieKopieren`, `PartieLoeschen`, `PartieOffenlegen`, `PartieOffenlegungZuruecknehmen`, `PartiePausieren`, `PartieSagen`, `PartieVergessen`, `PartieWuenschen`, `PartieZeigen`, `TischabzugPDFSpeichern`, `TischabzugSpeichern` | `TestEinePartieEntstehtAnJederFlaecheGleich`, `TestDieAbwahlEntstehtAnJederFlaecheGleich`, `TestEinSpielbefehlWirktAnJederFlaecheUndNurEinmal`, `TestDasLoeschenEinerPartieGiltAnJederFlaeche`, `TestDasKopierenEinerPartieGiltAnJederFlaeche` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden`, `TestEineUnbekannteAbwahlWirdAnJederFlaecheAbgewiesen` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | Gäste, Beitritt, Reden, Einwilligung und Tischabzug nur am Fenster — der Tisch überlebt den Ruf, die Einwilligung ist Sache eines Menschen |
| Kontakt/Peer | 19: `FernHorchBeenden`, `FernHorchStarten`, `HorchBeenden`, `HorchStarten`, `KontaktLoeschen`, `KontakteUebernehmen`, `KontakteinladungAnnehmen`, `PeeridentitaetRotieren`, `PeeridentitaetSicherstellen`, `PeerkontaktAdresse`, `PeerkontaktAufnehmen`, `PeerkontaktBestaetigen`, `PeerkontaktEntsperren`, `PeerkontaktNotiz`, `PeerkontaktSperren`, `PeerzweckEntziehen`, `PeerzweckErlauben`, `SichtbarWerden`, `SichtbarkeitBeenden` | `TestDieKontakteUeberJedeFlaeche`, `TestEinladungUeberJedeFlaeche` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | MCP bei den Vertrauensentscheidungen — sie gehören einem Menschen |
| Deckleihe/Bedienung | 17: `Anheften`, `BefehlAbbrechen`, `BefehlAusfuehren`, `DarstellungsstandMelden`, `DeckBenutzt`, `DeckleiheAnbieten`, `DeckleiheAnnehmen`, `DeckleiheWiderrufen`, `DeckleiheZurueckgeben`, `FensterbildEntfernen`, `FensterbildWaehlen`, `Loesen`, `ObjekthandlungenMelden`, `SetBenutzt`, `SkriptAbbrechen`, `TischbildEntfernen`, `TischbildWaehlen` | — | — | — | **nur am Fenster** — Bedienung ohne Bestand, die Leihe ist Sache eines Menschen |
| Sitzung/Anmeldung | 16: `Abmelden`, `Anmelden`, `BiometrieEinrichten`, `Entsperren`, `EntsperrenMitSensor`, `FaktorBestaetigen`, `FaktorEinrichten`, `FaktorEntfernen`, `KennwortWechseln`, `KennwortWiederherstellen`, `KontoAnmelden`, `KontoAnonym`, `KontoEintragen`, `KontoLoesen`, `LokalAnmelden`, `Sperren` | `TestDieKontenAntwortenAnJederFlaecheGleich` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | entfällt — Träger Nutzer | An- und Abmelden, Sperren und Kennwörter nur am Fenster — die Sitzung gehört dem Fenster |
| Turnier | 16: `TurnierAbschliessen`, `TurnierAnlegen`, `TurnierAnmelden`, `TurnierAusgabeSpeichern`, `TurnierAussteigen`, `TurnierBerichtSpeichern`, `TurnierEinlesenOeffnen`, `TurnierErgebnisSetzen`, `TurnierKopieren`, `TurnierLoeschen`, `TurnierMatchAusDerPartie`, `TurnierMatchBinden`, `TurnierMatchMelden`, `TurnierMatchUebernehmen`, `TurnierRundeLegen`, `TurnierWiedereroeffnen` | `TestEinTurnierSchliesstUeberallGleichAb`, `TestEinTurnierReistUeberallGleich`, `TestEinTurnierLaesstSichUeberallNachVorlageAnlegen` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestWeitereSchreibklassenFolgenDerRolle` | `TurnierMatch*` nur am Fenster — sie hängen an einer Partie am Tisch |
| Handel | 15: `HandelArchivieren`, `HandelEntwurfAendern`, `HandelEntwurfLoeschen`, `HandelEroeffnen`, `HandelFesthalten`, `HandelFreigabeErteilen`, `HandelKopieren`, `HandelPostbestaetigung`, `HandelSchicken`, `HandelSchritt`, `HandelVorschlagen`, `HandelZustellen`, `Handelsabgleich`, `Handelsbestaetigung`, `Handelsverbuchung` | `TestEinHandelKommtUeberBeideFlaechenBisZumAbschluss`, `TestDieVerbuchungWeistUeberJedeFlaecheGleichAb`, `TestDieHandelslisteIstUeberDreiFlaechenDieselbe` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | MCP schreibt keinen Handel; `HandelFreigabeErteilen` nur am Fenster ([0078](entscheidungen/)) |
| Karte/Feld | 16: `AblageAusschneiden`, `AblageEinfuegen`, `ErsetzungAusfuehren`, `FeldSetzen`, `FelderSetzenMitErwartung`, `KarteAnlegen`, `KarteKopieren`, `KarteLoeschen`, `KarteVerweisLoesen`, `KarteVerweisSetzen`, `KartenTauschen`, `KartenskriptAusfuehren`, `ListensortierungSetzen`, `ListenspaltenSetzen`, `WortWaehlen`, `ZusatzwertSetzen` | `TestDerKartenlebenslaufUeberJedeFlaeche`, `TestEineKarteWandertUeberJedeFlaecheOhneVerlust`, `TestEinFeldAufVielenKartenUeberJedeFlaeche`, `TestErsetzenUeberJedeFlaecheOhneFenster`, `TestEinZusatzwertGehtUeberDreiFlaechen` | `TestEineFremdeKarteWirdUeberallAbgewiesen`, `TestEinUnvollstaendigerAusdruckScheitertUeberallGleich`, `TestEinZusatzwertGehtUeberDreiFlaechen` | `TestWeitereSchreibklassenFolgenDerRolle`, `TestEinLeserSchreibtAuchUeberMCPUndDienstNicht` | — |
| Paket | 17: `InstallerEinlesen`, `PaketAnlegen`, `PaketBeziehen`, `PaketBezugAktualisieren`, `PaketBezugZurueck`, `PaketEinbindenVon`, `PaketEntfernen`, `PaketErsetzen`, `PaketKopieren`, `PaketWeitergeben`, `PaketZurueck`, `PaketdateiAnlegen`, `PaketdateiEinfuegen`, `PaketdateiEntfernen`, `PaketdateiSchreiben`, `PaketdateiUmbenennen`, `VorlagenUebernehmen` | `TestEinPaketWirdAnJederFlaecheAngelegt`, `TestEinPaketWirdAnJederFlaecheKopiert`, `TestEinPaketWirdAnJederFlaecheEingebunden`, `TestPlanErsetzenUndZurueckAnJederFlaeche`, `TestDieBezuegeStehenAnJederFlaeche` | `TestEinZweigStattEinesStandsWirdUeberallAbgewiesen`, `TestEinFremdesPaketWirdUeberallAbgewiesen`, `TestEinZweigWirdNirgendsAktualisiert`, `TestOhneAufgehobenenStandGibtEsKeinZurueck`, `TestEineQuelleOhneAuskunftWirdUeberallGleichAbgewiesen` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | `PaketdateiEinfuegen` nicht an MCP und Dienst — es liest eine Datei mit den Rechten dessen, der vor dem Rechner sitzt |
| Deck | 14: `DeckAnlegen`, `DeckAusSet`, `DeckFehlkartenWuenschen`, `DeckFormatWaehlen`, `DeckKopieren`, `DeckLoeschen`, `DeckUebernehmen`, `DeckUmbenennen`, `DeckZuordnen`, `DeckZuordnungLoesen`, `DecklisteUebernehmen`, `DeckplatzLegen`, `DeckplatzNehmen`, `DeckplatzVerschieben` | `TestEinDeckEntstehtUeberallGleich`, `TestEinDeckLaesstSichUeberallKopieren`, `TestAusEinemSetWirdUeberallEinDeck`, `TestDerBereichswechselGehtUeberAlleFlaechen` | `TestEinDeckMitVeralteterFassungWirdAnJederFlaecheAbgewiesen` | `TestDieRollenmatrixUeberDieFlaechen` | — |
| KI | 14: `KIBatchVorschlagen`, `KIBildAbbrechen`, `KIBildEinsetzen`, `KIBildVerwerfen`, `KIBildVorschlagen`, `KIEinschaetzen`, `KIFreigabeErteilen`, `KIGeraetemodellFuerSitzung`, `KIGruppeUebernehmen`, `KIKarteErkennen`, `KIModellPruefen`, `KIModellVergessen`, `KIVorschlagUebernehmen`, `KIVorschlagen` | `TestEineMessungZuVergessenGehtUeberJedeFlaeche`, `TestEineUebernahmeUeberJedeFlaeche` | `TestEinKIVorschlagOhneLaufzeitUeberJedeFlaeche`, `TestEinZuGrosserBatchUeberJedeFlaeche`, `TestEineGruppeUeberJedeFlaeche`, `TestEinFehlendesKartenfotoUeberJedeFlaeche` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | KI-Bild und `KIFreigabeErteilen` nur am Fenster |
| Symbol | 11: `BibliothekssymbolAufnehmen`, `BibliothekssymbolAusgeben`, `BibliothekssymbolEinlesen`, `BibliothekssymbolEinsetzen`, `BibliothekssymbolKopieren`, `BibliothekssymbolLoeschen`, `BibliothekssymbolNachziehen`, `BibliothekssymbolSichern`, `BibliothekssymbolUmbenennen`, `BibliothekssymbolZuruecknehmen`, `SymbolSetzen` | `TestDieSymbolbibliothekReistUeberallGleich`, `TestEinSymbolWirdUeberallGleichBearbeitet` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | — |
| Bestand | 13: `BestandSichern`, `Pflegen`, `PlatzFreigeben`, `RueckkehrQuittieren`, `RuecksetzungFreigabeErteilen`, `SicherungZurueckspielen`, `SicherungsplanBeenden`, `SicherungsplanEinrichten`, `StartWiederholen`, `VerschluesselungUmschalten`, `WechselVorbereiten`, `Zuruecksetzen`, `ZwischenspeicherLeeren` | `TestDenBestandSichernUeberJedeFlaeche`, `TestWeggeraeumtIstWeggeraeumtAnJederFlaeche` | `TestEineSicherungUeberschreibtNirgendsEtwas` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | Zurückspielen nicht an MCP und Dienst; Zurücksetzen, Pflegen und Rückkehr nicht an MCP; der erneute Start nur am Fenster (F05); der Plan der Sicherung nur am Fenster — an den übrigen Flächen ist er eine Einstellung (OPS01) |
| Profil/Freigabe | 9: `FreigabeLoeschen`, `FreigabeSetzen`, `ProfilAendern`, `ProfilAnlegen`, `ProfilLoeschen`, `ProfilWechseln`, `ProfilbadgeEntfernen`, `ProfilbadgeWaehlen`, `ProfilvorgabeSetzen` | `TestGesetztIstGesetztAnJederFlaeche`, `TestEineSelbstHerabgestufteFreigabeGiltAnJederFlaeche`, `TestDerGanzeAbzeichenwegAnJederFlaeche` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestDerLeserwechselGehtUeberMCPUndDienstNicht` | — |
| Exemplar/Sammlung | 8: `ExemplarAendern`, `ExemplarAnlegen`, `ExemplarLoeschen`, `ListeEinlesen`, `MasseAendern`, `MasseLoeschen`, `MassePreise`, `SammlungCSVEinlesen` | `TestEinExemplarUeberJedeFlaeche`, `TestEinMassenvorgangLaeuftUeberallGleich`, `TestLoeschenInMasseLaeuftUeberallGleich`, `TestEineListeWirdUeberallGleichEingelesen`, `TestEineMassenaenderungLaesstSichUeberallZuruecknehmen` | `TestEinExemplarMitVeralteterFassungWirdAnJederFlaecheAbgewiesen` | `TestWeitereSchreibklassenFolgenDerRolle` | — |
| Nutzer | 8: `NutzerAendern`, `NutzerAnlegen`, `NutzerLoeschen`, `NutzerWechseln`, `NutzerbildEntfernen`, `NutzerbildWaehlen`, `NutzereinstellungSetzen`, `PersonendatenSichern` | `TestEinNutzerWirdAnJederFlaecheAngelegt`, `TestJedeFlaecheGibtEinenBildnamenDerselbenGestalt` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | entfällt — Träger Nutzer | — |
| Dienste/Schnittstellen | 9: `Befehlszeile`, `MCPBeenden`, `MCPStarten`, `RESTBeenden`, `RESTStarten`, `SchnittstelleEinrichten`, `SchnittstelleWiderrufen`, `SchnittstellenAlleEinrichten`, `SchnittstellenAlleWiderrufen` | — | — | — | **nur am Fenster** — Start und Ende der Dienste, Einrichtung der Schnittstellen |
| Marke | 7: `MarkeEntfernen`, `MarkeSetzen`, `MarkeUeberallEntfernen`, `MarkeUmbenennen`, `MarkeVergeben`, `ObjektmarkeEntfernen`, `ObjektmarkeSetzen` | `TestEineMarkeErreichtJedeFlaeche`, `TestMarkenAnJedemObjektUeberJedeFlaeche` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | MCP bei `MarkeUmbenennen` und `MarkeUeberallEntfernen` — ein Werkzeug schreibt nur, was es nicht überschreibt ([0044](entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)) |
| Set | 7: `BeispielsetOeffnen`, `SetAnlegen`, `SetAusDatei`, `SetLoeschen`, `SetSpeichern`, `SetSpeichernUnter`, `SetwertSetzen` | `TestEinNeuesSetUeberJedeFlaeche`, `TestEinSetLoeschenUeberJedeFlaecheOhneFenster`, `TestDieSetangabenUeberJedeFlaeche` | `TestEineSetangabeOhneNamenWirdUeberallAbgewiesen`, `TestBeideFlaechenWeisenDasselbeMitDemselbenGrundAb` | `TestWeitereSchreibklassenFolgenDerRolle` | — |
| Ausgabe | 6: `BildSchreiben`, `BogenLegen`, `ExportoptionSetzen`, `MTGJSONOeffnen`, `MTGJSONSchreiben`, `VorlageAusfuehren` | `TestEinDruckbogenEntstehtAnJederFlaeche`, `TestEineExportvorlageLaeuftAnJederFlaeche`, `TestMTGJSONWirdAnJederFlaecheGeschrieben` | `TestBeideFlaechenWeisenDasselbeMitDemselbenGrundAb` | `TestDieRollenmatrixUeberDieFlaechen` | — |
| Sprachpaket | 5: `SprachpaketAendern`, `SprachpaketAnlegen`, `SprachpaketAusgabeSpeichern`, `SprachpaketEinlesenOeffnen`, `SprachpaketLoeschen` | `TestEineEigeneUebersetzungGehtUeberallGleich` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | entfällt — Träger Rechner | — |
| Token | 5: `TokenAendern`, `TokenAlleWiderrufen`, `TokenAnlegen`, `TokenLoeschen`, `TokenRotieren` | `TestRotationUndSammelwiderrufUeberZeileUndDienst` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestEinLesetokenSchreibtAnKeinerStelle`, `TestKeinSchreibenderWegNimmtEinLesetokenAn`, `TestEinTokenWirktNurAnSeinemProfil` | — |
| Umschlag | 5: `UmschlagDeck`, `UmschlagKarte`, `UmschlagSchnueren`, `UmschlagUebernehmen`, `UmschlagVorlage` | `TestDenUmschlagUeberJedeFlaeche`, `TestEinUmschlagKommtAnJederFlaecheAlsNeuesSetHerein` | `TestEineFremdeDateiWirdUeberallAbgewiesen` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | `UmschlagDeck`, `-Karte`, `-Vorlage` nur am Fenster |
| Autoersetzung | 4: `AutoersetzungLoeschen`, `AutoersetzungSchalten`, `AutoersetzungSetzen`, `AutoersetzungenZuruecksetzen` | `TestEineEigeneAutoersetzungEntstehtUeberallGleich`, `TestSchalterUndZuruecksetzenGreifenUeberall` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestWeitereSchreibklassenFolgenDerRolle` | — |
| Betrieb | 4: `AuftragAbbrechen`, `IndexAbbrechen`, `IndexAufbauen`, `NeueFassungLaden` | `TestDerSuchindexGehtUeberDreiFlaechen`, `TestDieAktualisierungUeberJedeFlaeche` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | MCP baut keinen Index und lädt keine Fassung |
| Diagnose | 4: `DiagnoseSichern`, `DiagnosepaketLoeschen`, `DiagnosepaketSichern`, `Leistungsprofil` | — | — | entfällt — Träger Rechner | **nur am Fenster** — Text und Zustimmung stehen am Fenster |
| Einstellung | 4: `EinstellungSetzen`, `EinstellungenSichern`, `EinstellungenUebernehmen`, `SpracheSetzen` | `TestEineEinstellungGiltAnAllenFlaechen`, `TestDenNetzwegUeberJedeFlaeche` | `TestEineEinstellungOhneSchluesselWirdUeberallAbgewiesen`, `TestEinProxyOhneSchemaWirdUeberallAbgewiesen`, `TestJederSchalterWeistAnDreiFlaechenGleichAb` | `TestEinLeserLiestUndSchreibtNichtAnDerZeile`, `TestWeitereSchreibklassenFolgenDerRolle` | — |
| Preise/Markt | 4: `Fremdpreisverlauf`, `Marktangebote`, `Preisbulk`, `PreiseHolen` | `TestPreiseHolenUndLesenUeberJedeFlaeche`, `TestDerBulkbezugUeberJedeFlaeche` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | MCP und Dienst nur lesend bei `Marktangebote` |
| Scryfall | 4: `ScryfallKarteAnlegen`, `ScryfallSammeldatenHolen`, `ScryfallSatzEinlesen`, `ScryfallUebernehmen` | `TestEineKarteAnlegenUeberJedeFlaeche`, `TestEinenSatzEinlesenUeberJedeFlaeche`, `TestKartendatenUebernehmenUeberJedeFlaeche`, `TestDieSammeldateiHolenUeberJedeFlaeche` | `TestEinUnbekannterKartennameScheitertUeberallGleich`, `TestEinUnbekannterSatzScheitertUeberallGleich`, `TestEinUnbekanntesSpielScheitertUeberallGleich`, `TestEineErgebnisloseScryfallsucheScheitertUeberallGleich` | `TestDieRollenmatrixUeberDieFlaechen` | — |
| Zugang/Geheimnis | 4: `AnbietergeheimnisLoeschen`, `AnbietergeheimnisSetzen`, `ZugangLoeschen`, `ZugangSetzen` | — | — | entfällt — Träger Nutzer | **nur am Fenster** — der Schlüsselbund fragt je Prozess, und ein Geheimnis reist in keinem Aufruf |
| Format | 3: `FormatAusBauart`, `FormatLoeschen`, `FormatSetzen` | `TestEinFormatEntstehtUeberallGleich`, `TestEinFormatLaesstSichUeberallFortschreiben` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestWeitereSchreibklassenFolgenDerRolle` | — |
| Kartenbild | 3: `BildEinsetzen`, `BildSetzen`, `ScryfallBildEinsetzen` | `TestEinKartenbildWirdAnJederFlaecheEingesetzt` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | — |
| Schlüsselwort | 3: `SchluesselwortAnlegen`, `SchluesselwortLoeschen`, `SchluesselwortSetzen` | `TestDieSchluesselwoerterUeberJedeFlaeche`, `TestEinSchluesselwortAnlegenUndLoeschenUeberJedeFlaeche` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestWeitereSchreibklassenFolgenDerRolle` | — |
| Stil | 3: `StilWechseln`, `StiloptionSetzen`, `StiloptionZuruecksetzen` | `TestDieStileUeberJedeFlaeche`, `TestDieStiloptionenUeberJedeFlaeche` | `TestEinFremderStilWirdUeberallAbgewiesen` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | — |
| Sicht/Freigabe Sammlung | 2: `SammlungFreigeben`, `SammlungsfreigabeWiderrufen` | `TestDieAngabenEinerSichtGeltenAnJederFlaeche`, `TestDieFreigegebeneSichtUeberJedeFlaeche` | `TestEineAuswahlOhneDruckeWirdAnJederFlaecheAbgewiesen` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | MCP — ein Modell entscheidet nicht, wer die Sammlung sieht |
| Verlauf | 2: `Wiederherstellen`, `Zuruecknehmen` | `TestDieRuecknahmeEinerWortaenderungUeberJedeFlaeche`, `TestEineMassenaenderungGehtGezieltZurueckUeberJedeFlaeche` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestWeitereSchreibklassenFolgenDerRolle` | — |
| Kosten | 1: `KostenbudgetSetzen` | `TestDasKostenjournalGehtUeberDreiFlaechen` | `TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden` | `TestDieUebrigenSchreibklassenFolgenDerRolle` | MCP setzt keine Grenze — die Grenze ist die Freigabe |
