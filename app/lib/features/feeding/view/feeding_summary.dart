import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../../l10n/app_localizations.dart';
import '../data/feeding_model.dart';

class FeedingSummary extends StatelessWidget {
  final Feeding feeding;
  final AppLocalizations l10n;
  final bool showDate;
  final String? currentUserName;

  const FeedingSummary({
    super.key,
    required this.feeding,
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

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (showDate) ...[
          Text(
            DateFormat.yMMMd(
              Localizations.localeOf(context).toString(),
            ).add_Hm().format(feeding.fedAt),
            style: textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
          ),
          if (feeding.fedByName != null &&
              feeding.fedByName != currentUserName) ...[
            const SizedBox(height: 2),
            Text(
              l10n.feedingFedBy(feeding.fedByName!),
              style: bodyStyle,
            ),
          ],
          const SizedBox(height: 8),
        ],
        Text(l10n.feedingType, style: labelStyle),
        const SizedBox(height: 2),
        Text(
          '${feeding.feedType} · ${feeding.amount}',
          style: bodyStyle,
        ),
        if (feeding.notes.isNotEmpty) ...[
          const SizedBox(height: 8),
          Text(l10n.feedingNote, style: labelStyle),
          const SizedBox(height: 2),
          Text(
            feeding.notes,
            style: bodyStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ],
    );
  }
}
