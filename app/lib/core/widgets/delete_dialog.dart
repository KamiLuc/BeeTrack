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
  String? confirmLabel,
}) async {
  if (!withPuzzle) {
    return await _showSimpleConfirm(context,
        title: title, warning: warning, l10n: l10n, confirmLabel: confirmLabel);
  }
  return await _showPuzzleConfirm(context,
      title: title, warning: warning, l10n: l10n, confirmLabel: confirmLabel);
}

Future<bool> _showSimpleConfirm(
  BuildContext context, {
  required String title,
  required String warning,
  required AppLocalizations l10n,
  String? confirmLabel,
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
            confirmLabel ?? l10n.generalDelete,
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
  String? confirmLabel,
}) async {
  final result = await showDialog<bool>(
    context: context,
    builder: (ctx) => _PuzzleDialog(
      title: title,
      warning: warning,
      l10n: l10n,
      confirmLabel: confirmLabel,
    ),
  );
  return result ?? false;
}

class _PuzzleDialog extends StatefulWidget {
  final String title;
  final String warning;
  final AppLocalizations l10n;
  final String? confirmLabel;

  const _PuzzleDialog({
    required this.title,
    required this.warning,
    required this.l10n,
    this.confirmLabel,
  });

  @override
  State<_PuzzleDialog> createState() => _PuzzleDialogState();
}

class _PuzzleDialogState extends State<_PuzzleDialog> {
  late final int _a;
  late final int _b;
  late final int _expected;
  final _controller = TextEditingController();
  String? _error;

  @override
  void initState() {
    super.initState();
    final rng = Random();
    _a = rng.nextInt(9) + 1;
    _b = rng.nextInt(9) + 1;
    _expected = _a + _b;
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _tryConfirm() {
    if (int.tryParse(_controller.text.trim()) == _expected) {
      Navigator.of(context).pop(true);
    } else {
      setState(() => _error = widget.l10n.deletePuzzleWrong);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = widget.l10n;
    return AlertDialog(
      title: Text(widget.title),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(widget.warning),
          const SizedBox(height: 16),
          Text(l10n.deletePuzzlePrompt),
          const SizedBox(height: 8),
          Row(
            children: [
              Text(
                '$_a + $_b = ',
                style: const TextStyle(fontWeight: FontWeight.w600),
              ),
              Expanded(
                child: TextField(
                  controller: _controller,
                  keyboardType: TextInputType.number,
                  inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                  decoration: InputDecoration(
                    isDense: true,
                    border: const OutlineInputBorder(),
                    errorText: _error,
                  ),
                  autofocus: true,
                  onChanged: (_) {
                    if (_error != null) setState(() => _error = null);
                  },
                  onSubmitted: (_) => _tryConfirm(),
                ),
              ),
            ],
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(false),
          child: Text(l10n.generalCancel),
        ),
        TextButton(
          onPressed: _tryConfirm,
          child: Text(
            widget.confirmLabel ?? l10n.generalDelete,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ),
      ],
    );
  }
}
