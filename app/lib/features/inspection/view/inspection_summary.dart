import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../../l10n/app_localizations.dart';
import '../data/inspection_model.dart';

class InspectionSummary extends StatelessWidget {
  final Inspection inspection;
  final bool showDate;

  const InspectionSummary({
    super.key,
    required this.inspection,
    this.showDate = false,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    final style = textTheme.bodyMedium?.copyWith(
      color: colorScheme.onSurfaceVariant,
    );

    // Grouped rows of plain text
    final rows = <String>[];

    // Observations: queen · brood · aggressiveness
    final obs = <String>[];
    if (inspection.queenSeen.isNotEmpty) {
      obs.add(
        inspection.queenSeen == 'seen'
            ? l10n.inspectionQueenStatusSeen
            : l10n.inspectionQueenStatusNotSeen,
      );
    }
    if (inspection.broodPattern.isNotEmpty) {
      obs.add('${l10n.inspectionBroodPattern}: ${_broodLabel(l10n, inspection.broodPattern)}');
    }
    if (inspection.aggressiveness.isNotEmpty) {
      obs.add(_aggressivenessLabel(l10n, inspection.aggressiveness));
    }
    if (obs.isNotEmpty) rows.add(obs.join(' · '));

    // Current frame counts
    final frames = <String>[];
    if (inspection.framesBrood != null) {
      frames.add('${l10n.inspectionFramesBrood}: ${inspection.framesBrood}');
    }
    if (inspection.framesHoney != null) {
      frames.add('${l10n.inspectionFramesHoney}: ${inspection.framesHoney}');
    }
    if (inspection.framesPollen != null) {
      frames.add('${l10n.inspectionFramesPollen}: ${inspection.framesPollen}');
    }
    if (frames.isNotEmpty) rows.add(frames.join(' · '));

    // Added frames
    final added = <String>[];
    if (inspection.framesAddedDrawn != null) {
      added.add('${l10n.inspectionFramesAddedDrawn}: ${inspection.framesAddedDrawn}');
    }
    if (inspection.framesAddedFoundation != null) {
      added.add('${l10n.inspectionFramesAddedFoundation}: ${inspection.framesAddedFoundation}');
    }
    if (inspection.framesAddedHoney != null) {
      added.add('${l10n.inspectionFramesAddedHoney}: ${inspection.framesAddedHoney}');
    }
    if (added.isNotEmpty) rows.add(added.join(' · '));

    // Queen cells + queen added
    final queen = <String>[];
    if ((inspection.queenCellsCount ?? 0) > 0) {
      queen.add('${l10n.inspectionQueenCellsCount}: ${inspection.queenCellsCount}');
    }
    if (inspection.queenAdded) {
      queen.add(l10n.inspectionQueenAdded);
    }
    if (queen.isNotEmpty) rows.add(queen.join(' · '));

    final hasContent = showDate || rows.isNotEmpty || inspection.notes.isNotEmpty;
    if (!hasContent) return const SizedBox.shrink();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (showDate) ...[
          Text(
            DateFormat.yMMMd(
              Localizations.localeOf(context).toString(),
            ).format(inspection.inspectedAt),
            style: textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
          ),
          if (rows.isNotEmpty || inspection.notes.isNotEmpty) const SizedBox(height: 2),
        ],
        for (int i = 0; i < rows.length; i++) ...[
          if (i > 0) const SizedBox(height: 2),
          Text(rows[i], style: style),
        ],
        if (inspection.notes.isNotEmpty) ...[
          if (rows.isNotEmpty) const SizedBox(height: 2),
          Text(
            '${l10n.inspectionNote}: ${inspection.notes}',
            style: style,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ],
    );
  }
}

String _broodLabel(AppLocalizations l10n, String v) => switch (v) {
      'none' => l10n.inspectionBroodNone,
      'poor' => l10n.inspectionBroodPoor,
      'good' => l10n.inspectionBroodGood,
      'excellent' => l10n.inspectionBroodExcellent,
      _ => v,
    };

String _aggressivenessLabel(AppLocalizations l10n, String v) => switch (v) {
      'calm' => l10n.inspectionAggressivenessCalm,
      'mild' => l10n.inspectionAggressivenessMild,
      'aggressive' => l10n.inspectionAggressivenessAggressive,
      'very_aggressive' => l10n.inspectionAggressivenessVeryAggressive,
      _ => v,
    };
