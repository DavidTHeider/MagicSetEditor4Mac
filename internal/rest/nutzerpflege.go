// **Nutzer und Profile ändern und löschen im Dienst** — Werkstück 5 von
// Phase 4.12
// (#1351).

package rest

import "net/http"

// nutzerAendern schreibt die Angaben eines Nutzers fort.
//
// **Die Kennung steht im Weg, die Angaben im Rumpf** — Werkstück 8 (#1351).
// Ein Rumpf mit elf Feldern ist kein Grund, die Adresse zu ändern: der Ort
// bleibt der Nutzer.
func (d *Dienst) nutzerAendern(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf Nutzerangaben

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	rumpf.NutzerID = anfrage.PathValue("nutzerID")

	geaendert, fehler := d.naehte.Nutzer.NutzerAendern(anfrage.Context(), rumpf)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"nutzer": geaendert})
}

// profilAendern schreibt den Namen eines Profils fort.
//
// **`/profile/{id}` und nicht `/nutzer/{id}/profile/{id}`**: ein Profil
// wechselt seinen Nutzer nicht, und ein Weg, der ihn nennt, ließe sich mit
// einem falschen Nutzer aufrufen — der Dienst müsste dann prüfen, was der
// Bestand ohnehin weiß.
func (d *Dienst) profilAendern(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf Profilangaben

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	rumpf.ProfilID = anfrage.PathValue("profilID")

	geaendert, fehler := d.naehte.Nutzer.ProfilAendern(anfrage.Context(), rumpf)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"profil": geaendert})
}

// nutzerLoeschen nimmt einen Nutzer samt seinen Profilen heraus.
//
// **Der Name steht in der Abfrage** — ein `DELETE` hat keinen Rumpf, und die
// Bestätigung muss trotzdem mitkommen (#1218).
func (d *Dienst) nutzerLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	if fehler := d.naehte.Nutzer.NutzerLoeschen(anfrage.Context(),
		anfrage.PathValue("nutzerID"),
		anfrage.URL.Query().Get("name")); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	schreiben.WriteHeader(http.StatusNoContent)
}

// profilLoeschen nimmt ein Profil heraus.
func (d *Dienst) profilLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	if fehler := d.naehte.Nutzer.ProfilLoeschen(anfrage.Context(),
		anfrage.PathValue("profilID"),
		anfrage.URL.Query().Get("name")); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	schreiben.WriteHeader(http.StatusNoContent)
}
