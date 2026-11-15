// **Zwei Zusagen, die für sich stehen** —
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md).
//
// **Eine eigene Datei**, seit `werkzeugumfang_test.go` mit der Restparität
// über die Zeilengrenze aus §2 lief
// (#1291).
// Der Schnitt folgt der Frage: dort steht die **Prüfung** samt der Liste aller
// Zusagen, hier die **Begründung** der beiden, die am meisten Text brauchen.
package mcp_test

// ohneUeberschreiben sind die, die schreiben, ohne zu überschreiben (0044).
//
// Sie stehen nicht bei „eine Karte je Aufruf", weil sie eine andere Zusage
// geben: nicht „eine Karte", sondern „nichts, was schon da ist".
func ohneUeberschreiben() map[string]bool {
	return map[string]bool{
		"schreiben.set.einfahren":  true,
		"schreiben.set.sichern":    true,
		"schreiben.karte.zeichnen": true,

		// **Und der Druckbogen** (#1291). Er legt eine PDF-Datei an einem
		// freien Ort an und rührt den Bestand nicht an — dieselbe Zusage wie
		// beim Zeichnen, nur über ein ganzes Set.
		"schreiben.set.bogen": true,

		// **Und die Weitergabe als Installer** (FMT-02.AK-04, #1691): sie
		// legt eine Datei an einem freien Ort an — die Naht reserviert sie
		// vorher —, und das Einlesen **übergeht**, was schon eingebunden ist.
		"schreiben.paket.weitergeben": true,
		"schreiben.paket.einlesen":    true,

		// **Und das Einbinden und Übernehmen von Paketen** (#1291). Beide
		// legen Verzeichnisse an einem freien Ort an: ein Paket, das schon
		// eingebunden ist, wird abgewiesen (Einbinden) oder **übergangen**
		// (Übernahme) — überschrieben wird keines.
		// **Und die beiden Nutzerwerkzeuge**
		// (#1351).
		// Dieselbe Zusage: sie legen einen **neuen** Nutzer beziehungsweise
		// ein **neues** Profil an; ein vorhandener Name ist ein Konflikt und
		// wird nicht überschrieben.
		"schreiben.nutzer.neu": true,
		"schreiben.profil.neu": true,

		// **Und die Decklisteneinfuhr** (T8, #1473). Dieselbe Zusage: sie
		// **legt** Plätze in ein Deck und nimmt keinen heraus — derselbe
		// Bezug im selben Bereich wird eine höhere Menge, kein zweiter Platz
		// (`storage.Deckablage.PlatzSetzen`). Und sie ist **umkehrbar**, der
		// zweite Maßstab aus 0044: was hineinkam, steht mit seiner
		// Platzkennung da und geht über `schreiben.deck.platz.nehmen` wieder
		// heraus.
		//
		// **Besitz entsteht dabei nicht.** Ein Deck verweist auf Karten; die
		// Sammlung bleibt unberührt, und ein Lauf über die Flächen hält das
		// fest (`internal/flaechen/deckeinfuhr_test.go`).
		"schreiben.deckliste": true,

		// **Und das anonyme Konto** — Werkstück 6 desselben Issues. Dieselbe
		// Zusage: es legt einen **neuen** Nutzer an und rührt keinen
		// vorhandenen an. Seine Kennung ist zufällig, und **deshalb** kann es
		// keinen treffen, den es schon gibt.
		"schreiben.konto.anonym": true,

		// **Und das Herausnehmen eines zweiten Faktors** — Werkstück 7
		// desselben Issues. Der zweite Maßstab aus 0044, **umkehrbar**: der
		// Nutzer bleibt, seine Daten bleiben. **Was nicht zurückkommt, ist
		// das Geheimnis** — wer den Faktor wieder will, richtet ihn neu ein
		// und scannt ein neues Muster. Das ist keine Ausnahme von der
		// Umkehrbarkeit, sondern ihr Preis: ein Geheimnis, das ein Löschen
		// überlebte, wäre keines.
		//
		// **Und die Richtung ist die sichere**: das Werkzeug macht ein Konto
		// **leichter** zugänglich, nicht schwerer. Es sperrt niemanden aus.
		"schreiben.faktor.loeschen": true,

		// **Und der Profilwechsel** — Werkstück 3 desselben Issues, mit einer
		// **anderen** Begründung: er legt nichts an. Er stellt einen Zeiger
		// um, und was er verdeckt, ist nach dem Zurückwechseln wieder da. Von
		// den beiden Maßstäben aus 0044 ist das der zweite — **umkehrbar**,
		// und zwar vollständig und mit demselben Werkzeug.
		"schreiben.profil.wechseln": true,

		// **Und der Nutzerwechsel** — Werkstück 17 von Phase 4.12
		// (#1400),
		// aus demselben Grund wie der Profilwechsel darüber: er legt nichts an
		// und stellt einen Zeiger um. **Umkehrbar mit demselben Werkzeug** —
		// wer zurückwill, wechselt zurück.
		//
		// **Der zweite Faktor macht ihn nicht unumkehrbar**, sondern
		// **schwerer zu tun**: er ist der Ausweis des Ziels, und ohne ihn
		// findet der Wechsel gar nicht statt.
		"schreiben.nutzer.wechseln": true,

		// **Und das Lösen eines Anbieterkontos** — Werkstück 4 desselben
		// Issues. Auch dies ist der zweite Maßstab aus 0044, **umkehrbar**:
		// der Nutzer bleibt, seine Daten bleiben, nur der Weg über diesen
		// Anbieter fällt weg. Wer sich danach wieder anmeldet, ist derselbe
		// Mensch — dieselbe fremde Kennung findet denselben Nutzer.
		"schreiben.konto.loesen": true,

		// **Und die vier Pflegewerkzeuge** — Werkstück 5 desselben Issues.
		//
		// **Ändern** ist der zweite Maßstab aus 0044: ein Name, der gesetzt
		// wird, ersetzt genau den einen, der dastand — und der Rufer hat ihn
		// gerade gelesen. **Löschen** ist der Grenzfall, und deshalb steht der
		// **Name** als Bestätigung davor: wer sich vertippt, trifft niemanden.
		// Zwei Zusagen halten den Bestand offen — der letzte Nutzer bleibt,
		// und das letzte Profil eines Nutzers bleibt.
		"schreiben.nutzer.aendern":  true,
		"schreiben.profil.aendern":  true,
		"schreiben.nutzer.loeschen": true,
		"schreiben.profil.loeschen": true,

		// **Und die beiden am Profilbild** — Werkstück 8 desselben Issues,
		// mit derselben Begründung wie das Ändern: der zweite Maßstab aus
		// 0044, **umkehrbar**. Ein Bild ersetzt genau das eine, das dastand;
		// wer es wegnimmt, kann dasselbe wieder setzen. **Was nicht
		// zurückkommt, ist die alte Datei** — und das steht hier, weil es
		// der Unterschied zu einem Feldwert ist: den hält der Verlauf, eine
		// überschriebene Bilddatei niemand. Deshalb ändern beide Werkzeuge
		// **einen** Nutzer, den ihre Argumente nennen, und keinen zweiten.
		"schreiben.nutzer.bild":          true,
		"schreiben.nutzer.bild.loeschen": true,

		// **Und die beiden am Abzeichen eines Profils** — Werkstück 9
		// (#1351), Punkt 13. Dieselbe Zusage aus demselben Grund: sie ändern
		// **ein** Profil, das ihre Argumente nennen, und was sie
		// überschreiben, lässt sich mit demselben Werkzeug wieder setzen.
		// **Was nicht zurückkommt, ist die alte Datei** — wie beim
		// Profilbild.
		"schreiben.profil.badge":          true,
		"schreiben.profil.badge.loeschen": true,

		// **Und die Einstellung eines Menschen** — Punkte 9 und 11 desselben
		// Werkstücks. Der zweite Maßstab aus 0044, **umkehrbar**: sie setzt
		// **einen** Schlüssel, den ihre Argumente nennen, auf einen Wert, den
		// `lesen.nutzereinstellungen` genannt hat — und derselbe Weg stellt
		// ihn zurück. Was sie **nicht** kann, ist eine Einstellung des
		// Profils ändern; dafür gibt es `schreiben.einstellung`.
		"schreiben.nutzereinstellung": true,

		// **Und die drei an den Zugriffstoken** — Werkstück 10 desselben
		// Issues, Punkt 13. **Und sie stehen hier mit drei verschiedenen
		// Begründungen**, nicht mit einer.
		//
		// **Anlegen** ist der erste Maßstab aus 0044 — schreiben, ohne zu
		// überschreiben: es entsteht ein **neues** Token, und kein
		// bestehendes wird angefasst. Umkehrbar ist es über die Kennung, die
		// es zurückgibt.
		//
		// **Ändern** ist der zweite — umkehrbar: Bezeichnung, Erlaubnis und
		// Ablauf ersetzen genau das, was der Rufer gerade gelesen hat, und
		// derselbe Weg stellt es zurück. **Das Geheimnis rührt es nicht an**;
		// niemand verliert seinen Zugang, weil jemand umbenannt hat.
		//
		// **Löschen ist der Grenzfall, und es ist NICHT umkehrbar.** Was
		// verlorengeht, ist das Geheimnis — es gibt es nirgends mehr, und ein
		// neues Token ist ein anderes. Erlaubt ist es trotzdem, und zwar aus
		// dem Grund, um dessentwillen es diese Werkzeuge gibt: **ein Token
		// zurückziehen zu können, ist der Sinn der Sache**. Wer es nicht
		// abschalten kann, hat kein verwaltbares Geheimnis, sondern eines,
		// das für immer gilt. Der Schaden ist begrenzt und benannt — was mit
		// dem Token angelegt wurde, bleibt im Bestand.
		"schreiben.token.anlegen":         true,
		"schreiben.token.aendern":         true,
		"schreiben.token.loeschen":        true,
		"schreiben.token.widerrufen.alle": true,

		// **Und die Einfuhr von Einstellungen** — Werkstück O3 von Phase 5.2
		// (#1427). Sie verlangt die Marke aus der Vorschau und lässt alles
		// stehen, wenn sie nicht mehr stimmt: das ist das Schreiben ohne
		// Überschreiben aus 0044, an Einstellungen statt an Karten.
		"schreiben.einstellungen.einfuhr": true,

		// **Und die Profilvorgaben** — Werkstück 15 desselben Issues, Punkte
		// 4 bis 7. Der zweite Maßstab aus 0044, **umkehrbar**: sie setzen die
		// Werte **einer** Art, die ihre Argumente nennen, auf Werte, die
		// `lesen.profilvorgaben` genannt hat — und derselbe Weg stellt sie
		// zurück.
		//
		// **Was sie nicht anfassen, ist das, worauf sie zeigen.** Ein
		// Präferenzdeck bleibt ein Deck; die Vorgabe sagt nur, dass dieses
		// Profil es bevorzugt.
		"schreiben.profilvorgabe": true,

		// **Und die beiden an den Freigaben** — Werkstück 11 desselben
		// Issues. Der zweite Maßstab aus 0044, **umkehrbar**: sie setzen die
		// Rolle **eines** Nutzers an **einem** Profil, die ihre Argumente
		// nennen, und derselbe Weg stellt sie zurück.
		//
		// **Was sie nicht anfassen, sind die Daten dahinter.** Wer eine
		// Freigabe verliert, verliert den Zugang — nicht, was er darin
		// angelegt hat.
		//
		// **Und ein Modell kann nichts vergeben, was der Mensch dahinter
		// nicht hat**: nur ein `verwalter` kommt durch, und wer das ist, steht
		// fest, bevor ein Werkzeug gerufen wird.
		"schreiben.freigabe":          true,
		"schreiben.freigabe.loeschen": true,

		// **Und die Deckübernahme** — Werkstück 12 desselben Issues, nach
		// demselben Maßstab: sie legt ein **neues** Deck an und überschreibt
		// nichts. Weder im eigenen Profil (dort entsteht etwas) noch im
		// fremden (dort wird nur gelesen).
		//
		// **Umkehrbar über `schreiben.deck.loeschen`** — und zwar
		// vollständig: das übernommene Deck ist neu, es hat nichts verdrängt.
		//
		// **Und ein Modell kann nichts holen, wohin der Mensch dahinter nicht
		// darf**: es braucht `mitarbeiter` an dem fremden Profil, und wer das
		// hat, steht fest, bevor ein Werkzeug gerufen wird.
		"schreiben.deck.uebernehmen": true,

		"schreiben.paket.installieren": true,
		"schreiben.paket.uebernehmen":  true,

		// **Und das Anlegen eines neuen Pakets**
		// (#1350).
		// Dieselbe Zusage aus demselben Grund: es legt ein Verzeichnis an
		// einem **freien** Ort an und endet in derselben Installation wie das
		// Einbinden — ein Paket, das schon da ist, wird abgewiesen und nicht
		// überschrieben.
		"schreiben.paket.anlegen": true,

		// **Und die eigene Kopie** (FMT-02, #1691): sie bekommt einen freien
		// Namen oder wird abgewiesen, und das Original bleibt, wie es war.
		"schreiben.paket.kopieren": true,

		// **Und MTG JSON in beide Richtungen** (#1291). Das Lesen legt Sets
		// an und überschreibt keines — „eine Quelle, ein Set" macht den
		// zweiten Ruf zu einem Lesevorgang, wie beim Einfahren. Das
		// Schreiben legt eine Datei an einem freien Ort an, wie das
		// Sichern.
		"schreiben.mtgjson.lesen":     true,
		"schreiben.mtgjson.schreiben": true,

		// **Und das Anlegen aus einem Scryfall-Namen** (#1032). Es legt eine
		// **neue** Karte an und füllt ihre leeren Felder; überschrieben wird
		// nichts, weil vorher nichts dastand. Umkehrbar ist es über die
		// Kennung, die es zurückgibt — `schreiben.karte.loeschen` nimmt sie
		// entgegen (0010, 0044).
		"schreiben.scryfall.karte": true,

		// **Und ein ganzer Satz** (#1033). „Eine Quelle, ein Set" (#886): ein
		// zweiter Aufruf mit demselben Satzcode legt kein zweites an, sondern
		// meldet das vorhandene — dieselbe Zusage wie `schreiben.set.einfahren`.
		"schreiben.scryfall.set": true,
	}
}

