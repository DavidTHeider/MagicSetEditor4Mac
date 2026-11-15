// The **data hygiene** texts — work item O8 of phase 5.2
// (#1432).

export const datenpflege = {
  titel: 'Наведение порядка',
  hinweis:
    'Со временем в библиотеке накапливаются строки, которые больше ни на что не указывают. ' +
    'Здесь показано, что это такое — и что из этого можно безопасно убрать.',

  nachsehen: 'Посмотреть',
  nachsehenLaeuft: 'Идёт проверка …',
  sauber: 'Ничего не найдено. Библиотека в порядке.',
  gueltigbis: 'Этот обзор действителен до {zeit}.',

  arten: {
    kartenwerteOhneKarte: 'Значения полей без своей карты',
    setangabenOhneSet: 'Данные набора без своего набора',
    suchzeilenOhneKarte: 'Поисковые строки без своей карты',
    deckplaetzeOhneDeck: 'Ячейки колоды без своей колоды',
    zuordnungenOhneExemplar: 'Привязки к колодам без своего экземпляра',
    verlaufOhneSet: 'Записи истории для несуществующих наборов',
    abgelaufeneNachweise: 'Истёкшие подтверждения',
    abgelaufeneToken: 'Истёкшие токены доступа',
    moeglicheDubletten: 'Похоже на дубликаты',
    offeneRueckkehren: 'Незавершённые восстановления',
  } as Record<string, string>,

  klassen: {
    sicher: 'Можно безопасно убрать',
    entscheidung: 'Решаешь ты',
    nurInformieren: 'Не будет затронуто',
  } as Record<string, string>,

  klassenhinweise: {
    sicher:
      'Эти строки ни на что не указывают и ничего не содержат. Они выбраны заранее.',
    entscheidung:
      'Это похоже на остаток — и может оказаться записью. Посмотри ' +
      'примеры, прежде чем выбирать это.',
    nurInformieren:
      'Это не убирается. Одинаковые номера карт — не дубликаты: ' +
      'состояние, язык, отделка, приобретение и место хранения их ' +
      'различают, — а незавершённое восстановление является записью.',
  } as Record<string, string>,

  anzahl: 'Мест в библиотеке: {anzahl}',
  beispieleTitel: 'Например',

  aufraeumen: 'Убрать выбранное',
  aufraeumenLaeuft: 'Идёт уборка …',
  aufraeumenHinweis:
    'Сначала создаётся резервная копия библиотеки. После этого проверка выполняется снова.',

  bilanzTitel: 'Что произошло',
  gesichert: 'Резервная копия сохранена в {ort}',

  ausgaenge: {
    erledigt: 'Убрано',
    konflikt: 'Изменилось с момента обзора — не затронуто',
    schonWeg: 'Уже отсутствовало',
    uebersprungen: 'Пропущено',
    gescheitert: 'Не удалось',
    bereit: 'Готово',
  } as Record<string, string>,

  bilanzZeile: '{art}: {ausgang} ({anzahl})',
  bilanzStaende: 'Было находок: {vorher}, стало: {nachher}.',
}
