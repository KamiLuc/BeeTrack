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

  /// No description provided for @profileNameUpdated.
  ///
  /// In pl, this message translates to:
  /// **'Nazwa zaktualizowana'**
  String get profileNameUpdated;

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

  /// No description provided for @hiveFrames.
  ///
  /// In pl, this message translates to:
  /// **'Ramki'**
  String get hiveFrames;

  /// No description provided for @hiveFramesWarning.
  ///
  /// In pl, this message translates to:
  /// **'Liczba ramek w inspekcji przekracza pojemność ula'**
  String get hiveFramesWarning;

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

  /// No description provided for @generalLoadMore.
  ///
  /// In pl, this message translates to:
  /// **'Załaduj więcej'**
  String get generalLoadMore;

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

  /// No description provided for @inspectionFramesHoney.
  ///
  /// In pl, this message translates to:
  /// **'Ramki z miodem'**
  String get inspectionFramesHoney;

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

  /// No description provided for @inspectionFramesAddedHoney.
  ///
  /// In pl, this message translates to:
  /// **'Dodane ramki z miodem'**
  String get inspectionFramesAddedHoney;

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
