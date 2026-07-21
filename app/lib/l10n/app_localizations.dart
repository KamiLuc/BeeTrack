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

  /// No description provided for @profileLanguage.
  ///
  /// In pl, this message translates to:
  /// **'Język'**
  String get profileLanguage;

  /// No description provided for @profileLanguageEn.
  ///
  /// In pl, this message translates to:
  /// **'Angielski'**
  String get profileLanguageEn;

  /// No description provided for @profileLanguagePl.
  ///
  /// In pl, this message translates to:
  /// **'Polski'**
  String get profileLanguagePl;

  /// No description provided for @profileDisplayName.
  ///
  /// In pl, this message translates to:
  /// **'Nazwa wyświetlana'**
  String get profileDisplayName;

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

  /// No description provided for @generalClose.
  ///
  /// In pl, this message translates to:
  /// **'Zamknij'**
  String get generalClose;

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

  /// No description provided for @generalRetry.
  ///
  /// In pl, this message translates to:
  /// **'Spróbuj ponownie'**
  String get generalRetry;

  /// No description provided for @generalLoading.
  ///
  /// In pl, this message translates to:
  /// **'Ładowanie...'**
  String get generalLoading;

  /// No description provided for @deletePuzzlePrompt.
  ///
  /// In pl, this message translates to:
  /// **'Aby potwierdzić, rozwiąż:'**
  String get deletePuzzlePrompt;

  /// No description provided for @deletePuzzleWrong.
  ///
  /// In pl, this message translates to:
  /// **'Zła odpowiedź'**
  String get deletePuzzleWrong;

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

  /// No description provided for @authEmailNotVerified.
  ///
  /// In pl, this message translates to:
  /// **'Potwierdź swój adres e-mail przed zalogowaniem'**
  String get authEmailNotVerified;

  /// No description provided for @authCheckEmail.
  ///
  /// In pl, this message translates to:
  /// **'Sprawdź pocztę'**
  String get authCheckEmail;

  /// No description provided for @authCheckEmailMessage.
  ///
  /// In pl, this message translates to:
  /// **'Wysłaliśmy link weryfikacyjny na adres {email}. Sprawdź skrzynkę odbiorczą.'**
  String authCheckEmailMessage(String email);

  /// No description provided for @authResendEmail.
  ///
  /// In pl, this message translates to:
  /// **'Wyślij ponownie'**
  String get authResendEmail;

  /// No description provided for @authBackToLogin.
  ///
  /// In pl, this message translates to:
  /// **'Wróć do logowania'**
  String get authBackToLogin;

  /// No description provided for @authForgotPassword.
  ///
  /// In pl, this message translates to:
  /// **'Nie pamiętasz hasła?'**
  String get authForgotPassword;

  /// No description provided for @authForgotPasswordTitle.
  ///
  /// In pl, this message translates to:
  /// **'Resetowanie hasła'**
  String get authForgotPasswordTitle;

  /// No description provided for @authForgotPasswordSubtitle.
  ///
  /// In pl, this message translates to:
  /// **'Podaj swój e-mail, a wyślemy Ci link do resetowania hasła.'**
  String get authForgotPasswordSubtitle;

  /// No description provided for @authForgotPasswordSent.
  ///
  /// In pl, this message translates to:
  /// **'Sprawdź skrzynkę — link do resetowania hasła został wysłany.'**
  String get authForgotPasswordSent;

  /// No description provided for @authSendResetLink.
  ///
  /// In pl, this message translates to:
  /// **'Wyślij link'**
  String get authSendResetLink;

  /// No description provided for @authVerifyingEmail.
  ///
  /// In pl, this message translates to:
  /// **'Weryfikacja adresu e-mail...'**
  String get authVerifyingEmail;

  /// No description provided for @authEmailVerified.
  ///
  /// In pl, this message translates to:
  /// **'Adres e-mail zweryfikowany!'**
  String get authEmailVerified;

  /// No description provided for @authEmailVerifiedMessage.
  ///
  /// In pl, this message translates to:
  /// **'Twoje konto jest aktywne. Możesz się zalogować.'**
  String get authEmailVerifiedMessage;

  /// No description provided for @authVerificationFailed.
  ///
  /// In pl, this message translates to:
  /// **'Weryfikacja nie powiodła się'**
  String get authVerificationFailed;

  /// No description provided for @authVerificationFailedMessage.
  ///
  /// In pl, this message translates to:
  /// **'Link mógł wygasnąć lub już był użyty.'**
  String get authVerificationFailedMessage;

  /// No description provided for @authGoToLogin.
  ///
  /// In pl, this message translates to:
  /// **'Przejdź do logowania'**
  String get authGoToLogin;

  /// No description provided for @authNewPassword.
  ///
  /// In pl, this message translates to:
  /// **'Nowe hasło'**
  String get authNewPassword;

  /// No description provided for @authPasswordChanged.
  ///
  /// In pl, this message translates to:
  /// **'Hasło zostało zmienione!'**
  String get authPasswordChanged;

  /// No description provided for @authPasswordChangedMessage.
  ///
  /// In pl, this message translates to:
  /// **'Możesz teraz zalogować się nowym hasłem.'**
  String get authPasswordChangedMessage;

  /// No description provided for @authInvalidResetToken.
  ///
  /// In pl, this message translates to:
  /// **'Ten link wygasł lub już był użyty.'**
  String get authInvalidResetToken;

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

  /// No description provided for @invitationTitle.
  ///
  /// In pl, this message translates to:
  /// **'Zaproszenia'**
  String get invitationTitle;

  /// No description provided for @invitationMembers.
  ///
  /// In pl, this message translates to:
  /// **'Członkowie'**
  String get invitationMembers;

  /// No description provided for @invitationPending.
  ///
  /// In pl, this message translates to:
  /// **'Oczekujące zaproszenia'**
  String get invitationPending;

  /// No description provided for @invitationInvite.
  ///
  /// In pl, this message translates to:
  /// **'Zarządzaj członkami'**
  String get invitationInvite;

  /// No description provided for @invitationEmailHint.
  ///
  /// In pl, this message translates to:
  /// **'Adres e-mail'**
  String get invitationEmailHint;

  /// No description provided for @invitationSend.
  ///
  /// In pl, this message translates to:
  /// **'Wyślij zaproszenie'**
  String get invitationSend;

  /// No description provided for @invitationSentSuccess.
  ///
  /// In pl, this message translates to:
  /// **'Zaproszenie wysłane'**
  String get invitationSentSuccess;

  /// No description provided for @invitationAlreadyPending.
  ///
  /// In pl, this message translates to:
  /// **'Zaproszenie dla tego adresu e-mail już oczekuje'**
  String get invitationAlreadyPending;

  /// No description provided for @invitationAlreadyMember.
  ///
  /// In pl, this message translates to:
  /// **'Ten użytkownik jest już członkiem'**
  String get invitationAlreadyMember;

  /// No description provided for @invitationCannotInviteSelf.
  ///
  /// In pl, this message translates to:
  /// **'Nie możesz zaprosić siebie'**
  String get invitationCannotInviteSelf;

  /// No description provided for @invitationUserNotFound.
  ///
  /// In pl, this message translates to:
  /// **'Nie znaleziono konta dla tego adresu e-mail'**
  String get invitationUserNotFound;

  /// No description provided for @invitationNoMembers.
  ///
  /// In pl, this message translates to:
  /// **'Brak członków'**
  String get invitationNoMembers;

  /// No description provided for @invitationNoPending.
  ///
  /// In pl, this message translates to:
  /// **'Brak oczekujących zaproszeń'**
  String get invitationNoPending;

  /// No description provided for @invitationRemove.
  ///
  /// In pl, this message translates to:
  /// **'Usuń'**
  String get invitationRemove;

  /// No description provided for @invitationAccept.
  ///
  /// In pl, this message translates to:
  /// **'Akceptuj'**
  String get invitationAccept;

  /// No description provided for @invitationDecline.
  ///
  /// In pl, this message translates to:
  /// **'Odrzuć'**
  String get invitationDecline;

  /// No description provided for @invitationFrom.
  ///
  /// In pl, this message translates to:
  /// **'z {apiary} od {name}'**
  String invitationFrom(String apiary, String name);

  /// No description provided for @invitationBadgeTooltip.
  ///
  /// In pl, this message translates to:
  /// **'Oczekujące zaproszenia'**
  String get invitationBadgeTooltip;

  /// No description provided for @leaveApiary.
  ///
  /// In pl, this message translates to:
  /// **'Opuść pasiekę'**
  String get leaveApiary;

  /// No description provided for @leaveApiaryConfirm.
  ///
  /// In pl, this message translates to:
  /// **'Opuścić pasiekę?'**
  String get leaveApiaryConfirm;

  /// No description provided for @leaveApiaryWarning.
  ///
  /// In pl, this message translates to:
  /// **'Utracisz dostęp do tej pasieki.'**
  String get leaveApiaryWarning;

  /// No description provided for @apiaryTitle.
  ///
  /// In pl, this message translates to:
  /// **'Pasieki'**
  String get apiaryTitle;

  /// No description provided for @marketplaceTitle.
  ///
  /// In pl, this message translates to:
  /// **'Ogłoszenia'**
  String get marketplaceTitle;

  /// No description provided for @marketplaceComingSoon.
  ///
  /// In pl, this message translates to:
  /// **'Wkrótce'**
  String get marketplaceComingSoon;

  /// No description provided for @marketplaceSearchHint.
  ///
  /// In pl, this message translates to:
  /// **'Szukaj ogłoszeń'**
  String get marketplaceSearchHint;

  /// No description provided for @marketplaceEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Brak ogłoszeń'**
  String get marketplaceEmpty;

  /// No description provided for @marketplaceResultsCount.
  ///
  /// In pl, this message translates to:
  /// **'{loaded}/{total}'**
  String marketplaceResultsCount(int loaded, int total);

  /// No description provided for @marketplaceMapTooltip.
  ///
  /// In pl, this message translates to:
  /// **'Mapa ogłoszeń'**
  String get marketplaceMapTooltip;

  /// No description provided for @marketplaceMapTitle.
  ///
  /// In pl, this message translates to:
  /// **'Mapa ogłoszeń'**
  String get marketplaceMapTitle;

  /// No description provided for @marketplaceMapEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Brak ogłoszeń z lokalizacją pasującą do filtrów'**
  String get marketplaceMapEmpty;

  /// No description provided for @marketplacePriceMinHint.
  ///
  /// In pl, this message translates to:
  /// **'Cena od'**
  String get marketplacePriceMinHint;

  /// No description provided for @marketplacePriceMaxHint.
  ///
  /// In pl, this message translates to:
  /// **'Cena do'**
  String get marketplacePriceMaxHint;

  /// No description provided for @marketplaceFiltersButton.
  ///
  /// In pl, this message translates to:
  /// **'Filtry'**
  String get marketplaceFiltersButton;

  /// No description provided for @marketplaceClearFilters.
  ///
  /// In pl, this message translates to:
  /// **'Wyczyść filtry'**
  String get marketplaceClearFilters;

  /// No description provided for @marketplacePostedWithinAny.
  ///
  /// In pl, this message translates to:
  /// **'Dowolny czas'**
  String get marketplacePostedWithinAny;

  /// No description provided for @marketplacePostedWithinToday.
  ///
  /// In pl, this message translates to:
  /// **'Dzisiaj'**
  String get marketplacePostedWithinToday;

  /// No description provided for @marketplacePostedWithin7Days.
  ///
  /// In pl, this message translates to:
  /// **'Ostatnie 7 dni'**
  String get marketplacePostedWithin7Days;

  /// No description provided for @marketplacePostedWithin14Days.
  ///
  /// In pl, this message translates to:
  /// **'Ostatnie 14 dni'**
  String get marketplacePostedWithin14Days;

  /// No description provided for @marketplacePostedWithin30Days.
  ///
  /// In pl, this message translates to:
  /// **'Ostatnie 30 dni'**
  String get marketplacePostedWithin30Days;

  /// No description provided for @marketplaceDistanceLabel.
  ///
  /// In pl, this message translates to:
  /// **'Odległość'**
  String get marketplaceDistanceLabel;

  /// No description provided for @marketplaceGpsUnavailable.
  ///
  /// In pl, this message translates to:
  /// **'GPS niedostępny na tym urządzeniu'**
  String get marketplaceGpsUnavailable;

  /// No description provided for @marketplaceDistanceAny.
  ///
  /// In pl, this message translates to:
  /// **'Dowolna odległość'**
  String get marketplaceDistanceAny;

  /// No description provided for @marketplaceDistance5Km.
  ///
  /// In pl, this message translates to:
  /// **'Do 5 km'**
  String get marketplaceDistance5Km;

  /// No description provided for @marketplaceDistance10Km.
  ///
  /// In pl, this message translates to:
  /// **'Do 10 km'**
  String get marketplaceDistance10Km;

  /// No description provided for @marketplaceDistance25Km.
  ///
  /// In pl, this message translates to:
  /// **'Do 25 km'**
  String get marketplaceDistance25Km;

  /// No description provided for @marketplaceDistance50Km.
  ///
  /// In pl, this message translates to:
  /// **'Do 50 km'**
  String get marketplaceDistance50Km;

  /// No description provided for @marketplaceDistance100Km.
  ///
  /// In pl, this message translates to:
  /// **'Do 100 km'**
  String get marketplaceDistance100Km;

  /// No description provided for @marketplaceApiaryFilterLabel.
  ///
  /// In pl, this message translates to:
  /// **'Tylko ogłoszenia z powiązaną pasieką'**
  String get marketplaceApiaryFilterLabel;

  /// No description provided for @marketplaceDistanceAway.
  ///
  /// In pl, this message translates to:
  /// **'{km} km stąd'**
  String marketplaceDistanceAway(String km);

  /// No description provided for @marketplacePriceOnRequest.
  ///
  /// In pl, this message translates to:
  /// **'Cena do negocjacji'**
  String get marketplacePriceOnRequest;

  /// No description provided for @marketplacePriceFree.
  ///
  /// In pl, this message translates to:
  /// **'Za darmo'**
  String get marketplacePriceFree;

  /// No description provided for @marketplaceCategoryAll.
  ///
  /// In pl, this message translates to:
  /// **'Wszystkie'**
  String get marketplaceCategoryAll;

  /// No description provided for @marketplaceCategoryHoney.
  ///
  /// In pl, this message translates to:
  /// **'Miód'**
  String get marketplaceCategoryHoney;

  /// No description provided for @marketplaceCategoryPollen.
  ///
  /// In pl, this message translates to:
  /// **'Pyłek'**
  String get marketplaceCategoryPollen;

  /// No description provided for @marketplaceCategoryBeeColonies.
  ///
  /// In pl, this message translates to:
  /// **'Rodziny pszczele'**
  String get marketplaceCategoryBeeColonies;

  /// No description provided for @marketplaceCategoryQueenBees.
  ///
  /// In pl, this message translates to:
  /// **'Matki pszczele'**
  String get marketplaceCategoryQueenBees;

  /// No description provided for @marketplaceCategoryBeehives.
  ///
  /// In pl, this message translates to:
  /// **'Ule'**
  String get marketplaceCategoryBeehives;

  /// No description provided for @marketplaceCategoryEquipment.
  ///
  /// In pl, this message translates to:
  /// **'Sprzęt'**
  String get marketplaceCategoryEquipment;

  /// No description provided for @marketplaceCategoryExtractionEquipment.
  ///
  /// In pl, this message translates to:
  /// **'Sprzęt do wirowania miodu'**
  String get marketplaceCategoryExtractionEquipment;

  /// No description provided for @marketplaceCategoryFeed.
  ///
  /// In pl, this message translates to:
  /// **'Pokarm dla pszczół'**
  String get marketplaceCategoryFeed;

  /// No description provided for @marketplaceCategorySupplies.
  ///
  /// In pl, this message translates to:
  /// **'Zaopatrzenie'**
  String get marketplaceCategorySupplies;

  /// No description provided for @marketplaceCategoryWaxFoundation.
  ///
  /// In pl, this message translates to:
  /// **'Węza'**
  String get marketplaceCategoryWaxFoundation;

  /// No description provided for @marketplaceCategoryBeeswax.
  ///
  /// In pl, this message translates to:
  /// **'Wosk pszczeli'**
  String get marketplaceCategoryBeeswax;

  /// No description provided for @marketplaceCategoryPropolis.
  ///
  /// In pl, this message translates to:
  /// **'Propolis'**
  String get marketplaceCategoryPropolis;

  /// No description provided for @marketplaceCategoryServices.
  ///
  /// In pl, this message translates to:
  /// **'Usługi'**
  String get marketplaceCategoryServices;

  /// No description provided for @marketplaceCategoryOther.
  ///
  /// In pl, this message translates to:
  /// **'Inne'**
  String get marketplaceCategoryOther;

  /// No description provided for @marketplaceFavoriteAdd.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj do ulubionych'**
  String get marketplaceFavoriteAdd;

  /// No description provided for @marketplaceFavoriteRemove.
  ///
  /// In pl, this message translates to:
  /// **'Usuń z ulubionych'**
  String get marketplaceFavoriteRemove;

  /// No description provided for @marketplaceDescriptionLabel.
  ///
  /// In pl, this message translates to:
  /// **'Opis'**
  String get marketplaceDescriptionLabel;

  /// No description provided for @marketplaceContactLabel.
  ///
  /// In pl, this message translates to:
  /// **'Kontakt'**
  String get marketplaceContactLabel;

  /// No description provided for @marketplaceCallButton.
  ///
  /// In pl, this message translates to:
  /// **'Zadzwoń'**
  String get marketplaceCallButton;

  /// No description provided for @marketplaceWriteButton.
  ///
  /// In pl, this message translates to:
  /// **'Napisz'**
  String get marketplaceWriteButton;

  /// No description provided for @marketplaceApiaryLabel.
  ///
  /// In pl, this message translates to:
  /// **'Pasieka powiązana z ogłoszeniem'**
  String get marketplaceApiaryLabel;

  /// No description provided for @marketplaceQuantityLabel.
  ///
  /// In pl, this message translates to:
  /// **'Ilość'**
  String get marketplaceQuantityLabel;

  /// No description provided for @marketplacePostedOn.
  ///
  /// In pl, this message translates to:
  /// **'Dodano {date}'**
  String marketplacePostedOn(String date);

  /// No description provided for @marketplaceCreateScreenTitle.
  ///
  /// In pl, this message translates to:
  /// **'Nowe ogłoszenie'**
  String get marketplaceCreateScreenTitle;

  /// No description provided for @marketplaceFieldTitle.
  ///
  /// In pl, this message translates to:
  /// **'Tytuł'**
  String get marketplaceFieldTitle;

  /// No description provided for @marketplaceFieldTitleRequired.
  ///
  /// In pl, this message translates to:
  /// **'Tytuł jest wymagany'**
  String get marketplaceFieldTitleRequired;

  /// No description provided for @marketplaceFieldCategory.
  ///
  /// In pl, this message translates to:
  /// **'Kategoria'**
  String get marketplaceFieldCategory;

  /// No description provided for @marketplaceFieldCategoryRequired.
  ///
  /// In pl, this message translates to:
  /// **'Wybierz kategorię'**
  String get marketplaceFieldCategoryRequired;

  /// No description provided for @marketplaceFieldPrice.
  ///
  /// In pl, this message translates to:
  /// **'Cena'**
  String get marketplaceFieldPrice;

  /// No description provided for @marketplaceFieldPriceInvalid.
  ///
  /// In pl, this message translates to:
  /// **'Podaj prawidłową cenę'**
  String get marketplaceFieldPriceInvalid;

  /// No description provided for @marketplaceFieldPriceRequired.
  ///
  /// In pl, this message translates to:
  /// **'Cena jest wymagana'**
  String get marketplaceFieldPriceRequired;

  /// No description provided for @marketplaceFieldPriceTooLarge.
  ///
  /// In pl, this message translates to:
  /// **'Cena musi być mniejsza niż 100 000 000'**
  String get marketplaceFieldPriceTooLarge;

  /// No description provided for @marketplaceFieldAddress.
  ///
  /// In pl, this message translates to:
  /// **'Adres'**
  String get marketplaceFieldAddress;

  /// No description provided for @marketplaceFieldLatitude.
  ///
  /// In pl, this message translates to:
  /// **'Szerokość geograficzna'**
  String get marketplaceFieldLatitude;

  /// No description provided for @marketplaceFieldLongitude.
  ///
  /// In pl, this message translates to:
  /// **'Długość geograficzna'**
  String get marketplaceFieldLongitude;

  /// No description provided for @marketplaceLocationRequired.
  ///
  /// In pl, this message translates to:
  /// **'Wybierz lokalizację na mapie lub użyj GPS'**
  String get marketplaceLocationRequired;

  /// No description provided for @locationPickerTitle.
  ///
  /// In pl, this message translates to:
  /// **'Wybierz lokalizację'**
  String get locationPickerTitle;

  /// No description provided for @locationPickerHint.
  ///
  /// In pl, this message translates to:
  /// **'Dotknij mapę, aby wybrać lokalizację'**
  String get locationPickerHint;

  /// No description provided for @locationPickerGpsButton.
  ///
  /// In pl, this message translates to:
  /// **'GPS'**
  String get locationPickerGpsButton;

  /// No description provided for @locationPickerMapButton.
  ///
  /// In pl, this message translates to:
  /// **'Mapa'**
  String get locationPickerMapButton;

  /// No description provided for @marketplaceFieldPhone.
  ///
  /// In pl, this message translates to:
  /// **'Telefon'**
  String get marketplaceFieldPhone;

  /// No description provided for @marketplaceFieldPhoneInvalid.
  ///
  /// In pl, this message translates to:
  /// **'Podaj prawidłowy numer telefonu'**
  String get marketplaceFieldPhoneInvalid;

  /// No description provided for @marketplaceFieldEmail.
  ///
  /// In pl, this message translates to:
  /// **'E-mail'**
  String get marketplaceFieldEmail;

  /// No description provided for @marketplaceContactRequired.
  ///
  /// In pl, this message translates to:
  /// **'Podaj numer telefonu lub adres e-mail'**
  String get marketplaceContactRequired;

  /// No description provided for @marketplaceApiaryNone.
  ///
  /// In pl, this message translates to:
  /// **'Brak'**
  String get marketplaceApiaryNone;

  /// No description provided for @marketplacePhotosLabel.
  ///
  /// In pl, this message translates to:
  /// **'Zdjęcia'**
  String get marketplacePhotosLabel;

  /// No description provided for @marketplaceAddPhoto.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj zdjęcie'**
  String get marketplaceAddPhoto;

  /// No description provided for @marketplacePhotoSourceGallery.
  ///
  /// In pl, this message translates to:
  /// **'Wybierz z galerii'**
  String get marketplacePhotoSourceGallery;

  /// No description provided for @marketplacePhotoSourceCamera.
  ///
  /// In pl, this message translates to:
  /// **'Zrób zdjęcie'**
  String get marketplacePhotoSourceCamera;

  /// No description provided for @marketplaceEditScreenTitle.
  ///
  /// In pl, this message translates to:
  /// **'Edytuj ogłoszenie'**
  String get marketplaceEditScreenTitle;

  /// No description provided for @myListingsTitle.
  ///
  /// In pl, this message translates to:
  /// **'Moje ogłoszenia'**
  String get myListingsTitle;

  /// No description provided for @myListingsEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Nie masz jeszcze żadnych ogłoszeń'**
  String get myListingsEmpty;

  /// No description provided for @favoritesTitle.
  ///
  /// In pl, this message translates to:
  /// **'Ulubione'**
  String get favoritesTitle;

  /// No description provided for @favoritesEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Nie masz jeszcze żadnych ulubionych ogłoszeń'**
  String get favoritesEmpty;

  /// No description provided for @marketplaceHiddenBadge.
  ///
  /// In pl, this message translates to:
  /// **'Prywatne'**
  String get marketplaceHiddenBadge;

  /// No description provided for @marketplaceStatusPending.
  ///
  /// In pl, this message translates to:
  /// **'Oczekuje na weryfikację'**
  String get marketplaceStatusPending;

  /// No description provided for @marketplaceStatusApproved.
  ///
  /// In pl, this message translates to:
  /// **'Aktywne'**
  String get marketplaceStatusApproved;

  /// No description provided for @marketplaceStatusRejected.
  ///
  /// In pl, this message translates to:
  /// **'Odrzucone'**
  String get marketplaceStatusRejected;

  /// No description provided for @marketplaceStatusRemoved.
  ///
  /// In pl, this message translates to:
  /// **'Usunięte przez administratora'**
  String get marketplaceStatusRemoved;

  /// No description provided for @marketplaceHideListing.
  ///
  /// In pl, this message translates to:
  /// **'Ustaw jako prywatne'**
  String get marketplaceHideListing;

  /// No description provided for @marketplaceShowListing.
  ///
  /// In pl, this message translates to:
  /// **'Ustaw jako publiczne'**
  String get marketplaceShowListing;

  /// No description provided for @marketplaceDeleteConfirm.
  ///
  /// In pl, this message translates to:
  /// **'Usunąć to ogłoszenie?'**
  String get marketplaceDeleteConfirm;

  /// No description provided for @marketplaceDeleteWarning.
  ///
  /// In pl, this message translates to:
  /// **'Tej operacji nie można cofnąć.'**
  String get marketplaceDeleteWarning;

  /// No description provided for @apiaryMapTitle.
  ///
  /// In pl, this message translates to:
  /// **'Mapa pasiek'**
  String get apiaryMapTitle;

  /// No description provided for @apiaryMapTooltip.
  ///
  /// In pl, this message translates to:
  /// **'Pokaż na mapie'**
  String get apiaryMapTooltip;

  /// No description provided for @apiaryLocationTitle.
  ///
  /// In pl, this message translates to:
  /// **'Lokalizacja pasieki'**
  String get apiaryLocationTitle;

  /// No description provided for @apiaryCopy.
  ///
  /// In pl, this message translates to:
  /// **'Skopiuj pasiekę'**
  String get apiaryCopy;

  /// No description provided for @apiaryCopySuffix.
  ///
  /// In pl, this message translates to:
  /// **'kopia'**
  String get apiaryCopySuffix;

  /// No description provided for @apiaryCopyNewName.
  ///
  /// In pl, this message translates to:
  /// **'Nowa nazwa'**
  String get apiaryCopyNewName;

  /// No description provided for @apiaryCopied.
  ///
  /// In pl, this message translates to:
  /// **'Pasieka skopiowana'**
  String get apiaryCopied;

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

  /// No description provided for @apiaryNameRequired.
  ///
  /// In pl, this message translates to:
  /// **'Nazwa pasieki nie może być pusta'**
  String get apiaryNameRequired;

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

  /// No description provided for @apiaryGridHivesWillMove.
  ///
  /// In pl, this message translates to:
  /// **'{count, plural, =1{1 ul zostanie przeniesiony, aby zmieścić się w nowej siatce.} few{{count} ule zostaną przeniesione, aby zmieścić się w nowej siatce.} many{{count} uli zostanie przeniesionych, aby zmieścić się w nowej siatce.} other{{count} uli zostanie przeniesionych, aby zmieścić się w nowej siatce.}}'**
  String apiaryGridHivesWillMove(int count);

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

  /// No description provided for @hiveQueenless.
  ///
  /// In pl, this message translates to:
  /// **'Bezmateczny'**
  String get hiveQueenless;

  /// No description provided for @hiveReadyForHarvest.
  ///
  /// In pl, this message translates to:
  /// **'Gotowy do zbioru'**
  String get hiveReadyForHarvest;

  /// No description provided for @hiveSick.
  ///
  /// In pl, this message translates to:
  /// **'Chory'**
  String get hiveSick;

  /// No description provided for @hiveFilterTooltip.
  ///
  /// In pl, this message translates to:
  /// **'Filtruj ule'**
  String get hiveFilterTooltip;

  /// No description provided for @hiveListTooltip.
  ///
  /// In pl, this message translates to:
  /// **'Lista uli'**
  String get hiveListTooltip;

  /// No description provided for @apiaryCenterView.
  ///
  /// In pl, this message translates to:
  /// **'Wyśrodkuj widok'**
  String get apiaryCenterView;

  /// No description provided for @hiveDiseases.
  ///
  /// In pl, this message translates to:
  /// **'Choroby'**
  String get hiveDiseases;

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

  /// No description provided for @hiveDetailViewInspections.
  ///
  /// In pl, this message translates to:
  /// **'Pokaż wszystkie'**
  String get hiveDetailViewInspections;

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

  /// No description provided for @hiveDetailFeedings.
  ///
  /// In pl, this message translates to:
  /// **'Podkarmianie'**
  String get hiveDetailFeedings;

  /// No description provided for @hiveDetailNoFeedings.
  ///
  /// In pl, this message translates to:
  /// **'Brak podkarmiań'**
  String get hiveDetailNoFeedings;

  /// No description provided for @hiveDetailLogFeeding.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj podkarmianie'**
  String get hiveDetailLogFeeding;

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

  /// No description provided for @hiveChangeApiary.
  ///
  /// In pl, this message translates to:
  /// **'Zmień pasiekę'**
  String get hiveChangeApiary;

  /// No description provided for @hiveChangeApiaryTitle.
  ///
  /// In pl, this message translates to:
  /// **'Przenieś ul'**
  String get hiveChangeApiaryTitle;

  /// No description provided for @hiveChangeApiaryNoSpace.
  ///
  /// In pl, this message translates to:
  /// **'Docelowa pasieka nie ma wolnego miejsca'**
  String get hiveChangeApiaryNoSpace;

  /// No description provided for @hiveDuplicateName.
  ///
  /// In pl, this message translates to:
  /// **'Ul o tej nazwie już istnieje w tej pasiece'**
  String get hiveDuplicateName;

  /// No description provided for @generalRequired.
  ///
  /// In pl, this message translates to:
  /// **'Wymagane'**
  String get generalRequired;

  /// No description provided for @generalLoadMore.
  ///
  /// In pl, this message translates to:
  /// **'Załaduj więcej'**
  String get generalLoadMore;

  /// No description provided for @generalFieldTooLong.
  ///
  /// In pl, this message translates to:
  /// **'{field} może mieć maksymalnie {max} znaków'**
  String generalFieldTooLong(String field, int max);

  /// No description provided for @generalValueTooLarge.
  ///
  /// In pl, this message translates to:
  /// **'{field} musi być mniejsze lub równe {max}'**
  String generalValueTooLarge(String field, String max);

  /// No description provided for @generalPhotoTooLarge.
  ///
  /// In pl, this message translates to:
  /// **'Zdjęcie jest za duże. Maksymalny rozmiar to {max}.'**
  String generalPhotoTooLarge(String max);

  /// No description provided for @inspectionTitle.
  ///
  /// In pl, this message translates to:
  /// **'Inspekcje'**
  String get inspectionTitle;

  /// No description provided for @inspectionAdd.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj inspekcję'**
  String get inspectionAdd;

  /// No description provided for @inspectionEdit.
  ///
  /// In pl, this message translates to:
  /// **'Edytuj inspekcję'**
  String get inspectionEdit;

  /// No description provided for @inspectionDeleteConfirm.
  ///
  /// In pl, this message translates to:
  /// **'Usunąć inspekcję?'**
  String get inspectionDeleteConfirm;

  /// No description provided for @inspectionDeleteWarning.
  ///
  /// In pl, this message translates to:
  /// **'Ta operacja trwale usunie inspekcję.'**
  String get inspectionDeleteWarning;

  /// No description provided for @inspectionEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Brak inspekcji'**
  String get inspectionEmpty;

  /// No description provided for @inspectionDate.
  ///
  /// In pl, this message translates to:
  /// **'Data inspekcji'**
  String get inspectionDate;

  /// No description provided for @inspectionQueenSeen.
  ///
  /// In pl, this message translates to:
  /// **'Matka widziana'**
  String get inspectionQueenSeen;

  /// No description provided for @inspectionQueenStatusSeen.
  ///
  /// In pl, this message translates to:
  /// **'Matka widziana'**
  String get inspectionQueenStatusSeen;

  /// No description provided for @inspectionQueenStatusNotSeen.
  ///
  /// In pl, this message translates to:
  /// **'Matka niewidziana'**
  String get inspectionQueenStatusNotSeen;

  /// No description provided for @inspectionBroodPattern.
  ///
  /// In pl, this message translates to:
  /// **'Czerw'**
  String get inspectionBroodPattern;

  /// No description provided for @inspectionBroodExcellent.
  ///
  /// In pl, this message translates to:
  /// **'Dużo'**
  String get inspectionBroodExcellent;

  /// No description provided for @inspectionBroodGood.
  ///
  /// In pl, this message translates to:
  /// **'Średnio'**
  String get inspectionBroodGood;

  /// No description provided for @inspectionBroodPoor.
  ///
  /// In pl, this message translates to:
  /// **'Mało'**
  String get inspectionBroodPoor;

  /// No description provided for @inspectionBroodNone.
  ///
  /// In pl, this message translates to:
  /// **'Brak'**
  String get inspectionBroodNone;

  /// No description provided for @inspectionAggressiveness.
  ///
  /// In pl, this message translates to:
  /// **'Agresywność'**
  String get inspectionAggressiveness;

  /// No description provided for @inspectionAggressivenessCalm.
  ///
  /// In pl, this message translates to:
  /// **'Spokojne'**
  String get inspectionAggressivenessCalm;

  /// No description provided for @inspectionAggressivenessMild.
  ///
  /// In pl, this message translates to:
  /// **'Łagodne'**
  String get inspectionAggressivenessMild;

  /// No description provided for @inspectionAggressivenessAggressive.
  ///
  /// In pl, this message translates to:
  /// **'Agresywne'**
  String get inspectionAggressivenessAggressive;

  /// No description provided for @inspectionAggressivenessVeryAggressive.
  ///
  /// In pl, this message translates to:
  /// **'Bardzo agresywne'**
  String get inspectionAggressivenessVeryAggressive;

  /// No description provided for @inspectionFramesBrood.
  ///
  /// In pl, this message translates to:
  /// **'Ramki z czerwiem'**
  String get inspectionFramesBrood;

  /// No description provided for @inspectionFramesFeed.
  ///
  /// In pl, this message translates to:
  /// **'Ramki z pokarmem'**
  String get inspectionFramesFeed;

  /// No description provided for @inspectionFramesPollen.
  ///
  /// In pl, this message translates to:
  /// **'Ramki z pyłkiem'**
  String get inspectionFramesPollen;

  /// No description provided for @inspectionFramesAddedDrawn.
  ///
  /// In pl, this message translates to:
  /// **'Dodane puste ramki'**
  String get inspectionFramesAddedDrawn;

  /// No description provided for @inspectionFramesAddedFoundation.
  ///
  /// In pl, this message translates to:
  /// **'Dodana węza'**
  String get inspectionFramesAddedFoundation;

  /// No description provided for @inspectionFramesAddedBrood.
  ///
  /// In pl, this message translates to:
  /// **'Dodane ramki z czerwiem'**
  String get inspectionFramesAddedBrood;

  /// No description provided for @inspectionFramesAddedFeed.
  ///
  /// In pl, this message translates to:
  /// **'Dodane ramki z pokarmem'**
  String get inspectionFramesAddedFeed;

  /// No description provided for @inspectionFramesTakenDrawn.
  ///
  /// In pl, this message translates to:
  /// **'Zabrane puste ramki'**
  String get inspectionFramesTakenDrawn;

  /// No description provided for @inspectionFramesTakenFoundation.
  ///
  /// In pl, this message translates to:
  /// **'Zabrana węza'**
  String get inspectionFramesTakenFoundation;

  /// No description provided for @inspectionFramesTakenBrood.
  ///
  /// In pl, this message translates to:
  /// **'Zabrane ramki z czerwiem'**
  String get inspectionFramesTakenBrood;

  /// No description provided for @inspectionFramesTakenFeed.
  ///
  /// In pl, this message translates to:
  /// **'Zabrane ramki z pokarmem'**
  String get inspectionFramesTakenFeed;

  /// No description provided for @inspectionQueenCellsCount.
  ///
  /// In pl, this message translates to:
  /// **'Mateczniki'**
  String get inspectionQueenCellsCount;

  /// No description provided for @inspectionQueenAdded.
  ///
  /// In pl, this message translates to:
  /// **'Poddano matkę'**
  String get inspectionQueenAdded;

  /// No description provided for @inspectionSectionObservations.
  ///
  /// In pl, this message translates to:
  /// **'Obserwacje'**
  String get inspectionSectionObservations;

  /// No description provided for @inspectionSectionFrames.
  ///
  /// In pl, this message translates to:
  /// **'Ramki'**
  String get inspectionSectionFrames;

  /// No description provided for @inspectionSectionHealth.
  ///
  /// In pl, this message translates to:
  /// **'Zdrowie'**
  String get inspectionSectionHealth;

  /// No description provided for @inspectionSectionHiveState.
  ///
  /// In pl, this message translates to:
  /// **'Stan ula'**
  String get inspectionSectionHiveState;

  /// No description provided for @inspectionNotes.
  ///
  /// In pl, this message translates to:
  /// **'Notatki'**
  String get inspectionNotes;

  /// No description provided for @inspectionNote.
  ///
  /// In pl, this message translates to:
  /// **'Notatka'**
  String get inspectionNote;

  /// No description provided for @inspectionInspectedBy.
  ///
  /// In pl, this message translates to:
  /// **'Przez {name}'**
  String inspectionInspectedBy(String name);

  /// No description provided for @inspectionDiseases.
  ///
  /// In pl, this message translates to:
  /// **'Choroby'**
  String get inspectionDiseases;

  /// No description provided for @inspectionDiseaseVarroa.
  ///
  /// In pl, this message translates to:
  /// **'Warroza'**
  String get inspectionDiseaseVarroa;

  /// No description provided for @inspectionDiseaseNosema.
  ///
  /// In pl, this message translates to:
  /// **'Nosemoza'**
  String get inspectionDiseaseNosema;

  /// No description provided for @inspectionDiseaseDwv.
  ///
  /// In pl, this message translates to:
  /// **'Wirusy (DWV)'**
  String get inspectionDiseaseDwv;

  /// No description provided for @inspectionDiseaseAmericanFoulbrood.
  ///
  /// In pl, this message translates to:
  /// **'Zgnilec amerykański'**
  String get inspectionDiseaseAmericanFoulbrood;

  /// No description provided for @inspectionDiseaseChalkbrood.
  ///
  /// In pl, this message translates to:
  /// **'Grzybica wapienna'**
  String get inspectionDiseaseChalkbrood;

  /// No description provided for @inspectionDiseaseEuropeanFoulbrood.
  ///
  /// In pl, this message translates to:
  /// **'Zgnilec europejski'**
  String get inspectionDiseaseEuropeanFoulbrood;

  /// No description provided for @inspectionDiseaseLayingWorkers.
  ///
  /// In pl, this message translates to:
  /// **'Strutowienie rodziny'**
  String get inspectionDiseaseLayingWorkers;

  /// No description provided for @inspectionNotSet.
  ///
  /// In pl, this message translates to:
  /// **'Nie ustawiono'**
  String get inspectionNotSet;

  /// No description provided for @inspectionPhotos.
  ///
  /// In pl, this message translates to:
  /// **'Zdjęcia'**
  String get inspectionPhotos;

  /// No description provided for @inspectionPhotoSourceGallery.
  ///
  /// In pl, this message translates to:
  /// **'Galeria'**
  String get inspectionPhotoSourceGallery;

  /// No description provided for @inspectionPhotoSourceCamera.
  ///
  /// In pl, this message translates to:
  /// **'Aparat'**
  String get inspectionPhotoSourceCamera;

  /// No description provided for @inspectionAddPhoto.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj zdjęcie'**
  String get inspectionAddPhoto;

  /// No description provided for @inspectionNoPhotos.
  ///
  /// In pl, this message translates to:
  /// **'Brak zdjęć'**
  String get inspectionNoPhotos;

  /// No description provided for @inspectionDeletePhoto.
  ///
  /// In pl, this message translates to:
  /// **'Usunąć zdjęcie?'**
  String get inspectionDeletePhoto;

  /// No description provided for @inspectionDeletePhotoWarning.
  ///
  /// In pl, this message translates to:
  /// **'Ta operacja trwale usunie zdjęcie.'**
  String get inspectionDeletePhotoWarning;

  /// No description provided for @inspectionPhotoCount.
  ///
  /// In pl, this message translates to:
  /// **'{count, plural, =1{1 zdjęcie} few{{count} zdjęcia} other{{count} zdjęć}}'**
  String inspectionPhotoCount(int count);

  /// No description provided for @hiveTypeRequired.
  ///
  /// In pl, this message translates to:
  /// **'Typ ula jest wymagany'**
  String get hiveTypeRequired;

  /// No description provided for @treatmentTitle.
  ///
  /// In pl, this message translates to:
  /// **'Zabiegi'**
  String get treatmentTitle;

  /// No description provided for @treatmentAdd.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj zabieg'**
  String get treatmentAdd;

  /// No description provided for @treatmentEdit.
  ///
  /// In pl, this message translates to:
  /// **'Edytuj zabieg'**
  String get treatmentEdit;

  /// No description provided for @treatmentEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Brak zabiegów'**
  String get treatmentEmpty;

  /// No description provided for @treatmentDeleteConfirm.
  ///
  /// In pl, this message translates to:
  /// **'Usunąć zabieg?'**
  String get treatmentDeleteConfirm;

  /// No description provided for @treatmentDeleteWarning.
  ///
  /// In pl, this message translates to:
  /// **'Ta operacja trwale usunie wpis zabiegu.'**
  String get treatmentDeleteWarning;

  /// No description provided for @treatmentDate.
  ///
  /// In pl, this message translates to:
  /// **'Data zabiegu'**
  String get treatmentDate;

  /// No description provided for @treatmentMedicine.
  ///
  /// In pl, this message translates to:
  /// **'Preparat'**
  String get treatmentMedicine;

  /// No description provided for @treatmentMedicineRequired.
  ///
  /// In pl, this message translates to:
  /// **'Nazwa preparatu jest wymagana'**
  String get treatmentMedicineRequired;

  /// No description provided for @treatmentDose.
  ///
  /// In pl, this message translates to:
  /// **'Dawka'**
  String get treatmentDose;

  /// No description provided for @treatmentDoseRequired.
  ///
  /// In pl, this message translates to:
  /// **'Dawka jest wymagana'**
  String get treatmentDoseRequired;

  /// No description provided for @treatmentNote.
  ///
  /// In pl, this message translates to:
  /// **'Notatka'**
  String get treatmentNote;

  /// No description provided for @treatmentDoseCount.
  ///
  /// In pl, this message translates to:
  /// **'{count, plural, =1{1 dawka} few{{count} dawki} many{{count} dawek} other{{count} dawek}}'**
  String treatmentDoseCount(int count);

  /// No description provided for @treatmentTreatedBy.
  ///
  /// In pl, this message translates to:
  /// **'Przez {name}'**
  String treatmentTreatedBy(String name);

  /// No description provided for @treatmentTreatAllHives.
  ///
  /// In pl, this message translates to:
  /// **'Lecz ule'**
  String get treatmentTreatAllHives;

  /// No description provided for @treatmentBulkSuccess.
  ///
  /// In pl, this message translates to:
  /// **'{count, plural, =1{Leczenie zapisano dla 1 ula} few{Leczenie zapisano dla {count} uli} many{Leczenie zapisano dla {count} uli} other{Leczenie zapisano dla {count} uli}}'**
  String treatmentBulkSuccess(int count);

  /// No description provided for @feedingTitle.
  ///
  /// In pl, this message translates to:
  /// **'Podkarmianie'**
  String get feedingTitle;

  /// No description provided for @feedingAdd.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj podkarmianie'**
  String get feedingAdd;

  /// No description provided for @feedingEdit.
  ///
  /// In pl, this message translates to:
  /// **'Edytuj podkarmianie'**
  String get feedingEdit;

  /// No description provided for @feedingEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Brak podkarmiań'**
  String get feedingEmpty;

  /// No description provided for @feedingDeleteConfirm.
  ///
  /// In pl, this message translates to:
  /// **'Usunąć podkarmianie?'**
  String get feedingDeleteConfirm;

  /// No description provided for @feedingDeleteWarning.
  ///
  /// In pl, this message translates to:
  /// **'Ta operacja trwale usunie wpis podkarmiania.'**
  String get feedingDeleteWarning;

  /// No description provided for @feedingDate.
  ///
  /// In pl, this message translates to:
  /// **'Data podkarmiania'**
  String get feedingDate;

  /// No description provided for @feedingType.
  ///
  /// In pl, this message translates to:
  /// **'Pokarm'**
  String get feedingType;

  /// No description provided for @feedingTypeRequired.
  ///
  /// In pl, this message translates to:
  /// **'Rodzaj pokarmu jest wymagany'**
  String get feedingTypeRequired;

  /// No description provided for @feedingAmount.
  ///
  /// In pl, this message translates to:
  /// **'Ilość'**
  String get feedingAmount;

  /// No description provided for @feedingAmountRequired.
  ///
  /// In pl, this message translates to:
  /// **'Ilość jest wymagana'**
  String get feedingAmountRequired;

  /// No description provided for @feedingNote.
  ///
  /// In pl, this message translates to:
  /// **'Notatka'**
  String get feedingNote;

  /// No description provided for @feedingFedBy.
  ///
  /// In pl, this message translates to:
  /// **'Przez {name}'**
  String feedingFedBy(String name);

  /// No description provided for @feedingFeedAllHives.
  ///
  /// In pl, this message translates to:
  /// **'Podkarm ule'**
  String get feedingFeedAllHives;

  /// No description provided for @feedingBulkSuccess.
  ///
  /// In pl, this message translates to:
  /// **'{count, plural, =1{Podkarmianie zapisano dla 1 ula} few{Podkarmianie zapisano dla {count} uli} many{Podkarmianie zapisano dla {count} uli} other{Podkarmianie zapisano dla {count} uli}}'**
  String feedingBulkSuccess(int count);

  /// No description provided for @bulkSelectHives.
  ///
  /// In pl, this message translates to:
  /// **'Wybierz ule'**
  String get bulkSelectHives;

  /// No description provided for @harvestTitle.
  ///
  /// In pl, this message translates to:
  /// **'Zbiory'**
  String get harvestTitle;

  /// No description provided for @harvestAdd.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj zbiór'**
  String get harvestAdd;

  /// No description provided for @harvestEdit.
  ///
  /// In pl, this message translates to:
  /// **'Edytuj zbiór'**
  String get harvestEdit;

  /// No description provided for @harvestEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Brak zbiorów'**
  String get harvestEmpty;

  /// No description provided for @harvestDeleteConfirm.
  ///
  /// In pl, this message translates to:
  /// **'Usunąć zbiór?'**
  String get harvestDeleteConfirm;

  /// No description provided for @harvestDeleteWarning.
  ///
  /// In pl, this message translates to:
  /// **'Ta operacja trwale usunie wpis zbioru.'**
  String get harvestDeleteWarning;

  /// No description provided for @harvestDate.
  ///
  /// In pl, this message translates to:
  /// **'Data zbioru'**
  String get harvestDate;

  /// No description provided for @harvestFrames.
  ///
  /// In pl, this message translates to:
  /// **'Ramki'**
  String get harvestFrames;

  /// No description provided for @harvestHalfFrames.
  ///
  /// In pl, this message translates to:
  /// **'Półramki'**
  String get harvestHalfFrames;

  /// No description provided for @harvestFramesCount.
  ///
  /// In pl, this message translates to:
  /// **'{count, plural, =1{1 ramka} few{{count} ramki} many{{count} ramek} other{{count} ramek}}'**
  String harvestFramesCount(int count);

  /// No description provided for @harvestHalfFramesCount.
  ///
  /// In pl, this message translates to:
  /// **'{count, plural, =1{1 półramka} few{{count} półramki} many{{count} półramek} other{{count} półramek}}'**
  String harvestHalfFramesCount(int count);

  /// No description provided for @harvestKilograms.
  ///
  /// In pl, this message translates to:
  /// **'Kilogramy (kg)'**
  String get harvestKilograms;

  /// No description provided for @harvestKilogramsRequired.
  ///
  /// In pl, this message translates to:
  /// **'Kilogramy są wymagane'**
  String get harvestKilogramsRequired;

  /// No description provided for @harvestNote.
  ///
  /// In pl, this message translates to:
  /// **'Notatka'**
  String get harvestNote;

  /// No description provided for @harvestFramesRequired.
  ///
  /// In pl, this message translates to:
  /// **'Wymagana co najmniej jedna ramka'**
  String get harvestFramesRequired;

  /// No description provided for @harvestHarvestedBy.
  ///
  /// In pl, this message translates to:
  /// **'Przez {name}'**
  String harvestHarvestedBy(String name);

  /// No description provided for @honeyBatchTitle.
  ///
  /// In pl, this message translates to:
  /// **'Partie miodu'**
  String get honeyBatchTitle;

  /// No description provided for @honeyBatchEmpty.
  ///
  /// In pl, this message translates to:
  /// **'Brak partii miodu'**
  String get honeyBatchEmpty;

  /// No description provided for @honeyBatchAdd.
  ///
  /// In pl, this message translates to:
  /// **'Dodaj partię miodu'**
  String get honeyBatchAdd;

  /// No description provided for @honeyBatchEditTitle.
  ///
  /// In pl, this message translates to:
  /// **'Edytuj partię miodu'**
  String get honeyBatchEditTitle;

  /// No description provided for @honeyBatchHoneyType.
  ///
  /// In pl, this message translates to:
  /// **'Rodzaj miodu'**
  String get honeyBatchHoneyType;

  /// No description provided for @honeyBatchHoneyTypeRequired.
  ///
  /// In pl, this message translates to:
  /// **'Rodzaj miodu jest wymagany'**
  String get honeyBatchHoneyTypeRequired;

  /// No description provided for @honeyBatchProcessingMethod.
  ///
  /// In pl, this message translates to:
  /// **'Metoda przetwarzania'**
  String get honeyBatchProcessingMethod;

  /// No description provided for @honeyBatchMethodRaw.
  ///
  /// In pl, this message translates to:
  /// **'Surowy'**
  String get honeyBatchMethodRaw;

  /// No description provided for @honeyBatchMethodFiltered.
  ///
  /// In pl, this message translates to:
  /// **'Filtrowany'**
  String get honeyBatchMethodFiltered;

  /// No description provided for @honeyBatchMethodPasteurized.
  ///
  /// In pl, this message translates to:
  /// **'Pasteryzowany'**
  String get honeyBatchMethodPasteurized;

  /// No description provided for @honeyBatchGatheringDate.
  ///
  /// In pl, this message translates to:
  /// **'Data pozyskania'**
  String get honeyBatchGatheringDate;

  /// No description provided for @honeyBatchAmountKg.
  ///
  /// In pl, this message translates to:
  /// **'Ilość (kg)'**
  String get honeyBatchAmountKg;

  /// No description provided for @honeyBatchAmountRequired.
  ///
  /// In pl, this message translates to:
  /// **'Ilość jest wymagana'**
  String get honeyBatchAmountRequired;

  /// No description provided for @honeyBatchAmountInvalid.
  ///
  /// In pl, this message translates to:
  /// **'Podaj prawidłową ilość'**
  String get honeyBatchAmountInvalid;

  /// No description provided for @honeyBatchPdfLabel.
  ///
  /// In pl, this message translates to:
  /// **'PDF z badania laboratoryjnego'**
  String get honeyBatchPdfLabel;

  /// No description provided for @honeyBatchNoPdf.
  ///
  /// In pl, this message translates to:
  /// **'Brak'**
  String get honeyBatchNoPdf;

  /// No description provided for @honeyBatchCertify.
  ///
  /// In pl, this message translates to:
  /// **'Certyfikuj'**
  String get honeyBatchCertify;

  /// No description provided for @honeyBatchCertifyConfirmTitle.
  ///
  /// In pl, this message translates to:
  /// **'Certyfikować tę partię?'**
  String get honeyBatchCertifyConfirmTitle;

  /// No description provided for @honeyBatchCertifyConfirmMessage.
  ///
  /// In pl, this message translates to:
  /// **'Po certyfikacji tej partii miodu nie będzie już można jej edytować.'**
  String get honeyBatchCertifyConfirmMessage;

  /// No description provided for @honeyBatchRetry.
  ///
  /// In pl, this message translates to:
  /// **'Ponów certyfikację'**
  String get honeyBatchRetry;

  /// No description provided for @honeyBatchNotCertified.
  ///
  /// In pl, this message translates to:
  /// **'Niecertyfikowane'**
  String get honeyBatchNotCertified;

  /// No description provided for @honeyBatchInProgress.
  ///
  /// In pl, this message translates to:
  /// **'Certyfikacja w toku'**
  String get honeyBatchInProgress;

  /// No description provided for @honeyBatchDeleteConfirm.
  ///
  /// In pl, this message translates to:
  /// **'Usunąć partię miodu?'**
  String get honeyBatchDeleteConfirm;

  /// No description provided for @honeyBatchDeleteWarning.
  ///
  /// In pl, this message translates to:
  /// **'Ta operacja trwale usunie wpis partii miodu.'**
  String get honeyBatchDeleteWarning;

  /// No description provided for @honeyBatchStatusQueued.
  ///
  /// In pl, this message translates to:
  /// **'W kolejce'**
  String get honeyBatchStatusQueued;

  /// No description provided for @honeyBatchStatusSubmitting.
  ///
  /// In pl, this message translates to:
  /// **'Wysyłanie'**
  String get honeyBatchStatusSubmitting;

  /// No description provided for @honeyBatchStatusSubmitted.
  ///
  /// In pl, this message translates to:
  /// **'Wysłano'**
  String get honeyBatchStatusSubmitted;

  /// No description provided for @honeyBatchStatusPendingConfirmation.
  ///
  /// In pl, this message translates to:
  /// **'Oczekuje na potwierdzenie'**
  String get honeyBatchStatusPendingConfirmation;

  /// No description provided for @honeyBatchStatusConfirmed.
  ///
  /// In pl, this message translates to:
  /// **'Potwierdzono'**
  String get honeyBatchStatusConfirmed;

  /// No description provided for @honeyBatchStatusFailed.
  ///
  /// In pl, this message translates to:
  /// **'Niepowodzenie'**
  String get honeyBatchStatusFailed;

  /// No description provided for @honeyBatchStatusReverted.
  ///
  /// In pl, this message translates to:
  /// **'Wycofano'**
  String get honeyBatchStatusReverted;

  /// No description provided for @honeyBatchViewQr.
  ///
  /// In pl, this message translates to:
  /// **'Pokaż kod QR'**
  String get honeyBatchViewQr;

  /// No description provided for @honeyBatchDownloadQr.
  ///
  /// In pl, this message translates to:
  /// **'Pobierz kod QR'**
  String get honeyBatchDownloadQr;
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
