// **Die Wege, die an einem Nutzer, einem Profil oder einem Konto hängen** —
// Phase 4.12 (#1351).
//
// **Eine eigene Datei seit dem 2026-09-05**, als `endpunkteneben_test.go` mit
// den drei Kontoarten aus Werkstück 6 über die Zeilengrenze aus §2 ging. Der
// Schnitt folgt der Frage: nebenan stehen die Wege am **Bestand** vorbei,
// hier die, die sagen, **wer** ihn benutzt.

package rest_test

import "net/http"

// endpunkteNutzer sind die Wege an Nutzern, Profilen und Konten.
func endpunkteNutzer() []endpunkt {
	return []endpunkt{
		// **Nutzer und Profile** (#1351). Das Profil hängt am Nutzer, und der
		// Pfad sagt es.
		{
			name: "Nutzer lesen", methode: http.MethodGet, weg: "/nutzer",
			erfolg: http.StatusOK,
		},
		{
			name: "Nutzer anlegen", methode: http.MethodPost, weg: "/nutzer",
			rumpf: `{"name":"Zweite"}`, erfolg: http.StatusCreated,
		},
		{
			name: "Profil anlegen", methode: http.MethodPost, weg: "/nutzer/1/profile",
			rumpf: `{"name":"Turnier"}`, erfolg: http.StatusCreated,
		},
		// **Der Profilwechsel** — Werkstück 3 desselben Issues.
		//
		// **Einzahl und ohne Kennung im Weg**: es gibt genau eines, in dem
		// gearbeitet wird. `PUT` und `200`, weil nichts entsteht: derselbe Ruf
		// zweimal führt zum selben Zustand.
		{
			name: "gezeigtes Profil lesen", methode: http.MethodGet, weg: "/profil",
			erfolg: http.StatusOK,
		},
		{
			name: "Profil wechseln", methode: http.MethodPut, weg: "/profil",
			rumpf: `{"profilID":"20"}`, erfolg: http.StatusOK,
		},
		// **Die Anbieterkonten hängen am Nutzer**, und der Pfad sagt es —
		// Werkstück 4 (#1351). **204 beim Lösen**: es bleibt nichts zu zeigen.
		{
			name: "Konten lesen", methode: http.MethodGet, weg: "/nutzer/1/konten",
			erfolg: http.StatusOK,
		},
		{
			name: "Konto lösen", methode: http.MethodDelete,
			weg: "/nutzer/1/konten/github", erfolg: http.StatusNoContent,
		},
		// **Ändern und Löschen** — Werkstück 5 (#1351). **Der Name steht in
		// der Abfrage**: ein `DELETE` hat keinen Rumpf (#1218).
		{
			name: "Nutzer ändern", methode: http.MethodPut, weg: "/nutzer/1",
			rumpf: `{"name":"Umbenannt"}`, erfolg: http.StatusOK,
		},
		{
			name: "Profil ändern", methode: http.MethodPut, weg: "/profile/1",
			rumpf: `{"name":"Legacy"}`, erfolg: http.StatusOK,
		},
		{
			name: "Nutzer löschen", methode: http.MethodDelete,
			weg: "/nutzer/1?name=Ich", erfolg: http.StatusNoContent,
		},
		{
			name: "Profil löschen", methode: http.MethodDelete,
			weg: "/profile/1?name=Mein+Profil", erfolg: http.StatusNoContent,
		},
		// **Die vier Kontoarten** — Werkstück 6 (#1351). **Das Kennwort steht
		// im Rumpf und nie im Weg**: eine Anfragezeile landet in jedem
		// Zugriffsprotokoll.
		{
			name: "Konto eintragen", methode: http.MethodPost, weg: "/konten",
			rumpf:  `{"kennung":"wer@example.org","kennwort":"einlangesgeheimnis"}`,
			erfolg: http.StatusCreated,
		},
		{
			// **`200` und nicht `201`**: hier entsteht nichts.
			name: "Konto anmelden", methode: http.MethodPost, weg: "/konten/anmeldung",
			rumpf:  `{"kennung":"wer@example.org","kennwort":"einlangesgeheimnis"}`,
			erfolg: http.StatusOK,
		},
		{
			// **Ohne Rumpf, weil es nichts anzugeben gibt.**
			name: "Konto anonym", methode: http.MethodPost, weg: "/konten/anonym",
			erfolg: http.StatusCreated,
		},
		// **Der zweite Faktor hängt am Nutzer** — Werkstück 7 (#1351).
		// **204 beim Bestätigen, Einrichten der Biometrie und Entfernen**: es
		// bleibt nichts zu zeigen.
		{
			name: "Faktoren lesen", methode: http.MethodGet,
			weg: "/nutzer/1/faktoren", erfolg: http.StatusOK,
		},
		{
			name: "Faktor einrichten", methode: http.MethodPost,
			weg: "/nutzer/1/faktoren", rumpf: `{"art":"totp"}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "Biometrie einrichten", methode: http.MethodPost,
			weg: "/nutzer/1/faktoren", rumpf: `{"art":"biometrie"}`,
			erfolg: http.StatusNoContent,
		},
		{
			name: "Faktor bestätigen", methode: http.MethodPut,
			weg: "/nutzer/1/faktoren/totp", rumpf: `{"code":"123456"}`,
			erfolg: http.StatusNoContent,
		},
		{
			name: "Faktor entfernen", methode: http.MethodDelete,
			weg: "/nutzer/1/faktoren/totp", erfolg: http.StatusNoContent,
		},
		// **Das Profilbild hängt am Nutzer** — Werkstück 8 (#1351).
		//
		// **`PUT` und `200`, nicht `POST` und `201`**: es entsteht nichts
		// Neues, denn ein Mensch hat **ein** Bild — derselbe Ruf zweimal
		// führt zum selben Zustand. **Und `200` auch beim Löschen**, weil
		// der Nutzer zurückkommt: seine Fassung ist eine andere geworden,
		// und wer sie nicht bekäme, schriebe als Nächstes gegen einen Stand,
		// den es nicht mehr gibt (§7, Regel 2).
		{
			name: "Profilbild setzen", methode: http.MethodPut,
			weg: "/nutzer/1/bild", rumpf: `{"pfad":"/tmp/gesicht.png","fassung":"1"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Profilbild entfernen", methode: http.MethodDelete,
			weg: "/nutzer/1/bild?fassung=1", erfolg: http.StatusOK,
		},
		// **Die Einstellungen eines Menschen und seine Auskunft** —
		// Werkstück 8 (#1351), Punkte 9, 10 und 11.
		//
		// **`PUT` auf den Schlüssel**: er ist der Ort, und derselbe Ruf
		// zweimal führt zum selben Zustand. **Und `200` statt `204`**, weil
		// alle Einstellungen zurückkommen — wer nur einen Erfolg bekäme,
		// müsste danach noch einmal lesen (§7, Regel 2).
		{
			name: "Nutzereinstellungen lesen", methode: http.MethodGet,
			weg: "/nutzer/1/einstellungen", erfolg: http.StatusOK,
		},
		{
			name: "Nutzereinstellung setzen", methode: http.MethodPut,
			weg:   "/nutzer/1/einstellungen/barrierefreiheit.kontrast",
			rumpf: `{"wert":"hoch"}`, erfolg: http.StatusOK,
		},
		{
			name: "Personendaten lesen", methode: http.MethodGet,
			weg: "/nutzer/1/personendaten", erfolg: http.StatusOK,
		},
		// **Das Abzeichen hängt am Profil** — Werkstück 9 (#1351), Punkt 13.
		//
		// **`/profile/{id}/badge` und nicht unter dem Nutzer**: es gehört
		// einer **Rolle** und nicht dem Menschen — dieselbe Überlegung wie
		// bei `PUT /profile/{profilID}`.
		{
			name: "Profilbadge setzen", methode: http.MethodPut,
			weg: "/profile/1/badge", rumpf: `{"pfad":"/tmp/wappen.png","fassung":"1"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Profilbadge entfernen", methode: http.MethodDelete,
			weg: "/profile/1/badge?fassung=1", erfolg: http.StatusOK,
		},

		// **Die Zugriffstoken hängen ebenfalls am Profil** — Werkstück 10
		// (#1351), Punkt 13. **`POST` gibt `201`**: es entsteht etwas, und
		// zwar das einzige Mal, dass der Klartext herauskommt.
		//
		// **Und das Ändern und Löschen hängt am Token selbst**, nicht unter
		// dem Profil: ein Token wechselt sein Profil nicht, und ein Weg, der
		// es nennt, ließe sich mit einem falschen Profil aufrufen.
		{
			name: "Token lesen", methode: http.MethodGet,
			weg: "/profile/1/token", erfolg: http.StatusOK,
		},
		{
			name: "Token anlegen", methode: http.MethodPost,
			weg:    "/profile/1/token",
			rumpf:  `{"bezeichnung":"Nachtlauf","principal":"Rechner","scope":"lesen"}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "Token ändern", methode: http.MethodPut,
			weg:    "/token/1",
			rumpf:  `{"bezeichnung":"Tageslauf","principal":"Rechner","scope":"lesen","fassung":"1"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Token löschen", methode: http.MethodDelete,
			weg: "/token/1", erfolg: http.StatusOK,
		},

		// **Rotation, Sammelwiderruf und Ablaufwahlen** — Werkstück O2 von
		// Phase 5.2 (#1426). Die Rotation antwortet mit `201` wie das
		// Anlegen: es entsteht ein Geheimnis.
		{
			name: "Token rotieren", methode: http.MethodPost,
			weg: "/token/1/rotation", erfolg: http.StatusCreated,
		},
		{
			name: "alle Token eines Profils widerrufen", methode: http.MethodDelete,
			weg: "/profile/1/token", erfolg: http.StatusOK,
		},
		{
			name: "die Ablaufwahlen", methode: http.MethodGet,
			weg: "/token/ablaufwahlen", erfolg: http.StatusOK,
		},

		// **Die Vorgaben, die auf Zeilen zeigen** — Werkstück 15 (#1351),
		// Punkte 4 bis 7. **`PUT` auf die Art**: sie ist der Ort, und
		// derselbe Ruf zweimal führt zum selben Zustand.
		{
			name: "Vorgaben lesen", methode: http.MethodGet,
			weg: "/profile/1/vorgaben", erfolg: http.StatusOK,
		},
		{
			name: "Vorgabe setzen", methode: http.MethodPut,
			weg: "/profile/1/vorgaben/deck", rumpf: `{"werte":["7"]}`,
			erfolg: http.StatusOK,
		},

		// **Die Freigaben** — Werkstück 11 (#1351). **`PUT` auf den Nutzer**:
		// er ist der Ort, und derselbe Ruf zweimal führt zum selben Zustand.
		{
			name: "Freigaben lesen", methode: http.MethodGet,
			weg: "/profile/1/freigaben", erfolg: http.StatusOK,
		},
		{
			name: "Freigabe setzen", methode: http.MethodPut,
			weg: "/profile/1/freigaben/2", rumpf: `{"rolle":"leser"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Freigabe zurückziehen", methode: http.MethodDelete,
			weg: "/profile/1/freigaben/2", erfolg: http.StatusOK,
		},

		// **Die Deckübernahme** — Werkstück 12 desselben Issues. **`POST`
		// und 201**: es entsteht ein Deck, und ein zweiter Ruf legte ein
		// zweites an. **Unter `/decks`**, weil der Weg nennt, wo etwas
		// entsteht; das fremde Profil steht im Rumpf.
		{
			name: "Deck übernehmen", methode: http.MethodPost,
			weg: "/decks/uebernehmen", rumpf: `{"quellProfil":"2","quellDeck":"7"}`,
			erfolg: http.StatusCreated,
		},

		// **Und seine lesende Hälfte** (UX-08.AK-04, #1680): die Decks des
		// fremden Profils zur Wahl. **Unter dem Profil**, denn gelesen wird
		// dort; hinter derselben Erlaubnis wie die Übernahme.
		{
			name: "Decks zur Übernahme", methode: http.MethodGet,
			weg: "/profile/2/decks", erfolg: http.StatusOK,
		},
	}
}
