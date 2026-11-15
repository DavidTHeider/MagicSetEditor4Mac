package mcp

// **Die abgestufte Rücksetzung über MCP — lesend, und das ist die
// Entscheidung** — Werkstück O4 von Phase 5.2
// (#1428).
//
// # Warum die Vorschau da ist und die Ausführung nicht
//
// **Dieselbe Trennung wie bei der Datenpflege** und aus demselben Grund
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)):
// die Vorschau ändert nichts und beantwortet eine Frage, die ein Modell für
// jemanden beantworten kann — „was verliere ich, wenn ich das tue?".
//
// **Die Ausführung löscht ein Profil, einen Menschen oder alles.** Und für
// die schwerste Art gibt es ohnehin keinen Weg: ein Nachweis entsteht am
// Fenster, von einem Menschen, und eine Maschine bekommt keinen (O1,
// [0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)). Ein Werkzeug,
// das die anderen fünf ausführte, machte aus dieser Grenze eine Beschriftung.

import (
	"context"
	"encoding/json"
)

// Betroffenes ist eine Gruppe, die eine Rücksetzung anfasst.
type Betroffenes struct {
	Was    string   `json:"was"`
	Anzahl int      `json:"anzahl"`
	Namen  []string `json:"namen"`
}

// Ruecksetzungsvorschau ist, was eine Rücksetzung anfassen würde.
type Ruecksetzungsvorschau struct {
	Art                string        `json:"art"`
	Betroffen          []Betroffenes `json:"betroffen"`
	Hindernisse        []Hindernis   `json:"hindernisse"`
	Bereit             bool          `json:"bereit"`
	NameNoetig         bool          `json:"nameNoetig"`
	NachweisNoetig     bool          `json:"nachweisNoetig"`
	SicherungEmpfohlen bool          `json:"sicherungEmpfohlen"`
}

// Ruecksetzungsauskunft ist die Naht zur Anwendung.
type Ruecksetzungsauskunft interface {
	Ruecksetzungsvorschau(ctx context.Context, art, profilID, nutzerID string) (
		Ruecksetzungsvorschau, error)
}

func ruecksetzungswerkzeug(naht Ruecksetzungsauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.ruecksetzungsvorschau", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, was eine Rücksetzung anfassen würde — **ändert " +
			"nichts**. »art« ist eine von sechs, von der leichtesten zur " +
			"schwersten: »layout« (nur das Aussehen des Fensters), " +
			"»profileinstellungen« (die Wahl des Profils, danach wird wieder " +
			"geerbt), »nutzereinstellungen«, »profil« (löscht es samt Sets, " +
			"Karten, Decks, Token), »nutzer« (löscht ihn samt seinen Profilen) " +
			"und »anwendung« (legt den ganzen Bestand beiseite). »betroffen« " +
			"zählt und nennt bis zu fünf Namen je Gruppe, »hindernisse« sagt, " +
			"was im Weg steht. AUSGEFÜHRT WIRD HIER NICHTS: das Löschen gehört " +
			"an eine Fläche, an der ein Mensch sitzt.",
		Schema: objekt(map[string]any{
			"art":      text("die Art der Rücksetzung — eine der sechs"),
			"profilID": text("wahlweise die Kennung des Profils; leer heißt: das gezeigte"),
			"nutzerID": text("wahlweise die Kennung des Menschen; leer heißt: der eigene"),
		}, "art"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var eingabe struct {
				Art      string `json:"art"`
				ProfilID string `json:"profilID"`
				NutzerID string `json:"nutzerID"`
			}

			if fehler := json.Unmarshal(roh, &eingabe); fehler != nil {
				return "", fehler
			}

			vorschau, fehler := naht.Ruecksetzungsvorschau(
				ctx, eingabe.Art, eingabe.ProfilID, eingabe.NutzerID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(vorschau)
		},
	}}
}
