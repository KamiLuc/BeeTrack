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
  String get marketplaceMapTooltip => 'Mapa ogłoszeń';

  @override
  String get marketplaceMapTitle => 'Mapa ogłoszeń';

  @override
  String get marketplaceMapEmpty =>
      'Brak ogłoszeń z lokalizacją pasującą do filtrów';

  @override
  String get marketplacePriceMinHint => 'Cena od';

  @override
  String get marketplacePriceMaxHint => 'Cena do';

  @override
  String get marketplaceFiltersButton => 'Filtry';

  @override
  String get marketplaceClearFilters => 'Wyczyść filtry';

  @override
  String get marketplacePostedWithinAny => 'Dowolny czas';

  @override
  String get marketplacePostedWithinToday => 'Dzisiaj';

  @override
  String get marketplacePostedWithin7Days => 'Ostatnie 7 dni';

  @override
  String get marketplacePostedWithin14Days => 'Ostatnie 14 dni';

  @override
  String get marketplacePostedWithin30Days => 'Ostatnie 30 dni';

  @override
  String get marketplaceDistanceLabel => 'Odległość';

  @override
  String get marketplaceGpsUnavailable => 'GPS niedostępny na tym urządzeniu';

  @override
  String get marketplaceDistanceAny => 'Dowolna odległość';

  @override
  String get marketplaceDistance5Km => 'Do 5 km';

  @override
  String get marketplaceDistance10Km => 'Do 10 km';

  @override
  String get marketplaceDistance25Km => 'Do 25 km';

  @override
  String get marketplaceDistance50Km => 'Do 50 km';

  @override
  String get marketplaceDistance100Km => 'Do 100 km';

  @override
  String get marketplaceApiaryFilterLabel =>
      'Tylko ogłoszenia z powiązaną pasieką';

  @override
  String marketplaceDistanceAway(String km) {
    return '$km km stąd';
  }

  @override
  String get marketplacePriceOnRequest => 'Cena do negocjacji';

  @override
  String get marketplacePriceFree => 'Za darmo';

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
  String get marketplaceCallButton => 'Zadzwoń';

  @override
  String get marketplaceWriteButton => 'Napisz';

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
  String get marketplaceFieldPriceRequired => 'Cena jest wymagana';

  @override
  String get marketplaceFieldPriceTooLarge =>
      'Cena musi być mniejsza niż 100 000 000';

  @override
  String get marketplaceFieldAddress => 'Adres';

  @override
  String get marketplaceFieldLatitude => 'Szerokość geograficzna';

  @override
  String get marketplaceFieldLongitude => 'Długość geograficzna';

  @override
  String get marketplaceLocationRequired =>
      'Wybierz lokalizację na mapie lub użyj GPS';

  @override
  String get locationPickerTitle => 'Wybierz lokalizację';

  @override
  String get locationPickerHint => 'Dotknij mapę, aby wybrać lokalizację';

  @override
  String get locationPickerGpsButton => 'GPS';

  @override
  String get locationPickerMapButton => 'Mapa';

  @override
  String get marketplaceFieldPhone => 'Telefon';

  @override
  String get marketplaceFieldPhoneInvalid => 'Podaj prawidłowy numer telefonu';

  @override
  String get marketplaceFieldEmail => 'E-mail';

  @override
  String get marketplaceContactRequired =>
      'Podaj numer telefonu lub adres e-mail';

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
  String get marketplaceEditScreenTitle => 'Edytuj ogłoszenie';

  @override
  String get myListingsTitle => 'Moje ogłoszenia';

  @override
  String get myListingsEmpty => 'Nie masz jeszcze żadnych ogłoszeń';

  @override
  String get favoritesTitle => 'Ulubione';

  @override
  String get favoritesEmpty => 'Nie masz jeszcze żadnych ulubionych ogłoszeń';

  @override
  String get marketplaceHiddenBadge => 'Prywatne';

  @override
  String get marketplaceHideListing => 'Ustaw jako prywatne';

  @override
  String get marketplaceShowListing => 'Ustaw jako publiczne';

  @override
  String get marketplaceDeleteConfirm => 'Usunąć to ogłoszenie?';

  @override
  String get marketplaceDeleteWarning => 'Tej operacji nie można cofnąć.';

  @override
  String get apiaryMapTitle => 'Mapa pasiek';

  @override
  String get apiaryMapTooltip => 'Pokaż na mapie';

  @override
  String get apiaryLocationTitle => 'Lokalizacja pasieki';

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
  String get hiveDetailFeedings => 'Podkarmianie';

  @override
  String get hiveDetailNoFeedings => 'Brak podkarmiań';

  @override
  String get hiveDetailLogFeeding => 'Dodaj podkarmianie';

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
  String generalFieldTooLong(String field, int max) {
    return '$field może mieć maksymalnie $max znaków';
  }

  @override
  String generalValueTooLarge(String field, String max) {
    return '$field musi być mniejsze lub równe $max';
  }

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
  String get treatmentTreatAllHives => 'Lecz ule';

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
  String get feedingTitle => 'Podkarmianie';

  @override
  String get feedingAdd => 'Dodaj podkarmianie';

  @override
  String get feedingEdit => 'Edytuj podkarmianie';

  @override
  String get feedingEmpty => 'Brak podkarmiań';

  @override
  String get feedingDeleteConfirm => 'Usunąć podkarmianie?';

  @override
  String get feedingDeleteWarning =>
      'Ta operacja trwale usunie wpis podkarmiania.';

  @override
  String get feedingDate => 'Data podkarmiania';

  @override
  String get feedingType => 'Pokarm';

  @override
  String get feedingTypeRequired => 'Rodzaj pokarmu jest wymagany';

  @override
  String get feedingAmount => 'Ilość';

  @override
  String get feedingAmountRequired => 'Ilość jest wymagana';

  @override
  String get feedingNote => 'Notatka';

  @override
  String feedingFedBy(String name) {
    return 'Przez $name';
  }

  @override
  String get feedingFeedAllHives => 'Podkarm ule';

  @override
  String feedingBulkSuccess(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'Podkarmianie zapisano dla $count uli',
      many: 'Podkarmianie zapisano dla $count uli',
      few: 'Podkarmianie zapisano dla $count uli',
      one: 'Podkarmianie zapisano dla 1 ula',
    );
    return '$_temp0';
  }

  @override
  String get bulkSelectHives => 'Wybierz ule';

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
  String harvestFramesCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count ramek',
      many: '$count ramek',
      few: '$count ramki',
      one: '1 ramka',
    );
    return '$_temp0';
  }

  @override
  String harvestHalfFramesCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count półramek',
      many: '$count półramek',
      few: '$count półramki',
      one: '1 półramka',
    );
    return '$_temp0';
  }

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

  @override
  String get honeyBatchTitle => 'Partie miodu';

  @override
  String get honeyBatchEmpty => 'Brak partii miodu';

  @override
  String get honeyBatchAdd => 'Dodaj partię miodu';

  @override
  String get honeyBatchEditTitle => 'Edytuj partię miodu';

  @override
  String get honeyBatchHoneyType => 'Rodzaj miodu';

  @override
  String get honeyBatchHoneyTypeRequired => 'Rodzaj miodu jest wymagany';

  @override
  String get honeyBatchProcessingMethod => 'Metoda przetwarzania';

  @override
  String get honeyBatchMethodRaw => 'Surowy';

  @override
  String get honeyBatchMethodFiltered => 'Filtrowany';

  @override
  String get honeyBatchMethodPasteurized => 'Pasteryzowany';

  @override
  String get honeyBatchGatheringDate => 'Data pozyskania';

  @override
  String get honeyBatchAmountKg => 'Ilość (kg)';

  @override
  String get honeyBatchAmountRequired => 'Ilość jest wymagana';

  @override
  String get honeyBatchAmountInvalid => 'Podaj prawidłową ilość';

  @override
  String get honeyBatchPdfLabel => 'PDF z badania laboratoryjnego';

  @override
  String get honeyBatchNoPdf => 'Brak';

  @override
  String get honeyBatchCertify => 'Certyfikuj';

  @override
  String get honeyBatchCertifyConfirmTitle => 'Certyfikować tę partię?';

  @override
  String get honeyBatchCertifyConfirmMessage =>
      'Po certyfikacji tej partii miodu nie będzie już można jej edytować.';

  @override
  String get honeyBatchRetry => 'Ponów certyfikację';

  @override
  String get honeyBatchNotCertified => 'Niecertyfikowane';

  @override
  String get honeyBatchInProgress => 'Certyfikacja w toku';

  @override
  String get honeyBatchDeleteConfirm => 'Usunąć partię miodu?';

  @override
  String get honeyBatchDeleteWarning =>
      'Ta operacja trwale usunie wpis partii miodu.';

  @override
  String get honeyBatchStatusQueued => 'W kolejce';

  @override
  String get honeyBatchStatusSubmitting => 'Wysyłanie';

  @override
  String get honeyBatchStatusSubmitted => 'Wysłano';

  @override
  String get honeyBatchStatusPendingConfirmation => 'Oczekuje na potwierdzenie';

  @override
  String get honeyBatchStatusConfirmed => 'Potwierdzono';

  @override
  String get honeyBatchStatusFailed => 'Niepowodzenie';

  @override
  String get honeyBatchStatusReverted => 'Wycofano';

  @override
  String get honeyBatchViewQr => 'Pokaż kod QR';

  @override
  String get honeyBatchDownloadQr => 'Pobierz kod QR';
}
