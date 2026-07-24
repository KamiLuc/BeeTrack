import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/dashboard/view/dashboard_screen.dart';
import 'package:app/features/hive/data/hive_model.dart';
import 'package:app/l10n/app_localizations.dart';

class _RecordingAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  final Map<int, DateTime> treatedAtByHive;

  _RecordingAdapter({this.treatedAtByHive = const {}});

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);

    if (options.path.contains('/invitations/count')) {
      return ResponseBody.fromString(
        jsonEncode({'count': 0}),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }

    if (options.path.contains('/treatments')) {
      final hiveId = int.parse(
        RegExp(r'/hives/(\d+)/treatments').firstMatch(options.path)!.group(1)!,
      );
      final treatedAt = treatedAtByHive[hiveId];
      final items = treatedAt == null
          ? []
          : [
              {
                'id': hiveId,
                'hive_id': hiveId,
                'treated_at': treatedAt.toUtc().toIso8601String(),
                'medicine_name': 'Oxalic acid',
                'dose': '1',
                'notes': '',
              }
            ];
      return ResponseBody.fromString(
        jsonEncode({'items': items, 'total': items.length}),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }

    return ResponseBody.fromString(
      jsonEncode({'items': [], 'total': 0}),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

Future<(ApiClient, _RecordingAdapter)> _fakeApiClient({
  Map<int, DateTime> treatedAtByHive = const {},
}) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient =
      ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
  final adapter = _RecordingAdapter(treatedAtByHive: treatedAtByHive);
  apiClient.dio.httpClientAdapter = adapter;
  return (apiClient, adapter);
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

const _hiveA = Hive(
  id: 1,
  apiaryId: 1,
  name: 'Alpha',
  type: 'langstroth',
  active: true,
  queenless: false,
  readyForHarvest: false,
  needsFood: false,
  gridRow: 0,
  gridCol: 0,
);

const _hiveB = Hive(
  id: 2,
  apiaryId: 1,
  name: 'Beta',
  type: 'langstroth',
  active: true,
  queenless: false,
  readyForHarvest: false,
  needsFood: false,
  gridRow: 0,
  gridCol: 1,
);

const _inactiveHive = Hive(
  id: 3,
  apiaryId: 1,
  name: 'Gamma',
  type: 'langstroth',
  active: false,
  queenless: false,
  readyForHarvest: false,
  needsFood: false,
  gridRow: 0,
  gridCol: 2,
);

void main() {
  group('DashboardScreen category selection', () {
    testWidgets('keeps the last selected category chip selected',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();
      await tester.pumpWidget(_wrap(
        apiClient,
        const DashboardScreen(apiaryId: 1, hives: [_hiveA]),
      ));
      await tester.pumpAndSettle();

      Future<void> tapChip(String label) async {
        await tester.tap(find.widgetWithText(FilterChip, label));
        await tester.pumpAndSettle();
      }

      await tapChip('Inspections');
      await tapChip('Feedings');
      await tapChip('Treatments');
      // Only "Harvests" remains selected; tapping it must not deselect it.
      await tapChip('Harvests');

      final chip = tester.widget<FilterChip>(
        find.widgetWithText(FilterChip, 'Harvests'),
      );
      expect(chip.selected, isTrue);
    });

    testWidgets('other categories can be deselected freely', (tester) async {
      final (apiClient, _) = await _fakeApiClient();
      await tester.pumpWidget(_wrap(
        apiClient,
        const DashboardScreen(apiaryId: 1, hives: [_hiveA]),
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(FilterChip, 'Inspections'));
      await tester.pumpAndSettle();

      final chip = tester.widget<FilterChip>(
        find.widgetWithText(FilterChip, 'Inspections'),
      );
      expect(chip.selected, isFalse);
    });
  });

  group('DashboardScreen hive selection', () {
    testWidgets('excludes inactive hives from the hive list', (tester) async {
      final (apiClient, _) = await _fakeApiClient();
      await tester.pumpWidget(_wrap(
        apiClient,
        const DashboardScreen(
          apiaryId: 1,
          hives: [_hiveA, _inactiveHive],
        ),
      ));
      await tester.pumpAndSettle();

      expect(find.text('Alpha'), findsOneWidget);
      expect(find.text('Gamma'), findsNothing);
    });

    testWidgets('disables Generate report when no hives are selected',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();
      await tester.pumpWidget(_wrap(
        apiClient,
        const DashboardScreen(apiaryId: 1, hives: [_hiveA]),
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(CheckboxListTile));
      await tester.pumpAndSettle();

      final button = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, 'Generate report'),
      );
      expect(button.onPressed, isNull);
    });
  });

  group('DashboardScreen report generation', () {
    testWidgets('filters records outside the selected date range',
        (tester) async {
      final now = DateTime.now();
      final today = DateTime(now.year, now.month, now.day);
      final (apiClient, _) = await _fakeApiClient(treatedAtByHive: {
        _hiveA.id: today.subtract(const Duration(days: 5)),
        _hiveB.id: today.subtract(const Duration(days: 20)),
      });
      await tester.pumpWidget(_wrap(
        apiClient,
        const DashboardScreen(apiaryId: 1, hives: [_hiveA, _hiveB]),
      ));
      await tester.pumpAndSettle();

      for (final label in ['Inspections', 'Feedings', 'Harvests']) {
        await tester.tap(find.widgetWithText(FilterChip, label));
        await tester.pumpAndSettle();
      }

      await tester.tap(find.widgetWithText(ElevatedButton, 'Generate report'));
      await tester.pumpAndSettle();

      // "Alpha" appears both in the hive checklist and the report heading;
      // "Beta" only appears once (the checklist) since it has no results.
      expect(find.text('Alpha'), findsNWidgets(2));
      expect(find.text('Beta'), findsOneWidget);
    });

    testWidgets('shows the no-results message when nothing matches',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();
      await tester.pumpWidget(_wrap(
        apiClient,
        const DashboardScreen(apiaryId: 1, hives: [_hiveA]),
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ElevatedButton, 'Generate report'));
      await tester.pumpAndSettle();

      expect(find.text('No records match these filters'), findsOneWidget);
    });
  });
}
