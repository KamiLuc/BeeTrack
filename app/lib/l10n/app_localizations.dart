import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_pl.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('pl'),
  ];

  /// No description provided for @appName.
  ///
  /// In pl, this message translates to:
  /// **'BeeTrack'**
  String get appName;

  /// No description provided for @generalSave.
  ///
  /// In pl, this message translates to:
  /// **'Zapisz'**
  String get generalSave;

  /// No description provided for @generalCancel.
  ///
  /// In pl, this message translates to:
  /// **'Anuluj'**
  String get generalCancel;

  /// No description provided for @generalDelete.
  ///
  /// In pl, this message translates to:
  /// **'Usuń'**
  String get generalDelete;

  /// No description provided for @generalEdit.
  ///
  /// In pl, this message translates to:
  /// **'Edytuj'**
  String get generalEdit;

  /// No description provided for @generalConfirm.
  ///
  /// In pl, this message translates to:
  /// **'Potwierdź'**
  String get generalConfirm;

  /// No description provided for @generalError.
  ///
  /// In pl, this message translates to:
  /// **'Wystąpił błąd. Spróbuj ponownie.'**
  String get generalError;

  /// No description provided for @generalLoading.
  ///
  /// In pl, this message translates to:
  /// **'Ładowanie...'**
  String get generalLoading;

  /// No description provided for @authEmail.
  ///
  /// In pl, this message translates to:
  /// **'E-mail'**
  String get authEmail;

  /// No description provided for @authPassword.
  ///
  /// In pl, this message translates to:
  /// **'Hasło'**
  String get authPassword;

  /// No description provided for @authName.
  ///
  /// In pl, this message translates to:
  /// **'Imię i nazwisko'**
  String get authName;

  /// No description provided for @authLogin.
  ///
  /// In pl, this message translates to:
  /// **'Zaloguj się'**
  String get authLogin;

  /// No description provided for @authRegister.
  ///
  /// In pl, this message translates to:
  /// **'Zarejestruj się'**
  String get authRegister;

  /// No description provided for @authLogout.
  ///
  /// In pl, this message translates to:
  /// **'Wyloguj się'**
  String get authLogout;

  /// No description provided for @authNoAccount.
  ///
  /// In pl, this message translates to:
  /// **'Nie masz konta? Zarejestruj się'**
  String get authNoAccount;

  /// No description provided for @authHaveAccount.
  ///
  /// In pl, this message translates to:
  /// **'Masz już konto? Zaloguj się'**
  String get authHaveAccount;

  /// No description provided for @authInvalidEmail.
  ///
  /// In pl, this message translates to:
  /// **'Nieprawidłowy adres e-mail'**
  String get authInvalidEmail;

  /// No description provided for @authWeakPassword.
  ///
  /// In pl, this message translates to:
  /// **'Hasło musi mieć co najmniej 8 znaków'**
  String get authWeakPassword;

  /// No description provided for @authInvalidCredentials.
  ///
  /// In pl, this message translates to:
  /// **'Nieprawidłowy e-mail lub hasło'**
  String get authInvalidCredentials;

  /// No description provided for @authEmailTaken.
  ///
  /// In pl, this message translates to:
  /// **'Ten adres e-mail jest już zajęty'**
  String get authEmailTaken;

  /// No description provided for @roleOwner.
  ///
  /// In pl, this message translates to:
  /// **'Właściciel'**
  String get roleOwner;

  /// No description provided for @roleMember.
  ///
  /// In pl, this message translates to:
  /// **'Członek'**
  String get roleMember;

  /// No description provided for @apiaryTitle.
  ///
  /// In pl, this message translates to:
  /// **'Pasieki'**
  String get apiaryTitle;

  /// No description provided for @apiaryAdd.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj pasiekę'**
  String get apiaryAdd;

  /// No description provided for @apiaryName.
  ///
  /// In pl, this message translates to:
  /// **'Pasieka'**
  String get apiaryName;

  /// No description provided for @apiaryLatitude.
  ///
  /// In pl, this message translates to:
  /// **'Szerokość geograficzna'**
  String get apiaryLatitude;

  /// No description provided for @apiaryLongitude.
  ///
  /// In pl, this message translates to:
  /// **'Długość geograficzna'**
  String get apiaryLongitude;

  /// No description provided for @apiaryGpsUnavailable.
  ///
  /// In pl, this message translates to:
  /// **'GPS niedostępny na tym urządzeniu'**
  String get apiaryGpsUnavailable;

  /// No description provided for @apiaryLocation.
  ///
  /// In pl, this message translates to:
  /// **'Lokalizacja (opcjonalnie)'**
  String get apiaryLocation;

  /// No description provided for @apiaryGridRows.
  ///
  /// In pl, this message translates to:
  /// **'Wiersze siatki'**
  String get apiaryGridRows;

  /// No description provided for @apiaryGridCols.
  ///
  /// In pl, this message translates to:
  /// **'Kolumny siatki'**
  String get apiaryGridCols;

  /// No description provided for @apiaryEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Nie masz jeszcze żadnych pasiek'**
  String get apiaryEmpty;

  /// No description provided for @apiaryEdit.
  ///
  /// In pl, this message translates to:
  /// **'Edytuj pasiekę'**
  String get apiaryEdit;

  /// No description provided for @apiaryDeleteConfirm.
  ///
  /// In pl, this message translates to:
  /// **'Usunąć pasiekę?'**
  String get apiaryDeleteConfirm;

  /// No description provided for @apiaryDeleteWarning.
  ///
  /// In pl, this message translates to:
  /// **'Ta operacja trwale usunie pasiekę i wszystkie jej dane.'**
  String get apiaryDeleteWarning;

  /// No description provided for @apiaryGridTooSmall.
  ///
  /// In pl, this message translates to:
  /// **'Nowa siatka jest za mała, aby pomieścić wszystkie ule.'**
  String get apiaryGridTooSmall;

  /// No description provided for @hiveTitle.
  ///
  /// In pl, this message translates to:
  /// **'Ule'**
  String get hiveTitle;

  /// No description provided for @hiveCount.
  ///
  /// In pl, this message translates to:
  /// **'{count, plural, =1{1 ul} few{{count} ule} many{{count} uli} other{{count} uli}}'**
  String hiveCount(int count);

  /// No description provided for @hiveAdd.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj ul'**
  String get hiveAdd;

  /// No description provided for @hiveName.
  ///
  /// In pl, this message translates to:
  /// **'Nazwa ula'**
  String get hiveName;

  /// No description provided for @hiveType.
  ///
  /// In pl, this message translates to:
  /// **'Typ ula'**
  String get hiveType;

  /// No description provided for @hiveActive.
  ///
  /// In pl, this message translates to:
  /// **'Aktywny'**
  String get hiveActive;

  /// No description provided for @hiveInactive.
  ///
  /// In pl, this message translates to:
  /// **'Nieaktywny'**
  String get hiveInactive;

  /// No description provided for @hiveEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Brak uli w tej pasiece'**
  String get hiveEmpty;

  /// No description provided for @hiveEdit.
  ///
  /// In pl, this message translates to:
  /// **'Edytuj ul'**
  String get hiveEdit;

  /// No description provided for @hiveDeleteConfirm.
  ///
  /// In pl, this message translates to:
  /// **'Usunąć ul?'**
  String get hiveDeleteConfirm;

  /// No description provided for @hiveDeleteWarning.
  ///
  /// In pl, this message translates to:
  /// **'Ta operacja trwale usunie ul.'**
  String get hiveDeleteWarning;

  /// No description provided for @hiveDefaultName.
  ///
  /// In pl, this message translates to:
  /// **'Ul {index}'**
  String hiveDefaultName(int index);

  /// No description provided for @hiveStatus.
  ///
  /// In pl, this message translates to:
  /// **'Status'**
  String get hiveStatus;

  /// No description provided for @hiveDetailInspections.
  ///
  /// In pl, this message translates to:
  /// **'Inspekcje'**
  String get hiveDetailInspections;

  /// No description provided for @hiveDetailNoInspections.
  ///
  /// In pl, this message translates to:
  /// **'Brak inspekcji'**
  String get hiveDetailNoInspections;

  /// No description provided for @hiveDetailAddInspection.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj inspekcję'**
  String get hiveDetailAddInspection;

  /// No description provided for @hiveDetailTreatments.
  ///
  /// In pl, this message translates to:
  /// **'Leczenia'**
  String get hiveDetailTreatments;

  /// No description provided for @hiveDetailNoTreatments.
  ///
  /// In pl, this message translates to:
  /// **'Brak aktywnych leczeń'**
  String get hiveDetailNoTreatments;

  /// No description provided for @hiveDetailLogTreatment.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj leczenie'**
  String get hiveDetailLogTreatment;

  /// No description provided for @hiveDetailHarvests.
  ///
  /// In pl, this message translates to:
  /// **'Zbiory'**
  String get hiveDetailHarvests;

  /// No description provided for @hiveDetailNoHarvests.
  ///
  /// In pl, this message translates to:
  /// **'Brak zbiorów'**
  String get hiveDetailNoHarvests;

  /// No description provided for @hiveDetailLogHarvest.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj zbiór'**
  String get hiveDetailLogHarvest;

  /// No description provided for @generalRequired.
  ///
  /// In pl, this message translates to:
  /// **'Wymagane'**
  String get generalRequired;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'pl'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'pl':
      return AppLocalizationsPl();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
