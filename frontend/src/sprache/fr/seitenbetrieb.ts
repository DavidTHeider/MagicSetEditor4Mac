// **Run times, consumption and the performance profile** — OPS-05
// (#1722).

export const betrieb = {
  ueberschrift: 'Temps d’exécution et consommation',
  zeitraum: 'Relevé du {seit} au {bis}.',
  aufwand: 'Coût du relevé : {mikros} µs',
  speicher: 'Mémoire',
  speicherWert:
    '{fussabdruck} d’empreinte — dont {heap} de tas Go, {system} fournis par le système',
  speicherOhneFussabdruck:
    '{heap} de tas Go, {system} fournis par le système — l’empreinte n’est pas mesurée',
  web: 'Processus web',
  webWert: '{summe} au total — {teile}',
  webTeil: '{art} {groesse}',
  webArten: { webinhalt: 'Contenu web', gpu: 'Graphisme', netz: 'Réseau' },
  ohneWeb: 'Ce processus ne contient aucune vue web.',
  webUngemessen: 'Non mesuré : {grund}',
  goroutinen: 'Routines concurrentes',
  rechenzeit: 'Temps de calcul depuis le démarrage',
  datenbank: 'Connexions à la base de données',
  datenbankWert:
    '{offen} ouvertes, {inBenutzung} utilisées, {gewartet} attentes, {dauer} au total',
  ohneBestand: 'Ce processus ne détient aucune base de données.',
  zwischenspeicher: 'Caches des services externes',
  zwischenspeicherWert:
    '{treffer} succès, {geholt} récupérés, {veraltet} servis périmés, {fehlgeschlagen} échecs',
  laufzeiten: 'Temps d’exécution par interface',
  laufzeitenLeer:
    'Aucun appel n’est mesuré ici — seuls le service et le serveur d’outils sont mesurés.',
  ungemessen: 'Non mesuré : {liste}.',
  ungemessenNamen: {
    fenster:
      'les liaisons de la fenêtre (elles n’ont pas d’endroit commun où un appel se termine)',
    kommandozeile:
      'la ligne de commande (elle ne vit que le temps d’une commande)',
    speicher: 'l’empreinte de ce processus (le système ne l’a pas fournie)',
  },
  spalteArt: 'Type',
  spalteAnzahl: 'Appels',
  spalteFehler: 'Taux d’erreur',
  spalteMittel: 'Moyenne',
  spalteVerteilung: 'Répartition',
  arten: { lesend: 'lecture', schreibend: 'écriture' },
  eimer: 'jusqu’à {grenze} ms',
  verteilung: '{eimer} : {anzahl} appels',
  eimerOffen: 'au-delà',
  spurAus:
    'La trace est désactivée. Active-la sous Réglages › Diagnostic — elle s’applique à partir du prochain démarrage.',
  spurAn:
    'Trace activée : {aufgenommen} appels sur {gesehen}, échantillonnage à {abtastung} %. Au plus {grenze} entrées pendant {minuten} min — {verdraengt} évincées, {abgelaufen} expirées.',
}

export const leistungsprofil = {
  ueberschrift: 'Profil de performance',
  erklaerung:
    'Un profil CPU de cette fenêtre sur la durée choisie, au format de « go tool pprof ». Il est conservé à côté de la base de données ; le nombre de profils conservés se règle sous Réglages › Diagnostic.',
  sekunden: 'Durée en secondes',
  grenze: 'Un nombre entier de 1 à 60.',
  aufzeichnen: 'Enregistrer le profil',
  laeuft: 'Enregistrement en cours — la fenêtre reste utilisable.',
  fertig: 'Écrit : {datei} ({groesse})',
}

export const zwischenspeicher = {
  ueberschrift: 'Caches',
  erklaerung:
    'Ce que cette application conserve pour ne pas avoir à le récupérer ou à le recalculer deux fois — sur le disque et en mémoire. Rien de tout cela ne constitue tes données : une fois vidé, chaque cache se remplit de nouveau à la prochaine utilisation.',
  spalteSpeicher: 'Cache',
  spalteOrt: 'Emplacement',
  spalteQuote: 'Taux de succès',
  spalteEintraege: 'Entrées',
  spalteGroesse: 'Taille',
  spalteGrenze: 'Limite',
  keineQuote: 'aucune requête pour l’instant',
  keineGrenze: 'un fichier qui est remplacé',
  namen: {
    'scryfall.antworten': 'Réponses de Scryfall',
    'scryfall.bilder': 'Images de cartes de Scryfall',
    'scryfall.sammeldaten': 'Fichier de données groupées de Scryfall',
    'render.ebenen': 'Calques échantillonnés',
    'pakete.bilder': 'Images des paquets de modèles',
    'skript.baeume': 'Scripts analysés',
    'vorschau.bilder': 'Aperçus de cartes',
    'tisch.texturen': 'Images de cartes à la table',
  },
  orte: { platte: 'Disque', arbeitsspeicher: 'Mémoire' },
  leerenHinweis:
    'Vider supprime tout. Sans réseau, les réponses et les images de Scryfall ne reviennent qu’une fois le réseau rétabli.',
  leeren: 'Tout vider',
  geleert: 'Vidés — ils se remplissent de nouveau à la prochaine utilisation.',
}
