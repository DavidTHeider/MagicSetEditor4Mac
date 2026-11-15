package domain

// **Was ein Teilnehmer später sehen darf** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Die Zusage, für die es diese Rechnung gibt
//
// Das Issue sagt sie zweimal: „kein Replay exportiert unbemerkt fremde private
// Informationen" und „Replay/Chat/Logs eines Teilnehmers enthalten nur seine
// freigegebenen Informationen."
//
// **Die Gefahr ist nicht der laufende Tisch, sondern das Protokoll danach.**
// Am Tisch rechnet [Partie.Sicht] je Sitz — dort kommt gar nicht erst an, was
// nicht hingehört. Ein Ereignisstrom dagegen liegt **vollständig** vor: er ist
// die Wahrheit, aus der sich jeder Stand nachrechnen lässt, und genau deshalb
// trägt er alles. Wer ihn ungeprüft ausgibt, verrät im Nachhinein, was während
// des Spiels niemand sehen durfte.
//
// # Warum das Öffentliche bleibt
//
// **Dass jemand gezogen hat, ist kein Geheimnis** — die Reihenfolge der Züge
// ist der Verlauf der Partie. Was er gezogen hat, steht in den Ergänzungen,
// und die sind empfängergebunden (siehe [Ergaenzung]).
//
// **Und die Kette bleibt auch.** Sie ist der Beleg, dass nichts fehlt; wer sie
// aus einer Rückschau striche, nähme dem Leser die Möglichkeit nachzurechnen,
// dass er einen vollständigen Verlauf vor sich hat.

// RueckschauFuer gibt den Strom, wie ein Empfänger ihn sehen darf.
//
// **Der Strom selbst bleibt unberührt.** Wer ihn beim Filtern änderte, hätte
// nach der ersten Rückschau nur noch diese eine — und die nächste Sicht sähe
// ein Protokoll, aus dem schon etwas entfernt wurde.
//
// # Warum hier keine zweite Rechnung steht
//
// **Weil [Ereignis.Fuer] dieselbe ist.** Der erste Anlauf schrieb daneben ein
// `fuerRueckschau`, das Wort für Wort dasselbe tat — und genau davor warnt
// §10: zwei Stellen mit derselben Regel laufen beim ersten neuen Feld
// auseinander, und die Rückschau wäre dann die, die es **nicht** mitbekommt.
// Was der Rückschau eigen ist, steht im Kopf dieser Datei: nicht die
// Rechnung, sondern der Anlass.
//
// **Ein leerer Empfänger bekommt keine Ergänzung** — das ist der Fall beim
// Ausführen eines Protokolls ohne Angabe, und er darf nicht versehentlich
// alles zeigen. [Ereignis.FuerMitleser] hält das; hier steht, dass es hier
// gebraucht wird.
//
// # Und warum die Freigaben hier ankommen und nicht am Tisch
//
// **Weil eine Einwilligung von vorhin keine von jetzt ist.** [Offenlegung]
// sagt, dass ein Sitz seine privaten Ergänzungen **später** hergibt — für die
// Nachlese. Wer sie am laufenden Tisch wirken ließe, zeigte ein Blatt während
// des Spiels, für das jemand die Nachlese freigegeben hat.
//
// **Der Wächter dafür ist, dass [Partie.Sicht] sie gar nicht kennt.**
func RueckschauFuer(
	strom []Ereignis, fuer string, freigaben []Offenlegung,
) []Ereignis {
	darf := Mitleser(freigaben, fuer)

	gesehen := make([]Ereignis, 0, len(strom))

	for _, ereignis := range strom {
		gesehen = append(gesehen, ereignis.FuerMitleser(darf))
	}

	return gesehen
}
