package domain

// **Wie viel vom Kartenformular dasteht** — UX-05.AK-01
// (#1677).
//
// # Warum eine eigene Datei
//
// **Weil `einstellung.go` bei 398 Zeilen steht** (§2). Der Schnitt folgt der
// Frage: dort steht, **was eine Einstellung ist**, hier, **was die drei
// Tiefen bedeuten** — und das ist mehr als ein Eintrag in einer Tabelle.
//
// # Die Zuordnung steht im Fenster und nicht hier
//
// **Weil sie aus Feldmetadaten gerechnet wird** und nicht aus einer Liste von
// Feldnamen: `frontend/src/seiten/feldtiefe.ts` entscheidet je Feld, welche
// Tiefe es braucht. Hier steht nur, **welche es gibt** und welche gilt,
// solange niemand etwas gewählt hat — das ist die Einstellung, und die
// gehört dem Bestand.
//
// # Und warum MSE2 dazu nichts sagt
//
// **Nachgelesen am 2026-09-21** in `src/data/field.hpp`, Zeilen 46–59,
// Fassung f9d9356 vom 2021-07-06: ein `Field` trägt `editable`,
// `save_value`, `show_statistics`, `identifying`, `card_list_visible` und
// `card_list_allow` — **kein** Merkmal für Komplexität. MSE2 zeigt alle
// Felder in einem Bogen. Die Einteilung ist damit **neu** (§15); gesucht
// wurde in `field.hpp`, `card.hpp` und `card_editor.cpp`.

// EinstellungFeldtiefe ist, wie viel das Kartenformular zeigt.
//
// **Auf Profilebene und nicht je Gerät**: wer erfahren genug für den
// Expertenmodus ist, ist es an jedem Rechner. Das Kriterium sagt es
// ausdrücklich — „ein je Nutzer oder Profil änderbarer Standardmodus".
const EinstellungFeldtiefe = "feldtiefe"

// Die drei Werte von [EinstellungFeldtiefe].
//
// **Sie liegen ineinander**: `normal ⊆ erweitert ⊆ experte`. Wer tiefer
// geht, sieht mehr — nie anderes. AK-02 verlangt, dass ein Moduswechsel
// keine versteckten Werte verliert; so ist es keine Zusage, sondern eine
// Eigenschaft der Ordnung.
const (
	// FeldtiefeNormal zeigt, was jemand ausfüllt — die Vorgabe.
	FeldtiefeNormal = "normal"

	// FeldtiefeErweitert nimmt dazu, was ein Skript füllt.
	FeldtiefeErweitert = "erweitert"

	// FeldtiefeExperte nimmt dazu, was das Spielpaket nicht erklärt.
	FeldtiefeExperte = "experte"
)

// init hängt die Tiefe an die Einstellungsreihe.
//
// **Wie die Dienstschalter in `einstellungsreihe.go`**, und aus demselben
// Grund: die Reihe steht in `einstellung.go`, und die Datei ist voll. Sie
// landet damit **am Ende** der Anzeigereihenfolge — das ist keine Aussage
// über ihre Wichtigkeit, sondern die Stelle, an der ein Anhang steht.
func init() {
	einstellungen = append(einstellungen, einstellung{
		Schluessel: EinstellungFeldtiefe,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte: []string{
			FeldtiefeNormal, FeldtiefeErweitert, FeldtiefeExperte,
		},
		Vorgabe: FeldtiefeNormal,
	})
}
