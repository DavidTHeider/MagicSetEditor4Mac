// Die Texte der Technikseite über **Dienste und Wege** — welches Zeugnis
// ein Dienst vorzeigt (NET-03, #1707) und wie ein Gegenüber über fremde
// Netze angebunden ist (NET-02, #1706). Aus `seitentechnik.ts` herausgelöst,
// weil die Datei an der Zeilengrenze aus §2 stand.

export const dienstzeugnis = {
  eigenes: 'Eigenes Zeugnis von {aussteller}, gültig bis {bis}',
  identitaet:
    'Zeugnis auf der Identität dieses Profils — wer sich verbindet, vergleicht den Fingerabdruck.',
}

export const fernwege = {
  titel: 'Wie die Gegenüber angebunden sind',
  keine: 'Noch hat niemand diese Einladung benutzt.',
  direkt: 'Direkt, über {adresse}',
  relais: 'Über das Relais {adresse} — noch ohne direkten Weg',
  aufbau: 'Bekannt, aber noch über keinen Weg erreichbar',
  unbestimmt: 'Angebunden: {anzahl} — über welchen Weg, steht noch nicht fest.',
}

// **Kommandozeile und Werkzeugserver per Knopf** (OPS-09, #1907).
export const schnittstellen = {
  titel: 'Kommandozeile und Werkzeugserver einrichten',
  satz: 'Ein Knopf trägt ein, was Shell, Codex, Claude und der Dienst brauchen, ein zweiter nimmt genau das wieder heraus. Vorher steht da, welche Datei und welche Zeilen; fremde Einträge bleiben, wie sie sind.',
  namen: {
    befehle: 'Die Verweise auf mse und mse-mcp in ~/.local/bin',
    befehlszeile: 'Die Kommandozeile im PATH der Anmeldeshell',
    codex: 'Der Werkzeugserver für Codex',
    claudedesktop: 'Der Werkzeugserver für Claude Desktop',
    claudecode: 'Der Werkzeugserver für Claude Code',
    dienst: 'Der Dienst für Skripte als LaunchAgent',
  },
  // **Der Stand ist eine Marke und der Satz steht darunter** — UXR-19
  // (#1927):
  // als Fließtext las sich „eingerichtet" wie jede andere Zeile der Karte.
  staende: {
    offen: 'nicht eingerichtet',
    eingerichtet: 'eingerichtet',
    veraltet: 'veraltet',
    fremd: 'fremder Eintrag',
    nichtGefunden: 'nicht auf diesem Mac',
    unnoetig: 'nicht nötig',
  },
  staendeSaetze: {
    veraltet: 'Die Anwendung liegt woanders; Einrichten repariert es.',
    fremd: 'Ein fremder Eintrag steht schon da — er bleibt, wie er ist.',
    unnoetig: 'Es steht schon im PATH.',
  },
  zeilen: 'Diese Zeilen kommen hinzu',
  vorschau: {
    json: 'Dieser Eintrag kommt in die Datei',
    befehl: 'Dieser Befehl läuft',
    agent: 'Diese Datei entsteht, und dieser Befehl lädt sie',
  },
  nurEintrag:
    'In {name} hatte sich seither etwas geändert — genommen ist nur der eigene Eintrag, der Rest bleibt, wie du ihn hinterlassen hast.',
  einrichten: 'Einrichten',
  widerrufen: 'Widerrufen',
  // **Alles mit einem Knopf und die Wahl der Schnittstellen** — G08
  // (#2272).
  alle: {
    titel: 'Alles auf einmal',
    satz: 'Richtet jedes Ziel ein, das es auf diesem Mac gibt und das noch fehlt — oder nimmt alles Eingerichtete zurück. Fremde Einträge bleiben, wie sie sind.',
    vorschau: 'Was dabei geschrieben wird ({anzahl} Ziele)',
    vorschauLeer: 'Es fehlt nichts — eingerichtet ist, was es hier gibt.',
    einrichten: 'Alles einrichten',
    widerrufen: 'Alles zurücknehmen',
    ergebnis: 'Was geschah',
    ausgaenge: {
      eingerichtet: 'eingerichtet',
      widerrufen: 'zurückgenommen',
      fremd: 'fremder Eintrag, unverändert',
      unveraendert: 'unverändert',
      gescheitert: 'gescheitert',
      abgebrochen: 'abgebrochen',
    },
  },
  lauschwahl: {
    titel: 'Wo die Dienste lauschen',
    satz: 'Gilt für den Dienst für Skripte — im Fenster gestartet und als LaunchAgent — und für den Werkzeugserver über HTTP, ab ihrem nächsten Start.',
    wahl: 'Erreichbar über',
    schleife: 'Nur dieser Mac (127.0.0.1)',
    auswahl: 'Ausgewählte Adressen',
    alle: 'Alle Schnittstellen (0.0.0.0)',
    alleSatz: 'Jede Schnittstelle, auch eine, die später dazukommt.',
    aussen:
      'Eine gewählte Adresse verlässt diesen Mac. Die Dienste verlangen dann TLS und einen Zugang (Token), und macOS fragt womöglich nach der Erlaubnis für das lokale Netzwerk.',
    inaktiv: 'ausgeschaltet',
    ohneAdresse: 'keine Adresse',
    liste: 'Schnittstellen dieses Mac',
    arten: {
      schleife: 'Schleife',
      netz: 'Netz',
      tunnel: 'Tunnel (VPN)',
      bruecke: 'Brücke',
      direkt: 'Direktfunk',
    },
    speichern: 'Wahl übernehmen',
    gespeichert: 'Übernommen — wirkt beim nächsten Start eines Dienstes.',
    keineGewaehlt: 'Wähle mindestens eine Adresse aus.',
    platzhalter: 'wie unter „Wo die Dienste lauschen“',
    weitere: 'Lauscht außerdem auf',
    ohne: 'Ohne {adresse}: {grund}',
  },
}
