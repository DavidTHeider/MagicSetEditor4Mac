package rest

// **Pakete als `.mse-installer` weitergeben und einlesen** — FMT-02.AK-04
// (#1691).
//
// Einordnung nach §16: **verträglich** — zwei Wege kommen hinzu. **`ziel`
// und `datei` sind Pfadfelder** (`domain.IstPfadfeld`): der Dienst nimmt nur
// Orte im Zuhause und in den eingestellten Ordnern.

import "net/http"

// weitergabewege trägt die zwei Wege ein — eine eigene Stelle, weil
// `wegetafel.go` an der Zeilengrenze aus §2 steht.
func (d *Dienst) weitergabewege(weg *http.ServeMux) {
	weg.HandleFunc("POST /pakete/weitergabe", d.paketeWeitergeben)
	weg.HandleFunc("POST /pakete/eingelesen", d.installerEinlesen)
}

// paketeWeitergeben schreibt die genannten Pakete als Installer.
func (d *Dienst) paketeWeitergeben(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Pakete []string `json:"pakete"`
		Ziel   string   `json:"ziel"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	datei, fehler := d.naehte.Pakete.PaketWeitergeben(anfrage.Context(), rumpf.Pakete, rumpf.Ziel)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]string{"datei": datei})
}

// installerEinlesen legt die Pakete eines Installers ins Verzeichnis.
func (d *Dienst) installerEinlesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Datei string `json:"datei"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	uebernahme, fehler := d.naehte.Pakete.PaketEinlesen(anfrage.Context(), rumpf.Datei)

	// **Der Teilerfolg geht mit dem Fehler hinaus** — wie bei der Übernahme.
	if fehler != nil {
		fehlerhaftMit(schreiben, fehler, uebernahme)

		return
	}

	ja(schreiben, http.StatusCreated, uebernahme)
}
