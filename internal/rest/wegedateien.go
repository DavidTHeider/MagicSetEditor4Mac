// **Die Wege, die an keinem Set hängen** — die Sicherung des Bestands
// (#1169),
// der portable Umschlag
// (#1457)
// und der Netzweg
// (#1571).
//
// **Eine eigene Datei, weil `wegetafel.go` mit ihnen über die Zeilengrenze
// aus §2 ginge** — und der Schnitt folgt der Frage: sie fassen eine **Datei
// auf der Platte** an oder beschreiben diesen **Prozess**, und keine von
// ihnen kennt ein Set.

package rest

import "net/http"

// wegeAnGanzenDateien trägt sie ein.
func (d *Dienst) wegeAnGanzenDateien(weg *http.ServeMux) {
	// **Die Aufträge der langen Datei-Läufe und die Vorschau vor dem
	// Einfahren** (FMT-03, #1692).
	d.auftragswegeEintragen(weg)
	weg.HandleFunc("POST /sets/vorschau", d.setvorschau)

	// **Und der Ausgang** (FMT-03.AK-03) — was eine Ausgabe ohne `ziel`
	// geschrieben hat, `dateiuebertragung.go`.
	weg.HandleFunc("GET /ausgang/{name}", d.ausgangAbholen)
	weg.HandleFunc("DELETE /ausgang/{name}", d.ausgangVergessen)

	// **Die Sicherung des ganzen Bestands** (#1169) — deshalb `/bestand` und
	// nicht `/sets/{id}/…`.
	weg.HandleFunc("POST /bestand/sicherung", d.bestandSichern)
	weg.HandleFunc("GET /bestand/sicherung", d.sicherungSchauen)
	weg.HandleFunc("GET /bestand/verschluesselung", d.verschluesselungSchauen)
	weg.HandleFunc("GET /bestand/sicherungsstand", d.sicherungsstandSchauen)

	// **Eine Datei ansehen** (UX-13, #1685) — der Pfadvertrag gilt wie
	// überall, und Bilddaten gehen nicht mit.
	weg.HandleFunc("GET /dateien/ansicht", d.dateiAnsehen)

	// **Der portable Umschlag** (N4, #1457) — der Weg hinaus, das Ansehen und
	// die ausdrückliche Übernahme.
	weg.HandleFunc("POST /umschlag", d.umschlagSchnueren)
	weg.HandleFunc("GET /umschlag", d.umschlagSchauen)
	weg.HandleFunc("POST /umschlag/uebernahme", d.umschlagUebernehmen)

	// **Der Netzweg** (N5, #1571) — rein lesend; geändert wird über die
	// Einstellungen.
	weg.HandleFunc("GET /netz", d.netzwegSchauen)

	// **Und die Schnittstellen** (OPS-09, #1907) — rein lesend.
	weg.HandleFunc("GET /schnittstellen", d.schnittstellenSchauen)

	// **Und die Schalter** (OPS-10, #1908) — lesend; geschaltet wird über
	// die Einstellungen.
	weg.HandleFunc("GET /schalter", d.schalterSchauen)

	// **Und die Regelfreigabe** (G4, #1461) — rein lesend.
	weg.HandleFunc("GET /regelfreigabe", d.regelfreigabeSchauen)

	// **Die Partien** (G1, #1458). **Sie hängen an keinem Set**, und genau
	// das ist der Grund, warum sie hier stehen: ein Deck wird beim Anlegen
	// eingefroren, und danach kennt eine Partie kein Set mehr.
	weg.HandleFunc("POST /partien", d.partieAnlegen)
	weg.HandleFunc("GET /partien", d.partienZeigen)
	weg.HandleFunc("GET /partien/{partieID}", d.partieZeigen)

	// **Der Verlauf** — Werkstück G5 von Phase 5.5
	// (#1462).
	weg.HandleFunc("GET /partien/{partieID}/verlauf", d.partieVerlauf)
	weg.HandleFunc("POST /partien/{partieID}/befehle", d.partieBefehl)
	weg.HandleFunc("POST /partien/{partieID}/stand", d.partieStand)

	// **Die Verwaltung** — PLAY-03.AK-05 (#1711). `DELETE` und nicht
	// `POST …/loeschen`: es ist die Bedeutung, die HTTP dafür hat, und ein
	// eigener Weg daneben wäre ein zweiter Name für dieselbe Sache.
	weg.HandleFunc("DELETE /partien/{partieID}", d.partieLoeschen)
	weg.HandleFunc("POST /partien/{partieID}/kopie", d.partieKopieren)
	weg.HandleFunc("GET /partien/{partieID}/ausfuhr", d.partieAusfuhr)
	weg.HandleFunc("POST /partien/einfuhr", d.partieEinfuhr)
}
