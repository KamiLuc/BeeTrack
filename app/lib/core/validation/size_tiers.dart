import 'package:image_picker/image_picker.dart';

import '../../l10n/app_localizations.dart';

/// A named text-length limit shared across all form validation, so fields
/// pick a tier (e.g. [SizeTier.small]) instead of a hand-picked character count.
enum SizeTier { tiny, superSmall, small, medium, large, extraLarge }

extension SizeTierLimit on SizeTier {
  /// The maximum character count allowed for this tier. [SizeTier.tiny] is
  /// frontend-only (short numeric-text fields like frame counts); every
  /// other tier mirrors the limits enforced by the backend's
  /// `internal/validation` package.
  int get maxLength => switch (this) {
        SizeTier.tiny => 2,
        SizeTier.superSmall => 20,
        SizeTier.small => 50,
        SizeTier.medium => 150,
        SizeTier.large => 500,
        SizeTier.extraLarge => 5000,
      };

  /// The x/N counter is only shown for the larger tiers, where the limit
  /// isn't obvious at a glance; short fields would just add visual noise.
  bool get showsCounter => this == SizeTier.large || this == SizeTier.extraLarge;

  /// `counterText` value for an [InputDecoration]: `null` shows the default
  /// x/N counter, `''` hides it entirely.
  String? get counterText => showsCounter ? null : '';
}

/// Returns the localized "too long" error if [value] exceeds [tier], else null.
/// [fieldLabel] names the field in the error message (e.g. l10n.marketplaceFieldTitle).
String? validateSizeTier(
  String? value,
  SizeTier tier,
  String fieldLabel,
  AppLocalizations l10n,
) {
  if (value == null || value.length <= tier.maxLength) return null;
  return l10n.generalFieldTooLong(fieldLabel, tier.maxLength);
}

/// Matches bcrypt's effective input limit (72 bytes) on the backend — passwords
/// longer than this are silently truncated by bcrypt, so anything past it is
/// pointless and rejected outright instead.
const int maxPasswordLength = 72;

/// Largest amount of honey (kg) a single harvest record may claim. Mirrors
/// `maxHarvestKilograms` in the backend's `internal/service/harvest.go`.
const double maxHarvestKilograms = 1000.0;

/// Returns the localized "too long" error if [value] exceeds [maxPasswordLength].
String? validatePasswordLength(String? value, AppLocalizations l10n) {
  if (value == null || value.length <= maxPasswordLength) return null;
  return l10n.generalFieldTooLong(l10n.authPassword, maxPasswordLength);
}

/// Largest image file accepted for inspection and listing photo uploads.
/// Mirrors `MaxImageBytes` in the backend's `internal/service/inspection_image.go`.
const int maxImageBytes = 5 * 1024 * 1024;

/// Human-readable form of [maxImageBytes] for error messages, e.g. "5 MB".
final String maxImageSizeLabel = '${maxImageBytes ~/ (1024 * 1024)} MB';

/// Returns the localized "photo too large" error if [file] exceeds [maxImageBytes], else null.
Future<String?> validateImageFileSize(XFile file, AppLocalizations l10n) async {
  final size = await file.length();
  if (size <= maxImageBytes) return null;
  return l10n.generalPhotoTooLarge(maxImageSizeLabel);
}

/// Largest lab PDF file accepted for honey batch uploads. Mirrors
/// `maxLabPDFBytes` in the backend's `internal/service/honey_batch.go`.
const int maxPdfBytes = 10 * 1024 * 1024;

/// Human-readable form of [maxPdfBytes] for error messages, e.g. "10 MB".
final String maxPdfSizeLabel = '${maxPdfBytes ~/ (1024 * 1024)} MB';

/// Returns the localized "PDF too large" error if [byteLength] exceeds [maxPdfBytes], else null.
String? validatePdfFileSize(int byteLength, AppLocalizations l10n) {
  if (byteLength <= maxPdfBytes) return null;
  return l10n.generalPdfTooLarge(maxPdfSizeLabel);
}
