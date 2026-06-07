import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../../l10n/app_localizations.dart';
import '../data/inspection_model.dart';

class InspectionSummary extends StatelessWidget {
  final Inspection inspection;
  final bool showDate;
  final String? currentUserName;

  const InspectionSummary({
    super.key,
    required this.inspection,
    this.showDate = false,
    this.currentUserName,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
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

    // Added frames — only show non-zero values
    final added = <String>[];
    if ((inspection.framesAddedDrawn ?? 0) > 0) {
      added.add('${l10n.inspectionFramesAddedDrawn}: ${inspection.framesAddedDrawn}');
    }
    if ((inspection.framesAddedFoundation ?? 0) > 0) {
      added.add('${l10n.inspectionFramesAddedFoundation}: ${inspection.framesAddedFoundation}');
    }
    if ((inspection.framesAddedHoney ?? 0) > 0) {
      added.add('${l10n.inspectionFramesAddedHoney}: ${inspection.framesAddedHoney}');
    }

    // Queen cells + queen added
    final queen = <String>[];
    if ((inspection.queenCellsCount ?? 0) > 0) {
      queen.add('${l10n.inspectionQueenCellsCount}: ${inspection.queenCellsCount}');
    }
    if (inspection.queenAdded) {
      queen.add(l10n.inspectionQueenAdded);
    }

    final otherInspector = currentUserName != null &&
            inspection.inspectedByName != null &&
            inspection.inspectedByName != currentUserName
        ? inspection.inspectedByName
        : null;

    final hasContent = showDate ||
        otherInspector != null ||
        obs.isNotEmpty ||
        frames.isNotEmpty ||
        added.isNotEmpty ||
        queen.isNotEmpty ||
        inspection.notes.isNotEmpty;
    if (!hasContent) return const SizedBox.shrink();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (showDate) ...[
          Text(
            DateFormat.yMMMd(
              Localizations.localeOf(context).toString(),
            ).add_Hm().format(inspection.inspectedAt),
            style: textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 2),
        ],
        if (otherInspector != null) ...[
          Text(
            l10n.inspectionInspectedBy(otherInspector),
            style: bodyStyle,
          ),
          const SizedBox(height: 6),
        ] else if (showDate)
          const SizedBox(height: 4),
        if (obs.isNotEmpty) ...[
          Text(l10n.inspectionSectionObservations, style: labelStyle),
          const SizedBox(height: 2),
          Text(obs.join(' · '), style: bodyStyle),
          const SizedBox(height: 8),
        ],
        if (frames.isNotEmpty || added.isNotEmpty) ...[
          Text(l10n.inspectionSectionFrames, style: labelStyle),
          const SizedBox(height: 2),
          if (frames.isNotEmpty) Text(frames.join(' · '), style: bodyStyle),
          if (frames.isNotEmpty && added.isNotEmpty) const SizedBox(height: 2),
          if (added.isNotEmpty)
            Text('+ ${added.join(' · ')}', style: bodyStyle),
          const SizedBox(height: 8),
        ],
        if (queen.isNotEmpty) ...[
          Text(queen.join(' · '), style: bodyStyle),
          const SizedBox(height: 8),
        ],
        if (inspection.notes.isNotEmpty) ...[
          Text(l10n.inspectionNote, style: labelStyle),
          const SizedBox(height: 2),
          Text(
            inspection.notes,
            style: bodyStyle,
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
