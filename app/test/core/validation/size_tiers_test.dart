import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/core/validation/size_tiers.dart';
import 'package:app/l10n/app_localizations.dart';

void main() {
  group('validatePdfFileSize', () {
    late AppLocalizations l10n;

    setUpAll(() async {
      l10n = await AppLocalizations.delegate.load(const Locale('en'));
    });

    test('returns null when byteLength is at or below the limit', () {
      expect(validatePdfFileSize(maxPdfBytes, l10n), isNull);
      expect(validatePdfFileSize(1024, l10n), isNull);
    });

    test('returns the localized error when byteLength exceeds the limit', () {
      expect(
        validatePdfFileSize(maxPdfBytes + 1, l10n),
        l10n.generalPdfTooLarge(maxPdfSizeLabel),
      );
    });
  });
}
