package rest

// **Die Wege rund um Peers** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine eigene Datei**, weil `wegetafel.go` mit ihnen über die Zeilengrenze
// aus §2 ging. Der Schnitt folgt der Frage: dort stehen die Wege in den
// Bestand, hier die zu **anderen Rechnern**.

import "net/http"

// peerwege hängt die eigene Identität und die Kontakte an die Tafel.
func (d *Dienst) peerwege(weg *http.ServeMux) {
	// **Die eigene Peeridentität** (N1, #1454) — lesen, anlegen, ablösen.
	weg.HandleFunc("GET /peer", d.peerZeigen)
	weg.HandleFunc("POST /peer", d.peerAnlegen)
	weg.HandleFunc("POST /peer/rotation", d.peerRotieren)

	// **Und die bekannten Gegenüber** (N1, #1454). Die Sperre ist eine
	// Ressource und kein Befehl: `PUT` setzt sie, `DELETE` nimmt sie.
	weg.HandleFunc("GET /peer/kontakte", d.kontakteZeigen)
	weg.HandleFunc("POST /peer/kontakte", d.kontaktAufnehmen)
	weg.HandleFunc("POST /peer/kontakte/{kennung}/bestaetigung",
		d.kontaktBestaetigen)
	weg.HandleFunc("PUT /peer/kontakte/{kennung}/sperre", d.kontaktSperren)
	weg.HandleFunc("DELETE /peer/kontakte/{kennung}/sperre",
		d.kontaktEntsperren)
	weg.HandleFunc("PUT /peer/kontakte/{kennung}/zwecke/{zweck}",
		d.zweckErlauben)
	weg.HandleFunc("DELETE /peer/kontakte/{kennung}/zwecke/{zweck}",
		d.zweckEntziehen)

	// **Einladen, prüfen, annehmen** (NET-01, #1705) — der Code steht im
	// Rumpf und nie im Pfad.
	weg.HandleFunc("POST /peer/einladungen", d.einladungErstellen)
	weg.HandleFunc("POST /peer/einladungen/pruefung", d.einladungPruefen)
	weg.HandleFunc("POST /peer/einladungen/annahme", d.einladungAnnehmen)

	// **Und wo eines erreichbar ist** (N2, #1455) — ein `PUT`, weil es
	// genau eine gibt; eine leere löscht sie.
	weg.HandleFunc("PUT /peer/kontakte/{kennung}/adresse", d.kontaktAdresse)

	// **Und was ein Gegenüber sehen darf** (T3, #1468). Eine Freigabe ist
	// eine **Ressource**: `POST` erteilt, `DELETE` nimmt zurück, `GET` sagt,
	// was steht. Die Vorschau daneben ist eine Frage an dieselbe Sache — mit
	// dem Zweck in der Abfrage, denn dieselbe Sammlung sieht für „teilen"
	// anders aus als für „handel".
	weg.HandleFunc("GET /peer/kontakte/{kennung}/freigaben", d.freigabenZeigen)
	weg.HandleFunc("POST /peer/kontakte/{kennung}/freigaben", d.freigabeErteilen)
	weg.HandleFunc("DELETE /peer/kontakte/{kennung}/freigaben/{zweck}/{umfang}",
		d.freigabeWiderrufen)
	weg.HandleFunc("GET /peer/kontakte/{kennung}/sicht", d.freigabesichtZeigen)

	// **`POST`, weil der fremde Stand mitkommt** — und geschrieben wird
	// dabei nichts (T4, #1469).
	weg.HandleFunc("POST /peer/kontakte/{kennung}/abgleich", d.tauschabgleichZeigen)

	// **`GET`, weil nichts entsteht** — der Stand ist eine Auskunft über
	// das, was schon dasteht (T5, #1470).
	weg.HandleFunc("GET /peer/kontakte/{kennung}/post", d.poststandZeigen)
	weg.HandleFunc("GET /peer/kontakte/{kennung}/handel", d.handelsvorgaengeZeigen)

	// **Der Handel ist die Ressource, der Schritt sein Zustand** (T6,
	// #1471). `POST` auf denselben Weg trägt im Rumpf, **welcher** Schritt
	// gemeint ist; ein Pfad je Schritt machte aus einer Zustandsmaschine
	// drei Endpunkte, die nichts voneinander wissen.
	weg.HandleFunc("POST /handel/{vorgang}/schritt", d.handelsschrittGehen)

	// **Und die Schritte davor gehen denselben Weg** (T1, #1466,
	// angeschlossen mit #1637). Ein Vorgang entsteht mit `POST /handel`, ein
	// Vorschlag ist eine **neue Revision** und deshalb ein eigener Weg, und
	// der Stand ist ein `GET` — es entsteht nichts.
	weg.HandleFunc("POST /handel", d.handelEroeffnen)
	weg.HandleFunc("GET /handel/{vorgang}", d.handelsstandZeigen)
	weg.HandleFunc("POST /handel/{vorgang}/vorschlag", d.handelVorschlagen)

	// **Die Liste und die Entwürfe** (TRADE-01, #1715): ein Entwurf ist eine
	// Ressource des Vorgangs, das Archiv ein Zustand, den `PUT` setzt.
	weg.HandleFunc("GET /handel", d.handelsuebersichtZeigen)
	weg.HandleFunc("PUT /handel/{vorgang}/entwurf", d.handelEntwurfAendern)
	weg.HandleFunc("DELETE /handel/{vorgang}/entwurf", d.handelEntwurfLoeschen)
	weg.HandleFunc("POST /handel/{vorgang}/kopie", d.handelKopieren)
	weg.HandleFunc("PUT /handel/{vorgang}/archiv", d.handelArchivieren(true))
	weg.HandleFunc("DELETE /handel/{vorgang}/archiv", d.handelArchivieren(false))

	// **Der Ausgang ist eine Ressource des Vorgangs** (T5, #1470,
	// angeschlossen mit #1637): `POST` legt hinein, und die Bestätigung
	// räumt **eine benannte Nachricht** weg — deshalb hängt sie an ihr und
	// nicht am Vorgang.
	weg.HandleFunc("POST /handel/{vorgang}/versand", d.handelVersenden)
	weg.HandleFunc("POST /handel/{vorgang}/wiederaufnahme",
		d.handelWiederaufnehmen)
	weg.HandleFunc("POST /handelspost/{nachrichtId}/bestaetigung",
		d.handelspostBestaetigen)

	// **Die Zusagen sind eine Ressource des Vorgangs** (#1643): `POST` hält
	// eine benannte Besitzzeile fest, `GET` sagt, was festgehalten ist.
	weg.HandleFunc("POST /handel/{vorgang}/zusagen", d.handelFesthalten)
	weg.HandleFunc("GET /handel/{vorgang}/zusagen", d.handelszusagenZeigen)

	// **Und das Journal ist ein `GET`** — es entsteht nichts.
	weg.HandleFunc("GET /handel/{vorgang}/journal", d.handelsjournalZeigen)

	// **Und wer unterschrieben hat** — gelesen, ohne einen Schritt zu gehen
	// (Hüllenprobe `nat04handel`, #1672).
	weg.HandleFunc("GET /handel/{vorgang}/beleg", d.handelsbelegZeigen)

	// **Und wie er dorthin kam** (TRADE-03.AK-03, #1717).
	weg.HandleFunc("GET /handel/{vorgang}/protokoll", d.handelsprotokollZeigen)

	// **Und was die Seiten ungefähr wert sind** (TRADE-02b, #1951).
	weg.HandleFunc("GET /handel/{vorgang}/anhalt", d.handelsanhaltZeigen)

	// **Wohin eine Karte außerhalb führt** (T8, #1473) — ein `GET`, denn es
	// entsteht nichts. **Geöffnet wird hier nichts**: wer über einen Port
	// anfragt, sitzt nicht vor dem Bildschirm, auf dem sich ein Fenster
	// öffnete.
	weg.HandleFunc("GET /drucke/{printing}/links", d.marktlinkeZeigen)

	// **Das eigene Konto bei einem Marktplatz** (T11, #1476) — ein `GET`
	// und keine zweite Methode.
	weg.HandleFunc("GET /markt/{quelle}/konto", d.marktkontoZeigen)

	// **Und die Notiz zu einem Gegenüber** (T3, #1468) — ein `PUT`, weil es
	// genau eine gibt; eine leere löscht sie.
	weg.HandleFunc("PUT /peer/kontakte/{kennung}/notiz", d.kontaktNotiz)

	// **Und wer sich im selben Netz meldet** (N2, #1455) — ein `GET`, denn
	// es ist eine Frage. Dass sie ein Paket ins Netz schickt, ändert daran
	// nichts: sie ändert nichts an diesem Rechner.
	weg.HandleFunc("GET /peer/nachbarn", d.nachbarnZeigen)
}
