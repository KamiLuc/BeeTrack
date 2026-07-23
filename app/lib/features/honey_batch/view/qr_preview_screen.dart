import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../l10n/app_localizations.dart';

/// Shows the QR code at [imageUrl] in a dialog on top of the current screen,
/// with a download action pointed at [downloadUrl] and a link to
/// [verificationUrl] — the same public page the QR code itself points to.
Future<void> showQrPreviewDialog(
  BuildContext context, {
  required String title,
  required String imageUrl,
  required String downloadUrl,
  required String verificationUrl,
}) {
  final l10n = AppLocalizations.of(context)!;
  final size = MediaQuery.sizeOf(context);
  return showDialog(
    context: context,
    builder: (dialogContext) => Dialog(
      insetPadding: const EdgeInsets.all(24),
      child: SizedBox(
        width: size.width * 0.9,
        height: size.height * 0.7,
        child: Column(
          children: [
            AppBar(
              automaticallyImplyLeading: false,
              title: Text(title, overflow: TextOverflow.ellipsis),
              actions: [
                IconButton(
                  icon: const Icon(Icons.download),
                  tooltip: l10n.honeyBatchDownloadQr,
                  onPressed: () => launchQrDownload(downloadUrl),
                ),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.of(dialogContext).pop(),
                ),
              ],
            ),
            Expanded(
              child: Center(
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Image.network(
                    imageUrl,
                    fit: BoxFit.contain,
                    errorBuilder: (context, error, stack) => Center(
                      child: Text(l10n.generalError),
                    ),
                    loadingBuilder: (context, child, progress) {
                      if (progress == null) return child;
                      return const Center(child: CircularProgressIndicator());
                    },
                  ),
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
              child: SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  icon: const Icon(Icons.open_in_new),
                  label: Text(l10n.honeyBatchOpenPublicPage),
                  onPressed: () => launchVerificationPage(verificationUrl),
                ),
              ),
            ),
          ],
        ),
      ),
    ),
  );
}

Future<void> launchQrDownload(String downloadUrl) {
  return launchUrl(Uri.parse(downloadUrl), webOnlyWindowName: '_blank');
}

Future<void> launchVerificationPage(String verificationUrl) {
  return launchUrl(Uri.parse(verificationUrl), webOnlyWindowName: '_blank');
}
