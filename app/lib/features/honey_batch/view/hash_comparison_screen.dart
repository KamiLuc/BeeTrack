import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../data/honey_batch_certification_model.dart';
import '../data/honey_batch_repository.dart';

enum _HashCheck { match, mismatch, unavailable }

_HashCheck _compare(String stored, String? onChain) {
  if (onChain == null || onChain.isEmpty) return _HashCheck.unavailable;
  return stored.toLowerCase() == onChain.toLowerCase()
      ? _HashCheck.match
      : _HashCheck.mismatch;
}

/// Shows a dialog comparing the batch's stored hashes (computed by our
/// backend when the batch was created) against the live on-chain values read
/// back from the blockchain, so the user can see at a glance whether the
/// record has been tampered with.
Future<void> showHashComparisonDialog(
  BuildContext context, {
  required String title,
  required int batchId,
  required String storedPdfHash,
  required String storedMetadataHash,
}) {
  final repo = HoneyBatchRepository(api: context.read<ApiClient>());
  return showDialog(
    context: context,
    builder: (dialogContext) => Dialog(
      insetPadding: const EdgeInsets.all(24),
      child: SizedBox(
        width: AppLayout.dialogWidth(context),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            AppBar(
              automaticallyImplyLeading: false,
              title: Text(title, overflow: TextOverflow.ellipsis),
              actions: [
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.of(dialogContext).pop(),
                ),
              ],
            ),
            Flexible(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: FutureBuilder<List<HoneyBatchCertificationModel>>(
                  future: repo.getCertifications(batchId),
                  builder: (context, snapshot) {
                    final l10n = AppLocalizations.of(context)!;
                    if (snapshot.connectionState != ConnectionState.done) {
                      return const Padding(
                        padding: EdgeInsets.all(24),
                        child: Center(child: CircularProgressIndicator()),
                      );
                    }
                    final items = snapshot.data ?? const [];
                    final onChainPdfHash =
                        items.isNotEmpty ? items.first.onChainPdfHash : null;
                    final onChainMetadataHash =
                        items.isNotEmpty ? items.first.onChainMetadataHash : null;
                    return Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _HashRow(
                          label: l10n.honeyBatchPdfHashLabel,
                          explainer: l10n.honeyBatchPdfHashExplainer,
                          stored: storedPdfHash,
                          onChain: onChainPdfHash,
                        ),
                        const SizedBox(height: 20),
                        _HashRow(
                          label: l10n.honeyBatchMetadataHashLabel,
                          explainer: l10n.honeyBatchMetadataHashExplainer,
                          stored: storedMetadataHash,
                          onChain: onChainMetadataHash,
                        ),
                      ],
                    );
                  },
                ),
              ),
            ),
          ],
        ),
      ),
    ),
  );
}

class _HashRow extends StatelessWidget {
  final String label;
  final String explainer;
  final String stored;
  final String? onChain;

  const _HashRow({
    required this.label,
    required this.explainer,
    required this.stored,
    required this.onChain,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    final check = _compare(stored, onChain);
    final (checkLabel, checkColor) = switch (check) {
      _HashCheck.match => (l10n.honeyBatchHashMatch, Colors.green),
      _HashCheck.mismatch => (l10n.honeyBatchHashMismatch, colorScheme.error),
      _HashCheck.unavailable => (
          l10n.honeyBatchHashCheckUnavailable,
          colorScheme.onSurfaceVariant,
        ),
    };

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w600)),
        const SizedBox(height: 2),
        Text(
          explainer,
          style: textTheme.bodySmall?.copyWith(color: colorScheme.onSurfaceVariant),
        ),
        const SizedBox(height: 8),
        _HashValue(label: l10n.honeyBatchStoredHash, value: stored),
        const SizedBox(height: 8),
        _HashValue(label: l10n.honeyBatchOnChainHash, value: onChain ?? '—'),
        const SizedBox(height: 6),
        Row(
          children: [
            Icon(
              switch (check) {
                _HashCheck.match => Icons.check_circle,
                _HashCheck.mismatch => Icons.error,
                _HashCheck.unavailable => Icons.help_outline,
              },
              size: 16,
              color: checkColor,
            ),
            const SizedBox(width: 6),
            Flexible(
              child: Text(
                checkLabel,
                style: textTheme.bodySmall?.copyWith(
                  color: checkColor,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }
}

class _HashValue extends StatelessWidget {
  final String label;
  final String value;

  const _HashValue({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: textTheme.labelSmall?.copyWith(color: colorScheme.onSurfaceVariant),
        ),
        SelectableText(
          value,
          style: textTheme.bodySmall?.copyWith(fontFamily: 'monospace'),
        ),
      ],
    );
  }
}
