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
  String get profileLanguage => 'Język';

  @override
  String get profileLanguageEn => 'Angielski';

  @override
  String get profileLanguagePl => 'Polski';

  @override
  String get profileDisplayName => 'Nazwa wyświetlana';

  @override
  String get profileNameUpdated => 'Nazwa zaktualizowana';

  @override
  String get generalSave => 'Zapisz';

  @override
  String get generalCancel => 'Anuluj';

  @override
  String get generalClose => 'Zamknij';

  @override
  String get generalDelete => 'Usuń';

  @override
  String get generalEdit => 'Edytuj';

  @override
  String get generalConfirm => 'Potwierdź';

  @override
  String get generalError => 'Wystąpił błąd. Spróbuj ponownie.';

  @override
  String get generalRetry => 'Spróbuj ponownie';

  @override
  String get generalLoading => 'Ładowanie...';

  @override
  String get deletePuzzlePrompt => 'Aby potwierdzić, rozwiąż:';

  @override
  String get deletePuzzleWrong => 'Zła odpowiedź';

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
  String get authEmailNotVerified =>
      'Potwierdź swój adres e-mail przed zalogowaniem';

  @override
  String get authCheckEmail => 'Sprawdź pocztę';

  @override
  String authCheckEmailMessage(String email) {
    return 'Wysłaliśmy link weryfikacyjny na adres $email. Sprawdź skrzynkę odbiorczą.';
  }

  @override
  String get authResendEmail => 'Wyślij ponownie';

  @override
  String get authBackToLogin => 'Wróć do logowania';

  @override
  String get authForgotPassword => 'Nie pamiętasz hasła?';

  @override
  String get authForgotPasswordTitle => 'Resetowanie hasła';

  @override
  String get authForgotPasswordSubtitle =>
      'Podaj swój e-mail, a wyślemy Ci link do resetowania hasła.';

  @override
  String get authForgotPasswordSent =>
      'Sprawdź skrzynkę — link do resetowania hasła został wysłany.';

  @override
  String get authSendResetLink => 'Wyślij link';

  @override
  String get authVerifyingEmail => 'Weryfikacja adresu e-mail...';

  @override
  String get authEmailVerified => 'Adres e-mail zweryfikowany!';

  @override
  String get authEmailVerifiedMessage =>
      'Twoje konto jest aktywne. Możesz się zalogować.';

  @override
  String get authVerificationFailed => 'Weryfikacja nie powiodła się';

  @override
  String get authVerificationFailedMessage =>
      'Link mógł wygasnąć lub już był użyty.';

  @override
  String get authGoToLogin => 'Przejdź do logowania';

  @override
  String get authNewPassword => 'Nowe hasło';

  @override
  String get authPasswordChanged => 'Hasło zostało zmienione!';

  @override
  String get authPasswordChangedMessage =>
      'Możesz teraz zalogować się nowym hasłem.';

  @override
  String get authInvalidResetToken => 'Ten link wygasł lub już był użyty.';

  @override
  String get roleOwner => 'Właściciel';

  @override
  String get roleMember => 'Członek';

  @override
  String get invitationTitle => 'Zaproszenia';

  @override
  String get invitationMembers => 'Członkowie';

  @override
  String get invitationPending => 'Oczekujące zaproszenia';

  @override
  String get invitationInvite => 'Zarządzaj członkami';

  @override
  String get invitationEmailHint => 'Adres e-mail';

  @override
  String get invitationSend => 'Wyślij zaproszenie';

  @override
  String get invitationSentSuccess => 'Zaproszenie wysłane';

  @override
  String get invitationAlreadyPending =>
      'Zaproszenie dla tego adresu e-mail już oczekuje';

  @override
  String get invitationAlreadyMember => 'Ten użytkownik jest już członkiem';

  @override
  String get invitationCannotInviteSelf => 'Nie możesz zaprosić siebie';

  @override
  String get invitationUserNotFound =>
      'Nie znaleziono konta dla tego adresu e-mail';

  @override
  String get invitationNoMembers => 'Brak członków';

  @override
  String get invitationNoPending => 'Brak oczekujących zaproszeń';

  @override
  String get invitationRemove => 'Usuń';

  @override
  String get invitationAccept => 'Akceptuj';

  @override
  String get invitationDecline => 'Odrzuć';

  @override
  String invitationFrom(String apiary, String name) {
    return 'z $apiary od $name';
  }

  @override
  String get invitationBadgeTooltip => 'Oczekujące zaproszenia';

  @override
  String get leaveApiary => 'Opuść pasiekę';

  @override
  String get leaveApiaryConfirm => 'Opuścić pasiekę?';

  @override
  String get leaveApiaryWarning => 'Utracisz dostęp do tej pasieki.';

  @override
  String get apiaryTitle => 'Pasieki';

  @override
  String get marketplaceTitle => 'Ogłoszenia';

  @override
  String get marketplaceComingSoon => 'Wkrótce';

  @override
  String get marketplaceSearchHint => 'Szukaj ogłoszeń';

  @override
  String get marketplaceEmpty => 'Brak ogłoszeń';

  @override
  String get marketplaceMapTooltip => 'Widok mapy (wkrótce)';

  @override
  String get marketplacePriceOnRequest => 'Cena do negocjacji';

  @override
  String get marketplaceCategoryAll => 'Wszystkie';

  @override
  String get marketplaceCategoryHoney => 'Miód';

  @override
  String get marketplaceCategoryPollen => 'Pyłek';

  @override
  String get marketplaceCategoryBeeColonies => 'Rodziny pszczele';

  @override
  String get marketplaceCategoryQueenBees => 'Matki pszczele';

  @override
  String get marketplaceCategoryBeehives => 'Ule';

  @override
  String get marketplaceCategoryEquipment => 'Sprzęt';

  @override
  String get marketplaceCategoryExtractionEquipment =>
      'Sprzęt do wirowania miodu';

  @override
  String get marketplaceCategoryFeed => 'Pokarm dla pszczół';

  @override
  String get marketplaceCategorySupplies => 'Zaopatrzenie';

  @override
  String get marketplaceCategoryWaxFoundation => 'Węza';

  @override
  String get marketplaceCategoryBeeswax => 'Wosk pszczeli';

  @override
  String get marketplaceCategoryPropolis => 'Propolis';

  @override
  String get marketplaceCategoryServices => 'Usługi';

  @override
  String get marketplaceCategoryOther => 'Inne';

  @override
  String get marketplaceFavoriteAdd => 'Dodaj do ulubionych';

  @override
  String get marketplaceFavoriteRemove => 'Usuń z ulubionych';

  @override
  String get marketplaceDescriptionLabel => 'Opis';

  @override
  String get marketplaceContactLabel => 'Kontakt';

  @override
  String get marketplaceApiaryLabel => 'Pasieka powiązana z ogłoszeniem';

  @override
  String get marketplaceQuantityLabel => 'Ilość';

  @override
  String marketplacePostedOn(String date) {
    return 'Dodano $date';
  }

  @override
  String get marketplaceCreateScreenTitle => 'Nowe ogłoszenie';

  @override
  String get marketplaceFieldTitle => 'Tytuł';

  @override
  String get marketplaceFieldTitleRequired => 'Tytuł jest wymagany';

  @override
  String get marketplaceFieldCategory => 'Kategoria';

  @override
  String get marketplaceFieldCategoryRequired => 'Wybierz kategorię';

  @override
  String get marketplaceFieldPrice => 'Cena';

  @override
  String get marketplaceFieldPriceInvalid => 'Podaj prawidłową cenę';

  @override
  String get marketplaceFieldAddress => 'Adres';

  @override
  String get marketplaceFieldPhone => 'Telefon';

  @override
  String get marketplaceFieldEmail => 'E-mail';

  @override
  String get marketplaceApiaryNone => 'Brak';

  @override
  String get marketplacePhotosLabel => 'Zdjęcia';

  @override
  String get marketplaceAddPhoto => 'Dodaj zdjęcie';

  @override
  String get marketplacePhotoSourceGallery => 'Wybierz z galerii';

  @override
  String get marketplacePhotoSourceCamera => 'Zrób zdjęcie';

  @override
  String get apiaryMapTitle => 'Mapa pasiek';

  @override
  String get apiaryMapTooltip => 'Pokaż na mapie';

  @override
  String get apiaryCopy => 'Skopiuj pasiekę';

  @override
  String get apiaryCopySuffix => 'kopia';

  @override
  String get apiaryCopyNewName => 'Nowa nazwa';

  @override
  String get apiaryCopied => 'Pasieka skopiowana';

  @override
  String get apiaryAdd => 'Dodaj pasiekę';

  @override
  String get apiaryName => 'Pasieka';

  @override
  String get apiaryNameRequired => 'Nazwa pasieki nie może być pusta';

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
  String apiaryGridHivesWillMove(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other:
          '$count uli zostanie przeniesionych, aby zmieścić się w nowej siatce.',
      many:
          '$count uli zostanie przeniesionych, aby zmieścić się w nowej siatce.',
      few: '$count ule zostaną przeniesione, aby zmieścić się w nowej siatce.',
      one: '1 ul zostanie przeniesiony, aby zmieścić się w nowej siatce.',
    );
    return '$_temp0';
  }

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
  String get hiveQueenless => 'Bezmateczny';

  @override
  String get hiveReadyForHarvest => 'Gotowy do zbioru';

  @override
  String get hiveSick => 'Chory';

  @override
  String get hiveFilterTooltip => 'Filtruj ule';

  @override
  String get hiveListTooltip => 'Lista uli';

  @override
  String get apiaryCenterView => 'Wyśrodkuj widok';

  @override
  String get hiveDiseases => 'Choroby';

  @override
  String get hiveStatus => 'Status';

  @override
  String get hiveDetailInspections => 'Inspekcje';

  @override
  String get hiveDetailNoInspections => 'Brak inspekcji';

  @override
  String get hiveDetailAddInspection => 'Dodaj inspekcję';

  @override
  String get hiveDetailViewInspections => 'Pokaż wszystkie';

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
  String get hiveChangeApiary => 'Zmień pasiekę';

  @override
  String get hiveChangeApiaryTitle => 'Przenieś ul';

  @override
  String get hiveChangeApiaryNoSpace =>
      'Docelowa pasieka nie ma wolnego miejsca';

  @override
  String get hiveDuplicateName => 'Ul o tej nazwie już istnieje w tej pasiece';

  @override
  String get generalRequired => 'Wymagane';

  @override
  String get generalLoadMore => 'Załaduj więcej';

  @override
  String get inspectionTitle => 'Inspekcje';

  @override
  String get inspectionAdd => 'Dodaj inspekcję';

  @override
  String get inspectionEdit => 'Edytuj inspekcję';

  @override
  String get inspectionDeleteConfirm => 'Usunąć inspekcję?';

  @override
  String get inspectionDeleteWarning => 'Ta operacja trwale usunie inspekcję.';

  @override
  String get inspectionEmpty => 'Brak inspekcji';

  @override
  String get inspectionDate => 'Data inspekcji';

  @override
  String get inspectionQueenSeen => 'Matka widziana';

  @override
  String get inspectionQueenStatusSeen => 'Matka widziana';

  @override
  String get inspectionQueenStatusNotSeen => 'Matka niewidziana';

  @override
  String get inspectionBroodPattern => 'Czerw';

  @override
  String get inspectionBroodExcellent => 'Dużo';

  @override
  String get inspectionBroodGood => 'Średnio';

  @override
  String get inspectionBroodPoor => 'Mało';

  @override
  String get inspectionBroodNone => 'Brak';

  @override
  String get inspectionAggressiveness => 'Agresywność';

  @override
  String get inspectionAggressivenessCalm => 'Spokojne';

  @override
  String get inspectionAggressivenessMild => 'Łagodne';

  @override
  String get inspectionAggressivenessAggressive => 'Agresywne';

  @override
  String get inspectionAggressivenessVeryAggressive => 'Bardzo agresywne';

  @override
  String get inspectionFramesBrood => 'Ramki z czerwiem';

  @override
  String get inspectionFramesFeed => 'Ramki z pokarmem';

  @override
  String get inspectionFramesPollen => 'Ramki z pyłkiem';

  @override
  String get inspectionFramesAddedDrawn => 'Dodane puste ramki';

  @override
  String get inspectionFramesAddedFoundation => 'Dodana węza';

  @override
  String get inspectionFramesAddedBrood => 'Dodane ramki z czerwiem';

  @override
  String get inspectionFramesAddedFeed => 'Dodane ramki z pokarmem';

  @override
  String get inspectionFramesTakenDrawn => 'Zabrane puste ramki';

  @override
  String get inspectionFramesTakenFoundation => 'Zabrana węza';

  @override
  String get inspectionFramesTakenBrood => 'Zabrane ramki z czerwiem';

  @override
  String get inspectionFramesTakenFeed => 'Zabrane ramki z pokarmem';

  @override
  String get inspectionQueenCellsCount => 'Mateczniki';

  @override
  String get inspectionQueenAdded => 'Poddano matkę';

  @override
  String get inspectionSectionObservations => 'Obserwacje';

  @override
  String get inspectionSectionFrames => 'Ramki';

  @override
  String get inspectionSectionHealth => 'Zdrowie';

  @override
  String get inspectionSectionHiveState => 'Stan ula';

  @override
  String get inspectionNotes => 'Notatki';

  @override
  String get inspectionNote => 'Notatka';

  @override
  String inspectionInspectedBy(String name) {
    return 'Przez $name';
  }

  @override
  String get inspectionDiseases => 'Choroby';

  @override
  String get inspectionDiseaseVarroa => 'Warroza';

  @override
  String get inspectionDiseaseNosema => 'Nosemoza';

  @override
  String get inspectionDiseaseDwv => 'Wirusy (DWV)';

  @override
  String get inspectionDiseaseAmericanFoulbrood => 'Zgnilec amerykański';

  @override
  String get inspectionDiseaseChalkbrood => 'Grzybica wapienna';

  @override
  String get inspectionDiseaseEuropeanFoulbrood => 'Zgnilec europejski';

  @override
  String get inspectionDiseaseLayingWorkers => 'Strutowienie rodziny';

  @override
  String get inspectionNotSet => 'Nie ustawiono';

  @override
  String get inspectionPhotos => 'Zdjęcia';

  @override
  String get inspectionPhotoSourceGallery => 'Galeria';

  @override
  String get inspectionPhotoSourceCamera => 'Aparat';

  @override
  String get inspectionAddPhoto => 'Dodaj zdjęcie';

  @override
  String get inspectionNoPhotos => 'Brak zdjęć';

  @override
  String get inspectionDeletePhoto => 'Usunąć zdjęcie?';

  @override
  String get inspectionDeletePhotoWarning =>
      'Ta operacja trwale usunie zdjęcie.';

  @override
  String inspectionPhotoCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count zdjęć',
      few: '$count zdjęcia',
      one: '1 zdjęcie',
    );
    return '$_temp0';
  }

  @override
  String get hiveTypeRequired => 'Typ ula jest wymagany';

  @override
  String get treatmentTitle => 'Zabiegi';

  @override
  String get treatmentAdd => 'Dodaj zabieg';

  @override
  String get treatmentEdit => 'Edytuj zabieg';

  @override
  String get treatmentEmpty => 'Brak zabiegów';

  @override
  String get treatmentDeleteConfirm => 'Usunąć zabieg?';

  @override
  String get treatmentDeleteWarning =>
      'Ta operacja trwale usunie wpis zabiegu.';

  @override
  String get treatmentDate => 'Data zabiegu';

  @override
  String get treatmentMedicine => 'Preparat';

  @override
  String get treatmentMedicineRequired => 'Nazwa preparatu jest wymagana';

  @override
  String get treatmentDose => 'Dawka';

  @override
  String get treatmentDoseRequired => 'Dawka jest wymagana';

  @override
  String get treatmentNote => 'Notatka';

  @override
  String treatmentDoseCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count dawek',
      many: '$count dawek',
      few: '$count dawki',
      one: '1 dawka',
    );
    return '$_temp0';
  }

  @override
  String treatmentTreatedBy(String name) {
    return 'Przez $name';
  }

  @override
  String get treatmentTreatAllHives => 'Lecz wszystkie ule';

  @override
  String treatmentBulkSuccess(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'Leczenie zapisano dla $count uli',
      many: 'Leczenie zapisano dla $count uli',
      few: 'Leczenie zapisano dla $count uli',
      one: 'Leczenie zapisano dla 1 ula',
    );
    return '$_temp0';
  }

  @override
  String get harvestTitle => 'Zbiory';

  @override
  String get harvestAdd => 'Dodaj zbiór';

  @override
  String get harvestEdit => 'Edytuj zbiór';

  @override
  String get harvestEmpty => 'Brak zbiorów';

  @override
  String get harvestDeleteConfirm => 'Usunąć zbiór?';

  @override
  String get harvestDeleteWarning => 'Ta operacja trwale usunie wpis zbioru.';

  @override
  String get harvestDate => 'Data zbioru';

  @override
  String get harvestFrames => 'Ramki';

  @override
  String get harvestHalfFrames => 'Półramki';

  @override
  String get harvestKilograms => 'Kilogramy (kg)';

  @override
  String get harvestKilogramsRequired => 'Kilogramy są wymagane';

  @override
  String get harvestNote => 'Notatka';

  @override
  String get harvestFramesRequired => 'Wymagana co najmniej jedna ramka';

  @override
  String harvestHarvestedBy(String name) {
    return 'Przez $name';
  }
}
