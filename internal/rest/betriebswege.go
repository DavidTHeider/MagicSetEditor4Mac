package rest

// **Die Wege über den Bestand selbst** — abgetrennt am 2026-09-10, weil
// `wegetafel.go` mit der Rücksetzung (O4, #1428) über die Zeilengrenze aus §2
// ging.
//
// # Der Schnitt folgt der Frage
//
// Nebenan stehen die Wege über die **Inhalte** — Sets, Karten, Decks, Bilder.
// Hier stehen die Wege über den **Bestand und den Prozess**: was er verträgt,
// wie es ihm geht, was er gewirkt hat, was in ihm herumliegt, was zurückgehen
// kann und was sich zurücksetzen lässt.
//
// **Sie haben gemeinsam, dass keiner von ihnen eine Karte anfasst.** Wer eine
// Auskunft über den Bestand sucht, sucht sie hier.

import (
	"net/http"
)

// betriebswegeEintragen legt die Wege über den Bestand an.
func (d *Dienst) betriebswegeEintragen(weg *http.ServeMux) {
	// **Die Grenzen** (#1429) hängen an keiner Naht — sie gelten immer.
	weg.HandleFunc("GET /grenzen", d.grenzenAuskunft)
	// **Der Betriebszustand** (O9, #1433) — `/gesundheit` als Auskunft,
	// `/bereit` mit dem Statuscode als Antwort. Beide hinter der Marke; der
	// Grund steht im Kopf von `gesundheit.go`.
	weg.HandleFunc("GET /gesundheit", d.gesundheit)
	weg.HandleFunc("GET /bereit", d.bereit)

	// **Die Metriken** (O9, #1433) — endlich viele Reihen, hinter der Marke.
	weg.HandleFunc("GET /metriken", d.metriken)

	// **Wer antwortet hier** (O9, #1433) — Fassung, Plattform, Fläche.
	weg.HandleFunc("GET /meta", d.meta)

	// **Die Verträglichkeit** (O6, #1430) — kann die neueste Freigabe
	// diesen Bestand öffnen?
	weg.HandleFunc("GET /vertraeglichkeit", d.vertraeglichkeit)
	weg.HandleFunc("GET /ruecksprung/arten", d.ruecksprungarten)
	weg.HandleFunc("GET /ruecksprung/lage", d.ruecksprunglage)
	weg.HandleFunc("GET /rueckkehren", d.rueckkehren)
	weg.HandleFunc("GET /pflege/vorschau", d.pflegevorschau)
	weg.HandleFunc("POST /pflege", d.pflegen)

	// **Die abgestufte Rücksetzung** (O4, #1428) — nachsehen und ausführen.
	// **Der lokale Suchindex** (A15, #1514) — drei lesende Wege und ein
	// schreibender. `DELETE` auf den Aufbauweg bricht ihn ab; ein zweiter
	// Bezeichner für dieselbe Sache wäre eine zweite Wahrheit.
	weg.HandleFunc("GET /index", d.indexstand)
	weg.HandleFunc("GET /index/suche", d.indexsuche)
	weg.HandleFunc("POST /index/aufbau", d.indexaufbau)
	weg.HandleFunc("DELETE /index/aufbau", d.indexabbruch)

	// **Und die Recherche darüber** (A16, #1515) — zwei lesende Wege und
	// einer, der wie ein Schreiben aussieht und keines ist: `POST
	// /kontextpaket` gibt zurück, **was an ein Modell ginge**, und sendet
	// nichts (`recherche.go`).
	weg.HandleFunc("GET /recherche", d.recherche)
	weg.HandleFunc("GET /recherche/aehnlich/{karte}", d.aehnlichekarten)
	weg.HandleFunc("POST /kontextpaket", d.kontextpaket)

	weg.HandleFunc("GET /ruecksetzung/arten", d.ruecksetzungsarten)
	weg.HandleFunc("GET /ruecksetzung/{art}/vorschau", d.ruecksetzungsvorschau)
	weg.HandleFunc("POST /ruecksetzung/{art}", d.zuruecksetzen)
	weg.HandleFunc("POST /rueckkehren/{id}/quittung", d.rueckkehrQuittieren)

	// **Das Kostenjournal** (O12, #1439) — der Stand, die Posten, die Grenze.
	weg.HandleFunc("GET /kosten/{konto}", d.kostenstand)
	weg.HandleFunc("GET /kosten/{konto}/posten", d.kostenposten)
	weg.HandleFunc("PUT /kosten/{konto}/budget", d.kostenbudget)
}