// nimmtEinPaketWeg ist die Zusage über `schreiben.paket.entfernen` (#1291).
//
// # Warum das eine eigene Zusage ist
//
// Aus demselben Grund wie `nimmtEinSetWeg`: es ist **unumkehrbar**. Nach ihr
// liegt das Paketverzeichnis nicht mehr da, und es gibt keinen Papierkorb.
//
// **Der Unterschied zum Set**: es trifft **kein** Set, aber womöglich viele —
// ein entfernter Stil fehlt jedem Set, das ihn benutzt.
//
// **Was diese Zusage sagt**: der Aufruf nimmt **ein** Paket und verlangt dafür
// den **Pfad aus `lesen.pakete`** als Bestätigung; stimmt er nicht, geschieht
// nichts. Er entfernt nur aus dem **verwalteten** Verzeichnis — die
// mitgelieferte Saat weist der Dienst ab
// (#58).
//
// **Was sie nicht sagt**: dass die Sets weiter zeichnen. Wer den Stil
// entfernt, den ein Set benutzt, bekommt danach den Behelfsplan.
//
// **Und wo die Bestätigung geprüft wird**: an der Wurzel (`cmd/mse-mcp`), nicht
// in diesem Paket — sie ist eine Aussage über die Platte, und `internal/mcp`
// ist ein Blatt (0044).
func nimmtEinPaketWeg() map[string]bool {
	return map[string]bool{"schreiben.paket.entfernen": true}
}
