package domain

// **Die Schlüssel der eigenen Übersetzungen** — LOC-02
// (#1688).
// Eine eigene Datei aus demselben Grund wie `fehlerschluesselsymbol.go`;
// angemeldet werden sie in `fehlerschluesselwege.go`.

const (
	// SchluesselSprachkennungUngueltig heißt: die Kennung ist keine
	// BCP-47-Kennung in der Schreibweise, die dieses Programm führt. Wert:
	// `kennung`.
	SchluesselSprachkennungUngueltig Schluessel = "sprachkennungUngueltig"

	// SchluesselSprachkennungEingebaut heißt: die Kennung gehört einer
	// eingebauten Sprache. Wert: `kennung`.
	SchluesselSprachkennungEingebaut Schluessel = "sprachkennungEingebaut"

	// SchluesselPlatzhalterAbweichend heißt: eine Übersetzung nennt andere
	// Platzhalter als ihre Vorlage. Werte: `pfad` und `erwartet`.
	SchluesselPlatzhalterAbweichend Schluessel = "platzhalterAbweichend"

	// SchluesselAuszeichnungUnsicher heißt: eine Übersetzung trägt, was
	// nicht Text ist — ein Tag, eine Skriptadresse, ein Steuerzeichen, oder
	// eine Auszeichnung, die nicht schließt. Wert: `pfad`.
	SchluesselAuszeichnungUnsicher Schluessel = "auszeichnungUnsicher"

	// SchluesselSprachpaketNeuer heißt: die Datei stammt aus einem neueren
	// Programm. Wert: `format`.
	SchluesselSprachpaketNeuer Schluessel = "sprachpaketNeuer"

	// SchluesselSprachpaketOhneLizenz heißt: ein Paket ohne Lizenzhinweis
	// geht nicht hinaus (AK-04).
	SchluesselSprachpaketOhneLizenz Schluessel = "sprachpaketOhneLizenz"
)
