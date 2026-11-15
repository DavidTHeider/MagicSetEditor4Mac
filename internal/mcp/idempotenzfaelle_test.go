// **Die Argumente, mit denen jedes idempotente Werkzeug zweimal gerufen
// wird** — abgetrennt am 2026-09-02, weil `idempotenz_test.go` mit dem
// Symboleditor über die Zeilengrenze aus §2 ging
// (#1391).
//
// # Der Schnitt folgt der Frage
//
// Dort steht die **Prüfung** — zweimal rufen, vergleichen; hier die
// **Fälle**. Die Tabelle wächst mit jedem Werkzeug, die Prüfung nicht.

package mcp_test

// idempotenzfaelle sind die Argumente, mit denen ein Werkzeug zweimal gerufen
// wird.
//
// **Je Werkzeug einer, und die Liste ist vollständig zu halten** — ein
// Werkzeug ohne Fall käme sonst still durch. Genau das prüft
// [TestJedesWerkzeugHatEinenIdempotenzfall].
// **`#nosec G101`**: seit Werkstück 10 (#1351) steht das Wort `token` in
// dieser Tabelle, und gosec hält jede Zeile mit diesem Wort für ein
// eingebautes Geheimnis. Was hier steht, sind **Argumente** für Werkzeuge.
var idempotenzfaelle = mitHandelsfaellen(mitPaketfaellen(map[string]string{ // #nosec G101 -- Werkzeugargumente
	// **Ein Set, das es nicht gibt.** Der Fall darf hier nichts löschen: der
	// Wächter ruft zweimal, und die Antworten sollen vergleichbar sein.
	// `schreiben.set.loeschen` sagt keine Idempotenz zu — beide Rufe scheitern
	// gleich, und genau das ist ehrlich: nach dem ersten Löschen gibt es das
	// Set nicht mehr.
	"schreiben.set.loeschen":     `{"setID": 4711, "name": "Gibt es nicht"}`,
	"lesen.sets":                 `{}`,
	"lesen.karten":               `{"setID": 1}`,
	"lesen.felder":               `{"setID": 1, "karteID": 7}`,
	"lesen.suche":                `{"text": "Flammenball"}`,
	"lesen.setangaben":           `{"setID": 1}`,
	"lesen.vorbehalte":           `{"setID": 1}`,
	"lesen.set.probleme":         `{"setID": 1}`,
	"lesen.set.probleme.pruefen": `{"setID": 1}`,
	"lesen.beispielsets":         `{}`,
	"lesen.stile":                `{"setID": 1}`,
	"lesen.stilfolgen":           `{"setID": 1, "paket": "probe-einzeiler"}`,
	"lesen.stiloptionen":         `{"setID": 1}`,
	"lesen.pakete":               `{}`,
	"lesen.statistik":            `{"setID": 1}`,
	"lesen.verteilung":           `{"setID": 1, "feld": "name"}`,
	"lesen.einstellungen":        `{}`,

	// **Der Platz** — Werkstück 4 von Phase 4.13 (#1352). Das Wegräumen nennt
	// einen Posten, den es **gibt**, aber der Bestand der Vorrichtung hat
	// keine Vorsicherungen: beide Rufe finden nichts und antworten gleich.
	// Genau das ist die Zusage.
	"lesen.platz":                        `{}`,
	"schreiben.platz.freigeben":          `{"posten": "vorsicherungen"}`,
	"schreiben.stil":                     `{"setID": 1, "paket": "probe-einzeiler"}`,
	"schreiben.stiloption":               `{"setID": 1, "feld": "rand", "wert": "breit"}`,
	"schreiben.stiloption.zuruecksetzen": `{"setID": 1, "feld": "rand"}`,
	"schreiben.paket.beziehen":           `{"adresse": "https://example.invalid/x", "stand": "abc1234", "unterverzeichnis": "data", "name": "probe"}`,
	"schreiben.zuruecknehmen":            `{"setID": 1}`,
	"schreiben.wiederherstellen":         `{"setID": 1}`,
	"schreiben.einstellung":              `{"schluessel": "sprache", "wert": "en"}`,
	"schreiben.setangabe":                `{"setID": 1, "feld": "title", "wert": "x"}`,
	"schreiben.feld":                     `{"setID": 1, "karteID": 7, "feld": "name", "wert": "x"}`,
	"schreiben.karte.anlegen":            `{"setID": 1}`,
	"schreiben.karte.kopieren":           `{"setID": 1, "karteID": 7}`,
	"schreiben.karte.loeschen":           `{"setID": 1, "karteID": 7}`,
	"schreiben.karten.tauschen":          `{"setID": 1, "eineID": 7, "andereID": 8}`,
	"schreiben.set.einfahren":            `{"pfad": "/tmp/probe.mse-set"}`,
	"schreiben.set.sichern":              `{"setID": 1, "ziel": "/tmp/raus.mse-set"}`,
	"schreiben.karte.zeichnen":           `{"karteID": 7, "ziel": "/tmp/karte.png"}`,
	"schreiben.set.bogen":                `{"setID": 1, "ziel": "/tmp/bogen.pdf"}`,
	"lesen.mse2orte":                     `{}`,
	"lesen.vorlagen":                     `{"setID": 1}`,
	"schreiben.vorlage.ausfuehren": `{"setID": 1, ` +
		`"paket": "probe-liste.mse-export-template", "ziel": "/tmp/raus.txt"}`,
	"schreiben.vorlagenoption": `{"vorlage": "probe-liste.mse-export-template", ` +
		`"feld": "trenner", "wert": ", "}`,
	"schreiben.karte.bild": `{"setID": 1, "karteID": 7, "feld": "kunst", ` +
		`"pfad": "/tmp/bild.png"}`,
	"lesen.nutzer":         `{}`,
	"schreiben.nutzer.neu": `{"name": "Zweite"}`,
	"schreiben.profil.neu": `{"nutzerID": "1", "name": "Turnier"}`,

	// **Das anonyme Konto ist nicht idempotent** — Werkstück 6 (#1351), und
	// das ist seine Sache: jeder Ruf legt einen **anderen** Nutzer an, weil
	// jeder eine eigene Zufallskennung bekommt. Zweimal dasselbe zu bekommen
	// hieße, zwei Menschen wären derselbe.
	"schreiben.konto.anonym": `{}`,

	// **Das Herausnehmen ist idempotent** — Werkstück 7 (#1351): im zweiten
	// Lauf gibt es den Faktor nicht mehr, und „gibt es nicht" ist derselbe
	// Zustand wie „entfernt".
	"lesen.faktoren":            `{"nutzerID": "1"}`,
	"schreiben.faktor.loeschen": `{"nutzerID": "1", "art": "totp"}`,

	// **Der Wechsel ist idempotent** — Werkstück 3 (#1351): zweimal auf
	// dasselbe Profil zu wechseln ist einmal wechseln, und das zweite Mal
	// schreibt gar nicht (siehe `service.Nutzerverwaltung.Wechseln`).
	"lesen.profil":              `{}`,
	"schreiben.profil.wechseln": `{"profilID": "1"}`,

	// **Und der Nutzerwechsel ebenso** — Werkstück 17 (#1400): wer schon
	// dieser Mensch ist, wird es nicht noch einmal, und der Ruf schreibt gar
	// nicht (siehe `service.Nutzerverwaltung.NutzerWechseln`).
	"lesen.werarbeitet":         `{}`,
	"lesen.kontext":             `{}`,
	"schreiben.nutzer.wechseln": `{"nutzerID": "1"}`,

	// **Und das Lösen ist idempotent** — im zweiten Lauf gibt es das Konto
	// nicht mehr, und „gibt es nicht" ist derselbe Zustand wie „gelöst".
	"lesen.anbieterkonten":   `{"nutzerID": "1"}`,
	"schreiben.konto.loesen": `{"nutzerID": "1", "anbieter": "github"}`,

	// **Ändern ist idempotent** — derselbe Name zweimal gesetzt ist derselbe
	// Zustand. **Und Löschen auch**: beim zweiten Mal gibt es das Ziel nicht
	// mehr, und „gibt es nicht" ist derselbe Zustand wie „gelöscht".
	"schreiben.nutzer.aendern":  `{"nutzerID": "1", "name": "Ich"}`,
	"schreiben.profil.aendern":  `{"profilID": "1", "name": "Mein Profil"}`,
	"schreiben.nutzer.loeschen": `{"nutzerID": "2", "name": "Zweite"}`,
	"schreiben.profil.loeschen": `{"profilID": "2", "name": "Turnier"}`,

	// **Das Profilbild ist idempotent** — Werkstück 8 (#1351). Dieselbe
	// Datei zweimal abgelegt ergibt dieselbe Datei unter demselben Namen;
	// zweimal weggenommen ergibt zweimal „keines".
	"schreiben.nutzer.bild":          `{"nutzerID": "1", "pfad": "/tmp/gesicht.png"}`,
	"schreiben.nutzer.bild.loeschen": `{"nutzerID": "1"}`,

	// **Das Abzeichen eines Profils** — Werkstück 9 (#1351). Dieselbe
	// Überlegung wie beim Profilbild darüber.
	"schreiben.profil.badge":          `{"profilID": "1", "pfad": "/tmp/wappen.png"}`,
	"schreiben.profil.badge.loeschen": `{"profilID": "1"}`,

	// **Die Einstellungen eines Menschen** — Werkstück 8 (#1351). Lesen ist
	// idempotent, weil es nichts ändert; Setzen, weil derselbe Wert zweimal
	// derselbe Zustand ist.
	"lesen.nutzereinstellungen":   `{"nutzerID": "1"}`,
	"lesen.personendaten":         `{"nutzerID": "1"}`,
	"schreiben.nutzereinstellung": `{"nutzerID": "1", "schluessel": "barrierefreiheit.kontrast", "wert": "hoch"}`,

	// **Die Zugriffstoken** — Werkstück 10 (#1351). Lesen ändert nichts;
	// Ändern setzt zweimal denselben Zustand; Löschen führt zweimal auf „gibt
	// es nicht". **Das Anlegen steht ohne Zusage da** und trotzdem hier: der
	// Wächter über die Vollständigkeit verlangt für jedes Werkzeug einen Fall,
	// und ohne ihn bliebe eine Lücke, die niemand sieht.
	"lesen.token": `{"profilID": "1"}`,
	"schreiben.token.anlegen": `{"profilID": "1", "bezeichnung": "Nachtlauf", "principal": "Rechner", ` +
		`"scope": "lesen"}`,
	"schreiben.token.aendern": `{"tokenID": "1", "bezeichnung": "Nachtlauf", "principal": "Rechner", ` +
		`"scope": "lesen", "fassung": "1"}`,
	"schreiben.token.loeschen": `{"tokenID": "1"}`,

	// **Werkstück O2** (#1426): die Ablaufwahlen lesen nur die Uhr. Der
	// Sammelwiderruf sagt keine Idempotenz zu — der zweite Ruf zieht null,
	// und das ist eine andere Antwort; der Fall steht hier, damit der Lauf
	// ihn zweimal ruft und den Unterschied sieht.
	"lesen.token.ablaufwahlen":        `{}`,
	"schreiben.token.widerrufen.alle": `{"profilID": "1"}`,

	// **Der Einstellungsaustausch** — Werkstück O3 von Phase 5.2 (#1427).
	// Ausfuhr und Vorschau lesen nur; die Einfuhr sagt keine Idempotenz zu,
	// weil die Marke nach der ersten Übernahme eine andere ist — der zweite
	// Ruf mit derselben Marke wird abgewiesen, und das ist die Zusage aus
	// 0044, nicht ihr Gegenteil.
	"lesen.einstellungen.ausfuhr":     `{}`,
	"lesen.einstellungen.vorschau":    `{"inhalt": "{}"}`,
	"schreiben.einstellungen.einfuhr": `{"inhalt": "{}", "revision": "abc123"}`,

	// **Die Profilvorgaben** — Werkstück 15 (#1351). Lesen ändert nichts;
	// Setzen setzt zweimal denselben Zustand. **Ein leeres `werte`**, weil
	// dieser Bestand kein Angebot hat und jeder Wert abgewiesen würde — das
	// Abräumen ist der Fall, der ohne Angebot durchgeht.
	"lesen.profilvorgaben": `{"profilID": "1"}`,
	"schreiben.profilvorgabe": `{"profilID": "1", "art": "deck", ` +
		`"werte": []}`,

	// **Die Freigaben** — Werkstück 11 (#1351). Lesen ändert nichts; eine
	// Rolle zweimal gesetzt ist derselbe Zustand; zweimal zurückgezogen führt
	// zweimal auf „gibt es nicht".
	"lesen.freigaben": `{"profilID": "1"}`,

	// **Die Wege nach draußen** (T8, #1473) — zweimal gefragt, zweimal
	// dieselben Adressen: sie entstehen auf diesem Rechner.
	"lesen.marktlinks": `{"printing": "bolt-m10", "name": "Lightning Bolt"}`,
	"lesen.marktkonto": `{"quelle": "cardtrader", "seite": 1, "menge": 20}`,

	// **Der Handel** steht in `idempotenzfaellehandel_test.go`.

	// **Die Decklisteneinfuhr** (T8, #1473). Das schreibende sagt keine
	// Idempotenz zu — der Stand aus der Vorschau gilt nach dem ersten Ruf
	// nicht mehr, und der zweite wird abgewiesen. Der Fall steht trotzdem
	// hier: die Liste ist vollständig zu halten.
	"lesen.deckliste":     `{"deckID": 1, "text": "4 Flammenball"}`,
	"schreiben.deckliste": `{"deckID": 1, "stand": "abdruck", "zeilen": []}`,
	"schreiben.freigabe": `{"profilID": "1", "nutzerID": "2", ` +
		`"rolle": "leser"}`,
	"schreiben.freigabe.loeschen": `{"profilID": "1", "nutzerID": "2"}`,

	// **Die Deckübernahme sagt keine Idempotenz zu** (#1351, Werkstück 12):
	// ein zweiter Ruf legt ein zweites Deck an; ihr Fall prüft nur, dass er
	// durchgeht. **Ihre lesende Hälfte** (UX-08.AK-04) nennt zweimal dieselben.
	"lesen.decks.uebernehmbar":   `{"quellProfil": "2"}`,
	"schreiben.deck.uebernehmen": `{"quellProfil": "2", "quellDeck": "7"}`,

	"lesen.mtgjson.spiele":    `{}`,
	"schreiben.mtgjson.lesen": `{"pfad": "/tmp/alle.json", "spiel": "magic"}`,
	"schreiben.mtgjson.schreiben": `{"setID": 1, "spiel": "magic", ` +
		`"ziel": "/tmp/raus.json"}`,

	"lesen.scryfall.vorschlag": `{"setID": 1, "karteID": 7, "name": "Prüfstein"}`,
	"schreiben.scryfall.uebernehmen": `{"setID": 1, "karteID": 7, "name": "Prüfstein", ` +
		`"felder": ["rule text"]}`,
	"schreiben.scryfall.bild": `{"setID": 1, "karteID": 7, "feld": "image", ` +
		`"name": "Prüfstein", "art": "normal"}`,
	// **Nicht idempotent, und das ist der Punkt** (#1032): ein zweiter Aufruf
	// legt eine **zweite** Karte an. Das Werkzeug meldet es nicht anders — es
	// gibt die Kennung zurück, und die ist beim zweiten Mal eine andere.
	"schreiben.scryfall.karte": `{"setID": 1, "name": "Prüfstein"}`,

	// **Nicht idempotent, aber ohne zweite Wirkung** (#1033): „eine Quelle,
	// ein Set" — der zweite Aufruf legt nichts an und meldet das vorhandene.
	"schreiben.scryfall.set": `{"code": "lea", "spiel": "magic"}`,

	// **Lesend und damit idempotent** (#1070): ein Ausdruck rechnet und gibt
	// zurück; zweimal derselbe Ausdruck ergibt dasselbe.
	"lesen.skript": `{"karteID": 7, "ausdruck": "1 + 1"}`,

	// **Lesend und damit idempotent** (#1164): eine Aufzählung dessen, was
	// die Umgebung kennt. Sie ändert sich, wenn jemand die Vorlage wechselt —
	// nicht dadurch, dass man zweimal fragt.
	"lesen.skriptnamen": `{"karteID": 7}`,

	// **Die zwei lesenden der Paketwerkstatt** (#1164) — Aufzählen und Lesen
	// ändern nichts.
	"lesen.paketdateien": `{"paket": "probe.mse-style"}`,
	"lesen.paketdatei":   `{"paket": "probe.mse-style", "datei": "style"}`,

	// **Zweimal derselbe Text, zweimal derselbe Befund.** Die Prüfung liest
	// nichts von der Platte und schreibt nichts — sie ist die reinste Form
	// eines lesenden Werkzeugs, und genau deshalb steht sie hier.
	// **Zweimal dasselbe Feld auf dieselben Karten** — das Ergebnis ist
	// beide Male dasselbe, und der zweite Lauf ändert nichts mehr (#1390).
	// Dass er trotzdem `schreiben.` heißt, ist richtig: er **kann** ändern.
	"schreiben.felder": `{"setID": 1, "karteIDs": [1, 2], "feld": "artist", "wert": "Gleich"}`,
	// **Die Vorschau dazu schreibt nichts** (U6, #1417) — zweimal dieselbe
	// Frage, zweimal dieselben Zeilen.
	"lesen.felder.vorschau": `{"setID": 1, "karteIDs": [1, 2], "feld": "artist", "wert": "Gleich"}`,
	// **Und der Verlauf** (U6): lesen ändert ihn nicht.
	"lesen.verlauf": `{"setID": 1}`,

	// **Der Symboleditor** (#1391) — acht Rechnungen und das Lesen. Sie
	// rechnen ohne Bestand: dieselben Teile geben dasselbe Ergebnis, und
	// zwar beliebig oft. Genau das ist Idempotenz in ihrer reinsten Form.
	"lesen.symbolpalette": `{"karteID": 1, "feld": "casting cost"}`,
	"lesen.symbol":        `{"setID": 1}`,

	// **Und das Ablegen.** Es heißt `schreiben.`, weil es ändern **kann** —
	// idempotent ist es trotzdem: dieselben Teile unter demselben Namen
	// geben dieselbe Datei.
	"schreiben.symbol":             `{"setID": 1, "teile": [], "unter": "probe.mse-symbol"}`,
	"lesen.symbol.grundform":       `{"art": "Rechteck"}`,
	"lesen.symbol.griff":           `{"teile": [{"art":"Form","name":"a","kombination":"overlap","punkte":[{"x":0.1,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.9,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"}]}], "teil": 0, "punkt": 0, "x": 0.5, "y": 0.5}`,
	"lesen.symbol.sperre":          `{"teile": [{"art":"Form","name":"a","kombination":"overlap","punkte":[{"x":0.1,"y":0.1,"griffDavorX":-0.1,"griffDavorY":0,"griffDanachX":0.2,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.9,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"}]}], "teil": 0, "punkt": 0, "sperre": "size"}`,
	"lesen.symbol.punkt":           `{"teile": [{"art":"Form","name":"a","kombination":"overlap","punkte":[{"x":0.1,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.9,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"}]}], "teil": 0, "punkt": 0, "weise": "einfuegen"}`,
	"lesen.symbol.gruppe":          `{"teile": [{"art":"Form","name":"a","kombination":"overlap","punkte":[{"x":0.1,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.9,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"}]},{"art":"Form","name":"b","kombination":"overlap","punkte":[{"x":0.1,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.9,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"}]}], "stellen": [0, 1], "weise": "gruppieren"}`,
	"lesen.symbol.reihenfolge":     `{"teile": [{"art":"Form","name":"a","kombination":"overlap","punkte":[{"x":0.1,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.9,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"}]}], "von": 0, "nach": 0}`,
	"lesen.symbol.symmetrie":       `{"teile": [{"art":"Form","name":"a","kombination":"overlap","punkte":[{"x":0.1,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.9,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"}]}], "stelle": 0, "weise": "anlegen"}`,
	"lesen.symbol.symmetriekopien": `{"teile": [{"art":"Symmetrie","name":"s","spiegelart":"rotation","kopien":2,"mitteX":0.5,"mitteY":0.5,"griffX":0.5,"griffY":0.1,"teile":[{"art":"Form","name":"a","kombination":"overlap","punkte":[{"x":0.1,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"},{"x":0.9,"y":0.9,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"art":"corner"}]}]}], "stelle": 0}`,
	"lesen.symbol.zuschnitt": `{"quelleBreite": 100, "quelleHoehe": 100, ` +
		`"zielBreite": 50, "zielHoehe": 50, "auswahlX": 0, "auswahlY": 0, ` +
		`"breite": 50, "hoehe": 50}`,

	"lesen.skriptbefund": `{"inhalt": "mse version: 0.3.8\nscript: f(1, 2\n"}`,

	// **Und das schreibende ist es nicht** — und das ist Absicht: der zweite
	// Aufruf mit demselben Stand wird abgewiesen, weil der Stand nach dem
	// ersten nicht mehr stimmt. Genau das ist der Schutz.
	"schreiben.paketdatei": `{"paket": "probe.mse-style", "datei": "style", ` +
		`"inhalt": "x", "stand": "s"}`,

	// **Lesend und damit idempotent** (#1074).
	"lesen.wortlisten": `{"karteID": 7, "feld": "type"}`,

	// **Lesend und damit idempotent** (#1062).
	"lesen.schluesselwoerter": `{"setID": 1}`,

	// **Beide lesend und damit idempotent** (#889, Befund 7 des Zweitblicks
	// über Phase W). Ziehen ändert nichts am Bestand — es rechnet aus einem
	// Samen eine Auswahl, und **derselbe Same gibt dieselbe Ziehung**. Das ist
	// hier nicht nur eine Randbemerkung: ohne den Samen wäre der Aufruf keine
	// Funktion und ließe sich nicht als idempotent führen.
	"lesen.packarten":   `{"setID": 1}`,
	"lesen.pack.ziehen": `{"setID": 1, "art": "alles", "same": "4242"}`,

	// **Die drei schreibenden sagen keine Idempotenz zu** (#1062) — kein
	// schreibendes tut das. Die Fälle stehen hier, damit der Wächter sie
	// zählt: anlegen ist ohnehin nicht idempotent (ein zweiter Aufruf legt ein
	// zweites an), setzen und löschen ändern die Datei.
	"schreiben.schluesselwort.anlegen": `{"setID": 1}`,

	// **Nicht idempotent, und das ist Absicht** (#891): zwei Aufrufe mit
	// demselben Namen legen **zwei** Sets an. Ein Set hat keinen Schlüssel
	// außer seiner Kennung, und zwei Sets desselben Namens sind erlaubt.
	"schreiben.set.neu": `{"name": "Neu", "spiel": "probe"}`,
	"schreiben.schluesselwort.setzen": `{"setID": 1, "wortID": 1, ` +
		`"feld": "muster", "wert": "Beben"}`,
	"schreiben.schluesselwort.loeschen": `{"setID": 1, "wortID": 1}`,

	// **Ausdrücklich nicht idempotent** (#1074): eine Wahl auf ein **Präfix**
	// schaltet um. Zweimal `Legendary ` gewählt heißt: hin und wieder zurück
	// — und das ist keine Eigenheit, sondern MSEs Verhalten
	// (`DropDownWordList::select`, `src/gui/value/text.cpp`, Zeilen 276–297,
	// Fassung f9d9356 vom 2021-07-06). Deshalb sagt das Werkzeug keine
	// Idempotenz zu, und der Fall steht hier nur, damit der Wächter ihn zählt.
	"schreiben.wortwahl": `{"setID": 1, "karteID": 7, "feld": "type", ` +
		`"bereich": 0, "weg": [0]}`,

	"lesen.scryfall.regeln":     `{"name": "Prüfstein"}`,
	"lesen.scryfall.manakosten": `{"kosten": "{2}{R}"}`,

	// **Beide ohne Argumente**: es gibt nur eine Sammeldatei, auf die dieses
	// Programm zurückfällt, und die Wahl wäre eine ohne Auswahl.
	"lesen.scryfall.sammeldaten":     `{}`,
	"schreiben.scryfall.sammeldaten": `{}`,

	// **Die Sammlung** (#1165). Sie hängt an keinem Set — die Kennung ist die
	// eines **Drucks**.
	"lesen.sammlung":     `{}`,
	"lesen.sammlungfuer": `{"printing": "probe-druck"}`,

	// **Anlegen sagt keine Idempotenz zu**, und der Fall steht hier nur,
	// damit der Wächter ihn zählt: dieselbe Karte in demselben Zustand gibt es
	// einmal, ein zweiter Ruf wird abgewiesen.
	"schreiben.exemplar": `{"printing": "neuer-druck", "menge": 1}`,

	"schreiben.exemplaraendern": `{"id": 1, "printing": "probe-druck", ` +
		`"menge": 3, "zustand": "nearMint", "sprache": "en"}`,

	// **Löschen sagt ebenfalls keine zu**: der zweite Ruf sagt, dass es die
	// Zeile nicht gibt. Die Wirkung ist dieselbe, die Antwort nicht.
	"schreiben.exemplarloeschen": `{"id": 1}`,

	// **Die Preise** (#1165). Drei lesen, eines holt — und das holende sagt
	// keine Idempotenz zu, weil ein zweiter Abruf nach Ablauf des
	// Zwischenspeichers eine **neue** Momentaufnahme ablegt.
	"lesen.preise":    `{"printing": "probe-druck"}`,
	"lesen.portfolio": `{}`,
	// **Eine Quelle, die es nicht gibt.** Der Fall darf hier nichts holen: der
	// Wächter ruft zweimal, und ein echter Bezug lüde zweimal eine große
	// Datei. `schreiben.preise.bulk` sagt keine Idempotenz zu — ein zweiter
	// Ruf ist ein **neuer** Abruf und legt neue Momentaufnahmen ab.
	"schreiben.preise.bulk": `{"quelle": "gibtesnicht", "umfang": "heute"}`,

	"lesen.preisverlauf": `{"printing": "probe-druck", "quelle": "scryfall", ` +
		`"markt": "paper", "art": "retail", "waehrung": "EUR", ` +
		`"finish": "nonfoil"}`,
	"lesen.marktangebote":    `{"quelle": "cardtrader", "printing": "probe-druck"}`,
	"lesen.anbieterstand":    `{"dienst": "gibtesnicht"}`,
	"lesen.ki.modelle":       `{}`,
	"schreiben.fremdverlauf": `{"quelle": "gibtesnicht", "printing": "probe-druck"}`,

	// **Ohne eingeschaltete Laufzeit scheitern beide Rufe gleich** (A7,
	// #1446) — die Vorgabe von `ki.laufzeit` ist `aus`, und das Werkzeug
	// sagt **keine** Idempotenz zu: mit Laufzeit gäbe dasselbe Modell zwei
	// verschiedene Antworten, und genau das ist der Zweck einer Generierung.
	"schreiben.ki.vorschlag": `{"karteId": 1, "alias": "gemma4:8b", "auftrag": "x"}`,
	"schreiben.ki.batch":     `{"karteIds": [1, 2], "alias": "gemma4:8b", "auftrag": "x"}`,
	"schreiben.ki.einschaetzung": `{"karteId": 1, "art": "regel", ` +
		`"alias": "gemma4:8b", "frage": "Wie wirkt sie?"}`,

	// **Eine Gruppenübernahme ist nicht idempotent, und beide Rufe scheitern
	// hier gleich** (A13.1b, #1452): den Auftrag `a1` gibt es nicht, und das
	// ist beim zweiten Ruf genauso wahr wie beim ersten.
	//
	// **Sie sagt auch mit vorhandenem Auftrag keine Idempotenz zu** — im
	// Gegenteil: die Kennung ist **konsumierbar**, und der zweite Ruf muss
	// scheitern. Zweimal übernehmen hieße, denselben Vorschlag auf einen
	// Stand zu legen, der ihn schon trägt.
	"schreiben.ki.gruppe": `{"wahl": [{"kennung": "a1", "felder": ["name"]}]}`,

	// **Eine Erkennung ohne Bilddatei scheitert zweimal gleich** (A13.3,
	// #1452). **Mit** einer sagt sie keine Idempotenz zu: dasselbe Foto
	// zweimal gelesen gibt zwei Antworten — das ist der Zweck einer
	// Generierung.
	"schreiben.ki.erkennung": `{"pfad": "/gibtesnicht.png", ` +
		`"anbieter": "anthropic", "alias": "claude-probe"}`,

	// **Die Messung ist von sich aus idempotent** (A9, #1448): eine zweite
	// ersetzt die erste, und `vergessen` nimmt jede zurück.
	"schreiben.ki.modell.pruefen":   `{"anbieter": "openai", "alias": "gpt-probe"}`,
	"schreiben.ki.modell.vergessen": `{"anbieter": "openai", "alias": "gpt-probe"}`,

	// **Der Sitzungsschalter ist ein Stand** (AI-05, #2042): zweimal aus ist aus.
	"schreiben.ki.geraetemodell.sitzung": `{"an": false}`,
	"schreiben.preiseholen":              `{"printing": "probe-druck"}`,
}))
