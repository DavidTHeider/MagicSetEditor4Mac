package app

// **Die Zwischenspeicher des Zeichenwerks** — OPS-08
// (#1725):
// die abgetasteten Ebenen, die Bilder der Pakete und die Spielbilder am Tisch.

// werkspeicher gibt Auskunft und Leeren der drei.
func werkspeicher(werk *Zeichenwerk) (quellen []func() ZwischenspeicherDTO, leerer []func()) {
	quellen = []func() ZwischenspeicherDTO{
		func() ZwischenspeicherDTO {
			zahlen := werk.Rasterer.Ebenenzahlen()

			return ZwischenspeicherDTO{
				Name: SpeicherEbenen, Ort: OrtArbeitsspeicher, Treffer: zahlen.Treffer,
				Fehlgriffe: zahlen.Fehlgriffe, Eintraege: zahlen.Eintraege, Bytes: zahlen.Bytes,
				GrenzeBytes: zahlen.GrenzeBytes,
			}
		},
		func() ZwischenspeicherDTO {
			zahlen := werk.Bildquelle.Zahlen()

			return ZwischenspeicherDTO{
				Name: SpeicherPaketbilder, Ort: OrtArbeitsspeicher, Treffer: zahlen.Treffer,
				Fehlgriffe: zahlen.Fehlgriffe, Eintraege: zahlen.Eintraege, Bytes: zahlen.Bytes,
				GrenzeEintraege: zahlen.GrenzeEintraege, GrenzeBytes: zahlen.GrenzeBytes,
			}
		},
		func() ZwischenspeicherDTO {
			zahlen := werk.Spielbildfach.Zahlen()

			return ZwischenspeicherDTO{
				Name: SpeicherTexturen, Ort: OrtArbeitsspeicher, Treffer: zahlen.Treffer,
				Fehlgriffe: zahlen.Fehlgriffe, Eintraege: zahlen.Eintraege, Bytes: zahlen.Bytes,
				GrenzeBytes: zahlen.GrenzeBytes,
			}
		},
	}

	return quellen, []func(){werk.Rasterer.EbenenLeeren, werk.Bildquelle.Leeren, werk.Spielbildfach.Raeume}
}
