package mcp_test

// **Die letzten fünf Zusagen** — abgetrennt am 2026-09-13, weil
// `zusagenliste_test.go` mit der Gruppenübernahme über die Zeilengrenze aus
// §2 ging (A13.1b, #1452).
//
// # Der Schnitt folgt der Frage
//
// Dort stehen die Zusagen über **Karte, Set und Bestand** — die Gruppen, die
// sich am Aufbau des Bestands entlang unterscheiden. Hier stehen die fünf,
// die **danebengreifen**: Preise eines fremden Dienstes, eine Gruppe über
// mehrere Karten, ein Symbol als Datei, das Wegräumen alter Sicherungen und
// der Tausch einer Paketfassung.
//
// **Die Tabelle bleibt eine.** `zusagen()` ruft diese Funktion auf; wer eine
// Zusage sucht, findet sie über die Aufzählung dort.

// weitereZusagen gibt die fünf, die nicht am Aufbau des Bestands hängen.
func weitereZusagen() map[string]bool {
	// **Und eines holt Preise** (#1165).
	//
	// # Warum das eine elfte Zusage ist
	//
	// Es ist das zweite Werkzeug nach `schreiben.paket.beziehen`, das ins
	// **Netz** geht — und das einzige, das eine **Momentaufnahme** ablegt.
	// Beides ist ein anderer Umfang als „eine Zeile ändern": ein Aufruf legt
	// bis zu sieben Zeilen an, eine je Währung und Ausführung.
	//
	// **Was diese Zusage sagt**: der Aufruf fragt **einen** Druck ab, legt die
	// Preise dieses Drucks als Momentaufnahmen ab und ändert **nichts
	// Vorhandenes** — ein Stand mit demselben Zeitpunkt wird übergangen, ein
	// älterer bleibt stehen. Ein Verlauf entsteht dadurch und wird nicht
	// überschrieben.
	//
	// **Was sie nicht sagt**: dass der Preis stimmt. Die Quelle sagt
	// ausdrücklich das Gegenteil, und der Vorbehalt steht in der Erklärung.
	//
	// **Und der Bulkbezug gehört in dieselbe Gruppe** (A3, #1442): er legt
	// Momentaufnahmen an — viele auf einmal —, und er **überschreibt nichts**.
	// Was er anfasst, ist die Preistafel und keine Karte.
	//
	// **Und der Fremdverlauf ebenso** (A5, #1444) — mit **einer** Zugabe: er
	// merkt sich beim ersten Mal die Zuordnung des Drucks zur Karte des
	// Anbieters. Das ist eine **Kennung**, kein Feldwert, und sie ist
	// umkehrbar: wer sie entfernt, sucht sie beim nächsten Abruf wieder.
	holtPreise := map[string]bool{
		"schreiben.preiseholen":  true,
		"schreiben.preise.bulk":  true,
		"schreiben.fremdverlauf": true,
	}

	// **Und das eine, das mehrere Karten auf einmal ändert** (#1390,
	// [0064](../../docs/entscheidungen/0064-ein-feld-auf-vielen-karten-darf-ueber-mcp.md)).
	//
	// # Warum das eine zwölfte Zusage ist
	//
	// „Eine Karte je Aufruf" trifft ausdrücklich nicht zu — das ist der Satz,
	// den [0040](../../docs/entscheidungen/0040-der-lebenszyklus-einer-karte-gehoert-dazu.md)
	// von 0010 stehen ließ und den 0064 fortschreibt. „Nichts, was schon da
	// ist" trifft auch nicht zu: der vorige Wert wird überschrieben.
	//
	// **Was diese Zusage sagt**: der Aufruf ändert **ein** Feld an den
	// Karten, die er **nennt**, in **einem** Set und in **einer**
	// Transaktion — und der Speicher legt ihn als **eine Gruppe im Verlauf**
	// ab.
	//
	// **Was sie nicht sagt, seit dem 2026-09-03**: dass ein einziges
	// `schreiben.zuruecknehmen` ihn ganz zurückholt. Hier stand das, und es
	// war zu stark — Befund 5 des Nachlaufs über die Phasen 4.0 bis 4.9.
	// Vermerk 0064 sagt es inzwischen richtig: die Rücknahme bekommt nur eine
	// **Setkennung** und sucht den jüngsten geltenden Schritt
	// (`internal/storage/verlaufbewegen.go`, `neuesterSchritt`). Schreibt ein
	// zweiter Prozess dazwischen — und [0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)
	// sieht mehrere ausdrücklich vor —, trifft sie dessen Schritt.
	//
	// **Der Name sagt es deshalb auch nicht mehr.** Er hieß
	// `aendertVieleKartenInEinerGruppe`; er heißt jetzt nach dem, was gilt.
	//
	// **Und sie sagt nicht**, dass die vorigen Werte in der Antwort stehen.
	// Sie stehen nicht darin — sie können je Karte andere sein, und eine
	// Liste davon wäre nur über Einzelaufrufe wieder einzuspielen.
	//
	// **Und die Gruppenübernahme gehört dazu** (A13.1b, #1452,
	// [0109](../../docs/entscheidungen/0109-eine-gruppe-darf-ueber-setgrenzen-gehen.md)).
	// Derselbe Maßstab: der Umfang steht in den Argumenten — jede Kennung
	// ihrer `wahl` ist ein Auftrag, und der trägt seine Zielkarte samt
	// eingefrorener Revision seit dem Anlegen.
	//
	// **Zwei Unterschiede zu `schreiben.felder`**: sie schreibt **mehrere**
	// Felder je Karte, und ihre Karten dürfen aus **mehreren** Sets kommen —
	// das eine ist kein neuer Maßstab, das andere entscheidet 0109. Der
	// Grund: die Setkennung kommt hier nicht vom Aufrufer, sondern wird an
	// der Karte gelesen, und was nicht angegeben wird, kann nicht falsch
	// sein.
	//
	// **Und was sie nicht sagt**: dass eine Rücknahme sie als Gruppe trifft.
	// Vorwärts ist sie atomar, rückwärts nicht — eine Gruppe über drei Sets
	// hat drei Verlaufsschritte. Das steht in 0109 als Preis.
	aendertVieleKartenInEinerGruppe := map[string]bool{
		"schreiben.felder":    true,
		"schreiben.ki.gruppe": true,
	}

	// **Und das eine, das ein Symbol neben das Set legt** (#1391).
	//
	// # Warum das eine dreizehnte Zusage ist
	//
	// „Eine Karte je Aufruf" trifft nicht zu — es sind null: ein Symbol
	// hängt am **Set** und an keiner Karte. „Nichts, was schon da ist"
	// trifft auch nicht zu: ein Symbol desselben Namens wird ersetzt.
	//
	// **Was diese Zusage sagt**: der Aufruf schreibt **eine** Datei neben
	// **ein** Set und ändert keine Karte. Der vorige Stand ist über
	// `lesen.symbol` zu holen, bevor geschrieben wird — die Erklärung sagt
	// das ausdrücklich.
	//
	// **Was sie nicht sagt**: dass ein Verlauf ihn hält. Ein Symbol liegt
	// auf der Platte, nicht im Bestand; `schreiben.zuruecknehmen` erreicht
	// es nicht.
	legtEinSymbolAb := map[string]bool{
		"schreiben.symbol": true,
	}

	// **Und das eine, das alte Sicherungen wegräumt** — Werkstück 4 von
	// Phase 4.13
	// (#1352).
	//
	// # Warum das eine vierzehnte Zusage ist
	//
	// „Eine Karte je Aufruf" trifft nicht zu — es sind null. „Ohne
	// Bestandswirkung" auch nicht, aber anders als bei allen übrigen: es
	// ändert **nichts im Bestand** und löscht **Dateien daneben**.
	//
	// **Was diese Zusage sagt**: der Aufruf löscht Sicherungen, die vor
	// Migrationen entstanden sind — **bis auf die jüngste**. Sie ist der
	// Rückweg, und wer sie mitnähme, nähme genau die Sicherung, für die es
	// sie gibt. Kein Set, keine Karte, keine Einstellung wird angefasst.
	//
	// **Und sie sagt nicht**, dass es umkehrbar ist. Eine gelöschte Datei ist
	// fort; die Erklärung des Werkzeugs sagt „LÖSCHT DATEIEN", und
	// `lesen.platz` sagt vorher, wie viele es sind.
	raeumtAlteSicherungen := map[string]bool{
		"schreiben.platz.freigeben": true,
	}

	// **Und zwei tauschen die Fassung eines Pakets** — DOM-01
	// (#1696).
	//
	// # Warum das eine fünfzehnte Zusage ist
	//
	// „Ohne Überschreiben" trifft nicht zu: das Ersetzen legt eine neue
	// Fassung **über** die vorhandene. Aber es ist umkehrbar wie
	// `schreiben.paketdatei` — die bisherige Fassung wird als **vorige**
	// aufgehoben, und `schreiben.paket.zurueck` tauscht beide.
	//
	// **Was diese Zusage sagt**: der Aufruf tauscht **ein** Paket und hebt
	// den vorigen Stand auf; ein **lokal geändertes** Paket wird nur mit
	// `bestaetigt` ersetzt, und auch dann liegt die Änderung als vorige
	// Fassung da. Kein Set und keine Karte wird angefasst.
	//
	// **Was sie nicht sagt**: dass die Sets danach gleich zeichnen. Was eine
	// andere Fassung an ihnen ändert, nennt `lesen.paket.plan` vorher.
	tauschtEinePaketfassung := map[string]bool{
		"schreiben.paket.ersetzen": true,
		"schreiben.paket.zurueck":  true,
	}

	return zusammengefasst(
		holtPreise, aendertVieleKartenInEinerGruppe, legtEinSymbolAb,
		raeumtAlteSicherungen, tauschtEinePaketfassung, fuehrtDieSymbolbibliothek(),
	)
}

