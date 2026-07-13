import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/marketplace/data/listing_price.dart';
import 'package:app/l10n/app_localizations.dart';

void main() {
  group('listingPriceLabel', () {
    late AppLocalizations l10n;

    setUpAll(() async {
      l10n = await AppLocalizations.delegate.load(const Locale('en'));
    });

    test('returns "Price on request" when price is null', () {
      expect(listingPriceLabel(l10n, null), l10n.marketplacePriceOnRequest);
    });

    test('returns "Free" when price is exactly 0', () {
      expect(listingPriceLabel(l10n, 0), l10n.marketplacePriceFree);
    });

    test('returns formatted price otherwise', () {
      expect(listingPriceLabel(l10n, 42.5), '42.50');
    });
  });
}
