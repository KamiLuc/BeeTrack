import 'package:flutter/material.dart';

import '../../l10n/app_localizations.dart';

/// Returns true if [text] would exceed [maxLines] when rendered with [style]
/// inside [maxWidth] — i.e. it would be cut off with an ellipsis. Used to
/// decide whether a note preview should be clickable to see the full text.
bool isTextTruncated(
  String text,
  TextStyle? style,
  double maxWidth, {
  int maxLines = 2,
}) {
  if (maxWidth <= 0) return false;
  final painter = TextPainter(
    text: TextSpan(text: text, style: style),
    maxLines: maxLines,
    textDirection: TextDirection.ltr,
  )..layout(maxWidth: maxWidth);
  return painter.didExceedMaxLines;
}

/// Shows a read-only dialog with [title] and [note]. Used by record cards
/// whose note preview is truncated, so the full text stays reachable.
Future<void> showNoteDialog(
  BuildContext context, {
  required String title,
  required String note,
}) {
  final l10n = AppLocalizations.of(context)!;
  return showDialog<void>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: Text(title),
      content: SingleChildScrollView(child: Text(note)),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(ctx).pop(),
          child: Text(l10n.generalClose),
        ),
      ],
    ),
  );
}
