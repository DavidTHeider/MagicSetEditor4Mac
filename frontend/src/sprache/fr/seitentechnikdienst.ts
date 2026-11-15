// Die Texte der Technikseite über **Dienste und Wege** — welches Zeugnis
// ein Dienst vorzeigt (NET-03, #1707) und wie ein Gegenüber über fremde
// Netze angebunden ist (NET-02, #1706). Aus `seitentechnik.ts` herausgelöst,
// weil die Datei an der Zeilengrenze aus §2 stand.

export const dienstzeugnis = {
  eigenes: 'Certificat propre de {aussteller}, valable jusqu’au {bis}',
  identitaet:
    'Certificat fondé sur l’identité de ce profil — quiconque se connecte en compare l’empreinte.',
}

export const fernwege = {
  titel: 'Comment les correspondants sont connectés',
  keine: 'Personne n’a encore utilisé cette invitation.',
  direkt: 'Direct, via {adresse}',
  relais: 'Via le relais {adresse} — pas encore de chemin direct',
  aufbau: 'Connu, mais aucun chemin ne permet encore de l’atteindre',
  unbestimmt: 'Connexions : {anzahl} — le chemin n’est pas encore connu.',
}

// **Command line and tool server with one click** (OPS-09, #1907).
export const schnittstellen = {
  titel: 'Configurer la ligne de commande et le serveur d’outils',
  satz: 'Un bouton inscrit ce dont un programme de ce Mac a besoin pour trouver la ligne de commande ou le serveur d’outils — un second retire exactement cela. Avant, tu vois quel fichier et quelles lignes sont concernés ; les entrées d’autres sources restent telles quelles.',
  namen: {
    befehle: 'Les liens vers mse et mse-mcp dans ~/.local/bin',
    befehlszeile: 'La ligne de commande dans le PATH du shell de connexion',
    codex: 'Le serveur d’outils pour Codex',
    claudedesktop: 'Claude Desktop',
    claudecode: 'Claude Code',
    dienst: 'Le service pour scripts en LaunchAgent',
  },
  staende: {
    offen: 'non configuré',
    eingerichtet: 'configuré',
    veraltet: 'obsolète',
    fremd: 'entrée tierce',
    nichtGefunden: 'absent de ce Mac',
    unnoetig: 'inutile',
  },
  staendeSaetze: {
    veraltet:
      'L’application se trouve ailleurs maintenant ; Configurer le répare.',
    fremd: 'Une entrée tierce est déjà là — elle reste telle quelle.',
    unnoetig: 'Déjà dans le PATH.',
  },
  zeilen: 'Ces lignes sont ajoutées',
  vorschau: {
    json: 'Cette entrée est ajoutée au fichier',
    befehl: 'Cette commande s’exécute',
    agent: 'Ce fichier est créé, puis cette commande le charge',
  },
  nurEintrag:
    '{name} avait changé depuis — seule l’entrée propre a été retirée ; le reste reste tel que tu l’as laissé.',
  einrichten: 'Configurer',
  widerrufen: 'Révoquer',
  alle: {
    titel: 'Tout à la fois',
    satz: 'Configure chaque cible présente sur ce Mac et encore manquante — ou retire tout ce qui est configuré. Les entrées étrangères restent telles quelles.',
    vorschau: 'Ce qui sera écrit ({anzahl} cibles)',
    vorschauLeer: 'Rien ne manque : ce qui existe ici est configuré.',
    einrichten: 'Tout configurer',
    widerrufen: 'Tout retirer',
    ergebnis: 'Ce qui s’est passé',
    ausgaenge: {
      eingerichtet: 'configuré',
      widerrufen: 'retiré',
      fremd: 'entrée étrangère, inchangée',
      unveraendert: 'inchangé',
      gescheitert: 'échoué',
      abgebrochen: 'annulé',
    },
  },
  lauschwahl: {
    titel: 'Où écoutent les services',
    satz: 'S’applique au service pour scripts — lancé dans la fenêtre et comme LaunchAgent — et au serveur d’outils via HTTP, à partir de leur prochain démarrage.',
    wahl: 'Accessible via',
    schleife: 'Ce Mac uniquement (127.0.0.1)',
    auswahl: 'Adresses choisies',
    alle: 'Toutes les interfaces (0.0.0.0)',
    alleSatz: 'Chaque interface, y compris une ajoutée plus tard.',
    aussen:
      'Une adresse choisie quitte ce Mac : les services exigent alors TLS et un accès (jeton), et macOS peut demander l’autorisation réseau local.',
    inaktiv: 'désactivée',
    ohneAdresse: 'aucune adresse',
    liste: 'Interfaces de ce Mac',
    arten: {
      schleife: 'Boucle locale',
      netz: 'Réseau',
      tunnel: 'Tunnel (VPN)',
      bruecke: 'Pont',
      direkt: 'Sans-fil direct',
    },
    speichern: 'Appliquer le choix',
    gespeichert: 'Appliqué — prend effet au prochain démarrage d’un service.',
    keineGewaehlt: 'Choisis au moins une adresse.',
    platzhalter: 'comme sous « Où écoutent les services »',
    weitere: 'Écoute aussi sur',
    ohne: 'Sans {adresse} : {grund}',
  },
}
