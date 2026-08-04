import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../../l10n/app_localizations.dart';
import '../data/harvest_model.dart';

class HarvestSummary extends StatelessWidget {
  final Harvest harvest;
  final AppLocalizations l10n;
  final bool showDate;
  final String? currentUserName;

  const HarvestSummary({
    super.key,
    required this.harvest,
    required this.l10n,
    this.showDate = true,
    this.currentUserName,
  });

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    final bodyStyle = textTheme.bodyMedium?.copyWith(
      color: colorScheme.onSurfaceVariant,
    );
    final labelStyle = textTheme.labelSmall?.copyWith(
      color: colorScheme.primary,
      fontWeight: FontWeight.w600,
      letterSpacing: 0.4,
    );

    final otherHarvester =
        harvest.harvestedByName != null &&
                harvest.harvestedByName!.isNotEmpty &&
                harvest.harvestedByName != currentUserName
            ? harvest.harvestedByName
            : null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (showDate) ...[
          Text(
            DateFormat.yMMMd(
              Localizations.localeOf(context).toString(),
            ).add_Hm().format(harvest.harvestedAt),
            style: textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
          ),
          if (otherHarvester != null) ...[
            const SizedBox(height: 2),
            Text(
              l10n.harvestHarvestedBy(otherHarvester),
              style: bodyStyle,
            ),
          ],
          const SizedBox(height: 8),
        ],
        Text(l10n.harvestFrames, style: labelStyle),
        const SizedBox(height: 2),
        Text(
          harvest.halfFrames > 0
              ? '${l10n.harvestFramesCount(harvest.frames)} + ${l10n.harvestHalfFramesCount(harvest.halfFrames)}'
              : l10n.harvestFramesCount(harvest.frames),
          style: bodyStyle,
        ),
        const SizedBox(height: 8),
        Text(l10n.harvestKilograms, style: labelStyle),
        const SizedBox(height: 2),
        Text(
          '${harvest.kilograms.toStringAsFixed(2)} kg',
          style: bodyStyle,
        ),
        if (harvest.notes.isNotEmpty) ...[
          const SizedBox(height: 8),
          Text(l10n.harvestNote, style: labelStyle),
          const SizedBox(height: 2),
          Text(
            harvest.notes,
            style: bodyStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ],
    );
  }
}
