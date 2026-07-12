import '../../../l10n/app_localizations.dart';

const List<String> listingCategories = [
  'HONEY',
  'POLLEN',
  'BEE_COLONIES',
  'QUEEN_BEES',
  'BEEHIVES',
  'POPULATED_BEEHIVES',
  'EQUIPMENT',
  'EXTRACTION_EQUIPMENT',
  'FEED',
  'SUPPLIES',
  'WAX_FOUNDATION',
  'BEESWAX',
  'PROPOLIS',
  'SERVICES',
  'OTHER',
];

String listingCategoryLabel(AppLocalizations l10n, String category) {
  return switch (category) {
    'HONEY' => l10n.marketplaceCategoryHoney,
    'POLLEN' => l10n.marketplaceCategoryPollen,
    'BEE_COLONIES' => l10n.marketplaceCategoryBeeColonies,
    'QUEEN_BEES' => l10n.marketplaceCategoryQueenBees,
    'BEEHIVES' => l10n.marketplaceCategoryBeehives,
    'POPULATED_BEEHIVES' => l10n.marketplaceCategoryPopulatedBeehives,
    'EQUIPMENT' => l10n.marketplaceCategoryEquipment,
    'EXTRACTION_EQUIPMENT' => l10n.marketplaceCategoryExtractionEquipment,
    'FEED' => l10n.marketplaceCategoryFeed,
    'SUPPLIES' => l10n.marketplaceCategorySupplies,
    'WAX_FOUNDATION' => l10n.marketplaceCategoryWaxFoundation,
    'BEESWAX' => l10n.marketplaceCategoryBeeswax,
    'PROPOLIS' => l10n.marketplaceCategoryPropolis,
    'SERVICES' => l10n.marketplaceCategoryServices,
    _ => l10n.marketplaceCategoryOther,
  };
}
