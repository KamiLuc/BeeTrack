import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../../l10n/app_localizations.dart';
import '../data/treatment_model.dart';

class TreatmentSummary extends StatelessWidget {
  final Treatment treatment;
  final AppLocalizations l10n;
  final bool showDate;
  final String? currentUserName;

  const TreatmentSummary({
    super.key,
    required this.treatment,
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

    final doseCount = int.tryParse(treatment.dose);
    final doseDisplay = doseCount != null
        ? l10n.treatmentDoseCount(doseCount)
        : treatment.dose;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (showDate) ...[
          Text(
            DateFormat.yMMMd(
              Localizations.localeOf(context).toString(),
            ).add_Hm().format(treatment.treatedAt),
            style: textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
          ),
          if (treatment.treatedByName != null &&
              treatment.treatedByName != currentUserName) ...[
            const SizedBox(height: 2),
            Text(
              l10n.treatmentTreatedBy(treatment.treatedByName!),
              style: bodyStyle,
            ),
          ],
          const SizedBox(height: 8),
        ],
        Text(l10n.treatmentMedicine, style: labelStyle),
        const SizedBox(height: 2),
        Text(
          '${treatment.medicineName} · $doseDisplay',
          style: bodyStyle,
        ),
        if (treatment.notes.isNotEmpty) ...[
          const SizedBox(height: 8),
          Text(l10n.treatmentNote, style: labelStyle),
          const SizedBox(height: 2),
          Text(
            treatment.notes,
            style: bodyStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ],
    );
  }
}
