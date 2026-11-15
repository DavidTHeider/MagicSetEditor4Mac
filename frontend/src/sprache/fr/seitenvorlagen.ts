// Die Texte der **Paketwerkstatt** (#1164) — eine Vorlage öffnen, ändern,
// speichern.
//
// **Eine eigene Datei**, weil `seiten.ts` an der Zeilengrenze aus §2 steht.
// Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als **einer**
// ankommt (§9).
export const vorlagen = {
  titel: 'Modèles',
  erklaerung:
    'Ce qu’une carte sait faire et son apparence sont définis dans des fichiers texte du paquet de modèles — champs, scripts, dimensions. Ils peuvent être modifiés ici. L’état précédent reste comme sauvegarde à côté du paquet.',
  paket: 'Paquet',
  keinPaket: 'Aucun paquet choisi pour l’instant.',
  waehleEines:
    'Choisis un paquet ci-dessus pour voir et modifier ses fichiers texte — champs, scripts, dimensions.',
  keineDateien: 'Ce paquet n’a aucun fichier à modifier.',
  dateien: 'Fichiers',
  datei: 'Fichier',
  inhalt: 'Contenu',
  speichern: 'Enregistrer',
  gespeichert: 'Enregistré.',
  ungespeichert: 'Non enregistré.',
  warnung:
    'Ce fichier vaut pour tous les sets qui utilisent ce paquet — pas seulement pour celui qui est ouvert.',
  konflikt:
    'Le fichier a changé depuis son ouverture. Charge sa version — ton texte est alors perdu — ou garde la tienne : les modifications montrent alors ce que tu écrases dans la sienne.',
  neuLaden: 'Charger sa version',
  behalten: 'Garder la mienne',
  vorschau: 'Aperçu',
  vorschauErklaerung:
    'La première carte du premier set qui utilise ce modèle — dessinée avec le texte affiché ici. Rien n’est enregistré pour autant.',
  vorschauLaeuft: 'Dessin en cours…',
  vorschauKeineKarte:
    'Aucun set ouvert n’utilise ce modèle. Ouvre sous « Sets » un set qui le nomme — tu verras alors ici l’effet d’une modification.',

  arbeitsplatz: 'Atelier de modèles',
  baumBezeichnung: 'Paquets et fichiers',
  inspectorBezeichnung: 'Aperçu et diagnostic',
  baumSuche: 'Rechercher des paquets et des fichiers',
  baumSuchePlatzhalter: 'Paquet ou fichier…',
  baumLeer: 'Aucun paquet ni fichier ne correspond à la recherche.',
  baumOhnePakete:
    'Aucun paquet à modifier pour l’instant. Crées-en un sous « Paquets » ou récupères-en un — il apparaîtra alors ici.',
  nurLesbar:
    'Fourni — en lecture seule. Une copie personnelle peut être modifiée.',
  baumNurMitgeliefert:
    'Tous les paquets ici sont fournis. Crées-en une copie personnelle ou un nouveau sous « Paquets » — les deux peuvent être modifiés et apparaissent alors en haut.',
  dateienLaden: 'Lecture des fichiers…',
  inspectorLeer:
    'L’aperçu, le diagnostic et l’origine du fichier ouvert s’affichent ici.',
  waehleImBaum:
    'Choisis un fichier dans l’arborescence à gauche — champs, scripts et dimensions s’y trouvent sous forme de texte et peuvent être modifiés ici.',
  verwerfenFrage:
    'Le fichier ouvert contient des modifications non enregistrées. Les abandonner et ouvrir {datei} ?',
  verwerfen: 'Abandonner et ouvrir',
  hierBleiben: 'Rester ici',
  testkarte: 'Carte de test',
  testkarteVorgabe: 'Première carte du premier set',
  testkarteOhneNamen: 'Carte {nummer}',
  testkarteErklaerung:
    'Un modèle casse rarement sur la première carte. Le set d’exemple « Werkstatt-Testkarten » apporte un texte long, une mise en page spéciale, une seconde face, une image manquante et des caractères spéciaux — ouvre-le sous « Sets » et change son style pour ce paquet.',
  veraltet:
    'Périmé — le fichier a un constat à la ligne {zeile} ; le dernier état valide est affiché.',
  wirdNeuGezeichnet: 'Nouveau dessin en cours…',
  diagnose: 'Diagnostic',
  diagnoseFrei: 'Aucun constat dans ce fichier.',
  feldartFehlt:
    'Ce champ n’a pas de « type: ». MSE2 arrête la lecture du paquet ici ; cette app le traite comme un champ de texte.',
  feldartUnbekannt:
    'MSE2 ne connaît pas le type de champ « {0} » et arrête la lecture du paquet ici. Cette app traite le champ comme du texte. Types connus : {1}.',
  diagnoseStelle: 'Ligne {zeile}, colonne {spalte}',
  zurStelle: 'Aller à l’emplacement',
  herkunft: 'D’où vient un nom',
  herkunftKeine:
    'Place le curseur sur un nom — tu verras ici d’où il vient et qui l’utilise.',
  herkunftUnbekannt:
    '« {name} » n’est pas défini par ce paquet — le nom vient de la bibliothèque ou d’un paquet inclus.',
  herkunftStelle: '{art} dans {datei}, ligne {zeile}',
  zurDefinition: 'Aller à la définition',
  verwendungen: 'Références',
  verwendungenKeine: 'Aucun autre endroit ne nomme ceci.',
  verwendungLaden: 'Recherche des références…',
  symbolarten: {
    'card field': 'Champ de carte',
    'set field': 'Champ de set',
    'styling field': 'Option de style',
    'extra card field': 'Champ supplémentaire',
    script: 'Nom de script',
  } as Record<string, string>,
  symbolartSonst: 'Défini ici',
  aenderungen: 'Modifications avant l’enregistrement',
  aenderungenZeigen: 'Afficher les modifications',
  aenderungenVerbergen: 'Masquer les modifications',
  aenderungenKeine: 'Aucune modification par rapport à l’état lu.',
  aenderungenZusammenfassung: '{neu} ligne(s) ajoutée(s), {weg} supprimée(s).',
  aenderungenAbschnitt: 'Modification {nummer}',
  zeileWeg: 'supprimée',
  zeileNeu: 'ajoutée',
  kopieHinweis: 'Origine du paquet',
  kopieren: 'Créer une copie personnelle',
  kopieMitgeliefert:
    'Ce paquet est fourni avec l’application et ne peut pas être modifié. Crées-en une copie personnelle — elle apparaîtra parmi tes paquets, et l’original reste tel quel.',
  kopieFremd:
    'Ce paquet vient de l’extérieur. Une modification ici compte comme modification locale et s’oppose à la nouvelle version lors du prochain téléchargement. Crée une copie personnelle si elle doit rester.',
  kopieOriginal:
    'L’original reste. Tu travailles sur une copie — {original} lui-même est inchangé.',
  dateiaktionen: 'Fichiers du paquet',
  dateiNeu: 'Nouveau fichier',
  dateiName: 'Nom du nouveau fichier',
  dateiAnlegen: 'Créer',
  dateiAbbrechen: 'Annuler',
  dateiHinzufuegen: 'Ajouter une image ou une police…',
  dateiUmbenennen: 'Renommer',
  dateiNeuerName: 'Nouveau nom',
  dateiUmbenennenJa: 'Renommer',
  dateiLoeschen: 'Supprimer…',
  dateiLoeschenFrage:
    'Supprimer ? Le fichier se trouvera ensuite dans la sauvegarde à côté du paquet.',
  dateiLoeschenJa: 'Oui, supprimer',
  dateiErstSichern: 'Enregistre ou abandonne d’abord.',
  beiwerk: 'Annexe',
  beiwerkHinweis:
    'Pas du texte — le paquet utilise ce fichier tel quel. On peut le renommer et le supprimer ici.',
  beiwerkZeile: 'Annexes ({anzahl})',
  weitereZeigen: 'Afficher plus ({gezeigt} sur {gesamt})',
} as const
