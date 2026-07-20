import 'package:flutter/material.dart';

import '../../../l10n/app_localizations.dart';
import '../data/honey_batch_certification_model.dart';

class CertificationStatusBadge extends StatelessWidget {
  final CertificationStatus? status;

  const CertificationStatusBadge({super.key, required this.status});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;

    if (status == null) {
      return _badge(
        context,
        label: l10n.honeyBatchNotCertified,
        bg: colorScheme.surfaceContainerHighest,
        fg: colorScheme.onSurfaceVariant,
      );
    }

    final (label, color) = switch (status!) {
      CertificationStatus.queued => (l10n.honeyBatchStatusQueued, Colors.amber),
      CertificationStatus.submitting => (l10n.honeyBatchStatusSubmitting, Colors.amber),
      CertificationStatus.submitted => (l10n.honeyBatchStatusSubmitted, Colors.amber),
      CertificationStatus.pendingConfirmation =>
        (l10n.honeyBatchStatusPendingConfirmation, Colors.amber),
      CertificationStatus.confirmed => (l10n.honeyBatchStatusConfirmed, Colors.green),
      CertificationStatus.failed => (l10n.honeyBatchStatusFailed, colorScheme.error),
      CertificationStatus.reverted => (l10n.honeyBatchStatusReverted, colorScheme.error),
    };

    return _badge(context, label: label, bg: color.withValues(alpha: 0.15), fg: color);
  }

  Widget _badge(
    BuildContext context, {
    required String label,
    required Color bg,
    required Color fg,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(color: bg, borderRadius: BorderRadius.circular(20)),
      child: Text(
        label,
        style: Theme.of(context)
            .textTheme
            .labelSmall
            ?.copyWith(color: fg, fontWeight: FontWeight.w600),
      ),
    );
  }
}
