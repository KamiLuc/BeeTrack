// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appName => 'BeeTrack';

  @override
  String get generalSave => 'Save';

  @override
  String get generalCancel => 'Cancel';

  @override
  String get generalDelete => 'Delete';

  @override
  String get generalEdit => 'Edit';

  @override
  String get generalConfirm => 'Confirm';

  @override
  String get generalError => 'Something went wrong. Please try again.';

  @override
  String get generalLoading => 'Loading...';

  @override
  String get authEmail => 'Email';

  @override
  String get authPassword => 'Password';

  @override
  String get authName => 'Full name';

  @override
  String get authLogin => 'Log in';

  @override
  String get authRegister => 'Register';

  @override
  String get authLogout => 'Log out';

  @override
  String get authNoAccount => 'Don\'t have an account? Register';

  @override
  String get authHaveAccount => 'Already have an account? Log in';

  @override
  String get authInvalidEmail => 'Invalid email address';

  @override
  String get authWeakPassword => 'Password must be at least 8 characters';

  @override
  String get authInvalidCredentials => 'Invalid email or password';

  @override
  String get authEmailTaken => 'This email is already taken';

  @override
  String get roleOwner => 'Owner';

  @override
  String get roleMember => 'Member';

  @override
  String get apiaryTitle => 'Apiaries';

  @override
  String get apiaryAdd => 'Add apiary';

  @override
  String get apiaryName => 'Apiary';

  @override
  String get apiaryLatitude => 'Latitude';

  @override
  String get apiaryLongitude => 'Longitude';

  @override
  String get apiaryGpsUnavailable => 'GPS not available on this device';

  @override
  String get apiaryLocation => 'Location (optional)';

  @override
  String get apiaryGridRows => 'Grid rows';

  @override
  String get apiaryGridCols => 'Grid columns';

  @override
  String get apiaryEmpty => 'You have no apiaries yet';

  @override
  String get hiveTitle => 'Hives';

  @override
  String get hiveAdd => 'Add hive';

  @override
  String get hiveName => 'Hive name';

  @override
  String get hiveType => 'Hive type';

  @override
  String get hiveActive => 'Active';

  @override
  String get hiveInactive => 'Inactive';

  @override
  String get hiveEmpty => 'No hives in this apiary';
}
