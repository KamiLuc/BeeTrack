import 'package:flutter/material.dart';

import '../../../l10n/app_localizations.dart';

const List<String> listingCategories = [
  'HONEY',
  'POLLEN',
  'BEE_COLONIES',
  'QUEEN_BEES',
  'BEEHIVES',
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

IconData listingCategoryIcon(String? category) {
  return switch (category) {
    'HONEY' => Icons.water_drop_outlined,
    'POLLEN' => Icons.grain_outlined,
    'BEE_COLONIES' => Icons.emoji_nature_outlined,
    'QUEEN_BEES' => Icons.workspace_premium_outlined,
    'BEEHIVES' => Icons.hive_outlined,
    'EQUIPMENT' => Icons.construction_outlined,
    'EXTRACTION_EQUIPMENT' => Icons.autorenew,
    'FEED' => Icons.restaurant_outlined,
    'SUPPLIES' => Icons.inventory_2_outlined,
    'WAX_FOUNDATION' => Icons.grid_4x4_outlined,
    'BEESWAX' => Icons.local_fire_department_outlined,
    'PROPOLIS' => Icons.science_outlined,
    'SERVICES' => Icons.handyman_outlined,
    'OTHER' => Icons.more_horiz,
    _ => Icons.apps_outlined,
  };
}
