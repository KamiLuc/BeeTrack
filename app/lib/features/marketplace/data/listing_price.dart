import '../../../l10n/app_localizations.dart';

String listingPriceLabel(AppLocalizations l10n, double? price) {
  if (price == null) return l10n.marketplacePriceOnRequest;
  if (price == 0) return l10n.marketplacePriceFree;
  return price.toStringAsFixed(2);
}
