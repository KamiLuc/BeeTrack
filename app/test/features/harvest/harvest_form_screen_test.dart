import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/harvest/view/harvest_form_screen.dart';
import 'package:app/features/hive/data/hive_model.dart';
import 'package:app/l10n/app_localizations.dart';

Future<ApiClient> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  return ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
}

Widget _wrap(ApiClient apiClient, Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: RepositoryProvider<ApiClient>.value(
        value: apiClient,
        child: child,
      ),
    );

const _hive = Hive(
  id: 10,
  apiaryId: 1,
  name: 'Alpha',
  type: 'langstroth',
  active: true,
  queenNeedsReplacement: false,
  readyForHarvest: false,
  needsFood: false,
  boxNeedsAdding: false,
  gridRow: 0,
  gridCol: 0,
);

void main() {
  group('HarvestFormScreen onSaveProposed', () {
    testWidgets(
        'calls onSaveProposed with form state and pops true, without hitting the repository',
        (tester) async {
      final apiClient = await _fakeApiClient();
      Map<String, dynamic>? received;

      await tester.pumpWidget(_wrap(
        apiClient,
        Navigator(
          onGenerateRoute: (settings) => MaterialPageRoute(
            builder: (_) => HarvestFormScreen(
              apiaryId: 1,
              hive: _hive,
              onSaveProposed: (args) async {
                received = args;
              },
            ),
          ),
        ),
      ));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextFormField).at(0), '10');
      await tester.enterText(find.byType(TextFormField).at(1), '2');
      await tester.enterText(find.byType(TextFormField).at(2), '15.5');
      await tester.enterText(find.byType(TextFormField).at(3), 'good yield');
      await tester.pump();

      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      expect(received, isNotNull);
      expect(received!['frames'], 10);
      expect(received!['half_frames'], 2);
      expect(received!['kilograms'], 15.5);
      expect(received!['notes'], 'good yield');
      expect(find.byType(HarvestFormScreen), findsNothing);
    });

    testWidgets(
        'shows an error and resets loading state when onSaveProposed throws',
        (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        apiClient,
        HarvestFormScreen(
          apiaryId: 1,
          hive: _hive,
          onSaveProposed: (args) async {
            throw Exception('boom');
          },
        ),
      ));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextFormField).at(0), '10');
      await tester.enterText(find.byType(TextFormField).at(2), '15.5');
      await tester.pump();

      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      expect(find.byType(HarvestFormScreen), findsOneWidget);
      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.byIcon(Icons.check), findsOneWidget);
    });
  });
}
