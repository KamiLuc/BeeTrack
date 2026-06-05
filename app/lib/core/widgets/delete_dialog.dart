import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../l10n/app_localizations.dart';

/// Shows a delete confirmation dialog.
/// When [withPuzzle] is true the user must solve a simple addition before
/// the delete action is confirmed.
Future<bool> showDeleteDialog(
  BuildContext context, {
  required String title,
  required String warning,
  required AppLocalizations l10n,
  bool withPuzzle = false,
}) async {
  if (!withPuzzle) {
    return await _showSimpleConfirm(context,
        title: title, warning: warning, l10n: l10n);
  }
  return await _showPuzzleConfirm(context,
      title: title, warning: warning, l10n: l10n);
}

Future<bool> _showSimpleConfirm(
  BuildContext context, {
  required String title,
  required String warning,
  required AppLocalizations l10n,
}) async {
  final result = await showDialog<bool>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: Text(title),
      content: Text(warning),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(ctx).pop(false),
          child: Text(l10n.generalCancel),
        ),
        TextButton(
          onPressed: () => Navigator.of(ctx).pop(true),
          child: Text(
            l10n.generalDelete,
            style: TextStyle(color: Theme.of(ctx).colorScheme.error),
          ),
        ),
      ],
    ),
  );
  return result ?? false;
}

Future<bool> _showPuzzleConfirm(
  BuildContext context, {
  required String title,
  required String warning,
  required AppLocalizations l10n,
}) async {
  final rng = Random();
  final a = rng.nextInt(9) + 1;
  final b = rng.nextInt(9) + 1;
  final expected = a + b;
  final controller = TextEditingController();
  String? inputError;

  final result = await showDialog<bool>(
    context: context,
    builder: (ctx) => StatefulBuilder(
      builder: (ctx, setState) => AlertDialog(
        title: Text(title),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(warning),
            const SizedBox(height: 16),
            Text(l10n.deletePuzzlePrompt),
            const SizedBox(height: 8),
            Row(
              children: [
                Text(
                  '$a + $b = ',
                  style: const TextStyle(fontWeight: FontWeight.w600),
                ),
                Expanded(
                  child: TextField(
                    controller: controller,
                    keyboardType: TextInputType.number,
                    inputFormatters: [
                      FilteringTextInputFormatter.digitsOnly,
                    ],
                    decoration: InputDecoration(
                      isDense: true,
                      border: const OutlineInputBorder(),
                      errorText: inputError,
                    ),
                    autofocus: true,
                    onSubmitted: (_) {
                      if (int.tryParse(controller.text.trim()) == expected) {
                        Navigator.of(ctx).pop(true);
                      } else {
                        setState(() => inputError = l10n.deletePuzzleWrong);
                      }
                    },
                  ),
                ),
              ],
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: Text(l10n.generalCancel),
          ),
          TextButton(
            onPressed: () {
              if (int.tryParse(controller.text.trim()) == expected) {
                Navigator.of(ctx).pop(true);
              } else {
                setState(() => inputError = l10n.deletePuzzleWrong);
              }
            },
            child: Text(
              l10n.generalDelete,
              style: TextStyle(color: Theme.of(ctx).colorScheme.error),
            ),
          ),
        ],
      ),
    ),
  );
  controller.dispose();
  return result ?? false;
}
