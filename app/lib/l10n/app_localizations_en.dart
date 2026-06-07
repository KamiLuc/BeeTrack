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
  String get profileLanguage => 'Language';

  @override
  String get profileLanguageEn => 'English';

  @override
  String get profileLanguagePl => 'Polish';

  @override
  String get profileDisplayName => 'Display name';

  @override
  String get profileNameUpdated => 'Name updated';

  @override
  String get generalSave => 'Save';

  @override
  String get generalCancel => 'Cancel';

  @override
  String get generalClose => 'Close';

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
  String get deletePuzzlePrompt => 'To confirm, solve:';

  @override
  String get deletePuzzleWrong => 'Wrong answer';

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
  String get authEmailNotVerified =>
      'Please verify your email before logging in';

  @override
  String get authCheckEmail => 'Check your email';

  @override
  String authCheckEmailMessage(String email) {
    return 'We sent a verification link to $email. Check your inbox.';
  }

  @override
  String get authResendEmail => 'Resend email';

  @override
  String get authBackToLogin => 'Back to log in';

  @override
  String get authForgotPassword => 'Forgot password?';

  @override
  String get authForgotPasswordTitle => 'Reset password';

  @override
  String get authForgotPasswordSubtitle =>
      'Enter your email and we\'ll send you a reset link.';

  @override
  String get authForgotPasswordSent => 'Check your inbox for the reset link.';

  @override
  String get authSendResetLink => 'Send reset link';

  @override
  String get authVerifyingEmail => 'Verifying your email...';

  @override
  String get authEmailVerified => 'Email verified!';

  @override
  String get authEmailVerifiedMessage =>
      'Your account is now active. You can log in.';

  @override
  String get authVerificationFailed => 'Verification failed';

  @override
  String get authVerificationFailedMessage =>
      'The link may have expired or already been used.';

  @override
  String get authGoToLogin => 'Go to login';

  @override
  String get authNewPassword => 'New password';

  @override
  String get authPasswordChanged => 'Password changed!';

  @override
  String get authPasswordChangedMessage =>
      'You can now log in with your new password.';

  @override
  String get authInvalidResetToken =>
      'This link has expired or already been used.';

  @override
  String get roleOwner => 'Owner';

  @override
  String get roleMember => 'Member';

  @override
  String get invitationTitle => 'Invitations';

  @override
  String get invitationMembers => 'Members';

  @override
  String get invitationPending => 'Pending invitations';

  @override
  String get invitationInvite => 'Manage members';

  @override
  String get invitationEmailHint => 'Email address';

  @override
  String get invitationSend => 'Send invitation';

  @override
  String get invitationSentSuccess => 'Invitation sent';

  @override
  String get invitationAlreadyPending =>
      'An invitation is already pending for this email';

  @override
  String get invitationAlreadyMember => 'This user is already a member';

  @override
  String get invitationCannotInviteSelf => 'You cannot invite yourself';

  @override
  String get invitationUserNotFound =>
      'No account found for that email address';

  @override
  String get invitationNoMembers => 'No members yet';

  @override
  String get invitationNoPending => 'No pending invitations';

  @override
  String get invitationRemove => 'Remove';

  @override
  String get invitationAccept => 'Accept';

  @override
  String get invitationDecline => 'Decline';

  @override
  String invitationFrom(String apiary, String name) {
    return 'from $apiary by $name';
  }

  @override
  String get invitationBadgeTooltip => 'Pending invitations';

  @override
  String get leaveApiary => 'Leave apiary';

  @override
  String get leaveApiaryConfirm => 'Leave apiary?';

  @override
  String get leaveApiaryWarning => 'You will lose access to this apiary.';

  @override
  String get apiaryTitle => 'Apiaries';

  @override
  String get apiaryMapTitle => 'Apiaries map';

  @override
  String get apiaryMapTooltip => 'Show on map';

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
  String get apiaryEdit => 'Edit apiary';

  @override
  String get apiaryDeleteConfirm => 'Delete apiary?';

  @override
  String get apiaryDeleteWarning =>
      'This will permanently delete the apiary and all its data.';

  @override
  String get apiaryGridTooSmall =>
      'The new grid is too small to fit all existing hives.';

  @override
  String apiaryGridHivesWillMove(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count hives will be relocated to fit the new grid.',
      one: '1 hive will be relocated to fit the new grid.',
    );
    return '$_temp0';
  }

  @override
  String get hiveTitle => 'Hives';

  @override
  String hiveCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count hives',
      one: '1 hive',
    );
    return '$_temp0';
  }

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

  @override
  String get hiveEdit => 'Edit hive';

  @override
  String get hiveDeleteConfirm => 'Delete hive?';

  @override
  String get hiveDeleteWarning => 'This will permanently delete the hive.';

  @override
  String hiveDefaultName(int index) {
    return 'Bee house $index';
  }

  @override
  String get hiveQueenless => 'Queenless';

  @override
  String get hiveReadyForHarvest => 'Ready for harvest';

  @override
  String get hiveSick => 'Sick';

  @override
  String get hiveFilterTooltip => 'Filter hives';

  @override
  String get hiveListTooltip => 'Hive list';

  @override
  String get apiaryCenterView => 'Center view';

  @override
  String get hiveDiseases => 'Diseases';

  @override
  String get hiveFrames => 'Frames';

  @override
  String get hiveFramesWarning =>
      'Frame count in inspection exceeds hive capacity';

  @override
  String get hiveStatus => 'Status';

  @override
  String get hiveDetailInspections => 'Inspections';

  @override
  String get hiveDetailNoInspections => 'No inspections yet';

  @override
  String get hiveDetailAddInspection => 'Add inspection';

  @override
  String get hiveDetailViewInspections => 'View all';

  @override
  String get hiveDetailTreatments => 'Treatments';

  @override
  String get hiveDetailNoTreatments => 'No active treatments';

  @override
  String get hiveDetailLogTreatment => 'Log treatment';

  @override
  String get hiveDetailHarvests => 'Harvests';

  @override
  String get hiveDetailNoHarvests => 'No harvests yet';

  @override
  String get hiveDetailLogHarvest => 'Log harvest';

  @override
  String get generalRequired => 'Required';

  @override
  String get generalLoadMore => 'Load more';

  @override
  String get inspectionTitle => 'Inspections';

  @override
  String get inspectionAdd => 'Add inspection';

  @override
  String get inspectionEdit => 'Edit inspection';

  @override
  String get inspectionDeleteConfirm => 'Delete inspection?';

  @override
  String get inspectionDeleteWarning =>
      'This will permanently delete the inspection.';

  @override
  String get inspectionEmpty => 'No inspections yet';

  @override
  String get inspectionDate => 'Inspection date';

  @override
  String get inspectionQueenSeen => 'Queen seen';

  @override
  String get inspectionQueenStatusSeen => 'Queen seen';

  @override
  String get inspectionQueenStatusNotSeen => 'Queen not seen';

  @override
  String get inspectionBroodPattern => 'Brood';

  @override
  String get inspectionBroodExcellent => 'Lots';

  @override
  String get inspectionBroodGood => 'Medium';

  @override
  String get inspectionBroodPoor => 'Little';

  @override
  String get inspectionBroodNone => 'None';

  @override
  String get inspectionAggressiveness => 'Aggressiveness';

  @override
  String get inspectionAggressivenessCalm => 'Calm';

  @override
  String get inspectionAggressivenessMild => 'Mild';

  @override
  String get inspectionAggressivenessAggressive => 'Aggressive';

  @override
  String get inspectionAggressivenessVeryAggressive => 'Very aggressive';

  @override
  String get inspectionFramesBrood => 'Brood frames';

  @override
  String get inspectionFramesHoney => 'Honey frames';

  @override
  String get inspectionFramesPollen => 'Pollen frames';

  @override
  String get inspectionFramesAddedDrawn => 'Added empty frames';

  @override
  String get inspectionFramesAddedFoundation => 'Added foundation';

  @override
  String get inspectionFramesAddedHoney => 'Added honey frames';

  @override
  String get inspectionQueenCellsCount => 'Queen cells';

  @override
  String get inspectionQueenAdded => 'Queen added';

  @override
  String get inspectionSectionObservations => 'Observations';

  @override
  String get inspectionSectionFrames => 'Frames';

  @override
  String get inspectionSectionHealth => 'Health';

  @override
  String get inspectionSectionHiveState => 'Hive state';

  @override
  String get inspectionNotes => 'Notes';

  @override
  String get inspectionNote => 'Note';

  @override
  String get inspectionDiseases => 'Diseases';

  @override
  String get inspectionDiseaseVarroa => 'Varroa';

  @override
  String get inspectionDiseaseNosema => 'Nosema';

  @override
  String get inspectionDiseaseDwv => 'DWV viruses';

  @override
  String get inspectionDiseaseAmericanFoulbrood => 'American foulbrood';

  @override
  String get inspectionDiseaseChalkbrood => 'Chalkbrood';

  @override
  String get inspectionDiseaseEuropeanFoulbrood => 'European foulbrood';

  @override
  String get inspectionDiseaseLayingWorkers => 'Laying workers';

  @override
  String get inspectionNotSet => 'Not set';

  @override
  String get inspectionPhotos => 'Photos';

  @override
  String get inspectionPhotoSourceGallery => 'Gallery';

  @override
  String get inspectionPhotoSourceCamera => 'Camera';

  @override
  String get inspectionAddPhoto => 'Add photo';

  @override
  String get inspectionNoPhotos => 'No photos yet';

  @override
  String get inspectionDeletePhoto => 'Delete photo?';

  @override
  String get inspectionDeletePhotoWarning =>
      'This will permanently delete the photo.';

  @override
  String inspectionPhotoCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count photos',
      one: '1 photo',
    );
    return '$_temp0';
  }
}
