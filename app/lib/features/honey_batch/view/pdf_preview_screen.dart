import 'dart:math';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:printing/printing.dart';

/// Shows [bytes] as a PDF in a dialog on top of the current screen, instead
/// of navigating to a separate full-screen route.
Future<void> showPdfPreviewDialog(
  BuildContext context, {
  required String title,
  required Uint8List bytes,
}) {
  final size = MediaQuery.sizeOf(context);
  return showDialog(
    context: context,
    builder: (dialogContext) => Dialog(
      insetPadding: const EdgeInsets.all(24),
      child: SizedBox(
        // A PDF page reads more comfortably wider than the app's usual
        // dialog cap, so this uses its own (larger) ceiling instead of
        // AppLayout.dialogWidth.
        width: min(800.0, size.width * 0.9),
        height: size.height * 0.85,
        child: Column(
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
            Expanded(
              child: PdfPreview(
                build: (_) async => bytes,
                canChangeOrientation: false,
                canChangePageFormat: false,
                canDebug: false,
                allowPrinting: false,
                allowSharing: false,
                actions: [
                  PdfShareAction(
                    icon: const Icon(Icons.download),
                    filename: title,
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    ),
  );
}
