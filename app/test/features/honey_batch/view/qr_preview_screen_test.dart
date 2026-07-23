import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:url_launcher_platform_interface/link.dart';
import 'package:url_launcher_platform_interface/url_launcher_platform_interface.dart';

import 'package:app/features/honey_batch/view/qr_preview_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _FakeUrlLauncherPlatform extends UrlLauncherPlatform {
  final List<String> launchedUrls = [];

  @override
  LinkDelegate? get linkDelegate => null;

  @override
  Future<bool> canLaunch(String url) async => true;

  @override
  Future<bool> launch(
    String url, {
    required bool useSafariVC,
    required bool useWebView,
    required bool enableJavaScript,
    required bool enableDomStorage,
    required bool universalLinksOnly,
    required Map<String, String> headers,
    String? webOnlyWindowName,
  }) async {
    launchedUrls.add(url);
    return true;
  }
}

void main() {
  late AppLocalizations l10n;
  late _FakeUrlLauncherPlatform fakeLauncher;

  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  setUp(() {
    fakeLauncher = _FakeUrlLauncherPlatform();
    UrlLauncherPlatform.instance = fakeLauncher;
  });

  Widget wrap() => MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Scaffold(
          body: Builder(
            builder: (context) => ElevatedButton(
              onPressed: () => showQrPreviewDialog(
                context,
                title: 'Batch #1 QR',
                imageUrl: 'https://example.com/qr.png',
                downloadUrl: 'https://example.com/qr.png?download=1',
                verificationUrl: 'https://example.com/verify/token-1',
              ),
              child: const Text('open'),
            ),
          ),
        ),
      );

  group('showQrPreviewDialog', () {
    testWidgets('shows the "open public verification page" button', (
      tester,
    ) async {
      await tester.pumpWidget(wrap());
      await tester.tap(find.text('open'));
      await tester.pump();

      expect(find.text(l10n.honeyBatchOpenPublicPage), findsOneWidget);

      final button = tester.widget<OutlinedButton>(
        find.ancestor(
          of: find.text(l10n.honeyBatchOpenPublicPage),
          matching: find.byType(OutlinedButton),
        ),
      );
      expect(button.onPressed, isNotNull);
    });

    testWidgets(
        'tapping the button launches the verification url', (tester) async {
      await tester.pumpWidget(wrap());
      await tester.tap(find.text('open'));
      await tester.pump();

      await tester.tap(find.text(l10n.honeyBatchOpenPublicPage));
      await tester.pump();

      expect(fakeLauncher.launchedUrls, ['https://example.com/verify/token-1']);
    });
  });
}
