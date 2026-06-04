// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Polish (`pl`).
class AppLocalizationsPl extends AppLocalizations {
  AppLocalizationsPl([String locale = 'pl']) : super(locale);

  @override
  String get appName => 'BeeTrack';

  @override
  String get generalSave => 'Zapisz';

  @override
  String get generalCancel => 'Anuluj';

  @override
  String get generalDelete => 'Usuń';

  @override
  String get generalEdit => 'Edytuj';

  @override
  String get generalConfirm => 'Potwierdź';

  @override
  String get generalError => 'Wystąpił błąd. Spróbuj ponownie.';

  @override
  String get generalLoading => 'Ładowanie...';

  @override
  String get authEmail => 'E-mail';

  @override
  String get authPassword => 'Hasło';

  @override
  String get authName => 'Imię i nazwisko';

  @override
  String get authLogin => 'Zaloguj się';

  @override
  String get authRegister => 'Zarejestruj się';

  @override
  String get authLogout => 'Wyloguj się';

  @override
  String get authNoAccount => 'Nie masz konta? Zarejestruj się';

  @override
  String get authHaveAccount => 'Masz już konto? Zaloguj się';

  @override
  String get authInvalidEmail => 'Nieprawidłowy adres e-mail';

  @override
  String get authWeakPassword => 'Hasło musi mieć co najmniej 8 znaków';

  @override
  String get authInvalidCredentials => 'Nieprawidłowy e-mail lub hasło';

  @override
  String get authEmailTaken => 'Ten adres e-mail jest już zajęty';

  @override
  String get roleOwner => 'Właściciel';

  @override
  String get roleMember => 'Członek';

  @override
  String get apiaryTitle => 'Pasieki';

  @override
  String get apiaryAdd => 'Dodaj pasiekę';

  @override
  String get apiaryName => 'Pasieka';

  @override
  String get apiaryLatitude => 'Szerokość geograficzna';

  @override
  String get apiaryLongitude => 'Długość geograficzna';

  @override
  String get apiaryGpsUnavailable => 'GPS niedostępny na tym urządzeniu';

  @override
  String get apiaryLocation => 'Lokalizacja (opcjonalnie)';

  @override
  String get apiaryGridRows => 'Wiersze siatki';

  @override
  String get apiaryGridCols => 'Kolumny siatki';

  @override
  String get apiaryEmpty => 'Nie masz jeszcze żadnych pasiek';

  @override
  String get apiaryEdit => 'Edytuj pasiekę';

  @override
  String get apiaryDeleteConfirm => 'Usunąć pasiekę?';

  @override
  String get apiaryDeleteWarning =>
      'Ta operacja trwale usunie pasiekę i wszystkie jej dane.';

  @override
  String get apiaryGridTooSmall =>
      'Nowa siatka jest za mała, aby pomieścić wszystkie ule.';

  @override
  String get hiveTitle => 'Ule';

  @override
  String hiveCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count uli',
      many: '$count uli',
      few: '$count ule',
      one: '1 ul',
    );
    return '$_temp0';
  }

  @override
  String get hiveAdd => 'Dodaj ul';

  @override
  String get hiveName => 'Nazwa ula';

  @override
  String get hiveType => 'Typ ula';

  @override
  String get hiveActive => 'Aktywny';

  @override
  String get hiveInactive => 'Nieaktywny';

  @override
  String get hiveEmpty => 'Brak uli w tej pasiece';

  @override
  String get hiveEdit => 'Edytuj ul';

  @override
  String get hiveDeleteConfirm => 'Usunąć ul?';

  @override
  String get hiveDeleteWarning => 'Ta operacja trwale usunie ul.';

  @override
  String hiveDefaultName(int index) {
    return 'Ul $index';
  }

  @override
  String get hiveStatus => 'Status';

  @override
  String get hiveDetailInspections => 'Inspekcje';

  @override
  String get hiveDetailNoInspections => 'Brak inspekcji';

  @override
  String get hiveDetailAddInspection => 'Dodaj inspekcję';

  @override
  String get hiveDetailTreatments => 'Leczenia';

  @override
  String get hiveDetailNoTreatments => 'Brak aktywnych leczeń';

  @override
  String get hiveDetailLogTreatment => 'Dodaj leczenie';

  @override
  String get hiveDetailHarvests => 'Zbiory';

  @override
  String get hiveDetailNoHarvests => 'Brak zbiorów';

  @override
  String get hiveDetailLogHarvest => 'Dodaj zbiór';

  @override
  String get generalRequired => 'Wymagane';
}
