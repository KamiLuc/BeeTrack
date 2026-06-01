import 'package:flutter/material.dart';

import '../../../l10n/app_localizations.dart';

class ApiaryGridSection extends StatelessWidget {
  final int rows;
  final int cols;
  final ValueChanged<int> onRowsChanged;
  final ValueChanged<int> onColsChanged;
  final AppLocalizations l10n;

  const ApiaryGridSection({
    super.key,
    required this.rows,
    required this.cols,
    required this.onRowsChanged,
    required this.onColsChanged,
    required this.l10n,
  });

  @override
  Widget build(BuildContext context) {
    final items = List.generate(25, (i) => i + 1)
        .map((v) => DropdownMenuItem(value: v, child: Text('$v')))
        .toList();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            Expanded(
              child: DropdownButtonFormField<int>(
                initialValue: rows,
                menuMaxHeight: 48 * 10,
                decoration: InputDecoration(labelText: l10n.apiaryGridRows),
                items: items,
                onChanged: (v) { if (v != null) onRowsChanged(v); },
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: DropdownButtonFormField<int>(
                initialValue: cols,
                menuMaxHeight: 48 * 10,
                decoration: InputDecoration(labelText: l10n.apiaryGridCols),
                items: items,
                onChanged: (v) { if (v != null) onColsChanged(v); },
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        _GridPreview(rows: rows, cols: cols),
      ],
    );
  }
}

class _GridPreview extends StatelessWidget {
  final int rows;
  final int cols;

  const _GridPreview({required this.rows, required this.cols});

  @override
  Widget build(BuildContext context) {
    const maxWidth = 240.0;
    final cellSize = (maxWidth / cols).clamp(4.0, 24.0);

    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: List.generate(rows, (r) {
          return Row(
            mainAxisSize: MainAxisSize.min,
            children: List.generate(cols, (c) {
              return Container(
                width: cellSize,
                height: cellSize,
                margin: const EdgeInsets.all(1),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.primaryContainer,
                  borderRadius: BorderRadius.circular(2),
                ),
              );
            }),
          );
        }),
      ),
    );
  }
}

class ApiaryLocationSection extends StatelessWidget {
  final TextEditingController latController;
  final TextEditingController lngController;
  final bool locating;
  final VoidCallback onGps;
  final VoidCallback onMap;
  final AppLocalizations l10n;

  const ApiaryLocationSection({
    super.key,
    required this.latController,
    required this.lngController,
    required this.locating,
    required this.onGps,
    required this.onMap,
    required this.l10n,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextFormField(
          controller: latController,
          enabled: false,
          decoration: InputDecoration(labelText: l10n.apiaryLatitude),
        ),
        const SizedBox(height: 12),
        TextFormField(
          controller: lngController,
          enabled: false,
          decoration: InputDecoration(labelText: l10n.apiaryLongitude),
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: OutlinedButton.icon(
                onPressed: locating ? null : onGps,
                icon: locating
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.my_location, size: 18),
                label: const Text('GPS'),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: OutlinedButton.icon(
                onPressed: onMap,
                icon: const Icon(Icons.map, size: 18),
                label: const Text('Mapa'),
              ),
            ),
          ],
        ),
      ],
    );
  }
}