// fuehrtDieSymbolbibliothek ist die Zusage über die sechs schreibenden
// Werkzeuge der Symbolbibliothek — DOM-02
// (#1697).
//
// # Mit vier Begründungen, nicht mit einer
//
// **Aufnehmen, Duplizieren und Einlesen** sind der erste Maßstab aus 0044 —
// schreiben, ohne zu überschreiben: es entsteht ein **neues** Symbol, und
// kein bestehendes wird angefasst.
//
// **Umbenennen** ist der zweite — umkehrbar: derselbe Weg stellt den Namen
// zurück, den `lesen.symbolbibliothek` genannt hat.
//
// **Einsetzen** legt eine **neue** Datei neben das Set und ändert **einen**
// Wert eines Sets, das in den Argumenten steht. Umkehrbar über
// `schreiben.setangabe` mit dem Wert, den `lesen.setangaben` vorher nannte —
// die vorige Datei bleibt liegen.
//
// **Sichern und Zurücknehmen** sind umkehrbar: die bisherige Fassung wird die
// vorige, und derselbe Weg zurück tauscht beide. **Nachziehen** schreibt die
// Symboldatei der Sets neu, die die vorige Fassung tragen — umkehrbar über
// Zurücknehmen und ein zweites Nachziehen.
//
// **Löschen ist der Grenzfall, und es ist NICHT umkehrbar.** Erlaubt ist es,
// weil der Schaden begrenzt und benannt ist: ein Symbol, das Sets tragen,
// geht nur mit `bestaetigt`, die Absage nennt sie, und **sie behalten ihre
// Kopie** — verloren geht der Eintrag der Bibliothek, nicht ein Symbol eines
// Sets. Dieselbe Abwägung wie beim Token.
func fuehrtDieSymbolbibliothek() map[string]bool {
	return map[string]bool{
		"schreiben.symbolbibliothek.aufnehmen":   true,
		"schreiben.symbolbibliothek.duplizieren": true,
		"schreiben.symbolbibliothek.einlesen":    true,
		"schreiben.symbolbibliothek.umbenennen":  true,
		"schreiben.symbolbibliothek.einsetzen":   true,
		"schreiben.symbolbibliothek.loeschen":    true,

		"schreiben.symbolbibliothek.sichern":       true,
		"schreiben.symbolbibliothek.zuruecknehmen": true,
		"schreiben.symbolbibliothek.nachziehen":    true,
	}
}
