package domain

// **Die Schlüssel, unter denen ein Gegenüber und das Netz nicht antworten**
// — N2 (#1455)
// und #2491. Eine eigene Datei, weil `fehlerschluessel.go` an der
// Zeilengrenze aus §2 stand, als #2491 zwei dazubrachte; angemeldet werden
// sie mit allen übrigen in `fehlerschluesselwege.go`.

const (
	// SchluesselOhnePeeridentitaet heißt: dieses Profil hat keine Identität
	// (N2, #1455).
	//
	// **Er steht neben `nichtGefunden` und nicht darin**, aus demselben Grund
	// wie der darüber: dessen Satz rät, die Liste neu zu laden, weil jemand
	// anderes etwas gelöscht haben könnte. Hier ist nichts gelöscht worden —
	// es gibt noch nichts, und der nächste Schritt ist ein Kasten weiter
	// oben. Angestrichen von der Sichtprüfung zu N2.
	SchluesselOhnePeeridentitaet Schluessel = "ohnePeeridentitaet"

	// SchluesselPeerNichtErreichbar heißt: an der Adresse eines Gegenübers
	// nahm niemand an ([ErrPeerNichtErreichbar], #2491). **Neben
	// `nichtGefunden`** wie der darüber: gelöscht ist nichts, und der Rat
	// ist, zu fragen, ob es horcht und ob seine Adresse stimmt.
	SchluesselPeerNichtErreichbar Schluessel = "peerNichtErreichbar"

	// SchluesselPeerOhneAdresse heißt: für ein Gegenüber ist keine Adresse
	// eingetragen ([ErrPeerOhneAdresse], #2491) — der Rat ist das Feld
	// „Adresse und Notiz" am Kontakt.
	SchluesselPeerOhneAdresse Schluessel = "peerOhneAdresse"

	// SchluesselOhneNetzerlaubnis heißt: die Suche im lokalen Netz ging
	// nicht hinaus (N2, #1455).
	//
	// **Er steht neben `validierung` und nicht darin**: dort geht es um eine
	// Eingabe, die jemand ändern kann. Hier ist die Eingabe in Ordnung und
	// die Erlaubnis fehlt — der nächste Schritt liegt in den
	// Systemeinstellungen und nicht im Formular. Angestrichen von der
	// Sichtprüfung zu N2.
	SchluesselOhneNetzerlaubnis Schluessel = "ohneNetzerlaubnis"

	// SchluesselOhneNetzroute heißt: ins lokale Netz gibt es keinen Weg
	// (NET-02.NAT-03, #1706).
	SchluesselOhneNetzroute Schluessel = "ohneNetzroute"
)
