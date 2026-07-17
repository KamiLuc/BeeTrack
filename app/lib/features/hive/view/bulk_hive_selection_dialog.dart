import 'package:flutter/material.dart';

import '../../../l10n/app_localizations.dart';
import '../data/hive_model.dart';

Future<List<int>?> showBulkHiveSelectionDialog(
  BuildContext context, {
  required List<Hive> hives,
}) {
  return showDialog<List<int>>(
    context: context,
    builder: (_) => _BulkHiveSelectionDialog(hives: hives),
  );
}

class _BulkHiveSelectionDialog extends StatefulWidget {
  final List<Hive> hives;

  const _BulkHiveSelectionDialog({required this.hives});

  @override
  State<_BulkHiveSelectionDialog> createState() =>
      _BulkHiveSelectionDialogState();
}

class _BulkHiveSelectionDialogState extends State<_BulkHiveSelectionDialog> {
  late Set<int> _selected = widget.hives.map((h) => h.id).toSet();

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isWide = MediaQuery.sizeOf(context).width >= 600;

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: isWide ? 480 : 380),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(24, 24, 8, 8),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      l10n.bulkSelectHives,
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close),
                    visualDensity: VisualDensity.compact,
                    onPressed: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
            ),
            const Divider(height: 1),
            Flexible(
              child: ListView(
                shrinkWrap: true,
                children: widget.hives
                    .map(
                      (h) => CheckboxListTile(
                        value: _selected.contains(h.id),
                        title: Text(h.name),
                        onChanged: (v) => setState(() {
                          if (v ?? false) {
                            _selected.add(h.id);
                          } else {
                            _selected.remove(h.id);
                          }
                        }),
                      ),
                    )
                    .toList(),
              ),
            ),
            const Divider(height: 1),
            Padding(
              padding: const EdgeInsets.all(16),
              child: SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _selected.isEmpty
                      ? null
                      : () => Navigator.of(context).pop(_selected.toList()),
                  child: Text(l10n.generalConfirm),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
