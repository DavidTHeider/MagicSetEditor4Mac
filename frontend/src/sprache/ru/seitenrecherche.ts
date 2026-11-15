// The research texts — work package A16 of phase 5.3
// (#1515).
export const recherche = {
  rechercheTitel: 'Поиск информации',
  rechercheErklaerung:
    'Ищет в текстах карт, разъяснениях и правилах на этом компьютере — полнотекстовый поиск и векторы вместе. Модель эмбеддингов участвует, локально; модель, пишущая тексты, не запрашивается. Ничто не покидает этот компьютер, и ничто не меняется.',

  rechercheFall: 'Для чего',
  rechercheFallKarten: 'Карты',
  rechercheFallRegeln: 'Правила',
  rechercheFallRulings: 'Разъяснения',

  rechercheFrage: 'Вопрос',
  rechercheSuchen: 'Найти',
  rechercheSprache: 'Язык',
  rechercheAlleSprachen: 'Любой',

  rechercheModus: 'Искали',
  rechercheModusHybrid: 'полнотекстовый поиск и векторы вместе',
  rechercheModusLexikalisch: 'только полный текст — без модели эмбеддингов',
  rechercheModusOhneIndex: 'вообще не искали — поискового индекса нет',

  rechercheLeereNichtsGefunden: 'Попробуй другое слово?',
  rechercheLeereOhneIndex:
    'Его можно построить в разделе «Настройки › Диагностика».',
  rechercheLeereOhneModell:
    'Без модели эмбеддингов поиск находит слова, а не сходство.',
  rechercheLeereOhneQuelle:
    'Для этого вопроса нужен источник, которого нет в индексе.',
  rechercheLeereGesperrt: 'Доступ к этому источнику не предоставлен.',
  rechercheLeereBudget: 'Этот запуск достиг своего предела.',

  rechercheHerkunft:
    'Ревизия {revision}, действительна {datum}, получена {abgerufen}',
  rechercheWeitere: 'Ещё разделов той же карты: {anzahl}',
  rechercheOeffnen: 'Открыть карту',
  rechercheEinschraenkung: 'Не выполнено',
  rechercheStaende: 'Ревизии',
  rechercheIndex: 'Поисковый индекс',
  rechercheMehr: 'Загрузить ещё',

  aehnlichTitel: 'Похожие карты',
  aehnlichErklaerung:
    'Ищет по тексту этой карты, а не по её названию. Сама карта в список не попадает.',
  aehnlichSuchen: 'Найти похожие',
  aehnlichLeer: 'Пока ничего не искали.',

  freigabeTitel: 'Что получила бы модель',
  freigabeErklaerung:
    'Собирает то, что увидела бы модель, — и не отправляет. Кто потом запустит генерацию, тот знает, с чем именно она работала.',
  freigabeAufgabe: 'Задача',
  freigabeAnbieter: 'Провайдер',
  freigabeModell: 'Модель',
  freigabeFenster: 'Контекстное окно',
  freigabeReserve: 'Резерв для ответа',
  freigabeZeigen: 'Собрать',
  freigabeOrtLokal: 'Вычисляется на этом компьютере.',
  freigabeOrtCloud: 'Вычисляется у внешнего провайдера.',
  freigabeBudget: '{verbraucht} из {grenze} токенов, {reserve} в резерве',
  freigabeZaehlweise: 'оценка: {zaehlweise}',
  freigabeZieladresse: 'Целевой адрес',
  freigabeKategorien: 'Категории данных',
  freigabePrivat: 'Включено собственное содержимое.',
  freigabeOeffentlich: 'Только общедоступные источники.',
  freigabeBelege: 'Ссылки на источники',
  freigabeKeineBelege:
    'Нет ссылок на источники — собранная так, модель ничего бы не увидела.',
  freigabeVeraltet: 'Не активная ревизия',
  freigabeLaeuftAb: 'Действительно до {zeit}',

  // **Releasing your own collection** — A13.4
  // (#1452).
  freigabeEigeneKarten: 'Искать и среди моих карт',
  freigabeBestandErklaerung:
    'Без этой отметки твои собственные карты остаются в стороне — поиск тогда охватывает только общедоступные источники, и ответ об этом сообщает. Всё, что при этом передаётся, покидает этот компьютер, если модель работает в другом месте.',
}
